package ai_refine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProcessDeliverableEnhancementStreaming handles deliverable node enhancement requests with streaming
// This directly streams the AI response back to the properties panel for deliverable node enhancement
func (h *Handler) ProcessDeliverableEnhancementStreaming(c *gin.Context, agencyID string, userMessage string, metadata map[string]interface{}) {
	h.logger.Info("📦 HANDLER CALLED: ProcessDeliverableEnhancementStreaming")

	// Extract node information from metadata
	nodeName, _ := metadata["nodeName"].(string)
	nodeType, _ := metadata["nodeType"].(string)

	if nodeName == "" {
		h.logger.Error("Node name not provided in metadata")
		c.SSEvent("error", `{"error": "Missing node information"}`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"node_name": nodeName,
		"node_type": nodeType,
	}).Info("🔍 Processing deliverable node enhancement request")

	// Setup SSE
	h.setupSSE(c)

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

	// Fetch agency for context
	ag, spec, err := h.fetchAgencyAndSpec(c, agencyID)
	if err != nil {
		c.SSEvent("error", `{"error": "Agency not found"}`)
		return
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

	// Build enhancement request to use work items builder's stream capability
	// We'll use the work items builder's LLM client but with the deliverables prompt
	enhancementReq := &builder.RefineWorkItemsRequest{
		AgencyID:      agencyID,
		UserMessage:   userMessage,
		AgencyContext: ag,
	}

	// Stream the AI response using the deliverables enhancement prompt
	chunkCount := 0
	var fullResponse string

	// Use the work item builder's stream capability with deliverables prompt
	_, err = h.workItemBuilder.RefineWorkItemsStream(
		c.Request.Context(),
		enhancementReq,
		builderContextData,
		func(chunk string) error {
			chunkCount++
			fullResponse += chunk // Accumulate the full response
			c.SSEvent("chunk", chunk)
			c.Writer.Flush()
			return nil
		},
	)

	if err != nil {
		h.logger.WithError(err).Error("❌ Streaming deliverable enhancement failed")
		c.SSEvent("error", fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}

	h.logger.WithField("total_chunks", chunkCount).Info("✅ Streaming completed")

	// Parse the JSON response from the AI
	enhancementJSON, err := extractJSONFromResponse(fullResponse)
	if err != nil {
		h.logger.WithError(err).Warn("⚠️  Failed to extract JSON from AI response")
		// Continue without saving - still show the response to user
	} else {
		// Parse the enhancement data
		var enhancement DeliverableEnhancement
		if err := json.Unmarshal([]byte(enhancementJSON), &enhancement); err != nil {
			h.logger.WithError(err).Error("❌ Failed to parse enhancement JSON")
		} else {
			// Save the enhancement to the database
			if err := h.saveDeliverableEnhancement(c.Request.Context(), agencyID, nodeName, &enhancement); err != nil {
				h.logger.WithError(err).Error("❌ Failed to save deliverable enhancement")
			} else {
				h.logger.WithFields(logrus.Fields{
					"node_name":   nodeName,
					"was_changed": enhancement.WasChanged,
				}).Info("✅ Deliverable enhancement saved successfully")
			}
		}
	}

	// Format message for conversation
	chatMessage := fmt.Sprintf("✨ **Deliverable Enhanced**: %s\n\nEnhanced deliverable structure returned as JSON.", nodeName)

	// Add to conversation
	if err := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); err != nil {
		h.logger.WithError(err).Error("Failed to add message to conversation")
	}

	// Send completion - the enhancement detector will pick up the JSON from the stream
	completionData := map[string]interface{}{
		"was_changed":     true, // Assume enhancement was made
		"explanation":     chatMessage,
		"message":         chatMessage,
		"conversation_id": conversation.ID,
	}

	c.SSEvent("complete", completionData)
	c.Writer.Flush()

	h.logger.Info("✅ Streaming deliverable enhancement completed")
}

// DeliverableEnhancement represents the AI's enhancement response
type DeliverableEnhancement struct {
	Name                   string                   `json:"name"`
	Description            string                   `json:"description"`
	PromptInstructions     string                   `json:"prompt_instructions"`
	SuggestedChildren      []models.DeliverableNode `json:"suggested_children"`
	EnhancementExplanation string                   `json:"enhancement_explanation"`
	WasChanged             bool                     `json:"was_changed"`
}

// extractJSONFromResponse extracts JSON from a code fence in the AI response
func extractJSONFromResponse(response string) (string, error) {
	// Look for ```json ... ``` code fence
	re := regexp.MustCompile("```json\\s*\\n([\\s\\S]*?)\\n```")
	matches := re.FindStringSubmatch(response)

	if len(matches) < 2 {
		return "", fmt.Errorf("no JSON code fence found in response")
	}

	return strings.TrimSpace(matches[1]), nil
}

// saveDeliverableEnhancement saves the enhanced deliverable to the database
func (h *Handler) saveDeliverableEnhancement(ctx context.Context, agencyID, nodeName string, enhancement *DeliverableEnhancement) error {
	// Get the agency and specification using the existing helper
	// We need a gin.Context for fetchAgencyAndSpec, but we only have context.Context
	// So we'll use the agency service directly
	spec, err := h.agencyService.GetSpecification(ctx, agencyID)
	if err != nil {
		return fmt.Errorf("failed to get specification: %w", err)
	}

	if spec == nil {
		return fmt.Errorf("agency specification not found")
	}

	// Find the work item that contains this deliverable
	// We need to search through all work items' DeliverablesStructured trees
	var targetWorkItem *models.WorkItem
	var targetNode *models.DeliverableNode

	for i := range spec.WorkItems {
		workItem := &spec.WorkItems[i]

		// Search in the deliverables tree
		for j := range workItem.DeliverablesStructured {
			node := findNodeByName(&workItem.DeliverablesStructured[j], nodeName)
			if node != nil {
				targetWorkItem = workItem
				targetNode = node
				break
			}
		}

		if targetNode != nil {
			break
		}
	}

	if targetWorkItem == nil || targetNode == nil {
		return fmt.Errorf("deliverable node '%s' not found in any work item", nodeName)
	}

	h.logger.WithFields(logrus.Fields{
		"work_item_code": targetWorkItem.Code,
		"node_name":      nodeName,
	}).Info("🎯 Found deliverable node in work item")

	// Update the node with the enhancement
	if enhancement.WasChanged {
		targetNode.Name = enhancement.Name
		targetNode.Description = enhancement.Description
		targetNode.PromptInstructions = enhancement.PromptInstructions

		// Update or add children if suggested
		if len(enhancement.SuggestedChildren) > 0 {
			targetNode.Children = enhancement.SuggestedChildren
			// Ensure all new children have IDs
			for i := range targetNode.Children {
				targetNode.Children[i].EnsureAllIDs()
			}
		}

		// Recompute paths for the entire tree
		for i := range targetWorkItem.DeliverablesStructured {
			targetWorkItem.DeliverablesStructured[i].ComputeAllPaths("")
		}

		h.logger.WithFields(logrus.Fields{
			"updated_name":       enhancement.Name,
			"children_count":     len(targetNode.Children),
			"suggested_children": len(enhancement.SuggestedChildren),
		}).Info("📝 Updated deliverable node")
	}

	// Save the updated work item back to the database
	updatedWorkItems := make([]models.WorkItem, len(spec.WorkItems))
	copy(updatedWorkItems, spec.WorkItems)

	_, err = h.agencyService.UpdateSpecificationWorkItems(ctx, agencyID, updatedWorkItems, "deliverable-enhancement")
	if err != nil {
		return fmt.Errorf("failed to save updated work items: %w", err)
	}

	h.logger.Info("💾 Saved updated work items to database")
	return nil
}

// findNodeByName recursively searches for a node by name in the deliverable tree
func findNodeByName(node *models.DeliverableNode, name string) *models.DeliverableNode {
	if node.Name == name {
		return node
	}

	for i := range node.Children {
		if found := findNodeByName(&node.Children[i], name); found != nil {
			return found
		}
	}

	return nil
}
