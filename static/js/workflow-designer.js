// Workflow Designer - Alpine.js Component with jsPlumb
// Integrates with GitOps work items architecture

function workflowDesigner() {
    return {
        jsPlumbInstance: null,
        selectedNode: null,
        nodes: [],
        edges: [],
        workflowId: null,
        agencyId: null,
        nodeCounter: 0,

        init() {
            this.agencyId = this.$el.dataset.agencyId;
            this.workflowId = this.$el.dataset.workflowId;
            this.initializeJsPlumb();
            if (this.workflowId) {
                this.loadWorkflow();
            }
        },

        initializeJsPlumb() {
            const container = document.getElementById('workflow-canvas');

            // Initialize jsPlumb instance
            this.jsPlumbInstance = jsPlumb.newInstance({
                container: container,
                connector: ['Bezier', { curviness: 50 }],
                paintStyle: { stroke: '#3273dc', strokeWidth: 2 },
                hoverPaintStyle: { stroke: '#48c774', strokeWidth: 3 },
                endpoint: ['Dot', { radius: 5 }],
                endpointStyle: { fill: '#3273dc' },
                endpointHoverStyle: { fill: '#48c774' },
                dragOptions: { cursor: 'move', zIndex: 2000 },
                connectionsDetachable: true,
                reattachConnections: true,
            });

            // Connection event handler
            this.jsPlumbInstance.bind('connection', (info) => {
                this.onConnectionCreated(info);
            });

            // Connection detach handler
            this.jsPlumbInstance.bind('connectionDetached', (info) => {
                this.onConnectionDetached(info);
            });

            // Add minimap
            const minimap = new XYFlow.MiniMap({
                nodeColor: (node) => {
                    const colors = {
                        document: '#3298dc',
                        software: '#48c774',
                        proposal: '#ffdd57',
                        analysis: '#f14668'
                    };
                    return colors[node.data.type] || '#ccc';
                }
            });
            container.appendChild(minimap);
        },

        createWorkItemNode() {
            const self = this;
            return {
                render: (node) => {
                    const typeIcons = {
                        document: 'fa-file-alt',
                        software: 'fa-code',
                        proposal: 'fa-file-contract',
                        analysis: 'fa-chart-line'
                    };

                    const typeColors = {
                        document: 'has-background-info-light',
                        software: 'has-background-success-light',
                        proposal: 'has-background-warning-light',
                        analysis: 'has-background-danger-light'
                    };

                    const typeBorderColors = {
                        document: 'has-border-info',
                        software: 'has-border-success',
                        proposal: 'has-border-warning',
                        analysis: 'has-border-danger'
                    };

                    const autoMergeIcon = node.data.gitea_config?.auto_merge
                        ? '<i class="fas fa-bolt has-text-success" title="Auto-merge enabled"></i>'
                        : '<i class="fas fa-hand-paper has-text-grey" title="Manual merge required"></i>';

                    return `
            <div class="work-item-node box p-3 ${typeColors[node.data.type]} ${typeBorderColors[node.data.type]}" style="min-width: 200px; border-left: 4px solid;">
              <div class="is-flex is-justify-content-between is-align-items-center mb-2">
                <div class="is-flex is-align-items-center">
                  <span class="icon mr-2">
                    <i class="fas ${typeIcons[node.data.type]}"></i>
                  </span>
                  <strong class="is-size-6">${node.data.title || 'Untitled'}</strong>
                </div>
                ${autoMergeIcon}
              </div>
              <p class="is-size-7 has-text-grey mb-2" style="max-width: 180px; overflow: hidden; text-overflow: ellipsis;">
                ${node.data.description?.substring(0, 60) || 'No description'}${node.data.description?.length > 60 ? '...' : ''}
              </p>
              <div class="tags are-small">
                ${node.data.labels?.slice(0, 3).map(l => `<span class="tag">${l}</span>`).join('') || ''}
                ${node.data.labels?.length > 3 ? `<span class="tag">+${node.data.labels.length - 3}</span>` : ''}
              </div>
            </div>
          `;
                }
            };
        },

        addNode(type) {
            const nodeId = `node-${Date.now()}`;
            const position = this.getNextNodePosition();

            const newNode = {
                id: nodeId,
                type: 'workItem',
                position: position,
                data: {
                    type: type,
                    title: `New ${type} work item`,
                    description: '',
                    labels: ['work-item', type],
                    labelsText: `work-item, ${type}`,
                    gitea_config: {
                        repo: 'main-repo',
                        branch_pattern: 'issue-{issue_id}-{slug}',
                        merge_strategy: type === 'document' ? 'squash' : 'merge',
                        auto_merge: type === 'document',
                        require_reviews: type === 'document' ? 0 : 1
                    }
                }
            };

            this.nodes.push(newNode);
            this.flowInstance.setNodes([...this.nodes]);
            this.selectedNode = newNode;
        },

        getNextNodePosition() {
            if (this.nodes.length === 0) {
                return { x: 250, y: 100 };
            }

            // Place new node to the right of the last node
            const lastNode = this.nodes[this.nodes.length - 1];
            return {
                x: lastNode.position.x + 300,
                y: lastNode.position.y
            };
        },

        addEdge(params) {
            // Check if edge already exists
            const edgeExists = this.edges.some(e =>
                e.source === params.source && e.target === params.target
            );

            if (edgeExists) {
                this.showNotification('Connection already exists', 'warning');
                return;
            }

            // Check for circular dependency
            if (this.wouldCreateCycle(params.source, params.target)) {
                this.showNotification('Cannot create circular dependency', 'danger');
                return;
            }

            const newEdge = {
                id: `edge-${params.source}-${params.target}`,
                source: params.source,
                target: params.target,
                type: 'smoothstep',
                animated: true,
                label: 'depends on',
                markerEnd: {
                    type: 'arrowclosed',
                }
            };

            this.edges.push(newEdge);
            this.flowInstance.setEdges([...this.edges]);
        },

        wouldCreateCycle(source, target) {
            // Check if adding edge from source to target would create a cycle
            const visited = new Set();
            const recStack = new Set();

            const dfs = (nodeId) => {
                visited.add(nodeId);
                recStack.add(nodeId);

                // Get all outgoing edges from this node
                const outgoing = this.edges.filter(e => e.source === nodeId);

                // Add the potential new edge
                if (nodeId === source) {
                    outgoing.push({ source, target });
                }

                for (const edge of outgoing) {
                    if (!visited.has(edge.target)) {
                        if (dfs(edge.target)) return true;
                    } else if (recStack.has(edge.target)) {
                        return true;
                    }
                }

                recStack.delete(nodeId);
                return false;
            };

            return dfs(source);
        },

        updateNode() {
            if (this.selectedNode) {
                // Parse labels text
                this.selectedNode.data.labels = this.selectedNode.data.labelsText
                    .split(',')
                    .map(l => l.trim())
                    .filter(l => l);

                // Update flow
                this.flowInstance.setNodes([...this.nodes]);
            }
        },

        deleteNode() {
            if (!this.selectedNode) return;

            if (confirm(`Delete node "${this.selectedNode.data.title}"?`)) {
                // Remove node
                this.nodes = this.nodes.filter(n => n.id !== this.selectedNode.id);

                // Remove connected edges
                this.edges = this.edges.filter(e =>
                    e.source !== this.selectedNode.id && e.target !== this.selectedNode.id
                );

                this.flowInstance.setNodes([...this.nodes]);
                this.flowInstance.setEdges([...this.edges]);
                this.selectedNode = null;
            }
        },

        async saveWorkflow() {
            if (!this.validateWorkflow()) {
                return;
            }

            const workflow = {
                id: this.workflowId,
                agency_id: this.agencyId,
                name: document.getElementById('workflow-name')?.value || 'Untitled Workflow',
                description: document.getElementById('workflow-description')?.value || '',
                nodes: this.nodes.map(n => ({
                    id: n.id,
                    type: n.data.type,
                    title: n.data.title,
                    description: n.data.description,
                    labels: n.data.labels,
                    position: n.position,
                    gitea_config: n.data.gitea_config
                })),
                edges: this.edges.map(e => ({
                    id: e.id,
                    source: e.source,
                    target: e.target
                }))
            };

            try {
                const url = this.workflowId
                    ? `/api/v1/agencies/${this.agencyId}/workflows/${this.workflowId}`
                    : `/api/v1/agencies/${this.agencyId}/workflows`;

                const method = this.workflowId ? 'PUT' : 'POST';

                const response = await fetch(url, {
                    method: method,
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(workflow)
                });

                if (response.ok) {
                    const data = await response.json();
                    this.workflowId = data._key;
                    this.showNotification('Workflow saved successfully', 'success');
                } else {
                    const error = await response.json();
                    throw new Error(error.error || 'Save failed');
                }
            } catch (error) {
                this.showNotification(`Failed to save: ${error.message}`, 'danger');
            }
        },

        validateWorkflow() {
            // Check if workflow has nodes
            if (this.nodes.length === 0) {
                this.showNotification('Workflow must have at least one node', 'warning');
                return false;
            }

            // Check for cycles
            if (this.detectCycle()) {
                this.showNotification('Workflow has circular dependencies', 'danger');
                return false;
            }

            // Check required fields
            const invalidNodes = this.nodes.filter(n =>
                !n.data.title || !n.data.description || !n.data.gitea_config.repo
            );

            if (invalidNodes.length > 0) {
                this.showNotification(`${invalidNodes.length} node(s) missing required fields`, 'danger');
                return false;
            }

            return true;
        },

        detectCycle() {
            const visited = new Set();
            const recStack = new Set();

            const hasCycleUtil = (nodeId) => {
                visited.add(nodeId);
                recStack.add(nodeId);

                const outgoingEdges = this.edges.filter(e => e.source === nodeId);
                for (const edge of outgoingEdges) {
                    if (!visited.has(edge.target)) {
                        if (hasCycleUtil(edge.target)) return true;
                    } else if (recStack.has(edge.target)) {
                        return true;
                    }
                }

                recStack.delete(nodeId);
                return false;
            };

            for (const node of this.nodes) {
                if (!visited.has(node.id)) {
                    if (hasCycleUtil(node.id)) return true;
                }
            }

            return false;
        },

        async exportWorkflow() {
            const workflow = {
                version: '1.0',
                name: document.getElementById('workflow-name')?.value || 'Untitled Workflow',
                description: document.getElementById('workflow-description')?.value || '',
                nodes: this.nodes,
                edges: this.edges,
                metadata: {
                    created: new Date().toISOString(),
                    generator: 'CodeValdCortex Workflow Designer',
                    agency_id: this.agencyId
                }
            };

            const blob = new Blob([JSON.stringify(workflow, null, 2)], {
                type: 'application/json'
            });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `workflow-${this.workflowId || 'new'}.json`;
            a.click();
            URL.revokeObjectURL(url);

            this.showNotification('Workflow exported', 'success');
        },

        async loadWorkflow() {
            if (!this.workflowId) return;

            try {
                const response = await fetch(`/api/v1/agencies/${this.agencyId}/workflows/${this.workflowId}`);

                if (!response.ok) {
                    throw new Error('Failed to load workflow');
                }

                const workflow = await response.json();

                // Set workflow metadata
                if (document.getElementById('workflow-name')) {
                    document.getElementById('workflow-name').value = workflow.name || '';
                }
                if (document.getElementById('workflow-description')) {
                    document.getElementById('workflow-description').value = workflow.description || '';
                }

                // Load nodes
                this.nodes = workflow.nodes.map(n => ({
                    ...n,
                    type: 'workItem',
                    data: {
                        ...n,
                        labelsText: n.labels?.join(', ') || ''
                    }
                }));

                // Load edges
                this.edges = workflow.edges;

                this.flowInstance.setNodes(this.nodes);
                this.flowInstance.setEdges(this.edges);
                this.flowInstance.fitView();

            } catch (error) {
                this.showNotification(`Failed to load workflow: ${error.message}`, 'danger');
            }
        },

        async executeWorkflow() {
            if (!this.workflowId) {
                this.showNotification('Save workflow first', 'warning');
                return;
            }

            if (!confirm('Execute this workflow? This will create Gitea issues for all nodes.')) {
                return;
            }

            try {
                const response = await fetch(
                    `/api/v1/agencies/${this.agencyId}/workflows/${this.workflowId}/execute`,
                    { method: 'POST' }
                );

                if (response.ok) {
                    this.showNotification('Workflow execution started', 'success');
                } else {
                    const error = await response.json();
                    throw new Error(error.error || 'Execution failed');
                }
            } catch (error) {
                this.showNotification(`Execution failed: ${error.message}`, 'danger');
            }
        },

        showNotification(message, type) {
            const container = document.getElementById('notification-container') || document.body;
            const notification = document.createElement('div');
            notification.className = `notification is-${type} is-light`;
            notification.innerHTML = `
        <button class="delete"></button>
        ${message}
      `;

            container.appendChild(notification);

            // Delete button handler
            notification.querySelector('.delete').addEventListener('click', () => {
                notification.remove();
            });

            // Auto-remove after 5 seconds
            setTimeout(() => {
                if (notification.parentElement) {
                    notification.remove();
                }
            }, 5000);
        }
    };
}
