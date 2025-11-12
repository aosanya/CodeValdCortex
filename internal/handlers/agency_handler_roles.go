package handlers

import (
"net/http"
"sort"

"github.com/aosanya/CodeValdCortex/internal/agency/models"
"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
"github.com/gin-gonic/gin"
)

// GetAgencyRoles handles GET /api/v1/agencies/:id/roles
func (h *AgencyHandler) GetAgencyRoles(c *gin.Context) {
agencyID := c.Param("id")

// Get agency specification which contains roles
spec, err := h.service.GetSpecification(c.Request.Context(), agencyID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, spec.Roles)
}

// GetAgencyRolesHTML handles GET /api/v1/agencies/:id/roles/html
// Returns rendered HTML fragment for HTMX/JavaScript rendering
func (h *AgencyHandler) GetAgencyRolesHTML(c *gin.Context) {
agencyID := c.Param("id")

// Get agency specification which contains roles
spec, err := h.service.GetSpecification(c.Request.Context(), agencyID)
if err != nil {
h.logger.WithError(err).Error("Failed to get specification")
c.String(http.StatusInternalServerError, "Error loading roles")
return
}

// Convert to pointers for template
rolePtrs := make([]*models.Role, len(spec.Roles))
for i := range spec.Roles {
rolePtrs[i] = &spec.Roles[i]
}

// Sort roles by Code
sort.Slice(rolePtrs, func(i, j int) bool {
return rolePtrs[i].Code < rolePtrs[j].Code
})

// Render the roles list template
component := agency_designer.AgencyRolesList(rolePtrs)
c.Header("Content-Type", "text/html")
component.Render(c.Request.Context(), c.Writer)
}

// CreateAgencyRole handles POST /api/v1/agencies/:id/roles
func (h *AgencyHandler) CreateAgencyRole(c *gin.Context) {
agencyID := c.Param("id")

var req models.CreateRoleRequest
if err := c.ShouldBindJSON(&req); err != nil {
h.logger.WithError(err).Error("Failed to bind role data")
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role data"})
return
}

// Get current specification
spec, err := h.service.GetSpecification(c.Request.Context(), agencyID)
if err != nil {
h.logger.WithError(err).Error("Failed to get specification", "agency_id", agencyID)
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

// Create new role
newRole := models.Role{
AgencyID:      agencyID,
Code:          req.Code,
Name:          req.Name,
Description:   req.Description,
Tags:          req.Tags,
AutonomyLevel: req.AutonomyLevel,
TokenBudget:   req.TokenBudget,
Icon:          req.Icon,
Color:         req.Color,
IsActive:      req.IsActive,
}

// Append to existing roles
updatedRoles := append(spec.Roles, newRole)

// Update specification with new roles list
_, err = h.service.UpdateSpecificationRoles(c.Request.Context(), agencyID, updatedRoles, "api")
if err != nil {
h.logger.WithError(err).Error("Failed to create role", "agency_id", agencyID)
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

h.logger.Info("Role created", "agency_id", agencyID, "role_code", newRole.Code)
c.JSON(http.StatusCreated, newRole)
}

// GetAgencyRole handles GET /api/v1/agencies/:id/roles/:key
func (h *AgencyHandler) GetAgencyRole(c *gin.Context) {
agencyID := c.Param("id")
key := c.Param("key")

// Get specification
spec, err := h.service.GetSpecification(c.Request.Context(), agencyID)
if err != nil {
h.logger.WithError(err).Error("Failed to get specification", "key", key)
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

// Find role by key
for _, role := range spec.Roles {
if role.Key == key {
c.JSON(http.StatusOK, role)
return
}
}

c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
}

// UpdateAgencyRole handles PUT /api/v1/agencies/:id/roles/:key
func (h *AgencyHandler) UpdateAgencyRole(c *gin.Context) {
agencyID := c.Param("id")
key := c.Param("key")

var req models.UpdateRoleRequest
if err := c.ShouldBindJSON(&req); err != nil {
h.logger.WithError(err).Error("Failed to bind role data")
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role data"})
return
}

// Get current specification
spec, err := h.service.GetSpecification(c.Request.Context(), agencyID)
if err != nil {
h.logger.WithError(err).Error("Failed to get specification", "key", key)
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

// Find and update role
found := false
for i := range spec.Roles {
if spec.Roles[i].Key == key {
spec.Roles[i].Code = req.Code
spec.Roles[i].Name = req.Name
spec.Roles[i].Description = req.Description
spec.Roles[i].Tags = req.Tags
spec.Roles[i].AutonomyLevel = req.AutonomyLevel
spec.Roles[i].TokenBudget = req.TokenBudget
spec.Roles[i].Icon = req.Icon
spec.Roles[i].Color = req.Color
spec.Roles[i].IsActive = req.IsActive
found = true
break
}
}

if !found {
c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
return
}

// Update specification
_, err = h.service.UpdateSpecificationRoles(c.Request.Context(), agencyID, spec.Roles, "api")
if err != nil {
h.logger.WithError(err).Error("Failed to update role", "agency_id", agencyID, "key", key)
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

h.logger.Info("Role updated", "agency_id", agencyID, "role_key", key)
c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

// DeleteAgencyRole handles DELETE /api/v1/agencies/:id/roles/:key
func (h *AgencyHandler) DeleteAgencyRole(c *gin.Context) {
agencyID := c.Param("id")
key := c.Param("key")

// Get current specification
spec, err := h.service.GetSpecification(c.Request.Context(), agencyID)
if err != nil {
h.logger.WithError(err).Error("Failed to get specification", "key", key)
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

// Find and remove role
newRoles := make([]models.Role, 0, len(spec.Roles))
found := false
for _, role := range spec.Roles {
if role.Key == key {
found = true
continue
}
newRoles = append(newRoles, role)
}

if !found {
c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
return
}

// Update specification
_, err = h.service.UpdateSpecificationRoles(c.Request.Context(), agencyID, newRoles, "api")
if err != nil {
h.logger.WithError(err).Error("Failed to delete role", "agency_id", agencyID, "key", key)
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

h.logger.Info("Role deleted", "agency_id", agencyID, "role_key", key)
c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}
