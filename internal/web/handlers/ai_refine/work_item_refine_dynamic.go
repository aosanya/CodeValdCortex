package ai_refine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

	// Use AI to parse the user request and create a work item
	workItemData, err := h.parseWorkItemRequest(c.Request.Context(), req.UserMessage, ag)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse work item request with AI")
		responseMessage := "❌ **Unable to Process Request**\n\n" +
			"I had trouble understanding your work item request. Please try rephrasing it with more details about:\n" +
			"• What the work item is about\n" +
			"• What needs to be delivered\n" +
			"• Any dependencies or requirements"

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
			placeholderReq := agency.CreateWorkItemRequest{
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
	createReq := agency.CreateWorkItemRequest{
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

// parseWorkItemRequest uses AI to parse a natural language work item request
func (h *Handler) parseWorkItemRequest(ctx context.Context, userMessage string, ag *agency.Agency) (*workItemParseResult, error) {
	// Get agency overview for context
	overview, err := h.agencyService.GetAgencyOverview(ctx, ag.ID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get agency overview, continuing without it")
		overview = &agency.Overview{}
	}

	systemPrompt := `You are an AI assistant that helps parse natural language requests for creating work items in a project management system.

Given a user's request, extract the following information:
- title: A clear, concise title for the work item (required)
- description: A detailed description of what needs to be done (required)
- deliverables: A list of concrete deliverables or outputs (array of strings)
- dependencies: ONLY include dependencies if the user explicitly mentions them or if they reference existing work items. Otherwise, leave this as an empty array. (array of strings)
- tags: Relevant tags or categories (array of strings)

IMPORTANT: For dependencies, only include them if the user's request explicitly mentions prerequisite work or references existing work items. Most new work items should have an empty dependencies array.

Return ONLY a JSON object with these fields. Do not include any markdown formatting or code blocks.

Example input: "Add a work item for: User add a new issue"
Example output:
{
  "title": "User Add New Issue",
  "description": "Implement functionality for users to add new issues through the work items interface",
  "deliverables": ["UI form for adding issues", "Backend API endpoint", "Database schema update", "Input validation"],
  "dependencies": [],
  "tags": ["feature", "user-interface", "work-items"]
}`

	introduction := "No introduction available"
	if overview.Introduction != "" {
		introduction = overview.Introduction
	}

	userPrompt := fmt.Sprintf(`Agency Context:
- Name: %s
- Introduction: %s

User Request: %s

Parse this request and return a JSON object with the work item details.`, ag.Name, introduction, userMessage)

	// Call AI service via the introductionRefiner's LLM client (all builders share the same client)
	response, err := h.callAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}

	// Parse JSON response
	cleanedContent := strings.TrimSpace(response)
	// Remove markdown code blocks if present
	cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
	cleanedContent = strings.TrimPrefix(cleanedContent, "```")
	cleanedContent = strings.TrimSuffix(cleanedContent, "```")
	cleanedContent = strings.TrimSpace(cleanedContent)

	var result workItemParseResult
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

// workItemParseResult represents the AI-parsed work item data
type workItemParseResult struct {
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
