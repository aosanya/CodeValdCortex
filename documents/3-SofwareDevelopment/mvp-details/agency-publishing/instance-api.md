# Instance Management: HTTP API Endpoints

**Related Task**: MVP-PUB-007  
**Component**: API Layer  
**Research Reference**: See [instance-research-session.md](instance-research-session.md) for architectural Q&A

---

## API Endpoints

```go
// internal/handlers/instance_handler.go

POST   /api/v1/agencies/:id/instances                        // Start new instance from tag
GET    /api/v1/agencies/:id/instances                        // List instances (supports filtering)
GET    /api/v1/agencies/:id/instances/:instance_id           // Get instance dashboard data
DELETE /api/v1/agencies/:id/instances/:instance_id           // Soft delete stopped instance
POST   /api/v1/agencies/:id/instances/:instance_id/stop      // Stop instance (graceful shutdown)
POST   /api/v1/agencies/:id/instances/:instance_id/restart   // Restart instance
GET    /api/v1/agencies/:id/instances/:instance_id/health    // Get health status (calculated on-demand)
GET    /api/v1/agencies/:id/instances/:instance_id/agents    // List instance agent references
POST   /api/v1/agencies/:id/instances/:instance_id/accept-job // Check if instance accepts new jobs
```

---

## Endpoint Details

### 1. POST /api/v1/agencies/:id/instances

**Description**: Start a new instance from a tag

**Request Body**:
```json
{
  "tag_name": "v1.0.0",
  "name": "Production Instance",
  "description": "Main production deployment",
  "tags": ["production", "stable"]
}
```

**Response** (201 Created):
```json
{
  "instance_id": "inst-uuid-001",
  "name": "Production Instance",
  "state": "running",
  "tag_id": "tag-uuid-123",
  "tag_name": "v1.0.0",
  "deployed_at": "2025-11-24T10:00:00Z",
  "accepts_new_jobs": true
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or tag not found
- `409 Conflict`: Instance name already exists for this agency

**Handler Implementation**:
```go
func (h *InstanceHandler) StartInstance(c *gin.Context) {
    agencyID := c.Param("id")
    
    var req services.StartInstanceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }
    
    instance, err := h.instanceService.StartInstance(c.Request.Context(), agencyID, req.TagName, &req)
    if err != nil {
        if strings.Contains(err.Error(), "unique") {
            c.JSON(409, gin.H{"error": err.Error()})
        } else {
            c.JSON(400, gin.H{"error": err.Error()})
        }
        return
    }
    
    c.JSON(201, instance)
}
```

---

### 2. GET /api/v1/agencies/:id/instances

**Description**: List all instances for an agency with optional filtering

**Query Parameters**:
- `tag_id` (optional): Filter by tag ID
- `state` (optional): Filter by state (running, stopping, stopped, failed)
- `from_date` (optional): Filter instances deployed after this date
- `to_date` (optional): Filter instances deployed before this date
- `limit` (optional): Max results (default: 50)
- `offset` (optional): Pagination offset

**Response** (200 OK):
```json
{
  "instances": [
    {
      "instance_id": "inst-uuid-001",
      "name": "Production Instance",
      "state": "running",
      "tag_name": "v1.0.0",
      "deployed_at": "2025-11-24T10:00:00Z",
      "agent_count": 5,
      "uptime_seconds": 3600
    }
  ],
  "total": 10,
  "limit": 50,
  "offset": 0
}
```

---

### 3. GET /api/v1/agencies/:id/instances/:instance_id

**Description**: Get instance details for dashboard

**Response** (200 OK):
```json
{
  "instance_id": "inst-uuid-001",
  "name": "Production Instance",
  "description": "Main production deployment",
  "state": "running",
  "health_status": "healthy",
  "tag_id": "tag-uuid-123",
  "tag_name": "v1.0.0",
  "deployed_at": "2025-11-24T10:00:00Z",
  "deployed_by": "user@example.com",
  "started_at": "2025-11-24T10:00:00Z",
  "uptime_seconds": 3600,
  "agent_count": 5,
  "workflow_count": 3,
  "accepts_new_jobs": true,
  "tags": ["production", "stable"]
}
```

**Error Responses**:
- `404 Not Found`: Instance not found

---

### 4. POST /api/v1/agencies/:id/instances/:instance_id/stop

**Description**: Stop instance with graceful shutdown (30s timeout)

**Response** (200 OK):
```json
{
  "message": "Instance entering graceful shutdown",
  "timeout_seconds": 30
}
```

**Error Responses**:
- `400 Bad Request`: Instance already stopped
- `404 Not Found`: Instance not found

**Handler Implementation**:
```go
func (h *InstanceHandler) StopInstance(c *gin.Context) {
    agencyID := c.Param("id")
    instanceID := c.Param("instance_id")
    
    err := h.instanceService.StopInstance(c.Request.Context(), agencyID, instanceID)
    if err != nil {
        if strings.Contains(err.Error(), "not found") {
            c.JSON(404, gin.H{"error": err.Error()})
        } else {
            c.JSON(400, gin.H{"error": err.Error()})
        }
        return
    }
    
    c.JSON(200, gin.H{
        "message": "Instance entering graceful shutdown",
        "timeout_seconds": 30,
    })
}
```

---

### 5. POST /api/v1/agencies/:id/instances/:instance_id/restart

**Description**: Restart a stopped instance

**Response** (200 OK):
```json
{
  "message": "Instance restarted",
  "state": "running"
}
```

**Error Responses**:
- `400 Bad Request`: Instance is not stopped
- `404 Not Found`: Instance not found

---

### 6. GET /api/v1/agencies/:id/instances/:instance_id/health

**Description**: Get instance health status (calculated on-demand)

**Response** (200 OK):
```json
{
  "instance_id": "inst-uuid-001",
  "health_status": "healthy",
  "agents_healthy": 5,
  "agents_degraded": 0,
  "agents_unhealthy": 0,
  "last_check": "2025-11-24T11:30:00Z"
}
```

---

### 7. GET /api/v1/agencies/:id/instances/:instance_id/agents

**Description**: List agent references for an instance

**Response** (200 OK):
```json
{
  "agent_references": [
    {
      "agent_role_code": "developer-agent-001",
      "tag_id": "tag-uuid-123",
      "state": "referenced",
      "added_at": "2025-11-24T10:00:00Z"
    }
  ]
}
```

---

### 8. DELETE /api/v1/agencies/:id/instances/:instance_id

**Description**: Soft delete a stopped instance

**Response** (200 OK):
```json
{
  "message": "Instance deleted",
  "instance_id": "inst-uuid-001"
}
```

**Error Responses**:
- `400 Bad Request`: Instance is not stopped
- `404 Not Found`: Instance not found

---

### 9. POST /api/v1/agencies/:id/instances/:instance_id/accept-job

**Description**: Check if instance can accept new jobs

**Response** (200 OK):
```json
{
  "accepts_jobs": true,
  "state": "running"
}
```

**Response** (503 Service Unavailable):
```json
{
  "accepts_jobs": false,
  "state": "stopping",
  "reason": "Instance is shutting down"
}
```

---

## Handler Structure

```go
type InstanceHandler struct {
    instanceService services.InstanceService
    agencyRepo      agency.Repository
    logger          *logrus.Logger
}

func NewInstanceHandler(
    instanceService services.InstanceService,
    agencyRepo agency.Repository,
    logger *logrus.Logger,
) *InstanceHandler {
    return &InstanceHandler{
        instanceService: instanceService,
        agencyRepo:      agencyRepo,
        logger:          logger,
    }
}

func (h *InstanceHandler) RegisterRoutes(router *gin.RouterGroup) {
    instances := router.Group("/agencies/:id/instances")
    {
        instances.POST("", h.StartInstance)
        instances.GET("", h.ListInstances)
        instances.GET("/:instance_id", h.GetInstance)
        instances.DELETE("/:instance_id", h.DeleteInstance)
        instances.POST("/:instance_id/stop", h.StopInstance)
        instances.POST("/:instance_id/restart", h.RestartInstance)
        instances.GET("/:instance_id/health", h.GetInstanceHealth)
        instances.GET("/:instance_id/agents", h.GetInstanceAgents)
        instances.POST("/:instance_id/accept-job", h.CanAcceptJob)
    }
}
```

---

## Error Handling

All endpoints follow consistent error response format:

```json
{
  "error": "descriptive error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

**Common Error Codes**:
- `INSTANCE_NOT_FOUND`: Instance doesn't exist
- `INSTANCE_NAME_EXISTS`: Duplicate instance name
- `INVALID_STATE_TRANSITION`: Can't perform operation in current state
- `TAG_NOT_FOUND`: Referenced tag doesn't exist
- `VALIDATION_ERROR`: Invalid request parameters

---

## Related Files

- **Data Models**: [instance-data-models.md](instance-data-models.md)
- **Service Layer**: [instance-services.md](instance-services.md)
- **UI Components**: [instance-ui.md](instance-ui.md)
- **Task Overview**: [instance-management.md](instance-management.md)
