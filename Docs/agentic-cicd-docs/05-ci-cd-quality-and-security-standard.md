# CI/CD Quality and Security Standard

**Status:** Proposed normative baseline  
**Applies to:** All repositories onboarded to Council (the Agentic CI/CD Platform)  
**Keywords:** MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY are normative requirements.

## 1. Purpose

This standard defines the minimum controls a change must satisfy from pull request through production. It is intentionally independent of programming language and CI vendor. GitHub Actions is the initial control plane, while all substantive checks are invoked through the repository-owned `ci/run` interface.

Agent output is treated as untrusted proposed work. Agents may create patches and findings, but only deterministic policy, protected GitHub rules, and authorized humans can approve exceptions or production promotion.

## 2. Control hierarchy

When controls disagree, the following precedence applies:

1. Safety, legal, credential, and data-protection policy.
2. Deterministic compilation, tests, contracts, and security policy.
3. Approved specification and acceptance manifest.
4. Human review decisions.
5. Agent findings and scores.

An agent verdict MUST NOT waive or downgrade a higher-precedence failure.

## 3. Source-control governance

Every production repository MUST have:

- A protected default branch with direct pushes disabled.
- Pull requests required for changes.
- Required status checks that cannot be bypassed by the implementation agent.
- At least one human approval for high-risk changes, security-sensitive areas, policy changes, release workflows, and protected evals.
- `CODEOWNERS` coverage for workflow, policy, security, deployment, and hidden-eval paths.
- Stale approval dismissal after relevant changes.
- Conversation resolution before merge.
- A documented merge strategy; squash merge is the recommended pilot default.
- Automated dependency-update PRs through Dependabot, Renovate, or an approved equivalent.
- A release-tag and versioning convention.

Repositories SHOULD require signed release tags. Commit signing MAY be required based on risk and contributor workflow.

## 4. Workflow security

GitHub workflows MUST:

- Declare explicit minimal `permissions` at workflow or job level.
- Pin every third-party Action to a full immutable commit SHA.
- Use official GitHub Actions where they reduce external supply-chain exposure.
- Run `actionlint` and `zizmor` against workflow changes.
- Set explicit job timeouts.
- Use concurrency groups to cancel stale runs where cancellation is safe.
- Avoid interpolating untrusted event data directly into shell scripts.
- Use `persist-credentials: false` on read-only checkouts unless a job demonstrably needs Git writes.
- Avoid `pull_request_target` for any job that checks out or executes pull-request content.
- Keep credentialed comment/publication work in a small job that does not execute repository code.

The agent-review workflow SHOULD be defined on the protected default branch and triggered only after deterministic CI. It may read an exact candidate SHA and sanitized evidence, but it MUST NOT execute candidate-controlled hooks or pipeline scripts with personal AI credentials present.

GitHub states that full-length commit-SHA pinning is the only immutable way to reference an Action and warns that persistent self-hosted runners can be compromised by untrusted workflow code. See [GitHub secure use](https://docs.github.com/en/actions/reference/security/secure-use).

## 5. Standard `ci/run` contract

Each repository MUST expose these canonical umbrella commands, even if their internal implementation varies:

```text
ci/run detect
ci/run bootstrap
ci/run fast
ci/run full
ci/run security
ci/run eval
ci/run package
ci/run evidence
```

`fast`, `full`, `security`, `eval`, and `evidence` are profiles whose exact adapter capabilities appear in the output of `detect`. `evidence` accepts the repository's configured PR, nightly, or release profile rather than creating a separate incompatible command name.

Repositories SHOULD expose granular selectors beneath the same interface for targeted local or matrix execution, for example `ci/run check format`, `ci/run check lint`, `ci/run test unit`, and `ci/run test contract`. They MAY expose additional selectors such as `fuzz`, `mutation`, `dast`, `performance`, `sbom`, `sign`, and `deploy-smoke`. Unsupported capabilities MUST return a structured `unsupported` result rather than silently succeeding.

The interface MUST:

- Return zero only when the requested operation passes.
- Emit human-readable logs and machine-readable results.
- Avoid changing tracked source during check-only operations.
- Record exact tool versions.
- Support clean, noninteractive execution.
- Produce the same result locally and in CI when inputs, platform, toolchain, and advisory snapshots are equivalent.
- Separate infrastructure errors from quality failures.

## 6. Required check profiles

### 6.1 Fast local profile

Target: under two minutes for a typical repository.

- Specification and manifest schema validation.
- Formatting check.
- Linting.
- Type checking where supported.
- Incremental build.
- Affected unit tests.
- Fast secret and dependency checks.

This profile provides feedback but does not replace required PR checks.

### 6.2 Pull-request profile

The following are blocking unless explicitly marked non-applicable in repository policy:

| Gate | Minimum rule | Standard output |
|---|---|---|
| Specification traceability | Every mandatory acceptance ID has an executable verifier | JSON manifest |
| Workflow validation | No syntax errors or unwaived high-confidence security findings | SARIF/text |
| Format | No formatter diff | Patch/text |
| Lint | Zero errors; no unauthorized warning regression | SARIF |
| Type check | Zero errors for typed scope | SARIF/text |
| Build | Clean locked build succeeds | Log/metadata |
| Unit tests | Zero failures without retry-to-green | JUnit |
| Integration tests | Required service boundaries pass | JUnit |
| Contracts/schemas | Valid schemas; no unauthorized breaking change | JUnit/JSON |
| Coverage | No material regression; changed-code policy satisfied | LCOV/Cobertura |
| Property tests | No counterexample; seed recorded | JUnit/JSON |
| Secrets | No verified secret | SARIF |
| SAST | No new unwaived critical/high finding | SARIF |
| Dependency vulnerabilities | No new policy-violating vulnerability | SARIF/JSON |
| Agent/eval PR suite | All hard invariants and minimum quality floor pass | JUnit/JSON |

Deterministic failures MUST NOT be automatically retried. A network, rate-limit, registry, or runner error may be retried only after it is classified as infrastructure failure.

### 6.3 Nightly profile

Nightly checks SHOULD include:

- Full operating-system and supported-runtime matrix.
- Full integration and end-to-end suites.
- Mutation testing over changed or critical modules.
- Fuzzing, sanitizer runs, and stored-corpus replay.
- DAST against an isolated preview deployment.
- Performance and resource-budget checks in a controlled environment.
- Fresh vulnerability database scan.
- Full agent and LLM evaluation corpus.
- Flaky-test detection and trend analysis.
- Backup, artifact-verification, or restore tests where applicable.

Nightly failures MUST create a visible issue or alert with an owner. A nightly failure that invalidates release confidence MUST block releases until dispositioned.

### 6.4 Release profile

Release jobs MUST:

- Start from a protected commit or tag.
- Build in a clean environment from locked dependencies.
- Produce an immutable artifact digest.
- Generate an SBOM from the final artifact or image.
- Scan the final artifact with current vulnerability data.
- Apply license policy.
- Sign the artifact and associated attestations.
- Produce provenance identifying source revision, builder, and recipe.
- Verify every required policy over the normalized evidence bundle.
- Publish artifacts to a controlled registry.
- Deploy the same digest that was scanned and approved.

Use either Trivy as the consolidated artifact-scanning family, or Syft plus Grype as a modular family. Duplicate blocking scanners SHOULD NOT be added unless they provide distinct coverage or are being evaluated deliberately.

### 6.5 Deployment profile

Production deployment MUST require:

- A protected GitHub environment or equivalent approval boundary.
- Short-lived workload identity such as OIDC where the deployment target supports it.
- Immutable artifact references, never floating tags alone.
- Signature, signer-identity, provenance, and policy verification.
- Pre-deployment compatibility checks.
- Staging or an equivalent validation environment.
- Automated smoke and health checks.
- A documented, tested rollback mechanism.
- Deployment records linked to commit, artifact digest, evidence, and approver.

GitHub's [OIDC guidance](https://docs.github.com/en/actions/concepts/security/openid-connect) describes using short-lived cloud credentials rather than long-lived deployment secrets.

## 7. Language adapters

Repositories MUST select exactly one primary formatter and one primary linter per language unless a second tool has explicitly non-overlapping coverage.

| Ecosystem | Formatting/lint/type baseline | Test and advanced-test baseline |
|---|---|---|
| Python | Ruff formatter and linter; Pyrefly, mypy, or Pyright | pytest, coverage.py, Hypothesis, mutmut |
| JavaScript/TypeScript | Biome, or ESLint plus Prettier; `tsc` | Vitest/Jest, c8, fast-check, StrykerJS |
| Go | gofmt, go vet, staticcheck or golangci-lint | `go test -race -cover`, native fuzzing, govulncheck |
| Rust | rustfmt, Clippy | cargo test, llvm-cov, Proptest, cargo-fuzz, cargo-mutants |
| Java/Kotlin | Spotless/Checkstyle, SpotBugs | JUnit, JaCoCo, PIT |
| .NET | `dotnet format`, Roslyn analyzers | `dotnet test`, Coverlet, FsCheck, Stryker.NET |
| C/C++ | clang-format, clang-tidy, compiler warnings | CTest, sanitizers, libFuzzer, coverage |

Pyrefly is a type checker; it complements rather than replaces Ruff. Pylint MAY remain for project-specific checks not supplied by Ruff.

An adapter MUST pin or lock its toolchain and SHOULD use a repository-declared version manager or reproducible tool image.

## 8. Cross-language security baseline

The standard baseline is:

- **Workflow:** actionlint and zizmor.
- **Secrets:** Gitleaks.
- **SAST:** Semgrep Community Edition plus native analyzers.
- **Source dependency/SCA:** OSV-Scanner.
- **IaC/container:** Trivy or an approved equivalent.
- **API/web DAST:** OWASP ZAP where applicable.
- **SBOM:** CycloneDX or SPDX using Trivy or Syft.
- **Signing:** Cosign/Sigstore.
- **Policy:** OPA/Conftest or an equivalent deterministic engine.

CodeQL MAY be added for supported languages. It is not assumed in the no-extra-cost private-repository baseline because private-repository availability depends on GitHub Code Security licensing.

Security findings MUST carry a stable identifier, affected component, severity, evidence, owner, status, and—when waived—an expiry.

## 9. Coverage, mutation, and testing policy

- Coverage is a signal, not proof of correctness.
- Changed-code coverage SHOULD be used alongside a repository-specific global floor.
- A coverage percentage MUST NOT compensate for missing acceptance tests.
- Surviving mutations in security-critical logic MUST be reviewed.
- Property and fuzz failures MUST preserve the seed or input as a permanent regression fixture.
- Quarantined tests MUST have an owner, issue, rationale, noncritical classification, and expiry.
- Quarantined tests are reported separately and MUST NOT count as passing.
- Tests MUST NOT be rerun automatically merely to produce a green status.

## 10. Artifact, evidence, and cache policy

Every CI run MUST record:

- Repository, base SHA, candidate SHA, and patch hash.
- Specification and acceptance-manifest hashes.
- Toolchain and scanner versions.
- Runner or container image digest.
- Test, analysis, coverage, and eval outputs.
- Advisory database version or timestamp where relevant.
- Waivers applied to the decision.
- Final policy decision and reason codes.

Use standard formats wherever possible:

- JUnit for tests/evals.
- SARIF for static/security findings.
- LCOV or Cobertura for coverage.
- CycloneDX or SPDX for SBOMs.
- in-toto/SLSA-compatible statements for provenance.
- OpenTelemetry for execution traces.

Caches MUST be namespaced by repository, trust level, lockfile/toolchain hash, and relevant platform. Untrusted jobs MUST NOT write caches later trusted by release jobs. Release correctness MUST NOT depend on a mutable cache.

Raw agent transcripts SHOULD remain local and SHOULD NOT be retained by default. Structured findings and decisions may be retained after secret and personal-data checks.

## 11. Supply-chain integrity

- Dependencies MUST be locked where the ecosystem supports lockfiles.
- Build tools and scanner binaries MUST be version-pinned.
- A centrally maintained quality-tools image SHOULD be referenced by digest.
- The final SBOM MUST describe the released artifact, not only the source tree.
- Vulnerability exceptions SHOULD use expiring VEX or waiver records rather than permanent ignore files.
- The system MUST verify the expected signer and builder identity, not merely the presence of any signature.
- Production promotion MUST preserve the exact artifact digest.

A home-administered runner can provide useful traceability but cannot honestly claim the strongest hosted-builder isolation when the same administrator can alter the build and its provenance. Official releases requiring stronger trust SHOULD use an independently isolated builder. See the [SLSA build track](https://slsa.dev/spec/v1.2/build-track-basics) and [GitHub artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations).

## 12. Agent-specific controls

Agent jobs MUST:

- Receive the exact specification, diff, and evidence hashes.
- Use a fresh reviewer context independent of the writer.
- Return output conforming to the approved schema.
- Remain within explicit time, turn, and repair-cycle limits.
- Operate with the least filesystem, command, and network permissions possible.
- Avoid access to protected hidden tests, policy write paths, production secrets, and merge credentials.
- Report uncertainty and missing evidence rather than manufacturing a pass.

An implementation agent MUST NOT approve its own change. Reviewer consensus MUST NOT replace deterministic verification or required human review.

## 13. Waivers

Waiver eligibility is declared by policy, not inferred from severity. The final policy statuses are:

- `pass`: all required controls pass.
- `pass_with_waiver`: a policy-declared waivable finding remains failed but an authorized, scoped, unexpired waiver supplies compensating controls.
- `blocked`: a required or non-waivable control fails.
- `pending`: required evidence, agent review, or human approval is unavailable.

The following are non-waivable for the candidate under evaluation:

- Missing, stale, malformed, or hash-mismatched required evidence.
- Compilation failure or failure of a hard acceptance invariant.
- Confirmed credential disclosure, sandbox escape, or protected-network access.
- Unauthorized change to a protected specification, policy, hidden test, workflow, or approval rule.
- Artifact digest, signature, signer, provenance, or deployment-identity mismatch.
- Unauthorized production deployment or inability to execute the required rollback for a risky change.

Changing an intended acceptance invariant requires an approved specification/policy revision and a complete rerun; it is not a waiver of the old result. A vulnerability or operational threshold may be waivable only when its policy explicitly allows it and all requirements below are satisfied.

A waiver is permitted only when it:

- Identifies one exact policy or finding.
- States affected scope and business justification.
- Names an accountable owner and approver.
- Links a remediation issue.
- Has an expiry date.
- Records compensating controls.
- Is visible in the evidence bundle and final decision.

Agents may draft waiver text but MUST NOT approve or apply it. Use the [policy waiver template](templates/policy-waiver-template.md).

## 14. Required status checks for the pilot

The pilot repository should expose these stable GitHub check names:

```text
spec / traceability
ci / format-lint-type
ci / build-unit
ci / integration-contract
security / secrets-sast-sca
eval / acceptance
policy / decision
```

`policy / decision` is the single normalized final gate, but the underlying checks SHOULD also remain visible so failure causes are immediately understandable.

## 15. Review cadence

This standard SHOULD be reviewed:

- After the pilot retrospective.
- At least quarterly thereafter.
- After a runner or credential compromise.
- After a material GitHub, Claude, Codex, ARC, or policy-engine change.
- Before claiming a higher supply-chain assurance level.

Changes to this standard require security-owner approval and must be exercised against a test repository before adoption.

## References

- [GitHub secure use reference](https://docs.github.com/en/actions/reference/security/secure-use)
- [GitHub Actions workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [Semgrep Community Edition](https://semgrep.dev/docs/tags/semgrep-ce)
- [OSV-Scanner](https://google.github.io/osv-scanner/usage/)
- [Cosign quick start](https://docs.sigstore.dev/quickstart/quickstart-cosign/)
- [OPA for CI/CD](https://www.openpolicyagent.org/docs/cicd)
- [SLSA build provenance](https://slsa.dev/spec/v1.2/build-provenance)
