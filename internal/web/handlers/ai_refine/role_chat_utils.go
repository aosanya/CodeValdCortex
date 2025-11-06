package ai_refine

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProcessRoleChatRequest handles chat-based role interactions
// This is similar to ProcessGoalChatRequest but for roles in chat context
func (h *Handler) ProcessRoleChatRequest(c *gin.Context) {
	h.logger.Info("🔵 HANDLER CALLED: ProcessRoleChatRequest")

	agencyID := c.Param("id")

	// Get user's chat message/request
	userRequest := c.PostForm("user-request")
	if userRequest == "" {
		userRequest = c.PostForm("message")
	}

	if userRequest == "" {
		h.logger.Error("No user request provided for role chat")
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
	}).Info("Processing chat-based role request")

	// Get agency context
	ctx := c.Request.Context()
	_, err := h.agencyService.GetAgency(ctx, agencyID)
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

	// For now, return a placeholder response until RefineRoles is implemented in the builder
	responseMessage := fmt.Sprintf(`
		<strong>Role Processing Request Received:</strong><br>
		<em>%s</em><br><br>
		<strong>Status:</strong> Role processing is under construction.<br>
		<strong>Planned Capabilities:</strong><br>
		• Refine specific roles<br>
		• Generate new roles based on work items<br>
		• Consolidate duplicate roles<br>
		• Enhance all roles with AI analysis<br><br>
		<em>This will be implemented once the RolesBuilder.RefineRoles method is complete.</em>
	`, userRequest)

	// Add AI response to conversation
	if err := h.designerService.AddMessage(conv.ID, "assistant", responseMessage); err != nil {
		h.logger.WithError(err).Error("Failed to add AI response to conversation")
	}

	// Return response
	c.Header("Content-Type", "text/html")
	responseHTML := fmt.Sprintf(`
		<div class="notification is-info">
			<div class="is-flex is-align-items-center">
				<span class="icon mr-2">
					<i class="fas fa-users"></i>
				</span>
				<div class="content">
					%s
				</div>
			</div>
		</div>
	`, responseMessage)

	c.String(http.StatusOK, responseHTML)
}
