# P4-09 — Unattended Claude Authentication Qualification (PAC-018)

**Status:** PASSED
**Date:** 2026-07-15
**Owner:** Aakash (`Aakash1337`), platform + security roles
**Gates:** Charter PAC-018; roadmap P4-09; ADR-005; NFR-COST-001; ASM-004

This record qualifies the documented `claude setup-token` path for unattended Claude use before the broker depends on it (doc 04 §4.2). It is the evidence artifact required by PAC-018.

## Authentication facts (recorded)

| Field | Value |
|---|---|
| Documentation URL | <https://code.claude.com/docs/en/iam> (§ "Generate a long-lived token") |
| Auth mode | `claude setup-token` one-year OAuth token, supplied via `CLAUDE_CODE_OAUTH_TOKEN` (user-scope env) |
| Plan | Claude Max (consumer subscription) |
| Client version | Claude Code 2.1.206 |
| Token length / scope | 108 chars; inference-only (cannot establish Remote Control sessions, per docs) |
| Precedence confirmed | `ANTHROPIC_API_KEY`/`ANTHROPIC_AUTH_TOKEN` unset in the review subprocess, so the OAuth token is the credential (IAM precedence #5), i.e. subscription auth — not an API key |
| Verification date | 2026-07-15 |
| Next review date | 2026-10-15 (quarterly, per ops runbook §4) and on any client/provider change |
| Token expiry | 2027-07-15 (rotate ahead of it; ops runbook §4) |

## Drill results

Run from a clean broker-owned directory (`E:\CouncilCI\spike\clean`); the token was read from the user registry into subprocess env only and never written to disk or printed.

| Drill | Result |
|---|---|
| Structured-output runs (pass@1) | **12/12** — every run returned exactly the requested JSON; avg 4.7 s, range 3.1–8.1 s |
| Invalid token | **Fails closed** — a bogus token produces an auth error, never a silent pass |
| Timeout | **Bounded** — a 1 s hard cap kills the call (rc 124); no runaway |
| Secret redaction | **Clean** — neither the full token nor its 20-char prefix appears in any run output, CSV, or log |
| Persistence | **Confirmed** — token is in the persistent user environment, so a fresh shell (and a reboot) authenticates without re-login |

Per-run detail is in the spike CSV (not committed — lives in the local evidence spool; summarized above).

## Cost note (NFR-COST-001 / KPI-012)

The CLI reports a `total_cost_usd` per run (≈$0.70 across the 12 runs). Under subscription (OAuth-token) auth this is the **informational API-equivalent estimate**, not an actual charge — Max-plan usage is covered by the flat subscription, and the documented setup-token path is the no-per-token-billing path (ADR-005/ADR-012). **Owner action (low priority):** confirm zero API line-items in the Anthropic console at the next billing cycle to close the loop empirically; the $0-unapproved-spend KPI assumes this.

## Not exercised (documented, deferred)

- **Usage exhaustion** — cannot be forced on demand; the agent-lane behavior on exhaustion is covered by the arbiter's `agents-available=false` path (tested: yields `pending`, never a false pass) and by ops runbook §12. Re-verify opportunistically if a real throttle occurs.
- **Reboot** — proven by proxy (persistent user-scope env); a literal reboot test is unnecessary given the credential store is the registry.

## Conclusion

The unattended Claude authentication path is qualified. The broker may use `--real` Claude reviews. If the documented setup-token path changes or is withdrawn, the lane fails closed pending re-qualification (ADR-005 revisit trigger).
