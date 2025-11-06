package ai_refine

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RefineSpecificGoal handles POST /api/v1/agencies/:id/goals/:goalKey/refine
// Wrapper that uses RefineGoals with a specific goal refinement prompt
func (h *Handler) RefineSpecificGoal(c *gin.Context) {
	goalKey := c.Param("goalKey")

	// Parse request body to get the current goal data
	var req struct {
		Description    string   `json:"description"`
		Scope          string   `json:"scope"`
		SuccessMetrics []string `json:"success_metrics"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse goal refinement request")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Invalid Request</strong>
						<p class="mb-0">Please provide goal description and details.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Create a dynamic request that specifically targets this goal
	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		GoalKeys    []string `json:"goal_keys"`
	}{
		UserMessage: "Refine and improve this specific goal to be clearer, more specific, and better aligned with the agency's purpose and introduction. Provide specific, measurable success metrics. Consider if the goal adequately covers its intended scope or if additional complementary goals might be needed.",
		GoalKeys:    []string{goalKey},
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineGoals handler
	h.RefineGoals(c)
}

// GenerateGoalWithPrompt handles POST /api/v1/agencies/:id/goals/generate
// Wrapper that uses RefineGoals with a goal generation prompt
func (h *Handler) GenerateGoalWithPrompt(c *gin.Context) {
	// Parse request body to get user input
	var req struct {
		UserInput string `json:"userInput" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse generate goal request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Create a dynamic request for goal generation
	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		GoalKeys    []string `json:"goal_keys"`
	}{
		UserMessage: "Generate one or more strategic goals based on this user request. Consider the agency's introduction and overall purpose to create comprehensive goals that cover the topic thoroughly: " + req.UserInput,
		GoalKeys:    []string{}, // Empty - we're creating new goals
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineGoals handler
	h.RefineGoals(c)
}

// ConsolidateGoalsWithPrompt handles POST /api/v1/agencies/:id/goals/consolidate
// Wrapper that uses RefineGoals with a consolidation prompt
func (h *Handler) ConsolidateGoalsWithPrompt(c *gin.Context) {
	// Create a dynamic request for goal consolidation
	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		GoalKeys    []string `json:"goal_keys"`
	}{
		UserMessage: "Consolidate and merge duplicate or overlapping goals into a lean, strategic list. Remove redundancy while preserving strategic value.",
		GoalKeys:    []string{}, // Empty - we're working with all goals
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineGoals handler
	h.RefineGoals(c)
}

// GenerateGoalsFromIntroduction handles comprehensive goal generation based on agency introduction
// This could be used as an additional endpoint for generating multiple goals from the introduction
func (h *Handler) GenerateGoalsFromIntroduction(c *gin.Context) {
	// Parse optional request body for additional context
	var req struct {
		AdditionalContext string `json:"additional_context"`
		GoalCount         int    `json:"goal_count"` // Suggested number of goals (optional)
	}
	// Don't require binding - this is optional
	c.ShouldBindJSON(&req)

	// Create a dynamic request for comprehensive goal generation
	message := "Generate 3-5 strategic goals based on the agency's introduction and purpose. Create comprehensive goals that cover all major aspects of the agency's mission and capabilities."

	if req.AdditionalContext != "" {
		message += " Additional context: " + req.AdditionalContext
	}

	if req.GoalCount > 0 {
		message = fmt.Sprintf("Generate %d strategic goals based on the agency's introduction and purpose. Create comprehensive goals that cover all major aspects of the agency's mission and capabilities.", req.GoalCount)
		if req.AdditionalContext != "" {
			message += " Additional context: " + req.AdditionalContext
		}
	}

	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		GoalKeys    []string `json:"goal_keys"`
	}{
		UserMessage: message,
		GoalKeys:    []string{}, // Empty - we're creating new goals
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineGoals handler
	h.RefineGoals(c)
}
