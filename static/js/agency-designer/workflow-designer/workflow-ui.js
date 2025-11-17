/**
 * Workflow Designer - Workflow Operations & UI Module
 * Handles drag-drop, canvas controls, workflow save/validate/execute, and history
 */

(function (window) {
    'use strict';

    if (!window.WorkflowDesigner) {
        window.WorkflowDesigner = {};
    }

    /**
     * Creates workflow and UI management functions
     * @param {Object} context - The Alpine.js component context (this)
     * @param {Object} state - State management object
     * @returns {Object} Workflow and UI methods
     */
    window.WorkflowDesigner.createWorkflowUI = function (context, state) {
        return {
            // ====== Drag and Drop ======

            /**
             * Toolbox drag start handler
             * @param {DragEvent} event - Drag event
             */
            onToolboxDragStart(event) {
                const toolboxItem = event.target.closest('.toolbox-item');
                const nodeType = toolboxItem.dataset.nodeType;

                let workItemKey = toolboxItem.dataset.workItemKey;

                // If dragging a work item and key is empty, try to get it from the element's index
                if (nodeType === 'work-item' && (!workItemKey || workItemKey === '')) {
                    const toolboxItems = Array.from(document.querySelectorAll('.toolbox-item[data-node-type="work-item"]'));
                    const itemIndex = toolboxItems.indexOf(toolboxItem);

                    if (itemIndex >= 0 && itemIndex < state.availableWorkItems.length) {
                        const workItem = state.availableWorkItems[itemIndex];
                        workItemKey = workItem.key || workItem._key || workItem.id || workItem._id;
                    }
                }

                event.dataTransfer.effectAllowed = 'copy';
                event.dataTransfer.setData('nodeType', nodeType);
                if (workItemKey) {
                    event.dataTransfer.setData('workItemKey', workItemKey);
                }
            },

            /**
             * Canvas drag over handler - visual feedback while dragging
             * @param {DragEvent} event - Drag event
             */
            onCanvasDragOver(event) {
                event.preventDefault();

                try {
                    // Calculate current drag position
                    const canvasRect = context.$refs.canvasViewport.getBoundingClientRect();
                    const transform = state.panzoomInstance.getTransform();
                    const x = (event.clientX - canvasRect.left - transform.x) / transform.scale;
                    const y = (event.clientY - canvasRect.top - transform.y) / transform.scale;

                    // Check if dragging over an edge
                    const nodeWidth = 150;
                    const nodeHeight = 70;
                    const nodeBoxX = x - nodeWidth / 2;
                    const nodeBoxY = y - nodeHeight / 2;

                    // Debug: Show bounds during toolbar drag
                    console.log('[MVP-052][TOOLBAR-DRAG-BOUNDS]', {
                        x: nodeBoxX,
                        y: nodeBoxY,
                        width: nodeWidth,
                        height: nodeHeight,
                        bounds: {
                            left: nodeBoxX,
                            top: nodeBoxY,
                            right: nodeBoxX + nodeWidth,
                            bottom: nodeBoxY + nodeHeight
                        }
                    });

                    const nearEdge = context.findNearbyEdgeWithBox(nodeBoxX, nodeBoxY, nodeWidth, nodeHeight);

                    // Cache the edge and timestamp for use in drop handler
                    if (nearEdge) {
                        state.lastDragOverEdge = nearEdge;
                        state.lastDragOverTime = Date.now();
                        context.highlightEdge(nearEdge.id);
                    } else {
                        context.clearEdgeHighlights();
                    }
                } catch (error) {
                    console.error('Error in drag over handler:', error);
                }
            },

            /**
             * Canvas drop handler
             * @param {DragEvent} event - Drop event
             */
            onCanvasDrop(event) {
                event.preventDefault();

                // Clear any edge highlights
                context.clearEdgeHighlights();

                const nodeType = event.dataTransfer.getData('nodeType');
                const workItemKey = event.dataTransfer.getData('workItemKey');

                if (!nodeType) {
                    return;
                }

                // Calculate drop position relative to canvas viewport
                const canvasRect = context.$refs.canvasViewport.getBoundingClientRect();
                const transform = state.panzoomInstance.getTransform();

                const x = (event.clientX - canvasRect.left - transform.x) / transform.scale;
                const y = (event.clientY - canvasRect.top - transform.y) / transform.scale;

                // Center node box on cursor
                const nodeWidth = 150;
                const nodeHeight = 70;
                const nodeBoxX = x - nodeWidth / 2;
                const nodeBoxY = y - nodeHeight / 2;

                // Determine which edge to split
                let nearEdge = null;
                const now = Date.now();
                const timeSinceLastDragOver = state.lastDragOverTime ? (now - state.lastDragOverTime) : Infinity;

                if (state.lastDragOverEdge && timeSinceLastDragOver < state.dragOverEdgeTimeout) {
                    // Validate that the cached edge still exists
                    const edgeStillExists = state.edges.find(e => e.id === state.lastDragOverEdge.id);

                    if (edgeStillExists) {
                        nearEdge = state.lastDragOverEdge;
                    } else {
                        nearEdge = context.findNearbyEdgeWithBox(nodeBoxX, nodeBoxY, nodeWidth, nodeHeight);
                    }
                } else {
                    nearEdge = context.findNearbyEdgeWithBox(nodeBoxX, nodeBoxY, nodeWidth, nodeHeight);
                }

                // Clear the cache
                state.lastDragOverEdge = null;
                state.lastDragOverTime = null;

                // Create node at the top-left corner of the centered box
                const nodeId = context.createNode(nodeType, nodeBoxX, nodeBoxY, workItemKey);

                // If dropped on an edge, split it
                if (nearEdge && nodeId) {
                    context.splitEdge(nearEdge, nodeId);
                }
            },

            // ====== Canvas Controls ======

            /**
             * Zoom in on the canvas
             */
            zoomIn() {
                const transform = state.panzoomInstance.getTransform();
                const newScale = transform.scale * 1.2;
                state.panzoomInstance.zoomAbs(0, 0, newScale);
            },

            /**
             * Zoom out on the canvas
             */
            zoomOut() {
                const transform = state.panzoomInstance.getTransform();
                const newScale = transform.scale * 0.8;
                state.panzoomInstance.zoomAbs(0, 0, newScale);
            },

            /**
             * Fit canvas to screen
             */
            fitToScreen() {
                state.panzoomInstance.moveTo(0, 0);
                state.panzoomInstance.zoomTo(0, 0, 1);
            },

            /**
             * Toggle minimap visibility
             */
            toggleMinimap() {
                state.showMinimap = !state.showMinimap;
            },

            /**
             * Auto-layout nodes horizontally
             */
            autoLayout() {
                let currentX = 50;
                const y = 100;
                const spacing = 250;

                state.nodes.forEach(node => {
                    node.position.x = currentX;
                    node.position.y = y;
                    currentX += spacing;

                    const nodeEl = document.getElementById(node.id);
                    if (nodeEl) {
                        nodeEl.style.left = `${node.position.x}px`;
                        nodeEl.style.top = `${node.position.y}px`;
                    }
                });

                state.jsPlumbInstance.repaintEverything();
                context.saveToHistory();
            },

            // ====== Workflow Operations ======

            /**
             * Validate workflow structure
             */
            validateWorkflow() {
                const errors = [];

                // Check for start node
                const startNodes = state.nodes.filter(n => n.type === 'start');
                if (startNodes.length === 0) {
                    errors.push('Workflow must have a Start node');
                } else if (startNodes.length > 1) {
                    errors.push('Workflow can only have one Start node');
                }

                // Check for end node
                const endNodes = state.nodes.filter(n => n.type === 'end');
                if (endNodes.length === 0) {
                    errors.push('Workflow must have at least one End node');
                }

                // Check for orphaned nodes
                const connectedNodes = new Set();
                state.edges.forEach(edge => {
                    connectedNodes.add(edge.source);
                    connectedNodes.add(edge.target);
                });

                state.nodes.forEach(node => {
                    if (node.type !== 'start' && !connectedNodes.has(node.id)) {
                        errors.push(`Node "${node.data.name}" is not connected`);
                    }
                });

                // Show results
                if (errors.length === 0) {
                    alert('✓ Workflow is valid');
                } else {
                    alert('Validation Errors:\n\n' + errors.join('\n'));
                }
            },

            /**
             * Save workflow to the server
             */
            async saveWorkflow() {
                state.saving = true;

                try {
                    // Fetch the current specification to get ALL workflows
                    const specResponse = await fetch(`/api/v1/agencies/${state.agencyId}/specification`);
                    if (!specResponse.ok) {
                        throw new Error('Failed to fetch current specification');
                    }

                    const spec = await specResponse.json();
                    const allWorkflows = spec.workflows || [];

                    // Build the updated workflow object
                    const updatedWorkflow = {
                        _key: state.workflowKey,
                        name: state.workflowName,
                        description: state.workflowDescription,
                        agency_id: state.agencyId,
                        version: state.workflowVersion,
                        nodes: state.nodes,
                        edges: state.edges
                    };

                    // Find and update the specific workflow, or add if new
                    const workflowIndex = allWorkflows.findIndex(wf =>
                        wf._key === state.workflowKey || wf.key === state.workflowKey
                    );

                    if (workflowIndex >= 0) {
                        allWorkflows[workflowIndex] = updatedWorkflow;
                    } else {
                        allWorkflows.push(updatedWorkflow);
                    }

                    // Save ALL workflows back to the specification
                    const response = await fetch(`/api/v1/agencies/${state.agencyId}/specification/workflows`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify({
                            workflows: allWorkflows,
                            updated_by: 'system'
                        })
                    });

                    if (response.ok) {
                        alert('✓ Workflow saved successfully');
                    } else {
                        const error = await response.json();
                        alert('Failed to save workflow: ' + (error.error || 'Unknown error'));
                    }
                } catch (error) {
                    alert('Failed to save workflow: ' + error.message);
                } finally {
                    state.saving = false;
                }
            },

            /**
             * Execute workflow (placeholder)
             */
            executeWorkflow() {
                alert('Workflow execution will be implemented in Phase 6');
            },

            // ====== History Management ======

            /**
             * Save current state to history
             */
            saveToHistory() {
                const historyState = {
                    nodes: JSON.parse(JSON.stringify(state.nodes)),
                    edges: JSON.parse(JSON.stringify(state.edges))
                };

                // Remove future states if we're not at the end
                state.history = state.history.slice(0, state.historyIndex + 1);

                state.history.push(historyState);
                state.historyIndex = state.history.length - 1;

                this.updateHistoryButtons();
            },

            /**
             * Update undo/redo button states
             */
            updateHistoryButtons() {
                state.canUndo = state.historyIndex > 0;
                state.canRedo = state.historyIndex < state.history.length - 1;
            },

            /**
             * Undo last action
             */
            undo() {
                if (state.canUndo) {
                    state.historyIndex--;
                    this.restoreState(state.history[state.historyIndex]);
                }
            },

            /**
             * Redo last undone action
             */
            redo() {
                if (state.canRedo) {
                    state.historyIndex++;
                    this.restoreState(state.history[state.historyIndex]);
                }
            },

            /**
             * Restore state from history
             * @param {Object} historyState - State to restore
             */
            restoreState(historyState) {
                state.nodes = JSON.parse(JSON.stringify(historyState.nodes));
                state.edges = JSON.parse(JSON.stringify(historyState.edges));

                // Clear canvas
                context.$refs.canvasViewport.innerHTML = '';
                state.jsPlumbInstance.reset();

                // Re-render
                context.renderNodes();
                this.updateHistoryButtons();
            }
        };
    };

})(window);
