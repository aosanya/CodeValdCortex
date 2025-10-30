---
title: Work Items Specification
path: /documents/2-SoftwareDesignAndArchitecture/work-items.md
---

# Work Items (WI) — Design Specification

This document specifies the Work Items subsystem used by CodeValdCortex agency operations. It complements the existing RACI and Goals documentation by describing:

- Work item types and their required fields
- Canonical lifecycle and SLA/SLO behaviour
- Assignment and routing rules (declarative)
- Concurrency, idempotence and reentrancy contracts
- Compensation (sagas) and rollback strategies
- Policy-driven validation and approvals with evidence capture
- Templating and versioned catalogs for standard processes

Where applicable the document includes example JSON/TypeScript interfaces and implementation notes for storage (ArangoDB documents) and runtime enforcement.

## 1. Goals and Acceptance

Acceptance criteria for this specification:

- CRUD for Work Items implemented and discoverable via API
- Work item types available in registry with schemas and templates
- Lifecycle state machine enforced with audit trail
- SLA/SLO fields exist and breach actions are actionable
- Declarative routing rules apply to assignment and escalation
- Concurrency controls prevent duplicate, conflicting external effects
- Compensating actions and saga orchestration supported for external mutations
- Policy-driven gates for privacy/compliance are configurable per template

## 2. Work Item Types & Contracts

Contract: every Work Item document must include a type field and adhere to the schema for that type. Types are registered in the `work_item_types` registry (analogous to `agent_types`).

Core types (recommended):

- Task — single, short-running unit of work executed by an agent or person
- Job — collection of tasks or long-running process (batch or pipeline)
- Investigation — exploratory analysis, typically human-led with evidence capture
- Change — planned change to systems (requires change control template)
- Remediation — corrective action after an incident or validation failure
- Experiment — controlled test with measured metrics and rollback plan

Each type example (TypeScript-like interface):

```typescript
interface WorkItemBase {
  _key?: string;                // ArangoDB key
  _id?: string;                 // e.g. work_items/AGENCY_WI_0001
  code: string;                 // Business code: WIFINSERV-001
  type: 'task'|'job'|'investigation'|'change'|'remediation'|'experiment';
  title: string;
  description?: string;
  agencyId: string;
  priority?: number;            // 0..100
  createdBy?: string;           // user id
  createdAt: ISODateTime;
  updatedAt?: ISODateTime;
  status: WorkItemStatus;       // see lifecycle below
  raci?: RACIMatrix;            // RACI per activity
  relatedGoals?: string[];      // goal codes
  dependencies?: string[];      // other work_item._id
  labels?: string[];            // free-form labels for routing
  metadata?: Record<string, any>;
}

interface TaskWorkItem extends WorkItemBase {
  type: 'task';
  assignee?: string;            // user id or agent id
  dueAt?: ISODateTime;
  timeoutMs?: number;           // runtime enforcement
  idempotenceKey?: string;      // prevents double processing
  payload?: Record<string, any>; // input for agent
  compensation?: CompensationPlan; // for external effects
}

interface ChangeWorkItem extends WorkItemBase {
  type: 'change';
  changeWindow?: {start: ISODateTime; end: ISODateTime};
  riskAssessment: {
    riskLevel: 'LOW'|'MEDIUM'|'HIGH'|'CRITICAL';
    mitigations: string[];
  };
  approvalPolicyId?: string; // references policy that gates activation
  templatesApplied?: {templateId:string; version:string}[];
}
```

Required fields per type (high-level):

- Task: title, type, createdAt, status, payload or description
- Job: title, type, tasks[] or workflow spec, schedule or trigger
- Investigation: title, scope, successCriteria, evidenceRequirements
- Change: title, changeWindow, approvalPolicyId, rollbackPlan
- Remediation: title, incidentRef, startAt, remediationPlan
- Experiment: title, hypothesis, metrics, successCriteria, rollbackPlan

Storage note: persist all work items in `work_items` collection. Keep type-specific schema in `work_item_types` registry as JSON Schema for validation on create/update.

## 3. Lifecycle & SLAs

Canonical lifecycle states:

Planned → In-Progress → Waiting → Review → Done | Failed | Rolled-back

Detailed statuses (string enum):

```go
type WorkItemStatus string
const (
  WIPlanned WorkItemStatus = "planned"
  WIInProgress WorkItemStatus = "in_progress"
  WIWaiting WorkItemStatus = "waiting"
  WIReview WorkItemStatus = "review"
  WIDone WorkItemStatus = "done"
  WIFailed WorkItemStatus = "failed"
  WIRolledBack WorkItemStatus = "rolled_back"
)
```

State machine rules (allowed transitions):

- planned -> in_progress (assign & start)  
- in_progress -> waiting (blocked: external wait, dependency)  
- waiting -> in_progress (unblocked)  
- in_progress -> review (requires evidence if policy demands)  
- review -> done (approval satisfied)  
- review -> failed (rejected)  
- in_progress -> failed (execution error after retries)  
- failed -> remediation (create remediation work item)  
- done -> rolled_back (if compensating actions executed)  

Timers, SLAs and SLOs:

Each work item may declare SLA/SLO fields. Example:

```json
"sla": {
  "targetMs": 3600000,            // target completion time from planned->done
  "breachAction": "escalate",   // escalate, auto-retry, human-intervention
  "breachEscalationPolicyId": "esc-001"
}
```

Runtime enforcement:

- Scheduler monitors SLA timers and triggers configured breachAction when target exceeded
- Breach actions supported: notify, escalate, cancel, open-remediation, pause, queue-human-approval

Timeouts:

- Tasks may include `timeoutMs` (per execution) and `overallTimeoutMs` (deadline for the WI)
- Agents must cancel execution when `timeoutMs` exceeded and return status `failed` or `timed_out`

Breach behavior examples:

- If SLA.breachAction = "escalate": system creates an escalation ticket and notifies escalation recipients defined in policy.
- If SLA.breachAction = "open-remediation": automatically create a `remediation` work item linked to the failing WI and assign to on-call.

SLA metrics to store (for reporting): createdAt, startedAt, completedAt, lastUpdatedAt, slaBreachedAt, slaBreachCount.

## 4. Assignment & Routing (Declarative Rules)

Routing rules are declarative objects that map a work item to candidate assignees (agents or people) and define escalation paths.

Routing rule example (JSON Schema):

```json
{
  "id": "route-001",
  "match": { "labels": ["db","urgent"], "type": ["remediation","change"] },
  "selectors": [
    { "kind": "skill", "value": "postgres-admin" },
    { "kind": "role", "value": "oncall" }
  ],
  "costBudget": {"currency":"USD","limit":1000},
  "dataResidency": {"allowedRegions":["KE","UG"]},
  "escalation": {
    "onTimeoutMs": 3600000,
    "steps": [
      {"afterMs": 0, "action":"notify", "to":["assignee"]},
      {"afterMs": 1800000, "action":"escalate", "to":["team_lead"]},
      {"afterMs": 3600000, "action":"escalate", "to":["agency_lead"]}
    ]
  },
  "hitl": {"required":true, "gates":["privacy_review","security_review"]}
}
```

Routing rules evaluation:

- Rules are evaluated at creation time and on re-routing triggers (failure or SLA breach)
- Selection strategies: priority, capacity-aware, round-robin, cost-minimizing
- Final assignment chosen by `selector` matching and runtime capability checks (e.g., agent has required capabilities)

Hand-offs and HITL:

- `hitl` property requires explicit human approval before certain transitions (e.g., change -> review/activate). HITL checkpoints record approver, timestamp, justification, and captured evidence.

Assignment auditing:

- All assignment decisions must be logged with `reason`, `ruleId`, `candidates`, and `chosenAssignee` for traceability.

## 5. Concurrency Controls & Idempotence

To prevent duplicate work or conflicting external effects, Work Items and Task executions must adhere to concurrency contracts.

Idempotence:

- Producers (work item creators) may include `idempotenceKey` on creation to avoid duplicate WIs when retried.
- Agents executing tasks should accept `idempotenceKey` in payload; idempotent handlers must detect and return previous result instead of re-executing side-effects.

Mutex scopes:

- `mutexScope` is an optional label that marks a resource or domain for mutual exclusion. Example: `mutexScope: "db:users:12345"`
- Runtime must acquire lease (Redis/ArangoDB TTL-based lock or etcd) before executing side-effecting operations. Locks include owner, ttlMs, and leaseRenewal.

Reentrancy contracts:

- Work item types should declare whether handlers are reentrant. If non-reentrant, the scheduler must serialize executions for same `idempotenceKey` or `mutexScope`.

Best practices:

- Use deterministic deterministic idempotenceKey generation including workItem.code + inputHash + actorId
- Short locks for high contention resources; prefer optimistic concurrency where possible

## 6. Compensation & Saga Orchestration

When a work item performs multiple operations that mutate external systems, use Saga patterns and defined compensation steps.

Compensation plan example:

```json
"compensation": {
  "steps": [
    {"actionId":"revoke_api_key", "on":"failure_of","targetStep":"create_api_key"},
    {"actionId":"delete_temp_records", "on":"any_failure","targetStep":"*"}
  ],
  "coordinator": "orchestration/saga-service",
  "retryPolicy": {"maxRetries":3, "backoffMs":1000}
}
```

Saga orchestration modes:

- Choreography: each step emits events; compensators subscribe and react. Simpler but harder to reason about for complex failures.
- Orchestration: a central coordinator (workflow/orchestration service) runs steps and calls compensations on failure. Preferred for work items that must maintain strict ordering and compensations.

Rollback semantics:

- If action A succeeds and action B fails, the coordinator executes compensation for A in reverse order.
- Compensation must itself be idempotent and able to tolerate partial failures; compensations must be logged and monitored.

Audit and evidence:

- Each saga run produces a run document: {runId, workItemId, steps: [{stepId, status, startedAt, endedAt, logs}], compensationStatus}

## 7. Validation & Approvals (Policy Gates)

Policy-driven gates allow compliance/privacy/risk checks before transitions (e.g., planned->in_progress or review->done). Policies are pluggable and versioned.

Gate example:

```json
"approvalPolicy": {
  "id": "policy-change-high-risk-v1",
  "checks": [
    {"kind":"compliance","framework":"PCI-DSS","required":true},
    {"kind":"privacy","dataClassification":"RESTRICTED","required":true},
    {"kind":"risk","maxRiskLevel":"MEDIUM"}
  ],
  "approvalSteps": [
    {"role":"Security Officer","action":"approve"},
    {"role":"Agency Lead","action":"approve"}
  ]
}
```

Evidence capture:

- For each gate the system stores evidence objects (documents) including attachments, signed approvals, scan results, and a digest reference to external artifacts.
- Evidence schema: {id, policyId, workItemId, evidenceType, createdBy, createdAt, metadata, artifacts:[artifactId,...]}

Automated validators:

- Validators may be synchronous (blocking transition until check completes) or asynchronous (allow transition but flag for audit and create remediation if failed).

## 8. Templating & Catalogs

Work Item templates allow rapid creation of standard workflows (e.g., PCI change, HIPAA export review). Templates are versioned artifacts stored in `work_item_templates` collection.

Template model:

```json
{
  "id": "tpl:change:pci:1.0.0",
  "name": "PCI Change Request",
  "version": "1.0.0",
  "type": "change",
  "parameters": {
    "system": {"type":"string","required":true},
    "changeWindow": {"type":"timespan","required":true}
  },
  "workflow": { /* orchestration/workflow spec or task list */ },
  "approvalPolicyId": "policy-change-pci-v1",
  "documentation": "..."
}
```

Template catalog features:

- Parameterization with typed parameters and defaults
- Versioning (SemVer) and migration notes
- Template inheritance and composition
- Template validation against registered `work_item_types` schema

## 9. APIs & Implementation Notes

Storage collections:

- `work_items` — the Work Item documents
- `work_item_types` — registered types with JSON Schemas
- `work_item_templates` — versioned templates
- `work_item_runs` / `saga_runs` — orchestration runtime records
- `work_item_evidence` — captured evidence for approvals

Recommended API endpoints (examples):

```
GET    /api/v1/agencies/{id}/work-items
POST   /api/v1/agencies/{id}/work-items
GET    /api/v1/work-items/{workItemId}
PUT    /api/v1/work-items/{workItemId}
POST   /api/v1/work-items/{workItemId}/transition  // change status via controlled API
POST   /api/v1/work-items/{workItemId}/evidence
POST   /api/v1/work-items/{workItemId}/compensate
POST   /api/v1/work-items/{workItemId}/reroute
```

Transitions should be validated server-side with policy checks and permission enforcement.

## 10. Observability & Metrics

Key metrics to export:

- Work items created per type
- Mean time to start (planned->in_progress)
- Mean time to complete (planned->done)
- SLA breach count and rate
- Compensation runs and success rate
- Approval latency per policy

Logs and tracing:

- Each work item run should include structured logs and a trace ID compatible with system tracing (OpenTelemetry). Saga steps should be traceable end-to-end.

## 11. Examples

- Example: Auto-remediation flow
  - Change WI created for patching. Approval policy requires security review. Agent performs patching (job) with compensation to rollback package install. If job fails at step 3, compensation invoked to revert changes and remediation WI created.

- Example: HIPAA data export (template)
  - Template enforces `approvalPolicy` requiring privacy officer sign-off, evidence capture of exported records, and data residency check.

## 12. Agent Types Taxonomy

Work items are executed by agents with well-defined capabilities, autonomy levels, and constraints. The complete agent types taxonomy is documented separately.

**See**: [Agent Types Taxonomy](./agent-types-taxonomy.md) for comprehensive documentation including:

- **Agent Type Classifications**: Stateless Tool-Caller, Planner/Coordinator, Data Access Agent, Long-Running Service, Sensor/Monitor, Actuator, Reviewer/HITL Proxy
- **Skills & Tools Contract**: Tool adapters, capability declarations, proficiency levels
- **Autonomy Levels (L0-L4)**: From manual to full autonomy with policy-bound action scopes
- **Budgeting**: Token/$ budgets, time/compute quotas, exhaustion behaviors
- **Data Boundaries**: Allowed datasets, masking rules, data residency, cross-border controls
- **Safety Constraints**: Allowed/prohibited actions, two-person rule, dry-run requirements
- **Identity & Tenancy**: OIDC/SPIFFE workload identity, attestation, tenant isolation guarantees

### 12.1 Integration with Work Items

Agent type selection for work item execution considers:

1. **Required Capabilities**: Work item type declares required agent capabilities
2. **Autonomy Requirements**: Work item risk level determines minimum autonomy level
3. **Budget Constraints**: Work item SLA maps to agent budget allocation
4. **Data Access Needs**: Work item data requirements match agent data boundaries
5. **Safety Requirements**: Work item operations validate against agent safety constraints

**Assignment Algorithm**:
```
1. Filter agents by required capabilities
2. Check autonomy level meets work item risk threshold
3. Verify budget availability for estimated work item cost
4. Validate data boundary permissions
5. Confirm safety constraints allow work item operations
6. Select optimal agent based on cost, capacity, and proficiency
```

## 13. Traceability & Validation

To ensure complete traceability from work items to agent actions to artifacts, the system must maintain explicit linkage documents and validate chains for completeness.

### 13.1 Traceability Schema

```typescript
interface TraceabilityValidation {
  validationId: string;
  timestamp: ISODateTime;
  
  checks: {
    goalToWorkItemLinkage: {
      passed: boolean;
      orphanedGoals: string[];  // Goals with no work items
    };
    
    workItemToActionLinkage: {
      passed: boolean;
      orphanedWorkItems: string[];  // Work items with no agent actions
    };
    
    actionToArtifactLinkage: {
      passed: boolean;
      orphanedActions: string[];  // Actions with no artifacts
    };
    
    deterministicIds: {
      passed: boolean;
      duplicateIds: string[];  // Non-unique IDs detected
      invalidFormats: string[];  // IDs not matching format spec
    };
    
    compensationTraceability: {
      passed: boolean;
      incompleteCompensations: string[]; // Saga runs with missing compensation logs
    };
    
    approvalTraceability: {
      passed: boolean;
      missingApprovals: string[]; // Work items requiring approval but lacking evidence
    };
  };
  
  overallStatus: "complete" | "incomplete" | "broken";
  recommendedActions: string[];
}
```

**Validation Schedule**:
- **Real-time**: On work item creation, transition, and completion
- **Daily**: Nightly batch validation of all active traceability chains
- **Monthly**: Comprehensive audit of archived work items and completed sagas

**Broken Chain Resolution**:
1. **Detection**: Validation check identifies missing link
2. **Notification**: Alert work item owner and agency lead
3. **Remediation**: Manual review to restore link or mark as invalid
4. **Prevention**: Enforce foreign key constraints and required fields in schema

### 13.2 Deterministic ID Generation

All work items, actions, and artifacts must have deterministic, globally unique IDs:

**Work Item ID Format**: `WI-{agencyCode}-{type}-{timestamp}-{hash}`
- Example: `WI-FINSERV-CHANGE-20251030-A3F9B2`

**Agent Action ID Format**: `ACT-{workItemId}-{agentId}-{sequence}`
- Example: `ACT-WI-FINSERV-CHANGE-20251030-A3F9B2-agent-01-001`

**Artifact ID Format**: `ART-{actionId}-{artifactType}-{hash}`
- Example: `ART-ACT-WI-FINSERV-CHANGE-20251030-A3F9B2-agent-01-001-log-5D8E3A`

Benefits:
- **Uniqueness**: Collision-resistant due to timestamp + hash
- **Determinism**: Same inputs always produce same ID
- **Readability**: Human-readable components for debugging
- **Traceability**: IDs encode parent relationships

## 14. Implementation Tasks (Next Steps)

### Phase 1: Core Schema & Registry (MVP-030)
- [ ] Add JSON Schemas for each `work_item_type` to `work_item_types` collection
- [ ] Implement agent type taxonomy fields in `agent_types` registry
- [ ] Add autonomy level, budget, and safety constraint fields to agent schema
- [ ] Create default agent types with example configurations

### Phase 2: Lifecycle & SLA Enforcement (MVP-031)
- [ ] Implement server-side transition validator and `POST /work-items/{id}/transition` endpoint
- [ ] Add SLA timer monitoring service with breach detection
- [ ] Implement breach action handlers (escalation, remediation creation)
- [ ] Add lifecycle audit trail to work item documents

### Phase 3: Assignment & Routing (MVP-032)
- [ ] Build routing engine to evaluate declarative rules
- [ ] Implement skill-based assignment with capacity awareness
- [ ] Add cost budget enforcement to assignment algorithm
- [ ] Create escalation path execution service

### Phase 4: Concurrency & Idempotence (MVP-033)
- [ ] Implement idempotence key deduplication layer
- [ ] Add distributed mutex/lock service (Redis or ArangoDB-based)
- [ ] Enforce reentrancy contracts in task execution
- [ ] Add mutex scope validation to work item creation

### Phase 5: Compensation & Sagas (MVP-034)
- [ ] Implement saga orchestration runner (orchestrator pattern)
- [ ] Add compensation step execution with retry logic
- [ ] Create saga run audit trail and visualization
- [ ] Build rollback testing framework

### Phase 6: Policy Gates & Evidence (MVP-035)
- [ ] Implement policy registry with versioned policies
- [ ] Add policy gate evaluation engine for transitions
- [ ] Build evidence capture UI and storage
- [ ] Integrate with external compliance scanners (optional)

### Phase 7: Templates & Catalog (MVP-036)
- [ ] Create template registry with versioning
- [ ] Implement template parameterization and instantiation
- [ ] Add industry templates (PCI change, HIPAA export, SOC2 review)
- [ ] Build template inheritance and composition system

### Phase 8: Traceability & Validation (MVP-037)
- [ ] Implement deterministic ID generation for all entities
- [ ] Add traceability validation service with scheduled checks
- [ ] Build broken chain detection and notification
- [ ] Create traceability dashboard and reports

---

**Document Version**: 0.2.0 — Draft (Enhanced with Agent Taxonomy & Traceability)  
**Last Updated**: 2025-10-30  
**Owner**: CodeValdCortex Architecture Team  
**Status**: Draft - Ready for Review
