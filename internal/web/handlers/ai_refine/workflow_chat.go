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

// WorkflowsResponse captures the JSON response from RefineWorkflows
type WorkflowsResponse struct {
	Action         string `json:"action"`
	Explanation    string `json:"explanation"`
	NoActionNeeded bool   `json:"no_action_needed"`
}

// ProcessWorkflowsChatRequestStreaming handles chat-based workflow interactions with streaming
// Uses real AI streaming exactly like goals and work items for consistency
func (h *Handler) ProcessWorkflowsChatRequestStreaming(c *gin.Context) {
	h.logger.Info("🌊 HANDLER CALLED: ProcessWorkflowsChatRequestStreaming")

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
		WorkflowKeys []string `json:"workflow_keys"`
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
		"agency_id":     agencyID,
		"user_message":  userMessage,
		"workflow_keys": req.WorkflowKeys,
	}).Info("Processing streaming chat-based workflow request")

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

	// Build RefineWorkflowsRequest
	existingWorkflows := make([]*models.Workflow, len(spec.Workflows))
	for i := range spec.Workflows {
		existingWorkflows[i] = &spec.Workflows[i]
	}

	var targetWorkflows []*models.Workflow
	if len(req.WorkflowKeys) > 0 {
		workflowKeyMap := make(map[string]bool)
		for _, key := range req.WorkflowKeys {
			workflowKeyMap[key] = true
		}
		for _, workflow := range existingWorkflows {
			if workflowKeyMap[workflow.Key] {
				targetWorkflows = append(targetWorkflows, workflow)
			}
		}
	}

	goals := make([]*models.Goal, len(spec.Goals))
	for i := range spec.Goals {
		goals[i] = &spec.Goals[i]
	}

	workItems := make([]*models.WorkItem, len(spec.WorkItems))
	for i := range spec.WorkItems {
		workItems[i] = &spec.WorkItems[i]
	}

	refineReq := &builder.RefineWorkflowsRequest{
		AgencyID:          agencyID,
		UserMessage:       userMessage,
		TargetWorkflows:   targetWorkflows,
		ExistingWorkflows: existingWorkflows,
		Goals:             goals,
		WorkItems:         workItems,
		AgencyContext:     ag,
	}

	// Setup SSE
	h.setupSSE(c)

	// Stream refinement (real AI streaming from backend, exactly like goals and work items)
	chunkCount := 0
	totalChunkBytes := 0

	h.logger.Info("🔍 DEBUG: Starting workflow refinement stream")

	result, err := h.workflowBuilder.RefineWorkflowsStream(
		c.Request.Context(),
		refineReq,
		builderContextData,
		func(chunk string) error {
			chunkCount++
			chunkBytes := len(chunk)
			totalChunkBytes += chunkBytes

			// Log every 10 chunks to track progress
			if chunkCount%10 == 0 {
				h.logger.WithFields(logrus.Fields{
					"chunk_number":  chunkCount,
					"chunk_bytes":   chunkBytes,
					"total_bytes":   totalChunkBytes,
					"chunk_preview": truncateForLog(chunk, 50),
				}).Debug("🔍 DEBUG: Forwarding workflow chunk to SSE")
			}

			c.SSEvent("chunk", chunk)
			c.Writer.Flush()
			return nil
		},
	)

	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"total_chunks": chunkCount,
			"total_bytes":  totalChunkBytes,
		}).Error("❌ Streaming workflow refinement failed")
		c.SSEvent("error", fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"total_chunks": chunkCount,
		"total_bytes":  totalChunkBytes,
	}).Info("✅ Streaming completed")

	// Save changes if any were made
	if !result.NoActionNeeded {
		h.logger.Info("Workflows were modified, applying and saving changes...")
		if err := h.applyAndSaveWorkflows(c.Request.Context(), agencyID, result, existingWorkflows); err != nil {
			h.logger.WithError(err).Error("Failed to save workflows")
			c.SSEvent("error", fmt.Sprintf(`{"error": "Failed to save workflows: %s"}`, err.Error()))
			return
		}
	}

	// Format message for conversation history
	workflowsResp := WorkflowsResponse{
		Action:         result.Action,
		Explanation:    result.Explanation,
		NoActionNeeded: result.NoActionNeeded,
	}
	chatMessage := formatWorkflowsChatMessage(workflowsResp)

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

	h.logger.Info("✅ Streaming workflows chat completed")
}

// formatWorkflowsChatMessage formats the workflows AI response for chat display
func formatWorkflowsChatMessage(resp WorkflowsResponse) string {
	var message strings.Builder

	// Add emoji and title based on action
	if resp.NoActionNeeded {
		message.WriteString("✅ **Workflows Review Complete**\n\n")
	} else {
		switch resp.Action {
		case "refine":
			message.WriteString("✨ **Workflows Refined**\n\n")
		case "generate":
			message.WriteString("🎯 **Workflows Generated**\n\n")
		case "consolidate":
			message.WriteString("📊 **Workflows Consolidated**\n\n")
		case "enhance_all":
			message.WriteString("⚡ **Workflows Enhanced**\n\n")
		case "under_construction":
			message.WriteString("🚧 **Feature Under Construction**\n\n")
		default:
			message.WriteString("✨ **Workflows Updated**\n\n")
		}
	}

	// Add explanation
	if resp.Explanation != "" {
		message.WriteString(resp.Explanation)
	}

	return message.String()
}

// applyAndSaveWorkflows applies the AI recommendations and saves workflows to the database
// Handles all actions: refine, generate, consolidate, remove, enhance_all
func (h *Handler) applyAndSaveWorkflows(ctx context.Context, agencyID string, result *builder.RefineWorkflowsResponse, existingWorkflows []*models.Workflow) error {
	h.logger.WithFields(logrus.Fields{
		"action":             result.Action,
		"refined_count":      len(result.RefinedWorkflows),
		"generated_count":    len(result.GeneratedWorkflows),
		"existing_workflows": len(existingWorkflows),
	}).Info("🔧 Applying workflow changes")

	var updatedWorkflows []models.Workflow
	workflowsModified := false

	switch result.Action {
	case "refine", "enhance_all":
		// Create a map of refined workflows by original key for quick lookup
		refinedMap := make(map[string]*builder.RefinedWorkflowResult)
		for i := range result.RefinedWorkflows {
			refinedMap[result.RefinedWorkflows[i].OriginalKey] = &result.RefinedWorkflows[i]
		}

		// Apply refinements to existing workflows
		for _, workflow := range existingWorkflows {
			if refined, exists := refinedMap[workflow.Key]; exists && refined.WasChanged {
				h.logger.WithFields(logrus.Fields{
					"key":      workflow.Key,
					"old_name": workflow.Name,
					"new_name": refined.RefinedName,
				}).Info("✏️ Refining workflow")

				updatedWorkflow := *workflow
				updatedWorkflow.Name = refined.RefinedName
				updatedWorkflow.Description = refined.RefinedDescription
				if len(refined.RefinedNodes) > 0 {
					updatedWorkflow.Nodes = refined.RefinedNodes
				}
				if len(refined.RefinedEdges) > 0 {
					updatedWorkflow.Edges = refined.RefinedEdges
				}

				updatedWorkflows = append(updatedWorkflows, updatedWorkflow)
				workflowsModified = true
			} else {
				// Keep workflow unchanged
				updatedWorkflows = append(updatedWorkflows, *workflow)
			}
		}

	case "generate":
		// Keep all existing workflows
		for _, workflow := range existingWorkflows {
			updatedWorkflows = append(updatedWorkflows, *workflow)
		}

		// Add generated workflows
		for _, gwf := range result.GeneratedWorkflows {
			workflowKey := generateWorkflowKey(gwf.Name)
			h.logger.WithFields(logrus.Fields{
				"name": gwf.Name,
				"key":  workflowKey,
			}).Info("🆕 Adding generated workflow")

			newWorkflow := models.Workflow{
				Key:         workflowKey,
				AgencyID:    agencyID,
				Name:        gwf.Name,
				Description: gwf.Description,
				Version:     gwf.Version,
				Nodes:       gwf.Nodes,
				Edges:       gwf.Edges,
			}

			updatedWorkflows = append(updatedWorkflows, newWorkflow)
			workflowsModified = true
		}

	case "consolidate", "remove":
		if result.ConsolidatedData != nil {
			// Create a set of removed workflow keys for quick lookup
			removedKeys := make(map[string]bool)
			for _, removedKey := range result.ConsolidatedData.RemovedWorkflows {
				removedKeys[removedKey] = true
				h.logger.Info("🔍 DEBUG: Marking workflow for removal", "key", removedKey)
			}

			h.logger.WithFields(logrus.Fields{
				"total_existing_workflows": len(existingWorkflows),
				"workflows_to_remove":      len(removedKeys),
			}).Info("🔍 DEBUG: Processing workflow removal/consolidation")

			// Keep workflows that are NOT in the removed list
			for _, wf := range existingWorkflows {
				if !removedKeys[wf.Key] {
					updatedWorkflows = append(updatedWorkflows, *wf)
					h.logger.Info("🔍 DEBUG: Keeping workflow", "key", wf.Key, "name", wf.Name)
				} else {
					h.logger.Info("🗑️ Removing workflow", "key", wf.Key, "name", wf.Name)
					workflowsModified = true
				}
			}

			// Add consolidated workflows (these are new or updated workflows)
			for _, cwf := range result.ConsolidatedData.ConsolidatedWorkflows {
				workflowKey := generateWorkflowKey(cwf.Name)
				h.logger.WithFields(logrus.Fields{
					"name": cwf.Name,
					"key":  workflowKey,
				}).Info("🔄 Adding consolidated workflow")

				newWorkflow := models.Workflow{
					Key:         workflowKey,
					AgencyID:    agencyID,
					Name:        cwf.Name,
					Description: cwf.Description,
					Version:     cwf.Version,
					Nodes:       cwf.Nodes,
					Edges:       cwf.Edges,
				}

				updatedWorkflows = append(updatedWorkflows, newWorkflow)
				workflowsModified = true
			}
		}
	}

	// Save the updated workflows list if modified
	if workflowsModified {
		h.logger.WithFields(logrus.Fields{
			"previous_count": len(existingWorkflows),
			"updated_count":  len(updatedWorkflows),
		}).Info("💾 Saving updated workflows to database")

		_, err := h.agencyService.UpdateSpecificationWorkflows(ctx, agencyID, updatedWorkflows, "ai-refine")
		if err != nil {
			h.logger.WithError(err).Error("❌ Failed to save workflows to database")
			return fmt.Errorf("failed to save workflows: %w", err)
		}
		h.logger.Info("✅ Successfully saved workflows to database")
	} else {
		h.logger.Info("ℹ️ No workflows modifications needed")
	}

	return nil
}

// generateWorkflowKey creates a key from a workflow name
func generateWorkflowKey(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	key := strings.ToLower(name)
	key = strings.ReplaceAll(key, " ", "-")
	// Remove special characters (keep only alphanumeric and hyphens)
	var cleaned strings.Builder
	for _, r := range key {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}
	result := cleaned.String()
	// Limit to reasonable length
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

// truncateForLog returns a truncated version of a string for logging
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
