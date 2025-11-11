# MVP-A2A-000: a2a-go SDK Integration

**Task ID**: MVP-A2A-000  
**Title**: Integrate a2a-go SDK for A2A Protocol Implementation  
**Status**: Not Started  
**Priority**: P0  
**Effort**: Low (8-12 hours)  
**Dependencies**: None

---

## Overview

Integrate the official `a2a-go` SDK from the A2A Project to accelerate A2A protocol implementation and ensure compliance with the A2A specification.

**SDK Repository**: https://github.com/a2aproject/a2a-go

## Objectives

1. Add `a2a-go` dependency to `go.mod`
2. Create wrapper interfaces for CodeValdCortex-specific integration
3. Implement protocol translation layer between internal and A2A formats
4. Establish foundation for subsequent A2A tasks (MVP-A2A-001 through MVP-A2A-009)

## Benefits

- ✅ **40-60% faster implementation** vs. building from scratch
- ✅ **Guaranteed protocol compliance** with A2A specification
- ✅ **Upstream updates** and security patches
- ✅ **Battle-tested code** used by A2A ecosystem participants
- ✅ **Reduced maintenance burden** - no need to track spec changes

## Implementation Steps

### 1. Add SDK Dependency

```bash
# Add a2a-go SDK to project
go get github.com/a2aproject/a2a-go@latest

# Verify installation
go list -m github.com/a2aproject/a2a-go
```

**Expected `go.mod` change**:
```go
module github.com/aosanya/CodeValdCortex

go 1.23.0

require (
    // ... existing dependencies ...
    github.com/a2aproject/a2a-go v1.0.0  // Add this
)
```

### 2. Create Package Structure

```
internal/a2a/
├── gateway/
│   ├── gateway.go           # Main A2A Gateway implementation
│   ├── gateway_test.go      # Integration tests
│   └── config.go            # Gateway configuration
├── translator/
│   ├── translator.go        # Protocol translation layer
│   ├── to_a2a.go           # Internal → A2A format conversion
│   ├── from_a2a.go         # A2A → Internal format conversion
│   └── translator_test.go  # Translation tests
├── card/
│   ├── manager.go          # Agent Card management
│   ├── generator.go        # Card generation logic
│   └── validator.go        # Card validation
└── registry/
    ├── external.go         # External agent registry
    └── health.go           # Health check system
```

### 3. Implement Gateway Wrapper

**File**: `internal/a2a/gateway/gateway.go`

```go
package gateway

import (
    "context"
    "fmt"
    
    a2aclient "github.com/a2aproject/a2a-go/client"
    a2aserver "github.com/a2aproject/a2a-go/server"
    "github.com/a2aproject/a2a-go/protocol"
    
    "github.com/aosanya/CodeValdCortex/internal/orchestration"
    "github.com/aosanya/CodeValdCortex/internal/registry"
)

// A2AGateway bridges CodeValdCortex internal systems with A2A protocol
type A2AGateway struct {
    // a2a-go SDK components
    server *a2aserver.Server  // Expose internal agents
    client *a2aclient.Client  // Consume external agents
    
    // CodeValdCortex components
    translator   *Translator
    orchestrator *orchestration.Engine
    cardManager  *card.Manager
}

// NewGateway creates a new A2A Gateway
func NewGateway(config *Config) (*A2AGateway, error) {
    // Initialize a2a-go server
    server, err := a2aserver.NewServer(a2aserver.Config{
        Address:   config.ServerAddress,
        TLSConfig: config.TLSConfig,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create a2a server: %w", err)
    }
    
    // Initialize a2a-go client
    client := a2aclient.NewClient(a2aclient.Config{
        Timeout:     config.ClientTimeout,
        RetryPolicy: config.RetryPolicy,
    })
    
    return &A2AGateway{
        server:       server,
        client:       client,
        translator:   NewTranslator(),
        orchestrator: config.Orchestrator,
        cardManager:  config.CardManager,
    }, nil
}

// Start starts the A2A Gateway
func (g *A2AGateway) Start(ctx context.Context) error {
    // Register internal agents with A2A server
    if err := g.registerInternalAgents(); err != nil {
        return fmt.Errorf("failed to register internal agents: %w", err)
    }
    
    // Start A2A server
    return g.server.Start(ctx)
}

// RegisterInternalAgent exposes an internal agent via A2A protocol
func (g *A2AGateway) RegisterInternalAgent(agent *registry.Agent) error {
    // Generate A2A agent card
    card, err := g.cardManager.GenerateCard(agent)
    if err != nil {
        return fmt.Errorf("failed to generate card: %w", err)
    }
    
    // Create protocol handler
    handler := &protocol.AgentHandler{
        AgentID: agent.ID,
        Card:    card,
        ExecuteFunc: func(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error) {
            // Translate A2A → Internal
            task := g.translator.ToInternalTask(req)
            
            // Execute via orchestrator
            result := g.orchestrator.ExecuteTask(ctx, agent.ID, task)
            
            // Translate Internal → A2A
            return g.translator.ToA2AResponse(result), nil
        },
    }
    
    return g.server.RegisterAgent(handler)
}

// ExecuteOnExternalAgent delegates a task to an external A2A agent
func (g *A2AGateway) ExecuteOnExternalAgent(
    ctx context.Context,
    agentURL string,
    task *orchestration.Task,
) (*orchestration.TaskResult, error) {
    // Translate Internal → A2A
    a2aReq := g.translator.ToA2ARequest(task)
    
    // Execute using a2a-go client
    a2aResp, err := g.client.ExecuteTask(ctx, agentURL, a2aReq)
    if err != nil {
        return nil, fmt.Errorf("external execution failed: %w", err)
    }
    
    // Translate A2A → Internal
    return g.translator.ToInternalResult(a2aResp), nil
}
```

### 4. Implement Protocol Translator

**File**: `internal/a2a/translator/translator.go`

```go
package translator

import (
    "github.com/a2aproject/a2a-go/protocol"
    "github.com/aosanya/CodeValdCortex/internal/orchestration"
)

// Translator converts between CodeValdCortex internal formats and A2A protocol
type Translator struct{}

// NewTranslator creates a new protocol translator
func NewTranslator() *Translator {
    return &Translator{}
}

// ToA2ARequest converts internal task to A2A protocol request
func (t *Translator) ToA2ARequest(task *orchestration.Task) *protocol.TaskRequest {
    return &protocol.TaskRequest{
        TaskID:   task.ID,
        AgentID:  task.TargetAgentID,
        Input:    task.Input,
        Priority: t.convertPriority(task.Priority),
        Timeout:  int(task.Timeout.Seconds()),
    }
}

// ToInternalTask converts A2A protocol request to internal task
func (t *Translator) ToInternalTask(req *protocol.TaskRequest) *orchestration.Task {
    return &orchestration.Task{
        ID:            req.TaskID,
        TargetAgentID: req.AgentID,
        Input:         req.Input,
        Priority:      t.convertPriorityFromA2A(req.Priority),
        // ... map other fields
    }
}

// ToA2AResponse converts internal result to A2A protocol response
func (t *Translator) ToA2AResponse(result *orchestration.TaskResult) *protocol.TaskResponse {
    resp := &protocol.TaskResponse{
        TaskID: result.TaskID,
        Status: result.Status,
        Output: result.Output,
    }
    
    if result.Error != nil {
        resp.Error = &protocol.Error{
            Code:    result.Error.Code,
            Message: result.Error.Message,
        }
    }
    
    return resp
}

// ToInternalResult converts A2A protocol response to internal result
func (t *Translator) ToInternalResult(resp *protocol.TaskResponse) *orchestration.TaskResult {
    result := &orchestration.TaskResult{
        TaskID: resp.TaskID,
        Status: resp.Status,
        Output: resp.Output,
    }
    
    if resp.Error != nil {
        result.Error = &orchestration.TaskError{
            Code:    resp.Error.Code,
            Message: resp.Error.Message,
        }
    }
    
    return result
}

// Helper functions
func (t *Translator) convertPriority(p orchestration.Priority) string {
    // Map internal priority to A2A protocol priority
    switch p {
    case orchestration.PriorityCritical:
        return "critical"
    case orchestration.PriorityHigh:
        return "high"
    case orchestration.PriorityMedium:
        return "medium"
    default:
        return "low"
    }
}

func (t *Translator) convertPriorityFromA2A(p string) orchestration.Priority {
    // Map A2A protocol priority to internal priority
    switch p {
    case "critical":
        return orchestration.PriorityCritical
    case "high":
        return orchestration.PriorityHigh
    case "medium":
        return orchestration.PriorityMedium
    default:
        return orchestration.PriorityLow
    }
}
```

### 5. Configuration

**File**: `config/a2a.yaml`

```yaml
a2a:
  enabled: true
  
  # SDK configuration
  sdk:
    log_level: "info"
  
  # Server (expose internal agents)
  server:
    address: ":8083"
    tls_enabled: true
    tls_cert_file: "/etc/certs/a2a.crt"
    tls_key_file: "/etc/certs/a2a.key"
  
  # Client (consume external agents)
  client:
    timeout: "30s"
    max_connections: 100
    retry_attempts: 3
    retry_backoff: "exponential"
```

### 6. Testing

**File**: `internal/a2a/gateway/gateway_test.go`

```go
package gateway_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    a2aserver "github.com/a2aproject/a2a-go/server"
    "github.com/a2aproject/a2a-go/protocol"
    
    "github.com/aosanya/CodeValdCortex/internal/a2a/gateway"
)

func TestGatewayIntegration(t *testing.T) {
    // Setup mock external A2A agent using SDK
    mockAgent := startMockA2AAgent(t)
    defer mockAgent.Stop()
    
    // Initialize gateway
    gw, err := gateway.NewGateway(testConfig())
    require.NoError(t, err)
    
    // Test executing task on external agent
    task := &orchestration.Task{
        ID:    "test-001",
        Input: map[string]interface{}{"query": "test"},
    }
    
    result, err := gw.ExecuteOnExternalAgent(context.Background(), mockAgent.URL, task)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "completed", result.Status)
}

// Mock A2A agent using a2a-go SDK
func startMockA2AAgent(t *testing.T) *MockAgent {
    server, err := a2aserver.NewServer(a2aserver.Config{
        Address: ":0", // Random available port
    })
    require.NoError(t, err)
    
    // Register mock handler
    server.RegisterAgent(&protocol.AgentHandler{
        AgentID: "mock-agent",
        Card: &protocol.AgentCard{
            Name:         "Mock Agent",
            Version:      "1.0.0",
            Capabilities: []string{"test"},
        },
        ExecuteFunc: func(ctx context.Context, req *protocol.TaskRequest) (*protocol.TaskResponse, error) {
            return &protocol.TaskResponse{
                TaskID: req.TaskID,
                Status: "completed",
                Output: map[string]interface{}{"result": "success"},
            }, nil
        },
    })
    
    go server.Start(context.Background())
    
    return &MockAgent{Server: server}
}
```

## Acceptance Criteria

- [ ] `a2a-go` SDK added to `go.mod`
- [ ] `internal/a2a/` package structure created
- [ ] `A2AGateway` wrapper implemented
- [ ] Protocol translator implemented and tested
- [ ] Configuration system in place
- [ ] Integration tests passing with mock A2A agent
- [ ] Documentation updated with SDK usage patterns
- [ ] All tests passing (`go test ./internal/a2a/...`)

## Validation Steps

```bash
# 1. Verify SDK installation
go list -m github.com/a2aproject/a2a-go

# 2. Run unit tests
go test ./internal/a2a/translator/... -v

# 3. Run integration tests
go test ./internal/a2a/gateway/... -v

# 4. Build project to verify no errors
go build ./cmd/...

# 5. Verify gateway can start
go run ./cmd/a2a-gateway
```

## Next Steps

After completion, this task unblocks:
- **MVP-A2A-001**: A2A Agent Card Generator
- **MVP-A2A-002**: External Agent Registry
- **MVP-A2A-003**: A2A Gateway Service

## Resources

- **SDK Repository**: https://github.com/a2aproject/a2a-go
- **A2A Protocol Spec**: https://github.com/linuxfoundation/a2a-protocol
- **Go Documentation**: https://pkg.go.dev/github.com/a2aproject/a2a-go
- **Integration Spec**: `/documents/2-SoftwareDesignAndArchitecture/a2a-protocol-integration.md`

## Estimated Timeline

- **Setup & Research**: 2 hours
- **Package Structure**: 2 hours
- **Gateway Implementation**: 3 hours
- **Translator Implementation**: 2 hours
- **Testing**: 2 hours
- **Documentation**: 1 hour

**Total**: 12 hours
