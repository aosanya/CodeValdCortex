# MVP-051: Compliance Agent Architecture

## Overview

Design and implement the foundational architecture for **intelligent compliance agents** that provide context-aware, dynamic regulatory enforcement. This replaces the static compliance framework configuration with AI-powered agents that understand regulations, analyze context, and generate appropriate compliance workflows.

## Problem Statement

### Current Limitations (Static Configuration)
- ❌ **No Intelligence**: Compliance frameworks as hardcoded arrays in config
- ❌ **One-Size-Fits-All**: Same rules for all scenarios (medical data = marketing data)
- ❌ **Manual Maintenance**: Humans must update configs when regulations change
- ❌ **No Reasoning**: Cannot explain WHY a step is required
- ❌ **Context-Blind**: Cannot adapt to jurisdiction, data sensitivity, use case

### Agent-Based Solution
- ✅ **Context-Aware**: Analyzes data types, jurisdiction, purpose before enforcement
- ✅ **Explainable**: Provides reasoning for each requirement with article/control references
- ✅ **Adaptive**: Updates automatically when regulations change
- ✅ **Intelligent**: Understands lawful basis, risk levels, cross-border transfers
- ✅ **Auditable**: Complete reasoning trail for compliance officers

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Compliance Agents                        │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │   GDPR     │  │   SOC2     │  │   HIPAA    │  ...       │
│  │   Agent    │  │   Agent    │  │   Agent    │            │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘            │
└────────┼───────────────┼───────────────┼────────────────────┘
         │               │               │
         │  ┌────────────▼───────────────▼──────┐
         └─►│  Agent-to-Workflow Bridge         │
            │  - Converts plans to workflows    │
            │  - Creates work items             │
            └────────────┬──────────────────────┘
                         │
         ┌───────────────▼──────────────────────┐
         │     Workflow Designer (MVP-052)      │
         │  - Visualizes workflows              │
         │  - Executes compliance steps         │
         │  - Monitors compliance status        │
         └──────────────────────────────────────┘
```

### Component Architecture

```
ComplianceAgent (Interface)
├── KnowledgeBase           # Regulation articles, controls, guidelines
├── ContextAnalyzer         # Analyzes use case context
├── PolicyEngine            # Determines applicable rules
├── ReasoningEngine         # Explains requirements
├── PlanGenerator           # Creates compliance plans
└── UpdateMonitor           # Tracks regulation changes

Flow:
1. Context Analysis:    Identify data types, jurisdiction, purpose
2. Rule Selection:      Determine applicable articles/controls
3. Plan Generation:     Create context-specific compliance steps
4. Reasoning:           Explain why each step is needed
5. Workflow Conversion: Bridge converts plan to executable workflow
6. Execution:           Workflow engine executes compliance steps
7. Monitoring:          Track compliance status and violations
```

## Implementation

### 1. Core Interfaces

```go
// internal/compliance/agent.go

package compliance

import (
	"context"
	"time"
)

// ComplianceAgent is the base interface for all compliance agents
type ComplianceAgent interface {
	// Type returns the agent type (e.g., "gdpr", "soc2", "hipaa")
	Type() string
	
	// Analyze evaluates context and produces compliance assessment
	Analyze(ctx context.Context, input AnalysisInput) (*ComplianceAssessment, error)
	
	// GeneratePlan creates executable compliance plan from assessment
	GeneratePlan(ctx context.Context, assessment *ComplianceAssessment) (*CompliancePlan, error)
	
	// Monitor checks ongoing compliance during execution
	Monitor(ctx context.Context, execution *Execution) (*ComplianceReport, error)
	
	// Explain provides human-readable explanation for a step
	Explain(ctx context.Context, step ComplianceStep) (string, error)
	
	// Update processes regulatory updates
	Update(ctx context.Context, regulation RegulationUpdate) error
}

// AnalysisInput contains context for compliance analysis
type AnalysisInput struct {
	// Use case information
	UseCaseID   string                 `json:"use_case_id"`
	Purpose     string                 `json:"purpose"`
	Description string                 `json:"description"`
	
	// Data flow information
	DataTypes       []DataType         `json:"data_types"`
	DataSensitivity DataSensitivity    `json:"data_sensitivity"`
	
	// Jurisdictional information
	Jurisdictions   []string           `json:"jurisdictions"`
	DataResidency   string             `json:"data_residency"`
	
	// Operational context
	Users           []UserProfile      `json:"users"`
	Actions         []ActionType       `json:"actions"`
	ThirdParties    []ThirdParty       `json:"third_parties"`
	
	// Additional context
	Metadata        map[string]interface{} `json:"metadata"`
}

// ComplianceAssessment represents the result of analyzing context
type ComplianceAssessment struct {
	AgentType       string             `json:"agent_type"`
	AnalyzedAt      time.Time          `json:"analyzed_at"`
	
	// Identified requirements
	ApplicableRules []RuleReference    `json:"applicable_rules"`
	RiskLevel       RiskLevel          `json:"risk_level"`
	
	// Context-specific findings
	LawfulBasis     string             `json:"lawful_basis,omitempty"`      // GDPR
	TrustCriteria   []string           `json:"trust_criteria,omitempty"`    // SOC2
	PHIClassification string           `json:"phi_classification,omitempty"` // HIPAA
	
	// Recommendations
	RequiredSteps   []string           `json:"required_steps"`
	OptionalSteps   []string           `json:"optional_steps"`
	Warnings        []string           `json:"warnings"`
}

// CompliancePlan is the executable compliance workflow
type CompliancePlan struct {
	ID              string             `json:"id"`
	AgentType       string             `json:"agent_type"`
	CreatedAt       time.Time          `json:"created_at"`
	
	// Plan metadata
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Context         AnalysisInput      `json:"context"`
	Assessment      *ComplianceAssessment `json:"assessment"`
	
	// Executable steps
	Steps           []ComplianceStep   `json:"steps"`
	
	// Monitoring & reporting
	Monitoring      MonitoringConfig   `json:"monitoring"`
	ReportingFreq   time.Duration      `json:"reporting_frequency"`
}

// ComplianceStep represents a single compliance action
type ComplianceStep struct {
	ID              string             `json:"id"`
	Order           int                `json:"order"`
	
	// Step definition
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Type            StepType           `json:"type"` // validation|documentation|enforcement|audit
	
	// Regulatory basis
	RuleReference   RuleReference      `json:"rule_reference"`
	Mandatory       bool               `json:"mandatory"`
	Reasoning       string             `json:"reasoning"`
	
	// Execution details
	Implementation  string             `json:"implementation"` // How to execute
	ValidationCriteria []string        `json:"validation_criteria"`
	
	// Dependencies
	DependsOn       []string           `json:"depends_on"` // Step IDs
	
	// Metadata
	EstimatedDuration time.Duration    `json:"estimated_duration"`
	Priority         int               `json:"priority"`
}

// RuleReference links step to specific regulation article/control
type RuleReference struct {
	Framework       string             `json:"framework"` // GDPR, SOC2, HIPAA, ISO27001
	Article         string             `json:"article"`   // e.g., "Article 6(1)(a)"
	Section         string             `json:"section,omitempty"`
	ControlID       string             `json:"control_id,omitempty"`
	Description     string             `json:"description"`
	URL             string             `json:"url,omitempty"` // Link to official text
}

// ComplianceReport tracks compliance status
type ComplianceReport struct {
	PlanID          string             `json:"plan_id"`
	GeneratedAt     time.Time          `json:"generated_at"`
	
	Status          ComplianceStatus   `json:"status"`
	Progress        float64            `json:"progress"` // 0.0 - 1.0
	
	CompletedSteps  []string           `json:"completed_steps"`
	InProgressSteps []string           `json:"in_progress_steps"`
	BlockedSteps    []string           `json:"blocked_steps"`
	
	Violations      []ComplianceViolation `json:"violations"`
	Risks           []ComplianceRisk   `json:"risks"`
	
	Evidence        []EvidenceItem     `json:"evidence"`
}

// Supporting types
type DataType string
type DataSensitivity string
type RiskLevel string
type StepType string
type ComplianceStatus string

const (
	DataTypePersonal     DataType = "personal"
	DataTypeMedical      DataType = "medical"
	DataTypeFinancial    DataType = "financial"
	DataTypePublic       DataType = "public"
	
	SensitivityPublic     DataSensitivity = "public"
	SensitivityInternal   DataSensitivity = "internal"
	SensitivityConfidential DataSensitivity = "confidential"
	SensitivityRestricted DataSensitivity = "restricted"
	
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
	
	StepValidation    StepType = "validation"
	StepDocumentation StepType = "documentation"
	StepEnforcement   StepType = "enforcement"
	StepAudit         StepType = "audit"
	
	StatusPending    ComplianceStatus = "pending"
	StatusActive     ComplianceStatus = "active"
	StatusCompliant  ComplianceStatus = "compliant"
	StatusViolation  ComplianceStatus = "violation"
)
```

### 2. Base Agent Implementation

```go
// internal/compliance/base_agent.go

package compliance

import (
	"context"
	"fmt"
)

// BaseComplianceAgent provides common functionality for all agents
type BaseComplianceAgent struct {
	agentType       string
	knowledgeBase   *KnowledgeBase
	contextAnalyzer *ContextAnalyzer
	policyEngine    *PolicyEngine
	reasoningEngine *ReasoningEngine
	updateMonitor   *UpdateMonitor
}

func NewBaseComplianceAgent(
	agentType string,
	kb *KnowledgeBase,
) *BaseComplianceAgent {
	return &BaseComplianceAgent{
		agentType:       agentType,
		knowledgeBase:   kb,
		contextAnalyzer: NewContextAnalyzer(),
		policyEngine:    NewPolicyEngine(kb),
		reasoningEngine: NewReasoningEngine(kb),
		updateMonitor:   NewUpdateMonitor(agentType),
	}
}

func (b *BaseComplianceAgent) Type() string {
	return b.agentType
}

// Common analysis helpers
func (b *BaseComplianceAgent) analyzeDataTypes(input AnalysisInput) []DataClassification {
	return b.contextAnalyzer.ClassifyData(input.DataTypes, input.DataSensitivity)
}

func (b *BaseComplianceAgent) determineJurisdiction(input AnalysisInput) JurisdictionInfo {
	return b.contextAnalyzer.AnalyzeJurisdiction(input.Jurisdictions, input.DataResidency)
}

func (b *BaseComplianceAgent) assessRisk(input AnalysisInput) RiskLevel {
	return b.contextAnalyzer.CalculateRisk(
		input.DataTypes,
		input.DataSensitivity,
		input.Actions,
		input.ThirdParties,
	)
}
```

### 3. Agent-to-Workflow Bridge

```go
// internal/compliance/workflow_bridge.go

package compliance

import (
	"context"
	"fmt"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

// WorkflowBridge converts compliance plans to executable workflows
type WorkflowBridge struct {
	workflowBuilder *models.WorkflowBuilder
	workItemFactory *WorkItemFactory
}

func NewWorkflowBridge() *WorkflowBridge {
	return &WorkflowBridge{
		workflowBuilder: models.NewWorkflowBuilder(),
		workItemFactory: NewWorkItemFactory(),
	}
}

// ConvertToWorkflow transforms a compliance plan into an executable workflow
func (b *WorkflowBridge) ConvertToWorkflow(
	ctx context.Context,
	plan *CompliancePlan,
	agencyID string,
) (*models.Workflow, error) {
	workflow := &models.Workflow{
		Name:        plan.Name,
		Description: plan.Description,
		AgencyID:    agencyID,
		Steps:       []models.Step{},
	}
	
	// Convert each compliance step to a work item
	for i, complianceStep := range plan.Steps {
		// Create work item from compliance step
		workItem := b.workItemFactory.CreateFromCompliance(complianceStep)
		
		// Add compliance metadata for traceability
		workItem.Metadata = map[string]interface{}{
			"compliance_agent":    plan.AgentType,
			"regulation_framework": complianceStep.RuleReference.Framework,
			"regulation_article":  complianceStep.RuleReference.Article,
			"control_id":          complianceStep.RuleReference.ControlID,
			"reasoning":           complianceStep.Reasoning,
			"mandatory":           complianceStep.Mandatory,
			"rule_url":            complianceStep.RuleReference.URL,
		}
		
		// Add to workflow
		workflow.Steps = append(workflow.Steps, models.Step{
			ID:    fmt.Sprintf("step_%d", i+1),
			Order: i + 1,
			Items: []models.StepItem{{
				ID:           fmt.Sprintf("item_%d", i+1),
				WorkItemID:   workItem.ID,
				WorkItemKey:  workItem.Key,
				WorkItemName: workItem.Title,
			}},
		})
	}
	
	return workflow, nil
}

// WorkItemFactory creates work items from compliance steps
type WorkItemFactory struct{}

func NewWorkItemFactory() *WorkItemFactory {
	return &WorkItemFactory{}
}

func (f *WorkItemFactory) CreateFromCompliance(step ComplianceStep) *models.WorkItem {
	return &models.WorkItem{
		Title:       step.Name,
		Description: step.Description,
		Type:        string(step.Type),
		Priority:    step.Priority,
		
		// Add validation criteria as acceptance criteria
		AcceptanceCriteria: step.ValidationCriteria,
		
		// Add regulatory reference in notes
		Notes: fmt.Sprintf(
			"Regulatory Basis: %s %s\n\nReasoning: %s\n\nImplementation: %s",
			step.RuleReference.Framework,
			step.RuleReference.Article,
			step.Reasoning,
			step.Implementation,
		),
	}
}
```

### 4. Knowledge Base System

```go
// internal/compliance/knowledge_base.go

package compliance

import "time"

// KnowledgeBase stores regulatory knowledge for compliance agents
type KnowledgeBase struct {
	framework       string
	articles        map[string]Article
	controls        map[string]Control
	guidelines      map[string]Guideline
	lastUpdated     time.Time
}

// Article represents a regulation article
type Article struct {
	ID              string
	Number          string // e.g., "Article 6"
	Subsection      string // e.g., "(1)(a)"
	Title           string
	Text            string
	Interpretation  string
	Scope           []string // What it applies to
	Penalties       string
	OfficialURL     string
	LastReviewed    time.Time
}

// Control represents a compliance control (SOC2, ISO27001)
type Control struct {
	ID              string
	Framework       string
	Category        string
	Title           string
	Description     string
	Implementation  string
	TestProcedures  []string
	Evidence        []string
	Frequency       string
}

// Guideline represents regulatory guidance/interpretation
type Guideline struct {
	ID              string
	Authority       string // e.g., "ICO", "CNIL", "DPA"
	Topic           string
	Summary         string
	Details         string
	Jurisdiction    string
	PublishedDate   time.Time
	URL             string
}

// GetApplicableArticles returns articles relevant to context
func (kb *KnowledgeBase) GetApplicableArticles(
	dataTypes []DataType,
	jurisdiction JurisdictionInfo,
) []Article {
	// Implementation: Query knowledge base based on context
	return nil
}
```

## File Structure

```
internal/compliance/
├── agent.go                  # Core interfaces
├── base_agent.go            # Base implementation
├── knowledge_base.go        # Regulatory knowledge storage
├── context_analyzer.go      # Context analysis engine
├── policy_engine.go         # Policy application logic
├── reasoning_engine.go      # Explanation generation
├── update_monitor.go        # Regulation update tracking
├── workflow_bridge.go       # Plan-to-workflow conversion
└── agents/
    ├── gdpr_agent.go        # GDPR implementation (MVP-052)
    ├── soc2_agent.go        # SOC2 implementation (MVP-053)
    ├── hipaa_agent.go       # HIPAA implementation (MVP-053)
    └── iso27001_agent.go    # ISO27001 implementation (MVP-053)
```

## API Endpoints

```go
// internal/web/handlers/compliance_handler.go

// POST /api/compliance/analyze
// Request: AnalysisInput
// Response: ComplianceAssessment
func (h *ComplianceHandler) AnalyzeContext(c *gin.Context)

// POST /api/compliance/plan
// Request: ComplianceAssessment
// Response: CompliancePlan
func (h *ComplianceHandler) GeneratePlan(c *gin.Context)

// POST /api/compliance/plan/:id/workflow
// Converts compliance plan to executable workflow
func (h *ComplianceHandler) ConvertToWorkflow(c *gin.Context)

// GET /api/compliance/plan/:id/report
// Response: ComplianceReport
func (h *ComplianceHandler) GetReport(c *gin.Context)

// GET /api/compliance/agents
// Lists available compliance agents
func (h *ComplianceHandler) ListAgents(c *gin.Context)
```

## Success Metrics

- ✅ **Agent Interface**: Clean, extensible interface for all compliance agents
- ✅ **Knowledge Base**: Structured storage for regulatory knowledge
- ✅ **Bridge Service**: Successful conversion of compliance plans to workflows
- ✅ **Context Analysis**: Accurate classification of data types and jurisdictions
- ✅ **Explainability**: Clear reasoning for each compliance requirement
- ✅ **Testability**: >80% code coverage for core components

## Testing Strategy

1. **Unit Tests**: Test each component in isolation
2. **Integration Tests**: Test agent-to-workflow flow
3. **Mock Compliance Scenarios**: Various data types, jurisdictions
4. **Edge Cases**: Cross-border transfers, multi-jurisdiction
5. **Performance**: Handle 1000+ compliance steps efficiently

## Dependencies

- MVP-050: AI Policy Layer - Advanced Features (provides policy infrastructure)
- Internal: Workflow engine, work item system
- External: Regulatory knowledge sources (APIs, databases)

## Timeline

- **Week 1**: Core interfaces, base agent implementation
- **Week 2**: Knowledge base system, context analyzer
- **Week 3**: Workflow bridge, API endpoints
- **Week 4**: Testing, documentation, refinement

## Next Steps

After MVP-051 completion:
- **MVP-052**: GDPR Compliance Agent (specific implementation)
- **MVP-053**: Multi-Framework System (SOC2, HIPAA, ISO27001)
