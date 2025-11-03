package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
)

// GetAgencyRoles handles GET /api/v1/agencies/:id/roles
func (h *AgencyHandler) GetAgencyRoles(c *gin.Context) {
	// For now, list all roles from the registry
	// TODO: Filter by agency when agency-specific roles are implemented
	roles, err := h.roleService.ListTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter out system roles (core, monitoring, etc.)
	userRoles := make([]*registry.Role, 0)
	for _, role := range roles {
		if !role.IsSystemType {
			userRoles = append(userRoles, role)
		}
	}

	c.JSON(http.StatusOK, userRoles)
}

// GetAgencyRolesHTML handles GET /api/v1/agencies/:id/roles/html
// Returns rendered HTML fragment for HTMX/JavaScript rendering
func (h *AgencyHandler) GetAgencyRolesHTML(c *gin.Context) {
	// For now, list all roles from the registry
	// TODO: Filter by agency when agency-specific roles are implemented
	roles, err := h.roleService.ListTypes(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list roles")
		c.String(http.StatusInternalServerError, "Error loading roles")
		return
	}

	// Filter out system roles (core, monitoring, etc.)
	userRoles := make([]*registry.Role, 0)
	for _, role := range roles {
		if !role.IsSystemType {
			userRoles = append(userRoles, role)
		}
	}

	h.logger.Infof("Returning %d user-defined roles for HTML rendering", len(userRoles))

	// Render the roles list template
	component := agency_designer.AgencyRolesList(userRoles)
	c.Header("Content-Type", "text/html")
	component.Render(c.Request.Context(), c.Writer)
}

// CreateAgencyRole handles POST /api/v1/agencies/:id/roles
func (h *AgencyHandler) CreateAgencyRole(c *gin.Context) {
	agencyID := c.Param("id")

	var role registry.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		h.logger.WithError(err).Error("Failed to bind role data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role data"})
		return
	}

	// Set creation metadata
	role.CreatedBy = "api"
	role.IsSystemType = false
	role.IsEnabled = true

	// Register the role
	if err := h.roleService.RegisterType(c.Request.Context(), &role); err != nil {
		h.logger.WithError(err).Error("Failed to create role", "agency_id", agencyID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Role created", "agency_id", agencyID, "role_id", role.ID)
	c.JSON(http.StatusCreated, role)
}

// GetAgencyRole handles GET /api/v1/agencies/:id/roles/:key
func (h *AgencyHandler) GetAgencyRole(c *gin.Context) {
	key := c.Param("key")

	role, err := h.roleService.GetType(c.Request.Context(), key)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get role", "key", key)
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// UpdateAgencyRole handles PUT /api/v1/agencies/:id/roles/:key
func (h *AgencyHandler) UpdateAgencyRole(c *gin.Context) {
	agencyID := c.Param("id")
	key := c.Param("key")

	// Check if role exists and is not a system role
	existingRole, err := h.roleService.GetType(c.Request.Context(), key)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get role for update", "key", key)
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Prevent editing of system roles
	if existingRole.IsSystemType {
		h.logger.Warn("Attempted to update system role", "key", key)
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot edit system roles"})
		return
	}

	var role registry.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		h.logger.WithError(err).Error("Failed to bind role data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role data"})
		return
	}

	// Set the key/ID
	role.Key = key
	role.ID = key

	// Update the role
	if err := h.roleService.UpdateType(c.Request.Context(), &role); err != nil {
		h.logger.WithError(err).Error("Failed to update role", "agency_id", agencyID, "key", key)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Role updated", "agency_id", agencyID, "role_key", key)
	c.JSON(http.StatusOK, role)
}

// DeleteAgencyRole handles DELETE /api/v1/agencies/:id/roles/:key
func (h *AgencyHandler) DeleteAgencyRole(c *gin.Context) {
	agencyID := c.Param("id")
	key := c.Param("key")

	// Check if role exists and is not a system role
	role, err := h.roleService.GetType(c.Request.Context(), key)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get role for deletion", "key", key)
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Prevent deletion of system roles
	if role.IsSystemType {
		h.logger.Warn("Attempted to delete system role", "key", key)
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete system roles"})
		return
	}

	if err := h.roleService.UnregisterType(c.Request.Context(), key); err != nil {
		h.logger.WithError(err).Error("Failed to delete role", "agency_id", agencyID, "key", key)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Role deleted", "agency_id", agencyID, "role_key", key)
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}
