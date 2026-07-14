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

## Invariants (proven by tests + CLI drills)

- A failed hard gate blocks even when both reviewers approve (PAC-005).
- Missing/tampered evidence fails closed; the bundle hash is verified
  before any review is consumed.
- A reproduced critical/high finding blocks regardless of the other
  model's verdict; a lone unrefuted high routes to a human; a refuted
  one annotates.
- Agent unavailability yields `pending`, never a false pass.

## Runners

`agent.MockRunner` replays fixtures for deterministic tests and offline
development (so broker CI never spends subscription capacity).
`agent.CLIRunner` invokes the real `claude`/`codex` clients, which own
authentication — the broker never handles tokens.

See `internal/arbiter` for the decision precedence and its test suite.
