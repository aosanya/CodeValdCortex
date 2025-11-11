# a2a-go SDK Integration Strategy

**Document Version**: 1.0  
**Date**: November 11, 2025  
**Status**: Planning Phase

---

## Overview

This document outlines the integration strategy for using the official `a2a-go` SDK from the A2A Project (https://github.com/a2aproject/a2a-go) as the foundation for CodeValdCortex's A2A protocol implementation.

## Why Use a2a-go SDK?

### Decision Rationale

Instead of building a custom A2A protocol implementation from scratch, CodeValdCortex will leverage the official SDK for the following reasons:

| Aspect | Custom Implementation | Using a2a-go SDK | Decision |
|--------|----------------------|------------------|----------|
| **Development Time** | 12-16 weeks | 6-8 weeks | ✅ 40-50% faster |
| **Protocol Compliance** | Manual validation required | Guaranteed by SDK | ✅ Reduced risk |
| **Maintenance** | Must track spec changes | Upstream updates | ✅ Lower overhead |
| **Community Support** | Internal only | A2A ecosystem | ✅ Better support |
| **Security Patches** | Manual implementation | Upstream patches | ✅ Faster fixes |
| **Testing** | Custom test suite needed | SDK battle-tested | ✅ Higher quality |

**Conclusion**: Using the SDK reduces implementation time by 40-50% while ensuring protocol compliance and reducing long-term maintenance burden.

---

## Integration Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│              CodeValdCortex Platform                        │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          A2A Gateway (Our Implementation)            │  │
│  │                                                      │  │
│  │  ┌────────────────┐        ┌──────────────────┐    │  │
│  │  │  a2a-go Server │        │  a2a-go Client   │    │  │
│  │  │  (SDK)         │        │  (SDK)           │    │  │
│  │  └────────┬───────┘        └────────┬─────────┘    │  │
│  │           │                         │               │  │
│  │  ┌────────▼─────────────────────────▼─────────┐    │  │
│  │  │    Protocol Translator                     │    │  │
│  │  │    (Our Implementation)                    │    │  │
│  │  │    - Internal ↔ A2A format conversion      │    │  │
│  │  └────────┬───────────────────────┬───────────┘    │  │
│  └───────────┼───────────────────────┼────────────────┘  │
│              │                       │                    │
│     ┌────────▼────────┐     ┌────────▼────────┐          │
│     │   Orchestrator  │     │  Agent Registry │          │
│     │   (Internal)    │     │  (Internal)     │          │
│     └─────────────────┘     └─────────────────┘          │
└─────────────────────────────────────────────────────────────┘
         ▲                                 │
         │ A2A Protocol                    │ A2A Protocol
         │ (via SDK)                       ▼ (via SDK)
    ┌────┴─────┐                     ┌──────────┐
    │ External │                     │ External │
    │ Agent A  │                     │ Agent B  │
    └──────────┘                     └──────────┘
```

### Component Responsibilities

| Component | Responsibility | Implementation |
|-----------|---------------|----------------|
| **a2a-go Server** | Expose internal agents via A2A protocol | SDK (off-the-shelf) |
| **a2a-go Client** | Consume external A2A agents | SDK (off-the-shelf) |
| **Protocol Translator** | Convert between internal and A2A formats | Custom (CodeValdCortex) |
| **A2A Gateway** | Coordinate SDK components and internal systems | Custom (CodeValdCortex) |
| **Card Manager** | Generate and manage A2A Agent Cards | Custom using SDK types |
| **External Registry** | Catalog external A2A agents | Custom using SDK client |

---

## Implementation Plan

### Phase 1: SDK Integration Foundation (MVP-A2A-000)

**Duration**: 1-2 weeks  
**Effort**: 12 hours

**Tasks**:
1. Add `github.com/a2aproject/a2a-go` to `go.mod`
2. Create `internal/a2a/` package structure
3. Implement `A2AGateway` wrapper around SDK
4. Build protocol translator layer
5. Write integration tests with mock A2A agents

**Deliverables**:
- [ ] SDK dependency in `go.mod`
- [ ] `internal/a2a/gateway/gateway.go`
- [ ] `internal/a2a/translator/translator.go`
- [ ] Integration tests passing
- [ ] Configuration system (`config/a2a.yaml`)

**See Details**: [MVP-A2A-000 Task Specification](../3-SofwareDevelopment/mvp-details/MVP-A2A-000_a2a_go_sdk_integration.md)

### Phase 2: Agent Card Generation (MVP-A2A-001)

**Duration**: 1 week  
**Effort**: 24 hours

**Tasks**:
1. Implement card generator using `protocol.AgentCard` from SDK
2. Build HTTP endpoint to serve agent cards
3. Integrate with internal agent registry
4. Add card validation

**SDK Types Used**:
```go
import "github.com/a2aproject/a2a-go/protocol"

// Use SDK's AgentCard type
card := &protocol.AgentCard{
    Name:        agent.Name,
    Version:     agent.Version,
    Endpoint:    serviceURL,
    // ... SDK ensures compliance
}
```

### Phase 3: External Agent Integration (MVP-A2A-002, MVP-A2A-003)

**Duration**: 2 weeks  
**Effort**: 48 hours

**Tasks**:
1. Use `a2aclient.Client` to register external agents
2. Implement health check system
3. Build discovery UI
4. Create task delegation system

**SDK Usage**:
```go
import a2aclient "github.com/a2aproject/a2a-go/client"

client := a2aclient.NewClient(config)
response, err := client.ExecuteTask(ctx, agentURL, taskRequest)
```

---

## SDK Components Overview

### Available SDK Packages

| Package | Purpose | CodeValdCortex Usage |
|---------|---------|---------------------|
| `a2a-go/server` | Run A2A-compliant server | Expose internal agents |
| `a2a-go/client` | A2A protocol client | Call external agents |
| `a2a-go/protocol` | Protocol types & schemas | Type-safe message handling |
| `a2a-go/auth` | Authentication helpers | OAuth 2.0 / JWT integration |
| `a2a-go/discovery` | Agent discovery | External agent registration |

### Key SDK Interfaces

```go
// Server interface (from SDK)
type Server interface {
    Start(ctx context.Context) error
    RegisterAgent(handler *AgentHandler) error
    Stop() error
}

// Client interface (from SDK)
type Client interface {
    ExecuteTask(ctx context.Context, agentURL string, req *TaskRequest) (*TaskResponse, error)
    GetAgentCard(ctx context.Context, agentURL string) (*AgentCard, error)
}

// Our wrapper
type A2AGateway struct {
    server Server  // SDK server
    client Client  // SDK client
    // ... our custom components
}
```

---

## Configuration

### Environment Variables

```bash
# Enable A2A integration
CVXC_A2A_ENABLED=true

# SDK configuration
CVXC_A2A_SDK_LOG_LEVEL=info

# Server configuration
CVXC_A2A_SERVER_ADDRESS=:8083
CVXC_A2A_SERVER_TLS_ENABLED=true
CVXC_A2A_SERVER_TLS_CERT=/etc/certs/a2a.crt
CVXC_A2A_SERVER_TLS_KEY=/etc/certs/a2a.key

# Client configuration
CVXC_A2A_CLIENT_TIMEOUT=30s
CVXC_A2A_CLIENT_MAX_CONNECTIONS=100
CVXC_A2A_CLIENT_RETRY_ATTEMPTS=3
```

### Configuration File

```yaml
# config/a2a.yaml
a2a:
  enabled: true
  
  sdk:
    log_level: "info"
  
  server:
    address: ":8083"
    tls_enabled: true
    tls_cert_file: "/etc/certs/a2a.crt"
    tls_key_file: "/etc/certs/a2a.key"
  
  client:
    timeout: "30s"
    max_connections: 100
    retry_attempts: 3
    retry_backoff: "exponential"
  
  agent_selection:
    algorithm: "weighted_score"
    weights:
      capability_match: 0.4
      trust_score: 0.3
      cost_efficiency: 0.2
      latency: 0.1
```

---

## Testing Strategy

### Unit Tests

Test our wrapper code without SDK:

```go
func TestProtocolTranslator(t *testing.T) {
    translator := NewTranslator()
    
    // Test internal → A2A conversion
    task := &orchestration.Task{ID: "test-001"}
    a2aReq := translator.ToA2ARequest(task)
    assert.Equal(t, "test-001", a2aReq.TaskID)
}
```

### Integration Tests

Test with real SDK components:

```go
func TestGatewayWithSDK(t *testing.T) {
    // Use real a2a-go SDK for testing
    mockAgent := startMockA2AAgent(t) // Using SDK server
    defer mockAgent.Stop()
    
    gateway := setupGateway(t)
    result, err := gateway.ExecuteOnExternalAgent(ctx, mockAgent.URL, task)
    
    assert.NoError(t, err)
    assert.Equal(t, "completed", result.Status)
}
```

### SDK Compatibility Tests

Ensure compatibility with SDK updates:

```go
func TestSDKVersionCompatibility(t *testing.T) {
    // Verify we're using compatible SDK version
    version := a2ago.Version()
    assert.True(t, semver.Compare(version, "v1.0.0") >= 0)
}
```

---

## Migration Timeline

### Immediate (Week 1-2)
- ✅ Decision to use a2a-go SDK documented
- 🔄 Add SDK to `go.mod` (MVP-A2A-000)
- 🔄 Create wrapper interfaces
- 🔄 Implement protocol translator

### Short-term (Week 3-6)
- 🔲 Agent Card generation using SDK types (MVP-A2A-001)
- 🔲 External agent registry using SDK client (MVP-A2A-002)
- 🔲 Gateway service with SDK integration (MVP-A2A-003)

### Medium-term (Week 7-10)
- 🔲 Task delegation system (MVP-A2A-004)
- 🔲 Enhanced orchestration (MVP-A2A-005)
- 🔲 Security & compliance (MVP-A2A-006)

### Long-term (Week 11-12)
- 🔲 Monitoring & observability (MVP-A2A-007)
- 🔲 Performance optimization (MVP-A2A-008)
- 🔲 Documentation & developer experience (MVP-A2A-009)

---

## Risk Mitigation

### SDK Dependency Risks

| Risk | Mitigation |
|------|-----------|
| SDK breaking changes | Pin to stable version, monitor releases, maintain compatibility tests |
| SDK bugs/issues | Contribute fixes upstream, maintain fork if necessary |
| SDK abandonment | Low risk (Linux Foundation project), but could fork if needed |
| API incompatibility | Protocol translator provides abstraction layer |

### Technical Risks

| Risk | Mitigation |
|------|-----------|
| Performance overhead | Benchmark SDK vs. direct implementation, optimize translator |
| Integration complexity | Thorough testing, gradual rollout, fallback to internal agents |
| Learning curve | Study SDK documentation, create internal guides, code examples |

---

## Success Metrics

### Implementation Metrics
- [ ] SDK integrated and building successfully
- [ ] All unit tests passing (>80% coverage)
- [ ] Integration tests with mock A2A agents passing
- [ ] Documentation complete and reviewed

### Performance Metrics
- [ ] P95 latency < 500ms for A2A task delegation
- [ ] Support 1000 concurrent A2A tasks
- [ ] < 10MB memory per external agent connection

### Business Metrics
- [ ] 40-50% reduction in A2A implementation time
- [ ] Zero protocol compliance issues
- [ ] Successful integration with at least 1 external A2A agent

---

## References

### Official Resources
- **SDK Repository**: https://github.com/a2aproject/a2a-go
- **A2A Protocol Spec**: https://github.com/linuxfoundation/a2a-protocol
- **SDK Documentation**: https://pkg.go.dev/github.com/a2aproject/a2a-go
- **A2A Project**: https://www.linuxfoundation.org/projects/a2a

### Internal Documentation
- [A2A Protocol Integration Spec](./a2a-protocol-integration.md)
- [MVP Task List](../3-SofwareDevelopment/mvp.md#a2a-protocol-integration-p1---strategic)
- [MVP-A2A-000 Task Details](../3-SofwareDevelopment/mvp-details/MVP-A2A-000_a2a_go_sdk_integration.md)
- [Backend Architecture](./backend-architecture.md)

---

## Appendix: Example Code

### Basic SDK Usage

```go
package main

import (
    "context"
    "log"
    
    a2aserver "github.com/a2aproject/a2a-go/server"
    "github.com/a2aproject/a2a-go/protocol"
)

func main() {
    // Create A2A server
    server, err := a2aserver.NewServer(a2aserver.Config{
        Address: ":8083",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Register an agent
    server.RegisterAgent(&protocol.AgentHandler{
        AgentID: "example-agent",
        Card: &protocol.AgentCard{
            Name:         "Example Agent",
            Version:      "1.0.0",
            Capabilities: []string{"example"},
        },
        ExecuteFunc: func(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error) {
            return &protocol.TaskResponse{
                TaskID: req.TaskID,
                Status: "completed",
                Output: map[string]interface{}{"result": "success"},
            }, nil
        },
    })
    
    // Start server
    server.Start(context.Background())
}
```

### CodeValdCortex Integration

```go
// internal/a2a/gateway/gateway.go
type A2AGateway struct {
    server *a2aserver.Server
    client *a2aclient.Client
}

func (g *A2AGateway) RegisterInternalAgent(agent *registry.Agent) error {
    card := generateCard(agent)
    
    handler := &protocol.AgentHandler{
        AgentID: agent.ID,
        Card:    card,
        ExecuteFunc: func(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error) {
            // Bridge to internal orchestration
            task := g.translator.ToInternalTask(req)
            result := g.orchestrator.ExecuteTask(ctx, agent.ID, task)
            return g.translator.ToA2AResponse(result), nil
        },
    }
    
    return g.server.RegisterAgent(handler)
}
```

---

**Document Status**: ✅ Approved for Implementation  
**Next Action**: Begin MVP-A2A-000 (SDK Integration)
