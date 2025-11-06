package ai_refine

import (
	"fmt"
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/workflow"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RefineWorkflows handles POST /api/v1/agencies/:id/workflows/refine-dynamic
// Dynamically determines and executes the appropriate workflow operation based on user message
func (h *Handler) RefineWorkflows(c *gin.Context) {
	agencyID := c.Param("id")

	h.logger.WithField("agency_id", agencyID).Info("Processing dynamic AI workflow refinement request")

	var req struct {
		UserMessage  string   `json:"user_message" binding:"required"` // Natural language instruction
		WorkflowKeys []string `json:"workflow_keys"`                   // Optional: specific workflows to operate on
	}

	// Check if there's a preset request from wrapper methods
	if dynamicReq, exists := c.Get("dynamic_request"); exists {
		if presetReq, ok := dynamicReq.(struct {
			UserMessage  string   `json:"user_message"`
			WorkflowKeys []string `json:"workflow_keys"`
		}); ok {
			req.UserMessage = presetReq.UserMessage
			req.WorkflowKeys = presetReq.WorkflowKeys
			h.logger.WithField("source", "wrapper").Info("Using preset request from wrapper method")
		}
	} else {
		// Parse request body for direct calls
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.WithError(err).Error("Failed to parse dynamic workflow refinement request")
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, `
				<div class="notification is-danger">
					<div class="is-flex is-align-items-center">
						<span class="icon has-text-danger mr-2">
							<i class="fas fa-exclamation-circle"></i>
						</span>
						<div>
							<strong>Invalid Request</strong>
							<p class="mb-0">Please provide a message describing what you want to do with the workflows.</p>
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
		h.logger.WithError(err).Error("Failed to get agency")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Error</strong>
						<p class="mb-0">Failed to load agency information.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Fetch existing workflows
	existingWorkflows, err := h.workflowService.GetWorkflowsByAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch workflows")
		existingWorkflows = []*workflow.Workflow{} // Continue with empty list
	}

	// Get work items for context
	workItems, err := h.agencyService.GetWorkItems(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch work items for workflow context")
		workItems = []*agency.WorkItem{}
	}

	// Convert pointers to values for AI builder
	workItemValues := make([]agency.WorkItem, len(workItems))
	for i, wi := range workItems {
		workItemValues[i] = *wi
	}

	// Get overview for additional context (note: there's no GetOverview method, so we'll skip this)
	var overview *agency.Overview = nil

	// Call the AI to determine the operation type
	h.logger.WithFields(logrus.Fields{
		"user_message":    req.UserMessage,
		"workflow_count":  len(existingWorkflows),
		"work_item_count": len(workItemValues),
	}).Info("Determining workflow operation type")

	// For now, default to generating new workflows from context
	// TODO: Implement AI-powered operation determination (generate, refine, suggest improvements)

	if len(existingWorkflows) == 0 {
		// No workflows exist, generate from context
		h.generateWorkflowsFromContext(c, ag, overview, workItemValues)
		return
	}

	// If workflows exist and user wants to refine a specific one
	if len(req.WorkflowKeys) == 1 {
		h.refineSpecificWorkflow(c, req.UserMessage, req.WorkflowKeys[0], workItemValues)
		return
	}

	// Default: generate new workflow from prompt
	h.generateWorkflowFromPrompt(c, ag, req.UserMessage, workItemValues)
}

// generateWorkflowsFromContext generates workflows based on agency goals and work items
func (h *Handler) generateWorkflowsFromContext(c *gin.Context, ag *agency.Agency, overview *agency.Overview, workItems []agency.WorkItem) {
	// Note: GenerateWorkflowsFromContext doesn't take goals parameter
	h.logger.WithFields(logrus.Fields{
		"agency_id":  ag.ID,
		"work_items": len(workItems),
	}).Info("Generating workflows from agency context")

	workflows, err := h.workflowBuilder.GenerateWorkflowsFromContext(c.Request.Context(), ag, overview, workItems)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate workflows from context")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Generation Failed</strong>
						<p class="mb-0">Failed to generate workflows from context.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Save generated workflows
	savedCount := 0
	for _, wf := range workflows {
		wf.AgencyID = ag.ID
		if err := h.workflowService.CreateWorkflow(c.Request.Context(), &wf); err != nil {
			h.logger.WithError(err).WithField("workflow_name", wf.Name).Error("Failed to save generated workflow")
			continue
		}
		savedCount++
	}

	h.logger.WithField("count", savedCount).Info("Successfully generated and saved workflows")

	// Return success HTML
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
		<div class="notification is-success">
			<div class="is-flex is-align-items-center">
				<span class="icon has-text-success mr-2">
					<i class="fas fa-check-circle"></i>
				</span>
				<div>
					<strong>Workflows Generated</strong>
					<p class="mb-0">Successfully generated %d workflow(s) from your agency context.</p>
				</div>
			</div>
		</div>
	`, savedCount))
}

// generateWorkflowFromPrompt generates a single workflow from user's natural language prompt
func (h *Handler) generateWorkflowFromPrompt(c *gin.Context, ag *agency.Agency, userPrompt string, workItems []agency.WorkItem) {
	h.logger.WithFields(logrus.Fields{
		"agency_id": ag.ID,
		"prompt":    userPrompt,
	}).Info("Generating workflow from user prompt")

	wf, err := h.workflowBuilder.GenerateWorkflowWithPrompt(c.Request.Context(), ag, userPrompt, workItems)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate workflow from prompt")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Generation Failed</strong>
						<p class="mb-0">Failed to generate workflow from your prompt.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Save workflow
	wf.AgencyID = ag.ID
	if err := h.workflowService.CreateWorkflow(c.Request.Context(), wf); err != nil {
		h.logger.WithError(err).Error("Failed to save generated workflow")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Save Failed</strong>
						<p class="mb-0">Failed to save the generated workflow.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	h.logger.WithField("workflow_id", wf.ID).Info("Successfully generated and saved workflow")

	// Return success HTML
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
		<div class="notification is-success">
			<div class="is-flex is-align-items-center">
				<span class="icon has-text-success mr-2">
					<i class="fas fa-check-circle"></i>
				</span>
				<div>
					<strong>Workflow Generated</strong>
					<p class="mb-0">Successfully generated workflow: %s</p>
				</div>
			</div>
		</div>
	`, wf.Name))
}

// refineSpecificWorkflow refines an existing workflow based on user feedback
func (h *Handler) refineSpecificWorkflow(c *gin.Context, userFeedback string, workflowKey string, workItems []agency.WorkItem) {
	h.logger.WithFields(logrus.Fields{
		"workflow_key": workflowKey,
		"feedback":     userFeedback,
	}).Info("Refining specific workflow")

	// Get the workflow
	wf, err := h.workflowService.GetWorkflow(c.Request.Context(), workflowKey)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch workflow for refinement")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusNotFound, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Workflow Not Found</strong>
						<p class="mb-0">Could not find the workflow to refine.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Call AI to refine (RefineWorkflow takes 3 args: ctx, workflow, prompt)
	refined, err := h.workflowBuilder.RefineWorkflow(c.Request.Context(), wf, userFeedback)
	if err != nil {
		h.logger.WithError(err).Error("Failed to refine workflow")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Refinement Failed</strong>
						<p class="mb-0">Failed to refine the workflow.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	// Update workflow
	if err := h.workflowService.UpdateWorkflow(c.Request.Context(), refined); err != nil {
		h.logger.WithError(err).Error("Failed to save refined workflow")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusInternalServerError, `
			<div class="notification is-danger">
				<div class="is-flex is-align-items-center">
					<span class="icon has-text-danger mr-2">
						<i class="fas fa-exclamation-circle"></i>
					</span>
					<div>
						<strong>Save Failed</strong>
						<p class="mb-0">Failed to save the refined workflow.</p>
					</div>
				</div>
			</div>
		`)
		return
	}

	h.logger.WithField("workflow_id", refined.ID).Info("Successfully refined workflow")

	// Return success HTML
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
		<div class="notification is-success">
			<div class="is-flex is-align-items-center">
				<span class="icon has-text-success mr-2">
					<i class="fas fa-check-circle"></i>
				</span>
				<div>
					<strong>Workflow Refined</strong>
					<p class="mb-0">Successfully refined workflow: %s</p>
				</div>
			</div>
		</div>
	`, refined.Name))
}
