package ai_refine

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RefineGoals handles POST /api/v1/agencies/:id/goals/refine-dynamic
// Dynamically determines and executes the appropriate goal operation based on user message
func (h *Handler) RefineGoals(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithField("agency_id", agencyID).Info("Processing dynamic AI goal refinement request")

	// Check if this is a wrapper call with a preset request
	var req struct {
		UserMessage string   `json:"user_message" binding:"required"` // Natural language instruction
		GoalKeys    []string `json:"goal_keys"`                       // Optional: specific goals to operate on
	}

	// First, check if there's a preset request from wrapper methods
	if dynamicReq, exists := c.Get("dynamic_request"); exists {
		if presetReq, ok := dynamicReq.(struct {
			UserMessage string   `json:"user_message"`
			GoalKeys    []string `json:"goal_keys"`
		}); ok {
			req.UserMessage = presetReq.UserMessage
			req.GoalKeys = presetReq.GoalKeys
			h.logger.WithField("source", "wrapper").Info("Using preset request from wrapper method")
		}
	} else {
		// Parse request body for direct calls
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.WithError(err).Error("Failed to parse dynamic refinement request")
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
				<div class="notification is-danger">
					<div class="is-flex is-align-items-center">
						<span class="icon has-text-danger mr-2">
							<i class="fas fa-exclamation-circle"></i>
						</span>
						<div>
							<strong>Invalid Request</strong>
							<p class="mb-0">Please provide a user message describing what you want to do with the goals.</p>
						</div>
					</div>
				</div>
			`)
			return
		}
	}

	// Get agency context
	ag, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch agency")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusNotFound, `
			<div class="notification is-warning">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-warning mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>Agency Not Found</strong>
						<p class="mb-0">The requested agency could not be found.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Get unified specification (replaces separate GetGoals, GetWorkItems, GetOverview calls)
	spec, err := h.agencyService.GetSpecification(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch specification")
		spec = &models.AgencySpecification{
			Introduction: "",
			Goals:        []models.Goal{},
			WorkItems:    []models.WorkItem{},
		}
	}

	// Convert goals from []Goal to []*Goal for compatibility
	existingGoals := make([]*models.Goal, len(spec.Goals))
	for i := range spec.Goals {
		existingGoals[i] = &spec.Goals[i]
	}

	// Filter target goals if specific keys were provided
	var targetGoals []*models.Goal
	if len(req.GoalKeys) > 0 {
		goalKeyMap := make(map[string]bool)
		for _, key := range req.GoalKeys {
			goalKeyMap[key] = true
		}
		for _, goal := range existingGoals {
			if goalKeyMap[goal.Key] {
				targetGoals = append(targetGoals, goal)
			}
		}
		h.logger.WithFields(logrus.Fields{
			"requested_keys": len(req.GoalKeys),
			"found_goals":    len(targetGoals),
		}).Info("Filtered target goals")
	}

	// Convert work items from []WorkItem to []*WorkItem for compatibility
	workItems := make([]*models.WorkItem, len(spec.WorkItems))
	for i := range spec.WorkItems {
		workItems[i] = &spec.WorkItems[i]
	}

	// Build the AI builder context
	builderContext, err := h.contextBuilder.BuildBuilderContext(
		c.Request.Context(),
		ag,
		spec.Introduction,
		req.UserMessage,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build context")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Context Error</strong>
						<p class="mb-0">Failed to gather agency context for AI processing.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Create the refinement request
	refineReq := &builder.RefineGoalsRequest{
		AgencyID:      agencyID,
		UserMessage:   req.UserMessage,
		TargetGoals:   targetGoals,
		ExistingGoals: existingGoals,
		WorkItems:     workItems,
		AgencyContext: ag,
	}

	// Call the AI service
	result, err := h.goalRefiner.RefineGoals(c.Request.Context(), refineReq, builderContext)
	if err != nil {
		h.logger.WithError(err).Error("AI refinement failed")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>AI Processing Failed</strong>
						<p class="mb-0">The AI service encountered an error processing your request.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"action":           result.Action,
		"refined_count":    len(result.RefinedGoals),
		"generated_count":  len(result.GeneratedGoals),
		"no_action_needed": result.NoActionNeeded,
		"has_consolidated": result.ConsolidatedData != nil,
	}).Info("Dynamic goal refinement completed")

	// Return the result as JSON
	c.JSON(http.StatusOK, gin.H{
		"action":            result.Action,
		"refined_goals":     result.RefinedGoals,
		"generated_goals":   result.GeneratedGoals,
		"consolidated_data": result.ConsolidatedData,
		"explanation":       result.Explanation,
		"no_action_needed":  result.NoActionNeeded,
		"summary":           h.buildSummaryMessage(result),
	})
}

// buildSummaryMessage creates a user-friendly summary of what was done
func (h *Handler) buildSummaryMessage(result *builder.RefineGoalsResponse) string {
	if result.NoActionNeeded {
		return "✓ No changes needed - your goals are already well-defined and strategically aligned."
	}

	var parts []string

	switch result.Action {
	case "refine":
		changedCount := 0
		for _, rg := range result.RefinedGoals {
			if rg.WasChanged {
				changedCount++
			}
		}
		if changedCount > 0 {
			parts = append(parts, "✓ Refined "+pluralize(changedCount, "goal", "goals"))
		} else {
			parts = append(parts, "✓ Reviewed goals - no changes needed")
		}

	case "generate":
		if len(result.GeneratedGoals) > 0 {
			parts = append(parts, "✓ Generated "+pluralize(len(result.GeneratedGoals), "new goal", "new goals"))
		}

	case "consolidate":
		if result.ConsolidatedData != nil {
			consolidated := len(result.ConsolidatedData.ConsolidatedGoals)
			removed := len(result.ConsolidatedData.RemovedGoals)
			if consolidated > 0 {
				parts = append(parts, "✓ Consolidated into "+pluralize(consolidated, "goal", "goals"))
			}
			if removed > 0 {
				parts = append(parts, "✓ Removed "+pluralize(removed, "duplicate", "duplicates"))
			}
		}

	case "enhance_all":
		changedCount := 0
		for _, rg := range result.RefinedGoals {
			if rg.WasChanged {
				changedCount++
			}
		}
		if changedCount > 0 {
			parts = append(parts, "✓ Enhanced "+pluralize(changedCount, "goal", "goals"))
		} else {
			parts = append(parts, "✓ Reviewed all goals - no changes needed")
		}
	}

	if len(parts) == 0 {
		return "✓ Processing completed"
	}

	return strings.Join(parts, " • ")
}

// pluralize returns singular or plural form based on count
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d %s", count, plural)
}
