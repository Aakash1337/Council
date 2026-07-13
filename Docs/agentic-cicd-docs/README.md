# Council — Agentic CI/CD Platform

Project documentation for **Council**, a local-first, specification-driven CI/CD system in which coding agents can implement, independently review, challenge, and repair changes while deterministic checks remain the final authority. The name reflects the core review model: independent agents deliberate over the same sealed evidence, but the deterministic policy engine — not the council — holds final authority.

**Document status:** Proposed baseline  
**Version:** 0.1  
**Last updated:** 2026-07-13  
**Initial operating assumption:** Private GitHub repositories, a capable home server, ChatGPT Pro, and Claude Max with Claude Code.  
**Project owner:** TBD  
**Security owner:** TBD  
**Pilot repository:** TBD

## Executive decision

The platform will use GitHub as its source-control and workflow control plane, with disposable or dedicated local runners as the primary execution plane. GitHub-hosted runners provide a controlled fallback for deterministic workloads and remain the default for code from public forks. Personal Claude and ChatGPT subscription credentials stay only on a separate trusted agent host and are never exposed to pull-request jobs.

One coding agent normally implements a change. Claude and Codex then perform fresh-context, independent reviews against the same frozen specification, diff, and deterministic evidence. After both initial reports are sealed, each may assess the other's findings. A deterministic policy engine decides whether the change passes. Agent agreement is evidence; it cannot override compilation, tests, security controls, or policy.

## Documentation map

| Document | Purpose |
|---|---|
| [01 — Project charter and requirements](01-project-charter-and-requirements.md) | Business case, scope, requirements, constraints, success measures, and MVP definition |
| [02 — System architecture](02-system-architecture.md) | Components, trust zones, data flow, review sequence, and architectural boundaries |
| [03 — Specification and evaluation framework](03-spec-and-eval-framework.md) | Executable specifications, eval design, evidence formats, scoring, baselines, and flakiness rules |
| [04 — Agent collaboration protocol](04-agent-collaboration-protocol.md) | Writer/reviewer roles, blind review, cross-review, arbitration, repair loops, budgets, and shared memory |
| [05 — CI/CD quality and security standard](05-ci-cd-quality-and-security-standard.md) | Required GitHub, build, test, analysis, supply-chain, and deployment controls |
| [06 — Infrastructure and deployment](06-infrastructure-and-deployment.md) | Home-server topology, runner lifecycle, routing/fallback, storage, network isolation, and delivery architecture |
| [07 — Implementation roadmap](07-implementation-roadmap.md) | Phases, deliverables, milestones, exit criteria, dependencies, and backlog |
| [08 — Operations runbook](08-operations-runbook.md) | Day-two operations, outages, credential expiry, compromise response, waivers, rollback, and maintenance |
| [09 — Risk register and decisions](09-risk-register-and-decisions.md) | Active risks, mitigations, decision log, and architecture-decision index |

Supporting assets:

| Asset | Purpose |
|---|---|
| [Change specification template](templates/change-spec-template.md) | Standard starting point for a feature, fix, refactor, or operational change |
| [Architecture decision template](templates/adr-template.md) | Records a consequential technical decision and its tradeoffs |
| [Policy waiver template](templates/policy-waiver-template.md) | Time-bounded, owned exception to a quality or security policy |
| [Agent review schema](schemas/agent-review.schema.json) | Machine-validatable independent blind-review output |
| [Agent cross-review schema](schemas/agent-cross-review.schema.json) | Machine-validatable second-pass response to sealed findings |
| [Change manifest schema](schemas/change-manifest.schema.json) | Machine-validatable link between a specification and its verifiers |
| [Evidence bundle schema](schemas/evidence-bundle.schema.json) | Machine-validatable immutable run evidence and policy decision |
| [Example change manifest](examples/change-manifest.example.yaml) | Concrete example of executable acceptance traceability |
| [Example agent review](examples/agent-review.example.json) | Schema-valid independent review report with reproducible evidence |
| [Example agent cross-review](examples/agent-cross-review.example.json) | Schema-valid response to another reviewer's sealed findings |
| [Example evidence bundle](examples/evidence-bundle.example.json) | Schema-valid run manifest linking checks, reviews, and decision |

## Core principles

1. **Specifications are versioned inputs.** Every run records the exact specification and commit hashes it evaluated.
2. **Acceptance criteria are executable.** Every mandatory criterion maps to a test, contract, static rule, policy, or scored evaluation.
3. **Deterministic evidence wins.** Agents cannot vote away a failed hard gate.
4. **Review begins independently.** Reviewers do not see one another's initial findings, reducing anchoring and false consensus.
5. **Credentials and untrusted code do not mix.** Build runners receive no personal AI credentials; agent hosts do not execute arbitrary pull-request scripts.
6. **Automation is bounded.** Agent turns, wall-clock time, retries, repair cycles, and concurrent subscription use are limited.
7. **The repository owns its pipeline.** GitHub, Bitbucket, local shells, and future systems call the same checked-in `ci/run` interface.
8. **Memory is curated evidence.** Specifications, ADRs, regression tests, and accepted findings persist; unverified model claims and private reasoning do not.
9. **Build once, promote by digest.** Release artifacts are scanned, signed, attested, and moved unchanged between environments.
10. **Recovery is designed in.** The system has explicit behavior for offline runners, rate limits, expired logins, flaky services, and failed deployments.

## High-level lifecycle

1. Author and approve a change specification.
2. Freeze its acceptance manifest and base revision.
3. Implement in an isolated worktree.
4. Execute deterministic PR gates.
5. Produce independent Claude and Codex reviews.
6. Exchange structured findings for an evidence-backed second pass.
7. Apply deterministic arbitration.
8. Repair at most twice, rerunning all affected gates each time.
9. Merge through protected GitHub rules.
10. Build, scan, sign, attest, stage, verify, and promote one immutable artifact.

## Proposed repository integration

```text
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
```

Workflow YAML should remain orchestration-only. Compilation, testing, analysis, evaluation, and evidence normalization belong behind `ci/run`, so local execution and other CI providers behave consistently.

## Initial technology decisions

- **SCM/control plane:** GitHub and GitHub Actions.
- **Pilot local runner:** Official GitHub self-hosted runner inside a dedicated Ubuntu VM.
- **Target runner architecture:** Ephemeral runner pods through Actions Runner Controller for trusted scale-out; disposable Incus VMs through GARM where stronger isolation is required.
- **Specification system:** GitHub Spec Kit for the pilot; OpenSpec is the documented alternative for heavily brownfield or multi-repository use.
- **Interactive agent bridge:** OpenAI's official `codex-plugin-cc` for Claude Code.
- **Automated review:** A small broker invoking only the official Claude Code and Codex clients and validating schema-constrained output.
- **Policy:** OPA/Conftest or an equivalently deterministic policy evaluator over a normalized evidence bundle.
- **Evaluation:** Native tests and hidden acceptance cases first; Promptfoo for LLM/agent behavior; Phoenix for local OpenTelemetry traces and experiments.
- **Evidence:** JUnit, SARIF, LCOV/Cobertura, CycloneDX/SPDX, in-toto/SLSA-style provenance, and OpenTelemetry.

## Non-negotiable safety rules

- Never mount `~/.claude/.credentials.json`, `~/.codex/auth.json`, SSH agent sockets, cloud credentials, or the host Docker socket into a pull-request runner.
- Never run public-fork code on a persistent home runner.
- Never use `pull_request_target` to execute a pull request's code.
- Never let an agent approve its own policy waiver or directly merge to a protected branch.
- Never retry semantic failures until a passing sample appears.
- Never promote an artifact different from the one scanned and approved.
- Never allow model consensus to override a failed hard gate.

## Primary research anchors

- [OpenAI Codex plugin for Claude Code](https://github.com/openai/codex-plugin-cc)
- [Codex authentication for trusted private CI](https://developers.openai.com/codex/auth/ci-cd-auth)
- [Claude Code authentication](https://code.claude.com/docs/en/iam)
- [GitHub Spec Kit](https://github.com/github/spec-kit)
- [GitHub self-hosted runner security](https://docs.github.com/en/actions/reference/security/secure-use)
- [GitHub Actions Runner Controller](https://docs.github.com/en/actions/concepts/runners/actions-runner-controller)
- [SLSA build track](https://slsa.dev/spec/v1.2/build-track-basics)
