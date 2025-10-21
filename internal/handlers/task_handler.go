package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/runtime"
	"github.com/aosanya/CodeValdCortex/internal/task"
	"github.com/gin-gonic/gin"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	manager *runtime.Manager
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(manager *runtime.Manager) *TaskHandler {
	return &TaskHandler{
		manager: manager,
	}
}

// AdvancedTaskRequest represents an advanced task submission request
type AdvancedTaskRequest struct {
	Type         string                 `json:"type" binding:"required"`
	Name         string                 `json:"name,omitempty"`
	Payload      map[string]interface{} `json:"payload,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
	Timeout      string                 `json:"timeout,omitempty"` // Duration string like "5m"
	Dependencies []string               `json:"dependencies,omitempty"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
}

// SubmitAdvancedTask godoc
// @Summary Submit an advanced task to an agent
// @Description Submit a task with advanced features like priority, dependencies, and custom handlers
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param task body AdvancedTaskRequest true "Task details"
// @Success 202 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/advanced-tasks [post]
func (h *TaskHandler) SubmitAdvancedTask(c *gin.Context) {
	agentID := c.Param("id")

	var req AdvancedTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create task
	taskReq := &task.Task{
		Type:         req.Type,
		Name:         req.Name,
		Payload:      req.Payload,
		Priority:     req.Priority,
		Dependencies: req.Dependencies,
		Metadata:     req.Metadata,
	}

	// Parse timeout if provided
	if req.Timeout != "" {
		timeout, err := parseTimeout(req.Timeout)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timeout format"})
			return
		}
		taskReq.Timeout = timeout
	}

	// Submit task
	err := h.manager.SubmitTaskToAgent(c.Request.Context(), agentID, taskReq)
	if err != nil {
		if err.Error() == "agent not found: "+agentID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"task_id":  taskReq.ID,
		"agent_id": agentID,
		"status":   "submitted",
	})
}

// GetTask godoc
// @Summary Get task details
// @Description Retrieve details of a specific task
// @Tags tasks
// @Produce json
// @Param id path string true "Agent ID"
// @Param taskId path string true "Task ID"
// @Success 200 {object} task.Task
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/tasks/{taskId} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	agentID := c.Param("id")
	taskID := c.Param("taskId")

	task, err := h.manager.GetAgentTask(c.Request.Context(), agentID, taskID)
	if err != nil {
		if err.Error() == "agent not found: "+agentID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetTaskResult godoc
// @Summary Get task execution result
// @Description Retrieve the result of a completed task
// @Tags tasks
// @Produce json
// @Param id path string true "Agent ID"
// @Param taskId path string true "Task ID"
// @Success 200 {object} task.TaskResult
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/tasks/{taskId}/result [get]
func (h *TaskHandler) GetTaskResult(c *gin.Context) {
	agentID := c.Param("id")
	taskID := c.Param("taskId")

	result, err := h.manager.GetAgentTaskResult(c.Request.Context(), agentID, taskID)
	if err != nil {
		if err.Error() == "agent not found: "+agentID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Task result not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CancelTask godoc
// @Summary Cancel a task
// @Description Cancel a pending or running task
// @Tags tasks
// @Produce json
// @Param id path string true "Agent ID"
// @Param taskId path string true "Task ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/tasks/{taskId}/cancel [post]
func (h *TaskHandler) CancelTask(c *gin.Context) {
	agentID := c.Param("id")
	taskID := c.Param("taskId")

	err := h.manager.CancelAgentTask(c.Request.Context(), agentID, taskID)
	if err != nil {
		if err.Error() == "agent not found: "+agentID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  "cancelled",
	})
}

// ListTasks godoc
// @Summary List agent tasks
// @Description List all tasks for an agent with optional filtering
// @Tags tasks
// @Produce json
// @Param id path string true "Agent ID"
// @Param status query []string false "Filter by task status"
// @Param type query string false "Filter by task type"
// @Param min_priority query int false "Filter by minimum priority"
// @Param limit query int false "Limit number of results"
// @Success 200 {array} task.Task
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	agentID := c.Param("id")

	// Parse query parameters
	filters := task.TaskFilters{}

	// Status filter
	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		for _, status := range statuses {
			filters.Status = append(filters.Status, task.TaskStatus(status))
		}
	}

	// Type filter
	if taskType := c.Query("type"); taskType != "" {
		filters.Type = taskType
	}

	// Priority filter
	if minPriorityStr := c.Query("min_priority"); minPriorityStr != "" {
		if minPriority, err := strconv.Atoi(minPriorityStr); err == nil {
			filters.MinPriority = minPriority
		}
	}

	// Limit filter
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	tasks, err := h.manager.ListAgentTasks(c.Request.Context(), agentID, filters)
	if err != nil {
		if err.Error() == "agent not found: "+agentID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTaskMetrics godoc
// @Summary Get agent task metrics
// @Description Retrieve task execution metrics for an agent
// @Tags tasks
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} task.AgentTaskMetrics
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /agents/{id}/task-metrics [get]
func (h *TaskHandler) GetTaskMetrics(c *gin.Context) {
	agentID := c.Param("id")

	metrics, err := h.manager.GetAgentTaskMetrics(c.Request.Context(), agentID)
	if err != nil {
		if err.Error() == "agent not found: "+agentID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetTaskManagerStatus godoc
// @Summary Get task manager status
// @Description Get the status of the agent's task manager
// @Tags tasks
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /agents/{id}/task-manager/status [get]
func (h *TaskHandler) GetTaskManagerStatus(c *gin.Context) {
	agentID := c.Param("id")

	status := h.manager.GetAgentTaskManagerStatus(agentID)
	if status["error"] != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// RegisterRoutes registers all task-related routes
func (h *TaskHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// Task management routes
		v1.POST("/agents/:id/advanced-tasks", h.SubmitAdvancedTask)
		v1.GET("/agents/:id/tasks", h.ListTasks)
		v1.GET("/agents/:id/tasks/:taskId", h.GetTask)
		v1.GET("/agents/:id/tasks/:taskId/result", h.GetTaskResult)
		v1.POST("/agents/:id/tasks/:taskId/cancel", h.CancelTask)
		v1.GET("/agents/:id/task-metrics", h.GetTaskMetrics)
		v1.GET("/agents/:id/task-manager/status", h.GetTaskManagerStatus)
	}
}

// parseTimeout parses a timeout string into a time.Duration
func parseTimeout(timeoutStr string) (time.Duration, error) {
	return time.ParseDuration(timeoutStr)
}
