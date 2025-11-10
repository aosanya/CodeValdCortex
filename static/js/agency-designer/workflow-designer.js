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
            console.log('Initializing workflow designer...');

            // Load workflow data from global variable
            if (typeof workflowData !== 'undefined') {
                this.workflowId = workflowData.id;
                this.agencyId = workflowData.agencyId;
                this.nodes = workflowData.nodes || [];
                this.edges = workflowData.edges || [];
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

                console.log('jsPlumb initialized');
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

            console.log('Pan/zoom initialized');
        },

        // Load available work items from API
        async loadWorkItems() {
            try {
                const workItems = await window.specificationAPI.getWorkItems();
                this.availableWorkItems = workItems || [];
                console.log('Loaded work items:', this.availableWorkItems.length);
            } catch (error) {
                console.error('Failed to load work items:', error);
            }
        },

        // Toolbox drag start
        onToolboxDragStart(event) {
            const nodeType = event.target.closest('.toolbox-item').dataset.nodeType;
            const workItemKey = event.target.closest('.toolbox-item').dataset.workItemKey;

            event.dataTransfer.effectAllowed = 'copy';
            event.dataTransfer.setData('nodeType', nodeType);
            if (workItemKey) {
                event.dataTransfer.setData('workItemKey', workItemKey);
            }
        },

        // Canvas drop handler
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
                const workItem = this.availableWorkItems.find(wi => wi.key === workItemKey);
                if (workItem) {
                    nodeName = workItem.name;
                    nodeData = {
                        name: workItem.name,
                        description: workItem.description || '',
                        workItemKey: workItemKey,
                        type: workItem.type
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

            console.log('Created node:', node);
        },

        // Render all nodes
        renderNodes() {
            this.nodes.forEach(node => this.renderNode(node));

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
                case 'work-item':
                    icon = 'fa-tasks';
                    iconColor = 'has-text-link';
                    break;
            }

            nodeEl.innerHTML = `
                <div class="node-content">
                    <div class="node-header">
                        <span class="node-icon ${iconColor}"><i class="fas ${icon}"></i></span>
                        <span class="node-title">${node.data.name || node.type}</span>
                    </div>
                    ${node.data.description ? `<div class="node-description">${node.data.description}</div>` : ''}
                </div>
            `;

            // Add to canvas
            this.$refs.canvasViewport.appendChild(nodeEl);

            // Make node draggable
            this.jsPlumbInstance.draggable(nodeEl, {
                containment: true,
                grid: [10, 10],
                stop: (params) => {
                    // Update node position
                    const nodeObj = this.nodes.find(n => n.id === node.id);
                    if (nodeObj) {
                        nodeObj.position.x = params.pos[0];
                        nodeObj.position.y = params.pos[1];
                        this.saveToHistory();
                    }
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

            if (confirm('Delete this node?')) {
                // Remove connections
                this.jsPlumbInstance.deleteConnectionsForElement(this.selectedNode.id);

                // Remove from nodes array
                this.nodes = this.nodes.filter(n => n.id !== this.selectedNode.id);

                // Remove from DOM
                const nodeEl = document.getElementById(this.selectedNode.id);
                if (nodeEl) nodeEl.remove();

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

            console.log('Connection created:', edge);
        },

        // Connection removed
        onConnectionRemoved(info) {
            this.edges = this.edges.filter(e =>
                !(e.source === info.source.id && e.target === info.target.id)
            );
            this.saveToHistory();

            console.log('Connection removed');
        },

        // Zoom controls
        zoomIn() {
            this.panzoomInstance.zoomIn();
        },

        zoomOut() {
            this.panzoomInstance.zoomOut();
        },

        fitToScreen() {
            this.panzoomInstance.moveTo(0, 0);
            this.panzoomInstance.zoomAbs(0, 0, 1);
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
                const workflowData = {
                    nodes: this.nodes,
                    edges: this.edges
                };

                const response = await fetch(`/api/v1/workflows/${this.workflowId}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        agency_id: this.agencyId,
                        nodes: this.nodes,
                        edges: this.edges
                    })
                });

                if (response.ok) {
                    alert('✓ Workflow saved successfully');
                } else {
                    const error = await response.json();
                    alert('Failed to save workflow: ' + (error.error || 'Unknown error'));
                }
            } catch (error) {
                console.error('Save error:', error);
                alert('Failed to save workflow');
            } finally {
                this.saving = false;
            }
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
