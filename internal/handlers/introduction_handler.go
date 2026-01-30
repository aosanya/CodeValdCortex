package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// IntroductionHandler handles introduction-related HTTP requests
type IntroductionHandler struct {
	service *service.IntroductionService
	logger  *logrus.Logger
}

// NewIntroductionHandler creates a new introduction handler
func NewIntroductionHandler(service *service.IntroductionService, logger *logrus.Logger) *IntroductionHandler {
	return &IntroductionHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers introduction routes with the router
func (h *IntroductionHandler) RegisterRoutes(router *gin.RouterGroup) {
	introductions := router.Group("/agencies/:agency_id/introduction")
	{
		// Core introduction routes
		introductions.GET("", h.GetIntroduction)
		introductions.POST("", h.CreateIntroduction)
		introductions.PUT("", h.UpdateIntroduction)
		introductions.DELETE("", h.DeleteIntroduction)

		// Section management
		introductions.POST("/sections", h.AddSection)
		introductions.PUT("/sections/:section_id", h.UpdateSection)
		introductions.DELETE("/sections/:section_id", h.DeleteSection)
		introductions.PUT("/sections/reorder", h.ReorderSections)
		introductions.GET("/sections/:section_id", h.GetSection)

		// Template operations
		introductions.POST("/apply-template", h.ApplyTemplate)
		introductions.GET("/templates", h.ListTemplates)
	}

	// AI generation routes (global)
	ai := router.Group("/introduction/ai")
	{
		ai.POST("/generate", h.GenerateIntroduction)
		ai.POST("/refine-section", h.RefineSection)
	}
}

// GetIntroduction handles GET /api/v1/agencies/:agency_id/introduction
func (h *IntroductionHandler) GetIntroduction(c *gin.Context) {
	agencyID := c.Param("agency_id")

	intro, err := h.service.GetByAgencyID(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Failed to get introduction")
		c.JSON(http.StatusNotFound, gin.H{"error": "Introduction not found"})
		return
	}

	c.JSON(http.StatusOK, intro)
}

// CreateIntroduction handles POST /api/v1/agencies/:agency_id/introduction
func (h *IntroductionHandler) CreateIntroduction(c *gin.Context) {
	agencyID := c.Param("agency_id")

	var intro models.AgencyIntroduction
	if err := c.ShouldBindJSON(&intro); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	intro.AgencyID = agencyID

	if err := h.service.Create(c.Request.Context(), &intro); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Failed to create introduction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, intro)
}

// UpdateIntroduction handles PUT /api/v1/agencies/:agency_id/introduction
func (h *IntroductionHandler) UpdateIntroduction(c *gin.Context) {
	agencyID := c.Param("agency_id")

	var intro models.AgencyIntroduction
	if err := c.ShouldBindJSON(&intro); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	intro.AgencyID = agencyID

	if err := h.service.Update(c.Request.Context(), &intro); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Failed to update introduction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, intro)
}

// DeleteIntroduction handles DELETE /api/v1/agencies/:agency_id/introduction
func (h *IntroductionHandler) DeleteIntroduction(c *gin.Context) {
	agencyID := c.Param("agency_id")

	if err := h.service.Delete(c.Request.Context(), agencyID); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Failed to delete introduction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AddSection handles POST /api/v1/agencies/:agency_id/introduction/sections
func (h *IntroductionHandler) AddSection(c *gin.Context) {
	agencyID := c.Param("agency_id")

	var section models.IntroductionSection
	if err := c.ShouldBindJSON(&section); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddSection(c.Request.Context(), agencyID, &section); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":  agencyID,
			"section_id": section.ID,
			"error":      err.Error(),
		}).Error("Failed to add section")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, section)
}

// UpdateSection handles PUT /api/v1/agencies/:agency_id/introduction/sections/:section_id
func (h *IntroductionHandler) UpdateSection(c *gin.Context) {
	agencyID := c.Param("agency_id")
	sectionID := c.Param("section_id")

	var section models.IntroductionSection
	if err := c.ShouldBindJSON(&section); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":  agencyID,
			"section_id": sectionID,
			"error":      err.Error(),
		}).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateSection(c.Request.Context(), agencyID, sectionID, &section); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":  agencyID,
			"section_id": sectionID,
			"error":      err.Error(),
		}).Error("Failed to update section")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, section)
}

// DeleteSection handles DELETE /api/v1/agencies/:agency_id/introduction/sections/:section_id
func (h *IntroductionHandler) DeleteSection(c *gin.Context) {
	agencyID := c.Param("agency_id")
	sectionID := c.Param("section_id")

	if err := h.service.DeleteSection(c.Request.Context(), agencyID, sectionID); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":  agencyID,
			"section_id": sectionID,
			"error":      err.Error(),
		}).Error("Failed to delete section")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ReorderSectionsRequest represents the request body for reordering sections
type ReorderSectionsRequest struct {
	SectionIDs []string `json:"section_ids" binding:"required"`
}

// ReorderSections handles PUT /api/v1/agencies/:agency_id/introduction/sections/reorder
func (h *IntroductionHandler) ReorderSections(c *gin.Context) {
	agencyID := c.Param("agency_id")

	var req ReorderSectionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReorderSections(c.Request.Context(), agencyID, req.SectionIDs); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Failed to reorder sections")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sections reordered successfully"})
}

// GetSection handles GET /api/v1/agencies/:agency_id/introduction/sections/:section_id
func (h *IntroductionHandler) GetSection(c *gin.Context) {
	agencyID := c.Param("agency_id")
	sectionID := c.Param("section_id")

	section, err := h.service.GetSectionByID(c.Request.Context(), agencyID, sectionID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id":  agencyID,
			"section_id": sectionID,
			"error":      err.Error(),
		}).Error("Failed to get section")
		c.JSON(http.StatusNotFound, gin.H{"error": "Section not found"})
		return
	}

	c.JSON(http.StatusOK, section)
}

// ApplyTemplateRequest represents the request body for applying a template
type ApplyTemplateRequest struct {
	Template string `json:"template" binding:"required"`
}

// ApplyTemplate handles POST /api/v1/agencies/:agency_id/introduction/apply-template
func (h *IntroductionHandler) ApplyTemplate(c *gin.Context) {
	agencyID := c.Param("agency_id")

	var req ApplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"error":     err.Error(),
		}).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	intro, err := h.service.ApplyTemplate(c.Request.Context(), agencyID, req.Template)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"agency_id": agencyID,
			"template":  req.Template,
			"error":     err.Error(),
		}).Error("Failed to apply template")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, intro)
}

// ListTemplates handles GET /api/v1/agencies/:agency_id/introduction/templates
func (h *IntroductionHandler) ListTemplates(c *gin.Context) {
	templates := h.service.GetAvailableTemplates()
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// GenerateIntroductionRequest represents the request body for AI generation
type GenerateIntroductionRequest struct {
	AgencyID string   `json:"agency_id" binding:"required"`
	Keywords []string `json:"keywords" binding:"required"`
	Template string   `json:"template"`
}

// GenerateIntroduction handles POST /api/v1/introduction/ai/generate
func (h *IntroductionHandler) GenerateIntroduction(c *gin.Context) {
	var req GenerateIntroductionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement AI generation in Phase 4
	// For now, return a placeholder response
	h.logger.WithFields(logrus.Fields{
		"agency_id": req.AgencyID,
		"keywords":  req.Keywords,
		"template":  req.Template,
	}).Info("AI generation requested (not yet implemented)")

	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "AI generation will be implemented in Phase 4",
		"request": req,
	})
}

// RefineSectionRequest represents the request body for AI section refinement
type RefineSectionRequest struct {
	AgencyID   string                     `json:"agency_id" binding:"required"`
	SectionID  string                     `json:"section_id" binding:"required"`
	Section    models.IntroductionSection `json:"section" binding:"required"`
	Refinement string                     `json:"refinement" binding:"required"`
}

// RefineSection handles POST /api/v1/introduction/ai/refine-section
func (h *IntroductionHandler) RefineSection(c *gin.Context) {
	var req RefineSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement AI refinement in Phase 4
	// For now, return a placeholder response
	h.logger.WithFields(logrus.Fields{
		"agency_id":  req.AgencyID,
		"section_id": req.SectionID,
		"refinement": req.Refinement,
	}).Info("AI refinement requested (not yet implemented)")

	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "AI refinement will be implemented in Phase 4",
		"request": req,
	})
}
