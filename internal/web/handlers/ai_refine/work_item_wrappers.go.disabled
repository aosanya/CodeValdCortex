package ai_refine

import (
	"github.com/gin-gonic/gin"
)

// RefineSpecificWorkItem handles POST /api/v1/agencies/:id/work-items/:workItemKey/refine
// Wrapper around RefineWorkItems with a preset prompt for specific work item refinement
func (h *Handler) RefineSpecificWorkItem(c *gin.Context) {
	workItemKey := c.Param("workItemKey")

	// Set up the dynamic request for RefineWorkItems with preset prompt
	dynamicReq := struct {
		UserMessage  string   `json:"user_message"`
		WorkItemKeys []string `json:"work_item_keys"`
	}{
		UserMessage:  "Please refine and improve this specific work item to be clearer, more detailed, and better aligned with the agency's goals. Focus on improving the title, description, deliverables, and ensuring it follows best practices.",
		WorkItemKeys: []string{workItemKey},
	}

	// Store the preset request in the context for RefineWorkItems to use
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineWorkItems handler
	h.RefineWorkItems(c)
}

// GenerateWorkItemWithPrompt handles POST /api/v1/agencies/:id/work-items/generate
// Wrapper around RefineWorkItems with a preset prompt for work item generation
func (h *Handler) GenerateWorkItemWithPrompt(c *gin.Context) {
	// Set up the dynamic request for RefineWorkItems with preset prompt
	dynamicReq := struct {
		UserMessage  string   `json:"user_message"`
		WorkItemKeys []string `json:"work_item_keys"`
	}{
		UserMessage:  "Based on the agency's goals and current context, generate new work items that will help achieve the strategic objectives. Focus on creating actionable, well-defined work items with clear deliverables and appropriate effort estimates.",
		WorkItemKeys: nil, // No specific work items, generate new ones
	}

	// Store the preset request in the context for RefineWorkItems to use
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineWorkItems handler
	h.RefineWorkItems(c)
}

// ConsolidateWorkItemsWithPrompt handles POST /api/v1/agencies/:id/work-items/consolidate
// Wrapper around RefineWorkItems with a preset prompt for work item consolidation
func (h *Handler) ConsolidateWorkItemsWithPrompt(c *gin.Context) {
	// Set up the dynamic request for RefineWorkItems with preset prompt
	dynamicReq := struct {
		UserMessage  string   `json:"user_message"`
		WorkItemKeys []string `json:"work_item_keys"`
	}{
		UserMessage:  "Analyze all work items and consolidate any that are duplicate, overlapping, or can be combined without losing important details. Focus on reducing redundancy while maintaining comprehensive coverage of all necessary tasks and deliverables.",
		WorkItemKeys: nil, // Analyze all work items for consolidation
	}

	// Store the preset request in the context for RefineWorkItems to use
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineWorkItems handler
	h.RefineWorkItems(c)
}

// EnhanceAllWorkItems handles POST /api/v1/agencies/:id/work-items/enhance-all
// Wrapper around RefineWorkItems with a preset prompt for enhancing all work items
func (h *Handler) EnhanceAllWorkItems(c *gin.Context) {
	// Set up the dynamic request for RefineWorkItems with preset prompt
	dynamicReq := struct {
		UserMessage  string   `json:"user_message"`
		WorkItemKeys []string `json:"work_item_keys"`
	}{
		UserMessage:  "Review and enhance all work items to ensure they are comprehensive, well-defined, and strategically aligned. Improve titles, descriptions, deliverables, effort estimates, and tags. Ensure each work item has clear acceptance criteria and measurable outcomes.",
		WorkItemKeys: nil, // Enhance all work items
	}

	// Store the preset request in the context for RefineWorkItems to use
	c.Set("dynamic_request", dynamicReq)

	// Call the main RefineWorkItems handler
	h.RefineWorkItems(c)
}
