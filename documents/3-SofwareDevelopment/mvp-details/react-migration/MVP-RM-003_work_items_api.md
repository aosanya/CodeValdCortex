# MVP-RM-003: Work Items REST API

**Domain**: React Migration  
**Priority**: P1  
**Effort**: Medium  
**Skills**: Go, Gin Framework, ArangoDB  
**Dependencies**: MVP-RM-002 ✅

---

## Overview

Implement RESTful API endpoints for Work Items domain. This provides the backend API layer that the React frontend will consume for CRUD operations on work items.

---

## API Specification

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### 1. List Work Items
```http
GET /api/v1/agencies/{agencyID}/work-items
```

**Query Parameters:**
- `type` (optional): Filter by work item type
- `status` (optional): Filter by status
- `search` (optional): Search in title/description
- `limit` (optional): Pagination limit (default: 50)
- `offset` (optional): Pagination offset (default: 0)

**Response (200 OK):**
```json
{
  "data": [{
    "_key": "work-item-001",
    "title": "Implement auth",
    "description": "Add JWT auth",
    "agency_id": "UC-INFRA-001",
    "type": "task",
    "status": "active",
    "autonomy_level": "1-supervised",
    "created_at": "2026-01-20T10:00:00Z",
    "updated_at": "2026-01-25T15:30:00Z"
  }],
  "pagination": {
    "total": 42,
    "limit": 50,
    "offset": 0,
    "has_more": false
  }
}
```

#### 2. Get Single Work Item
```http
GET /api/v1/agencies/{agencyID}/work-items/{workItemID}
```

#### 3. Create Work Item
```http
POST /api/v1/agencies/{agencyID}/work-items
```

**Request Body:**
```json
{
  "title": "New work item",
  "description": "Detailed description",
  "type": "task",
  "status": "active",
  "autonomy_level": "1-supervised"
}
```

#### 4. Update Work Item
```http
PUT /api/v1/agencies/{agencyID}/work-items/{workItemID}
```

#### 5. Delete Work Item
```http
DELETE /api/v1/agencies/{agencyID}/work-items/{workItemID}
```

**Response:** 204 No Content

---

## Implementation

### Package Structure
```
internal/api/v1/workitems/
├── handler.go          # HTTP handlers
├── routes.go           # Route registration
└── validation.go       # Request validation
```

### Handler Implementation

**Handler (`internal/api/v1/workitems/handler.go`):**
```go
package workitems

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/yourusername/codevaldcortex/internal/agency"
    "github.com/yourusername/codevaldcortex/internal/api/response"
)

type Handler struct {
    workItemService *agency.WorkItemService
}

func NewHandler(workItemService *agency.WorkItemService) *Handler {
    return &Handler{
        workItemService: workItemService,
    }
}

func (h *Handler) List(c *gin.Context) {
    agencyID := c.Param("agencyID")
    
    // Parse query params
    filters := parseFilters(c)
    
    // Get work items from service
    items, total, err := h.workItemService.List(c.Request.Context(), agencyID, filters)
    if err != nil {
        response.InternalError(c, err)
        return
    }
    
    // Build pagination metadata
    pagination := &response.PaginationMeta{
        Total:   total,
        Limit:   filters.Limit,
        Offset:  filters.Offset,
        HasMore: filters.Offset+len(items) < total,
    }
    
    response.SuccessWithPagination(c, http.StatusOK, items, pagination)
}

func (h *Handler) Get(c *gin.Context) {
    agencyID := c.Param("agencyID")
    workItemID := c.Param("workItemID")
    
    item, err := h.workItemService.GetByID(c.Request.Context(), agencyID, workItemID)
    if err != nil {
        response.NotFoundError(c, "work item", workItemID)
        return
    }
    
    response.Success(c, http.StatusOK, item)
}

func (h *Handler) Create(c *gin.Context) {
    agencyID := c.Param("agencyID")
    
    var req CreateWorkItemRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "body", "Invalid request format")
        return
    }
    
    if err := validateCreateRequest(&req); err != nil {
        response.ValidationError(c, err.Field, err.Message)
        return
    }
    
    item := req.ToWorkItem(agencyID)
    created, err := h.workItemService.Create(c.Request.Context(), agencyID, item)
    if err != nil {
        response.InternalError(c, err)
        return
    }
    
    response.Success(c, http.StatusCreated, created)
}

func (h *Handler) Update(c *gin.Context) {
    agencyID := c.Param("agencyID")
    workItemID := c.Param("workItemID")
    
    var req UpdateWorkItemRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "body", "Invalid request format")
        return
    }
    
    updated, err := h.workItemService.Update(c.Request.Context(), agencyID, workItemID, req.ToUpdates())
    if err != nil {
        response.InternalError(c, err)
        return
    }
    
    response.Success(c, http.StatusOK, updated)
}

func (h *Handler) Delete(c *gin.Context) {
    agencyID := c.Param("agencyID")
    workItemID := c.Param("workItemID")
    
    err := h.workItemService.Delete(c.Request.Context(), agencyID, workItemID)
    if err != nil {
        response.InternalError(c, err)
        return
    }
    
    c.Status(http.StatusNoContent)
}
```

**Routes (`internal/api/v1/workitems/routes.go`):**
```go
package workitems

import (
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *Handler) {
    agencies := rg.Group("/agencies/:agencyID")
    {
        workItems := agencies.Group("/work-items")
        {
            workItems.GET("", handler.List)
            workItems.POST("", handler.Create)
            workItems.GET("/:workItemID", handler.Get)
            workItems.PUT("/:workItemID", handler.Update)
            workItems.DELETE("/:workItemID", handler.Delete)
        }
    }
}
```

**Validation (`internal/api/v1/workitems/validation.go`):**
```go
package workitems

import "errors"

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return e.Message
}

func validateCreateRequest(req *CreateWorkItemRequest) *ValidationError {
    if req.Title == "" {
        return &ValidationError{Field: "title", Message: "Title is required"}
    }
    if len(req.Title) > 200 {
        return &ValidationError{Field: "title", Message: "Title must be under 200 characters"}
    }
    if req.Type == "" {
        return &ValidationError{Field: "type", Message: "Type is required"}
    }
    return nil
}
```

---

## Acceptance Criteria

- [ ] All 5 endpoints implemented and working
- [ ] Proper error handling with standardized responses
- [ ] Input validation for create/update operations
- [ ] Pagination working correctly with metadata
- [ ] Query parameter filtering implemented
- [ ] Returns correct HTTP status codes
- [ ] Agency-specific data isolation enforced
- [ ] Tests written for all endpoints
- [ ] API integrated with v1 router
- [ ] Documentation updated with examples

---

## Testing

### Integration Tests

```bash
# List work items
curl http://localhost:8080/api/v1/agencies/UC-INFRA-001/work-items

# Create work item
curl -X POST http://localhost:8080/api/v1/agencies/UC-INFRA-001/work-items \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Item","type":"task","status":"active"}'

# Get single work item
curl http://localhost:8080/api/v1/agencies/UC-INFRA-001/work-items/work-item-001

# Update work item
curl -X PUT http://localhost:8080/api/v1/agencies/UC-INFRA-001/work-items/work-item-001 \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title","status":"completed"}'

# Delete work item
curl -X DELETE http://localhost:8080/api/v1/agencies/UC-INFRA-001/work-items/work-item-001
```

---

**Created**: January 26, 2026
