# Workbench Chat Panel & Workflows Section Integration

## Overview
This document outlines the design and integration plan for adding a chat panel and workflows section to the Workbench, matching the patterns used in the Agency Designer.

## Design Summary
- **Chat Panel**: Reuse `ChatPanel` from `agency_designer/chat_panel.templ` for consistent UI/UX and backend logic.
- **Workflows Section**: Display workflows from agency tags using a Bulma panel, matching the designer’s left panel layout.
- **Layout**: Use shared layout components (`LayoutWithAgency`) for page structure.
- **Backend Wiring**: All chat and issue creation logic should follow the same service/repository/handler pattern as goals.
- **Dynamic Updates**: Use HTMX for UI updates, scoped to agency context.

## Implementation Steps
1. Update workbench template to use shared layout and include workflows panel and chat panel.
2. Fetch workflows from agency tags via agency specification service.
3. Integrate `ChatPanel` with agency context and conversation state.
4. Wire up backend for chat/issue creation using the goals pattern.
5. Ensure HTMX is used for dynamic updates.
6. Test for multi-tenancy, UI consistency, and correct issue creation.

## Open Questions / Gaps
- Future Kanban features for workflows panel (drag-and-drop, inline editing) are deferred.
- Ensure chat panel supports referencing and creating issues with context from workflows/goals.

## References
- `internal/web/pages/agency_designer/chat_panel.templ`
- `internal/web/components/layout_with_agency.templ`
- Agency specification service/repository
- Goals panel implementation
