# Goals System Specification

## Overview

This document provides a comprehensive specification for the Goals system within CodeValdCortex, addressing schema definitions, lifecycle management, prioritization, dependency semantics, impact analysis, and traceability requirements.

## 1. Goal Schema & Versioning

### 1.1 Formal Goal Schema

**Goal Document Structure (ArangoDB Collection: `goals`)**:

```typescript
interface Goal {
  // Identity & Versioning
  _key: string;                    // ArangoDB document key (auto-generated)
  _id: string;                     // Full document ID: "goals/{_key}"
  _rev: string;                    // ArangoDB revision (for optimistic locking)
  
  // Core Identification
  code: string;                    // Business identifier (e.g., "GOAL-001", "RISK-ANALYSIS-01")
  version: SemanticVersion;        // Semantic version (major.minor.patch)
  status: GoalStatus;              // Current lifecycle state
  
  // Ownership & Accountability
  owner: {
    role: string;                  // RACI role (e.g., "Agency Lead", "Technical Lead")
    userId?: string;               // Optional user ID for individual assignment
    agencyId: string;              // Owning agency ID
  };
  
  // Goal Definition
  title: string;                   // Short, descriptive title (max 120 chars)
  description: string;             // Detailed goal description (Markdown supported)
  scope: {
    boundaries: string[];          // What is included in this goal
    constraints: string[];         // Limitations and restrictions
    assumptions: string[];         // Underlying assumptions
  };
  
  // Success Measurement
  successMetrics: SuccessMetric[]; // Key Performance Indicators (KPIs) / OKRs
  
  // Risk & Compliance
  riskLevel: RiskLevel;            // LOW, MEDIUM, HIGH, CRITICAL
  complianceTags: ComplianceTag[]; // Regulatory requirements
  dataClassification: DataClassification; // Data sensitivity level
  dataDomains: string[];           // Affected data domains (e.g., "customer", "financial")
  
  // Business Context
  businessValue: {
    category: BusinessValueCategory;  // Cost reduction, Revenue growth, etc.
    estimatedImpact: number;          // Quantitative estimate (currency, time saved, etc.)
    impactUnit: string;                // Unit of measurement (USD, hours, etc.)
    confidenceLevel: ConfidenceLevel; // LOW, MEDIUM, HIGH
  };
  
  // Non-Goals (Context-Specific)
  nonGoals: string[];              // Explicitly excluded objectives
  
  // Relationships
  dependencies: GoalDependency[];  // Dependencies on other goals
  
  // Metadata
  createdAt: ISODateTime;
  createdBy: string;
  updatedAt: ISODateTime;
  updatedBy: string;
  approvedAt?: ISODateTime;
  approvedBy?: string;
  
  // Change Control
  changeHistory: ChangeHistoryEntry[];
  migrationRules?: MigrationRule[];
}

// Supporting Types
type SemanticVersion = {
  major: number;  // Breaking changes to goal definition or success criteria
  minor: number;  // Additive changes (new success metrics, refined scope)
  patch: number;  // Clarifications, typo fixes, non-semantic changes
};

type GoalStatus = 
  | "draft"       // Initial creation, not yet proposed
  | "proposed"    // Submitted for review
  | "approved"    // Accepted by governance board
  | "active"      // Currently being pursued
  | "paused"      // Temporarily suspended
  | "completed"   // Success criteria met
  | "archived"    // No longer relevant/superseded
  | "rejected";   // Proposal declined

type RiskLevel = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";

interface ComplianceTag {
  framework: string;     // e.g., "SOC2", "HIPAA", "GDPR", "ISO27001"
  requirements: string[]; // Specific requirement IDs
  mandatory: boolean;     // Must-have vs nice-to-have
}

interface DataClassification {
  level: "PUBLIC" | "INTERNAL" | "CONFIDENTIAL" | "RESTRICTED";
  handlingRequirements: string[];
}

type BusinessValueCategory = 
  | "cost_reduction"
  | "revenue_growth"
  | "risk_mitigation"
  | "compliance_requirement"
  | "customer_satisfaction"
  | "operational_efficiency"
  | "innovation_enablement";

type ConfidenceLevel = "LOW" | "MEDIUM" | "HIGH";

interface SuccessMetric {
  id: string;
  name: string;
  description: string;
  type: "OKR" | "KPI" | "METRIC";
  target: {
    value: number;
    unit: string;
    operator: ">=" | ">" | "=" | "<" | "<=";
  };
  current?: number;
  measurementFrequency: "realtime" | "daily" | "weekly" | "monthly" | "quarterly";
  dataSource: string;  // Where this metric is collected from
}

interface GoalDependency {
  goalCode: string;
  dependencyType: "hard" | "soft";
  description: string;
  blockingRules: BlockingRule[];
}

interface BlockingRule {
  condition: string;  // e.g., "goal_status != 'completed'"
  action: "block_activation" | "show_warning" | "require_approval";
  message: string;
}

interface ChangeHistoryEntry {
  timestamp: ISODateTime;
  userId: string;
  changeType: "created" | "updated" | "status_change" | "approval" | "version_bump";
  previousVersion?: SemanticVersion;
  newVersion?: SemanticVersion;
  changes: Record<string, any>;  // JSON diff of changes
  reason: string;
}

interface MigrationRule {
  fromVersion: SemanticVersion;
  toVersion: SemanticVersion;
  migrationScript: string;  // Function name or script to transform data
  backwardCompatible: boolean;
}

type ISODateTime = string;  // ISO 8601 format
```

### 1.2 Versioning Rules (SemVer)

**Version Number Format**: `MAJOR.MINOR.PATCH`

**Version Increment Rules**:

| Change Type | Version Increment | Examples |
|-------------|------------------|----------|
| **Breaking Changes** | MAJOR | - Success criteria fundamentally changed<br>- Goal scope drastically expanded/reduced<br>- Owner/accountability transferred<br>- Goal merged with or split from others |
| **Additive Changes** | MINOR | - New success metrics added<br>- Scope refined without changing core objective<br>- New compliance tags added<br>- Risk level increased/decreased<br>- Dependencies added |
| **Non-Semantic Changes** | PATCH | - Typo fixes in description<br>- Formatting improvements<br>- Clarification without changing meaning<br>- Metadata updates (tags, classifications) |

**Versioning Workflow**:
1. Draft goals start at version `0.1.0`
2. First approval bumps to `1.0.0`
3. All subsequent changes follow SemVer rules
4. Version history maintained in `changeHistory` array

**Breaking Change Migration**:
- Major version changes require migration rules for dependent work items
- System validates all work item references are compatible with new version
- Automatic notification to work item owners when goal version changes

### 1.3 Schema Validation

**Validation Rules**:
- `code` must be unique within agency and match pattern: `^[A-Z0-9-]+$`
- `version` must follow SemVer format and increment monotonically
- `status` transitions must follow lifecycle state machine (see Section 2)
- `owner.role` must exist in agency RACI matrix
- `successMetrics` must have at least one entry for `active` status
- `complianceTags` must reference valid compliance frameworks
- `dependencies` must reference existing goal codes
- `riskLevel` must be assessed before approval
- `approvedBy` required when `status` = "approved" or "active"

## 2. Goal Lifecycle Management

### 2.1 Canonical States

**State Machine Definition**:

```
┌──────────┐
│  Draft   │◄────────────────────────────────┐
└────┬─────┘                                 │
     │ submit_for_review()                   │ reject()
     ▼                                       │
┌──────────┐                                 │
│ Proposed │─────────────────────────────────┤
└────┬─────┘                                 │
     │ approve()                             │
     ▼                                       │
┌──────────┐                          ┌──────────┐
│ Approved │─────────activate()──────►│  Active  │
└──────────┘                          └────┬─────┘
                                           │
                ┌──────────────────────────┼──────────────────────┐
                │ pause()                  │ complete()           │ archive()
                ▼                          ▼                      ▼
          ┌──────────┐             ┌──────────┐          ┌──────────┐
          │  Paused  │             │Completed │          │ Archived │
          └────┬─────┘             └──────────┘          └──────────┘
               │
               │ resume()
               ▼
          ┌──────────┐
          │  Active  │
          └──────────┘
```

**State Definitions**:

| State | Description | Entry Requirements | Exit Requirements | Allowed Actions |
|-------|-------------|-------------------|-------------------|-----------------|
| **Draft** | Initial creation, editing allowed | None | Submit for review | Edit all fields, delete, submit |
| **Proposed** | Submitted for governance review | - All required fields completed<br>- At least 1 success metric<br>- Risk level assessed | Approval or rejection | View, comment, request changes |
| **Approved** | Accepted by governance, ready for activation | - Governance approval<br>- Approver designated<br>- Dependencies validated | Activation | Activate, edit (creates new version) |
| **Active** | Currently being pursued, work items in progress | - All hard dependencies satisfied<br>- Resources allocated | Pause, complete, or archive | Pause, complete, edit (requires approval) |
| **Paused** | Temporarily suspended, may resume | Active status | Resume or archive | Resume, archive, view progress |
| **Completed** | Success criteria met | - All success metrics achieved<br>- Final report submitted | Archive (after review period) | View, generate report, archive |
| **Archived** | No longer relevant or superseded | Completed or obsolete | N/A (terminal state) | View only (read-only) |
| **Rejected** | Proposal declined by governance | Proposed status | Resubmit (as new draft) | View, copy to new draft |

### 2.2 State Transition Rules

**Required Approvals by Transition**:

| Transition | Approver Role | Quorum | Veto Rights |
|------------|--------------|--------|-------------|
| Draft → Proposed | Goal Owner | 1 | N/A |
| Proposed → Approved | Governance Board | Majority (51%) | Agency Lead, Compliance Officer |
| Proposed → Rejected | Governance Board | Majority (51%) | N/A |
| Approved → Active | Resource Allocator | 1 | N/A (automated if resources available) |
| Active → Paused | Agency Lead or Goal Owner | 1 | N/A |
| Active → Completed | Goal Owner + QA Lead | 2 (both required) | N/A |
| Active → Archived | Agency Lead | 1 | N/A |
| Paused → Active | Agency Lead | 1 | Resource Allocator (if resources unavailable) |
| Completed → Archived | Governance Board | 1 (any member) | N/A (after 30-day review period) |

### 2.3 Change Control

**Who Can Change What When**:

| Field/Action | Draft | Proposed | Approved | Active | Paused | Completed | Archived |
|--------------|-------|----------|----------|--------|--------|-----------|----------|
| Edit description | Owner | ❌ | Owner† | Owner† | Owner† | ❌ | ❌ |
| Edit scope | Owner | ❌ | Owner† | Owner† | Owner† | ❌ | ❌ |
| Add success metrics | Owner | ❌ | Owner† | Owner† | Owner† | ❌ | ❌ |
| Change risk level | Owner | ❌ | Compliance† | Compliance† | Compliance† | ❌ | ❌ |
| Add compliance tags | Owner | ❌ | Compliance† | Compliance† | Compliance† | ❌ | ❌ |
| Add dependencies | Owner | ❌ | Owner† | Owner† | ❌ | ❌ | ❌ |
| Change owner | Lead | Lead | Lead† | Lead† | Lead† | ❌ | ❌ |
| Delete | Owner | Owner | Lead | ❌ | ❌ | ❌ | ❌ |
| Add comments | Anyone | Anyone | Anyone | Anyone | Anyone | Anyone | View only |

**Legend**:
- ✓ = Allowed without approval
- † = Allowed with approval + version bump (minor or major)
- ❌ = Not allowed
- Owner = Goal owner
- Lead = Agency Lead
- Compliance = Compliance Officer

**Approval Requirements for Changes**:
- **Minor Changes** (PATCH version): Owner approval only
- **Moderate Changes** (MINOR version): Owner + 1 governance board member
- **Breaking Changes** (MAJOR version): Full governance board approval (majority vote)

## 3. Prioritization & Capacity Management

### 3.1 Priority Classes

**Priority Tiers** (inspired by Kubernetes QoS classes):

| Priority Class | Description | Resource Guarantee | Preemption | Use Cases |
|----------------|-------------|-------------------|------------|-----------|
| **P0 - Critical** | Business-critical, compliance-required | 100% (guaranteed resources) | Cannot be preempted | Regulatory compliance, system stability, security incidents |
| **P1 - High** | Important strategic goals | 80% (high allocation) | Can preempt P2/P3 | Strategic initiatives, major product features |
| **P2 - Medium** | Standard operational goals | 50% (fair share) | Can preempt P3 | Standard feature development, improvements |
| **P3 - Low** | Nice-to-have, experimental | Best effort (no guarantee) | Can be preempted by all | Research, optimization, technical debt |

**Priority Assignment Criteria**:
```typescript
interface PriorityAssignment {
  priorityClass: "P0" | "P1" | "P2" | "P3";
  factors: {
    businessValue: number;      // 0-100 (from businessValue.estimatedImpact)
    complianceRequired: boolean; // Compliance tags present
    riskLevel: RiskLevel;       // LOW/MEDIUM/HIGH/CRITICAL
    dependencies: number;        // Count of goals depending on this
    deadline?: ISODateTime;      // Hard deadline (if any)
  };
  calculatedScore: number;      // 0-100 (weighted combination)
  manualOverride?: {
    reason: string;
    approvedBy: string;
    date: ISODateTime;
  };
}
```

**Automatic Priority Calculation**:
```
Priority Score = (
  businessValue * 0.30 +
  (complianceRequired ? 40 : 0) +
  (riskLevel == CRITICAL ? 30 : riskLevel == HIGH ? 20 : riskLevel == MEDIUM ? 10 : 0) +
  (dependencies * 2) +  // 2 points per dependent goal
  (hasDeadline ? 10 : 0)
)

Priority Class Assignment:
- Score >= 80: P0 (Critical)
- Score >= 60: P1 (High)
- Score >= 30: P2 (Medium)
- Score < 30:  P3 (Low)
```

### 3.2 Work-In-Progress (WIP) Limits

**WIP Limit Definitions**:

```typescript
interface WIPLimits {
  agency: {
    maxActivateGoals: number;        // Total active goals across all priorities
    byPriority: {
      P0: number;  // No limit (infinite)
      P1: number;  // e.g., 5 goals
      P2: number;  // e.g., 10 goals
      P3: number;  // e.g., 3 goals
    };
    byOwner: {
      maxPerOwner: number;  // e.g., 3 active goals per owner
    };
  };
  
  violations: {
    action: "block" | "warn" | "require_approval";
    notifyRoles: string[];
  };
}
```

**WIP Enforcement Rules**:
1. **Hard Limits (Block)**: Cannot activate new goal if limit reached
2. **Soft Limits (Warn)**: Show warning but allow activation with justification
3. **Approval Override**: Agency Lead can override limits with documented reason

**Example WIP Configuration**:
```yaml
wip_limits:
  max_active_goals: 20
  by_priority:
    P0: unlimited
    P1: 5
    P2: 10
    P3: 3
  by_owner:
    max_per_owner: 3
  
  enforcement:
    P0_P1_violations: block      # Cannot exceed critical/high limits
    P2_P3_violations: warn       # Show warning for medium/low
    override_role: "Agency Lead"
```

### 3.3 Scheduling Policy (Agent Capacity Constraints)

**Capacity-Aware Scheduling Algorithm**:

```typescript
interface AgentCapacity {
  totalAgents: number;
  availableAgents: number;
  allocatedByGoal: Map<string, number>;  // goalCode -> agent count
  
  reservations: CapacityReservation[];
}

interface CapacityReservation {
  goalCode: string;
  priorityClass: "P0" | "P1" | "P2" | "P3";
  requestedAgents: number;
  guaranteedAgents: number;
  actualAgents: number;
}
```

**Scheduling Policy Rules**:

1. **Guaranteed Capacity** (P0/P1):
   - P0 goals get first claim on available agents (100% guarantee)
   - P1 goals get 80% of remaining capacity
   
2. **Fair Share** (P2):
   - P2 goals split remaining capacity equally
   
3. **Best Effort** (P3):
   - P3 goals use any leftover capacity
   - Can be preempted if P0/P1 goals activate

4. **Preemption Rules**:
   ```
   IF (P0 goal activates AND insufficient capacity):
     Preempt agents from P3 goals (lowest priority first)
     IF still insufficient:
       Preempt from P2 goals (least recently activated)
     IF still insufficient:
       Queue P0 goal and notify Agency Lead (manual decision)
   
   IF (P1 goal activates AND insufficient capacity):
     Preempt agents from P3 goals only
     IF still insufficient:
       Queue P1 goal or request capacity increase
   ```

5. **Capacity Monitoring**:
   - Real-time dashboard showing capacity utilization by priority
   - Alerts when utilization exceeds 85% for 24+ hours
   - Recommendations for capacity scaling or goal deprioritization

## 4. Dependency Semantics

### 4.1 Dependency Types

**Hard Dependencies** (Blocking):
```typescript
interface HardDependency {
  type: "hard";
  goalCode: string;
  blockingCondition: DependencyCondition;
  enforcementLevel: "strict" | "advisory";
}

type DependencyCondition = 
  | { type: "status"; requiredStatus: GoalStatus[] }
  | { type: "completion"; minimumProgress: number }  // 0-100%
  | { type: "metric"; metricId: string; minimumValue: number }
  | { type: "custom"; evaluationFunction: string };
```

**Soft Dependencies** (Informational):
```typescript
interface SoftDependency {
  type: "soft";
  goalCode: string;
  relationship: "benefits_from" | "related_to" | "follows";
  description: string;
  recommendedDelay?: number;  // Days to wait after dependency completes
}
```

### 4.2 Blocking Rules

**Rule Engine Configuration**:

```typescript
interface BlockingRule {
  ruleId: string;
  condition: DependencyCondition;
  action: BlockingAction;
  message: string;
  overrideRole?: string;  // Role that can override this rule
}

type BlockingAction = 
  | { type: "block_activation"; allowOverride: boolean }
  | { type: "show_warning"; requireAcknowledgment: boolean }
  | { type: "require_approval"; approverRoles: string[] }
  | { type: "auto_pause"; resumeWhen: DependencyCondition };
```

**Example Blocking Rules**:

```json
{
  "rules": [
    {
      "ruleId": "HARD_DEP_NOT_COMPLETE",
      "condition": {
        "type": "status",
        "requiredStatus": ["completed"]
      },
      "action": {
        "type": "block_activation",
        "allowOverride": false
      },
      "message": "Cannot activate: dependency {goalCode} must be completed first"
    },
    {
      "ruleId": "HARD_DEP_PROGRESS_INSUFFICIENT",
      "condition": {
        "type": "completion",
        "minimumProgress": 80
      },
      "action": {
        "type": "require_approval",
        "approverRoles": ["Agency Lead", "Technical Lead"]
      },
      "message": "Dependency {goalCode} only {actualProgress}% complete. Approval required to proceed."
    },
    {
      "ruleId": "SOFT_DEP_NOT_READY",
      "condition": {
        "type": "status",
        "requiredStatus": ["active", "completed"]
      },
      "action": {
        "type": "show_warning",
        "requireAcknowledgment": true
      },
      "message": "Recommended dependency {goalCode} is not yet active. Consider waiting."
    }
  ]
}
```

### 4.3 Conflict Detection

**Conflict Types**:

```typescript
interface GoalConflict {
  conflictType: ConflictType;
  goals: string[];  // Conflicting goal codes
  severity: "critical" | "high" | "medium" | "low";
  description: string;
  detectedAt: ISODateTime;
  resolutionRequired: boolean;
}

type ConflictType = 
  | "circular_dependency"       // A→B→C→A
  | "resource_contention"       // Both need same exclusive resource
  | "contradictory_objectives"  // Success criteria mutually exclusive
  | "scope_overlap"             // Duplicate/overlapping goals
  | "compliance_violation"      // Goals violate same compliance rule
  | "priority_inversion";       // Low priority blocks high priority
```

**Conflict Detection Algorithms**:

1. **Circular Dependency Detection** (Graph Cycle Detection):
   ```
   Algorithm: Depth-First Search with visited/stack tracking
   Complexity: O(V + E) where V=goals, E=dependencies
   Trigger: On dependency addition or goal activation
   ```

2. **Resource Contention Detection**:
   ```
   Algorithm: Check for exclusive resource claims in capacity reservations
   Trigger: On goal activation or resource allocation
   Action: Block lower-priority goal or require manual resolution
   ```

3. **Contradictory Objectives** (Semantic Analysis):
   ```
   Algorithm: NLP-based similarity + keyword contradiction detection
   Example Contradictions:
     - "increase security" vs "reduce latency" (inherent trade-off)
     - "minimize cost" vs "maximize features" (resource conflict)
   Trigger: On goal creation or major version change
   Action: Warn goal owner, require explicit trade-off documentation
   ```

4. **Scope Overlap Detection**:
   ```
   Algorithm: Cosine similarity of goal descriptions + scope keywords
   Threshold: >80% similarity triggers review
   Trigger: On goal proposal submission
   Action: Suggest merging goals or clarifying distinctions
   ```

**Conflict Resolution Workflow**:
```
Conflict Detected → Notify Owners → Review Meeting → Resolution Options:
  1. Merge Goals (if duplicate)
  2. Reprioritize (if resource contention)
  3. Adjust Dependencies (if circular)
  4. Document Trade-offs (if contradictory)
  5. Escalate to Governance Board (if unresolved)
```

## 5. Impact Analysis Specification

### 5.1 Graph Algorithms

**Algorithm Suite**:

| Algorithm | Purpose | Complexity | Implementation |
|-----------|---------|------------|----------------|
| **Reachability (DFS/BFS)** | Find all work items affected by goal change | O(V + E) | ArangoDB graph traversal |
| **Shortest Path (Dijkstra)** | Find most direct path from goal to work item | O(V² ) or O(E + V log V) | ArangoDB shortest_path |
| **Centrality (Betweenness)** | Identify critical goals/work items | O(V³) or O(VE) | Custom implementation |
| **Connected Components** | Find isolated goal clusters | O(V + E) | ArangoDB graph analysis |
| **Cut-Set Analysis** | Find minimum set of goals to remove to disconnect graph | O(V³) | Custom implementation |
| **Transitive Closure** | All indirect dependencies | O(V³) | ArangoDB pattern matching |

**Impact Analysis Query Examples**:

```javascript
// 1. Reachability: Find all work items affected by goal change
FOR v, e, p IN 1..10 OUTBOUND 'goals/GOAL-001' 
  goal_work_item_relationships
  OPTIONS {uniqueVertices: 'path'}
  FILTER IS_SAME_COLLECTION('work_items', v)
  RETURN {
    workItem: v,
    path: p.edges,
    impactLevel: e.impact_level
  }

// 2. Centrality: Find most critical goals (most work items depend on them)
FOR goal IN goals
  LET workItemCount = LENGTH(
    FOR v IN 1..1 OUTBOUND goal goal_work_item_relationships
      FILTER IS_SAME_COLLECTION('work_items', v)
      RETURN 1
  )
  SORT workItemCount DESC
  LIMIT 10
  RETURN {goal: goal.code, criticalityScore: workItemCount}

// 3. Cut-Set: Find goals that, if removed, would disconnect work items
FOR goal IN goals
  LET beforeConnectivity = LENGTH(
    FOR v IN 1..10 ANY 'work_items/WI-001' goal_work_item_relationships
      RETURN 1
  )
  LET afterConnectivity = LENGTH(
    FOR v IN 1..10 ANY 'work_items/WI-001' goal_work_item_relationships
      FILTER v._id != goal._id
      RETURN 1
  )
  FILTER afterConnectivity < beforeConnectivity
  RETURN {goal: goal.code, criticalityType: 'cut_vertex'}
```

### 5.2 False Positive/Negative Rates

**Acceptable Error Rates** (tuned per algorithm):

| Analysis Type | Acceptable False Positive Rate | Acceptable False Negative Rate | Rationale |
|---------------|--------------------------------|-------------------------------|-----------|
| **Reachability** | 5% | 1% | Prefer over-reporting impact (safer) |
| **Centrality** | 10% | 5% | Approximate rankings acceptable |
| **Conflict Detection** | 15% | 2% | Prefer false alarms over missed conflicts |
| **Scope Overlap** | 20% | 5% | Human review filters false positives |
| **Dependency Validation** | 0% | 0% | Must be deterministic (critical path) |

**Error Rate Monitoring**:
```typescript
interface AnalysisMetrics {
  algorithmName: string;
  totalRuns: number;
  falsePositives: number;      // User-reported incorrect results
  falseNegatives: number;      // User-reported missed results
  avgExecutionTime: number;    // milliseconds
  lastCalibration: ISODateTime;
  
  // Computed rates
  falsePositiveRate: number;   // falsePositives / totalRuns
  falseNegativeRate: number;   // falseNegatives / totalRuns
  
  // Alert if rates exceed thresholds
  alertTriggered: boolean;
}
```

**Calibration Process**:
1. **Monthly Review**: Sample 100 impact analyses, manually validate results
2. **Threshold Tuning**: Adjust similarity thresholds, graph traversal depth limits
3. **Algorithm Updates**: If error rates exceed acceptable levels, update implementation
4. **User Feedback Loop**: "Was this analysis helpful?" prompt after each impact report

### 5.3 User Overrides

**Override Mechanism**:

```typescript
interface ImpactAnalysisOverride {
  analysisId: string;
  overrideType: "exclude_item" | "include_item" | "adjust_impact_level" | "ignore_conflict";
  targetId: string;          // Goal or work item ID
  reason: string;
  overriddenBy: string;      // User ID
  overriddenAt: ISODateTime;
  expiresAt?: ISODateTime;   // Auto-revert if not renewed
}
```

**Override Use Cases**:
1. **False Positive Suppression**: "This work item is NOT affected by goal change"
2. **False Negative Correction**: "This work item IS affected but wasn't detected"
3. **Impact Level Adjustment**: "Change impact level from 'primary' to 'secondary'"
4. **Conflict Dismissal**: "This conflict is acknowledged and acceptable"

**Override Audit Trail**:
- All overrides logged in `changeHistory`
- Monthly report of overrides sent to governance board
- Overrides inform algorithm calibration (training data)

## 6. Traceability Contract

### 6.1 End-to-End Linkage Requirements

**Required Traceability Path**:
```
Business KPI → Goal → Work Item → Agent Action → Artifact/Log
```

**Traceability Schema**:

```typescript
interface TraceabilityChain {
  // Level 1: Business KPI
  businessKPI: {
    id: string;              // e.g., "KPI-2024-Q4-001"
    name: string;            // e.g., "Reduce fraud detection time by 60%"
    targetValue: number;
    currentValue: number;
    measurementUnit: string;
  };
  
  // Level 2: Goal
  goal: {
    code: string;            // e.g., "GOAL-001"
    title: string;
    successMetrics: SuccessMetric[];  // Must map to KPI
  };
  
  // Level 3: Work Item
  workItems: WorkItemTrace[];
  
  // Level 4: Agent Actions
  agentActions: AgentActionTrace[];
  
  // Level 5: Artifacts & Logs
  artifacts: ArtifactTrace[];
}

interface WorkItemTrace {
  workItemCode: string;
  relationshipType: "achieves" | "supports" | "enables" | "advances" | "mitigates";
  contributionPercentage: number;  // % of goal achieved by this work item
  status: "not_started" | "in_progress" | "completed";
}

interface AgentActionTrace {
  actionId: string;            // Deterministic ID
  agentId: string;
  workItemCode: string;
  actionType: string;          // e.g., "data_collection", "risk_calculation"
  timestamp: ISODateTime;
  inputData: Record<string, any>;
  outputData: Record<string, any>;
  executionMetrics: {
    duration: number;          // milliseconds
    resourcesUsed: {
      cpu: number;
      memory: number;
    };
  };
}

interface ArtifactTrace {
  artifactId: string;          // Deterministic ID
  artifactType: "log" | "report" | "dataset" | "model" | "code";
  actionId: string;            // Links to AgentActionTrace
  storageLocation: string;     // URI to artifact
  contentHash: string;         // SHA-256 for integrity
  createdAt: ISODateTime;
}
```

### 6.2 Deterministic ID Generation

**ID Format Standards**:

```typescript
// Goal IDs
goal._id = `goals/${agency_id}_${sequential_number}`
// Example: "goals/FINSERV_001"

// Work Item IDs
workItem._id = `work_items/${agency_id}_WI_${sequential_number}`
// Example: "work_items/FINSERV_WI_001"

// Agent Action IDs (deterministic from inputs)
actionId = SHA256(`${agentId}:${workItemCode}:${timestamp_ms}:${input_hash}`)
// Example: "agent_action_a3f9c8d7e2b1..."

// Artifact IDs
artifactId = `artifact_${action_id}_${artifact_type}_${sequence}`
// Example: "artifact_agent_action_a3f9c8d7e2b1_log_0001"

// Relationship IDs
relationshipId = `${goal_code}_${relationship_type}_${work_item_code}`
// Example: "GOAL-001_achieves_WI-001"
```

**ID Properties**:
- **Uniqueness**: Globally unique across all agencies
- **Determinism**: Same inputs always produce same ID
- **Readability**: Human-readable components for debugging
- **Traceability**: IDs encode parent relationships

### 6.3 Traceability Queries

**Common Traceability Queries**:

```javascript
// Query 1: KPI to Artifacts (full chain)
FOR kpi IN kpis FILTER kpi.id == 'KPI-2024-Q4-001'
  FOR goal IN goals FILTER goal.businessValue.kpiId == kpi.id
    FOR wi IN work_items 
      FOR v, e IN 1..1 OUTBOUND goal goal_work_item_relationships
        FILTER v._id == wi._id
        FOR action IN agent_actions FILTER action.workItemCode == wi.code
          FOR artifact IN artifacts FILTER artifact.actionId == action.actionId
            RETURN {
              kpi: kpi.name,
              goal: goal.code,
              workItem: wi.code,
              action: action.actionId,
              artifact: artifact.artifactId,
              chain: "complete"
            }

// Query 2: Artifact to KPI (reverse trace)
FOR artifact IN artifacts FILTER artifact.artifactId == 'artifact_xyz'
  FOR action IN agent_actions FILTER action.actionId == artifact.actionId
    FOR wi IN work_items FILTER wi.code == action.workItemCode
      FOR v, e IN 1..1 INBOUND wi goal_work_item_relationships
        FOR goal IN goals FILTER goal._id == v._id
          FOR kpi IN kpis FILTER kpi.id == goal.businessValue.kpiId
            RETURN {
              artifact: artifact.artifactId,
              action: action.actionId,
              workItem: wi.code,
              goal: goal.code,
              kpi: kpi.name,
              chain: "complete"
            }

// Query 3: Impact of Goal Change on KPIs
FOR goal IN goals FILTER goal.code == 'GOAL-001'
  FOR kpi IN kpis FILTER kpi.id == goal.businessValue.kpiId
    LET affectedWorkItems = (
      FOR v IN 1..1 OUTBOUND goal goal_work_item_relationships
        RETURN v.code
    )
    LET affectedActions = (
      FOR wi IN affectedWorkItems
        FOR action IN agent_actions FILTER action.workItemCode == wi
          RETURN action
    )
    RETURN {
      goal: goal.code,
      kpi: kpi.name,
      impactedWorkItems: LENGTH(affectedWorkItems),
      impactedActions: LENGTH(affectedActions),
      estimatedKPIChange: goal.businessValue.estimatedImpact
    }
```

### 6.4 Traceability Validation

**Validation Rules**:

```typescript
interface TraceabilityValidation {
  validationId: string;
  timestamp: ISODateTime;
  
  checks: {
    kpiToGoalLinkage: {
      passed: boolean;
      missingLinks: string[];  // KPI IDs without goal mappings
    };
    
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
  };
  
  overallStatus: "complete" | "incomplete" | "broken";
  recommendedActions: string[];
}
```

**Validation Schedule**:
- **Real-time**: On goal activation, work item creation, agent action logging
- **Daily**: Nightly batch validation of all active traceability chains
- **Monthly**: Comprehensive audit of archived goals and completed work items

**Broken Chain Resolution**:
1. **Detection**: Validation check identifies missing link
2. **Notification**: Alert goal owner and agency lead
3. **Remediation**: Manual review to restore link or mark as invalid
4. **Prevention**: Enforce foreign key constraints and required fields in schema

## 7. Implementation Roadmap

### Phase 1: Schema & Versioning (MVP-029)
- [ ] Implement formal goal schema in ArangoDB
- [ ] Add semantic versioning fields and validation
- [ ] Create migration rules for existing goals
- [ ] Implement change history tracking

### Phase 2: Lifecycle Management (MVP-029)
- [ ] Build state machine for goal lifecycle
- [ ] Implement approval workflow and role-based permissions
- [ ] Create change control rules and audit trail

### Phase 3: Prioritization & Capacity (MVP-029)
- [ ] Add priority classes and scoring algorithm
- [ ] Implement WIP limits and enforcement
- [ ] Build capacity-aware scheduling engine

### Phase 4: Dependency & Conflicts (MVP-031)
- [ ] Add dependency types (hard/soft) to schema
- [ ] Implement blocking rules and validation
- [ ] Build conflict detection algorithms (circular dependencies, resource contention)

### Phase 5: Impact Analysis (MVP-032)
- [ ] Implement graph traversal algorithms (reachability, centrality, cut-sets)
- [ ] Add false positive/negative rate monitoring
- [ ] Build user override mechanism

### Phase 6: Traceability (MVP-032)
- [ ] Implement end-to-end traceability schema
- [ ] Add deterministic ID generation
- [ ] Build traceability validation and reporting

## 8. Success Metrics

### Technical Metrics
- **Schema Compliance**: 100% of goals conform to formal schema
- **Lifecycle Violations**: <1% of state transitions violate rules
- **Conflict Detection Accuracy**: >95% true positive rate, <5% false negative rate
- **Impact Analysis Performance**: <2 seconds for 1000-node graph traversal
- **Traceability Completeness**: >98% of active goals have complete KPI→artifact chains

### Business Metrics
- **Goal Approval Time**: Reduce from average 2 weeks to <5 days
- **Resource Utilization**: Maintain >80% agent capacity utilization
- **WIP Limit Adherence**: >95% compliance with defined limits
- **Dependency Blocking**: <10% of goal activations blocked by unmet dependencies
- **Audit Trail Coverage**: 100% of goal changes logged with full context

---

**Document Version**: 1.0.0  
**Last Updated**: 2024-10-30  
**Owner**: CodeValdCortex Architecture Team  
**Status**: Draft - Pending Review
