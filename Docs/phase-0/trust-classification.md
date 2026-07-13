# P0-03 — Code Trust Classification and Execution Matrix

**Status:** Proposed — ratified by owner approval of the Phase 0 pull request
**Date:** 2026-07-13
**Roadmap reference:** [Phase 0 work item P0-03](../agentic-cicd-docs/07-implementation-roadmap.md)

This instantiates the trust classes from [doc 06 §7.2](../agentic-cicd-docs/06-infrastructure-and-deployment.md) for the actual repositories in scope.

## 1. Repository trust classes

| Repository | Visibility | Trust class | Fork PRs accepted? |
|---|---|---|---|
| `Aakash1337/CustomDNS` (pilot) | Public → **private** (owner action pending) | `private-owner` | No — none expected; policy below applies if one arrives |
| `Aakash1337/Council` (platform docs) | Public (intentional) | `public` | Possible once shared |
| All other local projects | — | Out of platform scope for the pilot | — |

## 2. Execution matrix

| Event | Repository class | Runner | Notes |
|---|---|---|---|
| PR from owner branch | `private-owner` | Local `ci-pilot` VM when enabled (P3); GitHub-hosted before P3 or via `force_hosted` | The normal pilot path |
| PR from fork or unknown contributor | any | **GitHub-hosted only** | CON-004 / FR-RUN-001; workflow policy must make the local runner label unselectable (PAC-010) |
| Any event, `public` repo | `public` | **GitHub-hosted only** | Applies to Council itself if it ever gets CI (e.g., markdown lint) |
| Protected-branch nightly | `private-owner` | Local | After P3 |
| Agent review | `private-owner`, deterministic gates passed | Agent host only (P4+) | Never on build runners; never for `confidential-restricted` content |
| Release | reviewed commit | Protected release identity (P7) | Not in pilot scope |

## 3. Explicit decisions

1. **Public/fork code never executes on the persistent local runner.** This is the G0 "public/fork code policy explicitly decided" item. Enforcement: runner-group restriction to the private pilot repo plus workflow-level trust classification (PAC-010 test in P3).
2. **Council remains public** by owner intent (shareable documentation). Consequence: any CI added to Council runs GitHub-hosted, and Council never receives agent-lane credentials or the local runner label.
3. **`confidential-restricted` material** (anything under the `cybiccrm` org or other employer/client code) is out of platform scope entirely — not onboarded, not used as fixtures, never sent through the consumer agent lane.
4. **Trust class is derived, never user-supplied** (doc 06 §7.2): the workflow computes it from event type, repo visibility, and contributor association.
