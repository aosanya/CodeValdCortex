package ai_refine

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProcessAIGoalRequest handles POST /api/v1/agencies/:id/goals/ai-process
// Processes multiple AI operations on goals (create, enhance, consolidate)
func (h *Handler) ProcessAIGoalRequest(c *gin.Context) {
	h.logger.Info("🔵 HANDLER CALLED: ProcessAIGoalRequest")

	agencyID := c.Param("id")

	// Parse request body
	var req struct {
		Operations  []string `json:"operations" binding:"required"`
		GoalKeys    []string `json:"goal_keys"`    // Optional: specific goals to enhance/consolidate
		UserRequest string   `json:"user_request"` // Optional: user's request for creating/modifying goals
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse AI process request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":    agencyID,
		"operations":   req.Operations,
		"goal_keys":    req.GoalKeys,
		"user_request": req.UserRequest,
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

	// Filter goals if specific keys were provided
	var goalsToProcess []*agency.Goal
	if len(req.GoalKeys) > 0 {
		goalKeyMap := make(map[string]bool)
		for _, key := range req.GoalKeys {
			goalKeyMap[key] = true
		}
		for _, goal := range existingGoals {
			if goalKeyMap[goal.Key] {
				goalsToProcess = append(goalsToProcess, goal)
			}
		}
		h.logger.Info("Filtered goals for processing",
			"agencyID", agencyID,
			"requestedKeys", len(req.GoalKeys),
			"foundGoals", len(goalsToProcess))
	} else {
		goalsToProcess = existingGoals
	}

	// Get units of work for context
	workItems, err := h.agencyService.GetWorkItems(ctx, agencyID)
	if err != nil {
		h.logger.Warn("Failed to get units of work", "agencyID", agencyID, "error", err)
		workItems = []*agency.WorkItem{}
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
			h.logger.Info("🔵 CALLING: processCreateOperation", "agencyID", agencyID)
			h.processCreateOperation(c, agencyID, ag, overview, existingGoals, workItems, req.UserRequest, results, &createdGoals)
		case "enhance":
			h.logger.Info("🔵 CALLING: processEnhanceOperation", "agencyID", agencyID)
			h.processEnhanceOperation(c, agencyID, ag, goalsToProcess, workItems, results, &enhancedGoals)
		case "consolidate":
			h.logger.Info("🔵 CALLING: processConsolidateOperation", "agencyID", agencyID)
			h.processConsolidateOperation(c, agencyID, ag, goalsToProcess, workItems, results, &createdGoals)
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
	workItems []*agency.WorkItem,
	userRequest string,
	results map[string]interface{},
	createdGoals *[]*agency.Goal,
) {
	h.logger.Info("🟢 FUNCTION ENTRY: processCreateOperation", "agencyID", agencyID)

	// Generate new goals based on introduction or user request
	var userInput string
	if userRequest != "" {
		// User provided specific request for goal creation
		userInput = userRequest
		h.logger.Info("Using user request for goal generation",
			"agencyID", agencyID,
			"userRequest", userRequest)
	} else if overview.Introduction != "" {
		// Fall back to using introduction
		userInput = "Based on the agency introduction: " + overview.Introduction
		h.logger.Info("Using introduction for goal generation",
			"agencyID", agencyID,
			"introductionLength", len(overview.Introduction))
	} else {
		h.logger.Warn("No introduction or user request found for goal generation", "agencyID", agencyID)
		results["create_error"] = "No introduction or user request found. Please add an introduction or provide a specific goal request."
		return
	}

	h.logger.Info("Starting multiple goal generation from introduction",
		"agencyID", agencyID,
		"userInputLength", len(userInput),
		"existingGoalsCount", len(existingGoals))

	// Build AI context
	builderContext, err := h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, "", userInput)
	if err != nil {
		h.logger.Error("Failed to build AI context", "agencyID", agencyID, "error", err)
		results["create_error"] = fmt.Sprintf("Failed to build context: %v", err)
		return
	}

	// Generate multiple goals in one AI call
	genReq := &builder.GenerateGoalRequest{
		AgencyID:      agencyID,
		AgencyContext: ag,
		ExistingGoals: existingGoals,
		WorkItems:     workItems,
		UserInput:     userInput,
	}

	h.logger.Info("Calling AI to generate multiple goals", "agencyID", agencyID)

	result, err := h.goalRefiner.GenerateGoals(c.Request.Context(), genReq, builderContext)
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

func (h *Handler) processEnhanceOperation(
	c *gin.Context,
	agencyID string,
	ag *agency.Agency,
	existingGoals []*agency.Goal,
	workItems []*agency.WorkItem,
	results map[string]interface{},
	enhancedGoals *[]*agency.Goal,
) {
	h.logger.Info("🟢 FUNCTION ENTRY: processEnhanceOperation", "agencyID", agencyID)

	// Check if there are goals to enhance
	if len(existingGoals) == 0 {
		h.logger.Warn("No goals to enhance", "agencyID", agencyID)
		results["enhance_error"] = "No goals found. Please create goals first."
		return
	}

	h.logger.Info("Starting goal enhancement",
		"agencyID", agencyID,
		"goalsCount", len(existingGoals))

	var enhancementResults []string
	changedCount := 0

	// Enhance each goal individually
	for i, goal := range existingGoals {
		h.logger.Info("Enhancing goal",
			"agencyID", agencyID,
			"goalIndex", i+1,
			"goalKey", goal.Key,
			"goalCode", goal.Code)

		// Build AI context for this goal refinement
		builderContext, err := h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, "", "")
		if err != nil {
			h.logger.Error("Failed to build AI context",
				"agencyID", agencyID,
				"goalKey", goal.Key,
				"error", err)
			enhancementResults = append(enhancementResults, fmt.Sprintf("Failed to build context for %s: %s", goal.Code, err.Error()))
			continue
		}

		// Build refinement request
		refineReq := &builder.RefineGoalRequest{
			AgencyID:       agencyID,
			CurrentGoal:    goal,
			Description:    goal.Description,
			Scope:          goal.Scope,
			SuccessMetrics: goal.SuccessMetrics,
			ExistingGoals:  existingGoals,
			WorkItems:      workItems,
			AgencyContext:  ag,
		}

		// Call AI to refine the goal
		refinedResult, err := h.goalRefiner.RefineGoal(c.Request.Context(), refineReq, builderContext)
		if err != nil {
			h.logger.Error("Failed to enhance goal",
				"agencyID", agencyID,
				"goalKey", goal.Key,
				"error", err)
			enhancementResults = append(enhancementResults, fmt.Sprintf("Failed to enhance %s: %s", goal.Code, err.Error()))
			continue
		}

		// Only update if the AI made changes
		if refinedResult.WasChanged {
			h.logger.Info("Goal was enhanced by AI, updating...",
				"agencyID", agencyID,
				"goalKey", goal.Key,
				"goalCode", goal.Code,
				"explanation", refinedResult.Explanation)

			// Update the goal with refined content
			goal.Description = refinedResult.RefinedDescription
			goal.Scope = refinedResult.RefinedScope
			goal.SuccessMetrics = refinedResult.RefinedMetrics

			// For now, just update description using the existing UpdateGoal method
			// TODO: Extend UpdateGoal to support scope and metrics
			err := h.agencyService.UpdateGoal(c.Request.Context(), agencyID, goal.Key, goal.Code, refinedResult.RefinedDescription)
			if err != nil {
				h.logger.Error("Failed to save enhanced goal",
					"agencyID", agencyID,
					"goalKey", goal.Key,
					"error", err)
				enhancementResults = append(enhancementResults, fmt.Sprintf("Failed to save %s: %s", goal.Code, err.Error()))
				continue
			}

			changedCount++
			*enhancedGoals = append(*enhancedGoals, goal)
			enhancementResults = append(enhancementResults, fmt.Sprintf("Enhanced %s: %s", goal.Code, refinedResult.Explanation))

			h.logger.Info("Goal enhanced successfully",
				"agencyID", agencyID,
				"goalKey", goal.Key,
				"goalCode", goal.Code)
		} else {
			h.logger.Info("Goal did not need enhancement",
				"agencyID", agencyID,
				"goalKey", goal.Key,
				"goalCode", goal.Code)
			enhancementResults = append(enhancementResults, fmt.Sprintf("%s: No changes needed", goal.Code))
		}
	}

	h.logger.Info("Completed goal enhancement",
		"agencyID", agencyID,
		"totalProcessed", len(existingGoals),
		"changedCount", changedCount)

	// Build explanation from all enhancement results
	explanation := strings.Join(enhancementResults, ". ")

	results["enhance_success"] = fmt.Sprintf("Enhanced %d of %d goals", changedCount, len(existingGoals))
	results["ai_explanation"] = explanation
	results["changed_count"] = changedCount
	results["unchanged_count"] = len(existingGoals) - changedCount
}

func (h *Handler) processConsolidateOperation(
	c *gin.Context,
	agencyID string,
	ag *agency.Agency,
	existingGoals []*agency.Goal,
	workItems []*agency.WorkItem,
	results map[string]interface{},
	createdGoals *[]*agency.Goal,
) {
	h.logger.Info("🟢 FUNCTION ENTRY: processConsolidateOperation", "agencyID", agencyID)

	// Consolidate goals into a lean, strategic list
	if len(existingGoals) < 2 {
		h.logger.Warn("Too few goals to consolidate", "count", len(existingGoals))
		results["consolidate_error"] = "Need at least 2 goals to consolidate"
		return
	}

	h.logger.Info("Starting goal consolidation",
		"agencyID", agencyID,
		"currentGoalsCount", len(existingGoals))

	// Build AI context for consolidation
	builderContext, err := h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, "", "")
	if err != nil {
		h.logger.Error("Failed to build AI context", "agencyID", agencyID, "error", err)
		results["consolidate_error"] = fmt.Sprintf("Failed to build context: %v", err)
		return
	}

	// Perform consolidation
	consolidationReq := &builder.ConsolidateGoalsRequest{
		AgencyID:      agencyID,
		AgencyContext: ag,
		CurrentGoals:  existingGoals,
		WorkItems:     workItems,
	}

	consolidationResult, err := h.goalRefiner.ConsolidateGoals(c.Request.Context(), consolidationReq, builderContext)
	if err != nil {
		h.logger.Error("Failed to consolidate goals", "agencyID", agencyID, "error", err)
		results["consolidate_error"] = err.Error()
		return
	}

	h.logger.Info("AI consolidation analysis complete",
		"agencyID", agencyID,
		"originalCount", len(existingGoals),
		"consolidatedCount", len(consolidationResult.ConsolidatedGoals),
		"removedCount", len(consolidationResult.RemovedGoals))

	// Check if AI decided no consolidation is needed
	if len(consolidationResult.ConsolidatedGoals) == 0 {
		h.logger.Info("AI determined no consolidation needed - goals are already distinct",
			"agencyID", agencyID,
			"goalCount", len(existingGoals))

		results["consolidate_success"] = "No consolidation needed - goals are already well-defined and distinct"
		results["consolidation_summary"] = consolidationResult.Summary
		results["ai_explanation"] = consolidationResult.Explanation
		results["removed_count"] = 0
		results["new_count"] = 0
		return
	}

	// Determine which goals to delete
	goalsToDelete := consolidationResult.RemovedGoals

	// If AI didn't specify which goals to remove, delete all input goals that were consolidated
	if len(consolidationResult.RemovedGoals) == 0 && len(consolidationResult.ConsolidatedGoals) > 0 {
		h.logger.Warn("AI did not specify which goals to remove, will delete all input goals",
			"agencyID", agencyID,
			"originalCount", len(existingGoals))

		// Delete all the goals that were sent for consolidation
		for _, goal := range existingGoals {
			goalsToDelete = append(goalsToDelete, goal.Key)
		}
	}

	// Delete goals that were merged/consolidated
	for _, removedGoalKey := range goalsToDelete {
		h.logger.Info("Deleting consolidated goal",
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
			"mergedFrom", len(consolidatedGoal.ConsolidatedFrom))

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

func (h *Handler) formatExplanationAsBullets(explanation string) string {
	// Split by common sentence delimiters or line breaks
	sentences := strings.Split(explanation, ". ")

	var bullets []string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Add period back if it was removed by split
		if !strings.HasSuffix(sentence, ".") && !strings.HasSuffix(sentence, "!") && !strings.HasSuffix(sentence, "?") {
			sentence += "."
		}

		// Format as bullet point
		bullets = append(bullets, "• "+sentence)
	}

	return strings.Join(bullets, "\n")
}

func (h *Handler) addExplanationToChat(c *gin.Context, agencyID string, explanation string, createdGoalsCount int) {
	h.logger.Info("🟢 FUNCTION ENTRY: addExplanationToChat",
		"agencyID", agencyID,
		"createdGoalsCount", createdGoalsCount,
		"explanationLength", len(explanation))

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

	// Format explanation as bullet points
	formattedExplanation := h.formatExplanationAsBullets(explanation)

	// Build appropriate message based on whether goals were created
	var chatMessage string
	if createdGoalsCount > 0 {
		chatMessage = fmt.Sprintf("✨ **Created %d Goals**\n\n%s", createdGoalsCount, formattedExplanation)
	} else {
		chatMessage = fmt.Sprintf("✨ **Goal Analysis**\n\n%s", formattedExplanation)
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

// ProcessGoalChatRequest handles chat-based goal interactions
// This is similar to RefineIntroduction but for goals in chat context
func (h *Handler) ProcessGoalChatRequest(c *gin.Context) {
	h.logger.Info("🔵 HANDLER CALLED: ProcessGoalChatRequest")

	agencyID := c.Param("id")

	// Get user's chat message/request
	userRequest := c.PostForm("user-request")
	if userRequest == "" {
		userRequest = c.PostForm("message")
	}

	if userRequest == "" {
		h.logger.Error("No user request provided for goal chat")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, `
			<div class="notification is-warning">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-warning mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>No Request Provided</strong>
						<p class="mb-0">Please provide a message or request.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":    agencyID,
		"user_request": userRequest,
	}).Info("Processing chat-based goal request")

	// Get agency context
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
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

	// Get conversation context for recent chat messages
	conv, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil || conv == nil {
		h.logger.WithError(err).Warn("No conversation found for agency")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, `
			<div class="notification is-warning">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-warning mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>No Conversation</strong>
						<p class="mb-0">Please start a conversation first.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Add user message to conversation
	if addErr := h.designerService.AddMessage(conv.ID, "user", userRequest); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add user message to conversation")
	}

	// Build AI context
	builderContext, err := h.contextBuilder.BuildBuilderContext(ctx, ag, "", userRequest)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build AI context")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
					</span>
					<div>
						<strong>Context Build Failed</strong>
						<p class="mb-0">Failed to gather necessary context data.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Determine operation from user request
	// Simple keyword-based routing for now
	userRequestLower := strings.ToLower(userRequest)
	var operation string
	if strings.Contains(userRequestLower, "create") || strings.Contains(userRequestLower, "generate") || strings.Contains(userRequestLower, "add") {
		operation = "create"
	} else if strings.Contains(userRequestLower, "consolidate") || strings.Contains(userRequestLower, "merge") || strings.Contains(userRequestLower, "combine") {
		operation = "consolidate"
	} else if strings.Contains(userRequestLower, "status") || strings.Contains(userRequestLower, "list") || strings.Contains(userRequestLower, "show") {
		operation = "status"
	} else {
		// Default to enhance for general requests (including remove, delete, improve, refine, etc.)
		operation = "enhance"
	}

	h.logger.Info("Determined operation from user request",
		"operation", operation,
		"userRequest", userRequest)

	// Get existing goals for context
	existingGoals := builderContext.Goals

	// Process based on operation
	var responseMessage string
	switch operation {
	case "create":
		// Generate new goals
		genReq := &builder.GenerateGoalRequest{
			AgencyID:      agencyID,
			AgencyContext: ag,
		}
		result, genErr := h.goalRefiner.GenerateGoals(ctx, genReq, builderContext)
		if genErr != nil {
			h.logger.WithError(genErr).Error("Failed to generate goals")
			responseMessage = fmt.Sprintf("❌ Failed to generate goals: %v", genErr)
		} else {
			// Create goals using agency service
			createdCount := 0
			for _, goalResp := range result.Goals {
				_, createErr := h.agencyService.CreateGoal(ctx, agencyID, goalResp.SuggestedCode, goalResp.Description)
				if createErr != nil {
					h.logger.WithError(createErr).Error("Failed to create goal", "goalCode", goalResp.SuggestedCode)
				} else {
					createdCount++
				}
			}
			responseMessage = fmt.Sprintf("✨ **Created %d Goals**\n\n%s", createdCount, result.Explanation)
		}

	case "consolidate":
		// Consolidate goals
		if len(existingGoals) < 2 {
			responseMessage = "ℹ️ Need at least 2 goals to consolidate."
		} else {
			consolidateReq := &builder.ConsolidateGoalsRequest{
				AgencyID:      agencyID,
				AgencyContext: ag,
			}
			result, consolidateErr := h.goalRefiner.ConsolidateGoals(ctx, consolidateReq, builderContext)
			if consolidateErr != nil {
				h.logger.WithError(consolidateErr).Error("Failed to consolidate goals")
				responseMessage = fmt.Sprintf("❌ Failed to consolidate goals: %v", consolidateErr)
			} else {
				// Format consolidation suggestions
				suggestions := make([]string, len(result.ConsolidatedGoals))
				for i, cGoal := range result.ConsolidatedGoals {
					suggestions[i] = fmt.Sprintf("**%s**: %s", cGoal.SuggestedCode, cGoal.Description)
				}
				responseMessage = fmt.Sprintf("💡 **Consolidation Suggestions**\n\n%s\n\n**Removed Goals**: %s\n\n%s",
					strings.Join(suggestions, "\n"),
					strings.Join(result.RemovedGoals, ", "),
					result.Summary)
			}
		}

	case "enhance":
		// Enhance/refine goals based on user request (includes remove, delete, improve, refine)
		if len(existingGoals) == 0 {
			responseMessage = "ℹ️ No goals found to work with. Try creating some goals first."
		} else {
			h.logger.Info("Processing enhance/refine request",
				"agencyID", agencyID,
				"userRequest", userRequest,
				"existingGoalsCount", len(existingGoals))

			// General enhancement - provide guidance
			responseMessage = fmt.Sprintf("🔧 **Goal Enhancement**\n\nI can help you with the following:\n\n"+
				"• **Remove specific goals**: Specify the goal code (e.g., 'remove G013')\n"+
				"• **Improve goal descriptions**: I can enhance clarity and specificity\n"+
				"• **Refine scope and metrics**: I can suggest better success criteria\n\n"+
				"Current goals (%d):\n", len(existingGoals))

			for i, goal := range existingGoals {
				responseMessage += fmt.Sprintf("%d. **%s**: %s\n", i+1, goal.Code, goal.Description)
			}

			responseMessage += "\n💡 Try asking me to 'remove G013' or 'consolidate goals'."
		}

	case "status":
		// Provide status of current goals
		if len(existingGoals) == 0 {
			responseMessage = "ℹ️ No goals have been defined yet. Would you like me to create some goals based on the agency introduction?"
		} else {
			goalsList := make([]string, len(existingGoals))
			for i, goal := range existingGoals {
				goalsList[i] = fmt.Sprintf("- **%s**: %s", goal.Code, goal.Description)
			}
			responseMessage = fmt.Sprintf("📊 **Current Goals (%d)**\n\n%s\n\nHow can I help with these goals? I can create new ones, consolidate them, or help refine them.", len(existingGoals), strings.Join(goalsList, "\n"))
		}

	default:
		responseMessage = "❓ I'm not sure what you want to do with the goals. Try asking to create, consolidate, or enhance goals, or ask for the current status."
	}

	// Add AI response to conversation
	if addErr := h.designerService.AddMessage(conv.ID, "assistant", responseMessage); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add AI response to conversation")
	}

	// Render chat messages (user + assistant)
	c.Header("Content-Type", "text/html")

	// Get updated conversation to retrieve messages
	updatedConv, err := h.designerService.GetConversation(conv.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get updated conversation")
		c.String(http.StatusInternalServerError, "Failed to render messages")
		return
	}

	// Render last 2 messages (user + assistant)
	messageCount := len(updatedConv.Messages)
	if messageCount >= 2 {
		// Render user message
		userMsg := &updatedConv.Messages[messageCount-2]
		if renderErr := agency_designer.UserMessage(*userMsg).Render(ctx, c.Writer); renderErr != nil {
			h.logger.WithError(renderErr).Error("Failed to render user message")
		}
		// Render assistant message
		aiMsg := &updatedConv.Messages[messageCount-1]
		if renderErr := agency_designer.AIMessage(*aiMsg).Render(ctx, c.Writer); renderErr != nil {
			h.logger.WithError(renderErr).Error("Failed to render AI message")
		}
	}

	h.logger.Info("Successfully processed chat-based goal request",
		"agencyID", agencyID,
		"operation", operation)
}
