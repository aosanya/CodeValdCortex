package ai_refine

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/gin-gonic/gin"
)

// RefineRACIMappings is the main dynamic router for all RACI operations
// It analyzes the user's message to determine what action to take
// and handles RACI refinement, generation, consolidation, and creation
func (h *Handler) RefineRACIMappings(c *gin.Context) {
	agencyID := c.Param("id")
	if agencyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agency ID is required"})
		return
	}

	var req builder.RefineRACIMappingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the agency ID from the URL parameter
	req.AgencyID = agencyID

	// Build context for the RACI processing
	ctx := c.Request.Context()
	ag, err := h.agencyService.GetAgency(ctx, agencyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	// Build AI context data
	builderContext, err := h.contextBuilder.BuildBuilderContext(ctx, ag, "", req.UserMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build context"})
		return
	}

	// Set context data in the request if not provided
	if req.ExistingAssignments == nil {
		// RACI assignments are now in the specification.RACIMatrix, not separate edges
		req.ExistingAssignments = []*models.RACIAssignment{}
	}
	if req.WorkItems == nil {
		req.WorkItems = builderContext.WorkItems
	}
	if req.Roles == nil {
		req.Roles = builderContext.Roles
	}
	if req.AgencyContext == nil {
		req.AgencyContext = ag
	}

	// For now, return a placeholder response until RefineRACIMappings is implemented
	response := &builder.RefineRACIMappingsResponse{
		Action:         "under_construction",
		Explanation:    "RACI processing is under construction. The following operations will be supported: refine specific RACI assignments, generate new assignments based on work items and roles, consolidate duplicate assignments, create complete RACI matrix.",
		NoActionNeeded: false,
	}

	c.JSON(http.StatusOK, response)
}
