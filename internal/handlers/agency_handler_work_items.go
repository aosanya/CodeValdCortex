package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
)

// GetWorkItems handles GET /api/v1/agencies/:id/work-items
func (h *AgencyHandler) GetWorkItems(c *gin.Context) {
	id := c.Param("id")

	workItems, err := h.service.GetWorkItems(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workItems)
}

// GetWorkItemsHTML handles GET /api/v1/agencies/:id/work-items/html
// Returns rendered HTML fragment for HTMX/JavaScript rendering
func (h *AgencyHandler) GetWorkItemsHTML(c *gin.Context) {
	id := c.Param("id")

	workItems, err := h.service.GetWorkItems(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading work items")
		return
	}

	// Render the work items list template
	component := agency_designer.WorkItemsList(workItems)
	c.Header("Content-Type", "text/html")
	component.Render(c.Request.Context(), c.Writer)
}

// CreateWorkItem handles POST /api/v1/agencies/:id/work-items
func (h *AgencyHandler) CreateWorkItem(c *gin.Context) {
	id := c.Param("id")

	var req agency.CreateWorkItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	workItem, err := h.service.CreateWorkItem(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, workItem)
}

// UpdateWorkItem handles PUT /api/v1/agencies/:id/work-items/:key
func (h *AgencyHandler) UpdateWorkItem(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	var req agency.UpdateWorkItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := h.service.UpdateWorkItem(c.Request.Context(), id, key, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated work item
	workItem, err := h.service.GetWorkItem(c.Request.Context(), id, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workItem)
}

// DeleteWorkItem handles DELETE /api/v1/agencies/:id/work-items/:key
func (h *AgencyHandler) DeleteWorkItem(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	if err := h.service.DeleteWorkItem(c.Request.Context(), id, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Work item deleted successfully"})
}

// ValidateWorkItemDependencies handles POST /api/v1/agencies/:id/work-items/validate-deps
func (h *AgencyHandler) ValidateWorkItemDependencies(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		WorkItemKey  string   `json:"work_item_key"`
		Dependencies []string `json:"dependencies"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	err := h.service.ValidateWorkItemDependencies(c.Request.Context(), id, req.WorkItemKey, req.Dependencies)
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"valid": true,
	})
}