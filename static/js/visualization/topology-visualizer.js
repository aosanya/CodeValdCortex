/**
 * Network Topology Visualizer - INFRA-017
 * 
 * High-performance network topology visualization using Deck.gl + d3-hierarchy
 * 
 * Features:
 * - WebGL-accelerated rendering for large networks (1000+ nodes)
 * - Real-time updates with JSON Patch support
 * - Multiple layout algorithms (force-directed, hierarchical, geographic)
 * - Interactive pan, zoom, and selection
 * - Agent state visualization with color coding
 * - Edge animations for message flows
 * 
 * Dependencies:
 * - deck.gl (WebGL rendering)
 * - d3-hierarchy (layout algorithms)
 * - d3 (data utilities)
 */

class TopologyVisualizer {
    constructor(containerId, config = {}) {
        this.containerId = containerId;
        this.config = {
            width: config.width || 800,
            height: config.height || 600,
            initialViewState: config.initialViewState || {
                longitude: 0,
                latitude: 0,
                zoom: 1,
                pitch: 0,
                bearing: 0
            },
            layoutAlgorithm: config.layoutAlgorithm || 'force', // 'force', 'hierarchical', 'geographic', 'grid'
            enableInteraction: config.enableInteraction !== false,
            updateInterval: config.updateInterval || 1000,
            animateTransitions: config.animateTransitions !== false,
            ...config
        };

        // Initialize state
        this.nodes = [];
        this.edges = [];
        this.selectedNode = null;
        this.hoveredNode = null;

        // Deck.gl instance
        this.deck = null;

        // Layout engine
        this.layoutEngine = new LayoutEngine(this.config.layoutAlgorithm);

        // Initialize
        this.init();
    }

    /**
     * Initialize the visualizer
     */
    init() {
        const container = document.getElementById(this.containerId);
        if (!container) {
            console.error(`Container #${this.containerId} not found`);
            return;
        }

        // Create Deck.gl instance with MapLibre
        this.deck = new deck.DeckGL({
            container: container,
            width: this.config.width,
            height: this.config.height,
            initialViewState: this.config.initialViewState,
            controller: this.config.enableInteraction,
            layers: [],
            onHover: this.onHover.bind(this),
            onClick: this.onClick.bind(this),
            getCursor: ({ isDragging, isHovering }) =>
                isDragging ? 'grabbing' : isHovering ? 'pointer' : 'grab',
            // Use MapLibre with detailed OpenStreetMap style (roads, buildings, labels)
            mapLib: maplibregl,
            mapStyle: 'https://tiles.openfreemap.org/styles/liberty'
        });

        console.log('Topology Visualizer initialized with MapLibre + OpenStreetMap');
    }

    /**
     * Load data and render
     */
    async loadData(agents, edges) {
        // Transform agents to nodes
        this.nodes = agents.map(agent => {
            const metadata = agent.metadata || {};

            // Extract coordinates from metadata (latitude/longitude as strings)
            let coordinates = null;
            let position = null;

            if (metadata.latitude && metadata.longitude) {
                const lat = parseFloat(metadata.latitude);
                const lng = parseFloat(metadata.longitude);
                if (!isNaN(lat) && !isNaN(lng)) {
                    coordinates = { latitude: lat, longitude: lng };
                    position = [lng, lat]; // [longitude, latitude] for map rendering
                }
            } else if (agent.coordinates) {
                coordinates = agent.coordinates;
                position = [agent.coordinates.longitude, agent.coordinates.latitude];
            }

            return {
                id: agent.id,
                name: agent.name || agent.id,
                type: agent.agent_type || agent.type,
                status: agent.status || 'unknown',
                metadata: metadata,
                coordinates: coordinates,
                position: position
            };
        });

        // Transform edges
        this.edges = edges.map(edge => ({
            source: edge.source || edge.from,
            target: edge.target || edge.to,
            type: edge.type || 'connection',
            metadata: edge.metadata || {}
        }));

        // Compute layout
        this.computeLayout();

        // Render
        this.render();
    }

    /**
     * Compute node positions using layout algorithm
     */
    computeLayout() {
        if (this.config.layoutAlgorithm === 'geographic') {
            // Use coordinates if available
            this.nodes.forEach(node => {
                if (!node.position && node.coordinates) {
                    node.position = [node.coordinates.longitude, node.coordinates.latitude];
                }
            });
        } else if (this.config.layoutAlgorithm === 'hierarchical') {
            // Use d3-hierarchy for tree layout
            const positions = this.layoutEngine.computeHierarchical(this.nodes, this.edges);
            this.nodes.forEach((node, i) => {
                node.position = positions[node.id] || [0, 0];
            });
        } else if (this.config.layoutAlgorithm === 'force') {
            // Use force-directed layout
            const positions = this.layoutEngine.computeForceDirected(this.nodes, this.edges);
            this.nodes.forEach(node => {
                node.position = positions[node.id] || [0, 0];
            });
        } else if (this.config.layoutAlgorithm === 'grid') {
            // Simple grid layout
            const positions = this.layoutEngine.computeGrid(this.nodes);
            this.nodes.forEach(node => {
                node.position = positions[node.id] || [0, 0];
            });
        }

        // Ensure all nodes have positions
        this.nodes.forEach((node, i) => {
            if (!node.position) {
                node.position = [i * 0.1, 0];
            }
        });
    }

    /**
     * Render layers
     */
    render() {
        const layers = [
            // Edge layer
            new deck.LineLayer({
                id: 'edges',
                data: this.edges,
                getSourcePosition: d => {
                    const source = this.nodes.find(n => n.id === d.source);
                    return source ? source.position : [0, 0];
                },
                getTargetPosition: d => {
                    const target = this.nodes.find(n => n.id === d.target);
                    return target ? target.position : [0, 0];
                },
                getColor: d => this.getEdgeColor(d),
                getWidth: 2,
                opacity: 0.6,
                pickable: false
            }),

            // Node layer
            new deck.ScatterplotLayer({
                id: 'nodes',
                data: this.nodes,
                getPosition: d => d.position,
                getRadius: d => this.getNodeRadius(d),
                getFillColor: d => this.getNodeColor(d),
                getLineColor: [255, 255, 255],
                lineWidthMinPixels: 2,
                radiusMinPixels: 5,
                radiusMaxPixels: 30,
                pickable: true,
                opacity: 0.8,
                stroked: true,
                updateTriggers: {
                    getFillColor: [this.selectedNode, this.hoveredNode],
                    getRadius: [this.selectedNode, this.hoveredNode]
                }
            }),

            // Label layer (for zoomed in view)
            new deck.TextLayer({
                id: 'labels',
                data: this.nodes.filter(n =>
                    this.selectedNode === n.id ||
                    this.hoveredNode === n.id
                ),
                getPosition: d => d.position,
                getText: d => d.name,
                getColor: [255, 255, 255],
                getSize: 12,
                getAngle: 0,
                getTextAnchor: 'middle',
                getAlignmentBaseline: 'bottom',
                getPixelOffset: [0, -20],
                pickable: false
            })
        ];

        this.deck.setProps({ layers });
    }

    /**
     * Get node color based on status
     */
    getNodeColor(node) {
        if (this.selectedNode === node.id) {
            return [255, 215, 0, 255]; // Gold for selected
        }
        if (this.hoveredNode === node.id) {
            return [255, 255, 255, 255]; // White for hovered
        }

        // Status-based colors
        const statusColors = {
            active: [76, 175, 80, 255],      // Green
            idle: [158, 158, 158, 255],      // Gray
            error: [244, 67, 54, 255],       // Red
            warning: [255, 152, 0, 255],     // Orange
            offline: [97, 97, 97, 255],      // Dark Gray
            unknown: [33, 150, 243, 255]     // Blue
        };

        return statusColors[node.status] || statusColors.unknown;
    }

    /**
     * Get node radius
     */
    getNodeRadius(node) {
        const baseRadius = 100;
        if (this.selectedNode === node.id) {
            return baseRadius * 1.5;
        }
        if (this.hoveredNode === node.id) {
            return baseRadius * 1.3;
        }
        return baseRadius;
    }

    /**
     * Get edge color
     */
    getEdgeColor(edge) {
        const typeColors = {
            connection: [158, 158, 158, 100],  // Gray
            message: [33, 150, 243, 150],      // Blue
            data: [76, 175, 80, 150],          // Green
            control: [255, 152, 0, 150]        // Orange
        };

        return typeColors[edge.type] || typeColors.connection;
    }

    /**
     * Handle hover events
     */
    onHover(info) {
        if (info.object && info.object.id) {
            this.hoveredNode = info.object.id;
            this.render();

            // Show tooltip
            this.showTooltip(info);
        } else if (this.hoveredNode) {
            this.hoveredNode = null;
            this.render();
            this.hideTooltip();
        }
    }

    /**
     * Handle click events
     */
    onClick(info) {
        if (info.object && info.object.id) {
            this.selectedNode = info.object.id;
            this.render();

            // Emit event
            this.emit('nodeSelected', info.object);
        } else {
            this.selectedNode = null;
            this.render();
        }
    }

    /**
     * Show tooltip
     */
    showTooltip(info) {
        const tooltip = document.getElementById('topology-tooltip') || this.createTooltip();
        const node = info.object;

        tooltip.innerHTML = `
            <div class="tooltip-title">${node.name}</div>
            <div class="tooltip-content">
                <div><strong>Type:</strong> ${node.type}</div>
                <div><strong>Status:</strong> ${node.status}</div>
                <div><strong>ID:</strong> ${node.id}</div>
            </div>
        `;

        tooltip.style.left = `${info.x}px`;
        tooltip.style.top = `${info.y}px`;
        tooltip.style.display = 'block';
    }

    /**
     * Hide tooltip
     */
    hideTooltip() {
        const tooltip = document.getElementById('topology-tooltip');
        if (tooltip) {
            tooltip.style.display = 'none';
        }
    }

    /**
     * Create tooltip element
     */
    createTooltip() {
        const tooltip = document.createElement('div');
        tooltip.id = 'topology-tooltip';
        tooltip.className = 'topology-tooltip';
        tooltip.style.cssText = `
            position: absolute;
            display: none;
            background: rgba(0, 0, 0, 0.9);
            color: white;
            padding: 10px;
            border-radius: 4px;
            font-size: 12px;
            pointer-events: none;
            z-index: 1000;
        `;
        document.body.appendChild(tooltip);
        return tooltip;
    }

    /**
     * Update visualization with new data
     */
    update(agents, edges) {
        this.loadData(agents, edges);
    }

    /**
     * Apply JSON Patch to update specific nodes/edges
     */
    applyPatch(patches) {
        // Implementation for incremental updates
        patches.forEach(patch => {
            if (patch.op === 'replace' && patch.path.startsWith('/agents/')) {
                const nodeId = patch.path.split('/')[2];
                const node = this.nodes.find(n => n.id === nodeId);
                if (node) {
                    Object.assign(node, patch.value);
                }
            }
        });

        this.render();
    }

    /**
     * Event emitter
     */
    emit(event, data) {
        const customEvent = new CustomEvent(`topology:${event}`, { detail: data });
        document.dispatchEvent(customEvent);
    }

    /**
     * Cleanup
     */
    destroy() {
        if (this.deck) {
            this.deck.finalize();
            this.deck = null;
        }

        const tooltip = document.getElementById('topology-tooltip');
        if (tooltip) {
            tooltip.remove();
        }
    }
}

/**
 * Layout Engine - Computes node positions using various algorithms
 */
class LayoutEngine {
    constructor(algorithm = 'force') {
        this.algorithm = algorithm;
    }

    /**
     * Force-directed layout using simple physics simulation
     */
    computeForceDirected(nodes, edges) {
        const positions = {};
        const iterations = 100;
        const spacing = 1.0;

        // Initialize random positions
        nodes.forEach((node, i) => {
            const angle = (i / nodes.length) * Math.PI * 2;
            const radius = 5;
            positions[node.id] = {
                x: Math.cos(angle) * radius,
                y: Math.sin(angle) * radius,
                vx: 0,
                vy: 0
            };
        });

        // Build adjacency list
        const adjacency = new Map();
        edges.forEach(edge => {
            if (!adjacency.has(edge.source)) adjacency.set(edge.source, []);
            if (!adjacency.has(edge.target)) adjacency.set(edge.target, []);
            adjacency.get(edge.source).push(edge.target);
            adjacency.get(edge.target).push(edge.source);
        });

        // Simulate
        for (let iter = 0; iter < iterations; iter++) {
            const alpha = 0.1 * (1 - iter / iterations);

            // Repulsion between all nodes
            Object.keys(positions).forEach(nodeId1 => {
                const pos1 = positions[nodeId1];
                Object.keys(positions).forEach(nodeId2 => {
                    if (nodeId1 === nodeId2) return;

                    const pos2 = positions[nodeId2];
                    const dx = pos1.x - pos2.x;
                    const dy = pos1.y - pos2.y;
                    const dist = Math.sqrt(dx * dx + dy * dy) || 0.01;
                    const force = spacing / (dist * dist);

                    pos1.vx += (dx / dist) * force;
                    pos1.vy += (dy / dist) * force;
                });
            });

            // Attraction along edges
            edges.forEach(edge => {
                const pos1 = positions[edge.source];
                const pos2 = positions[edge.target];
                if (!pos1 || !pos2) return;

                const dx = pos2.x - pos1.x;
                const dy = pos2.y - pos1.y;
                const dist = Math.sqrt(dx * dx + dy * dy) || 0.01;
                const force = dist * 0.01;

                pos1.vx += (dx / dist) * force;
                pos1.vy += (dy / dist) * force;
                pos2.vx -= (dx / dist) * force;
                pos2.vy -= (dy / dist) * force;
            });

            // Update positions
            Object.values(positions).forEach(pos => {
                pos.x += pos.vx * alpha;
                pos.y += pos.vy * alpha;
                pos.vx *= 0.8;
                pos.vy *= 0.8;
            });
        }

        // Convert to [lng, lat] format
        const result = {};
        Object.keys(positions).forEach(nodeId => {
            result[nodeId] = [positions[nodeId].x, positions[nodeId].y];
        });

        return result;
    }

    /**
     * Hierarchical layout using d3-hierarchy
     */
    computeHierarchical(nodes, edges) {
        const positions = {};

        // Build tree structure
        const nodeMap = new Map(nodes.map(n => [n.id, { ...n, children: [] }]));
        const roots = [];

        // Find root nodes (no incoming edges)
        const hasIncoming = new Set();
        edges.forEach(edge => hasIncoming.add(edge.target));

        nodes.forEach(node => {
            if (!hasIncoming.has(node.id)) {
                roots.push(nodeMap.get(node.id));
            }
        });

        // Build parent-child relationships
        edges.forEach(edge => {
            const parent = nodeMap.get(edge.source);
            const child = nodeMap.get(edge.target);
            if (parent && child) {
                parent.children = parent.children || [];
                parent.children.push(child);
            }
        });

        // Use d3.tree layout
        if (roots.length > 0) {
            const width = 20;
            const height = 20;
            const treeLayout = d3.tree().size([width, height]);

            roots.forEach((root, rootIndex) => {
                const hierarchy = d3.hierarchy(root);
                const tree = treeLayout(hierarchy);

                tree.descendants().forEach(node => {
                    positions[node.data.id] = [
                        node.x + rootIndex * width,
                        node.y
                    ];
                });
            });
        } else {
            // Fallback to grid if no hierarchy
            return this.computeGrid(nodes);
        }

        return positions;
    }

    /**
     * Grid layout
     */
    computeGrid(nodes) {
        const positions = {};
        const cols = Math.ceil(Math.sqrt(nodes.length));
        const spacing = 2;

        nodes.forEach((node, i) => {
            const row = Math.floor(i / cols);
            const col = i % cols;
            positions[node.id] = [col * spacing, row * spacing];
        });

        return positions;
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { TopologyVisualizer, LayoutEngine };
}
