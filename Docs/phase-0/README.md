# Phase 0 — Discovery and Threat Model

Deliverables for [roadmap Phase 0](../agentic-cicd-docs/07-implementation-roadmap.md), produced 2026-07-13. Approval of the Phase 0 pull request by the project owner ratifies the proposed decisions.

| Work item | Deliverable | Status |
|---|---|---|
| P0-01 Inventory infrastructure | [Environment inventory](environment-inventory.md) | Complete (owner-input items marked) |
| P0-02 Select pilot | [Pilot decision record](pilot-decision.md) — `Aakash1337/CustomDNS` | Proposed |
| P0-03 Classify code trust | [Trust classification and execution matrix](trust-classification.md) | Proposed |
| P0-04 Define change risk levels | [Risk-tier policy](risk-tier-policy.md) | Proposed |
| P0-05 Establish owners | [Roles and owners](owners.md) | Proposed |
| P0-06 Define budget policy | [Budget policy](budget-policy.md) | Proposed |
| P0-07 Review initial risk register | [Risk review and assumption validation](risk-review.md); R-038–R-040 added to the register | Proposed |

## Exit gate G0 checklist

- [x] Pilot repository selected — `Aakash1337/CustomDNS` (proposed; ratified by PR approval)
- [x] Server and network inventory complete — with three **[OWNER INPUT]** items: router VLAN capability, NAS presence, UPS
- [x] Public/fork code policy explicitly decided — GitHub-hosted only, never the persistent local runner
- [x] Required owners named — all logical roles held by `Aakash1337` with the solo-operator provision
- [x] Initial threat model approved — risk review complete; approval occurs via this PR
- [x] No unresolved critical assumption blocks implementation — ASM-001 converts to a pre-P1 owner action; all others validated

## Owner actions carried forward

| # | Action | Needed before | Status |
|---|---|---|---|
| 1 | Flip `CustomDNS` to private | P1 | **Done** 2026-07-14 |
| 2 | Confirm or correct the `httpserver`/`cybiccrm` classification assumption | G0 | **Confirmed** — not the owner's repo; disqualification stands |
| 3 | Verify both providers' model-improvement controls; record date | P4 | **Verified** 2026-07-14 — disabled since account creation; quarterly recheck 2026-10-14 |
| 4 | Free or add ~150 GB disk (R-039) | P3 | **Decision recorded, risk stays open** — location is E: (116 GB free). Not capacity-resolved: the doc-06 VM sizing alone is 60–120 GB. Mitigation: pilot VM capped at a 60 GB dynamically-allocated disk, cache/evidence quotas, free-space monitoring; R-039 moves to *Mitigating* and closes only when measured headroom exists (or capacity is added) before P3 exit |
| 5 | Confirm/supply: router VLAN capability, NAS presence, UPS status; separate home server | P3 | **Answered** 2026-07-14 — Xfinity default gateway (no VLANs); NAS dual-boots on the future server box; no UPS; server migration planned post-build |
| 6 | Verify GitHub plan and monthly Actions minutes | P2 | **Verified** — Free plan, 2,000 min + 0.5 GB/month |
| 7 | **decide merge-protection approach (R-041/F-03):** GitHub Free cannot enforce branch protection on the private pilot repo. | P2 exit (G2) | **Decided 2026-07-15** — Option B (process-discipline waiver) for CustomDNS; Council (public) got enforced protection. See [decision #7 record](../phase-4/decision-07-repo-strategy.md). Precondition: enforced protection before any unattended agent has merge access. |
