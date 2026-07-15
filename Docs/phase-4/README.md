# Phase 4 — Supervised Two-Model Workflow

**Status:** Auth unblocked 2026-07-15 — Codex reviewer proven live; Claude unattended auth qualified ([P4-09/PAC-018](p4-09-auth-qualification.md)); real two-model review is the remaining step to close G4
**Roadmap reference:** [Phase 4 work items](../agentic-cicd-docs/07-implementation-roadmap.md)

Phase 4 validates the review protocol with the real first-party clients before the unattended broker (P5) depends on it.

## Work-item disposition

| Work item | Deliverable | Status |
|---|---|---|
| P4-01 Authenticate Claude Code | `CLAUDE_CODE_OAUTH_TOKEN` (setup-token) | **Done 2026-07-15** — token set by owner; unattended auth qualified (P4-09). |
| P4-02 Authenticate Codex | — | **Done** — `codex exec` runs headless and authenticated. |
| P4-03 Codex-for-Claude plugin | Supervised interactive path | Deferred (needs Claude auth) |
| P4-04 Review JSON schema | `agent-review.schema.json` (from P5 broker) | Done |
| P4-05 Evidence bundle | Broker `freeze` | Done (P5) |
| P4-06 Adversarial review with seeded defect | Real Codex review of a seeded data race (**detection** proven; repair+reconfirm pending) | **Partial — see below** |
| P4-07 Iteration budgets | Time/turn limits in `CLIRunner` | Done (P5) |
| P4-08 Review-quality metrics | Detection recorded on the seeded corpus | Partial (single case) |
| P4-09 Unattended-auth qualification | `claude setup-token` spike (PAC-018) | **PASSED 2026-07-15** — [qualification record](p4-09-auth-qualification.md): 12/12 pass@1, all failure drills fail closed |

## Live evidence: real Codex review of a seeded defect (P4-06 / G4)

A seeded concurrency defect ([`broker/testdata/p4-seeded/cache.go`](../../broker/testdata/p4-seeded/cache.go)) — a DNS response cache whose map is read and written from concurrent goroutines with no synchronization — was reviewed by the **real Codex CLI** (`codex exec`, read-only sandbox, ~12k tokens):

- **Codex found the exact seeded defect:** a high-severity `concurrency` finding — "concurrent Get/Put on the cache map is a data race" — with a reproducer (run goroutines calling Get/Put under `-race`). It also surfaced three secondary issues (unbounded growth, shared byte-slice mutation). This is the G4 "seeded high-severity defect is found" criterion.
- **Raw output needed one schema repair** (doc 04 §7.2): the model emitted `verdict: "request_changes"` and `reproducer.kind: "race_test"`, outside the schema's controlled vocabulary. A bounded normalization (documented synonym map — substance unchanged) mapped them to `changes_required` / `test`; the repaired review then **validates against the strict `agent-review.schema.json`**. Both raw and repaired artifacts are retained.

Artifacts: [`review-prompt.md`](../../broker/testdata/p4-seeded/review-prompt.md), [`codex-review.json`](../../broker/testdata/p4-seeded/codex-review.json) (raw), [`codex-review-repaired.json`](../../broker/testdata/p4-seeded/codex-review-repaired.json) (schema-valid).

## The Claude-auth finding (resolved 2026-07-15)

Earlier, probing both CLIs headlessly showed `codex exec` authenticated but `claude -p` returning `OAuth session expired and could not be refreshed` — precisely the **agent-unavailable** state the platform is designed for (doc 04 §9, ops runbook §11, risk R-005), during which deterministic CI stayed authoritative and the broker correctly yielded `pending` (never treating one provider as consensus).

**Resolved:** the owner ran `claude setup-token` and set `CLAUDE_CODE_OAUTH_TOKEN`. The [P4-09 qualification](p4-09-auth-qualification.md) then passed all drills (12/12 structured runs, invalid-token/timeout fail closed, redaction clean, persistence confirmed). The broker may now run `--real` Claude reviews, so the agent-unavailable path returns to being an exception rather than the steady state.

## Exit gate G4

| Criterion | Status |
|---|---|
| A reviewer produces schema-valid structured findings | **Met** (Codex, after one bounded repair) |
| A seeded high-severity defect is **found** | **Met** (Codex found the data race) |
| The seeded defect is **repaired** and re-review confirms the fix | **Not met** — no repair/final-review artifact yet; the roadmap's G4 requires found *and* repaired, so this gate is not closed |
| The loop stops on budget/no-progress | Met (P5 budgets) |
| Auth artifacts never appear in repos/logs/build runners | Met (broker handles no tokens; runner credential-free per G3) |
| Both models review independently | **Blocked** on Claude auth (owner P4-09) |

**G4 is not yet closed.** The Codex reviewer and protocol mechanics are proven (detection of the seeded defect, schema validation, single-provider `pending` handling), but two of the gate's criteria remain open: the repair-and-reconfirm loop, and independent two-model review — both of which need the Claude reviewer, i.e. the owner's `claude setup-token` action (P4-09). This page must not be used to close P4 until those land.
