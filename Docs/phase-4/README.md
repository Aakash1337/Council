# Phase 4 — Supervised Two-Model Workflow

**Status:** Partially delivered 2026-07-14 — Codex reviewer proven live; Claude reviewer blocked on expired auth (owner action)
**Roadmap reference:** [Phase 4 work items](../agentic-cicd-docs/07-implementation-roadmap.md)

Phase 4 validates the review protocol with the real first-party clients before the unattended broker (P5) depends on it.

## Work-item disposition

| Work item | Deliverable | Status |
|---|---|---|
| P4-01 Authenticate Claude Code | — | **Blocked** — the `claude` CLI's OAuth session is expired ("could not be refreshed"). Owner must re-login or run `claude setup-token` (P4-09). |
| P4-02 Authenticate Codex | — | **Done** — `codex exec` runs headless and authenticated. |
| P4-03 Codex-for-Claude plugin | Supervised interactive path | Deferred (needs Claude auth) |
| P4-04 Review JSON schema | `agent-review.schema.json` (from P5 broker) | Done |
| P4-05 Evidence bundle | Broker `freeze` | Done (P5) |
| P4-06 Adversarial review with seeded defect | Real Codex review of a seeded data race | **Done — see below** |
| P4-07 Iteration budgets | Time/turn limits in `CLIRunner` | Done (P5) |
| P4-08 Review-quality metrics | Detection recorded on the seeded corpus | Partial (single case) |
| P4-09 Unattended-auth qualification | `claude setup-token` spike (PAC-018) | **Owner action** — see below |

## Live evidence: real Codex review of a seeded defect (P4-06 / G4)

A seeded concurrency defect ([`broker/testdata/p4-seeded/cache.go`](../../broker/testdata/p4-seeded/cache.go)) — a DNS response cache whose map is read and written from concurrent goroutines with no synchronization — was reviewed by the **real Codex CLI** (`codex exec`, read-only sandbox, ~12k tokens):

- **Codex found the exact seeded defect:** a high-severity `concurrency` finding — "concurrent Get/Put on the cache map is a data race" — with a reproducer (run goroutines calling Get/Put under `-race`). It also surfaced three secondary issues (unbounded growth, shared byte-slice mutation). This is the G4 "seeded high-severity defect is found" criterion.
- **Raw output needed one schema repair** (doc 04 §7.2): the model emitted `verdict: "request_changes"` and `reproducer.kind: "race_test"`, outside the schema's controlled vocabulary. A bounded normalization (documented synonym map — substance unchanged) mapped them to `changes_required` / `test`; the repaired review then **validates against the strict `agent-review.schema.json`**. Both raw and repaired artifacts are retained.

Artifacts: [`review-prompt.md`](../../broker/testdata/p4-seeded/review-prompt.md), [`codex-review.json`](../../broker/testdata/p4-seeded/codex-review.json) (raw), [`codex-review-repaired.json`](../../broker/testdata/p4-seeded/codex-review-repaired.json) (schema-valid).

## The Claude-auth finding (why two-model is blocked)

Probing both CLIs headlessly:

- `codex exec` → authenticated, produces output.
- `claude -p` → `{"is_error": true, "result": "Failed to authenticate: OAuth session expired and could not be refreshed"}`.

This is precisely the **agent-unavailable** state the platform is designed for (doc 04 §9, ops runbook §11, risk R-005). Consequences, all already implemented:

- Deterministic CI is unaffected and remains authoritative.
- With only one provider available, the broker yields `pending` — a single-provider review is **not** treated as two-agent consensus (ops runbook §12).
- The fix is the documented P4-09 owner action: run `claude setup-token` on the agent host to mint a one-year `CLAUDE_CODE_OAUTH_TOKEN`, then complete the PAC-018 qualification (10–20 unattended runs; expiry/invalid-token/exhaustion/timeout/malformed/redaction cases fail closed).

## Exit gate G4

| Criterion | Status |
|---|---|
| A reviewer produces schema-valid structured findings | **Met** (Codex, after one bounded repair) |
| A seeded high-severity defect is found and reported | **Met** (Codex found the data race) |
| The loop stops on budget/no-progress | Met (P5 budgets) |
| Auth artifacts never appear in repos/logs/build runners | Met (broker handles no tokens; runner credential-free per G3) |
| Both models review independently | **Blocked** on Claude auth (owner P4-09) |

G4 is met for the Codex reviewer and the protocol mechanics; the two-model portion resumes once Claude re-authenticates.
