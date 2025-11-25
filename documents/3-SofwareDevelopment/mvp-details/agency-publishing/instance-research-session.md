# Research Session: Agency Instance Management (MVP-PUB-007)

**Date**: November 24, 2025  
**Status**: In Progress  
**Focus**: Architecture clarification and design decisions

---

## Session Goals

Explore and document the architecture, business logic, and implementation details of the Agency Instance Management system to ensure complete understanding before finalizing documentation.

---

## Questions & Answers

### Q1: Instance Startup State Transition

**Category**: Architecture & Design  
**Status**: ✅ Answered

**Question**: When an instance transitions from "starting" to "running" state, what specific criteria must be met?

**Context**: Documentation mentions instances start in "running" state immediately on creation, but this seemed to conflict with typical startup sequences where agents need to be spawned.

**Answer**: **A) Optimistic start approach**
- Instances start in "running" state **immediately** upon creation
- No "starting" state exists
- Agent spawning is NOT part of instance creation
- "Running" means "ready to accept jobs and spawn agents on-demand"

**Implications**:
- Lightweight instance creation (no heavy agent spawning)
- Instances are immediately available for work
- Agent spawning happens asynchronously when jobs arrive

---

### Q2: Work Assignment During Startup

**Category**: Architecture & Design  
**Status**: ✅ Answered

**Question**: If instances start "running" immediately (before agents are spawned), how do we prevent work from being assigned to instances that don't have agents ready yet?

**Context**: With optimistic start, there's a window where instance state = "running" but no agents exist yet.

**Answer**: **Lazy agent initialization pattern**
- Agents are **NOT pre-spawned** during instance creation
- Agents are defined in the tag's database configuration
- Agents are **spawned on-demand** when jobs arrive
- "Running" state means "ready to accept jobs and spawn agents as needed"

**Implications**:
- No gap/race condition - agents spawn when needed
- First job on an instance will have higher latency (agent cold start)
- Health calculation can't rely on "agents currently running" count
- Instance deletion is soft (agents clean up when jobs complete)
- No explicit "ready" flag needed - "running" = ready

---

### Q3: Multi-Instance Resource Isolation

**Category**: Business Logic  
**Status**: ✅ Answered

**Question**: When multiple instances of the same tag are running simultaneously, how are agent resources isolated between instances?

**Context**: If Instance-A and Instance-B both run from tag "release-v1.0.0", and both receive jobs triggering the same agent role (e.g., "NW-MON"):
- Do they share a single agent pool?
- Do they get separate agent instances?
- How do we prevent instance A's agents from processing instance B's jobs?

**Answer**: **Separate agent instances per deployment instance**
- Each instance spawns its own agents independently
- Complete isolation between instances
- No shared agent pools

**Implications**:
- True multi-tenancy at instance level
- Resource consumption scales linearly with instance count
- Instance failures don't affect other instances
- Each instance can have different agent configurations from different tags

---

### Q4: Agent Identification and Tracking

**Category**: Data Models  
**Status**: ✅ Answered

**Question**: How are separate agent instances identified and tracked in the database to prevent collisions?

**Context**: Multiple instances spawning agents with the same role codes need unique database identifiers.

**Answer**: **Instance ID field-based filtering**
- Agent documents have an `instance_id` field
- Agents stored in flat collection with filtering: `FILTER a.instance_id == @instanceId`
- Clean query patterns without complex namespacing

**Implications**:
- Simple ArangoDB queries for instance-scoped agent lists
- Agents table/collection remains unified (not split per instance)
- Index on `instance_id` field recommended for performance
- Agent IDs themselves can be UUIDs (globally unique)

**Noted Gap**: Need to clarify if agent IDs are UUIDs or role-based with collision prevention

---

### Q5: Job Routing to Instances

**Category**: Business Logic  
**Status**: 📝 Noted for Later

**Question**: When a job arrives for a specific instance, how does the system route that job to the correct instance's agents?

**Context**: With agents filtered by `instance_id`, jobs must carry instance context to ensure:
- Jobs for Instance-A only trigger Instance-A's agents
- Instance-B's agents don't pick up Instance-A's jobs
- Work orchestrator targets correct instance

**What We're Looking For**:
- Job instance_id field for filtering
- Instance-specific job submission endpoints
- Instance-scoped job queues
- Workflow execution context binding
- Combination of approaches

**Importance**: Critical for complete instance isolation story

**Note**: This is part of the broader job/workflow system architecture. Will document expected approach based on instance isolation pattern established (instance_id field filtering).

---

## Key Architectural Insights Discovered

### 1. Lazy Agent Initialization
**Discovery**: Agents are not pre-spawned; they're instantiated on-demand when jobs arrive.

**Impact**:
- Instance creation is fast and lightweight
- Resource consumption scales with actual workload, not instance count
- Cold start latency on first job per agent role
- Health checks must use tag configuration, not live agent count

### 2. Optimistic State Model
**Discovery**: Instances start directly in "running" state (no "starting" phase).

**Impact**:
- Simplified state machine (fewer transitions)
- Instances immediately available for work routing
- No need for readiness probes before work assignment
- State transitions: draft → running → stopped → deleted

### 3. Database-Driven Agent Configuration
**Discovery**: Agents are defined in tag database configuration, not in runtime memory.

**Impact**:
- Tag immutability ensures consistent agent definitions
- Multiple instances can reference same tag config
- Agent behavior determined at spawn time from tag data
- Configuration changes require new tag (proper versioning)

### 4. Instances Dashboard UI Structure
**Discovery**: Centralized dashboard at `/instances` route for monitoring all instances.

**Components**:
- **Tags List Page**: Shows all tags with running instances nested underneath
- **Start Instance Dialog**: Modal for creating instances with name, description, tags fields
- **Instance Dashboard**: Per-instance monitoring with 5 panels:
  - Overview Card (state, health, uptime, deployment info)
  - Agent References Panel (role codes from tag)
  - Workflow Execution Panel (active workflows with progress)
  - Recent Activity Feed (timeline of events)
  - Control Actions (stop/restart/delete buttons)

**Impact**:
- Single location to view all instance activity across tags
- Real-time uptime calculation (current_time - started_at)
- Graceful shutdown controls with 30s timeout
- Instance creation pre-fills name with `{tag_name} - {current_date}`

### 5. Instance-Scoped Agent Filtering
**Discovery**: Agents use `instance_id` field for database-level instance isolation.

**Implementation**:
- Agent documents include `instance_id` field
- Flat agents collection (not per-instance collections)
- Query pattern: `FOR a IN agents FILTER a.instance_id == @instanceId`
- Agent IDs are globally unique (likely UUIDs)

**Impact**:
- Simple query patterns for fetching instance agents
- Unified agent storage (easier to manage, backup, query across instances)
- Performance depends on indexing `instance_id` field
- Cross-instance agent queries possible for admin/monitoring

---

## Remaining Topics to Explore

### High Priority
- [ ] Multi-instance resource isolation mechanism (Q3 in progress)
- [ ] Instance health calculation implementation
- [ ] Graceful shutdown sequence details
- [ ] Job routing to specific instances

### Medium Priority
- [ ] Instance naming constraints and validation
- [ ] Soft delete behavior and cleanup
- [ ] Uptime calculation edge cases
- [ ] Error handling for agent spawn failures

### Low Priority
- [ ] UI state synchronization approach
- [ ] Dashboard refresh intervals
- [ ] Tag compatibility checks before instance creation
- [ ] Migration path from single-instance to multi-instance

---

## Next Steps

1. Continue Q3: Answer multi-instance isolation question
2. Deep dive into health calculation logic
3. Explore graceful shutdown mechanics
4. Document job routing and instance selection

---

## Notes

- Session follows one-question-at-a-time research format
- Each answer may trigger follow-up questions (DEEPER path)
- Gaps noted for later exploration (NOTE path)
- Will review and summarize every 5-7 questions
