# Infrastructure and Deployment

| Field | Value |
|---|---|
| Document ID | ACI-INFRA-006 |
| Version | 0.1 |
| Status | Proposed baseline |
| Date | 2026-07-13 |
| Audience | Home-lab operator, platform engineer, security reviewer, and release owner |
| Scope | Runner infrastructure, network and storage boundaries, GitHub routing, credentials, release supply chain, deployment, and recovery |
| Related documents | [System architecture](02-system-architecture.md), [CI/CD quality and security standard](05-ci-cd-quality-and-security-standard.md), [Risk register](09-risk-register-and-decisions.md) |

## 1. Purpose

This document turns the system architecture into an operable local-first infrastructure plan. It defines:

- A safe pilot deployment that can be built without first operating a Kubernetes cluster.
- A target runner topology with disposable execution environments.
- The decision boundary between Actions Runner Controller and GARM with Incus.
- GitHub-hosted fallback behavior when the home platform is unavailable or a change is untrusted.
- Network, credential, storage, cache, evidence, and observability boundaries.
- A build-once release path using OIDC, SBOMs, vulnerability scanning, signatures, provenance, staged deployment, health verification, and rollback.

The design is hypervisor-neutral. Proxmox, Incus, libvirt, VMware, or another home-server virtualization layer can host the pilot and MVP as long as it supports dedicated networks, snapshots, and automated VM replacement.

## 2. Deployment principles

1. GitHub is the workflow and review control plane; the home server is the preferred execution plane.
2. Repository code receives no Claude, Codex, signing, production, hypervisor, NAS, or router credentials.
3. No Internet-facing inbound port is opened for a GitHub runner.
4. Persistent runners execute only private, owner-controlled changes during the pilot and MVP.
5. Public repositories, forked pull requests, and unknown contributors use GitHub-hosted runners until disposable local VMs are ready.
6. A runner is disposable infrastructure, not a pet workstation.
7. Release jobs never reuse a pull-request workspace.
8. Artifacts are built once and promoted by immutable digest.
9. Caches improve speed but are never treated as evidence or trusted release inputs.
10. Every fallback, retry, waiver, and deployment decision is recorded.

## 3. Pilot infrastructure

### 3.1 Topology

~~~mermaid
flowchart TD
    GH["GitHub Actions and repository"]

    subgraph HOME["Home-server virtualization"]
        CI["ci-pilot Ubuntu VM"]
        AGENT["agent-pilot VM"]
        REL["Optional protected release VM"]
        STORE["Optional local evidence mirror"]
    end

    REG["GHCR or artifact registry"]
    ENV["Staging and production"]
    HOSTED["GitHub-hosted fallback"]

    GH --> CI
    GH --> HOSTED
    CI --> STORE
    STORE --> AGENT
    GH --> REL
    REL --> REG
    REG --> ENV
~~~

### 3.2 Initial virtual machines

| VM | Suggested starting size | Trust and lifecycle | Access |
|---|---|---|---|
| ci-pilot | 4-8 vCPU, 8-16 GB RAM, 60-120 GB ephemeral disk | Persistent only for the pilot; rebuild from template regularly | Outbound GitHub and package sources/proxies; no trusted LAN |
| agent-pilot | 4-8 vCPU, 8-16 GB RAM, encrypted 40-80 GB disk | Persistent login state in dedicated local OS accounts | Provider endpoints, read-only evidence, isolated worktree storage |
| release-protected, when CD is enabled | 4-8 vCPU, 8-16 GB RAM, 60-120 GB ephemeral disk | Powered on or registered only for protected releases | GitHub, registry, evidence store, OIDC issuer/STS, deployment gateway |
| local evidence mirror, optional before hardened v1 | Capacity based on retention; begin with 100-500 GB only when needed | Persistent and backed up | Private service network only |

These are starting points, not fixed requirements. Collect CPU, memory, disk, queue, and cache metrics for several weeks before increasing capacity.

### 3.3 Pilot restrictions

- The ci-pilot runner label is usable only by private repositories in an allowlisted GitHub organization or account.
- Workflows from forks cannot select it.
- It does not mount the host Docker socket, NAS, source archive, home directory, or hypervisor API.
- If container builds require privileged Docker, they run inside the disposable CI VM; the entire VM is considered compromised afterward and is rebuilt.
- The agent VM is invoked manually or only from a protected default-branch workflow.
- The agent VM never runs a pull request's build scripts.
- Release jobs use a fresh checkout of a protected commit and a separate runner identity.
- Manual force_hosted is the initial outage path. Automated local-to-hosted routing is deferred until the failure classifier is tested.

GitHub's [secure-use reference](https://docs.github.com/en/actions/reference/security/secure-use) explains why persistent self-hosted runners should not execute untrusted pull-request code.

## 4. Target infrastructure

~~~mermaid
flowchart TD
    GH["GitHub Actions control plane"]
    ROUTER["Hosted router job"]

    subgraph HOME["Home execution plane"]
        ARC["ARC on k3s"]
        GARM["GARM controller"]
        INCUS["Disposable Incus VMs"]
        BROKER["Agent broker"]
        EVIDENCE["S3-compatible evidence store"]
        OBS["OpenTelemetry and Phoenix"]
        RELEASE["Disposable release runner"]
    end

    HOSTED["GitHub-hosted runners"]
    REGISTRY["Artifact registry"]
    TARGETS["Deployment environments"]

    GH --> ROUTER
    ROUTER --> ARC
    ROUTER --> GARM
    GARM --> INCUS
    ROUTER --> HOSTED
    ARC --> EVIDENCE
    INCUS --> EVIDENCE
    EVIDENCE --> BROKER
    EVIDENCE --> OBS
    GH --> RELEASE
    RELEASE --> REGISTRY
    REGISTRY --> TARGETS
~~~

GitHub Actions is the workflow scheduler and control plane; it is not simply a Docker launcher. A standard GitHub-hosted Linux job generally receives a fresh virtual machine, and a workflow may optionally place job steps in a container inside that environment. A self-hosted runner instead executes on the machine or VM where its runner service is installed, so lifecycle, isolation, cleanup, and network policy are the operator's responsibility. This design preserves Actions workflow semantics while choosing the local isolation unit by workload risk. See [GitHub-hosted runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners) and [job containers](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#jobsjob_idcontainer).

The target design is hybrid:

- ARC is the preferred autoscaler for high-throughput, owner-controlled jobs where Kubernetes pod isolation is acceptable.
- GARM with Incus is the preferred local path for lower-trust pull-request code, Docker-in-Docker, compiler plugins, or other workloads that deserve a VM boundary.
- GitHub-hosted runners remain the default for public/fork PRs and the fallback for deterministic jobs during local outages.
- Agent and release workloads remain in dedicated pools and never share nodes with general CI.

## 5. ARC versus GARM with Incus

### 5.1 Decision matrix

| Criterion | Actions Runner Controller | GARM with Incus |
|---|---|---|
| Upstream support | GitHub-supported Kubernetes operator | Community project and provider |
| Isolation unit | Kubernetes pod, normally sharing a kernel | Disposable VM or system container; VM offers a distinct kernel |
| Startup speed | Usually faster | Usually slower than a pod |
| Density | High | Lower |
| Operational prerequisite | Kubernetes expertise and k3s/cluster operations | Incus and VM image lifecycle expertise |
| Scale-to-zero | Native runner scale-set pattern | Supported through managed pools |
| Docker build handling | DinD often needs privileged operation; risk depends on node boundary | Privilege can be contained inside a disposable VM |
| Best fit | Trusted internal matrices, unit tests, ordinary builds | Lower-trust PRs, native toolchains, kernel-sensitive or privileged builds |
| Maturity risk | Lower for GitHub integration | Higher; pin versions and test upgrades |
| Recommended role | First autoscaler when throughput justifies k3s | Stronger-isolation local pool introduced for risky jobs |

GitHub documents ARC as its Kubernetes autoscaling solution in [Actions Runner Controller](https://docs.github.com/en/actions/concepts/runners/actions-runner-controller). GARM and its Incus integration are maintained in the [GARM repository](https://github.com/cloudbase/garm) and [GARM Incus provider](https://github.com/cloudbase/garm-provider-incus).

### 5.2 Chosen rollout

1. Use one dedicated VM for the pilot.
2. Add ARC only when concurrent owner-controlled jobs or scale-to-zero justify operating k3s.
3. Continue sending forks and public PRs to GitHub-hosted runners.
4. Add GARM/Incus when local execution of lower-trust code becomes a real requirement.
5. Do not deploy both autoscalers merely for feature completeness. Each pool must have an identified workload class and owner.

Containers are a reproducibility mechanism and a useful containment layer, but are not treated as the only security boundary for hostile code.

## 6. Runner pools and labels

| Pool | Suggested labels | Workload | Secrets | Lifecycle |
|---|---|---|---|---|
| GitHub hosted | ubuntu-latest or explicit hosted image | Public/fork PRs, outage fallback, clean independent checks | Job-scoped GitHub token only | GitHub-managed fresh environment |
| Pilot local | self-hosted, linux, x64, ci-pilot | Private owner-controlled PRs | No AI/release/deploy secrets | Persistent pilot VM; template rebuild |
| ARC trusted | self-hosted, linux, x64, arc-trusted | Owner-controlled test matrices | No AI/release/deploy secrets | One pod per job |
| Incus isolated | self-hosted, linux, x64, vm-isolated | Lower-trust local code, nested builds | No AI/release/deploy secrets | One disposable VM per job |
| Agent broker | self-hosted, linux, x64, agent-broker | Review and patch generation only | Native Claude/Codex login state | Dedicated host; serialized sessions |
| Release | self-hosted, linux, x64, release-protected | Protected tag or approved release | Short-lived OIDC/signing identity | Fresh VM or reverted snapshot per release |

Repository workflows should reference centrally approved reusable workflows rather than choosing sensitive labels freely. Organization/repository runner groups restrict which repositories may target each pool. GitHub documents labels, groups, and routing in [Using self-hosted runners in a workflow](https://docs.github.com/en/actions/how-tos/manage-runners/self-hosted-runners/use-in-a-workflow).

### 6.1 Runner lifecycle requirements

- Base images are built from version-controlled configuration.
- Toolchains are pinned or captured in a toolchain manifest.
- Each job starts with an empty workspace and ends with secure disposal.
- Ephemeral runner registration is preferred. ARC supplies this lifecycle for its pods.
- Runner update policy is tested in staging before production pools.
- Node images are patched on a defined cadence and immediately for critical exposure.
- Runner processes use dedicated unprivileged accounts.
- No runner account can administer the hypervisor or create arbitrary long-lived infrastructure.
- Logs are redacted and shipped before the VM is destroyed.
- A runner is quarantined, not reused, after unexpected persistence, network-policy violation, or evidence-integrity failure.

## 7. GitHub routing and fallback

### 7.1 Native limitation

GitHub runs a job on a self-hosted runner only when all requested labels match. A label array is an AND condition, not an OR condition, so there is no native expression meaning self-hosted if available, otherwise ubuntu-latest. See [Workflow syntax for runs-on](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#jobsjob_idruns-on).

A self-hosted job that has no matching online runner can remain queued rather than automatically moving to a hosted runner. Fallback must therefore be an explicit platform feature.

### 7.2 Pilot routing

The pilot exposes a `workflow_dispatch` input:

| Input | Values | Behavior |
|---|---|---|
| execution_target | auto, local, hosted | auto selects the policy default; local and hosted provide explicit operator control |
| force_hosted | false, true | Emergency override for deterministic jobs |
| trust_class | Derived, never user-trusted | fork, public, private-owner, protected, release |

The policy is:

| Event | Default runner |
|---|---|
| Public repository PR | GitHub hosted |
| Pull request from a fork | GitHub hosted |
| Private owner-controlled PR | Local ci-pilot when manually enabled; otherwise hosted during pilot |
| Protected-branch nightly | Local |
| Agent review | Agent broker only after deterministic prerequisites |
| Release | Protected release runner or approved hosted release builder |
| Manual force_hosted | GitHub hosted for deterministic jobs only |

Subscription-authenticated agent work never falls back to a hosted runner.

### 7.3 Target automated router

The target adds:

1. A local controller publishes a signed heartbeat containing pool name, capacity, image version, policy version, timestamp, and expiry.
2. A tiny GitHub-hosted router job runs without repository checkout and verifies the heartbeat with a public key committed on the protected default branch.
3. The router combines trust class, required capabilities, heartbeat freshness, queue policy, and explicit operator override.
4. It returns one allowed runner selection to a reusable workflow.
5. An external watchdog observes jobs selected for local execution.
6. If a job does not start within the queue grace period, or ends with a classified infrastructure failure, the watchdog cancels and redispatches it once with hosted selection.

The heartbeat publisher uses a narrowly scoped GitHub App. Its key exists only on the local controller, not on build runners. The router's public verification key is not secret.

### 7.4 Failure classification

| Classification | Examples | Hosted redispatch? |
|---|---|---:|
| Infrastructure | Runner lost, VM provisioning failure, disk failure, heartbeat expired before dispatch | Yes, once, if policy permits |
| Provider availability | GitHub service issue, package registry outage | Not automatically unless the alternate path removes the dependency |
| Semantic | Compile, test, lint, policy, security, or eval failure | No |
| Agent availability | Claude/Codex usage exhausted, login expired, provider unavailable | No paid or hosted fallback |
| Security | Evidence mismatch, unexpected network connection, credential access attempt | No; quarantine and investigate |
| Operator cancellation | Human stopped the run | No unless explicitly redispatched |

The watchdog records both attempts under one logical run ID. It never masks the local failure or treats a successful hosted attempt as proof that the local infrastructure was healthy.

## 8. Network boundaries

### 8.1 Segmentation

Use dedicated networks or VLANs:

- ci-untrusted: build and test workers.
- agent-trusted: agent broker and authentication state.
- release-protected: release builder and deployment gateway.
- services: evidence store, cache proxies, registry mirror, telemetry.
- management: hypervisor and infrastructure administration.
- production: deployed services and runtime data.

The CI, agent, and release identities cannot initiate connections to the management network. Administrative access enters through the existing private remote-access path with MFA and is not exposed through a runner.

~~~mermaid
flowchart LR
    INTERNET["GitHub and approved registries"]
    CI["CI network"]
    SERVICES["Internal services"]
    AGENT["Agent network"]
    RELEASE["Release network"]
    PROD["Production"]
    MGMT["Management"]

    CI --> INTERNET
    CI --> SERVICES
    AGENT --> INTERNET
    AGENT --> SERVICES
    RELEASE --> INTERNET
    RELEASE --> SERVICES
    RELEASE --> PROD
    MGMT --> CI
    MGMT --> AGENT
    MGMT --> RELEASE
~~~

Arrows represent permitted initiation direction. CI, agent, release, and production networks do not initiate management connections.

### 8.2 Flow allowlist

| Source | Destination | Port/protocol | Purpose |
|---|---|---|---|
| Runners | GitHub-required endpoints | TCP 443 | Runner control, Actions downloads, artifacts, checks |
| Build runners | Local package/cache proxies | TCP 443 or service-specific TLS | Dependency restore |
| Build runners | Evidence store | TLS | Append evidence under run-scoped identity |
| Build runners | Approved test endpoints | Explicit TLS allowlist | Contract/e2e tests that truly require external services |
| Agent broker | OpenAI and Anthropic client endpoints | TCP 443 | First-party client login and inference |
| Agent broker | Evidence store | TLS, read-only | Retrieve frozen review inputs |
| Release runner | Registry | TCP 443 | Push signed artifact by digest |
| Release runner | OIDC issuer and target STS | TCP 443 | Obtain short-lived deployment credentials |
| Release runner | Deployment gateway | TLS or SSH with short-lived certificate | Deploy immutable digest |
| All systems | Internal DNS, NTP, telemetry | Restricted service ports | Resolution, time, observability |

Default-deny rules include:

- No runner-to-agent connection.
- No CI-to-production or CI-to-management connection.
- No agent-to-production, registry-push, hypervisor, NAS, or management connection.
- No PR job access to cloud metadata endpoints.
- No general SMB/NFS mount on a runner.
- No inbound Internet forwarding to runner services.

For maximum reproducibility, split dependency acquisition and test execution: restore pinned dependencies through controlled proxies, then disable general egress while compiling and testing. Network-required tests receive an explicit, reviewed allowlist.

## 9. Local-first storage and caching

### 9.1 Storage roles

| Store | Contents | Authority | Backup |
|---|---|---|---|
| Git | Source, specifications, policies, ADRs, regression tests | Canonical project truth | Provided by GitHub plus repository backup policy |
| GitHub artifacts/checks | User-visible summaries and portable evidence needed for hosted fallback | Canonical deterministic workflow evidence during the pilot, retention-limited | GitHub-managed |
| MVP encrypted evidence directory | Content-addressed agent reviews, decisions, and hashes | Canonical agent-decision bundle during MVP; backed up before unattended use | Yes |
| Local S3-compatible evidence store | Full immutable run bundles, review reports, logs, SBOMs, provenance | Canonical detailed run evidence after hardened-v1 migration; optional mirror earlier | Yes |
| Package proxies | Downloaded dependencies by ecosystem | Performance optimization only | No; reconstructable |
| Build cache | Compiler, BuildKit, test, and analysis caches | Performance optimization only | No; reconstructable |
| Artifact registry | Release artifacts and container manifests by digest | Canonical deployable artifact | Yes according to release policy |
| Telemetry store | Metrics and traces | Operational evidence | Optional backup; retain according to observability policy |

GitHub's dependency cache is a GitHub service even when the job runs on a self-hosted runner. It remains useful for hosted fallback, but does not satisfy a purely local cache requirement. See [Caching dependencies to speed up workflows](https://docs.github.com/en/actions/using-workflows/caching-dependencies-to-speed-up-workflows).

### 9.2 Cache policy

Cache keys include:

~~~text
repository / trust-class / operating-system / architecture /
toolchain-version / lockfile-hash / cache-schema-version
~~~

Rules:

- Fork and untrusted jobs never write a cache consumed by release jobs.
- Protected main builds seed trusted caches.
- Release correctness never depends on a cache hit.
- Dependency checksums and lockfiles are verified after restore.
- Cache writers have quota and TTL limits.
- A suspected poisoned cache invalidates its complete trust namespace.
- Cache services are isolated from evidence and artifact-registry credentials.
- Local proxies may include ecosystem-specific services such as devpi, Verdaccio, Athens, or a general repository manager; introduce only those justified by measured traffic.

### 9.3 Suggested retention

| Data | Suggested default |
|---|---:|
| Failed PR detailed evidence | 30 days |
| Successful PR detailed evidence | 14-30 days |
| Protected-branch evidence | 90 days |
| Release SBOM, provenance, signature, decision | Life of release plus organizational retention |
| Agent review and resolution records | 90 days by default; confirmed regression artifacts move to Git. Structured sealed review/cross-review reports and decision records MAY be assigned an indefinite `research` retention class (after secret and personal-data checks) to support longitudinal review-efficacy analysis |
| Operational traces | 14-30 days |
| Build caches | 7-30 days, quota-driven |

Retention is configuration, not policy hardcoding. Legal, customer, or incident-response requirements can supersede these defaults.

### 9.4 Evidence integrity

- Objects are written under a run-scoped prefix and become read-only after finalization.
- The manifest contains a hash and media type for every object.
- Review and release stages validate the manifest before consuming evidence.
- Storage versioning and object lock are enabled when supported.
- Deletion uses a separate retention identity, not a CI runner identity.
- Evidence uploads are spooled locally only long enough to survive a transient store outage; a run cannot be approved until finalization succeeds.

## 10. Credential provisioning

### 10.1 GitHub runner identity

- Runner registration credentials are used only during provisioning.
- Runner directories and service accounts are unique per pool.
- Organization or repository runner groups restrict eligible repositories.
- Workflow tokens receive explicit minimal permissions.
- Checkout uses persist-credentials: false unless a narrowly scoped job must write.
- Third-party Actions are pinned to a full commit SHA in protected workflows.

### 10.2 Agent subscription identity

- Claude Code and Codex are installed and logged in interactively on the dedicated agent VM using first-party clients.
- Unattended Claude work authenticates through a `claude setup-token`-generated one-year OAuth token supplied as `CLAUDE_CODE_OAUTH_TOKEN`. The token is generated on the agent VM itself (the CLI prints it once and does not persist it), stored only in the dedicated account's protected environment, and never transits another machine, GitHub Secrets, or ordinary runners.
- Claude state, the setup-token, and Codex CODEX_HOME are readable only by their dedicated account.
- The agent broker refers to credential locations; it never copies their contents.
- One Codex authentication state is processed serially unless current provider guidance explicitly permits otherwise; the Claude setup-token is serialized under the same rule.
- Setup-token rotation is scheduled ahead of the one-year expiry; login/token-expiry monitoring raises a maintenance alert before unattended work is attempted.
- Provider usage, login, or token failure stops the agent lane; a change to the documented setup-token path fails the lane closed pending review.
- The official Claude and Codex API-key Actions are not used in the subscription-funded path, and `--bare` is not used because it does not read `CLAUDE_CODE_OAUTH_TOKEN`.

OpenAI describes the trusted-private-infrastructure pattern and password-equivalent auth file in [Codex CI/CD authentication](https://developers.openai.com/codex/auth/ci-cd-auth). Anthropic describes Claude Code login and stored credentials in [Identity and access management](https://code.claude.com/docs/en/iam).

### 10.3 Release and deployment identity

Use GitHub Actions OIDC wherever the deployment target or secret broker supports it:

1. The protected deployment job receives id-token: write and only the additional repository permissions it needs.
2. GitHub issues a job-specific OIDC token.
3. The target STS, Vault, or deployment gateway verifies issuer, audience, repository, ref, workflow, actor, and environment claims.
4. It returns a narrowly scoped, short-lived credential.
5. The credential expires after the deployment window and cannot create broader credentials.

GitHub documents this flow in [OpenID Connect in cloud providers](https://docs.github.com/en/actions/concepts/security/openid-connect).

For home infrastructure without a native cloud STS, use a deployment gateway or Vault JWT/OIDC role that validates GitHub's token and issues a short-lived SSH certificate or deployment token. Do not replace OIDC with a long-lived root SSH key in GitHub Secrets.

Runtime application secrets are delivered by the target environment's secret manager. They are not baked into images, SBOMs, provenance, caches, or workflow artifacts.

## 11. Release supply chain

### 11.1 Build and promotion flow

~~~mermaid
sequenceDiagram
    participant G as GitHub
    participant B as Release builder
    participant R as Registry
    participant D as Deploy controller
    participant T as Target

    G->>B: Protected commit and approval
    B->>B: Clean build, tests, SBOM, scan
    B->>B: Provenance and signature
    B->>R: Push immutable artifact and attestations
    R-->>D: Verified digest
    D->>T: Deploy digest to staging
    T-->>D: Health and smoke results
    D->>T: Promote same digest to production
    T-->>G: Deployment status and evidence
~~~

### 11.2 Release stages

1. Verify the protected commit and required checks.
2. Start a clean release environment.
3. Restore pinned dependencies from trusted namespaces.
4. Run the release test subset and policy checks.
5. Build the artifact exactly once.
6. Generate a CycloneDX or SPDX SBOM with a tool such as [Syft](https://github.com/anchore/syft).
7. Scan the artifact and dependencies with one selected primary scanner, such as Grype or Trivy, using a recorded database version.
8. Produce an in-toto provenance statement containing source, builder, invocation, dependency, output digest, and timestamps.
9. Sign the artifact digest and provenance with [Cosign](https://docs.sigstore.dev/cosign/signing/overview/).
10. Push artifact, SBOM, provenance, scan result, and signature to the registry/evidence store.
11. Verify them independently before deployment.
12. Deploy the immutable digest to staging.
13. Run health, smoke, migration-compatibility, and rollback-readiness checks.
14. Obtain required protected-environment approval.
15. Promote the same digest to production.
16. Monitor the defined observation window and automatically or manually roll back on policy breach.

GitHub artifact attestations can provide integrated provenance when the repository and plan are eligible; availability differs for public and private repositories. The portable baseline is self-managed in-toto provenance plus Cosign. See [GitHub artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations).

### 11.3 Provenance claims

The platform records useful provenance but makes no unsupported assurance claim. A home-server administrator can alter a builder, image, clock, network, or provenance generator. Stronger SLSA build levels require controls beyond merely emitting a provenance file. The applicable requirements are defined by [SLSA build track basics](https://slsa.dev/spec/v1.2/build-track-basics).

For releases needing stronger independent trust, route the clean release build to a qualifying hosted builder while leaving ordinary CI local.

## 12. Deployment environments

| Environment | Trigger | Approval | Deployment method | Rollback |
|---|---|---|---|---|
| Development | Merge to development branch or manual | Automated policy | Deploy digest to disposable/test namespace | Replace with previous known-good or destroy |
| Staging | Release candidate | Automated checks; optional owner approval | Deploy exact signed digest | Redeploy previous staging digest |
| Production | Protected environment after staging | Required human or policy owner | Promote exact staging digest | Redeploy recorded previous production digest |

GitHub environments should define protection rules, allowed branches/tags, approvers, and environment-scoped configuration. See [Managing environments for deployment](https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments).

### 12.1 Deployment strategies

Choose per service:

| Strategy | Use when | Rollback behavior |
|---|---|---|
| Rolling | Stateless services with robust readiness checks | Stop rollout and restore prior digest |
| Blue/green | Fast switch and near-zero downtime justify duplicate capacity | Route traffic back to previous color |
| Canary | Production signals are required before full rollout | Reduce canary to zero and restore prior routing |
| Recreate | Small internal service tolerates interruption | Start previous digest |

The strategy is part of the service's deployment specification. The pipeline must not infer it from agent output.

### 12.2 Database migrations

- Prefer backward-compatible expand/migrate/contract changes.
- Run pre-deployment compatibility checks.
- Back up or snapshot according to the application's recovery plan.
- A deployment cannot assume that rolling back application code reverses a destructive schema migration.
- Destructive migrations require explicit approval, a verified restore path, and a maintenance/traffic plan.

### 12.3 Rollback record

For each environment, store:

- Current and prior artifact digest.
- Configuration revision.
- Migration revision and compatibility status.
- Deployment run ID and approver.
- Health-check result and observation window.
- Exact rollback command or controller operation.

Rollback uses an already verified digest; it does not rebuild an old Git commit.

## 13. Observability

### 13.1 Required metrics

| Area | Metrics |
|---|---|
| Runner platform | Online capacity, provisioning latency, queue latency, job duration, CPU, memory, disk, image version |
| Routing | Local/hosted selections, stale heartbeats, manual overrides, watchdog redispatches |
| Quality | Failure category, flaky rate, coverage delta, security finding age, eval pass rate |
| Cache | Hit rate, bytes served, restore time, eviction, trust namespace |
| Agents | Availability, turns, elapsed time, review findings, accepted/rejected ratio, repair cycles, no-progress stops |
| Supply chain | SBOM generation, scan database age, signature verification, provenance validation |
| Deployment | Lead time, change failure rate, health-check latency, rollback count, recovery time |
| Storage | Evidence finalization latency, quota, object-lock failures, retention deletions |

Propagate one logical run ID from router through deterministic execution, evidence, agent review, release, and deployment. OpenTelemetry is the interchange format; a local Phoenix deployment may provide initial trace and experiment exploration.

### 13.2 Alerts

Alert on:

- No eligible runner beyond the queue threshold.
- Heartbeat signature or freshness failure.
- Repeated infrastructure redispatch.
- Agent authentication nearing expiry or unavailable.
- Evidence finalization/hash failure.
- Storage or runner disk above threshold.
- Security policy or signature validation failure.
- Deployment health degradation or rollback failure.

GitHub check summaries remain the developer-facing status. Infrastructure alerts go to the platform owner through the selected notification channel.

## 14. Backup and disaster recovery

### 14.1 Back up

- Infrastructure-as-code and runner image definitions.
- GitHub App configuration and encrypted private-key recovery material.
- Evidence-store metadata and release evidence according to retention.
- Artifact registry metadata and release artifacts not reproducible from another trusted registry.
- Deployment manifests and environment state records.
- Signing-key recovery material if key-based signing is used.
- Telemetry configuration and dashboards.

Do not spend backup capacity on reconstructable compiler caches or disposable runner disks.

### 14.2 Recovery objectives

Suggested pilot objectives:

| Service | RPO | RTO |
|---|---:|---:|
| Git repository | Governed by GitHub and repository backup | 4 hours |
| Evidence store | 24 hours for non-release evidence; zero loss desired for finalized release evidence | 8 hours |
| Runner capacity | No durable state | 4 hours or immediate hosted fallback |
| Agent broker | Credential/login state may require manual recovery | 1 business day |
| Artifact registry | Zero loss for active release digests | 4 hours |
| Deployment controller state | Last successful deployment record | 4 hours |

Test restoration quarterly and before materially changing storage, hypervisor, or signing architecture.

## 15. Failure-mode runbook

| Symptom | Immediate response | Recovery | Prevention/verification |
|---|---|---|---|
| Local pool offline | Route permitted deterministic jobs to hosted | Repair controller/host; verify image and heartbeat | Test forced outage monthly |
| Job queued despite healthy host | Inspect label/group mismatch and capacity | Correct routing metadata; redispatch logical run once | Contract-test runner labels |
| Runner lost mid-job | Mark infrastructure failure and destroy workspace | Hosted redispatch if allowed | Monitor host pressure and runner heartbeat |
| Persistent runner suspected compromised | Disable runner and network access | Rebuild from known image; rotate runner identity | Move workload to ephemeral pool |
| Cache poisoning suspected | Disable namespace | Delete cache and rebuild from lockfiles | Trust-separated namespaces |
| Evidence hash mismatch | Quarantine run | Reproduce from clean checkout | Hash at producer and consumer |
| Evidence store full | Stop finalization and releases | Extend capacity or expire approved data | Quotas and advance alerts |
| Claude/Codex login expired | Pause agent lane | Interactive first-party login | Expiry maintenance alert |
| Subscription usage exhausted | Pause agent lane | Wait for plan reset or request human review | Budgets and two-cycle cap |
| Provider outage | Preserve pending evidence | Resume later; do not switch to paid API | Agent lane is non-authoritative for deterministic CI |
| OIDC denied | Stop deployment | Inspect claim policy and environment/ref | Policy tests in staging |
| SBOM/scan/signing failure | Block publication | Repair tooling/database/identity and rebuild cleanly | Pin and monitor supply-chain tools |
| Staging health failure | Stop promotion | Roll back staging digest | Representative smoke tests |
| Production regression | Freeze promotion | Route/redeploy prior digest | Canary/blue-green and observation window |
| Rollback failure | Declare incident | Execute service recovery plan | Quarterly rollback and restore exercise |

## 16. Configuration variables

Names are illustrative and should be centralized in protected repository or organization configuration. Secrets are explicitly excluded from ordinary variables.

| Variable | Example/default | Sensitivity | Purpose |
|---|---|---:|---|
| CI_EXECUTION_TARGET | auto | Normal | auto, local, or hosted selection |
| CI_TRUST_CLASS | Derived | Normal | fork, public, private-owner, protected, release |
| LOCAL_RUNNER_LABEL | ci-pilot | Normal | Pilot local selector |
| ISOLATED_RUNNER_LABEL | vm-isolated | Normal | GARM/Incus selector |
| ARC_RUNNER_LABEL | arc-trusted | Normal | ARC selector |
| HOSTED_RUNNER_IMAGE | ubuntu-latest | Normal | Hosted fallback image |
| RUNNER_HEARTBEAT_TTL_SECONDS | 120 | Normal | Maximum accepted heartbeat age |
| LOCAL_QUEUE_GRACE_SECONDS | 300 | Normal | Time before watchdog considers redispatch |
| INFRA_RETRY_LIMIT | 1 | Normal | Maximum automatic infrastructure redispatch |
| AGENT_REVIEW_MODE | blind-cross-review | Normal | Agent protocol |
| AGENT_MAX_REPAIR_ROUNDS | 2 | Normal | Loop limit |
| AGENT_TIMEOUT_MINUTES | 30 | Normal | Per-agent wall-clock limit |
| CODEX_HOME | Dedicated local path | Sensitive path, not content | Codex state location on broker |
| CLAUDE_CONFIG_DIR | Dedicated local path | Sensitive path, not content | Claude state location on broker |
| EVIDENCE_ENDPOINT | Internal TLS URL | Internal | Local object store |
| EVIDENCE_BUCKET | agentic-ci-evidence | Normal | Evidence namespace |
| EVIDENCE_RETENTION_DAYS | 30 | Normal | Default non-release retention |
| CACHE_SCHEMA_VERSION | v1 | Normal | Cache invalidation boundary |
| CACHE_TRUST_NAMESPACE | Derived | Normal | Separates fork, PR, main, release |
| OTEL_EXPORTER_OTLP_ENDPOINT | Internal TLS URL | Internal | Telemetry destination |
| ARTIFACT_REGISTRY | ghcr.io/owner | Normal | Release registry |
| RELEASE_ENVIRONMENT | staging or production | Normal | GitHub protected environment |
| DEPLOY_OBSERVATION_SECONDS | Service-specific | Normal | Post-deploy health window |
| PREVIOUS_GOOD_DIGEST | Controller-managed | Protected | Rollback artifact |

Secrets and private keys belong in their dedicated systems:

- Claude/Codex login state and the `claude setup-token` OAuth token (`CLAUDE_CODE_OAUTH_TOKEN`): agent VM only.
- GitHub App private key: local router/watchdog controller only.
- OIDC trust configuration: deployment target, STS, Vault, or gateway.
- Runtime secrets: deployment environment secret manager.
- Signing keys, if not keyless: KMS/HSM or protected signing service.

## 17. Infrastructure work packages

Document 07 is authoritative for project phase and gate numbering. The infrastructure work packages below map to it and are not a second delivery-phase system.

### I1 — Safe local pilot foundation (Roadmap P3)

- Create segmented CI, agent, release, and service networks.
- Build three VM templates.
- Register ci-pilot in a restricted runner group.
- Deploy evidence storage and basic quotas.
- Implement manual hosted fallback.
- Verify no AI or deployment credentials exist on ci-pilot.

### I2 — Standard deterministic pipeline support (Roadmap P1–P2)

- Connect ci/run and language adapters.
- Publish JUnit, SARIF, coverage, logs, and manifests.
- Add cache trust namespaces.
- Pin Actions and enforce minimal workflow permissions.
- Exercise local and hosted runs against the same commit.

### I3 — Agent lane (Roadmap P4–P5)

- Install and authenticate first-party clients in dedicated local OS accounts on agent-pilot.
- Restrict evidence input to read-only.
- Enforce schemas, timeouts, serialized auth use, and repair limits.
- Ensure patch execution occurs only on credential-free CI.
- Test login expiry, usage exhaustion, and provider outage.

### I4 — Ephemeral runners (Roadmap P6)

- Add ARC for measured trusted workload demand.
- Add GARM/Incus only for workload classes requiring stronger local isolation.
- Replace persistent ci-pilot use with disposable capacity.
- Add image pipeline, patch cadence, and quarantine automation.

### I5 — Automated fallback (Roadmap P6)

- Deploy signed heartbeat publisher.
- Implement no-checkout hosted router.
- Deploy external queue watchdog.
- Test stale heartbeat, offline host, capacity exhaustion, runner loss, and semantic failure classification.

### I6 — Release and deployment (Roadmap P7)

- Deploy protected release runner.
- Add SBOM, selected vulnerability scanner, provenance, and Cosign.
- Configure OIDC trust and protected environments.
- Implement staging, smoke tests, promotion by digest, observation, and rollback.
- Run a full disaster-recovery and rollback exercise.

## 18. Infrastructure acceptance tests

Before declaring hardened v1 operational:

- Disconnect the home server and verify deterministic jobs have the documented hosted/manual path.
- Submit a fork-like test and prove it cannot select a self-hosted persistent runner.
- Attempt access from CI to the agent, management, NAS, and production networks; all must fail.
- Verify the agent host cannot deploy or push to the release registry.
- Run the same commit locally and hosted and compare normalized evidence.
- Corrupt one evidence object and verify review/release fail closed.
- Expire or remove one provider login and verify there is no API fallback.
- Poison an untrusted cache namespace and verify release cannot consume it.
- Rebuild a runner from its template.
- Issue an OIDC deployment credential and verify scope and expiry.
- Verify a signed artifact, SBOM, and provenance independently.
- Deploy to staging by digest, fail a health check, and demonstrate rollback.
- Restore the current production digest and release evidence from backup.

## 19. Primary references

- [GitHub self-hosted runners](https://docs.github.com/en/actions/reference/runners/self-hosted-runners)
- [GitHub Actions secure use reference](https://docs.github.com/en/actions/reference/security/secure-use)
- [Actions Runner Controller](https://docs.github.com/en/actions/concepts/runners/actions-runner-controller)
- [Workflow syntax for runs-on and job containers](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [Using self-hosted runners in a workflow](https://docs.github.com/en/actions/how-tos/manage-runners/self-hosted-runners/use-in-a-workflow)
- [GitHub Actions dependency caching](https://docs.github.com/en/actions/using-workflows/caching-dependencies-to-speed-up-workflows)
- [GitHub Actions OIDC](https://docs.github.com/en/actions/concepts/security/openid-connect)
- [GitHub artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations)
- [GitHub deployment environments](https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments)
- [GARM](https://github.com/cloudbase/garm)
- [GARM provider for Incus](https://github.com/cloudbase/garm-provider-incus)
- [OpenAI Codex CI/CD authentication](https://developers.openai.com/codex/auth/ci-cd-auth)
- [Anthropic Claude Code authentication](https://code.claude.com/docs/en/iam)
- [Sigstore Cosign signing overview](https://docs.sigstore.dev/cosign/signing/overview/)
- [Syft SBOM generator](https://github.com/anchore/syft)
- [SLSA build track basics](https://slsa.dev/spec/v1.2/build-track-basics)
