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
| NIC | NAT with host port-forward `127.0.0.1:2222 → 22`, `--natdnshostresolver1 on` | No bridged LAN exposure; DNS resolves via the NAT proxy (10.0.2.3) so the runner never queries a real LAN resolver |
| User | `runner` | Non-admin runner account (doc 06 §6.1) |

## Method: cloud image + NoCloud seed (what actually worked)

The VirtualBox `unattended install` path against the 24.04 **live-server ISO** was tried first and abandoned: VBox 7.0's Ubuntu template drives the legacy debian-installer preseed, which Ubuntu 24.04's subiquity installer does not fully honor — the installer stalled at an interactive prompt and never created the account. The reliable path is the official **cloud image**, which boots directly and configures itself from a NoCloud cloud-init seed (no installer at all).

```bash
VBM="/c/Program Files/Oracle/VirtualBox/VBoxManage.exe"

# 1. SSH key for headless management (private key stays on the host).
ssh-keygen -t ed25519 -f ~/.ssh/council_ci_pilot -N "" -C "council-ci-pilot"

# 2. NoCloud seed: user-data (with the pubkey) + meta-data -> CIDATA ISO.
#    Seed files live in E:\CouncilCI\seed\; build seed.iso with volume
#    label CIDATA (a tiny Go iso9660 writer is used since the host has no
#    genisoimage/xorriso). See seed/user-data in this directory.

# 3. Cloud image (QCOW2) -> VDI. VirtualBox reads QCOW2 natively via
#    clonemedium; convertfromraw does NOT (it would treat QCOW2 as raw).
"$VBM" clonemedium disk "E:\CouncilCI\iso\noble-server-cloudimg-amd64.img" \
  "E:\CouncilCI\vms\ci-pilot.vdi" --format VDI
"$VBM" modifymedium disk "E:\CouncilCI\vms\ci-pilot.vdi" --resize 61440

# 4. VM around the disk; EFI firmware (cloud image is EFI). NAT with the
#    DNS host-resolver proxy so DNS stays inside the allowed segment.
"$VBM" createvm --name ci-pilot --ostype Ubuntu_64 --register --basefolder "E:\CouncilCI\vms"
"$VBM" modifyvm ci-pilot --memory 8192 --cpus 4 --vram 16 --firmware efi \
  --graphicscontroller vmsvga --audio-driver none --usb off \
  --nic1 nat --natpf1 "ssh,tcp,127.0.0.1,2222,,22" \
  --natdnshostresolver1 on --natdnsproxy1 on
"$VBM" storagectl ci-pilot --name SATA --add sata --controller IntelAhci --portcount 2
"$VBM" storageattach ci-pilot --storagectl SATA --port 0 --device 0 --type hdd \
  --medium "E:\CouncilCI\vms\ci-pilot.vdi"
"$VBM" storageattach ci-pilot --storagectl SATA --port 1 --device 0 --type dvddrive \
  --medium "E:\CouncilCI\seed\seed.iso"   # cloud-init consumes this on first boot
"$VBM" startvm ci-pilot --type headless
```

Cloud-init installs sshd and injects the key on first boot (~1–2 min). The `runner` account has passwordless sudo and a locked password (key-only login). No console recovery password is needed with this method.

## Snapshot policy

After the runner is registered and network-locked (see `harden.sh`), take a golden snapshot:

```bash
"$VBM" snapshot ci-pilot take golden-runner-v1 --description "Registered runner, ufw locked, tools bootstrapped"
```

A suspected-compromise response (ops runbook §9) restores this snapshot rather than cleaning in place.
