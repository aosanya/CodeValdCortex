package ai_refine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
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
		UserMessage  string   `json:"user_message"`   // Natural language instruction
		WorkItemKeys []string `json:"work_item_keys"` // Optional: specific work items to operate on
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
		existingWorkItems = []*models.WorkItem{}
	}

	// Filter target work items if specific keys were provided
	var targetWorkItems []*models.WorkItem
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

	// Extract work item codes from context section of message (if present)
	// The chat form appends context like: "\n\n**Context:**\n\n1. **work-item** [WI-001]:\n   Description here"
	contextWorkItemCodes := extractWorkItemCodesFromContext(req.UserMessage)
	if len(contextWorkItemCodes) > 0 {
		h.logger.WithFields(logrus.Fields{
			"extracted_codes": contextWorkItemCodes,
		}).Info("Extracted work item codes from context")

		// Add these codes to the target work items if not already specified
		if len(req.WorkItemKeys) == 0 {
			req.WorkItemKeys = contextWorkItemCodes

			// Filter target work items by extracted codes
			workItemKeyMap := make(map[string]bool)
			for _, key := range contextWorkItemCodes {
				workItemKeyMap[key] = true
			}
			for _, workItem := range existingWorkItems {
				if workItemKeyMap[workItem.Code] {
					targetWorkItems = append(targetWorkItems, workItem)
				}
			}
		}
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":           agencyID,
		"user_message":        req.UserMessage,
		"target_work_items":   len(targetWorkItems),
		"existing_work_items": len(existingWorkItems),
		"context_codes":       len(contextWorkItemCodes),
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

	// Let AI decide what to do based on the message and context (creation vs refinement vs other operations)
	workItemData, err := h.processWorkItemRequest(c.Request.Context(), req.UserMessage, targetWorkItems, existingWorkItems, ag)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process work item request with AI")
		responseMessage := "❌ **Unable to Process Request**\n\n" +
			"I had trouble understanding your request. Please try rephrasing it with more details."

		if addErr := h.designerService.AddMessage(conversation.ID, "assistant", responseMessage); addErr != nil {
			h.logger.WithError(addErr).Error("Failed to add AI response to conversation")
		}

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, fmt.Sprintf(`
			<div class="ai-refine-response" 
				hx-trigger="load delay:100ms" 
				hx-get="/agencies/%s/chat-messages?agencyName=%s"
				hx-target="#chat-messages" 
				hx-swap="innerHTML">
			</div>
		`, agencyID, ag.Name))
		return
	}

	// Execute the operation determined by AI
	if workItemData.Operation == "update" {
		// UPDATE MODE: Refine existing work item
		h.logger.WithFields(logrus.Fields{
			"operation":  "update",
			"target_key": workItemData.TargetKey,
		}).Info("AI determined this is an update operation")

		// Update the work item
		updateReq := models.UpdateWorkItemRequest{
			Title:        workItemData.Title,
			Description:  workItemData.Description,
			Deliverables: workItemData.Deliverables,
			Tags:         workItemData.Tags,
		}

		err = h.agencyService.UpdateWorkItem(c.Request.Context(), agencyID, workItemData.TargetKey, updateReq)
		if err != nil {
			h.logger.WithError(err).Error("Failed to update work item")
			responseMessage := "❌ **Failed to Update Work Item**\n\n" +
				"I understood your request but couldn't save the changes. Please try again."

			if addErr := h.designerService.AddMessage(conversation.ID, "assistant", responseMessage); addErr != nil {
				h.logger.WithError(addErr).Error("Failed to add AI response to conversation")
			}

			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, fmt.Sprintf(`
				<div class="ai-refine-response" 
					hx-trigger="load delay:100ms" 
					hx-get="/agencies/%s/chat-messages?agencyName=%s"
					hx-target="#chat-messages" 
					hx-swap="innerHTML">
			</div>
			`, agencyID, ag.Name))
			return
		}

		// Get the updated work item
		updatedWorkItem, err := h.agencyService.GetWorkItem(c.Request.Context(), agencyID, workItemData.TargetKey)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to get updated work item for display")
		}

		// Prepare success response
		responseMessage := fmt.Sprintf("✅ **Work Item Updated Successfully!**\n\n"+
			"**%s** (%s)\n\n"+
			"%s\n\n"+
			"**Deliverables:**\n%s",
			workItemData.Title,
			workItemData.TargetKey,
			workItemData.Description,
			formatDeliverables(workItemData.Deliverables))

		if addErr := h.designerService.AddMessage(conversation.ID, "assistant", responseMessage); addErr != nil {
			h.logger.WithError(addErr).Error("Failed to add AI response to conversation")
		}

		h.logger.Info("Work item updated successfully via AI",
			"agencyID", agencyID,
			"workItemKey", updatedWorkItem.Key)

		// Return success with table refresh
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, fmt.Sprintf(`
			<div class="ai-refine-response" 
				hx-trigger="load delay:100ms" 
				hx-get="/agencies/%s/chat-messages?agencyName=%s"
				hx-target="#chat-messages" 
				hx-swap="innerHTML"
				hx-on::after-swap="
					const workItemsTable = document.getElementById('work-items-table-body');
					if (workItemsTable) {
						htmx.ajax('GET', '/api/v1/agencies/%s/work-items/html', {
							target: '#work-items-table-body',
							swap: 'innerHTML'
						});
					}
				">
			</div>
		`, agencyID, ag.Name, agencyID))
		return
	}

	// CREATE MODE: Create new work item
	h.logger.WithField("operation", "create").Info("AI determined this is a create operation")

	// Resolve dependencies: AI may return human-readable titles. We must convert them to work item codes.
	resolvedDeps := []string{}
	for _, dep := range workItemData.Dependencies {
		depTrim := strings.TrimSpace(dep)
		if depTrim == "" {
			continue
		}

		found := false
		for _, existing := range existingWorkItems {
			if strings.EqualFold(strings.TrimSpace(existing.Title), depTrim) || strings.EqualFold(existing.Code, depTrim) {
				resolvedDeps = append(resolvedDeps, existing.Code)
				found = true
				break
			}
		}

		if !found {
			// Create a placeholder work item for this dependency so the reference can be satisfied
			placeholderReq := models.CreateWorkItemRequest{
				Title:       depTrim,
				Description: "Placeholder work item created automatically as a dependency.",
			}
			ph, phErr := h.agencyService.CreateWorkItem(c.Request.Context(), agencyID, placeholderReq)
			if phErr != nil {
				h.logger.WithError(phErr).WithField("dependency", depTrim).Warn("Failed to create placeholder dependency; skipping dependency")
				continue
			}
			resolvedDeps = append(resolvedDeps, ph.Code)
			// Add to existingWorkItems so subsequent dependency resolution can find it
			existingWorkItems = append(existingWorkItems, ph)
		}
	}

	// Create the work item in the database
	createReq := models.CreateWorkItemRequest{
		Title:        workItemData.Title,
		Description:  workItemData.Description,
		Deliverables: workItemData.Deliverables,
		Dependencies: resolvedDeps,
		Tags:         workItemData.Tags,
	}

	newWorkItem, err := h.agencyService.CreateWorkItem(c.Request.Context(), agencyID, createReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create work item")
		responseMessage := "❌ **Failed to Create Work Item**\n\n" +
			"I understood your request but couldn't save it to the database. Please try again."

		if addErr := h.designerService.AddMessage(conversation.ID, "assistant", responseMessage); addErr != nil {
			h.logger.WithError(addErr).Error("Failed to add AI response to conversation")
		}

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, fmt.Sprintf(`
			<div class="ai-refine-response" 
				hx-trigger="load delay:100ms" 
				hx-get="/agencies/%s/chat-messages?agencyName=%s"
				hx-target="#chat-messages" 
				hx-swap="innerHTML">
			</div>
		`, agencyID, ag.Name))
		return
	}

	// Prepare success response message
	responseMessage := fmt.Sprintf("✅ **Work Item Created Successfully!**\n\n"+
		"**%s** (%s)\n\n"+
		"%s\n\n"+
		"**Deliverables:**\n%s\n\n"+
		"The work item has been added to your agency.",
		newWorkItem.Title,
		newWorkItem.Key,
		newWorkItem.Description,
		formatDeliverables(newWorkItem.Deliverables))

	// Add AI response to conversation
	if addErr := h.designerService.AddMessage(conversation.ID, "assistant", responseMessage); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add AI response to conversation")
	}

	h.logger.Info("Work item created successfully via AI",
		"agencyID", agencyID,
		"conversationID", conversation.ID,
		"workItemKey", newWorkItem.Key)

	// Return success - chat will be refreshed AND work items table will be reloaded
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
		<div class="ai-refine-response" 
			hx-trigger="load delay:100ms" 
			hx-get="/agencies/%s/chat-messages?agencyName=%s"
			hx-target="#chat-messages" 
			hx-swap="innerHTML"
			hx-on::after-swap="
				// Refresh work items table
				const workItemsTable = document.getElementById('work-items-table-body');
				if (workItemsTable) {
					htmx.ajax('GET', '/api/v1/agencies/%s/work-items/html', {
						target: '#work-items-table-body',
						swap: 'innerHTML'
					});
				}
			">
		</div>
	`, agencyID, ag.Name, agencyID))
}

// processWorkItemRequest uses AI to intelligently process work item requests
// The AI determines whether to create a new work item or refine an existing one based on context
func (h *Handler) processWorkItemRequest(ctx context.Context, userMessage string, targetWorkItems, existingWorkItems []*models.WorkItem, ag *models.Agency) (*workItemProcessResult, error) {
	// Get agency overview for context
	overview, err := h.agencyService.GetAgencyOverview(ctx, ag.ID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get agency overview, continuing without it")
		overview = &models.Overview{}
	}

	systemPrompt := `You are an AI assistant that helps manage work items in a project management system.

Based on the user's request and the provided context, determine the appropriate action and return work item details.

CONTEXT-AWARE BEHAVIOR:
- If work items are provided in the target context, the user wants to REFINE/UPDATE those items
- If no target work items are provided, the user wants to CREATE a new work item
- Use the user's message to understand what changes or improvements to make

Return a JSON object with these fields:
- operation: "create" or "update" (based on whether target work items exist)
- target_key: the work item key to update (only for update operations, empty for create)
- title: A clear, concise title (required)
- description: A detailed description (required)
- deliverables: A list of concrete deliverables (array of strings)
- dependencies: Dependencies as work item codes (array of strings) - only if explicitly mentioned
- tags: Relevant tags or categories (array of strings)

IMPORTANT RULES:
1. If target work items exist, operation MUST be "update" and target_key MUST be set
2. Preserve good existing content unless user explicitly asks to change it
3. For refinements: improve clarity, add details, fix issues based on user feedback
4. For new items: create comprehensive, well-structured work items
5. Do NOT add dependencies unless explicitly mentioned by the user

Return ONLY valid JSON. No markdown, no code blocks, no explanations.`

	introduction := "No introduction available"
	if overview.Introduction != "" {
		introduction = overview.Introduction
	}

	// Build context information
	targetContext := "None - this is a new work item request"
	if len(targetWorkItems) > 0 {
		targetContext = "Target Work Items to Refine:\n"
		for _, wi := range targetWorkItems {
			deliverables := "None"
			if len(wi.Deliverables) > 0 {
				deliverables = strings.Join(wi.Deliverables, "\n    - ")
			}
			targetContext += fmt.Sprintf(`
- Code: %s
  Title: %s
  Description: %s
  Deliverables:
    - %s
  Tags: %s
`, wi.Code, wi.Title, wi.Description, deliverables, strings.Join(wi.Tags, ", "))
		}
	}

	existingContext := "No existing work items"
	if len(existingWorkItems) > 0 {
		existingContext = fmt.Sprintf("Existing work items count: %d", len(existingWorkItems))
	}

	userPrompt := fmt.Sprintf(`Agency Context:
- Name: %s
- Introduction: %s
- %s

%s

User Request: %s

Determine the operation and return appropriate work item details as JSON.`,
		ag.Name, introduction, existingContext, targetContext, userMessage)

	// Call AI service
	response, err := h.callAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}

	// Parse JSON response
	cleanedContent := strings.TrimSpace(response)
	cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
	cleanedContent = strings.TrimPrefix(cleanedContent, "```")
	cleanedContent = strings.TrimSuffix(cleanedContent, "```")
	cleanedContent = strings.TrimSpace(cleanedContent)

	var result workItemProcessResult
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		h.logger.WithError(err).WithField("response", cleanedContent).Error("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// callAI is a helper method to make AI calls using the available LLM client
func (h *Handler) callAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Since we can't easily access the raw LLM client, use the designerService
	// but with a very strict JSON-only prompt to override its conversational nature
	conversation, err := h.designerService.StartConversation(ctx, "temp-json-parser")
	if err != nil {
		return "", fmt.Errorf("failed to create conversation: %w", err)
	}

	// Create an extremely strict prompt that forces JSON output
	strictPrompt := fmt.Sprintf(`YOU MUST RESPOND WITH ONLY VALID JSON. NO OTHER TEXT ALLOWED.

%s

%s

CRITICAL RULES:
1. Output ONLY the JSON object
2. NO markdown code blocks
3. NO explanations
4. NO conversational text
5. Just pure JSON

BEGIN JSON OUTPUT NOW:`, systemPrompt, userPrompt)

	response, err := h.designerService.SendMessage(ctx, conversation.ID, strictPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	return response.Content, nil
}

// workItemProcessResult represents the AI-processed work item operation
type workItemProcessResult struct {
	Operation    string   `json:"operation"`  // "create" or "update"
	TargetKey    string   `json:"target_key"` // Work item key to update (for update operations)
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Deliverables []string `json:"deliverables"`
	Dependencies []string `json:"dependencies"`
	Tags         []string `json:"tags"`
}

// formatDeliverables formats a list of deliverables for display
func formatDeliverables(deliverables []string) string {
	if len(deliverables) == 0 {
		return "• None specified"
	}
	var formatted []string
	for _, d := range deliverables {
		formatted = append(formatted, "• "+d)
	}
	return strings.Join(formatted, "\n")
}

// extractWorkItemCodesFromContext parses the context section of a message to extract work item codes
// Example context format:
// **Context:**
//  1. **work-item** [WI-001]:
//     Description here
func extractWorkItemCodesFromContext(message string) []string {
	var codes []string

	// Look for work item codes in square brackets [WI-XXX] or [MVP-XXX]
	// Common patterns: [WI-001], [MVP-001], etc.
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		// Look for patterns like "**work-item** [CODE]:" or just "[CODE]"
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			if start >= 0 && end > start {
				code := strings.TrimSpace(line[start+1 : end])
				// Validate it looks like a work item code (has a dash)
				if strings.Contains(code, "-") {
					codes = append(codes, code)
				}
			}
		}
	}

	return codes
}
