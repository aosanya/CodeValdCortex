package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ActivationHandler handles agency activation/deactivation HTTP requests
type ActivationHandler struct {
	activationSvc services.ActivationService
	logger        *log.Logger
}

// NewActivationHandler creates a new activation handler
func NewActivationHandler(activationSvc services.ActivationService, logger *log.Logger) *ActivationHandler {
	return &ActivationHandler{
		activationSvc: activationSvc,
		logger:        logger,
	}
}

// PauseAgency pauses an active agency
// POST /api/v1/agencies/:id/pause
func (h *ActivationHandler) PauseAgency(c *gin.Context) {
	agencyID := c.Param("id")

	if err := h.activationSvc.PauseAgency(c.Request.Context(), agencyID); err != nil {
		h.logger.WithError(err).WithField("agency_id", agencyID).Error("failed to pause agency")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to pause agency",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithField("agency_id", agencyID).Info("agency paused successfully")
	c.JSON(http.StatusOK, gin.H{
		"message":   "agency paused successfully",
		"agency_id": agencyID,
	})
}

// ResumeAgency resumes a paused agency
// POST /api/v1/agencies/:id/resume
func (h *ActivationHandler) ResumeAgency(c *gin.Context) {
	agencyID := c.Param("id")

	if err := h.activationSvc.ResumeAgency(c.Request.Context(), agencyID); err != nil {
		h.logger.WithError(err).WithField("agency_id", agencyID).Error("failed to resume agency")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to resume agency",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithField("agency_id", agencyID).Info("agency resumed successfully")
	c.JSON(http.StatusOK, gin.H{
		"message":   "agency resumed successfully",
		"agency_id": agencyID,
	})
}

// DrainAgency gracefully drains an agency (complete work, no new tasks)
// POST /api/v1/agencies/:id/drain
func (h *ActivationHandler) DrainAgency(c *gin.Context) {
	agencyID := c.Param("id")

	if err := h.activationSvc.DrainAgency(c.Request.Context(), agencyID); err != nil {
		h.logger.WithError(err).WithField("agency_id", agencyID).Error("failed to drain agency")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to drain agency",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithField("agency_id", agencyID).Info("agency drained successfully")
	c.JSON(http.StatusOK, gin.H{
		"message":   "agency drained successfully",
		"agency_id": agencyID,
	})
}

// StopAgency stops an agency
// POST /api/v1/agencies/:id/stop
type StopAgencyRequest struct {
	Force bool `json:"force" binding:"omitempty"`
}

func (h *ActivationHandler) StopAgency(c *gin.Context) {
	agencyID := c.Param("id")

	var req StopAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to graceful stop if no body provided
		req.Force = false
	}

	if err := h.activationSvc.StopAgency(c.Request.Context(), agencyID, req.Force); err != nil {
		h.logger.WithError(err).WithField("agency_id", agencyID).Error("failed to stop agency")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to stop agency",
			"details": err.Error(),
		})
		return
	}

	h.logger.WithFields(log.Fields{
		"agency_id": agencyID,
		"force":     req.Force,
	}).Info("agency stopped successfully")

	c.JSON(http.StatusOK, gin.H{
		"message":   "agency stopped successfully",
		"agency_id": agencyID,
		"force":     req.Force,
	})
}
