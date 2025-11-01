# MVP - Minimum Viable Product Task Breakdown

## Task Overview
- **Objective**: Define and execute the minimum set of tasks required to launch a functional product that delivers core value to users
- **Success Criteria**: Deployable system with essential features that satisfies primary user needs and business objectives
- **Dependencies**: Infrastructure foundation and core technical architecture decisions

## Foundation Tasks (P0 - Blocking)

*All foundation tasks completed. See `mvp_done.md` for details.*

## Core Agent Mechanics (P0 - Blocking)

*All core agent mechanics tasks completed. See `mvp_done.md` for details.*

## Core Functionality Tasks (P1 - Critical)

*All core functionality tasks completed. See `mvp_done.md` for details.*

## Platform Integration Tasks (P1 - Critical)

| Task ID | Title                 | Description                                                                        | Status      | Priority | Effort | Skills Required         | Dependencies |
| ------- | --------------------- | ---------------------------------------------------------------------------------- | ----------- | -------- | ------ | ----------------------- | ------------ |
| MVP-014 | Kubernetes Deployment | Create Kubernetes manifests and Helm charts for agent deployment                   | Not Started | P1       | High   | DevOps, Kubernetes      | MVP-010      |
| MVP-015 | Management Dashboard  | Build web interface with Templ+HTMX+Alpine.js for agent monitoring, real-time updates, and control | In Progress | P1       | Medium | Go, Frontend Dev, Templ | MVP-013      |
| MVP-023 | AI Agent Creator      | Implement AI-powered conversational interface for creating agents. AI asks questions, resolves details, and generates complete agent configurations through natural language dialogue | Not Started | P1       | Medium | Go, Templ, AI/LLM, Frontend Dev | MVP-025      |
| MVP-030 | Work Items Core Schema & Registry | Implement work item types registry, JSON schemas, and extend agent types with taxonomy fields (autonomy, budget, safety, identity) | Not Started | P1       | Medium | Go, ArangoDB, JSON Schema | MVP-029      |
| MVP-031 | Work Items Lifecycle & SLA | Implement state machine, timers, breach handlers, and SLA/SLO enforcement for work items | Not Started | P1       | Medium | Go, ArangoDB, Backend Dev | MVP-030      |
| MVP-032 | Work Items Assignment & Routing | Build declarative routing rules engine, skill matching, and agent selection algorithms | Not Started | P1       | Medium | Go, ArangoDB, Backend Dev | MVP-031      |
| MVP-033 | Agent Lifecycle FSM | Implement agent lifecycle states (Registered, Scheduled, Starting, Healthy, Degraded, Backoff, Draining, Quarantined, Stopped, Retired) with transitions, guards, and health probes | Not Started | P1       | High   | Go, Backend Dev, Health Checks | MVP-032      |
| MVP-034 | Run Execution FSM | Implement run states (Pending, Running, Waiting I/O, Waiting HITL, Succeeded, Failed, Compensating, Compensated, Orphaned) with retry/backoff logic | Not Started | P1       | High   | Go, Backend Dev, State Machine | MVP-033      |
| MVP-035 | Health & Circuit Breakers | Implement health probe framework (HTTP, TCP, exec, gRPC), circuit breaker integration, and degradation detection | Not Started | P1       | Medium | Go, Backend Dev, Monitoring | MVP-034      |
| MVP-036 | Quarantine System | Implement quarantine triggers, evidence capture, triage workflow, and re-enablement approval process | Not Started | P1       | Medium | Go, Security, Backend Dev | MVP-035      |
| MVP-037 | Deployment Rollouts | Implement blue-green, canary, and progressive delivery strategies with SLO-based rollback | Not Started | P1       | High   | Go, DevOps, Deployment | MVP-036      |
| MVP-038 | Namespace Isolation | Implement namespace hierarchy, resource quotas, network policies, and noisy neighbor protections | Not Started | P1       | High   | Go, Kubernetes, Networking | MVP-037      |
| MVP-039 | Organization & RBAC | Build org/BU/project hierarchy, role matrix, permission system, and approval chain engine | Not Started | P1       | High   | Go, Security, Backend Dev | MVP-038      |
| MVP-040 | Billing & Metering | Implement metering for all billing dimensions (agent-hours, storage, network, audit), cost allocation, and budget tracking | Not Started | P1       | Medium | Go, Backend Dev, Analytics | MVP-039      |
| MVP-041 | Multi-tenancy Hardening | Add advanced isolation (dedicated nodes, encryption), data residency controls, and compliance reporting | Not Started | P2       | Medium | Go, Security, Compliance | MVP-040      |

## Authentication & Security Tasks (P2 - Important)

| Task ID | Title                     | Description                                                | Status      | Priority | Effort | Skills Required       | Dependencies |
| ------- | ------------------------- | ---------------------------------------------------------- | ----------- | -------- | ------ | --------------------- | ------------ |
| MVP-026 | Basic User Authentication | Implement user registration, login, and session management | Not Started | P2       | Medium | Backend Dev, Security | MVP-014      |
| MVP-027 | Security Implementation   | Add input validation, HTTPS, and basic security headers    | Not Started | P2       | Medium | Security, Backend Dev | MVP-026      |
| MVP-028 | Access Control System     | Implement role-based access control for agent operations   | Not Started | P2       | Low    | Backend Dev, Security | MVP-027      |

## Agency Designer (P1 - Critical)

| Task ID | Title                     | Description                                                | Status      | Priority | Effort | Skills Required       | Dependencies |
| ------- | ------------------------- | ---------------------------------------------------------- | ----------- | -------- | ------ | --------------------- | ------------ |
| MVP-043 | Work Items UI Module      | Build complete Work Items management UI with CRUD operations, AI refinement, templates, deliverables/dependencies management, and filtering | Not Started | P1       | Medium | Go, Templ, Frontend Dev, HTMX | MVP-029      |
| MVP-044 | Agent Types UI Module     | Build Agent Types definition and management UI with type catalog, capability specification, taxonomy fields, and templates | Not Started | P1       | Medium | Go, Templ, Frontend Dev | MVP-043      |
| MVP-045 | RACI Matrix UI Editor     | Build interactive RACI matrix editor with role assignments, validation, templates, and visual matrix interface | Not Started | P1       | Medium | Go, Templ, Frontend Dev | MVP-044      |
| MVP-042 | AI-Powered Agency Creator | Implement AI-driven agency creation flow with text upload, selective generation (introduction, goals, work items, agent types, RACI), and batch AI generation | Not Started | P1       | High   | Go, Templ, AI/LLM, Frontend Dev | MVP-045      |

## Agency Management Feature (P1 - Critical)

*All agency management tasks completed. See `mvp_done.md` for details.*

---

### MVP-023: Create Agent Form

**Objective**: Implement an AI-powered conversational interface for creating new agents within an agency context. The AI assistant will ask intelligent questions, resolve ambiguities, suggest configurations, and guide users through agent creation with natural language interaction.

**Key Deliverables**:

1. **AI Conversational Interface (Chat-based Agent Creator)**:
   - Chat-style modal with AI assistant avatar
   - Natural language input for agent specifications
   - AI asks clarifying questions about:
     * Agent purpose and role
     * Agent type selection (with suggestions)
     * Required capabilities and behaviors
     * Configuration parameters
     * Location and metadata
   - Real-time suggestions and validation
   - Preview of agent configuration before creation
   - Context-aware recommendations based on agency type

2. **AI Agent Creation Flow**:
   ```
   User: "I need an agent to monitor water pressure"
   
   AI: "Great! I'll help you create a pressure monitoring agent. 
        Let me ask a few questions:
        
        1. What should we name this agent? 
           (Suggestion: Pressure-Monitor-Zone-A)"
   
   User: "Pressure Monitor A1"
   
   AI: "Perfect! Now, for pressure monitoring, I recommend a 
        'sensor' type agent. It can:
        - Monitor pressure readings continuously
        - Send alerts when thresholds are exceeded
        - Log historical data
        
        Should I configure it with these capabilities?"
   
   User: "Yes, and alert me if pressure drops below 30 PSI"
   
   AI: "Understood! I've configured:
        ✓ Agent Type: Sensor
        ✓ Name: Pressure-Monitor-A1
        ✓ Alert Threshold: < 30 PSI
        ✓ Monitoring: Continuous
        
        What location should I assign? (e.g., Zone A, Building 3)"
   ```

3. **AI Intelligence Features**:
   - **Intent Recognition**: Understand user's agent requirements from natural language
   - **Smart Suggestions**: Recommend agent types based on description
   - **Configuration Inference**: Auto-configure based on use case
   - **Validation**: Catch incomplete or conflicting requirements
   - **Learning**: Improve suggestions based on agency patterns
   - **Multi-turn Dialogue**: Handle complex configurations through conversation
   - **Ambiguity Resolution**: Ask clarifying questions when needed

4. **Backend AI Integration**:
   - LLM integration (OpenAI GPT-4, Claude, or local model)
   - Agent type knowledge base for recommendations
   - Configuration templates for common scenarios
   - Prompt engineering for agent creation domain
   - Context injection (agency info, existing agents)
   - Structured output generation (JSON config)

5. **Agent Configuration Generation**:
   - AI generates complete agent configuration from conversation
   - Validates against agency database schema
   - Suggests defaults for optional fields
   - Creates unique names if conflicts detected
   - Maps natural language to technical parameters

6. **API Endpoints**:
   ```
   POST /api/v1/agencies/{id}/agents/chat         # Send message to AI
   POST /api/v1/agencies/{id}/agents/create       # Create from AI-generated config
   GET  /api/v1/agencies/{id}/agents/suggestions  # Get AI suggestions
   ```

**UI Mockup Structure**:
```html
<!-- AI Agent Creator Chat Interface -->
<div class="modal is-active">
  <div class="modal-background"></div>
  <div class="modal-card agent-chat-creator">
    <header class="modal-card-head">
      <div class="media">
        <div class="media-left">
          <span class="icon is-large">🤖</span>
        </div>
        <div class="media-content">
          <p class="modal-card-title">AI Agent Creator</p>
          <p class="subtitle is-7">Let me help you create the perfect agent</p>
        </div>
      </div>
      <button class="delete" aria-label="close"></button>
    </header>
    
    <section class="modal-card-body chat-container">
      <!-- Chat Messages -->
      <div class="chat-messages">
        <!-- AI Welcome Message -->
        <div class="message ai-message">
          <div class="message-avatar">🤖</div>
          <div class="message-content">
            <p>Hi! I'm your AI assistant for creating agents in the 
               <strong>Water Distribution Network</strong> agency.</p>
            <p>Tell me what kind of agent you need, and I'll help you 
               configure it. For example:</p>
            <ul>
              <li>"Create a sensor to monitor pipe pressure"</li>
              <li>"I need an agent to control pump operations"</li>
              <li>"Set up a leak detection agent for Zone A"</li>
            </ul>
          </div>
        </div>
        
        <!-- User Message Example -->
        <div class="message user-message">
          <div class="message-content">
            I need a sensor to monitor water flow
          </div>
          <div class="message-avatar">👤</div>
        </div>
        
        <!-- AI Response with Suggestions -->
        <div class="message ai-message">
          <div class="message-avatar">🤖</div>
          <div class="message-content">
            <p>Perfect! I'll create a flow monitoring sensor for you.</p>
            <div class="suggestion-box">
              <strong>Suggested Configuration:</strong>
              <ul>
                <li>📊 Agent Type: Flow Sensor</li>
                <li>⏱️ Update Interval: 30 seconds</li>
                <li>🎯 Measurement Units: Gallons per minute</li>
              </ul>
            </div>
            <p>What should we name this agent?</p>
            <div class="quick-replies">
              <button class="button is-small is-outlined">
                Flow-Sensor-01
              </button>
              <button class="button is-small is-outlined">
                Main-Flow-Monitor
              </button>
              <button class="button is-small is-outlined">
                Custom name...
              </button>
            </div>
          </div>
        </div>
        
        <!-- Configuration Preview -->
        <div class="message ai-message">
          <div class="message-avatar">🤖</div>
          <div class="message-content">
            <p>✅ All set! Here's your agent configuration:</p>
            <div class="config-preview">
              <div class="columns is-mobile">
                <div class="column">
                  <strong>Name:</strong><br>Flow-Sensor-Zone-A
                </div>
                <div class="column">
                  <strong>Type:</strong><br>Flow Sensor
                </div>
              </div>
              <div class="columns is-mobile">
                <div class="column">
                  <strong>Location:</strong><br>Zone A, Sector 3
                </div>
                <div class="column">
                  <strong>Interval:</strong><br>30 seconds
                </div>
              </div>
            </div>
            <p>Should I create this agent now?</p>
            <div class="action-buttons">
              <button class="button is-success">
                ✓ Create Agent
              </button>
              <button class="button is-warning">
                ✏️ Modify Configuration
              </button>
              <button class="button is-light">
                ↩️ Start Over
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Typing Indicator -->
      <div class="typing-indicator" style="display: none;">
        <span></span><span></span><span></span>
      </div>
    </section>
    
    <!-- Chat Input -->
    <footer class="modal-card-foot chat-input-container">
      <div class="field has-addons is-fullwidth">
        <div class="control is-expanded">
          <input class="input" type="text" 
                 placeholder="Describe the agent you need..."
                 id="chat-input">
        </div>
        <div class="control">
          <button class="button is-primary" id="send-message">
            <span class="icon">📤</span>
          </button>
        </div>
      </div>
      <p class="help has-text-centered">
        💡 Tip: Be specific about what you want the agent to do
      </p>
    </footer>
  </div>
</div>
```

**AI Conversation Patterns**:

1. **Intent Recognition**:
   ```
   User: "monitor temperature"
   AI detects: monitoring + temperature → Sensor agent
   ```

2. **Clarifying Questions**:
   ```
   User: "create a pump agent"
   AI: "Should this pump be automatic or manual control?"
   ```

3. **Smart Defaults**:
   ```
   User: "leak detector"
   AI: Auto-configures threshold, sensitivity, alert settings
   ```

4. **Conflict Resolution**:
   ```
   User: "name it Sensor-A1"
   AI: "That name exists. How about Sensor-A1-v2?"
   ```

**Acceptance Criteria**:
- ✅ AI understands natural language agent descriptions
- ✅ Asks intelligent clarifying questions
- ✅ Suggests appropriate agent types
- ✅ Generates valid agent configurations
- ✅ Handles multi-turn conversations
- ✅ Validates and prevents conflicts
- ✅ Creates agents successfully from AI-generated config
- ✅ Provides helpful suggestions and examples
- ✅ Gracefully handles ambiguous requests
- ✅ Learns from agency-specific patterns

**Technical Implementation**:
```
/workspaces/CodeValdCortex/
├── internal/ai/
│   ├── agent_creator.go              # AI agent creation logic
│   ├── llm_client.go                 # LLM API integration
│   ├── prompts/
│   │   ├── agent_creation.txt        # Prompt templates
│   │   └── configuration.txt         # Config generation prompts
│   └── knowledge/
│       ├── agent_types.json          # Agent type knowledge base
│       └── templates.json            # Configuration templates
├── internal/web/
│   ├── pages/
│   │   └── ai_agent_creator.templ    # Chat interface
│   ├── handlers/
│   │   └── ai_agent_handler.go       # AI chat endpoints
│   └── components/
│       └── chat_message.templ        # Message components
└── static/
    ├── css/
    │   └── ai-chat.css               # Chat UI styles
    └── js/
        ├── ai-chat.js                # Chat interactions
        └── streaming-response.js     # Real-time AI responses
```

**AI Prompt Engineering**:
```
System Prompt:
You are an expert AI assistant for creating agents in the CodeValdCortex 
multi-agent system. Your role is to help users create perfectly configured 
agents through natural conversation.

Agency Context: {agency_name} - {agency_category}
Available Agent Types: {agent_types}
Existing Agents: {agent_count}

Guidelines:
1. Ask clarifying questions for ambiguous requests
2. Suggest appropriate agent types based on description
3. Provide sensible defaults for optional fields
4. Validate against existing agent names
5. Be concise but helpful
6. Use emojis to make responses engaging
7. Always confirm before creating the agent

When ready to create, output JSON:
{
  "name": "agent-name",
  "type": "agent-type",
  "config": {...},
  "metadata": {...}
}
```

**API Specification**:
```json
POST /api/v1/agencies/{agency_id}/agents/chat
Content-Type: application/json

Request:
{
  "message": "I need a sensor to monitor pipe pressure in Zone A",
  "conversation_id": "conv-uuid-123",
  "context": {
    "previous_messages": [...],
    "draft_config": {...}
  }
}

Response (200 OK):
{
  "message": "Great! I'll help you create a pressure monitoring sensor...",
  "suggestions": ["Pressure-Sensor-Zone-A", "Zone-A-Pressure-01"],
  "questions": ["What pressure threshold should trigger alerts?"],
  "draft_config": {
    "type": "sensor",
    "category": "pressure",
    "location": "Zone A"
  },
  "confidence": 0.95,
  "ready_to_create": false
}

POST /api/v1/agencies/{agency_id}/agents/create
{
  "config": {...},  // AI-generated configuration
  "conversation_id": "conv-uuid-123"
}
```

**Dependencies**:
- MVP-022 (agency selection and context)
- LLM API access (OpenAI, Claude, or local model)
- Agent type registry populated
- Streaming response support for real-time AI feedback

**Future Enhancements**:
- Voice input for agent creation
- Visual agent builder (AI-assisted drag-and-drop)
- Bulk agent creation from descriptions
- Agent cloning with AI modifications
- Multi-language support
- AI-powered agent optimization suggestions
- Learning from user corrections and preferences

---

### MVP-030: Work Items Core Schema & Registry

**Objective**: Implement work item types registry, JSON schemas for 6 work item types (Task, Job, Investigation, Change, Remediation, Experiment), and extend agent types schema with taxonomy fields from the comprehensive specifications.

**Key Deliverables**:

1. **Work Item Type Registry**:
   ```go
   type WorkItemType struct {
       TypeID          string                 `json:"type_id"`
       Name            string                 `json:"name"`
       Description     string                 `json:"description"`
       Schema          JSONSchema             `json:"schema"`
       DefaultValues   map[string]interface{} `json:"default_values"`
       RequiredFields  []string               `json:"required_fields"`
       AllowedStates   []string               `json:"allowed_states"`
       Version         string                 `json:"version"`
   }
   ```

2. **JSON Schemas for Work Item Types**:
   - **Task**: Short-lived, single-agent execution (hours to days)
   - **Job**: Multi-agent orchestration (days to weeks)
   - **Investigation**: Root cause analysis and evidence gathering
   - **Change**: Infrastructure/config changes with approvals
   - **Remediation**: Incident response and recovery actions
   - **Experiment**: A/B testing and hypothesis validation

3. **Agent Type Taxonomy Extension**:
   ```go
   type AgentType struct {
       TypeID           string           `json:"type_id"`
       Name             string           `json:"name"`
       Category         AgentCategory    `json:"category"`
       SkillsContract   SkillsContract   `json:"skills_contract"`
       AutonomyLevel    AutonomyLevel    `json:"autonomy_level"`
       Budget           BudgetLimits     `json:"budget"`
       DataBoundaries   DataBoundaries   `json:"data_boundaries"`
       SafetyConstraints SafetyConstraints `json:"safety_constraints"`
       Identity         IdentityConfig   `json:"identity"`
       TenantIsolation  IsolationPolicy  `json:"tenant_isolation"`
   }
   ```

4. **ArangoDB Collections**:
   - `work_item_types` collection
   - `agent_types` collection (extend existing)
   - `work_item_templates` collection
   - Create indexes and validation rules

5. **Backend Services**:
   - `WorkItemTypeRegistry` service
   - `AgentTypeRegistry` service (extend existing)
   - Schema validation service
   - Template management service

6. **API Endpoints**:
   ```
   GET    /api/v1/work-item-types
   GET    /api/v1/work-item-types/{typeId}
   POST   /api/v1/work-item-types
   PUT    /api/v1/work-item-types/{typeId}
   GET    /api/v1/agent-types
   GET    /api/v1/agent-types/{typeId}
   PUT    /api/v1/agent-types/{typeId}
   ```

**Acceptance Criteria**:
- [ ] All 6 work item types defined with complete JSON schemas
- [ ] Agent types extended with taxonomy fields (7 agent types)
- [ ] Registry services functional with CRUD operations
- [ ] Schema validation prevents invalid work item creation
- [ ] Default values and templates available
- [ ] Migration scripts for existing data structures
- [ ] Documentation for all schemas

**Dependencies**: MVP-029 (Goals Module)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items.md` and `agent-types-taxonomy.md`

---

### MVP-031: Work Items Lifecycle & SLA

**Objective**: Implement complete lifecycle state machine, timers, breach handlers, and SLA/SLO enforcement for work items.

**Key Deliverables**:

1. **State Machine Implementation**:
   - States: Planned, In-Progress, Waiting, Review, Done, Failed, Rolled-back
   - Transition validation and guards
   - State history tracking
   - Allowed transitions matrix

2. **SLA/SLO Fields & Enforcement**:
   ```go
   type WorkItemSLA struct {
       ResponseTimeMinutes   int       `json:"response_time_minutes"`
       CompletionTimeHours   int       `json:"completion_time_hours"`
       EscalationPolicy      Escalation `json:"escalation_policy"`
       BreachActions         []Action   `json:"breach_actions"`
       CreatedAt             time.Time  `json:"created_at"`
       FirstResponseAt       *time.Time `json:"first_response_at"`
       CompletedAt           *time.Time `json:"completed_at"`
   }
   ```

3. **Timer Service**:
   - Background timer for SLA monitoring
   - Breach detection and alerting
   - Escalation trigger execution
   - Timeout handling for waiting states

4. **Breach Actions**:
   - Auto-escalate to higher priority
   - Auto-retry with backoff
   - Trigger remediation work item
   - Notify stakeholders

5. **API Endpoints**:
   ```
   POST   /api/v1/work-items/{id}/transition
   GET    /api/v1/work-items/{id}/state-history
   GET    /api/v1/work-items/sla-breaches
   POST   /api/v1/work-items/{id}/escalate
   ```

**Acceptance Criteria**:
- [ ] State machine enforces valid transitions
- [ ] SLA timers track response and completion times
- [ ] Breach detection triggers configured actions
- [ ] Escalation policies execute correctly
- [ ] State history is complete and queryable
- [ ] Metrics collected for SLA compliance

**Dependencies**: MVP-030 (Work Items Core Schema & Registry)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items.md` Section 3

---

### MVP-032: Work Items Assignment & Routing

**Objective**: Build declarative routing rules engine with skill matching, cost optimization, and agent selection algorithms.

**Key Deliverables**:

1. **Routing Rules Engine**:
   ```go
   type RoutingRule struct {
       RuleID      string            `json:"rule_id"`
       WorkItemType string           `json:"work_item_type"`
       Conditions  []Condition       `json:"conditions"`
       AgentSelection AgentSelection `json:"agent_selection"`
       Priority    int               `json:"priority"`
   }
   ```

2. **Agent Selection Algorithms**:
   - Skills-based matching
   - Cost budget optimization
   - Data residency compliance
   - Load balancing across agents
   - Round-robin and least-loaded strategies

3. **Skill Matching**:
   ```go
   type SkillMatcher struct {
       RequiredSkills []Skill
       OptionalSkills []Skill
       SkillWeights   map[string]float64
   }
   ```

4. **Assignment Service**:
   - `WorkItemAssignmentService`
   - Rule evaluation engine
   - Agent availability checking
   - Fallback and retry logic

5. **API Endpoints**:
   ```
   POST   /api/v1/work-items/{id}/assign
   GET    /api/v1/work-items/{id}/candidate-agents
   POST   /api/v1/work-items/{id}/reassign
   GET    /api/v1/routing-rules
   POST   /api/v1/routing-rules
   ```

**Acceptance Criteria**:
- [ ] Routing rules evaluate correctly
- [ ] Skills matching finds qualified agents
- [ ] Cost budgets are respected
- [ ] Data residency rules enforced
- [ ] Load balancing works across agents
- [ ] Fallback mechanisms handle no-match scenarios

**Dependencies**: MVP-031 (Work Items Lifecycle & SLA)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items.md` Section 4

---

### MVP-033: Agent Lifecycle FSM

**Objective**: Implement comprehensive agent lifecycle finite state machine with 10 states, transitions, guards, timeouts, and health probes.

**Key Deliverables**:

1. **Agent Lifecycle States**:
   - Registered, Scheduled, Starting, Healthy, Degraded, Backoff, Draining, Quarantined, Stopped, Retired
   - State metadata and duration tracking
   - State transition validation

2. **Health Probes**:
   ```go
   type HealthProbe struct {
       Type            ProbeType   `json:"type"` // liveness, readiness
       Method          ProbeMethod `json:"method"` // http, tcp, exec, grpc
       Config          ProbeConfig `json:"config"`
       InitialDelay    int         `json:"initial_delay_ms"`
       Interval        int         `json:"interval_ms"`
       Timeout         int         `json:"timeout_ms"`
       SuccessThreshold int        `json:"success_threshold"`
       FailureThreshold int        `json:"failure_threshold"`
   }
   ```

3. **Timeouts & Heartbeats**:
   - Startup timeout configuration
   - Heartbeat monitoring
   - Drain timeout handling
   - Exponential backoff calculation

4. **Transition Guards**:
   - Validate preconditions before transitions
   - Check resource availability
   - Enforce approval requirements
   - Block invalid state changes

5. **API Endpoints**:
   ```
   GET    /api/v1/agents/{id}/state
   POST   /api/v1/agents/{id}/transition
   GET    /api/v1/agents/{id}/state-history
   GET    /api/v1/agents/{id}/health
   POST   /api/v1/agents/{id}/heartbeat
   ```

**Acceptance Criteria**:
- [ ] All 10 lifecycle states implemented
- [ ] State transitions follow FSM rules
- [ ] Health probes work for all probe types
- [ ] Timeouts trigger appropriate actions
- [ ] Heartbeat monitoring detects failures
- [ ] Guards prevent invalid transitions

**Dependencies**: MVP-032 (Work Items Assignment & Routing)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` Section 1

---

### MVP-034: Run Execution FSM

**Objective**: Implement run/task execution finite state machine with 9 states, retry/backoff logic, waiting states, and orphan recovery.

**Key Deliverables**:

1. **Run States**:
   - Pending, Running, Waiting I/O, Waiting HITL, Succeeded, Failed, Compensating, Compensated, Orphaned
   - Run metadata and execution context
   - State duration and timing metrics

2. **Retry & Backoff Logic**:
   ```go
   type RetryPolicy struct {
       MaxAttempts      int            `json:"max_attempts"`
       BackoffStrategy  string         `json:"backoff_strategy"` // fixed, exponential
       InitialDelay     int            `json:"initial_delay_ms"`
       MaxDelay         int            `json:"max_delay_ms"`
       Multiplier       float64        `json:"multiplier"`
       Jitter           bool           `json:"jitter"`
       RetryOn          []ErrorCategory `json:"retry_on"`
       DoNotRetryOn     []string       `json:"do_not_retry_on"`
   }
   ```

3. **Wait Conditions**:
   - I/O wait (external API, database)
   - HITL wait (human approval)
   - Dependency wait (other work items)
   - Rate limit wait

4. **Orphan Recovery**:
   - Heartbeat-based detection
   - State recovery from checkpoints
   - Idempotent operation retry
   - Manual review escalation

5. **Compensation/Saga Support**:
   ```go
   type CompensationPlan struct {
       CompensationSteps []CompensationStep `json:"compensation_steps"`
       Strategy          string             `json:"strategy"` // sequential, parallel
   }
   ```

6. **API Endpoints**:
   ```
   GET    /api/v1/runs/{id}/state
   POST   /api/v1/runs/{id}/transition
   POST   /api/v1/runs/{id}/retry
   POST   /api/v1/runs/{id}/compensate
   GET    /api/v1/runs/orphaned
   POST   /api/v1/runs/{id}/recover
   ```

**Acceptance Criteria**:
- [ ] All 9 run states implemented
- [ ] Retry policy with exponential backoff works
- [ ] Wait states handle timeouts correctly
- [ ] Orphan detection and recovery functional
- [ ] Compensation steps execute in order
- [ ] Run history and metrics collected

**Dependencies**: MVP-033 (Agent Lifecycle FSM)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` Section 2

---

### MVP-035: Health & Circuit Breakers

**Objective**: Implement health probe framework for multiple probe types and circuit breaker integration for external dependencies.

**Key Deliverables**:

1. **Health Probe Framework**:
   - HTTP probe (status endpoint)
   - TCP probe (port connectivity)
   - Exec probe (command execution)
   - gRPC probe (gRPC health check protocol)

2. **Circuit Breaker Service**:
   ```go
   type CircuitBreaker struct {
       Name       string          `json:"name"`
       State      CircuitState    `json:"state"` // closed, open, half_open
       Thresholds Thresholds      `json:"thresholds"`
       Timings    Timings         `json:"timings"`
       Metrics    CircuitMetrics  `json:"metrics"`
   }
   ```

3. **Degradation Detection**:
   - Performance degradation monitoring
   - Error rate threshold checking
   - Automatic transition to Degraded state
   - Recovery detection and restoration

4. **Integration Points**:
   - Database connection pools
   - External API clients
   - Message queue connections
   - Cache connections

5. **Monitoring Dashboard**:
   - Real-time circuit breaker status
   - Health probe results visualization
   - Degradation alerts and notifications

**Acceptance Criteria**:
- [ ] All 4 probe types functional
- [ ] Circuit breaker opens on threshold breach
- [ ] Half-open state tests recovery
- [ ] Degradation detection works correctly
- [ ] Agent transitions to Degraded when circuit opens
- [ ] Monitoring dashboard displays status

**Dependencies**: MVP-034 (Run Execution FSM)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` Sections 1.4-1.5

---

### MVP-036: Quarantine System

**Objective**: Implement comprehensive quarantine system with triggers, evidence capture, triage workflow, and re-enablement process.

**Key Deliverables**:

1. **Quarantine Triggers**:
   - Security violations (unauthorized access, credential exposure)
   - Policy violations (compliance breaches)
   - Anomaly detection (behavioral anomaly scores)
   - Resource abuse (CPU/memory/network)
   - Repeated failures (excessive error rates)

2. **Evidence Capture**:
   ```go
   type QuarantineEvidence struct {
       EvidenceID       string          `json:"evidence_id"`
       AgentState       AgentStateSnapshot `json:"agent_state"`
       RecentLogs       []LogEntry      `json:"recent_logs"`
       PerformanceMetrics Metrics       `json:"performance_metrics"`
       SecurityEvents   []SecurityEvent `json:"security_events"`
       StorageLocation  string          `json:"storage_location"`
   }
   ```

3. **Triage Workflow**:
   - Automated analysis and classification
   - Assignment to triage team
   - Manual investigation tools
   - Root cause determination
   - Remediation decision tracking

4. **Re-enablement Process**:
   - Checklist validation (root cause, remediation, testing, approvals)
   - Security and compliance approvals
   - Gradual rollout with monitoring
   - Rollback triggers for issues

5. **SLA Tracking**:
   - Response time by severity (Critical: 15min, High: 1hr, Medium: 4hr, Low: 24hr)
   - Resolution targets
   - Escalation automation

6. **API Endpoints**:
   ```
   POST   /api/v1/agents/{id}/quarantine
   GET    /api/v1/agents/{id}/quarantine/evidence
   POST   /api/v1/agents/{id}/quarantine/triage
   POST   /api/v1/agents/{id}/quarantine/re-enable
   GET    /api/v1/quarantine/dashboard
   ```

**Acceptance Criteria**:
- [ ] Quarantine triggers isolate agents correctly
- [ ] Evidence capture includes all required data
- [ ] Triage workflow assigns and tracks investigations
- [ ] Re-enablement requires all approvals
- [ ] SLA tracking monitors response times
- [ ] Post-mortem documentation generated

**Dependencies**: MVP-035 (Health & Circuit Breakers)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` Section 3

---

### MVP-037: Deployment Rollouts

**Objective**: Implement multiple deployment strategies (blue-green, canary, progressive delivery) with SLO-based automatic rollback.

**Key Deliverables**:

1. **Blue-Green Deployment**:
   - Instant traffic switchover
   - Warmup period for new version
   - Health check validation
   - Instant rollback capability

2. **Canary Deployment**:
   ```go
   type CanaryStage struct {
       Name            string          `json:"name"`
       Percentage      int             `json:"percentage"` // % traffic
       Duration        int             `json:"duration_ms"`
       SuccessCriteria SuccessCriteria `json:"success_criteria"`
   }
   ```

3. **Progressive Delivery**:
   - Start at 1% traffic
   - Increment by 10% every 5 minutes
   - Pause on error detection
   - Auto-rollback on failure

4. **SLO-Based Rollback**:
   ```go
   type RollbackTrigger struct {
       Metric           string  `json:"metric"`
       Operator         string  `json:"operator"`
       Threshold        float64 `json:"threshold"`
       EvaluationWindow int     `json:"evaluation_window_ms"`
   }
   ```

5. **Error Budget Integration**:
   - Track error budget consumption
   - Pause rollout when budget warning threshold reached
   - Auto-rollback when budget critical threshold reached

6. **Deployment Metrics**:
   - Error rate monitoring
   - Latency P95/P99 tracking
   - Success rate calculation
   - Custom metric evaluation

7. **API Endpoints**:
   ```
   POST   /api/v1/agents/{id}/deploy
   GET    /api/v1/deployments/{id}/status
   POST   /api/v1/deployments/{id}/promote
   POST   /api/v1/deployments/{id}/rollback
   GET    /api/v1/deployments/{id}/metrics
   ```

**Acceptance Criteria**:
- [ ] Blue-green deployments switch instantly
- [ ] Canary stages progress automatically
- [ ] Progressive delivery increments correctly
- [ ] Rollback triggers fire on threshold breach
- [ ] Error budget tracking prevents issues
- [ ] All deployment types support instant rollback

**Dependencies**: MVP-036 (Quarantine System)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` Section 4

---

### MVP-038: Namespace Isolation

**Objective**: Implement namespace hierarchy, resource quotas, network policies, and noisy neighbor protections for multi-tenant isolation.

**Key Deliverables**:

1. **Namespace Architecture**:
   ```go
   type Namespace struct {
       NamespaceID      string          `json:"namespace_id"`
       Type             NamespaceType   `json:"type"` // org, business_unit, project, environment
       OrganizationID   string          `json:"organization_id"`
       BusinessUnitID   string          `json:"business_unit_id,omitempty"`
       ProjectID        string          `json:"project_id,omitempty"`
       IsolationPolicy  IsolationPolicy `json:"isolation"`
       Quotas           ResourceQuotas  `json:"quotas"`
       NetworkPolicy    NetworkPolicy   `json:"network_policy"`
   }
   ```

2. **Resource Quotas**:
   - Compute: max agents, CPU cores, memory, GPUs
   - Storage: database size, artifact storage, backups
   - Network: egress/ingress bandwidth, connections
   - Work Items: active items, daily creation limit
   - API: rate limits, daily quotas

3. **Network Policies**:
   - Ingress rules (source namespaces, IPs, service accounts)
   - Egress rules (destination namespaces, IPs, domains)
   - DNS policy (allowed/blocked domains, DNSSEC)
   - Default allow/deny behavior

4. **Noisy Neighbor Protection**:
   ```go
   type NoisyNeighborProtection struct {
       Detection   DetectionConfig  `json:"detection"`
       Throttling  ThrottlingConfig `json:"throttling"`
       Fairness    FairnessPolicy   `json:"fairness"`
   }
   ```

5. **Quota Enforcement**:
   - Soft limits (warnings at 80%)
   - Hard limits (blocking at 100%)
   - Spillover to alternate namespaces
   - Alert notifications

6. **API Endpoints**:
   ```
   GET    /api/v1/namespaces
   POST   /api/v1/namespaces
   GET    /api/v1/namespaces/{id}/quotas
   PUT    /api/v1/namespaces/{id}/quotas
   GET    /api/v1/namespaces/{id}/network-policy
   PUT    /api/v1/namespaces/{id}/network-policy
   ```

**Acceptance Criteria**:
- [ ] Namespace hierarchy enforced
- [ ] Resource quotas prevent overuse
- [ ] Network policies block unauthorized access
- [ ] Noisy neighbor detection functional
- [ ] Throttling limits resource abuse
- [ ] Quota dashboard displays utilization

**Dependencies**: MVP-037 (Deployment Rollouts)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/multi-tenancy-org-model.md` Section 1

---

### MVP-039: Organization & RBAC

**Objective**: Build organizational hierarchy (Org → BU → Project → Environment), role-based access control with 8 standard roles, and multi-step approval chains.

**Key Deliverables**:

1. **Organizational Hierarchy**:
   ```go
   type Organization struct {
       OrganizationID  string         `json:"organization_id"`
       Name            string         `json:"name"`
       Domain          string         `json:"domain"`
       BusinessUnits   []BusinessUnit `json:"business_units"`
       Settings        OrgSettings    `json:"settings"`
       BillingAccount  BillingAccount `json:"billing_account"`
   }
   ```

2. **Role Matrix** (8 Standard Roles):
   - Organization Owner (full control)
   - Business Unit Lead (manage BU and projects)
   - Project Owner (full project control)
   - Developer (build and deploy agents)
   - Operator (operate and monitor)
   - Auditor (review and audit)
   - Risk Manager (manage risk policies)
   - Viewer (read-only access)

3. **Permission System**:
   ```go
   type Permission struct {
       Resource   ResourceType `json:"resource"`
       Actions    []Action     `json:"actions"`
       Conditions []Condition  `json:"conditions,omitempty"`
   }
   ```

4. **Approval Chains**:
   - Multi-step approval workflow
   - Role-based approvers
   - Dynamic approver resolution
   - Timeout and escalation
   - Veto rights for specific roles

5. **API Endpoints**:
   ```
   GET    /api/v1/organizations
   POST   /api/v1/organizations
   GET    /api/v1/organizations/{id}/business-units
   POST   /api/v1/organizations/{id}/business-units
   GET    /api/v1/projects
   POST   /api/v1/projects
   GET    /api/v1/roles
   POST   /api/v1/roles
   GET    /api/v1/approval-chains
   POST   /api/v1/approval-chains
   ```

**Acceptance Criteria**:
- [ ] Org hierarchy creation and management works
- [ ] All 8 roles implemented with correct permissions
- [ ] Permission checks enforce access control
- [ ] Approval chains execute multi-step workflows
- [ ] Timeout and escalation functional
- [ ] Role inheritance works across hierarchy

**Dependencies**: MVP-038 (Namespace Isolation)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/multi-tenancy-org-model.md` Section 2

---

### MVP-040: Billing & Metering

**Objective**: Implement comprehensive metering for all billing dimensions, cost allocation/chargebacks, and budget tracking with alerts.

**Key Deliverables**:

1. **Billing Dimensions**:
   - Agent-hours by agent type
   - CPU/memory/GPU hours
   - Storage (database, artifacts, backups, archive)
   - Network (egress/ingress by region)
   - Audit retention (logs, traces, metrics)

2. **Metering Service**:
   ```go
   type BillingDimensions struct {
       OrganizationID string         `json:"organization_id"`
       BillingPeriod  BillingPeriod  `json:"billing_period"`
       Compute        ComputeCosts   `json:"compute"`
       Storage        StorageCosts   `json:"storage"`
       Network        NetworkCosts   `json:"network"`
       Audit          AuditCosts     `json:"audit"`
       TotalCost      Cost           `json:"total_cost"`
   }
   ```

3. **Cost Allocation**:
   - By business unit
   - By project
   - By environment
   - By tags
   - Shared cost allocation (equal, proportional, weighted)

4. **Budget System**:
   ```go
   type Budget struct {
       BudgetID       string        `json:"budget_id"`
       Scope          BudgetScope   `json:"scope"`
       Amount         float64       `json:"amount"`
       Period         string        `json:"period"`
       Alerts         []BudgetAlert `json:"alerts"`
       Actions        []BudgetAction `json:"actions"`
       CurrentSpend   float64       `json:"current_spend"`
       ForecastedSpend float64      `json:"forecasted_spend"`
   }
   ```

5. **Pricing Models**:
   - OSS: Virtual "CodeVald Credits" for cost visibility
   - Commercial: Actual billing with tiered pricing

6. **Chargeback Reports**:
   - Executive summary
   - Detailed breakdown by hierarchy
   - Cost optimization recommendations
   - Export to CSV/PDF

7. **API Endpoints**:
   ```
   GET    /api/v1/billing/metrics
   GET    /api/v1/billing/allocation
   GET    /api/v1/budgets
   POST   /api/v1/budgets
   GET    /api/v1/budgets/{id}/alerts
   GET    /api/v1/billing/reports/chargeback
   ```

**Acceptance Criteria**:
- [ ] All billing dimensions metered accurately
- [ ] Cost allocation works for all hierarchy levels
- [ ] Budget alerts trigger at thresholds
- [ ] Budget actions execute (throttle, block)
- [ ] Chargeback reports generate correctly
- [ ] OSS credit system displays costs

**Dependencies**: MVP-039 (Organization & RBAC)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/multi-tenancy-org-model.md` Section 3

---

### MVP-041: Multi-tenancy Hardening

**Objective**: Add advanced isolation features, data residency controls, tenant migration tools, and compliance reporting dashboards.

**Key Deliverables**:

1. **Advanced Isolation**:
   - Dedicated compute nodes per tenant
   - CPU pinning and memory reservation
   - Dedicated storage volumes
   - Encryption at rest for all tenant data
   - Isolated backup storage

2. **Data Residency Controls**:
   ```go
   type DataResidency struct {
       AllowedRegions    []string `json:"allowed_regions"`
       PrimaryRegion     string   `json:"primary_region"`
       BackupRegions     []string `json:"backup_regions"`
       CrossBorderPolicy string   `json:"cross_border_policy"`
   }
   ```

3. **Tenant Migration Tools**:
   - Export tenant data
   - Import tenant data
   - Namespace migration
   - Zero-downtime migration
   - Migration validation

4. **Compliance Reporting**:
   - SOC2 compliance dashboard
   - HIPAA compliance dashboard
   - GDPR compliance dashboard
   - Audit trail reports
   - Data access logs

5. **API Endpoints**:
   ```
   POST   /api/v1/tenants/{id}/export
   POST   /api/v1/tenants/{id}/import
   POST   /api/v1/tenants/{id}/migrate
   GET    /api/v1/compliance/reports
   GET    /api/v1/compliance/{framework}/status
   ```

**Acceptance Criteria**:
- [ ] Dedicated nodes isolate tenant workloads
- [ ] Data residency rules enforced
- [ ] Tenant migration completes successfully
- [ ] Compliance dashboards show status
- [ ] Audit trails complete and queryable
- [ ] Encryption at rest functional

**Dependencies**: MVP-040 (Billing & Metering)

**Reference**: See `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/multi-tenancy-org-model.md` Section 4

---

**Note**: The original MVP-030, MVP-031, MVP-032, and MVP-033 task descriptions have been superseded by comprehensive specifications documented in:
- `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items.md`
- `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-types-taxonomy.md`
- `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md`
- `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/multi-tenancy-org-model.md`
- `/documents/2-SoftwareDesignAndArchitecture/agency-operation-framework/goals-specification.md`

The new MVP tasks (MVP-030 through MVP-041) implement these comprehensive specifications with proper phasing and dependencies.

---

## Agent Property Broadcasting Feature (P1 - Critical)

*Enables UC-TRACK-001 (Safiri Salama) and other real-time tracking/monitoring use cases*

| Task ID | Title                                    | Description                                                                                                      | Status      | Priority | Effort | Skills Required            | Dependencies |
| ------- | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | ----------- | -------- | ------ | -------------------------- | ------------ |
| MVP-016 | Core Broadcasting Infrastructure         | Implement BroadcastConfiguration, PropertyBroadcaster service, ContextEvaluator, and integration with PubSub    | Not Started | P1       | High   | Go, Backend Dev, PubSub    | MVP-013      |
| MVP-017 | Subscription Management                  | Build SubscriptionManager, subscriber filtering, favorite functionality, and subscription API endpoints          | Not Started | P1       | Medium | Go, Backend Dev, REST API  | MVP-016      |
| MVP-018 | Privacy & Security Controls              | Implement geofencing, property masking, permission model, audit logging, and encryption for sensitive properties | Not Started | P1       | Medium | Security, Backend Dev      | MVP-017      |
| MVP-019 | Performance Optimization & Scale         | Performance tuning, caching, load balancing for broadcasters, message broker optimization, monitoring & alerting | Not Started | P1       | Medium | Performance, DevOps        | MVP-018      |
| MVP-020 | UC-TRACK-001 Integration & Testing       | Implement Vehicle & Passenger agents, build mobile/web UI, SACCO management portal, end-to-end testing          | Not Started | P1       | High   | Full-stack, Mobile Dev     | MVP-019      |

**Note**: Detailed task specifications for MVP-016 through MVP-020 are available in `/documents/3-SofwareDevelopment/core-systems/agent-property-broadcasting.md`

---

## Authentication & Security Tasks (P2 - Important)
   - **Work Breakdown Structure**: AI automatically breaks down complex work into manageable items
   - **Deliverable Generation**: AI suggests specific, measurable deliverables
   - **Dependency Mapping**: AI identifies and suggests dependencies between work items
   - **Effort Estimation**: AI provides effort estimates based on similar work items
   - **Work Item Templates**: AI-generated templates based on work type and industry

3. **AI Work Item Creation Flow**:
   ```
   User: "We need to improve our customer portal"
   
   AI: "Great! I'll help you break this down into specific work items.
        
        For a customer portal improvement, I typically see these areas:
        
        🎨 User Interface/Experience
        ⚙️ Backend/API Development  
        🔐 Security & Authentication
        📱 Mobile Responsiveness
        📊 Analytics & Reporting
        🧪 Testing & QA
        
        Which aspects are most important for your project?"
   
   User: "Focus on UI improvements and mobile responsiveness"
   
   AI: "Perfect! I'll create focused work items:
        
        ✓ Work Item 1: UI/UX Design Audit & Redesign
          Deliverables:
          - Current UI assessment report
          - New design mockups and prototypes
          - User testing results
          
        ✓ Work Item 2: Mobile Responsive Implementation  
          Deliverables:
          - Mobile-first CSS framework
          - Cross-device testing results
          - Performance optimization report
          
        Dependencies: Work Item 2 depends on Work Item 1
        Estimated Effort: 3-4 weeks total
        
        Should I add more specific work items or refine these further?"
   ```

4. **AI Work Item Refinement Features**:
   - **Work Item Decomposition**: Break large work items into smaller, manageable tasks
   - **Deliverable Refinement**: AI suggests specific, measurable deliverables
   - **Dependency Analysis**: AI maps dependencies and suggests optimal sequencing
   - **Resource Estimation**: AI estimates time, effort, and resource requirements
   - **Risk Assessment**: AI identifies potential risks and mitigation strategies
   - **Quality Criteria**: AI suggests acceptance criteria and quality standards

5. **Backend Services**:
   - `WorkItemService` interface and implementation
   - `AIWorkItemService` for AI-powered features
   - CRUD operations with validation
   - Agency-scoped data access
   - Auto-numbering for work item sequences
   - Duplicate code prevention
   - Work item template management
   - Dependency validation and cycle detection

6. **API Endpoints**:
   ```
   GET    /api/v1/agencies/{id}/work-items
   POST   /api/v1/agencies/{id}/work-items
   PUT    /api/v1/agencies/{id}/work-items/{workItemKey}
   DELETE /api/v1/agencies/{id}/work-items/{workItemKey}
   GET    /api/v1/agencies/{id}/work-items/html                      # HTML for HTMX
   POST   /api/v1/agencies/{id}/work-items/{workItemKey}/refine      # AI refine endpoint (like overview/refine)
   POST   /api/v1/agencies/{id}/work-items/generate                  # AI generate from scratch
   POST   /api/v1/agencies/{id}/work-items/breakdown                 # AI work breakdown
   GET    /api/v1/agencies/{id}/work-items/ai/templates             # AI-generated templates
   ```

7. **Enhanced User Interface with AI Refine Pattern**:
   - **Work Item Definition Form**: Standard textarea with AI Sparkle button
   - **AI Refine Button**: HTMX-powered refinement following introduction_card pattern
   - **Work Items List**: Enhanced with AI generation and breakdown options
   - **Template Selector**: AI-generated work item templates
   - **Dependency Mapper**: Visual dependency mapping with AI suggestions

8. **AI Refine Implementation (Following Introduction Card Pattern)**:
   ```html
   <!-- Work Item Definition Card -->
   <div class="box">
     <div id="workitem-content">
       <div class="content">
         <div class="field">
           <label class="label">Work Item Description</label>
           <div class="control">
             <textarea
               class="textarea"
               id="workitem-editor"
               placeholder="Describe the work item and its objectives..."
               rows="12"
               style="font-family: monospace; font-size: 14px;">
               { workItem.Description }
             </textarea>
           </div>
         </div>
         
         <div class="field">
           <label class="label">Deliverables</label>
           <div class="control">
             <textarea
               class="textarea"
               id="deliverables-editor"
               placeholder="List specific deliverables and outcomes..."
               rows="8">
               { strings.Join(workItem.Deliverables, "\n") }
             </textarea>
           </div>
         </div>
         
         <div class="field">
           <label class="label">Dependencies</label>
           <div class="control">
             <textarea
               class="textarea"
               id="dependencies-editor"
               placeholder="List dependencies and prerequisites..."
               rows="5">
               { strings.Join(workItem.Dependencies, "\n") }
             </textarea>
           </div>
         </div>
       </div>
       
       <div class="buttons is-right">
         <button
           class="button is-primary"
           onclick="saveWorkItemDefinition()"
           id="save-workitem-btn">
           <span class="icon"><i class="fas fa-save"></i></span>
           <span>Save</span>
         </button>
         
         <!-- AI Refine / Sparkle button (following introduction_card pattern) -->
         <button
           class="button is-info"
           hx-post={ "/api/v1/agencies/" + currentAgency.ID + "/work-items/" + workItem.Key + "/refine" }
           hx-include="#workitem-editor, #deliverables-editor, #dependencies-editor"
           hx-target="#workitem-content"
           hx-indicator="#ai-process-status"
           hx-on::after-request="
             console.log('🏁 HTMX request completed, hiding status...');
             if (window.hideAIProcessStatus) {
               window.hideAIProcessStatus();
             } else {
               console.log('❌ hideAIProcessStatus not available');
               const status = document.getElementById('ai-process-status');
               if (status) {
                 status.style.display = 'none';
                 console.log('✅ Status hidden manually');
               }
             }
           "
           id="ai-sparkle-btn"
           onclick="window.handleRefineClick && window.handleRefineClick()"
           title="Refine with AI">
           <span class="icon"><i class="fas fa-magic"></i></span>
           <span>Refine</span>
         </button>
         
         <!-- AI Breakdown button for complex work items -->
         <button
           class="button is-warning"
           hx-post={ "/api/v1/agencies/" + currentAgency.ID + "/work-items/breakdown" }
           hx-include="#workitem-editor"
           hx-target="#workitem-content"
           hx-indicator="#ai-process-status"
           title="Break down into smaller work items">
           <span class="icon"><i class="fas fa-sitemap"></i></span>
           <span>Breakdown</span>
         </button>
         
         <button
           class="button"
           onclick="undoWorkItemDefinition()"
           id="undo-workitem-btn">
           <span class="icon"><i class="fas fa-undo"></i></span>
           <span>Undo</span>
         </button>
       </div>
       
       <!-- Loading indicator for AI refine (same as introduction_card) -->
       <div id="ai-refine-loading" class="htmx-indicator">
         <div class="notification is-info is-light">
           <div class="is-flex is-align-items-center">
             <span class="icon has-text-info mr-2">
               <i class="fas fa-spinner fa-spin"></i>
             </span>
             <span>AI is refining your work item...</span>
           </div>
         </div>
       </div>
     </div>
   </div>
   ```

9. **AI-Enhanced Features**:
   - **Smart Templates**: Context-aware work item templates
   - **Dependency Prediction**: AI predicts likely dependencies based on work type
   - **Resource Optimization**: AI suggests resource allocation and scheduling
   - **Risk Identification**: AI flags potential risks and bottlenecks
   - **Quality Assurance**: AI suggests testing and validation approaches
   - **Progress Tracking**: AI-enhanced progress monitoring and reporting

**Acceptance Criteria**:
- [ ] Users can create, edit, and delete work items
- [ ] AI can generate work items from high-level descriptions
- [ ] AI provides intelligent work breakdown suggestions
- [ ] Work item templates are generated based on context
- [ ] Deliverables are automatically suggested and refined
- [ ] Dependencies are intelligently mapped and validated
- [ ] Effort estimation is provided with AI assistance
- [ ] Work item codes are unique within agency
- [ ] Auto-numbering works correctly
- [ ] Search and filtering functional with AI-enhanced search
- [ ] Form validation prevents invalid data
- [ ] Deliverables and dependencies can be managed
- [ ] Agency-scoped security implemented
- [ ] AI conversation history is maintained per work item

**Dependencies**: MVP-029 (Goals Module)

---

### MVP-033: RACI Matrix Editor

**Objective**: Implement a comprehensive visual RACI matrix editor with role assignments, validation rules, and templates for work items.

**Key Deliverables**:

1. **RACI Data Models**:
   ```go
   type RACIMatrix struct {
       WorkItemKey string         `json:"work_item_key"`
       Activities  []RACIActivity `json:"activities"`
       CreatedAt   time.Time      `json:"created_at"`
       UpdatedAt   time.Time      `json:"updated_at"`
   }

   type RACIActivity struct {
       ID          string              `json:"id"`
       Name        string              `json:"name"`
       Description string              `json:"description"`
       Assignments map[string]RACIRole `json:"assignments"` // role_name -> RACI role
   }

   type RACIRole string
   const (
       Responsible RACIRole = "R"
       Accountable RACIRole = "A"
       Consulted   RACIRole = "C"
       Informed    RACIRole = "I"
   )
   ```

2. **Visual RACI Matrix Editor**:
   - Interactive table/grid interface
   - Role columns (Agency Lead, Technical Lead, Subject Matter Expert, etc.)
   - Activity rows with descriptions
   - Click-to-assign RACI roles
   - Color-coded role assignments
   - Drag-and-drop for reordering activities

3. **RACI Validation Engine**:
   - Each activity must have exactly one Accountable (A) role
   - Each activity must have at least one Responsible (R) role
   - Warn about activities with no Consulted (C) or Informed (I) roles
   - Validation error display with specific guidance
   - Real-time validation as user makes changes

4. **Role Templates**:
   - Predefined RACI templates for common work item types
   - Templates: "Software Development", "Research & Analysis", "Deployment", "Testing"
   - Template application with customization options
   - Save custom templates for agency reuse

5. **Backend Services**:
   - `RACIService` interface and implementation
   - RACI matrix CRUD operations
   - Template management service
   - Validation service with detailed error reporting

6. **API Endpoints**:
   ```
   GET    /api/v1/agencies/{id}/work-items/{workItemKey}/raci
   POST   /api/v1/agencies/{id}/work-items/{workItemKey}/raci
   PUT    /api/v1/agencies/{id}/work-items/{workItemKey}/raci
   DELETE /api/v1/agencies/{id}/work-items/{workItemKey}/raci
   GET    /api/v1/agencies/{id}/raci-templates
   POST   /api/v1/agencies/{id}/raci-templates
   ```

7. **Advanced Features**:
   - Activity templates library
   - Role responsibility descriptions
   - RACI matrix export to PDF/Excel
   - Activity dependency mapping
   - Role workload analysis

**UI Components**:
```html
<!-- RACI Matrix Editor -->
<div class="raci-matrix-editor">
  <div class="matrix-toolbar">
    <button class="button is-primary" onclick="addActivity()">
      <span class="icon">➕</span>
      <span>Add Activity</span>
    </button>
    <div class="dropdown">
      <div class="dropdown-trigger">
        <button class="button">
          <span>Templates</span>
          <span class="icon">🔽</span>
        </button>
      </div>
      <div class="dropdown-menu">
        <div class="dropdown-content">
          <a class="dropdown-item" onclick="loadTemplate('software-dev')">
            Software Development
          </a>
          <a class="dropdown-item" onclick="loadTemplate('research')">
            Research & Analysis
          </a>
        </div>
      </div>
    </div>
  </div>
  
  <div class="matrix-container">
    <table class="table is-striped is-hoverable raci-table">
      <thead>
        <tr>
          <th>Activity</th>
          <th>Agency Lead</th>
          <th>Technical Lead</th>
          <th>Subject Matter Expert</th>
          <th>Quality Assurance</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody id="raci-activities">
        <!-- Activities populated dynamically -->
        <tr class="activity-row" data-activity-id="act-1">
          <td class="activity-description">
            <strong>Requirements Gathering</strong>
            <p class="help">Collect and document functional requirements</p>
          </td>
          <td class="role-assignment">
            <div class="raci-selector">
              <button class="raci-btn active" data-role="A">A</button>
              <button class="raci-btn" data-role="R">R</button>
              <button class="raci-btn" data-role="C">C</button>
              <button class="raci-btn" data-role="I">I</button>
            </div>
          </td>
          <td class="role-assignment">
            <div class="raci-selector">
              <button class="raci-btn" data-role="A">A</button>
              <button class="raci-btn active" data-role="R">R</button>
              <button class="raci-btn" data-role="C">C</button>
              <button class="raci-btn" data-role="I">I</button>
            </div>
          </td>
          <!-- More role columns -->
          <td class="activity-actions">
            <button class="button is-small" onclick="editActivity('act-1')">
              ✏️
            </button>
            <button class="button is-small" onclick="deleteActivity('act-1')">
              🗑️
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
  
  <div class="validation-summary">
    <div class="notification is-warning" id="validation-warnings">
      <strong>⚠️ Validation Issues:</strong>
      <ul>
        <li>Activity "Testing" has no Accountable role assigned</li>
        <li>Activity "Deployment" has multiple Accountable roles</li>
      </ul>
    </div>
  </div>
</div>
```

**Acceptance Criteria**:
- [ ] Visual RACI matrix editor works smoothly
- [ ] RACI validation rules enforced in real-time
- [ ] Templates can be applied and customized
- [ ] Role assignments save and persist correctly
- [ ] Export functionality generates proper RACI documentation
- [ ] User can manage activities (add/edit/delete/reorder)
- [ ] Interface is intuitive and responsive
- [ ] Integration with work items seamless

**Dependencies**: MVP-030 (Work Items Basic Management)

---

### MVP-031: Graph Relationships System

**Objective**: Implement ArangoDB graph collections and relationship mapping between Problems and Work Items with rich metadata.

**Key Deliverables**:

1. **Graph Database Schema**:
   - `problems` collection (already from MVP-029)
   - `work_items` collection (already from MVP-030)
   - `problem_work_item_relationships` edge collection
   - Proper graph indices and constraints

2. **Relationship Data Model**:
   ```go
   type ProblemWorkItemRelationship struct {
       Key                     string           `json:"_key,omitempty"`
       ID                      string           `json:"_id,omitempty"`
       From                    string           `json:"_from"` // problems/{problem_key}
       To                      string           `json:"_to"`   // work_items/{work_item_key}
       RelationshipType        RelationshipType `json:"relationship_type"`
       ContributionDescription string           `json:"contribution_description"`
       ImpactLevel            ImpactLevel      `json:"impact_level"`
       CreatedAt              time.Time        `json:"created_at"`
       UpdatedAt              time.Time        `json:"updated_at"`
   }
   ```

3. **Relationship Types**:
   - `solves` - Work item directly addresses the problem
   - `supports` - Work item contributes to solving the problem
   - `enables` - Work item creates prerequisites for the problem
   - `mitigates` - Work item reduces impact/likelihood of the problem

4. **Graph Services**:
   - `RelationshipService` for edge management
   - Graph traversal queries using AQL
   - Relationship validation (ensure nodes exist)
   - Bulk relationship operations

5. **API Endpoints**:
   ```
   GET    /api/v1/agencies/{id}/relationships
   POST   /api/v1/agencies/{id}/relationships
   PUT    /api/v1/agencies/{id}/relationships/{relationshipKey}
   DELETE /api/v1/agencies/{id}/relationships/{relationshipKey}
   GET    /api/v1/agencies/{id}/problems/{problemKey}/work-items
   GET    /api/v1/agencies/{id}/work-items/{workItemKey}/problems
   ```

6. **Relationship Interface**:
   - Visual relationship editor
   - Problem/Work Item selection dropdowns
   - Relationship type and impact level selectors
   - Contribution description editor with templates
   - Validation and duplicate prevention

**Acceptance Criteria**:
- [ ] Graph collections created and indexed
- [ ] Relationships can be created between problems and work items
- [ ] Graph traversal queries work efficiently
- [ ] Relationship validation prevents invalid references
- [ ] Users can manage relationships through UI
- [ ] Relationship metadata is properly stored

**Dependencies**: MVP-033 (RACI Matrix Editor)

---

### MVP-032: Agency Operations Analytics

**Objective**: Implement graph visualization, coverage analysis, and relationship analytics to provide insights into agency operations.

**Key Deliverables**:

1. **Graph Visualization**:
   - Interactive graph diagram showing problem-work item relationships
   - Node types: Problems (circles), Work Items (squares)
   - Edge types: Different colors/styles for relationship types
   - Zoom, pan, and selection functionality
   - Graph layout algorithms (force-directed, hierarchical)

2. **Analytics Queries**:
   ```aql
   // Coverage Analysis - Problems without work items
   FOR p IN problems
     FILTER p.agency_id == @agencyId
     LET work_items = (
       FOR v IN 1..1 OUTBOUND p._id problem_work_item_relationships
         RETURN v
     )
     FILTER LENGTH(work_items) == 0
     RETURN p

   // Impact Analysis - Work items affecting multiple problems  
   FOR wi IN work_items
     FILTER wi.agency_id == @agencyId
     LET problems = (
       FOR v IN 1..1 INBOUND wi._id problem_work_item_relationships
         RETURN v
     )
     FILTER LENGTH(problems) > 1
     RETURN {work_item: wi, problem_count: LENGTH(problems)}
   ```

3. **Analytics Dashboard**:
   - Coverage metrics (problems without solutions)
   - Impact metrics (work items solving multiple problems)
   - Relationship type distribution
   - RACI role distribution across work items
   - Agency operations health score

4. **Export Functionality**:
   - Generate PDF reports with RACI matrices
   - Export relationship mappings to CSV
   - Generate problem-solution documentation
   - Export graph visualizations as images

5. **API Endpoints**:
   ```
   GET /api/v1/agencies/{id}/analytics/coverage
   GET /api/v1/agencies/{id}/analytics/impact  
   GET /api/v1/agencies/{id}/analytics/graph
   GET /api/v1/agencies/{id}/analytics/raci-summary
   POST /api/v1/agencies/{id}/export/documentation
   ```

**Acceptance Criteria**:
- [ ] Graph visualization displays correctly
- [ ] Coverage analysis identifies unaddressed problems
- [ ] Impact analysis shows multi-problem work items
- [ ] Export functionality generates proper documentation
- [ ] Analytics provide actionable insights
- [ ] Performance is acceptable for typical agency sizes

**Dependencies**: MVP-031 (Graph Relationships System)

---

## Agent Property Broadcasting Feature (P1 - Critical)

*Enables UC-TRACK-001 (Safiri Salama) and other real-time tracking/monitoring use cases*

| Task ID | Title                                    | Description                                                                                                      | Status      | Priority | Effort | Skills Required            | Dependencies |
| ------- | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | ----------- | -------- | ------ | -------------------------- | ------------ |
| MVP-016 | Core Broadcasting Infrastructure         | Implement BroadcastConfiguration, PropertyBroadcaster service, ContextEvaluator, and integration with PubSub    | Not Started | P1       | High   | Go, Backend Dev, PubSub    | MVP-013      |
| MVP-017 | Subscription Management                  | Build SubscriptionManager, subscriber filtering, favorite functionality, and subscription API endpoints          | Not Started | P1       | Medium | Go, Backend Dev, REST API  | MVP-016      |
| MVP-018 | Privacy & Security Controls              | Implement geofencing, property masking, permission model, audit logging, and encryption for sensitive properties | Not Started | P1       | Medium | Security, Backend Dev      | MVP-017      |
| MVP-019 | Performance Optimization & Scale         | Performance tuning, caching, load balancing for broadcasters, message broker optimization, monitoring & alerting | Not Started | P1       | Medium | Performance, DevOps        | MVP-018      |
| MVP-020 | UC-TRACK-001 Integration & Testing       | Implement Vehicle & Passenger agents, build mobile/web UI, SACCO management portal, end-to-end testing          | Not Started | P1       | High   | Full-stack, Mobile Dev     | MVP-019      |

### MVP-016: Core Broadcasting Infrastructure

**Objective**: Build the foundational broadcasting system that enables agents to automatically publish properties at configurable intervals.

**Key Deliverables**:
1. **Data Structures**:
   - `BroadcastConfiguration` type with rules, intervals, privacy controls
   - `BroadcastRule` with condition evaluation logic
   - `PropertyUpdateMessage` format
   - `BroadcastMetrics` for monitoring

2. **Core Services**:
   - `PropertyBroadcaster` service implementing lifecycle management
   - `ContextEvaluator` for rule matching and interval determination
   - `BroadcastConfigRepository` for persistent storage

3. **Agent Integration**:
   - Extend Agent base class with broadcasting methods
   - Add `EnableBroadcasting()`, `StartBroadcasting()`, `StopBroadcasting()`
   - Add `UpdateBroadcastInterval()`, `PauseBroadcasting()`, `BroadcastNow()`
   - Implement property collection from agent state

4. **PubSub Integration**:
   - Extend PubSubService with `PublishPropertyUpdate()`
   - Add topic routing for property updates
   - Implement message formatting and priority handling

**Acceptance Criteria**:
- Agent can configure and start broadcasting
- Context-aware interval adjustment works correctly
- Properties are published to message bus
- Basic metrics collection functional
- Unit tests for all core components (>80% coverage)

**Technical Details**: See `documents/3-SofwareDevelopment/core-systems/agent-property-broadcasting.md`

### MVP-017: Subscription Management

**Objective**: Enable agents to subscribe to property updates from other agents with filtering and notification preferences.

**Key Deliverables**:
1. **Subscription Service**:
   - `SubscriptionManager` interface and implementation
   - Subscribe/Unsubscribe operations
   - Favorite/Unfavorite functionality
   - Subscription approval workflow

2. **Filtering Logic**:
   - Property-based filtering
   - Priority-based filtering
   - Geofence-based filtering
   - Time window filtering
   - Subscriber type restrictions

3. **Data Models**:
   - `Subscriber` type with preferences
   - `SubscriptionFilters` configuration
   - `NotificationPreferences` settings

4. **REST API Endpoints**:
   ```
   POST   /api/v1/agents/{agentId}/broadcasting/subscribe
   DELETE /api/v1/agents/{agentId}/broadcasting/unsubscribe
   POST   /api/v1/agents/{agentId}/broadcasting/favorite
   DELETE /api/v1/agents/{agentId}/broadcasting/favorite
   GET    /api/v1/agents/{agentId}/broadcasting/subscribers
   GET    /api/v1/agents/{agentId}/subscribers/subscriptions
   ```

5. **Notification Delivery**:
   - Push notification integration
   - SMS notification integration
   - Email notification integration
   - In-app notification handling

**Acceptance Criteria**:
- Subscribers receive filtered property updates
- Favorite notifications work with priority
- API endpoints functional and documented
- Subscription persistence works correctly
- Integration tests for subscribe/receive flow

### MVP-018: Privacy & Security Controls

**Objective**: Implement comprehensive privacy and security features to protect sensitive location and property data.

**Key Deliverables**:
1. **Geofencing Service**:
   - Geofence zone definition and storage
   - Point-in-polygon detection
   - Integration with broadcast pause logic
   - Admin UI for geofence management

2. **Property Masking**:
   - Masking rule configuration per property
   - Context-aware masking (e.g., mask exact location, show only vicinity)
   - Redaction of sensitive fields

3. **Permission Model**:
   - `BroadcastPermissions` service
   - Subscriber authorization checks
   - Property-level access control
   - Approval workflow for sensitive publishers

4. **Audit Logging**:
   - Log all subscription requests
   - Track property access patterns
   - Record broadcast pause/resume events
   - Generate compliance reports

5. **Encryption**:
   - Encrypt sensitive properties in transit
   - Secure storage of subscriber credentials
   - API authentication and authorization

**Acceptance Criteria**:
- Broadcasting pauses in restricted geofences
- Property masking works correctly
- Unauthorized subscribers cannot access data
- Complete audit trail for all operations
- Security tests pass (no critical vulnerabilities)

### MVP-019: Performance Optimization & Scale

**Objective**: Ensure the broadcasting system can handle production-scale loads with low latency and high reliability.

**Key Deliverables**:
1. **Caching Strategy**:
   - Cache subscriber lists in memory with TTL
   - Cache broadcast configurations
   - Cache geofence definitions
   - Redis integration for distributed caching

2. **Performance Optimizations**:
   - Batch property updates where possible
   - Async message publishing
   - Connection pooling for message broker
   - Efficient subscriber filtering at source

3. **Load Balancing**:
   - Distribute agents across broadcaster instances
   - Sharding strategy for high-volume agents
   - Health checks and failover

4. **Monitoring & Alerting**:
   - Prometheus metrics integration
   - Grafana dashboards for broadcast metrics
   - Alert rules for anomalies (high latency, failures)
   - Performance profiling tools

5. **Resource Limits**:
   - Implement `BroadcastResourceLimits`
   - Rate limiting per agent
   - Throttling for misbehaving agents
   - Circuit breakers for external services

**Acceptance Criteria**:
- Support 10,000+ concurrent broadcasting agents
- Sub-second delivery latency (p99 < 500ms)
- Handle 100,000+ messages per minute
- Memory usage stays within bounds under load
- Passes load and stress tests

**Performance Targets**:
- Broadcast-to-delivery latency: <500ms (p99)
- Concurrent agents: 10,000+
- Subscribers per agent: 1,000+
- Message throughput: 100,000/min
- System uptime: 99.9%

### MVP-020: UC-TRACK-001 Integration & Testing

**Objective**: Complete end-to-end implementation of the Safiri Salama tracking system as the reference implementation.

**Key Deliverables**:
1. **Vehicle Agent Implementation**:
   - GPS location tracking integration
   - Context-aware broadcast configuration
   - Status state machine (idle, en_route, at_stop, etc.)
   - Driver dashboard (mobile/tablet)

2. **Passenger Agent Implementation**:
   - Favorite vehicle management
   - Smart notification handling
   - Trip rating and feedback
   - Commute pattern learning

3. **Parent Agent Implementation**:
   - Child bus tracking
   - Emergency alert handling
   - Proximity notifications
   - School communication portal

4. **Fleet Operator Portal**:
   - SACCO manager dashboard
   - Loyalty analytics
   - Vehicle performance metrics
   - Passenger engagement insights

5. **Mobile Applications**:
   - Passenger app (iOS/Android) - track favorites
   - Parent app (iOS/Android) - track school bus
   - Driver app (iOS/Android) - status control

6. **Web Portal**:
   - Admin dashboard for fleet operators
   - Real-time fleet map view
   - Analytics and reporting
   - Configuration management

7. **Use Case Deployment**:
   - UC-TRACK-001 folder structure
   - Agent JSON schemas (Vehicle, Passenger, Parent, etc.)
   - Environment configuration (.env)
   - Deployment scripts (start.sh)
   - Documentation and user guides

**Acceptance Criteria**:
- All 5 agent types functional (Vehicle, Parent, Passenger, Route Manager, Fleet Operator)
- Mobile apps published to app stores (beta)
- End-to-end tracking works: Vehicle → Broadcast → Passenger receives
- Favorite notification feature works
- Real-world pilot with 5+ vehicles and 50+ subscribers
- User feedback positive (>4.0 rating)
- System stable under real usage patterns

**Pilot Program**:
1. **School Pilot**: 2-3 schools, 5-10 buses, 200-300 parents
2. **Matatu Pilot**: 1-2 SACCOs, 10-15 vehicles, 500-1000 passengers
3. **Duration**: 4 weeks
4. **Success Metrics**: >70% active usage, <5% error rate, positive feedback

---

### MVP-043: Work Items UI Module

**Objective**: Build a complete Work Items management interface in the Agency Designer with CRUD operations, AI-powered refinement, smart templates, deliverables and dependencies management, and advanced filtering capabilities.

**Key Deliverables**:

1. **Work Items Data Model**:
   ```go
   type WorkItem struct {
       Key             string            `json:"key"`              // e.g., "WI-001"
       AgencyID        string            `json:"agency_id"`
       Title           string            `json:"title"`
       Description     string            `json:"description"`
       Type            WorkItemType      `json:"type"`             // Task, Feature, Epic, etc.
       Priority        Priority          `json:"priority"`         // P0, P1, P2, P3
       Status          WorkItemStatus    `json:"status"`           // Not Started, In Progress, Done
       Deliverables    []string          `json:"deliverables"`
       Dependencies    []string          `json:"dependencies"`     // References to other work item keys
       EstimatedEffort string            `json:"estimated_effort"` // e.g., "2 weeks", "40 hours"
       AssignedTo      string            `json:"assigned_to,omitempty"`
       Tags            []string          `json:"tags,omitempty"`
       CreatedAt       time.Time         `json:"created_at"`
       UpdatedAt       time.Time         `json:"updated_at"`
   }
   ```

2. **Work Items List View**:
   - Table/card view of all work items
   - Filtering by status, priority, type, tags
   - Search functionality
   - Sortable columns (priority, status, created date)
   - Bulk operations (delete, update status)
   - Quick actions (edit, delete, duplicate)
   - Color coding by priority and status

3. **Work Item Editor (Modal/Form)**:
   - Title and description fields
   - Type selector (Task, Feature, Epic, Bug, Research)
   - Priority selector (P0-P3)
   - Status selector with workflow
   - Deliverables list editor (add/remove/reorder)
   - Dependencies selector (dropdown of existing work items)
   - Estimated effort input
   - Tags input (multi-select or comma-separated)
   - Auto-generate work item key (WI-001, WI-002, etc.)

4. **AI Refinement Integration** (Following Introduction Card Pattern):
   ```html
   <!-- Work Item Editor with AI Sparkle Button -->
   <div class="box" id="workitem-editor-box">
     <div id="workitem-content">
       <div class="field">
         <label class="label">Description</label>
         <textarea id="workitem-description" class="textarea" rows="4"></textarea>
       </div>
       
       <div class="field">
         <label class="label">Deliverables</label>
         <textarea id="workitem-deliverables" class="textarea" rows="3"></textarea>
       </div>
       
       <div class="buttons is-right">
         <button class="button is-primary" onclick="saveWorkItem()">
           <span class="icon"><i class="fas fa-save"></i></span>
           <span>Save</span>
         </button>
         
         <!-- AI Refine Button (HTMX powered) -->
         <button class="button is-info"
                 hx-post="/api/v1/agencies/{id}/work-items/{key}/refine"
                 hx-include="#workitem-description, #workitem-deliverables"
                 hx-target="#workitem-content"
                 hx-indicator="#ai-process-status">
           <span class="icon"><i class="fas fa-magic"></i></span>
           <span>Refine with AI</span>
         </button>
       </div>
     </div>
   </div>
   ```

5. **Dependencies Visualization**:
   - Visual dependency graph
   - Dependency validation (prevent circular dependencies)
   - Highlight blocking work items
   - Suggest logical sequencing

6. **Backend Services**:
   - `WorkItemService` interface and implementation
   - CRUD operations with agency-scoped access
   - Auto-numbering for work item keys
   - Dependency cycle detection
   - Work item templates management

7. **API Endpoints**:
   ```
   GET    /api/v1/agencies/{id}/work-items
   POST   /api/v1/agencies/{id}/work-items
   PUT    /api/v1/agencies/{id}/work-items/{key}
   DELETE /api/v1/agencies/{id}/work-items/{key}
   POST   /api/v1/agencies/{id}/work-items/{key}/refine      # AI refinement
   GET    /api/v1/agencies/{id}/work-items/templates         # Work item templates
   POST   /api/v1/agencies/{id}/work-items/validate-deps     # Validate dependencies
   ```

8. **Template System**:
   - Predefined work item templates by type
   - Templates include suggested deliverables
   - Context-aware template suggestions
   - Save custom templates

**Acceptance Criteria**:
- [ ] Users can create, edit, and delete work items
- [ ] Work item keys auto-generate uniquely (WI-001, WI-002, etc.)
- [ ] Filtering and search work correctly
- [ ] Dependencies can be added and validated
- [ ] AI refinement improves descriptions and suggests deliverables
- [ ] No circular dependencies allowed
- [ ] Deliverables are editable as list items
- [ ] Work items are agency-scoped
- [ ] Real-time validation feedback
- [ ] Export work items list to CSV/PDF

**Dependencies**: MVP-029 (Goals Module - provides foundation patterns)

---

### MVP-044: Agent Types UI Module

**Objective**: Build a comprehensive Agent Types definition and management interface allowing users to define, categorize, and configure different types of agents that will be used in their agency.

**Key Deliverables**:

1. **Agent Type Data Model**:
   ```go
   type AgentType struct {
       Key              string            `json:"key"`               // e.g., "AT-001"
       AgencyID         string            `json:"agency_id"`
       Name             string            `json:"name"`              // e.g., "Sensor Agent"
       Category         string            `json:"category"`          // Monitor, Controller, Analyzer, etc.
       Description      string            `json:"description"`
       Capabilities     []string          `json:"capabilities"`      // List of what this agent can do
       AutonomyLevel    AutonomyLevel     `json:"autonomy_level"`   // Manual, SemiAutomated, FullyAutomated
       RequiredSkills   []string          `json:"required_skills,omitempty"`
       Configuration    map[string]string `json:"configuration,omitempty"` // Key-value config params
       Icon             string            `json:"icon,omitempty"`    // Font Awesome icon or emoji
       Color            string            `json:"color,omitempty"`   // For UI visualization
       CreatedAt        time.Time         `json:"created_at"`
       UpdatedAt        time.Time         `json:"updated_at"`
   }
   
   type AutonomyLevel string
   const (
       ManualControl      AutonomyLevel = "manual"
       SemiAutomated      AutonomyLevel = "semi_automated"
       FullyAutomated     AutonomyLevel = "fully_automated"
   )
   ```

2. **Agent Types Catalog View**:
   - Grid/card layout of agent types
   - Visual representation with icons and colors
   - Category grouping
   - Search and filter by category
   - Quick view of capabilities
   - Add new agent type button

3. **Agent Type Editor Form**:
   - Name and description fields
   - Category selector (Monitor, Controller, Analyzer, Coordinator, Reporter)
   - Autonomy level selector (Manual, Semi-Automated, Fully Automated)
   - Capabilities list editor (add/remove capabilities)
   - Required skills multi-select
   - Configuration parameters (key-value pairs editor)
   - Icon picker (emoji or Font Awesome)
   - Color picker for visual identification
   - Preview of agent type card

4. **Agent Type Templates**:
   - Predefined templates for common agent types:
     * Sensor Agent (monitoring, data collection)
     * Controller Agent (actuation, control logic)
     * Analyzer Agent (data analysis, pattern detection)
     * Coordinator Agent (workflow orchestration)
     * Reporter Agent (reporting, dashboards)
   - Template application with customization
   - Industry-specific templates (water, logistics, healthcare, etc.)

5. **Capability Management**:
   - Predefined capability library
   - Custom capability creation
   - Capability descriptions and examples
   - Capability validation

6. **Backend Services**:
   - `AgentTypeService` interface and implementation
   - CRUD operations with agency-scoped access
   - Auto-numbering for agent type keys
   - Template management
   - Capability library management

7. **API Endpoints**:
   ```
   GET    /api/v1/agencies/{id}/agent-types
   POST   /api/v1/agencies/{id}/agent-types
   PUT    /api/v1/agencies/{id}/agent-types/{key}
   DELETE /api/v1/agencies/{id}/agent-types/{key}
   GET    /api/v1/agencies/{id}/agent-types/templates       # Agent type templates
   GET    /api/v1/capabilities                              # Capability library
   ```

8. **UI Components**:
   ```html
   <!-- Agent Type Card in Catalog -->
   <div class="card agent-type-card" style="border-left: 4px solid {color}">
     <div class="card-content">
       <div class="media">
         <div class="media-left">
           <span class="icon is-large">{icon}</span>
         </div>
         <div class="media-content">
           <p class="title is-5">{name}</p>
           <p class="subtitle is-6">{category}</p>
         </div>
       </div>
       <div class="content">
         <p>{description}</p>
         <div class="tags">
           {#each capabilities}
             <span class="tag is-info">{capability}</span>
           {/each}
         </div>
         <p class="is-size-7">
           <span class="tag is-light">{autonomy_level}</span>
         </p>
       </div>
     </div>
     <footer class="card-footer">
       <a class="card-footer-item" onclick="editAgentType('{key}')">Edit</a>
       <a class="card-footer-item" onclick="duplicateAgentType('{key}')">Duplicate</a>
       <a class="card-footer-item has-text-danger" onclick="deleteAgentType('{key}')">Delete</a>
     </footer>
   </div>
   ```

**Acceptance Criteria**:
- [ ] Users can create, edit, and delete agent types
- [ ] Agent type keys auto-generate uniquely (AT-001, AT-002, etc.)
- [ ] Category filtering and search work correctly
- [ ] Templates can be applied and customized
- [ ] Capabilities can be added from library or custom created
- [ ] Icon and color pickers functional
- [ ] Visual catalog displays agent types clearly
- [ ] Agent types are agency-scoped
- [ ] Export agent types catalog to PDF
- [ ] Validation prevents duplicate names

**Dependencies**: MVP-043 (Work Items UI Module)

---

### MVP-045: RACI Matrix UI Editor

**Objective**: Build an interactive RACI (Responsible, Accountable, Consulted, Informed) matrix editor with visual role assignments, validation rules, templates, and integration with work items.

**Key Deliverables**:

1. **RACI Data Models**:
   ```go
   type RACIMatrix struct {
       Key          string           `json:"key"`
       AgencyID     string           `json:"agency_id"`
       WorkItemKey  string           `json:"work_item_key,omitempty"`  // Optional: link to specific work item
       Name         string           `json:"name"`
       Activities   []RACIActivity   `json:"activities"`
       Roles        []string         `json:"roles"`          // List of role names
       CreatedAt    time.Time        `json:"created_at"`
       UpdatedAt    time.Time        `json:"updated_at"`
   }

   type RACIActivity struct {
       ID          string              `json:"id"`
       Name        string              `json:"name"`
       Description string              `json:"description"`
       Assignments map[string]RACIRole `json:"assignments"` // role_name -> RACI role
   }

   type RACIRole string
   const (
       Responsible RACIRole = "R"  // Does the work
       Accountable RACIRole = "A"  // Ultimately answerable
       Consulted   RACIRole = "C"  // Provides input
       Informed    RACIRole = "I"  // Kept in the loop
   )
   ```

2. **Visual RACI Matrix Editor**:
   ```html
   <!-- Interactive RACI Matrix -->
   <div class="raci-matrix-container">
     <table class="table is-striped is-hoverable is-fullwidth raci-table">
       <thead>
         <tr>
           <th style="width: 30%">Activity</th>
           <th>Project Manager</th>
           <th>Tech Lead</th>
           <th>Developer</th>
           <th>QA Engineer</th>
           <th style="width: 80px">Actions</th>
         </tr>
       </thead>
       <tbody>
         <tr class="activity-row">
           <td>
             <strong>Requirements Gathering</strong>
             <p class="help is-size-7">Define project requirements</p>
           </td>
           <td class="role-cell">
             <div class="raci-selector">
               <button class="raci-btn" data-role="R">R</button>
               <button class="raci-btn active" data-role="A">A</button>
               <button class="raci-btn" data-role="C">C</button>
               <button class="raci-btn" data-role="I">I</button>
               <button class="raci-btn-clear" title="Clear">✕</button>
             </div>
           </td>
           <td class="role-cell">
             <div class="raci-selector">
               <button class="raci-btn" data-role="R">R</button>
               <button class="raci-btn" data-role="A">A</button>
               <button class="raci-btn active" data-role="C">C</button>
               <button class="raci-btn" data-role="I">I</button>
               <button class="raci-btn-clear" title="Clear">✕</button>
             </div>
           </td>
           <!-- More role cells -->
           <td>
             <div class="buttons are-small">
               <button class="button is-small" title="Edit"><i class="fas fa-edit"></i></button>
               <button class="button is-small" title="Delete"><i class="fas fa-trash"></i></button>
             </div>
           </td>
         </tr>
       </tbody>
     </table>
     
     <div class="buttons">
       <button class="button is-primary" onclick="addActivity()">
         <span class="icon"><i class="fas fa-plus"></i></span>
         <span>Add Activity</span>
       </button>
       <button class="button" onclick="addRole()">
         <span class="icon"><i class="fas fa-user-plus"></i></span>
         <span>Add Role</span>
       </button>
     </div>
   </div>
   ```

3. **RACI Validation Engine**:
   - **Rule 1**: Each activity must have exactly ONE Accountable (A)
   - **Rule 2**: Each activity must have at least ONE Responsible (R)
   - **Rule 3**: Warn if no Consulted (C) or Informed (I) roles
   - Real-time validation with visual feedback
   - Validation error display with specific guidance
   - Prevent saving invalid RACI matrices

4. **Role Management**:
   - Add/remove roles dynamically
   - Rename roles
   - Role descriptions
   - Default roles template (Project Manager, Tech Lead, Developer, QA, etc.)

5. **Activity Management**:
   - Add/remove activities
   - Edit activity name and description
   - Reorder activities (drag-and-drop)
   - Activity templates by work item type
   - Import activities from work items

6. **RACI Templates**:
   - Predefined templates for common scenarios:
     * Software Development Project
     * Research & Analysis
     * Infrastructure Deployment
     * Testing & QA
     * Change Management
   - Template application with customization
   - Save custom templates for reuse

7. **Backend Services**:
   - `RACIService` interface and implementation
   - CRUD operations for RACI matrices
   - Validation service with detailed error reporting
   - Template management service
   - Export service (PDF, Excel)

8. **API Endpoints**:
   ```
   GET    /api/v1/agencies/{id}/raci-matrices
   POST   /api/v1/agencies/{id}/raci-matrices
   PUT    /api/v1/agencies/{id}/raci-matrices/{key}
   DELETE /api/v1/agencies/{id}/raci-matrices/{key}
   POST   /api/v1/agencies/{id}/raci-matrices/{key}/validate # Validate matrix
   GET    /api/v1/agencies/{id}/raci-templates               # RACI templates
   POST   /api/v1/agencies/{id}/raci-matrices/{key}/export   # Export to PDF/Excel
   ```

9. **Validation Feedback UI**:
   ```html
   <!-- Validation Summary -->
   <div class="notification is-warning" id="raci-validation-warnings">
     <p class="has-text-weight-bold">⚠️ RACI Validation Issues:</p>
     <ul>
       <li>Activity "Design Architecture" has no Accountable (A) role assigned</li>
       <li>Activity "Code Review" has multiple Accountable (A) roles</li>
       <li>Activity "Testing" has no Responsible (R) role</li>
     </ul>
     <p class="help mt-2">
       Fix these issues before saving. Each activity must have exactly one Accountable 
       and at least one Responsible role.
     </p>
   </div>
   ```

10. **Export Functionality**:
    - Export to PDF with formatted table
    - Export to Excel/CSV
    - Include validation status in export
    - Professional formatting

**Acceptance Criteria**:
- [ ] Users can create and edit RACI matrices
- [ ] Click-to-assign RACI roles works smoothly
- [ ] Real-time validation enforces RACI rules
- [ ] Visual feedback for validation errors
- [ ] Activities and roles can be added/removed dynamically
- [ ] Templates can be applied and customized
- [ ] Export to PDF and Excel works correctly
- [ ] RACI matrices can be linked to work items
- [ ] Role descriptions are editable
- [ ] Drag-and-drop reordering of activities
- [ ] RACI matrices are agency-scoped
- [ ] Color coding for different RACI roles

**Dependencies**: MVP-044 (Agent Types UI Module)

---

### MVP-042: AI-Powered Agency Creator

**Objective**: Implement a comprehensive AI-driven agency creation flow that allows users to upload text documents (RFPs, SOWs, specifications), select which agency design areas to generate, and batch-generate complete agency designs including introduction, goals, work items, agent types, and RACI matrices.

**Key Deliverables**:

1. **Text Upload & Parsing**:
   - File upload interface (PDF, DOCX, TXT, MD)
   - Text extraction and preprocessing
   - Document structure analysis
   - Context extraction for AI prompts
   ```go
   type SourceDocument struct {
       DocumentID   string    `json:"document_id"`
       Filename     string    `json:"filename"`
       ContentType  string    `json:"content_type"`
       TextContent  string    `json:"text_content"`
       UploadedAt   time.Time `json:"uploaded_at"`
       ProcessedAt  time.Time `json:"processed_at"`
   }
   ```

2. **Generation Selection Interface**:
   - Checkbox-based selection for generation areas:
     * ☑️ Introduction (background, purpose, scope)
     * ☑️ Goals (structured goal catalog with SMART format)
     * ☑️ Work Items (breakdown with deliverables, dependencies)
     * ☑️ Agent Types (roles, capabilities, specifications)
     * ☑️ RACI Matrix (responsibility assignments)
   - Preview available source text
   - Configure generation parameters (detail level, formality, etc.)
   - Batch or individual generation modes

3. **AI Generation Engine**:
   ```go
   type GenerationRequest struct {
       AgencyID     string              `json:"agency_id"`
       SourceText   string              `json:"source_text"`
       Areas        []GenerationArea    `json:"areas"`
       Options      GenerationOptions   `json:"options"`
   }

   type GenerationArea string
   const (
       AreaIntroduction GenerationArea = "introduction"
       AreaGoals        GenerationArea = "goals"
       AreaWorkItems    GenerationArea = "work_items"
       AreaAgentTypes   GenerationArea = "agent_types"
       AreaRACIMatrix   GenerationArea = "raci_matrix"
   )

   type GenerationOptions struct {
       DetailLevel  string `json:"detail_level"` // brief, standard, detailed
       Formality    string `json:"formality"`    // casual, professional, formal
       AgencyType   string `json:"agency_type"`  // context for generation
   }
   ```

4. **Batch AI Processing**:
   - Parallel generation of selected areas
   - Progress tracking for each area
   - Error handling and retry logic
   - Streaming results as they complete
   - Validation of generated content

5. **Generated Content Review & Edit**:
   - Side-by-side view: Source text | Generated content
   - Edit generated content before saving
   - Regenerate individual sections
   - Accept/reject individual items
   - Bulk operations (accept all, reject all)

6. **Agency Design Organization**:
   - Separate tabs/sections in Agency Designer:
     * 📄 Introduction
     * 🎯 Goals
     * 📋 Work Items
     * 🤖 Agent Types
     * 📊 RACI Matrix
   - Navigation between sections
   - Completion indicators per section
   - Save progress at any point

7. **Integration with Existing Modules**:
   - Reuse Introduction module (MVP-025 - completed)
   - Reuse Goals module (MVP-029 - completed)
   - Integrate with Work Items module (MVP-030)
   - Connect to Agent Types system
   - Link to RACI Matrix editor (MVP-033)

8. **API Endpoints**:
   ```
   POST   /api/v1/agencies/{id}/upload-document
   POST   /api/v1/agencies/{id}/generate/batch
   GET    /api/v1/agencies/{id}/generation-status
   POST   /api/v1/agencies/{id}/generate/introduction
   POST   /api/v1/agencies/{id}/generate/goals
   POST   /api/v1/agencies/{id}/generate/work-items
   POST   /api/v1/agencies/{id}/generate/agent-types
   POST   /api/v1/agencies/{id}/generate/raci-matrix
   ```

**UI Components**:

```html
<!-- Agency Creation Wizard -->
<div class="agency-creator-wizard">
  <!-- Step 1: Upload Source Document -->
  <div class="wizard-step" id="step-upload">
    <h2 class="title is-4">Step 1: Upload Source Document</h2>
    <div class="file has-name is-fullwidth">
      <label class="file-label">
        <input class="file-input" type="file" 
               accept=".pdf,.doc,.docx,.txt,.md"
               id="source-document">
        <span class="file-cta">
          <span class="file-icon">📁</span>
          <span class="file-label">Choose a file…</span>
        </span>
        <span class="file-name" id="filename">
          No file selected
        </span>
      </label>
    </div>
    <div class="notification is-info is-light mt-4">
      <p><strong>Tip:</strong> Upload an RFP, SOW, specification document, 
         or any text describing your agency's purpose and requirements.</p>
    </div>
    
    <!-- Document Preview -->
    <div class="box mt-4" id="document-preview" style="display: none;">
      <h3 class="subtitle is-6">Document Preview</h3>
      <div class="content preview-text" style="max-height: 300px; overflow-y: auto;">
        <!-- Extracted text shown here -->
      </div>
    </div>
  </div>

  <!-- Step 2: Select Generation Areas -->
  <div class="wizard-step" id="step-select">
    <h2 class="title is-4">Step 2: Select What to Generate</h2>
    <div class="box">
      <label class="checkbox is-block mb-3">
        <input type="checkbox" name="areas" value="introduction" checked>
        <strong>📄 Introduction</strong>
        <p class="help">Generate agency background, purpose, and scope</p>
      </label>
      
      <label class="checkbox is-block mb-3">
        <input type="checkbox" name="areas" value="goals" checked>
        <strong>🎯 Goals</strong>
        <p class="help">Generate structured SMART goals catalog</p>
      </label>
      
      <label class="checkbox is-block mb-3">
        <input type="checkbox" name="areas" value="work_items" checked>
        <strong>📋 Work Items</strong>
        <p class="help">Generate work breakdown with deliverables and dependencies</p>
      </label>
      
      <label class="checkbox is-block mb-3">
        <input type="checkbox" name="areas" value="agent_types">
        <strong>🤖 Agent Types</strong>
        <p class="help">Generate agent roles and specifications</p>
      </label>
      
      <label class="checkbox is-block mb-3">
        <input type="checkbox" name="areas" value="raci_matrix">
        <strong>📊 RACI Matrix</strong>
        <p class="help">Generate responsibility assignments</p>
      </label>
    </div>
    
    <!-- Generation Options -->
    <div class="box mt-4">
      <h3 class="subtitle is-6">Generation Options</h3>
      <div class="field">
        <label class="label">Detail Level</label>
        <div class="control">
          <div class="select">
            <select name="detail_level">
              <option value="brief">Brief</option>
              <option value="standard" selected>Standard</option>
              <option value="detailed">Detailed</option>
            </select>
          </div>
        </div>
      </div>
      
      <div class="field">
        <label class="label">Formality</label>
        <div class="control">
          <div class="select">
            <select name="formality">
              <option value="casual">Casual</option>
              <option value="professional" selected>Professional</option>
              <option value="formal">Formal</option>
            </select>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Step 3: Generation Progress -->
  <div class="wizard-step" id="step-generate">
    <h2 class="title is-4">Step 3: Generating Your Agency Design</h2>
    
    <div class="generation-progress">
      <!-- Introduction -->
      <div class="box generation-item" data-area="introduction">
        <div class="level">
          <div class="level-left">
            <div class="level-item">
              <span class="icon"><i class="fas fa-file-alt"></i></span>
            </div>
            <div class="level-item">
              <strong>Introduction</strong>
            </div>
          </div>
          <div class="level-right">
            <div class="level-item">
              <span class="tag is-warning">In Progress</span>
              <span class="icon ml-2">
                <i class="fas fa-spinner fa-spin"></i>
              </span>
            </div>
          </div>
        </div>
        <progress class="progress is-small is-primary" value="65" max="100">65%</progress>
      </div>
      
      <!-- Goals -->
      <div class="box generation-item" data-area="goals">
        <div class="level">
          <div class="level-left">
            <div class="level-item">
              <span class="icon"><i class="fas fa-bullseye"></i></span>
            </div>
            <div class="level-item">
              <strong>Goals</strong>
            </div>
          </div>
          <div class="level-right">
            <div class="level-item">
              <span class="tag is-success">Complete</span>
              <span class="icon ml-2 has-text-success">
                <i class="fas fa-check-circle"></i>
              </span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Work Items -->
      <div class="box generation-item" data-area="work_items">
        <div class="level">
          <div class="level-left">
            <div class="level-item">
              <span class="icon"><i class="fas fa-tasks"></i></span>
            </div>
            <div class="level-item">
              <strong>Work Items</strong>
            </div>
          </div>
          <div class="level-right">
            <div class="level-item">
              <span class="tag is-light">Waiting</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Step 4: Review & Edit -->
  <div class="wizard-step" id="step-review">
    <h2 class="title is-4">Step 4: Review & Edit Generated Content</h2>
    
    <div class="tabs is-boxed">
      <ul>
        <li class="is-active"><a data-tab="introduction">📄 Introduction</a></li>
        <li><a data-tab="goals">🎯 Goals</a></li>
        <li><a data-tab="work-items">📋 Work Items</a></li>
        <li><a data-tab="agent-types">🤖 Agent Types</a></li>
        <li><a data-tab="raci">📊 RACI Matrix</a></li>
      </ul>
    </div>
    
    <div class="tab-content">
      <div class="columns">
        <div class="column">
          <h3 class="subtitle is-6">Source Text</h3>
          <div class="box source-text" style="max-height: 500px; overflow-y: auto;">
            <!-- Original text shown here -->
          </div>
        </div>
        <div class="column">
          <h3 class="subtitle is-6">Generated Content</h3>
          <div class="box generated-content" style="max-height: 500px; overflow-y: auto;">
            <!-- Generated content shown here (editable) -->
            <div class="buttons">
              <button class="button is-small is-primary">
                <span class="icon"><i class="fas fa-save"></i></span>
                <span>Save Changes</span>
              </button>
              <button class="button is-small is-warning">
                <span class="icon"><i class="fas fa-sync"></i></span>
                <span>Regenerate</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Wizard Navigation -->
  <div class="wizard-navigation buttons is-right mt-5">
    <button class="button" id="btn-prev">
      <span class="icon"><i class="fas fa-arrow-left"></i></span>
      <span>Previous</span>
    </button>
    <button class="button is-primary" id="btn-next">
      <span>Next</span>
      <span class="icon"><i class="fas fa-arrow-right"></i></span>
    </button>
    <button class="button is-success" id="btn-finish" style="display: none;">
      <span class="icon"><i class="fas fa-check"></i></span>
      <span>Finish & Save Agency</span>
    </button>
  </div>
</div>
```

**Backend Services**:

```go
// AIAgencyGeneratorService
type AIAgencyGeneratorService interface {
    UploadDocument(ctx context.Context, agencyID string, file io.Reader, filename string) (*SourceDocument, error)
    GenerateBatch(ctx context.Context, req GenerationRequest) (*GenerationResult, error)
    GenerateIntroduction(ctx context.Context, agencyID, sourceText string, opts GenerationOptions) (string, error)
    GenerateGoals(ctx context.Context, agencyID, sourceText string, opts GenerationOptions) ([]Goal, error)
    GenerateWorkItems(ctx context.Context, agencyID, sourceText string, opts GenerationOptions) ([]WorkItem, error)
    GenerateAgentTypes(ctx context.Context, agencyID, sourceText string, opts GenerationOptions) ([]AgentType, error)
    GenerateRACIMatrix(ctx context.Context, agencyID, sourceText string, workItems []WorkItem, opts GenerationOptions) (*RACIMatrix, error)
    GetGenerationStatus(ctx context.Context, agencyID, generationID string) (*GenerationStatus, error)
}

// Document processing utilities
type DocumentProcessor interface {
    ExtractText(file io.Reader, contentType string) (string, error)
    AnalyzeStructure(text string) (*DocumentStructure, error)
    ChunkForAI(text string, maxTokens int) ([]string, error)
}
```

**AI Prompt Templates**:

Each generation area should have specialized prompts:

1. **Introduction Generation**:
   - Extract background, purpose, scope
   - Identify stakeholders
   - Determine success criteria

2. **Goals Generation**:
   - Extract objectives from source
   - Structure as SMART goals
   - Categorize by type (Efficiency, Quality, Innovation, etc.)

3. **Work Items Generation**:
   - Break down into actionable items
   - Generate deliverables per item
   - Identify dependencies
   - Estimate effort

4. **Agent Types Generation**:
   - Identify required roles
   - Define capabilities per role
   - Specify autonomy levels
   - Determine communication patterns

5. **RACI Matrix Generation**:
   - Map work items to activities
   - Assign roles to activities
   - Ensure RACI validation rules

**Acceptance Criteria**:
- [ ] Document upload supports PDF, DOCX, TXT, MD formats
- [ ] Text extraction works correctly for all formats
- [ ] Checkbox selection allows any combination of areas
- [ ] Batch generation processes multiple areas in parallel
- [ ] Progress tracking shows real-time status
- [ ] Generated content is editable before saving
- [ ] Regeneration of individual areas works
- [ ] Side-by-side comparison view functional
- [ ] All generated content validates against schemas
- [ ] Agency Designer organized into clear sections
- [ ] Navigation between sections smooth
- [ ] Integration with existing modules seamless
- [ ] Error handling and retry logic robust
- [ ] Generation completes within reasonable time (<2 min per area)

**Dependencies**: MVP-029 (Goals Module - completed)

**Reference**: See existing introduction and goals modules for integration patterns

---

## Resource Requirements

### Team Members
- **Backend Developer**: API development, database design, security implementation
- **Frontend Developer**: UI/UX implementation, responsive design, user experience
- **DevOps Engineer**: Infrastructure setup, CI/CD, production deployment
- **QA Engineer**: Testing strategy, test automation, quality assurance

### Tools and Platforms
- **Development**: Git, Docker, VS Code/IDE of choice
- **Backend**: Node.js/Python/Go (TBD), REST/GraphQL APIs
- **Frontend**: React/Vue/Angular (TBD), CSS frameworks
- **Database**: PostgreSQL/MongoDB (TBD)
- **CI/CD**: GitHub Actions/GitLab CI (TBD)
- **Monitoring**: Basic logging and health checks

### Infrastructure
- **Hosting**: Cloud provider (AWS/GCP/Azure TBD)
- **Environments**: Development, staging, production
- **CDN**: Basic content delivery for static assets
- **SSL**: Certificate management and HTTPS enforcement

## Risk Assessment

### Identified Risks
- **Technical Debt**: Rushing MVP features may compromise code quality
- **Scope Creep**: Adding non-essential features that delay launch
- **Performance Issues**: Scalability problems under load
- **Security Vulnerabilities**: Inadequate security implementation
- **Integration Challenges**: Third-party service dependencies

### Mitigation Strategies
- **Code Reviews**: Mandatory peer review for all code changes
- **Feature Freeze**: Strict adherence to MVP scope definition
- **Load Testing**: Early performance testing with realistic data volumes
- **Security Audits**: Regular security reviews and penetration testing
- **Fallback Plans**: Alternative solutions for critical third-party dependencies

### Contingency Plans
- **MVP Scope Reduction**: Remove P2/P3 features if timeline is at risk
- **Technical Alternatives**: Backup technology choices for critical components
- **Extended Timeline**: Buffer time for unexpected complications
- **Resource Scaling**: Option to add temporary team members if needed

## MVP Success Metrics

### Technical Metrics
- **Uptime**: 99%+ availability during business hours
- **Response Time**: <2 seconds for critical user actions
- **Security**: Zero critical vulnerabilities at launch
- **Performance**: Support for 100+ concurrent users

### User Metrics
- **Registration**: >80% completion rate for sign-up flow
- **Workflow Completion**: >70% completion rate for primary user journey
- **User Retention**: >50% of users return within first week
- **Error Rate**: <5% user-facing errors

### Business Metrics
- **Timeline**: Launch within planned development window
- **Budget**: Stay within allocated development resources
- **Value Validation**: Demonstrate core value proposition to target users
- **Market Readiness**: Receive positive feedback from beta users

## Workflow Integration

### Task Management Process
1. **Task Assignment**: Pick tasks based on priority (P0 first) and dependencies
2. **Implementation**: Update "Status" column as work progresses (Not Started → In Progress → Testing → Complete)
3. **Completion Process** (MANDATORY):
   - Create detailed coding session document in `coding_sessions/` using format: `{TaskID}_{description}.md`
   - Add completed task to summary table in `mvp_done.md` with completion date
   - Remove completed task from this active `mvp.md` file
   - Update any dependent task references
   - Merge feature branch to main:
     ```bash
     # Merge when complete and tested
     git checkout main
     git merge feature/MVP-XXX_description
     git branch -d feature/MVP-XXX_description
     git push origin main
     ```
4. **Dependencies**: Ensure prerequisite tasks are completed before starting dependent work

### Branch Management (MANDATORY)
For each new task:
```bash
# Create feature branch
git checkout -b feature/MVP-XXX_description

# Work on task implementation
# ... development work ...

# Build validation before merge
# - Follow coding standards
# - Run linting and validation tools
# - Verify code follows established patterns
# - Check for deprecated API usage
# - Remove unused code/imports/variables
# - Run build processes and tests
# - Fix any build errors or warnings

# Merge when complete and tested
git checkout main
git merge feature/MVP-XXX_description
git branch -d feature/MVP-XXX_description
```

### Repository Structure
```
/workspaces/CodeValdCortex/
├── documents/3-SofwareDevelopment/
│   ├── mvp.md                    # This file - Active tasks only
│   ├── mvp_done.md              # Completed tasks archive
│   └── coding_sessions/         # Detailed implementation logs
├── [project code structure]     # Implementation code
└── [other project folders]      # Additional project resources
```

---

**Note**: This document contains only active and pending tasks. All completed tasks are moved to `mvp_done.md` to maintain a clean, actionable backlog.