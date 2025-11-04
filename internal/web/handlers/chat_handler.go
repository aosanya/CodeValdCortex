package handlers

import (
	"fmt"
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
	goalRefiner         *ai.GoalRefiner
	goalConsolidator    *ai.GoalConsolidator
	logger              *logrus.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(
	designerService *ai.AgencyDesignerService,
	agencyService agency.Service,
	roleService registry.RoleService,
	introductionRefiner *ai.IntroductionRefiner,
	goalRefiner *ai.GoalRefiner,
	goalConsolidator *ai.GoalConsolidator,
	logger *logrus.Logger,
) *ChatHandler {
	return &ChatHandler{
		designerService:     designerService,
		agencyService:       agencyService,
		roleService:         roleService,
		introductionRefiner: introductionRefiner,
		goalRefiner:         goalRefiner,
		goalConsolidator:    goalConsolidator,
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

	// Get the updated conversation to retrieve the user message that was added
	conversation, err = h.designerService.GetConversation(conversationID)
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
		"agency_id":        agencyID,
		"was_changed":      refinedResult.WasChanged,
		"explanation_len":  len(refinedResult.Explanation),
		"changed_sections": refinedResult.ChangedSections,
	}).Info("Introduction refinement completed")

	// Extract refined introduction from data
	var refinedIntro string
	if refinedResult.Data != nil && refinedResult.Data.Introduction != "" {
		refinedIntro = refinedResult.Data.Introduction
	} else {
		refinedIntro = overview.Introduction
	}

	// Check if AI returned empty introduction - keep original if so
	if strings.TrimSpace(refinedIntro) == "" {
		h.logger.Warn("AI returned empty introduction, keeping original")
		refinedIntro = overview.Introduction
		refinedResult.Explanation = "AI returned empty introduction, keeping original."
		refinedResult.WasChanged = false
	}

	// Save the refined introduction
	if refinedIntro != overview.Introduction {
		err = h.agencyService.UpdateAgencyOverview(ctx, agencyID, refinedIntro)
		if err != nil {
			return nil, err
		}
	}

	// Add messages to conversation
	h.designerService.AddMessage(c.Param("conversationId"), "user", userMessage)
	chatMessage := "✨ **Introduction Refined & Saved**\n\n" + refinedResult.Explanation
	h.designerService.AddMessage(c.Param("conversationId"), "assistant", chatMessage)

	// Trigger introduction reload on the frontend
	c.Header("HX-Trigger", "introductionUpdated")

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

// performGoalProcessing processes goal-related requests from chat
// Returns the response HTML or nil if processing failed
func (h *ChatHandler) performGoalProcessing(c *gin.Context, agencyID string, userMessage string, conversationID string) (*string, error) {
	ctx := c.Request.Context()

	// Get agency
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		return nil, err
	}

	// Get current overview/introduction for context
	overview, err := h.agencyService.GetAgencyOverview(ctx, agencyID)
	if err != nil {
		// Create empty overview if not found
		overview = &agency.Overview{
			AgencyID:     agencyID,
			Introduction: "",
		}
	}

	// Get existing goals
	existingGoals, err := h.agencyService.GetGoals(ctx, agencyID)
	if err != nil {
		h.logger.Error("Failed to get existing goals", "agencyID", agencyID, "error", err)
		existingGoals = []*agency.Goal{}
	}

	// Get units of work for context
	workItems, err := h.agencyService.GetWorkItems(ctx, agencyID)
	if err != nil {
		h.logger.Warn("Failed to get units of work", "agencyID", agencyID, "error", err)
		workItems = []*agency.WorkItem{}
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":           agencyID,
		"user_request":        userMessage,
		"existing_goals":      len(existingGoals),
		"introduction_length": len(overview.Introduction),
		"work_items_count":    len(workItems),
	}).Info("==== CHAT HANDLER - Processing goal request ====")

	// Build goal generation request with user's message as input
	genReq := &ai.GenerateGoalRequest{
		AgencyID:      agencyID,
		AgencyContext: ag,
		ExistingGoals: existingGoals,
		WorkItems:     workItems,
		UserInput:     userMessage, // Use user's chat message as input
	}

	// Call AI to generate goals
	result, err := h.goalRefiner.GenerateGoals(ctx, genReq)
	if err != nil {
		h.logger.Error("Failed to generate goals from AI", "agencyID", agencyID, "error", err)
		return nil, err
	}

	h.logger.Info("AI generated goals successfully",
		"agencyID", agencyID,
		"goalsCount", len(result.Goals),
		"explanation", result.Explanation)

	// Save each generated goal to database
	var createdGoals []*agency.Goal
	for i, goalData := range result.Goals {
		h.logger.Info("Saving generated goal to database",
			"agencyID", agencyID,
			"goalIndex", i+1,
			"goalCode", goalData.SuggestedCode,
			"descriptionLength", len(goalData.Description))

		goal, err := h.agencyService.CreateGoal(ctx, agencyID, goalData.SuggestedCode, goalData.Description)
		if err != nil {
			h.logger.Error("Failed to save generated goal",
				"agencyID", agencyID,
				"goalIndex", i+1,
				"goalCode", goalData.SuggestedCode,
				"error", err)
			// Continue with other goals even if one fails
			continue
		}

		h.logger.Info("Goal saved successfully",
			"agencyID", agencyID,
			"goalKey", goal.Key,
			"goalCode", goal.Code,
			"goalNumber", goal.Number)

		createdGoals = append(createdGoals, goal)
	}

	h.logger.Info("Completed creating goals",
		"agencyID", agencyID,
		"totalCreated", len(createdGoals),
		"requested", len(result.Goals))

	// Add messages to conversation
	h.designerService.AddMessage(conversationID, "user", userMessage)

	var chatMessage string
	if len(createdGoals) > 0 {
		chatMessage = fmt.Sprintf("✨ **Created %d Goal(s)**\n\n%s", len(createdGoals), result.Explanation)
	} else {
		chatMessage = fmt.Sprintf("✨ **Goal Analysis**\n\n%s", result.Explanation)
	}

	h.designerService.AddMessage(conversationID, "assistant", chatMessage)

	// Trigger goals reload on the frontend
	c.Header("HX-Trigger", "goalsUpdated")

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

	successResult := "success"
	return &successResult, nil
}

// handleContextSpecificProcessing handles context-specific processing for both new and existing conversations
// Returns (handled bool, error) where handled=true means the request was fully processed
func (h *ChatHandler) handleContextSpecificProcessing(c *gin.Context, agencyID, userMessage, context string, isNewConversation bool) (bool, error) {
	if context == "introduction" {
		h.logger.Info("User on introduction section - performing direct refinement")

		var conversationID string
		if isNewConversation {
			// Start conversation first for new conversations
			ctx := c.Request.Context()
			conversation, err := h.designerService.StartConversation(ctx, agencyID)
			if err != nil {
				h.logger.WithError(err).Error("Failed to start conversation for refinement")
				return false, err
			}
			conversationID = conversation.ID
			// Store conversation ID for the refinement method
			c.Params = append(c.Params, gin.Param{Key: "conversationId", Value: conversationID})
		}

		// Perform the refinement directly
		refined, err := h.performIntroductionRefinement(c, agencyID, userMessage)
		if err != nil {
			h.logger.WithError(err).Error("Failed to perform introduction refinement")
			return false, err
		}

		if refined != nil {
			return true, nil // Successfully handled
		}
		return false, nil // Not handled, fall through to normal chat

	} else if context == "goal-definition" {
		h.logger.Info("User on goal-definition section - processing goal request")

		var conversationID string
		if isNewConversation {
			// Start conversation first for new conversations
			ctx := c.Request.Context()
			conversation, err := h.designerService.StartConversation(ctx, agencyID)
			if err != nil {
				h.logger.WithError(err).Error("Failed to start conversation for goal processing")
				return false, err
			}
			conversationID = conversation.ID
		} else {
			// Get existing conversation ID
			conversationID = c.Param("conversationId")
		}

		// Perform the goal processing directly
		processed, err := h.performGoalProcessing(c, agencyID, userMessage, conversationID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to perform goal processing")
			return false, err
		}

		if processed != nil {
			return true, nil // Successfully handled
		}
		return false, nil // Not handled, fall through to normal chat
	}

	// Context not recognized or not handled
	return false, nil
}
