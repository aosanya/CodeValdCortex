package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TopologyVisualizerHandler handles topology visualization requests
type TopologyVisualizerHandler struct {
	runtime *runtime.Manager
	logger  *logrus.Logger
}

// NewTopologyVisualizerHandler creates a new topology visualizer handler
func NewTopologyVisualizerHandler(runtime *runtime.Manager, logger *logrus.Logger) *TopologyVisualizerHandler {
	return &TopologyVisualizerHandler{
		runtime: runtime,
		logger:  logger,
	}
}

// ShowTopologyVisualizer renders the topology visualizer page
func (h *TopologyVisualizerHandler) ShowTopologyVisualizer(c *gin.Context) {
	// Serve the static HTML file
	c.File("static/topology-visualizer-demo.html")
}

// ShowGeographicVisualizer renders the geographic/use-case aware visualizer
func (h *TopologyVisualizerHandler) ShowGeographicVisualizer(c *gin.Context) {
	// Serve the geographic visualizer HTML file
	c.File("static/geographic-visualizer.html")
}

// GetTopologyData returns agents and edges for visualization
func (h *TopologyVisualizerHandler) GetTopologyData(c *gin.Context) {
	agents := h.runtime.ListAgents()

	// Transform agents to topology nodes
	nodes := make([]gin.H, len(agents))
	for i, a := range agents {
		// Preserve original metadata and add runtime info
		metadata := a.Metadata
		if metadata == nil {
			metadata = make(map[string]string)
		}
		// Add runtime health info to metadata (convert bool to string)
		if a.IsHealthy() {
			metadata["healthy"] = "true"
		} else {
			metadata["healthy"] = "false"
		}
		metadata["last_heartbeat"] = a.LastHeartbeat.String()

		nodes[i] = gin.H{
			"id":         a.ID,
			"name":       a.Name,
			"agent_type": a.Type,
			"status":     string(a.GetState()),
			"metadata":   metadata,
		}
	}

	// Build edges based on agent relationships
	// This is a simplified version - you may want to query actual relationships
	edges := h.buildEdges(agents)

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
		"edges": edges,
	})
}

// buildEdges creates edges between agents based on their relationships
// This is a placeholder - you should implement based on your actual relationship model
func (h *TopologyVisualizerHandler) buildEdges(agents []*agent.Agent) []gin.H {
	edges := []gin.H{}

	// Example: Create edges based on agent types or dependencies
	// You can extend this to query actual relationships from your database
	for i := 0; i < len(agents)-1; i++ {
		// Example: Connect sequential agents
		if agents[i].Type == agents[i+1].Type {
			edges = append(edges, gin.H{
				"source": agents[i].ID,
				"target": agents[i+1].ID,
				"type":   "connection",
			})
		}
	}

	return edges
}

// GetTopologyUpdates provides real-time updates for topology changes
func (h *TopologyVisualizerHandler) GetTopologyUpdates(c *gin.Context) {
	// This can be extended to support Server-Sent Events (SSE) or WebSocket
	// For now, return current state
	h.GetTopologyData(c)
}
