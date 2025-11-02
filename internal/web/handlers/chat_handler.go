package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
) // ChatHandler handles web chat interactions
type ChatHandler struct {
	designerService *ai.AgencyDesignerService
	logger          *logrus.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(designerService *ai.AgencyDesignerService, logger *logrus.Logger) *ChatHandler {
	return &ChatHandler{
		designerService: designerService,
		logger:          logger,
	}
}

// SendMessage handles POST /api/v1/conversations/:conversationId/messages/web
// Returns HTML for HTMX to append to the chat
func (h *ChatHandler) SendMessage(c *gin.Context) {
	conversationID := c.Param("conversationId")
	userMessage := c.PostForm("message")

	if userMessage == "" {
		h.logger.Warn("Empty message received")
		c.String(http.StatusBadRequest, "Message cannot be empty")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"message_length":  len(userMessage),
	}).Info("Processing chat message")

	// Get AI response (this also adds the user message to the conversation)
	ctx := c.Request.Context()
	response, err := h.designerService.SendMessage(ctx, conversationID, userMessage)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get AI response")
		c.String(http.StatusInternalServerError, "Failed to process message")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"response_length": len(response.Content),
	}).Info("AI response received")

	// Get the updated conversation to retrieve the user message that was added
	conversation, err := h.designerService.GetConversation(conversationID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get conversation")
		c.String(http.StatusInternalServerError, "Failed to get conversation")
		return
	}

	// Find the user message (should be second to last, before the assistant response)
	var userMsg *ai.Message
	if len(conversation.Messages) >= 2 {
		// Get the second to last message (user message)
		userMsg = &conversation.Messages[len(conversation.Messages)-2]
	}

	// Render both user message and AI response
	c.Header("Content-Type", "text/html")

	userRole := "none"
	if userMsg != nil {
		userRole = userMsg.Role
	}

	h.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"user_msg_found":  userMsg != nil,
		"user_msg_role":   userRole,
	}).Info("Rendering messages for existing conversation")

	// Render user message if found
	if userMsg != nil && userMsg.Role == "user" {
		err = agency_designer.UserMessage(*userMsg).Render(c.Request.Context(), c.Writer)
		if err != nil {
			h.logger.WithError(err).Error("Failed to render user message")
			return
		}
		h.logger.Info("User message rendered successfully")
	}

	// Render AI response
	err = agency_designer.AIMessage(*response).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render AI message")
		return
	}
	h.logger.Info("AI message rendered successfully")
}

// StartConversation handles POST /api/v1/agencies/:id/designer/conversations/web
// Starts a new conversation and returns the first message
func (h *ChatHandler) StartConversation(c *gin.Context) {
	agencyID := c.Param("id")
	userMessage := c.PostForm("message")

	if userMessage == "" {
		h.logger.Warn("Empty message received for new conversation")
		c.String(http.StatusBadRequest, "Message cannot be empty")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":      agencyID,
		"message_length": len(userMessage),
	}).Info("Starting new conversation")

	ctx := c.Request.Context()

	// Start conversation
	conversation, err := h.designerService.StartConversation(ctx, agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to start conversation")
		c.String(http.StatusInternalServerError, "Failed to start conversation")
		return
	}

	// Get AI response (this also adds the user message to the conversation)
	response, err := h.designerService.SendMessage(ctx, conversation.ID, userMessage)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get AI response")
		c.String(http.StatusInternalServerError, "Failed to process message")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"conversation_id": conversation.ID,
		"response_length": len(response.Content),
	}).Info("Conversation started and AI response received")

	// Get the updated conversation to retrieve messages
	conversation, err = h.designerService.GetConversation(conversation.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get conversation")
		c.String(http.StatusInternalServerError, "Failed to get conversation")
		return
	}

	// Find the user message (should be second to last, before the assistant response)
	// Skip system message (first message)
	var userMsg *ai.Message
	for i := len(conversation.Messages) - 1; i >= 0; i-- {
		if conversation.Messages[i].Role == "user" {
			userMsg = &conversation.Messages[i]
			break
		}
	}

	// Render both user message and AI response
	c.Header("Content-Type", "text/html")

	userRole := "none"
	if userMsg != nil {
		userRole = userMsg.Role
	}

	h.logger.WithFields(logrus.Fields{
		"conversation_id": conversation.ID,
		"user_msg_found":  userMsg != nil,
		"user_msg_role":   userRole,
	}).Info("Rendering messages for new conversation")

	// Render user message if found
	if userMsg != nil {
		err = agency_designer.UserMessage(*userMsg).Render(c.Request.Context(), c.Writer)
		if err != nil {
			h.logger.WithError(err).Error("Failed to render user message")
			return
		}
		h.logger.Info("User message rendered successfully")
	}

	// Render AI response
	err = agency_designer.AIMessage(*response).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render AI message")
		return
	}
	h.logger.Info("AI message rendered successfully")
}
