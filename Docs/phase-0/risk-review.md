# P0-07 — Initial Risk Register Review and Assumption Validation

**Status:** Proposed — ratified by owner approval of the Phase 0 pull request
**Date:** 2026-07-13
**Roadmap reference:** [Phase 0 work item P0-07](../agentic-cicd-docs/07-implementation-roadmap.md)

## 1. Assumption validation (charter §12)

G0 requires each assumption validated or converted to a risk/open item.

| ID | Assumption | Result |
|---|---|---|
| ASM-001 | Pilot repository is private and owner-controlled | **Open item** — CustomDNS is owner-controlled but currently public; owner action 1 in the [pilot decision](pilot-decision.md) converts this to validated |
| ASM-002 | Home server can host a dedicated Linux VM | **Validated with deviation** — no separate server; the workstation hosts a VirtualBox Ubuntu VM instead (F-01/R-038); disk capacity must be remediated first (F-02/R-039) |
| ASM-003 | Reliable outbound HTTPS; no inbound port needed | **Validated** — standard home broadband, 1 Gbps LAN; runners are outbound-only |
| ASM-004 | Subscriptions active and permit first-party client use | **Validated for supervised use** — both clients installed, current, and operational; unattended qualification deliberately deferred to P4-09/PAC-018 |
| ASM-005 | Owner accepts subscription pauses | **Validated** — accepted in project direction; budget policy encodes it |
| ASM-006 | GitHub remains the control plane | **Validated** — Council and pilot repos live on GitHub |
| ASM-007 | Pilot has meaningful automated tests | **Validated** — CustomDNS has test files across config, resolver, and server packages |
| ASM-008 | Human review available for disputed findings | **Validated with provision** — solo operator; substitution and cooling-off rules in the [risk-tier policy](risk-tier-policy.md) |

## 2. New risks discovered in Phase 0

Added to the [risk register](../agentic-cicd-docs/09-risk-register-and-decisions.md) summary table:

| ID | Risk | Inherent | Treatment | Residual |
|---|---|---:|---|---:|
| R-038 | Workstation colocation: daily-driver machine also hosts runner VM and agent broker, weakening trust-zone separation | 4/4/16 | VirtualBox VM boundary for the runner; dedicated OS accounts for agent clients (P4); host-only/NAT VM networking; red-team reachability tests (ACT-004) before agent credentials; revisit if a separate server is added | 2/3/6 |
| R-039 | Disk exhaustion: <30 GB free on C:, ~116 GB on E: cannot hold runner VM + caches + evidence | 4/3/12 | Owner frees/adds ~150 GB before P3; quotas on caches and evidence; retention lifecycle from day one | 2/2/4 |
| R-040 | VirtualBox dependency on Windows 11 Home (no Hyper-V): community update cadence, manual lifecycle automation | 3/3/9 | Pin VirtualBox version; scripted VM template (Vagrant or VBoxManage); test updates before adopting; revisit if host OS upgrades to Pro | 2/2/4 |

## 3. Review of existing high/critical inherent risks

The pilot-relevant subset, with Phase 0 dispositions:

| Risk | Phase 0 disposition |
|---|---|
| R-001 (PR code obtains AI credentials) | Controls unchanged and achievable on this host **only after** dedicated OS accounts and VM isolation exist; until P3/P4, no unattended agent execution happens at all, so exposure window is nil |
| R-002/R-003/R-028 (runner persistence/escape/LAN movement) | VirtualBox VM + host-only networking is the pilot boundary; full VLAN segmentation deferred to hardened v1 as docs already state; ACT-004 reachability tests remain gating for G4 |
| R-004 (prompt injection) | No change; becomes operative at P4 |
| R-008 (false consensus) | No change; research instrumentation (doc 04 §13) now strengthens detection |
| R-012 (agent edits protected paths) | Path list now concrete — see [risk-tier policy](risk-tier-policy.md) §1 protected paths |
| R-014 (evidence leaks secrets) | Note: Council repo is public — evidence from the pilot must never be committed or published there; pilot evidence stays in the private pilot repo/local store |
| R-018 (Action supply chain) | Unchanged; enforced from the first `pr.yml` in P2 |
| R-025 (complexity prevents adoption) | Phase 0 kept to seven focused documents; thin-slice build order (P1→P2→P4) reaffirmed |
| R-037 (unauthorized provider transmission) | Now concrete: classification table in the [pilot decision](pilot-decision.md); `cybiccrm` material explicitly out of scope |

No risk's residual score rises to ≥17 (critical); no capability is blocked from proceeding.

## 4. Recommendation

Phase 0 finds no blocker to beginning Phase 1 immediately. The two items that must land before their respective phases: CustomDNS visibility flip (before P1 CI work meaningfully starts) and disk remediation (before P3).
