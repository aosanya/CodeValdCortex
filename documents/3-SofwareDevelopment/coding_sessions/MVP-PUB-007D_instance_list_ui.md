# Coding Session: MVP-PUB-007D - Instance List UI (Hybrid View)

**Task ID**: MVP-PUB-007D  
**Date**: 2025-11-25  
**Status**: ✅ Complete  
**Branch**: feature/MVP-PUB-007_agency_instance_management

---

## Objective

Implement the instance list UI with hybrid view (by-tag grouped cards + flat table with filters).

## Tasks Completed

### 1. Template File Creation
**File**: `internal/web/pages/instances_list.templ` (423 lines)

Created complete Templ template with:
- Main `InstancesList` component with header and tab structure
- **Tab 1 (By Tag)**: `instancesByTagView` - Tag cards with instance lists
  - `tagCardWithInstances` - Individual tag card component
  - `instanceListItem` - Instance card within tag group
- **Tab 2 (All Instances)**: `instancesTableView` - Flat table with filters
  - Filter controls: state, tag, search, sort
  - `instanceTableRow` - Table row with instance data and actions
- `startInstanceDialog` - Modal dialog for creating new instances
- Helper functions: `getInstancesForTag`, `getStateClass`, `getHealthClass`, `formatUptime`

### 2. JavaScript Implementation
**File**: `static/js/instances.js` (352 lines)

Implemented all client-side functionality:
- **Tab switching**: `switchTab()` - Toggle between by-tag and table views
- **Filtering**: `applyFilters()`, `sortRows()` - Client-side filter and sort logic
- **Dialog management**: `openStartInstanceDialog()`, `closeStartInstanceDialog()`, `startInstanceFromTag()`
- **Instance creation**: `submitStartInstance()` - POST to `/api/v1/agencies/:id/tags/:name/instances`
- **Instance control**: 
  - `viewInstance()` - Navigate to dashboard
  - `stopInstance()` - Graceful shutdown
  - `restartInstance()` - Restart instance
  - `deleteInstance()` - Soft delete
- **Notifications**: `showNotification()` - Toast messages
- **Initialization**: DOMContentLoaded event handler

### 3. Web Handler
**File**: `internal/web/handlers/instance_web_handler.go` (73 lines)

Created new web handler for instance UI:
- `InstanceWebHandler` struct with dependencies (InstanceService, AgencyService, TagService, Logger)
- `NewInstanceWebHandler()` constructor
- `ShowInstancesList()` method:
  - Fetches agency, tags, instances
  - Renders `pages.InstancesList` template
  - Error handling with fallbacks

### 4. Route Registration
**File**: `internal/app/app.go` (modified)

Registered web route:
```go
if a.instanceService != nil && a.tagService != nil {
    instanceWebHandler := webhandlers.NewInstanceWebHandler(a.instanceService, a.agencyService, *a.tagService, a.logger)
    router.GET("/agencies/:id/instances", instanceWebHandler.ShowInstancesList)
    a.logger.Info("Instance management web routes registered")
}
```

---

## Technical Details

### Template Architecture
- **Layout reuse**: Uses `@components.LayoutWithAgency()` - follows DRY principles
- **Hybrid tabs**: Two views of the same data (by-tag vs flat table)
- **Data attributes**: All rows/cards include `data-*` attributes for JavaScript filtering
- **Agency ID injection**: Uses `<meta name="agency-id" content={agency.ID}/>` for JavaScript access

### Data Flow
1. User navigates to `/agencies/:id/instances`
2. `ShowInstancesList` handler:
   - Fetches agency metadata
   - Fetches all tags for agency
   - Fetches all instances for agency
3. Renders `InstancesList` template with data
4. JavaScript initializes on page load:
   - Stores agency ID from meta tag
   - Sets default tab to "by-tag"
   - Attaches event handlers

### Instance Creation Flow
1. User clicks "New Instance" or "Start Instance" on tag card
2. Dialog opens with tag pre-selected (if from card)
3. User fills name and description
4. JavaScript POSTs to `/api/v1/agencies/:id/tags/:name/instances`
5. Backend creates instance and spawns agents (optimistic start)
6. Success notification shown, page reloads after 1.5s

### State Management
- **Client-side filtering**: All filtering/sorting happens in JavaScript (no server round-trips)
- **Tab state**: Stored in DOM classes (`is-active`)
- **Filter values**: Read directly from form inputs

---

## Files Created/Modified

### Created
1. `/workspaces/CodeValdCortex/internal/web/pages/instances_list.templ` (423 lines)
2. `/workspaces/CodeValdCortex/static/js/instances.js` (352 lines)
3. `/workspaces/CodeValdCortex/internal/web/handlers/instance_web_handler.go` (73 lines)

### Modified
1. `/workspaces/CodeValdCortex/internal/app/app.go` - Added web route registration

---

## Build & Test Results

### Build Status
```bash
$ make build
Building codevaldcortex...
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=2fa7800-dirty -X main.buildTime=2025-11-25T13:51:47Z -X main.gitCommit=2fa78004de5eafb0e5f2d757daf595f54c1857d6" -o bin/codevaldcortex ./cmd
```

✅ **Build succeeded** - All components compile correctly.

### Manual Testing
- ❓ **Pending**: Need to run application and manually test UI
- Required tests:
  - Page loads without errors
  - Both tabs render correctly
  - Tab switching works
  - Filters apply correctly
  - Dialog opens/closes
  - Instance creation workflow (requires existing tags)

---

## Dependencies

### Backend Dependencies
- ✅ MVP-PUB-007A: Instance data models and repository
- ✅ MVP-PUB-007B: Instance service layer
- ✅ MVP-PUB-007C: Instance API endpoints
- ✅ Tag service (for fetching tags)
- ✅ Agency service (for fetching agency metadata)

### Frontend Dependencies
- ✅ Bulma CSS (styling)
- ✅ Alpine.js (for interactive components in layout)
- ✅ HTMX (for partial page updates - not heavily used in this view)
- ✅ FontAwesome (icons)

---

## Next Steps

1. **MVP-PUB-007E**: Instance Dashboard (single instance view)
   - Create 5-panel dashboard
   - Implement HTMX polling for real-time updates
   - Add charts for resource metrics
   - Event timeline with color-coded markers

2. **MVP-PUB-007F**: Integration Testing
   - Test full instance lifecycle
   - Verify all UI interactions
   - Test edge cases (no instances, no tags, errors)

---

## Lessons Learned

1. **Type alignment**: Needed to use `agency.Service` instead of `services.AgencyService` - checked app.go field types first
2. **Interface parameters**: TagService.ListTags requires `*TagFilters` parameter (pass nil for all)
3. **Template compilation**: `templ generate` automatically runs in `make build`
4. **DRY compliance**: Reused `@components.LayoutWithAgency()` to avoid duplicating page structure

---

## Code Quality Notes

- ✅ **File sizes**: All files under limits (templ: 423 lines, JS: 352 lines, handler: 73 lines)
- ✅ **No code duplication**: Reused shared layout component
- ✅ **Error handling**: Graceful fallbacks for missing data
- ✅ **Clear separation**: HTML in .templ, logic in .js, data access in handler
- ✅ **Accessibility**: Proper ARIA labels, semantic HTML

---

## Documentation References

- [instance-ui-templates.md](../../mvp-details/agency-publishing/instance-ui-templates.md) - Template specifications
- [instance-ui-javascript.md](../../mvp-details/agency-publishing/instance-ui-javascript.md) - JavaScript specifications
- [instance-api.md](../../mvp-details/agency-publishing/instance-api.md) - API endpoint documentation
- [instance-data-models.md](../../mvp-details/agency-publishing/instance-data-models.md) - Data model definitions

---

**Session Complete**: MVP-PUB-007D ✅
