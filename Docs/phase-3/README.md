# Phase 3 — Local Execution Plane

**Status:** In progress 2026-07-14
**Roadmap reference:** [Phase 3 work items](../agentic-cicd-docs/07-implementation-roadmap.md)

Moves trusted deterministic workloads onto a home-hosted runner while preserving the GitHub-hosted recovery path. On this host (Windows 11 Home, no Hyper-V) the isolation unit is a VirtualBox VM (Phase 0 R-040).

## Work-item disposition

| Work item | Deliverable | Status |
|---|---|---|
| P3-01 Provision dedicated VM | `ci-pilot` Ubuntu 24.04.4 VM on E:, scripted in [`infra/ci-pilot/provision-vm.md`](../../infra/ci-pilot/provision-vm.md) | See status below |
| P3-02 Segment network | [`harden.sh`](../../infra/ci-pilot/harden.sh): ufw deny-by-default, egress allowlist (443/80/53/123), LAN denies | Scripted |
| P3-03 Register scoped runner | [`register-runner.md`](../../infra/ci-pilot/register-runner.md): repository-scoped, ephemeral, `ci-pilot` labels | Scripted |
| P3-04 Runner labels | `self-hosted,linux,x64,ci-pilot` | Scripted |
| P3-05 Ephemeral workspace | `--ephemeral` runner + per-job clean `_work` | Scripted |
| P3-06 Cache namespaces | Trust-tagged cache keys (deferred detail; pilot uses hosted cache only until local object store in P6) | Partial |
| P3-07 Manual fallback | `force_hosted` input on the local-routed workflow | In CustomDNS |
| P3-08 Monitoring | [`monitor.sh`](../../infra/ci-pilot/monitor.sh): disk/mem/load/runner/GitHub, threshold alerts | Scripted |
| P3-09 Failure drills | Offline-server and interrupted-job drills; credential red-team [`reachability-test.sh`](../../infra/ci-pilot/reachability-test.sh) | See drills below |

## Exit gate G3

| Criterion | Evidence | Status |
|---|---|---|
| Trusted CI runs locally with hosted-equivalent results | Same commit, local vs hosted normalized evidence compared | pending VM |
| Home server can be powered off without blocking a manual hosted run | `force_hosted` drill | pending VM |
| Runner cannot reach protected home networks | `reachability-test.sh` PASS (PAC-008) | pending VM |
| Workspace cleanup + cache separation verified | ephemeral runner re-register between jobs | pending VM |
| Compromised test container cannot obtain AI/release credentials | credential-absence assertions in reachability test | pending VM |
| Fork/public cannot select the local runner | repository-scoped runner (PAC-010) | by construction |

## Provisioning outcome and drill evidence

**VM built (2026-07-14):** Ubuntu 24.04.4 LTS, 4 vCPU / 8 GB / 58 GB root, on E:. Access via SSH key as `runner`.

**Provisioning pivot (recorded honestly).** The first attempt used VirtualBox `unattended install` against the live-server ISO. It failed exactly as the risk noted: VBox 7.0's legacy preseed does not drive Ubuntu 24.04's subiquity installer, which stalled at an interactive prompt and never created the account (SSH responded — the installer's own sshd — but no credential worked). The reliable path that succeeded: boot the official **cloud image** and configure it from a NoCloud cloud-init seed (see `provision-vm.md`, `seed/`). No installer involved.

**Two real bugs the drills caught (the point of testing):**

1. **ufw rule ordering.** The first `harden.sh` placed `allow out 443/tcp` before the LAN denies. Because ufw is first-match, a connection to a LAN host on `:443` matched the port allow before the deny — the reachability test flagged `10.0.0.1:443` still reachable. Fixed by ordering the private-range denies *before* the port allows.
2. **DNS via a real LAN resolver.** The home network hands out a resolver at `10.64.0.1` (inside 10/8). The correct LAN-deny then broke DNS, because that denied resolver was the only working one (the public `75.75.75.75` is unreachable from the NAT'd guest). Fixed architecturally: enable VirtualBox's NAT DNS host-resolver proxy (`--natdnshostresolver1 on`) and point the guest at `10.0.2.3` (inside the allowed NAT segment). DNS now resolves through the hypervisor proxy; no real LAN host is reachable.

**Final reachability red-team (`reachability-test.sh`) — PASS:** all home-LAN targets and cloud metadata unreachable, GitHub reachable, zero AI/deployment credentials present, all provider env vars unset. This is the PAC-008 / G3 isolation evidence.

## R-039 disk check (gates closure)

After the VM and a first local run, record E: free space; R-039 closes only if headroom is acceptable, else capacity is added.
