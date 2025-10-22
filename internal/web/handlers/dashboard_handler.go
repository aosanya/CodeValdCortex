package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	"github.com/aosanya/CodeValdCortex/internal/web/components"
	"github.com/aosanya/CodeValdCortex/internal/web/pages"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	agents := h.runtime.ListAgents()
	stats := h.calculateStats(agents)

	// Render Templ component
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := pages.Dashboard(agents, stats).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.Errorf("Failed to render dashboard: %v", err)
		c.String(http.StatusInternalServerError, "Failed to render dashboard")
		return
	}
}

// GetAgentsLive returns just the agents grid for HTMX updates
func (h *DashboardHandler) GetAgentsLive(c *gin.Context) {
	agents := h.runtime.ListAgents()

	// Return only the agent cards (partial HTML)
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, a := range agents {
		err := components.AgentCard(a).Render(c.Request.Context(), c.Writer)
		if err != nil {
			h.logger.Errorf("Failed to render agent card: %v", err)
			continue
		}
	}
}

// HandleAgentAction handles start/stop/restart/pause/resume actions via HTMX
func (h *DashboardHandler) HandleAgentAction(c *gin.Context) {
	agentID := c.Param("id")
	action := c.Param("action")

	h.logger.Infof("Agent action: %s on agent %s", action, agentID)

	agent, err := h.runtime.GetAgent(agentID)
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
	agent, err = h.runtime.GetAgent(agentID)
	if err != nil {
		h.logger.Errorf("Failed to get updated agent %s: %v", agentID, err)
		c.String(http.StatusInternalServerError, "Failed to get updated agent")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = components.AgentCard(agent).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.Errorf("Failed to render agent card: %v", err)
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
