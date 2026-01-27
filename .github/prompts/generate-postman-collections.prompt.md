---
agent: agent
---

# Generate Postman Collections from Routes

This prompt scans the CodeValdCortex API routes and generates organized Postman collection files.

## Objective

Create comprehensive Postman collection files by analyzing the route definitions in the codebase. Collections should be:
- **Well-organized** by domain/feature area
- **Max 500 lines per file** for maintainability
- **Complete** with examples, descriptions, and test scripts
- **Structured** in `/workspaces/CodeValdCortex/documents/4-QA/postmanCalls/`

## Steps

### 1. Scan Route Definitions

**Primary route files to analyze:**
- `/workspaces/CodeValdCortex/internal/app/routes_api.go` - Main API v1 routes
- `/workspaces/CodeValdCortex/internal/app/routes_web.go` - Web/Templ routes (if needed)
- Handler files in `/workspaces/CodeValdCortex/internal/handlers/` - Handler implementations
- Handler files in `/workspaces/CodeValdCortex/internal/web/handlers/` - Web handlers

**Extract from each route:**
- HTTP method (GET, POST, PUT, DELETE, PATCH, OPTIONS)
- Path/URL pattern (e.g., `/api/v1/agencies/:id`)
- Handler function name
- Request body structure (if POST/PUT/PATCH)
- Query parameters
- Path parameters
- Expected response format

### 2. Group Routes by Domain

**Organize routes into logical domains** (max 500 lines per file):

#### Domain Categories:
1. **Health & System** (`01-health-system.postman.json`)
   - Health checks
   - Status endpoints
   - CORS testing

2. **Agencies** (`02-agencies.postman.json`)
   - Agency CRUD
   - Specifications
   - Statistics
   - Roles management

3. **Work Items (Issues)** (`03-work-items.postman.json`)
   - List with pagination/filtering
   - CRUD operations
   - Assignment operations
   - Status updates

4. **Instances** (`04-instances.postman.json`)
   - Instance lifecycle
   - Instance management
   - Health checks

5. **Tags & Publications** (`05-tags-publications.postman.json`)
   - Tag management
   - Publication operations
   - Activation/deactivation

6. **Workflows** (`06-workflows.postman.json`)
   - Workflow CRUD
   - Workflow execution
   - Validation

7. **AI Services** (`07-ai-services.postman.json`)
   - AI refinement endpoints
   - Agency designer
   - Goal/Role/RACI generation

8. **Webhooks & Integration** (`08-webhooks.postman.json`)
   - Gitea webhooks
   - Work item integration
   - PR automation

### 3. Postman Collection Structure Template

Each collection file should follow this structure:

```json
{
  "info": {
    "_postman_id": "unique-id-{domain}",
    "name": "CodeValdCortex - {Domain Name}",
    "description": "Detailed description of domain and endpoints",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Endpoint Group",
      "description": "Group description",
      "item": [
        {
          "name": "Endpoint Name",
          "request": {
            "method": "GET|POST|PUT|DELETE|PATCH|OPTIONS",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json",
                "type": "text"
              },
              {
                "key": "Origin",
                "value": "http://localhost:5173",
                "type": "text",
                "description": "For CORS testing"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"example\": \"request body\"\n}",
              "options": {
                "raw": {
                  "language": "json"
                }
              }
            },
            "url": {
              "raw": "{{base_url}}/api/v1/path/:param?query=value",
              "host": ["{{base_url}}"],
              "path": ["api", "v1", "path", ":param"],
              "query": [
                {
                  "key": "query",
                  "value": "value",
                  "description": "Query parameter description"
                }
              ],
              "variable": [
                {
                  "key": "param",
                  "value": "{{param_variable}}"
                }
              ]
            },
            "description": "Detailed endpoint description with usage examples"
          },
          "response": [],
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test(\"Status code is 200\", function () {",
                  "    pm.response.to.have.status(200);",
                  "});",
                  "",
                  "pm.test(\"Response has CORS headers\", function () {",
                  "    pm.response.to.have.header(\"Access-Control-Allow-Origin\");",
                  "});"
                ],
                "type": "text/javascript"
              }
            }
          ]
        }
      ]
    }
  ],
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8082",
      "type": "string"
    }
  ]
}
```

### 4. Include Request Examples

For each endpoint, provide realistic examples:

**GET Requests:**
- Show query parameters with example values
- Include pagination examples (limit, offset)
- Include filter examples (status, type, etc.)

**POST/PUT Requests:**
- Provide complete request body examples
- Show required vs optional fields
- Include validation examples

**DELETE Requests:**
- Show confirmation patterns
- Include cascade deletion notes

### 5. Add Test Scripts

Include Postman test scripts for each request:

```javascript
// Status check
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

// CORS headers (MVP-RM-002)
pm.test("CORS headers present", function () {
    pm.response.to.have.header("Access-Control-Allow-Origin");
    pm.expect(pm.response.headers.get("Access-Control-Allow-Origin")).to.include("localhost:5173");
});

// Response structure validation
pm.test("Response has data property", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property("data");
});

// Pagination validation (for list endpoints)
pm.test("Response has pagination", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property("pagination");
    pm.expect(jsonData.pagination).to.have.property("total");
    pm.expect(jsonData.pagination).to.have.property("limit");
    pm.expect(jsonData.pagination).to.have.property("offset");
});

// Save IDs for subsequent requests
if (pm.response.code === 201) {
    var jsonData = pm.response.json();
    pm.environment.set("created_id", jsonData.data.id);
}
```

### 6. Collection Variables

Each collection should use these common variables:

```json
"variable": [
  {
    "key": "base_url",
    "value": "http://localhost:8082",
    "type": "string"
  },
  {
    "key": "agency_id",
    "value": "UC-INFRA-001",
    "type": "string"
  },
  {
    "key": "instance_id",
    "value": "",
    "type": "string"
  },
  {
    "key": "issue_id",
    "value": "",
    "type": "string"
  },
  {
    "key": "tag_name",
    "value": "v1.0.0",
    "type": "string"
  },
  {
    "key": "workflow_id",
    "value": "",
    "type": "string"
  }
]
```

### 7. File Organization

**Directory structure:**
```
/workspaces/CodeValdCortex/documents/4-QA/postmanCalls/
├── README.md                           # Overview and usage instructions
├── postman_environment_local.json      # Local environment variables
├── postman_environment_production.json # Production environment (optional)
├── collections/
│   ├── 01-health-system.postman.json
│   ├── 02-agencies.postman.json
│   ├── 03-work-items.postman.json
│   ├── 04-instances.postman.json
│   ├── 05-tags-publications.postman.json
│   ├── 06-workflows.postman.json
│   ├── 07-ai-services.postman.json
│   └── 08-webhooks.postman.json
└── examples/
    ├── sample-agency.json
    ├── sample-work-item.json
    └── sample-workflow.json
```

### 8. Route Analysis Strategy

**Read these files to extract routes:**

```bash
# Main API routes
grep -E "v1\.(GET|POST|PUT|DELETE|PATCH|OPTIONS)" /workspaces/CodeValdCortex/internal/app/routes_api.go

# Extract handler patterns
grep -E "func.*Handler.*\(c \*gin\.Context\)" /workspaces/CodeValdCortex/internal/handlers/*.go
grep -E "func.*Handler.*\(c \*gin\.Context\)" /workspaces/CodeValdCortex/internal/web/handlers/*.go

# Look for route groups
grep -E "Group\(" /workspaces/CodeValdCortex/internal/app/routes_api.go
```

### 9. Special Considerations

**CORS Testing (MVP-RM-002):**
- Include OPTIONS preflight requests
- Add Origin header to test requests
- Verify Access-Control-* headers in responses

**Pagination (MVP-RM-003):**
- Include limit/offset query parameters
- Test pagination metadata in responses
- Show hasMore flag usage

**Agency Isolation:**
- Always include agency_id in paths
- Show instance-scoped endpoints
- Document multi-tenancy patterns

**Error Handling:**
- Include examples of error responses
- Show validation error formats
- Document error codes

## Output Format

For each domain, create a separate Postman collection file with:
- **Name**: Following numbering convention (01-, 02-, etc.)
- **Size**: Max 500 lines (split if needed)
- **Structure**: Consistent with template above
- **Documentation**: Clear descriptions and examples
- **Tests**: Validation scripts for common scenarios

## Success Criteria

- ✅ All API routes from routes_api.go are documented
- ✅ Each collection file is ≤ 500 lines
- ✅ All endpoints include request examples
- ✅ CORS testing examples included (MVP-RM-002)
- ✅ Work items pagination examples included (MVP-RM-003)
- ✅ Test scripts validate responses
- ✅ Variables are used consistently
- ✅ README.md explains usage
- ✅ Files are organized by domain

## Example Output

**File**: `collections/03-work-items.postman.json`
**Size**: ~450 lines
**Contains**:
- List work items with pagination (GET)
- Get single work item (GET)
- Create work item (POST)
- Update work item (PUT)
- Delete work item (DELETE)
- Assign work item (POST)
- All with test scripts and examples

---

**Start by:**
1. Reading `/workspaces/CodeValdCortex/internal/app/routes_api.go`
2. Identifying all route groups
3. Creating the directory structure
4. Generating collection files one by one
5. Testing each collection works with actual API
