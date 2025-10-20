# CodeValdCortex - QA Testing Documentation

This directory contains testing resources and documentation for quality assurance of the CodeValdCortex platform.

## 📋 Contents

### Postman Collections

- **postman_agent_runtime.json** - Agent Runtime Environment API tests (MVP-002)
  - Agent lifecycle management: Create, Start, Stop
  - Task submission and tracking
  - Runtime metrics and health checks
  - Port: 8082 (local development)

### Environments

- **postman_environment_local.json** - Local development environment settings

## 🚀 Getting Started with Postman Tests

### Prerequisites

- Postman Desktop App or Postman Web
- CodeValdCortex running locally or access to a deployed instance

### Import Collection and Environment

1. **Open Postman**

2. **Import the Collection**:
   - Click "Import" button in Postman
   - Select `postman_agent_runtime.json` (MVP-002 Agent Runtime tests)
   - The collection will appear in your workspace

3. **Import the Environment**:
   - Click "Import" button
   - Select `postman_environment_local.json` (for local testing)

4. **Select the Environment**:
   - In the top-right corner, select the imported environment from the dropdown

### Running Tests

#### Run Entire Collection

1. Click on the "CodeValdCortex - Agent Runtime (MVP-002)" collection
2. Click "Run" button
3. Configure run settings:
   - Select all requests or specific folders
   - Set delay between requests (optional)
   - Choose number of iterations
4. Click "Run CodeValdCortex - Agent Runtime (MVP-002)"

#### Run Individual Requests

1. Navigate to specific request in the collection
2. Click "Send" button
3. View response and test results in the bottom panel

### Test Scenarios

#### 1. Health & Status Tests

**Purpose**: Verify system availability and basic health

**Endpoints**:
- `GET /health` - Application health check
- `GET /api/v1/status` - System status information

**Expected Results**:
- 200 OK response
- Response time < 200ms
- Status field shows "healthy"

#### 2. Agent Management Tests

**Purpose**: Validate agent lifecycle operations

**Test Flow**:
1. List existing agents
2. Create a new test agent
3. Get agent details
4. Update agent configuration
5. Scale agent replicas
6. Delete agent

**Expected Results**:
- Agent CRUD operations succeed
- Agent ID is properly generated and returned
- Configuration changes are reflected
- Scaling operations complete successfully

#### 3. Workflow Management Tests

**Purpose**: Test workflow orchestration capabilities

**Test Flow**:
1. List existing workflows
2. Create workflow definition
3. Execute workflow with parameters
4. Monitor execution status
5. Verify completion

**Expected Results**:
- Workflow creation returns workflow ID
- Execution returns execution ID
- Status updates correctly (pending → running → completed)

#### 4. Metrics & Monitoring Tests

**Purpose**: Verify observability endpoints

**Endpoints**:
- `GET /api/v1/metrics` - JSON metrics
- `GET /metrics` - Prometheus format

**Expected Results**:
- Metrics data is returned
- Data includes relevant system metrics

## 🔑 Authentication

If authentication is enabled:

1. **Obtain Auth Token**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "password"}'
   ```

2. **Set Token in Environment**:
   - Go to "Environments" in Postman
   - Edit your active environment
   - Set `auth_token` variable value to the received token

3. **Token Auto-Refresh**:
   - The collection includes pre-request scripts for token refresh (if implemented)

## 📊 Test Assertions

Each request includes automated test scripts that verify:

- **HTTP Status Codes**: Correct status codes for success/failure
- **Response Structure**: Required fields are present
- **Response Time**: Performance within acceptable limits
- **Data Validation**: Values match expected formats
- **State Management**: Operations update state correctly

### Example Test Script

```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has required fields", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('id');
    pm.expect(jsonData).to.have.property('status');
});

pm.test("Response time is acceptable", function () {
    pm.expect(pm.response.responseTime).to.be.below(200);
});
```

## 🔄 Environment Variables

### Automatic Variables

Variables automatically set during test execution:

- `agent_id` - Set after agent creation, used in subsequent requests
- `workflow_id` - Set after workflow creation
- `execution_id` - Set after workflow execution

### Manual Configuration

Variables you may need to set manually:

- `base_url` - API base URL (default: http://localhost:8080)
- `auth_token` - Authentication token (if auth is enabled)

## 🐛 Troubleshooting

### Connection Refused

**Problem**: Cannot connect to API

**Solution**:
```bash
# Verify service is running
curl http://localhost:8080/health

# Check if correct port
ps aux | grep codevaldcortex

# Start service if needed
make run-dev
```

### Authentication Failures

**Problem**: 401 Unauthorized responses

**Solution**:
- Verify `auth_token` is set correctly in environment
- Check token hasn't expired
- Obtain new token if necessary

### Test Failures

**Problem**: Tests failing unexpectedly

**Solution**:
1. Run requests individually to isolate issue
2. Check response body for error details
3. Verify environment variables are set correctly
4. Ensure service is in clean state (restart if needed)

## 📈 CI/CD Integration

### Running Tests in CI Pipeline

```yaml
# Example GitHub Actions integration
- name: Run API Tests
  run: |
    newman run documents/4-QA/postman_collection.json \
      -e documents/4-QA/postman_environment_local.json \
      --reporters cli,json \
      --reporter-json-export test-results.json
```

### Using Newman (CLI)

Install Newman:
```bash
npm install -g newman
```

Run collection:
```bash
newman run postman_collection.json \
  -e postman_environment_local.json \
  --reporters cli,htmlextra \
  --reporter-htmlextra-export report.html
```

## 📝 Best Practices

1. **Run Tests in Order**: Some tests depend on previous test results (e.g., agent creation before update)

2. **Clean State**: Reset environment between test runs for consistent results

3. **Monitor Performance**: Pay attention to response times and resource usage

4. **Update Collection**: Keep collection in sync with API changes

5. **Version Control**: Commit collection and environment files to git

6. **Document Changes**: Update this README when adding new tests

## 🤝 Contributing

When adding new tests:

1. Create descriptive test names
2. Include appropriate assertions
3. Add error handling
4. Document expected behavior
5. Update this README

## 📚 Additional Resources

- [Postman Documentation](https://learning.postman.com/)
- [Newman CLI Documentation](https://github.com/postmanlabs/newman)
- [CodeValdCortex API Documentation](../api.md)
- [Test Writing Guide](./test-writing-guide.md)

---

For questions or issues, please open a GitHub issue or contact the development team.