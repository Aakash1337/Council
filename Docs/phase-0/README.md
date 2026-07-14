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

| # | Action | Needed before |
|---|---|---|
| 1 | Flip `CustomDNS` to private | P1 (CI work on the pilot) |
| 2 | Confirm or correct the `httpserver`/`cybiccrm` classification assumption | G0 ratification |
| 3 | Verify both providers' model-improvement controls; record date | P4 (agent lane) |
| 4 | Free or add ~150 GB disk (R-039) | P3 (local runner) |
| 5 | Confirm/supply: router VLAN capability, NAS presence, UPS status; confirm no separate home server | P3 |
| 6 | Verify GitHub plan and monthly Actions minutes | P2 (hosted CI volume planning) |
