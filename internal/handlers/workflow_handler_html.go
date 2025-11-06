package handlers

import (
	"net/http"
	"sort"

	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
)

// GetWorkflowsHTML handles GET /api/v1/agencies/:id/workflows/html
// Returns rendered HTML fragment for HTMX/JavaScript rendering
func (h *WorkflowHandler) GetWorkflowsHTML(c *gin.Context) {
	agencyID := c.Param("id")

	workflows, err := h.service.GetWorkflowsByAgency(c.Request.Context(), agencyID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading workflows")
		return
	}

	// Sort workflows by Name
	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].Name < workflows[j].Name
	})

	// Render the workflows list template
	component := agency_designer.WorkflowsList(workflows)
	c.Header("Content-Type", "text/html")
	component.Render(c.Request.Context(), c.Writer)
}
