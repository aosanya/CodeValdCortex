package handlers

import (
	"net/http"
	"strconv"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/workflow"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WorkflowHandler handles HTTP requests for workflows
type WorkflowHandler struct {
	service *workflow.Service
	logger  *logrus.Logger
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(service *workflow.Service, logger *logrus.Logger) *WorkflowHandler {
	return &WorkflowHandler{
		service: service,
		logger:  logger,
	}
}

// CreateWorkflow handles POST /api/v1/agencies/:id/workflows
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	agencyID := c.Param("id")

	var req models.Workflow
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Set agency ID from URL
	req.AgencyID = agencyID

	// TODO: Get user from context/session
	req.CreatedBy = "system"

	if err := h.service.CreateWorkflow(c.Request.Context(), &req); err != nil {
		h.logger.WithError(err).Error("Failed to create workflow")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetWorkflows handles GET /api/v1/agencies/:id/workflows
func (h *WorkflowHandler) GetWorkflows(c *gin.Context) {
	agencyID := c.Param("id")

	workflows, err := h.service.GetWorkflowsByAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workflows")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflows"})
		return
	}

	c.JSON(http.StatusOK, workflows)
}

// GetWorkflow handles GET /api/v1/workflows/:id
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	id := c.Param("id")

	wf, err := h.service.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get workflow")
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, wf)
}

// UpdateWorkflow handles PUT /api/v1/workflows/:id
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id := c.Param("id")

	var req models.Workflow
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.ID = id

	if err := h.service.UpdateWorkflow(c.Request.Context(), &req); err != nil {
		h.logger.WithError(err).Error("Failed to update workflow")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, req)
}

// DeleteWorkflow handles DELETE /api/v1/workflows/:id
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteWorkflow(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).Error("Failed to delete workflow")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow deleted successfully"})
}

// DuplicateWorkflow handles POST /api/v1/workflows/:id/duplicate
func (h *WorkflowHandler) DuplicateWorkflow(c *gin.Context) {
	id := c.Param("id")

	duplicate, err := h.service.DuplicateWorkflow(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to duplicate workflow")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to duplicate workflow", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, duplicate)
}

// ValidateWorkflow handles POST /api/v1/workflows/validate
func (h *WorkflowHandler) ValidateWorkflow(c *gin.Context) {
	var req models.Workflow
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	result := h.service.ValidateWorkflowStructure(&req)
	c.JSON(http.StatusOK, result)
}

// ListWorkflows handles GET /api/v1/workflows with pagination
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	workflows, err := h.service.ListWorkflows(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list workflows")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list workflows"})
		return
	}

	c.JSON(http.StatusOK, workflows)
}

// StartExecution handles POST /api/v1/workflows/:id/execute
func (h *WorkflowHandler) StartExecution(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Context map[string]interface{} `json:"context"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Get user from context/session
	startedBy := "system"

	execution, err := h.service.StartExecution(c.Request.Context(), id, startedBy, req.Context)
	if err != nil {
		h.logger.WithError(err).Error("Failed to start execution")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start execution", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, execution)
}
