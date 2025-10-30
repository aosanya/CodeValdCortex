# Goals System Specification Update Summary

## Date
October 30, 2024

## Overview
This document summarizes the comprehensive Goals system specification created in response to detailed architectural feedback on schema formalization, lifecycle management, prioritization, dependencies, impact analysis, and traceability requirements.

## Feedback Addressed

### 1. ✅ Schema & Versioning
**Feedback**: *"Formal goal schema (id, code, owner, OKRs/KPIs, risk level, compliance tags, data domains). SemVer and migration rules."*

**Implementation**:
- **Formal TypeScript Schema**: Complete interface definition with 25+ fields including:
  - Identity & versioning (`_key`, `_id`, `_rev`, `code`, `version`)
  - Ownership & accountability (`owner.role`, `owner.userId`, `owner.agencyId`)
  - Success measurement (`successMetrics[]` with OKRs/KPIs/Metrics)
  - Risk & compliance (`riskLevel`, `complianceTags[]`, `dataClassification`, `dataDomains[]`)
  - Business value tracking (`businessValue.category`, `estimatedImpact`, `confidenceLevel`)
  
- **Semantic Versioning (SemVer)**: Full implementation with:
  - MAJOR version: Breaking changes (success criteria changed, scope drastically modified)
  - MINOR version: Additive changes (new metrics, refined scope, dependencies added)
  - PATCH version: Non-semantic changes (typos, formatting, clarifications)
  - Version history in `changeHistory[]` array
  
- **Migration Rules**: `MigrationRule[]` interface with:
  - Version transition tracking (`fromVersion`, `toVersion`)
  - Migration scripts for data transformation
  - Backward compatibility flags
  - Automatic work item reference validation

**Location**: `documents/1-SoftwareRequirements/introduction/goals-specification.md`, Section 1

---

### 2. ✅ Lifecycle Management
**Feedback**: *"Canonical states (Draft → Proposed → Approved → Active → Paused → Completed → Archived), required approvals, and change control (who can change what when)."*

**Implementation**:
- **State Machine**: 8 canonical states with defined transitions:
  ```
  Draft → Proposed → Approved → Active → {Paused, Completed, Archived}
  ```
  
- **State Definitions Table**: Each state includes:
  - Entry requirements (what must be satisfied to enter)
  - Exit requirements (what must be satisfied to leave)
  - Allowed actions (what users can do in each state)
  
- **Required Approvals**: Complete approval matrix by transition:
  - Draft → Proposed: Goal Owner (1 approver)
  - Proposed → Approved: Governance Board (majority vote, 51%)
  - Approved → Active: Resource Allocator (automated if resources available)
  - Active → Completed: Goal Owner + QA Lead (both required)
  - Etc.
  
- **Change Control Matrix**: "Who can change what when" table:
  - 12 fields/actions × 7 states = 84 permission rules
  - Differentiates between "allowed without approval" (✓), "allowed with approval" (†), and "not allowed" (❌)
  - Role-based permissions (Owner, Lead, Compliance Officer)
  
- **Approval Requirements**:
  - PATCH version: Owner approval only
  - MINOR version: Owner + 1 governance board member
  - MAJOR version: Full governance board (majority vote)

**Location**: `documents/1-SoftwareRequirements/introduction/goals-specification.md`, Section 2

---

### 3. ✅ Prioritization & Capacity Management
**Feedback**: *"Priority classes, WIP limits, and scheduling policy when agent capacity is constrained."*

**Implementation**:
- **Priority Classes** (Kubernetes-inspired QoS model):
  - **P0 (Critical)**: 100% guaranteed resources, cannot be preempted (compliance, security incidents)
  - **P1 (High)**: 80% high allocation, can preempt P2/P3 (strategic initiatives)
  - **P2 (Medium)**: 50% fair share, can preempt P3 (standard features)
  - **P3 (Low)**: Best effort, can be preempted by all (research, technical debt)
  
- **Automatic Priority Calculation**:
  ```
  Score = businessValue*0.30 + complianceRequired*40 + riskLevel + dependencies*2 + deadline*10
  P0: Score >= 80, P1: Score >= 60, P2: Score >= 30, P3: Score < 30
  ```
  
- **WIP Limits**: Multi-level limits with enforcement:
  - Total active goals: 20 (example)
  - By priority: P0=unlimited, P1=5, P2=10, P3=3
  - By owner: Max 3 active goals per owner
  - Violations: Block (P0/P1) vs Warn (P2/P3)
  
- **Capacity-Aware Scheduling**:
  - Guaranteed capacity for P0 (100%), P1 (80%)
  - Fair share for P2
  - Best effort for P3
  - Preemption rules: P0 can preempt P3/P2, P1 can preempt P3 only
  - Real-time dashboard with 85% utilization alerts

**Location**: `documents/1-SoftwareRequirements/introduction/goals-specification.md`, Section 3

---

### 4. ✅ Dependency Semantics
**Feedback**: *"Hard vs soft dependencies, blocking rules, and conflict detection across goals."*

**Implementation**:
- **Dependency Types**:
  - **Hard Dependencies** (blocking): Must be satisfied before activation
    - Condition types: status, completion percentage, metric threshold, custom function
    - Enforcement levels: strict (cannot override) vs advisory (can override with approval)
  
  - **Soft Dependencies** (informational): Recommended but not required
    - Relationship types: benefits_from, related_to, follows
    - Optional recommended delays (e.g., "wait 7 days after dependency completes")
  
- **Blocking Rules Engine**:
  - 4 action types: block_activation, show_warning, require_approval, auto_pause
  - Override roles specified per rule
  - Example rules: "Hard dependency not complete" (block), "Progress insufficient" (require approval)
  
- **Conflict Detection**: 6 conflict types with algorithms:
  1. **Circular Dependency**: DFS cycle detection, O(V+E)
  2. **Resource Contention**: Exclusive resource claim checking
  3. **Contradictory Objectives**: NLP semantic analysis + keyword contradiction
  4. **Scope Overlap**: Cosine similarity >80% triggers review
  5. **Compliance Violation**: Multiple goals violating same rule
  6. **Priority Inversion**: Low priority blocking high priority
  
- **Conflict Resolution Workflow**: 5-step process from detection to resolution (merge, reprioritize, adjust dependencies, document trade-offs, escalate)

**Location**: `documents/1-SoftwareRequirements/introduction/goals-specification.md`, Section 4

---

### 5. ✅ Impact Analysis Specification
**Feedback**: *"What algorithms (reachability, centrality, cut-sets) power 'impact analysis'? Define accepted false-positive/false-negative rates and user overrides."*

**Implementation**:
- **Graph Algorithm Suite**: 6 algorithms with complexity analysis:
  - **Reachability** (DFS/BFS): Find all affected work items, O(V+E)
  - **Shortest Path** (Dijkstra): Most direct goal→work item path, O(E + V log V)
  - **Betweenness Centrality**: Identify critical goals/work items, O(VE)
  - **Connected Components**: Find isolated goal clusters, O(V+E)
  - **Cut-Set Analysis**: Minimum removal set to disconnect graph, O(V³)
  - **Transitive Closure**: All indirect dependencies, O(V³)
  
- **ArangoDB Implementation**: Example queries provided for:
  - Reachability: "Find all work items affected by goal change"
  - Centrality: "Find most critical goals (most work items depend on them)"
  - Cut-Set: "Find goals that, if removed, would disconnect work items"
  
- **False Positive/Negative Rates**: Defined acceptable error rates per algorithm:
  - Reachability: 5% FP, 1% FN (prefer over-reporting for safety)
  - Centrality: 10% FP, 5% FN (approximate rankings acceptable)
  - Conflict Detection: 15% FP, 2% FN (prefer false alarms)
  - Scope Overlap: 20% FP, 5% FN (human review filters)
  - Dependency Validation: 0% FP, 0% FN (must be deterministic)
  
- **Error Rate Monitoring**: `AnalysisMetrics` interface tracking:
  - Total runs, false positives/negatives (user-reported)
  - Computed rates, alert triggers
  - Monthly calibration process (sample 100 analyses, tune thresholds)
  
- **User Overrides**: 4 override types:
  - Exclude item (false positive suppression)
  - Include item (false negative correction)
  - Adjust impact level (primary → secondary)
  - Ignore conflict (acknowledged and acceptable)
  - All overrides logged, expire if not renewed, inform calibration

**Location**: `documents/1-SoftwareRequirements/introduction/goals-specification.md`, Section 5

---

### 6. ✅ Traceability Contract
**Feedback**: *"Required linkage from business KPI → goal → work items → agent actions → artifacts/logs with deterministic IDs."*

**Implementation**:
- **End-to-End Traceability Chain**: 5-level linkage:
  ```
  Business KPI → Goal → Work Item → Agent Action → Artifact/Log
  ```
  
- **TraceabilityChain Interface**: Complete schema with:
  - Level 1: Business KPI (id, name, target, current value, measurement unit)
  - Level 2: Goal (code, title, success metrics mapped to KPI)
  - Level 3: Work Items (code, relationship type, contribution %)
  - Level 4: Agent Actions (actionId, agentId, timestamp, input/output, execution metrics)
  - Level 5: Artifacts (artifactId, type, storage location, content hash, createdAt)
  
- **Deterministic ID Generation**: Formal ID formats:
  - Goal IDs: `goals/{agency_id}_{sequential_number}`
  - Work Item IDs: `work_items/{agency_id}_WI_{sequential_number}`
  - Agent Action IDs: `SHA256({agentId}:{workItemCode}:{timestamp}:{input_hash})`
  - Artifact IDs: `artifact_{action_id}_{artifact_type}_{sequence}`
  - Relationship IDs: `{goal_code}_{relationship_type}_{work_item_code}`
  
- **ID Properties**: Uniqueness, determinism, readability, traceability
  
- **Traceability Queries**: 3 example ArangoDB queries:
  1. KPI to Artifacts (forward trace through full chain)
  2. Artifact to KPI (reverse trace)
  3. Impact of Goal Change on KPIs (affected work items, actions, KPI estimates)
  
- **Traceability Validation**: `TraceabilityValidation` interface with:
  - 5 linkage checks (KPI→Goal, Goal→WorkItem, WorkItem→Action, Action→Artifact, ID determinism)
  - Detection of missing links, orphaned entities, duplicate IDs
  - Validation schedule: Real-time (on creation), daily (batch), monthly (comprehensive audit)
  - Broken chain resolution workflow (detect → notify → remediate → prevent)

**Location**: `documents/1-SoftwareRequirements/introduction/goals-specification.md`, Section 6

---

## Implementation Roadmap

The specification includes a phased implementation plan mapped to existing MVP tasks:

- **Phase 1**: Schema & Versioning (MVP-029)
- **Phase 2**: Lifecycle Management (MVP-029)
- **Phase 3**: Prioritization & Capacity (MVP-029)
- **Phase 4**: Dependency & Conflicts (MVP-031)
- **Phase 5**: Impact Analysis (MVP-032)
- **Phase 6**: Traceability (MVP-032)

## Success Metrics

### Technical Metrics
- Schema Compliance: 100%
- Lifecycle Violations: <1%
- Conflict Detection Accuracy: >95% true positive, <5% false negative
- Impact Analysis Performance: <2 seconds for 1000-node graph
- Traceability Completeness: >98% of active goals have complete chains

### Business Metrics
- Goal Approval Time: <5 days (reduced from 2 weeks)
- Resource Utilization: >80%
- WIP Limit Adherence: >95%
- Dependency Blocking: <10%
- Audit Trail Coverage: 100%

## Document Location

**Primary Specification**: `/workspaces/CodeValdCortex/documents/1-SoftwareRequirements/introduction/goals-specification.md`

**Updated Index**: `/workspaces/CodeValdCortex/documents/1-SoftwareRequirements/README.md`

## Next Steps

1. **Review**: Share specification with architecture team and stakeholders
2. **Validation**: Validate schema against existing agency-operations-framework.md
3. **Implementation**: Begin Phase 1 (Schema & Versioning) as part of MVP-029
4. **Integration**: Update MVP task descriptions to reference this specification
5. **Tooling**: Build code generation tools for TypeScript interfaces → ArangoDB schemas

---

**Document Status**: Complete  
**Feedback Coverage**: 100% (all 6 feedback areas addressed comprehensively)  
**Word Count**: ~8,500 words (comprehensive specification)  
**Code Examples**: 15+ TypeScript interfaces, 10+ ArangoDB queries, 20+ tables/matrices
