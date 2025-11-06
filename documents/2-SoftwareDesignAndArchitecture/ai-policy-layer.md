# AI Policy Layer - Design Specification

## Overview

The AI Policy Layer provides runtime governance and guardrails for AI agent operations, addressing DORA's finding that **clear AI policies reduce friction and amplify AI benefits** across organizations. This layer enforces organizational AI stance, autonomy boundaries, approval workflows, and compliance requirements.

---

## 1. AI Policy Framework

### 1.1 Core Policy Dimensions

**1. Model Access Policy**
- Which AI models/providers are approved for use
- Data residency and sovereignty requirements
- Cost controls and budget allocations
- API key management and rotation

**2. Autonomy Level Policy**
- Default autonomy level per role (L0-L4)
- Escalation requirements for higher autonomy
- Human-in-the-loop (HITL) trigger conditions
- Approval chain configuration

**3. Data Access Policy**
- What data sources agents can access
- PII/sensitive data handling rules
- Data retention and deletion policies
- Cross-boundary data movement restrictions

**4. Action Authorization Policy**
- What actions agents can perform autonomously
- Which actions require pre-approval
- Which actions are explicitly prohibited
- Rollback/compensation requirements

**5. Risk & Compliance Policy**
- Industry-specific compliance requirements (HIPAA, SOC2, GDPR)
- Risk scoring and mitigation rules
- Audit trail requirements
- Incident response procedures

---

## 2. Policy Schema

### 2.1 Agency AI Policy Document

```yaml
apiVersion: policy.codevaldcortex.io/v1
kind: AIPolicy
metadata:
  agency_id: "agency_financial_risk_analysis"
  version: "1.0.0"
  created: "2025-11-06T10:00:00Z"
  owner: "compliance-team@example.com"

spec:
  # Overall AI stance for the organization
  stance:
    adoption_level: "controlled"  # conservative | controlled | progressive | innovative
    risk_tolerance: "low"          # low | medium | high
    compliance_frameworks: ["SOC2", "ISO27001", "GDPR"]
    
  # Model usage policies
  models:
    allowed_providers:
      - provider: "openai"
        models: ["gpt-4", "gpt-3.5-turbo"]
        data_residency: "US"
        max_tokens_per_request: 4096
        monthly_budget_usd: 50000
        
      - provider: "anthropic"
        models: ["claude-3-opus", "claude-3-sonnet"]
        data_residency: "US"
        max_tokens_per_request: 8192
        monthly_budget_usd: 30000
        
    denied_providers: ["provider-x"]  # Explicitly blocked
    
    fallback_behavior: "fail_safe"  # fail_safe | degrade | queue_for_review
    
  # Autonomy policies by role
  autonomy:
    default_level: "L1"  # Default for new roles
    
    role_overrides:
      - role: "data_collector"
        level: "L2"
        justification: "Low-risk read-only operations"
        
      - role: "financial_analyst"
        level: "L1"
        justification: "Recommendations only, no execution"
        
      - role: "trade_executor"
        level: "L0"
        justification: "High-risk financial transactions"
        requires_approval_from: ["trading_manager", "risk_officer"]
        
    escalation_rules:
      - trigger: "high_cost_action"  # > $10k transaction
        escalate_to_level: "L0"
        approval_required: true
        
      - trigger: "sensitive_data_access"
        escalate_to_level: "L1"
        audit_required: true
        
  # Data access policies
  data_access:
    classification_required: true
    
    rules:
      - classification: "public"
        allowed_operations: ["read", "write", "delete"]
        retention_days: 365
        
      - classification: "internal"
        allowed_operations: ["read", "write"]
        retention_days: 730
        requires_justification: true
        
      - classification: "confidential"
        allowed_operations: ["read"]
        retention_days: 2555  # 7 years
        requires_approval: true
        audit_all_access: true
        
      - classification: "restricted"
        allowed_operations: []
        explicit_grant_required: true
        dual_approval: true
        
    pii_handling:
      detect_pii: true
      anonymization_required: true
      cross_border_transfer: "deny"
      deletion_on_request: true
      
  # Action authorization policies
  actions:
    approval_workflows:
      - action_pattern: "deploy_*"
        requires_approval: true
        approvers: ["ops_manager"]
        timeout_minutes: 30
        
      - action_pattern: "delete_*"
        requires_approval: true
        approvers: ["data_steward", "ops_manager"]
        dual_approval: true
        timeout_minutes: 60
        
      - action_pattern: "financial_transaction"
        requires_approval: true
        approval_chain: ["agent_supervisor", "financial_controller", "cfo"]
        approval_threshold:
          - amount: 1000
            approvers: ["agent_supervisor"]
          - amount: 10000
            approvers: ["agent_supervisor", "financial_controller"]
          - amount: 100000
            approvers: ["agent_supervisor", "financial_controller", "cfo"]
            
    prohibited_actions:
      - "drop_database"
      - "modify_security_policy"
      - "grant_root_access"
      
    rollback_requirements:
      - action_pattern: "update_*"
        rollback_plan_required: true
        test_rollback: true
        
  # Risk management
  risk:
    scoring_enabled: true
    
    thresholds:
      low: 30
      medium: 60
      high: 85
      
    mitigation_requirements:
      high:
        - human_review: true
        - additional_approval: true
        - enhanced_monitoring: true
        
      critical:
        - requires_incident_response_plan: true
        - executive_notification: true
        - full_audit_trail: true
        
  # Compliance requirements
  compliance:
    frameworks:
      - name: "SOC2"
        controls:
          - id: "CC6.1"
            description: "Logical and physical access controls"
            enforcement: "strict"
            
      - name: "GDPR"
        controls:
          - id: "Article.17"
            description: "Right to erasure"
            enforcement: "strict"
            
    audit_requirements:
      log_all_actions: true
      immutable_audit_log: true
      retention_years: 7
      
    reporting:
      daily_summary: true
      weekly_compliance_report: true
      quarterly_risk_assessment: true
      
  # Monitoring and alerting
  monitoring:
    real_time_policy_violations: true
    
    alerts:
      - condition: "policy_violation"
        severity: "high"
        notify: ["security_team", "compliance_officer"]
        
      - condition: "budget_threshold_exceeded"
        threshold: 0.8  # 80% of budget
        severity: "medium"
        notify: ["finance_team"]
        
      - condition: "autonomy_escalation"
        severity: "medium"
        notify: ["operations_manager"]
```

---

## 3. First-Run AI Policy Wizard

### 3.1 Wizard Flow

**Step 1: Organization Context**
```
Question: What industry are you in?
Options: 
  - Financial Services
  - Healthcare
  - Manufacturing
  - Telecommunications
  - Government
  - Technology/SaaS
  - Other

Question: What compliance frameworks apply?
Options (multi-select):
  - SOC 2 Type II
  - ISO 27001
  - HIPAA
  - GDPR
  - PCI DSS
  - CCPA
  - FedRAMP
  - None currently

Result: Pre-selects compliance controls and data policies
```

**Step 2: Risk Tolerance**
```
Question: What's your organization's AI adoption stance?

Conservative:
  - AI provides recommendations only (L0-L1)
  - All AI outputs require human review
  - Strict approval workflows
  - Enhanced audit trails
  
Controlled (Recommended):
  - AI can perform low-risk actions autonomously (L1-L2)
  - Human review for medium/high-risk actions
  - Standard approval workflows
  - Comprehensive audit trails
  
Progressive:
  - AI can perform most actions autonomously (L2-L3)
  - Human review for high-risk actions only
  - Streamlined approval workflows
  - Standard audit trails
  
Innovative:
  - AI operates with high autonomy (L3-L4)
  - Humans intervene for edge cases only
  - Minimal approval friction
  - Real-time monitoring

Result: Sets default autonomy levels and approval thresholds
```

**Step 3: AI Model Selection**
```
Question: Which AI providers can your organization use?

Options:
  ☐ OpenAI (GPT-4, GPT-3.5)
  ☐ Anthropic (Claude)
  ☐ Google (Gemini)
  ☐ Azure OpenAI (Enterprise SLA)
  ☐ AWS Bedrock
  ☐ Self-hosted models
  
For each selected:
  - Data residency requirements: [Dropdown: US, EU, UK, etc.]
  - Monthly budget (USD): [Input]
  - Max tokens per request: [Input]

Result: Configures model access policies
```

**Step 4: Data Classification**
```
Question: How do you classify data sensitivity?

We'll use these standard classifications:
  - Public: Can be freely shared
  - Internal: For organization use only
  - Confidential: Restricted access, audit required
  - Restricted: Highest sensitivity, dual approval

Question: Does your organization handle PII?
  ☐ Yes → Enable PII detection and anonymization
  ☐ No

Question: Are there geographic restrictions on data?
  ☐ No cross-border data transfer
  ☐ EU data stays in EU
  ☐ Custom restrictions

Result: Configures data access policies
```

**Step 5: Approval Workflows**
```
Question: Who should approve high-risk agent actions?

Roles:
  - Operations Manager: [Name/Email]
  - Security Officer: [Name/Email]
  - Compliance Officer: [Name/Email]
  - Financial Controller: [Name/Email]

Question: What actions require approval?
  ☐ Database modifications
  ☐ Financial transactions above $[Input]
  ☐ Access to confidential data
  ☐ External API calls
  ☐ Deployments/configuration changes
  ☐ User account modifications

Result: Configures approval workflows
```

**Step 6: Review & Generate**
```
Summary:
  Industry: Financial Services
  Compliance: SOC 2, ISO 27001
  Stance: Controlled
  Approved Providers: OpenAI, Anthropic
  Default Autonomy: L1
  PII Handling: Enabled
  Approval Required: Database changes, Financial transactions >$10k

[Generate Policy] [Edit] [Cancel]

Result: Creates agency_ai_policy document in ArangoDB
```

---

## 4. Runtime Policy Enforcement

### 4.1 Policy Evaluation Engine

```go
// internal/policy/engine.go
package policy

type PolicyEngine struct {
    repo       *PolicyRepository
    evaluator  *PolicyEvaluator
    auditor    *AuditLogger
    notifier   *AlertNotifier
}

// EvaluateAction checks if an agent action is allowed under current policy
func (pe *PolicyEngine) EvaluateAction(ctx context.Context, req *ActionRequest) (*PolicyDecision, error) {
    // 1. Load agency policy
    policy, err := pe.repo.GetAgencyPolicy(ctx, req.AgencyID)
    if err != nil {
        return nil, err
    }
    
    // 2. Check model usage policy
    if !pe.evaluator.IsModelAllowed(policy, req.ModelProvider, req.ModelName) {
        return &PolicyDecision{
            Allowed: false,
            Reason: "Model not approved in organization policy",
            Remediation: "Contact compliance team to request model approval",
        }, nil
    }
    
    // 3. Check autonomy level
    requiredLevel := pe.evaluator.GetRequiredAutonomyLevel(policy, req.Action)
    agentLevel := req.Agent.AutonomyLevel
    
    if agentLevel < requiredLevel {
        return &PolicyDecision{
            Allowed: false,
            Reason: fmt.Sprintf("Action requires autonomy level %s, agent has %s", requiredLevel, agentLevel),
            RequiresEscalation: true,
            ApprovalRequired: true,
            Approvers: policy.GetApprovers(req.Action),
        }, nil
    }
    
    // 4. Check data access policy
    if req.DataAccess != nil {
        dataDecision := pe.evaluator.EvaluateDataAccess(policy, req.DataAccess)
        if !dataDecision.Allowed {
            return dataDecision, nil
        }
    }
    
    // 5. Check action authorization
    if pe.evaluator.IsProhibitedAction(policy, req.Action) {
        pe.auditor.LogPolicyViolation(ctx, req, "prohibited_action")
        return &PolicyDecision{
            Allowed: false,
            Reason: "Action is explicitly prohibited",
            Severity: "critical",
        }, nil
    }
    
    if pe.evaluator.RequiresApproval(policy, req.Action, req.Context) {
        return &PolicyDecision{
            Allowed: false,
            RequiresApproval: true,
            Approvers: policy.GetApprovers(req.Action),
            ApprovalTimeout: policy.GetApprovalTimeout(req.Action),
            Reason: "Action requires approval per organization policy",
        }, nil
    }
    
    // 6. Calculate risk score
    riskScore := pe.evaluator.CalculateRiskScore(req)
    if riskScore > policy.Risk.Thresholds.High {
        return &PolicyDecision{
            Allowed: false,
            RequiresRiskReview: true,
            RiskScore: riskScore,
            Reason: "Risk score exceeds threshold",
            Remediation: "Reduce action scope or request exception",
        }, nil
    }
    
    // 7. Check budget constraints
    if req.EstimatedCost > 0 {
        budgetOk := pe.evaluator.CheckBudget(policy, req.ModelProvider, req.EstimatedCost)
        if !budgetOk {
            return &PolicyDecision{
                Allowed: false,
                Reason: "Action would exceed monthly AI budget",
                Remediation: "Wait for budget reset or request additional budget",
            }, nil
        }
    }
    
    // All checks passed
    return &PolicyDecision{
        Allowed: true,
        RiskScore: riskScore,
        Conditions: pe.evaluator.GetMonitoringConditions(policy, riskScore),
    }, nil
}
```

### 4.2 Policy Decision Structure

```go
type PolicyDecision struct {
    Allowed             bool
    Reason              string
    Remediation         string
    
    // Approval workflow
    RequiresApproval    bool
    Approvers           []string
    ApprovalTimeout     time.Duration
    ApprovalWorkflowID  string
    
    // Escalation
    RequiresEscalation  bool
    EscalateToLevel     string
    
    // Risk assessment
    RequiresRiskReview  bool
    RiskScore           int
    RiskFactors         []string
    
    // Monitoring
    EnhancedMonitoring  bool
    Conditions          []MonitoringCondition
    
    // Compliance
    AuditRequired       bool
    ComplianceControls  []string
    
    // Metadata
    PolicyVersion       string
    EvaluatedAt         time.Time
    Severity            string  // info | medium | high | critical
}
```

---

## 5. UI Integration

### 5.1 Policy Status Indicators

**Agent Card UI Enhancement**
```html
<div class="agent-card">
    <div class="agent-header">
        <h3>Financial Analyst Agent</h3>
        <span class="autonomy-badge level-1">L1 - Assisted</span>
    </div>
    
    <!-- Policy Status -->
    <div class="policy-status">
        <div class="policy-indicator allowed">
            <i class="fas fa-shield-check"></i>
            <span>Policy Compliant</span>
        </div>
        
        <div class="policy-details">
            <span class="policy-item">
                <i class="fas fa-brain"></i> Models: GPT-4, Claude
            </span>
            <span class="policy-item">
                <i class="fas fa-lock"></i> Data: Internal & Public only
            </span>
            <span class="policy-item">
                <i class="fas fa-dollar-sign"></i> Budget: $2,400 / $5,000
            </span>
        </div>
    </div>
</div>
```

**Action Approval UI**
```html
<div class="approval-request">
    <div class="approval-header warning">
        <i class="fas fa-exclamation-triangle"></i>
        <h3>Approval Required</h3>
    </div>
    
    <div class="approval-body">
        <p class="reason">
            This action requires approval: <strong>Financial transaction exceeds $10,000</strong>
        </p>
        
        <div class="action-details">
            <dl>
                <dt>Agent:</dt>
                <dd>Trade Executor</dd>
                
                <dt>Action:</dt>
                <dd>Execute market order - Buy 1000 shares AAPL</dd>
                
                <dt>Risk Score:</dt>
                <dd><span class="risk-badge medium">65 / 100</span></dd>
                
                <dt>Estimated Cost:</dt>
                <dd>$175,450.00</dd>
                
                <dt>Policy Requirement:</dt>
                <dd>Dual approval: Financial Controller + CFO</dd>
            </dl>
        </div>
        
        <div class="approval-actions">
            <button class="button is-success">
                <i class="fas fa-check"></i> Approve
            </button>
            <button class="button is-danger">
                <i class="fas fa-times"></i> Deny
            </button>
            <button class="button is-info">
                <i class="fas fa-edit"></i> Request Changes
            </button>
        </div>
    </div>
</div>
```

**Policy Violation Alert**
```html
<div class="notification is-danger">
    <button class="delete"></button>
    <div class="notification-content">
        <div class="notification-header">
            <i class="fas fa-shield-virus"></i>
            <strong>Policy Violation Detected</strong>
        </div>
        <p>
            Agent "Data Processor" attempted to access RESTRICTED data classification without required dual approval.
        </p>
        <div class="notification-details">
            <span>Time: 2025-11-06 14:23:15</span>
            <span>Policy: GDPR Article 17</span>
            <span>Action Blocked: Yes</span>
        </div>
        <div class="notification-actions">
            <a href="/audit/incident/12345">View Incident Report</a>
            <a href="/policy/review">Review Policy</a>
        </div>
    </div>
</div>
```

---

## 6. Policy Enforcement Points

### 6.1 Integration Points

```
Agent Action Flow with Policy Enforcement:

┌─────────────────────┐
│  Agent Requests     │
│  Action             │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Policy Engine      │◄─── Loads agency AI policy
│  - Model check      │
│  - Autonomy check   │
│  - Data access chk  │
│  - Action auth chk  │
│  - Risk scoring     │
│  - Budget check     │
└──────────┬──────────┘
           │
     ┌─────┴─────┐
     │           │
  Allowed    Blocked/Approval
     │           │
     ▼           ▼
┌─────────┐  ┌─────────────┐
│ Execute │  │  Approval   │
│ Action  │  │  Workflow   │
└─────────┘  └──────┬──────┘
     │               │
     │         ┌─────┴──────┐
     │         │            │
     │      Approved    Denied
     │         │            │
     └─────────┴────────────┘
               │
               ▼
       ┌───────────────┐
       │  Audit Log    │
       │  - Decision   │
       │  - Reason     │
       │  - Approvers  │
       │  - Timestamp  │
       └───────────────┘
```

### 6.2 Enforcement Locations

1. **AI Builder Service** (`/internal/builder/ai/`)
   - Model selection enforcement
   - Token budget tracking
   - Provider allowlist check

2. **Agent Runtime** (`/internal/runtime/`)
   - Autonomy level enforcement
   - Action authorization check
   - Real-time policy evaluation

3. **Task Execution** (`/internal/task/`)
   - Pre-execution policy check
   - Risk scoring
   - Approval workflow integration

4. **Data Access Layer** (`/internal/database/`)
   - Data classification enforcement
   - Access audit logging
   - PII detection and masking

5. **API Gateway** (`/internal/api/`)
   - Budget enforcement
   - Rate limiting by policy
   - Compliance header injection

---

## 7. Implementation Roadmap

### Phase 1: Policy Foundation (MVP-048)
**Duration**: 2 weeks

**Deliverables**:
1. ✅ Policy schema definition
2. ✅ Policy repository (ArangoDB)
3. ✅ First-run policy wizard UI
4. ✅ Basic policy engine
5. ✅ Model and autonomy enforcement

**Files to Create**:
```
/internal/policy/
├── engine.go          # Policy evaluation engine
├── repository.go      # Policy CRUD operations
├── types.go          # Policy data structures
├── evaluator.go      # Policy rule evaluation
└── auditor.go        # Policy violation logging

/internal/web/handlers/policy/
├── wizard.go         # First-run wizard handler
├── policy_crud.go    # Policy management handlers
└── approval.go       # Approval workflow handlers

/internal/web/templates/policy/
├── wizard_step1.templ
├── wizard_step2.templ
├── wizard_step3.templ
├── wizard_step4.templ
├── wizard_step5.templ
├── wizard_step6.templ
└── policy_status.templ
```

### Phase 2: Runtime Enforcement (MVP-049)
**Duration**: 2 weeks

**Deliverables**:
1. ✅ Action authorization engine
2. ✅ Approval workflow system
3. ✅ Risk scoring calculator
4. ✅ Budget tracking and alerts
5. ✅ Policy violation handling

### Phase 3: Advanced Features (MVP-050)
**Duration**: 2 weeks

**Deliverables**:
1. ✅ Data classification engine
2. ✅ PII detection and masking
3. ✅ Compliance reporting
4. ✅ Policy versioning and rollback
5. ✅ Multi-policy inheritance

---

## 8. Success Metrics

### 8.1 Policy Adoption Metrics
- % of agencies with defined AI policy
- Time to complete first-run wizard (target: <10 minutes)
- Policy update frequency

### 8.2 Enforcement Metrics
- Policy evaluation latency (target: <50ms)
- Policy violation rate
- Approval request volume
- Approval resolution time

### 8.3 Compliance Metrics
- Audit log completeness (target: 100%)
- Policy exception rate
- Compliance control coverage
- Incident response time

### 8.4 Business Impact Metrics
- Reduced compliance certification time
- Decreased policy violation incidents
- Improved audit readiness
- Faster onboarding of new agencies

---

## 9. API Examples

### 9.1 Create Agency Policy

```bash
POST /api/v1/agencies/{agency_id}/policy
Content-Type: application/json

{
  "stance": {
    "adoption_level": "controlled",
    "risk_tolerance": "low",
    "compliance_frameworks": ["SOC2", "GDPR"]
  },
  "models": {
    "allowed_providers": [
      {
        "provider": "openai",
        "models": ["gpt-4"],
        "monthly_budget_usd": 5000
      }
    ]
  },
  "autonomy": {
    "default_level": "L1"
  }
}
```

### 9.2 Evaluate Action Against Policy

```bash
POST /api/v1/policy/evaluate
Content-Type: application/json

{
  "agency_id": "agency_financial_analysis",
  "agent_id": "agent_trader_001",
  "action": "financial_transaction",
  "action_details": {
    "type": "buy_stock",
    "amount": 15000,
    "symbol": "AAPL"
  },
  "model_provider": "openai",
  "model_name": "gpt-4",
  "estimated_cost": 0.50
}

Response:
{
  "allowed": false,
  "reason": "Action requires approval: Financial transaction exceeds $10,000",
  "requires_approval": true,
  "approvers": ["financial_controller", "cfo"],
  "approval_timeout": "30m",
  "risk_score": 65,
  "policy_version": "1.0.0"
}
```

### 9.3 Submit Approval Request

```bash
POST /api/v1/policy/approvals
Content-Type: application/json

{
  "agency_id": "agency_financial_analysis",
  "agent_id": "agent_trader_001",
  "action": "financial_transaction",
  "action_details": {...},
  "approvers": ["financial_controller", "cfo"],
  "requester": "agent_supervisor",
  "justification": "Market opportunity - AAPL dip below target price"
}

Response:
{
  "approval_id": "apr_12345",
  "status": "pending",
  "approvers": [
    {"user": "financial_controller", "status": "pending"},
    {"user": "cfo", "status": "pending"}
  ],
  "expires_at": "2025-11-06T15:00:00Z"
}
```

---

## 10. Key Insights (DORA Alignment)

### 10.1 How This Addresses DORA Concerns

✅ **Clear AI Stance**: First-run wizard establishes explicit organizational AI policy, reducing ambiguity and friction

✅ **Runtime Enforcement**: Policy engine provides real-time feedback ("why did this fail?"), addressing DORA's platform UX concern

✅ **Visible Affordances**: UI clearly shows allowed/denied actions, approval requirements, and policy status

✅ **Safety Nets**: Approval workflows, risk scoring, and rollback requirements serve as AI-specific guardrails

✅ **Default to Lower Autonomy**: Recommends L1-L2 default, addressing DORA's finding that L0-L2 paths are critical for adoption

✅ **Governance Framework**: Defines clear boundaries between agency teams, platform teams, and compliance teams

### 10.2 Integration with DORA Metrics

The policy layer will instrument:
- **Change Failure Rate (CFR)**: Track policy violations as failures
- **Failed Deployment Recovery Time (FDRT)**: Measure approval workflow resolution time
- **Lead Time**: Policy evaluation time contributes to overall lead time
- **Deployment Frequency**: Policy friction may throttle frequency (to be monitored)
- **Rework Rate**: Policy violations requiring rework tracked separately

---

## Conclusion

The AI Policy Layer transforms CodeValdCortex from an orchestration platform to a **governed AI platform** that meets DORA's requirements for:
1. Clear organizational AI stance
2. Runtime safety nets
3. Transparent policy enforcement
4. Developer-friendly feedback loops
5. Compliance-ready governance

This positions CodeValdCortex as an enterprise-grade platform that can credibly deliver on its "Kubernetes of AI Agents" promise while addressing the specific concerns raised in DORA's 2024/2025 research.
