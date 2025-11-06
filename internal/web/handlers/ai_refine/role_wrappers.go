package ai_refine

import (
	"github.com/gin-gonic/gin"
)

// RefineSpecificRole refines a specific role using a preset prompt
func (h *Handler) RefineSpecificRole(c *gin.Context) {
	// Create a dynamic request for role refinement
	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		RoleKeys    []string `json:"role_keys"`
	}{
		UserMessage: "REFINE_SPECIFIC_ROLE_PRESET: Please analyze and refine the provided role to improve its description, capabilities, autonomy level, and required skills based on the agency's work items and context.",
		RoleKeys:    []string{}, // Will be set based on request
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineRoles handler
	h.RefineRoles(c)
}

// GenerateRoleWithPrompt generates a new role using a preset prompt
func (h *Handler) GenerateRoleWithPrompt(c *gin.Context) {
	// Parse request body to get user input
	var req struct {
		UserInput string `json:"userInput" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse generate role request")
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}

	// Create a dynamic request for role generation
	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		RoleKeys    []string `json:"role_keys"`
	}{
		UserMessage: "GENERATE_ROLE_PRESET: Please analyze the agency's work items and generate a new role that would be valuable for executing the work. Consider what gaps exist in the current role coverage and what specialized capabilities are needed. User request: " + req.UserInput,
		RoleKeys:    []string{}, // Empty - we're creating new roles
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineRoles handler
	h.RefineRoles(c)
}

// ConsolidateRolesWithPrompt consolidates duplicate or overlapping roles using a preset prompt
func (h *Handler) ConsolidateRolesWithPrompt(c *gin.Context) {
	// Create a dynamic request for role consolidation
	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		RoleKeys    []string `json:"role_keys"`
	}{
		UserMessage: "CONSOLIDATE_ROLES_PRESET: Please analyze all existing roles and identify any that are duplicated, overlapping, or could be merged. Consolidate them into a more efficient role structure while preserving all necessary capabilities.",
		RoleKeys:    []string{}, // Empty - we're working with all roles
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineRoles handler
	h.RefineRoles(c)
}

// EnhanceAllRolesWithPrompt enhances all roles using a preset prompt
func (h *Handler) EnhanceAllRolesWithPrompt(c *gin.Context) {
	// Create a dynamic request for enhancing all roles
	dynamicReq := struct {
		UserMessage string   `json:"user_message"`
		RoleKeys    []string `json:"role_keys"`
	}{
		UserMessage: "ENHANCE_ALL_ROLES_PRESET: Please analyze and enhance all existing roles by improving their descriptions, refining their capabilities, optimizing their autonomy levels, and updating their required skills based on the agency's current work items and objectives.",
		RoleKeys:    []string{}, // Empty - we're working with all roles
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineRoles handler
	h.RefineRoles(c)
}
