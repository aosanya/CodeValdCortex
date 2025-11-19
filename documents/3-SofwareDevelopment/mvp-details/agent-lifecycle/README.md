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

## Tasks in This Domain

| Task | Title | Priority | Status | Session |
|------|-------|----------|--------|---------|
| [MVP-033](MVP-033.md) | Agent Lifecycle FSM | P0 | Not Started | - |
| [MVP-034](MVP-034.md) | Run Execution FSM | P0 | Not Started | - |
| [MVP-035](MVP-035.md) | Health & Circuit Breakers | P0 | Not Started | - |
| [MVP-036](MVP-036.md) | Quarantine System | P0 | Not Started | - |

## Key Concepts

### Lifecycle States (10)
Agents transition through: Registered → Scheduled → Starting → Healthy → Degraded → Backoff → Draining → Quarantined → Stopped → Retired

### Run States (9)
Tasks execute through: Pending → Running → Waiting (I/O/HITL) → Succeeded/Failed → Compensating → Compensated/Orphaned

### Health Probes (4 types)
- **HTTP**: Endpoint health checks
- **TCP**: Socket connectivity
- **Exec**: Command execution
- **gRPC**: Service health protocol

### Circuit Breaker States (3)
- **Closed**: Normal operation
- **Open**: Fail fast protection
- **Half-Open**: Testing recovery

### Quarantine Triggers (5 categories)
Security violations, policy violations, anomaly detection, resource abuse, repeated failures

## Cross-Cutting Concerns

### State Persistence
All state machines persist state transitions to ArangoDB for durability and auditability:
- Agent lifecycle states: `agent_states` collection
- Run execution states: `run_states` collection
- State history with timestamps for compliance and debugging

### Event Publishing
State transitions publish events to the event bus:
- `agent.lifecycle.transitioned` - Agent state change
- `run.execution.transitioned` - Run state change
- `health.probe.failed` - Health probe failure
- `circuit.breaker.opened` - Circuit breaker opened
- `quarantine.triggered` - Agent quarantined

### Monitoring & Alerting
Key metrics:
- Agent state distribution (healthy vs degraded vs quarantined)
- Run success/failure rates
- Circuit breaker state changes
- Quarantine frequency and resolution times
- SLA compliance rates

## Detailed Documentation

- [Architecture Details](architecture/) - State diagrams, transition rules, detailed flows
- [Examples](examples/) - Configuration samples, API examples, integration patterns

## Implementation History

| Date | Session | Summary |
|------|---------|---------|
| TBD | TBD | Agent lifecycle management implementation |

## Reference

- Architecture: `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md`
