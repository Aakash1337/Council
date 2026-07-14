# ci-pilot VM provisioning (VirtualBox, Windows 11 Home host)

Reproducible record of how the pilot local runner VM was created. This is the R-040 mitigation: the VM lifecycle is scripted and version-controlled rather than click-built. Host has no Hyper-V (Windows 11 Home), so VirtualBox 7.0.20 is the hypervisor (Phase 0 F-01/R-040).

## Parameters

| Setting | Value | Rationale |
|---|---|---|
| Name | `ci-pilot` | Matches the runner label |
| Base folder | `E:\CouncilCI\vms` | Disk decision (R-039); keeps VM off the full C: |
| ISO | `ubuntu-24.04.4-live-server-amd64.iso` (sha256 `e907d92e…8433`) | Verified against releases.ubuntu.com SHA256SUMS |
| vCPU / RAM | 4 / 8192 MB | Host is 12c/64GB; leaves ample headroom |
| Disk | 60 GB VDI, dynamically allocated | R-039 cap; grows only as used |
| NIC | NAT with host port-forward `127.0.0.1:2222 → 22` | No bridged LAN exposure; runner reaches GitHub outbound only |
| User | `runner` | Non-admin runner account (doc 06 §6.1) |

## Commands (Git Bash on the host)

```bash
VBM="/c/Program Files/Oracle/VirtualBox/VBoxManage.exe"

# SSH key for headless management (private key stays on the host, ~/.ssh)
ssh-keygen -t ed25519 -f ~/.ssh/council_ci_pilot -N "" -C "council-ci-pilot"

"$VBM" createvm --name ci-pilot --ostype Ubuntu_64 --register --basefolder "E:\CouncilCI\vms"
"$VBM" modifyvm ci-pilot --memory 8192 --cpus 4 --vram 16 \
  --graphicscontroller vmsvga --audio-driver none --usb off \
  --nic1 nat --natpf1 "ssh,tcp,127.0.0.1,2222,,22"
"$VBM" createmedium disk --filename "E:\CouncilCI\vms\ci-pilot\ci-pilot.vdi" --size 61440
"$VBM" storagectl ci-pilot --name SATA --add sata --controller IntelAhci --portcount 2
"$VBM" storageattach ci-pilot --storagectl SATA --port 0 --device 0 --type hdd \
  --medium "E:\CouncilCI\vms\ci-pilot\ci-pilot.vdi"

# Unattended install: injects the SSH public key and enables sshd.
# NOTE: VBox 7.0 uses --password (not --user-password).
PUBKEY=$(cat ~/.ssh/council_ci_pilot.pub)
"$VBM" unattended install ci-pilot \
  --iso="E:\CouncilCI\iso\ubuntu-24.04.4-live-server-amd64.iso" \
  --user=runner --password="<generated>" --full-user-name="CIRunner" \
  --hostname=ci-pilot.council.local --time-zone=UTC \
  --post-install-command="apt-get update && apt-get install -y openssh-server && install -d -m 700 -o runner -g runner /home/runner/.ssh && echo \"$PUBKEY\" > /home/runner/.ssh/authorized_keys && chown runner:runner /home/runner/.ssh/authorized_keys && chmod 600 /home/runner/.ssh/authorized_keys && systemctl enable ssh"
"$VBM" startvm ci-pilot --type headless
```

The console recovery password is generated randomly and stored **only** at `E:\CouncilCI\vms\ci-pilot-console-recovery.txt` on the host (never in the repo). SSH key auth is the normal access path; the password is console-only fallback.

## Snapshot policy

After the runner is registered and network-locked (see `harden.sh`), take a golden snapshot:

```bash
"$VBM" snapshot ci-pilot take golden-runner-v1 --description "Registered runner, ufw locked, tools bootstrapped"
```

A suspected-compromise response (ops runbook §9) restores this snapshot rather than cleaning in place.
