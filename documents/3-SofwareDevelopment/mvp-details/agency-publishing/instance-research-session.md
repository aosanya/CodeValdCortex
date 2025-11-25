# Research Session: Agency Instance Management (MVP-PUB-007)

**Date**: November 24, 2025  
**Status**: ✅ Complete  
**Focus**: Architecture clarification and design decisions

---

## Session Summary

**Total Questions Answered**: 11  
**Duration**: Full research cycle completed  
**Outcome**: Complete architectural specification ready for implementation

**Key Outcomes**:
1. ✅ **Instance Architecture**: Lazy agent initialization with optimistic start confirmed
2. ✅ **Database Strategy**: `instance_id` field filtering across all collections (no separate DBs)
3. ✅ **Navigation Structure**: Dual routes with hybrid list view (by-tag + flat table)
4. ✅ **Filtering Strategy**: Standard dropdowns + search (state, tag, search, sort)
5. ✅ **Data Loading**: Full server render for MVP (pagination planned for future)
6. ✅ **Polling Mechanism**: Opt-in auto-refresh with staggered intervals per panel
7. ✅ **UI Components**: 5 independent dashboard panels with configurable refresh

**Documentation Updated**:
- [instance-ui.md](instance-ui.md) - Complete Templ templates and JavaScript specifications (1,057 lines)
- [instance-management.md](instance-management.md) - Enhanced with research-driven design decisions
- [mvp.md](../../mvp.md) - Enhanced MVP-PUB-007 description with architectural insights

**Next Steps**: Implementation phase - create actual `.templ` files, Go handlers, JavaScript modules based on documented specifications.

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
**Status**: ✅ Answered with Critical Clarification

**Question**: When a job arrives for a specific instance, how does the system route that job to the correct instance's agents?

**Context**: With agents filtered by `instance_id`, jobs must carry instance context to ensure:
- Jobs for Instance-A only trigger Instance-A's agents
- Instance-B's agents don't pick up Instance-A's jobs
- Work orchestrator targets correct instance

**Answer**: **Critical architectural separation: Definitions vs Executions**

**The Key Distinction**:

1. **Workflow Definitions** (`internal/agency/models/workflow.go`):
   - Blueprints/templates stored in agency database
   - Agency-wide, NOT instance-specific
   - Define workflow structure, steps, dependencies
   - **DO NOT have `instance_id` field**
   - Shared across all instances as reusable templates

2. **Workflow Executions** (`internal/orchestration/types.go`):
   - Runtime instances of workflow definitions
   - **HAVE `instance_id` field** (optional)
   - Track execution state, current step, agent assignments
   - Link back to definition via `workflow_id`
   - Instance-scoped OR agency-wide

3. **Task Executions** (`internal/orchestration/types.go`):
   - Runtime task execution within workflow
   - Has `agent_id` field linking to instance's agents
   - Part of WorkflowExecution context

**Routing Mechanism**:
```go
// API accepts instance_id parameter
POST /api/v1/agencies/{agency_id}/workflows/{workflow_id}/execute
{
  "instance_id": "inst-uuid-123",  // Optional - if provided, scopes to instance
  "parameters": {...}
}

// Creates WorkflowExecution with instance_id
execution := &WorkflowExecution{
    WorkflowID: workflow_id,
    InstanceID: instance_id,  // This ties execution to instance
    ...
}

// Engine selects agents filtered by instance_id
agents := GetAgents(ctx, AgentFilters{InstanceID: execution.InstanceID})
```

**Behavior**:
- If `instance_id` provided: Execution uses only that instance's agents
- If `instance_id` null/empty: Execution uses agency-wide agent pool (no instance)
- Same workflow definition can run on multiple instances simultaneously
- Each execution is isolated by its instance_id

**Implications**:
- Workflow definitions remain pure templates (no runtime state)
- WorkflowExecution carries instance context
- Agent assignment respects instance isolation
- Job routing is execution-level, not definition-level

**References**:
- See `internal/orchestration/types.go` lines 256-300 for `WorkflowExecution` model
- See `internal/orchestration/engine.go` for execution engine implementation

---
### Q6: Navigation Structure

**Category**: Architecture & Design  
**Status**: ✅ Answered

**Question**: Where should the instances list page be located in the navigation - as a new top-level route `/instances` or nested under tags?

**Answer**: **Dual navigation approach**
- `/instances` - Top-level route showing ALL instances across all tags
- `/agencies/:id/instances/:instance_id` - Per-instance dashboard

**Implications**:
- Two entry points: global instance view + tag-scoped view
- Need handlers for both routes
- Navigation breadcrumbs should reflect entry point

---

### Q7: Instances List Page Layout

**Category**: User Experience  
**Status**: ✅ Answered

**Question**: On the `/instances` page, how should instances be organized and displayed?

**Answer**: **C) Hybrid view with tabs**
- Tab 1: "By Tag" - grouped view (tag cards with instances nested underneath)
- Tab 2: "All Instances" - flat table view (sortable/filterable)
- User can switch between views based on task

**Implications**:
- Need tab switching UI (Bulma tabs component)
- Two data presentation modes in same template
- Client-side tab switching (data pre-loaded)
- More complex template but better UX flexibility

---

### Q8: Filtering Capabilities

**Category**: User Experience  
**Status**: ✅ Answered

**Question**: For the "all instances" flat view, what filtering and sorting capabilities should be available?

**Answer**: **B) Standard filtering**
- State filter dropdown (running/stopping/stopped/failed)
- Tag filter dropdown (from available tags)
- Search box (by instance name)
- Sort by name/date (ascending/descending)

**Implications**:
- Filter controls in page header
- Client-side filtering (data already loaded)
- Need to populate tag dropdown from all available tags
- Standard Bulma select/input components

---

### Q9: Data Loading Strategy

**Category**: API & Interfaces  
**Status**: ✅ Answered

**Question**: Should the `/instances` page load all instances data on initial page render, or use HTMX to lazy-load?

**Answer**: **A) Full server render**
- Handler fetches all instances + tags
- Complete page rendered server-side
- JavaScript handles tab switching client-side (all data in DOM)
- **Future**: Pagination when instance counts grow

**Implications**:
- Single handler endpoint for instances list page
- No separate HTMX endpoints for tabs (initially)
- Simpler implementation for MVP
- May need refactor to pagination when scaling

---

### Q10: Dashboard Auto-Refresh

**Category**: User Experience  
**Status**: ✅ Answered

**Question**: For the instance dashboard, should the 5 panels auto-refresh, or manual refresh?

**Answer**: **B) Polling with user toggle**
- Default: Manual refresh (static page)
- Toggle button: "Enable Auto-Refresh" (user opt-in)
- When enabled: JavaScript polls and updates via HTMX

**Implications**:
- Toggle control in dashboard header
- Polling state stored in JavaScript (not persisted)
- Reduces default server load
- Need clear UI feedback when polling active

---

### Q11: Panel Refresh Intervals

**Category**: User Experience  
**Status**: ✅ Answered

**Question**: When polling is enabled, what should the refresh intervals be for each dashboard panel?

**Answer**: **B) Staggered intervals with configurable components**
- Each panel is separate component with own refresh interval
- Panels can be configured independently
- Different update frequencies based on data volatility

**Implications**:
- Need panel component abstraction (templ components)
- Each panel has HTMX endpoint for partial updates
- Panel configuration structure (interval, endpoint, enabled)
- JavaScript manages multiple polling timers

---

## Remaining Topics to Explore

**Answer**: **C) Hybrid view with tabs**
- Tab 1: "By Tag" - grouped view (tag cards with instances nested)
- Tab 2: "All Instances" - flat table view (sortable/filterable)
- User can switch between views based on task

**Implications**:
- Need tab switching UI (Bulma tabs component)
- Two data presentation modes in same template
- Client-side tab switching (data pre-loaded)
- More complex template but better UX flexibility

---

### Q8: Filtering Capabilities

**Category**: User Experience  
**Status**: ✅ Answered

**Question**: For the "all instances" flat view, what filtering and sorting capabilities should be available?

**Answer**: **B) Standard filtering**
- State filter dropdown (running/stopping/stopped/failed)
- Tag filter dropdown (from available tags)
- Search box (by instance name)
- Sort by name/date (ascending/descending)

**Implications**:
- Filter controls in page header
- Client-side filtering (data already loaded) OR server-side (API params)
- Need to populate tag dropdown from all available tags
- Standard Bulma select/input components

---

### Q9: Data Loading Strategy

**Category**: API & Interfaces  
**Status**: ✅ Answered

**Question**: Should the `/instances` page load all instances data on initial page render, or use HTMX to lazy-load instance data?

**Answer**: **A) Full server render**
- Handler fetches all instances + tags
- Complete page rendered server-side
- JavaScript handles tab switching client-side (all data in DOM)
- **Future**: Pagination when instance counts grow

**Implications**:
- Single handler endpoint for instances list page
- No separate HTMX endpoints for tabs (initially)
- Simpler implementation for MVP
- May need refactor to pagination when scaling

---

### Q10: Dashboard Auto-Refresh

**Category**: User Experience  
**Status**: ✅ Answered

**Question**: For the instance dashboard, should the 5 panels auto-refresh, or does the user manually refresh?

**Answer**: **B) Polling with user toggle**
- Default: Manual refresh (static page)
- Toggle button: "Enable Auto-Refresh" (user opt-in)
- When enabled: JavaScript polls and updates via HTMX

**Implications**:
- Toggle control in dashboard header
- Polling state stored in JavaScript (not persisted)
- Reduces default server load (most users won't enable)
- Need clear UI feedback when polling is active

---

### Q11: Panel Refresh Intervals

**Category**: User Experience  
**Status**: ✅ Answered

**Question**: When polling is enabled, what should the refresh intervals be for each dashboard panel?

**Answer**: **B) Staggered intervals with configurable components**
- Each panel is separate component with own refresh interval
- Panels can be configured independently
- Different update frequencies based on data volatility

**Implications**:
- Need panel component abstraction (templ components)
- Each panel has HTMX endpoint for partial updates
- Panel configuration structure (interval, endpoint, enabled)
- JavaScript manages multiple polling timers

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
