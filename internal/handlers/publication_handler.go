package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/aosanya/CodeValdCortex/internal/agency/services"
)

// PublicationHandler handles publication-related HTTP requests
type PublicationHandler struct {
	pubService services.PublicationService
	logger     *logrus.Logger
}

// NewPublicationHandler creates a new publication handler instance
func NewPublicationHandler(pubService services.PublicationService, logger *logrus.Logger) *PublicationHandler {
	return &PublicationHandler{
		pubService: pubService,
		logger:     logger,
	}
}

// ValidateForPublish validates if an agency is ready for publication
// POST /api/v1/agencies/:id/validate
func (h *PublicationHandler) ValidateForPublish(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"endpoint":  "ValidateForPublish",
	}).Info("validating agency for publication")

	// Call validation service
	result, err := h.pubService.ValidateForPublish(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("validation service error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to validate agency",
			"details": err.Error(),
		})
		return
	}

	// Return validation result
	statusCode := http.StatusOK
	if !result.Valid {
		statusCode = http.StatusUnprocessableEntity
	}

	c.JSON(statusCode, gin.H{
		"valid":            result.Valid,
		"errors":           result.Errors,
		"warnings":         result.Warnings,
		"recommendations":  result.Recommendations,
	})
}

// Publish publishes an agency to create an immutable publication
// POST /api/v1/agencies/:id/publish
func (h *PublicationHandler) Publish(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"endpoint":  "Publish",
	}).Info("publishing agency")

	// Parse request body
	var req services.PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Call publish service
	publication, err := h.pubService.Publish(c.Request.Context(), agencyID, &req)
	if err != nil {
		h.logger.WithError(err).Error("publish service error")
		
		// Check for validation errors
		if err.Error() == "validation failed: 0 errors found" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "agency validation failed",
				"details": err.Error(),
			})
			return
		}
		
		// Check for duplicate version
		if err.Error() == "publication with version " + req.Version + " already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "version conflict",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to publish agency",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":      agencyID,
		"publication_id": publication.Key,
		"version":        publication.Version,
	}).Info("agency published successfully")

	c.JSON(http.StatusCreated, gin.H{
		"publication": publication,
		"message":     "agency published successfully",
	})
}

// Activate activates a published agency
// POST /api/v1/agencies/:id/activate
func (h *PublicationHandler) Activate(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"endpoint":  "Activate",
	}).Info("activating agency")

	// Get publication ID from query param or use latest
	publicationID := c.Query("publication_id")

	// If no publication ID, get latest publication for agency
	if publicationID == "" {
		publications, err := h.pubService.GetPublicationHistory(c.Request.Context(), agencyID)
		if err != nil || len(publications) == 0 {
			h.logger.WithError(err).Error("no publications found")
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no publications found for agency",
			})
			return
		}
		publicationID = publications[0].Key
	}

	// Call activate service
	result, err := h.pubService.Activate(c.Request.Context(), publicationID)
	if err != nil {
		h.logger.WithError(err).Error("activation service error")
		
		// Check for state errors
		if err.Error() == "agency must be in published state to activate" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "invalid state",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to activate agency",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":      agencyID,
		"publication_id": publicationID,
	}).Info("agency activated successfully")

	c.JSON(http.StatusOK, gin.H{
		"result":  result,
		"message": "agency activated successfully",
	})
}

// Deactivate deactivates an active agency
// POST /api/v1/agencies/:id/deactivate
func (h *PublicationHandler) Deactivate(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"endpoint":  "Deactivate",
	}).Info("deactivating agency")

	// Parse request body (optional - graceful flag)
	type DeactivateRequest struct {
		Graceful bool `json:"graceful"`
	}

	var req DeactivateRequest
	_ = c.ShouldBindJSON(&req) // Optional body

	// Call deactivate service
	err := h.pubService.Deactivate(c.Request.Context(), agencyID, req.Graceful)
	if err != nil {
		h.logger.WithError(err).Error("deactivation service error")
		
		// Check for state errors
		if err.Error() == "agency must be in active or paused state to deactivate" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "invalid state",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to deactivate agency",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithField("agency_id", agencyID).Info("agency deactivated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "agency deactivated successfully",
		"graceful": req.Graceful,
	})
}

// GetPublicationHistory retrieves publication history for an agency
// GET /api/v1/agencies/:id/publications
func (h *PublicationHandler) GetPublicationHistory(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"endpoint":  "GetPublicationHistory",
	}).Info("retrieving publication history")

	// Call service
	publications, err := h.pubService.GetPublicationHistory(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("failed to retrieve publication history")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve publication history",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id": agencyID,
		"count":     len(publications),
	}).Info("publication history retrieved")

	c.JSON(http.StatusOK, gin.H{
		"publications": publications,
		"count":        len(publications),
	})
}

// ActivatePublication activates a specific publication by ID
// POST /api/v1/publications/:id/activate
func (h *PublicationHandler) ActivatePublication(c *gin.Context) {
	publicationID := c.Param("id")

	h.logger.WithFields(logrus.Fields{
		"publication_id": publicationID,
		"endpoint":       "ActivatePublication",
	}).Info("activating publication")

	// Call activate service
	result, err := h.pubService.Activate(c.Request.Context(), publicationID)
	if err != nil {
		h.logger.WithError(err).Error("activation service error")
		
		// Check for state errors
		if err.Error() == "agency must be in published state to activate" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "invalid state",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to activate publication",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithField("publication_id", publicationID).Info("publication activated successfully")

	c.JSON(http.StatusOK, gin.H{
		"result":  result,
		"message": "publication activated successfully",
	})
}
