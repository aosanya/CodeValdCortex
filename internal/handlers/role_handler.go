package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RoleHandler handles HTTP requests for role operations
type RoleHandler struct {
	service registry.RoleService
	logger  *logrus.Logger
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(service registry.RoleService, logger *logrus.Logger) *RoleHandler {
	return &RoleHandler{
		service: service,
		logger:  logger,
	}
}

// ListRoles handles GET /api/v1/roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
	ctx := c.Request.Context()

	// Check for category filter
	category := c.Query("category")
	enabledOnly := c.Query("enabled") == "true"

	var types []*registry.Role
	var err error

	if category != "" {
		types, err = h.service.ListTypesByCategory(ctx, category)
	} else if enabledOnly {
		types, err = h.service.ListTypes(ctx)
		// Filter enabled
		enabled := make([]*registry.Role, 0)
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
		h.logger.WithError(err).Error("Failed to list roles")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": types,
		"count": len(types),
	})
}

// GetRole handles GET /api/v1/roles/:id
func (h *RoleHandler) GetRole(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	agentType, err := h.service.GetType(ctx, id)
	if err != nil {
		h.logger.WithError(err).WithField("type_id", id).Warn("Agent type not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "role not found",
		})
		return
	}

	c.JSON(http.StatusOK, agentType)
}

// CreateRole handles POST /api/v1/roles
func (h *RoleHandler) CreateRole(c *gin.Context) {
	ctx := c.Request.Context()

	var agentType registry.Role
	if err := c.ShouldBindJSON(&agentType); err != nil {
		h.logger.WithError(err).Warn("Invalid role request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Set created by from context (would normally come from auth)
	agentType.CreatedBy = "api"

	if err := h.service.RegisterType(ctx, &agentType); err != nil {
		h.logger.WithError(err).Error("Failed to create role")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, agentType)
}

// UpdateRole handles PUT /api/v1/roles/:id
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var agentType registry.Role
	if err := c.ShouldBindJSON(&agentType); err != nil {
		h.logger.WithError(err).Warn("Invalid role request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Ensure ID matches
	agentType.ID = id

	if err := h.service.UpdateType(ctx, &agentType); err != nil {
		h.logger.WithError(err).Error("Failed to update role")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, agentType)
}

// DeleteRole handles DELETE /api/v1/roles/:id
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.service.UnregisterType(ctx, id); err != nil {
		h.logger.WithError(err).Error("Failed to delete role")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "role deleted successfully",
	})
}

// EnableRole handles POST /api/v1/roles/:id/enable
func (h *RoleHandler) EnableRole(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.service.EnableType(ctx, id); err != nil {
		h.logger.WithError(err).Error("Failed to enable role")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "role enabled successfully",
	})
}

// DisableRole handles POST /api/v1/roles/:id/disable
func (h *RoleHandler) DisableRole(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.service.DisableType(ctx, id); err != nil {
		h.logger.WithError(err).Error("Failed to disable role")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "role disabled successfully",
	})
}
