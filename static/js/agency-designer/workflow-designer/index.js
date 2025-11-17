/**
 * Workflow Designer - Main Orchestration Module
 * Combines all sub-modules into a complete Alpine.js component
 * 
 * Dependencies (load in this order):
 * 1. utils.js
 * 2. state.js
 * 3. init.js
 * 4. nodes.js
 * 5. edges.js
 * 6. workflow-ui.js
 * 7. index.js (this file)
 */

(function (window) {
    'use strict';

    if (!window.WorkflowDesigner) {
        window.WorkflowDesigner = {};
    }

    /**
     * Main entry point - creates complete Alpine.js component
     * @returns {Object} Alpine.js component data and methods
     */
    window.WorkflowDesigner.init = function () {
        // Create utility functions
        const utils = window.WorkflowDesigner.createUtils();

        // Create state management
        const state = window.WorkflowDesigner.createState();

        // Placeholder context - will be bound to Alpine component
        const context = {
            $refs: null,
            $nextTick: null
        };

        // Create module factories (they need context, so we'll call them after Alpine binds)
        const componentData = Object.assign({}, state, {
            /**
             * Alpine.js init method - called when component is initialized
             */
            init() {
                // Bind context references
                context.$refs = this.$refs;
                context.$nextTick = this.$nextTick.bind(this);

                // Create module instances with proper context binding
                const initModule = window.WorkflowDesigner.createInit(this, state);
                const nodesModule = window.WorkflowDesigner.createNodes(this, state);
                const edgesModule = window.WorkflowDesigner.createEdges(this, state, utils);
                const workflowUIModule = window.WorkflowDesigner.createWorkflowUI(this, state);

                // Bind all module methods to this component
                Object.assign(this, initModule, nodesModule, edgesModule, workflowUIModule);

                // Call initialization
                this.init();
            }
        });

        return componentData;
    };

})(window);
