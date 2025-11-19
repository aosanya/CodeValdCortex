package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
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

	// Try to load existing policy from specification
	var existingPolicy *policy.AIPolicy
	spec, err := h.agencyService.GetSpecification(c.Request.Context(), agencyID)
	if err == nil && spec.AIPolicy != nil {
		// Convert models.Policy to policy.AIPolicy for the wizard
		existingPolicy, err = convertModelPolicyToAIPolicy(spec.AIPolicy)
		if err != nil {
			h.logger.WithError(err).Debug("Failed to convert policy, will create new")
			existingPolicy = nil
		}
	}

	// Render wizard template
	component := agency_designer.AIPolicyWizard(currentAgency, existingPolicy)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render policy wizard")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render policy wizard"})
		return
	}
}

// GetPolicy returns the AI policy for an agency from the specification
func (h *AIPolicyWebHandler) GetPolicy(c *gin.Context) {
	agencyID := c.Param("id")

	// Get specification which contains the policy
	spec, err := h.agencyService.GetSpecification(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Debug("Failed to get specification")
		c.JSON(http.StatusNotFound, gin.H{"error": "Specification not found"})
		return
	}

	if spec.AIPolicy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}

	c.JSON(http.StatusOK, spec.AIPolicy)
}

// GetPolicySummary returns a summary of the policy for display
func (h *AIPolicyWebHandler) GetPolicySummary(c *gin.Context) {
	agencyID := c.Param("id")

	// Get specification which contains the policy
	spec, err := h.agencyService.GetSpecification(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Debug("Failed to get specification")
		c.JSON(http.StatusNotFound, gin.H{"error": "Specification not found"})
		return
	}

	if spec.AIPolicy == nil {
		h.logger.Debug("No policy configured yet")
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}

	// Build summary directly from models.Policy
	summary := buildPolicySummaryFromModelPolicy(spec.AIPolicy)

	// Add specification metadata
	summary["last_updated"] = spec.UpdatedAt
	summary["agency_id"] = spec.AgencyID

	c.JSON(http.StatusOK, summary)
}

// SavePolicy saves or updates an AI policy in the agency specification
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

	// Validate the policy
	if err := h.policyService.ValidatePolicy(&aiPolicy); err != nil {
		h.logger.WithError(err).Error("Policy validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid policy: %s", err.Error())})
		return
	}

	// Convert policy.AIPolicy to models.Policy for storage
	policyData, err := convertAIPolicyToModelPolicy(&aiPolicy)
	if err != nil {
		h.logger.WithError(err).Error("Failed to convert policy")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert policy"})
		return
	}

	// Create specification update request with AI policy
	updateReq := &models.SpecificationUpdateRequest{
		AIPolicy: policyData,
	}

	// Save the updated specification
	_, err = h.agencyService.UpdateSpecification(c.Request.Context(), agencyID, updateReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update specification with policy")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Policy saved successfully", "policy": aiPolicy})
}

// convertAIPolicyToModelPolicy converts a policy.AIPolicy to models.Policy for storage
func convertAIPolicyToModelPolicy(aiPolicy *policy.AIPolicy) (*models.Policy, error) {
	// Marshal to JSON and unmarshal to map
	data, err := json.Marshal(aiPolicy)
	if err != nil {
		return nil, err
	}

	var policyMap map[string]interface{}
	if err := json.Unmarshal(data, &policyMap); err != nil {
		return nil, err
	}

	// Extract sections
	modelPolicy := &models.Policy{
		Version: aiPolicy.Version,
		Owner:   aiPolicy.Owner,
	}

	if stance, ok := policyMap["stance"].(map[string]interface{}); ok {
		modelPolicy.Stance = stance
	}
	if models, ok := policyMap["models"].(map[string]interface{}); ok {
		modelPolicy.Models = models
	}
	if autonomy, ok := policyMap["autonomy"].(map[string]interface{}); ok {
		modelPolicy.Autonomy = autonomy
	}
	if dataAccess, ok := policyMap["data_access"].(map[string]interface{}); ok {
		modelPolicy.DataAccess = dataAccess
	}
	if actions, ok := policyMap["actions"].(map[string]interface{}); ok {
		modelPolicy.Actions = actions
	}
	if risk, ok := policyMap["risk"].(map[string]interface{}); ok {
		modelPolicy.Risk = risk
	}
	if compliance, ok := policyMap["compliance"].(map[string]interface{}); ok {
		modelPolicy.Compliance = compliance
	}
	if monitoring, ok := policyMap["monitoring"].(map[string]interface{}); ok {
		modelPolicy.Monitoring = monitoring
	}

	return modelPolicy, nil
}

// convertModelPolicyToAIPolicy converts a models.Policy back to policy.AIPolicy
func convertModelPolicyToAIPolicy(modelPolicy *models.Policy) (*policy.AIPolicy, error) {
	// Marshal the models.Policy to JSON and unmarshal to policy.AIPolicy
	data, err := json.Marshal(modelPolicy)
	if err != nil {
		return nil, err
	}

	var aiPolicy policy.AIPolicy
	if err := json.Unmarshal(data, &aiPolicy); err != nil {
		return nil, err
	}

	return &aiPolicy, nil
}

// buildPolicySummaryFromModelPolicy creates a summary from models.Policy
func buildPolicySummaryFromModelPolicy(modelPolicy *models.Policy) map[string]interface{} {
	summary := make(map[string]interface{})

	summary["version"] = modelPolicy.Version
	summary["owner"] = modelPolicy.Owner

	// Extract adoption level and risk tolerance from stance
	if modelPolicy.Stance != nil {
		if adoptionLevel, ok := modelPolicy.Stance["adoption_level"]; ok {
			summary["adoption_level"] = adoptionLevel
		}
		if riskTolerance, ok := modelPolicy.Stance["risk_tolerance"]; ok {
			summary["risk_tolerance"] = riskTolerance
		}
	}

	// Extract autonomy info
	if modelPolicy.Autonomy != nil {
		if defaultLevel, ok := modelPolicy.Autonomy["default_level"]; ok {
			summary["default_autonomy_level"] = defaultLevel
		}
		if roleOverrides, ok := modelPolicy.Autonomy["role_overrides"].([]interface{}); ok {
			summary["role_overrides_count"] = len(roleOverrides)
		} else {
			summary["role_overrides_count"] = 0
		}
	}

	// Extract allowed providers and calculate totals
	allowedProviders := []string{}
	totalModels := 0
	totalBudget := 0.0
	currentSpend := 0.0

	if modelPolicy.Models != nil {
		if providers, ok := modelPolicy.Models["allowed_providers"].([]interface{}); ok {
			for _, p := range providers {
				if providerMap, ok := p.(map[string]interface{}); ok {
					if provider, ok := providerMap["provider"].(string); ok {
						allowedProviders = append(allowedProviders, provider)
					}
					if models, ok := providerMap["models"].([]interface{}); ok {
						totalModels += len(models)
					}
					if budget, ok := providerMap["monthly_budget_usd"].(float64); ok {
						totalBudget += budget
					}
					if spend, ok := providerMap["current_spend_usd"].(float64); ok {
						currentSpend += spend
					}
				}
			}
		}
	}

	summary["allowed_providers"] = allowedProviders
	summary["total_models"] = totalModels
	summary["total_budget_usd"] = totalBudget
	summary["current_spend_usd"] = currentSpend

	// Extract monitoring and auditing status
	monitoringEnabled := false
	auditingEnabled := false

	if modelPolicy.Monitoring != nil {
		if enabled, ok := modelPolicy.Monitoring["real_time_policy_violations"].(bool); ok {
			monitoringEnabled = enabled
		}
	}

	if modelPolicy.Compliance != nil {
		if auditReqs, ok := modelPolicy.Compliance["audit_requirements"].(map[string]interface{}); ok {
			if logAll, ok := auditReqs["log_all_actions"].(bool); ok {
				auditingEnabled = logAll
			}
		}
	}

	summary["monitoring_enabled"] = monitoringEnabled
	summary["auditing_enabled"] = auditingEnabled

	return summary
}

// DeletePolicy deletes an AI policy from the specification
func (h *AIPolicyWebHandler) DeletePolicy(c *gin.Context) {
	agencyID := c.Param("id")

	// Get current specification
	spec, err := h.agencyService.GetSpecification(c.Request.Context(), agencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get specification")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get specification"})
		return
	}

	// Remove the policy
	spec.AIPolicy = nil

	// Update specification
	updateReq := &models.SpecificationUpdateRequest{
		AIPolicy: nil,
	}
	_, err = h.agencyService.UpdateSpecification(c.Request.Context(), agencyID, updateReq)
	if err != nil {
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
