# P0-06 — Budget Policy

**Status:** Proposed — ratified by owner approval of the Phase 0 pull request
**Date:** 2026-07-13
**Roadmap reference:** [Phase 0 work item P0-06](../agentic-cicd-docs/07-implementation-roadmap.md)

## Authorized spend

| Category | Policy |
|---|---|
| Claude Max subscription | Existing; the primary Claude capacity source. No plan upgrade authorized without a project-owner decision. |
| ChatGPT Pro subscription | Existing; the primary Codex capacity source. Same rule. |
| Anthropic / OpenAI API | **$0. No automatic fallback** (CON-001, ADR-012). Enabling paid API use requires a new ADR with a budget ceiling. |
| GitHub Actions (hosted) | Within the account plan's included minutes only. Plan and included minutes are **[OWNER INPUT]** (Free = 2,000 min/month for private repos). No paid minute overage authorized; if minutes run out, work queues for the local runner or waits for the monthly reset. |
| GitHub storage (artifacts/packages) | Within plan allowance; artifact retention set per doc 06 §9.3 defaults to stay under it. |
| Local storage | Constrained — see R-039. Evidence and caches live on E: with quotas; no new hardware purchase is assumed by this policy (owner may choose to add storage as the R-039 remediation). |
| New infrastructure/services | $0 assumed. Any paid service (object storage, registry, observability SaaS) requires a project-owner decision recorded in an ADR. |

## Usage guardrails

- **Subscription pacing:** agent work is bounded by doc 04 §9.2 budgets (one writer + two reviewers, two repair cycles max). Competing implementations remain reserved for explicitly classified high-risk changes.
- **Hosted-minutes alert:** when GitHub-hosted usage passes ~50% of the monthly allowance, prefer deferring non-urgent CI until the local runner (P3) absorbs the load. Checked during the weekly runbook review.
- **`force_hosted` is opt-in** and visible (NFR-COST-002); no silent redispatch to hosted runners in the MVP (ADR-013).
- **Usage exhaustion is a pause, not a downgrade** — deterministic CI continues; agent lanes queue (doc 04 §9.2).

## KPI linkage

KPI-012 (unapproved API or hosted-compute spend) target remains **$0** — measured as: zero Anthropic/OpenAI API invoices and zero paid GitHub Actions overage across the pilot.
