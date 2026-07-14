# P0-05 — Roles and Owners

**Status:** Proposed — ratified by owner approval of the Phase 0 pull request
**Date:** 2026-07-13
**Roadmap reference:** [Phase 0 work item P0-05](../agentic-cicd-docs/07-implementation-roadmap.md)

## Named owners

All human roles defined in [charter §7](../agentic-cicd-docs/01-project-charter-and-requirements.md) are filled by **Aakash (GitHub: `Aakash1337`)** for the pilot. The logical responsibilities remain separate, and every decision records which logical role made it (see the [solo-operator provision](risk-tier-policy.md)).

| Role | Holder | Contact |
|---|---|---|
| Project owner | Aakash | GitHub `@Aakash1337` |
| Platform owner | Aakash | — |
| Repository maintainer | Aakash | — |
| Security owner | Aakash | — |
| Release owner | Aakash | — |
| Specification owner | Aakash | — |
| Evaluation owner | Aakash | — |
| Human reviewer | Aakash | — |

Non-human actors (agent broker, Claude Code, Codex) hold no approval authority per charter §7.

## Escalation

Doc 08 §2 escalation chains collapse to self-escalation for the pilot. The operative substitutes:

- **SEV-1/SEV-2 events:** stop the affected system first, then investigate (the runbook's containment-before-diagnosis ordering is the discipline that replaces a second responder).
- **Critical-tier risk acceptance:** 24-hour cooling-off before effect (defined in the risk-tier policy).
- Response-time targets in doc 08 §3 describe operating priority, not a staffed SLA (already stated there).

In an organizational deployment this table is re-populated with distinct people before unattended operation; that re-population is a named pre-condition in the ops runbook ("Names and contact methods MUST be filled in before unattended operation").
