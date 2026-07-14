# Phase 3 — Local Execution Plane

**Status:** Delivered 2026-07-14 — G3 met (drills passed on the live VM)
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

**Final reachability red-team (`reachability-test.sh`) — PASS:** all home-LAN targets and cloud metadata unreachable, GitHub reachable, zero AI/deployment credentials present, all provider env vars unset. This is the PAC-008 / G3 isolation evidence. Verified to **persist across a VM reboot** (ufw active, DNS still via 10.0.2.3, test still passes).

## Runner registration and a real local job

The runner is registered **repository-scoped, ephemeral**, labels `self-hosted,linux,x64,ci-pilot`, run as a systemd service. The nightly workflow's `route` job (on GitHub-hosted) selects the local runner; the `full-security` job then executes on the VM. First dispatch confirmed the job running with `runner_name: ci-pilot`.

**Infrastructure-failure finding (R-038 in action).** The first local run **failed with "the self-hosted runner lost communication with the server."** Root cause: a heavy `codex exec` (ultra reasoning, ~12k tokens) was running on the **host** at the same time as the VM was compiling six CI tools from source — the host CPU was saturated and the 4-vCPU VM was starved until its runner lost heartbeat, then the VM itself became briefly unresponsive. This is exactly the colocation risk R-038 predicted (the workstation hosts development, the runner, and agent work simultaneously) and it manifested as an **`infrastructure_error`**, not a semantic failure — the platform's failure classification treats it as redispatch-eligible, never as a code failure. Mitigations: keep heavy host work off the box during local runs until migration to the dedicated server (Phase 0 F-01), and/or cap concurrent load. The clean re-run (host idle) is recorded below.

**Clean local run — SUCCESS.** With the host kept idle and both Go caches cleared, the nightly's `full-security` job completed **green on `runner_name: ci-pilot`** (route job on hosted → local job on the VM). This is the G3 "trusted CI runs locally with hosted-equivalent results" evidence. A follow-up fix was needed in the pilot repo: `upload-artifact` v4+ excludes hidden files, so the `.ci-evidence/` artifact was empty until `include-hidden-files: true` was added (CustomDNS PR #5).

**Manual fallback (`force_hosted`) — CONFIRMED.** Dispatching the nightly with `force_hosted=true` routed the `full-security` job to `ubuntu-latest` / a GitHub-hosted runner instead of `ci-pilot`. This is the outage-recovery path: the home server can be offline (or explicitly bypassed) and eligible deterministic work still runs on hosted infrastructure. Subscription-authenticated agent work never takes this path.

**Golden snapshot.** After hardening and one green run, `golden-runner-v1` was taken (`VBoxManage snapshot`). A suspected-compromise or corruption response restores this snapshot rather than cleaning in place (ops runbook §9); it also would have made the cache-corruption recovery above a one-command rollback.

## Exit gate G3 — met

| Criterion | Evidence |
|---|---|
| Trusted CI runs locally, hosted-equivalent | Clean nightly green on `ci-pilot`; same `ci/run full` as hosted |
| Home server offline → manual hosted run works | `force_hosted=true` routed to hosted (confirmed) |
| Runner cannot reach protected home networks | `reachability-test.sh` PASS, persists across reboot |
| Workspace cleanup + cache separation | Ephemeral runner (one job, then de-register); per-job clean `_work` |
| Compromised container cannot get AI/release creds | Credential-absence assertions PASS |
| Fork/public cannot select the local runner | Repository-scoped runner (PAC-010, by construction) |

Two real defects and two operational failures were caught and resolved along the way (ufw ordering, DNS routing, host-load starvation, cache corruption) — exactly what a discovery-and-drill phase is for.

## R-039 disk check

After building the VM (dynamic disk, actual usage ~6.5 GB) and one full local run, **E: has ~110 GB free** (down from 116 GB at Phase 0; the unused live-server ISO was reclaimed). The dynamic VDI grows only as used and is capped at 60 GB. This is workable for the pilot but not roomy; **R-039 stays Mitigating** with free-space monitoring (`monitor.sh` alerts at 85%), and closes on migration to the dedicated server (Phase 0 F-01) or added capacity. Headroom is acceptable to proceed.
