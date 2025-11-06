package ai_refine

import (
	"github.com/gin-gonic/gin"
)

// RefineSpecificRACIMapping refines a specific RACI assignment using a preset prompt
func (h *Handler) RefineSpecificRACIMapping(c *gin.Context) {
	// Create a dynamic request for RACI assignment refinement
	dynamicReq := struct {
		UserMessage        string   `json:"user_message"`
		TargetWorkItemKeys []string `json:"target_work_item_keys"`
		TargetRoleKeys     []string `json:"target_role_keys"`
	}{
		UserMessage:        "REFINE_SPECIFIC_RACI_PRESET: Please analyze and refine the provided RACI assignment to improve clarity, ensure proper responsibility distribution, and optimize role-to-work-item mappings based on the agency's objectives.",
		TargetWorkItemKeys: []string{}, // Will be set based on request
		TargetRoleKeys:     []string{}, // Will be set based on request
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineRACIMappings handler
	h.RefineRACIMappings(c)
}

// GenerateRACIMappingWithPrompt generates new RACI assignments using a preset prompt
func (h *Handler) GenerateRACIMappingWithPrompt(c *gin.Context) {
	// Parse request body to get user input
	var req struct {
		UserInput string `json:"userInput" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse generate RACI request")
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}

	// Create a dynamic request for RACI generation
	dynamicReq := struct {
		UserMessage        string   `json:"user_message"`
		TargetWorkItemKeys []string `json:"target_work_item_keys"`
		TargetRoleKeys     []string `json:"target_role_keys"`
	}{
		UserMessage:        "GENERATE_RACI_PRESET: Please analyze the agency's work items and roles to generate appropriate RACI assignments. Consider each role's capabilities and the requirements of each work item. User request: " + req.UserInput,
		TargetWorkItemKeys: []string{}, // Empty - we're creating new assignments
		TargetRoleKeys:     []string{}, // Empty - we're considering all roles
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineRACIMappings handler
	h.RefineRACIMappings(c)
}

// ConsolidateRACIMappingsWithPrompt consolidates duplicate or overlapping RACI assignments using a preset prompt
func (h *Handler) ConsolidateRACIMappingsWithPrompt(c *gin.Context) {
	// Create a dynamic request for RACI consolidation
	dynamicReq := struct {
		UserMessage        string   `json:"user_message"`
		TargetWorkItemKeys []string `json:"target_work_item_keys"`
		TargetRoleKeys     []string `json:"target_role_keys"`
	}{
		UserMessage:        "CONSOLIDATE_RACI_PRESET: Please analyze all existing RACI assignments and identify any that are duplicated, conflicting, or could be optimized. Consolidate them into a clear, efficient responsibility matrix while ensuring all work items have proper coverage.",
		TargetWorkItemKeys: []string{}, // Empty - we're working with all assignments
		TargetRoleKeys:     []string{}, // Empty - we're working with all roles
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineRACIMappings handler
	h.RefineRACIMappings(c)
}

// CreateCompleteRACIMatrixWithPrompt creates a complete RACI matrix using a preset prompt
func (h *Handler) CreateCompleteRACIMatrixWithPrompt(c *gin.Context) {
	// Create a dynamic request for complete RACI matrix creation
	dynamicReq := struct {
		UserMessage        string   `json:"user_message"`
		TargetWorkItemKeys []string `json:"target_work_item_keys"`
		TargetRoleKeys     []string `json:"target_role_keys"`
	}{
		UserMessage:        "CREATE_COMPLETE_RACI_PRESET: Please create a comprehensive RACI matrix that covers all work items and roles. Ensure each work item has exactly one Accountable role, appropriate Responsible roles, and relevant Consulted/Informed roles based on the agency's structure and objectives.",
		TargetWorkItemKeys: []string{}, // Empty - we're working with all work items
		TargetRoleKeys:     []string{}, // Empty - we're working with all roles
	}

	// Set the request body for the dynamic handler
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineRACIMappings handler
	h.RefineRACIMappings(c)
}
