# Instance Management: Service Layer & Repository

**Related Task**: MVP-PUB-007  
**Component**: Business Logic Layer

---

## Instance Service Interface

```go
// internal/agency/services/instance_service.go
type InstanceService interface {
    // StartInstance creates and starts a new instance from a tag
    StartInstance(ctx context.Context, agencyID, tagName string, req *StartInstanceRequest) (*models.AgencyInstance, error)
    
    // StopInstance gracefully stops a running instance
    StopInstance(ctx context.Context, agencyID, instanceID string) error
    
    // RestartInstance stops and restarts an instance
    RestartInstance(ctx context.Context, agencyID, instanceID string) error
    
    // GetInstance retrieves instance details
    GetInstance(ctx context.Context, agencyID, instanceID string) (*models.AgencyInstance, error)
    
    // ListInstances lists instances with filtering
    ListInstances(ctx context.Context, agencyID string, filters *InstanceFilters) ([]*models.AgencyInstance, error)
    
    // GetInstanceHealth retrieves health status
    GetInstanceHealth(ctx context.Context, agencyID, instanceID string) (*InstanceHealth, error)
    
    // DeleteInstance soft-deletes a stopped instance
    DeleteInstance(ctx context.Context, agencyID, instanceID string) error
}
```

---

## Service Implementation

### StartInstance Flow

```go
func (s *instanceService) StartInstance(ctx, agencyID, tagName, req) (*models.AgencyInstance, error) {
    // 1. Validate instance name is unique per agency
    exists, err := s.instanceRepo.ExistsByName(ctx, agencyID, req.Name, agencyDB)
    if exists {
        return nil, errors.New("instance name must be unique per agency")
    }
    
    // 2. Retrieve tag from agency-specific database
    tag, err := s.tagRepo.GetByAgencyAndName(ctx, agencyID, tagName, agencyDB)
    if err != nil {
        return nil, fmt.Errorf("tag not found: %w", err)
    }
    
    // 3. Create instance record with immediate "running" state
    now := time.Now()
    instance := &models.AgencyInstance{
        AgencyID:       agencyID,
        TagID:          tag.ID,
        TagName:        tag.Name,
        InstanceID:     generateInstanceID(),
        Name:           req.Name,
        Description:    req.Description,
        State:          models.InstanceStateRunning, // Running immediately
        DeployedAt:     now,
        StartedAt:      &now,
        DeployedBy:     getCurrentUser(ctx),
        AcceptsNewJobs: true,
        Tags:           req.Tags,
    }
    
    // 4. Save instance to agency_instances collection
    err = s.instanceRepo.Create(ctx, instance, agencyID, agencyDB)
    if err != nil {
        return nil, fmt.Errorf("failed to create instance: %w", err)
    }
    
    // 5. Store agent references from tag snapshot (async)
    go s.storeAgentReferences(ctx, instance, tag)
    
    return instance, nil
}
```

### StopInstance Flow (Graceful Shutdown)

```go
func (s *instanceService) StopInstance(ctx, agencyID, instanceID) error {
    // 1. Get instance
    instance, err := s.instanceRepo.GetByInstanceID(ctx, instanceID, agencyDB)
    if err != nil {
        return fmt.Errorf("instance not found: %w", err)
    }
    
    // 2. Validate instance can be stopped
    if instance.State == models.InstanceStateStopped {
        return errors.New("instance already stopped")
    }
    
    // 3. Mark instance as stopping and reject new jobs
    instance.State = models.InstanceStateStopping
    instance.AcceptsNewJobs = false
    err = s.instanceRepo.Update(ctx, instance, agencyDB)
    if err != nil {
        return fmt.Errorf("failed to update instance state: %w", err)
    }
    
    // 4. Signal graceful shutdown (async)
    go s.gracefulShutdown(ctx, instance, 30*time.Second) // 30s timeout
    
    return nil
}

func (s *instanceService) gracefulShutdown(ctx, instance, timeout) {
    // Wait for current tasks with timeout
    select {
    case <-s.waitForCompletion(instance):
        // All tasks completed gracefully
        s.logger.Info("instance stopped gracefully", "instance_id", instance.InstanceID)
    case <-time.After(timeout):
        // Force stop after timeout
        s.logger.Warn("instance force stopped after timeout", "instance_id", instance.InstanceID)
        s.forceStop(instance)
    }
    
    // Mark as stopped
    instance.State = models.InstanceStateStopped
    now := time.Now()
    instance.StoppedAt = &now
    s.instanceRepo.Update(ctx, instance, agencyDB)
}
```

### GetInstanceHealth (On-Demand Calculation)

```go
func (s *instanceService) GetInstanceHealth(ctx, agencyID, instanceID) (*InstanceHealth, error) {
    // 1. Get instance
    instance, err := s.instanceRepo.GetByInstanceID(ctx, instanceID, agencyDB)
    if err != nil {
        return nil, err
    }
    
    // 2. Get agent references
    agentRefs, err := s.instanceRepo.GetInstanceAgentReferences(ctx, instanceID, agencyDB)
    if err != nil {
        return nil, err
    }
    
    // 3. Calculate health based on agent states
    health := &InstanceHealth{
        InstanceID:   instanceID,
        LastCheck:    time.Now(),
        AgentsHealthy: 0,
        AgentsDegraded: 0,
        AgentsUnhealthy: 0,
    }
    
    for _, agentRef := range agentRefs {
        // Query actual agent state (if instantiated)
        agentState := s.getAgentState(ctx, agentRef, agencyDB)
        switch agentState {
        case "healthy":
            health.AgentsHealthy++
        case "degraded":
            health.AgentsDegraded++
        case "unhealthy":
            health.AgentsUnhealthy++
        }
    }
    
    // 4. Determine overall health
    totalAgents := len(agentRefs)
    if totalAgents == 0 {
        health.HealthStatus = "healthy" // No agents yet
    } else if health.AgentsUnhealthy > totalAgents/2 {
        health.HealthStatus = "unhealthy"
    } else if health.AgentsDegraded > 0 || health.AgentsUnhealthy > 0 {
        health.HealthStatus = "degraded"
    } else {
        health.HealthStatus = "healthy"
    }
    
    return health, nil
}
```

---

## Instance Repository Interface

```go
// internal/agency/arangodb/instance_repository.go
type InstanceRepository interface {
    // CRUD operations
    Create(ctx context.Context, instance *models.AgencyInstance, agencyID, agencyDB string) error
    GetByID(ctx context.Context, agencyID, instanceID, agencyDB string) (*models.AgencyInstance, error)
    GetByInstanceID(ctx context.Context, instanceID, agencyDB string) (*models.AgencyInstance, error)
    ExistsByName(ctx context.Context, agencyID, name, agencyDB string) (bool, error)
    List(ctx context.Context, agencyID, agencyDB string, filters *services.InstanceFilters) ([]*models.AgencyInstance, error)
    Update(ctx context.Context, instance *models.AgencyInstance, agencyDB string) error
    SoftDelete(ctx context.Context, agencyID, instanceID, deletedBy, agencyDB string) error
    
    // Agent reference tracking
    LinkAgentReference(ctx context.Context, instanceID, agentRoleCode, tagID, agencyDB string) error
    GetInstanceAgentReferences(ctx context.Context, instanceID, agencyDB string) ([]AgentReference, error)
}

type AgentReference struct {
    InstanceID    string    `json:"instance_id"`
    AgentRoleCode string    `json:"agent_role_code"`
    TagID         string    `json:"tag_id"`
    State         string    `json:"state"`
    AddedAt       time.Time `json:"added_at"`
}
```

---

## Key Implementation Details

### Graceful Shutdown

**Process**:
1. Instance state → `stopping`
2. `AcceptsNewJobs` → `false`
3. Wait for current tasks to complete (30s timeout)
4. Force stop if timeout exceeded
5. Instance state → `stopped`

**Job Rejection**:
```go
func (s *instanceService) CanAcceptJob(ctx, instanceID) (bool, error) {
    instance, err := s.instanceRepo.GetByInstanceID(ctx, instanceID, agencyDB)
    if err != nil {
        return false, err
    }
    
    return instance.AcceptsNewJobs && instance.State == models.InstanceStateRunning, nil
}
```

### Agent Reference Storage

```go
func (s *instanceService) storeAgentReferences(ctx, instance, tag) {
    // Extract agent configurations from tag
    for _, roleConfig := range tag.Roles {
        err := s.instanceRepo.LinkAgentReference(
            ctx,
            instance.InstanceID,
            roleConfig.RoleCode,
            tag.ID,
            agencyDB,
        )
        if err != nil {
            s.logger.Error("failed to link agent reference", 
                "instance_id", instance.InstanceID,
                "role_code", roleConfig.RoleCode,
                "error", err)
        }
    }
}
```

### Unique Name Validation

```go
func (r *instanceRepository) ExistsByName(ctx, agencyID, name, agencyDB) (bool, error) {
    query := `
        FOR instance IN agency_instances
        FILTER instance.agency_id == @agencyID
        FILTER instance.name == @name
        FILTER instance.is_deleted == false
        LIMIT 1
        RETURN instance
    `
    
    cursor, err := r.db.Query(ctx, query, map[string]interface{}{
        "agencyID": agencyID,
        "name":     name,
    })
    
    return cursor.HasMore(), err
}
```

---

## Related Files

- **Data Models**: [instance-data-models.md](instance-data-models.md)
- **API Endpoints**: [instance-api.md](instance-api.md)
- **UI Components**: [instance-ui.md](instance-ui.md)
- **Task Overview**: [instance-management.md](instance-management.md)
