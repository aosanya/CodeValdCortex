package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AgencyHandler handles agency-related HTTP requests
type AgencyHandler struct {
	service agency.Service
	logger  *logrus.Logger
}

// NewAgencyHandler creates a new agency handler
func NewAgencyHandler(service agency.Service, logger *logrus.Logger) *AgencyHandler {
	return &AgencyHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers agency routes with the router
func (h *AgencyHandler) RegisterRoutes(router *gin.RouterGroup) {
	agencies := router.Group("/agencies")
	{
		// Core agency routes
		agencies.GET("", h.ListAgencies)
		agencies.POST("", h.CreateAgency)
		agencies.GET("/:id", h.GetAgency)
		agencies.PUT("/:id", h.UpdateAgency)
		agencies.DELETE("/:id", h.DeleteAgency)
		agencies.POST("/:id/activate", h.ActivateAgency)
		agencies.GET("/active", h.GetActiveAgency)
		agencies.GET("/:id/statistics", h.GetAgencyStatistics)

		// Unified Specification routes (replaces separate overview/goals/work-items)
		agencies.GET("/:id/specification", h.GetSpecification)
		agencies.PUT("/:id/specification", h.UpdateSpecification)
		agencies.PUT("/:id/specification/introduction", h.UpdateIntroduction)
		agencies.PUT("/:id/specification/goals", h.UpdateGoals)
		agencies.PUT("/:id/specification/work-items", h.UpdateWorkItems)
		agencies.PUT("/:id/specification/roles", h.UpdateRoles)
		agencies.PUT("/:id/specification/raci-matrix", h.UpdateRACIMatrixSection)

		// RACI Matrix CRUD endpoints
		agencies.GET("/:id/raci-matrix", h.GetRACIMatrix)
		agencies.POST("/:id/raci-matrix", h.SaveRACIMatrix)
	}
}

// CreateAgency handles POST /api/v1/agencies
func (h *AgencyHandler) CreateAgency(c *gin.Context) {
	var req models.CreateAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Sanitize ID - ensure it has the "agency_" prefix and remove hyphens from UUID part
	if !strings.HasPrefix(req.ID, "agency_") {
		// If no prefix, add it
		req.ID = "agency_" + strings.ReplaceAll(req.ID, "-", "")
	} else {
		// If prefix exists, just remove hyphens from the UUID part
		parts := strings.SplitN(req.ID, "_", 2)
		if len(parts) == 2 {
			req.ID = parts[0] + "_" + strings.ReplaceAll(parts[1], "-", "")
		}
	}

	// Set default icon based on category if not provided
	icon := req.Icon
	if icon == "" {
		icon = getCategoryIcon(req.Category)
	}

	// Set default metadata values
	metadata := req.Metadata
	if metadata.APIEndpoint == "" {
		metadata.APIEndpoint = fmt.Sprintf("/api/v1/agencies/%s", req.ID)
	}

	// Set default settings
	settings := req.Settings
	if !hasSettings(req.Settings) {
		settings = models.AgencySettings{
			AutoStart:         false,
			MonitoringEnabled: true,
			DashboardEnabled:  true,
			VisualizerEnabled: true,
		}
	}

	newAgency := &models.Agency{
		ID:          req.ID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Category:    req.Category,
		Icon:        icon,
		Status:      models.AgencyStatusActive,
		// Database field will be set by service with proper prefix
		Metadata:  metadata,
		Settings:  settings,
		CreatedBy: "system", // TODO: Get from auth context
	}

	if err := h.service.CreateAgency(c.Request.Context(), newAgency); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newAgency)
}

// hasSettings checks if settings have been provided
func hasSettings(settings models.AgencySettings) bool {
	return settings.AutoStart || settings.MonitoringEnabled || settings.DashboardEnabled || settings.VisualizerEnabled
}

// getCategoryIcon returns the default icon for a category
func getCategoryIcon(category string) string {
	categoryIcons := map[string]string{
		"infrastructure": "🏗️",
		"agriculture":    "🌾",
		"logistics":      "📦",
		"transportation": "🚗",
		"healthcare":     "🏥",
		"education":      "🎓",
		"finance":        "💰",
		"retail":         "🛒",
		"energy":         "⚡",
		"other":          "📋",
	}

	icon, ok := categoryIcons[category]
	if !ok {
		return "📋"
	}
	return icon
}

// GetAgency handles GET /api/v1/agencies/:id
func (h *AgencyHandler) GetAgency(c *gin.Context) {
	id := c.Param("id")

	agency, err := h.service.GetAgency(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	c.JSON(http.StatusOK, agency)
}

// ListAgencies handles GET /api/v1/agencies
func (h *AgencyHandler) ListAgencies(c *gin.Context) {
	// Parse query parameters
	filters := models.AgencyFilters{
		Category: c.Query("category"),
		Status:   models.AgencyStatus(c.Query("status")),
		Search:   c.Query("search"),
	}

	// Parse limit and offset
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil {
			filters.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil {
			filters.Offset = offset
		}
	}

	agencies, err := h.service.ListAgencies(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agencies)
}

// UpdateAgency handles PUT /api/v1/agencies/:id
func (h *AgencyHandler) UpdateAgency(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updates := models.AgencyUpdates(req)

	if err := h.service.UpdateAgency(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated agency
	updated, err := h.service.GetAgency(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteAgency handles DELETE /api/v1/agencies/:id
func (h *AgencyHandler) DeleteAgency(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteAgency(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ActivateAgency handles POST /api/v1/agencies/:id/activate
func (h *AgencyHandler) ActivateAgency(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.SetActiveAgency(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	agency, err := h.service.GetAgency(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agency)
}

// GetActiveAgency handles GET /api/v1/agencies/active
func (h *AgencyHandler) GetActiveAgency(c *gin.Context) {
	agency, err := h.service.GetActiveAgency(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active agency set"})
		return
	}

	c.JSON(http.StatusOK, agency)
}

// GetAgencyStatistics handles GET /api/v1/agencies/:id/statistics
func (h *AgencyHandler) GetAgencyStatistics(c *gin.Context) {
	id := c.Param("id")

	stats, err := h.service.GetAgencyStatistics(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSpecification handles GET /api/v1/agencies/:id/specification
func (h *AgencyHandler) GetSpecification(c *gin.Context) {
	id := c.Param("id")

	spec, err := h.service.GetSpecification(c.Request.Context(), id)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": id,
			"error":     err.Error(),
			"method":    "GetSpecification",
		}).Error("Failed to get specification")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// UpdateSpecification handles PUT /api/v1/agencies/:id/specification
func (h *AgencyHandler) UpdateSpecification(c *gin.Context) {
	id := c.Param("id")

	var req models.SpecificationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	spec, err := h.service.UpdateSpecification(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// UpdateIntroduction handles PUT /api/v1/agencies/:id/specification/introduction
func (h *AgencyHandler) UpdateIntroduction(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Introduction string `json:"introduction"`
		UpdatedBy    string `json:"updated_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	spec, err := h.service.UpdateIntroduction(c.Request.Context(), id, req.Introduction, req.UpdatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// UpdateGoals handles PUT /api/v1/agencies/:id/specification/goals
func (h *AgencyHandler) UpdateGoals(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Goals     []models.Goal `json:"goals"`
		UpdatedBy string        `json:"updated_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	spec, err := h.service.UpdateSpecificationGoals(c.Request.Context(), id, req.Goals, req.UpdatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// UpdateWorkItems handles PUT /api/v1/agencies/:id/specification/work-items
func (h *AgencyHandler) UpdateWorkItems(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		WorkItems []models.WorkItem `json:"work_items"`
		UpdatedBy string            `json:"updated_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	spec, err := h.service.UpdateSpecificationWorkItems(c.Request.Context(), id, req.WorkItems, req.UpdatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// UpdateRoles handles PUT /api/v1/agencies/:id/specification/roles
func (h *AgencyHandler) UpdateRoles(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Roles     []models.Role `json:"roles"`
		UpdatedBy string        `json:"updated_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	spec, err := h.service.UpdateSpecificationRoles(c.Request.Context(), id, req.Roles, req.UpdatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// UpdateRACIMatrixSection handles PUT /api/v1/agencies/:id/specification/raci-matrix
func (h *AgencyHandler) UpdateRACIMatrixSection(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		RACIMatrix *models.RACIMatrix `json:"raci_matrix"`
		UpdatedBy  string             `json:"updated_by"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	spec, err := h.service.UpdateSpecificationRACIMatrix(c.Request.Context(), id, req.RACIMatrix, req.UpdatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// GetRACIMatrix handles GET /api/v1/agencies/:id/raci-matrix
func (h *AgencyHandler) GetRACIMatrix(c *gin.Context) {
	id := c.Param("id")

	spec, err := h.service.GetSpecification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if spec == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency specification not found"})
		return
	}

	// Extract assignments from the RACI matrix if it exists
	assignments := make(map[string]interface{})
	if spec.RACIMatrix != nil && len(spec.RACIMatrix.Activities) > 0 {
		// Convert the RACI matrix activities to the format expected by JavaScript
		for _, activity := range spec.RACIMatrix.Activities {
			if len(activity.Assignments) > 0 {
				// Convert to JavaScript format: map[roleKey]RACIRole -> map[roleKey]builder.RACIAssignment
				jsAssignments := make(map[string]builder.RACIAssignment)
				for roleKey, raciRole := range activity.Assignments {
					jsAssignments[roleKey] = builder.RACIAssignment{
						RACI:      string(raciRole),
						Objective: "", // TODO: Store objectives in the model
					}
				}
				// Use activity ID as the work item key
				assignments[activity.ID] = jsAssignments
			}
		}
	}

	// Return the RACI assignments in the format expected by JavaScript
	response := gin.H{
		"assignments": assignments,
	}

	c.JSON(http.StatusOK, response)
}

// SaveRACIMatrix handles POST /api/v1/agencies/:id/raci-matrix
func (h *AgencyHandler) SaveRACIMatrix(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Assignments map[string]map[string]builder.RACIAssignment `json:"assignments"`
		UpdatedBy   string                                       `json:"updated_by,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Convert the assignments to RACIMatrix format
	activities := make([]models.RACIActivity, 0)
	for workItemKey, roleAssignments := range req.Assignments {
		// Convert the JavaScript format to models.RACIRole format
		modelAssignments := make(map[string]models.RACIRole)
		for roleKey, assignment := range roleAssignments {
			modelAssignments[roleKey] = models.RACIRole(assignment.RACI)
		}

		activity := models.RACIActivity{
			ID:          workItemKey,
			Name:        workItemKey, // Use work item key as name for now
			Assignments: modelAssignments,
		}
		activities = append(activities, activity)
	}

	raciMatrix := &models.RACIMatrix{
		AgencyID:   id,
		Name:       "RACI Matrix",
		Activities: activities,
		IsValid:    true, // TODO: Add validation
	}

	// Use default user if not provided
	updatedBy := req.UpdatedBy
	if updatedBy == "" {
		updatedBy = "system"
	}

	spec, err := h.service.UpdateSpecificationRACIMatrix(c.Request.Context(), id, raciMatrix, updatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":     "RACI matrix saved successfully",
		"raci_matrix": spec.RACIMatrix,
	})
}
