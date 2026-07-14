# Phase 1 — Specification and Local Quality Loop

**Status:** Delivered 2026-07-14 — pending owner ratification via [CustomDNS PR #1](https://github.com/Aakash1337/CustomDNS/pull/1)
**Roadmap reference:** [Phase 1 work items](../agentic-cicd-docs/07-implementation-roadmap.md)

Phase 1 lives in the pilot repository (`ci/run` is repository-owned per ADR-010); this page is the Council-side status record.

## Work-item disposition

| Work item | Deliverable | Where |
|---|---|---|
| P1-01 Initialize Spec Kit | Constitution + `specs/changes/<ID>/` workflow per doc 03 §3. **Deviation:** the Spec Kit *tool* was not installed; the doc-03 repository contract was implemented directly. Revisit tool adoption when agent-driven spec authoring starts (P4). | `specs/` in CustomDNS |
| P1-02 Change manifest schema | Council schema copied to `ci/schemas/`; validated by the traceability gate (JSON Schema 2020-12, ECMA-262 patterns via regexp2) | CustomDNS |
| P1-03 Repository agent rules | `agents/rules.md` canonical; `AGENTS.md`/`CLAUDE.md` thin pointers (ADR-019) | CustomDNS |
| P1-04 `ci/run detect` | Capability manifest incl. explicit `unsupported` list | CustomDNS |
| P1-05 Language adapter (Go) | gofmt, go vet, staticcheck, build, gotestsum (JUnit), coverage (LCOV), `-race`, config validation | CustomDNS |
| P1-06 Security-fast adapter | gitleaks, osv-scanner, govulncheck — SARIF outputs, pinned versions | CustomDNS |
| P1-07 Normalized evidence | `.ci-evidence/<run-id>/` with JUnit/LCOV/SARIF/logs + SHA-256 `manifest.json` binding commit, tool versions, checks | CustomDNS |
| P1-08 Developer hooks | **Deferred** (optional per roadmap); `ci/run fast` documented as the pre-commit habit | — |
| P1-09 First representative spec | CDNS-001 baseline contract, approved, 16 hard criteria → existing tests; CDNS-002 caching draft staged as the first agentic change | CustomDNS |

## Exit gate G1 evidence

- `ci/run full`: 10/10 checks pass noninteractively from bootstrap on a clean checkout.
- Traceability: 16/16 mandatory criteria resolve to executable verifiers; frozen spec hash verified.
- Seeded failure 1: reverting name normalization → `test/unit` fails, pipeline blocks.
- Seeded failure 2: tampering `specification.md` → spec-hash mismatch, traceability blocks.
- The new gates immediately caught real issues: a staticcheck error-string finding and an outdated `golang.org/x/net` flagged by osv-scanner — both fixed in the PR.
- Setup and troubleshooting: `docs/ci.md` in CustomDNS.

## Notes for later phases

- `councilci` (traceability + evidence manifest) is the seed of the P5 broker's validation layer; keep it dependency-light.
- The evidence manifest is explicitly marked as a Phase 1 precursor to the full evidence-bundle schema (P5).
- Scored evals remain unsupported until the CDNS-002 implementation brings the eval harness (recorded in `detect` output).
