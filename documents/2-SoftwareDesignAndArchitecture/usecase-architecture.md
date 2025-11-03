# Use Case Architecture - Configuration-Only Approach

## Overview

Use cases in CodeValdCortex are **extremely thin** - they should contain **only configuration files** with minimal to no custom code. The framework handles all the heavy lifting including agent lifecycle, data coordination, messaging, monitoring, and orchestration.

## Core Principle: Configuration Over Code

### Philosophy

**Use cases should be declarative, not imperative.**

- ✅ **DO**: Define roles, schemas, and behaviors through JSON/YAML configuration
- ✅ **DO**: Use framework-provided capabilities and built-in roles
- ✅ **DO**: Leverage the Role Registry for dynamic type registration
- ❌ **DON'T**: Write custom agent implementation code in use cases
- ❌ **DON'T**: Create use case-specific lifecycle management
- ❌ **DON'T**: Implement custom messaging or coordination logic

### What the Framework Provides

The framework (`/workspaces/CodeValdCortex/internal/`) provides all the infrastructure:

1. **Agent Lifecycle Management** (`internal/agent/`, `internal/runtime/`)
   - Agent creation, starting, stopping, monitoring
   - Health checks and heartbeats
   - Task execution and queuing
   - State management

2. **Data Coordination** (`internal/database/`, `internal/memory/`)
   - ArangoDB integration
   - Document storage and retrieval
   - Change stream processing
   - Memory services

3. **Communication** (`internal/communication/`)
   - Message routing and delivery
   - Pub/Sub patterns
   - Request/Response patterns
   - Message matching and filtering

4. **Role Registry** (`internal/registry/`)
   - Dynamic role registration
   - JSON Schema validation
   - Type-based agent creation
   - Capability management

5. **Configuration Management** (`internal/configuration/`)
   - Configuration loading and validation
   - Template processing
   - Environment variable handling

6. **Orchestration** (`internal/orchestration/`)
   - Multi-agent coordination
   - Workflow execution
   - Pool management

## Use Case Structure

A properly structured use case should look like this:

```
Usecases/UC-XXX-use-case-name/
├── .env                           # Environment configuration
├── README.md                      # Use case documentation
└── config/
    ├── agents/                    # Agent type definitions (JSON)
    │   ├── agent-type-1.json     # Agent type schema, capabilities, validation
    │   ├── agent-type-2.json
    │   └── agent-type-3.json
    ├── instances/                 # Agent instance configurations (optional)
    │   ├── agent-instances.json  # Specific agent instances to create
    │   └── templates.json        # Instance templates with variables
    └── workflows/                 # Workflow definitions (optional)
        └── workflows.json         # Multi-agent workflows
```

### No Code Required

Notice what's **NOT** in the use case directory:
- ❌ No `cmd/main.go` with custom agent factories
- ❌ No `internal/agents/` with custom agent implementations
- ❌ No custom Go code at all
- ❌ No package dependencies

## Example: UC-INFRA-001 Water Distribution Network

### Current Structure (Correct)

```
UC-INFRA-001-water-distribution-network/
├── .env                          # Environment with USECASE_CONFIG_DIR
├── README.md                     # Documentation
└── config/
    └── agents/
        └── pipe.json             # Pipe role definition
```

### What Gets Loaded

The framework automatically:

1. **Reads** `.env` to get `USECASE_CONFIG_DIR`
2. **Scans** `config/agents/*.json` for role definitions
3. **Validates** each JSON file against the AgentType schema
4. **Registers** roles in the Role Registry
5. **Makes available** for runtime agent creation

### Role Definition (pipe.json)

```json
{
    "id": "pipe",
    "name": "Pipe Agent",
    "description": "Water distribution pipe infrastructure agent",
    "category": "infrastructure",
    "version": "1.0.0",
    "schema": {
        "$schema": "http://json-schema.org/draft-07/schema#",
        "type": "object",
        "required": ["pipe_id", "material", "diameter", "length"],
        "properties": {
            "pipe_id": {
                "type": "string",
                "pattern": "^PIPE-[0-9]+$"
            },
            "material": {
                "type": "string",
                "enum": ["PVC", "steel", "copper", "cast_iron"]
            },
            "diameter": {
                "type": "integer",
                "minimum": 50,
                "maximum": 2000
            }
            // ... more properties
        }
    },
    "capabilities": [
        "monitor_pressure",
        "monitor_flow",
        "detect_leaks",
        "report_status"
    ],
    "default_config": {
        "monitoring_interval": 60,
        "alert_threshold": 5.0
    },
    "validation_rules": {
        "pressure_range": {"min": 0, "max": 200},
        "flow_rate_max": 2000
    }
}
```

This single JSON file defines **everything** about the pipe role:
- What data it requires (schema)
- What it can do (capabilities)
- How to validate instances (validation_rules)
- Default behavior (default_config)

## How Agents Are Created at Runtime

### From Database Records

When the framework queries ArangoDB and finds pipe records:

```go
// Framework code (internal/app/app.go or similar)
// User does NOT write this - it's in the framework

func loadAgentsFromDatabase(ctx context.Context) error {
    // 1. Query ArangoDB for pipe infrastructure
    query := "FOR pipe IN water_infrastructure_pipes RETURN pipe"
    cursor, _ := db.Query(ctx, query, nil)
    
    // 2. For each record, create agent using registry
    for cursor.HasMore() {
        var pipeData map[string]interface{}
        cursor.ReadDocument(ctx, &pipeData)
        
        // 3. Framework validates against registered schema
        agentType, _ := agentTypeService.GetType(ctx, "pipe")
        if err := agentType.ValidateConfig(pipeData); err != nil {
            continue // Invalid data
        }
        
        // 4. Framework creates agent instance
        agent := agent.New(
            pipeData["pipe_id"].(string),
            "pipe",
            agent.Config{/* from default_config */}
        )
        
        // 5. Framework starts and manages agent
        runtimeManager.RegisterAgent(agent)
        agent.Start()
    }
    
    return nil
}
```

### From API Requests

```bash
# Create a pipe agent instance via API
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "type": "pipe",
    "config": {
      "pipe_id": "PIPE-001",
      "material": "PVC",
      "diameter": 200,
      "length": 100.0
    }
  }'
```

The framework:
1. Looks up the "pipe" role from the registry
2. Validates the config against the JSON schema
3. Creates and starts the agent
4. Returns the agent ID

## Environment Configuration

The `.env` file in each use case configures the framework:

```bash
# Use case identification
USECASE_ID=UC-INFRA-001
USECASE_NAME=water-distribution-network
USECASE_CONFIG_DIR=/workspaces/CodeValdCortex/usecases/UC-INFRA-001-water-distribution-network

# Database
DB_NAME=water_distribution_network

# Framework behavior
LOAD_DEFAULT_AGENT_TYPES=true
MONITORING_INTERVAL=60
HEARTBEAT_INTERVAL=30

# Domain-specific thresholds
PRESSURE_DEVIATION_ALERT=5.0
FLOW_EFFICIENCY_CRITICAL=0.5
```

The framework reads these values and configures itself accordingly.

## Loading Roles at Startup

The framework automatically loads use case roles:

```go
// In internal/app/app.go - New() function

// Register default roles
ctx := context.Background()
if err := registry.InitializeDefaultAgentTypes(ctx, agentTypeService, logger); err != nil {
    logger.Warn("Failed to initialize default roles")
}

// Load use case-specific roles from config directory
useCaseConfigDir := os.Getenv("USECASE_CONFIG_DIR")
if useCaseConfigDir != "" {
    agentTypesDir := filepath.Join(useCaseConfigDir, "config", "agents")
    if err := loadAgentTypesFromDirectory(ctx, agentTypesDir, agentTypeService, logger); err != nil {
        logger.Warn("Failed to load use case roles")
    }
}
```

**Result**: All roles (12 defaults + use case-specific) are available in the registry and ready to use.

## Benefits of Configuration-Only Use Cases

### 1. Simplicity
- No Go code to maintain
- No compilation or build process
- No dependencies to manage
- Easy to understand and modify

### 2. Portability
- Use cases can be shared as simple JSON files
- No language-specific knowledge required
- Can be edited with any text editor
- Version controlled easily

### 3. Validation
- JSON Schema ensures correctness
- Framework validates before registration
- Type-safe at runtime
- Early error detection

### 4. Flexibility
- Add new roles without code changes
- Modify schemas dynamically
- Enable/disable types via API
- Hot-reload configurations

### 5. Consistency
- All use cases follow same pattern
- Framework handles all complexity
- Uniform error handling
- Standard monitoring and logging

## Anti-Patterns to Avoid

### ❌ Creating Custom Agent Code in Use Cases

**Wrong**:
```
Usecases/UC-INFRA-001/
├── cmd/main.go                    # ❌ Don't create this
├── internal/
│   └── agents/
│       └── pipe/
│           ├── pipe_agent.go      # ❌ Don't implement custom agents
│           └── monitoring.go      # ❌ Don't add custom logic
```

**Right**:
```
Usecases/UC-INFRA-001/
├── .env
├── README.md
└── config/
    └── agents/
        └── pipe.json              # ✅ Just configuration
```

### ❌ Implementing Custom Lifecycle Management

**Wrong**:
```go
// DON'T write this in use cases
func (f *AgentFactory) CreatePipeAgentFromConfig(ctx context.Context, pipeData PipeData) (*pipe.PipeAgent, error) {
    baseAgent := agent.New(...)
    pipeConfig := pipe.Config{...}
    pipeAgent, err := pipe.New(pipeConfig, baseAgent, ...)
    // ... custom initialization
}
```

**Right**: Let the framework handle it through the registry and runtime manager.

### ❌ Custom Database Queries in Use Cases

**Wrong**:
```go
// DON'T write this in use cases
func loadPipesFromDatabase(ctx context.Context, db *database.ArangoClient) error {
    query := "FOR pipe IN water_infrastructure_pipes RETURN pipe"
    cursor, _ := db.Query(ctx, query, nil)
    // ... custom processing
}
```

**Right**: Framework provides query services and automatic agent loading from collections.

## When Custom Code IS Needed

In rare cases where the framework doesn't provide required functionality:

1. **Extend the Framework** - Add the capability to `internal/` so all use cases benefit
2. **Create a Plugin** - Framework should support plugins for domain-specific logic
3. **Contribute Back** - Submit enhancements to the core framework

**Never** add custom code to individual use cases. If you need it, others probably will too.

## Migration Guide

If you have existing use cases with custom code:

### Step 1: Extract Role Definition
Convert Go structs to JSON Schema:

```go
// Before: Go code
type PipeConfig struct {
    PipeID   string  `json:"pipe_id"`
    Material string  `json:"material"`
    Diameter int     `json:"diameter"`
}
```

```json
// After: JSON Schema in config/agents/pipe.json
{
    "schema": {
        "properties": {
            "pipe_id": {"type": "string"},
            "material": {"type": "string"},
            "diameter": {"type": "integer"}
        }
    }
}
```

### Step 2: Move Capabilities to Configuration
```go
// Before: Go methods
func (p *PipeAgent) MonitorPressure() { ... }
func (p *PipeAgent) DetectLeaks() { ... }
```

```json
// After: Capabilities in JSON
{
    "capabilities": [
        "monitor_pressure",
        "detect_leaks"
    ]
}
```

### Step 3: Delete Custom Code
Remove all use case-specific Go code - the framework handles it all.

### Step 4: Test
```bash
# Set environment
export USECASE_CONFIG_DIR=/path/to/usecase

# Start framework
./bin/codevaldcortex

# Verify types loaded
curl http://localhost:8080/api/v1/roles
```

## Conclusion

**Use cases are configuration, not code.**

By keeping use cases as pure configuration, we achieve:
- Maximum flexibility and portability
- Minimal maintenance burden  
- Consistent behavior across all use cases
- Easy sharing and collaboration
- Framework-driven best practices

The framework is powerful and complete - use it, don't reinvent it.

## Integration with Agency Operations Framework

Use cases within CodeValdCortex agencies should be designed with the Agency Operations Framework in mind. Each use case should consider:

### Goals Integration
- **Goal Mapping**: Each use case should map to one or more defined goals in the agency's Goals Module
- **Solution Scope**: Use case configuration should align with the problem scope and success metrics
- **Traceability**: Clear links between use case agents and the problems they help solve

### Work Items (WI) Implementation
- **Work Package Alignment**: Use case roles should support the defined Work Items
- **Agent Capabilities**: Agent capabilities in JSON configuration should match WI requirements
- **Deliverable Support**: Agent schemas should capture data needed for WI deliverables

### RACI Matrix Considerations
When configuring use case agents, consider the RACI matrix requirements:
- **Responsible Agents**: Agent types that execute work (R roles)
- **Accountable Agents**: Agent types that verify completion and approve results (A roles)  
- **Consulted Agents**: Agent types that provide input and expertise (C roles)
- **Informed Agents**: Agent types that receive status updates and notifications (I roles)

### Example: Financial Risk Analysis Use Case with RACI

```json
{
    "id": "risk_analyzer",
    "name": "Financial Risk Analyzer",
    "description": "Responsible for executing risk analysis calculations",
    "raci_roles": ["R"],
    "capabilities": [
        "calculate_ratios",
        "analyze_trends",
        "detect_anomalies"
    ],
    "reports_to": ["risk_supervisor"],
    "consults_with": ["market_data_agent", "regulatory_agent"],
    "informs": ["stakeholder_notification_agent"]
}
```

This integration ensures that use case configurations support the broader agency operational framework and RACI-based responsibility assignments.

For detailed information about the Agency Operations Framework, see [agency-operations-framework.md](agency-operations-framework.md).
