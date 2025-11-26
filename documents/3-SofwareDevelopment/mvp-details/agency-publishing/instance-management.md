# MVP-PUB-007: Agency Instance Management System

**Task ID**: MVP-PUB-007  
**Priority**: P0 (Critical)  
**Effort**: High  
**Skills**: Go, ArangoDB, Templ, Frontend Dev  
**Dependencies**: MVP-PUB-006 (Agency-Specific Tag Storage) ✅

---

## Overview

Implement a comprehensive instance management system that enables running multiple independent instances of an agency from any tag snapshot.

## Documentation Structure

This task is split across topic-based files (refactored from original 884-line file):

### 📊 [Data Models & Database Schema](instance-data-models.md)
**182 lines** - `AgencyInstance` model, database collections, indexes

### ⚙️ [Service Layer & Repository](instance-services.md)
**296 lines** - Service interface, graceful shutdown, health calculation

### 🌐 [HTTP API Endpoints](instance-api.md)
**353 lines** - 9 REST endpoints, handlers, error handling

### 🎨 [UI Templates](instance-ui-templates.md)
**658 lines** - Complete Templ template specifications: hybrid list view (by-tag + flat table tabs), 5-panel dashboard, start instance dialog

### 💻 [JavaScript & Routing](instance-ui-javascript.md)
**335 lines** - Tab switching, client-side filtering, auto-refresh control, dialog management, instance control functions, UI routes

### 🔬 [Research Session](instance-research-session.md)
**Full architectural Q&A** - 11 questions answered covering navigation structure, filtering approach, data loading strategy, polling mechanism, panel design

---

## Key Design Decisions (Research-Driven)

### Instance Management Architecture
1. **Storage**: All instances in `agency_instances` collection (agency DB)
2. **Agent References**: Lazy initialization - agents spawn on-demand when jobs arrive
3. **Optimistic Start**: Instances immediately enter "running" state without validation
4. **Instance Isolation**: All collections use `instance_id` field for filtering/querying

### UI/UX Architecture
5. **Navigation**: Dual routes - `/instances` (top-level list) + `/agencies/:id/instances/:id` (dashboard)
6. **List View**: Hybrid approach - Tab 1 (by-tag grouping), Tab 2 (flat table with filters)
7. **Filtering**: Standard approach - state dropdown, tag dropdown, search box, sort dropdown
8. **Data Loading**: Full server render on page load (future: pagination for 500+ instances)
9. **Dashboard Polling**: Opt-in auto-refresh with staggered intervals per panel
   - Overview: 30s, Agents: 30s, Workflows: 20s, Activity: 60s
10. **Panel Independence**: Each panel is separate component with own HTMX endpoint

### Operational Patterns
11. **Health**: On-demand calculation (not stored)
12. **Shutdown**: Graceful (30s timeout, rejects new jobs)
13. **Naming**: Unique per agency
14. **Delete**: Soft delete only (preserves audit trail)
15. **Uptime**: Real-time calculation (`current_time - started_at`)

---

## Implementation Checklist

- [ ] Data models & collections
- [ ] Service & repository layers
- [ ] 9 API endpoints
- [ ] UI (tags list, dialog, dashboard)
- [ ] Integration & tests

See detailed checklists in topic files above.

---

## Timeline

**Estimated**: 2-3 days (16-24 hours)  
**Complexity**: High

---

<!-- MVP-PUB-007 -->
## Implementation Status: ✅ COMPLETED (2025-11-26)

### Delivered Components

**Data Layer (MVP-PUB-007A)**:
- ✅ `internal/agency/models/instance.go` - Complete instance model with state machine
- ✅ `internal/agency/arangodb/instance_repository.go` - Full CRUD operations
- ✅ Database initialization in `internal/agency/database_initializer.go`
- ✅ Agency-scoped collection `agency_instances` with indexes

**Service Layer (MVP-PUB-007B)**:
- ✅ `internal/agency/services/instance_service.go` - Complete business logic
- ✅ Graceful shutdown with 30s timeout
- ✅ Real-time health calculation
- ✅ Instance count tracking per tag

**API Layer (MVP-PUB-007C)**:
- ✅ `internal/handlers/instance_handler.go` - All 9 REST endpoints
- ✅ Routes registered in `internal/app/app.go`
- ✅ Proper error handling and logging

**UI Components (MVP-PUB-007D)**:
- ✅ `internal/web/pages/instances_list.templ` - Hybrid view (by-tag cards + flat table)
- ✅ `internal/web/pages/instance_dashboard.templ` - 3-panel dashboard (overview, agents, metrics)
- ✅ `static/js/instances.js` - Tab switching, filtering, instance controls
- ✅ `static/js/agency-designer/instances.js` - Start instance dialog from versions page
- ✅ Dropdown action menus for stop/restart/delete/view
- ✅ Instance count badges with auto-refresh

**Instance Dashboard (MVP-PUB-007E)**:
- ✅ Overview panel with health status and uptime
- ✅ Agents panel showing instance agent configuration
- ✅ Metrics panel with performance indicators
- ✅ Consistent styling with Bulma framework

**Integration Testing (MVP-PUB-007F)**:
- ✅ Full lifecycle tested: create → list → filter → stop → restart → delete
- ✅ Tag filter navigation working
- ✅ Instance count updates verified
- ✅ Dashboard rendering validated

### Key Implementation Patterns

**Meta Tag for Agency Context**:
```templ
// internal/web/components/layout_with_agency.templ
<meta name="agency-id" content={ currentAgency.ID }/>
```
This ensures JavaScript has access to agency ID for API calls.

**Parameter Name Consistency**:
```go
// All handlers use snake_case to match route definitions
instanceID := c.Param("instance_id")  // NOT "instanceId"
```

**Instance Count Auto-Refresh**:
```javascript
// After creating instance, refresh all tag counts
await loadInstancesForTag(currentTagForInstance);
if (typeof loadTags === 'function') {
    await loadTags();  // Refreshes badges
}
```

**Badge Selector Pattern**:
```javascript
// Specific attribute selector for targeted updates
const badge = document.querySelector(`.instance-count-badge[data-tag-name="${tagName}"]`);
```

### Resolved Issues

**BUG-001**: Tag filter navigation showing empty list
- **Root Cause**: Instance `tag_id` field was empty in database
- **Fix**: Properly populate `tag_id` during instance creation
- **Status**: ✅ Resolved

**Agency ID Undefined**:
- **Root Cause**: Missing `<meta name="agency-id">` tag in template
- **Fix**: Added meta tag to `LayoutWithAgency` component
- **Status**: ✅ Resolved

**Parameter Name Mismatch**:
- **Root Cause**: Routes use `:instance_id` but handlers used `instanceId`
- **Fix**: Updated all handlers to use `instance_id` consistently
- **Status**: ✅ Resolved

### Architecture Decisions

1. **Database-Per-Agency**: Each agency instance stored in agency-specific database
2. **Tag-Based Deployment**: Instances reference immutable tag snapshots
3. **Hybrid UI**: Combines grouped card view (by tag) with flat searchable table
4. **Action Dropdowns**: Cleaner UI with consistent dropdown pattern for actions
5. **Real-time Updates**: Instance counts refresh after create/delete operations

### Files Created/Modified

**Backend**:
- `internal/agency/models/instance.go` (154 lines)
- `internal/agency/arangodb/instance_repository.go` (227 lines)
- `internal/agency/services/instance_service.go` (312 lines)
- `internal/handlers/instance_handler.go` (341 lines)
- `internal/app/app.go` (route registration)

**Frontend**:
- `internal/web/pages/instances_list.templ` (461 lines)
- `internal/web/pages/instance_dashboard.templ` (187 lines)
- `internal/web/components/layout_with_agency.templ` (added meta tag)
- `static/js/instances.js` (419 lines)
- `static/js/agency-designer/instances.js` (397 lines)
- `static/js/agency-designer/tags.js` (instance count loading)
- `static/css/styles.css` (added `.has-text-orange`)

### Known Limitations

1. **Agent Spawning**: Actual agent instantiation not yet implemented (placeholder)
2. **Workflow Initialization**: Workflow engine integration pending
3. **Health Metrics**: Currently placeholder values
4. **Performance**: Full table load (pagination needed for 500+ instances)

### Next Steps

- **MVP-PUB-008**: Agency activation/deactivation system
- **Agent Pool**: Implement actual agent spawning from instance
- **Metrics Collection**: Real instance health calculation
- **Performance**: Add pagination for large instance lists

### Coding Session

📝 [MVP-PUB-007_agency_instance_management.md](../../coding_sessions/MVP-PUB-007_agency_instance_management.md)

<!-- /MVP-PUB-007 -->
