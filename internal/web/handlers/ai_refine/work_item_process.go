package ai_refine

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProcessAIWorkItemRequest handles POST /api/v1/agencies/:id/work-items/ai-process
// Processes multiple AI operations on work items (create, enhance, consolidate)
func (h *Handler) ProcessAIWorkItemRequest(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse request body
	var req struct {
		Operations   []string `json:"operations" binding:"required"`
		WorkItemKeys []string `json:"work_item_keys"` // Optional: specific work items to enhance/consolidate
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse AI process request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":      agencyID,
		"operations":     req.Operations,
		"work_item_keys": req.WorkItemKeys,
	}).Info("Processing AI work item operations")

	// Validate agency exists and get context
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		h.logger.Error("Agency not found", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	// Get existing work items
	existingWorkItems, err := h.agencyService.GetWorkItems(ctx, agencyID)
	if err != nil {
		h.logger.Error("Failed to get existing work items", "agencyID", agencyID, "error", err)
		existingWorkItems = []*agency.WorkItem{}
	}

	// Filter work items if specific keys were provided
	var workItemsToProcess []*agency.WorkItem
	if len(req.WorkItemKeys) > 0 {
		workItemKeyMap := make(map[string]bool)
		for _, key := range req.WorkItemKeys {
			workItemKeyMap[key] = true
		}
		for _, workItem := range existingWorkItems {
			if workItemKeyMap[workItem.Key] {
				workItemsToProcess = append(workItemsToProcess, workItem)
			}
		}
		h.logger.Info("Filtered work items for processing",
			"agencyID", agencyID,
			"requestedKeys", len(req.WorkItemKeys),
			"foundWorkItems", len(workItemsToProcess))
	} else {
		workItemsToProcess = existingWorkItems
	}

	results := make(map[string]interface{})
	var createdWorkItems []*agency.WorkItem
	var enhancedWorkItems []*agency.WorkItem

	// Process each operation
	for _, operation := range req.Operations {
		h.logger.Info("Processing operation", "operation", operation, "agencyID", agencyID)

		switch operation {
		case "create":
			h.processCreateWorkItemsOperation(c, agencyID, ag, results, &createdWorkItems)
		case "enhance":
			h.processEnhanceWorkItemsOperation(c, agencyID, ag, workItemsToProcess, results, &enhancedWorkItems)
		case "consolidate":
			h.processConsolidateWorkItemsOperation(c, agencyID, ag, workItemsToProcess, results, &createdWorkItems)
		}
	}

	// Add AI explanation to chat conversation if there's an explanation
	explanation, hasExplanation := results["ai_explanation"].(string)
	if hasExplanation && explanation != "" {
		h.addWorkItemExplanationToChat(c, agencyID, explanation, len(createdWorkItems))
	} else {
		h.logger.Warn("No explanation to add to chat",
			"agencyID", agencyID,
			"hasExplanation", hasExplanation,
			"createdWorkItemsCount", len(createdWorkItems))
	}

	// Build response
	response := gin.H{
		"success": true,
		"results": results,
	}

	if len(createdWorkItems) > 0 {
		response["created_work_items"] = createdWorkItems
		response["created_count"] = len(createdWorkItems)
	}

	if len(enhancedWorkItems) > 0 {
		response["enhanced_work_items"] = enhancedWorkItems
		response["enhanced_count"] = len(enhancedWorkItems)
	}

	h.logger.Info("AI work item operations completed",
		"agencyID", agencyID,
		"created", len(createdWorkItems),
		"enhanced", len(enhancedWorkItems))

	c.JSON(http.StatusOK, response)
}

func (h *Handler) processCreateWorkItemsOperation(
	c *gin.Context,
	agencyID string,
	ag *agency.Agency,
	results map[string]interface{},
	createdWorkItems *[]*agency.WorkItem,
) {
	// Build AI context to get goals
	builderContext, err := h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, "", "")
	if err != nil {
		h.logger.Error("Failed to build context for work item generation", "agencyID", agencyID, "error", err)
		results["create_error"] = err.Error()
		return
	}

	// Generate new work items based on goals from context
	if len(builderContext.Goals) == 0 {
		h.logger.Warn("No goals found for work item generation", "agencyID", agencyID)
		results["create_error"] = "No goals found. Please create goals first."
		return
	}

	h.logger.Info("Starting work item generation from goals",
		"agencyID", agencyID,
		"goalsCount", len(builderContext.Goals),
		"existingWorkItemsCount", len(builderContext.WorkItems))

	// Build user input from goals
	var goalsContext strings.Builder
	goalsContext.WriteString("Based on the following goals:\n")
	for _, goal := range builderContext.Goals {
		goalsContext.WriteString(fmt.Sprintf("- %s: %s\n", goal.Code, goal.Description))
	}

	// Rebuild context with user input
	builderContext, err = h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, "", goalsContext.String())
	if err != nil {
		h.logger.Error("Failed to build context for work item generation", "agencyID", agencyID, "error", err)
		results["create_error"] = err.Error()
		return
	}

	// Generate multiple work items in one AI call
	genReq := &builder.GenerateWorkItemRequest{
		AgencyID:  agencyID,
		UserInput: goalsContext.String(),
	}

	h.logger.Info("Calling AI to generate multiple work items", "agencyID", agencyID)

	result, err := h.workItemBuilder.GenerateWorkItems(c.Request.Context(), genReq, builderContext)
	if err != nil {
		h.logger.Error("Failed to generate work items from AI", "agencyID", agencyID, "error", err)
		results["create_error"] = err.Error()
		return
	}

	h.logger.Info("AI generated work items successfully",
		"agencyID", agencyID,
		"workItemsCount", len(result.WorkItems),
		"explanation", result.Explanation)

	// Save each generated work item to database
	for i, workItemData := range result.WorkItems {
		h.logger.Info("Saving generated work item to database",
			"agencyID", agencyID,
			"workItemIndex", i+1,
			"workItemCode", workItemData.SuggestedCode,
			"titleLength", len(workItemData.Title))

		req := agency.CreateWorkItemRequest{
			Title:        workItemData.Title,
			Description:  workItemData.Description,
			Deliverables: workItemData.Deliverables,
			Tags:         workItemData.SuggestedTags,
		}

		savedWorkItem, err := h.agencyService.CreateWorkItem(c.Request.Context(), agencyID, req)
		if err != nil {
			h.logger.Error("Failed to save generated work item",
				"agencyID", agencyID,
				"workItemIndex", i+1,
				"workItemCode", workItemData.SuggestedCode,
				"error", err)
			// Continue with other work items even if one fails
			continue
		}

		h.logger.Info("Work item saved successfully",
			"agencyID", agencyID,
			"workItemKey", savedWorkItem.Key,
			"workItemCode", savedWorkItem.Code,
			"workItemNumber", savedWorkItem.Number)

		*createdWorkItems = append(*createdWorkItems, savedWorkItem)
	}

	h.logger.Info("Completed creating multiple work items",
		"agencyID", agencyID,
		"totalCreated", len(*createdWorkItems),
		"requested", len(result.WorkItems))

	results["create_success"] = fmt.Sprintf("Created %d work items", len(*createdWorkItems))
	results["ai_explanation"] = result.Explanation
}

func (h *Handler) processEnhanceWorkItemsOperation(
	c *gin.Context,
	agencyID string,
	ag *agency.Agency,
	existingWorkItems []*agency.WorkItem,
	results map[string]interface{},
	enhancedWorkItems *[]*agency.WorkItem,
) {
	// Check if there are work items to enhance
	if len(existingWorkItems) == 0 {
		h.logger.Warn("No work items to enhance", "agencyID", agencyID)
		results["enhance_error"] = "No work items found. Please create work items first."
		return
	}

	h.logger.Info("Starting work item enhancement",
		"agencyID", agencyID,
		"workItemsCount", len(existingWorkItems))

	var enhancementResults []string
	changedCount := 0

	// Enhance each work item individually
	for i, workItem := range existingWorkItems {
		h.logger.Info("Enhancing work item",
			"agencyID", agencyID,
			"workItemIndex", i+1,
			"workItemKey", workItem.Key,
			"workItemCode", workItem.Code)

		// Build AI context for refinement
		builderContext, err := h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, "", "")
		if err != nil {
			h.logger.Error("Failed to build context for work item refinement", "agencyID", agencyID, "error", err)
			enhancementResults = append(enhancementResults, fmt.Sprintf("Failed to build context for %s: %s", workItem.Code, err.Error()))
			continue
		}

		// Build refinement request
		refineReq := &builder.RefineWorkItemRequest{
			AgencyID:     agencyID,
			Title:        workItem.Title,
			Description:  workItem.Description,
			Deliverables: workItem.Deliverables,
		}

		// Call AI to refine the work item
		refinedResult, err := h.workItemBuilder.RefineWorkItem(c.Request.Context(), refineReq, builderContext)
		if err != nil {
			h.logger.Error("Failed to enhance work item",
				"agencyID", agencyID,
				"workItemKey", workItem.Key,
				"error", err)
			enhancementResults = append(enhancementResults, fmt.Sprintf("Failed to enhance %s: %s", workItem.Code, err.Error()))
			continue
		}

		// Only update if the AI made changes
		if refinedResult.WasChanged {
			h.logger.Info("Work item was enhanced by AI, updating...",
				"agencyID", agencyID,
				"workItemKey", workItem.Key,
				"workItemCode", workItem.Code,
				"explanation", refinedResult.Explanation)

			// Update the work item with refined content
			updateReq := agency.UpdateWorkItemRequest{
				Title:        refinedResult.RefinedTitle,
				Description:  refinedResult.RefinedDescription,
				Deliverables: refinedResult.RefinedDeliverables,
				Dependencies: workItem.Dependencies,
				Tags:         refinedResult.SuggestedTags,
			}

			// Update work item in database
			err := h.agencyService.UpdateWorkItem(c.Request.Context(), agencyID, workItem.Key, updateReq)
			if err != nil {
				h.logger.Error("Failed to save enhanced work item",
					"agencyID", agencyID,
					"workItemKey", workItem.Key,
					"error", err)
				enhancementResults = append(enhancementResults, fmt.Sprintf("Failed to save %s: %s", workItem.Code, err.Error()))
				continue
			}

			// Update the local workItem struct for response
			workItem.Title = refinedResult.RefinedTitle
			workItem.Description = refinedResult.RefinedDescription
			workItem.Deliverables = refinedResult.RefinedDeliverables
			workItem.Tags = refinedResult.SuggestedTags

			changedCount++
			*enhancedWorkItems = append(*enhancedWorkItems, workItem)
			enhancementResults = append(enhancementResults, fmt.Sprintf("Enhanced %s: %s", workItem.Code, refinedResult.Explanation))

			h.logger.Info("Work item enhanced successfully",
				"agencyID", agencyID,
				"workItemKey", workItem.Key,
				"workItemCode", workItem.Code)
		} else {
			h.logger.Info("Work item did not need enhancement",
				"agencyID", agencyID,
				"workItemKey", workItem.Key,
				"workItemCode", workItem.Code)
			enhancementResults = append(enhancementResults, fmt.Sprintf("%s: No changes needed", workItem.Code))
		}
	}

	h.logger.Info("Completed work item enhancement",
		"agencyID", agencyID,
		"totalProcessed", len(existingWorkItems),
		"changedCount", changedCount)

	// Build explanation from all enhancement results
	explanation := strings.Join(enhancementResults, ". ")

	results["enhance_success"] = fmt.Sprintf("Enhanced %d of %d work items", changedCount, len(existingWorkItems))
	results["ai_explanation"] = explanation
	results["changed_count"] = changedCount
	results["unchanged_count"] = len(existingWorkItems) - changedCount
}

func (h *Handler) processConsolidateWorkItemsOperation(
	c *gin.Context,
	agencyID string,
	ag *agency.Agency,
	existingWorkItems []*agency.WorkItem,
	results map[string]interface{},
	createdWorkItems *[]*agency.WorkItem,
) {
	// Consolidate work items into a lean, manageable list
	if len(existingWorkItems) < 2 {
		h.logger.Warn("Too few work items to consolidate", "count", len(existingWorkItems))
		results["consolidate_error"] = "Need at least 2 work items to consolidate"
		return
	}

	h.logger.Info("Starting work item consolidation",
		"agencyID", agencyID,
		"currentWorkItemsCount", len(existingWorkItems))

	// Build AI context for consolidation
	builderContext, err := h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, "", "")
	if err != nil {
		h.logger.Error("Failed to build context for work item consolidation", "agencyID", agencyID, "error", err)
		results["consolidate_error"] = err.Error()
		return
	}

	// Perform consolidation
	consolidationReq := &builder.ConsolidateWorkItemsRequest{
		AgencyID: agencyID,
	}

	consolidationResult, err := h.workItemBuilder.ConsolidateWorkItems(c.Request.Context(), consolidationReq, builderContext)
	if err != nil {
		h.logger.Error("Failed to consolidate work items", "agencyID", agencyID, "error", err)
		results["consolidate_error"] = err.Error()
		return
	}

	h.logger.Info("AI consolidation analysis complete",
		"agencyID", agencyID,
		"originalCount", len(existingWorkItems),
		"consolidatedCount", len(consolidationResult.ConsolidatedWorkItems),
		"removedCount", len(consolidationResult.RemovedWorkItems))

	// Check if AI decided no consolidation is needed
	if len(consolidationResult.ConsolidatedWorkItems) == 0 {
		h.logger.Info("AI determined no consolidation needed - work items are already distinct",
			"agencyID", agencyID,
			"workItemCount", len(existingWorkItems))

		results["consolidate_success"] = "No consolidation needed - work items are already well-defined and distinct"
		results["consolidation_summary"] = consolidationResult.Summary
		results["ai_explanation"] = consolidationResult.Explanation
		results["removed_count"] = 0
		results["new_count"] = 0
		return
	}

	// Determine which work items to delete
	workItemsToDelete := consolidationResult.RemovedWorkItems

	// If AI didn't specify which work items to remove, delete all input work items that were consolidated
	if len(consolidationResult.RemovedWorkItems) == 0 && len(consolidationResult.ConsolidatedWorkItems) > 0 {
		h.logger.Warn("AI did not specify which work items to remove, will delete all input work items",
			"agencyID", agencyID,
			"originalCount", len(existingWorkItems))

		// Delete all the work items that were sent for consolidation
		for _, workItem := range existingWorkItems {
			workItemsToDelete = append(workItemsToDelete, workItem.Key)
		}
	}

	// Delete work items that were merged/consolidated
	for _, removedWorkItemKey := range workItemsToDelete {
		h.logger.Info("Deleting consolidated work item",
			"agencyID", agencyID,
			"workItemKey", removedWorkItemKey)

		if err := h.agencyService.DeleteWorkItem(c.Request.Context(), agencyID, removedWorkItemKey); err != nil {
			h.logger.Error("Failed to delete work item",
				"agencyID", agencyID,
				"workItemKey", removedWorkItemKey,
				"error", err)
			// Continue with other deletions even if one fails
		}
	}

	// Create new consolidated work items
	var consolidatedWorkItems []*agency.WorkItem
	for i, consolidatedWorkItem := range consolidationResult.ConsolidatedWorkItems {
		h.logger.Info("Creating consolidated work item",
			"agencyID", agencyID,
			"workItemIndex", i+1,
			"workItemCode", consolidatedWorkItem.SuggestedCode,
			"consolidatedFrom", len(consolidatedWorkItem.ConsolidatedFrom))

		req := agency.CreateWorkItemRequest{
			Title:        consolidatedWorkItem.Title,
			Description:  consolidatedWorkItem.Description,
			Deliverables: consolidatedWorkItem.Deliverables,
			Tags:         consolidatedWorkItem.SuggestedTags,
		}

		savedWorkItem, err := h.agencyService.CreateWorkItem(c.Request.Context(), agencyID, req)
		if err != nil {
			h.logger.Error("Failed to create consolidated work item",
				"agencyID", agencyID,
				"workItemIndex", i+1,
				"workItemCode", consolidatedWorkItem.SuggestedCode,
				"error", err)
			// Continue with other work items even if one fails
			continue
		}

		h.logger.Info("Consolidated work item created successfully",
			"agencyID", agencyID,
			"workItemKey", savedWorkItem.Key,
			"workItemCode", savedWorkItem.Code)

		consolidatedWorkItems = append(consolidatedWorkItems, savedWorkItem)
	}

	h.logger.Info("Completed work item consolidation",
		"agencyID", agencyID,
		"originalCount", len(existingWorkItems),
		"finalCount", len(consolidatedWorkItems),
		"deleted", len(consolidationResult.RemovedWorkItems))

	results["consolidate_success"] = fmt.Sprintf("Consolidated from %d to %d work items",
		len(existingWorkItems), len(consolidatedWorkItems))
	results["consolidation_summary"] = consolidationResult.Summary
	results["ai_explanation"] = consolidationResult.Explanation
	results["removed_count"] = len(consolidationResult.RemovedWorkItems)
	results["new_count"] = len(consolidatedWorkItems)

	// Add consolidated work items to response
	if len(consolidatedWorkItems) > 0 {
		*createdWorkItems = append(*createdWorkItems, consolidatedWorkItems...)
	}
}

func (h *Handler) addWorkItemExplanationToChat(c *gin.Context, agencyID string, explanation string, createdWorkItemsCount int) {
	h.logger.Info("Attempting to add AI explanation to chat",
		"agencyID", agencyID,
		"createdWorkItemsCount", createdWorkItemsCount,
		"explanationLength", len(explanation))

	// Get or create conversation
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		h.logger.Warn("No conversation exists, creating new one",
			"agencyID", agencyID,
			"error", err)
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation for AI work item generation message")
			return
		}
	}

	if conversation == nil {
		h.logger.Error("Conversation is nil after creation attempt", "agencyID", agencyID)
		return
	}

	// Format explanation as bullet points
	formattedExplanation := h.formatExplanationAsBullets(explanation)

	// Build appropriate message based on whether work items were created
	var chatMessage string
	if createdWorkItemsCount > 0 {
		chatMessage = fmt.Sprintf("✨ **Created %d Work Items**\n\n%s", createdWorkItemsCount, formattedExplanation)
	} else {
		chatMessage = fmt.Sprintf("✨ **Work Item Analysis**\n\n%s", formattedExplanation)
	}

	h.logger.Info("Adding message to chat",
		"agencyID", agencyID,
		"conversationID", conversation.ID,
		"messageLength", len(chatMessage))

	if addErr := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add work item generation explanation to chat")
	} else {
		h.logger.Info("Successfully added AI explanation to chat",
			"agencyID", agencyID,
			"conversationID", conversation.ID)
	}
}
