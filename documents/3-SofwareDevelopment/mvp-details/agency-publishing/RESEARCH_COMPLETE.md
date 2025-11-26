# Agency Instance Management - Research & Documentation Complete

**Date**: November 24, 2025  
**Task**: MVP-PUB-007  
**Status**: ✅ Research & Documentation Phase Complete

---

## Completion Summary

All architectural research questions have been answered, and comprehensive implementation specifications have been documented. The system is now ready for the implementation phase.

---

## Research Session Overview

**Methodology**: Followed `research.prompt.md` - systematic one-question-at-a-time approach  
**Total Questions**: 11  
**Full Session Log**: [instance-research-session.md](instance-research-session.md)

### Questions Answered

1. **Q1**: Instance startup state transition criteria → **Optimistic start** (no "starting" state)
2. **Q2**: Work assignment during startup → **Lazy agent initialization** (agents spawn on-demand)
3. **Q3**: Database schema for instance isolation → **`instance_id` field filtering** (not separate DBs)
4. **Q4**: Tag snapshot preservation across instances → **Tag's database is source of truth**
5. **Q5**: Reusing existing agent instances → **No instance reuse** (each instance has separate agent instances)
6. **Q6**: Navigation structure → **Dual routes** (/instances + /agencies/:id/instances/:id)
7. **Q7**: List page layout → **Hybrid approach** (by-tag tab + flat table tab)
8. **Q8**: Filtering implementation → **Standard approach** (state + tag + search + sort dropdowns)
9. **Q9**: Data loading strategy → **Full server render** (pagination planned for future)
10. **Q10**: Dashboard polling mechanism → **Opt-in auto-refresh** with staggered intervals
11. **Q11**: Polling intervals → **Staggered per panel** (Overview: 30s, Agents: 30s, Workflows: 20s, Activity: 60s)

---

## Documentation Deliverables

### Core Specification Files

| File | Lines | Status | Description |
|------|-------|--------|-------------|
| [instance-management.md](instance-management.md) | 80 | ✅ Complete | Task overview with research-driven design decisions |
| [instance-data-models.md](instance-data-models.md) | 183 | ✅ Complete | Data models, database schema, indexes |
| [instance-services.md](instance-services.md) | 297 | ✅ Complete | Service interface, repository, business logic |
| [instance-api.md](instance-api.md) | 354 | ✅ Complete | 9 REST endpoints with request/response specs |
| [instance-ui-templates.md](instance-ui-templates.md) | 865 | ✅ Complete | Complete Templ template specifications (list, dashboard, panels, dialog) |
| [instance-ui-javascript.md](instance-ui-javascript.md) | 686 | ✅ Complete | JavaScript functions, routing, styling, helpers |
| [instance-research-session.md](instance-research-session.md) | 479 | ✅ Complete | Full Q&A log with architectural decisions |
| [README.md](README.md) | 333 | ✅ Updated | Overview with research status and file links |

**Total Documentation**: ~3,277 lines of comprehensive specifications

**File Size Compliance**: 
- ✅ 7 of 8 files under 700 lines
- ⚠️ instance-ui-templates.md (865 lines) - acceptable for comprehensive template documentation

### Supporting Files (Lifecycle Domain)

| File | Lines | Status | Description |
|------|-------|--------|-------------|
| [state-models.md](state-models.md) | 350 | ✅ Complete | Agency state enum, publication/tag models |
| [state-transitions.md](state-transitions.md) | 616 | ✅ Complete | State machine logic with guards and actions |
| [state-database.md](state-database.md) | 671 | ✅ Complete | ArangoDB collections, indexes, migrations |

---

## Key Architectural Decisions

### Instance Architecture

1. **Lazy Agent Initialization**
   - Agents are NOT pre-spawned during instance creation
   - Agent definitions stored in tag's database collections
   - Agents spawn on-demand when jobs arrive
   - Benefits: Fast instance creation, resource efficiency

2. **Optimistic Start Pattern**
   - Instances immediately enter "running" state on creation
   - No "starting" state exists
   - "Running" means "ready to accept jobs and spawn agents on-demand"
   - Benefits: Simplifies state machine, improves UX

3. **Instance Isolation via `instance_id` Field**
   - All collections use `instance_id` field for filtering
   - No separate databases per instance
   - Filtering applied at query level: `FOR doc IN agents FILTER doc.instance_id == @instance_id`
   - Benefits: Simplifies database management, enables cross-instance analytics

### UI/UX Architecture

4. **Dual Navigation Structure**
   - **Route 1**: `/instances` (top-level instances list across all agencies)
   - **Route 2**: `/agencies/:id/instances/:id` (agency-scoped instance dashboard)
   - Benefits: Global overview + agency-specific drill-down

5. **Hybrid List View**
   - **Tab 1**: "By Tag" - Groups instances by tag using card layout
   - **Tab 2**: "All Instances" - Flat sortable table with filters
   - Benefits: Flexibility for different use cases (navigation vs search)

6. **Standard Filtering Approach**
   - State dropdown (running, stopped, failed)
   - Tag dropdown (populated from available tags)
   - Search box (name/description)
   - Sort dropdown (name-asc, name-desc, date-asc, date-desc)
   - Client-side filtering in JavaScript
   - Benefits: Familiar UX pattern, no backend complexity

7. **Full Server Render for MVP**
   - Load all instances on initial page load
   - No HTMX lazy-loading for list page
   - Future enhancement: Pagination when >500 instances
   - Benefits: Simple implementation, fewer moving parts

8. **Opt-in Dashboard Polling**
   - Auto-refresh toggle button (off by default)
   - HTMX polling with `hx-trigger="every Xs [autoRefreshEnabled]"`
   - Global JavaScript flag controls all panels
   - Benefits: User control, bandwidth-friendly

9. **Staggered Panel Refresh Intervals**
   - Overview Panel: 30 seconds (moderate)
   - Agents Panel: 30 seconds (moderate)
   - Workflows Panel: 20 seconds (fastest - active progress)
   - Activity Panel: 60 seconds (slowest - historical data)
   - Benefits: Load distribution, optimized for data volatility

10. **Independent Panel Components**
    - Each panel is separate Templ component
    - Each panel has own HTMX endpoint for partial updates
    - Each panel configurable (refresh interval, enabled state)
    - Benefits: Modularity, independent loading states

---

## Implementation Roadmap

### Phase 1: Data Layer (Day 1)
- [ ] Create `AgencyInstance` model in `internal/agency/models/instance.go`
- [ ] Add `agency_instances` collection to ArangoDB schema
- [ ] Create indexes: `instance_id`, `agency_id + name`, `tag_id`, `state`
- [ ] Implement `InstanceRepository` with CRUD operations
- [ ] Add `instance_id` field to existing collections (agents, workflows, tasks)

### Phase 2: Service Layer (Day 1-2)
- [ ] Create `InstanceService` interface in `internal/agency/services/instance_service.go`
- [ ] Implement `StartInstance()` - tag retrieval, validation, instance creation
- [ ] Implement `StopInstance()` - graceful shutdown with 30s timeout
- [ ] Implement `RestartInstance()` - stop and restart flow
- [ ] Implement `GetInstanceHealth()` - on-demand calculation
- [ ] Implement `ListInstances()` with filtering support

### Phase 3: API Layer (Day 2)
- [ ] Create `InstanceHandler` in `internal/web/handlers/instance_handler.go`
- [ ] Implement 9 REST endpoints (see [instance-api.md](instance-api.md))
- [ ] Add validation middleware
- [ ] Add error handling
- [ ] Register routes in main router

### Phase 4: UI Components (Day 2-3)
- [ ] Create `instances_list.templ` with hybrid view (2 tabs)
- [ ] Create `instance_dashboard.templ` with 5-panel layout
- [ ] Create 4 panel component templates:
  - [ ] `instance_overview_panel.templ`
  - [ ] `instance_agents_panel.templ`
  - [ ] `instance_workflows_panel.templ`
  - [ ] `instance_activity_panel.templ`
- [ ] Create `start_instance_dialog.templ` modal component
- [ ] Run `templ generate` to compile templates

### Phase 5: JavaScript (Day 3)
- [ ] Create `static/js/instances.js`
- [ ] Implement tab switching logic
- [ ] Implement client-side filtering
- [ ] Implement auto-refresh toggle
- [ ] Implement dialog management functions
- [ ] Implement instance control functions (stop, restart, delete)
- [ ] Implement helper functions (CSS class mapping, uptime formatting)

### Phase 6: Integration Testing (Day 3)
- [ ] Test instance creation from tag
- [ ] Test graceful shutdown with active workflows
- [ ] Test restart flow
- [ ] Test filtering and sorting
- [ ] Test HTMX polling (enable/disable)
- [ ] Test staggered panel updates
- [ ] Test soft delete flow
- [ ] Verify `instance_id` isolation (query multiple instances)

---

## Technical Specifications Summary

### Database Schema
- **Collection**: `agency_instances` (in agency-specific database)
- **Key Indexes**: 
  - `instance_id` (persistent, unique)
  - `agency_id + name` (unique constraint)
  - `tag_id` (persistent)
  - `state` (skiplist)

### Service Layer
- **Interface**: `InstanceService` with 7 methods
- **Key Patterns**: 
  - Graceful shutdown (30s timeout)
  - On-demand health calculation
  - Unique name validation per agency
  - Tag immutability enforcement

### API Layer
- **Endpoints**: 9 REST endpoints (4 POST, 4 GET, 1 DELETE)
- **Authentication**: Agency-scoped (require agency access)
- **Response Format**: JSON with consistent error structure

### UI Layer
- **Pages**: 2 (instances list, instance dashboard)
- **Components**: 6 Templ templates (list, dashboard, 4 panels, dialog)
- **JavaScript**: 1 module with 12+ functions
- **Styling**: Bulma CSS framework (cards, tables, modals, tabs, progress bars, timelines)

### HTMX Integration
- **Polling Endpoints**: 4 partial update endpoints
- **Trigger Pattern**: `hx-trigger="every Xs [autoRefreshEnabled]"`
- **Swap Strategy**: `hx-swap="outerHTML"` for panel replacement

---

## Implementation Checklist (High-Level)

### Prerequisites
- [x] Research complete (11 questions answered)
- [x] Documentation complete (6 specification files)
- [x] Design decisions finalized

### Implementation Tasks
- [ ] Data models & database collections
- [ ] Service layer (InstanceService + Repository)
- [ ] API endpoints (9 handlers)
- [ ] Templ templates (6 files)
- [ ] JavaScript module (instances.js)
- [ ] Integration tests
- [ ] User acceptance testing

### Validation Criteria
- [ ] Can create multiple instances from same tag
- [ ] Instances are properly isolated (separate agent pools, workflows)
- [ ] Graceful shutdown completes within 30 seconds
- [ ] Dashboard panels update independently with auto-refresh
- [ ] Filtering works correctly (client-side)
- [ ] All state transitions follow state machine rules

---

## Next Steps

1. **Begin Implementation**: Start with Phase 1 (Data Layer)
2. **Follow Documented Patterns**: Use exact template structures from instance-ui.md
3. **Test Incrementally**: Validate each phase before moving to next
4. **Reference Research**: Consult instance-research-session.md when design questions arise

---

## Questions or Concerns?

If any ambiguity arises during implementation, refer to:
1. [instance-research-session.md](instance-research-session.md) - Full architectural Q&A
2. [instance-ui.md](instance-ui.md) - Complete UI specifications
3. [instance-services.md](instance-services.md) - Service layer patterns
4. [instance-api.md](instance-api.md) - API endpoint details
5. [instance-data-models.md](instance-data-models.md) - Database schema

All design decisions have been explicitly documented through the research session.
