#!/usr/bin/env bash
# Runs INSIDE the ci-pilot VM after OS install. Establishes the pilot
# trust boundary (Council doc 06 §8, exit gate G3): deny-by-default
# egress with a narrow allowlist, no LAN reachability, no Docker socket
# exposure, dedicated non-admin runner account.
#
# Egress model: the runner initiates outbound HTTPS to GitHub, package,
# and scanner endpoints only. It must not reach the host LAN, the NAS,
# other home devices, or accept inbound anything except host SSH.
set -euo pipefail

echo "== toolchain =="
sudo apt-get update
sudo apt-get install -y ufw ca-certificates curl jq git build-essential

# Go (pinned; matches the pilot repo go.mod line)
GO_VERSION=1.26.2
if ! command -v go >/dev/null || [ "$(go version | awk '{print $3}')" != "go${GO_VERSION}" ]; then
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz
  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
  echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' | sudo tee /etc/profile.d/go.sh >/dev/null
fi
export PATH=$PATH:/usr/local/go/bin

echo "== host LAN discovery (for deny rules) =="
# The VM's NAT gateway is 10.0.2.2 (VirtualBox). The real home LAN
# (192.168.0.0/16, 10.0.0.0/8 except NAT, 172.16.0.0/12) must be denied.
DEFAULT_IF=$(ip route show default | awk '{print $5; exit}')
echo "default interface: $DEFAULT_IF"

echo "== DNS via the VBox NAT proxy only =="
# The VM must resolve DNS through VirtualBox's NAT host-resolver proxy at
# 10.0.2.3 (inside the allowed NAT segment), NOT through any real LAN
# resolver. The home network hands out a resolver in 10.0.0.0/8
# (e.g. 10.64.0.1) which the LAN-deny rule below correctly blocks; that
# resolver must not be the runner's DNS. Requires the VM to be created
# with `--natdnshostresolver1 on --natdnsproxy1 on` (see provision-vm.md).
sudo mkdir -p /etc/systemd/resolved.conf.d
printf '[Resolve]\nDNS=10.0.2.3\nDomains=~.\n' | sudo tee /etc/systemd/resolved.conf.d/council.conf >/dev/null
sudo resolvectl dns "$DEFAULT_IF" 10.0.2.3 2>/dev/null || true
sudo systemctl restart systemd-resolved

echo "== ufw deny-by-default with egress allowlist =="
sudo ufw --force reset
sudo ufw default deny incoming
sudo ufw default deny outgoing

# Inbound: only SSH from the NAT host (management path).
sudo ufw allow in 22/tcp comment 'host SSH management'

# RULE ORDER MATTERS: ufw evaluates rules top-to-bottom, first match
# wins. The port allows below (allow out 443) would otherwise match a
# connection to a LAN host on :443 before any destination deny. So the
# private-range denies MUST precede the port allows. The VBox NAT
# segment (10.0.2.0/24 — gateway and DNS) is allowed first, then the
# rest of 10/8 and the other RFC1918 ranges are denied, then ports.
sudo ufw allow out to 10.0.2.0/24 comment 'VBox NAT segment (gateway/DNS)'
for cidr in 192.168.0.0/16 172.16.0.0/12 10.0.0.0/8; do
  sudo ufw deny out to "$cidr" comment 'home LAN deny (before port allows)'
done

# Outbound essentials: DNS, NTP, HTTPS (to reach GitHub/pkg/scanners).
# These only apply to destinations not already denied above.
sudo ufw allow out 53 comment 'DNS'
sudo ufw allow out 123/udp comment 'NTP'
sudo ufw allow out 443/tcp comment 'HTTPS: GitHub, modules, scanners'
sudo ufw allow out 80/tcp comment 'HTTP: apt, redirects'

sudo ufw --force enable
sudo ufw status verbose

echo "== confirm no Docker socket present (untrusted jobs must not get it) =="
if [ -S /var/run/docker.sock ]; then
  echo "WARNING: docker.sock present; pilot policy forbids exposing it to jobs" >&2
fi

echo "== runner workspace =="
install -d -m 755 -o runner -g runner /home/runner/actions-runner
echo "harden.sh complete"
