# How to Ship a Clear AI Policy Layer - Quick Guide

## TL;DR

Ship a clear AI policy layer in **3 phases over 6 weeks**:

1. **Week 1-2**: First-run wizard + policy schema + basic enforcement
2. **Week 3-4**: Runtime enforcement + approval workflows + risk scoring  
3. **Week 5-6**: Advanced features + compliance reporting + PII handling

**Key Insight**: The policy layer isn't just a feature—it's the **governance backbone** that transforms CodeValdCortex from "AI orchestration tool" to "enterprise-grade governed AI platform."

---

## What Is the AI Policy Layer?

A runtime governance system that:
- ✅ Establishes **explicit organizational AI stance** (DORA requirement)
- ✅ Enforces **autonomy boundaries** (L0-L4 levels)
- ✅ Provides **real-time feedback** ("why was this blocked?")
- ✅ Enables **approval workflows** for high-risk actions
- ✅ Creates **audit trails** for compliance

Think: **"RBAC + Risk Scoring + Budget Controls + Approval Workflows = AI Policy Layer"**

---

## Why DORA Demands This

DORA's 2024/2025 research found:

> "Teams excel when the org's AI policy is clear and socialized; ambiguity creates both under- and over-use."

**Without Policy Layer**:
- ❌ Inconsistent AI usage across teams
- ❌ Compliance violations (no guardrails)
- ❌ Friction from ambiguity ("Can I use this model?")
- ❌ No audit trail for AI decisions
- ❌ Black box operations (no trust)

**With Policy Layer**:
- ✅ Explicit rules ("These models are approved")
- ✅ Runtime enforcement (blocked before violation)
- ✅ Clear feedback ("Action requires L2 autonomy")
- ✅ Complete audit trail (who, what, when, why)
- ✅ Trust through transparency

---

## The 6-Step Wizard (User's First Experience)

Every new agency goes through this **10-minute wizard**:

### Step 1: Industry & Compliance
```
Question: What industry are you in?
→ Pre-loads relevant compliance frameworks

Question: What frameworks apply?
☐ SOC 2   ☐ HIPAA   ☐ GDPR   ☐ ISO 27001
→ Configures required controls
```

### Step 2: AI Adoption Stance
```
Choose your organization's AI philosophy:

🐢 Conservative: AI recommends, humans execute (L0-L1)
🚶 Controlled: AI does low-risk, humans approve high-risk (L1-L2) [RECOMMENDED]
🏃 Progressive: AI mostly autonomous, humans for edge cases (L2-L3)
🚀 Innovative: AI fully autonomous, humans audit after (L3-L4)

→ Sets default autonomy levels
```

### Step 3: Model Approval
```
Which AI providers can your org use?
☐ OpenAI (GPT-4, GPT-3.5)
☐ Anthropic (Claude)
☐ Azure OpenAI (Enterprise SLA)

For each:
- Data residency: [US / EU / UK]
- Monthly budget: $______
- Max tokens/request: ______

→ Creates allowlist, budget tracking
```

### Step 4: Data Classification
```
How sensitive is your data?
☐ We handle PII → Enable detection & masking
☐ Geographic restrictions → No cross-border transfer
☐ Industry data rules → HIPAA / GDPR constraints

→ Configures data access policies
```

### Step 5: Approval Workflows
```
Who approves high-risk AI actions?

Roles to configure:
- Operations Manager: [email]
- Security Officer: [email]
- Compliance Officer: [email]

What requires approval?
☐ Database modifications
☐ Financial transactions > $______
☐ Confidential data access

→ Sets up approval chains
```

### Step 6: Review & Generate
```
Summary:
  Industry: Financial Services
  Compliance: SOC 2, GDPR
  Stance: Controlled (L1-L2 default)
  Approved Models: OpenAI, Anthropic
  Budget: $50k/month
  Approvals: DB changes, Financial >$10k

[Generate Policy] ← Creates YAML, saves to database
```

---

## Runtime Enforcement (How It Works)

Every agent action goes through the **Policy Engine**:

```
┌─────────────────────────┐
│ Agent: "Buy 1000 shares │
│ of AAPL for $175k"      │
└───────────┬─────────────┘
            │
            ▼
┌───────────────────────────────────────────┐
│ Policy Engine Checks:                     │
│                                            │
│ 1. ✅ Model allowed? (OpenAI GPT-4)       │
│ 2. ✅ Autonomy sufficient? (Needs L1)     │
│ 3. ❌ Amount > $10k threshold!            │
│ 4. ❌ Requires approval!                  │
│                                            │
│ Risk Score: 65/100 (Medium)               │
└───────────┬───────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────┐
│ BLOCKED - Approval Required             │
│                                          │
│ Reason: Financial transaction >$10k     │
│ Approvers Needed:                       │
│   - Financial Controller                │
│   - CFO                                 │
│ Timeout: 30 minutes                     │
│                                          │
│ [Request Approval] [Modify] [Cancel]    │
└─────────────────────────────────────────┘
```

**Key**: Agent gets **immediate, actionable feedback**. No silent failures. No confusion.

---

## The 5 Policy Dimensions

### 1. Model Access Policy
```yaml
models:
  allowed_providers:
    - provider: "openai"
      models: ["gpt-4"]
      data_residency: "US"
      monthly_budget_usd: 50000
```

**Enforces**: Which AI models can be used, where data goes, spending limits

### 2. Autonomy Level Policy
```yaml
autonomy:
  default_level: "L1"  # Assisted
  role_overrides:
    - role: "data_collector"
      level: "L2"       # Conditional autonomy
    - role: "trade_executor"
      level: "L0"       # Manual only
```

**Enforces**: How much independence each role has, escalation rules

### 3. Data Access Policy
```yaml
data_access:
  rules:
    - classification: "confidential"
      allowed_operations: ["read"]
      requires_approval: true
      audit_all_access: true
```

**Enforces**: What data agents can access, PII handling, retention rules

### 4. Action Authorization Policy
```yaml
actions:
  approval_workflows:
    - action_pattern: "financial_transaction"
      approval_threshold:
        - amount: 10000
          approvers: ["financial_controller", "cfo"]
```

**Enforces**: Which actions need approval, who approves, timeouts

### 5. Risk & Compliance Policy
```yaml
risk:
  thresholds:
    high: 85
  mitigation_requirements:
    high:
      - human_review: true
      - enhanced_monitoring: true

compliance:
  frameworks: ["SOC2", "GDPR"]
  audit_requirements:
    log_all_actions: true
    retention_years: 7
```

**Enforces**: Risk scoring, compliance controls, audit requirements

---

## Implementation: 3 MVPs

### MVP-048: Foundation (2 weeks)
**What**: Wizard + Schema + Basic Enforcement

**Delivers**:
- ✅ 6-step policy wizard
- ✅ Policy YAML schema
- ✅ ArangoDB collections
- ✅ Model allowlist enforcement
- ✅ Autonomy level enforcement
- ✅ UI indicators ("Policy Compliant" badges)

**Files Created**:
```
/internal/policy/
├── engine.go          # Core policy evaluation
├── repository.go      # Database operations
├── types.go           # Data structures
└── evaluator.go       # Rule evaluation

/internal/web/handlers/policy/
├── wizard.go          # Wizard flow
└── policy_crud.go     # Policy management

/internal/web/templates/policy/
└── wizard_*.templ     # Wizard UI
```

---

### MVP-049: Runtime Enforcement (2 weeks)
**What**: Approval Workflows + Risk Scoring + Auditing

**Delivers**:
- ✅ Approval workflow engine
- ✅ Risk scoring calculator
- ✅ Budget tracking per model
- ✅ Policy violation logging
- ✅ Real-time alerts
- ✅ Approval UI (request/approve/deny)

**User Experience**:
```
Agent tries high-risk action
  ↓
Policy blocks it
  ↓
Shows: "Why blocked" + "How to fix"
  ↓
One-click approval request
  ↓
Approver notified immediately
  ↓
Approves/denies with reason
  ↓
Agent proceeds or receives feedback
```

---

### MVP-050: Advanced Features (2 weeks)
**What**: PII Detection + Compliance Reporting + Versioning

**Delivers**:
- ✅ Data classification engine
- ✅ PII detection (regex + ML)
- ✅ Automated masking
- ✅ Compliance reports (SOC2, GDPR)
- ✅ Policy versioning (audit changes)
- ✅ Multi-policy inheritance

---

## UI Examples

### Policy Status (Always Visible)
```
┌──────────────────────────────────────┐
│ 🛡️ Policy Status                     │
├──────────────────────────────────────┤
│ ✅ Compliant                          │
│                                       │
│ 🧠 Models: GPT-4, Claude              │
│ 🔒 Data: Internal & Public only       │
│ 💰 Budget: $2,400 / $5,000 (48%)     │
│ 🎚️ Autonomy: L1 - Assisted           │
└──────────────────────────────────────┘
```

### Blocked Action Feedback
```
┌──────────────────────────────────────┐
│ ⚠️ Action Blocked                     │
├──────────────────────────────────────┤
│ Reason:                               │
│ Financial transaction exceeds $10k    │
│                                       │
│ Requirements:                         │
│ • Approval from Financial Controller  │
│ • Approval from CFO                   │
│ • Dual approval policy                │
│                                       │
│ Risk Score: 65/100 (Medium)           │
│                                       │
│ [Request Approval]  [Modify Action]   │
└──────────────────────────────────────┘
```

### Approval Request (for Approvers)
```
┌──────────────────────────────────────┐
│ 🔔 Approval Request                   │
├──────────────────────────────────────┤
│ Agent: Trade Executor                 │
│ Action: Buy 1000 shares AAPL          │
│ Amount: $175,450                      │
│ Risk: 65/100 (Medium)                 │
│                                       │
│ Justification:                        │
│ "Market opportunity - AAPL below      │
│  target price, strong Q4 forecast"    │
│                                       │
│ Policy: Transactions >$10k require    │
│ dual approval per SOC2 controls       │
│                                       │
│ [✅ Approve]  [❌ Deny]  [✏️ Comment] │
└──────────────────────────────────────┘
```

---

## DORA Alignment (Why This Matters)

| DORA Finding | How AI Policy Layer Addresses It |
|--------------|----------------------------------|
| "Clear AI stance reduces friction" | First-run wizard establishes explicit policy |
| "Ambiguity creates under/over-use" | Runtime enforcement removes ambiguity |
| "Platform UX is weakest point" | Clear feedback: "Why blocked + how to fix" |
| "Safety nets needed for AI" | Approval workflows, risk scoring, rollback |
| "Most teams use L0-L2, not L3-L4" | Wizard defaults to L1 (Controlled stance) |
| "Governance unclear in multi-team" | Policy defines boundaries and responsibilities |

---

## Success Metrics

After shipping AI Policy Layer, measure:

### Adoption Metrics
- ✅ 100% of new agencies complete wizard
- ✅ <10 minutes average wizard completion time
- ✅ <5% wizard abandonment rate

### Enforcement Metrics
- ✅ Policy evaluation latency <50ms
- ✅ 0 unhandled policy violations
- ✅ 100% of blocks include remediation guidance

### Business Metrics
- ✅ Compliance certification time reduced
- ✅ Policy violation incidents decreased
- ✅ Developer satisfaction (NPS) improved
- ✅ Audit readiness rating: "Excellent"

---

## Bottom Line

**Shipping a clear AI policy layer means**:

1. ✅ **10-minute wizard** establishes organizational AI stance
2. ✅ **Runtime enforcement** blocks violations before they happen
3. ✅ **Clear feedback** tells users why and how to fix
4. ✅ **Approval workflows** for high-risk actions
5. ✅ **Audit trails** for compliance

**Result**: Transform CodeValdCortex from "AI tool" to "governed AI platform" that enterprises can trust and scale.

**Timeline**: 6 weeks to ship all 3 phases  
**Priority**: P0 (blocking enterprise adoption)  
**DORA Impact**: Directly addresses 4 of 7 critical gaps

---

## Next Steps

1. ⚡ **This week**: Create MVP-048 branch, implement wizard
2. ⚡ **Next week**: Implement policy engine and basic enforcement
3. 📋 **Week 3**: Begin MVP-049 (runtime enforcement)
4. 📋 **Week 5**: Begin MVP-050 (advanced features)
5. 📋 **Week 7**: Public launch with documentation

**Questions?** See full specification: `/documents/2-SoftwareDesignAndArchitecture/ai-policy-layer.md`

---

**Remember**: This isn't "compliance theater"—it's the governance layer that makes AI agents safe, auditable, and scalable in enterprise environments. That's what DORA demands, and what the market will pay for.
