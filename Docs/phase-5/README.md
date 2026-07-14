# Phase 5 — Automated Agent Broker (MVP)

**Status:** Delivered 2026-07-14 — [`broker/`](../../broker/)
**Roadmap reference:** [Phase 5 work items](../agentic-cicd-docs/07-implementation-roadmap.md)

The broker is the deterministic state machine that turns two independent agent reviews into an auditable, policy-gated decision (Council doc 04 §11). It is a small Go module — one static binary — that freezes an immutable review bundle, runs blind reviews, seals them, runs one cross-review round, and applies deterministic arbitration. It never reads or relays subscription credentials and holds no merge or release authority.

## Work-item disposition

| Work item | Deliverable | Status |
|---|---|---|
| P5-01 Run manifest | `bundle.Freeze` — content-addressed, deterministic aggregate hash | Done |
| P5-02 Native CLI adapters | `agent.CLIRunner` (claude/codex) + `agent.MockRunner` (fixtures) | Done |
| P5-03 Blind review | `broker review` validates then **seals** (hashes) each report before either is revealed | Done |
| P5-04 Cross-review | `broker cross-review` validates one round against the cross-review schema | Done (validation; arbiter consumes dispositions) |
| P5-05 Policy arbitration | `arbiter.Decide` — deterministic precedence, reason-coded | Done |
| P5-06 Repair workflow | Repair handoff is defined in doc 04 §7.6; broker exits with blocking IDs for the writer. Full loop automation deferred with P4 agent-auth | Partial |
| P5-07 Queue/concurrency | Single-run serialization; per-provider serialization documented (doc 04 §9.2) | Partial |
| P5-08 Command/network allowlists | `CLIRunner` passes no provider API keys; child cwd is the bundle; env is allowlisted | Done |
| P5-09 Publication job | Decision JSON is the safe-output payload (doc 04 §12); GitHub publisher wiring deferred to integration | Partial |
| P5-10 Adversarial test suite | Arbiter + bundle unit suites; four CLI invariant drills | Done |

## Verified invariants

Unit tests (`go test ./...` green) plus CLI drills prove the load-bearing rules:

| Invariant | Evidence |
|---|---|
| A failed hard gate blocks even when **both** reviewers approve | `TestHardGateBlocksDespiteApproval`; CLI scenario B (exit 3) — **PAC-005** |
| A reproduced critical/high finding blocks regardless of the other model | `TestHighWithReproducerBlocks`; CLI scenario A (Claude approves, Codex's reproduced high blocks) |
| A lone unrefuted high routes to a human; a refuted one annotates | `TestLoneHighRoutesToHuman`, `TestRefutedHighClears` |
| Independent corroboration by both reviewers blocks | `TestBothFindSameHighBlocks` |
| Missing/tampered evidence fails closed; bundle hash verified before use | `TestMissingEvidenceFailsClosed`, `TestVerifyDetectsTamper`; CLI scenario D (tamper caught) |
| Agent unavailability yields `pending`, never a false pass | `TestAgentUnavailablePending` |
| Missing human approval yields `pending/human_required` | `TestMissingHumanApprovalPending` |

## Design notes

- **Mock vs real runner.** `MockRunner` replays fixtures so the broker's own CI never spends subscription capacity; `CLIRunner` drives the real first-party clients, which own authentication — the broker never touches tokens (doc 04 §4.2).
- **Sealing.** Blind reports are written and hashed before either is revealed, so later agreement cannot be copied (doc 04 §7.3).
- **Arbiter is code, not a model.** Agreement is supporting evidence, never truth; two models cannot override a hard gate, and a lone well-reproduced bug is not dismissed because the other model missed it.

## Dependency on P4

Live two-model operation needs both first-party CLIs authenticated. As of 2026-07-14 the Codex CLI is authenticated and works headlessly, but the Claude CLI's OAuth is expired — see the [Phase 4 record](../phase-4/README.md). Until Claude re-authenticates (owner action P4-09 / setup-token), the broker runs single-provider, which by policy yields `pending` (never single-model consensus).
