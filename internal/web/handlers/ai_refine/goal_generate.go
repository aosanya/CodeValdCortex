package ai_refine

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/gin-gonic/gin"
)

// GenerateGoal handles POST /api/v1/agencies/:id/goals/generate
// Generates a new goal using AI based on user input
func (h *Handler) GenerateGoal(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse request body to get user input
	var req struct {
		UserInput string `json:"userInput" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse generate goal request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate agency exists and get context
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		h.logger.Error("Agency not found", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	// Get existing goals for context
	existingGoals, err := h.agencyService.GetGoals(ctx, agencyID)
	if err != nil {
		h.logger.Error("Failed to get existing goals", "agencyID", agencyID, "error", err)
		existingGoals = []*agency.Goal{} // Continue with empty list
	}

	// Get units of work for context
	unitsOfWork, err := h.agencyService.GetUnitsOfWork(ctx, agencyID)
	if err != nil {
		h.logger.Error("Failed to get units of work", "agencyID", agencyID, "error", err)
		unitsOfWork = []*agency.UnitOfWork{} // Continue with empty list
	}

	// Build generation request
	genReq := &ai.GenerateGoalRequest{
		AgencyID:      agencyID,
		AgencyContext: ag,
		ExistingGoals: existingGoals,
		UnitsOfWork:   unitsOfWork,
		UserInput:     req.UserInput,
	}

	// Generate goal using AI
	result, err := h.goalRefiner.GenerateGoal(ctx, genReq)
	if err != nil {
		h.logger.Error("Failed to generate goal", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate goal"})
		return
	}

	// Save the generated goal to database
	createdGoal, err := h.agencyService.CreateGoal(ctx, agencyID, result.SuggestedCode, result.Description)
	if err != nil {
		h.logger.Error("Failed to save generated goal", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save generated goal"})
		return
	}

	h.logger.Info("Goal generated successfully",
		"agencyID", agencyID,
		"goalKey", createdGoal.Key,
		"code", createdGoal.Code)

	// Return the generated goal data as JSON for HTMX to handle
	c.JSON(http.StatusOK, gin.H{
		"goal":               createdGoal,
		"scope":              result.Scope,
		"success_metrics":    result.SuccessMetrics,
		"suggested_priority": result.SuggestedPriority,
		"suggested_category": result.SuggestedCategory,
		"suggested_tags":     result.SuggestedTags,
		"explanation":        result.Explanation,
		"success":            true,
	})
}
