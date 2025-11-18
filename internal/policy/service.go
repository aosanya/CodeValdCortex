package policy

import (
	"context"
	"fmt"
	"time"
)

// Service provides high-level policy management operations
type Service struct {
	repo   Repository
	engine *Engine
}

// NewService creates a new policy service
func NewService(repo Repository) *Service {
	return &Service{
		repo:   repo,
		engine: NewEngine(repo),
	}
}

// CreatePolicy creates a new AI policy for an agency
func (s *Service) CreatePolicy(ctx context.Context, policy *AIPolicy) error {
	if err := s.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	return s.repo.Create(ctx, policy)
}

// GetPolicy retrieves an AI policy by agency ID
func (s *Service) GetPolicy(ctx context.Context, agencyID string) (*AIPolicy, error) {
	return s.repo.Get(ctx, agencyID)
}

// UpdatePolicy updates an existing AI policy
func (s *Service) UpdatePolicy(ctx context.Context, policy *AIPolicy) error {
	if err := s.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	return s.repo.Update(ctx, policy)
}

// DeletePolicy deletes an AI policy
func (s *Service) DeletePolicy(ctx context.Context, agencyID string) error {
	return s.repo.Delete(ctx, agencyID)
}

// ListPolicies lists all AI policies
func (s *Service) ListPolicies(ctx context.Context) ([]*AIPolicy, error) {
	return s.repo.ListAll(ctx)
}

// EvaluatePolicy evaluates a request against the agency's policy
func (s *Service) EvaluatePolicy(ctx context.Context, req *EvaluateRequest) (*PolicyResult, error) {
	return s.engine.Evaluate(ctx, req)
}

// GetViolations retrieves policy violations for an agency
func (s *Service) GetViolations(ctx context.Context, agencyID string, limit int) ([]*PolicyViolation, error) {
	return s.repo.GetViolations(ctx, agencyID, limit)
}

// GetEvaluations retrieves policy evaluations for an agency
func (s *Service) GetEvaluations(ctx context.Context, agencyID string, limit int) ([]*PolicyEvaluation, error) {
	return s.repo.GetEvaluations(ctx, agencyID, limit)
}

// GetPolicySummary returns a summary of the policy for display
func (s *Service) GetPolicySummary(ctx context.Context, agencyID string) (*PolicySummary, error) {
	policy, err := s.repo.Get(ctx, agencyID)
	if err != nil {
		return nil, err
	}

	return s.buildPolicySummary(policy), nil
}

// validatePolicy validates policy configuration
func (s *Service) validatePolicy(policy *AIPolicy) error {
	if policy.AgencyID == "" {
		return fmt.Errorf("agency_id is required")
	}

	if policy.Owner == "" {
		return fmt.Errorf("owner is required")
	}

	// Validate stance
	if !isValidAdoptionLevel(policy.Stance.AdoptionLevel) {
		return fmt.Errorf("invalid adoption level: %s", policy.Stance.AdoptionLevel)
	}

	if !isValidRiskTolerance(policy.Stance.RiskTolerance) {
		return fmt.Errorf("invalid risk tolerance: %s", policy.Stance.RiskTolerance)
	}

	// Validate autonomy levels
	if !isValidAutonomyLevel(policy.Autonomy.DefaultLevel) {
		return fmt.Errorf("invalid default autonomy level: %s", policy.Autonomy.DefaultLevel)
	}

	// Filter out empty role overrides
	validOverrides := []RoleOverride{}
	for _, override := range policy.Autonomy.RoleOverrides {
		if override.RoleName != "" {
			if !isValidAutonomyLevel(override.Level) {
				return fmt.Errorf("invalid autonomy level in override: %s", override.Level)
			}
			validOverrides = append(validOverrides, override)
		}
	}
	policy.Autonomy.RoleOverrides = validOverrides

	// Validate model providers - filter out empty providers first
	validProviders := []AllowedProvider{}
	for _, provider := range policy.Models.AllowedProviders {
		if provider.Provider != "" {
			validProviders = append(validProviders, provider)
		}
	}
	
	if len(validProviders) == 0 {
		return fmt.Errorf("at least one allowed provider is required")
	}
	
	// Update policy with only valid providers
	policy.Models.AllowedProviders = validProviders

	for _, provider := range policy.Models.AllowedProviders {
		if len(provider.Models) == 0 {
			return fmt.Errorf("at least one model is required for provider %s", provider.Provider)
		}
		if provider.MonthlyBudgetUSD < 0 {
			return fmt.Errorf("monthly budget cannot be negative for provider %s", provider.Provider)
		}
	}

	// Validate data access rules
	for _, rule := range policy.DataAccess.Rules {
		if !isValidDataClassification(rule.Classification) {
			return fmt.Errorf("invalid data classification: %s", rule.Classification)
		}
		if len(rule.AllowedOperations) == 0 {
			return fmt.Errorf("at least one allowed operation is required for classification %s", rule.Classification)
		}
	}

	// Validate risk thresholds
	if policy.Risk.ScoringEnabled {
		if policy.Risk.Thresholds.Low >= policy.Risk.Thresholds.Medium ||
			policy.Risk.Thresholds.Medium >= policy.Risk.Thresholds.High ||
			policy.Risk.Thresholds.High >= policy.Risk.Thresholds.Critical {
			return fmt.Errorf("risk thresholds must be in ascending order")
		}
	}

	return nil
}

// buildPolicySummary creates a summary of the policy for UI display
func (s *Service) buildPolicySummary(policy *AIPolicy) *PolicySummary {
	summary := &PolicySummary{
		AgencyID:      policy.AgencyID,
		Version:       policy.Version,
		AdoptionLevel: policy.Stance.AdoptionLevel,
		RiskTolerance: policy.Stance.RiskTolerance,
		LastUpdated:   policy.UpdatedAt,
	}

	// Count allowed providers and models
	for _, provider := range policy.Models.AllowedProviders {
		summary.AllowedProviders = append(summary.AllowedProviders, provider.Provider)
		summary.TotalModels += len(provider.Models)
		summary.TotalBudgetUSD += provider.MonthlyBudgetUSD
		summary.CurrentSpendUSD += provider.CurrentSpendUSD
	}

	// Autonomy info
	summary.DefaultAutonomyLevel = policy.Autonomy.DefaultLevel
	summary.RoleOverridesCount = len(policy.Autonomy.RoleOverrides)

	// Compliance frameworks
	for _, framework := range policy.Compliance.Frameworks {
		if framework.Enabled {
			summary.ComplianceFrameworks = append(summary.ComplianceFrameworks, framework.Name)
		}
	}

	// Monitoring status
	summary.MonitoringEnabled = policy.Monitoring.RealTimePolicyViolations
	summary.AuditingEnabled = policy.Compliance.AuditRequirements.LogAllActions

	return summary
}

// Helper validation functions

func isValidAdoptionLevel(level string) bool {
	switch level {
	case AdoptionConservative, AdoptionControlled, AdoptionProgressive, AdoptionInnovative:
		return true
	}
	return false
}

func isValidRiskTolerance(tolerance string) bool {
	switch tolerance {
	case RiskToleranceLow, RiskToleranceMedium, RiskToleranceHigh:
		return true
	}
	return false
}

func isValidAutonomyLevel(level string) bool {
	switch level {
	case AutonomyL0, AutonomyL1, AutonomyL2, AutonomyL3, AutonomyL4:
		return true
	}
	return false
}

func isValidDataClassification(classification string) bool {
	switch classification {
	case DataClassPublic, DataClassInternal, DataClassConfidential, DataClassRestricted:
		return true
	}
	return false
}

// PolicySummary represents a summary of the policy for UI display
type PolicySummary struct {
	AgencyID             string    `json:"agency_id"`
	Version              string    `json:"version"`
	AdoptionLevel        string    `json:"adoption_level"`
	RiskTolerance        string    `json:"risk_tolerance"`
	DefaultAutonomyLevel string    `json:"default_autonomy_level"`
	RoleOverridesCount   int       `json:"role_overrides_count"`
	AllowedProviders     []string  `json:"allowed_providers"`
	TotalModels          int       `json:"total_models"`
	TotalBudgetUSD       float64   `json:"total_budget_usd"`
	CurrentSpendUSD      float64   `json:"current_spend_usd"`
	ComplianceFrameworks []string  `json:"compliance_frameworks"`
	MonitoringEnabled    bool      `json:"monitoring_enabled"`
	AuditingEnabled      bool      `json:"auditing_enabled"`
	LastUpdated          time.Time `json:"last_updated"`
}
