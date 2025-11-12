package ai_refine

import (
	"context"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WorkItemsResponse captures the JSON response from RefineWorkItems
type WorkItemsResponse struct {
	Action         string `json:"action"`
	Explanation    string `json:"explanation"`
	NoActionNeeded bool   `json:"no_action_needed"`
}

// ProcessWorkItemsChatRequestStreaming handles chat-based work item interactions with streaming
// Uses real AI streaming exactly like goals for consistency
func (h *Handler) ProcessWorkItemsChatRequestStreaming(c *gin.Context) {
	h.logger.Info("🌊 HANDLER CALLED: ProcessWorkItemsChatRequestStreaming")

	agencyID := c.Param("id")

	// Get user message from dynamic_request (set by chat_context_processor)
	dynamicReq, exists := c.Get("dynamic_request")
	if !exists {
		h.logger.Error("No dynamic_request found in context")
		c.SSEvent("error", `{"error": "Missing request data"}`)
		return
	}

	req, ok := dynamicReq.(struct {
		UserMessage  string   `json:"user_message"`
		WorkItemKeys []string `json:"work_item_keys"`
	})
	if !ok {
		h.logger.Error("Failed to cast dynamic_request to expected type")
		c.SSEvent("error", `{"error": "Invalid request format"}`)
		return
	}

	userMessage := req.UserMessage
	if userMessage == "" {
		h.logger.Error("No user message provided")
		c.SSEvent("error", `{"error": "No message provided"}`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":      agencyID,
		"user_message":   userMessage,
		"work_item_keys": req.WorkItemKeys,
	}).Info("Processing streaming chat-based work item request")

	// Fetch agency and specification
	ag, spec, err := h.fetchAgencyAndSpec(c, agencyID)
	if err != nil {
		c.SSEvent("error", `{"error": "Agency not found"}`)
		return
	}

	// Get or create conversation
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		h.logger.Warn("No conversation exists, creating new one", "agencyID", agencyID)
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation")
			c.SSEvent("error", `{"error": "Failed to initialize conversation"}`)
			return
		}
	}

	// Build AI context
	builderContextData, err := h.contextBuilder.BuildBuilderContext(
		c.Request.Context(),
		ag,
		spec.Introduction,
		userMessage,
	)
	if err != nil {
		c.SSEvent("error", `{"error": "Failed to build context"}`)
		return
	}

	// Build RefineWorkItemsRequest
	existingWorkItems := make([]*models.WorkItem, len(spec.WorkItems))
	for i := range spec.WorkItems {
		existingWorkItems[i] = &spec.WorkItems[i]
	}

	var targetWorkItems []*models.WorkItem
	if len(req.WorkItemKeys) > 0 {
		workItemKeyMap := make(map[string]bool)
		for _, key := range req.WorkItemKeys {
			workItemKeyMap[key] = true
		}
		for _, workItem := range existingWorkItems {
			if workItemKeyMap[workItem.Key] {
				targetWorkItems = append(targetWorkItems, workItem)
			}
		}
	}

	goals := make([]*models.Goal, len(spec.Goals))
	for i := range spec.Goals {
		goals[i] = &spec.Goals[i]
	}

	refineReq := &builder.RefineWorkItemsRequest{
		AgencyID:          agencyID,
		UserMessage:       userMessage,
		TargetWorkItems:   targetWorkItems,
		ExistingWorkItems: existingWorkItems,
		Goals:             goals,
		AgencyContext:     ag,
	}

	// Setup SSE
	h.setupSSE(c)

	// Stream refinement (real AI streaming from backend, exactly like goals)
	chunkCount := 0
	result, err := h.workItemBuilder.RefineWorkItemsStream(
		c.Request.Context(),
		refineReq,
		builderContextData,
		func(chunk string) error {
			chunkCount++
			c.SSEvent("chunk", chunk)
			c.Writer.Flush()
			return nil
		},
	)

	if err != nil {
		h.logger.WithError(err).Error("❌ Streaming work item refinement failed")
		c.SSEvent("error", fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}

	h.logger.WithField("total_chunks", chunkCount).Info("✅ Streaming completed")

	// Save changes if any were made
	if !result.NoActionNeeded {
		h.logger.Info("Work items were modified, applying and saving changes...")
		if err := h.applyAndSaveWorkItems(c.Request.Context(), agencyID, result, existingWorkItems); err != nil {
			h.logger.WithError(err).Error("Failed to save work items")
			c.SSEvent("error", fmt.Sprintf(`{"error": "Failed to save work items: %s"}`, err.Error()))
			return
		}
	}

	// Format message for conversation history
	workItemsResp := WorkItemsResponse{
		Action:         result.Action,
		Explanation:    result.Explanation,
		NoActionNeeded: result.NoActionNeeded,
	}
	chatMessage := formatWorkItemsChatMessage(workItemsResp)

	// Add to conversation
	if err := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); err != nil {
		h.logger.WithError(err).Error("Failed to add message to conversation")
	}

	// Send completion
	completionData := map[string]interface{}{
		"was_changed":     !result.NoActionNeeded,
		"explanation":     result.Explanation,
		"message":         chatMessage,
		"conversation_id": conversation.ID,
	}

	c.SSEvent("complete", completionData)
	c.Writer.Flush()

	h.logger.Info("✅ Streaming work items chat completed")
}

// formatWorkItemsChatMessage formats the work items AI response for chat display
func formatWorkItemsChatMessage(resp WorkItemsResponse) string {
	var message strings.Builder

	// Add emoji and title based on action
	if resp.NoActionNeeded {
		message.WriteString("✅ **Work Items Review Complete**\n\n")
	} else {
		switch resp.Action {
		case "refine":
			message.WriteString("✨ **Work Items Refined**\n\n")
		case "generate":
			message.WriteString("🎯 **Work Items Generated**\n\n")
		case "consolidate":
			message.WriteString("📊 **Work Items Consolidated**\n\n")
		case "enhance_all":
			message.WriteString("⚡ **Work Items Enhanced**\n\n")
		case "under_construction":
			message.WriteString("🚧 **Feature Under Construction**\n\n")
		default:
			message.WriteString("✨ **Work Items Updated**\n\n")
		}
	}

	// Add explanation
	if resp.Explanation != "" {
		message.WriteString(resp.Explanation)
	}

	return message.String()
}

// applyAndSaveWorkItems applies the AI recommendations and saves work items to the database
// Handles all actions: refine, generate, consolidate, remove, enhance_all
func (h *Handler) applyAndSaveWorkItems(ctx context.Context, agencyID string, result *builder.RefineWorkItemsResponse, existingWorkItems []*models.WorkItem) error {
	h.logger.WithFields(logrus.Fields{
		"action":              result.Action,
		"refined_count":       len(result.RefinedWorkItems),
		"generated_count":     len(result.GeneratedWorkItems),
		"existing_work_items": len(existingWorkItems),
	}).Info("🔧 Applying work item changes")

	var updatedWorkItems []models.WorkItem
	workItemsModified := false

	switch result.Action {
	case "refine", "enhance_all":
		// Create a map of refined work items by original key for quick lookup
		refinedMap := make(map[string]*builder.RefinedWorkItemResult)
		for i := range result.RefinedWorkItems {
			refinedMap[result.RefinedWorkItems[i].OriginalKey] = &result.RefinedWorkItems[i]
		}

		// Apply refinements to existing work items
		for _, workItem := range existingWorkItems {
			if refined, exists := refinedMap[workItem.Key]; exists && refined.WasChanged {
				h.logger.WithFields(logrus.Fields{
					"key":       workItem.Key,
					"old_title": workItem.Title,
					"new_title": refined.RefinedTitle,
					"old_code":  workItem.Code,
					"new_code":  refined.SuggestedCode,
				}).Info("✏️ Refining work item")

				updatedWorkItem := *workItem
				updatedWorkItem.Title = refined.RefinedTitle
				updatedWorkItem.Description = refined.RefinedDescription
				updatedWorkItem.Deliverables = refined.RefinedDeliverables
				if refined.SuggestedCode != "" {
					updatedWorkItem.Code = refined.SuggestedCode
				}
				if len(refined.SuggestedTags) > 0 {
					updatedWorkItem.Tags = refined.SuggestedTags
				}
				updatedWorkItems = append(updatedWorkItems, updatedWorkItem)
				workItemsModified = true
			} else {
				// Keep work item unchanged
				updatedWorkItems = append(updatedWorkItems, *workItem)
			}
		}

	case "generate":
		// Keep all existing work items
		for _, workItem := range existingWorkItems {
			updatedWorkItems = append(updatedWorkItems, *workItem)
		}

		// Add generated work items
		for _, gwi := range result.GeneratedWorkItems {
			h.logger.WithFields(logrus.Fields{
				"code":  gwi.SuggestedCode,
				"title": gwi.Title,
			}).Info("🆕 Adding generated work item")

			newWorkItem := models.WorkItem{
				Code:         gwi.SuggestedCode,
				Title:        gwi.Title,
				Description:  gwi.Description,
				Deliverables: gwi.Deliverables,
				Tags:         gwi.SuggestedTags,
			}
			updatedWorkItems = append(updatedWorkItems, newWorkItem)
			workItemsModified = true
		}

	case "consolidate", "remove":
		if result.ConsolidatedData != nil {
			// Create a set of removed work item keys for quick lookup
			removedKeys := make(map[string]bool)
			for _, removedKey := range result.ConsolidatedData.RemovedWorkItems {
				removedKeys[removedKey] = true
				h.logger.Info("🔍 DEBUG: Marking work item for removal", "key", removedKey)
			}

			h.logger.WithFields(logrus.Fields{
				"total_existing_work_items": len(existingWorkItems),
				"work_items_to_remove":      len(removedKeys),
			}).Info("🔍 DEBUG: Processing work item removal/consolidation")

			// Keep work items that are NOT in the removed list
			for _, wi := range existingWorkItems {
				if !removedKeys[wi.Key] {
					updatedWorkItems = append(updatedWorkItems, *wi)
					h.logger.Info("🔍 DEBUG: Keeping work item", "key", wi.Key, "code", wi.Code)
				} else {
					h.logger.Info("🗑️ Removing work item", "key", wi.Key, "code", wi.Code)
					workItemsModified = true
				}
			}

			// Add consolidated work items (these are new or updated work items)
			for _, cwi := range result.ConsolidatedData.ConsolidatedWorkItems {
				h.logger.WithFields(logrus.Fields{
					"code":  cwi.SuggestedCode,
					"title": cwi.Title,
				}).Info("🔄 Adding consolidated work item")

				newWorkItem := models.WorkItem{
					Code:         cwi.SuggestedCode,
					Title:        cwi.Title,
					Description:  cwi.Description,
					Deliverables: cwi.Deliverables,
					Tags:         cwi.SuggestedTags,
				}
				updatedWorkItems = append(updatedWorkItems, newWorkItem)
				workItemsModified = true
			}

			h.logger.WithFields(logrus.Fields{
				"work_items_modified":   workItemsModified,
				"final_work_item_count": len(updatedWorkItems),
			}).Info("🔍 DEBUG: Consolidation/removal complete")
		}
	}

	// Save the updated work items list if modified
	if workItemsModified {
		h.logger.WithFields(logrus.Fields{
			"previous_count": len(existingWorkItems),
			"updated_count":  len(updatedWorkItems),
		}).Info("💾 Saving updated work items to database")

		_, err := h.agencyService.UpdateSpecificationWorkItems(ctx, agencyID, updatedWorkItems, "ai-refine")
		if err != nil {
			h.logger.WithError(err).Error("❌ Failed to save work items to database")
			return fmt.Errorf("failed to save work items: %w", err)
		}
		h.logger.Info("✅ Successfully saved work items to database")
	} else {
		h.logger.Info("ℹ️ No work items modifications needed")
	}

	return nil
}
