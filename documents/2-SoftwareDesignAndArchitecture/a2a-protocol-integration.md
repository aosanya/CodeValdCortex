# A2A Protocol Integration - Technical Specification
## CodeValdCortex Enterprise Multi-Agent AI Orchestration Platform

**Document Version**: 1.0  
**Date**: November 11, 2025  
**Status**: Draft - For Review  
**Epic**: A2A Integration (v1.2)

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Integration Architecture](#2-integration-architecture)
3. [Technical Requirements](#3-technical-requirements)
4. [Implementation Phases](#4-implementation-phases)
5. [API Specifications](#5-api-specifications)
6. [Security & Compliance](#6-security--compliance)
7. [Testing Strategy](#7-testing-strategy)
8. [Deployment Guide](#8-deployment-guide)
9. [Appendices](#9-appendices)

---

## 1. Executive Summary

### 1.1 Purpose

This specification defines the integration of the Agent2Agent (A2A) Protocol into CodeValdCortex, enabling seamless interoperability between internal agents and external A2A-compatible agents across vendor boundaries.

### 1.2 Business Value

**Strategic Positioning**: Transform CodeValdCortex from a single-vendor orchestration platform to the **"Kubernetes of Multi-Vendor AI Agents"**

**Key Benefits**:
- **Vendor Independence**: Orchestrate agents from multiple vendors without lock-in
- **Ecosystem Access**: Tap into the growing A2A agent marketplace
- **Competitive Differentiation**: First enterprise-grade A2A orchestration platform
- **Future-Proof Architecture**: Align with Linux Foundation open standards

**Target ROI**:
- 40% reduction in custom integration development costs
- 60% faster time-to-value for new agent capabilities
- 3x expansion of addressable agent ecosystem

### 1.3 Scope

**In Scope**:
- A2A server implementation (expose CodeValdCortex agents)
- A2A client implementation (consume external agents)
- Agent Card generation and discovery
- Task delegation and lifecycle management
- Security, authentication, and audit integration
- Multi-platform orchestration capabilities

**Out of Scope** (Future Phases):
- A2A marketplace/registry hosting
- Advanced agent negotiation protocols
- Real-time bidirectional streaming (SSE implementation in Phase 1 is one-way)
- Payment/billing integration for commercial agents

### 1.4 Dependencies

**Completed Prerequisites**:
- ✅ MVP-008: Agent Registry System
- ✅ MVP-010: Agent Communication Layer
- ✅ MVP-012: Orchestration Engine
- ✅ MVP-029: Goals Module
- ✅ MVP-044: Roles Module

**Required Prerequisites** (must complete before A2A integration):
- 📋 MVP-028: Access Control System (RBAC)
- 📋 MVP-026: User Authentication
- 📋 MVP-027: Security Implementation

---

## 2. Integration Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    CodeValdCortex Platform                      │
│                                                                 │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐│
│  │   Internal   │      │     A2A      │      │   External   ││
│  │   Agents     │◄────►│  Gateway     │◄────►│   Agents     ││
│  │  (Go native) │      │  (HTTP/SSE)  │      │ (A2A Proto)  ││
│  └──────────────┘      └──────────────┘      └──────────────┘│
│         │                      │                      │        │
│         ▼                      ▼                      ▼        │
│  ┌────────────────────────────────────────────────────────┐  │
│  │         Enhanced Orchestration Engine                  │  │
│  │  - Internal routing (Go channels)                      │  │
│  │  - External routing (A2A protocol)                     │  │
│  │  - Intelligent agent selection                         │  │
│  │  - Task lifecycle management                           │  │
│  └────────────────────────────────────────────────────────┘  │
│         │                      │                      │        │
│         ▼                      ▼                      ▼        │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐│
│  │   Agent      │      │  A2A Agent   │      │   Security   ││
│  │   Registry   │      │  Card Store  │      │  & Audit     ││
│  │  (ArangoDB)  │      │  (ArangoDB)  │      │  (ArangoDB)  ││
│  └──────────────┘      └──────────────┘      └──────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │   External A2A Ecosystem      │
              │  - Vendor A agents            │
              │  - Vendor B agents            │
              │  - Open source agents         │
              └───────────────────────────────┘
```

### 2.2 Component Breakdown

#### 2.2.1 A2A Gateway Service

**Responsibility**: Bidirectional translation between internal Go-native communication and external A2A protocol

**Key Features**:
- HTTP/HTTPS server for incoming A2A requests
- HTTP/SSE client for outgoing A2A requests
- Protocol translation (Go channels ↔ JSON-RPC)
- Connection pooling and rate limiting
- Circuit breaker for external agent failures

**Technology Stack**:
- Go 1.21+ with `net/http` and `nhooyr.io/websocket` for SSE
- JSON-RPC 2.0 implementation
- OpenTelemetry for distributed tracing

#### 2.2.2 Agent Card Manager

**Responsibility**: Generate, store, and serve A2A-compliant Agent Cards for internal agents

**Schema** (A2A v1.0 compliant):
```json
{
  "name": "string",
  "description": "string",
  "version": "string",
  "service_endpoint": "string (URL)",
  "supported_modalities": ["text", "image", "audio"],
  "capabilities": ["string"],
  "authentication": {
    "type": "oauth2 | jwt | api_key",
    "endpoint": "string (URL)",
    "required_scopes": ["string"]
  },
  "metadata": {
    "autonomy_level": "L0|L1|L2|L3|L4",
    "agency_id": "string",
    "compliance_tags": ["SOC2", "HIPAA", "GDPR"]
  }
}
```

#### 2.2.3 Discovery Service

**Responsibility**: Discover and catalog external A2A-compatible agents

**Discovery Methods**:
1. **Manual Registration**: Admin adds external agent endpoint
2. **DNS-SD (Service Discovery)**: Automatic discovery via DNS TXT records
3. **Registry Integration**: Query external A2A registries/marketplaces
4. **Agent Referrals**: Agents recommend other agents

**Storage**:
- ArangoDB collection: `a2a_external_agents`
- Cached in Redis for performance (TTL: 5 minutes)
- Periodic health checks (every 60 seconds)

#### 2.2.4 Enhanced Orchestration Engine

**New Capabilities**:
- **Hybrid Routing**: Evaluate internal vs external agent suitability
- **Cost-Aware Selection**: Consider latency, cost, trust score
- **Fallback Chains**: Retry with alternative agents on failure
- **Transaction Management**: Coordinate multi-agent tasks across platforms

**Agent Selection Algorithm**:
```
Score = (CapabilityMatch × 0.4) + (TrustScore × 0.3) + (CostEfficiency × 0.2) + (Latency × 0.1)

Where:
- CapabilityMatch: % of required capabilities the agent supports (0-1)
- TrustScore: Historical success rate + compliance certifications (0-1)
- CostEfficiency: Normalized cost per task (0-1, inverted)
- Latency: Response time score (0-1, inverted)
```

### 2.3 Data Model Extensions

#### 2.3.1 Enhanced Agent Registry Schema

```go
// pkg/registry/agent.go
type Agent struct {
    // Existing fields
    ID            string    `json:"id"`
    Name          string    `json:"name"`
    AutonomyLevel string    `json:"autonomy_level"`
    
    // New A2A fields
    A2AEnabled    bool      `json:"a2a_enabled"`
    A2ACard       *A2ACard  `json:"a2a_card,omitempty"`
    IsExternal    bool      `json:"is_external"`
    
    // For external agents
    ExternalEndpoint string  `json:"external_endpoint,omitempty"`
    LastHealthCheck  time.Time `json:"last_health_check,omitempty"`
    HealthStatus     string  `json:"health_status,omitempty"` // healthy, degraded, down
}

type A2ACard struct {
    Name              string             `json:"name"`
    Description       string             `json:"description"`
    Version           string             `json:"version"`
    ServiceEndpoint   string             `json:"service_endpoint"`
    SupportedModalities []string         `json:"supported_modalities"`
    Capabilities      []string           `json:"capabilities"`
    Authentication    A2AAuthentication  `json:"authentication"`
    Metadata          map[string]string  `json:"metadata"`
}

type A2AAuthentication struct {
    Type           string   `json:"type"` // oauth2, jwt, api_key
    Endpoint       string   `json:"endpoint,omitempty"`
    RequiredScopes []string `json:"required_scopes,omitempty"`
}
```

#### 2.3.2 A2A Task Schema

```go
// pkg/a2a/task.go
type A2ATask struct {
    TaskID          string                 `json:"task_id"`
    AgentID         string                 `json:"agent_id"`
    Input           map[string]interface{} `json:"input"`
    RequiredCapabilities []string          `json:"required_capabilities"`
    Priority        string                 `json:"priority"` // low, medium, high, critical
    TimeoutSeconds  int                    `json:"timeout_seconds"`
    
    // Task lifecycle
    Status          string                 `json:"status"` // pending, running, completed, failed
    CreatedAt       time.Time              `json:"created_at"`
    CompletedAt     *time.Time             `json:"completed_at,omitempty"`
    
    // Results
    Output          map[string]interface{} `json:"output,omitempty"`
    Error           *A2AError              `json:"error,omitempty"`
    
    // Tracking
    SourceAgentID   string                 `json:"source_agent_id"` // Internal agent that initiated
    IsExternal      bool                   `json:"is_external"`     // Task delegated to external agent
}

type A2AError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

#### 2.3.3 ArangoDB Collections

```javascript
// New collections for A2A support

// Collection: a2a_agent_cards
{
  "_key": "agent_001",
  "agent_id": "risk-analysis-agent-001",
  "card": { /* A2A Card JSON */ },
  "created_at": "2025-11-11T10:00:00Z",
  "updated_at": "2025-11-11T10:00:00Z"
}

// Collection: a2a_external_agents
{
  "_key": "ext_agent_001",
  "name": "FraudDetector Pro",
  "vendor": "VendorX",
  "service_endpoint": "https://vendor-x.com/agents/fraud-detector",
  "card": { /* A2A Card JSON */ },
  "trust_score": 0.95,
  "health_status": "healthy",
  "last_health_check": "2025-11-11T10:30:00Z",
  "registered_at": "2025-11-01T00:00:00Z"
}

// Collection: a2a_task_logs
{
  "_key": "task_log_001",
  "task_id": "task_12345",
  "internal_agent_id": "risk-analysis-agent-001",
  "external_agent_id": "ext_agent_001",
  "request_payload": { /* JSON */ },
  "response_payload": { /* JSON */ },
  "status": "completed",
  "latency_ms": 450,
  "created_at": "2025-11-11T10:35:00Z",
  "completed_at": "2025-11-11T10:35:00.450Z"
}

// Edge collection: a2a_task_dependencies
{
  "_from": "tasks/task_parent",
  "_to": "tasks/task_child",
  "delegation_type": "a2a_external",
  "created_at": "2025-11-11T10:35:00Z"
}
```

---

## 3. Technical Requirements

### 3.1 Functional Requirements

#### FR-A2A-001: Agent Card Generation
**Priority**: P0 (Must Have)

**Description**: System shall automatically generate A2A-compliant Agent Cards for all internal agents

**Acceptance Criteria**:
- Agent Cards generated on agent registration
- Cards include all required A2A v1.0 fields
- Cards served via HTTP endpoint: `GET /api/v1/agents/{agent_id}/a2a-card`
- Cards updated automatically when agent configuration changes
- Support for custom metadata fields

**Technical Details**:
```go
// pkg/a2a/card_generator.go
func GenerateA2ACard(agent *registry.Agent, config *A2AConfig) (*A2ACard, error) {
    card := &A2ACard{
        Name:        agent.Name,
        Description: agent.Description,
        Version:     agent.Version,
        ServiceEndpoint: fmt.Sprintf("%s/api/v1/a2a/agents/%s", config.BaseURL, agent.ID),
        SupportedModalities: inferModalities(agent),
        Capabilities: agent.Capabilities,
        Authentication: A2AAuthentication{
            Type:           "oauth2",
            Endpoint:       fmt.Sprintf("%s/oauth/token", config.BaseURL),
            RequiredScopes: []string{"agent.execute"},
        },
        Metadata: map[string]string{
            "autonomy_level":  agent.AutonomyLevel,
            "agency_id":       agent.AgencyID,
            "compliance_tags": strings.Join(agent.ComplianceTags, ","),
        },
    }
    
    return card, validateA2ACard(card)
}
```

---

#### FR-A2A-002: External Agent Discovery
**Priority**: P0 (Must Have)

**Description**: System shall discover and register external A2A-compatible agents

**Acceptance Criteria**:
- Manual registration via Admin UI or API
- Health checks every 60 seconds
- Automatic de-registration after 3 consecutive health check failures
- Discovery results cached in Redis (TTL: 5 minutes)
- Support for paginated discovery results

**API Endpoints**:
```
POST   /api/v1/a2a/external-agents          # Register external agent
GET    /api/v1/a2a/external-agents          # List discovered agents
GET    /api/v1/a2a/external-agents/{id}     # Get agent details
DELETE /api/v1/a2a/external-agents/{id}     # Deregister agent
POST   /api/v1/a2a/external-agents/{id}/health # Manual health check
```

**Request/Response**:
```json
// POST /api/v1/a2a/external-agents
{
  "service_endpoint": "https://vendor-x.com/agents/fraud-detector",
  "name": "FraudDetector Pro",
  "vendor": "VendorX",
  "metadata": {
    "cost_per_task": "0.05",
    "sla_guarantee": "99.9%"
  }
}

// Response 201 Created
{
  "id": "ext_agent_001",
  "name": "FraudDetector Pro",
  "status": "healthy",
  "card": { /* A2A Card JSON */ },
  "registered_at": "2025-11-11T10:00:00Z"
}
```

---

#### FR-A2A-003: Task Delegation
**Priority**: P0 (Must Have)

**Description**: System shall delegate tasks to external A2A-compatible agents

**Acceptance Criteria**:
- Support for synchronous and asynchronous task execution
- Automatic retries on transient failures (3 attempts, exponential backoff)
- Timeout enforcement (configurable, default: 30 seconds)
- Complete audit trail of all external task delegations
- Fallback to internal agents on external agent failure

**Flow**:
```
1. Internal Agent → Orchestration Engine: Task request
2. Orchestration Engine: Evaluate internal vs external agents
3. If External Selected:
   a. Orchestration Engine → A2A Gateway: Delegate task
   b. A2A Gateway → External Agent: HTTP POST /tasks
   c. External Agent → A2A Gateway: Task result (sync or async)
   d. A2A Gateway → Orchestration Engine: Task result
   e. Orchestration Engine → Internal Agent: Task result
4. Orchestration Engine: Log to a2a_task_logs collection
```

**Task Execution API**:
```json
// POST /api/v1/a2a/agents/{agent_id}/tasks
{
  "task_id": "task_12345",
  "input": {
    "transaction_id": "txn_98765",
    "amount": 50000,
    "merchant": "Suspicious Store"
  },
  "required_capabilities": ["fraud-detection"],
  "priority": "high",
  "timeout_seconds": 30
}

// Response 200 OK (synchronous)
{
  "task_id": "task_12345",
  "status": "completed",
  "output": {
    "fraud_score": 0.87,
    "risk_level": "high",
    "reasons": ["unusual_amount", "new_merchant"]
  },
  "completed_at": "2025-11-11T10:35:00.450Z"
}

// Response 202 Accepted (asynchronous)
{
  "task_id": "task_12345",
  "status": "running",
  "status_url": "/api/v1/a2a/tasks/task_12345/status"
}
```

---

#### FR-A2A-004: Intelligent Agent Selection
**Priority**: P1 (Should Have)

**Description**: System shall automatically select the best agent (internal or external) for each task

**Acceptance Criteria**:
- Selection algorithm considers capability match, trust score, cost, latency
- Configurable weights for selection criteria
- Admin override to force internal or external execution
- Selection decision logged for auditing
- A/B testing support for selection algorithm optimization

**Selection Configuration**:
```yaml
# config/a2a_selection.yaml
selection:
  algorithm: weighted_score
  weights:
    capability_match: 0.4
    trust_score: 0.3
    cost_efficiency: 0.2
    latency: 0.1
  
  preferences:
    prefer_internal: true          # Prefer internal agents when scores are equal
    cost_threshold: 1.0            # Max cost per task (USD)
    latency_threshold_ms: 5000     # Max acceptable latency
    min_trust_score: 0.7           # Minimum trust score for external agents
  
  fallback:
    enabled: true
    max_retries: 3
    retry_delay_ms: 1000
    prefer_internal_on_failure: true
```

---

#### FR-A2A-005: Authentication & Authorization
**Priority**: P0 (Must Have)

**Description**: System shall enforce secure authentication for all A2A communications

**Acceptance Criteria**:
- Support OAuth 2.0, JWT, and API Key authentication
- Integration with existing RBAC system (MVP-028)
- Scoped access tokens (read, write, execute)
- Token rotation and expiration enforcement
- Audit log of all authentication attempts

**Supported Authentication Types**:

1. **OAuth 2.0 (Preferred)**:
```
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&
client_id=external_agent_001&
client_secret=secret&
scope=agent.execute
```

2. **JWT Bearer Token**:
```
POST /api/v1/a2a/agents/{agent_id}/tasks
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

3. **API Key** (For backward compatibility):
```
POST /api/v1/a2a/agents/{agent_id}/tasks
X-API-Key: cvxc_sk_1234567890abcdef
```

**Token Scopes**:
- `agent.read`: Read agent cards and discovery
- `agent.execute`: Execute tasks on agents
- `agent.admin`: Manage agent configuration

---

### 3.2 Non-Functional Requirements

#### NFR-A2A-001: Performance
- **Latency**: P95 latency for A2A task delegation < 500ms (excluding external agent execution time)
- **Throughput**: Support 1000 concurrent A2A tasks
- **Connection Pooling**: Reuse HTTP connections to external agents

#### NFR-A2A-002: Reliability
- **Availability**: 99.9% uptime for A2A Gateway
- **Circuit Breaker**: Open after 5 consecutive failures, half-open after 30 seconds
- **Retry Logic**: Exponential backoff (1s, 2s, 4s) for transient failures

#### NFR-A2A-003: Security
- **Encryption**: TLS 1.3 for all external communications
- **Input Validation**: Validate all incoming A2A requests against JSON schema
- **Rate Limiting**: 100 requests/minute per external agent

#### NFR-A2A-004: Observability
- **Distributed Tracing**: OpenTelemetry spans for all A2A operations
- **Metrics**: Prometheus metrics for latency, success rate, external agent health
- **Logging**: Structured logs (JSON) for all A2A events

---

## 4. Implementation Phases

### Phase 1: Foundation (MVP-A2A-001 to MVP-A2A-003)
**Duration**: 3 weeks  
**Effort**: 120 hours

#### MVP-A2A-001: A2A Agent Card Generator
**Effort**: 40 hours  
**Priority**: P0

**Tasks**:
1. Design A2A Card schema and validation (4h)
2. Implement card generator service (12h)
3. Add HTTP endpoint to serve cards (8h)
4. Integrate with agent registry (8h)
5. Unit and integration tests (8h)

**Deliverables**:
- `pkg/a2a/card_generator.go`
- `pkg/a2a/card_validator.go`
- API endpoint: `GET /api/v1/agents/{id}/a2a-card`
- ArangoDB collection: `a2a_agent_cards`

**Testing**:
```bash
# Test card generation
curl http://localhost:8082/api/v1/agents/risk-001/a2a-card

# Expected response
{
  "name": "Risk Analysis Agent",
  "version": "1.0.0",
  "service_endpoint": "http://localhost:8082/api/v1/a2a/agents/risk-001",
  ...
}
```

---

#### MVP-A2A-002: External Agent Registry
**Effort**: 48 hours  
**Priority**: P0

**Tasks**:
1. Design external agent data model (4h)
2. Implement registration API (12h)
3. Build health check system (12h)
4. Create discovery UI in Agency Designer (12h)
5. Testing and documentation (8h)

**Deliverables**:
- `pkg/a2a/external_registry.go`
- `pkg/a2a/health_checker.go`
- API endpoints for CRUD operations
- Admin UI for external agent management
- ArangoDB collection: `a2a_external_agents`

**UI Mockup**:
```
┌─────────────────────────────────────────────────────────┐
│ External Agents                          [+ Register]   │
├─────────────────────────────────────────────────────────┤
│ Name              Vendor   Status    Last Check  Actions│
│ FraudDetector Pro VendorX  ●Healthy  2m ago     [Edit]  │
│ SentimentAnalyzer VendorY  ●Healthy  5m ago     [Edit]  │
│ DataEnricher      VendorZ  ⚠Degraded 1m ago     [Edit]  │
└─────────────────────────────────────────────────────────┘
```

---

#### MVP-A2A-003: A2A Gateway Service
**Effort**: 32 hours  
**Priority**: P0

**Tasks**:
1. Design gateway architecture (4h)
2. Implement HTTP/SSE server (12h)
3. Implement HTTP/SSE client (8h)
4. Add protocol translation layer (4h)
5. Testing and benchmarking (4h)

**Deliverables**:
- `pkg/a2a/gateway/server.go`
- `pkg/a2a/gateway/client.go`
- `pkg/a2a/gateway/protocol.go`
- Docker image: `codevaldcortex-a2a-gateway:v1.0`

**Architecture**:
```go
// pkg/a2a/gateway/server.go
type A2AGatewayServer struct {
    httpServer    *http.Server
    authService   *auth.Service
    orchestrator  *orchestration.Engine
    metrics       *prometheus.Registry
}

func (s *A2AGatewayServer) HandleTaskRequest(w http.ResponseWriter, r *http.Request) {
    // 1. Authenticate request
    agentID, err := s.authService.AuthenticateA2ARequest(r)
    
    // 2. Parse task request
    var taskReq A2ATaskRequest
    json.NewDecoder(r.Body).Decode(&taskReq)
    
    // 3. Delegate to internal orchestration
    result := s.orchestrator.ExecuteTask(agentID, taskReq)
    
    // 4. Return response
    json.NewEncoder(w).Encode(result)
}
```

---

### Phase 2: Core Functionality (MVP-A2A-004 to MVP-A2A-006)
**Duration**: 4 weeks  
**Effort**: 160 hours

#### MVP-A2A-004: Task Delegation System
**Effort**: 64 hours  
**Priority**: P0

**Tasks**:
1. Design task lifecycle management (8h)
2. Implement synchronous task execution (16h)
3. Implement asynchronous task execution with SSE (16h)
4. Add retry and timeout logic (12h)
5. Create audit logging system (8h)
6. Testing and performance optimization (4h)

**Deliverables**:
- `pkg/a2a/task_manager.go`
- `pkg/a2a/task_executor.go`
- API endpoints for task submission and status
- ArangoDB collection: `a2a_task_logs`

---

#### MVP-A2A-005: Enhanced Orchestration
**Effort**: 56 hours  
**Priority**: P1

**Tasks**:
1. Design agent selection algorithm (8h)
2. Implement scoring system (16h)
3. Add fallback and retry logic (12h)
4. Integrate with existing orchestration engine (12h)
5. Create selection configuration UI (4h)
6. Testing and tuning (4h)

**Deliverables**:
- `pkg/orchestration/a2a_orchestrator.go`
- `pkg/orchestration/agent_selector.go`
- Configuration file: `config/a2a_selection.yaml`
- Metrics dashboard for agent selection

---

#### MVP-A2A-006: Security & Compliance
**Effort**: 40 hours  
**Priority**: P0

**Tasks**:
1. Implement OAuth 2.0 server (16h)
2. Integrate JWT validation (8h)
3. Add API key management (8h)
4. Create compliance audit reports (4h)
5. Security testing and penetration testing (4h)

**Deliverables**:
- `pkg/a2a/auth/oauth.go`
- `pkg/a2a/auth/jwt.go`
- `pkg/a2a/auth/apikey.go`
- OAuth endpoints: `/oauth/token`, `/oauth/authorize`
- Compliance report generator

---

### Phase 3: Production Readiness (MVP-A2A-007 to MVP-A2A-009)
**Duration**: 3 weeks  
**Effort**: 120 hours

#### MVP-A2A-007: Monitoring & Observability
**Effort**: 40 hours  
**Priority**: P1

**Tasks**:
1. Add Prometheus metrics for A2A operations (12h)
2. Create Grafana dashboards (12h)
3. Implement distributed tracing (8h)
4. Add structured logging (4h)
5. Create alerting rules (4h)

**Metrics**:
```
# Counter: Total A2A tasks executed
a2a_tasks_total{agent_type="internal|external", status="success|failure"}

# Histogram: Task execution latency
a2a_task_duration_seconds{agent_type="internal|external"}

# Gauge: Active external agents
a2a_external_agents_active{status="healthy|degraded|down"}

# Counter: Authentication attempts
a2a_auth_attempts_total{method="oauth2|jwt|apikey", status="success|failure"}
```

**Grafana Dashboard**:
```
┌────────────────────────────────────────────────┐
│ A2A Operations Dashboard                       │
├────────────────────────────────────────────────┤
│ ┌─────────────────┐  ┌─────────────────┐      │
│ │ Tasks/sec       │  │ Avg Latency     │      │
│ │      125        │  │     450ms       │      │
│ └─────────────────┘  └─────────────────┘      │
│                                                │
│ ┌─────────────────────────────────────────┐  │
│ │ Task Success Rate (Last 24h)            │  │
│ │ █████████████████████████████░░ 98.5%   │  │
│ └─────────────────────────────────────────┘  │
│                                                │
│ ┌─────────────────────────────────────────┐  │
│ │ External Agent Health                    │  │
│ │ ● Healthy: 8   ⚠ Degraded: 1   ● Down: 0│  │
│ └─────────────────────────────────────────┘  │
└────────────────────────────────────────────────┘
```

---

#### MVP-A2A-008: Performance Optimization
**Effort**: 40 hours  
**Priority**: P1

**Tasks**:
1. Implement connection pooling for external agents (12h)
2. Add caching layer for agent cards and discovery (8h)
3. Optimize serialization/deserialization (8h)
4. Load testing and benchmarking (8h)
5. Performance tuning based on results (4h)

**Deliverables**:
- Connection pool manager
- Redis cache integration
- Performance benchmarks
- Optimization report

**Performance Targets**:
- P95 latency < 500ms
- 1000 concurrent tasks
- 100 external agents registered
- < 10MB memory per connection

---

#### MVP-A2A-009: Documentation & Developer Experience
**Effort**: 40 hours  
**Priority**: P1

**Tasks**:
1. Write comprehensive API documentation (12h)
2. Create integration guide for external agents (8h)
3. Build example A2A client/server (8h)
4. Create troubleshooting guide (4h)
5. Record video tutorials (8h)

**Deliverables**:
- OpenAPI/Swagger specification
- Integration guide with code examples
- Sample A2A agent implementations (Python, Node.js)
- Troubleshooting playbook
- Video tutorials

---

## 5. API Specifications

### 5.1 Agent Card API

#### Get Agent Card
```http
GET /api/v1/agents/{agent_id}/a2a-card
Authorization: Bearer {token}
```

**Response 200 OK**:
```json
{
  "name": "Risk Analysis Agent",
  "description": "Analyzes financial transactions for risk indicators",
  "version": "1.0.0",
  "service_endpoint": "https://codevaldcortex.com/api/v1/a2a/agents/risk-001",
  "supported_modalities": ["text", "structured_data"],
  "capabilities": [
    "fraud-detection",
    "risk-scoring",
    "transaction-analysis"
  ],
  "authentication": {
    "type": "oauth2",
    "endpoint": "https://codevaldcortex.com/oauth/token",
    "required_scopes": ["agent.execute"]
  },
  "metadata": {
    "autonomy_level": "L3",
    "agency_id": "financial-services-agency-001",
    "compliance_tags": "SOC2,PCI-DSS"
  }
}
```

---

### 5.2 External Agent Management API

#### Register External Agent
```http
POST /api/v1/a2a/external-agents
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "service_endpoint": "https://vendor-x.com/agents/fraud-detector",
  "name": "FraudDetector Pro",
  "vendor": "VendorX",
  "metadata": {
    "cost_per_task": "0.05",
    "sla_guarantee": "99.9%"
  }
}
```

**Response 201 Created**:
```json
{
  "id": "ext_agent_001",
  "name": "FraudDetector Pro",
  "vendor": "VendorX",
  "service_endpoint": "https://vendor-x.com/agents/fraud-detector",
  "health_status": "healthy",
  "trust_score": 0.0,
  "card": {
    "name": "FraudDetector Pro",
    "version": "2.1.0",
    ...
  },
  "registered_at": "2025-11-11T10:00:00Z",
  "last_health_check": "2025-11-11T10:00:00Z"
}
```

---

#### List External Agents
```http
GET /api/v1/a2a/external-agents?status=healthy&page=1&limit=20
Authorization: Bearer {token}
```

**Response 200 OK**:
```json
{
  "agents": [
    {
      "id": "ext_agent_001",
      "name": "FraudDetector Pro",
      "vendor": "VendorX",
      "health_status": "healthy",
      "trust_score": 0.95,
      "last_health_check": "2025-11-11T10:30:00Z"
    },
    ...
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

---

### 5.3 Task Delegation API

#### Execute Task (Synchronous)
```http
POST /api/v1/a2a/agents/{agent_id}/tasks
Authorization: Bearer {token}
Content-Type: application/json

{
  "task_id": "task_12345",
  "input": {
    "transaction_id": "txn_98765",
    "amount": 50000,
    "merchant": "Suspicious Store"
  },
  "required_capabilities": ["fraud-detection"],
  "priority": "high",
  "timeout_seconds": 30
}
```

**Response 200 OK**:
```json
{
  "task_id": "task_12345",
  "status": "completed",
  "output": {
    "fraud_score": 0.87,
    "risk_level": "high",
    "reasons": ["unusual_amount", "new_merchant"],
    "recommended_action": "block_transaction"
  },
  "metadata": {
    "execution_time_ms": 450,
    "agent_version": "1.0.0"
  },
  "completed_at": "2025-11-11T10:35:00.450Z"
}
```

---

#### Execute Task (Asynchronous)
```http
POST /api/v1/a2a/agents/{agent_id}/tasks
Authorization: Bearer {token}
Content-Type: application/json
X-Execution-Mode: async

{
  "task_id": "task_12346",
  "input": { ... },
  "callback_url": "https://my-service.com/callbacks/task_12346"
}
```

**Response 202 Accepted**:
```json
{
  "task_id": "task_12346",
  "status": "running",
  "status_url": "/api/v1/a2a/tasks/task_12346/status",
  "created_at": "2025-11-11T10:35:00Z"
}
```

---

#### Check Task Status
```http
GET /api/v1/a2a/tasks/{task_id}/status
Authorization: Bearer {token}
```

**Response 200 OK**:
```json
{
  "task_id": "task_12346",
  "status": "running",
  "progress": 65,
  "estimated_completion": "2025-11-11T10:36:00Z",
  "created_at": "2025-11-11T10:35:00Z"
}
```

---

## 6. Security & Compliance

### 6.1 Authentication Methods

#### OAuth 2.0 Client Credentials Flow
```
1. External Agent → CodeValdCortex: POST /oauth/token
   {
     "grant_type": "client_credentials",
     "client_id": "ext_agent_001",
     "client_secret": "secret",
     "scope": "agent.execute"
   }

2. CodeValdCortex → External Agent: 200 OK
   {
     "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
     "token_type": "Bearer",
     "expires_in": 3600,
     "scope": "agent.execute"
   }

3. External Agent → CodeValdCortex: POST /api/v1/a2a/agents/{id}/tasks
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

### 6.2 Authorization & Access Control

**Role-Based Access Control (RBAC)**:

| Role | Permissions |
|------|-------------|
| `a2a_admin` | Register/deregister external agents, manage all tasks |
| `a2a_executor` | Execute tasks on agents, view task status |
| `a2a_reader` | Read agent cards, view task logs (read-only) |

**Token Scopes**:
- `agent.read`: Read agent cards and discovery
- `agent.execute`: Execute tasks on agents
- `agent.admin`: Manage agent configuration
- `task.read`: Read task status and logs
- `task.write`: Submit new tasks

---

### 6.3 Security Best Practices

1. **TLS 1.3 Enforcement**: All A2A communications use TLS 1.3
2. **Input Validation**: JSON schema validation on all incoming requests
3. **Rate Limiting**: 100 requests/minute per external agent
4. **Request Signing**: Optional request signature verification (HMAC-SHA256)
5. **IP Whitelisting**: Restrict external agent access by IP ranges
6. **Audit Logging**: Complete audit trail of all A2A operations

---

### 6.4 Compliance Features

**SOC 2 Type II**:
- Complete audit logs stored for 90 days
- Encryption at rest (AES-256) and in transit (TLS 1.3)
- Access control enforcement with RBAC

**GDPR**:
- Data minimization (only store necessary task data)
- Right to deletion (task log retention policy)
- Data portability (export API)

**HIPAA** (for healthcare use cases):
- PHI encryption and access controls
- BAA-compliant audit logging
- Secure agent-to-agent communication

---

## 7. Testing Strategy

### 7.1 Unit Testing

**Coverage Target**: 80% code coverage

**Key Test Cases**:
```go
// pkg/a2a/card_generator_test.go
func TestGenerateA2ACard(t *testing.T) {
    agent := &registry.Agent{
        ID:   "test-001",
        Name: "Test Agent",
        // ...
    }
    
    card, err := GenerateA2ACard(agent, testConfig)
    assert.NoError(t, err)
    assert.Equal(t, "Test Agent", card.Name)
    assert.NotEmpty(t, card.ServiceEndpoint)
}

// pkg/a2a/external_registry_test.go
func TestRegisterExternalAgent(t *testing.T) {
    req := &RegisterAgentRequest{
        ServiceEndpoint: "https://test.com/agent",
        Name: "External Test Agent",
    }
    
    agent, err := registry.RegisterExternal(req)
    assert.NoError(t, err)
    assert.Equal(t, "healthy", agent.HealthStatus)
}
```

---

### 7.2 Integration Testing

**Test Scenarios**:

1. **End-to-End Task Delegation**:
   - Register external agent
   - Submit task to external agent
   - Verify task execution and response
   - Check audit logs

2. **Health Check System**:
   - Register external agent
   - Simulate agent downtime
   - Verify automatic de-registration after 3 failures

3. **Authentication Flow**:
   - Request OAuth 2.0 token
   - Execute task with token
   - Verify token expiration handling

**Test Environment**:
```yaml
# docker-compose.test.yml
services:
  codevaldcortex:
    image: codevaldcortex:test
    environment:
      CVXC_A2A_ENABLED: true
      CVXC_A2A_BASE_URL: http://localhost:8082
  
  mock_external_agent:
    image: mock-a2a-agent:latest
    ports:
      - "9001:9001"
```

---

### 7.3 Performance Testing

**Load Testing with k6**:
```javascript
// tests/a2a/load_test.js
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '1m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '1m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
  },
};

export default function () {
  const payload = JSON.stringify({
    task_id: `task_${__VU}_${__ITER}`,
    input: { test_data: 'value' },
  });
  
  const res = http.post(
    'http://localhost:8082/api/v1/a2a/agents/test-001/tasks',
    payload,
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
}
```

**Performance Targets**:
- 1000 concurrent tasks
- P95 latency < 500ms
- 99.9% success rate
- < 100MB memory usage per 1000 agents

---

### 7.4 Security Testing

**Penetration Testing Checklist**:
- [ ] SQL injection attempts on API endpoints
- [ ] XSS attempts in task input fields
- [ ] Unauthorized access to admin endpoints
- [ ] Token manipulation and replay attacks
- [ ] Rate limiting bypass attempts
- [ ] TLS downgrade attacks

**Security Scanning Tools**:
- OWASP ZAP for API vulnerability scanning
- Trivy for container image scanning
- GoSec for Go code security analysis

---

## 8. Deployment Guide

### 8.1 Docker Deployment

**Dockerfile**:
```dockerfile
# Dockerfile.a2a-gateway
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o a2a-gateway ./cmd/a2a-gateway

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/a2a-gateway .
COPY --from=builder /app/config/a2a_selection.yaml ./config/

EXPOSE 8083
CMD ["./a2a-gateway"]
```

**Docker Compose**:
```yaml
# docker-compose.a2a.yml
version: '3.8'

services:
  a2a-gateway:
    build:
      context: .
      dockerfile: Dockerfile.a2a-gateway
    ports:
      - "8083:8083"
    environment:
      CVXC_A2A_ENABLED: "true"
      CVXC_A2A_BASE_URL: "http://localhost:8083"
      CVXC_DATABASE_HOST: "arangodb"
      CVXC_REDIS_HOST: "redis"
    depends_on:
      - arangodb
      - redis
    networks:
      - codevaldcortex

  codevaldcortex:
    image: codevaldcortex:latest
    environment:
      CVXC_A2A_GATEWAY_URL: "http://a2a-gateway:8083"
    networks:
      - codevaldcortex

networks:
  codevaldcortex:
    driver: bridge
```

---

### 8.2 Kubernetes Deployment

**Kubernetes Manifests**:
```yaml
# k8s/a2a-gateway-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: a2a-gateway
  namespace: codevaldcortex
spec:
  replicas: 3
  selector:
    matchLabels:
      app: a2a-gateway
  template:
    metadata:
      labels:
        app: a2a-gateway
    spec:
      containers:
      - name: a2a-gateway
        image: codevaldcortex/a2a-gateway:v1.0
        ports:
        - containerPort: 8083
        env:
        - name: CVXC_A2A_ENABLED
          value: "true"
        - name: CVXC_A2A_BASE_URL
          valueFrom:
            configMapKeyRef:
              name: a2a-config
              key: base_url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8083
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8083
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: a2a-gateway
  namespace: codevaldcortex
spec:
  selector:
    app: a2a-gateway
  ports:
  - protocol: TCP
    port: 8083
    targetPort: 8083
  type: LoadBalancer

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: a2a-config
  namespace: codevaldcortex
data:
  base_url: "https://a2a.codevaldcortex.com"
  selection_config: |
    selection:
      algorithm: weighted_score
      weights:
        capability_match: 0.4
        trust_score: 0.3
        cost_efficiency: 0.2
        latency: 0.1
```

---

### 8.3 Configuration Management

**Environment Variables**:
```bash
# A2A Gateway Configuration
export CVXC_A2A_ENABLED=true
export CVXC_A2A_BASE_URL=https://a2a.codevaldcortex.com
export CVXC_A2A_PORT=8083

# Security
export CVXC_A2A_TLS_CERT=/etc/certs/tls.crt
export CVXC_A2A_TLS_KEY=/etc/certs/tls.key
export CVXC_A2A_JWT_SECRET=your-secret-key

# Performance
export CVXC_A2A_MAX_CONNECTIONS=1000
export CVXC_A2A_CONNECTION_TIMEOUT=30s
export CVXC_A2A_TASK_TIMEOUT=30s

# Observability
export CVXC_A2A_METRICS_ENABLED=true
export CVXC_A2A_TRACING_ENABLED=true
export CVXC_A2A_LOG_LEVEL=info
```

---

### 8.4 Monitoring & Alerting

**Prometheus Alerts**:
```yaml
# alerts/a2a_alerts.yaml
groups:
- name: a2a_alerts
  interval: 30s
  rules:
  - alert: A2AHighErrorRate
    expr: |
      rate(a2a_tasks_total{status="failure"}[5m]) 
      / 
      rate(a2a_tasks_total[5m]) > 0.05
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "A2A task error rate above 5%"
      description: "{{ $value }}% of A2A tasks are failing"

  - alert: A2AHighLatency
    expr: |
      histogram_quantile(0.95, 
        rate(a2a_task_duration_seconds_bucket[5m])
      ) > 0.5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "A2A P95 latency above 500ms"
      description: "P95 latency is {{ $value }}s"

  - alert: ExternalAgentDown
    expr: |
      a2a_external_agents_active{status="down"} > 0
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "External A2A agent is down"
      description: "{{ $value }} external agents are unreachable"
```

---

## 9. Appendices

### Appendix A: Glossary

| Term | Definition |
|------|------------|
| **A2A Protocol** | Agent-to-Agent communication protocol for interoperability |
| **Agent Card** | Metadata document describing agent capabilities and endpoints |
| **A2A Gateway** | Service translating between internal and A2A protocol |
| **External Agent** | Agent hosted outside CodeValdCortex that uses A2A protocol |
| **Task Delegation** | Process of assigning work to an agent (internal or external) |
| **Trust Score** | Metric (0-1) representing reliability of external agent |

---

### Appendix B: Related Specifications

- [A2A Protocol Specification (Linux Foundation)](https://github.com/linuxfoundation/a2a-protocol)
- [CodeValdCortex Agent Registry](./backend-architecture.md#agent-registry)
- [CodeValdCortex Orchestration Engine](./backend-architecture.md#orchestration)
- [Security & Compliance Framework](../1-SoftwareRequirements/security-requirements.md)

---

### Appendix C: Migration Path

**For Existing CodeValdCortex Deployments**:

1. **Phase 1: Enable A2A Support** (Week 1-2)
   - Deploy A2A Gateway service
   - Generate Agent Cards for existing agents
   - No impact on existing functionality

2. **Phase 2: Register External Agents** (Week 3-4)
   - Register external agents via Admin UI
   - Run health checks and validate connectivity
   - External agents available but not selected by default

3. **Phase 3: Enable Intelligent Selection** (Week 5-6)
   - Configure agent selection algorithm weights
   - Enable A2A task delegation for specific agencies
   - Monitor performance and adjust configuration

4. **Phase 4: Production Rollout** (Week 7-8)
   - Enable A2A for all agencies
   - Decommission redundant internal agents
   - Optimize costs and performance

---

### Appendix D: FAQ

**Q: Will A2A integration affect existing agent performance?**  
A: No. Internal agents continue using Go channels for direct communication. A2A only applies to external agents.

**Q: How do I ensure security when using external agents?**  
A: All external agents are authenticated via OAuth 2.0, rate-limited, and audited. You can also configure IP whitelisting and minimum trust score requirements.

**Q: Can I use A2A with on-premise deployments?**  
A: Yes. A2A Gateway can be deployed on-premise and configured to communicate with both cloud and on-premise external agents.

**Q: What happens if an external agent fails?**  
A: The orchestration engine automatically falls back to internal agents or retries with alternative external agents based on your configuration.

**Q: How are costs tracked for external agent usage?**  
A: Task logs include cost metadata. You can generate cost reports grouped by agency, agent, or time period.

---

**Document History**:
- v1.0 (2025-11-11): Initial specification created
- Future versions will be tracked in Git commit history

**Authors**: CodeValdCortex Architecture Team  
**Reviewers**: TBD  
**Approvers**: TBD
