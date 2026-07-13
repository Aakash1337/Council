# Operations Runbook

**Status:** Proposed  
**Audience:** Platform owner, repository maintainer, security owner, and release owner  
**Scope:** GitHub control plane, home runner infrastructure, deterministic CI, trusted agent broker, evidence storage, and deployment workflows

## 1. Operating objectives

The platform should fail closed for quality, credentials, and deployment integrity while remaining recoverable when local infrastructure is unavailable. The normal recovery path must distinguish an infrastructure interruption from a real code, test, security, or evaluation failure.

Primary operating objectives:

- No personal AI or deployment credential is exposed to pull-request code.
- Required checks accurately describe the candidate commit evaluated.
- A local outage does not prevent authorized hosted deterministic execution.
- Agent loops terminate on budget, timeout, no progress, or iteration limit.
- Release promotion uses the exact artifact digest that passed policy.
- Every policy exception is owned, visible, and time-bounded.

## 2. Roles and escalation

| Role | Responsibilities | Escalates to |
|---|---|---|
| Platform owner | Runner/controller health, storage, routing, upgrades, backups | Security or project owner |
| Repository maintainer | Specifications, adapters, tests, branch rules, triage | Project owner |
| Security owner | Credential incidents, high-severity findings, waivers, trust policy | Project owner |
| Release owner | Artifact approval, environment protection, rollback | Project owner/security owner |
| Project owner | Priority, scope, residual-risk acceptance | Human decision authority |

Names and contact methods are deployment-specific and MUST be filled in before unattended operation.

## 3. Severity model

| Severity | Definition | Initial response target |
|---|---|---:|
| SEV-1 | Confirmed credential disclosure, runner escape into protected network, unauthorized production deployment, or artifact-integrity failure | Stop affected systems immediately; acknowledge within 15 minutes when attended |
| SEV-2 | Persistent inability to enforce required gates, suspected runner compromise, release pipeline unavailable, or incorrect production behavior requiring rollback | Acknowledge within 1 hour |
| SEV-3 | Local runner outage with hosted recovery, agent service outage, evidence-store degradation, or recurring flaky infrastructure | Acknowledge within 1 business day |
| SEV-4 | Nonblocking warning, documentation defect, capacity trend, or minor tooling issue | Add to backlog |

For a personal/home system, targets describe operating priority rather than a staffed 24/7 SLA.

## 4. Routine operating checklist

### Daily or before active development

- Confirm the local runner/controller is online if local execution is expected.
- Confirm disk usage, memory pressure, and object-store health are within limits.
- Confirm the signed health heartbeat is current once automatic routing is enabled.
- Review failed required checks and infrastructure-error classifications.
- Check whether Claude or Codex reports authentication expiry or usage limits.
- Verify no unexpected runner is registered to the repository or organization.

### Weekly

- Review queued and abandoned jobs.
- Review high/critical security findings and waiver expirations.
- Inspect agent time, turn, false-positive, and repair-loop metrics.
- Confirm backups completed and one recent backup is readable.
- Check for runner/controller, GitHub Action, scanner, and quality-tools updates.
- Review storage retention and remove expired nonessential evidence through approved lifecycle policy.
- Inspect cache size and trust namespaces.

### Monthly

- Restore one representative evidence bundle from backup.
- Exercise a hosted fallback run.
- Verify branch rules, required checks, CODEOWNERS, and environment reviewers.
- Verify third-party Actions remain pinned to reviewed SHAs.
- Review authorized egress destinations and firewall rules.
- Review repository and runner access membership.
- Review metrics and adjust capacity or time budgets.
- Test an artifact-signature and provenance verification path.

### Quarterly

- Conduct an offline-runner drill and a failed-deployment rollback drill.
- Reassess the threat model and risk register.
- Review provider authentication and subscription terms for changed behavior; verify the documented `claude setup-token` path is unchanged and rotate the token ahead of its one-year expiry (fail the agent lane closed if the documented path changes).
- Review the quality/security standard and language adapters.
- Calibrate agent reviewers and any model graders against human-labeled cases.
- Rotate long-lived credentials that cannot yet be replaced by workload identity.

## 5. Normal pull-request operation

1. Verify the change specification and manifest are present and schema-valid.
2. Confirm the candidate SHA shown by GitHub matches the run manifest.
3. Let deterministic CI finish before invoking model review.
4. Classify failures:
   - `quality_failure`
   - `security_failure`
   - `policy_failure`
   - `infrastructure_error`
   - `agent_unavailable`
5. Do not rerun the first three merely to obtain a different result.
6. For an infrastructure error, correct or confirm the transient condition and redispatch once.
7. If deterministic CI passes, begin independent reviews.
8. Confirm both initial review files validate and were sealed before cross-review.
9. Verify the final decision lists all applied evidence and waivers.
10. Merge only through protected rules.

## 6. Failure classification rules

| Symptom | Classification | Retry policy |
|---|---|---|
| Compiler, linter, test, contract, or policy returns a normal failure | Quality/policy failure | No automatic retry |
| Secret, SAST, SCA, or artifact policy finds a violation | Security failure | No automatic retry |
| Registry/DNS/network timeout before the check executes | Infrastructure error | Bounded retry allowed |
| Runner disconnects mid-job | Infrastructure error unless evidence indicates workload crash | Cancel and redispatch once |
| Model provider rate limit or native client authentication expiry | Agent unavailable | Pause or bounded retry after backoff/reauthentication |
| Model returns a valid but unfavorable finding | Review result | Never retry solely to seek a pass |
| Model output violates JSON schema | Agent protocol failure | One constrained regeneration allowed; then human/unavailable |
| Test passes only on second execution without a known infrastructure cause | Suspected flake | Original failure remains visible; quarantine process required |

## 7. Local runner offline before dispatch

### Detection

- Manual observation during the pilot.
- Later: signed heartbeat missing or expired.
- GitHub runner API status may be used as supporting information but is not sufficient for scale-to-zero health.

### Response

1. Confirm whether the outage is planned.
2. If deterministic work is urgent, rerun with `force_hosted=true`.
3. Do not transfer AI subscription credentials to the hosted runner.
4. Keep the agent-review stage pending or perform it interactively on the trusted host after recovery.
5. Inspect hypervisor, VM, controller, disk, network, and GitHub registration state.
6. Restore the local runner and execute a canary workflow.
7. Clear the forced-hosted override after the canary succeeds.

### Close criteria

- Runner/controller healthy.
- Canary completes.
- No stale registration or duplicate runner remains.
- Root cause and duration recorded for recurring outages.

## 8. Runner failure after job assignment

GitHub does not live-migrate an assigned self-hosted job.

1. Wait only for the configured disconnect grace period.
2. Cancel the abandoned run.
3. Verify no side effect—such as package publication or deployment—occurred before failure.
4. Classify the failure as infrastructure-related.
5. Redispatch to hosted or another disposable local runner.
6. If the job was not idempotent, require human review before redispatch.
7. Remove the stale runner registration if it does not clean itself up.

Never redispatch a normal test failure under the infrastructure-fallback mechanism.

## 9. Suspected runner compromise

Treat unexpected processes, persistence, unknown network connections, modified runner binaries, cache anomalies, secret-access attempts, or lateral movement as compromise indicators.

### Immediate containment

1. Disable routing to the runner pool.
2. Isolate the VM, pod, or host at the network layer; do not merely stop the runner service.
3. Revoke the runner registration token/session and remove its GitHub registration.
4. Stop releases and invalidate affected caches.
5. Preserve logs, snapshots, run IDs, commit hashes, and network metadata.
6. Determine whether any credentialed or release workload used the same trust boundary.

### Eradication and recovery

1. Rebuild from a known-good template; do not clean and reuse the compromised instance.
2. Rotate any credential that may have been reachable.
3. Rebuild affected artifacts on a clean independent runner.
4. Compare provenance and artifact digests.
5. Review runner-network rules and the triggering workflow.
6. Re-enable only after a clean canary and security-owner approval.

### Escalation

Any suspected access to AI credentials, repository write credentials, signing material, deployment identity, NAS data, or hypervisor administration is SEV-1.

## 10. AI credential disclosure or suspected access

1. Stop the trusted agent broker and revoke its ability to accept new work.
2. Revoke/log out the affected Claude or Codex session using the provider's supported mechanism.
3. Remove exposed credentials from logs and artifacts after preserving restricted incident evidence.
4. Search repository history, Actions logs, caches, evidence storage, and runner workspaces for the secret.
5. Reauthenticate on a rebuilt or verified trusted agent host.
6. Audit all agent runs since the earliest possible disclosure.
7. Add or strengthen a deterministic secret-detection rule.
8. Record the incident and remediation without storing the credential itself.

Never paste provider credential-file contents into an issue, PR, chat, or incident document.

## 11. Claude or Codex login or token expired

### Symptoms

- Native client requests interactive login.
- Headless execution stops with an authentication error.
- The `CLAUDE_CODE_OAUTH_TOKEN` setup-token is expired, revoked, or rejected.
- Review jobs remain queued or return `agent_unavailable`.

### Response

1. Mark the agent lane unavailable; do not fail deterministic CI.
2. For interactive login state, reauthenticate interactively on the trusted agent host using the official client. For an expired or revoked setup-token, regenerate with `claude setup-token` on the agent host itself; the token prints once and is not persisted by the client, so place it directly into the dedicated account's protected environment.
3. Confirm credential-file and token permissions and ensure no copy was created elsewhere; the token must not appear in shell history, GitHub Secrets, ordinary runners, artifacts, or logs.
4. Execute a read-only canary against a non-sensitive fixture repository.
5. Resume one queued job and confirm structured output.
6. Update the rotation record: verification date and next review date ahead of the one-year expiry.

Do not replace the subscription login with an API key unless the project owner has explicitly approved the separate spend and secret-management model.

## 12. Subscription limit or provider rate limit

1. Record the provider, time, operation, and native error category without storing sensitive response data.
2. Stop new jobs for that provider and retain queue order.
3. Apply provider-appropriate backoff.
4. Allow deterministic CI and the other provider's independent review to complete.
5. Do not treat a single-provider review as two-agent consensus.
6. Offer a human-review path for urgent work.
7. Review whether unnecessary dual-writing, excessive context, or repeated unchanged prompts caused the usage.

Budgets should favor one writer plus two reviewers. Competing implementations are reserved for high-risk changes.

## 13. Agent loop stuck or making no progress

Terminate a loop when any condition occurs:

- Maximum wall-clock time reached.
- Maximum turns reached.
- Two repair cycles completed.
- The same finding reappears without materially new evidence.
- Patch hash does not change after a requested repair.
- The deterministic failure signature is unchanged.
- Review output repeatedly fails schema validation.
- Provider usage limit or authentication failure occurs.

On termination:

1. Preserve the last valid evidence and decision.
2. Mark the policy conclusion `pending` with review coverage `human_required`, rather than failed or passed.
3. Summarize unresolved finding IDs and exact deterministic failures.
4. Stop all child/background agent tasks.
5. Release worktree and queue leases after artifacts are committed.

## 14. Prompt-injection or prohibited tool request

Examples include repository text instructing a reviewer to ignore policy, read credential files, contact an external host, modify hidden tests, or approve the change.

1. Deny the requested operation.
2. Record the source file/path and requested capability as a security finding.
3. Stop the agent run if a protected resource was accessed or attempted repeatedly.
4. Confirm network, filesystem, and command controls prevented the action.
5. Review the repository content to determine whether it was malicious or an accidental test fixture.
6. Add a regression case to the agent-integrity evaluation set.
7. Escalate to the security owner for high-risk attempts.

Agent instructions inside the candidate repository never outrank platform policy or the protected review prompt.

## 15. Malformed or conflicting evidence

1. Recompute hashes for the specification, candidate, patch, and evidence files.
2. Confirm every report targets the same candidate SHA.
3. Reject evidence produced for another revision.
4. Regenerate only the malformed evidence source; do not selectively rerun semantic failures.
5. If evidence was modified after sealing, invalidate the entire decision and investigate.
6. Re-run the deterministic policy engine from the complete immutable bundle.

## 16. Scanner database changed

Vulnerability and advisory checks are time-dependent.

- PR comparisons SHOULD record a known database version or timestamp.
- Nightly and release scans SHOULD use current data.
- A new release-time finding must be evaluated even if the PR scan passed.
- A release may proceed only under the current policy or an approved expiring waiver.
- The evidence bundle must show which database informed the decision.

## 17. Flaky test handling

1. Preserve the first failure, logs, seed, environment, and candidate SHA.
2. Determine whether an external dependency or runner failure caused it.
3. If the test itself is flaky, create an issue with owner and target date.
4. Quarantine only if the test is noncritical and an authorized maintainer approves.
5. Exclude quarantined tests from the passing count and report them separately.
6. Add expiry to the quarantine.
7. Restore the test as blocking after the fix and verify repeated stability.

Security, authorization, data-integrity, migration, and rollback tests SHOULD NOT be quarantined merely for convenience.

## 18. Policy waiver procedure

1. Open a waiver using the standard template.
2. Identify the exact finding or control and affected revision/artifact.
3. Document business need, risk, compensating controls, owner, approver, issue, and expiry.
4. Obtain security-owner approval for security controls and release-owner approval for release controls.
5. Add the waiver as structured input to the policy decision.
6. Confirm it does not apply beyond its explicit scope.
7. Alert before expiry.
8. Remove the waiver after remediation; do not silently extend it.

Agents may recommend remediation or draft a waiver but cannot approve it.

## 19. Cache poisoning or corruption

1. Disable the affected cache namespace.
2. Rebuild from a clean environment with caching disabled.
3. Compare dependency locks, toolchain versions, outputs, and artifact digests.
4. Determine whether untrusted work could write a cache later read by trusted or release jobs.
5. Purge the namespace and rotate cache keys.
6. Repair trust-level separation before re-enabling writes.

A release must remain buildable without a mutable cache.

## 20. Evidence-store outage

1. Stop final policy decisions that require unavailable evidence.
2. Continue checks only if they can safely retain local temporary outputs within capacity.
3. Do not substitute incomplete GitHub summaries for the required evidence bundle.
4. Restore the object store or fail over to an approved target.
5. Verify object hashes and index consistency.
6. Resume policy evaluation from the complete bundle.

## 21. Backup and recovery

Back up:

- Platform configuration and infrastructure-as-code.
- Policy, schema, and quality-tool definitions.
- Object-store evidence within approved retention.
- Phoenix or other trace/dataset metadata where it is not reproducible.
- Signing configuration, without placing private signing material in ordinary backups.
- Operational dashboards and alert definitions.

Do not back up disposable runner filesystems or provider credential files into general-purpose storage.

Recovery order:

1. GitHub repository and protected configuration.
2. Runner/controller infrastructure.
3. Policy and evidence schemas.
4. Object store and indexes.
5. Observability.
6. Trusted agent host, followed by interactive reauthentication.
7. Release/deployment capability.

## 22. Release procedure

1. Confirm the source revision is protected and all required gates passed.
2. Confirm no waiver is expired or broader than the release scope.
3. Build in a clean release environment.
4. Record the immutable artifact digest.
5. Generate the SBOM from that final artifact.
6. Scan the artifact with current data.
7. Sign the artifact, SBOM, and required attestations.
8. Generate provenance linking source, workflow, builder, and digest.
9. Evaluate release policy.
10. Publish to the registry using an immutable reference.
11. Deploy the same digest to staging.
12. Run smoke, compatibility, and migration checks.
13. Obtain production approval.
14. Promote the same digest.
15. Verify production health and record the deployment.

## 23. Failed deployment and rollback

### Rollback triggers

- Smoke or health check fails.
- Error rate, latency, or resource consumption exceeds the declared budget.
- Data migration fails or compatibility invariant is violated.
- Signature/provenance verification fails at admission.
- A confirmed critical security issue is discovered.

### Rollback procedure

1. Stop further rollout or traffic increase.
2. Capture deployment and application evidence.
3. Revert traffic to the last known-good immutable digest.
4. Execute the documented data rollback or forward-fix strategy; do not reverse an irreversible migration blindly.
5. Verify health and critical transactions.
6. Mark the failed release as non-promotable.
7. Open an incident and link commit, digest, deployment, metrics, and logs.
8. Require a new candidate and full policy evaluation for the fix.

## 24. Upgrade procedure

Applies to runners, ARC/GARM, native agent clients, broker, scanners, quality image, policy engine, object store, and observability services.

1. Read release notes and security advisories.
2. Pin the proposed version or digest.
3. Test in a nonproduction runner group or fixture repository.
4. Exercise success, deterministic failure, timeout, and cleanup paths.
5. Compare evidence formats for breaking changes.
6. Upgrade one component/pool at a time.
7. Retain a tested rollback version.
8. Update the version inventory and relevant ADR.

Native Claude or Codex changes that affect authentication, permissions, sandboxing, output schema, or noninteractive behavior require an agent-integrity regression run before broad adoption.

## 25. Capacity management

Track:

- Queue duration by job class.
- Runner startup and teardown time.
- CPU, memory, disk I/O, and network saturation.
- Storage growth and retention effectiveness.
- Cache hit rate and cache size.
- Hosted fallback frequency and minutes.
- Agent job duration, turns, concurrency, and throttling.

Capacity changes should preserve trust separation. Do not solve queue pressure by allowing untrusted builds onto credentialed or release hosts.

## 26. Audit checklist

- [ ] Default branch protection enabled.
- [ ] Required status checks current and unbypassable by agents.
- [ ] CODEOWNERS protects workflows, policy, evals, and deployment.
- [ ] Third-party Actions pinned to full SHAs.
- [ ] Job permissions minimal and explicit.
- [ ] Public/fork code excluded from persistent home runners.
- [ ] AI credential host separate from build/test runners.
- [ ] Release pool separate from PR execution.
- [ ] Network segmentation verified.
- [ ] No host Docker socket exposed to PR jobs.
- [ ] Evidence hashes and retention policy active.
- [ ] Cache namespaces separated by trust level.
- [ ] Waivers have owners and expiries.
- [ ] Backups and restore test current.
- [ ] Hosted fallback and rollback drills current.
- [ ] Artifact signature and provenance verification tested.

## 27. Incident record minimum fields

```text
Incident ID:
Severity:
Opened/closed timestamps:
Owner:
Affected repositories/runners/environments:
Base and candidate SHA:
Run IDs:
Artifact digests:
Detection source:
Observed impact:
Containment actions:
Credential exposure assessment:
Evidence locations and hashes:
Root cause:
Corrective actions:
Regression tests added:
Risk/ADR updates:
Approvals to restore service:
```

## 28. Reference procedures

- [GitHub self-hosted runners](https://docs.github.com/en/actions/reference/runners/self-hosted-runners)
- [GitHub secure use](https://docs.github.com/en/actions/reference/security/secure-use)
- [Codex trusted private CI authentication](https://developers.openai.com/codex/auth/ci-cd-auth)
- [Claude Code authentication](https://code.claude.com/docs/en/iam)
- [Cosign verification](https://docs.sigstore.dev/quickstart/quickstart-cosign/)
- [GitHub OIDC](https://docs.github.com/en/actions/concepts/security/openid-connect)
