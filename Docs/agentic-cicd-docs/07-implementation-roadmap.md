# Implementation Roadmap

**Status:** Proposed  
**Planning horizon:** Pilot through hardened v1  
**Estimated effort:** Approximately one focused week for the pilot and two to four weeks for a hardened v1, subject to the home-server hypervisor, repository complexity, and deployment targets.

## 1. Delivery strategy

The platform will be delivered as a sequence of independently useful increments. Deterministic CI comes before unattended agent orchestration. The first agent workflow is supervised and uses official native clients. Ephemeral infrastructure, automated fallback, full observability, and release hardening follow only after the base loop is reliable.

The pilot should use one private repository with representative build and test behavior but no irreplaceable production secrets. A small Go or Python service with a database boundary, API contract, container image, and deployable test environment is a good candidate.

## 2. Phase overview

| Phase | Outcome | Indicative duration | Exit gate |
|---|---|---:|---|
| 0 — Discovery and threat model | Known environment, risk tiers, and pilot | 0.5–1 day | G0 |
| 1 — Specification and local quality loop | Executable spec and portable `ci/run` | 1–2 days | G1 |
| 2 — GitHub deterministic CI | Protected, evidence-producing PR pipeline | 2–3 days | G2 |
| 3 — Local execution plane | Dedicated home runner with recovery path | 1–2 days | G3 |
| 4 — Supervised two-model workflow | Claude/Codex review with bounded repair | 1–2 days | G4 |
| 5 — Automated agent broker | Blind and cross-review with deterministic arbitration | 3–5 days | G5 |
| 6 — Ephemeral scaling and observability | Disposable runners, local evidence, health routing | 3–7 days | G6 |
| 7 — Release and deployment hardening | Signed, attested, staged, reversible delivery | 2–5 days | G7 |

Durations are estimates, not deadlines. Each gate should be passed before increasing autonomy.

### 2.1 Named releases

These names are authoritative across the documentation set:

| Release | Roadmap phases | Capability boundary |
|---|---|---|
| Pilot | P0–P4 | Specification-driven deterministic CI, dedicated local runner, manual hosted fallback, and supervised Claude-to-Codex review through official clients |
| MVP | P0–P5 | Pilot plus automated native-client broker, sealed blind/cross-review, deterministic arbitration, bounded repair, and immutable evidence bundle |
| Hardened v1 | P0–P7 | MVP plus disposable runner capacity, local evidence/telemetry services, automated health routing, and signed/attested reversible delivery |

Document 07 owns phase numbers and gate IDs. Component documents may define work packages, but they must map them to these phases rather than introduce another project-phase sequence.

## 3. Phase 0 — Discovery and threat model

### Objectives

- Confirm what the home server can safely host.
- Select a pilot repository and define trust boundaries.
- Establish risk tiers and ownership.
- Decide which costs are permitted.

### Inputs to collect

- Server CPU, RAM, storage, networking, power protection, and backup capability.
- Host operating system and hypervisor: Proxmox, Incus/LXD, KVM/libvirt, VMware, Hyper-V, bare Ubuntu, or other.
- Whether k3s/Kubernetes is already present.
- Repository visibility and whether fork-based contributions are accepted.
- Initial languages, package managers, services, databases, and container requirements.
- GitHub plan, Actions minutes, artifact limits, and registry choice.
- Deployment targets: home lab, cloud, Kubernetes, VMs, serverless, or mixed.
- Data classifications and compliance obligations.
- Acceptable API spend. The initial assumption is zero optional provider API spend.

### Work items

| ID | Work item | Deliverable |
|---|---|---|
| P0-01 | Inventory infrastructure | Environment inventory |
| P0-02 | Select pilot | Pilot decision record |
| P0-03 | Classify code trust | Trusted/private/fork execution matrix |
| P0-04 | Define change risk levels | Low/medium/high/critical policy |
| P0-05 | Establish owners | Project, security, infrastructure, release owners |
| P0-06 | Define budget policy | Subscription/API/Actions/storage limits |
| P0-07 | Review initial risk register | Approved mitigations and deferred risks |

### Exit gate G0

- Pilot repository selected.
- Server and network inventory complete.
- Public/fork code policy explicitly decided.
- Required owners named.
- Initial threat model approved.
- No unresolved critical assumption blocks implementation.

## 4. Phase 1 — Specification and local quality loop

### Objectives

- Make specifications versioned and acceptance criteria executable.
- Provide one consistent local CI interface.
- Establish deterministic checks before agent automation.

### Work items

| ID | Work item | Deliverable |
|---|---|---|
| P1-01 | Initialize Spec Kit | Constitution and specification workflow |
| P1-02 | Add change manifest schema | Machine validation for acceptance mappings |
| P1-03 | Define repository agent rules | Canonical `agents/rules.md` plus minimal tool entry points |
| P1-04 | Implement `ci/run detect` | Ecosystem and toolchain detection |
| P1-05 | Implement language adapter | Format, lint, type, build, and tests |
| P1-06 | Add security-fast adapter | Secrets, SAST, and dependency checks |
| P1-07 | Normalize evidence | JUnit, SARIF, coverage, and manifest output |
| P1-08 | Add developer hooks | Optional pre-commit or Lefthook integration |
| P1-09 | Create first representative spec | Approved change with hard/scored criteria |

### Design notes

- Choose Spec Kit or OpenSpec, not both.
- Workflow logic belongs behind `ci/run`, not duplicated in CI YAML.
- `act` may provide local Actions preflight, but it is not an acceptance environment because its behavior differs from GitHub in documented areas.
- Tool versions and lockfiles must be checked in before the results are treated as reproducible.

### Exit gate G1

- A clean clone can bootstrap and execute `ci/run full` noninteractively.
- Every mandatory acceptance criterion in the pilot spec resolves to a verifier.
- The adapter emits machine-readable evidence.
- A deliberately broken change produces the expected failure.
- Local documentation covers setup and common failures.

## 5. Phase 2 — GitHub deterministic CI

### Objectives

- Make quality enforcement independent of an agent's opinion.
- Establish standard GitHub governance and evidence publication.

### Work items

| ID | Work item | Deliverable |
|---|---|---|
| P2-01 | Create `pr.yml` | Required PR checks |
| P2-02 | Create `nightly.yml` | Deeper scheduled checks |
| P2-03 | Configure ruleset | Protected default branch and required statuses |
| P2-04 | Add CODEOWNERS | Ownership for sensitive paths |
| P2-05 | Harden Actions | Minimal permissions and full-SHA pins |
| P2-06 | Add workflow scanners | actionlint and zizmor |
| P2-07 | Configure artifacts | JUnit, SARIF, coverage, sanitized logs |
| P2-08 | Configure concurrency/timeouts | Stale-run cancellation and bounded jobs |
| P2-09 | Add dependency automation | Dependabot or Renovate |
| P2-10 | Implement final policy gate | Normalized pass/fail with reason codes |

### Required pilot checks

```text
spec / traceability
ci / format-lint-type
ci / build-unit
ci / integration-contract
security / secrets-sast-sca
eval / acceptance
policy / decision
```

### Exit gate G2

- Direct default-branch pushes are blocked.
- All required checks execute on GitHub-hosted infrastructure.
- Third-party Actions are pinned and workflow permissions are explicit.
- Test, analysis, and policy failures are visible and correctly block merge.
- An agent-generated patch has no path around required checks.

## 6. Phase 3 — Local execution plane

### Objectives

- Move trusted deterministic workloads to the home server.
- Preserve cloud recovery while establishing isolation.

### Work items

| ID | Work item | Deliverable |
|---|---|---|
| P3-01 | Provision dedicated Ubuntu VM | Isolated runner host |
| P3-02 | Segment network | Firewall/VLAN rules and egress policy |
| P3-03 | Register scoped runner | Repository or runner-group registration |
| P3-04 | Apply runner labels | Explicit trust and capability labels |
| P3-05 | Configure ephemeral workspace cleanup | Per-run work directory lifecycle |
| P3-06 | Establish cache namespaces | Trust-separated build caches |
| P3-07 | Add manual fallback | `force_hosted` workflow input |
| P3-08 | Add monitoring | Disk, CPU, queue, service, and runner status |
| P3-09 | Run failure drills | Offline server and interrupted-job tests |

### Initial network policy

- Runner-initiated outbound HTTPS to required GitHub, registry, package, and scanner endpoints.
- No unsolicited inbound Internet exposure.
- No access to NAS administration, hypervisor management, personal devices, or unrelated home services.
- No host Docker socket for pull-request jobs.
- No AI subscription credentials, deployment credentials, or long-lived repository write tokens.

### Exit gate G3

- Trusted CI runs locally with results equivalent to the hosted baseline.
- The home server can be powered off without preventing a manual hosted run.
- The runner cannot reach protected home-management networks.
- Workspace cleanup and cache separation are verified.
- A compromised test container cannot obtain AI or release credentials.

## 7. Phase 4 — Supervised two-model workflow

### Objectives

- Validate the human-supervised review protocol before unattended orchestration.
- Use current Claude Max and ChatGPT Pro subscriptions through official clients.

### Work items

| ID | Work item | Deliverable |
|---|---|---|
| P4-01 | Install and authenticate Claude Code | User-authenticated native client under a dedicated local OS account |
| P4-02 | Install and authenticate Codex CLI | User-authenticated native client under a dedicated local OS account |
| P4-03 | Install official Codex Claude plugin | `/codex:review` workflow |
| P4-04 | Add review JSON schema | Validated structured findings |
| P4-05 | Define evidence bundle | Frozen spec/diff/check package |
| P4-06 | Exercise adversarial review | Versioned seeded-defect corpus with known defect labels, enabling detection and recall measurement per severity and category |
| P4-07 | Apply iteration budgets | Maximum turns, time, and two repair cycles |
| P4-08 | Measure review quality | Detection, false positive, and usage metrics; blind-review overlap and cross-review disposition-change tracking from the first run |
| P4-09 | Authentication qualification spike | Recorded unattended-auth qualification: generate a token with `claude setup-token` on the agent host; run Claude from a clean broker-owned directory with curated inputs; complete 10–20 unattended structured-output runs; test reboot persistence, invalid tokens, usage exhaustion, timeouts, malformed output, and secret redaction; record documentation URL, plan, client version, auth mode, verification date, and next review date |

### Exit gate G4

- Claude can implement and invoke a fresh Codex review through the official plugin.
- Review results validate against the schema.
- A seeded high-severity defect is found and repaired.
- The loop stops on budget, no progress, or the second failed repair.
- Authentication artifacts never appear in repositories, logs, or build runners.
- The unattended Claude authentication path is qualified per P4-09 (PAC-018): repeated unattended structured-output runs succeed, and expiry, invalid-token, usage-exhaustion, timeout, and malformed-output cases fail closed without exposing the token. The lane fails closed if the documented path changes.

## 8. Phase 5 — Automated agent broker

### Objectives

- Add independent blind reviews, controlled cross-review, and deterministic arbitration.
- Keep native subscription clients local and prevent credential relay.

### Suggested implementation

A small Go service or command-line broker is recommended because it can be distributed as one static binary and can enforce concurrency, subprocess boundaries, JSON Schema validation, hashing, timeouts, and append-only output without requiring the target repository's language runtime.

### Work items

| ID | Work item | Deliverable |
|---|---|---|
| P5-01 | Implement run manifest | Content-addressed run identity |
| P5-02 | Build native CLI adapters | Claude and Codex subprocess launchers |
| P5-03 | Enforce blind review | Sealed initial reports |
| P5-04 | Implement cross-review | Accept/reject/modify responses |
| P5-05 | Implement policy arbitration | Deterministic reason-coded decision |
| P5-06 | Implement repair workflow | Patch handoff and full rerun |
| P5-07 | Add queue and concurrency controls | Subscription-safe serialization |
| P5-08 | Add command/network allowlists | Hardened reviewer environment |
| P5-09 | Add publication job | Sanitized PR report without code execution |
| P5-10 | Create adversarial test suite | Injection, malformed output, timeout, disagreement |

### Exit gate G5

- Initial reviewers cannot see each other's output.
- Cross-review occurs only after both reports validate and are sealed.
- Deterministic failures always block regardless of agent verdict.
- Disputed high-severity findings route to a reproducer or human.
- The broker never reads or exports provider credential files.
- Injection tests cannot cause access to prohibited paths or networks.
- Every decision is reproducible from its stored evidence and policy version.

## 9. Phase 6 — Ephemeral scaling and observability

### Objectives

- Replace persistent build hosts with disposable execution where practical.
- Add local evidence retention and robust health-based routing.

### Decision branch

- Choose **k3s + ARC** when GitHub-supported autoscaling, operational simplicity, and trusted/private workloads are primary.
- Choose **GARM + Incus VMs** for workloads requiring a stronger VM boundary or when Kubernetes is undesirable.
- Continue using GitHub-hosted runners for public/fork code even after local scaling exists.

### Work items

| ID | Work item | Deliverable |
|---|---|---|
| P6-01 | Select runtime architecture | ADR and capacity model |
| P6-02 | Deploy disposable runner pool | Scale-to-zero execution |
| P6-03 | Deploy local object storage | Evidence and log retention |
| P6-04 | Deploy registry/cache services | OCI, package, BuildKit, or `sccache` acceleration |
| P6-05 | Deploy Phoenix/OpenTelemetry | Agent and pipeline traces |
| P6-06 | Implement signed health heartbeat | Stale-expiring local health signal |
| P6-07 | Implement hosted router | Dynamic `runs-on` output |
| P6-08 | Implement watchdog | Cancel/redispatch infrastructure failures |
| P6-09 | Add backup and restore | Config and evidence recovery |
| P6-10 | Load and chaos test | Capacity, outage, storage, and network drills |

### Exit gate G6

- Ordinary runner instances are disposable and leave no persistent workspace.
- Scale-to-zero and scale-up meet the agreed queue SLO.
- The router handles healthy, stale, forced-hosted, and offline states.
- The watchdog retries only classified infrastructure failures.
- Local evidence can be restored from backup.
- Sensitive evidence is not uploaded to GitHub inadvertently.

## 10. Phase 7 — Release and deployment hardening

### Objectives

- Implement industry-standard artifact security and reversible deployment.

### Work items

| ID | Work item | Deliverable |
|---|---|---|
| P7-01 | Create `release.yml` | Clean protected build |
| P7-02 | Generate SBOM | CycloneDX or SPDX artifact |
| P7-03 | Scan final artifact | Policy-evaluated vulnerability report |
| P7-04 | Sign and attest | Cosign signatures and provenance |
| P7-05 | Configure environment protection | Approvers and separation of duties |
| P7-06 | Configure OIDC | Short-lived deployment identity |
| P7-07 | Create `deploy.yml` | Digest-based promotion |
| P7-08 | Add staging/smoke/canary | Automated validation |
| P7-09 | Implement rollback | Tested restoration procedure |
| P7-10 | Conduct release drill | End-to-end evidence and recovery test |

### Exit gate G7

- The artifact is rebuilt cleanly from a protected source revision.
- SBOM, scan, signature, provenance, and policy pass against one digest.
- Staging uses the same digest proposed for production.
- Deployment uses protected approval and short-lived identity.
- Rollback is tested, timed, and documented.
- Release evidence links source, build, artifact, approver, and deployment.

## 11. Cross-phase verification tests

The following test scenarios should be introduced progressively:

1. Linter and compiler failure.
2. Unit-test failure that passes on a retry, verifying retry-to-green is rejected.
3. Secret inserted in source.
4. Vulnerable dependency introduced.
5. Workflow changed to increase token permissions.
6. Pull-request instruction attempting to read provider credentials.
7. Claude and Codex agreeing on an incorrect claim.
8. Reviewers disagreeing about a real high-severity defect.
9. Malformed or schema-invalid agent output.
10. Agent process timeout and subscription limit exhaustion.
11. Home server offline before dispatch.
12. Runner failure after job assignment.
13. Poisoned cache candidate.
14. Scanner advisory database changed between PR and release.
15. Deployment health check failure requiring rollback.

## 12. Metrics collected from the pilot

| Category | Measures |
|---|---|
| Delivery | Lead time, PR cycle time, merge frequency, deployment frequency |
| Reliability | CI pass rate, infrastructure-error rate, flaky-test rate, rollback rate |
| Quality | Escaped defects, changed-code coverage, mutation trend, security findings |
| Agent efficacy | Pass@1, repair success, independently confirmed findings, false-positive rate, diff churn |
| Efficiency | Queue time, runner utilization, cache hit rate, local-vs-hosted minutes |
| Subscription usage | Agent runs, turns, throttles, authentication interruptions |
| Governance | Waiver count, expired waivers, unowned failures, policy bypass attempts |

The pilot should establish baselines rather than arbitrary optimization targets. Targets are approved after two to four weeks of representative data.

## 13. Deferred backlog

Do not place these on the MVP critical path:

- A vector database for agent memory.
- A self-hosted Git forge replacing GitHub.
- Simultaneous deployment of Promptfoo, DeepEval, Phoenix, and Langfuse.
- Large swarm frameworks or dozens of agent personas.
- Automatic merge based only on model consensus.
- Custom provider OAuth/token relay.
- Dagger migration before the repository-owned command interface stabilizes.
- Strong SLSA claims from a builder controlled by the same administrator producing provenance.
- Multi-cloud deployment unless a concrete use case requires it.

## 14. Immediate next actions

1. Complete the Phase 0 environment questionnaire.
2. Select the pilot repository.
3. Name project, security, infrastructure, and release owners.
4. Approve or modify the initial architectural decisions in Document 09.
5. Create the pilot repository's first specification using the provided template.
6. Implement `ci/run` and one language adapter.
