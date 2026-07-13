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

Proposed baseline (v0.1) — documentation phase; implementation has not started.

## Branches

- `main` — the canonical, organization-oriented platform design.
- `home` — adaptations for single-operator/home-lab use. The diff from `main` serves as the divergence record; commit messages on this branch record the rationale for each deviation.
