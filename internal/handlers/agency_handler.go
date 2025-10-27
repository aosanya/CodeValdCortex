package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/gin-gonic/gin"
)

// AgencyHandler handles agency-related HTTP requests
type AgencyHandler struct {
	service agency.Service
}

// NewAgencyHandler creates a new agency handler
func NewAgencyHandler(service agency.Service) *AgencyHandler {
	return &AgencyHandler{
		service: service,
	}
}

// RegisterRoutes registers agency routes with the router
func (h *AgencyHandler) RegisterRoutes(router *gin.RouterGroup) {
	agencies := router.Group("/agencies")
	{
		agencies.GET("", h.ListAgencies)
		agencies.POST("", h.CreateAgency)
		agencies.GET("/:id", h.GetAgency)
		agencies.PUT("/:id", h.UpdateAgency)
		agencies.DELETE("/:id", h.DeleteAgency)
		agencies.POST("/:id/activate", h.ActivateAgency)
		agencies.GET("/active", h.GetActiveAgency)
		agencies.GET("/:id/statistics", h.GetAgencyStatistics)
		agencies.GET("/:id/overview", h.GetOverview)
		agencies.PUT("/:id/overview", h.UpdateOverview)
	}
}

// CreateAgency handles POST /api/v1/agencies
func (h *AgencyHandler) CreateAgency(c *gin.Context) {
	var req agency.CreateAgencyRequest
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
		settings = agency.AgencySettings{
			AutoStart:         false,
			MonitoringEnabled: true,
			DashboardEnabled:  true,
			VisualizerEnabled: true,
		}
	}

	newAgency := &agency.Agency{
		ID:          req.ID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Category:    req.Category,
		Icon:        icon,
		Status:      agency.AgencyStatusActive,
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
func hasSettings(settings agency.AgencySettings) bool {
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
	filters := agency.AgencyFilters{
		Category: c.Query("category"),
		Status:   agency.AgencyStatus(c.Query("status")),
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

	var req agency.UpdateAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updates := agency.AgencyUpdates{
		DisplayName: req.DisplayName,
		Description: req.Description,
		Category:    req.Category,
		Icon:        req.Icon,
		Status:      req.Status,
		Metadata:    req.Metadata,
		Settings:    req.Settings,
	}

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

// GetOverview handles GET /api/v1/agencies/:id/overview
func (h *AgencyHandler) GetOverview(c *gin.Context) {
	id := c.Param("id")

	overview, err := h.service.GetAgencyOverview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// UpdateOverview handles PUT /api/v1/agencies/:id/overview
func (h *AgencyHandler) UpdateOverview(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Introduction string `json:"introduction" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := h.service.UpdateAgencyOverview(c.Request.Context(), id, req.Introduction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated overview
	overview, err := h.service.GetAgencyOverview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}
