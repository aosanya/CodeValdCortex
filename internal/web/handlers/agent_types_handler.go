package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/aosanya/CodeValdCortex/internal/web/components"
	"github.com/aosanya/CodeValdCortex/internal/web/pages"
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

// ShowRoles renders the agent types page
func (h *RolesWebHandler) ShowRoles(c *gin.Context) {
	ctx := c.Request.Context()

	agentTypes, err := h.service.ListTypes(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list agent types")
		c.String(http.StatusInternalServerError, "Failed to load agent types")
		return
	}

	// Render Templ component
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = pages.RolesPage(agentTypes).Render(ctx, c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render agent types page")
		c.String(http.StatusInternalServerError, "Failed to render page")
		return
	}
}

// GetRolesLive returns agent types grid for HTMX updates
func (h *RolesWebHandler) GetRolesLive(c *gin.Context) {
	ctx := c.Request.Context()

	// Check for filters
	category := c.Query("category")
	enabledOnly := c.Query("enabled") == "true"

	var agentTypes []*registry.Role
	var err error

	if category != "" {
		agentTypes, err = h.service.ListTypesByCategory(ctx, category)
	} else {
		agentTypes, err = h.service.ListTypes(ctx)
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to list agent types")
		c.String(http.StatusInternalServerError, "Failed to load agent types")
		return
	}

	// Filter by enabled if requested
	if enabledOnly {
		filtered := make([]*registry.Role, 0)
		for _, t := range agentTypes {
			if t.IsEnabled {
				filtered = append(filtered, t)
			}
		}
		agentTypes = filtered
	}

	// Return only the agent type cards (partial HTML)
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, agentType := range agentTypes {
		err := components.RoleCard(agentType).Render(ctx, c.Writer)
		if err != nil {
			h.logger.WithError(err).Error("Failed to render agent type card")
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
