/**
 * Workflow Designer - Node Management Module
 * Handles node creation, rendering, selection, and deletion
 */

(function (window) {
    'use strict';

    if (!window.WorkflowDesigner) {
        window.WorkflowDesigner = {};
    }

    /**
     * Creates node management functions
     * @param {Object} context - The Alpine.js component context (this)
     * @param {Object} state - State management object
     * @returns {Object} Node management methods
     */
    window.WorkflowDesigner.createNodes = function (context, state) {
        return {
            /**
             * Create a new node at the specified position
             * @param {string} type - Node type (start, end, decision, parallel, work-item)
             * @param {number} x - X position
             * @param {number} y - Y position
             * @param {string|null} workItemKey - Work item key for work-item nodes
             * @returns {string} Node ID
             */
            createNode(type, x, y, workItemKey = null) {
                console.log('[MVP-052][CREATE-NODE] Starting', { type, x, y, workItemKey });

                state.nodeCounter++;
                const nodeId = `node_${Date.now()}_${state.nodeCounter}`;

                let nodeName = type.charAt(0).toUpperCase() + type.slice(1);
                let nodeData = { name: nodeName };

                // For work item nodes, get work item details
                if (type === 'work-item' && workItemKey) {
                    const workItem = state.availableWorkItems.find(wi =>
                        wi._key === workItemKey || wi.key === workItemKey
                    );

                    if (workItem) {
                        nodeName = workItem.title;
                        nodeData = {
                            name: workItem.title,
                            description: workItem.description || '',
                            work_item_key: workItemKey,
                            type: workItem.type
                        };
                    } else {
                        nodeData = {
                            name: 'Loading...',
                            work_item_key: workItemKey
                        };
                    }
                }

                const node = {
                    id: nodeId,
                    type: type,
                    position: { x, y },
                    data: nodeData
                };

                state.nodes.push(node);
                console.log('[MVP-052][CREATE-NODE] Calling renderNode', nodeId);
                this.renderNode(node);
                console.log('[MVP-052][CREATE-NODE] Node created', nodeId);
                context.saveToHistory();

                return nodeId;
            },

            /**
             * Render all nodes on the canvas
             */
            renderNodes() {
                if (state.nodes.length === 0) {
                    return;
                }

                state.nodes.forEach(node => {
                    this.renderNode(node);
                });

                // Restore connections after nodes are rendered
                context.$nextTick(() => {
                    context.renderConnections();
                });
            },

            /**
             * Render a single node
             * @param {Object} node - Node object to render
             */
            renderNode(node) {
                console.log('[MVP-052][RENDER-NODE] Starting', {
                    nodeId: node.id,
                    hasViewport: !!context.$refs?.canvasViewport,
                    hasJsPlumb: !!state.jsPlumbInstance
                });

                const nodeEl = document.createElement('div');
                nodeEl.id = node.id;
                nodeEl.className = `workflow-node node-${node.type}`;
                nodeEl.style.left = `${node.position.x}px`;
                nodeEl.style.top = `${node.position.y}px`;

                // Get node icon and color based on type
                const { icon, iconColor } = getNodeIconAndColor(node.type);

                // Get display name for the node
                const displayName = getNodeDisplayName(node, state.availableWorkItems);

                nodeEl.innerHTML = `
                    <div class="node-content">
                        <div class="node-header">
                            <span class="node-icon ${iconColor}"><i class="fas ${icon}"></i></span>
                            <span class="node-title">${displayName}</span>
                        </div>
                    </div>
                `;

                // Add to canvas
                console.log('[MVP-052][RENDER-NODE] Appending to DOM');
                context.$refs.canvasViewport.appendChild(nodeEl);
                console.log('[MVP-052][RENDER-NODE] Appended successfully');

                // Make node draggable using jsPlumb v6 API
                state.jsPlumbInstance.setDraggable(nodeEl, true);

                // Add drag stop listener to update position
                state.jsPlumbInstance.on(nodeEl, 'stop', (params) => {
                    const nodeObj = state.nodes.find(n => n.id === node.id);
                    if (nodeObj) {
                        nodeObj.position.x = params.pos.x;
                        nodeObj.position.y = params.pos.y;

                        // Check auto-disconnect and auto-connect
                        context.checkAutoDisconnect(node.id);
                        context.checkAutoConnect(node.id);

                        context.saveToHistory();
                    }
                });

                // Add source endpoint (right side)
                state.jsPlumbInstance.addEndpoint(nodeEl, {
                    anchor: 'Right',
                    source: true,
                    maxConnections: node.type === 'decision' || node.type === 'parallel' ? -1 : 1,
                    cssClass: 'endpoint source'
                });

                // Add target endpoint (left side) - except for start node
                if (node.type !== 'start') {
                    state.jsPlumbInstance.addEndpoint(nodeEl, {
                        anchor: 'Left',
                        target: true,
                        maxConnections: -1,
                        cssClass: 'endpoint target'
                    });
                }

                // Click to select
                nodeEl.addEventListener('click', (e) => {
                    e.stopPropagation();
                    this.selectNode(node.id);
                });
            },

            /**
             * Select a node
             * @param {string} nodeId - ID of node to select
             */
            selectNode(nodeId) {
                // Deselect all
                document.querySelectorAll('.workflow-node').forEach(el => {
                    el.classList.remove('selected');
                });

                // Select this node
                const nodeEl = document.getElementById(nodeId);
                if (nodeEl) {
                    nodeEl.classList.add('selected');
                    const node = state.nodes.find(n => n.id === nodeId);
                    state.selectedNode = node;
                    // Update Alpine reactive property
                    context.selectedNode = node;
                }
            },

            /**
             * Update node property
             * @param {string} property - Property name to update
             * @param {*} value - New value
             */
            updateNodeProperty(property, value) {
                if (state.selectedNode) {
                    state.selectedNode.data[property] = value;
                    context.saveToHistory();

                    // Update visual
                    if (property === 'name') {
                        const nodeEl = document.getElementById(state.selectedNode.id);
                        if (nodeEl) {
                            const titleEl = nodeEl.querySelector('.node-title');
                            if (titleEl) titleEl.textContent = value;
                        }
                    }
                }
            },

            /**
             * Delete the selected node
             */
            deleteSelectedNode() {
                if (!state.selectedNode) return;

                // Don't allow deleting Start or End nodes
                if (state.selectedNode.type === 'start' || state.selectedNode.type === 'end') {
                    alert('Cannot delete Start or End nodes');
                    return;
                }

                if (confirm('Delete this node?')) {
                    const nodeId = state.selectedNode.id;
                    const nodeEl = document.getElementById(nodeId);

                    if (nodeEl) {
                        // Remove all connections first
                        state.jsPlumbInstance.removeAllEndpoints(nodeEl);

                        // Remove from jsPlumb management
                        state.jsPlumbInstance.unmanage(nodeEl);

                        // Remove from DOM
                        nodeEl.remove();
                    }

                    // Remove from nodes array
                    state.nodes = state.nodes.filter(n => n.id !== nodeId);

                    // Remove any edges connected to this node
                    state.edges = state.edges.filter(e =>
                        e.source !== nodeId && e.target !== nodeId
                    );

                    state.selectedNode = null;
                    // Clear Alpine reactive property
                    context.selectedNode = null;
                    context.saveToHistory();
                }
            }
        };
    };

    /**
     * Get icon and color for node type
     * @param {string} type - Node type
     * @returns {Object} Icon class and color class
     */
    function getNodeIconAndColor(type) {
        const iconMap = {
            'start': { icon: 'fa-play-circle', iconColor: 'has-text-success' },
            'end': { icon: 'fa-stop-circle', iconColor: 'has-text-danger' },
            'decision': { icon: 'fa-question-circle', iconColor: 'has-text-warning' },
            'parallel': { icon: 'fa-code-branch', iconColor: 'has-text-info' },
            'work_item': { icon: 'fa-tasks', iconColor: 'has-text-link' },
            'work-item': { icon: 'fa-tasks', iconColor: 'has-text-link' }
        };

        return iconMap[type] || { icon: 'fa-circle', iconColor: 'has-text-grey' };
    }

    /**
     * Get display name for a node
     * @param {Object} node - Node object
     * @param {Array} availableWorkItems - List of available work items
     * @returns {string} Display name
     */
    function getNodeDisplayName(node, availableWorkItems) {
        let displayName = node.data.name || node.type;

        // For work items, prioritize the title lookup from availableWorkItems
        if ((node.type === 'work_item' || node.type === 'work-item') && node.data.work_item_key) {
            const workItem = availableWorkItems.find(wi =>
                wi._key === node.data.work_item_key || wi.key === node.data.work_item_key
            );

            if (workItem) {
                displayName = workItem.title;
            } else if (node.data.name && node.data.name !== 'work_item' && node.data.name !== 'work-item') {
                displayName = node.data.name;
            } else if (availableWorkItems.length === 0) {
                displayName = 'Loading...';
            } else {
                displayName = `Work Item (${node.data.work_item_key.substring(0, 8)}...)`;
            }
        }

        return displayName;
    }

})(window);
