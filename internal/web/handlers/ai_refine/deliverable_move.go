package ai_refine

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// MoveDeliverableRequest represents the request to move a deliverable node
type MoveDeliverableRequest struct {
	SourceWorkItemCode string `json:"source_work_item_code" binding:"required"`
	TargetWorkItemCode string `json:"target_work_item_code" binding:"required"`
	NodeID             string `json:"node_id" binding:"required"`
	NodeName           string `json:"node_name"`
}

// MoveDeliverableResponse represents the response after moving a deliverable
type MoveDeliverableResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	SourceWorkItem string `json:"source_work_item"`
	TargetWorkItem string `json:"target_work_item"`
	MovedNodeName  string `json:"moved_node_name"`
	MovedNodeID    string `json:"moved_node_id"`
}

// MoveDeliverable handles POST /api/v1/agencies/:id/deliverables/move
// Moves a deliverable node from one work item to another
func (h *Handler) MoveDeliverable(c *gin.Context) {
	agencyID := c.Param("id")

	var req MoveDeliverableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Validate that source and target are different
	if req.SourceWorkItemCode == req.TargetWorkItemCode {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source and target work items must be different"})
		return
	}

	// Get the agency specification
	spec, err := h.agencyService.GetSpecification(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get specification")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get specification"})
		return
	}

	if spec == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency specification not found"})
		return
	}

	// Find source and target work items
	var sourceWorkItem, targetWorkItem *models.WorkItem
	for i := range spec.WorkItems {
		if spec.WorkItems[i].Code == req.SourceWorkItemCode {
			sourceWorkItem = &spec.WorkItems[i]
		}
		if spec.WorkItems[i].Code == req.TargetWorkItemCode {
			targetWorkItem = &spec.WorkItems[i]
		}
	}

	if sourceWorkItem == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Source work item '%s' not found", req.SourceWorkItemCode)})
		return
	}

	if targetWorkItem == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Target work item '%s' not found", req.TargetWorkItemCode)})
		return
	}

	// Find and remove the node from source work item
	movedNode, removed := h.removeNodeByID(&sourceWorkItem.DeliverablesStructured, req.NodeID)
	if !removed {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Deliverable node '%s' not found in source work item", req.NodeID)})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"node_id":          req.NodeID,
		"node_name":        movedNode.Name,
		"source_work_item": req.SourceWorkItemCode,
		"target_work_item": req.TargetWorkItemCode,
	}).Info("Moving deliverable node between work items")

	// Check for duplicate names in target work item (case-insensitive)
	if h.hasNodeWithName(&targetWorkItem.DeliverablesStructured, movedNode.Name) {
		// Restore node to source (since we removed it)
		sourceWorkItem.DeliverablesStructured = append(sourceWorkItem.DeliverablesStructured, *movedNode)

		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("A deliverable named '%s' already exists in target work item '%s'", movedNode.Name, req.TargetWorkItemCode),
		})
		return
	}

	// Add node to target work item (append to root level)
	targetWorkItem.DeliverablesStructured = append(targetWorkItem.DeliverablesStructured, *movedNode)

	// Recompute paths for both work items
	for i := range sourceWorkItem.DeliverablesStructured {
		sourceWorkItem.DeliverablesStructured[i].ComputeAllPaths("")
	}
	for i := range targetWorkItem.DeliverablesStructured {
		targetWorkItem.DeliverablesStructured[i].ComputeAllPaths("")
	}

	// Save the updated specification
	updatedWorkItems := make([]models.WorkItem, len(spec.WorkItems))
	copy(updatedWorkItems, spec.WorkItems)

	_, err = h.agencyService.UpdateSpecificationWorkItems(c.Request.Context(), agencyID, updatedWorkItems, "deliverable-move")
	if err != nil {
		h.logger.WithError(err).Error("Failed to save updated work items")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save changes"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"moved_node_id":   movedNode.ID,
		"moved_node_name": movedNode.Name,
		"from":            req.SourceWorkItemCode,
		"to":              req.TargetWorkItemCode,
	}).Info("Successfully moved deliverable node")

	c.JSON(http.StatusOK, MoveDeliverableResponse{
		Success:        true,
		Message:        fmt.Sprintf("Moved '%s' from %s to %s", movedNode.Name, req.SourceWorkItemCode, req.TargetWorkItemCode),
		SourceWorkItem: req.SourceWorkItemCode,
		TargetWorkItem: req.TargetWorkItemCode,
		MovedNodeName:  movedNode.Name,
		MovedNodeID:    movedNode.ID,
	})
}

// removeNodeByID recursively searches for and removes a node by ID
// Returns the removed node and a boolean indicating if removal was successful
func (h *Handler) removeNodeByID(nodes *[]models.DeliverableNode, nodeID string) (*models.DeliverableNode, bool) {
	for i := range *nodes {
		if (*nodes)[i].ID == nodeID {
			// Found the node - remove it
			removedNode := (*nodes)[i]
			*nodes = append((*nodes)[:i], (*nodes)[i+1:]...)
			return &removedNode, true
		}

		// Search in children
		if len((*nodes)[i].Children) > 0 {
			if removed, found := h.removeNodeByID(&(*nodes)[i].Children, nodeID); found {
				return removed, true
			}
		}
	}

	return nil, false
}

// hasNodeWithName checks if a node with the given name exists (case-insensitive)
func (h *Handler) hasNodeWithName(nodes *[]models.DeliverableNode, name string) bool {
	nameLower := strings.ToLower(name)

	for i := range *nodes {
		if strings.ToLower((*nodes)[i].Name) == nameLower {
			return true
		}

		// Search in children
		if len((*nodes)[i].Children) > 0 {
			if h.hasNodeWithName(&(*nodes)[i].Children, name) {
				return true
			}
		}
	}

	return false
}
