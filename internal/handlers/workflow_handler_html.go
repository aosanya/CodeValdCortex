package handlers

import (
	"net/http"
	"sort"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetWorkflowsHTML handles GET /api/v1/agencies/:id/workflows/html
// Returns rendered HTML fragment for HTMX/JavaScript rendering
func (h *WorkflowHandler) GetWorkflowsHTML(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithField("agency_id", agencyID).Info("🔍 GetWorkflowsHTML: Fetching specification")

	// Fetch workflows from specification (standardized with goals/work-items/roles)
	spec, err := h.agencyService.GetSpecification(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get specification")
		c.String(http.StatusInternalServerError, "Error loading workflows")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"total_workflows": len(spec.Workflows),
	}).Info("🔍 GetWorkflowsHTML: Retrieved specification")

	// Log each workflow for debugging
	for i, wf := range spec.Workflows {
		h.logger.WithFields(logrus.Fields{
			"index":       i,
			"key":         wf.Key,
			"name":        wf.Name,
			"agency_id":   wf.AgencyID,
			"description": truncateForLog(wf.Description, 50),
			"nodes_count": len(wf.Nodes),
			"edges_count": len(wf.Edges),
		}).Info("  🔹 Workflow retrieved")
	}

	// Convert []Workflow to []*Workflow for compatibility with template
	workflows := make([]*models.Workflow, len(spec.Workflows))
	for i := range spec.Workflows {
		workflows[i] = &spec.Workflows[i]
		h.logger.WithFields(logrus.Fields{
			"index":     i,
			"key":       workflows[i].Key,
			"name":      workflows[i].Name,
			"agency_id": workflows[i].AgencyID,
			"has_key":   workflows[i].Key != "",
		}).Info("  🔸 Workflow pointer created for template")
	}

	// Sort workflows by Name
	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].Name < workflows[j].Name
	})

	h.logger.WithFields(logrus.Fields{
		"agency_id":       agencyID,
		"workflows_count": len(workflows),
	}).Info("📋 Rendering workflows HTML")

	// Render the workflows list template
	component := agency_designer.WorkflowsList(workflows)
	c.Header("Content-Type", "text/html")
	component.Render(c.Request.Context(), c.Writer)
}

// truncateForLog returns a truncated version of a string for logging
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
