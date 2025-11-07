package handlers

import (
	"net/http"
	"sort"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
)

// GetGoals handles GET /api/v1/agencies/:id/goals
func (h *AgencyHandler) GetGoals(c *gin.Context) {
	id := c.Param("id")

	goals, err := h.service.GetGoals(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, goals)
}

// GetGoalsHTML handles GET /api/v1/agencies/:id/goals/html
// Returns rendered HTML fragment for HTMX/JavaScript rendering
func (h *AgencyHandler) GetGoalsHTML(c *gin.Context) {
	id := c.Param("id")

	goals, err := h.service.GetGoals(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading goals")
		return
	}

	// Sort goals by Code
	sort.Slice(goals, func(i, j int) bool {
		return goals[i].Code < goals[j].Code
	})

	// Render the goals list template
	component := agency_designer.GoalsList(goals)
	c.Header("Content-Type", "text/html")
	component.Render(c.Request.Context(), c.Writer)
}

// CreateGoal handles POST /api/v1/agencies/:id/goals
func (h *AgencyHandler) CreateGoal(c *gin.Context) {
	id := c.Param("id")

	var req models.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	goal, err := h.service.CreateGoal(c.Request.Context(), id, req.Code, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, goal)
}

// UpdateGoal handles PUT /api/v1/agencies/:id/goals/:goalKey
func (h *AgencyHandler) UpdateGoal(c *gin.Context) {
	id := c.Param("id")
	goalKey := c.Param("goalKey")

	var req models.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := h.service.UpdateGoal(c.Request.Context(), id, goalKey, req.Code, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Goal updated successfully"})
}

// DeleteGoal handles DELETE /api/v1/agencies/:id/goals/:goalKey
func (h *AgencyHandler) DeleteGoal(c *gin.Context) {
	id := c.Param("id")
	goalKey := c.Param("goalKey")

	if err := h.service.DeleteGoal(c.Request.Context(), id, goalKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Goal deleted successfully"})
}
