# System Architecture

| Field | Value |
|---|---|
| Document ID | ACI-ARCH-002 |
| Version | 0.1 |
| Status | Proposed baseline |
| Date | 2026-07-13 |
| Audience | Maintainers, platform engineers, security reviewers, and agent-workflow authors |
| Scope | Local-first, language-agnostic CI/CD with specification-driven development and two-agent review |
| Related documents | [Project charter](01-project-charter-and-requirements.md), [Specification and evaluation framework](03-spec-and-eval-framework.md), [Infrastructure and deployment](06-infrastructure-and-deployment.md), [Risk register](09-risk-register-and-decisions.md) |

## 1. Purpose

This document defines the system architecture for a GitHub-controlled, home-server-executed CI/CD platform that combines deterministic software assurance with Claude Code and Codex. The platform is designed to:

- Keep ordinary build and test execution local when safe and available.
- Preserve GitHub pull requests, checks, rulesets, environments, and hosted-runner fallback.
- Treat repository contents, generated code, dependencies, and model output as untrusted inputs.
- Use specifications and executable acceptance criteria as the definition of done.
- Allow Claude and Codex to implement, independently review, challenge, and repair changes without letting model consensus override a failed deterministic gate.
- Reuse the same repository-owned commands from a developer workstation, GitHub Actions, Bitbucket Pipelines, or another CI controller.
- Use existing ChatGPT Pro and Claude Max access through their first-party clients without silently incurring API charges.

The design deliberately separates the GitHub control plane, code-execution plane, credential-bearing agent plane, evidence plane, and release plane.

## 2. Architectural decisions

| Decision | Rationale | Consequence |
|---|---|---|
| GitHub remains the control plane | It already provides pull requests, checks, rulesets, environments, audit history, artifacts, and hosted runners | The project does not need to recreate a forge or workflow scheduler |
| Pipeline behavior lives behind ci/run and language adapters | YAML is a poor place for complex, portable business logic | Local, GitHub, and future Bitbucket runs invoke the same commands |
| Deterministic checks are authoritative | Agents are nondeterministic and can agree on an incorrect conclusion | Compile, test, security, policy, and provenance failures always block |
| Agent reviews start blind | Early sharing can anchor one model to the other model's answer | Independent findings are sealed before cross-review |
| Credentials and arbitrary code execution occupy different trust zones | Pull requests and dependencies can contain prompt injection or malicious scripts | The target architecture uses a credential broker plus credential-free execution sandboxes |
| Evidence is immutable and content-addressed | Review conclusions must be attributable to an exact specification, commit, and tool run | Every report references hashes and is rejected if its inputs do not match |
| Build once, promote by digest | Rebuilding per environment creates drift and weakens rollback | Staging and production receive the same signed artifact |
| Subscription clients do not fall back to paid APIs | The initial operating budget assumes existing subscriptions only | Agent jobs pause on exhausted usage, provider outage, or expired login |
| Memory is curated, not an unrestricted transcript | Raw conversational history is noisy, can contain secrets, and creates instruction-injection risk | Durable memory consists of accepted specifications, ADRs, tests, decisions, and resolved findings |

## 3. Pilot and target architecture

The pilot establishes a useful, auditable loop without requiring a Kubernetes or VM-fleet project first. The target architecture hardens isolation and automates routing only after the pipeline contracts are stable.

| Concern | Pilot | Target state |
|---|---|---|
| GitHub control plane | GitHub Actions and protected branches | Same |
| Local deterministic runner | Dedicated Ubuntu VM for private, owner-controlled branches | Ephemeral ARC pods for trusted workloads plus disposable Incus VMs for lower-trust jobs |
| Fork or public PR execution | GitHub-hosted ubuntu-latest | Same |
| Agent execution | Dedicated agent VM; manual or protected-branch invocation | Policy-enforcing broker with isolated worktrees and credential-free command workers |
| Claude/Codex collaboration | Official Codex plugin for Claude Code, supervised, with loop limits | Direct first-party CLIs, strict schemas, blind reviews, cross-review, and deterministic arbitration |
| Local/hosted fallback | Manual force_hosted input | Signed runner heartbeat, router, and external watchdog |
| Evidence storage | GitHub artifacts plus a local filesystem mirror | Immutable S3-compatible local object storage plus GitHub check summaries |
| Observability | Structured logs and run manifests | OpenTelemetry traces and a local Phoenix deployment |
| Release | Separate protected workflow and runner | Disposable release worker, signing, SBOM, provenance, policy, staged promotion, and rollback automation |

The pilot is restricted to private repositories and trusted contributors on the local persistent runner. A public repository, forked pull request, or unknown contributor must use a fresh GitHub-hosted runner until disposable local VMs are operational.

## 4. Trust model

### 4.1 Assumptions

The architecture assumes:

- A pull request can contain hostile workflow files, build scripts, tests, compiler plugins, package lifecycle hooks, and prompt-injection text.
- A compromised dependency can behave like hostile repository code.
- Claude and Codex can hallucinate, miss defects, follow malicious repository instructions, or reveal data made available to them.
- A self-hosted runner is not clean merely because a job completed.
- A home-lab administrator can alter the builder and provenance. Locally produced provenance is still useful for traceability, but it should not be presented as a high SLSA build level without the required isolation and controls.
- A failed hard check cannot be waived by an agent. Waivers require an identified human owner, reason, scope, and expiration.

GitHub warns that persistent self-hosted runners can be compromised by untrusted workflow code and recommends especially careful treatment of public and forked pull requests. See [Secure use reference](https://docs.github.com/en/actions/reference/security/secure-use).

### 4.2 Trust zones

| Zone | Trust level | May execute repository code? | Sensitive credentials | Main controls |
|---|---:|---:|---|---|
| GitHub control plane | High | Workflow metadata only | GitHub-managed token, narrowly scoped GitHub App if needed | Protected default branch, rulesets, required checks, pinned Actions, minimal permissions |
| Hosted untrusted runner | Low and disposable | Yes | None beyond job-scoped read token | Fresh GitHub-hosted environment, no subscription credentials, no deployment secrets |
| Local build runner | Low | Yes | No AI or deployment credentials | Disposable VM or restricted pilot VM, egress controls, no LAN access, no Docker socket |
| Agent broker | High | No arbitrary repository commands | Claude and Codex first-party login state | Dedicated OS identity, read-only inputs, tool allowlist, outbound provider access only |
| Agent worktree | Medium | File reads and controlled writes | No directly readable credentials | Sandboxed client, no network for child tools, patch-only output, exact commit binding |
| Evidence store | High integrity | No | Storage service identity | Append-only objects, content hashes, retention policy, separate write/read roles |
| Release plane | Highest | Reviewed build recipe only | Short-lived OIDC identity or signing identity | Protected environment, clean checkout, immutable input digest, no PR-triggered secrets |
| Deployment target | Production trust | Artifact startup and health checks | Runtime secrets | Environment isolation, policy checks, canary or staged rollout, rollback |

### 4.3 Topology

~~~mermaid
flowchart TD
    subgraph GH["GitHub control plane"]
        REPO["Repository, PR, and rulesets"]
        ROUTER["Workflow router"]
        CHECKS["Checks and summaries"]
    end

    subgraph EXEC["Credential-free execution"]
        HOSTED["GitHub-hosted runner"]
        LOCAL["Local disposable runner"]
        HARNESS["ci/run and language adapters"]
    end

    subgraph AGENT["Trusted agent zone"]
        BROKER["Agent broker"]
        CLAUDE["Claude Code"]
        CODEX["Codex CLI"]
    end

    subgraph DATA["Evidence plane"]
        BUNDLE["Immutable evidence bundle"]
        MEMORY["Curated project memory"]
    end

    subgraph RELEASE["Protected release plane"]
        BUILDER["Clean release builder"]
        TARGET["Staging and production"]
    end

    REPO --> ROUTER
    ROUTER --> HOSTED
    ROUTER --> LOCAL
    HOSTED --> HARNESS
    LOCAL --> HARNESS
    HARNESS --> BUNDLE
    BUNDLE --> CHECKS
    BUNDLE --> BROKER
    BROKER --> CLAUDE
    BROKER --> CODEX
    CLAUDE --> BUNDLE
    CODEX --> BUNDLE
    BUNDLE --> MEMORY
    CHECKS --> BUILDER
    BUILDER --> TARGET
~~~

No inbound Internet port is required for an ordinary GitHub self-hosted runner; the runner initiates outbound communication to GitHub. Network policy should nevertheless be based on GitHub's current [self-hosted runner communication requirements](https://docs.github.com/en/actions/reference/runners/self-hosted-runners).

## 5. Component responsibilities

| Component | Responsibilities | Explicit non-responsibilities |
|---|---|---|
| GitHub repository and rulesets | Source history, pull requests, owners, branch protection, workflow definitions, required checks | Running arbitrary privileged repair code |
| Workflow router | Classify trust, select local or hosted execution, expose manual override | Holding Claude, Codex, signing, or deployment credentials |
| ci/run | Stable entry point for detect, bootstrap, fast, full, security, eval, package, and evidence commands | Provider authentication and merge decisions |
| Language adapters | Ecosystem-specific formatting, linting, typing, compilation, tests, coverage, fuzzing, and mutation hooks | Reimplementing package managers |
| Evidence collector | Normalize JUnit, SARIF, coverage, SBOM, logs, and manifests; hash every input and output | Deciding whether an agent's prose is correct |
| Agent broker | Start official clients, enforce schemas and budgets, seal initial reviews, mediate cross-review | Executing uncontrolled repository shell commands or overriding hard gates |
| Claude Code and Codex | Implement or review against a supplied evidence bundle; produce evidence-linked structured output | Acting as final merge authority |
| Policy arbiter | Combine deterministic gate state with structured findings and waiver records | Free-form model debate |
| Evidence store | Persist immutable per-run bundles and expose read-only retrieval | Becoming an uncurated vector memory |
| Curated memory | Store accepted ADRs, regression cases, resolved findings, and project conventions | Raw hidden reasoning or transient chat |
| Release builder | Rebuild from a protected commit, scan, generate SBOM/provenance, sign, and publish | Reusing an untrusted PR workspace |
| Deployment controller | Promote an immutable digest, run health checks, and roll back | Building a new artifact in each environment |

## 6. Repository interface

The repository presents a controller-neutral interface:

~~~text
.github/workflows/
  pr.yml
  agent-review.yml
  nightly.yml
  release.yml
  deploy.yml

ci/
  run
  adapters/
  schemas/
    change-manifest.schema.json
    agent-review.schema.json
    agent-cross-review.schema.json
    evidence-bundle.schema.json
  containers/

specs/
evals/
policy/
agents/
  rules.md

AGENTS.md
CLAUDE.md
~~~

The ci/run contract should remain small and stable:

| Command | Purpose | Expected outputs |
|---|---|---|
| ci/run detect | Identify ecosystem, toolchains, lockfiles, and supported capabilities | Machine-readable capability manifest |
| ci/run bootstrap | Install or verify the pinned toolchain and dependencies without running quality gates | Bootstrap manifest and logs |
| ci/run fast | Format check, lint, type check, compile, unit tests, changed-code coverage | JUnit, SARIF, coverage, logs |
| ci/run full | Integration, contract, e2e, and complete coverage suite | JUnit, service logs, coverage |
| ci/run security | SAST, secrets, dependency, workflow, IaC, and container scans as applicable | SARIF and CycloneDX/SPDX references |
| ci/run eval | Specification-linked deterministic and agent/product evaluations | JUnit plus eval result JSON |
| ci/run package | Reproducible build from a clean source tree | Artifact digest and metadata |
| ci/run evidence | Validate and assemble the immutable evidence bundle; release behavior is selected by profile/configuration | Signed or hashed manifest |

Exit codes remain authoritative. A renderer may transform results into a GitHub check, but it must not reinterpret a failing exit code as success.

## 7. Pull-request execution flow

~~~mermaid
sequenceDiagram
    participant G as GitHub
    participant R as Router
    participant X as Isolated runner
    participant E as Evidence store
    participant P as Policy arbiter

    G->>R: PR event and immutable head SHA
    R->>R: Classify trust and runner health
    R->>X: Run protected workflow at exact SHA
    X->>X: Spec, build, test, scan, and eval
    X->>E: Upload manifest and hashed evidence
    E-->>P: Verified evidence bundle
    P-->>G: Required check conclusion
~~~

The workflow must:

1. Resolve and record the exact base SHA, head SHA, merge SHA if used, specification hash, workflow revision, and lockfile hashes.
2. Use persist-credentials: false for checkouts that do not need to push.
3. Avoid executing a pull request's modified privileged workflow in a credential-bearing context.
4. Run the deterministic suite without Claude, Codex, release, or deployment credentials.
5. Upload evidence even on failure, while preventing logs and artifacts from containing secrets.
6. Trigger agent review only after the required deterministic preconditions pass.

A workflow_run-triggered agent workflow can obtain privileges unavailable to the original workflow. GitHub explicitly warns that using untrusted code with workflow_run can create security vulnerabilities. The agent workflow therefore uses its protected default-branch definition, consumes only validated artifacts, and does not run code or scripts from the pull request. See [Events that trigger workflows](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_run).

## 8. Agent collaboration flow

~~~mermaid
sequenceDiagram
    participant B as Broker
    participant C as Claude
    participant X as Codex
    participant P as Arbiter
    participant W as Writer worktree

    B->>C: Frozen spec, diff, and evidence
    B->>X: Same frozen inputs
    C-->>B: Sealed review A
    X-->>B: Sealed review B
    B->>C: Review B for verification
    B->>X: Review A for verification
    C-->>P: Accept, reject, or amend findings
    X-->>P: Accept, reject, or amend findings
    P->>P: Apply deterministic policy
    P->>W: Accepted findings and repair budget
    W-->>B: Patch only
~~~

### 8.1 Review rules

- Initial reviews are independent and cannot read each other's output.
- Findings use a strict JSON Schema and include an ID, severity, confidence, file and line when applicable, acceptance-criterion IDs, claim, evidence, reproducer, and suggested fix.
- Cross-review exchanges only evidence-backed conclusions, not hidden chain-of-thought.
- A deterministic failure always blocks.
- A reproduced critical or high finding blocks.
- Agreement without evidence is not sufficient to block or approve.
- A disputed high-severity finding requires a reproducer, a fresh adjudication run, or a human decision.
- The writer receives accepted findings and evidence, not an unrestricted transcript.
- The complete deterministic suite runs after every repair.
- Review is limited to two repair cycles. A no-progress loop, usage exhaustion, or unresolved high-severity dispute escalates to a human.

The supervised pilot path uses OpenAI's official [Codex plugin for Claude Code](https://github.com/openai/codex-plugin-cc), including review or adversarial-review commands. The unattended target path invokes the first-party CLIs directly because [Codex noninteractive mode](https://developers.openai.com/codex/noninteractive) and the [Claude Code CLI](https://code.claude.com/docs/en/cli-reference) support machine-readable output and bounded execution.

## 9. Evidence and memory

### 9.1 Per-run bundle

Every run produces an immutable directory or object prefix:

~~~text
runs/<run-id>/
  manifest.json
  spec/
  source/
  checks/
  evals/
  reviews/
    claude.json
    codex.json
  cross-review/
  decision.json
  patches/
  release/
~~~

The manifest contains at minimum:

- Run ID and triggering GitHub event.
- Repository, pull request, base SHA, head SHA, and tested SHA.
- Specification and acceptance-manifest hashes.
- Workflow and ci/run revisions.
- Toolchain, analyzer, action, agent-client, and schema versions.
- Trust classification and runner identity.
- Start/end timestamps, timeouts, attempt numbers, and reason for retry.
- Hash, media type, producer, and size of each evidence object.
- Agent role, provider, budget, and verdict, without storing authentication material.
- Policy decision and any human waiver identity, reason, scope, and expiration.

### 9.2 Standard formats

| Evidence | Preferred format |
|---|---|
| Tests and evaluations | JUnit XML plus structured JSON details |
| Static and security analysis | SARIF |
| Coverage | LCOV or Cobertura |
| Dependency and artifact inventory | CycloneDX or SPDX |
| Build provenance | in-toto statement following the applicable SLSA provenance schema |
| Telemetry | OpenTelemetry traces, metrics, and logs |

### 9.3 Durable project memory

Git remains the canonical durable memory for specifications, architecture decisions, policy, regression tests, and approved waivers. A local database may index finding IDs and evidence metadata, but it is not authoritative.

Only the following are promoted to long-term memory:

- Accepted architecture decisions.
- Confirmed defects and their regression tests.
- Resolved review findings and the evidence that resolved them.
- Stable project conventions.
- Explicitly approved exceptions with expiration.

Raw prompts, hidden reasoning, secrets, rejected speculative findings, and unlimited transcripts are excluded.

## 10. Credential isolation

### 10.1 Subscription authentication

OpenAI documents ChatGPT-account authentication for trusted private Codex automation, while recommending API keys for general CI. The local authentication state is password-equivalent, must not be uploaded or logged, and should not be shared concurrently. See [Codex authentication for CI/CD](https://developers.openai.com/codex/auth/ci-cd-auth) and [Codex authentication](https://developers.openai.com/codex/auth).

Claude Code supports Pro and Max login through its first-party client. Its credential state remains on the dedicated agent host. For unattended operation, Anthropic documents `claude setup-token`: a one-year OAuth token for CI pipelines and scripts, supplied through the `CLAUDE_CODE_OAUTH_TOKEN` environment variable, supported on Pro/Max/Team/Enterprise plans, and scoped to inference only. The token is printed once at generation and never persisted by the client; it is generated on the agent host itself, treated as password-equivalent, and serialized like Codex authentication state. Bare mode (`--bare`) does not read `CLAUDE_CODE_OAUTH_TOKEN` and is not used in this lane. When reauthentication is required, the agent result cannot approve; deterministic CI remains independently reported and final merge policy follows the risk-tier substitution rules in Document 04. See [Claude Code authentication](https://code.claude.com/docs/en/iam).

### 10.2 Enforcement rules

- Agent authentication files never enter GitHub Secrets, workflow artifacts, container images, caches, repository files, or hosted runners.
- The broker runs under a dedicated, non-admin OS account.
- Each provider receives a separate home/config directory with mode 0600 credential files.
- Codex work using one authentication state is serialized unless the provider documentation explicitly permits another pattern.
- The broker can contact provider endpoints but cannot contact production, the hypervisor API, the NAS, or administrative networks.
- Child commands launched by an agent are denied network access and do not inherit readable authentication files.
- Review mode mounts source and evidence read-only.
- Repair mode permits writes only to an isolated worktree and returns a patch; credential-free CI executes it.
- A workflow cannot cause the broker to push, merge, release, or deploy directly.
- Provider outage, usage limit, or expired login produces a neutral agent-unavailable status, never an automatic paid-API fallback.

OpenAI describes sandboxing, network controls, and prompt-injection considerations in [Agent approvals and security](https://developers.openai.com/codex/agent-approvals-security). Anthropic provides corresponding [Claude Code security guidance](https://code.claude.com/docs/en/security) and [secure deployment guidance](https://code.claude.com/docs/en/agent-sdk/secure-deployment).

## 11. Merge and release decision policy

The arbiter evaluates gates in order:

1. Input integrity: hashes, schema versions, exact commit, and required evidence are valid.
2. Hard deterministic gates: specification traceability, compilation, required tests, policy, and critical security checks pass.
3. Quantitative gates: coverage, performance, mutation, and scored evaluations meet both absolute and regression thresholds.
4. Agent findings: no unresolved reproduced critical/high finding remains.
5. Human controls: required code owners and environment approvals are present.
6. Release controls: artifact digest, SBOM, scan, provenance, signature, and deployment policy agree.

No weighted average can offset a hard-gate failure. Flaky-test retries are recorded separately from semantic retries; the system does not retry a semantic failure until it happens to pass.

## 12. Failure behavior

| Failure | Required behavior |
|---|---|
| No eligible local runner | Use hosted execution only when policy permits; otherwise queue with a clear reason |
| Runner disappears mid-job | Mark infrastructure failure, discard workspace, and redispatch once through the watchdog |
| Test or policy failure | Fail without infrastructure retry |
| Evidence hash mismatch | Quarantine the bundle and block agent/release stages |
| Missing required evidence | Fail closed |
| Agent login expired or usage exhausted | Mark agent lane unavailable and request human action; do not use a paid API |
| Claude and Codex disagree on high severity | Seek a reproducer or human adjudication |
| Agent repair makes no progress | Stop after the configured cycle limit |
| Object store unavailable | Preserve local spool if safe; do not approve or release without committed evidence |
| Cache corruption | Delete the affected namespace and rebuild from lockfiles |
| Signing or provenance failure | Do not publish or deploy |
| Deployment health check failure | Stop promotion and roll back to the last verified digest |
| Production rollback fails | Freeze further deployment and invoke the incident procedure |

## 13. Security properties and limitations

This architecture provides:

- Separation of code execution, model credentials, and deployment authority.
- Exact linkage between specification, commit, evidence, reviews, and release digest.
- Reproducible controller-neutral commands.
- Independent model review with bounded repair.
- Hosted fallback for deterministic jobs without copying subscription credentials.
- A path from a simple dedicated VM to disposable local runners.

It does not guarantee:

- That two agreeing models are correct.
- That containers alone safely isolate hostile workloads.
- That a home-administered builder qualifies for a high SLSA build level.
- That GitHub-hosted fallback can migrate a job already running locally.
- That subscription availability or limits are suitable for an always-on production service.

The SLSA project distinguishes useful build traceability from stronger build-platform isolation requirements; claims must match the controls actually deployed. See [SLSA build track basics](https://slsa.dev/spec/v1.2/build-track-basics).

## 14. Architecture acceptance criteria

The architecture is ready for the pilot when:

- A private test repository can run ci/run fast and ci/run security identically on a workstation and a GitHub runner.
- The local runner holds no Claude, Codex, signing, or deployment credentials.
- Fork and public PR jobs cannot select the persistent local runner.
- A run manifest binds all results to the exact source and specification hashes.
- Claude and Codex each produce schema-valid, evidence-linked reviews.
- A deterministic failure remains blocking even when both agents approve.
- A repair is tested in a credential-free environment.
- Agent loops stop at the configured maximum.
- The release workflow starts from a protected commit and promotes by immutable digest.
- Disabling or disconnecting the home server has a documented, tested fallback outcome.

## 15. Primary references

- [GitHub-hosted runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)
- [Self-hosted runners](https://docs.github.com/en/actions/reference/runners/self-hosted-runners)
- [Workflow syntax and job containers](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#jobsjob_idcontainer)
- [GitHub Actions secure use reference](https://docs.github.com/en/actions/reference/security/secure-use)
- [Actions Runner Controller](https://docs.github.com/en/actions/concepts/runners/actions-runner-controller)
- [GitHub Spec Kit](https://github.com/github/spec-kit)
- [OpenAI Codex plugin for Claude Code](https://github.com/openai/codex-plugin-cc)
- [OpenAI Codex CI/CD authentication](https://developers.openai.com/codex/auth/ci-cd-auth)
- [Anthropic Claude Code authentication](https://code.claude.com/docs/en/iam)
- [SLSA build track basics](https://slsa.dev/spec/v1.2/build-track-basics)
