# Quality Assurance - MVP-013 REST API Layer

## Overview

This directory contains comprehensive testing resources for the CodeValdCortex REST API Layer (MVP-013). The API provides complete agent management, configuration, templates, task orchestration, communication, and monitoring capabilities.

## Test Collections

### 1. MVP-013 REST API Collection (`postman_mvp013_rest_api.json`)

**Purpose**: Complete API testing for MVP-013 REST API Layer
**Version**: 3.0.0
**Base URL**: `http://localhost:8080`

#### Test Categories:

##### Health & System
- **Health Check**: `GET /health` - Basic system health status
- **System Info**: `GET /api/v1/info` - System information and capabilities

##### Agent Management
- **CRUD Operations**: Create, Read, Update, Delete agents
- **Lifecycle Control**: Start, stop, restart, pause, resume agents
- **Status Monitoring**: Agent status, health, metrics, memory usage
- **Pool Management**: Agent pool operations

##### Configuration Management
- **Configuration CRUD**: Create, read, update, delete configurations
- **Version Control**: Configuration versions and rollback
- **Template Integration**: Create configurations from templates
- **Import/Export**: Configuration data portability
- **Validation**: Configuration validation and compatibility checks

##### Template Management
- **Template CRUD**: Create, read, update, delete templates
- **Rendering**: Template rendering with variable substitution
- **Validation**: Template syntax and structure validation
- **Variable Management**: Template variable definitions and defaults

##### Task Management
- **Task CRUD**: Create, read, update, cancel tasks
- **Task Control**: Retry, result retrieval, log access
- **Workflow Management**: Multi-step workflow creation and execution
- **Workflow Visualization**: Workflow graph and dependency tracking

##### Communication
- **Message Management**: Send, receive, list messages
- **Channel Management**: Create and manage communication channels
- **Statistics**: Communication metrics and statistics

##### Monitoring & Metrics
- **System Metrics**: Overall system performance metrics
- **Agent Metrics**: Agent-specific performance data
- **Resource Metrics**: System resource utilization
- **Health Monitoring**: Service health checks (agents, services, database)

##### Administration
- **System Configuration**: Configuration management and reload
- **System Statistics**: Runtime statistics and diagnostics
- **Maintenance**: System maintenance operations
- **Diagnostics**: System diagnostic information

#### Test Variables:

```json
{
    "base_url": "http://localhost:8080",
    "agent_id": "",
    "config_id": "",
    "template_id": "",
    "task_id": "",
    "workflow_id": "",
    "message_id": "",
    "channel_id": ""
}
```

### 2. Legacy MVP-002 Collection (`postman_agent_runtime.json`)

**Purpose**: Original agent runtime testing (maintained for backward compatibility)
**Version**: 2.0.0
**Base URL**: `http://localhost:8082`

This collection contains the original MVP-002 agent runtime tests and should be used for regression testing when making changes to core agent functionality.

## Environment Configuration

### Local Development (`postman_environment_local.json`)

- **Base URL**: `http://localhost:8080`
- **API Version**: `v1`
- **Environment**: Development mode with debug logging

### Production Environment

Create a separate environment file for production testing with:
- HTTPS endpoints
- Authentication tokens
- Production-specific configuration

## Testing Strategy

### 1. Smoke Tests
Run basic health and system info endpoints to verify API server is operational:
```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/info
```

### 2. Integration Tests
Use the complete Postman collection to test all API endpoints:
1. Import `postman_mvp013_rest_api.json` into Postman
2. Import `postman_environment_local.json` as environment
3. Run the entire collection to test all endpoints

### 3. Load Testing
For performance testing:
1. Use tools like `wrk` or `artillery` for load testing
2. Focus on high-traffic endpoints (agent listing, metrics)
3. Test concurrent agent operations

### 4. Security Testing
- Verify proper HTTP headers are set
- Test input validation and sanitization
- Check for proper error handling
- Validate CORS configuration

## Current Implementation Status

### ✅ Implemented
- Complete API server infrastructure
- Middleware stack (logging, CORS, security headers, recovery)
- Standardized response formats and error handling
- Comprehensive endpoint routing structure
- Health check and system info endpoints

### 🚧 In Progress (MVP-013)
- Agent management endpoint implementations
- Configuration management endpoint implementations
- Template management endpoint implementations
- Task and workflow management
- Communication system integration
- Monitoring and metrics collection

### ⏳ Planned (Future MVPs)
- Authentication and authorization
- Rate limiting implementation
- WebSocket support for real-time updates
- Advanced monitoring dashboards
- API versioning strategy

## Usage Instructions

### 1. Start the API Server
```bash
go run examples/api_server.go
# or
go run cmd/main.go --api-only
```

### 2. Verify Server is Running
```bash
curl http://localhost:8080/health
```

### 3. Import Postman Collection
1. Open Postman
2. Import `postman_mvp013_rest_api.json`
3. Import `postman_environment_local.json`
4. Select the local environment
5. Run individual requests or the entire collection

### 4. Run Automated Tests
```bash
# Using Newman (Postman CLI)
newman run postman_mvp013_rest_api.json -e postman_environment_local.json
```

## Development Guidelines

### Adding New Endpoints
1. Define the endpoint in `internal/api/server.go`
2. Implement the handler function
3. Add corresponding Postman test in the collection
4. Update this README with the new endpoint documentation

### Test Data Management
- Use Postman variables for dynamic test data
- Include realistic test payloads in request bodies
- Set up proper test assertions for response validation

### Error Testing
- Test both successful and error scenarios
- Verify proper HTTP status codes
- Check error response format consistency

## Troubleshooting

### Common Issues
1. **Server not starting**: Check port 8080 is not in use
2. **404 errors**: Verify API server is running with correct routing
3. **CORS issues**: Check CORS middleware configuration
4. **Request timeout**: Increase timeout settings for slow operations

### Debug Mode
Start the server with debug logging:
```bash
go run examples/api_server.go --debug
```

### Logs Analysis
Check server logs for:
- Request/response details
- Error stack traces
- Performance metrics
- Middleware execution

## Contributing

When adding new API endpoints:
1. Follow RESTful conventions
2. Use consistent response formats
3. Add comprehensive Postman tests
4. Update documentation
5. Include error scenarios in tests

## References

- [REST API Design Guidelines](https://restfulapi.net/)
- [Postman Documentation](https://learning.postman.com/)
- [Go Gin Framework](https://gin-gonic.com/)
- [HTTP Status Codes](https://httpstatuses.com/)