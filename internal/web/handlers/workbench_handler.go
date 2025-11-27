package handlers

import (
	"fmt"
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/web/pages"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WorkbenchHandler handles workbench and issue management routes
type WorkbenchHandler struct {
	issueService     *services.IssueService
	workbenchService *services.WorkbenchService
	instanceService  services.InstanceService
	agencyService    agency.Service
	logger           *logrus.Logger
}

// NewWorkbenchHandler creates a new workbench handler
func NewWorkbenchHandler(
	issueService *services.IssueService,
	workbenchService *services.WorkbenchService,
	instanceService services.InstanceService,
	agencyService agency.Service,
	logger *logrus.Logger,
) *WorkbenchHandler {
	return &WorkbenchHandler{
		issueService:     issueService,
		workbenchService: workbenchService,
		instanceService:  instanceService,
		agencyService:    agencyService,
		logger:           logger,
	}
}

// ShowInstanceSelector displays a list of instances for workbench selection
func (h *WorkbenchHandler) ShowInstanceSelector(c *gin.Context) {
	agencyID := c.Param("id")

	// Get agency
	agency, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load agency",
		})
		return
	}

	// Get instances for this agency
	var instances []*models.AgencyInstance
	if h.instanceService != nil {
		instances, err = h.instanceService.ListInstances(c.Request.Context(), agencyID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to list instances")
			// Continue with empty list rather than failing completely
			instances = []*models.AgencyInstance{}
		}
	} else {
		h.logger.Warn("Instance service not available")
		instances = []*models.AgencyInstance{}
	}

	// Render instance selector page
	component := pages.WorkbenchInstanceSelector(agency, instances)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render workbench instance selector")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to render page",
		})
		return
	}
}

// ShowWorkbench displays the Kanban board for an instance
func (h *WorkbenchHandler) ShowWorkbench(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")
	tagName := c.Query("tag") // Optional: specific tag to use

	// Get agency
	agency, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load agency",
		})
		return
	}

	// Generate workbench board
	var board *models.WorkbenchBoard
	if tagName != "" {
		// Use specific tag
		board, err = h.workbenchService.GenerateBoard(c.Request.Context(), agencyID, instanceID, tagName)
	} else {
		// Use current specification
		board, err = h.workbenchService.GenerateBoardFromSpecification(c.Request.Context(), agencyID, instanceID)
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to generate workbench board")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to generate workbench board",
		})
		return
	}

	// Render workbench page
	component := pages.Workbench(agency, board)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render workbench")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to render page",
		})
		return
	}
}

// CreateIssue creates a new issue
func (h *WorkbenchHandler) CreateIssue(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")

	var req models.CreateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user ID from context/session
	createdBy := "system"

	issue, err := h.issueService.CreateIssue(c.Request.Context(), agencyID, instanceID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create issue"})
		return
	}

	// Set created_by
	issue.CreatedBy = createdBy

	c.JSON(http.StatusCreated, issue)
}

// GetIssue retrieves a specific issue
func (h *WorkbenchHandler) GetIssue(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")
	issueID := c.Param("issue_id")

	issue, err := h.issueService.GetIssue(c.Request.Context(), agencyID, instanceID, issueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get issue")
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}

	c.JSON(http.StatusOK, issue)
}

// UpdateIssue updates an existing issue
func (h *WorkbenchHandler) UpdateIssue(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")
	issueID := c.Param("issue_id")

	var req models.UpdateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	issue, err := h.issueService.UpdateIssue(c.Request.Context(), agencyID, instanceID, issueID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update issue"})
		return
	}

	c.JSON(http.StatusOK, issue)
}

// DeleteIssue deletes an issue
func (h *WorkbenchHandler) DeleteIssue(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")
	issueID := c.Param("issue_id")

	err := h.issueService.DeleteIssue(c.Request.Context(), agencyID, instanceID, issueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete issue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Issue deleted successfully"})
}

// AssignIssue assigns an issue to a worker
func (h *WorkbenchHandler) AssignIssue(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")
	issueID := c.Param("issue_id")

	var req models.AssignIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	issue, err := h.issueService.AssignIssue(c.Request.Context(), agencyID, instanceID, issueID, req.WorkerID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to assign issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign issue"})
		return
	}

	c.JSON(http.StatusOK, issue)
}

// ClaimIssue allows a worker to claim an available issue
func (h *WorkbenchHandler) ClaimIssue(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")
	issueID := c.Param("issue_id")

	// TODO: Get worker ID from context/session
	workerID := "current-user" // Placeholder

	issue, err := h.issueService.ClaimIssue(c.Request.Context(), agencyID, instanceID, issueID, workerID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to claim issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to claim issue"})
		return
	}

	c.JSON(http.StatusOK, issue)
}

// ListIssues retrieves all issues for an instance with optional filters
func (h *WorkbenchHandler) ListIssues(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")

	// Parse filters from query parameters
	filters := agency.IssueFilters{
		WorkflowID: c.Query("workflow_id"),
		Status:     c.Query("status"),
		AssignedTo: c.Query("assigned_to"),
	}

	// Parse pagination
	var offset, limit int
	if c.Query("offset") != "" {
		offset = parseIntOrDefault(c.Query("offset"), 0)
	}
	if c.Query("limit") != "" {
		limit = parseIntOrDefault(c.Query("limit"), 50)
	} else {
		limit = 50 // Default limit
	}

	filters.Offset = offset
	filters.Limit = limit

	issues, err := h.issueService.ListIssues(c.Request.Context(), agencyID, instanceID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list issues")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list issues"})
		return
	}

	// Convert to response format
	issueValues := make([]models.WorkIssue, len(issues))
	for i, issue := range issues {
		issueValues[i] = *issue
	}

	response := models.IssueListResponse{
		Issues: issueValues,
		Total:  len(issueValues), // TODO: Get actual total count from repository
		Offset: offset,
		Limit:  limit,
	}

	c.JSON(http.StatusOK, response)
}

// ProgressIssue moves an issue to the next workflow step
func (h *WorkbenchHandler) ProgressIssue(c *gin.Context) {
	agencyID := c.Param("id")
	instanceID := c.Param("instance_id")
	issueID := c.Param("issue_id")

	issue, err := h.issueService.ProgressIssue(c.Request.Context(), agencyID, instanceID, issueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to progress issue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to progress issue"})
		return
	}

	c.JSON(http.StatusOK, issue)
}

// Helper function to parse int with default value
func parseIntOrDefault(s string, defaultVal int) int {
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
		return defaultVal
	}
	return val
}
