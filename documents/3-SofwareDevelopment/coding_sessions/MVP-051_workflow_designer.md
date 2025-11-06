# MVP-051: Workflow Manager - List & CRUD

**Task ID**: MVP-051  
**Status**: Complete  
**Completion Date**: November 6, 2025  
**Branch**: `feature/MVP-051_work_item_workflow_designer`  
**Commits**: 8 commits (c8aed33, 81cc56b, 08c8fd7, 6991c02, 1e7cc8a, ad6970a, c6dcad9, d7645d3)

## Overview

Implemented a complete workflow management system for the Agency Designer, enabling users to list, create, edit, delete, and AI-generate workflows that orchestrate work items. This feature provides the foundation for the visual workflow designer (MVP-052).

## Implementation Summary

### Phase 1: Backend Foundation (Commits: c8aed33, 81cc56b)

**AI Workflows Builder** (`internal/builder/ai/workflows_builder.go`)
- Created `WorkflowsBuilder` for AI-powered workflow generation
- Implemented three generation modes:
  - `GenerateWorkflowsFromContext()`: Generate workflows from agency context and work items
  - `GenerateWorkflowWithPrompt()`: Generate single workflow from user's natural language prompt
  - `RefineWorkflow()`: Refine existing workflow based on user feedback
  - `SuggestWorkflowImprovements()`: Analyze and suggest improvements
- Configured LLM with strict constraints:
  - Temperature: 0.5 (focused responses)
  - MaxTokens: 2500 (prevent truncation)
  - Prompt: Generate ONLY 1 simple workflow with 5-7 nodes max
- Added JSON response cleaning and validation

**AI Refine Handler Integration** (`internal/web/handlers/ai_refine/workflow_refine_dynamic.go`)
- Created `RefineWorkflows()` handler for dynamic AI workflow operations
- Implemented operation type detection:
  - "create": Generate workflows from work items
  - "refine": Improve existing workflow structure
  - "suggest": Analyze and suggest improvements
- Integrated with `WorkflowsBuilder` for AI generation
- Added error handling and HTML response rendering

### Phase 2: Frontend Templates & JavaScript (Commit: 08c8fd7)

**Templates Created:**
1. `workflows_list_card.templ` - List view with search, filters, batch operations
2. `workflow_editor_card.templ` - Create/edit form with AI buttons
3. `agency_designer_workflows.templ` - Reusable row components (WorkflowsList, WorkflowItem, WorkflowStatusBadge)

**JavaScript Implementation** (`static/js/agency-designer/workflows.js` - 450+ lines)
- CRUD Operations:
  - `loadWorkflows()`: Fetch and display workflows
  - `showWorkflowEditor()`: Show add/edit form
  - `saveWorkflowFromEditor()`: Create/update workflows
  - `deleteWorkflow()`: Delete workflow
  - `duplicateWorkflow()`: Duplicate existing workflow
- AI Operations:
  - `processAIWorkflowOperation()`: Handle Create/Refine/Suggest operations
  - `generateWorkflowWithAI()`: Generate from description in editor
  - `refineWorkflowWithAI()`: Refine current workflow in editor
- UI Management:
  - Search/filter functionality
  - Checkbox selection tracking
  - Button enable/disable based on selection
  - Status badge color coding

**Navigation Integration** (`agency_designer_overview.templ`, `overview.js`)
- Added "Workflows" section to navigation context
- Positioned under RACI Matrix section
- Integrated loadWorkflows() into section switching logic

### Phase 3: Route Configuration & Bug Fixes (Commits: 6991c02, 1e7cc8a, ad6970a, c6dcad9, d7645d3)

**Route Parameter Fix** (6991c02)
- **Issue**: Gin router panic - `:agencyId` conflicts with `:id` in same route group
- **Solution**: Standardized all agency routes to use `:id` parameter
- Changed workflow routes from `/agencies/:agencyId/workflows` to `/agencies/:id/workflows`
- Updated handlers to use `c.Param("id")` instead of `c.Param("agencyId")`

**AI Error Handling Improvements** (1e7cc8a)
- **Issue**: LLM generating 3 large workflows causing JSON truncation and stuck spinner
- **Solutions**:
  - Frontend: Added `response.ok` check to detect 500 errors and hide spinner
  - Backend: Added truncation detection in `parseWorkflowsResponse()`
  - Reduced workflow count from 2-3 to 1
  - Lowered MaxTokens from 4000 to 2500
  - Reduced Temperature from 0.7 to 0.5
  - Added stricter prompt constraints (max 5-7 nodes, avoid parallel nodes)

**Delete Workflow URL Fix** (ad6970a)
- **Issue**: Using `deleteEntity()` helper created malformed URL: `/api/v1/agencies/:id//api/v1/workflows/:id/workflow`
- **Solution**: Replaced with direct `fetch()` DELETE to `/api/v1/workflows/:id`

**Save/Update Workflow URL Fix** (c6dcad9)
- **Issue**: Using `saveEntity()` helper created malformed URL with `[object Object]`
- **Solution**: 
  - Replaced with direct `fetch()` implementation
  - Correctly handles POST `/api/v1/agencies/:id/workflows` for create
  - Correctly handles PUT `/api/v1/workflows/:id` for update
  - Removed unused `saveEntity` and `deleteEntity` imports

**Agency ID Validation Fix** (d7645d3)
- **Issue**: Validation error "Agency ID is required" when updating workflows
- **Solution**: Always include `agency_id` field in workflow object for both create and update operations

## Technical Details

### API Endpoints Registered
```
POST   /api/v1/agencies/:id/workflows                  # Create workflow
GET    /api/v1/agencies/:id/workflows                  # List agency workflows
GET    /api/v1/agencies/:id/workflows/html             # Get workflows HTML fragment
GET    /api/v1/workflows/:id                           # Get workflow by ID
PUT    /api/v1/workflows/:id                           # Update workflow
DELETE /api/v1/workflows/:id                           # Delete workflow
POST   /api/v1/workflows/:id/duplicate                 # Duplicate workflow
POST   /api/v1/workflows/validate                      # Validate workflow structure
POST   /api/v1/workflows/:id/execute                   # Start workflow execution
POST   /api/v1/agencies/:id/workflows/refine-dynamic   # AI workflow operations
```

### Workflow Model Structure
```go
type Workflow struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Version     string                 `json:"version"`
    Description string                 `json:"description"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
    Status      WorkflowStatus         `json:"status"` // draft, active, paused, completed, failed
    Nodes       []Node                 `json:"nodes"`
    Edges       []Edge                 `json:"edges"`
    Variables   map[string]interface{} `json:"variables"`
    AgencyID    string                 `json:"agency_id"`
}
```

### Node Types Supported
- `start`: Workflow entry point (manual, scheduled, event trigger)
- `work_item`: Reference to work item to execute
- `decision`: Conditional branching based on conditions
- `parallel`: Execute multiple paths concurrently
- `end`: Workflow completion (success/failure status)

## Key Learnings & Decisions

1. **CRUD Helpers Limitation**: The generic `deleteEntity()` and `saveEntity()` helpers from `crud-helpers.js` assume a URL pattern of `/api/v1/agencies/:id/{entity-type}/:key`, but workflows use `/api/v1/workflows/:id` for individual operations. Custom implementations were needed.

2. **LLM Token Management**: Initial configuration allowed 2-3 workflows with 4000 tokens, which frequently caused truncation. Reducing to 1 workflow with 2500 tokens and lower temperature (0.5) ensured reliable, focused responses.

3. **Route Parameter Consistency**: Gin router requires consistent parameter naming within route groups. All agency routes must use the same parameter name (`:id`) to avoid panics.

4. **Validation Requirements**: Workflow validation requires `agency_id` field to be present for both create and update operations, not just creates.

5. **Template-First Architecture**: Following the codebase pattern, all HTML is generated via `.templ` templates, with JavaScript handling only UI interactions and data operations.

## Files Created
- `internal/builder/ai/workflows_builder.go` (309 lines)
- `internal/web/handlers/ai_refine/workflow_refine_dynamic.go` (190 lines)
- `internal/web/pages/agency_designer/workflows_list_card.templ` (85 lines)
- `internal/web/pages/agency_designer/workflow_editor_card.templ` (95 lines)
- `internal/web/pages/agency_designer/agency_designer_workflows.templ` (98 lines)
- `internal/handlers/workflow_handler_html.go` (45 lines)
- `static/js/agency-designer/workflows.js` (494 lines)

## Files Modified
- `internal/app/app.go` - Registered workflow routes, fixed parameter naming
- `internal/handlers/workflow_handler.go` - Fixed parameter extraction
- `internal/web/pages/agency_designer/agency_designer_overview.templ` - Added WorkflowsContent section
- `static/js/agency-designer/overview.js` - Integrated workflows navigation

## Testing Performed
- ✅ Workflow list loads correctly with search and filters
- ✅ Create new workflow via form
- ✅ Edit existing workflow (loads data, saves updates)
- ✅ Delete workflow with confirmation
- ✅ Duplicate workflow (creates copy with "(Copy)" suffix)
- ✅ AI workflow generation from agency context (creates 1 workflow)
- ✅ Status badges display correct colors
- ✅ Selection tracking enables/disables action buttons
- ✅ Error handling shows user-friendly messages
- ✅ Navigation between sections preserves state

## Dependencies Satisfied
- ✅ MVP-032: Work Items Assignment & Routing (provides work items to orchestrate)

## Enables Future Work
- **MVP-052**: Workflow Visual Designer - Can now load/save workflows; visual designer will enhance editing experience
- Workflow execution engine can reference workflows created through this interface
- AI-generated workflows provide starting point for visual customization

## Notes
- The workflow structure (nodes, edges) is currently edited as JSON in a textarea
- MVP-052 will replace this with a drag-and-drop visual designer using xyflow
- Current implementation focuses on CRUD operations and AI generation
- Execution visualization and runtime controls will be added in future iterations
