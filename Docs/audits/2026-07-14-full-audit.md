# Full Audit — 2026-07-14

**Scope:** everything delivered in the Phases 0–5 build session — re-verification of all claims, execution of all suites, and a quality audit of the test suites and evaluation criteria themselves.
**Auditor:** platform agent (self-audit, adversarial posture); all fixes landed via reviewed PRs.

## 1. Re-verification of prior claims (all confirmed)

| Claim | Re-verified how | Result |
|---|---|---|
| Broker builds; tests pass | `go vet`, `gofmt -l`, `go test -race -count=1 ./...` | Clean |
| Pilot pipeline green | `ci/run full` + `ci/run security` from clean main | 13/13 checks pass |
| Local runner isolated | `reachability-test.sh` re-run live on the VM | PASS (persists) |
| Golden snapshot exists | `VBoxManage snapshot list` | `golden-runner-v1` present |
| Evidence artifacts fixed | Artifact sizes on two completed CI runs | 12–16 KB (previously empty) |
| force_hosted fallback | Job-level runner attribution on run 29320608015 | `ubuntu-latest`, success |

## 2. Audit findings and fixes (all landed via PRs)

### Finding 1 — the untrusted-input entry point had 0% coverage (HIGH)
`server.handle` — the function that receives raw DNS packets — was never executed by any test; every CDNS-001 criterion was verified only at the resolver-unit level.
**Fix (CustomDNS #6):** six end-to-end tests over live loopback UDP/TCP sockets through the production resolver chain, including forwarding to an in-test fake upstream, SERVFAIL on resolver failure, and FORMERR on a question-less packet. `handle` 0%→88.9%. The CDNS-001 manifest gained wire-level criteria **AC-017..AC-022** so the *evaluation contract itself* now requires end-to-end evidence.

### Finding 2 — `pass_with_waiver` was specified but unimplemented (HIGH)
Doc 04 §7.5 row 4 and doc 05 §13 define the waiver path; the arbiter had no waiver input, so the platform could not express its own policy.
**Fix (Council #9):** waiver support with the standard's constraints enforced in code — exact finding-ID scope, named approver required, deterministic RFC3339 expiry (injected `--now`, fail-closed on unparseable dates), and structurally unable to clear hard-gate or evidence failures. Proven by unit tests and a live CLI drill (valid → `pass_with_waiver`/exit 0; expired → `blocked`/exit 3).

### Finding 3 — councilci was never tested in CI (HIGH)
The traceability gate (the executable half of PAC-001) had no unit tests, and being a separate Go module, `ci/run` never ran any tests for it.
**Fix (CustomDNS #7):** unit tests pinning each failure mode the gate exists to catch (unresolvable verifier, frozen-spec tamper, hard-criterion-with-eval-verifier, schema-invalid manifest) plus the real repository manifest as a fixture; new `test/councilci` hard check in `ci/run full`.

### Finding 4 — three broker packages at 0% coverage (MEDIUM)
`schema`, `agent`, and `cmd/broker` were exercised only by manual CLI drills.
**Fix (Council #9):** schema tests (fixtures validate; enum, provider/client pairing, and unknown-field rejection; **regexp2 lookahead proven to reject backslash paths**); agent tests (**the review env allowlist never passes `ANTHROPIC_API_KEY`/`OPENAI_API_KEY`/`GITHUB_TOKEN`/`CLAUDE_CODE_OAUTH_TOKEN`**); cmd tests (seal tamper/missing/absent, cross-review `new_findings`, waiver loading). Coverage: agent 0→84%, schema 0→89%, arbiter 90→96%.

### Finding 5 — unpinned decision-matrix rows (MEDIUM)
Several §7.5 precedence rows had no dedicated test (medium+violation blocks, medium-without annotates, low annotates, `accept`/`needs_reproducer` dispositions).
**Fix (Council #9):** one test per row; the broker README now carries the complete row→test map, so a future rule change that breaks a row fails a named test.

### Finding 6 — register/doc status drift (LOW)
R-039 read `Open` in the register but `Mitigating` in the Phase 0 record.
**Fix (this PR):** register row updated with the actual mitigation state.

## 3. Evaluation-criteria standard (after audit)

- **CDNS-001**: 22 hard criteria, each resolving to a real, compilable test — now spanning unit *and* wire level; traceability re-verified (22/22).
- **Traceability gate**: schema validation (JSON Schema 2020-12, ECMA-262 patterns), frozen-spec hash integrity, and verifier existence — each failure mode pinned by a unit test that runs in CI.
- **Broker decision policy**: every doc-04 §7.5 row and every doc-05 §13 waiver constraint has a named test; integrity invariants (bundle tamper, added-file injection, seal re-verification, env allowlist) are all tested.
- **Known remaining gaps (documented, not hidden):** `cmd/broker` sits at ~21% (CLI plumbing; core logic is tested at the package level); `bundle` at 76% (Write/Load round-trip untested); pilot total coverage 65% with `cmd/customdns` main untested (flag wiring); scored-eval criteria (CDNS-002 AC for latency) still await the eval harness; mutation testing remains a hardened-v1 item per doc 05 §6.3.

## 4. Verdict

All prior claims survived re-verification. The audit found no incorrect *behavior* in what was built, but found real gaps in what the tests *guaranteed* — most seriously the untested network entry point, the unimplemented waiver path, and a load-bearing tool outside CI. All are fixed, tested, and merged; the evaluation contracts (acceptance manifest, decision-matrix map) were strengthened so these guarantees are now enforced by named, CI-run tests rather than by claim.
