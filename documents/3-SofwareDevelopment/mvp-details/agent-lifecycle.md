# Agent Lifecycle Management

## Overview

The agent lifecycle management system controls the complete journey of autonomous agents from registration through retirement. This domain encompasses finite state machines for both agent lifecycle and run execution, health monitoring, circuit breakers for external dependencies, and a comprehensive quarantine system for security and policy enforcement.

The lifecycle system ensures agents operate reliably, recover gracefully from failures, and maintain compliance with organizational policies throughout their operational lifetime.

## Architecture

The agent lifecycle is built on two complementary finite state machines:

1. **Agent Lifecycle FSM**: Controls the operational state of the agent itself (10 states)
2. **Run Execution FSM**: Manages individual task execution within agents (9 states)

These FSMs interact with supporting systems:
- **Health Probe Framework**: Monitors agent readiness and liveness
- **Circuit Breaker Service**: Protects against cascading failures from external dependencies
- **Quarantine System**: Isolates problematic agents and manages triage/recovery

**State Flow**:
```
Registration → Scheduling → Starting → Healthy ←→ Degraded
                                    ↓
                              Running Tasks
                                    ↓
                    (on issues) → Quarantined → Triage → Re-enablement
                                    ↓
                              Draining → Stopped → Retired
```

---

<!-- MVP-033 -->
## Agent Lifecycle FSM (MVP-033)

**Priority**: P0  
**Effort**: High  
**Dependencies**: MVP-032 (Work Items Assignment & Routing)  
**Status**: Not Started

### Overview

The Agent Lifecycle FSM implements a comprehensive 10-state finite state machine that governs the complete operational lifecycle of autonomous agents. This system ensures agents transition through well-defined states with proper validation, health monitoring, and timeout handling.

### Lifecycle States

1. **Registered**: Agent exists in system but not scheduled
2. **Scheduled**: Agent assigned to run, awaiting startup
3. **Starting**: Agent initialization in progress
4. **Healthy**: Agent operational and accepting work
5. **Degraded**: Agent experiencing issues but operational
6. **Backoff**: Agent in cooldown after failures
7. **Draining**: Agent completing existing work, not accepting new
8. **Quarantined**: Agent isolated due to security/policy issues
9. **Stopped**: Agent shut down gracefully
10. **Retired**: Agent permanently decommissioned

### State Transitions

**Valid Transitions**:
- `Registered → Scheduled`: Work assignment received
- `Scheduled → Starting`: Initialization triggered
- `Starting → Healthy`: Startup successful, health probes passing
- `Healthy ↔ Degraded`: Health status changes
- `Degraded → Backoff`: Circuit breaker opens or repeated failures
- `Backoff → Scheduled`: Cooldown period elapsed
- `{Healthy|Degraded} → Draining`: Graceful shutdown initiated
- `Any → Quarantined`: Security/policy violation detected
- `Draining → Stopped`: All work completed
- `Stopped → Retired`: Permanent decommission

**Transition Guards**: Each transition validates preconditions before execution to prevent invalid state changes.

### Health Probe Models

The system supports multiple probe types for comprehensive health monitoring:

1. **HTTP Probe**: Checks HTTP endpoint (GET/POST), validates status code and response body
2. **TCP Probe**: Verifies TCP socket connectivity
3. **Exec Probe**: Executes command in agent container, checks exit code
4. **gRPC Probe**: Calls gRPC health check service

**Probe Configuration**:
```json
{
  "type": "http|tcp|exec|grpc",
  "interval": "10s",
  "timeout": "5s",
  "successThreshold": 2,
  "failureThreshold": 3,
  "initialDelay": "15s"
}
```

**Probe Types**:
- **Liveness Probe**: Detects if agent is alive (restart if failing)
- **Readiness Probe**: Determines if agent can accept work

### Heartbeat & Timeout Monitoring

**Heartbeat System**:
- Agents send periodic heartbeats (configurable interval, default 30s)
- Missed heartbeats trigger degradation detection
- Consecutive missed heartbeats (default: 3) transition agent to `Degraded` or `Backoff`

**Timeout Handling**:
- **Startup Timeout**: Maximum time for `Starting → Healthy` (default: 5 minutes)
- **Drain Timeout**: Maximum time for `Draining → Stopped` (default: 15 minutes)
- **Quarantine Review Timeout**: SLA for triage completion (default: 24 hours)

### API Endpoints

1. **GET /agents/:id/state**: Retrieve current agent state
2. **POST /agents/:id/transition**: Trigger manual state transition (with guard validation)
3. **GET /agents/:id/state-history**: Retrieve state transition history
4. **GET /agents/:id/health**: Get current health probe results
5. **POST /agents/:id/heartbeat**: Record agent heartbeat

### Acceptance Criteria

- [ ] All 10 lifecycle states implemented with proper data models
- [ ] State transitions follow FSM rules with guard validation
- [ ] Health probes work for all 4 probe types (HTTP, TCP, Exec, gRPC)
- [ ] Timeouts trigger appropriate state transitions
- [ ] Heartbeat monitoring detects failures within configured thresholds
- [ ] Guards prevent invalid state transitions
- [ ] API endpoints return correct state and history information

### Reference

- Architecture: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` Section 1

<!-- /MVP-033 -->

---

<!-- MVP-034 -->
## Run Execution FSM (MVP-034)

**Priority**: P0  
**Effort**: High  
**Dependencies**: MVP-033 (Agent Lifecycle FSM)  
**Status**: Not Started

### Overview

The Run Execution FSM manages individual task execution within agents, implementing a 9-state finite state machine with retry/backoff logic, waiting states for external dependencies, and orphan recovery mechanisms. This complements the Agent Lifecycle FSM by controlling granular task-level execution.

### Run States

1. **Pending**: Run queued, awaiting execution slot
2. **Running**: Run actively executing
3. **Waiting I/O**: Run blocked on external I/O (API call, database query)
4. **Waiting HITL**: Run blocked awaiting human-in-the-loop decision
5. **Succeeded**: Run completed successfully
6. **Failed**: Run failed after exhausting retries
7. **Compensating**: Run executing compensation/rollback steps
8. **Compensated**: Compensation completed successfully
9. **Orphaned**: Run lost connection to agent (recovery needed)

### Retry Policy & Backoff

**RetryPolicy Configuration**:
```json
{
  "maxAttempts": 5,
  "initialDelay": "1s",
  "maxDelay": "60s",
  "multiplier": 2.0,
  "retryableErrors": ["TIMEOUT", "NETWORK_ERROR", "RATE_LIMIT"]
}
```

**Backoff Strategy**:
- **Exponential Backoff**: Delay = min(initialDelay × multiplier^attempt, maxDelay)
- **Jitter**: Add random jitter (0-25%) to prevent thundering herd
- **Circuit Breaker Integration**: Respect circuit breaker open state

**Retry Logic**:
1. Run fails with retryable error
2. Calculate backoff delay based on attempt count
3. Transition to `Pending` after delay
4. Increment attempt counter
5. Fail permanently if `maxAttempts` exceeded

### Wait Conditions

Runs can enter waiting states for various reasons:

1. **Waiting I/O**: External service call in progress
   - Timeout: Configurable per operation (default: 5 minutes)
   - Recovery: Retry or fail based on error type

2. **Waiting HITL**: Human decision required
   - Timeout: Escalation SLA (default: 4 hours)
   - Recovery: Escalate to manager or fail with timeout error

3. **Waiting Dependency**: Blocking on another run/agent
   - Timeout: Configurable dependency timeout (default: 30 minutes)
   - Recovery: Fail or retry based on dependency status

4. **Waiting Rate-Limit**: Throttled by rate limiter
   - Timeout: Rate limit window expiration
   - Recovery: Automatic retry when quota available

**Wait State Management**:
- Track wait start time and reason
- Monitor timeout expiration
- Transition back to `Running` when condition resolved
- Support cancellation during wait

### Orphan Detection & Recovery

**Orphan Scenarios**:
- Agent crashes during run execution
- Network partition between agent and orchestrator
- Agent quarantined mid-execution

**Detection Mechanisms**:
1. **Heartbeat Monitoring**: Agent fails to update run status within threshold (default: 2 minutes)
2. **Agent State Monitoring**: Agent transitions to `Quarantined`, `Stopped`, or `Retired`
3. **Execution Timeout**: Run exceeds maximum execution time (default: 1 hour)

**Recovery Strategies**:
1. **Idempotent Retry**: Restart run on different agent if operation is idempotent
2. **Manual Recovery**: Flag for human review if non-idempotent
3. **Compensation**: Execute rollback steps if partial completion detected
4. **Abandonment**: Mark as failed if unrecoverable

### Compensation & Saga Support

**Saga Pattern Implementation**:
- Runs define compensation steps for rollback
- Compensation steps execute in reverse order
- Each step can succeed, fail, or require manual intervention

**Compensation Model**:
```json
{
  "runId": "run-123",
  "compensationSteps": [
    {
      "stepId": "refund-payment",
      "status": "pending|succeeded|failed",
      "retryPolicy": {...}
    },
    {
      "stepId": "release-inventory",
      "status": "pending|succeeded|failed",
      "retryPolicy": {...}
    }
  ]
}
```

**State Transitions**:
- `Running → Compensating`: Failure detected, rollback initiated
- `Compensating → Compensated`: All compensation steps succeeded
- `Compensating → Failed`: Compensation steps failed (requires manual intervention)

### Run History & Metrics

**Tracked Metrics**:
- Execution duration (start to completion)
- State transition timestamps
- Retry attempts and delays
- Wait time in each waiting state
- Resource consumption (CPU, memory)
- Error types and frequencies

**History Schema**:
```json
{
  "runId": "run-123",
  "agentId": "agent-456",
  "states": [
    {"state": "Pending", "timestamp": "2025-11-19T10:00:00Z"},
    {"state": "Running", "timestamp": "2025-11-19T10:00:05Z"},
    {"state": "Succeeded", "timestamp": "2025-11-19T10:02:30Z"}
  ],
  "attempts": 1,
  "totalDuration": "2m25s",
  "waitDuration": "0s"
}
```

### Acceptance Criteria

- [ ] All 9 run states implemented with proper transitions
- [ ] Retry policy with exponential backoff and jitter works correctly
- [ ] Wait states handle all 4 wait types (I/O, HITL, dependency, rate-limit)
- [ ] Wait state timeouts trigger appropriate actions
- [ ] Orphan detection identifies lost runs within threshold
- [ ] Orphan recovery strategies execute correctly
- [ ] Compensation steps execute in reverse order
- [ ] Run history captures all state transitions and metrics

### Reference

- Architecture: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` Section 2

<!-- /MVP-034 -->

---

<!-- MVP-035 -->
## Health & Circuit Breakers (MVP-035)

**Priority**: P0  
**Effort**: Medium  
**Dependencies**: MVP-034 (Run Execution FSM)  
**Status**: Not Started

### Overview

The Health & Circuit Breakers system implements a comprehensive health probe framework supporting multiple probe types and integrates circuit breakers to protect against cascading failures from external dependencies. This system enables graceful degradation and automatic recovery.

### Health Probe Framework

**Supported Probe Types**:

1. **HTTP Probe**:
   ```json
   {
     "type": "http",
     "url": "http://localhost:8080/health",
     "method": "GET",
     "headers": {"Authorization": "Bearer token"},
     "expectedStatus": 200,
     "expectedBody": "ok|healthy",
     "timeout": "5s"
   }
   ```

2. **TCP Probe**:
   ```json
   {
     "type": "tcp",
     "host": "localhost",
     "port": 5432,
     "timeout": "3s"
   }
   ```

3. **Exec Probe**:
   ```json
   {
     "type": "exec",
     "command": ["/bin/sh", "-c", "pg_isready -U postgres"],
     "timeout": "10s",
     "expectedExitCode": 0
   }
   ```

4. **gRPC Probe**:
   ```json
   {
     "type": "grpc",
     "service": "grpc.health.v1.Health",
     "method": "Check",
     "address": "localhost:50051",
     "timeout": "5s"
   }
   ```

**Probe Execution**:
- Probes run at configured intervals
- Consecutive successes/failures tracked against thresholds
- Probe results published to agent state manager
- Failed probes trigger agent degradation or health warnings

### Circuit Breaker Service

**Circuit Breaker States**:

1. **Closed**: Normal operation, requests flow through
2. **Open**: Threshold exceeded, requests fail fast
3. **Half-Open**: Testing recovery, limited requests allowed

**State Transitions**:
```
Closed --[failure threshold exceeded]--> Open
Open --[timeout elapsed]--> Half-Open
Half-Open --[success threshold met]--> Closed
Half-Open --[failure detected]--> Open
```

**Configuration**:
```json
{
  "failureThreshold": 5,
  "successThreshold": 2,
  "timeout": "60s",
  "halfOpenMaxRequests": 3,
  "rollingWindow": "10s"
}
```

**Failure Detection**:
- Track success/failure ratio in rolling window
- Count consecutive failures
- Distinguish between error types (timeout, 5xx, connection refused)
- Open circuit when threshold exceeded

### Degradation Detection & Recovery

**Degradation Triggers**:
1. Circuit breaker opens for critical dependency (database, cache)
2. Health probes fail consecutively
3. Resource exhaustion detected (CPU >90%, memory >85%)
4. High error rate in run executions (>10% failure rate)

**Agent Degradation Response**:
- Transition agent to `Degraded` state
- Reduce work acceptance rate (e.g., 50% capacity)
- Increase health probe frequency
- Alert monitoring systems
- Attempt automatic recovery

**Recovery Process**:
1. Circuit breaker transitions to `Half-Open`
2. Limited test requests sent
3. If successful: Circuit `Closed`, agent returns to `Healthy`
4. If failed: Circuit remains `Open`, retry later

### Integration Points

**Protected Dependencies**:

1. **Database (ArangoDB)**:
   - Circuit breaker per collection
   - Health probe: TCP connection + query test
   - Degradation: Switch to read-only mode

2. **External APIs** (Gitea, GitHub, etc.):
   - Circuit breaker per endpoint
   - Health probe: HTTP GET /health
   - Degradation: Queue requests, reduce rate

3. **Message Queues** (if applicable):
   - Circuit breaker per queue
   - Health probe: Connection test
   - Degradation: Pause consumption

4. **Cache (Redis/In-Memory)**:
   - Circuit breaker for cache operations
   - Health probe: PING command
   - Degradation: Bypass cache, direct DB access

**Integration Architecture**:
```
Agent → Circuit Breaker → External Dependency
  ↓
Health Probe Framework
  ↓
Agent Lifecycle FSM (degradation detection)
```

### Monitoring Dashboard

**Dashboard Components**:

1. **Circuit Breaker Status Panel**:
   - Current state (Closed/Open/Half-Open) per dependency
   - Failure count and threshold
   - Time until next state transition

2. **Probe Results Panel**:
   - Last probe execution time
   - Success/failure status
   - Probe latency trends
   - Consecutive failure count

3. **Agent Health Overview**:
   - Agents by state (Healthy/Degraded/Backoff)
   - Health score per agent
   - Recent state transitions

4. **Dependency Health Matrix**:
   - Grid of agents × dependencies
   - Color-coded status (green/yellow/red)
   - Alert indicators

### Acceptance Criteria

- [ ] All 4 probe types (HTTP, TCP, Exec, gRPC) functional and configurable
- [ ] Circuit breaker opens when failure threshold exceeded
- [ ] Circuit breaker transitions to `Half-Open` after timeout
- [ ] `Half-Open` state tests recovery with limited requests
- [ ] Successful recovery transitions circuit to `Closed`
- [ ] Degradation detection triggers agent state change to `Degraded`
- [ ] Agent returns to `Healthy` when circuit breaker closes
- [ ] Monitoring dashboard displays real-time circuit and probe status
- [ ] Integration points (DB, APIs, queues, cache) protected by circuit breakers

### Reference

- Architecture: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` Sections 1.4-1.5

<!-- /MVP-035 -->

---

<!-- MVP-036 -->
## Quarantine System (MVP-036)

**Priority**: P0  
**Effort**: Medium  
**Dependencies**: MVP-035 (Health & Circuit Breakers)  
**Status**: Not Started

### Overview

The Quarantine System provides comprehensive isolation and triage capabilities for agents that violate security policies, exhibit anomalous behavior, or pose risks to system integrity. This system captures evidence, manages investigation workflows, and controls re-enablement with proper approvals.

### Quarantine Triggers

**Trigger Categories**:

1. **Security Violations**:
   - Unauthorized API access attempts
   - Suspicious file system access
   - Privilege escalation attempts
   - Credential theft indicators
   - Malicious code execution patterns

2. **Policy Violations**:
   - SLA breach (repeated deadline misses)
   - Resource quota exceeded (CPU, memory, network)
   - Rate limit abuse
   - Data access policy violations
   - Compliance rule violations (GDPR, SOC2)

3. **Anomaly Detection**:
   - Behavioral deviation from baseline
   - Unusual API call patterns
   - Unexpected network connections
   - Abnormal resource consumption
   - Error rate spikes

4. **Resource Abuse**:
   - CPU consumption >95% sustained
   - Memory leak detected
   - Disk space exhaustion
   - Network bandwidth saturation
   - Database connection pool exhaustion

5. **Repeated Failures**:
   - Consecutive run failures >10
   - Circuit breaker open >1 hour
   - Crash-restart loop detected
   - Persistent degraded state >24 hours

**Trigger Configuration**:
```json
{
  "triggerId": "security-unauthorized-access",
  "category": "security",
  "severity": "critical|high|medium|low",
  "autoQuarantine": true,
  "alertChannels": ["pagerduty", "slack"],
  "evidenceCapture": ["logs", "metrics", "network", "files"]
}
```

### Evidence Capture Model

**Evidence Types**:

1. **Logs**:
   - Agent execution logs (last 1000 lines)
   - System logs (kernel, container runtime)
   - Application logs (API calls, errors)
   - Audit logs (authorization checks)

2. **Metrics**:
   - CPU/memory/network usage (last 24 hours)
   - Request rates and latencies
   - Error rates by type
   - Resource consumption trends

3. **Network Activity**:
   - Outbound connections made
   - DNS queries
   - Firewall rule violations
   - Traffic volume and patterns

4. **File System**:
   - Modified files and timestamps
   - File permissions changes
   - Unexpected file creations
   - Configuration file diffs

5. **State Snapshots**:
   - Agent configuration
   - Environment variables
   - Run execution history
   - Recent state transitions

**Evidence Storage Schema**:
```json
{
  "quarantineId": "quar-123",
  "agentId": "agent-456",
  "triggeredAt": "2025-11-19T10:00:00Z",
  "trigger": {
    "type": "security",
    "reason": "Unauthorized API access",
    "severity": "critical"
  },
  "evidence": {
    "logs": "s3://evidence/quar-123/logs.tar.gz",
    "metrics": "s3://evidence/quar-123/metrics.json",
    "network": "s3://evidence/quar-123/network-capture.pcap",
    "files": "s3://evidence/quar-123/files-snapshot.tar.gz",
    "state": "s3://evidence/quar-123/state.json"
  },
  "capturedAt": "2025-11-19T10:00:15Z"
}
```

### Automated Triage Workflow

**Triage Process**:

1. **Initial Assessment** (Automated):
   - Classify severity (critical/high/medium/low)
   - Identify affected resources (agents, runs, data)
   - Calculate blast radius
   - Generate preliminary investigation tasks

2. **Assignment** (Automated):
   - Route to appropriate team based on trigger type:
     - Security team: Security violations
     - SRE team: Resource abuse, anomalies
     - Compliance team: Policy violations
   - Create investigation ticket (Gitea issue or internal tracking)
   - Set SLA based on severity:
     - Critical: 1 hour
     - High: 4 hours
     - Medium: 24 hours
     - Low: 7 days

3. **Investigation** (Manual):
   - Review evidence package
   - Analyze logs, metrics, network activity
   - Determine root cause
   - Assess security implications
   - Recommend remediation actions

4. **Resolution** (Manual):
   - Document findings
   - Apply fixes (code changes, config updates, policy adjustments)
   - Create post-mortem report
   - Update detection rules if needed

**Triage Workflow Schema**:
```json
{
  "triageId": "triage-789",
  "quarantineId": "quar-123",
  "status": "pending|in_progress|resolved|escalated",
  "assignedTeam": "security|sre|compliance",
  "assignedTo": "user-id",
  "sla": {
    "responseDeadline": "2025-11-19T11:00:00Z",
    "resolutionDeadline": "2025-11-19T14:00:00Z"
  },
  "findings": "Root cause analysis...",
  "remediation": "Actions taken...",
  "postMortemUrl": "https://docs/post-mortem-123"
}
```

### Re-enablement Process

**Re-enablement Checklist**:
- [ ] Root cause identified and documented
- [ ] Remediation actions completed and verified
- [ ] Code/config changes tested in staging
- [ ] Security team approval (for security violations)
- [ ] Compliance review (for policy violations)
- [ ] Post-mortem document published
- [ ] Detection rules updated (if applicable)
- [ ] Monitoring alerts configured

**Approval Requirements**:
- **Critical Severity**: Security team + Engineering manager approval
- **High Severity**: Engineering manager approval
- **Medium Severity**: Team lead approval
- **Low Severity**: Automated re-enablement after remediation

**Re-enablement API**:
```http
POST /quarantine/:id/re-enable
{
  "approvals": [
    {"role": "security", "userId": "user-123", "approvedAt": "2025-11-19T12:00:00Z"},
    {"role": "manager", "userId": "user-456", "approvedAt": "2025-11-19T12:05:00Z"}
  ],
  "checklistCompleted": true,
  "postMortemUrl": "https://docs/post-mortem-123",
  "monitoringPlan": "Enhanced monitoring for 7 days..."
}
```

**Re-enablement Process**:
1. Verify all approvals received
2. Confirm checklist completed
3. Apply any configuration changes
4. Transition agent from `Quarantined` to `Scheduled`
5. Enable enhanced monitoring (increased probe frequency)
6. Notify stakeholders of re-enablement

### SLA Tracking & Escalation

**SLA Metrics**:
- **Time to Acknowledge**: Quarantine detected → Investigation started
- **Time to Resolve**: Investigation started → Root cause identified
- **Time to Re-enable**: Root cause identified → Agent operational

**SLA Thresholds** (by severity):
| Severity | Acknowledge | Resolve | Re-enable |
|----------|-------------|---------|-----------|
| Critical | 15 minutes  | 1 hour  | 4 hours   |
| High     | 1 hour      | 4 hours | 24 hours  |
| Medium   | 4 hours     | 24 hours| 7 days    |
| Low      | 24 hours    | 7 days  | 14 days   |

**Escalation Triggers**:
- SLA threshold exceeded by 50%
- No investigation progress for 2x SLA window
- Critical severity quarantine not acknowledged within 15 minutes
- Multiple quarantines from same root cause

**Escalation Actions**:
1. Notify manager/on-call engineer
2. Page incident response team
3. Create high-priority ticket
4. Schedule emergency review meeting
5. Auto-assign to senior engineer if unassigned

### Post-Mortem Documentation

**Required Sections**:
1. **Incident Summary**: What happened, when, impact
2. **Timeline**: Key events from detection to resolution
3. **Root Cause**: Technical explanation of underlying issue
4. **Contributing Factors**: Environment, config, dependencies
5. **Resolution**: Steps taken to fix
6. **Prevention**: How to prevent recurrence
7. **Action Items**: Follow-up tasks with owners and deadlines

**Post-Mortem Template**:
```markdown
# Quarantine Post-Mortem: [Agent ID] - [Trigger Type]

## Summary
- **Quarantine ID**: quar-123
- **Agent ID**: agent-456
- **Trigger**: Unauthorized API access
- **Severity**: Critical
- **Duration**: 2025-11-19 10:00 - 12:30 (2h 30m)
- **Impact**: 1 agent quarantined, 5 runs failed

## Timeline
- 10:00: Anomaly detected, agent quarantined
- 10:05: Evidence captured, security team notified
- 10:30: Investigation started
- 11:15: Root cause identified - misconfigured API key
- 11:45: Fix deployed to staging
- 12:00: Fix tested and approved
- 12:30: Agent re-enabled with updated config

## Root Cause
[Detailed technical explanation]

## Prevention
- [ ] Update API key rotation policy
- [ ] Add validation for API credentials
- [ ] Implement pre-deployment config checks

## Action Items
- [X] Update detection rule (owner: security-team, due: 2025-11-20)
- [ ] Document API key best practices (owner: docs-team, due: 2025-11-25)
```

### Acceptance Criteria

- [ ] Quarantine triggers isolate agents correctly for all 5 trigger categories
- [ ] Evidence capture includes all required data types (logs, metrics, network, files, state)
- [ ] Automated triage workflow assigns investigations to correct teams
- [ ] Triage workflow tracks investigation status and SLA compliance
- [ ] Re-enablement requires all approvals based on severity
- [ ] Re-enablement checklist enforced before agent returns to service
- [ ] SLA tracking monitors all three time metrics (acknowledge, resolve, re-enable)
- [ ] Escalation triggers fire when SLA thresholds exceeded
- [ ] Post-mortem documentation generated with all required sections

### Reference

- Architecture: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` Section 3

<!-- /MVP-036 -->

---

## Cross-Cutting Concerns

### State Persistence

All state machines persist state transitions to ArangoDB for durability and auditability:
- Agent lifecycle states stored in `agent_states` collection
- Run execution states stored in `run_states` collection
- State history with timestamps for compliance and debugging

### Event Publishing

State transitions publish events to the event bus for monitoring and integration:
- `agent.lifecycle.transitioned` - Agent state change
- `run.execution.transitioned` - Run state change
- `health.probe.failed` - Health probe failure
- `circuit.breaker.opened` - Circuit breaker opened
- `quarantine.triggered` - Agent quarantined

### Monitoring & Alerting

Key metrics and alerts:
- Agent state distribution (healthy vs degraded vs quarantined)
- Run success/failure rates
- Circuit breaker state changes
- Quarantine frequency and resolution times
- SLA compliance rates

## Implementation History

| Date | Session | Summary |
|------|---------|---------|
| TBD | TBD | Agent lifecycle management implementation |
