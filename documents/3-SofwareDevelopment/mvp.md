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
| MVP-023 | AI Agent Creator      | Implement AI-powered conversational interface for creating agents. AI asks questions, resolves details, and generates complete agent configurations through natural language dialogue | Not Started | P1       | Medium | Go, Templ, AI/LLM, Frontend Dev | MVP-022      |
| MVP-025 | AI Agency Designer    | Advanced AI-driven agency design tool that brainstorms agency structure, creates agent types, defines relationships, and generates complete agency architecture through intelligent conversation | ✅ Complete | P1       | High   | Go, AI/LLM, System Design | MVP-024      |

## Authentication & Security Tasks (P2 - Important)

| Task ID | Title                     | Description                                                | Status      | Priority | Effort | Skills Required       | Dependencies |
| ------- | ------------------------- | ---------------------------------------------------------- | ----------- | -------- | ------ | --------------------- | ------------ |
| MVP-026 | Basic User Authentication | Implement user registration, login, and session management | Not Started | P2       | Medium | Backend Dev, Security | MVP-014      |
| MVP-027 | Security Implementation   | Add input validation, HTTPS, and basic security headers    | Not Started | P2       | Medium | Security, Backend Dev | MVP-026      |
| MVP-028 | Access Control System     | Implement role-based access control for agent operations   | Not Started | P2       | Low    | Backend Dev, Security | MVP-027      |

## Agency Management Feature (P1 - Critical)

*All agency management tasks completed. See `mvp_done.md` for details.*

---

### MVP-025: AI Agency Designer

**Objective**: Create an advanced AI-powered tool that helps users design complete agency architectures through intelligent conversation. The AI brainstorms agency structure, creates agent types, defines relationships, and generates a comprehensive agency blueprint.

**Key Deliverables**:

1. **AI-Driven Agency Design Conversation**:
   - Conversational interface for agency design
   - AI asks strategic questions about:
     * Business problem and objectives
     * Key stakeholders and actors
     * Processes and workflows
     * Data sources and integrations
     * Scale and performance requirements
   - Brainstorming session for agent types
   - Relationship mapping between agents
   - Architecture visualization and validation

2. **AI Design Process**:
   ```
   User: "I want to build a warehouse management system"
   
   AI: "Excellent! Let's design your warehouse management agency together.
        First, help me understand your warehouse operations:
        
        1. What are the main areas in your warehouse?
           (e.g., receiving, storage, picking, packing, shipping)"
   
   User: "We have receiving dock, three storage zones, picking area, 
          packing station, and shipping dock"
   
   AI: "Perfect! I'm thinking we'll need several agent types:
        
        🚚 Receiving Agent - Manages incoming shipments
        📦 Storage Agent (x3) - One per zone, tracks inventory
        🎯 Picking Agent - Coordinates order fulfillment  
        📮 Packing Agent - Manages packing operations
        🚛 Shipping Agent - Handles outbound logistics
        🤖 Orchestrator Agent - Coordinates workflow
        
        Does this structure make sense? Any other operations?"
   
   User: "We also need inventory tracking and quality control"
   
   AI: "Great additions! Let me add:
        
        📊 Inventory Manager - Real-time stock tracking
        ✅ Quality Control Agent - Inspection workflows
        
        Now, let's define how they communicate:
        - Receiving → Inventory Manager (new stock)
        - Inventory → Storage Agents (placement)
        - Orders → Orchestrator → Picking → Packing → Shipping
        
        Should I generate the complete agency specification?"
   ```

3. **AI Intelligence Features**:
   - **Domain Analysis**: Understand business domain and requirements
   - **Agent Type Generation**: Create custom agent types with capabilities
   - **Workflow Design**: Map processes to agent interactions
   - **Relationship Mapping**: Define pub/sub topics and message flows
   - **Architecture Validation**: Check for gaps and redundancies
   - **Best Practices**: Suggest proven patterns for the domain
   - **Scalability Analysis**: Recommend agent counts and distribution

4. **Generated Artifacts**:
   - Complete agency specification (JSON/YAML)
   - Agent type definitions with schemas
   - Communication topology diagram
   - Workflow documentation
   - Deployment recommendations
   - Sample agent configurations
   - Initial agent instances

5. **Interactive Design Features**:
   - Visual architecture diagram (nodes = agents, edges = communication)
   - Drag-and-drop to reorganize
   - Click agents to edit properties
   - Add/remove agent types on the fly
   - Real-time validation and suggestions
   - Export/import design blueprints

6. **Backend AI Integration**:
   - Advanced LLM prompting for system design
   - Domain-specific knowledge bases (logistics, healthcare, etc.)
   - Architecture pattern library
   - Agent type template generation
   - Automatic relationship inference
   - Code generation for agent types

**UI Mockup**:
```html
<!-- AI Agency Designer Interface -->
<div class="agency-designer-container">
  <!-- Left Panel: AI Conversation -->
  <div class="design-chat-panel">
    <header>
      <h2>🧠 AI Agency Designer</h2>
      <p>Let's design your agency architecture together</p>
    </header>
    
    <div class="chat-messages">
      <!-- AI introduces design process -->
      <div class="message ai-message">
        <div class="avatar">🤖</div>
        <div class="content">
          <p>Hi! I'm your AI Agency Designer. I'll help you create a 
             complete multi-agent architecture for your use case.</p>
          <p>Tell me about the system you want to build. What problem 
             are you solving?</p>
        </div>
      </div>
      
      <!-- Design conversation continues... -->
    </div>
    
    <div class="chat-input">
      <input type="text" placeholder="Describe your requirements...">
      <button>Send</button>
    </div>
  </div>
  
  <!-- Center Panel: Visual Architecture -->
  <div class="architecture-canvas">
    <div class="canvas-toolbar">
      <button>🔍 Zoom</button>
      <button>↻ Auto-layout</button>
      <button>💾 Save Design</button>
      <button>🚀 Deploy Agency</button>
    </div>
    
    <svg class="architecture-diagram">
      <!-- Agent nodes -->
      <g class="agent-node" data-type="receiving">
        <circle cx="100" cy="100" r="40" fill="#3273dc"/>
        <text x="100" y="105">🚚 Receiving</text>
      </g>
      
      <g class="agent-node" data-type="storage">
        <circle cx="250" cy="100" r="40" fill="#48c774"/>
        <text x="250" y="105">📦 Storage-A</text>
      </g>
      
      <!-- Communication links -->
      <line x1="140" y1="100" x2="210" y2="100" 
            stroke="#666" stroke-width="2" marker-end="url(#arrow)"/>
    </svg>
  </div>
  
  <!-- Right Panel: Agent Details -->
  <div class="agent-details-panel">
    <h3>Agent Type: Receiving Agent</h3>
    
    <div class="agent-properties">
      <h4>Capabilities</h4>
      <ul>
        <li>✓ Scan incoming shipments</li>
        <li>✓ Validate against purchase orders</li>
        <li>✓ Create inventory records</li>
        <li>✓ Notify storage agents</li>
      </ul>
      
      <h4>Configuration</h4>
      <pre>{
  "type": "receiving",
  "interval": 60,
  "auto_scan": true
}</pre>
      
      <h4>Publishes To</h4>
      <ul>
        <li>inventory.new_stock</li>
        <li>storage.placement_request</li>
      </ul>
      
      <h4>Subscribes To</h4>
      <ul>
        <li>orders.purchase_orders</li>
        <li>shipping.arrivals</li>
      </ul>
    </div>
    
    <button class="button is-primary">Save Agent Type</button>
  </div>
</div>
```

**AI Design Conversation Patterns**:

1. **Requirements Gathering**:
   ```
   AI: "What are your main business processes?"
   AI: "Who are the key actors in your system?"
   AI: "What data do you need to track?"
   ```

2. **Agent Type Brainstorming**:
   ```
   AI: "I suggest these agent types: [list]"
   AI: "Should we add agents for [process]?"
   AI: "How many instances of [agent] do you need?"
   ```

3. **Workflow Mapping**:
   ```
   AI: "When [event] happens, which agents should react?"
   AI: "Should [Agent A] communicate directly with [Agent B]?"
   ```

4. **Architecture Validation**:
   ```
   AI: "I notice no agent handles [scenario]. Should we add one?"
   AI: "Multiple agents manage [resource]. Should we consolidate?"
   ```

**Acceptance Criteria**:
- ✅ AI understands business requirements
- ✅ Suggests appropriate agent types
- ✅ Maps workflows to agent interactions
- ✅ Generates complete agency specification
- ✅ Creates agent type definitions
- ✅ Visual architecture diagram
- ✅ Validates design completeness
- ✅ Exports agency blueprint
- ✅ Can deploy designed agency
- ✅ Saves design for later editing

**API Specification**:
```json
POST /api/v1/agencies/design/chat
{
  "message": "I want to build a warehouse management system",
  "design_id": "design-uuid-123",
  "context": {...}
}

Response:
{
  "message": "Let's design your warehouse agency...",
  "suggestions": [...],
  "draft_architecture": {
    "agent_types": [
      {
        "name": "receiving_agent",
        "capabilities": [...],
        "publishes": [...],
        "subscribes": [...]
      }
    ],
    "relationships": [...],
    "workflows": [...]
  }
}

POST /api/v1/agencies/design/deploy
{
  "design_id": "design-uuid-123",
  "agency_id": "UC-WMS-001"
}
```

**Technical Implementation**:
```
/workspaces/CodeValdCortex/
├── internal/ai/
│   ├── agency_designer.go          # AI design orchestration
│   ├── architecture_analyzer.go    # Design validation
│   ├── agent_type_generator.go     # Generate agent specs
│   └── prompts/
│       └── agency_design.txt       # Design prompts
├── internal/web/
│   ├── pages/
│   │   └── agency_designer.templ   # Designer UI
│   └── handlers/
│       └── agency_design_handler.go
└── static/
    ├── css/
    │   └── agency-designer.css
    └── js/
        ├── architecture-canvas.js   # Visual diagram
        └── design-chat.js          # Chat interface
```

**Dependencies**:
- MVP-024 (create agency form)
- LLM API access with advanced reasoning
- Graph visualization library
- Agent type schema definitions

**Future Enhancements**:
- Import existing agency for redesign
- Multi-user collaborative design
- Design templates library
- Simulation mode (test before deploy)
- Cost estimation
- Performance prediction
- Auto-scaling recommendations

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