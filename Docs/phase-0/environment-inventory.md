# P0-01 — Environment Inventory

**Status:** Complete, with owner-input items marked
**Collected:** 2026-07-13, directly from the target workstation by the platform agent
**Roadmap reference:** [Phase 0 work item P0-01](../agentic-cicd-docs/07-implementation-roadmap.md)

Items the agent could not observe are marked **[OWNER INPUT]** and listed again in the [G0 checklist](README.md).

## 1. Compute

| Item | Value |
|---|---|
| Machine | ASUS desktop (custom build) |
| CPU | AMD Ryzen 9 5900X — 12 cores / 24 threads |
| RAM | 64 GB |
| Disk C: | 930 GB NVMe/SSD, **27 GB free** |
| Disk E: | 3,726 GB, **116 GB free** |
| OS | Windows 11 Home Single Language 10.0.26200 |
| Virtualization firmware | Enabled (`VirtualizationFirmwareEnabled: true`, `HypervisorPresent: true`) |

**Finding F-01 — no separate home server discovered.** The documentation set assumes "a capable home server" distinct from the operator's workstation. No separate server is visible from this machine. ASM-002 is provisionally satisfied by **this workstation**, which therefore hosts, at different times: daily development, the local CI runner (P3), and the agent broker (P4/P5). This colocation is recorded as new risk **R-038** in the [risk register](../agentic-cicd-docs/09-risk-register-and-decisions.md). **[OWNER INPUT]** — confirm no separate server exists, or provide its inventory.

**Finding F-02 — disk capacity is the binding constraint.** Both volumes are >85% full. A runner VM (60–120 GB per doc 06 §3.2), build caches, and evidence storage do not fit comfortably in current free space. Recorded as new risk **R-039**. **[OWNER ACTION]** — free or add roughly 150 GB before Phase 3.

## 2. Virtualization options

| Option | Present | Assessment |
|---|---|---|
| Hyper-V | Not available | Windows 11 **Home** does not include the Hyper-V manager |
| WSL2 | 2.1.5, running (docker-desktop distros) | Fast; weaker isolation (NAT to host network, filesystem interop); acceptable for dev, not for the runner trust boundary |
| VirtualBox | 7.0.20 installed | **Recommended runner-VM path** — real VM boundary, snapshots, host-only/NAT network modes; satisfies ADR-004's isolation intent on this hardware |
| Docker Desktop | 25.0.3, running | Reproducibility layer only; containers are not the security boundary (CON-007) |

Recorded as new risk **R-040**: the platform's VM lifecycle depends on VirtualBox on this host (community tooling, manual update cadence) rather than the hypervisors the docs assume (Proxmox/Incus/libvirt). Doc 06 is hypervisor-neutral by design, so this is a documented deviation, not a violation.

## 3. Developer tooling

| Tool | Version |
|---|---|
| git | 2.53.0.windows.2 |
| gh | 2.95.0 (authenticated as `Aakash1337`; scopes: gist, read:org, repo, workflow) |
| node | 24.14.0 |
| python | 3.11.9 |
| go | 1.26.2 |
| docker | 25.0.3 (Docker Desktop) |
| claude (Claude Code) | 2.1.206 |
| codex (Codex CLI) | 0.144.1 |
| cargo/rust | not installed |

Both first-party agent clients are installed and current. They currently live in the operator's primary user profile, not the dedicated OS accounts doc 06 §10.2 requires — acceptable until P4, when the dedicated agent identity is created (see R-038).

## 4. Network

| Item | Value |
|---|---|
| NIC | Marvell AQC111C 5 GbE, linked at 1 Gbps |
| Inbound exposure | None required (self-hosted runners are outbound-only) |
| Router VLAN capability | **[OWNER INPUT]** |
| NAS / other LAN devices | **[OWNER INPUT]** |
| UPS / power protection | **[OWNER INPUT]** |

Full VLAN segmentation (doc 06 §8) is a hardened-v1 concern; for the pilot, VirtualBox host-only/NAT network modes plus Windows Firewall rules provide the interim boundary. The reachability red-team test (ACT-004) remains required before agent credentials are introduced.

## 5. GitHub account

| Item | Value |
|---|---|
| Account | `Aakash1337` |
| Plan | Not readable with current token scope — **[OWNER INPUT]**: confirm plan and monthly Actions minutes for private repositories (Free plan = 2,000 hosted minutes/month) |
| Council repo | `Aakash1337/Council`, **public** (intentional — shareable documentation) |
| Pilot candidate repo | `Aakash1337/CustomDNS`, currently **public** (see [pilot decision](pilot-decision.md)) |

## 6. Subscriptions

| Item | Status |
|---|---|
| Claude Max | Active (Claude Code 2.1.206 operational on this machine) |
| ChatGPT Pro | Active (Codex CLI 0.144.1 installed) |
| Approved API spend | $0 (CON-001 / ADR-012) |

Unattended-auth qualification (setup-token, P4-09/PAC-018) is deliberately deferred to Phase 4.
