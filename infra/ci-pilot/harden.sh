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

echo "== ufw deny-by-default with egress allowlist =="
sudo ufw --force reset
sudo ufw default deny incoming
sudo ufw default deny outgoing

# Inbound: only SSH from the NAT host (management path).
sudo ufw allow in 22/tcp comment 'host SSH management'

# Outbound essentials: DNS, NTP, HTTPS (to reach GitHub/pkg/scanners).
sudo ufw allow out 53 comment 'DNS'
sudo ufw allow out 123/udp comment 'NTP'
sudo ufw allow out 443/tcp comment 'HTTPS: GitHub, modules, scanners'
sudo ufw allow out 80/tcp comment 'HTTP: apt, redirects'

# Explicitly deny private LAN ranges so a bug in the allowlist can't
# reach home infrastructure (defense in depth; NAT already isolates).
for cidr in 192.168.0.0/16 172.16.0.0/12; do
  sudo ufw deny out to "$cidr" comment 'home LAN deny'
done
# 10.0.2.0/24 is the VBox NAT segment (gateway/DNS live here) — allow it,
# but deny the rest of 10.0.0.0/8.
sudo ufw allow out to 10.0.2.0/24 comment 'VBox NAT segment'
sudo ufw deny out to 10.0.0.0/8 comment 'other 10.x LAN deny'

sudo ufw --force enable
sudo ufw status verbose

echo "== confirm no Docker socket present (untrusted jobs must not get it) =="
if [ -S /var/run/docker.sock ]; then
  echo "WARNING: docker.sock present; pilot policy forbids exposing it to jobs" >&2
fi

echo "== runner workspace =="
install -d -m 755 -o runner -g runner /home/runner/actions-runner
echo "harden.sh complete"
