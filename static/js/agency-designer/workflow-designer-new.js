/**
 * Workflow Designer - Module Loader
 * 
 * This is a minimal loader that depends on modular components.
 * 
 * IMPORTANT: Load modules in this order BEFORE this file:
 * 1. ./workflow-designer/utils.js       - Pure utility functions
 * 2. ./workflow-designer/state.js       - State management
 * 3. ./workflow-designer/init.js        - Initialization
 * 4. ./workflow-designer/nodes.js       - Node management
 * 5. ./workflow-designer/edges.js       - Edge management
 * 6. ./workflow-designer/workflow-ui.js - Workflow operations & UI
 * 7. ./workflow-designer/index.js       - Orchestration (combines modules)
 * 8. ./workflow-designer.js (this file) - Alpine.js registration
 * 
 * See ./workflow-designer/README.md for complete documentation.
 * 
 * Features:
 * - Drag and drop nodes from toolbox
 * - Auto-split edges when dropping nodes on them
 * - Connect nodes with edges
 * - Pan and zoom canvas
 * - Undo/redo history
 * - Save/load workflows
 * 
 * Dependencies:
 * - Alpine.js v3
 * - jsPlumb v6
 * - panzoom
 */

document.addEventListener('alpine:init', () => {
    Alpine.data('workflowDesigner', () => window.WorkflowDesigner.init());
});
