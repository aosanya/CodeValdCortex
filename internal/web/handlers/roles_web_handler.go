package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/aosanya/CodeValdCortex/internal/web/components"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RolesWebHandler handles web interface requests for agent types
type RolesWebHandler struct {
	service registry.RoleService
	logger  *logrus.Logger
}

// NewRolesWebHandler creates a new agent types web handler
func NewRolesWebHandler(service registry.RoleService, logger *logrus.Logger) *RolesWebHandler {
	return &RolesWebHandler{
		service: service,
		logger:  logger,
	}
}

// ShowRoles renders a simple roles listing page (roles are primarily managed in Agency Designer)
func (h *RolesWebHandler) ShowRoles(c *gin.Context) {
	ctx := c.Request.Context()

	roles, err := h.service.ListTypes(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list roles")
		c.String(http.StatusInternalServerError, "Failed to load roles")
		return
	}

	// For now, return JSON until we create a dedicated roles page template
	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"count": len(roles),
	})
}

// GetRolesLive returns roles grid for HTMX updates
func (h *RolesWebHandler) GetRolesLive(c *gin.Context) {
	ctx := c.Request.Context()

	// Check for tags filter
	tagFilter := c.Query("tag")
	enabledOnly := c.Query("enabled") == "true"

	var roles []*registry.Role
	var err error

	if tagFilter != "" {
		roles, err = h.service.ListTypesByTags(ctx, []string{tagFilter})
	} else {
		roles, err = h.service.ListTypes(ctx)
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to list roles")
		c.String(http.StatusInternalServerError, "Failed to load roles")
		return
	}

	// Filter by enabled if requested
	if enabledOnly {
		filtered := make([]*registry.Role, 0)
		for _, t := range roles {
			if t.IsEnabled {
				filtered = append(filtered, t)
			}
		}
		roles = filtered
	}

	// Return only the role cards (partial HTML)
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, role := range roles {
		err := components.RoleCard(role).Render(ctx, c.Writer)
		if err != nil {
			h.logger.WithError(err).Error("Failed to render role card")
			continue
		}
	}
}

// HandleRoleAction handles enable/disable actions via HTMX
func (h *RolesWebHandler) HandleRoleAction(c *gin.Context) {
	ctx := c.Request.Context()
	typeID := c.Param("id")
	action := c.Param("action")

	h.logger.Infof("Agent type action: %s on type %s", action, typeID)

	var err error
	switch action {
	case "enable":
		err = h.service.EnableType(ctx, typeID)
	case "disable":
		err = h.service.DisableType(ctx, typeID)
	default:
		c.String(http.StatusBadRequest, "Invalid action")
		return
	}

	if err != nil {
		h.logger.WithError(err).Errorf("Failed to %s agent type %s", action, typeID)
		c.String(http.StatusInternalServerError, "Action failed")
		return
	}

	// Return updated agent type card
	agentType, err := h.service.GetType(ctx, typeID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get updated agent type")
		c.String(http.StatusInternalServerError, "Failed to refresh")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = components.RoleCard(agentType).Render(ctx, c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render agent type card")
		c.String(http.StatusInternalServerError, "Failed to render")
	}
}
