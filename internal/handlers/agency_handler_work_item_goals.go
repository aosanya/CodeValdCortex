package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/gin-gonic/gin"
)

// GetWorkItemGoalLinks handles GET /api/v1/agencies/:id/work-items/:key/goals
// Returns all goal links for a work item
func (h *AgencyHandler) GetWorkItemGoalLinks(c *gin.Context) {
	agencyID := c.Param("id")
	workItemKey := c.Param("key")

	links, err := h.service.GetWorkItemGoalLinks(c.Request.Context(), agencyID, workItemKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, links)
}

// CreateWorkItemGoalLink handles POST /api/v1/agencies/:id/work-items/:key/goals
// Creates a new link between a work item and a goal
func (h *AgencyHandler) CreateWorkItemGoalLink(c *gin.Context) {
	agencyID := c.Param("id")
	workItemKey := c.Param("key")

	var req models.CreateWorkItemGoalLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Ensure work item key matches
	req.WorkItemKey = workItemKey

	link := &models.WorkItemGoalLink{
		WorkItemKey:  req.WorkItemKey,
		GoalKey:      req.GoalKey,
		Relationship: req.Relationship,
	}

	if link.Relationship == "" {
		link.Relationship = "addresses"
	}

	err := h.service.CreateWorkItemGoalLink(c.Request.Context(), agencyID, link)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, link)
}

// DeleteWorkItemGoalLink handles DELETE /api/v1/agencies/:id/work-items/:key/goals/:linkKey
// Deletes a specific work item-goal link
func (h *AgencyHandler) DeleteWorkItemGoalLink(c *gin.Context) {
	agencyID := c.Param("id")
	linkKey := c.Param("linkKey")

	err := h.service.DeleteWorkItemGoalLink(c.Request.Context(), agencyID, linkKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Link deleted successfully"})
}

// DeleteWorkItemGoalLinks handles DELETE /api/v1/agencies/:id/work-items/:key/goals
// Deletes all goal links for a work item
func (h *AgencyHandler) DeleteWorkItemGoalLinks(c *gin.Context) {
	agencyID := c.Param("id")
	workItemKey := c.Param("key")

	err := h.service.DeleteWorkItemGoalLinks(c.Request.Context(), agencyID, workItemKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All links deleted successfully"})
}
