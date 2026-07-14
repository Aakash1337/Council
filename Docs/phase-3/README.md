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

## VM provisioning status

The VM install is driven by VirtualBox unattended mode. **Known risk being validated:** VBox 7.0's Ubuntu template uses the legacy debian-installer preseed path, which Ubuntu 24.04's subiquity installer may not fully honor. If the unattended install does not produce a reachable SSH host, the fallback is a cloud-init `autoinstall` seed ISO (documented in `provision-vm.md` once exercised). This file is updated with the concrete drill evidence once the VM is reachable.

## R-039 disk check (gates closure)

After the VM and a first local run, record E: free space; R-039 closes only if headroom is acceptable, else capacity is added.
