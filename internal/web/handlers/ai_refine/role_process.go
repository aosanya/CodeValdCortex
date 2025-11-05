package ai_refine

import (
	"fmt"
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/aosanya/CodeValdCortex/internal/registry"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProcessAIRoleRequest handles POST /api/v1/agencies/:id/roles/ai-process
// Processes multiple AI operations on roles (create, enhance, consolidate)
func (h *Handler) ProcessAIRoleRequest(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse request body
	var req struct {
		Operations []string `json:"operations" binding:"required"`
		RoleKeys   []string `json:"role_keys"` // Optional: specific roles to enhance/consolidate
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse AI process request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":  agencyID,
		"operations": req.Operations,
		"role_keys":  req.RoleKeys,
	}).Info("Processing AI role operations")

	// Validate agency exists and get context
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		h.logger.Error("Agency not found", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	// Get work items for context
	workItems, err := h.agencyService.GetWorkItems(ctx, agencyID)
	if err != nil {
		h.logger.Warn("Failed to get work items", "agencyID", agencyID, "error", err)
		workItems = []*agency.WorkItem{}
	}

	// TODO: Get existing roles from role service
	// For now, we'll process based on work items

	results := make(map[string]interface{})

	// Process each operation
	for _, operation := range req.Operations {
		h.logger.Info("Processing operation", "operation", operation, "agencyID", agencyID)

		switch operation {
		case "create":
			h.processCreateRolesOperation(c, agencyID, ag, workItems, results)
		case "enhance":
			results["enhance_status"] = "Enhancement operation not yet implemented"
			h.logger.Info("Role enhancement requested but not yet implemented", "agencyID", agencyID)
		case "consolidate":
			results["consolidate_status"] = "Consolidation operation not yet implemented"
			h.logger.Info("Role consolidation requested but not yet implemented", "agencyID", agencyID)
		}
	}

	// Add AI explanation to chat conversation if there's an explanation
	explanation, hasExplanation := results["ai_explanation"].(string)
	if hasExplanation && explanation != "" {
		h.addRoleExplanationToChat(c, agencyID, explanation)
	}

	// Build response
	response := gin.H{
		"success": true,
		"results": results,
	}

	h.logger.Info("AI role operations completed", "agencyID", agencyID)

	c.JSON(http.StatusOK, response)
}

func (h *Handler) processCreateRolesOperation(
	c *gin.Context,
	agencyID string,
	ag *agency.Agency,
	workItems []*agency.WorkItem,
	results map[string]interface{},
) {
	// Check if role creator is available
	if h.roleBuilder == nil {
		h.logger.Warn("Role creator not available", "agencyID", agencyID)
		results["create_error"] = "AI role generation service is not configured"
		return
	}

	// Check if we have work items
	if len(workItems) == 0 {
		h.logger.Warn("No work items found for role generation", "agencyID", agencyID)
		results["create_error"] = "No work items found. Please create work items first."
		return
	}

	h.logger.Info("Generating roles from work items",
		"agencyID", agencyID,
		"workItemsCount", len(workItems))

	// TODO: Get existing roles from role service to avoid duplicates
	// For now, we'll pass an empty list

	ctx := c.Request.Context()

	// Build AI context for role generation
	builderContext, err := h.contextBuilder.BuildBuilderContext(ctx, ag, "", "")
	if err != nil {
		h.logger.Error("Failed to build context for role generation", "agencyID", agencyID, "error", err)
		results["create_error"] = fmt.Sprintf("Failed to build context: %v", err)
		return
	}

	// Call AI role builder
	generateReq := &builder.GenerateRolesRequest{
		AgencyID: agencyID,
	}

	response, err := h.roleBuilder.GenerateRoles(ctx, generateReq, builderContext)
	if err != nil {
		h.logger.Error("Failed to generate roles", "agencyID", agencyID, "error", err)
		results["create_error"] = fmt.Sprintf("Failed to generate roles: %v", err)
		return
	}

	h.logger.Info("AI generated roles",
		"agencyID", agencyID,
		"rolesCount", len(response.Roles))

	// Save generated roles to registry
	createdCount := 0
	var createdRoles []string
	for _, genRole := range response.Roles {
		// Convert AI generated role to registry.Role
		role := &registry.Role{
			ID:                   genRole.Name, // Use name as ID since SuggestedCode not in builder type yet
			Name:                 genRole.Name,
			Description:          genRole.Description,
			Tags:                 genRole.Tags,
			Version:              "1.0",
			AutonomyLevel:        genRole.AutonomyLevel,
			RequiredSkills:       genRole.RequiredSkills,
			RequiredCapabilities: genRole.Capabilities,
			TokenBudget:          genRole.TokenBudget,
			Icon:                 "🤖",       // Default icon until builder type includes it
			Color:                "#3498db", // Default color until builder type includes it
			IsSystemType:         false,
			IsEnabled:            true,
		}

		// Register the role
		err = h.roleService.RegisterType(ctx, role)
		if err != nil {
			h.logger.Error("Failed to register role",
				"agencyID", agencyID,
				"roleName", role.Name,
				"error", err)
			continue
		}

		createdCount++
		createdRoles = append(createdRoles, role.Name)
		h.logger.Info("Successfully created role",
			"agencyID", agencyID,
			"roleName", role.Name,
			"roleID", role.ID)
	}

	// Build results
	results["create_status"] = fmt.Sprintf("Generated %d roles successfully", createdCount)
	results["created_roles"] = createdRoles
	results["ai_explanation"] = response.Explanation

	h.logger.Info("Role generation completed",
		"agencyID", agencyID,
		"generatedCount", len(response.Roles),
		"createdCount", createdCount)
}

func (h *Handler) addRoleExplanationToChat(c *gin.Context, agencyID string, explanation string) {
	h.logger.Info("Attempting to add role AI explanation to chat",
		"agencyID", agencyID,
		"explanationLength", len(explanation))

	// Get or create conversation
	conversation, err := h.designerService.GetConversationByAgencyID(agencyID)
	if err != nil {
		h.logger.Warn("No conversation exists, creating new one",
			"agencyID", agencyID,
			"error", err)
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create conversation for AI role generation message")
			return
		}
	}

	if conversation == nil {
		h.logger.Error("Conversation is nil after creation attempt", "agencyID", agencyID)
		return
	}

	// Build chat message
	chatMessage := fmt.Sprintf("✨ **Role Generation**\n\n%s", explanation)

	h.logger.Info("Adding role message to chat",
		"agencyID", agencyID,
		"conversationID", conversation.ID,
		"messageLength", len(chatMessage))

	if addErr := h.designerService.AddMessage(conversation.ID, "assistant", chatMessage); addErr != nil {
		h.logger.WithError(addErr).Error("Failed to add role generation explanation to chat")
	} else {
		h.logger.Info("Successfully added role AI explanation to chat",
			"agencyID", agencyID,
			"conversationID", conversation.ID)
	}
}
