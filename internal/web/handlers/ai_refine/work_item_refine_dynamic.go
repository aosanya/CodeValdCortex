package ai_refine

import (
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

	// Check if this is a wrapper call with a preset request
	var req struct {
		UserMessage  string   `json:"user_message" binding:"required"` // Natural language instruction
		WorkItemKeys []string `json:"work_item_keys"`                  // Optional: specific work items to operate on
	}

	// First, check if there's a preset request from wrapper methods
	if dynamicReq, exists := c.Get("dynamic_request"); exists {
		if presetReq, ok := dynamicReq.(struct {
			UserMessage  string   `json:"user_message"`
			WorkItemKeys []string `json:"work_item_keys"`
		}); ok {
			req.UserMessage = presetReq.UserMessage
			req.WorkItemKeys = presetReq.WorkItemKeys
			h.logger.WithField("source", "wrapper").Info("Using preset request from wrapper method")
		}
	} else {
		// Parse request body for direct calls
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
							<p class="mb-0">Please provide a user message describing what you want to do with the work items.</p>
						</div>
					</div>
				</div>
			`)
			return
		}
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

	// TODO: Implement RefineWorkItems method in WorkItemsBuilder
	// For now, provide a simple response
	h.logger.Info("Work item refinement placeholder - RefineWorkItems method not yet implemented")

	responseMessage := "🚧 **Work Item Processing**\n\nWork item dynamic processing is being implemented. This will support:\n\n• Refining existing work items for clarity and completeness\n• Generating new work items based on goals\n• Consolidating duplicate or overlapping work items\n• Removing unnecessary work items\n\nRequest received: " + req.UserMessage

	// Format response as bullets
	responseMessage = h.formatExplanationAsBullets(responseMessage)

	// Return success response with placeholder message
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
		<div class="notification is-info">
			<div class="is-flex is-align-items-center">
				<span class="icon has-text-info mr-2">
					<i class="fas fa-info-circle"></i>
				</span>
				<div>
					<strong>Work Item Processing</strong>
					<div style="white-space: pre-line; margin-top: 0.5rem;">%s</div>
				</div>
			</div>
		</div>
	`, strings.ReplaceAll(responseMessage, "\n", "<br>")))
}
