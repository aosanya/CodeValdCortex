package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/gin-gonic/gin"
)

// GetAgencyRACIMatrix handles GET /api/v1/agencies/:id/raci-matrix
// Returns RACI matrix assignments for the agency
func (h *AgencyHandler) GetAgencyRACIMatrix(c *gin.Context) {
	agencyID := c.Param("id")

	// Fetch all RACI assignments from edge collection
	assignments, err := h.service.GetAllRACIAssignments(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("[RACI GET] Failed to fetch RACI assignments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch RACI assignments", "details": err.Error()})
		return
	}

	// Transform edge collection format to frontend format
	// Frontend expects: assignments[workItemKey][roleKey] = {raci: "R", objective: "..."}
	assignmentsMap := make(map[string]map[string]map[string]string)

	for _, assignment := range assignments {
		if assignmentsMap[assignment.WorkItemKey] == nil {
			assignmentsMap[assignment.WorkItemKey] = make(map[string]map[string]string)
		}
		assignmentsMap[assignment.WorkItemKey][assignment.RoleKey] = map[string]string{
			"raci":      assignment.RACI,
			"objective": assignment.Objective,
		}
	}

	response := gin.H{
		"agency_id":   agencyID,
		"assignments": assignmentsMap,
	}

	c.JSON(http.StatusOK, response)
}

// SaveAgencyRACIMatrix handles POST /api/v1/agencies/:id/raci-matrix
// Saves RACI matrix assignments for the agency
func (h *AgencyHandler) SaveAgencyRACIMatrix(c *gin.Context) {
	agencyID := c.Param("id")

	// Parse request body
	var req struct {
		Assignments map[string]map[string]struct {
			RACI      string `json:"raci"`
			Objective string `json:"objective"`
		} `json:"assignments"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("[RACI SAVE] Failed to parse request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// First, get existing assignments to determine which need create vs update
	existingAssignments, err := h.service.GetAllRACIAssignments(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Warn("[RACI SAVE] Failed to fetch existing assignments, will create new ones")
		existingAssignments = []*models.RACIAssignment{}
	}

	// Build map of existing assignments for quick lookup
	existingMap := make(map[string]*models.RACIAssignment)
	for _, assignment := range existingAssignments {
		key := assignment.WorkItemKey + ":" + assignment.RoleKey
		existingMap[key] = assignment
	}

	// Process new assignments
	savedCount := 0
	for workItemKey, roles := range req.Assignments {
		for roleKey, data := range roles {
			lookupKey := workItemKey + ":" + roleKey

			// Check if assignment exists
			if existing, found := existingMap[lookupKey]; found {
				// Update existing assignment
				existing.RACI = data.RACI
				existing.Objective = data.Objective
				err := h.service.UpdateRACIAssignment(c.Request.Context(), agencyID, existing.Key, existing)
				if err != nil {
					h.logger.WithError(err).Errorf("[RACI SAVE] Failed to update assignment %s", lookupKey)
					continue
				}
				delete(existingMap, lookupKey) // Remove from map so we know it was processed
			} else {
				// Create new assignment
				newAssignment := &models.RACIAssignment{
					From:        "work_items/" + workItemKey,
					To:          "roles/" + roleKey,
					WorkItemKey: workItemKey,
					RoleKey:     roleKey,
					RACI:        data.RACI,
					Objective:   data.Objective,
				}
				err := h.service.CreateRACIAssignment(c.Request.Context(), agencyID, newAssignment)
				if err != nil {
					h.logger.WithError(err).Errorf("[RACI SAVE] Failed to create assignment %s", lookupKey)
					continue
				}
			}
			savedCount++
		}
	}

	// Delete assignments that were not in the request (removed by user)
	for _, orphaned := range existingMap {
		err := h.service.DeleteRACIAssignment(c.Request.Context(), agencyID, orphaned.Key)
		if err != nil {
			h.logger.WithError(err).Warnf("[RACI SAVE] Failed to delete orphaned assignment %s", orphaned.Key)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "RACI matrix saved successfully",
		"count":   savedCount,
	})
}
