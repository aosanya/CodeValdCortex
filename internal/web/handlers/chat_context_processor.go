package handlers

import (
	"github.com/gin-gonic/gin"
)

// performIntroductionRefinement delegates to the ai_refine handler for introduction refinement
// Returns the response HTML or nil if refinement failed
func (h *ChatHandler) performIntroductionRefinement(c *gin.Context, agencyID, userMessage string) (*string, error) {
	h.logger.Info("🔵 DELEGATING: Introduction refinement to ai_refine.Handler")

	// Ensure conversation exists - start one if this is a new conversation
	conversationID := c.Param("conversationId")
	if conversationID == "" {
		// Start conversation first for new conversations
		ctx := c.Request.Context()
		conversation, err := h.designerService.StartConversation(ctx, agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to start conversation for refinement")
			return nil, err
		}
		conversationID = conversation.ID
		// Store conversation ID in params so it's available downstream
		c.Params = append(c.Params, gin.Param{Key: "conversationId", Value: conversationID})
	}

	h.logger.Info("Conversation ready for refinement",
		"agencyID", agencyID,
		"conversationID", conversationID)

	// Set the agencyID in the params so the ai_refine handler can access it
	// Need to add/update the :id param since the URL path is /conversations/:conversationId/messages/web
	paramFound := false
	for i, param := range c.Params {
		if param.Key == "id" {
			c.Params[i].Value = agencyID
			paramFound = true
			break
		}
	}
	if !paramFound {
		c.Params = append(c.Params, gin.Param{Key: "id", Value: agencyID})
	}

	// Set the user request in the form so the ai_refine handler can access it
	c.Request.PostForm.Set("user-request", userMessage)

	// Delegate to the ai_refine handler which has the full logic
	h.aiRefineHandler.RefineIntroduction(c)

	// If we got here without panic, consider it successful
	result := "success"
	return &result, nil
}

// performGoalsRefinement delegates to the ai_refine handler for goals processing via chat
// Returns the response HTML or nil if refinement failed
func (h *ChatHandler) performGoalsRefinement(c *gin.Context, agencyID, userMessage string) (*string, error) {
	h.logger.Info("🔵 DELEGATING: Goals processing to ai_refine.Handler (chat mode)")

	// Ensure conversation exists - start one if this is a new conversation
	conversationID := c.Param("conversationId")
	if conversationID == "" {
		// Start conversation first for new conversations
		ctx := c.Request.Context()
		conversation, err := h.designerService.StartConversation(ctx, agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to start conversation for goals processing")
			return nil, err
		}
		conversationID = conversation.ID
		// Store conversation ID in params so it's available downstream
		c.Params = append(c.Params, gin.Param{Key: "conversationId", Value: conversationID})
	}

	h.logger.Info("Conversation ready for goals processing",
		"agencyID", agencyID,
		"conversationID", conversationID)

	// Set the agencyID in the params
	paramFound := false
	for i, param := range c.Params {
		if param.Key == "id" {
			c.Params[i].Value = agencyID
			paramFound = true
			break
		}
	}
	if !paramFound {
		c.Params = append(c.Params, gin.Param{Key: "id", Value: agencyID})
	}

	// Set the user request in the form so the ai_refine handler can access it
	c.Request.PostForm.Set("user-request", userMessage)
	c.Request.PostForm.Set("message", userMessage)

	// Delegate to the ai_refine chat-friendly goal handler
	h.aiRefineHandler.ProcessGoalChatRequest(c)

	// If we got here without panic, consider it successful
	result := "success"
	return &result, nil
}

// performWorkItemsProcessing delegates to the ai_refine handler for work items processing via chat
// Returns the response HTML or nil if processing failed
func (h *ChatHandler) performWorkItemsProcessing(c *gin.Context, agencyID, userMessage string) (*string, error) {
	h.logger.Info("🔵 DELEGATING: Work items processing to ai_refine.Handler (chat mode)")

	conversationID := c.Param("conversationId")

	// Ensure conversation exists - start one if this is a new conversation
	if conversationID == "" {
		// Start conversation first for new conversations
		ctx := c.Request.Context()
		conversation, err := h.designerService.StartConversation(ctx, agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to start conversation for work items processing")
			return nil, err
		}
		conversationID = conversation.ID
		// Store conversation ID in params so it's available downstream
		c.Params = append(c.Params, gin.Param{Key: "conversationId", Value: conversationID})
	}

	// Store agencyID in params so RefineWorkItems can access it
	paramFound := false
	for i, param := range c.Params {
		if param.Key == "id" {
			c.Params[i].Value = agencyID
			paramFound = true
			break
		}
	}
	if !paramFound {
		c.Params = append(c.Params, gin.Param{Key: "id", Value: agencyID})
	}

	h.logger.Info("Conversation ready for work items processing",
		"agencyID", agencyID,
		"conversationID", conversationID)

	// Set the user request in the form so the ai_refine handler can access it
	c.Request.PostForm.Set("user-request", userMessage)
	c.Request.PostForm.Set("message", userMessage)

	// Delegate to the ai_refine work items handler
	h.aiRefineHandler.RefineWorkItems(c)

	// If we got here without panic, consider it successful
	result := "success"
	return &result, nil
}

// handleContextSpecificProcessing handles context-specific processing for both new and existing conversations
// Returns (handled bool, error) where handled=true means the request was fully processed
func (h *ChatHandler) handleContextSpecificProcessing(c *gin.Context, agencyID, userMessage, context string, isNewConversation bool) (bool, error) {
	h.logger.Info("🟢 FUNCTION ENTRY: handleContextSpecificProcessing",
		"context", context,
		"isNewConversation", isNewConversation,
		"agencyID", agencyID)

	switch context {
	case "introduction":
		h.logger.Info("User on introduction section - performing direct refinement")
		// Perform the refinement directly (conversation handling is inside)
		h.logger.Info("🔵 CALLING: performIntroductionRefinement", "agencyID", agencyID)
		refined, err := h.performIntroductionRefinement(c, agencyID, userMessage)
		if err != nil {
			h.logger.WithError(err).Error("Failed to perform introduction refinement")
			return false, err
		}

		if refined != nil {
			return true, nil // Successfully handled
		}
		return false, nil // Not handled, fall through to normal chat

	case "goal-definition":
		h.logger.Info("User on goal-definition section - performing direct goals processing")
		// Perform the goals processing directly (conversation handling is inside)
		h.logger.Info("🔵 CALLING: performGoalsRefinement", "agencyID", agencyID)
		refined, err := h.performGoalsRefinement(c, agencyID, userMessage)
		if err != nil {
			h.logger.WithError(err).Error("Failed to perform goals processing")
			return false, err
		}

		if refined != nil {
			return true, nil // Successfully handled
		}
		return false, nil // Not handled, fall through to normal chat

	case "work-items", "workflows":
		h.logger.Info("User on work items/workflows section - performing direct work items processing")
		// Perform the work items processing directly (conversation handling is inside)
		h.logger.Info("🔵 CALLING: performWorkItemsProcessing", "agencyID", agencyID)
		processed, err := h.performWorkItemsProcessing(c, agencyID, userMessage)
		if err != nil {
			h.logger.WithError(err).Error("Failed to perform work items processing")
			return false, err
		}

		if processed != nil {
			return true, nil // Successfully handled
		}
		return false, nil // Not handled, fall through to normal chat
	}

	// Context not recognized or not handled
	h.logger.Info("⚠️  Context not recognized or not handled - falling through to normal chat", "context", context)
	return false, nil
}
