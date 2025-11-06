package ai_refine

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/gin-gonic/gin"
)

// RefineRoles is the main dynamic router for all role operations
// It analyzes the user's message to determine what action to take
// and handles role refinement, generation, consolidation, and enhancement
func (h *Handler) RefineRoles(c *gin.Context) {
	agencyID := c.Param("id")
	if agencyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agency ID is required"})
		return
	}

	var req builder.RefineRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the agency ID from the URL parameter
	req.AgencyID = agencyID

	// Build context for the role processing
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
	if req.ExistingRoles == nil {
		req.ExistingRoles = builderContext.Roles
	}
	if req.WorkItems == nil {
		req.WorkItems = builderContext.WorkItems
	}
	if req.AgencyContext == nil {
		req.AgencyContext = ag
	}

	// For now, return a placeholder response until RefineRoles is implemented
	response := &builder.RefineRolesResponse{
		Action:         "under_construction",
		Explanation:    "Role processing is under construction. The following operations will be supported: refine specific roles, generate new roles based on work items, consolidate duplicate roles, enhance all roles with AI analysis.",
		NoActionNeeded: false,
	}

	c.JSON(http.StatusOK, response)
}
