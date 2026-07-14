# P0-04 — Change Risk-Tier Policy (Pilot)

**Status:** Proposed — ratified by owner approval of the Phase 0 pull request
**Date:** 2026-07-13
**Roadmap reference:** [Phase 0 work item P0-04](../agentic-cicd-docs/07-implementation-roadmap.md)

This instantiates the generic tiers in [doc 04 §9.1](../agentic-cicd-docs/04-agent-collaboration-protocol.md) for the pilot repository `CustomDNS` and for the Council platform repository itself. The generic table remains normative; this document maps concrete paths to tiers.

## 1. Pilot repository (`CustomDNS`) path-to-tier map

| Paths / change type | Tier | Rationale |
|---|---|---|
| `README.md`, docs, comments, test-only additions outside protected paths | **Low** | Reversible, non-functional |
| `configs/**`, config-loading logic (`internal/config/`) | **Medium** | Behavior-affecting but reversible; no trust boundary |
| Application logic in `cmd/`, `internal/server/` (non-parsing) | **Medium** | Ordinary logic |
| `internal/resolver/**` — DNS parsing, forwarding, rules | **High** | Parses untrusted network input; concurrency; the service's security boundary |
| `.github/workflows/**`, `ci/**`, `policy/**`, `specs/**` acceptance manifests, `CODEOWNERS` | **High** | CI workflow/policy per the generic table |
| Anything touching credentials, signing, release identity, or weakening a protected verifier | **Critical** | Generic table applies unchanged |

A change inherits the highest tier of any path it touches. Tier is proposed by the change author and confirmed per doc 04 §9.1 before agent work.

## 2. Council repository (docs) map

| Paths | Tier |
|---|---|
| `Docs/**` prose | **Low** |
| `Docs/agentic-cicd-docs/schemas/**` | **High** (machine-validated contracts other components depend on) |
| Future `ci/**`, `.github/workflows/**` | **High** |

## 3. Solo-operator substitution provision

One person (the project owner) currently fills every human role. Per charter §7 and doc 04 §9.1, the logical roles remain separate:

- Every approval, waiver, and risk acceptance records the **logical role** exercised (e.g., "approved as security owner").
- The generic table's "two qualified humans" requirement for high-tier agent-unavailable substitution cannot be met solo; for the pilot, the documented substitution is: deterministic CI + one human review + an explicit `agent_substituted` record, and the change waits for the agent lane where practical.
- **Critical-tier self-acceptance cooling-off:** a critical-tier risk acceptance or waiver made by the solo operator takes effect no sooner than 24 hours after it is recorded, to substitute deliberation for the missing second approver. This is a pilot-specific control; it is expected to be replaced by real multi-party approval in an organizational deployment.
