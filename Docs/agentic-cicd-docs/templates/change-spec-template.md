# Change Specification: `<CHANGE-ID> — <Short title>`

**Status:** Draft | In review | Approved | Implemented | Archived  
**Owner:** `<name or team>`  
**Risk tier:** Low | Medium | High | Critical  
**Target repository:** `<owner/repository>`  
**Base revision:** `<full commit SHA>`  
**Created:** `<YYYY-MM-DD>`  
**Last updated:** `<YYYY-MM-DD>`

## 1. Problem statement

Describe the observable problem, who experiences it, and why it is worth solving. Avoid prescribing the implementation here.

## 2. Desired outcome

State the outcome in measurable terms.

## 3. Users and stakeholders

| Stakeholder | Need or responsibility |
|---|---|
| `<user/operator/team>` | `<need>` |

## 4. Scope

### In scope

- `<behavior, interface, component, or operational result>`

### Out of scope

- `<explicit exclusion>`

## 5. Current behavior

Describe the current state, including known limitations and relevant evidence.

## 6. Required behavior

Use normative, testable statements. Assign each statement a stable requirement ID.

| Requirement | Statement | Priority |
|---|---|---|
| REQ-001 | The system MUST `<required behavior>`. | Must |
| REQ-002 | The system SHOULD `<preferred behavior>`. | Should |

## 7. User stories and scenarios

### Scenario S-001 — `<name>`

**Given** `<initial condition>`  
**When** `<action or event>`  
**Then** `<observable result>`

Include normal, boundary, failure, concurrency, authorization, and recovery scenarios where applicable.

## 8. Acceptance criteria

Every mandatory criterion must map to an executable verifier in `acceptance.yaml`.

The table uses local criterion IDs. Review and evidence records qualify them as `<CHANGE-ID>/<AC-ID>` to prevent collisions with another change or with platform acceptance criteria.

| Acceptance ID | Criterion | Criticality | Planned verifier | Visibility |
|---|---|---|---|---|
| AC-001 | `<observable invariant>` | Hard | `test://...` | Public |
| AC-002 | `<quality target>` | Scored | `eval://...` | Public/hidden |

Allowed criticalities:

- **Hard:** Any valid failure blocks.
- **Scored:** Must meet an explicit threshold and regression budget.
- **Advisory:** Reported but not independently blocking.

## 9. Interfaces and contracts

Document affected APIs, events, command-line interfaces, file formats, database schemas, configuration, and backward-compatibility rules.

| Interface | Proposed change | Compatibility requirement |
|---|---|---|
| `<API/schema/event>` | `<change>` | `<rule>` |

## 10. Data and privacy

- Data created, read, changed, transmitted, or deleted:
- Classification and sensitivity:
- Retention impact:
- Logging/telemetry impact:
- Migration or backfill:
- Privacy or consent concerns:

## 11. Security requirements

- Authentication and authorization:
- Secret handling:
- Input validation and injection controls:
- Network trust changes:
- Abuse cases:
- Audit requirements:
- Supply-chain impact:

## 12. Reliability and recovery

- Availability expectation:
- Timeout and retry behavior:
- Idempotency:
- Failure isolation:
- Degraded mode:
- Rollback or forward-fix strategy:
- Backup/restore impact:

## 13. Performance and capacity

Define the workload, measurement environment, baseline, target, and tail metric. Prefer p95/p99 and error-rate budgets over averages alone.

| Metric | Baseline | Required target | Measurement method |
|---|---:|---:|---|
| `<metric>` | `<value>` | `<value>` | `<verifier>` |

## 14. Observability

- Required logs:
- Metrics and thresholds:
- Traces:
- Audit events:
- Alerts:
- Dashboard changes:

## 15. Proposed approach

Summarize the expected technical direction after requirements have been approved. Link a separate architecture decision for consequential choices.

## 16. Alternatives considered

| Alternative | Advantages | Disadvantages | Disposition |
|---|---|---|---|
| `<alternative>` | `<advantages>` | `<disadvantages>` | Rejected/deferred |

## 17. Dependencies and constraints

- Technical dependencies:
- Organizational dependencies:
- Provider/subscription constraints:
- Deployment windows:
- Known compatibility constraints:

## 18. Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| `<risk>` | Low/medium/high | Low/medium/high/critical | `<mitigation>` |

## 19. Delivery and rollback plan

1. `<delivery step>`
2. `<validation step>`
3. `<promotion step>`

Rollback trigger:

Rollback procedure:

Irreversible steps requiring approval:

## 20. Open questions

| Question | Owner | Due date | Resolution |
|---|---|---|---|
| `<question>` | `<owner>` | `<date>` | Open |

## 21. Approval

| Role | Name | Decision | Date |
|---|---|---|---|
| Product/request owner |  |  |  |
| Technical owner |  |  |  |
| Security owner, if required |  |  |  |
| Release/data owner, if required |  |  |  |

Approval freezes the specification and acceptance-manifest hashes used by implementation and review. Material intent changes require a new approved revision.
