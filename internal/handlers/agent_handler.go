package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AgentHandler handles HTTP requests for agent operations
type AgentHandler struct {
	runtime *runtime.Manager
	logger  *logrus.Logger
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(runtime *runtime.Manager, logger *logrus.Logger) *AgentHandler {
	return &AgentHandler{
		runtime: runtime,
		logger:  logger,
	}
}

// CreateAgentRequest represents the request body for creating an agent
type CreateAgentRequest struct {
	Name   string       `json:"name" binding:"required"`
	Type   string       `json:"type" binding:"required"`
	Config agent.Config `json:"config"`
}

// AgentResponse represents the response for agent operations
type AgentResponse struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	State         agent.State       `json:"state"`
	Metadata      map[string]string `json:"metadata"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	LastHeartbeat time.Time         `json:"last_heartbeat"`
	IsHealthy     bool              `json:"is_healthy"`
}

// toAgentResponse converts an agent to a response model
func toAgentResponse(a *agent.Agent) AgentResponse {
	return AgentResponse{
		ID:            a.ID,
		Name:          a.Name,
		Type:          a.Type,
		State:         a.GetState(),
		Metadata:      a.Metadata,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		LastHeartbeat: a.LastHeartbeat,
		IsHealthy:     a.IsHealthy(),
	}
}

// CreateAgent godoc
// @Summary Create a new agent
// @Description Creates a new agent with the specified configuration
// @Tags agents
// @Accept json
// @Produce json
// @Param agent body CreateAgentRequest true "Agent configuration"
// @Success 201 {object} AgentResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents [post]
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default config values if not provided
	if req.Config.MaxConcurrentTasks == 0 {
		req.Config.MaxConcurrentTasks = 5
	}
	if req.Config.TaskQueueSize == 0 {
		req.Config.TaskQueueSize = 100
	}
	if req.Config.HeartbeatInterval == 0 {
		req.Config.HeartbeatInterval = 30 * time.Second
	}
	if req.Config.TaskTimeout == 0 {
		req.Config.TaskTimeout = 5 * time.Minute
	}

	a, err := h.runtime.CreateAgent(req.Name, req.Type, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toAgentResponse(a))
}

// StartAgent godoc
// @Summary Start an agent
// @Description Starts an agent and begins task processing
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} AgentResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/start [post]
func (h *AgentHandler) StartAgent(c *gin.Context) {
	agentID := c.Param("id")

	if err := h.runtime.StartAgent(agentID); err != nil {
		if err == agent.ErrAgentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	a, _ := h.runtime.GetAgent(agentID)
	c.JSON(http.StatusOK, toAgentResponse(a))
}

// StopAgent godoc
// @Summary Stop an agent
// @Description Stops an agent gracefully
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} AgentResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/stop [post]
func (h *AgentHandler) StopAgent(c *gin.Context) {
	agentID := c.Param("id")

	if err := h.runtime.StopAgent(agentID); err != nil {
		if err == agent.ErrAgentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	a, _ := h.runtime.GetAgent(agentID)
	c.JSON(http.StatusOK, toAgentResponse(a))
}

// PauseAgent godoc
// @Summary Pause an agent
// @Description Pauses a running agent
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} AgentResponse
// @Failure 404 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/pause [post]
func (h *AgentHandler) PauseAgent(c *gin.Context) {
	agentID := c.Param("id")

	if err := h.runtime.PauseAgent(agentID); err != nil {
		if err == agent.ErrAgentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a, _ := h.runtime.GetAgent(agentID)
	c.JSON(http.StatusOK, toAgentResponse(a))
}

// ResumeAgent godoc
// @Summary Resume an agent
// @Description Resumes a paused agent
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} AgentResponse
// @Failure 404 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/resume [post]
func (h *AgentHandler) ResumeAgent(c *gin.Context) {
	agentID := c.Param("id")

	if err := h.runtime.ResumeAgent(agentID); err != nil {
		if err == agent.ErrAgentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a, _ := h.runtime.GetAgent(agentID)
	c.JSON(http.StatusOK, toAgentResponse(a))
}

// RestartAgent godoc
// @Summary Restart an agent
// @Description Stops and restarts an agent
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} AgentResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/restart [post]
func (h *AgentHandler) RestartAgent(c *gin.Context) {
	agentID := c.Param("id")

	if err := h.runtime.RestartAgent(agentID); err != nil {
		if err == agent.ErrAgentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	a, _ := h.runtime.GetAgent(agentID)
	c.JSON(http.StatusOK, toAgentResponse(a))
}

// GetAgent godoc
// @Summary Get agent details
// @Description Retrieves details of a specific agent
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} AgentResponse
// @Failure 404 {object} map[string]string
// @Router /agents/{id} [get]
func (h *AgentHandler) GetAgent(c *gin.Context) {
	agentID := c.Param("id")

	a, err := h.runtime.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, toAgentResponse(a))
}

// ListAgents godoc
// @Summary List all agents
// @Description Retrieves a list of all agents
// @Tags agents
// @Produce json
// @Success 200 {array} AgentResponse
// @Router /agents [get]
func (h *AgentHandler) ListAgents(c *gin.Context) {
	// Support optional pagination query params: page and limit
	page := 1
	limit := 50

	if p := c.Query("page"); p != "" {
		if pv, err := strconv.Atoi(p); err == nil && pv > 0 {
			page = pv
		}
	}

	if l := c.Query("limit"); l != "" {
		if lv, err := strconv.Atoi(l); err == nil && lv > 0 && lv <= 1000 {
			limit = lv
		}
	}

	allAgents := h.runtime.ListAgents()

	// Safety: ensure deterministic ordering before pagination
	sort.Slice(allAgents, func(i, j int) bool { return allAgents[i].ID < allAgents[j].ID })

	total := len(allAgents)
	startIdx := (page - 1) * limit
	endIdx := startIdx + limit
	if startIdx >= total {
		c.JSON(http.StatusOK, []AgentResponse{})
		return
	}
	if endIdx > total {
		endIdx = total
	}

	agents := allAgents[startIdx:endIdx]

	response := make([]AgentResponse, 0, len(agents))
	for _, a := range agents {
		response = append(response, toAgentResponse(a))
	}

	c.JSON(http.StatusOK, response)
}

// SubmitTaskRequest represents the request body for submitting a task
type SubmitTaskRequest struct {
	Type     string      `json:"type" binding:"required"`
	Payload  interface{} `json:"payload"`
	Priority int         `json:"priority"`
	Timeout  int         `json:"timeout"` // in seconds
}

// SubmitTask godoc
// @Summary Submit a task to an agent
// @Description Submits a task to the specified agent's queue
// @Tags agents
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param task body SubmitTaskRequest true "Task details"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /agents/{id}/tasks [post]
func (h *AgentHandler) SubmitTask(c *gin.Context) {
	agentID := c.Param("id")

	var req SubmitTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a, err := h.runtime.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	// Create task
	task := agent.Task{
		ID:        generateTaskID(),
		Type:      req.Type,
		Payload:   req.Payload,
		Priority:  req.Priority,
		Timeout:   time.Duration(req.Timeout) * time.Second,
		CreatedAt: time.Now().UTC(),
	}

	// Submit task to agent
	if err := a.SubmitTask(task); err != nil {
		if err == agent.ErrTaskQueueFull {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "task queue is full"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"task_id":  task.ID,
		"agent_id": agentID,
		"message":  "task submitted successfully",
	})
}

// GetMetrics godoc
// @Summary Get runtime metrics
// @Description Retrieves runtime metrics for all agents
// @Tags metrics
// @Produce json
// @Success 200 {object} runtime.Metrics
// @Router /metrics [get]
func (h *AgentHandler) GetMetrics(c *gin.Context) {
	metrics := h.runtime.GetMetrics()
	c.JSON(http.StatusOK, metrics)
}

// generateTaskID generates a unique task ID
func generateTaskID() string {
	return "task-" + uuid.New().String()
}

// RegisterRoutes registers all agent-related routes
func (h *AgentHandler) RegisterRoutes(router *gin.Engine) {
	agents := router.Group("/api/v1/agents")
	{
		agents.POST("", h.CreateAgent)
		agents.GET("", h.ListAgents)
		agents.GET("/:id", h.GetAgent)
		agents.POST("/:id/start", h.StartAgent)
		agents.POST("/:id/stop", h.StopAgent)
		agents.POST("/:id/pause", h.PauseAgent)
		agents.POST("/:id/resume", h.ResumeAgent)
		agents.POST("/:id/restart", h.RestartAgent)
		agents.POST("/:id/tasks", h.SubmitTask)
	}

	metrics := router.Group("/api/v1/metrics")
	{
		metrics.GET("", h.GetMetrics)
	}
}
