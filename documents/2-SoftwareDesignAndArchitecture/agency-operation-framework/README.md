# Agency Operation Framework

This directory contains the comprehensive specifications for CodeValdCortex's Agency Operation Framework, which defines how agencies, agents, and work items operate within the system.

## Documents

### 1. [Agency Operations Framework](./agency-operations-framework.md)
**Purpose**: High-level overview and foundational concepts

**Contains**:
- Goals module structure and cataloging
- Work Items (WI) definition and components
- Goal-Work Item relationship mapping (graph-based)
- RACI Matrix framework for work items
- Implementation guidelines and examples

---

### 2. [Work Items Specification](./work-items.md)
**Purpose**: Detailed specification for work item types, lifecycle, and execution

**Contains**:
- **Work Item Types**: Task, Job, Investigation, Change, Remediation, Experiment
- **Lifecycle & SLAs**: State machine, timers, breach actions, escalation
- **Assignment & Routing**: Declarative rules, skills matching, cost budgets, HITL checkpoints
- **Concurrency Controls**: Idempotence keys, mutex scopes, reentrancy contracts
- **Compensation & Sagas**: Rollback strategies, orchestration patterns
- **Validation & Approvals**: Policy-driven gates, evidence capture
- **Templating**: Industry templates (PCI, HIPAA, SOC2), versioning, parameterization
- **Traceability**: Validation schema, deterministic ID generation
- **Implementation Roadmap**: Phased MVP tasks (MVP-030 through MVP-037)

---

### 3. [Role Taxonomy](./role-taxonomy.md)
**Purpose**: Comprehensive taxonomy for agent capabilities, constraints, and governance

**Contains**:
- **Role Classifications**:
  - Stateless Tool-Caller
  - Planner/Coordinator
  - Data Access Agent
  - Long-Running Service
  - Sensor/Monitor
  - Actuator
  - Reviewer/HITL Proxy
  
- **Skills & Tools Contract**: Tool adapters, rate limits, side effects, cost tracking
- **Autonomy Levels (L0-L4)**: Manual to full autonomy with policy-bound scopes
- **Budgeting**: Token/$ limits, compute quotas, exhaustion behaviors
- **Data Boundaries**: Allowed datasets, PII masking, data residency controls
- **Safety Constraints**: Allowed/prohibited actions, two-person rule
- **Identity & Tenancy**: OIDC/SPIFFE workload identity, attestation, tenant isolation
- **Complete Examples**: Full role specification with all taxonomy elements

---

## Document Relationships

```
┌─────────────────────────────────────────────────┐
│  Agency Operations Framework                    │
│  (High-level concepts & RACI)                  │
└────────────┬────────────────────────────────────┘
             │
             ├──────────────────┬─────────────────┐
             │                  │                 │
             ▼                  ▼                 ▼
┌────────────────────┐  ┌────────────────┐  ┌──────────────────┐
│   Work Items       │  │  Roles   │  │  Goals Module    │
│   Specification    │◄─┤   Taxonomy     │  │  (future doc)    │
└────────────────────┘  └────────────────┘  └──────────────────┘
         │                       │
         │  agents execute work items
         └───────────────────────┘
```

## Key Concepts

### Goals
Strategic objectives that agencies pursue. Goals are tracked, versioned, and linked to work items that achieve them.

### Work Items
Discrete units of work (tasks, jobs, changes, etc.) that contribute to achieving goals. Work items have lifecycles, SLAs, and are executed by agents.

### Agents
AI/software entities that execute work items. Agents have types, capabilities, autonomy levels, and operate within defined constraints.

### RACI Matrix
Role assignment framework (Responsible, Accountable, Consulted, Informed) that clarifies accountability for each activity within a work item.

## Implementation Status

| Component | Status | MVP |
|-----------|--------|-----|
| Work Item Types & Schema | 📝 Specified | MVP-030 |
| Lifecycle & SLA | 📝 Specified | MVP-031 |
| Assignment & Routing | 📝 Specified | MVP-032 |
| Concurrency Controls | 📝 Specified | MVP-033 |
| Compensation & Sagas | 📝 Specified | MVP-034 |
| Policy Gates & Evidence | 📝 Specified | MVP-035 |
| Templates & Catalog | 📝 Specified | MVP-036 |
| Traceability & Validation | 📝 Specified | MVP-037 |
| Role Taxonomy | 📝 Specified | MVP-030 |

**Legend**: 📝 Specified | 🚧 In Progress | ✅ Complete

## Usage

These documents serve as:
1. **Design Reference**: For implementing agency operation features
2. **API Contracts**: Schema definitions for data models and APIs
3. **Governance Guide**: Policies for autonomy, safety, and compliance
4. **Implementation Roadmap**: Phased MVP tasks with clear deliverables

## Related Documentation

- **Software Requirements**: `/documents/1-SoftwareRequirements/`
- **Backend Architecture**: `/documents/2-SoftwareDesignAndArchitecture/backend-architecture.md`
- **Use Case Definitions**: `/usecases/`
- **Coding Sessions**: `/documents/3-SofwareDevelopment/coding_sessions/`

---

**Last Updated**: 2025-10-30  
**Maintained By**: CodeValdCortex Architecture Team  
**Version**: 1.0.0
