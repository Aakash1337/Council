# Council

A local-first, specification-driven agentic CI/CD platform in which coding agents implement, independently review, challenge, and repair changes — while deterministic checks remain the final authority.

The name reflects the core review model: independent agents (Claude Code and Codex) deliberate over the same sealed evidence, but the deterministic policy engine — not the council — holds final authority.

## Documentation

The full documentation set lives in [`Docs/agentic-cicd-docs/`](Docs/agentic-cicd-docs/README.md):

- Project charter and requirements
- System architecture and trust zones
- Specification and evaluation framework
- Agent collaboration protocol (blind review, cross-review, arbitration, bounded repair)
- CI/CD quality and security standard
- Infrastructure and deployment
- Implementation roadmap
- Operations runbook
- Risk register and architecture decisions

Plus machine-validatable JSON schemas, document templates, and worked examples.

## Status

**Pilot + MVP implemented (roadmap Phases 0–5).** The pilot repository is [`Aakash1337/CustomDNS`](https://github.com/Aakash1337/CustomDNS).

| Phase | Gate | State |
|---|---|---|
| P0 Discovery & threat model | G0 | Done — [Docs/phase-0](Docs/phase-0/README.md) |
| P1 Spec & local quality loop | G1 | Done — `ci/run` + Go adapter + traceability in the pilot |
| P2 GitHub deterministic CI | G2 | Done (enforcement clause pending owner decision R-041) — [Docs/phase-2](Docs/phase-2/README.md) |
| P3 Local execution plane | G3 | Done — hardened VirtualBox runner, isolation + fallback drills passed — [Docs/phase-3](Docs/phase-3/README.md) |
| P4 Supervised two-model workflow | G4 | Partial — real Codex review proven; two-model + repair loop gated on the owner's `claude setup-token` (P4-09) — [Docs/phase-4](Docs/phase-4/README.md) |
| P5 Automated agent broker | G5 | Core done — deterministic broker with proven invariants — [`broker/`](broker/), [Docs/phase-5](Docs/phase-5/README.md) |
| P6 Ephemeral scaling & observability | G6 | Deferred (hardened v1) — gated on the ARC-vs-GARM benchmark and measured need |
| P7 Release & deployment hardening | G7 | Deferred (hardened v1) — the pilot has no production deployment target yet |

Open owner actions: the R-041 branch-protection decision (GitHub Pro vs. a process-discipline waiver) and `claude setup-token` to unblock two-model review.

## Branches

- `main` — the canonical, organization-oriented platform design.
- `home` — adaptations for single-operator/home-lab use. The diff from `main` serves as the divergence record; commit messages on this branch record the rationale for each deviation.
- `phase-N-<name>` — each roadmap phase (see [implementation roadmap](Docs/agentic-cicd-docs/07-implementation-roadmap.md)) is developed on its own branch and merged into `main` through a pull request, e.g. `phase-0-discovery`, `phase-1-spec-loop`.
- `feature/<name>` — discrete features or fixes that don't map to a whole phase.

Phase and feature branches are created when work begins, not in advance, and are deleted after merge.
