# Council — Local-First Agentic CI/CD Platform

## Project charter and requirements

| Field | Value |
|---|---|
| Document ID | ACI-CHARTER-001 |
| Version | 0.1 |
| Status | Proposed baseline |
| Date | 2026-07-13 |
| Decision authority | Project owner |
| Technical owner | Platform owner |
| Related document | [Risk register and decision log](09-risk-register-and-decisions.md) |

## 1. Executive summary

This project will establish a local-first, language-agnostic CI/CD platform for specification- and evaluation-driven software development with coding agents. GitHub will remain the source-control and workflow control plane. A home server will provide the primary compute plane through isolated self-hosted runners, with an explicitly controlled GitHub-hosted fallback for eligible deterministic jobs.

The platform will combine conventional CI/CD controls—builds, tests, static analysis, dependency and secret scanning, artifacts, release promotion, software bills of materials (SBOMs), signatures, deployment approvals, and rollback—with a bounded multi-agent workflow. Claude Code and Codex will use the existing Claude Max and ChatGPT Pro subscriptions through their first-party command-line clients. Agents may implement, independently review, challenge findings, and perform a limited repair pass. Agent agreement will never override a failed deterministic gate.

The primary security boundary is architectural: untrusted repository code will not execute in the environment that stores subscription credentials. The initial release will prove the workflow on one private pilot repository before adding ephemeral runner fleets, more language adapters, and automated deployment.

## 2. Problem statement

Conventional pipelines establish whether code builds, passes tests, meets policy, and can be safely released. Coding agents add new requirements that ordinary CI/CD systems do not fully address:

- Requirements must be explicit enough for agents to implement and evaluators to verify.
- Agent output is nondeterministic and can be confidently wrong.
- Repository content may contain prompt injection or instructions designed to misuse agent tools.
- Personal subscription credentials must not be exposed to pull-request code, logs, caches, artifacts, or hosted runners.
- Two agents need a structured way to exchange evidence without creating an unbounded debate or treating consensus as proof.
- The same checks should work locally and in hosted CI without encoding the implementation only in workflow YAML.

The project will solve these problems without creating a new source-code forge or a general-purpose agent framework.

## 3. Purpose

The purpose of the platform is to make agent-authored changes at least as reviewable, reproducible, and policy-controlled as human-authored changes while using local compute and existing subscriptions wherever safely possible.

## 4. Goals

| ID | Goal |
|---|---|
| G-001 | Make a frozen specification and executable acceptance criteria the starting point for agentic changes. |
| G-002 | Provide a language-agnostic pipeline contract with replaceable ecosystem adapters. |
| G-003 | Run deterministic validation before spending agent-review capacity. |
| G-004 | Enable Claude Code and Codex to independently review the same immutable evidence and then verify one another's findings. |
| G-005 | Keep untrusted build execution separate from subscription credentials and release authority. |
| G-006 | Use the home server as the normal compute plane while preserving a controlled path to GitHub-hosted execution. |
| G-007 | Produce durable, machine-readable evidence for every decision, repair, release, and deployment. |
| G-008 | Support standard CI/CD capabilities including protected merges, reusable workflows, test matrices, caching, artifacts, security scanning, SBOMs, signing, staged deployment, and rollback. |
| G-009 | Avoid mandatory per-token API spending for the core agent workflow. |
| G-010 | Create a platform that can add languages, repositories, evaluators, and runner backends incrementally. |

## 5. Non-goals

| ID | Non-goal |
|---|---|
| NG-001 | Replacing GitHub with a self-hosted Git forge in the initial release. |
| NG-002 | Building a fully autonomous system that can merge or deploy solely because two models agree. |
| NG-003 | Providing subscription credentials to arbitrary third-party agent frameworks or hosted runners. |
| NG-004 | Guaranteeing that containers alone securely isolate hostile code. |
| NG-005 | Supporting every programming language in the MVP. The MVP establishes a stable adapter contract and proves it with the pilot ecosystem. |
| NG-006 | Adding a vector database before structured Git and run artifacts prove insufficient. |
| NG-007 | Replacing unit, integration, security, or release testing with LLM-as-judge scores. |
| NG-008 | Claiming a formal supply-chain assurance level that a home-administered builder cannot substantiate. |
| NG-009 | Automatically purchasing API capacity when subscription usage is exhausted. |
| NG-010 | Executing unrestricted agent-generated infrastructure or deployment commands. |

## 6. Guiding principles

1. **Evidence outranks opinion.** Reproducers, tests, traces, and tool output outrank agent confidence or consensus.
2. **Deterministic failures are hard failures.** An agent cannot waive a compile, test, policy, or security gate.
3. **Trust zones are explicit.** Build, agent, and release capabilities are separated by identity, runtime, network, and storage policy.
4. **Build once, promote by digest.** A reviewed artifact is promoted; production is not rebuilt from mutable source.
5. **Repository logic is portable.** Workflow YAML orchestrates a repository-owned command contract rather than containing all pipeline behavior.
6. **Memory is curated.** Only evidence-backed decisions, accepted findings, regression tests, and approved architecture records become durable memory.
7. **Automation is bounded.** Every agent task has time, turn, concurrency, permission, and repair-loop limits.
8. **Fallback is explicit.** A local failure must not silently expose credentials or create unapproved cloud cost.

## 7. Stakeholders and responsibilities

| Role | Responsibilities | Approval scope |
|---|---|---|
| Project owner | Defines outcomes, budget constraints, risk tolerance, and pilot repository | Charter, scope changes, production go-live |
| Platform owner | Designs and operates runner infrastructure, workflows, broker, observability, and recovery | Platform implementation and operational changes |
| Repository maintainer | Owns repository rules, adapters, tests, specifications, and merge decisions | Code changes and repository-specific policy |
| Security owner | Reviews trust boundaries, secrets handling, scanners, egress, signing, and incident response | Security exceptions and release-risk acceptance |
| Release owner | Owns environments, deployment identity, approvals, smoke tests, and rollback | Release and production promotion |
| Specification owner | Defines change intent and acceptance criteria | Specification readiness |
| Human reviewer | Resolves disputed or high-risk findings and approves protected changes | Merge approval within repository policy |
| Agent broker | Enforces schemas, budgets, immutable inputs, and sequencing | No independent policy-waiver authority |
| Claude Code / Codex | Implement or review under assigned role and permissions | Advisory only; no risk acceptance authority |

One person may fill multiple human roles during the pilot, but the logical responsibilities remain separate.

## 8. Scope

### 8.1 Pilot scope — Roadmap phases P0 through P4

The pilot includes:

- One private GitHub pilot repository.
- GitHub pull requests, rulesets or branch protection, required checks, and `CODEOWNERS`.
- One dedicated Ubuntu runner VM on the home server for trusted private work.
- GitHub-hosted execution only through a manual, eligible-job fallback.
- One specification system, initially [GitHub Spec Kit](https://github.com/github/spec-kit).
- A repository-owned `ci/run` contract and one complete pilot language adapter.
- Fast deterministic gates: formatting, linting, type/build validation, unit and integration tests, coverage, secrets scanning, software-composition analysis, SAST, and workflow linting.
- Native Claude Code and Codex clients authenticated through Claude Max and ChatGPT Pro.
- The official [Codex plugin for Claude Code](https://github.com/openai/codex-plugin-cc) for the supervised implementation/review loop.
- A structured Codex review of a Claude-authored or human-authored candidate, with a supervised repair and deterministic rerun.
- Hashed run manifests and schema-valid reviewer findings sufficient to evaluate the pilot.
- Human approval for high-risk changes, unresolved findings, merge, and production deployment.
- A nightly workflow scaffold and release workflow design, even where the pilot has no production target.

### 8.2 MVP scope — adds Roadmap phase P5

The MVP adds:

- A small native-client broker invoking Claude Code and Codex without exporting subscription credentials.
- Sealed blind first reviews, one cross-review round, deterministic arbitration, and a controlled repair workflow.
- One repair pass by default and a hard maximum of two repair cycles.
- An immutable evidence bundle connecting the specification, candidate, checks, reviews, repair, and decision.
- Risk-tier enforcement, agent-unavailable behavior, prompt-injection tests, and broker crash recovery.

### 8.3 Hardened v1 scope — adds Roadmap phases P6 and P7

Hardened v1 adds:

- Ephemeral runner scale sets using [Actions Runner Controller](https://docs.github.com/en/actions/concepts/runners/actions-runner-controller), or disposable VM runners using GARM/Incus where stronger isolation is required.
- Additional language adapters and operating-system/runtime matrices.
- Automated local health routing and watchdog-driven infrastructure retries.
- Local artifact/cache/object storage and OpenTelemetry-compatible observability.
- Product/agent evaluation suites using Promptfoo and trace/experiment analysis using Phoenix, when applicable.
- SBOM generation, artifact signing, provenance, staged deployment, OIDC federation, canary release, and automated rollback.
- Optional competing implementations in isolated worktrees for explicitly classified high-risk changes.

### 8.4 Out of scope until separately approved

- Public or forked pull-request code on persistent local runners.
- Autonomous production changes without protected-environment approval.
- Generic third-party OAuth relays for Claude Max or ChatGPT Pro credentials.
- A silent paid-API fallback.
- Self-hosting GitHub-equivalent source control solely for this project.

## 9. Target operating model

### 9.1 Trust zones

| Zone | Repository code execution | Credentials | Network policy | Primary output |
|---|---|---|---|---|
| Z-01 Router/control | No checkout or executable repository code | Narrow GitHub App or workflow token only | GitHub endpoints only | Runner selection and workflow dispatch |
| Z-02 Build/test | Yes, treated as untrusted until merged | No Claude, Codex, deployment, NAS, or administrative credentials | Deny by default; allow required package and source endpoints | Test, scan, coverage, and build evidence |
| Z-03 Agent review | Read-only repository/diff review by default; repair occurs in an isolated worktree | First-party Claude and Codex client credentials | Provider, GitHub, and approved documentation endpoints | Structured reviews and candidate patches |
| Z-04 Release | Reviewed source or immutable build inputs only | Short-lived deployment identity; signing policy | Artifact registry and deployment targets only | Signed release and deployment evidence |

### 9.2 Standard change flow

1. The specification author freezes a change specification and maps every acceptance criterion to a verifier.
2. A writer agent or human creates a candidate in an isolated worktree.
3. The fast deterministic pipeline evaluates the exact candidate commit.
4. If hard gates pass, Claude and Codex receive the same immutable evidence bundle in fresh review sessions.
5. Initial reviews are sealed before either reviewer sees the other report.
6. Reviewers accept, reject, or modify each other's findings with evidence.
7. The broker applies deterministic arbitration rules; a human resolves disputed high-risk findings.
8. The writer performs a bounded repair pass, and the entire deterministic suite reruns.
9. Required checks and human approvals control merge.
10. Release automation builds or selects an immutable artifact, emits supply-chain evidence, deploys through a protected environment, verifies health, and retains rollback capability.

## 10. Functional requirements

Priority uses **M** (Must), **S** (Should), **C** (Could), and **W** (Won't in current scope).

### 10.1 Specification and traceability

| ID | Pri. | Requirement |
|---|:---:|---|
| FR-SPEC-001 | M | Each agentic change shall reference a versioned specification and exact specification hash. |
| FR-SPEC-002 | M | Every normative acceptance criterion shall have a stable ID and at least one executable verifier URI such as `test://`, `contract://`, `eval://`, or `policy://`. |
| FR-SPEC-003 | M | The pipeline shall fail when required acceptance criteria are missing, unparseable, unmapped, or not executed. |
| FR-SPEC-004 | M | Specification, verifier, candidate commit, and result hashes shall be recorded in the run manifest. |
| FR-SPEC-005 | S | Protected specifications, policies, hidden tests, and waivers shall require owner approval when modified. |
| FR-SPEC-006 | S | Specifications shall capture risk classification, affected components, nonfunctional constraints, rollout, and rollback expectations. |

### 10.2 Portable CI contract and language adapters

| ID | Pri. | Requirement |
|---|:---:|---|
| FR-CI-001 | M | Repository pipeline behavior shall be exposed through a versioned `ci/run` interface callable locally and from GitHub Actions. |
| FR-CI-002 | M | The interface shall support the canonical `detect`, `bootstrap`, `fast`, `full`, `security`, `eval`, `package`, and `evidence` operations defined in Document 05, with unsupported adapter capabilities reported explicitly. |
| FR-CI-003 | M | Each language adapter shall emit normalized exit status and machine-readable evidence, including JUnit for tests and SARIF where supported. |
| FR-CI-004 | M | The pilot adapter shall implement formatter verification, linting, type or compiler checks, tests, coverage, dependency analysis, and build/package verification appropriate to its ecosystem. |
| FR-CI-005 | S | Runtime and operating-system matrices shall be declared as repository configuration, not embedded solely in opaque scripts. |
| FR-CI-006 | S | The pipeline shall support unit, integration, contract, property, fuzz, mutation, performance, and end-to-end test classes, even when some classes are scheduled rather than per-PR. |
| FR-CI-007 | M | Tool and dependency versions shall be locked or recorded so a result can be reproduced. |

### 10.3 GitHub integration and conventional CI/CD

| ID | Pri. | Requirement |
|---|:---:|---|
| FR-GH-001 | M | Pull requests shall be protected by required deterministic checks and at least the human approvals defined by repository risk policy. |
| FR-GH-002 | M | Workflow jobs shall use least-privilege permissions and pin third-party Actions to immutable commit SHAs. |
| FR-GH-003 | M | The system shall publish relevant test, scan, coverage, and build artifacts with retention and sensitivity classifications. |
| FR-GH-004 | M | Stale runs shall be cancelled through workflow concurrency controls; job and workflow timeouts shall be explicit. |
| FR-GH-005 | M | Infrastructure retries shall be distinguished from semantic test/eval retries; failing code shall not be retried until green. |
| FR-GH-006 | S | Reusable workflows or composite actions shall provide organization-wide defaults without hiding repository-owned commands. |
| FR-GH-007 | S | Dependency update automation shall be supported, with generated updates subject to the same gates as other changes. |
| FR-GH-008 | S | CI status, required action, and evidence links shall be visible on the pull request without exposing secrets or excessive logs. |
| FR-GH-009 | S | Nightly and release schedules shall be independently configurable from pull-request checks. |

### 10.4 Deterministic quality and security gates

| ID | Pri. | Requirement |
|---|:---:|---|
| FR-GATE-001 | M | Compilation/type checking, required tests, and policy checks shall be hard gates that agents cannot waive. |
| FR-GATE-002 | M | The PR pipeline shall include secret scanning, SAST, dependency vulnerability analysis, and GitHub workflow validation. |
| FR-GATE-003 | S | The full/nightly pipeline shall support mutation testing, fuzzing or sanitizers, DAST, and performance budgets where applicable. |
| FR-GATE-004 | M | A scanner finding shall include tool identity, rule, severity, location, evidence, candidate commit, and suppression state. |
| FR-GATE-005 | M | Suppressions and risk acceptances shall be versioned, scoped, time-bound, owned, and reviewable. |
| FR-GATE-006 | S | Changed-code coverage and repository-defined quality thresholds shall be enforced without allowing a global average to hide critical-component regressions. |
| FR-GATE-007 | S | Release workflows shall generate an SBOM in CycloneDX or SPDX format and scan the final artifact or image. |

### 10.5 Agent orchestration and review

| ID | Pri. | Requirement |
|---|:---:|---|
| FR-AGENT-001 | M | The system shall invoke only first-party Claude Code and Codex clients for subscription-authenticated work. |
| FR-AGENT-002 | M | Agent review shall begin only after the configured deterministic prerequisite gates pass. |
| FR-AGENT-003 | M | Claude and Codex shall receive identical, immutable review inputs and fresh task contexts for the blind review stage. |
| FR-AGENT-004 | M | Reviewer output shall validate against a versioned JSON schema containing verdict, finding ID, severity, confidence, affected acceptance IDs, location, claim, evidence, reproducer, and suggested resolution. |
| FR-AGENT-005 | M | Initial reviews shall be sealed before cross-review, and cross-review shall record `accept`, `reject`, or `modify` for every considered finding. |
| FR-AGENT-006 | M | The policy arbiter shall block deterministic failures, reproduced critical/high findings, and independently corroborated blocking findings according to repository policy. |
| FR-AGENT-007 | M | Agent consensus shall not waive deterministic failures or security policy. |
| FR-AGENT-008 | M | Every agent invocation shall have explicit model/client metadata, role, permissions, input hashes, time limit, turn limit, and output size limit. |
| FR-AGENT-009 | M | Repair loops shall default to one pass and stop after no more than two passes; unresolved blocking findings shall require a human decision. |
| FR-AGENT-010 | S | The writer role may alternate between Claude and Codex; dual independent implementations shall be reserved for high-risk or benchmarked changes. |
| FR-AGENT-011 | M | An unavailable, expired, or rate-limited subscription shall pause or skip the agent lane according to declared policy and shall not trigger unapproved API billing. |
| FR-AGENT-012 | M | Agent-generated changes shall be delivered as reviewable patches or branches and shall not bypass protected merge rules. |
| FR-AGENT-013 | M | Every repository shall have an owner-approved provider-transmission classification before source or evidence is sent through a consumer Claude Max or ChatGPT Pro account. |
| FR-AGENT-014 | M | Approved consumer-plan use shall verify provider model-improvement controls, prohibit credentials/regulated data, and record the authorization date without storing sensitive settings evidence. |

### 10.6 Memory and evidence

| ID | Pri. | Requirement |
|---|:---:|---|
| FR-MEM-001 | M | Each run shall create an immutable manifest linking specification, commit, deterministic evidence, reviews, cross-reviews, repair patches, and final decision. |
| FR-MEM-002 | M | Canonical durable knowledge shall live in versioned specifications, tests, architecture decision records, policies, and approved agent instructions. |
| FR-MEM-003 | M | Raw private reasoning or chain-of-thought shall not be requested, stored, or exchanged; agents shall share conclusions and supporting evidence. |
| FR-MEM-004 | M | Durable findings shall include provenance, source commit, owner, state, and, where applicable, expiry or supersession data. |
| FR-MEM-005 | S | Search shall begin with structured files and SQLite/PostgreSQL metadata; semantic vector retrieval shall require a demonstrated use case and separate decision. |
| FR-MEM-006 | S | A single canonical agent-rules source shall prevent conflicting `AGENTS.md` and `CLAUDE.md` instructions. |

### 10.7 Runner routing, isolation, and recovery

| ID | Pri. | Requirement |
|---|:---:|---|
| FR-RUN-001 | M | Persistent self-hosted runners shall be restricted to private, trusted repository contexts during the MVP. |
| FR-RUN-002 | M | Build/test runners shall not store Claude, Codex, deployment, signing, NAS, hypervisor, or home-network administrative credentials. |
| FR-RUN-003 | M | The agent runner shall be isolated from arbitrary PR execution and shall not expose its credential store as an artifact, cache, volume, or log. |
| FR-RUN-004 | M | The release runner shall be separately labeled and protected and shall never execute an unreviewed pull-request workflow. |
| FR-RUN-005 | M | The MVP shall offer a manual GitHub-hosted fallback for eligible deterministic jobs; fallback shall not transfer subscription credentials. |
| FR-RUN-006 | S | Automated routing shall use a signed, expiring health signal and classify failures before redispatching work. |
| FR-RUN-007 | M | Runner networks shall be segmented from home-user, NAS, management, and administrative networks, with egress restricted by job class. |
| FR-RUN-008 | M | Untrusted jobs shall not receive the host Docker socket; container membership shall not be treated as a hard security boundary. |
| FR-RUN-009 | S | The target runner fleet shall provide disposable environments per job, using pods for trusted isolation levels and VMs where hostile-code isolation is required. |
| FR-RUN-010 | S | Cache namespaces shall include repository, trust tier, toolchain, and lockfile identity; untrusted cache writes shall not feed release jobs without validation. |

### 10.8 Release and deployment

| ID | Pri. | Requirement |
|---|:---:|---|
| FR-REL-001 | S | Releases shall promote the same immutable artifact digest validated in pre-production. |
| FR-REL-002 | S | Deployment shall use short-lived federated identity such as GitHub OIDC where the target supports it. |
| FR-REL-003 | S | Release evidence shall include source and workflow identity, dependency lock state, SBOM, scan results, signature or attestation, and artifact digest. |
| FR-REL-004 | S | Production deployment shall require a protected GitHub environment and the approvals defined by change risk. |
| FR-REL-005 | S | Every production deployment shall run smoke/health checks and have a documented, tested rollback procedure. |
| FR-REL-006 | C | Eligible services may use canary or progressive delivery when the platform can evaluate health signals reliably. |

### 10.9 Evaluation and observability

| ID | Pri. | Requirement |
|---|:---:|---|
| FR-EVAL-001 | M | Agent and product evaluations shall distinguish deterministic assertions from scored or model-judged assertions. |
| FR-EVAL-002 | M | Critical deterministic assertions shall not be averaged into a composite score. |
| FR-EVAL-003 | S | Nondeterministic evals shall define sample count, absolute floor, regression threshold, and infrastructure retry policy. |
| FR-EVAL-004 | S | The project shall maintain a repository-specific regression corpus built from accepted bugs, escaped defects, and representative tasks. |
| FR-EVAL-005 | S | LLM judges shall be calibrated against human-labeled examples before becoming blocking gates. |
| FR-EVAL-006 | M | Pipeline and agent telemetry shall record queue duration, runtime, outcome, retry reason, tool versions, and redacted error class. |
| FR-EVAL-007 | S | Agent telemetry shall record task success, repair success, finding confirmation, false-positive rate, latency, turn count, and available subscription usage indicators without storing credentials. |

## 11. Nonfunctional requirements

| ID | Category | Requirement / target |
|---|---|---|
| NFR-SEC-001 | Security | No subscription, deployment, signing, GitHub administrative, NAS, or hypervisor credential may be available to untrusted repository execution. |
| NFR-SEC-002 | Security | All credentials shall be encrypted at rest where supported, redacted from output, minimally scoped, and periodically revalidated. |
| NFR-SEC-003 | Security | Network access shall be deny-by-default per trust zone, with documented and reviewed allowlists. |
| NFR-SEC-004 | Security | High-risk prompt-injection indicators, unexpected tool requests, or attempts to access protected paths shall terminate or quarantine the agent run. |
| NFR-REL-001 | Reliability | Once the pilot is stable, at least 95% of eligible PR runs shall finish without an infrastructure-caused failure over a rolling 30-day window. |
| NFR-REL-002 | Reliability | A home-runner outage shall not block deterministic validation for more than one business hour when the manual hosted fallback is permitted and intentionally invoked. |
| NFR-REL-003 | Reliability | Release rollback shall be executable without rebuilding source and shall meet the service-specific recovery objective. |
| NFR-PERF-001 | Performance | For the pilot repository, the fast deterministic path shall target p50 at or below 10 minutes and p95 at or below 20 minutes after the first 30-run baseline. |
| NFR-PERF-002 | Performance | When local capacity is healthy, queue delay shall target p95 at or below 5 minutes; concurrency shall be bounded to protect the server. |
| NFR-PERF-003 | Performance | Agent review shall have a 30-minute default wall-clock limit per reviewer unless repository policy sets a lower value. |
| NFR-REP-001 | Reproducibility | The same commit, specification, lockfiles, tool versions, and declared environment shall reproduce the same deterministic verdict, excluding documented nondeterministic tests. |
| NFR-PORT-001 | Portability | Core checks shall run through `ci/run` without requiring GitHub Actions-specific environment variables. |
| NFR-MNT-001 | Maintainability | A new language adapter shall conform to the common evidence contract and shall not require changes to agent-review schemas. |
| NFR-MNT-002 | Maintainability | Policy, schemas, workflows, and adapters shall be versioned and covered by validation tests. |
| NFR-OBS-001 | Observability | Every run shall have one correlation ID across GitHub jobs, local runners, agent invocations, evidence, and release records. |
| NFR-AUD-001 | Auditability | Blocking decisions and waivers shall be attributable to a policy version or named human approver and retained according to the evidence policy. |
| NFR-PRV-001 | Privacy | Source, prompts, diffs, and findings shall remain local or within the explicitly selected first-party provider/GitHub boundary; no additional telemetry export is allowed by default. |
| NFR-COST-001 | Cost | Core agent execution shall use Claude Max and ChatGPT Pro without automatic per-token API billing. Any paid fallback requires a separate approved decision and explicit ceiling. |
| NFR-COST-002 | Cost | GitHub-hosted fallback shall be opt-in in the MVP and expose that it may consume GitHub-hosted runner allowance or incur charges depending on the account plan. |
| NFR-UX-001 | Usability | A maintainer shall be able to reproduce a failed check locally with one documented command or receive an explicit explanation when the check requires protected infrastructure. |

## 12. Assumptions

| ID | Assumption | Validation point |
|---|---|---|
| ASM-001 | The pilot repository is private and under the project owner's administrative control. | Kickoff |
| ASM-002 | The home server can host at least one dedicated Linux VM with CPU, memory, and storage appropriate to the pilot. | Infrastructure discovery |
| ASM-003 | Reliable outbound HTTPS is available; no inbound public port is required for a GitHub self-hosted runner. | Runner installation |
| ASM-004 | Claude Max and ChatGPT Pro remain active and permit use through the first-party Claude Code and Codex clients, including the documented `claude setup-token` path for unattended Claude use. | P4 authentication qualification (P4-09/PAC-018) before agent-lane activation and periodically thereafter |
| ASM-005 | The project owner accepts that subscription limits, reauthentication, and provider-policy changes may temporarily pause the agent lane. | Charter approval |
| ASM-006 | GitHub remains the canonical source and pull-request control plane. | Charter approval |
| ASM-007 | The pilot has or can create meaningful automated tests and executable acceptance verifiers. | Pilot selection |
| ASM-008 | Human review is available for disputed high-risk findings and protected releases. | Operating-readiness review |

## 13. Constraints

| ID | Constraint |
|---|---|
| CON-001 | No automatic paid Claude or OpenAI API fallback is authorized. |
| CON-002 | ChatGPT Pro and Claude Max authentication material shall remain inside dedicated, trusted environments running the first-party clients. |
| CON-003 | GitHub-hosted workers shall not receive subscription credential files. |
| CON-004 | Public/fork code shall not run on a persistent credential-bearing or home-network-trusted runner. |
| CON-005 | Exact home server hardware, hypervisor, network segmentation, and deployment targets must be inventoried before final infrastructure sizing. |
| CON-006 | The MVP shall not depend on a large agent framework, vector database, or duplicate Git forge. |
| CON-007 | The project shall distinguish container reproducibility from security isolation; hostile-code isolation requires a stronger boundary. |
| CON-008 | Provider terms and authentication behavior may change and shall be treated as external dependencies. |

## 14. Success measures

| ID | Measure | MVP target | Hardened target |
|---|---|---:|---:|
| KPI-001 | Normative acceptance criteria mapped to executed verifiers | 100% | 100% |
| KPI-002 | Required deterministic checks represented in normalized evidence | 100% | 100% |
| KPI-003 | Runs exposing protected credentials to build logs/artifacts | 0 | 0 |
| KPI-004 | Agent reviews passing JSON-schema validation | At least 95% during pilot; invalid output blocks/requests one schema repair | At least 99% |
| KPI-005 | Blocking agent findings with a reproducer or independently verified evidence | At least 90% | At least 95% |
| KPI-006 | Repair loops exceeding the configured maximum | 0 | 0 |
| KPI-007 | Eligible deterministic PR runs without infrastructure-caused failure | Baseline collected | At least 95% rolling 30 days |
| KPI-008 | Fast deterministic pipeline duration | Baseline first 30 runs | p50 ≤10 min; p95 ≤20 min for pilot |
| KPI-009 | Escaped defects represented by a new regression test/eval and linked finding | 100% after process adoption | 100% |
| KPI-010 | Production releases with SBOM, digest, scan evidence, and rollback record | Not required unless pilot deploys | 100% |
| KPI-011 | Disputed high-risk findings silently resolved by model consensus | 0 | 0 |
| KPI-012 | Unapproved API or hosted-compute spend | $0 | $0 |

## 15. Project acceptance criteria

| ID | Acceptance criterion | Evidence |
|---|---|---|
| PAC-001 | A pilot pull request references a frozen specification whose acceptance IDs all resolve to executed verifiers. | Run manifest and traceability report; covers FR-SPEC-001 through FR-SPEC-004 |
| PAC-002 | The same fast-check command runs successfully on a developer machine and in GitHub Actions. | Command transcript, workflow run, and normalized evidence; covers FR-CI-001 through FR-CI-004 |
| PAC-003 | A deliberately failing test and a deliberately introduced high-confidence secret both block the pull request. | Negative test workflow run; covers FR-GATE-001 and FR-GATE-002 |
| PAC-004 | Claude and Codex independently review the same immutable candidate and produce schema-valid, sealed reports before cross-review. | Manifest hashes, timestamps, and review JSON; covers FR-AGENT-001 through FR-AGENT-005 |
| PAC-005 | A deterministic failure remains blocking even when both agents recommend approval. | Policy-engine test; covers FR-AGENT-006 and FR-AGENT-007 |
| PAC-006 | A repair is applied in an isolated worktree, the deterministic suite reruns, and a second review addresses only the delta and unresolved finding IDs. | Patch, rerun evidence, and final decision record; covers FR-AGENT-008 and FR-AGENT-012 |
| PAC-007 | The repair workflow stops at its configured cycle limit and requests human adjudication. | Loop-limit integration test; covers FR-AGENT-009 |
| PAC-008 | A red-team workflow attempting to read Claude/Codex credential paths from build execution cannot access them and produces no credential-bearing artifact or log. | Isolation test and redacted log review; covers FR-RUN-002 and FR-RUN-003 |
| PAC-009 | The local runner can be intentionally disabled and an eligible deterministic job can be manually rerun on GitHub-hosted infrastructure without transferring subscription credentials. | Recovery exercise; covers FR-RUN-005 and NFR-REL-002 |
| PAC-010 | A public/fork simulation is prevented from selecting the persistent local or agent runner. | Workflow policy test; covers FR-RUN-001 and CON-004 |
| PAC-011 | An immutable run record links specification, commit, tool versions, checks, both reviews, cross-review, patch, and final decision using one correlation ID. | Evidence-bundle inspection; covers FR-MEM-001 through FR-MEM-004 and NFR-OBS-001 |
| PAC-012 | Workflow validation confirms least privilege, immutable Action pins, explicit timeouts, and concurrency cancellation. | actionlint/zizmor/policy results; covers FR-GH-002 and FR-GH-004 |
| PAC-013 | No agent-unavailable path invokes a paid API; the workflow applies the normative risk-tier merge policy and reports the explicit agent state. | Expired-login/rate-limit simulation; covers FR-AGENT-011 and NFR-COST-001 |
| PAC-014 | A maintainer can reproduce each pilot deterministic failure locally using documented `ci/run` commands. | Runbook exercise; covers NFR-PORT-001 and NFR-UX-001 |
| PAC-015 | If release/CD is enabled, the workflow promotes one immutable digest through a protected environment, produces SBOM/scan evidence, runs smoke checks, and completes a rollback exercise. | Release and rollback records; covers FR-REL-001 through FR-REL-005 |
| PAC-016 | In a supervised session, Claude or a human supplies a candidate, Codex returns a schema-valid independent review through the official plugin, a repair is bounded, and deterministic checks rerun without exposing credentials. | Supervised pilot transcript, review JSON, redacted logs, and rerun evidence; covers the Pilot subset of FR-AGENT-001, FR-AGENT-002, FR-AGENT-004, FR-AGENT-008, and FR-AGENT-012 |
| PAC-017 | A confidential/regulated fixture is denied entry to the consumer-plan agent lane, while an approved private-personal fixture runs only after both providers' data-control verification is recorded. | Classification-policy integration test and redacted authorization record; covers FR-AGENT-013 and FR-AGENT-014 |
| PAC-018 | The unattended Claude authentication path is qualified: a `claude setup-token` OAuth token generated on the agent host completes the configured number of unattended structured-output runs, and expiry, invalid-token, usage-exhaustion, timeout, malformed-output, and redaction cases fail closed without exposing the token. | Authentication qualification record (documentation URL, plan, client version, auth mode, verification and next-review dates); covers FR-AGENT-001, FR-AGENT-011, and ASM-004 |

## 16. Delivery phases and exit gates

| Phase | Deliverables | Exit gate |
|---|---|---|
| P0 — Discovery and threat model | Server/hypervisor inventory; repository classification; network diagram; pilot selection; data classification; finalized risk owners | Assumptions ASM-001 through ASM-008 validated or converted to risks |
| P1 — Specification and local quality loop | Spec Kit setup; acceptance manifest; canonical `ci/run`; pilot adapter; normalized evidence | PAC-001, PAC-002, and PAC-014 pass locally |
| P2 — GitHub deterministic CI | PR workflow; rulesets; required checks; CODEOWNERS; artifacts; scanner baseline; retry/concurrency policy | PAC-003 and PAC-012 pass; no unowned blocking scanner findings |
| P3 — Local execution plane | Dedicated Ubuntu VM; labels; network restrictions; cache policy; manual hosted fallback; recovery runbook | PAC-008, PAC-009, and PAC-010 pass |
| P4 — Supervised two-model workflow | Provider-transmission classification; native client login; unattended-auth qualification spike; official Codex-for-Claude plugin; structured review; supervised repair; usage limits | PAC-016, PAC-017, and PAC-018 pass |
| P5 — Automated agent broker | Schemas; blind/cross-review; policy arbiter; isolated repair; MVP evidence bundle; loop limits | PAC-004, PAC-005, PAC-006, PAC-007, PAC-011, and PAC-013 pass |
| P6 — Ephemeral scaling and observability | ARC or VM runner decision; disposable fleet; durable local evidence; health routing; telemetry; chaos exercises | Reliability and performance targets measured for 30 runs and Roadmap G6 passes |
| P7 — Release and deployment hardening | Registry; SBOM; signatures; provenance; OIDC; protected environments; staging; smoke/canary; rollback | PAC-015 and Roadmap G7 pass where a deployment target exists |

### 16.1 Requirements traceability summary

Document 07 owns the work-package IDs. Before each gate closes, the project MUST export item-level traceability from every applicable Must requirement to an implemented work item and an evidence reference. This summary defines the initial grouping; it does not claim that one test proves every requirement in a range.

| Requirement group | Primary roadmap phases | Principal project acceptance/evidence |
|---|---|---|
| FR-SPEC-001–006 | P1 | PAC-001 plus spec/schema/code-owner review |
| FR-CI-001–007 | P1–P2 | PAC-002, PAC-003, PAC-014 plus adapter contract tests |
| FR-GH-001–009 | P2–P3 | PAC-009, PAC-010, PAC-012 plus protected-branch inspection |
| FR-GATE-001–007 | P2, P6–P7 | PAC-003, PAC-005, PAC-015 plus nightly/release evidence |
| FR-AGENT-001–014 | P4–P5 | PAC-004–007, PAC-013, PAC-016, PAC-017, PAC-018; each result records the exact requirement IDs actually exercised |
| FR-MEM-001–006 | P1, P5–P6 | PAC-011 plus retention, restoration, and instruction-source checks |
| FR-RUN-001–010 | P3, P6 | PAC-008–010 plus runner lifecycle, routing, and chaos tests |
| FR-REL-001–006 | P7 | PAC-015 plus signature, provenance, OIDC, promotion, and rollback records |
| FR-EVAL-001–007 | P1–P2, P6 | PAC-001, PAC-003, PAC-005 plus baseline, holdout, variance, and flake evidence |
| NFR-SEC-001–004 | P2–P5 | PAC-003, PAC-008, PAC-010, PAC-012, PAC-017 |
| NFR-REL-001–003 | P3, P6–P7 | PAC-009, PAC-015 plus recovery drills |
| NFR-PERF-001–003 | P2, P6 | Thirty-run baseline and Roadmap G6 capacity/latency report |
| NFR-REP-001, NFR-PORT-001 | P1–P2 | PAC-002 and PAC-014 |
| NFR-MNT-001–002 | P1–P2 | Adapter contract, version inventory, and maintainer review |
| NFR-OBS-001, NFR-AUD-001 | P5–P6 | PAC-011 plus audit/recovery inspection |
| NFR-PRV-001 | P4–P6 | PAC-017 plus retention/redaction tests |
| NFR-COST-001–002 | P3–P6 | PAC-009, PAC-013 plus hosted/subscription usage report |
| NFR-UX-001 | P1–P2 | PAC-014 and pilot maintainer exercise |

If a requirement is not applicable to the pilot, its row is explicitly marked `deferred` with an owner and target gate; it is never counted as passed merely because the group has other evidence.

## 17. Definition of done for an agentic change

An agentic change is complete only when:

- Its specification and acceptance criteria are frozen and traceable.
- The exact commit has passed all applicable hard deterministic gates.
- Required blind and cross-reviews are schema-valid and linked to the same commit/spec hashes.
- Accepted blocking findings are repaired or explicitly adjudicated by an authorized human.
- Repair cycles did not exceed the configured limit.
- Generated patches are human-reviewable and protected merge rules remain in force.
- Evidence is stored with a retention/sensitivity classification.
- Deployment and rollback material are updated when behavior or operations change.

## 18. External dependencies and source anchors

- GitHub documents that hosted jobs normally receive fresh managed runners, while self-hosted jobs run on machines the user manages: [GitHub-hosted runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners) and [self-hosted runners](https://docs.github.com/en/actions/reference/runners/self-hosted-runners).
- GitHub warns that persistent self-hosted runners do not provide the clean-environment guarantee of GitHub-hosted runners and should be carefully isolated from untrusted changes: [Secure use reference](https://docs.github.com/en/actions/reference/security/secure-use).
- GitHub's supported Kubernetes autoscaling path is [Actions Runner Controller](https://docs.github.com/en/actions/concepts/runners/actions-runner-controller).
- OpenAI documents ChatGPT-authenticated Codex on trusted private infrastructure, with the authentication file treated as a password-like secret: [Codex CI/CD authentication](https://developers.openai.com/codex/auth/ci-cd-auth) and [Codex authentication](https://developers.openai.com/codex/auth).
- Codex supports schema-constrained, noninteractive execution suitable for automation: [Codex noninteractive mode](https://developers.openai.com/codex/noninteractive).
- Anthropic documents Claude Code authentication and noninteractive/structured CLI operation, including `claude setup-token` — a one-year OAuth token for CI pipelines and scripts supplied through `CLAUDE_CODE_OAUTH_TOKEN`, supported on Pro/Max/Team/Enterprise plans and scoped to inference only: [Claude Code identity and access](https://code.claude.com/docs/en/iam), [common workflows](https://code.claude.com/docs/en/common-workflows), and [CLI reference](https://code.claude.com/docs/en/cli-reference).
- The official [Codex plugin for Claude Code](https://github.com/openai/codex-plugin-cc) supplies the initial supervised delegation and review path.
- [GitHub Spec Kit](https://github.com/github/spec-kit) supplies the initial specification workflow; [OpenSpec](https://github.com/Fission-AI/OpenSpec) remains an alternative if brownfield or cross-repository change management dominates.
- GitHub Agentic Workflows supplies useful safe-output and memory patterns, but its current [authentication reference](https://github.github.com/gh-aw/reference/auth/) requires API credentials for Claude and Codex and therefore does not satisfy the no-automatic-API-spend constraint.
- Release assurance claims shall follow the actual builder threat model rather than overstating a home builder's guarantees: [SLSA build track](https://slsa.dev/spec/v1.2/build-track-basics).

## 19. Governance and change control

- Requirement IDs are stable. A removed requirement is marked retired rather than reused.
- Material changes to trust boundaries, subscription authentication, automatic fallback, agent merge authority, release identity, or evidence retention require a new or superseding architecture decision in [09-risk-register-and-decisions.md](09-risk-register-and-decisions.md).
- Security exceptions require a named owner, rationale, compensating control, expiry date, and explicit approval.
- The charter shall be reviewed at the end of the pilot, before enabling public repositories, before enabling automatic hosted fallback, and before the first production deployment.
- Provider documentation and terms shall be rechecked before unattended subscription-authenticated automation is enabled or materially expanded.
