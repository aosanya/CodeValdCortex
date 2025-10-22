# Agent Type Registry System

**Version**: 1.0.0  
**Status**: Implemented  
**Date**: October 22, 2025

## Overview

The Agent Type Registry is a flexible system that allows defining and managing agent types dynamically, rather than using hardcoded lists. This enables the framework to support domain-specific agent types for different use cases (e.g., water distribution infrastructure, logistics, community management) without modifying core framework code.

## Key Features

- **Dynamic Agent Type Registration**: Define new agent types at runtime
- **Schema-Based Validation**: JSON Schema validation for agent configurations
- **Category Organization**: Group related agent types by category
- **Capability Management**: Define required and optional capabilities per type
- **Enable/Disable Control**: Toggle agent type availability without deletion
- **System vs Custom Types**: Protect core system types from deletion
- **REST API**: Full CRUD operations via HTTP endpoints

## Architecture

### Components

1. **AgentType** - Core type definition with schema and validation rules
2. **AgentTypeRepository** - Storage interface with in-memory implementation
3. **AgentTypeService** - Business logic layer for type management
4. **Agent Type Handlers** - REST API endpoints
5. **Validator Integration** - Automatic validation in configuration and template validators

### Data Model

```go
type AgentType struct {
    ID                    string                 // Unique identifier (e.g., "pipe", "sensor")
    Name                  string                 // Human-readable name
    Description           string                 // Purpose description
    Category              string                 // Grouping (e.g., "core", "infrastructure")
    Version               string                 // Schema version
    Schema                json.RawMessage        // JSON Schema for validation
    RequiredCapabilities  []string               // Must-have capabilities
    OptionalCapabilities  []string               // Nice-to-have capabilities
    DefaultConfig         map[string]interface{} // Default configuration values
    ValidationRules       []ValidationRule       // Custom validation rules
    Metadata              map[string]string      // Additional information
    IsSystemType          bool                   // Protected from deletion
    IsEnabled             bool                   // Available for agent creation
    CreatedAt             time.Time
    UpdatedAt             time.Time
    CreatedBy             string
}
```

## Pre-Registered Agent Types

### Core System Types (Category: "core")

1. **worker** - General-purpose task execution agent
2. **coordinator** - Agent orchestration and management
3. **monitor** - Monitoring and observability agent
4. **proxy** - External system integration agent
5. **gateway** - API gateway agent

### Infrastructure Types (Category: "infrastructure")

For Water Distribution Network use case:

1. **pipe** - Water distribution pipe infrastructure
2. **sensor** - IoT sensor for monitoring (pressure, flow, temperature, quality)
3. **valve** - Control valve for flow regulation
4. **pump** - Water pump for pressure management
5. **reservoir** - Water storage reservoir/tank
6. **hydrant** - Fire hydrant infrastructure
7. **meter** - Water consumption meter

## API Usage

### List All Agent Types

```bash
GET /api/v1/agent-types
GET /api/v1/agent-types?category=infrastructure
GET /api/v1/agent-types?enabled=true
```

Response:
```json
{
  "agent_types": [
    {
      "id": "pipe",
      "name": "Pipe Agent",
      "description": "Water distribution pipe infrastructure agent",
      "category": "infrastructure",
      "version": "1.0.0",
      "is_enabled": true,
      "is_system_type": false
    }
  ],
  "count": 12
}
```

### Get Specific Agent Type

```bash
GET /api/v1/agent-types/pipe
```

Response:
```json
{
  "id": "pipe",
  "name": "Pipe Agent",
  "description": "Water distribution pipe infrastructure agent",
  "category": "infrastructure",
  "version": "1.0.0",
  "schema": { ... },
  "required_capabilities": ["flow_monitoring", "pressure_monitoring"],
  "optional_capabilities": ["leak_detection", "condition_assessment"],
  "default_config": {
    "monitoring_interval": "60s"
  },
  "is_system_type": false,
  "is_enabled": true,
  "created_at": "2025-10-22T10:00:00Z",
  "updated_at": "2025-10-22T10:00:00Z"
}
```

### Create Custom Agent Type

```bash
POST /api/v1/agent-types
Content-Type: application/json

{
  "id": "custom-sensor",
  "name": "Custom Sensor Agent",
  "description": "Custom sensor for specific monitoring",
  "category": "monitoring",
  "version": "1.0.0",
  "schema": {
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "required": ["sensor_id"],
    "properties": {
      "sensor_id": {
        "type": "string",
        "pattern": "^CUSTOM-[0-9]+$"
      }
    }
  },
  "required_capabilities": ["data_collection"],
  "is_enabled": true
}
```

### Update Agent Type

```bash
PUT /api/v1/agent-types/custom-sensor
Content-Type: application/json

{
  "id": "custom-sensor",
  "name": "Custom Sensor Agent v2",
  "description": "Updated description",
  ...
}
```

### Enable/Disable Agent Type

```bash
POST /api/v1/agent-types/custom-sensor/enable
POST /api/v1/agent-types/custom-sensor/disable
```

### Delete Agent Type

```bash
DELETE /api/v1/agent-types/custom-sensor
```

Note: System types (is_system_type: true) cannot be deleted.

## Programmatic Usage

### Initialize Registry

```go
// Create repository
agentTypeRepo := registry.NewInMemoryAgentTypeRepository()

// Create service
agentTypeService := registry.NewAgentTypeService(agentTypeRepo, logger)

// Initialize default types
ctx := context.Background()
err := registry.InitializeDefaultAgentTypes(ctx, agentTypeService, logger)
```

### Register Custom Type

```go
agentType := &registry.AgentType{
    ID:          "custom-agent",
    Name:        "Custom Agent Type",
    Description: "Domain-specific agent",
    Category:    "custom",
    Version:     "1.0.0",
    RequiredCapabilities: []string{"task_execution"},
    IsEnabled:   true,
}

err := agentTypeService.RegisterType(ctx, agentType)
```

### Validate Agent Type

```go
isValid, err := agentTypeService.IsValidType(ctx, "pipe")
if isValid {
    // Agent type is registered and enabled
}
```

### Validate Agent Configuration

```go
config := map[string]interface{}{
    "pipe_id": "PIPE-001",
    "material": "PVC",
    "diameter": 300,
    "length": 500.0,
}

err := agentTypeService.ValidateAgentConfig(ctx, "pipe", config)
```

## Schema Examples

### Pipe Agent Schema

```json
{
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
    },
    "length": {
      "type": "number",
      "minimum": 0
    },
    "pressure_rating": {
      "type": "number",
      "minimum": 0
    }
  }
}
```

## Validator Integration

The configuration and template validators automatically use the Agent Type Registry when available:

```go
// Configuration validator
validator := configuration.NewDefaultValidatorWithTypeService(resourceChecker, agentTypeService)

// Template validator
templateValidator := templates.NewDefaultValidatorWithTypeService(agentTypeService)

// Or set after creation
validator.SetAgentTypeService(agentTypeService)
```

When the service is set, validators will:
1. Check if the agent type is registered
2. Verify it's enabled
3. Validate configuration against the type's JSON schema
4. Apply custom validation rules

If the service is not set, validators fall back to the hardcoded list of core types.

## Benefits

1. **Extensibility**: Add new agent types without code changes
2. **Use Case Support**: Each use case can define its domain-specific agents
3. **Validation**: Automatic schema-based validation ensures data integrity
4. **Organization**: Category-based organization for better management
5. **Safety**: System types are protected from accidental deletion
6. **Flexibility**: Enable/disable types without removing definitions

## Use Case: Water Distribution Network

The infrastructure agent types (pipe, sensor, valve, pump, reservoir, hydrant, meter) are now registered and ready to use. You can:

1. Create agents of these types via API
2. Configuration validates against the specific schemas
3. Templates can use these types
4. Each type has appropriate capabilities defined

Example agent creation:
```json
POST /api/v1/agents
{
  "name": "Main Distribution Line - North",
  "type": "pipe",
  "config": {
    "pipe_id": "PIPE-001",
    "material": "PVC",
    "diameter": 300,
    "length": 500.0,
    "pressure_rating": 200.0,
    "flow_capacity": 2000.0
  }
}
```

The system will:
1. Validate "pipe" is a registered and enabled type
2. Validate the config against the pipe schema
3. Create the agent if validation passes

## Future Enhancements

- [ ] ArangoDB persistence for agent types
- [ ] Versioning support for schema evolution
- [ ] Agent type dependencies (e.g., "sensor requires pipe")
- [ ] Capability validation framework
- [ ] Agent type templates/inheritance
- [ ] Import/export agent type definitions
- [ ] Web UI for agent type management
