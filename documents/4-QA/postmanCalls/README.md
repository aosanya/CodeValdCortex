# CodeValdCortex - Postman API Collections

Comprehensive Postman collections for testing and documenting the CodeValdCortex API endpoints.

## 📁 Collection Files

All collections are organized in the `collections/` directory:

| Collection | Description | Key Features |
|-----------|-------------|--------------|
| [00-authentication.postman.json](collections/00-authentication.postman.json) | Authentication & User Management | JWT dual-token auth, user registration/login, token refresh, logout (MVP-AUTH-001 to MVP-AUTH-004) |
| [01-health-system.postman.json](collections/01-health-system.postman.json) | Health checks and system status | Basic health endpoints, CORS testing |
| [02-agencies.postman.json](collections/02-agencies.postman.json) | Agency CRUD and specifications | Full specification management, roles, RACI matrix |
| [03-work-items.postman.json](collections/03-work-items.postman.json) | Work item/issue management | Pagination (MVP-RM-003), filtering, assignment operations |
| [04-instances.postman.json](collections/04-instances.postman.json) | Instance lifecycle | Start, stop, restart, health checks (MVP-PUB-007) |
| [05-tags-publications.postman.json](collections/05-tags-publications.postman.json) | Tags and publications | Versioning, activation, lifecycle (MVP-PUB-003, MVP-PUB-004) |
| [06-workflows.postman.json](collections/06-workflows.postman.json) | Workflow operations | CRUD, validation, duplication |
| [07a-ai-goals-roles.postman.json](collections/07a-ai-goals-roles.postman.json) | AI Services: Goals & Roles | Introduction, goals, roles refinement |
| [07b-ai-raci-deliverables.postman.json](collections/07b-ai-raci-deliverables.postman.json) | AI Services: RACI & Deliverables | RACI matrix and deliverable management |
| [08-webhooks.postman.json](collections/08-webhooks.postman.json) | Gitea webhooks | Issue sync, pull request automation |

## 🚀 Quick Start

### 1. Import Collections

**Option A: Import All Collections**
```bash
# In Postman, go to File → Import
# Select all .json files from collections/ directory
# Click Import
```

**Option B: Import Individual Collections**
- Open Postman
- Click **Import** button
- Drag & drop the desired collection file
- Click **Import**

### 2. Import Environment

Import the local environment configuration:
- File: [postman_environment_local.json](postman_environment_local.json)
- Contains variables for `base_url`, `agency_id`, etc.

### 3. Set Environment Variables

Update these variables in your Postman environment:

```json
{
  "base_url": "http://localhost:8082",
  "agency_id": "UC-INFRA-001",
  "tag_name": "v1.0.0",
  "instance_id": "",
  "issue_id": "",
  "workflow_id": ""
}
```

**Note**: `instance_id`, `issue_id`, and `workflow_id` are auto-populated by test scripts after creation.

## 📋 Collection Details

### Health & System (01)
- **Purpose**: Verify API availability and CORS configuration
- **Key Tests**:
  - Health check returns 200 with "healthy" status
  - CORS headers present (MVP-RM-002)
  - API status endpoint responds

### Agencies (02)
- **Purpose**: Manage agency lifecycle and specifications
- **Endpoints**: 25+ endpoints covering:
  - Agency CRUD operations
  - Unified specification management (introduction, goals, work items, workflows, roles, RACI)
  - Individual section updates
  - Role management
  - Statistics

### Work Items (03)
- **Purpose**: Issue/work item management with pagination
- **Key Features**:
  - **Pagination** (MVP-RM-003): `limit` and `offset` parameters
  - Test scripts validate pagination metadata (`total`, `limit`, `offset`, `hasMore`)
  - Filtering by status and assignee
  - Assignment and claiming operations
  - CORS header validation

**Example pagination request**:
```
GET /api/v1/agencies/UC-INFRA-001/instances/{instance_id}/issues?limit=10&offset=0&status=open
```

### Instances (04)
- **Purpose**: Instance lifecycle management (MVP-PUB-007)
- **Operations**:
  - Start instance from tag
  - Health checks
  - Stop/Restart operations
  - Agent listing
  - Job submission

### Tags & Publications (05)
- **Purpose**: Version control and publication workflow
- **Sections**:
  - **Tags**: Create snapshots, restore, compare
  - **Publications**: Validate, publish, activation
  - **Lifecycle**: Pause, resume, drain, stop (MVP-PUB-004)

### Workflows (06)
- **Purpose**: Workflow definition and management
- **Features**:
  - JSON and HTML output formats
  - Validation before creation
  - Workflow duplication

### AI Services (07)
- **Purpose**: AI-powered agency design and refinement
- **Capabilities**:
  - Introduction refinement
  - Dynamic goal operations (refine, generate, consolidate)
  - Role enhancement
  - RACI matrix generation
  - Natural language prompts for all operations

**Example AI request**:
```json
{
  "prompt": "Enhance all goals to be more specific and measurable"
}
```

### Webhooks (08)
- **Purpose**: Gitea integration for work tracking
- **Events**:
  - Issue events (opened, updated, closed, assigned)
  - Pull request events (opened, merged, closed)
- **Headers**: Includes `X-Gitea-Event` and `X-Gitea-Signature` for verification

## 🧪 Running Tests

### Test All Collections

1. Select collection in Postman
2. Click **Run** (Runner icon)
3. Select environment
4. Click **Run Collection**

### Test Scripts Included

All collections include test scripts that verify:
- ✅ Status codes (200, 201, 204)
- ✅ Response structure
- ✅ CORS headers (MVP-RM-002)
- ✅ Pagination metadata (MVP-RM-003)
- ✅ Auto-save created IDs to environment variables

**Example test script**:
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("CORS headers present", function () {
    pm.response.to.have.header("Access-Control-Allow-Origin");
});

pm.test("Response has pagination", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData.pagination).to.have.property('total');
    pm.expect(jsonData.pagination).to.have.property('hasMore');
});
```

## 📂 Example Files

The `examples/` directory contains sample request bodies:

- [sample-agency.json](examples/sample-agency.json) - Complete agency creation payload
- [sample-work-item.json](examples/sample-work-item.json) - Issue creation example
- [sample-workflow.json](examples/sample-workflow.json) - Workflow definition template

## 🔄 Common Workflows

### Create and Publish Agency

1. **Create Agency** (02-agencies → Create Agency)
2. **Update Specification** (02-agencies → Update Full Specification)
3. **Validate** (05-tags-publications → Validate for Publish)
4. **Create Tag** (05-tags-publications → Create Tag)
5. **Publish** (05-tags-publications → Publish Agency)
6. **Activate** (05-tags-publications → Activate Agency)

### Start Instance and Manage Work

1. **Start Instance** (04-instances → Start Instance from Tag)
   - Auto-saves `instance_id` to environment
2. **Create Issue** (03-work-items → Create Issue)
   - Auto-saves `issue_id` to environment
3. **Assign Issue** (03-work-items → Assign Issue to Agent)
4. **Progress Issue** (03-work-items → Progress Issue)
5. **Check Instance Health** (04-instances → Get Instance Health)

### AI-Powered Agency Design

1. **Refine Introduction** (07-ai-services → Refine Introduction)
2. **Generate Goals** (07-ai-services → Generate Goal with Prompt)
3. **Create Roles** (07-ai-services → Generate Role with Prompt)
4. **Generate RACI Matrix** (07-ai-services → Create Complete RACI Matrix)
5. **Enhance All** (07-ai-services → Enhance All Roles)

## 🛠️ Troubleshooting

### CORS Errors
- Ensure `Origin: http://localhost:5173` header is included in requests
- Check API is running with CORS middleware enabled
- Verify `Access-Control-Allow-Origin` header in responses

### Pagination Not Working
- Verify `limit` and `offset` query parameters are set
- Check response includes `pagination` object with `total`, `limit`, `offset`, `hasMore`
- See MVP-RM-003 requirements

### Webhook Signature Verification Fails
- Ensure `X-Gitea-Signature` header is correctly formatted
- Verify webhook secret matches configuration
- Check signature calculation method (SHA-256 HMAC)

### Environment Variables Not Set
- Check test scripts ran successfully (green checkmarks)
- Verify environment is selected in Postman
- Manually set variables if auto-save failed

## 📖 API Documentation

For detailed API specifications, see:
- [Backend Architecture](../../2-SoftwareDesignAndArchitecture/backend-architecture.md)
- [routes_api.go](/workspaces/CodeValdCortex/internal/app/routes_api.go) - Complete route definitions

## 🔗 Related MVP Requirements

- **MVP-RM-002**: CORS Support (tested in Health & System collection)
- **MVP-RM-003**: Pagination (tested in Work Items collection)
- **MVP-WI-008**: Work Item Management (Work Items collection)
- **MVP-PUB-003**: Publication Workflow (Tags & Publications collection)
- **MVP-PUB-004**: Activation Lifecycle (Tags & Publications collection)
- **MVP-PUB-007**: Instance Management (Instances collection)

## 📝 Notes

- All collections use `{{base_url}}` variable (default: `http://localhost:8082`)
- Test scripts auto-populate IDs (`instance_id`, `issue_id`, etc.) in environment
- CORS testing requires setting `Origin` header
- Webhook collections include sample Gitea payloads
- AI service endpoints require AI provider configuration in `config.yaml`

## 🤝 Contributing

When adding new endpoints:
1. Add to appropriate collection (or create new if >500 lines)
2. Include test scripts for status codes and response validation
3. Use environment variables for IDs and URLs
4. Add CORS header tests where applicable
5. Update this README with new endpoints

---

**Last Updated**: January 26, 2026  
**Version**: 1.0.0  
**Total Collections**: 8  
**Total Endpoints**: 80+
