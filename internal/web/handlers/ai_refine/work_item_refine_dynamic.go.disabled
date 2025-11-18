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

	// Process the changes from AI response
	var responseMessages []string

	// Handle introduction changes
	if workItemData.Changes.Introduction != nil {
		intro := workItemData.Changes.Introduction
		h.logger.WithFields(logrus.Fields{
			"key":         intro.Key,
			"explanation": intro.Explanation,
		}).Info("Processing introduction change")
		responseMessages = append(responseMessages, fmt.Sprintf("📝 **Introduction**: %s", intro.Explanation))
	}

	// Handle goal changes
	for _, goal := range workItemData.Changes.Goals {
		h.logger.WithFields(logrus.Fields{
			"key":         goal.Key,
			"code":        goal.Code,
			"explanation": goal.Explanation,
		}).Info("Processing goal change")
		responseMessages = append(responseMessages, fmt.Sprintf("🎯 **Goal %s**: %s", goal.Code, goal.Explanation))
	}

	// Handle work item changes
	for _, wi := range workItemData.Changes.WorkItems {
		h.logger.WithFields(logrus.Fields{
			"key":         wi.Key,
			"code":        wi.Code,
			"explanation": wi.Explanation,
		}).Info("Processing work item change")

		// Determine if this is an update or create
		var workItem *models.WorkItem
		var err error

		// Try to find the work item - prioritize Code lookup since AI often provides codes
		if wi.Code != "" {
			// Try by code first (e.g., "WI-001")
			workItem, err = h.agencyService.GetWorkItemByCode(c.Request.Context(), agencyID, wi.Code)
		} else if wi.Key != "" {
			// Fall back to database key lookup
			workItem, err = h.agencyService.GetWorkItem(c.Request.Context(), agencyID, wi.Key)
		}

		if err != nil || workItem == nil {
			// Work item not found - this is a create operation
			h.logger.WithFields(logrus.Fields{
				"key":  wi.Key,
				"code": wi.Code,
			}).Info("Work item not found - will create new")

			// Extract content fields for creation
			if wi.Content != nil {
				// Build create request from content
				createReq := models.CreateWorkItemRequest{
					Title:        "New Work Item", // Default title
					Description:  "",
					Deliverables: []string{},
					Tags:         []string{},
				}

				// Extract title
				if title, ok := wi.Content["title"].(string); ok && title != "" {
					createReq.Title = title
				}

				// Extract description
				if description, ok := wi.Content["description"].(string); ok && description != "" {
					createReq.Description = description
				}

				// Extract deliverables
				if deliverablesRaw, ok := wi.Content["deliverables"]; ok {
					if deliverablesList, ok := deliverablesRaw.([]interface{}); ok {
						deliverables := make([]string, 0, len(deliverablesList))
						for _, d := range deliverablesList {
							if str, ok := d.(string); ok {
								deliverables = append(deliverables, str)
							}
						}
						createReq.Deliverables = deliverables
					}
				}

				// Extract tags
				if tagsRaw, ok := wi.Content["tags"]; ok {
					if tagsList, ok := tagsRaw.([]interface{}); ok {
						tags := make([]string, 0, len(tagsList))
						for _, t := range tagsList {
							if str, ok := t.(string); ok {
								tags = append(tags, str)
							}
						}
						createReq.Tags = tags
					}
				}

				// Create the work item
				createdWorkItem, createErr := h.agencyService.CreateWorkItem(c.Request.Context(), agencyID, createReq)
				if createErr != nil {
					h.logger.WithError(createErr).Error("Failed to create work item")
					responseMessages = append(responseMessages, fmt.Sprintf("❌ **Work Item %s**: Failed to create", wi.Code))
				} else {
					h.logger.WithFields(logrus.Fields{
						"key":   createdWorkItem.Key,
						"code":  createdWorkItem.Code,
						"title": createdWorkItem.Title,
					}).Info("Successfully created work item")
					responseMessages = append(responseMessages, fmt.Sprintf("✨ **New Work Item %s**: %s", createdWorkItem.Code, wi.Explanation))
				}
			} else {
				h.logger.Warn("No content provided for work item creation", "code", wi.Code)
				responseMessages = append(responseMessages, fmt.Sprintf("⚠️ **Work Item %s**: No content provided for creation", wi.Code))
			}
		} else {
			// Work item found - this is an update operation
			h.logger.WithFields(logrus.Fields{
				"key":  workItem.Key,
				"code": workItem.Code,
			}).Info("Work item found - will update")

			// Apply content changes from AI if provided
			if wi.Content != nil {
				// Build update request with fields from content
				updateReq := models.UpdateWorkItemRequest{
					Title:        workItem.Title, // Start with current values
					Description:  workItem.Description,
					Deliverables: workItem.Deliverables,
					Tags:         workItem.Tags,
				}

				updated := false

				// Update title if provided
				if title, ok := wi.Content["title"].(string); ok && title != "" {
					updateReq.Title = title
					updated = true
				}

				// Update description if provided
				if description, ok := wi.Content["description"].(string); ok && description != "" {
					updateReq.Description = description
					updated = true
				}

				// Update deliverables if provided
				if deliverablesRaw, ok := wi.Content["deliverables"]; ok {
					if deliverablesList, ok := deliverablesRaw.([]interface{}); ok {
						deliverables := make([]string, 0, len(deliverablesList))
						for _, d := range deliverablesList {
							if str, ok := d.(string); ok {
								deliverables = append(deliverables, str)
							}
						}
						updateReq.Deliverables = deliverables
						updated = true
					}
				}

				// Update tags if provided
				if tagsRaw, ok := wi.Content["tags"]; ok {
					if tagsList, ok := tagsRaw.([]interface{}); ok {
						tags := make([]string, 0, len(tagsList))
						for _, t := range tagsList {
							if str, ok := t.(string); ok {
								tags = append(tags, str)
							}
						}
						updateReq.Tags = tags
						updated = true
					}
				}

				// Save the updated work item
				if updated {
					if updateErr := h.agencyService.UpdateWorkItem(c.Request.Context(), agencyID, workItem.Key, updateReq); updateErr != nil {
						h.logger.WithError(updateErr).Error("Failed to update work item")
						responseMessages = append(responseMessages, fmt.Sprintf("❌ **Work Item %s**: Failed to save updates", workItem.Code))
					} else {
						h.logger.WithFields(logrus.Fields{
							"key":  workItem.Key,
							"code": workItem.Code,
						}).Info("Successfully updated work item")
						responseMessages = append(responseMessages, fmt.Sprintf("✅ **Work Item %s**: %s", workItem.Code, wi.Explanation))
					}
				} else {
					responseMessages = append(responseMessages, fmt.Sprintf("ℹ️ **Work Item %s**: No content changes provided", workItem.Code))
				}
			} else {
				responseMessages = append(responseMessages, fmt.Sprintf("ℹ️ **Work Item %s**: %s (no content updates)", workItem.Code, wi.Explanation))
			}
		}
	}

	// Handle role changes
	for _, role := range workItemData.Changes.Roles {
		h.logger.WithFields(logrus.Fields{
			"key":         role.Key,
			"code":        role.Code,
			"explanation": role.Explanation,
		}).Info("Processing role change")
		responseMessages = append(responseMessages, fmt.Sprintf("👤 **Role %s**: %s", role.Code, role.Explanation))
	}

	// Handle RACI changes
	for _, raci := range workItemData.Changes.RACI {
		h.logger.WithFields(logrus.Fields{
			"key":         raci.Key,
			"code":        raci.Code,
			"explanation": raci.Explanation,
		}).Info("Processing RACI change")
		responseMessages = append(responseMessages, fmt.Sprintf("📋 **RACI %s**: %s", raci.Code, raci.Explanation))
	}

	// Handle workflow changes
	for _, workflow := range workItemData.Changes.Workflows {
		h.logger.WithFields(logrus.Fields{
			"key":         workflow.Key,
			"code":        workflow.Code,
			"explanation": workflow.Explanation,
		}).Info("Processing workflow change")
		responseMessages = append(responseMessages, fmt.Sprintf("🔄 **Workflow %s**: %s", workflow.Code, workflow.Explanation))
	}

	// Handle deletions
	deletedCount := 0
	deletedCodes := []string{}

	// Delete work items
	for _, codeOrKey := range workItemData.Deletions.WorkItems {
		h.logger.WithField("code_or_key", codeOrKey).Info("Processing work item deletion")

		// Try to find by code first, then by key
		var workItem *models.WorkItem
		var err error

		workItem, err = h.agencyService.GetWorkItemByCode(c.Request.Context(), agencyID, codeOrKey)
		if err != nil || workItem == nil {
			// Try by key
			workItem, err = h.agencyService.GetWorkItem(c.Request.Context(), agencyID, codeOrKey)
		}

		if err != nil || workItem == nil {
			h.logger.WithField("code_or_key", codeOrKey).Warn("Work item not found for deletion")
			responseMessages = append(responseMessages, fmt.Sprintf("⚠️ **Work Item %s**: Not found", codeOrKey))
			continue
		}

		// Delete the work item
		if delErr := h.agencyService.DeleteWorkItem(c.Request.Context(), agencyID, workItem.Key); delErr != nil {
			h.logger.WithError(delErr).Error("Failed to delete work item", "key", workItem.Key)
			responseMessages = append(responseMessages, fmt.Sprintf("❌ **Work Item %s**: Failed to delete", workItem.Code))
		} else {
			h.logger.WithFields(logrus.Fields{
				"key":  workItem.Key,
				"code": workItem.Code,
			}).Info("Successfully deleted work item")
			deletedCount++
			deletedCodes = append(deletedCodes, workItem.Code)
			responseMessages = append(responseMessages, fmt.Sprintf("🗑️ **Work Item %s**: Deleted", workItem.Code))
		}
	}

	// Prepare combined response message
	responseMessage := "✅ **Changes Processed Successfully!**\n\n"
	if len(responseMessages) > 0 {
		responseMessage += strings.Join(responseMessages, "\n\n")
	} else {
		responseMessage = "ℹ️ **No Changes Detected**\n\nI analyzed your request but didn't identify specific changes to make. Could you provide more details?"
	}

	// Add AI response to conversation
	if addErr := h.designerService.AddMessage(conversation.ID, "assistant", responseMessage); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add AI response to conversation")
	}

	h.logger.Info("AI changes processed successfully",
		"agencyID", agencyID,
		"conversationID", conversation.ID,
		"changesCount", len(responseMessages))

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

	systemPrompt := `You are an AI assistant that helps manage agency components (introduction, goals, work items, roles, RACI, workflows).

Based on the user's request and context, determine what to change and return a structured response.

Return a JSON object with these fields:
{
  "operation": "update" | "create" | "delete" | "mixed",
  "changes": {
    "introduction": {
      "key": "existing-key-for-update",  // Empty for creates
      "code": "INTRO",
      "explanation": "Updated the introduction to...",
      "content": {
        "introduction": "The actual refined introduction text here..."
      }
    },
    "goals": [
      {
        "key": "existing-key-for-update",  // Empty for creates
        "code": "G-001",
        "explanation": "Refined goal description to clarify...",
        "content": {
          "description": "The refined goal description text..."
        }
      }
    ],
    "work_items": [
      {
        "key": "existing-key-for-update",  // Empty for creates
        "code": "WI-001",
        "explanation": "Updated work item title and description to...",
        "content": {
          "title": "The new title",
          "description": "The refined description",
          "deliverables": ["Deliverable 1", "Deliverable 2"],
          "tags": ["tag1", "tag2"]
        }
      }
    ],
    "roles": [
      {
        "key": "existing-key-for-update",
        "code": "ROLE-001",
        "explanation": "Modified role to include...",
        "content": {
          "name": "Role name",
          "description": "Role description",
          "capabilities": ["capability1", "capability2"]
        }
      }
    ],
    "raci": [
      {
        "key": "existing-key-for-update",
        "code": "RACI-001",
        "explanation": "Updated RACI assignment...",
        "content": {
          "role_key": "role123",
          "work_item_key": "wi456",
          "responsibility": "R"
        }
      }
    ],
    "workflows": [
      {
        "key": "existing-key-for-update",
        "code": "WF-001",
        "explanation": "Adjusted workflow to...",
        "content": {
          "name": "Workflow name",
          "description": "Workflow description"
        }
      }
    ]
  },
  "deletions": {
    "goals": ["G-001", "G-002"],           // Codes or keys of goals to delete
    "work_items": ["WI-001", "WI-002"],   // Codes or keys of work items to delete
    "roles": ["ROLE-001"],                 // Codes or keys of roles to delete
    "raci": ["raci_key_123"],              // Keys of RACI assignments to delete
    "workflows": ["WF-001"]                // Codes or keys of workflows to delete
  }
}

IMPORTANT RULES:
1. For UPDATE operations: Include the "key" field OR the "code" field in "changes" (we'll look up by code if key is not provided)
2. For CREATE operations: Leave "key" empty in "changes", but provide "code" for a specific code
3. For DELETE operations: Add codes/keys to the "deletions" object arrays (NOT in "changes")
4. For "remove all X" requests: Set operation to "delete" and list all items in deletions.X
5. Only include changed components in the "changes" object
6. The "explanation" field should briefly describe what changed
7. The "content" field MUST contain the actual updated/new field values
8. For work items, only include fields that are being changed in "content"
9. Keep existing values for fields not mentioned in "content"
10. Use operation "mixed" when both updating and deleting items

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
- Key: %s
  Code: %s
  Title: %s
  Description: %s
  Deliverables:
    - %s
  Tags: %s
`, wi.Key, wi.Code, wi.Title, wi.Description, deliverables, strings.Join(wi.Tags, ", "))
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
type workItemProcessResult = AIProcessResult

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
