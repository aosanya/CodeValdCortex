package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/aosanya/CodeValdCortex/internal/workflow"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WorkflowDesignerHandler handles visual workflow designer pages
type WorkflowDesignerHandler struct {
	workflowService *workflow.Service
	logger          *logrus.Logger
}

// NewWorkflowDesignerHandler creates a new workflow designer web handler
func NewWorkflowDesignerHandler(workflowService *workflow.Service, logger *logrus.Logger) *WorkflowDesignerHandler {
	return &WorkflowDesignerHandler{
		workflowService: workflowService,
		logger:          logger,
	}
}

// ShowDesigner renders the visual workflow designer page
func (h *WorkflowDesignerHandler) ShowDesigner(c *gin.Context) {
	agencyID := c.Param("id")
	workflowID := c.Param("workflowId")

	// Get workflow from service
	wf, err := h.workflowService.GetWorkflow(c.Request.Context(), workflowID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workflow")
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Workflow not found",
		})
		return
	}

	// Verify workflow belongs to agency
	if wf.AgencyID != agencyID {
		h.logger.Warn("Workflow does not belong to agency")
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error": "Access denied",
		})
		return
	}

	// Render designer page
	component := agency_designer.WorkflowDesigner(agencyID, wf)
	component.Render(c.Request.Context(), c.Writer)
}
