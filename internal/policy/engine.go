package policy

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Engine evaluates policy compliance for AI operations
type Engine struct {
	repo Repository
}

// NewEngine creates a new policy evaluation engine
func NewEngine(repo Repository) *Engine {
	return &Engine{
		repo: repo,
	}
}

// EvaluateRequest represents a policy evaluation request
type EvaluateRequest struct {
	AgencyID string
	AgentID  string
	RoleID   string
	Action   string
	Resource string
	Context  map[string]interface{}

	// AI-specific fields
	ModelProvider       string
	ModelName           string
	TokensRequested     int
	EstimatedCostUSD    float64
	DataClassifications []string
	AutonomyLevel       string
}

// Evaluate evaluates a request against the agency's AI policy
func (e *Engine) Evaluate(ctx context.Context, req *EvaluateRequest) (*PolicyResult, error) {
	// Get agency policy
	policy, err := e.repo.Get(ctx, req.AgencyID)
	if err != nil {
		// No policy found - allow with warning
		result := &PolicyResult{
			Allowed:       true,
			Reason:        "No AI policy defined - defaulting to allow",
			RiskScore:     50,
			RiskFactors:   []string{"no_policy_defined"},
			Severity:      SeverityMedium,
			PolicyVersion: "none",
			EvaluatedAt:   time.Now(),
		}
		return result, nil
	}

	result := &PolicyResult{
		Allowed:            true,
		PolicyVersion:      policy.Version,
		EvaluatedAt:        time.Now(),
		RiskScore:          0,
		RiskFactors:        []string{},
		ComplianceControls: []string{},
		Conditions:         []MonitoringCondition{},
	}

	// Evaluate each policy dimension
	e.evaluateModelPolicy(policy, req, result)
	e.evaluateAutonomyPolicy(policy, req, result)
	e.evaluateDataAccessPolicy(policy, req, result)
	e.evaluateActionPolicy(policy, req, result)
	e.evaluateRiskPolicy(policy, req, result)
	e.evaluateCompliancePolicy(policy, req, result)
	e.evaluateMonitoringPolicy(policy, req, result)

	// Determine overall severity based on risk score
	result.Severity = e.calculateSeverity(result.RiskScore)

	// Log the evaluation
	evaluation := &PolicyEvaluation{
		AgencyID:      req.AgencyID,
		PolicyVersion: policy.Version,
		AgentID:       req.AgentID,
		RoleID:        req.RoleID,
		Action:        req.Action,
		Resource:      req.Resource,
		Result:        *result,
		EvaluatedAt:   result.EvaluatedAt,
	}

	if err := e.repo.LogEvaluation(ctx, evaluation); err != nil {
		// Don't fail evaluation if logging fails
		fmt.Printf("Warning: failed to log evaluation: %v\n", err)
	}

	// Log violation if request was blocked
	if !result.Allowed {
		violation := &PolicyViolation{
			AgencyID:      req.AgencyID,
			PolicyVersion: policy.Version,
			AgentID:       req.AgentID,
			RoleID:        req.RoleID,
			Action:        req.Action,
			Resource:      req.Resource,
			ViolationType: e.determineViolationType(result),
			Severity:      result.Severity,
			Reason:        result.Reason,
			Blocked:       true,
		}

		if err := e.repo.LogViolation(ctx, violation); err != nil {
			fmt.Printf("Warning: failed to log violation: %v\n", err)
		}
	}

	return result, nil
}

// evaluateModelPolicy checks if the requested model is allowed
func (e *Engine) evaluateModelPolicy(policy *AIPolicy, req *EvaluateRequest, result *PolicyResult) {
	if req.ModelProvider == "" || req.ModelName == "" {
		return // No model specified, skip evaluation
	}

	// Check denied providers first
	for _, denied := range policy.Models.DeniedProviders {
		if strings.EqualFold(denied, req.ModelProvider) {
			result.Allowed = false
			result.Reason = fmt.Sprintf("Provider %s is explicitly denied", req.ModelProvider)
			result.RiskScore += 50
			result.RiskFactors = append(result.RiskFactors, "denied_provider")
			return
		}
	}

	// Check allowed providers
	providerAllowed := false
	modelAllowed := false
	var providerConfig *AllowedProvider

	for i, provider := range policy.Models.AllowedProviders {
		if strings.EqualFold(provider.Provider, req.ModelProvider) {
			providerAllowed = true
			providerConfig = &policy.Models.AllowedProviders[i]

			// Check if specific model is allowed
			for _, model := range provider.Models {
				if strings.EqualFold(model, req.ModelName) || model == "*" {
					modelAllowed = true
					break
				}
			}
			break
		}
	}

	if !providerAllowed {
		result.Allowed = false
		result.Reason = fmt.Sprintf("Provider %s is not in allowed list", req.ModelProvider)
		result.RiskScore += 40
		result.RiskFactors = append(result.RiskFactors, "unauthorized_provider")
		return
	}

	if !modelAllowed {
		result.Allowed = false
		result.Reason = fmt.Sprintf("Model %s is not allowed for provider %s", req.ModelName, req.ModelProvider)
		result.RiskScore += 35
		result.RiskFactors = append(result.RiskFactors, "unauthorized_model")
		return
	}

	// Check token limits
	if providerConfig != nil && providerConfig.MaxTokensPerRequest > 0 {
		if req.TokensRequested > providerConfig.MaxTokensPerRequest {
			result.Allowed = false
			result.Reason = fmt.Sprintf("Token request (%d) exceeds limit (%d)", req.TokensRequested, providerConfig.MaxTokensPerRequest)
			result.RiskScore += 25
			result.RiskFactors = append(result.RiskFactors, "token_limit_exceeded")
			return
		}
	}

	// Check budget
	if providerConfig != nil && providerConfig.MonthlyBudgetUSD > 0 {
		projectedSpend := providerConfig.CurrentSpendUSD + req.EstimatedCostUSD
		if projectedSpend > providerConfig.MonthlyBudgetUSD {
			result.Allowed = false
			result.Reason = fmt.Sprintf("Request would exceed monthly budget ($%.2f + $%.2f > $%.2f)",
				providerConfig.CurrentSpendUSD, req.EstimatedCostUSD, providerConfig.MonthlyBudgetUSD)
			result.RiskScore += 30
			result.RiskFactors = append(result.RiskFactors, "budget_exceeded")
			return
		}

		// Warn if approaching budget (>80%)
		if projectedSpend > (providerConfig.MonthlyBudgetUSD * 0.8) {
			result.RiskScore += 10
			result.RiskFactors = append(result.RiskFactors, "budget_warning")
			result.EnhancedMonitoring = true
			result.Conditions = append(result.Conditions, MonitoringCondition{
				Type:      "budget_threshold",
				Threshold: 0.8,
				Metadata: map[string]interface{}{
					"current_spend": providerConfig.CurrentSpendUSD,
					"budget":        providerConfig.MonthlyBudgetUSD,
					"projected":     projectedSpend,
				},
			})
		}
	}
}

// evaluateAutonomyPolicy checks if the autonomy level is appropriate
func (e *Engine) evaluateAutonomyPolicy(policy *AIPolicy, req *EvaluateRequest, result *PolicyResult) {
	if req.AutonomyLevel == "" {
		req.AutonomyLevel = policy.Autonomy.DefaultLevel
	}

	// Check role overrides
	for _, override := range policy.Autonomy.RoleOverrides {
		if override.RoleID == req.RoleID {
			if req.AutonomyLevel != override.Level {
				result.RequiresEscalation = true
				result.EscalateToLevel = override.Level
				result.RiskScore += 15
				result.RiskFactors = append(result.RiskFactors, "autonomy_override_required")
			}

			// Check if approval required
			if len(override.RequiresApprovalFrom) > 0 {
				result.RequiresApproval = true
				result.Approvers = override.RequiresApprovalFrom
				result.AuditRequired = true
			}
			return
		}
	}

	// Check escalation rules
	for _, rule := range policy.Autonomy.EscalationRules {
		if e.matchesTrigger(rule.Trigger, req) {
			result.RequiresEscalation = true
			result.EscalateToLevel = rule.EscalateToLevel
			result.RiskScore += 20

			if rule.ApprovalRequired {
				result.RequiresApproval = true
				result.RiskFactors = append(result.RiskFactors, "escalation_requires_approval")
			}

			if rule.AuditRequired {
				result.AuditRequired = true
			}
		}
	}
}

// evaluateDataAccessPolicy checks data access permissions
func (e *Engine) evaluateDataAccessPolicy(policy *AIPolicy, req *EvaluateRequest, result *PolicyResult) {
	if len(req.DataClassifications) == 0 {
		return // No data access requested
	}

	for _, classification := range req.DataClassifications {
		// Find matching rule
		var rule *DataClassificationRule
		for i, r := range policy.DataAccess.Rules {
			if strings.EqualFold(r.Classification, classification) {
				rule = &policy.DataAccess.Rules[i]
				break
			}
		}

		if rule == nil {
			result.Allowed = false
			result.Reason = fmt.Sprintf("No policy rule for data classification: %s", classification)
			result.RiskScore += 40
			result.RiskFactors = append(result.RiskFactors, "unclassified_data_access")
			continue
		}

		// Check if operation is allowed
		operationAllowed := false
		for _, op := range rule.AllowedOperations {
			if strings.EqualFold(op, req.Action) {
				operationAllowed = true
				break
			}
		}

		if !operationAllowed {
			result.Allowed = false
			result.Reason = fmt.Sprintf("Action %s not allowed for %s data", req.Action, classification)
			result.RiskScore += 35
			result.RiskFactors = append(result.RiskFactors, "unauthorized_data_operation")
			continue
		}

		// Apply additional requirements
		if rule.RequiresJustification {
			result.RiskScore += 10
			result.RiskFactors = append(result.RiskFactors, "justification_required")
		}

		if rule.RequiresApproval {
			result.RequiresApproval = true
			result.RiskScore += 15
		}

		if rule.AuditAllAccess {
			result.AuditRequired = true
		}

		if rule.ExplicitGrantRequired {
			result.RequiresApproval = true
			result.RiskScore += 20
			result.RiskFactors = append(result.RiskFactors, "explicit_grant_required")
		}

		if rule.DualApproval {
			result.RequiresApproval = true
			result.RiskScore += 25
			result.RiskFactors = append(result.RiskFactors, "dual_approval_required")
		}

		// Add risk based on classification
		switch strings.ToLower(classification) {
		case DataClassRestricted:
			result.RiskScore += 30
		case DataClassConfidential:
			result.RiskScore += 20
		case DataClassInternal:
			result.RiskScore += 10
		}
	}

	// Check PII handling
	if policy.DataAccess.PIIHandling.DetectPII {
		result.EnhancedMonitoring = true
		result.Conditions = append(result.Conditions, MonitoringCondition{
			Type: "pii_detection",
		})
	}
}

// evaluateActionPolicy checks action approval requirements
func (e *Engine) evaluateActionPolicy(policy *AIPolicy, req *EvaluateRequest, result *PolicyResult) {
	// Check prohibited actions
	for _, prohibited := range policy.Actions.ProhibitedActions {
		if matchesPattern(prohibited, req.Action) {
			result.Allowed = false
			result.Reason = fmt.Sprintf("Action %s is prohibited", req.Action)
			result.RiskScore += 50
			result.RiskFactors = append(result.RiskFactors, "prohibited_action")
			return
		}
	}

	// Check approval workflows
	for _, workflow := range policy.Actions.ApprovalWorkflows {
		if matchesPattern(workflow.ActionPattern, req.Action) {
			if workflow.RequiresApproval {
				result.RequiresApproval = true
				result.Approvers = workflow.Approvers
				result.ApprovalTimeout = workflow.TimeoutMinutes
				result.RiskScore += 15

				if workflow.DualApproval {
					result.RiskFactors = append(result.RiskFactors, "dual_approval_required")
					result.RiskScore += 10
				}
			}
		}
	}

	// Check rollback requirements
	for _, rollback := range policy.Actions.RollbackRequirements {
		if matchesPattern(rollback.ActionPattern, req.Action) {
			if rollback.RollbackPlanRequired {
				result.RiskScore += 10
				result.RiskFactors = append(result.RiskFactors, "rollback_plan_required")
			}
		}
	}
}

// evaluateRiskPolicy calculates overall risk score
func (e *Engine) evaluateRiskPolicy(policy *AIPolicy, req *EvaluateRequest, result *PolicyResult) {
	if !policy.Risk.ScoringEnabled {
		return
	}

	// Risk score already accumulated from other evaluations
	// Now check thresholds and apply mitigations

	var riskLevel string
	if result.RiskScore >= policy.Risk.Thresholds.Critical {
		riskLevel = SeverityCritical
		result.RequiresRiskReview = true
	} else if result.RiskScore >= policy.Risk.Thresholds.High {
		riskLevel = SeverityHigh
		result.RequiresRiskReview = true
	} else if result.RiskScore >= policy.Risk.Thresholds.Medium {
		riskLevel = SeverityMedium
	} else {
		riskLevel = SeverityInfo
	}

	// Apply mitigations
	if mitigations, exists := policy.Risk.Mitigation[riskLevel]; exists {
		for _, mitigation := range mitigations {
			switch mitigation {
			case "require_approval":
				result.RequiresApproval = true
			case "enhanced_monitoring":
				result.EnhancedMonitoring = true
			case "audit_required":
				result.AuditRequired = true
			case "escalate":
				result.RequiresEscalation = true
			}
		}
	}
}

// evaluateCompliancePolicy checks compliance requirements
func (e *Engine) evaluateCompliancePolicy(policy *AIPolicy, req *EvaluateRequest, result *PolicyResult) {
	// Note: Compliance frameworks (SOC2, GDPR, HIPAA, ISO27001) will be implemented
	// as intelligent agents in MVP-051-053. For now, we only enforce audit requirements.

	// Check audit requirements
	if policy.Compliance.AuditRequirements.LogAllActions {
		result.AuditRequired = true
	}

	// If immutable audit log is required, increase monitoring
	if policy.Compliance.AuditRequirements.ImmutableAuditLog {
		result.EnhancedMonitoring = true
	}
}

// evaluateMonitoringPolicy sets up monitoring requirements
func (e *Engine) evaluateMonitoringPolicy(policy *AIPolicy, req *EvaluateRequest, result *PolicyResult) {
	if !policy.Monitoring.RealTimePolicyViolations {
		return
	}

	// Check alert conditions
	for _, alert := range policy.Monitoring.Alerts {
		if e.matchesAlertCondition(alert, req, result) {
			result.EnhancedMonitoring = true
			result.Conditions = append(result.Conditions, MonitoringCondition{
				Type:      alert.Condition,
				Threshold: alert.Threshold,
				Metadata: map[string]interface{}{
					"severity": alert.Severity,
					"notify":   alert.NotifyTo,
				},
			})
		}
	}
}

// Helper functions

func (e *Engine) matchesTrigger(trigger string, req *EvaluateRequest) bool {
	// Simple trigger matching
	switch trigger {
	case "high_cost_action":
		return req.EstimatedCostUSD > 10.0
	case "sensitive_data_access":
		for _, class := range req.DataClassifications {
			if strings.EqualFold(class, DataClassRestricted) || strings.EqualFold(class, DataClassConfidential) {
				return true
			}
		}
	case "model_change":
		return strings.Contains(req.Action, "model") || strings.Contains(req.Action, "deploy")
	}
	return false
}

func (e *Engine) matchesAlertCondition(alert MonitoringAlert, req *EvaluateRequest, result *PolicyResult) bool {
	switch alert.Condition {
	case "policy_violation":
		return !result.Allowed
	case "budget_threshold":
		return req.EstimatedCostUSD > alert.Threshold
	case "high_risk_score":
		return float64(result.RiskScore) >= alert.Threshold
	}
	return false
}

func (e *Engine) calculateSeverity(riskScore int) string {
	if riskScore >= 75 {
		return SeverityCritical
	} else if riskScore >= 50 {
		return SeverityHigh
	} else if riskScore >= 25 {
		return SeverityMedium
	}
	return SeverityInfo
}

func (e *Engine) determineViolationType(result *PolicyResult) string {
	if len(result.RiskFactors) > 0 {
		// Return the most severe risk factor
		for _, factor := range result.RiskFactors {
			if strings.Contains(factor, "denied") || strings.Contains(factor, "prohibited") {
				return factor
			}
		}
		return result.RiskFactors[0]
	}
	return "policy_violation"
}

func matchesPattern(pattern, value string) bool {
	// Simple glob pattern matching
	// For MVP: exact match or wildcard (*)
	if pattern == "*" {
		return true
	}
	if strings.Contains(pattern, "*") {
		// Simple prefix/suffix matching
		if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
			return strings.Contains(value, strings.Trim(pattern, "*"))
		}
		if strings.HasPrefix(pattern, "*") {
			return strings.HasSuffix(value, strings.TrimPrefix(pattern, "*"))
		}
		if strings.HasSuffix(pattern, "*") {
			return strings.HasPrefix(value, strings.TrimSuffix(pattern, "*"))
		}
	}
	return strings.EqualFold(pattern, value)
}
