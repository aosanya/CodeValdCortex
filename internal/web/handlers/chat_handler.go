package handlers

import (
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/ai"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ChatHandler handles web chat interactions
type ChatHandler struct {
	designerService     *ai.AgencyDesignerService
	agencyService       agency.Service
	roleService         registry.RoleService
	introductionRefiner *ai.IntroductionRefiner
	logger              *logrus.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(
	designerService *ai.AgencyDesignerService,
	agencyService agency.Service,
	roleService registry.RoleService,
	introductionRefiner *ai.IntroductionRefiner,
	logger *logrus.Logger,
) *ChatHandler {
	return &ChatHandler{
		designerService:     designerService,
		agencyService:       agencyService,
		roleService:         roleService,
		introductionRefiner: introductionRefiner,
		logger:              logger,
	}
}

// SendMessage handles POST /api/v1/conversations/:conversationId/messages/web
// Returns HTML for HTMX to append to the chat
func (h *ChatHandler) SendMessage(c *gin.Context) {
	conversationID := c.Param("conversationId")
	userMessage := c.PostForm("message")
	activeTab := c.PostForm("activeTab") // Get active tab from form

	if userMessage == "" {
		h.logger.Warn("Empty message received")
		c.String(http.StatusBadRequest, "Message cannot be empty")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"message_length":  len(userMessage),
		"active_tab":      activeTab,
	}).Info("Processing chat message")

	// If user is on the introduction tab, always perform refinement
	if activeTab == "introduction" {
		h.logger.Info("User on introduction tab - performing direct refinement")

		// Get conversation to find agency ID
		conversation, err := h.designerService.GetConversation(conversationID)
		if err == nil && conversation != nil {
			// Perform the refinement directly
			refined, err := h.performIntroductionRefinement(c, conversation.AgencyID, userMessage)
			if err != nil {
				h.logger.WithError(err).Error("Failed to perform introduction refinement")
				// Fall back to normal chat
			} else if refined != nil {
				// Refinement successful - return the result
				return
			}
		}
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
	activeTab := c.PostForm("activeTab") // Get active tab from form

	if userMessage == "" {
		h.logger.Warn("Empty message received for new conversation")
		c.String(http.StatusBadRequest, "Message cannot be empty")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":      agencyID,
		"message_length": len(userMessage),
		"active_tab":     activeTab,
	}).Info("Starting new conversation")

	// If user is on the introduction tab, always perform refinement
	if activeTab == "introduction" {
		h.logger.Info("User on introduction tab - performing direct refinement in new conversation")

		// Start conversation first
		ctx := c.Request.Context()
		conversation, err := h.designerService.StartConversation(ctx, agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to start conversation for refinement")
			c.String(http.StatusInternalServerError, "Failed to start conversation")
			return
		}

		// Store conversation ID for the refinement method
		c.Params = append(c.Params, gin.Param{Key: "conversationId", Value: conversation.ID})

		// Perform the refinement directly
		refined, err := h.performIntroductionRefinement(c, agencyID, userMessage)
		if err != nil {
			h.logger.WithError(err).Error("Failed to perform introduction refinement")
			// Fall back to normal chat - remove the conversationId param
			c.Params = c.Params[:len(c.Params)-1]
		} else if refined != nil {
			// Refinement successful - return
			return
		}
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

// isIntroductionRefinementRequest detects if a chat message is requesting introduction refinement
// Looks for: action keywords + introduction context markers
// performIntroductionRefinement directly refines the introduction based on chat request
// Returns the response HTML or nil if refinement failed
func (h *ChatHandler) performIntroductionRefinement(c *gin.Context, agencyID string, userMessage string) (*string, error) {
	ctx := c.Request.Context()

	// Get agency
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		return nil, err
	}

	// Get current overview/introduction
	overview, err := h.agencyService.GetAgencyOverview(ctx, agencyID)
	if err != nil {
		// Create empty overview if not found
		overview = &agency.Overview{
			AgencyID:     agencyID,
			Introduction: "",
		}
	}

	// Get goals and work items for context
	goals, _ := h.agencyService.GetGoals(ctx, agencyID)
	workItems, _ := h.agencyService.GetWorkItems(ctx, agencyID)

	// Get roles and assignments for context
	roles, _ := h.roleService.ListTypes(ctx)
	assignments, _ := h.agencyService.GetAllRACIAssignments(ctx, agencyID)

	// Extract user request - keep the full message with context for AI
	// The AI needs to see what the user wants to remove
	userRequest := userMessage // Use full message, not just extracted request

	h.logger.WithFields(logrus.Fields{
		"agency_id":          agencyID,
		"current_intro_len":  len(overview.Introduction),
		"current_intro_text": overview.Introduction,
		"user_request_len":   len(userRequest),
		"user_request_text":  userRequest,
		"goals_count":        len(goals),
		"work_items_count":   len(workItems),
	}).Info("==== CHAT HANDLER - Preparing introduction refinement request ====")

	// Build refinement request
	refineReq := &ai.RefineIntroductionRequest{
		AgencyID:      agencyID,
		CurrentIntro:  overview.Introduction,
		Goals:         goals,
		WorkItems:     workItems,
		Roles:         roles,
		Assignments:   assignments,
		AgencyContext: ag,
		UserRequest:   userRequest, // Full message with context
	}

	// Call AI refiner
	refinedResult, err := h.introductionRefiner.RefineIntroduction(ctx, refineReq)
	if err != nil {
		return nil, err
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":       agencyID,
		"refined_len":     len(refinedResult.RefinedIntroduction),
		"was_changed":     refinedResult.WasChanged,
		"explanation_len": len(refinedResult.Explanation),
	}).Info("Introduction refinement completed")

	// Check if AI returned empty introduction - keep original if so
	if strings.TrimSpace(refinedResult.RefinedIntroduction) == "" {
		h.logger.Warn("AI returned empty introduction, keeping original")
		refinedResult.RefinedIntroduction = overview.Introduction
		refinedResult.Explanation = "AI returned empty introduction, keeping original."
		refinedResult.WasChanged = false
	}

	// Save the refined introduction
	if refinedResult.RefinedIntroduction != overview.Introduction {
		err = h.agencyService.UpdateAgencyOverview(ctx, agencyID, refinedResult.RefinedIntroduction)
		if err != nil {
			return nil, err
		}
	}

	// Add messages to conversation
	h.designerService.AddMessage(c.Param("conversationId"), "user", userMessage)
	chatMessage := "✨ **Introduction Refined & Saved**\n\n" + refinedResult.Explanation
	h.designerService.AddMessage(c.Param("conversationId"), "assistant", chatMessage)

	// Render response
	userMsg := ai.Message{Role: "user", Content: userMessage}
	aiMsg := ai.Message{Role: "assistant", Content: chatMessage}

	c.Header("Content-Type", "text/html")
	err = agency_designer.UserMessage(userMsg).Render(ctx, c.Writer)
	if err != nil {
		return nil, err
	}
	err = agency_designer.AIMessage(aiMsg).Render(ctx, c.Writer)
	if err != nil {
		return nil, err
	}

	result := "success"
	return &result, nil
}
