/**
 * Workflow Designer - Initialization Module
 * Handles setup: data loading, jsPlumb, panzoom, work items, keyboard shortcuts
 */

(function (window) {
    'use strict';

    if (!window.WorkflowDesigner) {
        window.WorkflowDesigner = {};
    }

    /**
     * Creates initialization functions for the workflow designer
     * @param {Object} context - The Alpine.js component context (this)
     * @param {Object} state - State management object
     * @returns {Object} Initialization methods
     */
    window.WorkflowDesigner.createInit = function (context, state) {
        return {
            /**
             * Main initialization - loads data, sets up jsPlumb and panzoom
             */
            initWorkflow() {
                const container = document.querySelector('.designer-container');
                if (!container) {
                    return;
                }

                // Load workflow data from data attributes
                try {
                    state.agencyId = container.dataset.agencyId;
                    state.workflowId = container.dataset.workflowId || container.dataset.workflowKey;
                    state.workflowKey = container.dataset.workflowKey;
                    state.workflowName = container.dataset.workflowName;
                    state.workflowDescription = container.dataset.workflowDescription || '';
                    state.workflowVersion = container.dataset.workflowVersion;

                    // Parse nodes and edges from JSON
                    const nodesData = container.dataset.workflowNodes;
                    const edgesData = container.dataset.workflowEdges;

                    state.nodes = nodesData ? JSON.parse(nodesData) : [];
                    state.edges = edgesData ? JSON.parse(edgesData) : [];

                    // Deduplicate edges - remove duplicates with same source and target
                    const edgesBefore = state.edges.length;
                    const uniqueEdges = [];
                    const edgeKeys = new Set();

                    state.edges.forEach(edge => {
                        const key = `${edge.source}_${edge.target}`;
                        if (!edgeKeys.has(key)) {
                            edgeKeys.add(key);
                            uniqueEdges.push(edge);
                        }
                    });

                    state.edges = uniqueEdges;

                    // Initialize with Start and End nodes if empty
                    if (state.nodes.length === 0) {
                        state.nodes = [
                            {
                                id: 'start',
                                type: 'start',
                                position: { x: 100, y: 200 },
                                data: {}
                            },
                            {
                                id: 'end',
                                type: 'end',
                                position: { x: 800, y: 200 },
                                data: {}
                            }
                        ];
                    }

                    // Create default edge from start to end if needed
                    const hasStartNode = state.nodes.some(n => n.id === 'start' || n.type === 'start');
                    const hasEndNode = state.nodes.some(n => n.id === 'end' || n.type === 'end');
                    const hasStartToEndEdge = state.edges.some(e => e.source === 'start' && e.target === 'end');

                    if (hasStartNode && hasEndNode && !hasStartToEndEdge && state.edges.length === 0) {
                        state.edges.push({
                            id: 'edge_start_end',
                            source: 'start',
                            target: 'end',
                            type: 'sequential'
                        });
                    }

                    // Initialize specification API with agency ID
                    if (window.specificationAPI) {
                        window.specificationAPI.agencyId = state.agencyId;
                    }
                } catch (error) {

                }

                // Set up jsPlumb, panzoom, and load work items
                this.initJsPlumb();
                this.initPanZoom();
                this.loadWorkItems();

                // Render existing nodes and set up shortcuts
                context.renderNodes();
                this.setupKeyboardShortcuts();

            },

            /**
             * Initialize jsPlumb instance with configuration
             */
            initJsPlumb() {
                const { ready, newInstance } = jsPlumb;

                ready(() => {
                    state.jsPlumbInstance = newInstance({
                        container: context.$refs.canvasViewport,
                        connector: {
                            type: 'Flowchart',
                            options: {
                                cornerRadius: 5,
                                gap: 5
                            }
                        },
                        endpoint: {
                            type: 'Dot',
                            options: {
                                radius: 6
                            }
                        },
                        paintStyle: {
                            stroke: '#3273dc',
                            strokeWidth: 2
                        },
                        hoverPaintStyle: {
                            stroke: '#209cee',
                            strokeWidth: 3
                        },
                        endpointStyle: {
                            fill: '#3273dc',
                            stroke: '#ffffff',
                            strokeWidth: 2
                        },
                        endpointHoverStyle: {
                            fill: '#209cee'
                        }
                    });

                    // Connection event handlers
                    state.jsPlumbInstance.bind('connection', (info) => {
                        context.onConnectionCreated(info);
                    });

                    state.jsPlumbInstance.bind('connectionDetached', (info) => {
                        context.onConnectionRemoved(info);
                    });

                    // Drag event handlers for edge highlighting
                    state.jsPlumbInstance.bind('drag:move', (params) => {
                        const nodeWidth = params.el.offsetWidth;
                        const x = params.pos.x;
                        const left = x;
                        const right = x + nodeWidth;

                        // Check all edges to see if node overlaps horizontally
                        const edgesWithinXBounds = [];
                        state.edges.forEach(edge => {
                            const sourceEl = document.getElementById(edge.source);
                            const targetEl = document.getElementById(edge.target);

                            if (sourceEl && targetEl) {
                                const sourceRect = sourceEl.getBoundingClientRect();
                                const targetRect = targetEl.getBoundingClientRect();
                                const canvasRect = context.$refs.canvasViewport.getBoundingClientRect();
                                const transform = state.panzoomInstance.getTransform();

                                // Calculate edge endpoints in canvas coordinates
                                const x1 = (sourceRect.left + sourceRect.width - canvasRect.left - transform.x) / transform.scale;
                                const x2 = (targetRect.left - canvasRect.left - transform.x) / transform.scale;

                                // Check if the node overlaps with edge's horizontal span
                                const edgeLeft = Math.min(x1, x2);
                                const edgeRight = Math.max(x1, x2);
                                const nodeIsBetweenEdge = left <= edgeRight && right >= edgeLeft;

                                if (nodeIsBetweenEdge) {
                                    edgesWithinXBounds.push(edge);
                                }
                            }
                        });

                        // First, clear all previous highlights
                        state.edges.forEach(edge => {
                            const connections = state.jsPlumbInstance.getConnections({
                                source: edge.source,
                                target: edge.target
                            });
                            if (connections && connections.length > 0) {
                                connections[0].removeClass('edge-highlight');
                                // Reset to default paint style
                                connections[0].setPaintStyle({
                                    stroke: '#3273dc',
                                    strokeWidth: 2
                                });
                            }
                        });

                        if (edgesWithinXBounds.length > 0) {
                            // Highlight edges that overlap with dragged node
                            edgesWithinXBounds.forEach(edge => {
                                const connection = state.jsPlumbInstance.getConnections({
                                    source: edge.source,
                                    target: edge.target
                                })[0];

                                if (connection) {
                                    connection.setPaintStyle({
                                        stroke: '#ff3860',
                                        strokeWidth: 6
                                    });
                                    state.lastDragOverEdge = edge;
                                }
                            });
                        } else {
                            state.lastDragOverEdge = null;
                        }
                    });
                });
            },            /**
             * Initialize pan/zoom functionality for canvas
             */
            initPanZoom() {
                const viewport = context.$refs.canvasViewport;

                state.panzoomInstance = panzoom(viewport, {
                    maxZoom: 3,
                    minZoom: 0.3,
                    smoothScroll: false,
                    bounds: true,
                    boundsPadding: 0.1,
                    zoomDoubleClickSpeed: 1,
                    beforeWheel: (e) => {
                        // Allow panzoom only when Ctrl/Cmd is pressed
                        return !e.ctrlKey && !e.metaKey;
                    }
                });
            },

            /**
             * Load available work items from API
             */
            async loadWorkItems() {
                try {
                    const workItems = await window.specificationAPI.getWorkItems();

                    // Update both state and context (Alpine component) for reactivity
                    state.availableWorkItems = workItems || [];
                    context.availableWorkItems = workItems || [];

                    // Update node titles with work item names if we have them
                    if (state.availableWorkItems.length > 0) {
                        updateWorkItemNodeTitles(context, state);
                    }
                } catch (error) {

                }
            },

            /**
             * Set up keyboard shortcuts for workflow designer
             */
            setupKeyboardShortcuts() {
                document.addEventListener('keydown', (e) => {
                    // Delete selected node with Delete/Backspace
                    if ((e.key === 'Delete' || e.key === 'Backspace') && state.selectedNode) {
                        e.preventDefault();
                        context.deleteNode(state.selectedNode);
                    }

                    // Undo with Ctrl+Z / Cmd+Z
                    if ((e.ctrlKey || e.metaKey) && e.key === 'z' && !e.shiftKey) {
                        e.preventDefault();
                        context.undo();
                    }

                    // Redo with Ctrl+Shift+Z / Cmd+Shift+Z
                    if ((e.ctrlKey || e.metaKey) && e.key === 'z' && e.shiftKey) {
                        e.preventDefault();
                        context.redo();
                    }

                    // Save with Ctrl+S / Cmd+S
                    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                        e.preventDefault();
                        context.saveWorkflow();
                    }
                });
            }
        };
    };

    /**
     * Update work item node titles after work items are loaded
     * @param {Object} context - Alpine.js component context
     * @param {Object} state - State management object
     */
    function updateWorkItemNodeTitles(context, state) {
        let updatedCount = 0;

        state.nodes.forEach(node => {
            if ((node.type === 'work_item' || node.type === 'work-item') && node.data.work_item_key) {
                // ArangoDB uses _key as the primary identifier
                const workItem = state.availableWorkItems.find(wi =>
                    wi._key === node.data.work_item_key || wi.key === node.data.work_item_key
                );
                const nodeEl = document.getElementById(node.id);

                if (nodeEl && workItem) {
                    const titleEl = nodeEl.querySelector('.node-title');
                    if (titleEl) {
                        titleEl.textContent = workItem.title;
                        updatedCount++;
                    }
                }
            }
        });
    }

})(window);
