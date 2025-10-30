package ai_refine

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ConsolidateGoals handles POST /api/v1/agencies/:id/goals/consolidate
// Consolidates goals into a lean, strategic list
func (h *Handler) ConsolidateGoals(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithField("agency_id", agencyID).Info("Processing goal consolidation request")

	// Get agency context
	ag, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency for consolidation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agency"})
		return
	}

	// Get current goals
	goals, err := h.agencyService.GetGoals(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get goals for consolidation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get goals"})
		return
	}

	if len(goals) < 5 {
		h.logger.Info("Too few goals to consolidate", "count", len(goals))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Need at least 5 goals to consolidate"})
		return
	}

	// Get units of work for context
	units, err := h.agencyService.GetUnitsOfWork(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get units of work, continuing without them")
		units = []*agency.UnitOfWork{}
	}

	// Perform consolidation
	consolidationReq := &ai.ConsolidateGoalsRequest{
		AgencyID:      agencyID,
		AgencyContext: ag,
		CurrentGoals:  goals,
		UnitsOfWork:   units,
	}

	result, err := h.goalConsolidator.ConsolidateGoals(c.Request.Context(), consolidationReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to consolidate goals")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to consolidate goals"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":          agencyID,
		"original_count":     len(goals),
		"consolidated_count": len(result.ConsolidatedGoals),
		"removed_count":      len(result.RemovedGoals),
	}).Info("Goal consolidation completed successfully")

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"original_count":     len(goals),
		"consolidated_count": len(result.ConsolidatedGoals),
		"removed_count":      len(result.RemovedGoals),
		"consolidated_goals": result.ConsolidatedGoals,
		"removed_goals":      result.RemovedGoals,
		"summary":            result.Summary,
		"explanation":        result.Explanation,
	})
}
