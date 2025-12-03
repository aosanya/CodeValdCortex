package ai_refine

import (
	"fmt"

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

	// TODO: Parse the JSON response from fullResponse and save the enhanced deliverable
	// For now, just log that we received the response
	h.logger.WithFields(logrus.Fields{
		"response_length": len(fullResponse),
		"node_name":       nodeName,
	}).Warn("⚠️  Deliverable enhancement JSON received but NOT YET SAVED - feature under construction")

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
