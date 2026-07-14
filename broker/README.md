# Council agent broker (Phase 5 MVP)

A small deterministic state machine (Council doc 04 §11) that freezes an
immutable review bundle, runs blind first-party reviews, seals them,
runs one cross-review round, and applies deterministic arbitration. It
never reads or relays subscription credentials and holds no merge or
release authority.

## Commands

```
broker freeze        # content-address and hash the review bundle
broker review        # blind reviews (mock or real CLI), schema-validate, seal
broker cross-review  # validate one cross-review round
broker decide        # deterministic arbitration -> decision.json (exit 3 if blocked)
```

## Decision-matrix coverage (doc 04 §7.5 → tests)

Every precedence row is pinned by at least one test:

| §7.5 row | Test |
|---|---|
| Evidence missing/stale/malformed → block | `TestMissingEvidenceFailsClosed` |
| Hard gate fails → block regardless of votes | `TestHardGateBlocksDespiteApproval` (PAC-005) |
| Reproduced acceptance violation → block | `TestMediumWithReproducedAcceptanceViolationBlocks` |
| Waivable finding + valid waiver → `pass_with_waiver` | `TestValidWaiverYieldsPassWithWaiver`; expired/unowned/hard-gate cases in `TestExpiredWaiverStillBlocks`, `TestWaiverEdgeCases` |
| Critical/high with reproducer → block | `TestHighWithReproducerBlocks` |
| Same root cause found by both → block | `TestBothFindSameHighBlocks` |
| Confirmed by other, unreproduced → block pending | `TestHighAcceptedByOtherBlocks` |
| Critical/high disagreement → human | `TestLoneHighRoutesToHuman`, `TestNeedsReproducerRoutesToHuman`; refuted clears via `TestRefutedHighClears` |
| Medium without demonstrated violation → annotate | `TestMediumWithoutReproducerAnnotates` |
| Low/advisory → annotate | `TestLowFindingAnnotatesOnly` |
| Both approve, gates pass → approve | `TestCleanApprove` |
| Cross-review-discovered blocker reaches arbitration | `TestCrossReviewNewFindingBlocks`, `TestLoadCrossCarriesNewFindings` |
| Agent unavailable → `pending`, never false pass | `TestAgentUnavailablePending` |
| Human approval absent → `pending/human_required` | `TestMissingHumanApprovalPending` |

Integrity invariants: bundle tamper and added-file injection are caught
(`TestVerifyDetectsTamper`, `TestVerifyRejectsAddedInputFile`); sealed
reports are re-verified before every decision (`TestVerifySeals`); the
review subprocess env is a strict allowlist that never passes provider
API keys or tokens (`TestReviewEnvIsAllowlistOnly`); schema enums,
provider/client pairing, and ECMA-262 path patterns are enforced
(schema package tests).

Waivers (doc 05 §13): supplied via `decide --waivers-file --now`;
scoped to exact finding IDs, require a named approver, expire
deterministically (fail-closed on unparseable dates), and can never
clear a hard-gate or evidence failure.

## Runners

`agent.MockRunner` replays fixtures for deterministic tests and offline
development (so broker CI never spends subscription capacity).
`agent.CLIRunner` invokes the real `claude`/`codex` clients, which own
authentication — the broker never handles tokens.

See `internal/arbiter` for the decision precedence and its test suite.
