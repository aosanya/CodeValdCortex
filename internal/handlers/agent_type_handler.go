package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AgentTypeHandler handles HTTP requests for agent type operations
type AgentTypeHandler struct {
	service registry.AgentTypeService
	logger  *logrus.Logger
}

// NewAgentTypeHandler creates a new agent type handler
func NewAgentTypeHandler(service registry.AgentTypeService, logger *logrus.Logger) *AgentTypeHandler {
	return &AgentTypeHandler{
		service: service,
		logger:  logger,
	}
}

// ListAgentTypes handles GET /api/v1/agent-types
func (h *AgentTypeHandler) ListAgentTypes(c *gin.Context) {
	ctx := c.Request.Context()

	// Check for category filter
	category := c.Query("category")
	enabledOnly := c.Query("enabled") == "true"

	var types []*registry.AgentType
	var err error

	if category != "" {
		types, err = h.service.ListTypesByCategory(ctx, category)
	} else if enabledOnly {
		types, err = h.service.ListTypes(ctx)
		// Filter enabled
		enabled := make([]*registry.AgentType, 0)
		for _, t := range types {
			if t.IsEnabled {
				enabled = append(enabled, t)
			}
		}
		types = enabled
	} else {
		types, err = h.service.ListTypes(ctx)
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to list agent types")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list agent types",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agent_types": types,
		"count":       len(types),
	})
}

// GetAgentType handles GET /api/v1/agent-types/:id
func (h *AgentTypeHandler) GetAgentType(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	agentType, err := h.service.GetType(ctx, id)
	if err != nil {
		h.logger.WithError(err).WithField("type_id", id).Warn("Agent type not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "agent type not found",
		})
		return
	}

	c.JSON(http.StatusOK, agentType)
}

// CreateAgentType handles POST /api/v1/agent-types
func (h *AgentTypeHandler) CreateAgentType(c *gin.Context) {
	ctx := c.Request.Context()

	var agentType registry.AgentType
	if err := c.ShouldBindJSON(&agentType); err != nil {
		h.logger.WithError(err).Warn("Invalid agent type request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Set created by from context (would normally come from auth)
	agentType.CreatedBy = "api"

	if err := h.service.RegisterType(ctx, &agentType); err != nil {
		h.logger.WithError(err).Error("Failed to create agent type")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, agentType)
}

// UpdateAgentType handles PUT /api/v1/agent-types/:id
func (h *AgentTypeHandler) UpdateAgentType(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var agentType registry.AgentType
	if err := c.ShouldBindJSON(&agentType); err != nil {
		h.logger.WithError(err).Warn("Invalid agent type request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Ensure ID matches
	agentType.ID = id

	if err := h.service.UpdateType(ctx, &agentType); err != nil {
		h.logger.WithError(err).Error("Failed to update agent type")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, agentType)
}

// DeleteAgentType handles DELETE /api/v1/agent-types/:id
func (h *AgentTypeHandler) DeleteAgentType(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.service.UnregisterType(ctx, id); err != nil {
		h.logger.WithError(err).Error("Failed to delete agent type")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "agent type deleted successfully",
	})
}

// EnableAgentType handles POST /api/v1/agent-types/:id/enable
func (h *AgentTypeHandler) EnableAgentType(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.service.EnableType(ctx, id); err != nil {
		h.logger.WithError(err).Error("Failed to enable agent type")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "agent type enabled successfully",
	})
}

// DisableAgentType handles POST /api/v1/agent-types/:id/disable
func (h *AgentTypeHandler) DisableAgentType(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.service.DisableType(ctx, id); err != nil {
		h.logger.WithError(err).Error("Failed to disable agent type")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "agent type disabled successfully",
	})
}
