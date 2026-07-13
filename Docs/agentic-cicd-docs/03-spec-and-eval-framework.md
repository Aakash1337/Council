# Specification and Evaluation Framework

**Status:** Proposed baseline

**Scope:** Language-agnostic, repository-owned specification and evaluation policy

**Normative terms:** **MUST**, **SHOULD**, and **MAY** indicate requirement strength.

## 1. Purpose

This framework turns a feature request into an auditable chain:

```text
intent -> approved specification -> acceptance manifest -> implementation
       -> deterministic evidence -> scored evaluations -> release decision
```

The specification describes the required behavior. The acceptance manifest identifies how each claim will be verified. CI produces evidence against an immutable commit and specification revision. No agent's opinion substitutes for executable evidence.

## 2. Specification-system decision

Use [GitHub Spec Kit](https://github.com/github/spec-kit) for the pilot. Its workflow separates project principles, requirements, technical planning, tasks, consistency analysis, and implementation; it also supports many coding-agent integrations and both greenfield and brownfield work.

Use [OpenSpec](https://github.com/Fission-AI/OpenSpec) instead when the dominant workflow is small brownfield change proposals, change archives, or specifications shared across several repositories. Do not install both in one repository during the pilot. Two sources of specification truth create more risk than either tool removes.

The selected tool is an authoring layer, not the merge authority. Both feed the repository-owned acceptance manifest defined below.

### 2.1 Selection rule

| Situation | Choice |
|---|---|
| GitHub-centered project, structured product-to-task flow, mixed agents | Spec Kit |
| Existing systems, compact change deltas, multi-repository spec store | OpenSpec |
| Regulated or safety-critical project | Either, with a project-specific template and mandatory human approval |

Record the choice in an ADR. Changing systems requires a migration plan that preserves acceptance IDs and archived decisions.

## 3. Repository contract

Recommended layout:

```text
.specify/                    # Spec Kit configuration, if selected
specs/
  constitution.md            # Engineering principles and non-negotiable constraints
  changes/<change-id>/
    specification.md         # What and why
    plan.md                  # How
    tasks.md                 # Work breakdown
    acceptance.yaml          # Machine-readable verification contract
    decisions/               # Change-local ADRs
evals/
  public/                    # Visible cases and fixtures
  datasets/                  # Versioned, non-sensitive eval datasets
  promptfoo/                 # Promptfoo configuration, when applicable
ci/
  schemas/change-manifest.schema.json
  schemas/agent-review.schema.json
  schemas/agent-cross-review.schema.json
  schemas/evidence-bundle.schema.json
  run                         # Language-neutral entry point
policy/
  quality.rego               # Optional policy-as-code implementation
```

Specifications, manifests, schemas, and policy are reviewed code. Changes to them MUST appear in the pull-request diff. Protected policy changes require a code owner who is not the implementation agent.

## 4. Specification lifecycle

1. **Constitution:** Define quality, security, privacy, compatibility, observability, performance, and testing principles.
2. **Specify:** Capture actors, scenarios, expected outcomes, exclusions, failure behavior, and non-functional requirements without prematurely choosing implementation details.
3. **Clarify:** Resolve ambiguity. Unknowns are explicit; an agent MUST NOT silently invent a product decision that materially changes scope.
4. **Plan:** Select the architecture, data changes, migration/rollback strategy, dependencies, and risk tier.
5. **Create acceptance manifest:** Give every testable claim a stable acceptance ID and verifier.
6. **Analyze:** Check consistency among specification, plan, tasks, and acceptance manifest before implementation.
7. **Implement:** Work against the approved specification revision in an isolated branch or worktree.
8. **Verify:** Run public checks, protected holdouts where required, and risk-tier evaluation suites.
9. **Converge:** Compare the implementation and evidence with the specification; add unresolved work rather than rewriting history.
10. **Archive:** Preserve the final manifest, evidence index, decisions, and regression tests with the merged commit.

The commit SHA and SHA-256 digest of the specification bundle MUST be frozen before an unattended agent run. If either changes, prior evidence is stale.

## 5. Executable acceptance manifest

The manifest provides traceability; it does not embed all test logic. It maps each acceptance criterion to a stable verifier owned by the repository or protected evaluation service.

```yaml
schema_version: 1
change_id: AUTH-123
title: Atomic refresh-token rotation
risk_tier: high                 # low | medium | high | critical

subject:
  base_sha: 6c9ad4f63d14223c7a449f238b863fcb5e9bb1aa
  head_sha: 9e087cb9e20302af71f672559f8c25af990242bb
  spec_path: specs/changes/AUTH-123/specification.md
  spec_sha256: 6bc7d2a9f5f01d8ac83e43885655810d34a70be4dfef8a749f706b0244570ca6

criteria:
  - id: AC-001
    statement: A refresh token can be redeemed no more than once.
    criticality: hard           # hard | scored | advisory
    owner: team-auth
    verifier:
      kind: test                # test | contract | policy | eval | manual
      ref: test://tests/auth/test_rotation.py::test_single_redemption
      visibility: public        # public | hidden
    evidence:
      format: junit
      required: true

  - id: AC-002
    statement: Two concurrent redemptions produce exactly one new session.
    criticality: hard
    owner: team-auth
    verifier:
      kind: test
      ref: holdout://auth/concurrent-rotation/v3
      visibility: hidden
    evidence:
      format: junit
      required: true

  - id: AC-003
    statement: P95 rotation latency does not regress by more than 10%.
    criticality: scored
    owner: team-auth
    verifier:
      kind: eval
      ref: eval://auth/rotation-performance
      visibility: public
    threshold:
      metric: p95_ms_ratio_to_baseline
      operator: lte
      value: 1.10
      minimum_samples: 30
    evidence:
      format: eval-json
      required: true

waivers: []
```

### 5.1 Manifest validation rules

- Every normative, testable specification statement MUST map to at least one acceptance ID.
- Criterion IDs are local within a change manifest; their globally unique form is `<change_id>/<criterion_id>` (for example, `AUTH-123/AC-002`) and review/evidence records MUST use that qualified form.
- Every acceptance ID MUST have a verifier, owner, criticality, and required evidence type.
- A hard criterion MUST use a deterministic verifier unless a human approval is explicitly required.
- A hidden verifier uses an opaque handle; its test content and storage location are unavailable to the writer.
- A command embedded in an untrusted pull request MUST NOT be executed by the credential-bearing agent broker. Commands run only in the isolated build/test environment.
- Manual criteria require an approver identity, timestamp, reason, and evidence reference.
- Waivers require an owner, issue, justification, scope, and expiration. Agents cannot approve waivers.
- CI MUST fail closed if a required verifier is missing, skipped, malformed, or executed against a different SHA.

## 6. Public and hidden evaluations

### 6.1 Public suite

Public evaluations are committed to the repository. They teach contributors the intended behavior and support fast local iteration. Include representative happy paths, edge cases, contracts, property tests, and at least one failure case for each major capability.

Public tests are necessary but insufficient for agent-written changes: an implementation agent can overfit to visible examples or accidentally weaken them.

### 6.2 Withheld checks and protected holdouts

Non-public evaluations exist in two tiers with different threat models:

- **Withheld check:** unavailable to the implementation agent. Protects against accidental overfitting and prevents the writer from editing the exact test. This is the pilot and MVP requirement.
- **Protected holdout:** additionally designed to resist deliberate discovery by hostile code or contributors. Requires separately controlled repository or test-service infrastructure. This is **not** a pilot or MVP requirement.

In pilot scope, `visibility: hidden` in the acceptance manifest denotes a withheld check; the manifest schema is unchanged.

The pilot and MVP requirement is: at least one withheld check for a selected high-risk criterion, executed outside the implementation agent's authorized workspace and Git object scope.

Withheld checks MUST:

- Live outside the agent's **entire authorized filesystem and Git object scope**. Outside the worktree is insufficient: sibling directories, packfiles, alternates, and the reflog are all inside Git scope and can leak content.
- Run under a separate eval-runner OS identity, credential-free, and preferably without network access.
- Execute only against the frozen candidate commit; the candidate is frozen before the withheld suite is retrieved or run. The runner holds no AI, deployment, or repository-write credential; if retrieval requires authentication, use a narrowly scoped short-lived read credential that is removed before repository code runs.
- Record an automatic content digest of the suite in the evidence bundle (the existing artifact `sha256` fields) rather than managing an opaque suite version by hand.
- Return only criterion IDs and bounded failure information, without exposing secrets or complete fixtures.
- Be rotated after actual or suspected disclosure.
- Preserve failed cases as protected regression tests.
- Prefer property, metamorphic, differential, and fuzz tests with fresh seeds, which resist memorization better than fixed fixtures.

**Explicit limitation:** withheld checks do not protect against deliberately malicious candidate code probing for the oracle. That threat requires the protected-holdout tier.

Protected-holdout infrastructure moves to hardened v1 and is introduced only when triggered by: untrusted contributors, sensitive fixtures, release-critical holdouts, multiple repositories sharing a suite, repeated leakage, or evidence of deliberate gaming.

Withheld tests must not become mysterious product requirements. The public specification still describes the behavior being measured; only the exact cases remain withheld.

## 7. Gate classes and precedence

| Gate class | Examples | Merge behavior |
|---|---|---|
| Hard deterministic | Compile/build, schema validation, unit/contract tests, type checks, policy, secret detection, critical security findings | Any valid failure blocks |
| Hard threshold | Coverage floor, error budget, maximum binary size, license allowlist | Threshold breach blocks |
| Scored | Agent-task success, semantic quality, performance distribution, LLM-product quality | Blocks only under a versioned policy threshold |
| Advisory | Style suggestions, low-confidence model review, non-regressing informational metrics | Annotates; does not block |
| Human | Product ambiguity, irreversible migration, risk acceptance | Blocks until named approval |

Precedence is strict:

1. Missing or invalid evidence fails closed.
2. Hard gates cannot be offset by a high aggregate score.
3. Security, privacy, data-loss, and acceptance-invariant failures cannot be waived by agent consensus.
4. Scored gates are evaluated only after hard gates pass.
5. A waiver changes policy for a limited period; it does not change the underlying result.

## 8. Evaluation suites by cadence

### Pull request

- Acceptance-manifest schema and traceability.
- Format, lint, type, build, unit, focused integration and contract tests.
- Changed-code coverage and property-based tests.
- Fast SAST, secret and dependency checks.
- Small public agent/product eval set.
- Hidden smoke holdouts for high-risk changes.

Target a predictable duration. Slow checks move to nightly only when the PR suite retains a meaningful signal.

### Nightly

- Full integration and end-to-end suites.
- Mutation testing, fuzzing, sanitizers, deeper SAST/DAST, and performance tests.
- Full public and hidden eval corpus with repeated stochastic trials.
- Baseline comparisons and flaky-test detection.

### Release

- Clean rebuild of the reviewed commit.
- All release-critical acceptance criteria.
- Artifact/container scan, SBOM, signing and provenance checks.
- Staging deployment, smoke tests, rollback verification, and required human approvals.

## 9. Statistics and nondeterminism policy

### 9.1 Deterministic checks

- Do not retry a semantic failure until it turns green.
- Retry only a classified infrastructure failure, such as runner loss or registry timeout, and record every attempt.
- If a supposedly deterministic check produces different outcomes on identical inputs, classify it as flaky and fail high/critical-risk changes until it is repaired or formally quarantined.

### 9.2 Stochastic model and agent evaluations

Record every trial, including failures. The primary reliability metric is **pass@1**; pass@k MAY be reported as diagnostic information but MUST NOT hide poor first-attempt reliability.

Recommended minimums:

| Cadence/risk | Trials per case | Decision rule |
|---|---:|---|
| PR, low/medium | 1 public trial | Hard invariants plus score floor |
| PR, high/critical | 3 independent trials | No critical failure; capability floor met |
| Nightly/release | 5 or more independent trials | Absolute floor and baseline-regression policy |

Use fresh sessions and fixed evaluation inputs. Record model/client version, configuration, tool permissions, temperature or effort controls where available, and infrastructure metadata.

For binary outcomes, report the observed pass rate and a Wilson confidence interval when sample size supports it. For continuous metrics, report count, median, p95, dispersion, and a bootstrap 95% interval when there are at least 30 observations. Small samples use conservative point thresholds and are labeled insufficient for trend claims.

A scored gate SHOULD require both:

- An absolute minimum per critical capability, so an aggregate cannot conceal a collapsed subgroup.
- No statistically or operationally material regression from the approved baseline.

The baseline is a versioned result from the default branch, not “the last run.” An invalid, flaky, or infrastructure-failed run cannot replace it.

### 9.3 Flakiness and quarantine

A check is a quarantine candidate only after reproduced non-determinism on identical source, dependencies, seed, and environment. Quarantine requires an owner, issue, reason, scope, expiration, and non-blocking replacement signal.

Security tests, hidden acceptance invariants, data-integrity tests, and migration/rollback tests MUST NOT be silently quarantined. Repeated model-eval variance is handled with sampling and thresholds, not by deleting difficult cases.

## 10. Model-graded evaluation rules

LLM judges are useful for qualities that resist exact comparison, such as relevance or explanation quality. They are not trusted to establish security, compilation, data integrity, or protocol conformance.

- Apply deterministic assertions before an LLM judge.
- Validate judge output against JSON Schema.
- Give the judge only the evidence needed for the rubric and delimit untrusted content as data.
- Never allow repository text to rewrite the judge rubric or tool policy.
- Calibrate each rubric against a human-labeled set and track false-positive/false-negative rates.
- Blind the judge to provider identity and expected winner when comparing candidates.
- Preserve concise evidence and rationale, not private chain-of-thought.
- A judge-version change creates a new baseline; do not compare it silently with old scores.
- A sole model judgment cannot be the final blocker for a high-impact change. Require a deterministic reproducer, a second independent judgment, or human review.

## 11. Evidence model

Each CI attempt emits a content-addressed evidence bundle:

```text
runs/<run-id>/
  manifest.json
  acceptance.yaml
  checks/
    junit/
    sarif/
    coverage/
    evals/
  sbom/
  traces/
  index.json
```

The run manifest includes:

- Run, workflow, repository, base/head commit, tree, and specification hashes.
- Trigger and trust classification.
- Runner image/VM, OS, architecture, and toolchain versions.
- Dependency lockfile hashes and container image digests.
- Commands/adapter versions, start/end times, exit codes, and retry classification.
- Acceptance IDs evaluated and evidence paths/hashes.
- Model client/model identifiers, role, structured-output schema version, and usage/latency metadata when agents participate.
- Redaction status, retention class, signature/attestation reference, and parent attempt.

Preferred interoperable formats:

| Evidence | Format |
|---|---|
| Test results | JUnit XML |
| Static/security findings | [SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html) |
| Coverage | LCOV or Cobertura XML |
| Eval results | Versioned JSON plus optional JUnit projection |
| SBOM | [CycloneDX](https://cyclonedx.org/specification/overview/) or [SPDX](https://spdx.dev/use/specifications/) |
| Agent review | JSON validated against [JSON Schema 2020-12](https://json-schema.org/draft/2020-12) |
| Telemetry | [OpenTelemetry](https://opentelemetry.io/docs/specs/otel/) traces and metrics |

Evidence is immutable after the decision. Corrections create a new attempt linked to the prior one. Logs are artifacts, not authority: the structured evidence record and its hash drive policy.

## 12. Promptfoo and Phoenix roles

### Promptfoo: evaluation runner

[Promptfoo's coding-agent evaluation guide](https://www.promptfoo.dev/docs/guides/evaluate-coding-agents/) describes repository-level coding-agent tasks and assertions; its [assertion documentation](https://www.promptfoo.dev/docs/configuration/expected-outputs/) and [CI integration guide](https://www.promptfoo.dev/docs/integrations/ci-cd/) cover programmatic checks and pipeline output.

Use Promptfoo for:

- Versioned LLM/product test cases and rubrics.
- Public coding-task corpora and captured-agent-output evaluation.
- Deterministic assertions combined with carefully calibrated model grading.
- Baseline comparison and machine-readable CI reports.

Promptfoo is not the source of truth for ordinary compilation or unit tests, and it MUST NOT receive exported Claude/Codex subscription credentials. For subscription-funded coding-agent trials, either evaluate captured outputs or invoke an approved local adapter that starts the official native CLI while credentials remain in that client's protected account.

### Phoenix: local observability and experiment analysis

[Phoenix supports self-hosted deployment](https://arize.com/docs/phoenix/self-hosting) and OpenTelemetry-based tracing. Use it as the local UI and analysis store for agent traces, datasets, experiment comparisons, latency, errors, and token/usage metadata.

Phoenix is not the merge ledger. The immutable evidence bundle remains authoritative if Phoenix is unavailable or rebuilt. Redact secrets, credentials, personal data, hidden-test content, and unnecessary source before export. Begin with Phoenix alone; add another observability platform only when a measured requirement justifies the operational cost.

## 13. Quality-policy decision algorithm

```text
if subject hashes do not match manifest: BLOCK
if required evidence is absent, stale, or malformed: BLOCK
if any hard criterion fails: BLOCK
if required human approval is absent: BLOCK
if any critical-capability scored floor fails: BLOCK
if approved baseline regression exceeds policy: BLOCK
if only advisory findings remain: PASS WITH ANNOTATIONS
otherwise: PASS
```

An agent may propose a test, threshold, waiver, or baseline update. Only repository policy and authorized human review can accept it.

## 14. Definition of done for the MVP specification/evaluation capability

The framework is operational when one pilot repository can demonstrate all of the following:

- A frozen specification and machine-validated acceptance manifest.
- Complete acceptance-ID-to-verifier traceability.
- Public tests and at least one withheld check for a selected high-risk criterion, executed outside the implementation agent's authorized workspace and Git object scope; a separately controlled holdout service is not a pilot or MVP requirement.
- Separation of deterministic, scored, advisory, and human gates.
- Immutable evidence linked to the exact source and specification hashes.
- A reproducible baseline and documented flaky-test procedure.
- Promptfoo used only where semantic/model evaluation is needed.
- Repository-native evidence remains authoritative and CI remains functional without an observability service.
- When Phoenix is enabled in hardened v1, it receives only redacted traces and is not the merge ledger.
- A protected policy change cannot be approved solely by the implementing agent.

## 15. Primary references

- [GitHub Spec Kit](https://github.com/github/spec-kit)
- [OpenSpec](https://github.com/Fission-AI/OpenSpec)
- [Promptfoo: Evaluate coding agents](https://www.promptfoo.dev/docs/guides/evaluate-coding-agents/)
- [Promptfoo assertions and metrics](https://www.promptfoo.dev/docs/configuration/expected-outputs/)
- [Promptfoo CI/CD integration](https://www.promptfoo.dev/docs/integrations/ci-cd/)
- [Phoenix self-hosting](https://arize.com/docs/phoenix/self-hosting)
- [OpenTelemetry specification](https://opentelemetry.io/docs/specs/otel/)
- [SARIF 2.1.0 specification](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html)
- [CycloneDX specification overview](https://cyclonedx.org/specification/overview/)
- [SPDX specifications](https://spdx.dev/use/specifications/)
