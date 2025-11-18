package handlers

import (
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/policy"
	"github.com/aosanya/CodeValdCortex/internal/web/pages/agency_designer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AIPolicyWebHandler handles web requests for AI policy management
type AIPolicyWebHandler struct {
	policyService *policy.Service
	agencyService agency.Service
	logger        *logrus.Logger
}

// NewAIPolicyWebHandler creates a new AI policy web handler
func NewAIPolicyWebHandler(policyService *policy.Service, agencyService agency.Service, logger *logrus.Logger) *AIPolicyWebHandler {
	return &AIPolicyWebHandler{
		policyService: policyService,
		agencyService: agencyService,
		logger:        logger,
	}
}

// ShowPolicyWizard renders the AI policy wizard page
func (h *AIPolicyWebHandler) ShowPolicyWizard(c *gin.Context) {
	agencyID := c.Param("id")

	// Fetch the agency
	currentAgency, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get agency")
		c.JSON(http.StatusNotFound, gin.H{"error": "Agency not found"})
		return
	}

	// Try to load existing policy
	existingPolicy, err := h.policyService.GetPolicy(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Debug("No existing policy found, will create new")
	}

	// Render wizard template
	component := agency_designer.AIPolicyWizard(currentAgency, existingPolicy)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render policy wizard")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render policy wizard"})
		return
	}
}

// GetPolicy returns the AI policy for an agency
func (h *AIPolicyWebHandler) GetPolicy(c *gin.Context) {
	agencyID := c.Param("id")

	existingPolicy, err := h.policyService.GetPolicy(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Debug("No policy found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}

	c.JSON(http.StatusOK, existingPolicy)
}

// GetPolicySummary returns a summary of the policy for display
func (h *AIPolicyWebHandler) GetPolicySummary(c *gin.Context) {
	agencyID := c.Param("id")

	summary, err := h.policyService.GetPolicySummary(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Debug("No policy summary available")
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// SavePolicy saves or updates an AI policy
func (h *AIPolicyWebHandler) SavePolicy(c *gin.Context) {
	agencyID := c.Param("id")

	var aiPolicy policy.AIPolicy
	if err := c.ShouldBindJSON(&aiPolicy); err != nil {
		h.logger.WithError(err).Error("Failed to parse policy JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy format"})
		return
	}

	// Ensure agency ID matches
	aiPolicy.AgencyID = agencyID

	// Check if policy already exists
	_, err := h.policyService.GetPolicy(c.Request.Context(), agencyID)
	if err != nil {
		// Create new policy
		if err := h.policyService.CreatePolicy(c.Request.Context(), &aiPolicy); err != nil {
			h.logger.WithError(err).Error("Failed to create policy")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create policy"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Policy created successfully", "policy": aiPolicy})
	} else {
		// Update existing policy
		if err := h.policyService.UpdatePolicy(c.Request.Context(), &aiPolicy); err != nil {
			h.logger.WithError(err).Error("Failed to update policy")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update policy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Policy updated successfully", "policy": aiPolicy})
	}
}

// DeletePolicy deletes an AI policy
func (h *AIPolicyWebHandler) DeletePolicy(c *gin.Context) {
	agencyID := c.Param("id")

	if err := h.policyService.DeletePolicy(c.Request.Context(), agencyID); err != nil {
		h.logger.WithError(err).Error("Failed to delete policy")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Policy deleted successfully"})
}

// EvaluatePolicy evaluates a request against the agency's policy
func (h *AIPolicyWebHandler) EvaluatePolicy(c *gin.Context) {
	agencyID := c.Param("id")

	var req policy.EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse evaluation request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Ensure agency ID matches
	req.AgencyID = agencyID

	result, err := h.policyService.EvaluatePolicy(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to evaluate policy")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to evaluate policy"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetViolations returns recent policy violations
func (h *AIPolicyWebHandler) GetViolations(c *gin.Context) {
	agencyID := c.Param("id")
	limit := 100 // Default limit

	violations, err := h.policyService.GetViolations(c.Request.Context(), agencyID, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get violations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get violations"})
		return
	}

	c.JSON(http.StatusOK, violations)
}

// GetEvaluations returns recent policy evaluations
func (h *AIPolicyWebHandler) GetEvaluations(c *gin.Context) {
	agencyID := c.Param("id")
	limit := 100 // Default limit

	evaluations, err := h.policyService.GetEvaluations(c.Request.Context(), agencyID, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get evaluations")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get evaluations"})
		return
	}

	c.JSON(http.StatusOK, evaluations)
}

// RegisterRoutes registers the AI policy routes
func (h *AIPolicyWebHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Policy wizard UI
	router.GET("/agencies/:id/policy/wizard", h.ShowPolicyWizard)

	// Policy API endpoints
	router.GET("/api/agencies/:id/policy", h.GetPolicy)
	router.GET("/api/agencies/:id/policy/summary", h.GetPolicySummary)
	router.POST("/api/agencies/:id/policy", h.SavePolicy)
	router.DELETE("/api/agencies/:id/policy", h.DeletePolicy)

	// Policy evaluation and monitoring
	router.POST("/api/agencies/:id/policy/evaluate", h.EvaluatePolicy)
	router.GET("/api/agencies/:id/policy/violations", h.GetViolations)
	router.GET("/api/agencies/:id/policy/evaluations", h.GetEvaluations)
}
