/**
 * Workflow Designer - Visual workflow builder using jsPlumb
 * Integrates with Alpine.js for state management
 * Uses global specificationAPI
 */

// Alpine.js component for workflow designer
function workflowDesigner() {
    return {
        // State
        workflowId: '',
        workflowKey: '',
        workflowName: '',
        workflowDescription: '',
        workflowVersion: '',
        agencyId: '',
        nodes: [],
        edges: [],
        selectedNode: null,
        availableWorkItems: [],
        jsPlumbInstance: null,
        panzoomInstance: null,
        showMinimap: false,
        saving: false,

        // Undo/Redo
        history: [],
        historyIndex: -1,
        canUndo: false,
        canRedo: false,

        // Node counter for unique IDs
        nodeCounter: 0,

        // Initialize
        init() {
            // Get the designer container element
            const container = document.querySelector('.designer-container');
            if (!container) {
                return;
            }

            // Read workflow data from data attributes
            try {
                this.agencyId = container.dataset.agencyId;
                this.workflowId = container.dataset.workflowId || container.dataset.workflowKey; // Use key as fallback
                this.workflowKey = container.dataset.workflowKey;
                this.workflowName = container.dataset.workflowName;
                this.workflowDescription = container.dataset.workflowDescription || '';
                this.workflowVersion = container.dataset.workflowVersion;

                // Parse nodes and edges from JSON
                const nodesData = container.dataset.workflowNodes;
                const edgesData = container.dataset.workflowEdges;

                this.nodes = nodesData ? JSON.parse(nodesData) : [];
                this.edges = edgesData ? JSON.parse(edgesData) : [];

                // If workflow is empty, initialize with Start and End nodes
                if (this.nodes.length === 0) {
                    this.nodes = [
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

                // Initialize specification API with agency ID
                if (window.specificationAPI) {
                    window.specificationAPI.agencyId = this.agencyId;
                }
            } catch (error) {
            }

            // Initialize jsPlumb
            this.initJsPlumb();

            // Initialize panzoom for canvas navigation
            this.initPanZoom();

            // Load available work items
            this.loadWorkItems();

            // Render existing nodes
            this.renderNodes();

            // Set up keyboard shortcuts
            this.setupKeyboardShortcuts();
        },

        // Initialize jsPlumb instance
        initJsPlumb() {
            const { ready, newInstance } = jsPlumb;

            ready(() => {
                this.jsPlumbInstance = newInstance({
                    container: this.$refs.canvasViewport,
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
                this.jsPlumbInstance.bind('connection', (info) => {
                    this.onConnectionCreated(info);
                });

                this.jsPlumbInstance.bind('connectionDetached', (info) => {
                    this.onConnectionRemoved(info);
                });

            });
        },

        // Initialize pan/zoom
        initPanZoom() {
            const viewport = this.$refs.canvasViewport;

            this.panzoomInstance = panzoom(viewport, {
                maxZoom: 3,
                minZoom: 0.3,
                smoothScroll: false,
                bounds: true,
                boundsPadding: 0.1,
                zoomDoubleClickSpeed: 1,
                beforeWheel: (e) => {
                    // Allow panzoom only when Ctrl/Cmd is pressed
                    const shouldIgnore = !e.ctrlKey && !e.metaKey;
                    return shouldIgnore;
                }
            });
        },

        // Load available work items from API
        async loadWorkItems() {
            try {
                const workItems = await window.specificationAPI.getWorkItems();
                this.availableWorkItems = workItems || [];

                // Update node titles with work item names
                if (this.availableWorkItems.length > 0) {
                    this.updateWorkItemNodeTitles();
                }
            } catch (error) {
            }
        },

        // Update work item node titles after work items are loaded
        updateWorkItemNodeTitles() {
            let updatedCount = 0;

            this.nodes.forEach(node => {
                if ((node.type === 'work_item' || node.type === 'work-item') && node.data.work_item_key) {
                    // ArangoDB uses _key as the primary identifier
                    const workItem = this.availableWorkItems.find(wi => wi._key === node.data.work_item_key || wi.key === node.data.work_item_key);
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
        },

        // Toolbox drag start
        onToolboxDragStart(event) {
            const toolboxItem = event.target.closest('.toolbox-item');
            const nodeType = toolboxItem.dataset.nodeType;

            // For work items, find the work item key from Alpine data
            let workItemKey = toolboxItem.dataset.workItemKey;

            // If we're dragging a work item and the key is empty, try to get it from the element's index
            if (nodeType === 'work-item' && (!workItemKey || workItemKey === '')) {
                // Get the index of this toolbox item
                const toolboxItems = Array.from(document.querySelectorAll('.toolbox-item[data-node-type="work-item"]'));
                const itemIndex = toolboxItems.indexOf(toolboxItem);

                if (itemIndex >= 0 && itemIndex < this.availableWorkItems.length) {
                    const workItem = this.availableWorkItems[itemIndex];

                    // Try different possible key properties
                    workItemKey = workItem.key || workItem._key || workItem.id || workItem._id;
                }
            }

            event.dataTransfer.effectAllowed = 'copy';
            event.dataTransfer.setData('nodeType', nodeType);
            if (workItemKey) {
                event.dataTransfer.setData('workItemKey', workItemKey);
            }
        },        // Canvas drop handler
        onCanvasDrop(event) {
            event.preventDefault();

            const nodeType = event.dataTransfer.getData('nodeType');
            const workItemKey = event.dataTransfer.getData('workItemKey');

            if (!nodeType) return;

            // Calculate drop position relative to canvas viewport
            const canvasRect = this.$refs.canvasViewport.getBoundingClientRect();
            const transform = this.panzoomInstance.getTransform();

            const x = (event.clientX - canvasRect.left - transform.x) / transform.scale;
            const y = (event.clientY - canvasRect.top - transform.y) / transform.scale;

            // Create node
            this.createNode(nodeType, x, y, workItemKey);
        },

        // Create a new node
        createNode(type, x, y, workItemKey = null) {
            this.nodeCounter++;
            const nodeId = `node_${Date.now()}_${this.nodeCounter}`;

            let nodeName = type.charAt(0).toUpperCase() + type.slice(1);
            let nodeData = { name: nodeName };

            // If it's a work item node, get the work item details
            if (type === 'work-item' && workItemKey) {
                // ArangoDB uses _key as the primary identifier
                const workItem = this.availableWorkItems.find(wi => wi._key === workItemKey || wi.key === workItemKey);

                if (workItem) {
                    // Work item found - use its details
                    nodeName = workItem.title;
                    nodeData = {
                        name: workItem.title,
                        description: workItem.description || '',
                        work_item_key: workItemKey,
                        type: workItem.type
                    };
                } else {
                    // Work item not found yet, but still store the key
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

            this.nodes.push(node);
            this.renderNode(node);
            this.saveToHistory();

        },

        // Render all nodes
        renderNodes() {
            if (this.nodes.length === 0) {
                return;
            }

            this.nodes.forEach((node, index) => {
                this.renderNode(node);
            });

            // Restore connections after nodes are rendered
            this.$nextTick(() => {
                this.renderConnections();
            });
        },

        // Render a single node
        renderNode(node) {
            const nodeEl = document.createElement('div');
            nodeEl.id = node.id;
            nodeEl.className = `workflow-node node-${node.type}`;
            nodeEl.style.left = `${node.position.x}px`;
            nodeEl.style.top = `${node.position.y}px`;

            // Node icon based on type
            let icon = 'fa-circle';
            let iconColor = 'has-text-grey';

            switch (node.type) {
                case 'start':
                    icon = 'fa-play-circle';
                    iconColor = 'has-text-success';
                    break;
                case 'end':
                    icon = 'fa-stop-circle';
                    iconColor = 'has-text-danger';
                    break;
                case 'decision':
                    icon = 'fa-question-circle';
                    iconColor = 'has-text-warning';
                    break;
                case 'parallel':
                    icon = 'fa-code-branch';
                    iconColor = 'has-text-info';
                    break;
                case 'work_item': // Handle underscore version
                case 'work-item':  // Handle hyphen version
                    icon = 'fa-tasks';
                    iconColor = 'has-text-link';
                    break;
            }

            // Get display name for the node
            let displayName = node.data.name || node.type;

            // For work items, prioritize the title lookup from availableWorkItems
            if ((node.type === 'work_item' || node.type === 'work-item') && node.data.work_item_key) {
                // ArangoDB uses _key as the primary identifier
                const workItem = this.availableWorkItems.find(wi => wi._key === node.data.work_item_key || wi.key === node.data.work_item_key);

                if (workItem) {
                    // Found in availableWorkItems - use the current title
                    displayName = workItem.title;
                } else if (node.data.name && node.data.name !== 'work_item' && node.data.name !== 'work-item') {
                    // Use stored name if it exists and isn't just the type name
                    displayName = node.data.name;
                } else if (this.availableWorkItems.length === 0) {
                    // Work items haven't loaded yet
                    displayName = 'Loading...';
                } else {
                    // Work item key not found in loaded items
                    displayName = `Work Item (${node.data.work_item_key.substring(0, 8)}...)`;
                }
            }

            nodeEl.innerHTML = `
                <div class="node-content">
                    <div class="node-header">
                        <span class="node-icon ${iconColor}"><i class="fas ${icon}"></i></span>
                        <span class="node-title">${displayName}</span>
                    </div>
                </div>
            `;

            // Add to canvas
            this.$refs.canvasViewport.appendChild(nodeEl);

            // Make node draggable using jsPlumb v6 API
            this.jsPlumbInstance.setDraggable(nodeEl, true);

            // Add drag stop listener to update position
            this.jsPlumbInstance.on(nodeEl, 'stop', (params) => {
                // Update node position
                const nodeObj = this.nodes.find(n => n.id === node.id);
                if (nodeObj) {
                    nodeObj.position.x = params.pos.x;
                    nodeObj.position.y = params.pos.y;
                    this.saveToHistory();
                }
            });

            // Add source endpoint (right side)
            this.jsPlumbInstance.addEndpoint(nodeEl, {
                anchor: 'Right',
                source: true,
                maxConnections: node.type === 'decision' || node.type === 'parallel' ? -1 : 1,
                cssClass: 'endpoint source'
            });

            // Add target endpoint (left side) - except for start node
            if (node.type !== 'start') {
                this.jsPlumbInstance.addEndpoint(nodeEl, {
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

        // Render connections
        renderConnections() {
            this.edges.forEach(edge => {
                const sourceNode = document.getElementById(edge.source);
                const targetNode = document.getElementById(edge.target);

                if (sourceNode && targetNode) {
                    this.jsPlumbInstance.connect({
                        source: sourceNode,
                        target: targetNode,
                        type: edge.type || 'sequential'
                    });
                }
            });
        },

        // Select a node
        selectNode(nodeId) {
            // Deselect all
            document.querySelectorAll('.workflow-node').forEach(el => {
                el.classList.remove('selected');
            });

            // Select this node
            const nodeEl = document.getElementById(nodeId);
            if (nodeEl) {
                nodeEl.classList.add('selected');
                this.selectedNode = this.nodes.find(n => n.id === nodeId);
            }
        },

        // Update node property
        updateNodeProperty(property, value) {
            if (this.selectedNode) {
                this.selectedNode.data[property] = value;
                this.saveToHistory();

                // Update visual
                const nodeEl = document.getElementById(this.selectedNode.id);
                if (nodeEl && property === 'name') {
                    const titleEl = nodeEl.querySelector('.node-title');
                    if (titleEl) titleEl.textContent = value;
                }
            }
        },

        // Delete selected node
        deleteSelectedNode() {
            if (!this.selectedNode) return;

            // Don't allow deleting Start or End nodes
            if (this.selectedNode.type === 'start' || this.selectedNode.type === 'end') {
                alert('Cannot delete Start or End nodes');
                return;
            }

            if (confirm('Delete this node?')) {
                const nodeId = this.selectedNode.id;

                // Get the DOM element
                const nodeEl = document.getElementById(nodeId);

                if (nodeEl) {
                    // Remove all connections first
                    this.jsPlumbInstance.removeAllEndpoints(nodeEl);

                    // Remove from jsPlumb management
                    this.jsPlumbInstance.unmanage(nodeEl);

                    // Remove from DOM
                    nodeEl.remove();
                }

                // Remove from nodes array
                this.nodes = this.nodes.filter(n => n.id !== nodeId);

                // Remove any edges connected to this node
                this.edges = this.edges.filter(e =>
                    e.source !== nodeId && e.target !== nodeId
                );

                this.selectedNode = null;
                this.saveToHistory();
            }
        },

        // Connection created
        onConnectionCreated(info) {
            const edge = {
                id: `edge_${Date.now()}`,
                source: info.source.id,
                target: info.target.id,
                type: 'sequential'
            };

            this.edges.push(edge);
            this.saveToHistory();

        },

        // Connection removed
        onConnectionRemoved(info) {
            this.edges = this.edges.filter(e =>
                !(e.source === info.source.id && e.target === info.target.id)
            );
            this.saveToHistory();

        },

        // Zoom controls
        zoomIn() {
            const transform = this.panzoomInstance.getTransform();
            const newScale = transform.scale * 1.2;
            // Use zoomAbs to set absolute zoom level
            this.panzoomInstance.zoomAbs(0, 0, newScale);
        },

        zoomOut() {
            const transform = this.panzoomInstance.getTransform();
            const newScale = transform.scale * 0.8;
            // Use zoomAbs to set absolute zoom level
            this.panzoomInstance.zoomAbs(0, 0, newScale);
        },

        fitToScreen() {
            this.panzoomInstance.moveTo(0, 0);
            this.panzoomInstance.zoomTo(0, 0, 1);
        },

        toggleMinimap() {
            this.showMinimap = !this.showMinimap;
        },

        // Auto layout
        autoLayout() {
            // Simple horizontal layout
            let currentX = 50;
            const y = 100;
            const spacing = 250;

            this.nodes.forEach(node => {
                node.position.x = currentX;
                node.position.y = y;
                currentX += spacing;

                const nodeEl = document.getElementById(node.id);
                if (nodeEl) {
                    nodeEl.style.left = `${node.position.x}px`;
                    nodeEl.style.top = `${node.position.y}px`;
                }
            });

            this.jsPlumbInstance.repaintEverything();
            this.saveToHistory();
        },

        // Validate workflow
        validateWorkflow() {
            const errors = [];

            // Check for start node
            const startNodes = this.nodes.filter(n => n.type === 'start');
            if (startNodes.length === 0) {
                errors.push('Workflow must have a Start node');
            } else if (startNodes.length > 1) {
                errors.push('Workflow can only have one Start node');
            }

            // Check for end node
            const endNodes = this.nodes.filter(n => n.type === 'end');
            if (endNodes.length === 0) {
                errors.push('Workflow must have at least one End node');
            }

            // Check for orphaned nodes
            const connectedNodes = new Set();
            this.edges.forEach(edge => {
                connectedNodes.add(edge.source);
                connectedNodes.add(edge.target);
            });

            this.nodes.forEach(node => {
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

        // Save workflow
        async saveWorkflow() {
            this.saving = true;

            try {
                // First, fetch the current specification to get ALL workflows
                const specResponse = await fetch(`/api/v1/agencies/${this.agencyId}/specification`);
                if (!specResponse.ok) {
                    throw new Error('Failed to fetch current specification');
                }

                const spec = await specResponse.json();
                const allWorkflows = spec.workflows || [];

                // Build the updated workflow object
                const updatedWorkflow = {
                    _key: this.workflowKey,
                    name: this.workflowName,
                    description: this.workflowDescription,
                    agency_id: this.agencyId,
                    version: this.workflowVersion,
                    nodes: this.nodes,
                    edges: this.edges
                };

                // Find and update the specific workflow, or add if new
                const workflowIndex = allWorkflows.findIndex(wf => wf._key === this.workflowKey || wf.key === this.workflowKey);

                if (workflowIndex >= 0) {
                    allWorkflows[workflowIndex] = updatedWorkflow;
                } else {
                    allWorkflows.push(updatedWorkflow);
                }

                // Save ALL workflows back to the specification
                const response = await fetch(`/api/v1/agencies/${this.agencyId}/specification/workflows`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        workflows: allWorkflows,  // Send ALL workflows
                        updated_by: 'system'  // TODO: Get actual user
                    })
                });

                if (response.ok) {
                    const result = await response.json();
                    alert('✓ Workflow saved successfully');
                } else {
                    const error = await response.json();
                    alert('Failed to save workflow: ' + (error.error || 'Unknown error'));
                }
            } catch (error) {
                alert('Failed to save workflow: ' + error.message);
            } finally {
                this.saving = false;
            }
        },

        // Helper to find workflow index by key
        findWorkflowIndexByKey(key) {
            // This would be used if we're managing multiple workflows
            // For now, we're just saving the current workflow
            return 0;
        },

        // Execute workflow
        executeWorkflow() {
            alert('Workflow execution will be implemented in Phase 6');
        },

        // History management
        saveToHistory() {
            const state = {
                nodes: JSON.parse(JSON.stringify(this.nodes)),
                edges: JSON.parse(JSON.stringify(this.edges))
            };

            // Remove future states if we're not at the end
            this.history = this.history.slice(0, this.historyIndex + 1);

            this.history.push(state);
            this.historyIndex = this.history.length - 1;

            this.updateHistoryButtons();
        },

        updateHistoryButtons() {
            this.canUndo = this.historyIndex > 0;
            this.canRedo = this.historyIndex < this.history.length - 1;
        },

        undo() {
            if (this.canUndo) {
                this.historyIndex--;
                this.restoreState(this.history[this.historyIndex]);
            }
        },

        redo() {
            if (this.canRedo) {
                this.historyIndex++;
                this.restoreState(this.history[this.historyIndex]);
            }
        },

        restoreState(state) {
            this.nodes = JSON.parse(JSON.stringify(state.nodes));
            this.edges = JSON.parse(JSON.stringify(state.edges));

            // Clear canvas
            this.$refs.canvasViewport.innerHTML = '';
            this.jsPlumbInstance.reset();

            // Re-render
            this.renderNodes();
            this.updateHistoryButtons();
        },

        // Keyboard shortcuts
        setupKeyboardShortcuts() {
            document.addEventListener('keydown', (e) => {
                // Ctrl/Cmd + Z - Undo
                if ((e.ctrlKey || e.metaKey) && e.key === 'z' && !e.shiftKey) {
                    e.preventDefault();
                    this.undo();
                }

                // Ctrl/Cmd + Shift + Z or Ctrl/Cmd + Y - Redo
                if ((e.ctrlKey || e.metaKey) && (e.key === 'y' || (e.key === 'z' && e.shiftKey))) {
                    e.preventDefault();
                    this.redo();
                }

                // Delete - Delete selected node
                if (e.key === 'Delete' && this.selectedNode) {
                    this.deleteSelectedNode();
                }

                // Ctrl/Cmd + S - Save
                if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                    e.preventDefault();
                    this.saveWorkflow();
                }
            });

            // Deselect on canvas click
            this.$refs.canvasViewport.addEventListener('click', (e) => {
                if (e.target === this.$refs.canvasViewport) {
                    this.selectedNode = null;
                    document.querySelectorAll('.workflow-node').forEach(el => {
                        el.classList.remove('selected');
                    });
                }
            });
        }
    };
}
