# Data Coordination & Export Services

This document covers the ArangoDB integration, data coordination, configuration template management, and enterprise integration services.

## 3. Data Coordination Services

### 3.1 ArangoDB Integration and Change Streams

#### Data Coordination Service
```go
type DataCoordinationService struct {
    db              arangodb.Database
    changeProcessors map[string]*ChangeStreamProcessor
    conflictResolver *ConflictResolver
    eventStore      *EventStore
    logger          *zap.Logger
    mutex           sync.RWMutex
}

func NewDataCoordinationService(db arangodb.Database) *DataCoordinationService {
    return &DataCoordinationService{
        db:               db,
        changeProcessors: make(map[string]*ChangeStreamProcessor),
        conflictResolver: NewConflictResolver(),
        eventStore:      NewEventStore(db),
        logger:          zap.NewNop(),
    }
}

func (dcs *DataCoordinationService) RegisterAgent(
    ctx context.Context, agent *Agent) error {
    
    // Store agent information in document store
    agentDoc := map[string]interface{}{
        "_key":        agent.ID,
        "config":      agent.Config,
        "status":      agent.Status,
        "created_at":  agent.CreatedAt,
        "updated_at":  time.Now(),
        "version":     1,
    }
    
    collection := dcs.db.Collection("agents")
    _, err := collection.CreateDocument(ctx, agentDoc)
    if err != nil {
        return fmt.Errorf("failed to register agent: %w", err)
    }
    
    // Start change stream processor for this agent
    processor, err := dcs.startChangeStreamProcessor(agent.ID)
    if err != nil {
        return fmt.Errorf("failed to start change processor: %w", err)
    }
    
    dcs.mutex.Lock()
    dcs.changeProcessors[agent.ID] = processor
    dcs.mutex.Unlock()
    
    // Create event for agent registration
    event := &Event{
        Type:      EventTypeAgentRegistered,
        AgentID:   agent.ID,
        Timestamp: time.Now(),
        Data:      agentDoc,
    }
    
    if err := dcs.eventStore.StoreEvent(ctx, event); err != nil {
        dcs.logger.Warn("failed to store agent registration event",
            zap.String("agent_id", agent.ID),
            zap.Error(err))
    }
    
    return nil
}

func (dcs *DataCoordinationService) UpdateAgentState(
    ctx context.Context, agentID string, state interface{}) error {
    
    collection := dcs.db.Collection("agent_states")
    
    // Prepare state document with versioning
    stateDoc := map[string]interface{}{
        "_key":       agentID,
        "agent_id":   agentID,
        "state":      state,
        "timestamp":  time.Now(),
        "version":    time.Now().UnixNano(), // Use timestamp as version
    }
    
    // Try to update with conflict detection
    meta, err := collection.ReplaceDocument(ctx, agentID, stateDoc)
    if driver.IsConflict(err) {
        // Handle conflict using conflict resolution strategy
        resolvedState, err := dcs.conflictResolver.ResolveStateConflict(
            ctx, agentID, state, collection)
        if err != nil {
            return fmt.Errorf("conflict resolution failed: %w", err)
        }
        
        stateDoc["state"] = resolvedState
        stateDoc["version"] = time.Now().UnixNano()
        
        meta, err = collection.ReplaceDocument(ctx, agentID, stateDoc)
        if err != nil {
            return fmt.Errorf("failed to update after conflict resolution: %w", err)
        }
    } else if err != nil {
        return fmt.Errorf("failed to update agent state: %w", err)
    }
    
    dcs.logger.Debug("agent state updated",
        zap.String("agent_id", agentID),
        zap.String("revision", meta.Rev))
    
    return nil
}

func (dcs *DataCoordinationService) WatchAgentChanges(
    ctx context.Context, agentID string) (<-chan *StateChange, error) {
    
    changeStream := make(chan *StateChange, 100)
    
    go func() {
        defer close(changeStream)
        
        // Watch for changes in agent state collection
        query := `
            FOR doc IN agent_states
            FILTER doc.agent_id == @agentId
            RETURN { "new": doc, "old": OLD }
        `
        
        cursor, err := dcs.db.Query(ctx, query, map[string]interface{}{
            "agentId": agentID,
        })
        
        if err != nil {
            dcs.logger.Error("failed to create change stream",
                zap.String("agent_id", agentID),
                zap.Error(err))
            return
        }
        defer cursor.Close()
        
        for cursor.HasMore() {
            var changeDoc struct {
                New interface{} `json:"new"`
                Old interface{} `json:"old"`
            }
            
            if _, err := cursor.ReadDocument(ctx, &changeDoc); err != nil {
                dcs.logger.Error("failed to read change document",
                    zap.Error(err))
                continue
            }
            
            change := &StateChange{
                AgentID:   agentID,
                NewState:  changeDoc.New,
                OldState:  changeDoc.Old,
                Timestamp: time.Now(),
            }
            
            select {
            case changeStream <- change:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return changeStream, nil
}

func (dcs *DataCoordinationService) GetAgentPool(
    ctx context.Context, poolID string) (*AgentPool, error) {
    
    collection := dcs.db.Collection("agent_pools")
    
    var poolDoc map[string]interface{}
    _, err := collection.ReadDocument(ctx, poolID, &poolDoc)
    if err != nil {
        if driver.IsNotFound(err) {
            return nil, ErrAgentPoolNotFound
        }
        return nil, fmt.Errorf("failed to read agent pool: %w", err)
    }
    
    // Get agents in this pool
    agentQuery := `
        FOR agent IN agents
        FILTER agent.pool_id == @poolId
        RETURN agent
    `
    
    cursor, err := dcs.db.Query(ctx, agentQuery, map[string]interface{}{
        "poolId": poolID,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to query pool agents: %w", err)
    }
    defer cursor.Close()
    
    var agents []*Agent
    for cursor.HasMore() {
        var agentDoc map[string]interface{}
        if _, err := cursor.ReadDocument(ctx, &agentDoc); err != nil {
            continue
        }
        
        agent := dcs.mapDocumentToAgent(agentDoc)
        agents = append(agents, agent)
    }
    
    pool := &AgentPool{
        ID:          poolDoc["_key"].(string),
        Name:        poolDoc["name"].(string),
        Description: poolDoc["description"].(string),
        Agents:      agents,
        Template:    poolDoc["template"].(map[string]interface{}),
        ScalingPolicy: poolDoc["scaling_policy"].(map[string]interface{}),
        CreatedAt:   poolDoc["created_at"].(time.Time),
        UpdatedAt:   poolDoc["updated_at"].(time.Time),
    }
    
    return pool, nil
}
}
```

### 3.2 Configuration Template Management

#### Template Storage and Versioning Service
```go
type ConfigurationTemplateService struct {
    db              arangodb.Database
    templateStore   *TemplateStore
    versionManager  *VersionManager
    validator       *TemplateValidator
    logger          *zap.Logger
}

func NewConfigurationTemplateService(db arangodb.Database) *ConfigurationTemplateService {
    return &ConfigurationTemplateService{
        db:             db,
        templateStore:  NewTemplateStore(db),
        versionManager: NewVersionManager(),
        validator:      NewTemplateValidator(),
        logger:         zap.NewNop(),
    }
}

func (cts *ConfigurationTemplateService) CreateTemplate(
    ctx context.Context, template *AgentTemplate) error {
    
    // Validate template configuration
    if err := cts.validator.ValidateTemplate(template); err != nil {
        return fmt.Errorf("template validation failed: %w", err)
    }
    
    // Assign version
    template.Version = cts.versionManager.GenerateVersion()
    template.CreatedAt = time.Now()
    template.UpdatedAt = time.Now()
    
    // Store template in database
    templateDoc := map[string]interface{}{
        "_key":         template.ID,
        "name":         template.Name,
        "description":  template.Description,
        "version":      template.Version,
        "configuration": template.Configuration,
        "resource_requirements": template.ResourceRequirements,
        "environment_variables": template.EnvironmentVariables,
        "security_settings": template.SecuritySettings,
        "created_at":   template.CreatedAt,
        "updated_at":   template.UpdatedAt,
        "created_by":   template.CreatedBy,
        "tags":         template.Tags,
    }
    
    collection := cts.db.Collection("agent_templates")
    _, err := collection.CreateDocument(ctx, templateDoc)
    if err != nil {
        return fmt.Errorf("failed to store template: %w", err)
    }
    
    cts.logger.Info("agent template created",
        zap.String("template_id", template.ID),
        zap.String("version", template.Version))
    
    return nil
}

func (cts *ConfigurationTemplateService) GetTemplate(
    ctx context.Context, templateID string) (*AgentTemplate, error) {
    
    collection := cts.db.Collection("agent_templates")
    
    var templateDoc map[string]interface{}
    _, err := collection.ReadDocument(ctx, templateID, &templateDoc)
    if err != nil {
        if driver.IsNotFound(err) {
            return nil, ErrTemplateNotFound
        }
        return nil, fmt.Errorf("failed to read template: %w", err)
    }
    
    template := &AgentTemplate{
        ID:          templateDoc["_key"].(string),
        Name:        templateDoc["name"].(string),
        Description: templateDoc["description"].(string),
        Version:     templateDoc["version"].(string),
        Configuration: templateDoc["configuration"].(map[string]interface{}),
        ResourceRequirements: templateDoc["resource_requirements"].(map[string]interface{}),
        EnvironmentVariables: templateDoc["environment_variables"].(map[string]interface{}),
        SecuritySettings: templateDoc["security_settings"].(map[string]interface{}),
        CreatedBy:   templateDoc["created_by"].(string),
        Tags:        templateDoc["tags"].([]string),
    }
    
    // Parse timestamps
    if createdAt, ok := templateDoc["created_at"].(time.Time); ok {
        template.CreatedAt = createdAt
    }
    if updatedAt, ok := templateDoc["updated_at"].(time.Time); ok {
        template.UpdatedAt = updatedAt
    }
    
    return template, nil
}

func (cts *ConfigurationTemplateService) ListTemplates(
    ctx context.Context, filters *TemplateFilters) ([]*AgentTemplate, error) {
    
    query := `
        FOR template IN agent_templates
        FILTER @filters == null OR (
            (@filters.tags == null OR LENGTH(INTERSECTION(template.tags, @filters.tags)) > 0) AND
            (@filters.created_by == null OR template.created_by == @filters.created_by) AND
            (@filters.name_pattern == null OR LIKE(template.name, @filters.name_pattern))
        )
        SORT template.updated_at DESC
        RETURN template
    `
    
    bindVars := map[string]interface{}{
        "filters": filters,
    }
    
    cursor, err := cts.db.Query(ctx, query, bindVars)
    if err != nil {
        return nil, fmt.Errorf("template query failed: %w", err)
    }
    defer cursor.Close()
    
    var templates []*AgentTemplate
    for cursor.HasMore() {
        var templateDoc map[string]interface{}
        if _, err := cursor.ReadDocument(ctx, &templateDoc); err != nil {
            continue
        }
        
        template := cts.mapDocumentToTemplate(templateDoc)
        templates = append(templates, template)
    }
    
    return templates, nil
}

func (cts *ConfigurationTemplateService) mapDocumentToTemplate(
    doc map[string]interface{}) *AgentTemplate {
    
    template := &AgentTemplate{
        ID:          doc["_key"].(string),
        Name:        doc["name"].(string),
        Description: doc["description"].(string),
        Version:     doc["version"].(string),
        Configuration: doc["configuration"].(map[string]interface{}),
        ResourceRequirements: doc["resource_requirements"].(map[string]interface{}),
        EnvironmentVariables: doc["environment_variables"].(map[string]interface{}),
        SecuritySettings: doc["security_settings"].(map[string]interface{}),
        CreatedBy:   doc["created_by"].(string),
        Tags:        doc["tags"].([]string),
    }
    
    // Parse timestamps
    if createdAt, ok := doc["created_at"].(time.Time); ok {
        template.CreatedAt = createdAt
    }
    if updatedAt, ok := doc["updated_at"].(time.Time); ok {
        template.UpdatedAt = updatedAt
    }
    
    return template
}

func (cts *ConfigurationTemplateService) CloneTemplate(
    ctx context.Context, sourceTemplateID, newName string) (*AgentTemplate, error) {
    
    // Get source template
    sourceTemplate, err := cts.GetTemplate(ctx, sourceTemplateID)
    if err != nil {
        return nil, fmt.Errorf("failed to get source template: %w", err)
    }
    
    // Create new template with cloned configuration
    newTemplate := &AgentTemplate{
        ID:          generateTemplateID(),
        Name:        newName,
        Description: fmt.Sprintf("Cloned from %s", sourceTemplate.Name),
        Configuration: deepCopyMap(sourceTemplate.Configuration),
        ResourceRequirements: deepCopyMap(sourceTemplate.ResourceRequirements),
        EnvironmentVariables: deepCopyMap(sourceTemplate.EnvironmentVariables),
        SecuritySettings: deepCopyMap(sourceTemplate.SecuritySettings),
        CreatedBy:   sourceTemplate.CreatedBy,
        Tags:        append([]string{"cloned"}, sourceTemplate.Tags...),
    }
    
    // Save the new template
    if err := cts.CreateTemplate(ctx, newTemplate); err != nil {
        return nil, fmt.Errorf("failed to create cloned template: %w", err)
    }
    
    return newTemplate, nil
}

// Helper functions for template management
func generateTemplateID() string {
    return fmt.Sprintf("template-%d", time.Now().UnixNano())
}

func deepCopyMap(original map[string]interface{}) map[string]interface{} {
    copy := make(map[string]interface{})
    for key, value := range original {
        switch v := value.(type) {
        case map[string]interface{}:
            copy[key] = deepCopyMap(v)
        default:
            copy[key] = v
        }
    }
    return copy
}

type TemplateValidator struct {
    requiredFields []string
    validators     map[string]func(interface{}) error
}

func NewTemplateValidator() *TemplateValidator {
    return &TemplateValidator{
        requiredFields: []string{"name", "configuration", "resource_requirements"},
        validators: map[string]func(interface{}) error{
            "resource_requirements": validateResourceRequirements,
            "security_settings":     validateSecuritySettings,
            "environment_variables": validateEnvironmentVariables,
        },
    }
}

func (tv *TemplateValidator) ValidateTemplate(template *AgentTemplate) error {
    // Check required fields
    if template.Name == "" {
        return fmt.Errorf("template name is required")
    }
    
    if len(template.Configuration) == 0 {
        return fmt.Errorf("template configuration is required")
    }
    
    if len(template.ResourceRequirements) == 0 {
        return fmt.Errorf("resource requirements are required")
    }
    
    // Run field-specific validators
    for field, validator := range tv.validators {
        var value interface{}
        switch field {
        case "resource_requirements":
            value = template.ResourceRequirements
        case "security_settings":
            value = template.SecuritySettings
        case "environment_variables":
            value = template.EnvironmentVariables
        }
        
        if err := validator(value); err != nil {
            return fmt.Errorf("validation failed for %s: %w", field, err)
        }
    }
    
    return nil
}
```

## 4. Data Export and Integration

### 3.2 Enterprise Integration Services

#### API Gateway Service
```go
type APIGatewayService struct {
    router           *mux.Router
    authProvider     *EnterpriseAuthProvider
    rateLimiter      *RateLimiter
    orchestrator     *AgentOrchestratorService
    coordinator      *DataCoordinationService
    logger           *zap.Logger
    metricsCollector *MetricsCollector
}

func NewAPIGatewayService(
    authProvider *EnterpriseAuthProvider,
    orchestrator *AgentOrchestratorService,
    coordinator *DataCoordinationService) *APIGatewayService {
    
    svc := &APIGatewayService{
        router:           mux.NewRouter(),
        authProvider:     authProvider,
        rateLimiter:      NewRateLimiter(),
        orchestrator:     orchestrator,
        coordinator:      coordinator,
        logger:           zap.NewNop(),
        metricsCollector: NewMetricsCollector(),
    }
    
    svc.setupRoutes()
    return svc
}

func (ags *APIGatewayService) setupRoutes() {
    // Apply global middleware
    ags.router.Use(ags.authenticationMiddleware)
    ags.router.Use(ags.rateLimitingMiddleware)
    ags.router.Use(ags.auditLoggingMiddleware)
    ags.router.Use(ags.metricsMiddleware)
    
    // Agent management endpoints
    agentAPI := ags.router.PathPrefix("/api/v1/agents").Subrouter()
    agentAPI.HandleFunc("", ags.handleListAgents).Methods("GET")
    agentAPI.HandleFunc("", ags.handleDeployAgent).Methods("POST")
    agentAPI.HandleFunc("/{agentId}", ags.handleGetAgent).Methods("GET")
    agentAPI.HandleFunc("/{agentId}", ags.handleDeleteAgent).Methods("DELETE")
    agentAPI.HandleFunc("/{agentId}/scale", ags.handleScaleAgent).Methods("POST")
    agentAPI.HandleFunc("/{agentId}/state", ags.handleGetAgentState).Methods("GET")
    agentAPI.HandleFunc("/{agentId}/state", ags.handleUpdateAgentState).Methods("PUT")
    
    // Agent pool management
    poolAPI := ags.router.PathPrefix("/api/v1/pools").Subrouter()
    poolAPI.HandleFunc("", ags.handleListPools).Methods("GET")
    poolAPI.HandleFunc("", ags.handleCreatePool).Methods("POST")
    poolAPI.HandleFunc("/{poolId}", ags.handleGetPool).Methods("GET")
    poolAPI.HandleFunc("/{poolId}/scale", ags.handleScalePool).Methods("POST")
    
    // Configuration management
    configAPI := ags.router.PathPrefix("/api/v1/config").Subrouter()
    configAPI.HandleFunc("/agents/{agentId}", ags.handleUpdateAgentConfig).Methods("PUT")
    configAPI.HandleFunc("/pools/{poolId}", ags.handleUpdatePoolConfig).Methods("PUT")
    
    // Monitoring and health endpoints
    healthAPI := ags.router.PathPrefix("/api/v1/health").Subrouter()
    healthAPI.HandleFunc("/agents", ags.handleAgentHealthStatus).Methods("GET")
    healthAPI.HandleFunc("/system", ags.handleSystemHealth).Methods("GET")
    
    // Metrics and observability
    metricsAPI := ags.router.PathPrefix("/api/v1/metrics").Subrouter()
    metricsAPI.HandleFunc("/agents", ags.handleAgentMetrics).Methods("GET")
    metricsAPI.HandleFunc("/pools", ags.handlePoolMetrics).Methods("GET")
    metricsAPI.HandleFunc("/system", ags.handleSystemMetrics).Methods("GET")
}

func (ags *APIGatewayService) handleDeployAgent(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    var req DeployAgentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        ags.writeErrorResponse(w, http.StatusBadRequest, 
            "invalid request body", err)
        return
    }
    
    // Check permissions
    user := ags.getUserFromContext(ctx)
    if err := ags.authProvider.EnforcePermissions(user, "agents", "create"); err != nil {
        ags.writeErrorResponse(w, http.StatusForbidden, 
            "insufficient permissions", err)
        return
    }
    
    // Deploy agent through orchestrator
    response, err := ags.orchestrator.DeployAgent(ctx, &req)
    if err != nil {
        ags.writeErrorResponse(w, http.StatusInternalServerError, 
            "deployment failed", err)
        return
    }
    
    // Record metrics
    ags.metricsCollector.IncrementCounter("agents_deployed_total")
    
    ags.writeJSONResponse(w, http.StatusCreated, response)
}

func (ags *APIGatewayService) handleUpdateAgentState(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)
    agentID := vars["agentId"]
    
    var stateUpdate map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&stateUpdate); err != nil {
        ags.writeErrorResponse(w, http.StatusBadRequest, 
            "invalid state data", err)
        return
    }
    
    // Check permissions
    user := ags.getUserFromContext(ctx)
    if err := ags.authProvider.EnforcePermissions(user, "agents", "update"); err != nil {
        ags.writeErrorResponse(w, http.StatusForbidden, 
            "insufficient permissions", err)
        return
    }
    
    // Update agent state through coordinator
    if err := ags.coordinator.UpdateAgentState(ctx, agentID, stateUpdate); err != nil {
        ags.writeErrorResponse(w, http.StatusInternalServerError, 
            "state update failed", err)
        return
    }
    
    ags.writeJSONResponse(w, http.StatusOK, map[string]string{
        "status": "updated",
        "agent_id": agentID,
    })
}
```
