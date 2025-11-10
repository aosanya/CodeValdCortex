package handlers

import (
	"fmt"
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/aosanya/CodeValdCortex/internal/web/handlers/ai_refine"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ChatHandler handles web chat interactions
type ChatHandler struct {
	designerService     *ai.AgencyDesignerService
	agencyService       agency.Service
	roleService         registry.RoleService
	introductionRefiner *ai.IntroductionBuilder
	goalRefiner         *ai.GoalsBuilder
	contextBuilder      *ai_refine.BuilderContextBuilder
	aiRefineHandler     *ai_refine.Handler
	logger              *logrus.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(
	designerService *ai.AgencyDesignerService,
	agencyService agency.Service,
	roleService registry.RoleService,
	introductionRefiner *ai.IntroductionBuilder,
	goalRefiner *ai.GoalsBuilder,
	aiRefineHandler *ai_refine.Handler,
	logger *logrus.Logger,
) *ChatHandler {
	// Create context builder for shared AI context gathering
	contextBuilder := ai_refine.NewBuilderContextBuilder(agencyService, roleService, logger)

	return &ChatHandler{
		designerService:     designerService,
		agencyService:       agencyService,
		roleService:         roleService,
		introductionRefiner: introductionRefiner,
		goalRefiner:         goalRefiner,
		contextBuilder:      contextBuilder,
		aiRefineHandler:     aiRefineHandler,
		logger:              logger,
	}
}

// SendMessage handles POST /api/v1/conversations/:conversationId/messages/web
// Returns HTML for HTMX to append to the chat
func (h *ChatHandler) SendMessage(c *gin.Context) {
	conversationID := c.Param("conversationId")
	userMessage := c.PostForm("message")
	context := c.PostForm("context") // Get current section context (introduction, goal-definition, work-items, roles, raci-matrix)

	fmt.Printf("\n[CHAT HANDLER] ========== REQUEST RECEIVED ==========\n")
	fmt.Printf("[CHAT HANDLER] ConversationID: %s\n", conversationID)
	fmt.Printf("[CHAT HANDLER] Message: %s\n", userMessage)
	fmt.Printf("[CHAT HANDLER] Context: '%s'\n", context)
	fmt.Printf("[CHAT HANDLER] Context length: %d\n", len(context))
	fmt.Printf("[CHAT HANDLER] Context bytes: %v\n", []byte(context))
	fmt.Printf("[CHAT HANDLER] All form values: %+v\n", c.Request.PostForm)
	fmt.Printf("[CHAT HANDLER] =====================================\n\n")

	// TEMPORARY DEBUG: Return error with context info
	if context == "" {
		errMsg := fmt.Sprintf("DEBUG: Context is EMPTY! Form values: %+v", c.Request.PostForm)
		h.logger.Error(errMsg)
		c.String(http.StatusBadRequest, errMsg)
		return
	}

	if userMessage == "" {
		h.logger.Warn("Empty message received")
		c.String(http.StatusBadRequest, "Message cannot be empty")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"message_length":  len(userMessage),
		"context":         context,
		"context_empty":   context == "",
		"context_bytes":   []byte(context),
	}).Info("Processing chat message")

	fmt.Printf("\n[CHAT DEBUG] ========================================\n")
	fmt.Printf("[CHAT DEBUG] Processing message with context: '%s'\n", context)
	fmt.Printf("[CHAT DEBUG] Context is empty: %v\n", context == "")
	fmt.Printf("[CHAT DEBUG] Context bytes: %v\n", []byte(context))
	fmt.Printf("[CHAT DEBUG] Message: '%s'\n", userMessage)
	fmt.Printf("[CHAT DEBUG] Is introduction context? %v\n", context == "introduction")
	fmt.Printf("[CHAT DEBUG] Is goal-definition context? %v\n", context == "goal-definition")
	fmt.Printf("[CHAT DEBUG] ========================================\n")

	// Handle context-specific processing
	conversation, convErr := h.designerService.GetConversation(conversationID)
	if convErr == nil && conversation != nil {
		handled, processErr := h.handleContextSpecificProcessing(c, conversation.AgencyID, userMessage, context, false)
		if processErr != nil {
			h.logger.WithError(processErr).Error("Context-specific processing failed")
			// Fall through to normal chat
		} else if handled {
			// Processing successful - return
			return
		}
		// Note: Context enrichment is no longer needed here as the AI prompt builder
		// (FormatAgencyContextBlock) already includes all agency data (goals, work items, etc.)
	} else {
		h.logger.WithError(convErr).Warn("Could not get conversation for context processing")
	}

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

	// Render only AI response (user message is already added by JavaScript)
	c.Header("Content-Type", "text/html")

	h.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
	}).Info("Rendering AI message only (user message added by JS)")

	// Render AI response only
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
	h.logger.Info("🔵 HANDLER CALLED: StartConversation")

	agencyID := c.Param("id")
	userMessage := c.PostForm("message")
	context := c.PostForm("context") // Get current section context

	if userMessage == "" {
		h.logger.Warn("Empty message received for new conversation")
		c.String(http.StatusBadRequest, "Message cannot be empty")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":      agencyID,
		"message_length": len(userMessage),
		"context":        context,
	}).Info("Starting new conversation")

	// Handle context-specific processing (introduction, goal-definition, etc.)
	h.logger.Info("🔵 CALLING: handleContextSpecificProcessing", "context", context)
	handled, err := h.handleContextSpecificProcessing(c, agencyID, userMessage, context, true)
	if err != nil {
		h.logger.WithError(err).Error("Context-specific processing failed")
		// Fall through to normal chat
	} else if handled {
		// Processing successful - return
		return
	}

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

	// Render only AI response (user message is already added by JavaScript)
	c.Header("Content-Type", "text/html")

	h.logger.WithFields(logrus.Fields{
		"conversation_id": conversation.ID,
	}).Info("Rendering AI message only for new conversation (user message added by JS)")

	// Render AI response only
	err = agency_designer.AIMessage(*response).Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render AI message")
		return
	}
	h.logger.Info("AI message rendered successfully")
}
