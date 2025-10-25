package handlers

import (
	"net/http"
	"strconv"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	"github.com/aosanya/CodeValdCortex/internal/web/components"
	"github.com/aosanya/CodeValdCortex/internal/web/pages"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

// DashboardHandler handles web dashboard requests
type DashboardHandler struct {
	runtime *runtime.Manager
	logger  *logrus.Logger
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(runtime *runtime.Manager, logger *logrus.Logger) *DashboardHandler {
	return &DashboardHandler{
		runtime: runtime,
		logger:  logger,
	}
}

// ShowDashboard renders the main dashboard page
func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	// Get all agents
	allAgents := h.runtime.ListAgents()

	// Filter by agency if one is selected
	var agents []*agent.Agent
	if ag, exists := c.Get("agency"); exists {
		if agencyPtr, ok := ag.(*agency.Agency); ok {
			// Filter agents by agency
			for _, a := range allAgents {
				if agencyID, exists := a.Metadata["agency_id"]; exists && agencyID == agencyPtr.ID {
					agents = append(agents, a)
				}
			}
		} else {
			agents = allAgents
		}
	} else {
		agents = allAgents
	}

	stats := h.calculateStats(agents)

	// Get current agency from context (if available)
	var currentAgency *agency.Agency
	if ag, exists := c.Get("agency"); exists {
		if agencyPtr, ok := ag.(*agency.Agency); ok {
			currentAgency = agencyPtr
		}
	}

	// Render Templ component
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := pages.Dashboard(agents, stats, currentAgency).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.Errorf("Failed to render dashboard: %v", err)
		c.String(http.StatusInternalServerError, "Failed to render dashboard")
		return
	}
}

// GetAgentsLive returns OOB status updates for all agents without changing pagination
// React-like component updates: only updates changed data, preserves user interactions
func (h *DashboardHandler) GetAgentsLive(c *gin.Context) {
	allAgents := h.runtime.ListAgents()

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Check if context is already canceled before starting
	if c.Request.Context().Err() != nil {
		h.logger.Warnf("Request context already canceled, skipping agent status updates")
		return
	}

	// Send granular OOB updates per agent (React-like virtual DOM)
	// This updates only dynamic data:
	//   1. Status badge (state text + styling)
	//   2. Health badge (health text + styling)
	//   3. Heartbeat timestamp
	//   4. Action buttons (state-dependent)
	// Uses innerHTML swap to update only content, preserving DOM structure and Alpine.js state
	for _, a := range allAgents {
		// Check context before each render to avoid cascading errors
		if c.Request.Context().Err() != nil {
			break
		}

		// Update 1: Status badge (innerHTML - only text changes, classes update via full span replacement)
		err := components.AgentStatusOOB(a).Render(c.Request.Context(), c.Writer)
		if err != nil {
			h.logger.Warnf("Failed to render agent status %s: %v", a.ID, err)
			continue
		}

		// Update 2: Health badge (innerHTML - only text changes, classes update via full span replacement)
		err = components.AgentHealthOOB(a).Render(c.Request.Context(), c.Writer)
		if err != nil {
			h.logger.Warnf("Failed to render agent health %s: %v", a.ID, err)
			continue
		}

		// Update 3: Heartbeat timestamp (innerHTML - only text content)
		err = components.AgentHeartbeatOOB(a).Render(c.Request.Context(), c.Writer)
		if err != nil {
			h.logger.Warnf("Failed to render agent heartbeat %s: %v", a.ID, err)
			continue
		}

		// Update 4: Action buttons (innerHTML - state-dependent button set)
		err = components.AgentActionsOOB(a).Render(c.Request.Context(), c.Writer)
		if err != nil {
			h.logger.Warnf("Failed to render agent actions %s: %v", a.ID, err)
			continue
		}
	}
} // HandleAgentAction handles start/stop/restart/pause/resume actions via HTMX
func (h *DashboardHandler) HandleAgentAction(c *gin.Context) {
	agentID := c.Param("id")
	action := c.Param("action")

	h.logger.Infof("Agent action: %s on agent %s", action, agentID)

	// Verify agent exists
	_, err := h.runtime.GetAgent(agentID)
	if err != nil {
		h.logger.Errorf("Failed to get agent %s: %v", agentID, err)
		c.String(http.StatusNotFound, "Agent not found")
		return
	}

	switch action {
	case "start":
		err = h.runtime.StartAgent(agentID)
	case "stop":
		err = h.runtime.StopAgent(agentID)
	case "restart":
		err = h.runtime.RestartAgent(agentID)
	case "pause":
		err = h.runtime.PauseAgent(agentID)
	case "resume":
		err = h.runtime.ResumeAgent(agentID)
	default:
		c.String(http.StatusBadRequest, "Unknown action")
		return
	}

	if err != nil {
		h.logger.Errorf("Failed to %s agent %s: %v", action, agentID, err)
		c.String(http.StatusInternalServerError, "Action failed: "+err.Error())
		return
	}

	// Return updated agent card
	agent, err := h.runtime.GetAgent(agentID)
	if err != nil {
		h.logger.Errorf("Failed to get updated agent %s: %v", agentID, err)
		c.String(http.StatusInternalServerError, "Failed to get updated agent")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = components.AgentCard(agent).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.Warnf("Failed to render agent card for %s: %v", agentID, err)
		c.String(http.StatusInternalServerError, "Failed to render response")
		return
	}
}

func (h *DashboardHandler) calculateStats(agents []*agent.Agent) pages.DashboardStats {
	stats := pages.DashboardStats{
		Total: len(agents),
	}

	for _, a := range agents {
		state := a.GetState()
		switch state {
		case agent.StateRunning:
			stats.Running++
		case agent.StateStopped:
			stats.Stopped++
		case agent.StatePaused:
			stats.Paused++
		}

		if a.IsHealthy() {
			stats.Healthy++
		} else {
			stats.Unhealthy++
		}
	}

	return stats
}

// GetAgentsJSON returns agents data as JSON with pagination (more efficient for large datasets)
func (h *DashboardHandler) GetAgentsJSON(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	pageSize := defaultPageSize

	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := c.Query("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= maxPageSize {
			pageSize = s
		}
	}

	allAgents := h.runtime.ListAgents()
	totalAgents := len(allAgents)

	// Calculate pagination
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	// Handle out of bounds
	if startIdx >= totalAgents {
		c.JSON(http.StatusOK, gin.H{
			"agents":      []interface{}{},
			"total":       totalAgents,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (totalAgents + pageSize - 1) / pageSize,
		})
		return
	}

	if endIdx > totalAgents {
		endIdx = totalAgents
	}

	agents := allAgents[startIdx:endIdx]

	// Build lightweight response
	agentData := make([]gin.H, len(agents))
	for i, a := range agents {
		agentData[i] = gin.H{
			"id":             a.ID,
			"name":           a.Name,
			"type":           a.Type,
			"state":          string(a.GetState()),
			"healthy":        a.IsHealthy(),
			"created_at":     a.CreatedAt,
			"last_heartbeat": a.LastHeartbeat,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"agents":      agentData,
		"total":       totalAgents,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (totalAgents + pageSize - 1) / pageSize,
	})
}
