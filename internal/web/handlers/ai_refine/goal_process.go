package ai_refine

import (
	"fmt"
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProcessAIGoalRequest handles POST /api/v1/agencies/:id/goals/ai-process
// Processes multiple AI operations on goals (create, enhance, consolidate)
func (h *Handler) ProcessAIGoalRequest(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse request body
	var req struct {
		Operations []string `json:"operations" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse AI process request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":  agencyID,
		"operations": req.Operations,
	}).Info("Processing AI goal operations")

	// Validate agency exists and get context
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		h.logger.Error("Agency not found", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	// Get introduction for context
	overview, err := h.agencyService.GetAgencyOverview(ctx, agencyID)
	if err != nil {
		h.logger.Warn("Failed to get overview", "agencyID", agencyID, "error", err)
		overview = &agency.Overview{AgencyID: agencyID}
	}

	// Get existing goals
	existingGoals, err := h.agencyService.GetGoals(ctx, agencyID)
	if err != nil {
		h.logger.Error("Failed to get existing goals", "agencyID", agencyID, "error", err)
		existingGoals = []*agency.Goal{}
	}

	// Get units of work for context
	unitsOfWork, err := h.agencyService.GetUnitsOfWork(ctx, agencyID)
	if err != nil {
		h.logger.Warn("Failed to get units of work", "agencyID", agencyID, "error", err)
		unitsOfWork = []*agency.UnitOfWork{}
	}

	results := make(map[string]interface{})
	var createdGoals []*agency.Goal
	var enhancedGoals []*agency.Goal
	var consolidationSuggestions []string

	// Process each operation
	for _, operation := range req.Operations {
		h.logger.Info("Processing operation", "operation", operation, "agencyID", agencyID)

		switch operation {
		case "create":
			h.processCreateOperation(c, agencyID, ag, overview, existingGoals, unitsOfWork, results, &createdGoals)
		case "enhance":
			// TODO: Implement goal enhancement logic
			results["enhance_message"] = "Enhancement feature coming soon"
		case "consolidate":
			h.processConsolidateOperation(c, agencyID, ag, existingGoals, unitsOfWork, results, &createdGoals)
		}
	}

	// Add AI explanation to chat conversation if there's an explanation
	explanation, hasExplanation := results["ai_explanation"].(string)
	if hasExplanation && explanation != "" {
		h.addExplanationToChat(c, agencyID, explanation, len(createdGoals))
	} else {
		h.logger.Warn("No explanation to add to chat",
			"agencyID", agencyID,
			"hasExplanation", hasExplanation,
			"createdGoalsCount", len(createdGoals))
	}

	// Build response
	response := gin.H{
		"success": true,
		"results": results,
	}

	if len(createdGoals) > 0 {
		response["created_goals"] = createdGoals
		response["created_count"] = len(createdGoals)
	}

	if len(enhancedGoals) > 0 {
		response["enhanced_goals"] = enhancedGoals
		response["enhanced_count"] = len(enhancedGoals)
	}

	if len(consolidationSuggestions) > 0 {
		response["consolidation_suggestions"] = consolidationSuggestions
	}

	h.logger.Info("AI goal operations completed",
		"agencyID", agencyID,
		"created", len(createdGoals),
		"enhanced", len(enhancedGoals))

	c.JSON(http.StatusOK, response)
}

func (h *Handler) processCreateOperation(
	c *gin.Context,
	agencyID string,
	ag *agency.Agency,
	overview *agency.Overview,
	existingGoals []*agency.Goal,
	unitsOfWork []*agency.UnitOfWork,
	results map[string]interface{},
	createdGoals *[]*agency.Goal,
) {
	// Generate new goals based on introduction
	if overview.Introduction == "" {
		h.logger.Warn("No introduction found for goal generation", "agencyID", agencyID)
		results["create_error"] = "No introduction found. Please add an introduction first."
		return
	}

	h.logger.Info("Starting multiple goal generation from introduction",
		"agencyID", agencyID,
		"introductionLength", len(overview.Introduction),
		"existingGoalsCount", len(existingGoals))

	// Generate multiple goals in one AI call
	genReq := &ai.GenerateGoalRequest{
		AgencyID:      agencyID,
		AgencyContext: ag,
		ExistingGoals: existingGoals,
		UnitsOfWork:   unitsOfWork,
		UserInput:     "Based on the agency introduction: " + overview.Introduction,
	}

	h.logger.Info("Calling AI to generate multiple goals", "agencyID", agencyID)

	result, err := h.goalRefiner.GenerateGoals(c.Request.Context(), genReq)
	if err != nil {
		h.logger.Error("Failed to generate goals from AI", "agencyID", agencyID, "error", err)
		results["create_error"] = err.Error()
		return
	}

	h.logger.Info("AI generated goals successfully",
		"agencyID", agencyID,
		"goalsCount", len(result.Goals),
		"explanation", result.Explanation)

	// Save each generated goal to database
	for i, goalData := range result.Goals {
		h.logger.Info("Saving generated goal to database",
			"agencyID", agencyID,
			"goalIndex", i+1,
			"goalCode", goalData.SuggestedCode,
			"descriptionLength", len(goalData.Description))

		goal, err := h.agencyService.CreateGoal(c.Request.Context(), agencyID, goalData.SuggestedCode, goalData.Description)
		if err != nil {
			h.logger.Error("Failed to save generated goal",
				"agencyID", agencyID,
				"goalIndex", i+1,
				"goalCode", goalData.SuggestedCode,
				"error", err)
			// Continue with other goals even if one fails
			continue
		}

		h.logger.Info("Goal saved successfully",
			"agencyID", agencyID,
			"goalKey", goal.Key,
			"goalCode", goal.Code,
			"goalNumber", goal.Number)

		*createdGoals = append(*createdGoals, goal)
	}

	h.logger.Info("Completed creating multiple goals",
		"agencyID", agencyID,
		"totalCreated", len(*createdGoals),
		"requested", len(result.Goals))

	results["create_success"] = fmt.Sprintf("Created %d goals", len(*createdGoals))
	results["ai_explanation"] = result.Explanation
}

func (h *Handler) processConsolidateOperation(
	c *gin.Context,
	agencyID string,
	ag *agency.Agency,
	existingGoals []*agency.Goal,
	unitsOfWork []*agency.UnitOfWork,
	results map[string]interface{},
	createdGoals *[]*agency.Goal,
) {
	// Consolidate goals into a lean, strategic list
	if len(existingGoals) < 5 {
		h.logger.Warn("Too few goals to consolidate", "count", len(existingGoals))
		results["consolidate_error"] = "Need at least 5 goals to consolidate"
		return
	}

	h.logger.Info("Starting goal consolidation",
		"agencyID", agencyID,
		"currentGoalsCount", len(existingGoals))

	// Perform consolidation
	consolidationReq := &ai.ConsolidateGoalsRequest{
		AgencyID:      agencyID,
		AgencyContext: ag,
		CurrentGoals:  existingGoals,
		UnitsOfWork:   unitsOfWork,
	}

	consolidationResult, err := h.goalConsolidator.ConsolidateGoals(c.Request.Context(), consolidationReq)
	if err != nil {
		h.logger.Error("Failed to consolidate goals", "agencyID", agencyID, "error", err)
		results["consolidate_error"] = err.Error()
		return
	}

	h.logger.Info("AI consolidated goals successfully",
		"agencyID", agencyID,
		"originalCount", len(existingGoals),
		"consolidatedCount", len(consolidationResult.ConsolidatedGoals),
		"removedCount", len(consolidationResult.RemovedGoals))

	// Delete all goals that should be removed
	for _, removedGoalKey := range consolidationResult.RemovedGoals {
		h.logger.Info("Deleting removed goal",
			"agencyID", agencyID,
			"goalKey", removedGoalKey)

		if err := h.agencyService.DeleteGoal(c.Request.Context(), agencyID, removedGoalKey); err != nil {
			h.logger.Error("Failed to delete goal",
				"agencyID", agencyID,
				"goalKey", removedGoalKey,
				"error", err)
			// Continue with other deletions even if one fails
		}
	}

	// Create new consolidated goals
	var consolidatedGoals []*agency.Goal
	for i, consolidatedGoal := range consolidationResult.ConsolidatedGoals {
		h.logger.Info("Creating consolidated goal",
			"agencyID", agencyID,
			"goalIndex", i+1,
			"goalCode", consolidatedGoal.SuggestedCode,
			"mergedFrom", len(consolidatedGoal.MergedFromKeys))

		goal, err := h.agencyService.CreateGoal(c.Request.Context(), agencyID, consolidatedGoal.SuggestedCode, consolidatedGoal.Description)
		if err != nil {
			h.logger.Error("Failed to create consolidated goal",
				"agencyID", agencyID,
				"goalIndex", i+1,
				"goalCode", consolidatedGoal.SuggestedCode,
				"error", err)
			// Continue with other goals even if one fails
			continue
		}

		h.logger.Info("Consolidated goal created successfully",
			"agencyID", agencyID,
			"goalKey", goal.Key,
			"goalCode", goal.Code)

		consolidatedGoals = append(consolidatedGoals, goal)
	}

	h.logger.Info("Completed goal consolidation",
		"agencyID", agencyID,
		"originalCount", len(existingGoals),
		"finalCount", len(consolidatedGoals),
		"deleted", len(consolidationResult.RemovedGoals))

	results["consolidate_success"] = fmt.Sprintf("Consolidated from %d to %d goals",
		len(existingGoals), len(consolidatedGoals))
	results["consolidation_summary"] = consolidationResult.Summary
	results["ai_explanation"] = consolidationResult.Explanation
	results["removed_count"] = len(consolidationResult.RemovedGoals)
	results["new_count"] = len(consolidatedGoals)

	// Add consolidated goals to response
	if len(consolidatedGoals) > 0 {
		*createdGoals = append(*createdGoals, consolidatedGoals...)
	}
}

func (h *Handler) addExplanationToChat(c *gin.Context, agencyID string, explanation string, createdGoalsCount int) {
	h.logger.Info("Attempting to add AI explanation to chat",
		"agencyID", agencyID,
		"createdGoalsCount", createdGoalsCount,
		"explanationLength", len(explanation))

	// Get or create conversation
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		h.logger.Warn("No conversation exists, creating new one",
			"agencyID", agencyID,
			"error", err)
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation for AI goal generation message")
			return
		}
	}

	if conversation == nil {
		h.logger.Error("Conversation is nil after creation attempt", "agencyID", agencyID)
		return
	}

	// Build appropriate message based on whether goals were created
	var chatMessage string
	if createdGoalsCount > 0 {
		chatMessage = fmt.Sprintf("✨ **Created %d Goals**\n\n%s", createdGoalsCount, explanation)
	} else {
		chatMessage = fmt.Sprintf("✨ **Goal Analysis**\n\n%s", explanation)
	}

	h.logger.Info("Adding message to chat",
		"agencyID", agencyID,
		"conversationID", conversation.ID,
		"messageLength", len(chatMessage))

	if addErr := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add goal generation explanation to chat")
	} else {
		h.logger.Info("Successfully added AI explanation to chat",
			"agencyID", agencyID,
			"conversationID", conversation.ID)
	}
}
