# Decision #7 — Repository Protection Strategy (WAIVER-R041-CUSTOMDNS)

**Status:** Accepted 2026-07-15
**Decision authority:** Project owner (Aakash), recorded as project + security roles
**Relates to:** R-041, F-03, G2 enforcement clause, ADR-014

## Decision

The owner accepts a **time-bounded process-discipline waiver** for `CustomDNS` instead of purchasing GitHub Pro for enforced branch protection, for the remainder of the pilot. If a private repository later needs real enforcement, the owner will buy Pro then. The stated general direction is to work in **public** repositories, where branch protection is available on the Free plan (as already enabled and drilled on Council).

## Waiver record (doc 05 §13 shape)

| Field | Value |
|---|---|
| Waived control | Enforced branch protection / required status checks on `CustomDNS` `main` |
| Scope | `Aakash1337/CustomDNS` only, pilot duration |
| Justification | Solo owner is the only committer; GitHub Free cannot enforce protection on private repos; cost avoidance during the pilot |
| Compensating controls | (1) Owner is sole committer; (2) changes go through PRs by convention, and PR CI (`pr.yml` → `policy / decision`) runs and reports on every PR; (3) the merge is a deliberate human action |
| Owner / approver | Aakash (project owner; security-owner role for this acceptance) |
| Expiry / review | Pilot end, or immediately upon the precondition below |
| Residual risk | R-041 residual lowered to 1/3/3 (Accepted) |

## Non-negotiable precondition

**Enforced protection MUST be in place before any unattended agent has merge-capable access to a repository (P5).** A process-discipline waiver relies on a disciplined human reviewer; an autonomous writer is not that. Concretely, before the broker is given any path to merge:

- either the target repo is public (Free-plan protection, as on Council), or
- GitHub Pro enforcement is enabled on the private repo.

This is the same boundary the charter already draws (agents never hold merge authority; merges go through protected rules) — decision #7 does not relax it; it only defers *enforcement tooling* for the supervised, human-committer pilot.

## Consequence of the "use public repos" direction (flagged for the owner)

`CustomDNS` is **currently private** (flipped during Phase 0 on the platform's recommendation, per CON-004). "I'll use public repos" implies making it public again at some point. That is a repository-visibility change only the owner can make, and it carries a **provider-transmission consequence that must be checked first** (doc 04 §5.2, R-037):

- A public repo's contents are already world-readable, so agent transmission classification is trivially `public` — *good*.
- **But** making a currently-private repo public exposes its **entire history**, not just the current tree. Before flipping, the owner MUST confirm no secret, key, private data, or `confidential-restricted` material exists anywhere in `CustomDNS` history (a `gitleaks detect` over full history is the mechanical check; the platform already runs gitleaks with `fetch-depth: 0` in CI, so a clean `security / secrets-sast-sca` on a full-history scan is strong evidence).
- Once public, the **local-runner trust rule flips**: public/fork PRs must run on GitHub-hosted runners, never the persistent `ci-pilot` VM (FR-RUN-001, CON-004, PAC-010). The execution matrix in the [trust classification](../phase-0/trust-classification.md) already encodes this; it becomes active the moment the repo is public.

**Recommendation:** keep `CustomDNS` private under the process-discipline waiver for the remainder of the supervised pilot (simplest, and the local runner stays usable). Only make it public when you actually intend to, and run the full-history secret scan first. This is an owner decision; the platform will not change repository visibility.
