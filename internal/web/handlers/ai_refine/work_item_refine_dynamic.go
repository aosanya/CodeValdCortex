package ai_refine

import (
	"fmt"
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RefineWorkItems handles POST /api/v1/agencies/:id/work-items/refine-dynamic
// Dynamically determines and executes the appropriate work item operation based on user message
func (h *Handler) RefineWorkItems(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithField("agency_id", agencyID).Info("Processing dynamic AI work item refinement request")

	// Check if this is a wrapper call with a preset request or from chat
	var req struct {
		UserMessage  string   `json:"user_message"`        // Natural language instruction
		WorkItemKeys []string `json:"work_item_keys"`      // Optional: specific work items to operate on
	}

	// Check multiple sources for the user message
	// 1. Preset request from wrapper methods
	if dynamicReq, exists := c.Get("dynamic_request"); exists {
		if presetReq, ok := dynamicReq.(struct {
			UserMessage  string   `json:"user_message"`
			WorkItemKeys []string `json:"work_item_keys"`
		}); ok {
			req.UserMessage = presetReq.UserMessage
			req.WorkItemKeys = presetReq.WorkItemKeys
			h.logger.WithField("source", "wrapper").Info("Using preset request from wrapper method")
		}
	} else if userRequest := c.PostForm("user-request"); userRequest != "" {
		// 2. From chat form (user-request field)
		req.UserMessage = userRequest
		h.logger.WithField("source", "chat_form").Info("Using user request from chat form")
	} else if message := c.PostForm("message"); message != "" {
		// 3. From chat message field
		req.UserMessage = message
		h.logger.WithField("source", "chat_message").Info("Using message from chat")
	} else {
		// 4. Parse JSON request body for direct API calls
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
							<p class="mb-0">Please provide a message describing what you want to do with the work items.</p>
						</div>
					</div>
				</div>
			`)
			return
		}
	}

	if req.UserMessage == "" {
		h.logger.Error("No user message found in request")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Missing Message</strong>
						<p class="mb-0">Please provide a message describing what you want to do.</p>
					</div>
				</div>
			</div>
		`)
		return
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

	// Get all existing work items for context
	existingWorkItems, err := h.agencyService.GetWorkItems(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch existing work items")
		existingWorkItems = []*agency.WorkItem{}
	}

	// Filter target work items if specific keys were provided
	var targetWorkItems []*agency.WorkItem
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
		h.logger.WithFields(logrus.Fields{
			"requested_keys": len(req.WorkItemKeys),
			"found_items":    len(targetWorkItems),
		}).Info("Filtered work items by keys")
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":           agencyID,
		"user_message":        req.UserMessage,
		"target_work_items":   len(targetWorkItems),
		"existing_work_items": len(existingWorkItems),
	}).Info("Starting dynamic work item refinement")

	// Build AI context data using shared context builder (will be used when RefineWorkItems is implemented)
	_, err = h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, "", req.UserMessage)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build context")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-triangle"></i>
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

	// TODO: Implement RefineWorkItems method in WorkItemsBuilder
	// For now, add a message to the conversation indicating the request was received
	h.logger.Info("Work item refinement placeholder - RefineWorkItems method not yet implemented")

	// Get or create conversation for this agency
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		h.logger.Warn("No conversation exists for work items processing, creating new one",
			"agencyID", agencyID,
			"error", err)
		// No conversation exists, create one
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation for work items processing")
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
				<div class="notification is-danger">
					<div class="is-flex is-align-items-center">
						<span class="icon has-text-danger mr-2">
							<i class="fas fa-exclamation-triangle"></i>
						</span>
						<div>
							<strong>Conversation Error</strong>
							<p class="mb-0">Failed to create conversation for processing.</p>
						</div>
					</div>
				</div>
			`)
			return
		}
	}

	// Add user message to conversation
	if addErr := h.designerService.AddMessage(conversation.ID, "user", req.UserMessage); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add user message to conversation")
	}

	// Prepare AI response message
	responseMessage := "📋 **Work Item Request Received**\n\n" +
		"I've received your request to work on items. The full AI-powered work item processing is being implemented.\n\n" +
		"**Your request:** " + req.UserMessage + "\n\n" +
		"Once implemented, I'll be able to:\n" +
		"• Add new work items based on your description\n" +
		"• Refine existing work items for clarity\n" +
		"• Generate work items from agency goals\n" +
		"• Consolidate duplicate items\n\n" +
		"Thank you for your patience! 🚀"

	// Add AI response to conversation
	if addErr := h.designerService.AddMessage(conversation.ID, "assistant", responseMessage); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add AI response to conversation")
	}

	h.logger.Info("Work items request added to conversation",
		"agencyID", agencyID,
		"conversationID", conversation.ID)

	// Return success - chat will be refreshed via HTMX trigger
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
		<div class="ai-refine-response" 
			hx-trigger="load delay:100ms" 
			hx-get="/agencies/%s/chat-messages?agencyName=%s"
			hx-target="#chat-messages" 
			hx-swap="innerHTML">
		</div>
	`, agencyID, ag.Name))
}

