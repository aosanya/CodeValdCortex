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
**Status**: ⏸️ Paused (waiting for answer)

**Question**: When multiple instances of the same tag are running simultaneously, how are agent resources isolated between instances?

**Context**: If Instance-A and Instance-B both run from tag "release-v1.0.0", and both receive jobs triggering the same agent role (e.g., "NW-MON"):
- Do they share a single agent pool?
- Do they get separate agent instances?
- How do we prevent instance A's agents from processing instance B's jobs?

**What We're Looking For**:
- Agent ID namespacing mechanism
- Instance-scoped agent collections/tables
- Job routing metadata
- Other isolation approaches

**Importance**: Critical for ensuring true instance isolation

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
