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

**Finding F-01 — separate server exists but is deferred (owner answer, 2026-07-14).** A separate physical box exists, **dual-booting between a NAS role and a server role** (the two roles cannot run simultaneously). The owner's plan: build and test the platform entirely on this workstation, then migrate to the server box after the product is complete. Until migration, the workstation hosts daily development, the local CI runner (P3), and the agent broker (P4/P5) — colocation recorded as risk **R-038**. The dual-boot exclusivity (NAS offline while serving, and vice versa) must be inventoried before the migration is planned.

**Finding F-02 — disk location decided; capacity risk stays open (2026-07-14).** Runner VM disks, build caches, and evidence storage will live on **E:** (3.7 TB, ~116 GB currently free). This is a *location decision*, not a capacity resolution: doc-06 sizing gives the ci-pilot VM alone 60–120 GB. Pilot mitigation: the VM uses a dynamically-allocated disk capped at 60 GB, caches and evidence carry quotas from day one, and E: free space is monitored with an alert threshold. **R-039 remains open (Mitigating)** and closes only when measured post-VM headroom is acceptable or capacity is added — checked at P3 exit (G3).

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
| Router VLAN capability | **None** — Xfinity default gateway (owner answer, 2026-07-14). Hardened-v1 segmentation will use VM-level network isolation and host firewall rules, or a future managed switch, not router VLANs |
| NAS / other LAN devices | NAS exists on the same physical box as the future server (dual-boot; mutually exclusive with server role) |
| UPS / power protection | **None** (owner answer). Accepted for the pilot; R-015's outage handling covers power loss — an interrupted job is an infrastructure failure, never silently retried |

Full VLAN segmentation (doc 06 §8) is a hardened-v1 concern; for the pilot, VirtualBox host-only/NAT network modes plus Windows Firewall rules provide the interim boundary. The reachability red-team test (ACT-004) remains required before agent credentials are introduced.

## 5. GitHub account

| Item | Value |
|---|---|
| Account | `Aakash1337` |
| Plan | **Free** — 2,000 Actions minutes/month and 0.5 GB Actions storage included, $0 billable (owner billing screenshot, 2026-07-14; allowance resets on an ~monthly cycle) |
| Council repo | `Aakash1337/Council`, **public** (intentional — shareable documentation) |
| Pilot repo | `Aakash1337/CustomDNS`, **private** (flipped by owner, 2026-07-14) |

**Finding F-03 — GitHub Free cannot enforce branch protection on private repositories.** Branch protection and rulesets on private repos require GitHub Pro or higher. Consequence: Phase 2 can implement required-check *workflows*, but GitHub will not *block* a direct push or an unchecked merge to `main` on the private pilot repo. Owner decision required — see the [G0 checklist](README.md) owner-actions table (option A: GitHub Pro upgrade, ~$4/month, needs a budget-policy amendment; option B: accept process-discipline-only protection for the pilot and revisit before unattended agent work in P5; option C is not available — making the pilot public conflicts with CON-004). Recorded as risk **R-041**.

## 6. Subscriptions

| Item | Status |
|---|---|
| Claude Max | Active (Claude Code 2.1.206 operational on this machine) |
| ChatGPT Pro | Active (Codex CLI 0.144.1 installed) |
| Approved API spend | $0 (CON-001 / ADR-012) |

Unattended-auth qualification (setup-token, P4-09/PAC-018) is deliberately deferred to Phase 4.
