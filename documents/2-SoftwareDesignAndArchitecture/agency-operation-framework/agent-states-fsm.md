# Agent States Finite State Machine Specification

## Overview

This document provides a comprehensive specification for the Agent States system within CodeValdCortex, defining formal finite state machines (FSM) for both agent lifecycle management and task/run execution. It addresses state definitions, transitions, guards, timeouts, health monitoring, quarantine rules, and deployment rollout strategies.

## 1. Agent Lifecycle FSM

### 1.1 Formal State Definitions

**Agent Lifecycle States**:

```typescript
type AgentLifecycleState = 
  | "registered"      // Initial registration, awaiting validation
  | "scheduled"       // Assigned to work, awaiting resource allocation
  | "starting"        // Initialization in progress
  | "healthy"         // Fully operational, passing all health checks
  | "degraded"        // Operational but with warnings/reduced capacity
  | "backoff"         // Temporarily unavailable due to repeated failures
  | "draining"        // Graceful shutdown, completing in-flight tasks
  | "quarantined"     // Isolated due to policy violation or anomalies
  | "stopped"         // Intentionally stopped, can be restarted
  | "retired"         // Permanently decommissioned (terminal state);

interface AgentLifecycleMetadata {
  agentId: string;
  currentState: AgentLifecycleState;
  previousState?: AgentLifecycleState;
  stateEnteredAt: ISODateTime;
  stateDuration: number;  // milliseconds in current state
  
  // Health & Performance
  healthStatus: {
    lastHeartbeat: ISODateTime;
    liveness: ProbeResult;
    readiness: ProbeResult;
    degradationReasons?: string[];
  };
  
  // Failure Tracking
  failureMetrics: {
    consecutiveFailures: number;
    totalFailures: number;
    lastFailureAt?: ISODateTime;
    backoffUntil?: ISODateTime;
    backoffMultiplier: number;  // Current exponential backoff level
  };
  
  // Quarantine Details
  quarantineInfo?: QuarantineRecord;
  
  // Capacity & Load
  capacity: {
    maxConcurrentTasks: number;
    currentActiveTasks: number;
    utilizationPercentage: number;
  };
  
  // Metadata
  version: string;
  deploymentStrategy?: DeploymentStrategy;
  createdAt: ISODateTime;
  updatedAt: ISODateTime;
}

interface ProbeResult {
  status: "pass" | "fail" | "unknown";
  lastChecked: ISODateTime;
  consecutiveFailures: number;
  message?: string;
}
```

### 1.2 State Machine Diagram

```
┌─────────────┐
│ Registered  │◄──────────────────────────────────┐
└──────┬──────┘                                   │
       │ validate()                               │
       ▼                                          │
┌─────────────┐                                   │
│  Scheduled  │                                   │
└──────┬──────┘                                   │
       │ allocate_resources()                     │
       ▼                                          │
┌─────────────┐                                   │
│  Starting   │──────────start_failed()──────────►│
└──────┬──────┘                                   │
       │ startup_complete()                       │
       ▼                                          │
┌─────────────┐                                   │
│   Healthy   │◄───────recover()─────────┐        │
└──────┬──────┘                           │        │
       │                                  │        │
       ├──degrade()──►┌──────────┐       │        │
       │              │ Degraded  │───────┘        │
       │              └─────┬─────┘                │
       │                    │                      │
       │                    │ fail_repeatedly()    │
       │                    ▼                      │
       │              ┌──────────┐                 │
       │              │ Backoff  │─────────────────┤
       │              └─────┬────┘                 │
       │                    │ retry_allowed()      │
       │                    └──────────────────────┘
       │
       ├──violate_policy()──►┌─────────────┐
       │                      │ Quarantined │──────re_enable()────►┐
       │                      └─────────────┘                      │
       │                                                           │
       ├──drain()────────────►┌──────────┐                        │
       │                      │ Draining │──────complete()────────┤
       │                      └──────────┘                        │
       │                                                           │
       └──stop()─────────────►┌──────────┐                        │
                               │ Stopped  │◄───────────────────────┘
                               └────┬─────┘
                                    │ restart()
                                    ▼
                               ┌──────────┐
                               │ Starting │
                               └────┬─────┘
                                    │ retire()
                                    ▼
                               ┌──────────┐
                               │ Retired  │  (Terminal State)
                               └──────────┘
```

### 1.3 State Transition Rules

**Transition Table**:

| From State | Event/Trigger | Guards | To State | Actions |
|------------|--------------|--------|----------|---------|
| **Registered** | `validate()` | - Schema valid<br>- Identity verified<br>- Capabilities declared | Scheduled | - Assign to pool<br>- Reserve resources |
| **Scheduled** | `allocate_resources()` | - Resources available<br>- No conflicts | Starting | - Allocate CPU/memory<br>- Initialize runtime |
| **Starting** | `startup_complete()` | - Liveness probe passes<br>- Readiness probe passes<br>- Startup timeout not exceeded | Healthy | - Mark ready for tasks<br>- Register in load balancer |
| **Starting** | `start_failed()` | - Liveness probe fails<br>- Startup timeout exceeded<br>- Initialization error | Registered | - Log failure<br>- Increment failure count<br>- Apply backoff if retries exhausted |
| **Healthy** | `degrade()` | - Performance degradation detected<br>- Readiness probe fails<br>- Error rate above threshold | Degraded | - Reduce task allocation<br>- Trigger alerts<br>- Begin diagnostics |
| **Healthy** | `violate_policy()` | - Security policy violation<br>- Anomaly score > threshold<br>- Unauthorized action | Quarantined | - Isolate agent<br>- Capture evidence<br>- Notify security team |
| **Healthy** | `drain()` | - Graceful shutdown requested<br>- Version upgrade pending | Draining | - Stop accepting new tasks<br>- Wait for in-flight completion |
| **Healthy** | `stop()` | - Manual stop requested<br>- Emergency shutdown | Stopped | - Terminate immediately<br>- Save state |
| **Degraded** | `recover()` | - Readiness probe passes<br>- Error rate normalized<br>- Performance restored | Healthy | - Restore full capacity<br>- Clear warnings |
| **Degraded** | `fail_repeatedly()` | - Consecutive failures > max_retries<br>- Liveness probe fails | Backoff | - Calculate backoff duration<br>- Increment backoff multiplier |
| **Backoff** | `retry_allowed()` | - Backoff period expired<br>- Manual override | Registered | - Reset backoff multiplier (if successful)<br>- Attempt restart |
| **Backoff** | `max_backoff_reached()` | - Backoff multiplier > max_backoff<br>- Manual intervention required | Quarantined | - Escalate to ops team<br>- Require manual review |
| **Draining** | `complete()` | - All in-flight tasks finished<br>- Drain timeout not exceeded | Stopped | - Release resources<br>- Unregister from pool |
| **Draining** | `drain_timeout()` | - Drain timeout exceeded<br>- Tasks still running | Stopped | - Force terminate tasks<br>- Log incomplete tasks |
| **Quarantined** | `re_enable()` | - Triage completed<br>- Root cause remediated<br>- Security approval granted | Registered | - Clear quarantine flag<br>- Reset anomaly score |
| **Stopped** | `restart()` | - Manual restart<br>- Scheduled restart | Starting | - Re-allocate resources<br>- Begin initialization |
| **Stopped** | `retire()` | - Permanent decommission<br>- Version EOL reached | Retired | - Archive historical data<br>- Remove from registry |

### 1.4 Timeouts & Probes

**Timeout Configuration**:

```typescript
interface AgentTimeouts {
  startup: {
    duration: number;        // e.g., 60000 (60 seconds)
    action: "retry" | "quarantine";
  };
  
  heartbeat: {
    interval: number;        // e.g., 5000 (5 seconds)
    missedThreshold: number; // e.g., 3 consecutive misses
    action: "degrade" | "backoff";
  };
  
  drain: {
    duration: number;        // e.g., 300000 (5 minutes)
    action: "force_stop";
  };
  
  backoff: {
    initial: number;         // e.g., 1000 (1 second)
    max: number;             // e.g., 300000 (5 minutes)
    multiplier: number;      // e.g., 2.0 (exponential)
    maxRetries: number;      // e.g., 10
  };
  
  quarantine: {
    minDuration: number;     // e.g., 3600000 (1 hour)
    requiresManualReview: boolean;
  };
}
```

**Health Probes**:

```typescript
interface HealthProbe {
  type: "liveness" | "readiness";
  method: "http" | "tcp" | "exec" | "grpc";
  config: ProbeConfig;
  
  // Probe behavior
  initialDelay: number;      // Delay before first probe (milliseconds)
  interval: number;          // Time between probes (milliseconds)
  timeout: number;           // Probe timeout (milliseconds)
  successThreshold: number;  // Consecutive successes to pass
  failureThreshold: number;  // Consecutive failures to fail
}

interface ProbeConfig {
  // HTTP probe
  path?: string;
  port?: number;
  headers?: Record<string, string>;
  expectedStatusCode?: number;
  
  // TCP probe
  tcpPort?: number;
  
  // Exec probe
  command?: string[];
  
  // gRPC probe
  grpcService?: string;
}
```

**Probe Behavior**:

| Probe Type | Purpose | Failure Action | Success Criteria |
|------------|---------|----------------|------------------|
| **Liveness** | Agent is alive and not deadlocked | Restart agent (Starting → Healthy) | HTTP 200, TCP connect, command exit 0 |
| **Readiness** | Agent is ready to accept tasks | Remove from task pool (Healthy → Degraded) | HTTP 200 + custom readiness logic |

**Example Probe Configuration**:

```yaml
health_probes:
  liveness:
    type: http
    path: /health/live
    port: 8080
    initial_delay: 10s
    interval: 10s
    timeout: 5s
    failure_threshold: 3
  
  readiness:
    type: http
    path: /health/ready
    port: 8080
    initial_delay: 5s
    interval: 5s
    timeout: 3s
    failure_threshold: 2
    success_threshold: 1
```

### 1.5 Circuit Breaker Integration

**Circuit Breaker States** (for external dependencies):

```typescript
interface CircuitBreaker {
  name: string;
  state: "closed" | "open" | "half_open";
  
  thresholds: {
    failureRate: number;       // e.g., 0.5 (50%)
    slowCallRate: number;      // e.g., 0.5 (50%)
    slowCallDuration: number;  // milliseconds
    minimumCalls: number;      // Before evaluating thresholds
  };
  
  timings: {
    waitDurationInOpen: number;      // Before trying half_open
    permittedCallsInHalfOpen: number; // Test calls allowed
  };
  
  metrics: {
    totalCalls: number;
    failedCalls: number;
    slowCalls: number;
    lastStateChange: ISODateTime;
  };
}
```

**Circuit Breaker Impact on Agent State**:

| Circuit State | Agent Action | Rationale |
|---------------|--------------|-----------|
| **Closed** | Normal operation | All calls allowed |
| **Open** | Transition to Degraded if critical dependency | Prevent cascading failures |
| **Half-Open** | Monitor for recovery | Test if dependency recovered |

## 2. Task/Run Execution FSM

### 2.1 Run State Definitions

**Run Lifecycle States**:

```typescript
type RunState = 
  | "pending"        // Queued, awaiting agent assignment
  | "running"        // Actively executing
  | "waiting_io"     // Blocked on external I/O (API call, database query)
  | "waiting_hitl"   // Blocked on Human-in-the-Loop approval
  | "succeeded"      // Completed successfully
  | "failed"         // Execution failed
  | "compensating"   // Running compensation logic (saga rollback)
  | "compensated"    // Compensation completed successfully
  | "orphaned";      // Agent died, run state unknown (terminal state)

interface RunMetadata {
  runId: string;
  workItemCode: string;
  agentId: string;
  currentState: RunState;
  previousState?: RunState;
  
  // Timing
  queuedAt: ISODateTime;
  startedAt?: ISODateTime;
  completedAt?: ISODateTime;
  duration?: number;  // milliseconds
  
  // Execution Context
  executionAttempt: number;  // Retry count
  maxRetries: number;
  
  // Wait States
  waitingFor?: WaitCondition;
  
  // Compensation
  compensationPlan?: CompensationPlan;
  compensationRunId?: string;
  
  // Results
  output?: Record<string, any>;
  error?: ErrorRecord;
  
  // Metadata
  createdAt: ISODateTime;
  updatedAt: ISODateTime;
}

interface WaitCondition {
  type: "io" | "hitl" | "dependency" | "rate_limit";
  description: string;
  expectedResolution: ISODateTime;
  timeoutAt: ISODateTime;
  
  // Type-specific details
  ioDetails?: {
    endpoint: string;
    requestId: string;
    retryAfter?: number;
  };
  
  hitlDetails?: {
    approverRole: string;
    approverUserId?: string;
    requestedAt: ISODateTime;
    escalationPolicy: EscalationPolicy;
  };
}

interface CompensationPlan {
  compensationSteps: CompensationStep[];
  strategy: "sequential" | "parallel" | "custom";
}

interface CompensationStep {
  stepId: string;
  action: string;
  agentId?: string;
  status: "pending" | "running" | "completed" | "failed";
}

interface ErrorRecord {
  code: string;
  message: string;
  stackTrace?: string;
  retriable: boolean;
  category: "transient" | "permanent" | "unknown";
}
```

### 2.2 Run State Machine Diagram

```
┌─────────┐
│ Pending │
└────┬────┘
     │ assign_agent()
     ▼
┌─────────┐
│ Running │◄────────resume()──────────┐
└────┬────┘                            │
     │                                 │
     ├──wait_io()───►┌─────────────┐  │
     │               │ Waiting I/O │──┘
     │               └─────────────┘
     │                                 
     ├──wait_hitl()─►┌──────────────┐ 
     │               │ Waiting HITL │──┐
     │               └──────────────┘  │
     │                     │            │ approve()
     │                     │ timeout()  │
     │                     ▼            │
     │               ┌──────────┐      │
     │               │  Failed  │      │
     │               └────┬─────┘      │
     │                    │             │
     ├──succeed()───►┌───────────┐     │
     │               │ Succeeded │◄────┘
     │               └───────────┘
     │
     ├──fail()──────►┌──────────┐
     │               │  Failed  │
     │               └────┬─────┘
     │                    │ compensate()
     │                    ▼
     │               ┌──────────────┐
     │               │ Compensating │
     │               └──────┬───────┘
     │                      │
     │                      ├──success──►┌─────────────┐
     │                      │             │ Compensated │
     │                      │             └─────────────┘
     │                      │
     │                      └──fail──────►┌──────────┐
     │                                    │  Failed  │
     │                                    └──────────┘
     │
     └──agent_died()────────────────────►┌──────────┐
                                          │ Orphaned │
                                          └──────────┘
```

### 2.3 Run Transition Rules

**Transition Table**:

| From State | Event/Trigger | Guards | To State | Actions |
|------------|--------------|--------|----------|---------|
| **Pending** | `assign_agent()` | - Agent available<br>- Agent has required skills<br>- Agent not quarantined | Running | - Reserve agent capacity<br>- Load execution context<br>- Start execution timer |
| **Running** | `wait_io()` | - External I/O call initiated<br>- Async operation started | Waiting I/O | - Save execution state<br>- Release agent for other tasks<br>- Set timeout timer |
| **Running** | `wait_hitl()` | - Human approval required<br>- HITL gate triggered | Waiting HITL | - Notify approver<br>- Set escalation timer<br>- Save decision context |
| **Running** | `succeed()` | - All steps completed<br>- Output validation passed<br>- No errors | Succeeded | - Save output<br>- Release agent<br>- Trigger downstream work |
| **Running** | `fail()` | - Execution error<br>- Validation failed<br>- Timeout exceeded | Failed | - Log error<br>- Check retry policy<br>- Initiate compensation if needed |
| **Running** | `agent_died()` | - Agent heartbeat lost<br>- Agent quarantined<br>- Agent crashed | Orphaned | - Mark run as orphaned<br>- Attempt state recovery<br>- Reassign if idempotent |
| **Waiting I/O** | `resume()` | - I/O response received<br>- Timeout not exceeded | Running | - Restore execution context<br>- Continue execution |
| **Waiting I/O** | `timeout()` | - I/O timeout exceeded<br>- No response received | Failed | - Log timeout error<br>- Check retry policy |
| **Waiting HITL** | `approve()` | - Approver decision received<br>- Decision = approve | Running | - Log approval<br>- Continue execution |
| **Waiting HITL** | `reject()` | - Approver decision received<br>- Decision = reject | Failed | - Log rejection<br>- Execute rejection handler |
| **Waiting HITL** | `timeout()` | - Escalation timeout exceeded<br>- No response | Failed | - Execute escalation policy<br>- Auto-fail or reassign |
| **Failed** | `retry()` | - Retriable error<br>- Retry count < max_retries<br>- Backoff period expired | Pending | - Increment retry count<br>- Apply exponential backoff<br>- Re-queue |
| **Failed** | `compensate()` | - Compensation plan exists<br>- Saga pattern enabled<br>- Side effects to rollback | Compensating | - Load compensation plan<br>- Execute compensation steps |
| **Compensating** | `complete()` | - All compensation steps succeeded<br>- Rollback complete | Compensated | - Mark as fully rolled back<br>- Log compensation success |
| **Compensating** | `fail()` | - Compensation step failed<br>- Manual intervention required | Failed | - Escalate to ops team<br>- Require manual cleanup |

### 2.4 Retry & Backoff Logic

**Retry Policy Configuration**:

```typescript
interface RetryPolicy {
  maxAttempts: number;           // e.g., 3
  backoffStrategy: "fixed" | "exponential" | "custom";
  
  // Fixed backoff
  fixedDelay?: number;           // milliseconds
  
  // Exponential backoff
  initialDelay?: number;         // e.g., 1000 (1 second)
  maxDelay?: number;             // e.g., 60000 (1 minute)
  multiplier?: number;           // e.g., 2.0
  jitter?: boolean;              // Add randomness to prevent thundering herd
  
  // Retry conditions
  retryOn: ErrorCategory[];      // e.g., ["transient", "rate_limit"]
  doNotRetryOn: ErrorCode[];     // e.g., ["INVALID_INPUT", "UNAUTHORIZED"]
}
```

**Exponential Backoff Formula**:

```
delay = min(maxDelay, initialDelay * (multiplier ^ attemptNumber))

If jitter enabled:
  delay = delay * (1.0 + random(-0.2, 0.2))  // ±20% randomness
```

**Example Retry Configurations**:

```yaml
# Transient failures (network, timeout)
transient_retry_policy:
  max_attempts: 5
  backoff_strategy: exponential
  initial_delay: 1000      # 1s
  max_delay: 60000         # 60s
  multiplier: 2.0
  jitter: true
  retry_on: [transient, timeout, rate_limit]

# Non-retriable failures
permanent_failure_policy:
  max_attempts: 1
  backoff_strategy: fixed
  do_not_retry_on: [invalid_input, unauthorized, forbidden]
```

### 2.5 Orphaned Run Recovery

**Orphan Detection**:

```typescript
interface OrphanDetection {
  detectionMethod: "heartbeat" | "agent_state" | "timeout";
  
  // Heartbeat-based
  heartbeatInterval: number;     // e.g., 5000 (5 seconds)
  missedHeartbeats: number;      // e.g., 3 consecutive misses
  
  // Timeout-based
  maxExecutionTime: number;      // e.g., 3600000 (1 hour)
  
  // Recovery strategy
  recoveryAction: "reassign" | "fail" | "manual_review";
  reassignConditions: {
    idempotent: boolean;         // Can safely retry
    stateRecoverable: boolean;   // State saved & can be restored
    withinRetryLimit: boolean;   // Haven't exceeded max retries
  };
}
```

**Orphan Recovery Process**:

```
1. Detect orphaned run (agent died, heartbeat lost)
   ↓
2. Analyze run characteristics
   - Is the operation idempotent?
   - Is the state recoverable from checkpoints?
   - Are there side effects to consider?
   ↓
3. Decision:
   a) Idempotent + State Recoverable → Reassign to new agent
   b) Non-idempotent → Mark as Failed, initiate compensation
   c) Uncertain → Manual Review (alert ops team)
   ↓
4. Execute recovery action
   ↓
5. Update traceability (log orphan event, new agent assignment)
```

## 3. Quarantine System

### 3.1 Quarantine Triggers

**Policy Violation Categories**:

```typescript
interface QuarantineRule {
  ruleId: string;
  category: QuarantineCategory;
  severity: "low" | "medium" | "high" | "critical";
  trigger: QuarantineTrigger;
  action: QuarantineAction;
}

type QuarantineCategory = 
  | "security_violation"        // Unauthorized access, credential leak
  | "policy_violation"          // Compliance rule breach
  | "anomaly_detection"         // Behavioral anomaly score threshold
  | "resource_abuse"            // CPU/memory/network abuse
  | "data_exfiltration"         // Unauthorized data access
  | "repeated_failures"         // Excessive failure rate
  | "manual_quarantine";        // Operator-initiated

interface QuarantineTrigger {
  type: "threshold" | "pattern" | "manual";
  
  // Threshold-based
  metric?: string;               // e.g., "anomaly_score"
  operator?: ">=" | ">" | "=" | "<" | "<=";
  value?: number;
  
  // Pattern-based
  pattern?: string;              // Regex or rule expression
  
  // Manual
  initiatedBy?: string;          // User ID
  reason?: string;
}

interface QuarantineAction {
  isolate: boolean;              // Remove from task pool
  captureEvidence: boolean;      // Snapshot state, logs, memory
  notifyTeam: string[];          // Security, ops, compliance
  requireTriage: boolean;        // Requires manual review
  autoRemediation?: string;      // Script/function to attempt fix
}
```

**Example Quarantine Rules**:

```yaml
quarantine_rules:
  - rule_id: QR-001
    category: security_violation
    severity: critical
    trigger:
      type: pattern
      pattern: "unauthorized_api_call|credential_exposure"
    action:
      isolate: true
      capture_evidence: true
      notify_team: [security, ops]
      require_triage: true
  
  - rule_id: QR-002
    category: anomaly_detection
    severity: high
    trigger:
      type: threshold
      metric: anomaly_score
      operator: ">="
      value: 0.9
    action:
      isolate: true
      capture_evidence: true
      notify_team: [ops]
      require_triage: true
      auto_remediation: "restart_with_clean_state"
  
  - rule_id: QR-003
    category: repeated_failures
    severity: medium
    trigger:
      type: threshold
      metric: failure_rate_5m
      operator: ">="
      value: 0.8
    action:
      isolate: true
      capture_evidence: false
      notify_team: [ops]
      require_triage: false
      auto_remediation: "exponential_backoff"
```

### 3.2 Evidence Capture

**Evidence Collection**:

```typescript
interface QuarantineEvidence {
  evidenceId: string;
  agentId: string;
  quarantineRuleId: string;
  capturedAt: ISODateTime;
  
  // State Snapshot
  agentState: {
    lifecycleState: AgentLifecycleState;
    healthStatus: ProbeResult;
    activeRuns: RunMetadata[];
    memorySnapshot?: string;     // Base64-encoded memory dump
  };
  
  // Logs
  recentLogs: LogEntry[];        // Last 1000 log lines
  errorLogs: LogEntry[];         // All error-level logs
  
  // Metrics
  performanceMetrics: {
    cpu: number;
    memory: number;
    network: NetworkMetrics;
    diskIO: DiskMetrics;
  };
  
  // Security Context
  securityEvents: SecurityEvent[];
  accessPatterns: AccessLog[];
  
  // Audit Trail
  recentActions: AuditLogEntry[];
  
  // Storage
  storageLocation: string;       // S3/GCS URI
  retentionPolicy: {
    duration: number;            // milliseconds
    deleteAfter: ISODateTime;
  };
}

interface SecurityEvent {
  eventType: string;
  severity: "info" | "warning" | "critical";
  description: string;
  timestamp: ISODateTime;
  sourceIp?: string;
  targetResource?: string;
}
```

### 3.3 Triage Workflow

**Triage Process**:

```
┌──────────────────┐
│ Quarantine Event │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Evidence Capture │
└────────┬─────────┘
         │
         ▼
┌──────────────────────┐
│ Automated Analysis   │  ← ML/rule-based classification
│ - Root cause         │
│ - Severity           │
│ - Blast radius       │
└────────┬─────────────┘
         │
         ├──► Auto-remediable? ──Yes──► Apply Auto-Remediation ──► Monitor
         │                                                            │
         No                                                           │
         │                                                            │
         ▼                                                            ▼
┌──────────────────────┐                                    Success? ──Yes──► Re-enable
│ Assign to Triage     │                                            │
│ - Security analyst   │                                            No
│ - Ops engineer       │                                            │
│ - Compliance officer │                                            ▼
└────────┬─────────────┘                                    Escalate to Manual Triage
         │
         ▼
┌──────────────────────┐
│ Manual Investigation │
│ - Review evidence    │
│ - Reproduce issue    │
│ - Determine root cause│
└────────┬─────────────┘
         │
         ▼
┌──────────────────────┐
│ Remediation Decision │
│ - Re-enable          │
│ - Retire             │
│ - Require update     │
└────────┬─────────────┘
         │
         ▼
┌──────────────────────┐
│ Post-Incident Review │
│ - Update rules       │
│ - Improve detection  │
└──────────────────────┘
```

**Triage SLAs**:

| Severity | Initial Response | Resolution Target | Escalation |
|----------|------------------|-------------------|------------|
| **Critical** | 15 minutes | 2 hours | After 1 hour to VP Eng |
| **High** | 1 hour | 8 hours | After 4 hours to Eng Manager |
| **Medium** | 4 hours | 24 hours | After 12 hours to Team Lead |
| **Low** | 24 hours | 7 days | After 3 days to Team Lead |

### 3.4 Re-enablement Process

**Re-enablement Checklist**:

```typescript
interface ReenablementApproval {
  agentId: string;
  quarantineId: string;
  
  // Pre-requisites
  checks: {
    rootCauseIdentified: boolean;
    remediationApplied: boolean;
    testingCompleted: boolean;
    securityApproval: boolean;    // Required for security violations
    complianceApproval: boolean;  // Required for policy violations
    peerReview: boolean;          // Required for manual code changes
  };
  
  // Approvals
  approvals: {
    role: string;
    userId: string;
    approvedAt: ISODateTime;
    comments?: string;
  }[];
  
  // Conditions
  reenablementConditions: {
    gradualRollout: boolean;       // Canary deployment
    monitoringPeriod: number;      // Extended monitoring (milliseconds)
    rollbackTriggers: string[];    // Auto-rollback conditions
  };
  
  // Documentation
  postMortemCompleted: boolean;
  lessonsLearned: string[];
  preventiveMeasures: string[];
}
```

**Re-enablement Workflow**:

```
1. Complete all checklist items
   ↓
2. Gather approvals (based on quarantine category)
   ↓
3. Apply remediation (patch, config change, code update)
   ↓
4. Test in isolated environment
   ↓
5. Deploy with gradual rollout (canary)
   ↓
6. Extended monitoring period (e.g., 24 hours)
   ↓
7. If stable → Full re-enablement
   If issues → Auto-rollback → Return to triage
```

## 4. Deployment Rollout Strategies

### 4.1 Rollout Methods

**Supported Deployment Strategies**:

```typescript
interface DeploymentStrategy {
  strategyType: "blue_green" | "canary" | "progressive_delivery" | "rolling";
  
  // Blue-Green
  blueGreenConfig?: {
    switchoverType: "instant" | "gradual";
    warmupPeriod: number;        // milliseconds
    rollbackType: "instant" | "manual";
  };
  
  // Canary
  canaryConfig?: {
    stages: CanaryStage[];
    promoteOnSuccess: boolean;
    autoRollbackOnFailure: boolean;
  };
  
  // Progressive Delivery
  progressiveConfig?: {
    initialPercentage: number;   // e.g., 1%
    incrementPercentage: number; // e.g., 10%
    incrementInterval: number;   // milliseconds
    targetPercentage: number;    // e.g., 100%
    pauseOnError: boolean;
  };
  
  // Rolling Update
  rollingConfig?: {
    maxSurge: number;            // Max new instances above desired
    maxUnavailable: number;      // Max instances unavailable during update
    batchSize: number;           // Instances to update per batch
  };
  
  // Rollback Policy
  rollbackPolicy: RollbackPolicy;
}

interface CanaryStage {
  name: string;
  percentage: number;            // % of traffic to new version
  duration: number;              // milliseconds to run this stage
  successCriteria: SuccessCriteria;
}

interface SuccessCriteria {
  errorRateThreshold: number;    // e.g., 0.01 (1%)
  latencyP95Threshold: number;   // milliseconds
  successRateThreshold: number;  // e.g., 0.99 (99%)
  customMetrics?: CustomMetric[];
}

interface CustomMetric {
  metricName: string;
  operator: ">=" | ">" | "=" | "<" | "<=";
  threshold: number;
}

interface RollbackPolicy {
  autoRollback: boolean;
  rollbackTriggers: RollbackTrigger[];
  rollbackSpeed: "instant" | "gradual";
  notifyOnRollback: string[];    // Roles to notify
}

interface RollbackTrigger {
  metric: string;
  operator: ">=" | ">" | "=" | "<" | "<=";
  threshold: number;
  evaluationWindow: number;      // milliseconds
}
```

### 4.2 Blue-Green Deployment

**Blue-Green Flow**:

```
┌─────────────┐                    ┌─────────────┐
│    Blue     │ ◄─── 100% Traffic  │   Router    │
│ (Current)   │                    └─────────────┘
└─────────────┘
       │
       │ Deploy new version
       ▼
┌─────────────┐
│    Green    │
│    (New)    │ ◄─── Warmup & Health Checks
└─────────────┘
       │
       │ Health checks pass
       ▼
┌─────────────┐                    ┌─────────────┐
│    Blue     │                    │   Router    │
│ (Current)   │                    └──────┬──────┘
└─────────────┘                           │
                                          │ Switch traffic
┌─────────────┐                           ▼
│    Green    │ ◄─── 100% Traffic  ┌─────────────┐
│    (New)    │                    │   Router    │
└─────────────┘                    └─────────────┘
       │
       │ Monitor for issues
       │
       ├──► Success → Decommission Blue
       └──► Failure → Rollback to Blue (instant)
```

**Implementation**:

```typescript
interface BlueGreenDeployment {
  deploymentId: string;
  blueVersion: string;           // Current version
  greenVersion: string;          // New version
  
  // Phases
  phases: {
    warmup: {
      status: "pending" | "in_progress" | "completed";
      startedAt?: ISODateTime;
      completedAt?: ISODateTime;
      healthChecks: ProbeResult[];
    };
    
    switchover: {
      status: "pending" | "in_progress" | "completed";
      startedAt?: ISODateTime;
      completedAt?: ISODateTime;
      trafficPercentage: number;  // 0-100
    };
    
    monitoring: {
      status: "in_progress" | "completed" | "failed";
      startedAt?: ISODateTime;
      duration: number;            // Observation period
      metrics: DeploymentMetrics;
    };
  };
  
  // Rollback
  rollbackAvailable: boolean;
  rollbackExecuted: boolean;
}
```

### 4.3 Canary Deployment

**Canary Flow** (Multi-Stage):

```
Stage 1: 5% → Monitor → Success? → Stage 2
                            ↓ No
                            Rollback
                            
Stage 2: 25% → Monitor → Success? → Stage 3
                             ↓ No
                             Rollback
                             
Stage 3: 50% → Monitor → Success? → Stage 4
                             ↓ No
                             Rollback
                             
Stage 4: 100% → Monitor → Success → Complete
                              ↓ No
                              Rollback
```

**Implementation**:

```yaml
canary_deployment:
  stages:
    - name: initial_canary
      percentage: 5
      duration: 10m
      success_criteria:
        error_rate_threshold: 0.01      # 1%
        latency_p95_threshold: 500      # 500ms
        success_rate_threshold: 0.99    # 99%
    
    - name: expand_canary
      percentage: 25
      duration: 20m
      success_criteria:
        error_rate_threshold: 0.005     # 0.5%
        latency_p95_threshold: 400
    
    - name: majority_canary
      percentage: 50
      duration: 30m
      success_criteria:
        error_rate_threshold: 0.005
        latency_p95_threshold: 400
    
    - name: full_rollout
      percentage: 100
      duration: 60m
      success_criteria:
        error_rate_threshold: 0.01
        latency_p95_threshold: 500
  
  auto_rollback: true
  rollback_triggers:
    - metric: error_rate
      operator: ">="
      threshold: 0.02               # 2%
      evaluation_window: 5m
    
    - metric: latency_p95
      operator: ">="
      threshold: 1000               # 1s
      evaluation_window: 5m
```

### 4.4 Progressive Delivery

**Progressive Delivery Flow**:

```
Start: 1% of agents
   ↓
Monitor for 5 minutes
   ↓
Success? ─No─► Pause & Alert
   │ Yes
   ▼
Increase to 10%
   ↓
Monitor for 5 minutes
   ↓
Success? ─No─► Auto-Rollback
   │ Yes
   ▼
Continue increasing by 10% every 5 minutes
   ↓
Reach 100%
   ↓
Extended monitoring (24 hours)
   ↓
Complete
```

**Implementation**:

```typescript
interface ProgressiveDelivery {
  deploymentId: string;
  currentPercentage: number;
  targetPercentage: number;
  
  // Configuration
  config: {
    initialPercentage: number;     // e.g., 1
    incrementPercentage: number;   // e.g., 10
    incrementInterval: number;     // e.g., 300000 (5 minutes)
    pauseOnError: boolean;
    autoRollbackOnFailure: boolean;
  };
  
  // Current state
  currentStage: {
    stageNumber: number;
    percentage: number;
    startedAt: ISODateTime;
    metrics: DeploymentMetrics;
    status: "monitoring" | "promoting" | "paused" | "rolling_back";
  };
  
  // Decision engine
  decisionEngine: {
    evaluationInterval: number;    // milliseconds
    lastEvaluation: ISODateTime;
    decision: "promote" | "pause" | "rollback" | "continue";
    reason: string;
  };
}
```

### 4.5 SLO-Based Rollback

**Error Budget Integration**:

```typescript
interface ErrorBudget {
  slo: {
    targetAvailability: number;    // e.g., 0.999 (99.9%)
    targetErrorRate: number;       // e.g., 0.001 (0.1%)
    targetLatencyP95: number;      // milliseconds
  };
  
  budget: {
    totalBudget: number;           // Total error budget for period
    consumedBudget: number;        // Budget consumed so far
    remainingBudget: number;       // Budget remaining
    budgetPeriod: "day" | "week" | "month";
  };
  
  status: "healthy" | "warning" | "exhausted";
  
  // Rollback policy
  rollbackThresholds: {
    warningThreshold: number;      // e.g., 0.8 (80% budget consumed)
    criticalThreshold: number;     // e.g., 0.95 (95% budget consumed)
    action: "alert" | "pause_rollout" | "auto_rollback";
  };
}
```

**Rollback Decision Logic**:

```
Every evaluation interval (e.g., 1 minute):
  1. Collect metrics (error rate, latency, success rate)
  2. Compare against success criteria
  3. Calculate error budget consumption rate
  4. Decision:
     - All metrics within thresholds + budget healthy → Promote
     - One metric violates threshold + budget warning → Pause
     - Critical violation + budget exhausted → Auto-Rollback
     - Manual override → Execute override action
```

**Example Rollback Configuration**:

```yaml
rollback_policy:
  auto_rollback: true
  
  rollback_triggers:
    # Error rate spike
    - metric: error_rate_5m
      operator: ">="
      threshold: 0.02               # 2%
      evaluation_window: 5m
      action: auto_rollback
    
    # Latency degradation
    - metric: latency_p95_5m
      operator: ">="
      threshold: 1000               # 1s
      evaluation_window: 5m
      action: auto_rollback
    
    # Success rate drop
    - metric: success_rate_5m
      operator: "<="
      threshold: 0.95               # 95%
      evaluation_window: 5m
      action: pause_rollout
    
    # Error budget exhaustion
    - metric: error_budget_remaining
      operator: "<="
      threshold: 0.05               # 5%
      action: pause_rollout
  
  rollback_speed: instant
  notify_on_rollback: [ops, engineering_manager, on_call]
```

## 5. Implementation Roadmap

### Phase 1: Agent Lifecycle FSM (MVP-033)
- [ ] Implement formal agent lifecycle states and transitions
- [ ] Build state machine validation and enforcement
- [ ] Add timeout configuration and monitoring
- [ ] Implement liveness/readiness probes
- [ ] Create heartbeat monitoring system

### Phase 2: Run Execution FSM (MVP-034)
- [ ] Implement run lifecycle states (pending → running → succeeded/failed)
- [ ] Add waiting states (I/O, HITL) with timeout handling
- [ ] Build retry/backoff logic with exponential backoff
- [ ] Implement orphaned run detection and recovery
- [ ] Create compensation/saga execution engine

### Phase 3: Health & Circuit Breakers (MVP-035)
- [ ] Implement health probe framework (HTTP, TCP, exec, gRPC)
- [ ] Build circuit breaker integration for external dependencies
- [ ] Add degradation detection and auto-recovery
- [ ] Create agent capacity monitoring dashboard

### Phase 4: Quarantine System (MVP-036)
- [ ] Implement quarantine triggers (policy, anomaly, failure rate)
- [ ] Build evidence capture system (state, logs, metrics)
- [ ] Create triage workflow with SLA tracking
- [ ] Implement re-enablement approval process

### Phase 5: Deployment Rollouts (MVP-037)
- [ ] Implement blue-green deployment strategy
- [ ] Build canary deployment with multi-stage rollout
- [ ] Add progressive delivery with auto-promotion
- [ ] Create SLO-based rollback with error budget tracking
- [ ] Build deployment metrics dashboard

## 6. Success Metrics

### Technical Metrics
- **State Transition Accuracy**: 100% of transitions follow FSM rules
- **Health Probe Latency**: <100ms for liveness/readiness checks
- **Orphan Recovery Rate**: >95% of orphaned runs successfully recovered or compensated
- **Quarantine False Positive Rate**: <5% of quarantines are false positives
- **Rollback Speed**: <60 seconds for instant rollback, <5 minutes for gradual

### Operational Metrics
- **Agent Uptime**: >99.9% healthy state for production agents
- **Mean Time to Quarantine**: <5 minutes from violation to isolation
- **Mean Time to Re-enable**: <2 hours for high-severity, <24 hours for medium
- **Deployment Success Rate**: >98% of deployments complete without rollback
- **Error Budget Compliance**: <10% of deployments exceed error budget

### Business Metrics
- **Incident Reduction**: 50% reduction in agent-related incidents within 6 months
- **Deployment Velocity**: 3x increase in safe deployment frequency
- **Operational Cost**: 30% reduction in manual intervention for agent issues

---

**Document Version**: 1.0.0  
**Last Updated**: 2024-10-30  
**Owner**: CodeValdCortex Architecture Team  
**Status**: Draft - Pending Review
