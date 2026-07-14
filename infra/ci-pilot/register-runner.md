# Registering the ci-pilot self-hosted runner

Runs INSIDE the ci-pilot VM after `harden.sh`. Registers a repository-scoped, **ephemeral** GitHub Actions runner with the trust labels from doc 06 §6.

## Why repository-scoped and ephemeral

- **Repository-scoped** (not org/account-wide): only `Aakash1337/CustomDNS` can target it. A different repo — including a public one — cannot route jobs here (FR-RUN-001, PAC-010).
- **`--ephemeral`**: the runner processes exactly one job, then de-registers. Combined with a per-job clean workspace this approximates the disposable-runner target (ADR-004) even before ARC/Incus, and closes the persistence risk (R-002).

## Steps

```bash
# On the host: mint a short-lived registration token (never stored in the VM image).
gh api -X POST repos/Aakash1337/CustomDNS/actions/runners/registration-token --jq .token
# Copy the token into the VM for the next command (expires in ~1 hour).

# Inside the VM, as the runner user:
cd /home/runner/actions-runner
RUNNER_VERSION=2.330.0   # pin; update via reviewed change
curl -fsSL -o runner.tar.gz \
  "https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz"
tar xzf runner.tar.gz

./config.sh \
  --url https://github.com/Aakash1337/CustomDNS \
  --token "<registration-token>" \
  --name ci-pilot \
  --labels self-hosted,linux,x64,ci-pilot \
  --work _work \
  --ephemeral \
  --unattended \
  --replace

# Run once per job under a loop supervisor. For the pilot, a systemd
# service restarts the ephemeral runner after each job:
sudo ./svc.sh install runner
sudo ./svc.sh start
```

## Credential absence (verified by reachability-test.sh)

This VM holds **no** Claude, Codex, deployment, signing, or NAS credentials (FR-RUN-002). The only secret present is the GitHub runner registration/auth state, which is job-scoped and cannot push, merge, or deploy. The reachability test asserts the credential-absence invariant on every drill.

## De-registration

```bash
./config.sh remove --token "$(gh api -X POST repos/Aakash1337/CustomDNS/actions/runners/remove-token --jq .token)"
```
