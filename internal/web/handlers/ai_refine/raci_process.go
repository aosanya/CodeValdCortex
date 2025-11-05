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

// ProcessAIRACIRequest handles POST /api/v1/agencies/:id/raci-matrix/ai-generate
// Processes AI operations for RACI matrix creation
func (h *Handler) ProcessAIRACIRequest(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse request body
	var req struct {
		Operations []string `json:"operations" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse AI RACI request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"agency_id":  agencyID,
		"operations": req.Operations,
	}).Info("Processing AI RACI operations")

	// Validate agency exists and get context
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		h.logger.Error("Agency not found", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	// Get work items
	workItems, err := h.agencyService.GetWorkItems(ctx, agencyID)
	if err != nil {
		h.logger.Warn("Failed to get work items", "agencyID", agencyID, "error", err)
		workItems = []*agency.WorkItem{}
	}

	// Get roles (filter out system roles)
	allRoles, err := h.roleService.ListTypes(ctx)
	if err != nil {
		h.logger.Warn("Failed to get roles", "agencyID", agencyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles"})
		return
	}

	// Filter out system roles
	var roles []*registry.Role
	for _, role := range allRoles {
		if !role.IsSystemType {
			roles = append(roles, role)
		}
	}

	results := make(map[string]interface{})

	// Process each operation
	for _, operation := range req.Operations {
		h.logger.Info("Processing operation", "operation", operation, "agencyID", agencyID)

		switch operation {
		case "create":
			h.processCreateRACIMappingsOperation(c, agencyID, ag, workItems, roles, results)
		default:
			results[operation+"_status"] = "Operation not yet implemented"
		}
	}

	// Add AI explanation to chat conversation if there's an explanation
	explanation, hasExplanation := results["ai_explanation"].(string)
	if hasExplanation && explanation != "" {
		h.addRACIExplanationToChat(c, agencyID, explanation)
	}

	// Build response
	response := gin.H{
		"success": true,
		"results": results,
	}

	if assignments, ok := results["assignments"].(map[string]map[string]builder.RACIAssignment); ok {
		response["assignments"] = assignments
		response["mapping_count"] = countTotalAssignments(assignments)
	}

	h.logger.Info("AI RACI operations completed", "agencyID", agencyID)

	c.JSON(http.StatusOK, response)
}

func (h *Handler) processCreateRACIMappingsOperation(
	c *gin.Context,
	agencyID string,
	ag *agency.Agency,
	workItems []*agency.WorkItem,
	roles []*registry.Role,
	results map[string]interface{},
) {
	if len(workItems) == 0 {
		h.logger.Warn("No work items available for RACI mapping", "agencyID", agencyID)
		results["create_error"] = "No work items available. Please create work items first."
		return
	}

	if len(roles) == 0 {
		h.logger.Warn("No roles available for RACI mapping", "agencyID", agencyID)
		results["create_error"] = "No roles available. Please create roles first."
		return
	}

	// Check if we have a RACI builder
	if h.raciBuilder == nil {
		h.logger.Error("RACI builder not available", "agencyID", agencyID)
		results["create_error"] = "RACI creation service not available"
		return
	}

	// Get RACI assignments for context
	assignments, err := h.agencyService.GetAllRACIAssignments(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch RACI assignments, continuing without them")
		assignments = []*agency.RACIAssignment{}
	}

	// Build context for RACI creation
	builderContext := builder.BuilderContext{
		// Agency metadata
		AgencyName:        ag.DisplayName,
		AgencyCategory:    ag.Category,
		AgencyDescription: ag.Description,
		// Agency working data
		Introduction: "",  // Introduction not available in agency struct
		Goals:        nil, // We would need to fetch these
		WorkItems:    workItems,
		Roles:        roles,
		Assignments:  assignments,
		UserInput:    "",
	}

	// Create RACI mappings request
	createReq := &builder.CreateRACIMappingsRequest{
		AgencyID: agencyID,
	}

	h.logger.Info("Calling AI to generate RACI mappings",
		"agencyID", agencyID,
		"workItems", len(workItems),
		"roles", len(roles))

	result, err := h.raciBuilder.CreateRACIMappings(c.Request.Context(), createReq, builderContext)
	if err != nil {
		h.logger.Error("Failed to generate RACI mappings from AI", "agencyID", agencyID, "error", err)
		results["create_error"] = err.Error()
		return
	}

	h.logger.Info("AI generated RACI mappings successfully",
		"agencyID", agencyID,
		"totalAssignments", countTotalAssignments(result.Assignments),
		"workItemsMapped", len(result.Assignments))

	// Save RACI assignments as edges in ArangoDB
	savedCount := 0
	for workItemKey, roleAssignments := range result.Assignments {
		for roleKey, raciData := range roleAssignments {
			assignment := &agency.RACIAssignment{
				WorkItemKey: workItemKey,
				RoleKey:     roleKey,
				RACI:        raciData.RACI,
				Objective:   raciData.Objective,
			}

			err := h.agencyService.CreateRACIAssignment(c.Request.Context(), agencyID, assignment)
			if err != nil {
				h.logger.Error("Failed to save RACI assignment",
					"agencyID", agencyID,
					"workItemKey", workItemKey,
					"roleKey", roleKey,
					"error", err)
				// Continue with other assignments
				continue
			}
			savedCount++
		}
	}

	h.logger.Info("Saved RACI assignments to database",
		"agencyID", agencyID,
		"savedCount", savedCount,
		"totalGenerated", countTotalAssignments(result.Assignments))

	results["create_success"] = fmt.Sprintf("Created %d RACI mapping(s)", savedCount)
	results["ai_explanation"] = result.Explanation
	results["assignments"] = result.Assignments
}

func (h *Handler) addRACIExplanationToChat(c *gin.Context, agencyID string, explanation string) {
	// Create assistant message with RACI generation explanation
	message := fmt.Sprintf("I've generated RACI assignments for your work items and roles.\n\n%s\n\nPlease review the assignments and adjust as needed.", explanation)

	// Get or create conversation for this agency
	conversation, err := h.designerService.GetConversation(agencyID)
	if err != nil || conversation == nil {
		// Try to start a new conversation
		conversation, err = h.designerService.StartConversation(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.Error("Failed to get or start conversation",
				"agencyID", agencyID,
				"error", err)
			return
		}
	}

	err = h.designerService.AddMessage(conversation.ID, "assistant", message)
	if err != nil {
		h.logger.Error("Failed to add RACI explanation to chat",
			"agencyID", agencyID,
			"error", err)
	} else {
		h.logger.Info("Added RACI explanation to chat", "agencyID", agencyID)
	}
}

func countTotalAssignments(assignments map[string]map[string]builder.RACIAssignment) int {
	count := 0
	for _, roleAssignments := range assignments {
		count += len(roleAssignments)
	}
	return count
}
