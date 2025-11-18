# Workflow Designer - Modular Implementation

This directory contains the modularized version of `workflow-designer.js`.

## Module Structure

- **`utils.js`** - Utility functions (line intersection, geometry calculations)
  - Line segment intersection detection
  - Box-to-line intersection checking
  - Distance calculations
  - No dependencies

- **`state.js`** - State management & data structures
  - Creates the base state object with all workflow data
  - Workflow metadata (ID, name, description, version)
  - Nodes and edges arrays
  - UI state (selected node, saving status)
  - Drag state tracking for edge splitting
  - History state for undo/redo
  - Depends on: none

- **`init.js`** - Initialization logic
  - Parses workflow data from HTML attributes
  - Initializes jsPlumb instance
  - Sets up panzoom for canvas navigation
  - Loads available work items from API
  - Sets up keyboard shortcuts
  - Depends on: state

- **`nodes.js`** - Node CRUD operations & rendering
  - Create, render, update, delete nodes
  - Node selection and highlighting
  - Work item node title updates
  - jsPlumb draggable configuration
  - Depends on: state, utils

- **`edges.js`** - Edge/connection management & auto-split
  - Find nearby edges during drag
  - Split edges when dropping nodes
  - Edge highlighting (visual feedback)
  - Auto-connect and auto-disconnect logic
  - Connection event handlers
  - Render connections
  - Depends on: state, utils, nodes

- **`drag-drop.js`** - Drag and drop from toolbox
  - Toolbox drag start handler
  - Canvas dragover handler (visual feedback)
  - Canvas drop handler
  - Integration with edge splitting
  - Depends on: state, nodes, edges

- **`canvas.js`** - Canvas controls (zoom, pan, layout)
  - Zoom in/out functions
  - Reset zoom
  - Pan functions
  - Auto-layout
  - Depends on: state

- **`history.js`** - Undo/redo functionality
  - Save state to history
  - Undo operation
  - Redo operation
  - History state management
  - Depends on: state, nodes, edges

- **`workflow.js`** - Workflow save, validate, execute operations
  - Validate workflow structure
  - Save workflow to API
  - Execute workflow
  - Depends on: state, nodes, edges

- **`index.js`** - Main entry point combining all modules
  - Combines all module factories
  - Returns the complete `workflowDesigner()` function for Alpine.js
  - Depends on: ALL modules above

## Loading Order (Browser)

When using script tags, load in this order:

```html
<!-- 1. Utilities (no dependencies) -->
<script src="/static/js/agency-designer/workflow-designer/utils.js"></script>

<!-- 2. State (no dependencies) -->
<script src="/static/js/agency-designer/workflow-designer/state.js"></script>

<!-- 3. Init (depends on state) -->
<script src="/static/js/agency-designer/workflow-designer/init.js"></script>

<!-- 4. Nodes (depends on state, utils) -->
<script src="/static/js/agency-designer/workflow-designer/nodes.js"></script>

<!-- 5. Edges (depends on state, utils, nodes) -->
<script src="/static/js/agency-designer/workflow-designer/edges.js"></script>

<!-- 6. Drag-drop (depends on state, nodes, edges) -->
<script src="/static/js/agency-designer/workflow-designer/drag-drop.js"></script>

<!-- 7. Canvas (depends on state) -->
<script src="/static/js/agency-designer/workflow-designer/canvas.js"></script>

<!-- 8. History (depends on state, nodes, edges) -->
<script src="/static/js/agency-designer/workflow-designer/history.js"></script>

<!-- 9. Workflow operations (depends on state, nodes, edges) -->
<script src="/static/js/agency-designer/workflow-designer/workflow.js"></script>

<!-- 10. Main entry point (depends on all above) -->
<script src="/static/js/agency-designer/workflow-designer/index.js"></script>

<!-- 11. Original file (now just loads the modular version) -->
<script src="/static/js/agency-designer/workflow-designer.js"></script>
```

## Dependencies

```
utils.js          → (no dependencies)
state.js          → (no dependencies)
init.js           → state
nodes.js          → state, utils
edges.js          → state, utils, nodes
drag-drop.js      → state, nodes, edges
canvas.js         → state
history.js        → state, nodes, edges
workflow.js       → state, nodes, edges
index.js          → ALL modules above
```

## Module Pattern

All modules use the browser-compatible IIFE (Immediately Invoked Function Expression) pattern:

```javascript
(function(window) {
    'use strict';
    
    // Initialize namespace
    if (!window.WorkflowDesigner) {
        window.WorkflowDesigner = {};
    }
    
    // Attach factory function
    window.WorkflowDesigner.createModuleName = function(context) {
        return {
            method1: function() {
                // Implementation
            },
            method2: function() {
                // Implementation
            }
        };
    };
    
})(window);
```

## Migration Status

- [x] Backup created: `workflow-designer.monolithic.backup.js`
- [x] Module directory created
- [x] README.md created
- [ ] Modules created with browser-compatible IIFE format
- [ ] Original file replaced with loader
- [ ] HTML template updated with script tags
- [ ] Testing complete

## File Sizes

| File | Lines | Description |
|------|-------|-------------|
| Original | 1244 | Monolithic implementation |
| utils.js | ~100 | Geometry utilities |
| state.js | ~50 | State management |
| init.js | ~120 | Initialization |
| nodes.js | ~250 | Node operations |
| edges.js | ~350 | Edge operations |
| drag-drop.js | ~100 | Drag and drop |
| canvas.js | ~60 | Canvas controls |
| history.js | ~80 | Undo/redo |
| workflow.js | ~120 | Workflow operations |
| index.js | ~80 | Main entry |
| **Total modules** | **~1310** | (Slightly more due to module boilerplate) |
| New main file | ~40 | Minimal loader |

## Global Namespace

All modules attach to `window.WorkflowDesigner` namespace:

```javascript
window.WorkflowDesigner = {
    createState: function() { ... },
    createUtils: function() { ... },
    createInitMethods: function(context) { ... },
    createNodeMethods: function(context) { ... },
    createEdgeMethods: function(context) { ... },
    createDragDropMethods: function(context) { ... },
    createCanvasMethods: function(context) { ... },
    createHistoryMethods: function(context) { ... },
    createWorkflowMethods: function(context) { ... },
    init: function() { ... } // Main factory from index.js
};
```

## Usage

After all modules are loaded, the `workflowDesigner()` function is available globally for Alpine.js:

```html
<div x-data="workflowDesigner()">
    <!-- Workflow designer UI -->
</div>
```

## Notes

- All modules are browser-compatible (no build step required)
- No ES6 import/export keywords used
- Each module is self-contained and focused
- Clear dependency chain
- Maintains full backward compatibility with Alpine.js integration
