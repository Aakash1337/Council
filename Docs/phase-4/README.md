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

## Live two-model blind review through the broker (2026-07-15)

With Claude auth qualified (P4-09), the **full two-model pipeline ran end-to-end through the broker** (`broker review --real`) on the seeded CDNS-002 cache defect. Both first-party clients were launched by the broker, reviewed **blind and independently**, and their reports were sealed and arbitrated. Evidence: [`broker/testdata/p4-live-twomodel/`](../../broker/testdata/p4-live-twomodel/) (real `claude-review.json`, `codex-review.json`, `decision.json`).

- **Both models independently found the seeded data race** — Claude: `critical concurrency`; Codex: `high concurrency_failure_recovery` — plus additional real findings (missing tests, mutable-slice aliasing, unbounded growth).
- **Identity enforcement worked** (doc 04 §9.3): each report carries its true provider/client (`anthropic`/`claude-code`, `openai`/`codex-cli`); the broker rejects a report whose self-reported identity doesn't match the invocation. This was found necessary because a shared prompt initially made Claude echo Codex's identity.
- **One bounded schema-repair attempt** (doc 04 §7.2) was applied to each raw report — the models' first output missed schema details (severity casing, string-vs-object fields); a single re-prompt fixed the shape without changing substance. A second failure would be `infrastructure_error`, never a pass. This loop is now implemented in the broker.
- **Deterministic arbitration blocked** the change (exit 3): with hard gates passing, the reproduced critical/high concurrency findings block; realistically the `-race` hard gate also fails, blocking regardless of the agents. Agreement is evidence, not authority.
- **Transport, not substance** (doc 04 §11): the broker unwraps the `claude -p` JSON envelope and strips markdown fences (`ExtractReviewJSON`) but never rewrites findings.
- **Windows portability:** the broker needs explicit `--claude-bin`/`--codex-bin` paths because the clients are `.cmd` shims, not `.exe` (Go's `exec.LookPath` doesn't resolve shims).

## Exit gate G4

| Criterion | Status |
|---|---|
| A reviewer produces schema-valid structured findings | **Met** — both models, after one bounded schema-repair each |
| A seeded high-severity defect is **found** | **Met** — both models independently found the data race |
| Both models review independently (blind) | **Met** — live two-model run through the broker |
| Deterministic arbitration blocks despite/aligned-with agents | **Met** — blocked (exit 3) |
| The loop stops on budget/no-progress | Met (P5 budgets; repair-gate enforces the two-cycle cap) |
| Auth artifacts never appear in repos/logs/build runners | Met — token forwarded to the client only, absent from all artifacts (scanned) |
| The seeded defect is **repaired** and re-review confirms the fix | **Not yet** — the writer→repair→reconfirm loop has not been run live end-to-end (the `repair-gate` enforces its bounds; a live repair cycle remains) |

**G4 is substantially closed:** independent live two-model review, detection, identity enforcement, bounded schema repair, and deterministic arbitration are all proven end-to-end. The single remaining criterion is a live writer-repair-reconfirm cycle (the mechanics — repair-gate, reverification — exist and are tested; running one live is the last step).
