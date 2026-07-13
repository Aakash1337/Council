# P0-02 — Pilot Repository Decision Record

**Status:** Proposed — ratified by owner approval of the Phase 0 pull request
**Date:** 2026-07-13
**Decision authority:** Project owner
**Roadmap reference:** [Phase 0 work item P0-02](../agentic-cicd-docs/07-implementation-roadmap.md)

## Decision

**`Aakash1337/CustomDNS` is the pilot repository**, conditional on the owner actions below.

The roadmap's ideal pilot profile is "a small Go or Python service with a database boundary, API contract, container image, and deployable test environment" with meaningful automated tests (ASM-007) and no irreplaceable production secrets.

## Candidates surveyed

All directories under the operator's projects folder were surveyed; the four plausible candidates:

| Candidate | Language | Tests | Remote | Assessment |
|---|---|---|---|---|
| `httpserver` | Go | Yes | Bitbucket, `cybiccrm` org | **Disqualified on classification** (below) |
| `CustomDNS` | Go | Yes — 4 test files (config, resolver ×2, server) | GitHub, personal | **Selected** |
| `StockGame` | Go | Minimal (1 test file) | No remote | Runner-up; web app, weaker test base |
| `ascendio` | Python | Not evident | None visible | Ingests third-party financial data (SEC/FINRA/Yahoo); likely API keys and rate-limited external calls — poor fit for deterministic CI |

Larger repositories (Cybic, Relay, Website, Schedule, Odin, Telestrator) were excluded: the pilot should be small enough that pipeline behavior, not application complexity, dominates.

### Why not `httpserver`

Technically it is the best fit — Dockerfile, Makefile, existing GitHub Actions workflow, prior scanner-output experiments (gitleaks/osv/semgrep/trivy in `dist/`). However, its remote is `bitbucket.org/cybiccrm/test-pipeline`, an apparent employer/client organization. Under the [provider-transmission policy](../agentic-cicd-docs/04-agent-collaboration-protocol.md) (§5.2) that makes its provisional class **`confidential-restricted`**, which **prohibits the consumer-plan agent lane** — disqualifying for a pilot whose purpose is to exercise that lane. **[OWNER INPUT]** — if this assessment is wrong (e.g., the code is personally owned test material), say so in the PR review; the decision can be revisited.

### Why `CustomDNS`

- Small Go network service (~16 files) with a clean package layout (`cmd/`, `internal/config`, `internal/resolver`, `internal/server`) and a dedicated test helper package (`internal/dnstest`).
- Real test coverage across the config, resolver, and server boundaries — satisfies ASM-007.
- A configuration boundary (`configs/example.yaml`) and a network protocol boundary (DNS parsing of untrusted input) give the risk-tier policy meaningful high/medium/low distinctions.
- Personally owned → class **`private-personal`** once verified, eligible for the consumer-plan agent lane (FR-AGENT-013/014).
- No production deployment, no irreplaceable secrets.

Known gaps versus the ideal profile (acceptable, and useful platform work):

| Gap | Disposition |
|---|---|
| No database boundary | Accepted deviation — config + network protocol boundaries provide equivalent risk-tier variety |
| No Dockerfile | Added in P1 as part of adapter work |
| No CI workflow | Advantage: greenfield for the platform's `pr.yml`, no legacy to migrate |
| Repository is **public** | Blocking owner action, below |

## Owner actions required before G0 closes

1. **Flip `Aakash1337/CustomDNS` to private** (Settings → General → Danger Zone → Change visibility). Required by ASM-001 and CON-004; a public repo cannot use the persistent local runner (FR-RUN-001) and would undermine the local-first pilot. The agent does not change repository visibility on the owner's behalf.
2. **Confirm the `httpserver`/`cybiccrm` classification assumption** (or correct it).
3. **Verify provider model-improvement controls** for both consumer accounts and record the verification date (FR-AGENT-014, doc 04 §5.2): OpenAI "Improve the model for everyone" + Codex training control; Anthropic Model Improvement setting. Required before Phase 4, not before G0.

## Provider-transmission classification (FR-AGENT-013)

| Repository | Class | Consumer-plan agent lane |
|---|---|---|
| `Aakash1337/CustomDNS` | `private-personal` (pending action 3) | Allowed after verification |
| `Aakash1337/Council` | `public` (documentation only) | Allowed after secret scanning |
| `httpserver` / `cybiccrm` material | `confidential-restricted` (provisional) | Prohibited |
