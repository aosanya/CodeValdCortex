# Workflow Designer Refactoring Plan

## Current State
- `workflow-designer.js`: 1244 lines (original monolithic)
- `workflow-designer/index.js`: 1131 lines (❌ EXCEEDS 500-line limit)
- `workflow-designer/utils.js`: 125 lines (✅ Pure functions)
- `workflow-designer/state.js`: 57 lines (✅ State management)

## Target Module Structure (Per Updated Prompt)

### Module Breakdown (All ≤ 500 lines)

1. **utils.js** (125 lines) ✅ COMPLETE
   - Pure geometry functions
   - Line intersection calculations
   - **100% testable without mocks**

2. **state.js** (57 lines) ✅ COMPLETE
   - Initial state factory
   - No business logic

3. **init.js** (~150 lines) - NEEDS CREATION
   - Data loading from HTML attributes
   - jsPlumb initialization
   - Panzoom setup
   - Work items loading
   - Keyboard shortcuts
   - **Dependencies**: state

4. **nodes.js** (~300 lines) - NEEDS CREATION
   - createNode, renderNode, renderNodes
   - selectNode, updateNode, deleteNode
   - updateWorkItemNodeTitles
   - **Dependencies**: state, utils

5. **edges.js** (~400 lines) - NEEDS CREATION
   - findNearbyEdge, findNearbyEdgeWithBox
   - highlightEdge, clearEdgeHighlights
   - splitEdge
   - renderConnections
   - onConnectionCreated
   - checkAutoDisconnect, checkAutoConnect
   - **Dependencies**: state, utils, nodes

6. **drag-drop.js** (~150 lines) - NEEDS CREATION
   - onToolboxDragStart
   - onCanvasDragOver
   - onCanvasDrop
   - **Dependencies**: state, nodes, edges

7. **workflow.js** (~200 lines) - NEEDS CREATION
   - validateWorkflow
   - saveWorkflow  
   - executeWorkflow
   - **Dependencies**: state, nodes, edges

8. **history.js** (~100 lines) - NEEDS CREATION
   - saveHistory
   - undo, redo
   - **Dependencies**: state, nodes, edges

9. **canvas.js** (~80 lines) - NEEDS CREATION
   - Zoom controls
   - Pan controls
   - Auto-layout
   - **Dependencies**: state

10. **index.js** (~80 lines) - NEEDS RECREATION
    - Orchestrates all modules
    - Creates combined Alpine component
    - **Dependencies**: ALL above

## Total: ~1500 lines across 10 modules (avg 150 lines/module)

## Implementation Steps

1. ✅ Backup original: `workflow-designer.monolithic.backup.js`
2. ✅ Created utils.js and state.js
3. ⏳ Extract remaining modules from current index.js
4. ⏳ Create new orchestration index.js
5. ⏳ Replace workflow-designer.js with minimal loader
6. ⏳ Find and update HTML template

## Testability Improvements

- **Pure modules** (100% testable): utils.js
- **Stateful** (easy to test): state.js, nodes.js
- **Side effects** (need mocks): init.js, drag-drop.js, canvas.js, workflow.js
- **Orchestration**: index.js

## Dependencies Flow
```
utils.js (no deps)
state.js (no deps)
  ↓
init.js (state)
canvas.js (state)
  ↓
nodes.js (state, utils)
  ↓
edges.js (state, utils, nodes)
  ↓
drag-drop.js (state, nodes, edges)
workflow.js (state, nodes, edges)
history.js (state, nodes, edges)
  ↓
index.js (ALL)
```
