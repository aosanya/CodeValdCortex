package ai_refine

import (
	"context"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RolesResponse captures the JSON response from RefineRoles
type RolesResponse struct {
	Action         string `json:"action"`
	Explanation    string `json:"explanation"`
	NoActionNeeded bool   `json:"no_action_needed"`
}

// ProcessRolesChatRequestStreaming handles chat-based role interactions with streaming
// Uses real AI streaming exactly like goals and work items for consistency
func (h *Handler) ProcessRolesChatRequestStreaming(c *gin.Context) {
	h.logger.Info("🌊 HANDLER CALLED: ProcessRolesChatRequestStreaming")

	agencyID := c.Param("id")

	// Get user message from dynamic_request (set by chat_context_processor)
	dynamicReq, exists := c.Get("dynamic_request")
	if !exists {
		h.logger.Error("No dynamic_request found in context")
		c.SSEvent("error", `{"error": "Missing request data"}`)
		return
	}

	req, ok := dynamicReq.(struct {
		UserMessage string   `json:"user_message"`
		RoleKeys    []string `json:"role_keys"`
	})
	if !ok {
		h.logger.Error("Failed to cast dynamic_request to expected type")
		c.SSEvent("error", `{"error": "Invalid request format"}`)
		return
	}

	userMessage := req.UserMessage
	if userMessage == "" {
		h.logger.Error("No user message provided")
		c.SSEvent("error", `{"error": "No message provided"}`)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":    agencyID,
		"user_message": userMessage,
		"role_keys":    req.RoleKeys,
	}).Info("Processing streaming chat-based role request")

	// Fetch agency and specification
	ag, spec, err := h.fetchAgencyAndSpec(c, agencyID)
	if err != nil {
		c.SSEvent("error", `{"error": "Agency not found"}`)
		return
	}

	// Get or create conversation
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		h.logger.Warn("No conversation exists, creating new one", "agencyID", agencyID)
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation")
			c.SSEvent("error", `{"error": "Failed to initialize conversation"}`)
			return
		}
	}

	// Build AI context
	builderContextData, err := h.contextBuilder.BuildBuilderContext(
		c.Request.Context(),
		ag,
		spec.Introduction,
		userMessage,
	)
	if err != nil {
		c.SSEvent("error", `{"error": "Failed to build context"}`)
		return
	}

	// Build RefineRolesRequest
	existingRoles := make([]*models.Role, len(spec.Roles))
	for i := range spec.Roles {
		existingRoles[i] = &spec.Roles[i]
	}

	var targetRoles []*models.Role
	if len(req.RoleKeys) > 0 {
		roleKeyMap := make(map[string]bool)
		for _, key := range req.RoleKeys {
			roleKeyMap[key] = true
		}
		for _, role := range existingRoles {
			if roleKeyMap[role.Key] {
				targetRoles = append(targetRoles, role)
			}
		}
	}

	workItems := make([]*models.WorkItem, len(spec.WorkItems))
	for i := range spec.WorkItems {
		workItems[i] = &spec.WorkItems[i]
	}

	refineReq := &builder.RefineRolesRequest{
		AgencyID:      agencyID,
		UserMessage:   userMessage,
		TargetRoles:   targetRoles,
		ExistingRoles: existingRoles,
		WorkItems:     workItems,
		AgencyContext: ag,
	}

	// Setup SSE
	h.setupSSE(c)

	// Stream refinement (real AI streaming from backend, exactly like goals and work items)
	chunkCount := 0
	result, err := h.roleBuilder.RefineRolesStream(
		c.Request.Context(),
		refineReq,
		builderContextData,
		func(chunk string) error {
			chunkCount++
			c.SSEvent("chunk", chunk)
			c.Writer.Flush()
			return nil
		},
	)

	if err != nil {
		h.logger.WithError(err).Error("❌ Streaming role refinement failed")
		c.SSEvent("error", fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}

	h.logger.WithField("total_chunks", chunkCount).Info("✅ Streaming completed")

	// Save changes if any were made
	if !result.NoActionNeeded {
		h.logger.Info("Roles were modified, applying and saving changes...")
		if err := h.applyAndSaveRoles(c.Request.Context(), agencyID, result, existingRoles); err != nil {
			h.logger.WithError(err).Error("Failed to save roles")
			c.SSEvent("error", fmt.Sprintf(`{"error": "Failed to save roles: %s"}`, err.Error()))
			return
		}
	}

	// Format message for conversation history
	rolesResp := RolesResponse{
		Action:         result.Action,
		Explanation:    result.Explanation,
		NoActionNeeded: result.NoActionNeeded,
	}
	chatMessage := formatRolesChatMessage(rolesResp)

	// Add to conversation
	if err := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); err != nil {
		h.logger.WithError(err).Error("Failed to add message to conversation")
	}

	// Send completion
	completionData := map[string]interface{}{
		"was_changed":     !result.NoActionNeeded,
		"explanation":     result.Explanation,
		"message":         chatMessage,
		"conversation_id": conversation.ID,
	}

	c.SSEvent("complete", completionData)
	c.Writer.Flush()

	h.logger.Info("✅ Streaming roles chat completed")
}

// formatRolesChatMessage formats the roles AI response for chat display
func formatRolesChatMessage(resp RolesResponse) string {
	var message strings.Builder

	// Add emoji and title based on action
	if resp.NoActionNeeded {
		message.WriteString("✅ **Roles Review Complete**\n\n")
	} else {
		switch resp.Action {
		case "refine":
			message.WriteString("✨ **Roles Refined**\n\n")
		case "generate":
			message.WriteString("🎯 **Roles Generated**\n\n")
		case "consolidate":
			message.WriteString("📊 **Roles Consolidated**\n\n")
		case "enhance_all":
			message.WriteString("⚡ **Roles Enhanced**\n\n")
		case "remove":
			message.WriteString("🗑️ **Roles Removed**\n\n")
		case "under_construction":
			message.WriteString("🚧 **Feature Under Construction**\n\n")
		default:
			message.WriteString("✨ **Roles Updated**\n\n")
		}
	}

	// Add explanation
	if resp.Explanation != "" {
		message.WriteString(resp.Explanation)
	}

	return message.String()
}

// applyAndSaveRoles applies the AI recommendations and saves roles to the database
// Handles all actions: refine, generate, consolidate, remove, enhance_all
func (h *Handler) applyAndSaveRoles(ctx context.Context, agencyID string, result *builder.RefineRolesResponse, existingRoles []*models.Role) error {
	h.logger.WithFields(logrus.Fields{
		"action":          result.Action,
		"refined_count":   len(result.RefinedRoles),
		"generated_count": len(result.GeneratedRoles),
		"existing_roles":  len(existingRoles),
	}).Info("🔧 Applying role changes")

	var updatedRoles []models.Role
	rolesModified := false

	switch result.Action {
	case "refine", "enhance_all":
		// Create a map of refined roles by original key for quick lookup
		refinedMap := make(map[string]*builder.RefinedRoleResult)
		for i := range result.RefinedRoles {
			refinedMap[result.RefinedRoles[i].OriginalKey] = &result.RefinedRoles[i]
		}

		// Apply refinements to existing roles
		for _, role := range existingRoles {
			if refined, exists := refinedMap[role.Key]; exists && refined.WasChanged {
				h.logger.WithFields(logrus.Fields{
					"key":      role.Key,
					"old_name": role.Name,
					"new_name": refined.RefinedName,
				}).Info("✏️ Refining role")

				updatedRole := *role
				updatedRole.Name = refined.RefinedName
				updatedRole.Description = refined.RefinedDescription
				if refined.SuggestedAutonomyLevel != "" {
					updatedRole.AutonomyLevel = refined.SuggestedAutonomyLevel
				}
				if refined.SuggestedTokenBudget > 0 {
					updatedRole.TokenBudget = refined.SuggestedTokenBudget
				}
				if len(refined.SuggestedTags) > 0 {
					updatedRole.Tags = refined.SuggestedTags
				}
				updatedRoles = append(updatedRoles, updatedRole)
				rolesModified = true
			} else {
				// Keep role unchanged
				updatedRoles = append(updatedRoles, *role)
			}
		}

	case "generate":
		// Keep all existing roles
		for _, role := range existingRoles {
			updatedRoles = append(updatedRoles, *role)
		}

		// Add generated roles
		for _, gr := range result.GeneratedRoles {
			h.logger.WithFields(logrus.Fields{
				"name": gr.Name,
			}).Info("🆕 Adding generated role")

			newRole := models.Role{
				Code:          generateRoleCode(gr.Name),
				Name:          gr.Name,
				Description:   gr.Description,
				AutonomyLevel: gr.AutonomyLevel,
				TokenBudget:   gr.TokenBudget,
				Tags:          gr.Tags,
				IsActive:      true,
			}
			updatedRoles = append(updatedRoles, newRole)
			rolesModified = true
		}

	case "consolidate", "remove":
		if result.ConsolidatedData != nil {
			// Create a set of removed role keys for quick lookup
			removedKeys := make(map[string]bool)
			for _, removedKey := range result.ConsolidatedData.RemovedRoles {
				removedKeys[removedKey] = true
				h.logger.Info("🔍 DEBUG: Marking role for removal", "key", removedKey)
			}

			h.logger.WithFields(logrus.Fields{
				"total_existing_roles": len(existingRoles),
				"roles_to_remove":      len(removedKeys),
			}).Info("🔍 DEBUG: Processing role removal/consolidation")

			// Keep roles that are NOT in the removed list
			for _, role := range existingRoles {
				if !removedKeys[role.Key] {
					updatedRoles = append(updatedRoles, *role)
					h.logger.Info("🔍 DEBUG: Keeping role", "key", role.Key, "name", role.Name)
				} else {
					h.logger.Info("🗑️ Removing role", "key", role.Key, "name", role.Name)
					rolesModified = true
				}
			}

			// Add consolidated roles (these are new or updated roles)
			for _, cr := range result.ConsolidatedData.ConsolidatedRoles {
				h.logger.WithFields(logrus.Fields{
					"name": cr.Name,
				}).Info("🔄 Adding consolidated role")

				newRole := models.Role{
					Code:          generateRoleCode(cr.Name),
					Name:          cr.Name,
					Description:   cr.Description,
					AutonomyLevel: cr.AutonomyLevel,
					TokenBudget:   cr.TokenBudget,
					Tags:          cr.Tags,
					IsActive:      true,
				}
				updatedRoles = append(updatedRoles, newRole)
				rolesModified = true
			}

			h.logger.WithFields(logrus.Fields{
				"roles_modified":   rolesModified,
				"final_role_count": len(updatedRoles),
			}).Info("🔍 DEBUG: Consolidation/removal complete")
		}
	}

	// Save the updated roles list if modified
	if rolesModified {
		h.logger.WithFields(logrus.Fields{
			"previous_count": len(existingRoles),
			"updated_count":  len(updatedRoles),
		}).Info("💾 Saving updated roles to database")

		_, err := h.agencyService.UpdateSpecificationRoles(ctx, agencyID, updatedRoles, "ai-refine")
		if err != nil {
			h.logger.WithError(err).Error("❌ Failed to save roles to database")
			return fmt.Errorf("failed to save roles: %w", err)
		}
		h.logger.Info("✅ Successfully saved roles to database")
	} else {
		h.logger.Info("ℹ️ No role modifications needed")
	}

	return nil
}

// generateRoleCode creates a code from a role name
func generateRoleCode(name string) string {
	// Convert to uppercase and replace spaces with hyphens
	code := strings.ToUpper(name)
	code = strings.ReplaceAll(code, " ", "-")
	// Remove special characters (keep only alphanumeric and hyphens)
	var cleaned strings.Builder
	for _, r := range code {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}
	result := cleaned.String()
	// Limit to reasonable length
	if len(result) > 20 {
		result = result[:20]
	}
	return result
}
