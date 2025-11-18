package policy

import "time"

// AIPolicy represents the complete AI governance policy for an agency
type AIPolicy struct {
	Key       string    `json:"_key,omitempty" db:"_key"`
	ID        string    `json:"_id,omitempty" db:"_id"`
	Rev       string    `json:"_rev,omitempty" db:"_rev"`
	AgencyID  string    `json:"agency_id" db:"agency_id"`
	Version   string    `json:"version" db:"version"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Owner     string    `json:"owner" db:"owner"`

	// Policy configuration sections
	Stance     PolicyStance     `json:"stance" db:"stance"`
	Models     ModelPolicy      `json:"models" db:"models"`
	Autonomy   AutonomyPolicy   `json:"autonomy" db:"autonomy"`
	DataAccess DataAccessPolicy `json:"data_access" db:"data_access"`
	Actions    ActionPolicy     `json:"actions" db:"actions"`
	Risk       RiskPolicy       `json:"risk" db:"risk"`
	Compliance CompliancePolicy `json:"compliance" db:"compliance"`
	Monitoring MonitoringPolicy `json:"monitoring" db:"monitoring"`
}

// PolicyStance defines the organization's overall AI adoption approach
type PolicyStance struct {
	AdoptionLevel        string   `json:"adoption_level" db:"adoption_level"`               // conservative|controlled|progressive|innovative
	RiskTolerance        string   `json:"risk_tolerance" db:"risk_tolerance"`               // low|medium|high
	ComplianceFrameworks []string `json:"compliance_frameworks" db:"compliance_frameworks"` // SOC2, GDPR, HIPAA, etc.
}

// ModelPolicy defines which AI models and providers are allowed
type ModelPolicy struct {
	AllowedProviders []AllowedProvider `json:"allowed_providers" db:"allowed_providers"`
	DeniedProviders  []string          `json:"denied_providers" db:"denied_providers"`
	FallbackBehavior string            `json:"fallback_behavior" db:"fallback_behavior"` // fail_safe|degrade|queue_for_review
}

// AllowedProvider represents an approved AI provider configuration
type AllowedProvider struct {
	Provider            string   `json:"provider" db:"provider"`
	Models              []string `json:"models" db:"models"`
	DataResidency       string   `json:"data_residency" db:"data_residency"`
	MaxTokensPerRequest int      `json:"max_tokens_per_request" db:"max_tokens_per_request"`
	MonthlyBudgetUSD    float64  `json:"monthly_budget_usd" db:"monthly_budget_usd"`
	CurrentSpendUSD     float64  `json:"current_spend_usd" db:"current_spend_usd"`
}

// AutonomyPolicy defines autonomy levels and escalation rules
type AutonomyPolicy struct {
	DefaultLevel    string           `json:"default_level" db:"default_level"` // L0-L4
	RoleOverrides   []RoleOverride   `json:"role_overrides" db:"role_overrides"`
	EscalationRules []EscalationRule `json:"escalation_rules" db:"escalation_rules"`
}

// RoleOverride specifies a custom autonomy level for a specific role
type RoleOverride struct {
	RoleID               string   `json:"role_id" db:"role_id"`
	RoleName             string   `json:"role_name" db:"role_name"`
	Level                string   `json:"level" db:"level"` // L0-L4
	Justification        string   `json:"justification" db:"justification"`
	RequiresApprovalFrom []string `json:"requires_approval_from,omitempty" db:"requires_approval_from"`
}

// EscalationRule defines when to escalate autonomy level
type EscalationRule struct {
	Trigger          string                 `json:"trigger" db:"trigger"` // high_cost_action, sensitive_data_access, etc.
	EscalateToLevel  string                 `json:"escalate_to_level" db:"escalate_to_level"`
	ApprovalRequired bool                   `json:"approval_required" db:"approval_required"`
	AuditRequired    bool                   `json:"audit_required" db:"audit_required"`
	Conditions       map[string]interface{} `json:"conditions,omitempty" db:"conditions"`
}

// DataAccessPolicy defines data classification and access rules
type DataAccessPolicy struct {
	ClassificationRequired bool                     `json:"classification_required" db:"classification_required"`
	Rules                  []DataClassificationRule `json:"rules" db:"rules"`
	PIIHandling            PIIHandlingPolicy        `json:"pii_handling" db:"pii_handling"`
}

// DataClassificationRule defines access rules per data classification
type DataClassificationRule struct {
	Classification        string   `json:"classification" db:"classification"`         // public|internal|confidential|restricted
	AllowedOperations     []string `json:"allowed_operations" db:"allowed_operations"` // read|write|delete
	RetentionDays         int      `json:"retention_days" db:"retention_days"`
	RequiresJustification bool     `json:"requires_justification" db:"requires_justification"`
	RequiresApproval      bool     `json:"requires_approval" db:"requires_approval"`
	AuditAllAccess        bool     `json:"audit_all_access" db:"audit_all_access"`
	ExplicitGrantRequired bool     `json:"explicit_grant_required" db:"explicit_grant_required"`
	DualApproval          bool     `json:"dual_approval" db:"dual_approval"`
}

// PIIHandlingPolicy defines PII data handling requirements
type PIIHandlingPolicy struct {
	DetectPII             bool   `json:"detect_pii" db:"detect_pii"`
	AnonymizationRequired bool   `json:"anonymization_required" db:"anonymization_required"`
	CrossBorderTransfer   string `json:"cross_border_transfer" db:"cross_border_transfer"` // allow|deny|require_approval
	DeletionOnRequest     bool   `json:"deletion_on_request" db:"deletion_on_request"`
}

// ActionPolicy defines action authorization and approval workflows
type ActionPolicy struct {
	ApprovalWorkflows    []ApprovalWorkflow    `json:"approval_workflows" db:"approval_workflows"`
	ProhibitedActions    []string              `json:"prohibited_actions" db:"prohibited_actions"`
	RollbackRequirements []RollbackRequirement `json:"rollback_requirements" db:"rollback_requirements"`
}

// ApprovalWorkflow defines when and who must approve actions
type ApprovalWorkflow struct {
	ActionPattern     string              `json:"action_pattern" db:"action_pattern"` // regex pattern
	RequiresApproval  bool                `json:"requires_approval" db:"requires_approval"`
	Approvers         []string            `json:"approvers" db:"approvers"`
	DualApproval      bool                `json:"dual_approval" db:"dual_approval"`
	TimeoutMinutes    int                 `json:"timeout_minutes" db:"timeout_minutes"`
	ApprovalChain     []string            `json:"approval_chain,omitempty" db:"approval_chain"`
	ApprovalThreshold []ApprovalThreshold `json:"approval_threshold,omitempty" db:"approval_threshold"`
}

// ApprovalThreshold defines approval requirements based on action magnitude
type ApprovalThreshold struct {
	Amount    float64  `json:"amount" db:"amount"`
	Approvers []string `json:"approvers" db:"approvers"`
}

// RollbackRequirement specifies rollback planning requirements
type RollbackRequirement struct {
	ActionPattern        string `json:"action_pattern" db:"action_pattern"`
	RollbackPlanRequired bool   `json:"rollback_plan_required" db:"rollback_plan_required"`
	TestRollback         bool   `json:"test_rollback" db:"test_rollback"`
}

// RiskPolicy defines risk assessment and mitigation rules
type RiskPolicy struct {
	ScoringEnabled bool                `json:"scoring_enabled" db:"scoring_enabled"`
	Thresholds     RiskThresholds      `json:"thresholds" db:"thresholds"`
	Mitigation     map[string][]string `json:"mitigation" db:"mitigation"` // risk level -> mitigation requirements
}

// RiskThresholds defines risk score boundaries
type RiskThresholds struct {
	Low      int `json:"low" db:"low"`
	Medium   int `json:"medium" db:"medium"`
	High     int `json:"high" db:"high"`
	Critical int `json:"critical" db:"critical"`
}

// CompliancePolicy defines compliance framework requirements
type CompliancePolicy struct {
	Frameworks        []ComplianceFramework `json:"frameworks" db:"frameworks"`
	AuditRequirements AuditRequirements     `json:"audit_requirements" db:"audit_requirements"`
	Reporting         ReportingRequirements `json:"reporting" db:"reporting"`
}

// ComplianceFramework represents a specific compliance standard
type ComplianceFramework struct {
	Name     string              `json:"name" db:"name"` // SOC2, GDPR, HIPAA, etc.
	Controls []ComplianceControl `json:"controls" db:"controls"`
	Enabled  bool                `json:"enabled" db:"enabled"`
}

// ComplianceControl represents a specific control within a framework
type ComplianceControl struct {
	ID          string `json:"id" db:"id"`
	Description string `json:"description" db:"description"`
	Enforcement string `json:"enforcement" db:"enforcement"` // strict|moderate|advisory
}

// AuditRequirements defines audit logging requirements
type AuditRequirements struct {
	LogAllActions     bool `json:"log_all_actions" db:"log_all_actions"`
	ImmutableAuditLog bool `json:"immutable_audit_log" db:"immutable_audit_log"`
	RetentionYears    int  `json:"retention_years" db:"retention_years"`
}

// ReportingRequirements defines compliance reporting schedules
type ReportingRequirements struct {
	DailySummary            bool `json:"daily_summary" db:"daily_summary"`
	WeeklyComplianceReport  bool `json:"weekly_compliance_report" db:"weekly_compliance_report"`
	QuarterlyRiskAssessment bool `json:"quarterly_risk_assessment" db:"quarterly_risk_assessment"`
}

// MonitoringPolicy defines real-time monitoring and alerting
type MonitoringPolicy struct {
	RealTimePolicyViolations bool              `json:"real_time_policy_violations" db:"real_time_policy_violations"`
	Alerts                   []MonitoringAlert `json:"alerts" db:"alerts"`
}

// MonitoringAlert defines alert conditions and recipients
type MonitoringAlert struct {
	Condition string   `json:"condition" db:"condition"` // policy_violation, budget_threshold, etc.
	Threshold float64  `json:"threshold,omitempty" db:"threshold"`
	Severity  string   `json:"severity" db:"severity"` // low|medium|high|critical
	NotifyTo  []string `json:"notify" db:"notify"`
}

// PolicyResult represents the outcome of a policy evaluation
type PolicyResult struct {
	Allowed            bool                  `json:"allowed"`
	Reason             string                `json:"reason"`
	RequiresApproval   bool                  `json:"requires_approval"`
	Approvers          []string              `json:"approvers,omitempty"`
	ApprovalTimeout    int                   `json:"approval_timeout,omitempty"`
	ApprovalWorkflowID string                `json:"approval_workflow_id,omitempty"`
	RequiresEscalation bool                  `json:"requires_escalation"`
	EscalateToLevel    string                `json:"escalate_to_level,omitempty"`
	RequiresRiskReview bool                  `json:"requires_risk_review"`
	RiskScore          int                   `json:"risk_score"`
	RiskFactors        []string              `json:"risk_factors,omitempty"`
	EnhancedMonitoring bool                  `json:"enhanced_monitoring"`
	Conditions         []MonitoringCondition `json:"conditions,omitempty"`
	AuditRequired      bool                  `json:"audit_required"`
	ComplianceControls []string              `json:"compliance_controls,omitempty"`
	PolicyVersion      string                `json:"policy_version"`
	EvaluatedAt        time.Time             `json:"evaluated_at"`
	Severity           string                `json:"severity"` // info|medium|high|critical
}

// MonitoringCondition represents a specific monitoring requirement
type MonitoringCondition struct {
	Type      string                 `json:"type"`
	Threshold float64                `json:"threshold,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PolicyEvaluation represents an audit log entry for policy evaluation
type PolicyEvaluation struct {
	Key           string    `json:"_key,omitempty" db:"_key"`
	ID            string    `json:"_id,omitempty" db:"_id"`
	Rev           string    `json:"_rev,omitempty" db:"_rev"`
	AgencyID      string    `json:"agency_id" db:"agency_id"`
	PolicyVersion string    `json:"policy_version" db:"policy_version"`
	EvaluatedAt   time.Time `json:"evaluated_at" db:"evaluated_at"`

	// Context
	AgentID  string `json:"agent_id,omitempty" db:"agent_id"`
	RoleID   string `json:"role_id,omitempty" db:"role_id"`
	Action   string `json:"action" db:"action"`
	Resource string `json:"resource,omitempty" db:"resource"`

	// Result
	Result PolicyResult `json:"result" db:"result"`

	// Additional metadata
	RequestID string `json:"request_id,omitempty" db:"request_id"`
	UserID    string `json:"user_id,omitempty" db:"user_id"`
}

// PolicyViolation represents a recorded policy violation
type PolicyViolation struct {
	Key           string    `json:"_key,omitempty" db:"_key"`
	ID            string    `json:"_id,omitempty" db:"_id"`
	Rev           string    `json:"_rev,omitempty" db:"_rev"`
	AgencyID      string    `json:"agency_id" db:"agency_id"`
	PolicyVersion string    `json:"policy_version" db:"policy_version"`
	ViolatedAt    time.Time `json:"violated_at" db:"violated_at"`

	// Context
	AgentID  string `json:"agent_id,omitempty" db:"agent_id"`
	RoleID   string `json:"role_id,omitempty" db:"role_id"`
	Action   string `json:"action" db:"action"`
	Resource string `json:"resource,omitempty" db:"resource"`

	// Violation details
	ViolationType string `json:"violation_type" db:"violation_type"` // unauthorized_model, exceeded_budget, prohibited_action, etc.
	Severity      string `json:"severity" db:"severity"`             // low|medium|high|critical
	Reason        string `json:"reason" db:"reason"`
	Blocked       bool   `json:"blocked" db:"blocked"` // Was the action blocked?

	// Response
	NotifiedTo []string   `json:"notified_to,omitempty" db:"notified_to"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy string     `json:"resolved_by,omitempty" db:"resolved_by"`
	Resolution string     `json:"resolution,omitempty" db:"resolution"`
}

// Autonomy levels as constants
const (
	AutonomyL0 = "L0" // Full oversight - All decisions require approval
	AutonomyL1 = "L1" // Assisted - AI provides recommendations, human approves
	AutonomyL2 = "L2" // Monitored - AI acts autonomously with human oversight
	AutonomyL3 = "L3" // Delegated - AI operates independently, humans intervene on exceptions
	AutonomyL4 = "L4" // Autonomous - AI operates fully autonomously
)

// Adoption levels as constants
const (
	AdoptionConservative = "conservative"
	AdoptionControlled   = "controlled"
	AdoptionProgressive  = "progressive"
	AdoptionInnovative   = "innovative"
)

// Risk tolerance levels
const (
	RiskToleranceLow    = "low"
	RiskToleranceMedium = "medium"
	RiskToleranceHigh   = "high"
)

// Data classifications
const (
	DataClassPublic       = "public"
	DataClassInternal     = "internal"
	DataClassConfidential = "confidential"
	DataClassRestricted   = "restricted"
)

// Severities
const (
	SeverityInfo     = "info"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)
