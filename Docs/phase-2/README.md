# Phase 2 — GitHub Deterministic CI

**Status:** Delivered 2026-07-14 — live verification via [CustomDNS PR #2](https://github.com/Aakash1337/CustomDNS/pull/2)
**Roadmap reference:** [Phase 2 work items](../agentic-cicd-docs/07-implementation-roadmap.md)

## Work-item disposition

| Work item | Deliverable | Status |
|---|---|---|
| P2-01 `pr.yml` | Five jobs with doc-05 §14 check names; thin YAML over `ci/run`; evidence artifacts | Done |
| P2-02 `nightly.yml` | Scaffold: full + security with fresh advisory data, 02:17 UTC + manual dispatch | Done (matrix/mutation/fuzz/DAST noted for hardened v1) |
| P2-03 Ruleset / branch protection | **Blocked by R-041** — API returned `403: Upgrade to GitHub Pro or make this repository public` on the private pilot repo (GitHub Free). Owner decision #7 pending: Pro upgrade vs. pilot process discipline | Open |
| P2-04 CODEOWNERS | Protected paths (`/ci/`, `/specs/`, `/agents/`, `/.github/`) → owner | Done (enforcement follows R-041) |
| P2-05 Harden Actions | All actions pinned to full commit SHAs; `permissions: contents: read`; `persist-credentials: false` | Done |
| P2-06 Workflow scanners | actionlint v1.7.12 + zizmor v1.26.1, pinned, zero findings; run as required PR job | Done |
| P2-07 Artifacts | JUnit/LCOV/SARIF/logs + hashed manifest uploaded per job; 14-day PR / 30-day nightly retention (0.5 GB plan budget respected) | Done |
| P2-08 Concurrency/timeouts | Stale-PR-run cancellation; explicit `timeout-minutes` on every job | Done |
| P2-09 Dependency automation | Dependabot: gomod (service + councilci) and github-actions, weekly | Done |
| P2-10 Final policy gate | `policy / decision` job: normalized pass/blocked over all required checks, results table in the run summary | Done |

## Exit gate G2

| Criterion | Status |
|---|---|
| Direct default-branch pushes blocked | **Pending owner decision #7 (R-041)** — not enforceable on GitHub Free + private repo |
| All required checks execute on GitHub-hosted infrastructure | Verified by the live PR run |
| Third-party actions pinned; permissions explicit | Done (only official `actions/*`, all full-SHA pins) |
| Failures visible and correctly block merge | Check-level: verified (Phase 1 seeded failures + `policy / decision` aggregation). Merge-level blocking follows R-041 |
| No agent path around required checks | Follows R-041 (identical mechanism) |

**G2 is functionally complete; its enforcement clause waits on one owner decision.** Live verification: all five checks passed on the first run of [CustomDNS PR #2](https://github.com/Aakash1337/CustomDNS/pull/2) — build-test 4m44s, security 3m41s, traceability 3m43s, workflow-lint 40s, policy-decision 4s (~13 runner-minutes per PR against the 2,000/month allowance).

## Notes

- The billing screenshot confirmed Free plan: 2,000 Actions min + 0.5 GB storage/month. Current usage: 0. A Pro upgrade (option A) would also raise minutes to 3,000.
- Evidence artifacts are the portable copy of `.ci-evidence/<run-id>/`; the local mirror store arrives in P3/P6.
