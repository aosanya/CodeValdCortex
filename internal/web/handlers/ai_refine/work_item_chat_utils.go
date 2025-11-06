package ai_refine

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProcessWorkItemChatRequest handles chat-based work item interactions
// This is similar to ProcessGoalChatRequest but for work items in chat context
func (h *Handler) ProcessWorkItemChatRequest(c *gin.Context) {
	h.logger.Info("🔵 HANDLER CALLED: ProcessWorkItemChatRequest")

	agencyID := c.Param("id")

	// Get user's chat message/request
	userRequest := c.PostForm("user-request")
	if userRequest == "" {
		userRequest = c.PostForm("message")
	}

	if userRequest == "" {
		h.logger.Error("No user request provided for work item chat")
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
	}).Info("Processing chat-based work item request")

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

	// Get or create conversation
	conv, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		conv, err = h.designerService.StartConversation(ctx, agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation")
			c.Header("Content-Type", "text/html")
			c.String(http.StatusInternalServerError, `
				<div class="notification is-danger">
					<div class="is-flex is-align-items-center">
						<span class="icon has-text-danger mr-2">
							<i class="fas fa-exclamation-triangle"></i>
						</span>
						<div>
							<strong>Conversation Error</strong>
							<p class="mb-0">Failed to initialize conversation.</p>
						</div>
					</div>
				</div>
			`)
			return
		}
	}

	// Add user message to conversation
	if err := h.designerService.AddMessage(conv.ID, "user", userRequest); err != nil {
		h.logger.WithError(err).Error("Failed to add user message to conversation")
	}

	// Build AI context data using shared context builder (not used yet but will be needed for RefineWorkItems)
	_, err = h.contextBuilder.BuildBuilderContext(ctx, ag, "", userRequest)
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
	// For now, provide a simple response
	responseMessage := "🚧 **Work Item Processing**\n\nWork item dynamic processing is being implemented. This will support refining, generating, and consolidating work items based on your requests."

	// Format response as bullets if it contains multiple points
	responseMessage = h.formatExplanationAsBullets(responseMessage)

	// Add AI response to conversation
	if err := h.designerService.AddMessage(conv.ID, "assistant", responseMessage); err != nil {
		h.logger.WithError(err).Error("Failed to add AI response to conversation")
	}

	// Return the HTML response for HTMX
	c.Header("Content-Type", "text/html")
	workItemsCard := agency_designer.WorkItemsListCard()
	workItemsCard.Render(c.Request.Context(), c.Writer)
}
