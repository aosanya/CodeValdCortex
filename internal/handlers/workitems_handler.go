package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aosanya/CodeValdCortex/internal/agency/arangodb"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/api"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WorkItemsHandler handles work items REST API requests
type WorkItemsHandler struct {
	repo   *arangodb.WorkItemRepository
	logger *logrus.Logger
}

// NewWorkItemsHandler creates a new work items API handler
func NewWorkItemsHandler(repo *arangodb.WorkItemRepository, logger *logrus.Logger) *WorkItemsHandler {
	return &WorkItemsHandler{
		repo:   repo,
		logger: logger,
	}
}

// ListWorkItems handles GET /api/v1/agencies/:id/work-items
func (h *WorkItemsHandler) ListWorkItems(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse query parameters for filtering
	filters := make(map[string]interface{})
	if workType := c.Query("type"); workType != "" {
		filters["type"] = workType
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	// Parse pagination parameters
	limit := 50
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get work items from repository
	items, err := h.repo.ListWorkItems(c.Request.Context(), agencyID, filters)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list work items")
		api.InternalError(c, "Failed to list work items", err.Error())
		return
	}

	// Apply pagination
	total := len(items)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}
	paginatedItems := items[offset:end]

	// Build response with pagination metadata
	response := map[string]interface{}{
		"data": paginatedItems,
		"pagination": map[string]interface{}{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_more": end < total,
		},
	}

	api.SuccessResponse(c, response)
}

// GetWorkItem handles GET /api/v1/agencies/:id/work-items/:workItemID
func (h *WorkItemsHandler) GetWorkItem(c *gin.Context) {
	agencyID := c.Param("id")
	workItemID := c.Param("workItemID")

	item, err := h.repo.GetWorkItem(c.Request.Context(), agencyID, workItemID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"agency_id":    agencyID,
			"work_item_id": workItemID,
		}).Error("Failed to get work item")
		api.NotFoundError(c, "Work item not found")
		return
	}

	api.SuccessResponse(c, item)
}

// CreateWorkItem handles POST /api/v1/agencies/:id/work-items
func (h *WorkItemsHandler) CreateWorkItem(c *gin.Context) {
	agencyID := c.Param("id")

	var req models.CreateWorkItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind create work item request")
		api.BadRequestError(c, "Invalid request format", err.Error())
		return
	}

	// Validate request
	if err := validateCreateWorkItemRequest(&req); err != nil {
		api.ValidationError(c, err.Error(), nil)
		return
	}

	// Create work item model
	workItem := &models.WorkItem{
		AgencyID:               agencyID,
		Code:                   req.Code,
		Title:                  req.Title,
		Description:            req.Description,
		Deliverables:           req.Deliverables,
		DeliverablesStructured: req.DeliverablesStructured,
		GoalKeys:               req.GoalKeys,
		Tags:                   req.Tags,
	}

	// Ensure IDs in deliverables tree
	if len(workItem.DeliverablesStructured) > 0 {
		models.EnsureAllIDsInTree(workItem.DeliverablesStructured)
	}

	// Create in repository
	created, err := h.repo.CreateWorkItem(c.Request.Context(), workItem)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create work item")
		api.InternalError(c, "Failed to create work item", err.Error())
		return
	}

	c.JSON(http.StatusCreated, api.Response{
		Success: true,
		Data:    created,
		Metadata: &api.Metadata{
			Timestamp: created.CreatedAt,
			RequestID: api.GetRequestID(c),
			Version:   "v1",
		},
	})
}

// UpdateWorkItem handles PUT /api/v1/agencies/:id/work-items/:workItemID
func (h *WorkItemsHandler) UpdateWorkItem(c *gin.Context) {
	agencyID := c.Param("id")
	workItemID := c.Param("workItemID")

	var req models.UpdateWorkItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind update work item request")
		api.BadRequestError(c, "Invalid request format", err.Error())
		return
	}

	// Validate request
	if err := validateUpdateWorkItemRequest(&req); err != nil {
		api.ValidationError(c, err.Error(), nil)
		return
	}

	// Create update model
	updates := &models.WorkItem{
		Code:                   req.Code,
		Title:                  req.Title,
		Description:            req.Description,
		Deliverables:           req.Deliverables,
		DeliverablesStructured: req.DeliverablesStructured,
		GoalKeys:               req.GoalKeys,
		Tags:                   req.Tags,
	}

	// Ensure IDs in deliverables tree
	if len(updates.DeliverablesStructured) > 0 {
		models.EnsureAllIDsInTree(updates.DeliverablesStructured)
	}

	// Update in repository
	updated, err := h.repo.UpdateWorkItem(c.Request.Context(), agencyID, workItemID, updates)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"agency_id":    agencyID,
			"work_item_id": workItemID,
		}).Error("Failed to update work item")
		api.InternalError(c, "Failed to update work item", err.Error())
		return
	}

	api.SuccessResponse(c, updated)
}

// DeleteWorkItem handles DELETE /api/v1/agencies/:id/work-items/:workItemID
func (h *WorkItemsHandler) DeleteWorkItem(c *gin.Context) {
	agencyID := c.Param("id")
	workItemID := c.Param("workItemID")

	err := h.repo.DeleteWorkItem(c.Request.Context(), agencyID, workItemID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"agency_id":    agencyID,
			"work_item_id": workItemID,
		}).Error("Failed to delete work item")
		api.NotFoundError(c, "Work item not found")
		return
	}

	c.Status(http.StatusNoContent)
}

// Validation helpers

func validateCreateWorkItemRequest(req *models.CreateWorkItemRequest) error {
	if req.Code == "" {
		return fmt.Errorf("code is required")
	}
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(req.Title) > 200 {
		return fmt.Errorf("title must be under 200 characters")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(req.Description) > 5000 {
		return fmt.Errorf("description must be under 5000 characters")
	}
	return nil
}

func validateUpdateWorkItemRequest(req *models.UpdateWorkItemRequest) error {
	return validateCreateWorkItemRequest((*models.CreateWorkItemRequest)(req))
}
