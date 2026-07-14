# Council — Local-First Agentic CI/CD Platform

## Risk register and architecture decisions

| Field | Value |
|---|---|
| Document ID | ACI-RISK-009 |
| Version | 0.1 |
| Status | Proposed baseline |
| Date | 2026-07-13 |
| Risk authority | Project owner |
| Related document | [Project charter and requirements](01-project-charter-and-requirements.md) |

## 1. Purpose

This document records threats, operational risks, mitigations, ownership, contingency actions, and foundational architecture decisions for the local-first agentic CI/CD platform. It is both a working risk register and a lightweight architecture decision record (ADR) index.

Risks are evaluated against the target system described in the project charter: GitHub as control plane; local runners as primary compute; Claude Code and Codex through Claude Max and ChatGPT Pro; deterministic gates as the merge authority; and separate build, agent, and release trust zones.

## 2. Risk method

### 2.1 Scoring

Likelihood and impact use a five-point scale.

| Score | Likelihood | Impact |
|---:|---|---|
| 1 | Rare; exceptional conditions | Negligible; no material delay or data/security effect |
| 2 | Unlikely; may occur during platform lifetime | Minor; localized rework or short interruption |
| 3 | Possible; credible during normal operation | Moderate; blocked delivery, contained security incident, or material rework |
| 4 | Likely; expected without active control | Major; extended outage, credential incident, release defect, or substantial cost |
| 5 | Almost certain or repeatedly observed | Critical; broad compromise, destructive production effect, or loss of trust |

Inherent score is `likelihood × impact` before planned controls. Residual score is the expected score after controls operate.

| Score | Rating | Required treatment |
|---:|---|---|
| 1–4 | Low | Accept and monitor |
| 5–9 | Medium | Named owner and planned controls |
| 10–16 | High | Mitigate before the affected capability becomes operational |
| 17–25 | Critical | Capability cannot proceed without project-owner and security-owner acceptance |

### 2.2 Status values

- **Open:** treatment is not complete.
- **Mitigating:** control implementation is in progress.
- **Accepted:** authorized owner accepts the residual risk for a defined period.
- **Monitoring:** controls are active and indicators are watched.
- **Closed:** the threat is removed or no longer applicable.

### 2.3 Ownership rules

- Risk owners are accountable for treatment and escalation, even when another role implements the control.
- Critical residual risk requires project-owner and security-owner approval.
- High residual risk requires the project owner or affected release owner to approve before production use.
- Canonical roles are Project owner, Platform owner, Security owner, Repository maintainer, Specification owner, Evaluation owner, and Release owner. One person may fill several roles during the pilot, but each decision records the logical role used.
- No model or automated agent may accept risk.

## 3. Summary risk register

| ID | Risk | Inherent L/I/Score | Planned response | Residual L/I/Score | Owner | Status |
|---|---|---:|---|---:|---|---|
| R-001 | PR code obtains Claude or Codex credentials | 4/5/20 | Separate credential-free build and credential-bearing agent zones; red-team isolation tests | 1/5/5 | Security owner | Open |
| R-002 | Persistent self-hosted runner is compromised across jobs | 4/5/20 | Private/trusted MVP use; rebuildable VM; ephemeral runner target; no release credentials | 2/4/8 | Platform owner | Mitigating |
| R-003 | Docker socket or privileged container enables host escape | 4/5/20 | No host socket for untrusted jobs; VM boundary for hostile code; scan runner config | 1/5/5 | Platform owner | Open |
| R-004 | Prompt injection manipulates reviewers or tools | 5/4/20 | Treat repo as untrusted data; read-only review; tool allowlists; immutable evidence; terminate protected-path attempts | 3/3/9 | Security owner | Open |
| R-005 | Subscription login or setup-token expires during automation | 4/3/12 | Auth health check; scheduled setup-token rotation before one-year expiry; explicit paused state; documented first-party re-login/regeneration; no secret copying | 3/2/6 | Platform owner | Open |
| R-006 | Claude Max or ChatGPT Pro usage limits stop work | 4/3/12 | Deterministic gates first; bounded reviews; queues/concurrency; human continuation; no automatic API spend | 3/2/6 | Project owner | Open |
| R-007 | Provider terms or authentication behavior changes | 3/4/12 | Use first-party clients; quarterly/event-driven review; feature flag agent lane; retain human workflow | 2/3/6 | Project owner | Monitoring |
| R-008 | Agents agree on an incorrect conclusion | 4/4/16 | Blind reviews; evidence/reproducers; deterministic arbiter; human adjudication for disputed high risk | 2/4/8 | Repository maintainer | Open |
| R-009 | Agent review/repair loops consume excessive usage | 4/3/12 | One pass default; two-pass hard cap; time/turn/token-output limits; no-progress detection | 1/3/3 | Platform owner | Open |
| R-010 | Nondeterministic or flaky evals create false results | 4/4/16 | Separate hard vs scored assertions; fixed seeds where valid; sample policy; quarantine; no retry-until-green | 2/3/6 | Evaluation owner | Open |
| R-011 | Weak specifications or verifiers are gamed | 4/4/16 | Stable acceptance IDs; executable traceability; hidden/owner-controlled tests; negative and property tests | 2/4/8 | Specification owner | Open |
| R-012 | Agent edits hidden tests, policy, or acceptance thresholds | 4/5/20 | Protected paths/CODEOWNERS; default-branch workflow; separate test assets where necessary; diff policy | 1/5/5 | Repository maintainer | Open |
| R-013 | Shared memory becomes poisoned, stale, or contradictory | 4/4/16 | Structured provenance; append-only run data; curated promotion; canonical rules; expiry/supersession | 2/3/6 | Platform owner | Open |
| R-014 | Logs, prompts, caches, or artifacts leak secrets/source | 4/5/20 | Classification; redaction; minimal evidence; retention; no auth files; secret scanning of output | 2/4/8 | Security owner | Open |
| R-015 | Home power, network, hardware, or hypervisor outage blocks CI | 4/3/12 | Manual hosted fallback; recovery runbook; monitoring; backup runner config; UPS considered | 2/3/6 | Platform owner | Open |
| R-016 | Fallback causes unexpected GitHub-hosted cost | 3/3/9 | Manual opt-in MVP; visible destination; concurrency cap; usage alert; no silent redispatch | 1/3/3 | Project owner | Open |
| R-017 | Hosted fallback receives personal subscription credentials | 3/5/15 | Eligibility policy; separate workflows/labels; secret absence test; fail closed | 1/5/5 | Security owner | Open |
| R-018 | Third-party Action or dependency supply chain is compromised | 4/5/20 | Pin full SHAs; minimize Actions; dependency/SBOM scans; trusted registries; update review | 2/4/8 | Platform owner | Open |
| R-019 | Cache poisoning influences trusted build or release | 3/5/15 | Namespaces by repo/trust/lock/toolchain; read-only release caches; digest verification; purge capability | 1/4/4 | Platform owner | Open |
| R-020 | Home builder provenance is overstated | 3/4/12 | Document threat model; make accurate assurance claim; use stronger hosted/isolated release builder when required | 1/4/4 | Release owner | Open |
| R-021 | Toolchain/scanner update breaks or changes results | 4/3/12 | Pin versions; staged update workflow; baseline diff; rollback; documented exceptions | 2/2/4 | Repository maintainer | Open |
| R-022 | Matrix growth makes PR pipelines too slow or expensive | 4/3/12 | Fast representative PR matrix; full nightly/release matrix; change-aware selection; bounded concurrency | 2/2/4 | Evaluation owner | Open |
| R-023 | Language adapters produce inconsistent semantics/evidence | 4/3/12 | Adapter contract tests; normalized schemas; reference adapter; certification checklist | 2/2/4 | Platform owner | Open |
| R-024 | Queue saturation starves urgent work or overloads server | 4/3/12 | Concurrency limits; priority classes; cancel stale; scale-to-zero pool later; capacity dashboards | 2/2/4 | Platform owner | Open |
| R-025 | Platform complexity prevents adoption | 4/4/16 | Thin MVP; one pilot; one command; progressive gates; feedback; defer vector DB/swarm/forge | 2/3/6 | Project owner | Open |
| R-026 | Security scanners create unmanageable false positives | 4/3/12 | Baseline existing debt; tune rules; require evidence; time-bound waivers; measure precision | 2/2/4 | Security owner | Open |
| R-027 | Signing, attestation, or OIDC policy is misconfigured | 3/5/15 | Separate release identity; audience/claim restrictions; protected environment; verification and rollback exercises | 1/5/5 | Release owner | Open |
| R-028 | Compromised runner moves laterally into the home LAN/NAS | 4/5/20 | Dedicated VLAN/bridge; firewall deny; no shared mounts; non-admin hypervisor identity; periodic reachability test | 1/5/5 | Security owner | Open |
| R-029 | Agent or workflow bypasses merge/deployment approval | 3/5/15 | Rulesets; no admin token; protected environments; separate comment/write job; audit alerts | 1/5/5 | Repository maintainer | Open |
| R-030 | Evidence retention grows without bound or violates privacy | 4/3/12 | Classification, retention schedule, quotas, lifecycle deletion, minimal prompts, legal/project review | 2/2/4 | Project owner | Open |
| R-031 | Router health signal is stale or forged | 3/4/12 | Manual fallback first; signed expiring heartbeat; freshness check; fail to explicit queued/manual state | 1/3/3 | Platform owner | Open |
| R-032 | Workflow from a PR executes with privileged default-branch context | 3/5/15 | `workflow_run`/privileged workflow from protected default branch; exact SHA checkout; no PR-controlled scripts in write job | 1/5/5 | Security owner | Open |
| R-033 | Pull-request checkout leaks or reuses GitHub write credentials | 3/4/12 | Read-only token; `persist-credentials: false`; separate narrowly scoped output publisher | 1/4/4 | Platform owner | Open |
| R-034 | Agent-generated patch introduces excessive or unrelated change | 4/3/12 | Diff/size budgets; protected paths; spec-to-diff review; human approval; split oversized changes | 2/2/4 | Repository maintainer | Open |
| R-035 | Model or client version changes invalidate evaluation baselines | 4/3/12 | Record version metadata; canary new versions; parallel benchmark; threshold reapproval | 2/2/4 | Evaluation owner | Open |
| R-036 | Human reviewers over-trust polished agent evidence | 4/4/16 | Reproducer-first UI; show deterministic evidence; training; explicit uncertainty; audit sampled approvals | 2/3/6 | Project owner | Open |
| R-037 | Consumer-plan agent use transmits confidential, regulated, or unauthorized source/evidence | 4/5/20 | Mandatory transmission classification; data-owner approval; provider-control verification; secret/regulated-data denial; approved alternative for restricted repositories | 1/5/5 | Security owner | Open |
| R-038 | Workstation colocation: the daily-driver machine also hosts the runner VM and agent broker, weakening trust-zone separation (Phase 0 finding F-01) | 4/4/16 | VirtualBox VM boundary for the runner; dedicated OS accounts for agent clients; host-only/NAT VM networking; ACT-004 reachability tests before agent credentials; revisit if a separate server is added | 2/3/6 | Platform owner | Open |
| R-039 | Disk exhaustion on the pilot host blocks runner VM, caches, and evidence storage (Phase 0 finding F-02) | 4/3/12 | Free/add ~150 GB before P3; cache and evidence quotas; retention lifecycle from day one | 2/2/4 | Platform owner | Open |
| R-040 | VirtualBox dependency on Windows 11 Home (no Hyper-V): community update cadence and manual VM lifecycle | 3/3/9 | Pin VirtualBox version; scripted VM template; staged updates; revisit on host OS upgrade | 2/2/4 | Platform owner | Open |
| R-041 | GitHub Free cannot enforce branch protection/required checks on the private pilot repo; a direct push or unchecked merge to main is technically possible (Phase 0 finding F-03) | 3/4/12 | Owner decision: GitHub Pro upgrade or documented process discipline for the pilot; must be resolved before unattended agent work (P5); required-check workflows still run and report either way | 2/3/6 | Project owner | Open |

## 4. Detailed treatment and monitoring plan

### 4.1 Credential and execution isolation

#### R-001 — PR code obtains Claude or Codex credentials

- **Threat scenario:** A malicious or compromised pull request reads client authentication files, environment variables, process memory, mounted volumes, or broker IPC and exfiltrates subscription credentials.
- **Preventive controls:** Separate VMs or equivalent strong boundaries; no credential mounts in build jobs; read-only agent review; egress restrictions; first-party client stores only; explicit protected-path deny rules.
- **Detective controls:** Canary credential-path tests, outbound DNS/HTTP telemetry, secret scanning of logs and artifacts, unexpected file-access alerts.
- **Contingency:** Stop agent lane, revoke/re-login both clients, quarantine runner image and artifacts, rotate any affected GitHub credentials, perform incident review.
- **Indicators:** Protected-path access attempts, redaction events, unexpected provider-token use, credential file checksum/permission change.

#### R-002/R-003/R-028 — Runner persistence, container escape, and LAN movement

- **Threat scenario:** Code persists across jobs, uses a container runtime to gain host control, or accesses the NAS, hypervisor, workstation, or home services.
- **Preventive controls:** Rebuildable dedicated VM during MVP; no Docker socket; rootless/container isolation only as defense in depth; separate VLAN or bridge; firewall deny rules; no home mounts or SSH keys; disposable VM target for hostile code.
- **Detective controls:** Runner image drift checks, post-job cleanliness validation, network reachability tests, EDR/audit logs where practical.
- **Contingency:** Disable runner registration, isolate VLAN, destroy/recreate VM from trusted image, invalidate caches, inspect GitHub audit log.
- **Exit criterion:** PAC-008 and PAC-010 in the charter pass before unattended local execution expands.

#### R-014/R-030 — Evidence leakage and retention

- **Threat scenario:** Prompts or reports include private source, secrets, credentials, personal information, or excessively retained operational data.
- **Controls:** Evidence classification (`public`, `internal`, `sensitive`, `secret-prohibited`); default redaction; schema allowlists; maximum output sizes; retention periods by class; encrypted local object storage; no raw chain-of-thought; log access controls.
- **Contingency:** Delete exposed artifacts where possible, rotate affected secrets, suspend publisher job, create regression redaction tests.
- **Indicators:** Secret-scanner hits, artifact access anomalies, quota trends, unclassified evidence percentage.

### 4.2 Agent integrity and evaluation

#### R-004 — Prompt injection

- **Threat scenario:** Source comments, issues, documentation, generated files, or dependency metadata direct an agent to ignore policy, access credentials, or send data externally.
- **Controls:** Label repository text as untrusted; constrain tools and egress; read-only review sessions; present a curated evidence bundle rather than an unrestricted workspace; ignore repository-authored overrides of platform policy; terminate protected-path/network attempts.
- **Contingency:** Quarantine the run and evidence, inspect the injection source, update detection/regression corpus, require human review.
- **Indicators:** Requests to reveal system instructions or credentials, unexpected network/tool use, attempts to modify policy/test paths.

#### R-008/R-036 — False consensus and automation bias

- **Threat scenario:** Both models repeat the same flawed assumption, and a human accepts polished but unsupported findings.
- **Controls:** Blind first pass; different reviewer roles; structured evidence/reproducer fields; deterministic gates; human review of disputed high-severity findings; sampled audit of approved changes; agents cannot waive policy.
- **Contingency:** Reopen affected change, add a regression verifier, update reviewer rubric and human checklist.
- **Indicators:** Escaped defects both agents approved, finding confirmation rate, human override rate, repeated shared misconception.

#### R-009 — Unbounded agent loop

- **Controls:** One repair pass by default, absolute two-pass maximum, per-review time/turn/output limits, cancellation, no-progress detection, and human escalation.
- **Contingency:** Cancel current run; retain partial evidence; let a human choose repair, defer, or abandon.
- **Indicators:** Average turns, wall time, repeat finding IDs, subscription interruption frequency.

#### R-010/R-035 — Eval nondeterminism and version drift

- **Controls:** Separate programmatic and judged assertions; version all datasets/prompts/rubrics; preserve client/model metadata; set sample count and floors before execution; distinguish infrastructure retry from semantic sampling; benchmark upgrades before promotion.
- **Contingency:** Make the affected eval advisory, restore pinned version where possible, rerun a calibrated comparison, obtain threshold approval.
- **Indicators:** Variance, false pass/fail rate, score discontinuities after client/model changes, judge disagreement with human labels.

#### R-011/R-012 — Specification gaming and protected-verifier changes

- **Controls:** Traceability validation, CODEOWNERS, protected policy/test paths, negative/property tests, hidden checks when justified, and a default-branch-controlled workflow.
- **Contingency:** Reject the run, revert unauthorized verifier changes through normal review, expand the regression corpus, and inspect similar merged changes.
- **Indicators:** Acceptance criteria with no verifier, reduced thresholds in same patch, skipped verifiers, unusual hidden-test access.

### 4.3 Availability and external dependencies

#### R-005/R-006/R-007 — Subscription availability and provider changes

- **Controls:** Use only first-party clients; unattended Claude authentication uses the documented `claude setup-token` path with scheduled rotation before the one-year expiry; preflight authentication; serialized/bounded workload; deterministic checks before agents; explicit `agent_unavailable` state; periodically verify official documentation and provider terms and fail the lane closed if the documented path changes.
- **Contingency:** Continue deterministic and human review; pause agent-required merge gates according to declared repository policy; reauthenticate interactively; do not copy auth files or invoke paid APIs automatically.
- **Indicators:** Login failures, usage-limit events, client errors, provider deprecation notices, authentication-file format changes.
- **Source anchors:** [Codex CI/CD authentication](https://developers.openai.com/codex/auth/ci-cd-auth), [Codex noninteractive mode](https://developers.openai.com/codex/noninteractive), [Claude authentication](https://code.claude.com/docs/en/iam), and [Claude CLI reference](https://code.claude.com/docs/en/cli-reference).

#### R-037 — Unauthorized provider transmission under consumer plans

- **Controls:** Classify every repository before agent use; prohibit confidential/client/employer, regulated, credential, production-data, and protected-holdout transmission unless the data owner and governing policy explicitly authorize the consumer services; verify OpenAI and Anthropic model-improvement controls; minimize prompts; avoid sensitive-session feedback; reassess quarterly.
- **Contingency:** Disable the agent lane, notify the data/security owner, revoke or delete provider-side data where supported, review affected sessions and contractual obligations, and move the repository to an approved business/API/local environment before resuming.
- **Indicators:** Missing classification, expired settings verification, secret/PII scanner finding, repository ownership change, new NDA/client constraint, provider privacy-policy change.
- **Source anchors:** [OpenAI model-improvement policy](https://help.openai.com/en/articles/5722486-how-your-data-is-used-to-improve-model-performance), [OpenAI Data Controls](https://help.openai.com/en/articles/7730893-data-controls-faq), and [Anthropic consumer training policy](https://privacy.claude.com/en/articles/10023580-is-my-data-used-for-model-training).

#### R-015/R-031 — Local outage and routing failure

- **Controls:** Manual hosted fallback in MVP; health monitoring; recovery runbook; workflow timeout; signed expiring heartbeat before automatic routing; off-host backup of non-secret configuration.
- **Contingency:** Rerun eligible deterministic jobs manually on GitHub-hosted infrastructure; queue subscription-authenticated agent work; repair local capacity.
- **Indicators:** Runner offline duration, queue age, heartbeat freshness, fallback count, infrastructure failure rate.

### 4.4 Supply chain and release

#### R-018/R-019/R-021 — Actions, cache, and toolchain compromise

- **Controls:** Full-SHA Action pins, minimal third-party Actions, tool/version locks, provenance and checksums, trust-tier cache namespaces, immutable release inputs, staged dependency updates, secret/SAST/SCA scans.
- **Contingency:** Disable affected action/cache, purge cache namespaces, restore pinned version, rebuild from clean inputs, assess released artifacts.
- **Indicators:** Pin drift, checksum mismatch, unexpected binary/network access, scan changes after tool update.
- **Source anchor:** [GitHub secure use reference](https://docs.github.com/en/actions/reference/security/secure-use).

#### R-020/R-027 — Provenance claims and release identity

- **Controls:** State the actual builder threat model; do not claim stronger SLSA levels than supported; separate release runner; short-lived OIDC; protected environments; verify signature/attestation before promotion; rehearse rollback.
- **Contingency:** Halt release, revoke federated role/session, restore prior immutable digest, publish corrected evidence.
- **Indicators:** Unsigned artifacts, subject/digest mismatch, OIDC claim mismatch, provenance generated outside approved builder.
- **Source anchors:** [SLSA build track](https://slsa.dev/spec/v1.2/build-track-basics) and [GitHub OIDC security](https://docs.github.com/en/actions/concepts/security/openid-connect).

#### R-029/R-032/R-033 — Privilege crossover

- **Controls:** Privileged workflow definition from protected default branch; exact candidate SHA; `persist-credentials: false`; read-only token in analysis job; separate minimal publisher; protected environments and rulesets; no admin token.
- **Contingency:** Disable workflow, revoke token/App installation, inspect audit logs and affected PRs/releases, restore rulesets.
- **Indicators:** Workflow permission expansion, default-branch/PR SHA mismatch, direct pushes, skipped required check, unrecognized deployment actor.

## 5. Risk acceptance and escalation triggers

The following events require immediate escalation to the project owner and security owner:

- Any evidence that Claude, Codex, GitHub administrative, release, NAS, or hypervisor credentials reached pull-request execution.
- A self-hosted runner communicating with an unauthorized home-network endpoint.
- A release whose digest differs from the reviewed/promoted digest.
- An agent or workflow bypassing protected merge or environment approval.
- Unapproved API billing or automatic hosted-compute cost.
- A residual risk score rising to 17 or greater.

The following events require a formal risk review before continuing the affected feature:

- Moving from private repositories to public or forked pull requests.
- Enabling unattended subscription-authenticated agents.
- Enabling automatic runner fallback or redispatch.
- Adding write-capable tools or unrestricted network access to reviewers.
- Adding a new deployment target, signing authority, or production environment.
- Replacing first-party clients with a third-party authentication broker.
- Claiming a formal supply-chain assurance level.

## 6. Architecture decision index

| ADR | Decision | Status | Rationale summary | Revisit trigger |
|---|---|---|---|---|
| ADR-001 | GitHub remains the source and workflow control plane | Proposed | Preserves standard PR/ruleset/Actions ecosystem and avoids building a forge | GitHub no longer meets security, cost, or availability needs |
| ADR-002 | Home server is the primary compute plane; hosted execution is fallback | Proposed | Uses available compute and reduces routine hosted usage | Local reliability or operations cost misses targets |
| ADR-003 | Build, agent, and release are separate trust zones | Proposed | Prevents untrusted code from reaching subscription or deployment credentials | Never; only strengthen boundaries |
| ADR-004 | MVP uses a dedicated Ubuntu VM; target uses disposable runners | Proposed | Provides a simple pilot path and a migration toward clean environments | Pilot threat model requires VM-per-job immediately |
| ADR-005 | Subscription work uses only first-party Claude Code and Codex clients | Proposed | Fits Claude Max/ChatGPT Pro and reduces credential-relay risk | Provider terms/auth change or an approved API budget exists |
| ADR-006 | Deterministic policy outranks all agent verdicts | Proposed | Model agreement is not proof; hard controls must remain authoritative | Never without a new project charter |
| ADR-007 | Reviews are blind first, then evidence-based cross-review | Proposed | Reduces anchoring and captures independent discovery before debate | Measured benefit does not justify usage/latency |
| ADR-008 | Repair is bounded to one pass by default and two maximum | Proposed | Prevents usage-draining loops and endless patch churn | Empirical data supports a safer lower limit; increases require explicit approval |
| ADR-009 | GitHub Spec Kit is the pilot specification system | Proposed for pilot | Good GitHub/agent alignment; avoids maintaining two spec authorities | Brownfield or cross-repository use favors OpenSpec |
| ADR-010 | Pipeline logic lives behind `ci/run`; Actions YAML is thin | Proposed | Enables local reproduction and portability to other CI systems | A component cannot reasonably run outside a provider primitive |
| ADR-011 | Shared memory begins with Git plus structured immutable run artifacts | Proposed | Auditable, simple, and resistant to opaque semantic-memory drift | Structured search demonstrably fails at expected corpus size |
| ADR-012 | No silent paid API fallback | Proposed | Respects the $0 incremental testing budget and prevents surprise spend | Project owner approves budget, provider, and ceiling in a new ADR |
| ADR-013 | Manual GitHub-hosted fallback precedes automatic routing | Proposed | Avoids stale-health and surprise-cost failure modes in MVP | Recovery is proven and signed heartbeat design passes testing |
| ADR-014 | Public/fork work uses GitHub-hosted or disposable hostile-code runners | Proposed | Persistent home runners lack a clean-environment guarantee | Equivalent stronger isolation is independently validated |
| ADR-015 | Build once and promote the same immutable digest | Proposed | Prevents production rebuild drift and supports rollback/provenance | Never; implementation may evolve |
| ADR-016 | Release identity is short-lived and separate from agent/build identity | Proposed | Minimizes credential lifetime and privilege crossover | Target cannot support federation; requires documented exception |
| ADR-017 | Full-SHA pinning is mandatory for third-party Actions | Proposed | Reduces upstream tag-mutation risk | Provider supplies a stronger, verifiable immutable mechanism |
| ADR-018 | Promptfoo and Phoenix are optional focused tools, not MVP dependencies | Proposed | Keeps initial system lean while preserving an eval/trace path | Pilot is itself an LLM product or needs experiment management immediately |
| ADR-019 | Agent instructions have one canonical source | Proposed | Prevents `AGENTS.md`/`CLAUDE.md` drift | Tool limitations require a documented generated derivative |
| ADR-020 | A small policy-enforcing broker is preferred over a general agent swarm | Proposed | Required workflow is bounded and benefits from auditable schemas/state | An established framework meets auth, isolation, schema, and cost constraints |
| ADR-021 | Consumer-plan use is gated by provider-transmission classification and approval | Proposed | Local isolation does not prevent source from being processed by providers | Business-plan terms, provider policies, or repository ownership/classification changes |

## 7. Initial decision records

### ADR-001 — Keep GitHub as the control plane

- **Status:** Proposed
- **Context:** The platform needs pull requests, rulesets, required checks, review, Actions compatibility, artifacts, release environments, and an online fallback.
- **Decision:** Retain GitHub as the canonical forge and workflow controller. Run most jobs locally through official self-hosted runner integration.
- **Consequences:** GitHub remains an availability and governance dependency. The project avoids duplicating source-control, identity, and workflow administration.
- **Rejected alternatives:** A new Forgejo/Gitea/Woodpecker deployment in the MVP; building a custom Actions controller.

### ADR-003 — Separate trust zones

- **Status:** Proposed
- **Context:** Pull-request code is potentially hostile, while Claude Max/ChatGPT Pro credentials and release identities are high-value secrets.
- **Decision:** Operate at least three isolated zones: credential-free build/test, subscription-authenticated agent review, and protected release/deployment. A router/control job performs no repository checkout or execution.
- **Consequences:** Some artifacts must cross zones through a validated, immutable evidence interface. Infrastructure is more complex but compromise blast radius is materially reduced.
- **Related risks:** R-001, R-002, R-003, R-014, R-017, R-028, R-032.

### ADR-004 — Dedicated VM first; ephemeral execution as target

- **Status:** Proposed
- **Context:** A pilot should be operational quickly, but persistent self-hosted runners lack a guaranteed clean state between jobs. GitHub recommends careful isolation and provides ARC for autoscaled Kubernetes runners: [self-hosted runner security](https://docs.github.com/en/actions/reference/security/secure-use) and [Actions Runner Controller](https://docs.github.com/en/actions/concepts/runners/actions-runner-controller).
- **Decision:** Start with one rebuildable Ubuntu VM restricted to private/trusted contexts. Evaluate ARC for scalable trusted workloads and GARM/Incus or another VM-per-job mechanism for hostile-code isolation.
- **Consequences:** The MVP must not run public/fork code locally. The VM image, registration, cleanup, and recovery must be documented.
- **Related risks:** R-002, R-003, R-015, R-024, R-028.

### ADR-005 — Use first-party subscription clients

- **Status:** Proposed
- **Context:** The project has Claude Max and ChatGPT Pro but no approved per-token API budget. OpenAI documents ChatGPT-authenticated trusted private automation, and both providers support noninteractive native clients. Anthropic explicitly documents `claude setup-token` — a one-year, inference-scoped OAuth token for CI pipelines and scripts supplied through `CLAUDE_CODE_OAUTH_TOKEN`, supported on Pro/Max plans — which provides a documented first-party path for unattended Claude use: [Codex CI/CD authentication](https://developers.openai.com/codex/auth/ci-cd-auth), [Codex noninteractive mode](https://developers.openai.com/codex/noninteractive), [Claude authentication](https://code.claude.com/docs/en/iam), and [Claude CLI reference](https://code.claude.com/docs/en/cli-reference).
- **Decision:** Invoke the installed, logged-in `claude` and `codex` clients directly; unattended Claude work uses the documented setup-token path, qualified through the P4-09 authentication spike (PAC-018) before the broker depends on it. Do not extract tokens or route subscription credentials through a generic third-party service, expose Max authentication through a hosted multi-user product, or use the Agent SDK or API-key GitHub Actions in the subscription-funded lane. Use OpenAI's official [Codex plugin for Claude Code](https://github.com/openai/codex-plugin-cc) for the initial supervised integration.
- **Consequences:** Agent capacity is subject to subscription limits, client behavior, login/token expiry, and provider changes. The setup-token requires rotation ahead of its one-year expiry. If either provider's documented unattended path changes or is withdrawn, the agent lane fails closed pending review. The agent lane may pause while deterministic CI and human review continue.
- **Related risks:** R-001, R-005, R-006, R-007, R-009, R-017.

### ADR-006/ADR-007/ADR-008 — Deterministic authority and bounded two-agent review

- **Status:** Proposed
- **Context:** Independent agents can find different issues, but they may share blind spots or reinforce an incorrect conclusion. Unbounded repair loops consume subscriptions and generate noisy patches.
- **Decision:** Run deterministic checks first. Conduct sealed blind reviews, then structured cross-review. The policy arbiter—not model confidence—determines blocking status. Allow one repair pass by default and no more than two; then require human adjudication.
- **Consequences:** Review latency and usage are controlled. Some disputed issues intentionally stop for human judgment. Agent consensus cannot override a failing hard gate.
- **Related risks:** R-008, R-009, R-010, R-034, R-036.

### ADR-009 — Use Spec Kit for the pilot

- **Status:** Proposed for pilot
- **Context:** The project needs a versioned intent-first workflow, but using multiple competing specification authorities would create drift.
- **Decision:** Use [GitHub Spec Kit](https://github.com/github/spec-kit) for the first repository, then normalize its acceptance criteria into the platform's executable verifier manifest. Do not install OpenSpec simultaneously.
- **Consequences:** If brownfield, delta-based, or cross-repository changes dominate, conduct a focused comparison with [OpenSpec](https://github.com/Fission-AI/OpenSpec) and supersede this decision if warranted.
- **Related risks:** R-011, R-012, R-025.

### ADR-010 — Keep workflow YAML thin

- **Status:** Proposed
- **Context:** Pipeline logic embedded entirely in Actions YAML is difficult to reproduce locally and hard to reuse in another CI system.
- **Decision:** Implement a versioned `ci/run` command with language adapters and normalized evidence. GitHub Actions provides triggers, permissions, matrices, caching, artifacts, environments, and orchestration.
- **Consequences:** Adapter contract design becomes a core platform responsibility. Provider-specific capabilities remain in thin workflow wrappers.
- **Related risks:** R-021, R-022, R-023, R-025.

### ADR-011/ADR-019 — Structured memory and one instruction source

- **Status:** Proposed
- **Context:** Agents need previous decisions and findings, but free-form shared memory can be poisoned, become stale, and conflict across tool-specific files.
- **Decision:** Treat specifications, tests, ADRs, approved policies, and resolved findings as canonical memory. Store each run as immutable structured artifacts. Maintain one canonical agent-rules file and minimal generated/tool-specific entry points. Do not store raw chain-of-thought or deploy a vector database in the MVP.
- **Consequences:** Retrieval is deliberately simple and auditable. Semantic search can be added only after measurable need and provenance rules are defined.
- **Related risks:** R-013, R-014, R-030.

### ADR-012/ADR-013 — No paid fallback; manual hosted recovery first

- **Status:** Proposed
- **Context:** GitHub Actions does not natively express a reliable `self-hosted OR ubuntu-latest` failover. Automatic routing can act on stale health, and hosted execution or provider APIs may cost money.
- **Decision:** The MVP exposes a manual `force_hosted` option for eligible deterministic jobs. Subscription-authenticated agent work queues locally. No paid API is invoked automatically. Automatic routing requires a signed expiring heartbeat, cost guardrails, and a tested redispatch policy.
- **Consequences:** Some outages require manual action and agent review can pause. The design avoids credential transfer and surprise spending.
- **Related risks:** R-015, R-016, R-017, R-031.

### ADR-015/ADR-016 — Immutable promotion and federated release identity

- **Status:** Proposed
- **Context:** Rebuilding for production creates artifact drift. Long-lived release credentials increase impact if a runner is compromised.
- **Decision:** Build once, identify the artifact by digest, verify the digest and evidence at each stage, and deploy with short-lived target-bound identity such as GitHub OIDC. Use a protected release environment and a separately controlled runner.
- **Consequences:** Artifact storage and metadata become production dependencies. Deployment targets without federation need an explicit exception and compensating controls.
- **Related risks:** R-020, R-027, R-029.

### ADR-018 — Defer broad eval/observability stack

- **Status:** Proposed
- **Context:** Promptfoo is useful for agent/product evals, and Phoenix is useful for local OpenTelemetry traces and experiments, but making both mandatory before the core loop works increases operational surface.
- **Decision:** Use repository-native tests and normalized evidence first. Add Promptfoo when evaluating LLM behavior or coding-agent task corpora, and add Phoenix when trace/experiment analysis provides a clear benefit. Do not deploy overlapping eval platforms by default.
- **Consequences:** Initial observability is simpler. Instrumentation schemas must remain OpenTelemetry-compatible so richer tooling can be added later.
- **Related risks:** R-010, R-025, R-030, R-035.

### ADR-020 — Prefer a small deterministic broker

- **Status:** Proposed
- **Context:** The required agent workflow is a bounded state machine: collect immutable inputs, launch two first-party clients, validate JSON, seal reports, conduct cross-review, arbitrate, permit limited repair, and retain evidence.
- **Decision:** Build a small policy-enforcing broker, preferably as a statically compiled Go service or CLI, rather than adopting a general-purpose multi-agent swarm. The broker shall not hold merge or release authority.
- **Consequences:** The project owns a modest amount of orchestration code but gains precise schemas, testable state transitions, and a smaller attack surface. The decision remains proposed until the supervised plugin pilot clarifies which integration code is actually needed.
- **Related risks:** R-004, R-005, R-009, R-025.

### ADR-021 — Gate consumer-plan use by data classification

- **Status:** Proposed
- **Context:** ChatGPT Pro and Claude Max are consumer products. Approved source is processed on provider systems, and consumer model-improvement/retention controls differ from enterprise contractual protections.
- **Decision:** Require an owner-approved transmission class before either agent receives repository content. Public and approved private-personal code may use the native clients after data-control verification. Confidential/client/employer, regulated, credential-bearing, production-data, and protected-holdout content is denied unless the relevant data owner and governing policy explicitly authorize these consumer services; otherwise use an approved business/API/local alternative.
- **Consequences:** Some repositories cannot use the subscription-funded lane. Classification and settings verification become auditable prerequisites, and provider policies require periodic review.
- **Related risks:** R-007, R-014, R-030, R-037.

## 8. Deferred decisions

| Decision | Required evidence before deciding | Owner |
|---|---|---|
| ARC/k3s versus GARM/Incus for the first ephemeral fleet | Home server hypervisor/network inventory; workload isolation classes; concurrency and image-start benchmarks | Platform owner |
| Pilot language adapter | Pilot repository selection and existing toolchain | Project owner and repository maintainer |
| Blocking thresholds for model-judged evals | Human-labeled calibration set, variance analysis, and false-pass/false-fail costs | Evaluation owner |
| Evidence storage backend and retention periods | Data classification, expected run volume, recovery requirements, and available storage | Platform owner and security owner |
| Automatic hosted fallback | Manual recovery results, signed heartbeat design, cost limits, and failure classifier accuracy | Project owner |
| Formal supply-chain assurance target | Release consumers, threat model, builder isolation, and external compliance requirements | Security and release owners |
| Production deployment strategy | Target environment, service SLOs, health signals, and rollback mechanism | Release owner |
| Paid provider API fallback | Business need, budget ceiling, credential model, expected usage, and approval | Project owner |

## 9. Decision and risk review cadence

- Review open critical/high risks before each phase exit.
- Review all risks monthly during the pilot and quarterly after stabilization.
- Review provider/authentication risks whenever either client updates materially or official guidance changes.
- Review release risks before every new deployment target is enabled.
- Re-score risks after an incident, escaped defect, fallback exercise, red-team exercise, or major topology change.
- Supersede ADRs rather than editing history silently. Record the replacing ADR, rationale, approver, and effective date.

## 10. Immediate actions

| ID | Action | Owner | Due milestone |
|---|---|---|---|
| ACT-001 | Inventory home server, hypervisor, network segments, storage, UPS, and available VM capacity | Platform owner | P0 |
| ACT-002 | Select and classify one private pilot repository | Project owner | P0 |
| ACT-003 | Validate current Claude Code and Codex first-party authentication on the intended agent VM | Platform owner | P4 before G4 |
| ACT-004 | Create credential-path and LAN-reachability red-team tests before agent credentials are introduced | Security owner | P3–P4 before G4 |
| ACT-005 | Define the acceptance manifest schema and protected verifier paths | Repository maintainer | P1 |
| ACT-006 | Define normalized evidence and reviewer JSON schemas | Platform owner | P1 and P5 before G5 |
| ACT-007 | Establish scanner baselines, owners, and time-bound waiver format | Security owner | P2 |
| ACT-008 | Document local outage, hosted fallback, credential incident, and runner rebuild runbooks | Platform owner | P3 |
| ACT-009 | Build policy tests proving deterministic gates cannot be waived and loop limits cannot be exceeded | Platform owner | P5 before G5 |
| ACT-010 | Conduct a threat-model and risk-register review before unattended agent execution | Project and security owners | P5 before G5 |
