# Audit Round 2 — 2026-07-15

**Scope:** full re-verification of the 2026-07-14 state plus every improvement the platform could land autonomously (owner away; long-development mode).
**Result:** all prior claims held; five improvements landed; fuzzing found **two real config-validation bugs** within seconds of existing; the platform's own governance gap was closed and drilled.

## 1. Re-verification (all held)

Broker: gofmt/vet/staticcheck clean, `-race` suite green. Pilot: `ci/run full` 11/11 + `security` 3/3 from clean main. VM running, isolation intact. Zero broken relative links across all Docs. No open PRs at session start.

## 2. Improvements landed

### 2.1 Council now has CI and enforced branch protection (Council #11)
**Finding:** the platform repo itself had *no automated gates* — the broker merged on bot review alone, violating doc 05's "applies to all repositories."
**Landed:** `broker / checks` (gofmt, vet, staticcheck, build, race tests), `workflow / lint`, `docs / markdownlint`; then branch protection on `main` with those three checks required, strict up-to-date, **admin-enforced**, no force pushes.
**Proof it works twice over:** the CI's *first ever run* caught a real ST1005 in the broker that local-only testing had missed; and the protection drill — a direct push to main — was **rejected** (`protected branch hook declined`). R-041 is now closed for Council; it remains open only for CustomDNS (private + Free plan), pending owner decision #7.

### 2.2 Repair-cycle cap finally enforced in code (Council #11)
ADR-008's "one pass default, two maximum" existed only as prose. New `broker repair-gate`: per-change ledger, `--max > 2` refused outright, no-progress detection (same failed head twice → refused), corrupt ledger fails closed, cross-change ledger reuse refused. Full lifecycle pinned by tests (PAC-007).

### 2.3 Broker CLI pipeline now tested in-process (Council #11)
`cmdDecide`'s direct `os.Exit(3)` was untestable — refactored to a sentinel error (same exit contract). New tests run the real freeze→review(mock)→decide pipeline: blind invariant, head-scoped waiver → `pass_with_waiver`, post-seal tamper, tampered-bundle refusal. Bundle Write→Load round-trip + fail-closed loads. Coverage: cmd 20→68%, bundle 76→90%.

### 2.4 Fuzzing — and it immediately earned its keep (CustomDNS #8)
Two native fuzz targets over the untrusted-input surface, `ci/run fuzz`, and a 120 s nightly step. **`FuzzNormalizeName` found two real bugs within seconds:**
1. Names with empty labels (`..`, `a..b`) normalized unchanged and **passed config validation silently** — such a rule could never match a query. Now rejected (`dns.IsDomainName`) for blocked names, override names, and CNAME values.
2. Backslash presentation-format escapes make `dns.Fqdn` **non-idempotent** (`\` → `\.` → `\..`). Out of MVP scope; now rejected at validation.

Both crashers are committed as permanent corpus fixtures (doc 05 §9), with regression unit tests pinning every rejection.

### 2.5 Pilot entry point tested (CustomDNS #8)
`cmd/customdns` was at 0%: `main()` split into a testable `run()` with an injected signal context. Tests cover check-config on the repo example, missing/invalid config, unknown flag, and the full serve→shutdown lifecycle.

## 3. Operational event — R-039 fired live

The instrumented fuzz build **exhausted C:** (106 MB free at the worst point; Phase 0 recorded 27 GB). Applied the standing owner decision (E: for platform storage): `GOCACHE`/`GOMODCACHE` moved to `E:\CouncilCI\cache`, the orphaned C: module cache (~3.5 GB) reclaimed. C: recovered to ~4.2 GB free — still tight; the register keeps R-039 at **Mitigating** and the 85% monitor alert stands. The owner should still plan real headroom on C:.

## 4. Standing state after round 2

- Coverage: broker — arbiter 96, bundle 90, schema 89, agent 84, cmd 68; pilot — entry point and wire path both covered, fuzz corpus live.
- Two repos, both with green CI; Council additionally has enforced protection.
- Known remaining gaps: scored evals await the CDNS-002 harness; mutation testing stays hardened-v1; CustomDNS merge enforcement awaits owner decision #7; two-model live loop awaits `claude setup-token`.
