# Coding Session: MVP-PUB-007E - Instance Dashboard (Single Instance View)

**Task ID**: MVP-PUB-007E  
**Date**: 2025-11-25  
**Status**: ✅ Complete  
**Branch**: feature/MVP-PUB-007_agency_instance_management

---

## Objective

Create a detailed dashboard for viewing a single instance with comprehensive monitoring panels.

## Tasks Completed

### 1. Dashboard Template Creation
**File**: `internal/web/pages/instance_dashboard.templ` (234 lines)

Created dashboard page with:
- **Header Section**: Instance name, state/health badges, control buttons (stop/restart/delete), back button
- **Panel 1 (Overview)**: 3-column layout with:
  - State & health badges
  - Tag reference
  - Uptime display
  - Deployed at/by info
  - Job acceptance status
  - Description (if present)
  - Instance tags (if present)
- **Panel 2 (Agents)**: Agent references table with:
  - Role code
  - State
  - Spawned timestamp
  - Empty state message when no agents loaded
- **Panel 3 (Health & Metrics)**: System information with:
  - Instance ID (monospace)
  - Agent count
  - Workflow count
  - Last heartbeat timestamp
  - Instance tags display

**Helper functions**:
- `getStateBadgeClass()` - Returns Bulma class for state (success/warning/light/danger)
- `getHealthBadgeClass()` - Returns Bulma class for health status
- `formatUptimeDisplay()` - Formats seconds to "Xh Ym" or "Ym"

### 2. Web Handler Method
**File**: `internal/web/handlers/instance_web_handler.go` (added method)

Added `ShowInstanceDashboard()` method:
- Extracts `agencyID` and `instance_id` from URL params
- Fetches agency metadata
- Fetches instance by ID
- Fetches agent references for the instance
- Renders `pages.InstanceDashboard` template
- Error handling with 404 for missing instances

### 3. Route Registration
**File**: `internal/app/app.go` (modified)

Registered dashboard route:
```go
router.GET("/agencies/:id/instances/:instance_id", instanceWebHandler.ShowInstanceDashboard)
```

Route placed alongside the instances list route for logical grouping.

---

## Technical Details

### Template Architecture
- **Layout reuse**: Uses `@components.LayoutWithAgency()` - follows DRY principles
- **Meta tags**: Injects agency ID and instance ID for JavaScript access
- **JavaScript reuse**: Uses `/static/js/instances.js` for instance control functions
- **Conditional rendering**: Shows different buttons based on instance state

### Data Flow
1. User navigates to `/agencies/:id/instances/:instance_id`
2. `ShowInstanceDashboard` handler:
   - Fetches agency from master DB
   - Fetches instance from agency-specific DB
   - Fetches agent references for instance
3. Renders dashboard with 3 panels
4. JavaScript provides interactive controls (stop, restart, delete)

### Instance Control Actions
- **Running state**: Shows "Stop" button (graceful shutdown)
- **Stopped state**: Shows "Restart" and "Delete" buttons
- **All states**: Shows "Back to List" link
- **JavaScript functions**: Reused from `instances.js` (stopInstance, restartInstance, deleteInstance, viewInstance)

### Model Field Alignment
**Fixed issues**:
- Changed `InstanceAgent.Status` → `InstanceAgent.State` (correct field)
- Changed `InstanceAgent.AddedAt` → `InstanceAgent.SpawnedAt` (correct field)
- Changed `formatUptimeDisplay(int)` → `formatUptimeDisplay(int64)` (correct type)
- Removed invalid fields: `instance.HealthScore`, `instance.LastHealthCheck`, `instance.UpdatedAt`, `instance.Environment` (map)
- Used available fields: `WorkflowCount`, `LastHeartbeat`, `Tags` (string array)

---

## Files Created/Modified

### Created
1. `/workspaces/CodeValdCortex/internal/web/pages/instance_dashboard.templ` (234 lines)

### Modified
1. `/workspaces/CodeValdCortex/internal/web/handlers/instance_web_handler.go` - Added ShowInstanceDashboard method
2. `/workspaces/CodeValdCortex/internal/app/app.go` - Registered dashboard route

---

## Build & Test Results

### Build Status
```bash
$ templ generate && make build
(✓) Complete [ updates=0 duration=161.901667ms ]
Building codevaldcortex...
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=2fa7800-dirty -X main.buildTime=2025-11-25T13:57:17Z -X main.gitCommit=2fa78004de5eafb0e5f2d757daf595f54c1857d6" -o bin/codevaldcortex ./cmd
```

✅ **Build succeeded** - All components compile correctly.

### Manual Testing
- ❓ **Pending**: Need to run application and manually test dashboard
- Required tests:
  - Navigate from instances list to dashboard
  - Verify all panels display correctly
  - Test instance control buttons
  - Verify state-dependent button display
  - Check responsive layout

---

## Dependencies

### Backend Dependencies
- ✅ MVP-PUB-007A: Instance data models
- ✅ MVP-PUB-007B: Instance service layer (GetInstance, ListInstanceAgents)
- ✅ MVP-PUB-007C: Instance API endpoints
- ✅ MVP-PUB-007D: Instance list UI (navigation from list to dashboard)

### Frontend Dependencies
- ✅ Bulma CSS (card, columns, field, label, tags, buttons)
- ✅ FontAwesome icons
- ✅ instances.js (instance control functions)

---

## Design Decisions

### Simplified Panel Design
**Decision**: Started with 3 core panels instead of 5 specified panels
**Rationale**:
- Workflows panel requires workflow execution data (not yet implemented)
- Activity/Events panel requires event tracking system (not yet implemented)
- Core panels (Overview, Agents, Health) provide essential monitoring
- Additional panels can be added when backend support is ready

### No HTMX Polling (Yet)
**Decision**: Static page without auto-refresh polling
**Rationale**:
- Simplified initial implementation
- Manual refresh works for MVP
- HTMX polling can be added later with:
  - `hx-get` endpoint per panel
  - `hx-trigger="every Xs"` for periodic updates
  - Auto-refresh toggle button

### Field Availability
**Decision**: Only display fields that exist on models
**Rationale**:
- Avoided displaying non-existent fields
- Used available fields: WorkflowCount, LastHeartbeat, Tags
- Removed planned features that need additional backend work:
  - HealthScore (needs calculation)
  - Environment variables (field is string, not map)
  - Resource usage metrics (needs tracking system)

---

## Next Steps

1. **MVP-PUB-007F**: Integration Testing
   - Test full instance lifecycle (create, view, stop, restart, delete)
   - Verify navigation flows
   - Test all UI controls
   - Edge case testing (no agents, stopped instance, etc.)

2. **Future Enhancements** (post-MVP):
   - Add Workflows panel when workflow execution is implemented
   - Add Activity/Events panel when event tracking is implemented
   - Implement HTMX polling for real-time updates
   - Add resource usage charts (CPU, memory)
   - Add health score calculation and display
   - Environment variable display (parse string or change field type)

---

## Code Quality Notes

- ✅ **File size**: Template 234 lines (well under limit)
- ✅ **No code duplication**: Reused LayoutWithAgency, instances.js functions
- ✅ **Error handling**: 404 for missing instances, graceful fallbacks
- ✅ **Clear separation**: HTML in .templ, logic in handler, data access in service
- ✅ **Type safety**: All model fields validated against actual struct definitions

---

## Documentation References

- [instance-ui-templates.md](../../mvp-details/agency-publishing/instance-ui-templates.md) - Dashboard specifications
- [instance-ui-javascript.md](../../mvp-details/agency-publishing/instance-ui-javascript.md) - JavaScript functions
- [instance-data-models.md](../../mvp-details/agency-publishing/instance-data-models.md) - Data model reference
- [instance-api.md](../../mvp-details/agency-publishing/instance-api.md) - API endpoints

---

**Session Complete**: MVP-PUB-007E ✅

Dashboard provides comprehensive single-instance view with essential monitoring panels. Ready for integration testing.
