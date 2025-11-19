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

// validatePolicy validates policy configuration (used by web handlers)
func (s *Service) ValidatePolicy(policy *AIPolicy) error {
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
	AgencyID             string   `json:"agency_id"`
	Version              string   `json:"version"`
	AdoptionLevel        string   `json:"adoption_level"`
	RiskTolerance        string   `json:"risk_tolerance"`
	DefaultAutonomyLevel string   `json:"default_autonomy_level"`
	RoleOverridesCount   int      `json:"role_overrides_count"`
	AllowedProviders     []string `json:"allowed_providers"`
	TotalModels          int      `json:"total_models"`
	TotalBudgetUSD       float64  `json:"total_budget_usd"`
	CurrentSpendUSD      float64  `json:"current_spend_usd"`
	// ComplianceFrameworks field removed - see MVP-051-053 for agent-based compliance
	MonitoringEnabled bool      `json:"monitoring_enabled"`
	AuditingEnabled   bool      `json:"auditing_enabled"`
	LastUpdated       time.Time `json:"last_updated"`
}
