package handlers

import (
	"regexp"
	"strings"

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
	// The handler will check for ?stream=true query parameter
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

	// Set the user request in a dynamic request structure for goals chat processing
	// userMessage is used here in the struct (linter false positive for unused param)
	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		GoalKeys    []string `json:"goal_keys"`
	}{
		UserMessage: userMessage, // Used here - not unused!
		GoalKeys:    []string{},  // Empty means process all goals based on message context
	}

	// Store the request in the context so ProcessGoalsChatRequest can access it
	c.Set("dynamic_request", dynamicReq)

	// Check if streaming is requested
	streamMode := c.Query("stream") == "true"

	if streamMode {
		// Delegate to streaming version
		h.aiRefineHandler.ProcessGoalsChatRequestStreaming(c)
	} else {
		// Delegate to the ProcessGoalsChatRequest handler which wraps RefineGoals with chat formatting
		h.aiRefineHandler.ProcessGoalsChatRequest(c)
	}

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

	// Parse contexts from the userMessage if present
	contexts := parseContextsFromMessage(userMessage, h.logger)

	h.logger.WithFields(map[string]interface{}{
		"contexts_count": len(contexts),
		"contexts":       contexts,
	}).Info("🔍 Parsed contexts from message")

	// Check if we have a Deliverable Node context
	for _, ctx := range contexts {
		if ctxType, ok := ctx["type"].(string); ok {
			if ctxType == "Deliverable Node" {
				h.logger.WithFields(map[string]interface{}{
					"type":     ctx["type"],
					"code":     ctx["code"],
					"nodeName": ctx["nodeName"],
					"nodeType": ctx["nodeType"],
				}).Info("🎯 DELIVERABLE NODE CONTEXT DETECTED - This should route to deliverable handler!")
			}
		}
	}

	// Set the user request in a dynamic request structure for work items chat processing
	dynamicReq := struct {
		UserMessage  string                   `json:"user_message"`
		WorkItemKeys []string                 `json:"work_item_keys"`
		Contexts     []map[string]interface{} `json:"contexts,omitempty"`
	}{
		UserMessage:  userMessage,
		WorkItemKeys: []string{}, // Empty means process all work items based on message context
		Contexts:     contexts,   // Parsed contexts from message
	}

	// Store the request in the context so ProcessWorkItemsChatRequestStreaming can access it
	c.Set("dynamic_request", dynamicReq)

	// Check if streaming is requested
	streamMode := c.Query("stream") == "true"

	if streamMode {
		// Delegate to streaming version
		h.aiRefineHandler.ProcessWorkItemsChatRequestStreaming(c)
	} else {
		h.logger.Warn("Non-streaming work item processing not yet implemented")
		result := "non_streaming_not_implemented"
		return &result, nil
	}

	// If we got here without panic, consider it successful
	result := "success"
	return &result, nil
}

// performRolesProcessing delegates to the ai_refine handler for roles processing via chat
// Returns the response HTML or nil if processing failed
func (h *ChatHandler) performRolesProcessing(c *gin.Context, agencyID, userMessage string) (*string, error) {
	h.logger.Info("🔵 DELEGATING: Roles processing to ai_refine.Handler (chat mode)")

	conversationID := c.Param("conversationId")

	// Ensure conversation exists - start one if this is a new conversation
	if conversationID == "" {
		// Start conversation first for new conversations
		ctx := c.Request.Context()
		conversation, err := h.designerService.StartConversation(ctx, agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to start conversation for roles processing")
			return nil, err
		}
		conversationID = conversation.ID
		// Store conversation ID in params so it's available downstream
		c.Params = append(c.Params, gin.Param{Key: "conversationId", Value: conversationID})
	}

	// Store agencyID in params so ProcessRolesChatRequestStreaming can access it
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

	h.logger.Info("Conversation ready for roles processing",
		"agencyID", agencyID,
		"conversationID", conversationID)

	// Set the user request in a dynamic request structure for roles chat processing
	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		RoleKeys    []string `json:"role_keys"`
	}{
		UserMessage: userMessage,
		RoleKeys:    []string{}, // Empty means process all roles based on message context
	}

	// Store the request in the context so ProcessRolesChatRequestStreaming can access it
	c.Set("dynamic_request", dynamicReq)

	// Check if streaming is requested
	streamMode := c.Query("stream") == "true"

	if streamMode {
		// Delegate to streaming version
		h.aiRefineHandler.ProcessRolesChatRequestStreaming(c)
	} else {
		h.logger.Warn("Non-streaming role processing not yet implemented")
		result := "non_streaming_not_implemented"
		return &result, nil
	}

	// If we got here without panic, consider it successful
	result := "success"
	return &result, nil
}

// performWorkflowsProcessing delegates to the ai_refine handler for workflows processing via chat
// Returns the response HTML or nil if processing failed
func (h *ChatHandler) performWorkflowsProcessing(c *gin.Context, agencyID, userMessage string) (*string, error) {
	h.logger.Info("🔵 DELEGATING: Workflows processing to ai_refine.Handler (chat mode)")

	conversationID := c.Param("conversationId")

	// Ensure conversation exists - start one if this is a new conversation
	if conversationID == "" {
		// Start conversation first for new conversations
		ctx := c.Request.Context()
		conversation, err := h.designerService.StartConversation(ctx, agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to start conversation for workflows processing")
			return nil, err
		}
		conversationID = conversation.ID
		// Store conversation ID in params so it's available downstream
		c.Params = append(c.Params, gin.Param{Key: "conversationId", Value: conversationID})
	}

	// Store agencyID in params so ProcessWorkflowsChatRequestStreaming can access it
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

	h.logger.Info("Conversation ready for workflows processing",
		"agencyID", agencyID,
		"conversationID", conversationID)

	// Set the user request in a dynamic request structure for workflows chat processing
	dynamicReq := struct {
		UserMessage  string   `json:"user_message"`
		WorkflowKeys []string `json:"workflow_keys"`
	}{
		UserMessage:  userMessage,
		WorkflowKeys: []string{}, // Empty means process all workflows based on message context
	}

	// Store the request in the context so ProcessWorkflowsChatRequestStreaming can access it
	c.Set("dynamic_request", dynamicReq)

	// Check if streaming is requested
	streamMode := c.Query("stream") == "true"

	if streamMode {
		// Delegate to streaming version
		h.aiRefineHandler.ProcessWorkflowsChatRequestStreaming(c)
	} else {
		h.logger.Warn("Non-streaming workflow processing not yet implemented")
		result := "non_streaming_not_implemented"
		return &result, nil
	}

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
		h.logger.Info("User on work items/workflows section - performing direct work items/workflows processing")

		// Check which section we're actually on
		if context == "workflows" {
			// Perform workflows processing
			h.logger.Info("🔵 CALLING: performWorkflowsProcessing", "agencyID", agencyID)
			processed, err := h.performWorkflowsProcessing(c, agencyID, userMessage)
			if err != nil {
				h.logger.WithError(err).Error("Failed to perform workflows processing")
				return false, err
			}

			if processed != nil {
				return true, nil // Successfully handled
			}
		} else {
			// Perform work items processing
			h.logger.Info("🔵 CALLING: performWorkItemsProcessing", "agencyID", agencyID)
			processed, err := h.performWorkItemsProcessing(c, agencyID, userMessage)
			if err != nil {
				h.logger.WithError(err).Error("Failed to perform work items processing")
				return false, err
			}

			if processed != nil {
				return true, nil // Successfully handled
			}
		}
		return false, nil // Not handled, fall through to normal chat

	case "roles":
		h.logger.Info("User on roles section - performing direct roles processing")
		// Perform the roles processing directly (conversation handling is inside)
		h.logger.Info("🔵 CALLING: performRolesProcessing", "agencyID", agencyID)
		processed, err := h.performRolesProcessing(c, agencyID, userMessage)
		if err != nil {
			h.logger.WithError(err).Error("Failed to perform roles processing")
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

// parseContextsFromMessage extracts context objects from the formatted message text
// Looks for the **Context:** section and parses individual context entries
func parseContextsFromMessage(message string, logger interface{}) []map[string]interface{} {
	var contexts []map[string]interface{}

	// Look for the **Context:** section
	lines := strings.Split(message, "\n")
	inContextSection := false
	currentContext := make(map[string]interface{})
	currentContextType := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start of context section
		if strings.HasPrefix(trimmed, "**Context:**") {
			inContextSection = true
			continue
		}

		// If we're in the context section
		if inContextSection {
			// Check if it's a context entry line: "1. **Work Item** [WI-TAB]:" or "2. **Deliverable Node** [DELIV-...]:"
			re := regexp.MustCompile(`^(\d+)\.\s+\*\*(.+?)\*\*\s+\[(.+?)\]:`)
			matches := re.FindStringSubmatch(trimmed)

			if len(matches) == 4 {
				// Save previous context if exists
				if currentContextType != "" && len(currentContext) > 0 {
					contexts = append(contexts, currentContext)
				}

				// Start new context
				currentContextType = matches[2] // e.g., "Work Item" or "Deliverable Node"
				currentContext = make(map[string]interface{})
				currentContext["type"] = currentContextType
				currentContext["code"] = matches[3] // e.g., "WI-TAB" or "DELIV-node-..."
				continue
			}

			// Parse content lines within a context entry (they are indented)
			if currentContextType != "" && strings.HasPrefix(line, "   ") {
				contentLine := strings.TrimSpace(line)

				// Check for "Deliverable Enhancement: name (type)" pattern
				if strings.Contains(contentLine, "Deliverable Enhancement:") {
					parts := strings.SplitN(contentLine, "Deliverable Enhancement:", 2)
					if len(parts) == 2 {
						rest := strings.TrimSpace(parts[1])
						// Extract name and type from "name (type)" format
						if idx := strings.LastIndex(rest, "("); idx != -1 {
							nodeName := strings.TrimSpace(rest[:idx])
							nodeTypeWithParen := rest[idx:]
							nodeType := strings.Trim(nodeTypeWithParen, "()")

							currentContext["nodeName"] = nodeName
							currentContext["nodeType"] = nodeType
							currentContext["isEnhancementRequest"] = true
						}
					}
				} else {
					// Regular content line
					if content, ok := currentContext["content"].(string); ok {
						currentContext["content"] = content + "\n" + contentLine
					} else {
						currentContext["content"] = contentLine
					}
				}
			}

			// Check if we've reached the end of the context section (empty line or next section)
			if trimmed == "" && i > 0 {
				// Empty line might signal end of contexts
				if currentContextType != "" && len(currentContext) > 0 {
					contexts = append(contexts, currentContext)
					currentContextType = ""
					currentContext = make(map[string]interface{})
				}
			}
		}
	}

	// Save the last context if exists
	if currentContextType != "" && len(currentContext) > 0 {
		contexts = append(contexts, currentContext)
	}

	return contexts
}
