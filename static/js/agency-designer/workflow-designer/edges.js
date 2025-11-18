/**
 * Workflow Designer - Edge Management Module
 * Handles edge creation, rendering, splitting, and connection management
 */

(function (window) {
    'use strict';

    if (!window.WorkflowDesigner) {
        window.WorkflowDesigner = {};
    }

    /**
     * Creates edge management functions
     * @param {Object} context - The Alpine.js component context (this)
     * @param {Object} state - State management object
     * @param {Object} utils - Utility functions
     * @returns {Object} Edge management methods
     */
    window.WorkflowDesigner.createEdges = function (context, state, utils) {
        return {
            /**
             * Render all connections on the canvas
             */
            renderConnections() {
                state.edges.forEach(edge => {
                    const sourceNode = document.getElementById(edge.source);
                    const targetNode = document.getElementById(edge.target);

                    if (sourceNode && targetNode) {
                        state.jsPlumbInstance.connect({
                            source: sourceNode,
                            target: targetNode,
                            type: edge.type || 'sequential',
                            data: { edgeId: edge.id }
                        });
                    }
                });
            },

            /**
             * Find nearby edge using point-based detection
             * @param {number} x - X coordinate
             * @param {number} y - Y coordinate
             * @param {number} threshold - Distance threshold (default 120)
             * @returns {Object|null} Edge object or null
             */
            findNearbyEdge(x, y, threshold = 120) {
                for (const edge of state.edges) {
                    const sourceNode = state.nodes.find(n => n.id === edge.source);
                    const targetNode = state.nodes.find(n => n.id === edge.target);

                    if (!sourceNode || !targetNode) {
                        continue;
                    }

                    // Get node centers
                    const sourceX = sourceNode.position.x + 75;
                    const sourceY = sourceNode.position.y + 35;
                    const targetX = targetNode.position.x + 75;
                    const targetY = targetNode.position.y + 35;

                    // Calculate distance from point to line segment
                    const distance = utils.pointToLineDistance(
                        x, y,
                        sourceX, sourceY,
                        targetX, targetY
                    );

                    if (distance < threshold) {
                        return edge;
                    }
                }

                return null;
            },

            /**
             * Find nearby edge using box intersection detection
             * @param {number} x - X coordinate (top-left)
             * @param {number} y - Y coordinate (top-left)
             * @param {number} nodeWidth - Width of node box (default 150)
             * @param {number} nodeHeight - Height of node box (default 70)
             * @returns {Object|null} Edge object or null
             */
            findNearbyEdgeWithBox(x, y, nodeWidth = 150, nodeHeight = 70) {
                const boxLeft = x;
                const boxRight = x + nodeWidth;
                const boxTop = y;
                const boxBottom = y + nodeHeight;

                for (const edge of state.edges) {
                    const sourceNode = state.nodes.find(n => n.id === edge.source);
                    const targetNode = state.nodes.find(n => n.id === edge.target);

                    if (!sourceNode || !targetNode) continue;

                    // Get node centers
                    const sourceX = sourceNode.position.x + 75;
                    const sourceY = sourceNode.position.y + 35;
                    const targetX = targetNode.position.x + 75;
                    const targetY = targetNode.position.y + 35;

                    // Check if line segment intersects with the node box
                    const intersects = utils.lineIntersectsBox(
                        sourceX, sourceY, targetX, targetY,
                        boxLeft, boxTop, boxRight, boxBottom
                    );

                    if (intersects) {
                        return edge;
                    }
                }

                return null;
            },

            /**
             * Split an edge by inserting a node
             * @param {Object} edge - Edge to split
             * @param {string} newNodeId - ID of node to insert
             */
            splitEdge(edge, newNodeId) {
                // Remove the original edge from edges array
                state.edges = state.edges.filter(e => e.id !== edge.id);

                // Get the connection object from jsPlumb and remove it
                const connections = state.jsPlumbInstance.getConnections({
                    source: edge.source,
                    target: edge.target
                });

                connections.forEach(conn => {
                    state.jsPlumbInstance.deleteConnection(conn);
                });

                // Create two new edges: source -> newNode and newNode -> target
                context.$nextTick(() => {
                    const sourceEl = document.getElementById(edge.source);
                    const targetEl = document.getElementById(edge.target);
                    const newNodeEl = document.getElementById(newNodeId);

                    if (sourceEl && targetEl && newNodeEl) {
                        // Connect source to new node
                        const edgeId1 = `edge_${Date.now()}`;
                        state.jsPlumbInstance.connect({
                            source: sourceEl,
                            target: newNodeEl,
                            type: edge.type || 'sequential',
                            data: { edgeId: edgeId1 }
                        });

                        state.edges.push({
                            id: edgeId1,
                            source: edge.source,
                            target: newNodeId,
                            type: edge.type || 'sequential'
                        });

                        // Connect new node to target
                        const edgeId2 = `edge_${Date.now() + 1}`;
                        state.jsPlumbInstance.connect({
                            source: newNodeEl,
                            target: targetEl,
                            type: edge.type || 'sequential',
                            data: { edgeId: edgeId2 }
                        });

                        state.edges.push({
                            id: edgeId2,
                            source: newNodeId,
                            target: edge.target,
                            type: edge.type || 'sequential'
                        });

                        // Mark node as auto-connected for tracking
                        const newNode = state.nodes.find(n => n.id === newNodeId);
                        if (newNode) {
                            newNode.data.autoConnected = {
                                source: edge.source,
                                target: edge.target
                            };
                        }
                    }
                });
            },

            /**
             * Highlight an edge (visual feedback during drag)
             * @param {string} edgeId - ID of edge to highlight
             */
            highlightEdge(edgeId) {
                const connections = state.jsPlumbInstance.getConnections();

                connections.forEach(conn => {
                    const connEdgeId = conn.getData()?.edgeId;

                    if (connEdgeId === edgeId) {
                        // Highlight this edge
                        conn.setPaintStyle({
                            stroke: '#ffdd57',  // Yellow highlight
                            strokeWidth: 4
                        });
                    } else {
                        // Reset other edges to normal
                        conn.setPaintStyle({
                            stroke: '#3273dc',
                            strokeWidth: 2
                        });
                    }
                });
            },

            /**
             * Clear all edge highlights
             */
            clearEdgeHighlights() {
                const connections = state.jsPlumbInstance.getConnections();

                connections.forEach(conn => {
                    conn.setPaintStyle({
                        stroke: '#3273dc',
                        strokeWidth: 2
                    });
                });
            },

            /**
             * Connection created event handler
             * @param {Object} info - Connection info from jsPlumb
             */
            onConnectionCreated(info) {
                const edge = {
                    id: `edge_${Date.now()}`,
                    source: info.source.id,
                    target: info.target.id,
                    type: 'sequential'
                };

                state.edges.push(edge);
                context.saveToHistory();
            },

            /**
             * Connection removed event handler
             * @param {Object} info - Connection info from jsPlumb
             */
            onConnectionRemoved(info) {
                state.edges = state.edges.filter(e =>
                    !(e.source === info.source.id && e.target === info.target.id)
                );
                context.saveToHistory();
            },

            /**
             * Check if auto-connected node should be disconnected
             * @param {string} nodeId - ID of node to check
             */
            checkAutoDisconnect(nodeId) {
                const node = state.nodes.find(n => n.id === nodeId);

                if (!node || !node.data.autoConnected) return;

                const sourceNode = state.nodes.find(n => n.id === node.data.autoConnected.source);
                const targetNode = state.nodes.find(n => n.id === node.data.autoConnected.target);

                if (!sourceNode || !targetNode) {
                    return;
                }

                // Calculate if node is still near the original line
                const sourceX = sourceNode.position.x + 75;
                const sourceY = sourceNode.position.y + 35;
                const targetX = targetNode.position.x + 75;
                const targetY = targetNode.position.y + 35;
                const nodeX = node.position.x + 75;
                const nodeY = node.position.y + 35;

                const distance = utils.pointToLineDistance(
                    nodeX, nodeY,
                    sourceX, sourceY,
                    targetX, targetY
                );

                // If node moved too far (threshold: 50px), disconnect it
                if (distance > 50) {
                    // Remove connections
                    const connections1 = state.jsPlumbInstance.getConnections({
                        source: node.data.autoConnected.source,
                        target: nodeId
                    });
                    connections1.forEach(conn => {
                        state.jsPlumbInstance.deleteConnection(conn);
                    });

                    const connections2 = state.jsPlumbInstance.getConnections({
                        source: nodeId,
                        target: node.data.autoConnected.target
                    });
                    connections2.forEach(conn => {
                        state.jsPlumbInstance.deleteConnection(conn);
                    });

                    // Reconnect original edge
                    const sourceEl = document.getElementById(node.data.autoConnected.source);
                    const targetEl = document.getElementById(node.data.autoConnected.target);

                    if (sourceEl && targetEl) {
                        state.jsPlumbInstance.connect({
                            source: sourceEl,
                            target: targetEl,
                            type: 'sequential'
                        });
                    }

                    // Remove auto-connected flag
                    delete node.data.autoConnected;
                }
            },

            /**
             * Check if unconnected node should be auto-connected to a nearby edge
             * @param {string} nodeId - ID of node to check
             */
            checkAutoConnect(nodeId) {
                const node = state.nodes.find(n => n.id === nodeId);
                if (!node) return;

                // Skip if node already has auto-connected flag
                if (node.data.autoConnected) return;

                // Check if node has any connections
                const hasConnections = state.edges.some(e =>
                    e.source === nodeId || e.target === nodeId
                );

                // Only auto-connect if node has no connections
                if (hasConnections) {
                    return;
                }

                // Calculate node center position
                const nodeX = node.position.x + 75;
                const nodeY = node.position.y + 35;

                // Check if node is near any edge
                const nearEdge = this.findNearbyEdge(nodeX, nodeY, 50);

                if (nearEdge) {
                    this.splitEdge(nearEdge, nodeId);
                }
            }
        };
    };

})(window);
