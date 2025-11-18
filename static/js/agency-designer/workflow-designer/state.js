/**
 * Workflow Designer - State Module
 * State management & data structures
 * No dependencies
 */

(function (window) {
    'use strict';

    // Initialize namespace if it doesn't exist
    if (!window.WorkflowDesigner) {
        window.WorkflowDesigner = {};
    }

    /**
     * Create initial state object
     * Contains all workflow data and UI state
     */
    window.WorkflowDesigner.createState = function () {
        return {
            // Workflow metadata
            workflowId: '',
            workflowKey: '',
            workflowName: '',
            workflowDescription: '',
            workflowVersion: '',
            agencyId: '',

            // Workflow graph data
            nodes: [],
            edges: [],

            // UI state
            selectedNode: null,
            availableWorkItems: [],
            jsPlumbInstance: null,
            panzoomInstance: null,
            showMinimap: false,
            saving: false,

            // Drag state tracking for edge splitting
            lastDragOverEdge: null,
            lastDragOverTime: null,
            dragOverEdgeTimeout: 2000, // milliseconds - 2 seconds to allow for slow/careful drags

            // Undo/Redo state
            history: [],
            historyIndex: -1,
            canUndo: false,
            canRedo: false,

            // Node counter for unique IDs
            nodeCounter: 0
        };
    };

})(window);
