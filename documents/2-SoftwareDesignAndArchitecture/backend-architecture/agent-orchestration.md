# Agent Orchestration Services

This document covers the agent lifecycle management, health monitoring, communication systems, task execution, and pool management components of the CodeValdCortex backend architecture.

## 2. Agent Orchestration Services

### 2.1 Agent Lifecycle Management

#### Agent Orchestrator Service
```go
type AgentOrchestratorService struct {
    kubernetesClient kubernetes.Interface
    configManager    *ConfigurationManager
    healthMonitor    *HealthMonitor
    scaler          *AutoScaler
    coordinator     *DataCoordinator
    logger          *zap.Logger
}

func (aos *AgentOrchestratorService) DeployAgent(
    ctx context.Context, req *DeployAgentRequest) (*DeployAgentResponse, error) {
    
    // Validate deployment request
    if err := aos.validateDeploymentRequest(req); err != nil {
        return nil, status.Errorf(codes.InvalidArgument, 
            "invalid deployment request: %v", err)
    }
    
    // Generate agent configuration
    agentConfig, err := aos.configManager.GenerateAgentConfig(req)
    if err != nil {
        return nil, status.Errorf(codes.Internal, 
            "failed to generate agent config: %v", err)
    }
    
    // Deploy to Kubernetes cluster
    deployment, err := aos.deployToKubernetes(ctx, agentConfig)
    if err != nil {
        return nil, status.Errorf(codes.Internal, 
            "kubernetes deployment failed: %v", err)
    }
    
    // Register agent with coordination service
    agent := &Agent{
        ID:           agentConfig.ID,
        Config:       agentConfig,
        Deployment:   deployment,
        Status:       AgentStatusDeploying,
        CreatedAt:    time.Now(),
    }
    
    if err := aos.coordinator.RegisterAgent(ctx, agent); err != nil {
        // Rollback Kubernetes deployment
        aos.rollbackDeployment(ctx, deployment)
        return nil, status.Errorf(codes.Internal, 
            "agent registration failed: %v", err)
    }
    
    // Start health monitoring
    aos.healthMonitor.StartMonitoring(agent)
    
    aos.logger.Info("agent deployed successfully",
        zap.String("agent_id", agent.ID),
        zap.String("namespace", agentConfig.Namespace))
    
    return &DeployAgentResponse{
        AgentId: agent.ID,
        Status:  agent.Status,
    }, nil
}

func (aos *AgentOrchestratorService) ScaleAgentPool(
    ctx context.Context, req *ScalePoolRequest) (*ScalePoolResponse, error) {
    
    // Get current pool status
    pool, err := aos.coordinator.GetAgentPool(ctx, req.PoolId)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, 
            "agent pool not found: %v", err)
    }
    
    // Calculate scaling strategy
    scalingPlan, err := aos.scaler.CalculateScalingPlan(
        pool, req.TargetCount, req.ScalingPolicy)
    if err != nil {
        return nil, status.Errorf(codes.Internal, 
            "scaling calculation failed: %v", err)
    }
    
    // Execute scaling operations
    results := make([]*ScalingResult, 0, len(scalingPlan.Operations))
    for _, operation := range scalingPlan.Operations {
        result, err := aos.executeScalingOperation(ctx, operation)
        if err != nil {
            aos.logger.Error("scaling operation failed",
                zap.String("operation", operation.Type),
                zap.Error(err))
        }
        results = append(results, result)
    }
    
    return &ScalePoolResponse{
        PoolId:    req.PoolId,
        Results:   results,
        NewCount:  scalingPlan.FinalCount,
    }, nil
}

func (cts *ConfigurationTemplateService) UpdateTemplate(
    ctx context.Context, templateID string, updates *TemplateUpdate) error {
    
    // Get existing template
    existingTemplate, err := cts.GetTemplate(ctx, templateID)
    if err != nil {
        return fmt.Errorf("failed to get existing template: %w", err)
    }
    
    // Create new version
    newVersion := cts.versionManager.IncrementVersion(existingTemplate.Version)
    
    // Apply updates
    updatedTemplate := &AgentTemplate{
        ID:          existingTemplate.ID,
        Name:        updates.Name,
        Description: updates.Description,
        Version:     newVersion,
        Configuration: updates.Configuration,
        ResourceRequirements: updates.ResourceRequirements,
        EnvironmentVariables: updates.EnvironmentVariables,
        SecuritySettings: updates.SecuritySettings,
        CreatedAt:   existingTemplate.CreatedAt,
        UpdatedAt:   time.Now(),
        CreatedBy:   existingTemplate.CreatedBy,
        Tags:        updates.Tags,
    }
    
    // Validate updated template
    if err := cts.validator.ValidateTemplate(updatedTemplate); err != nil {
        return fmt.Errorf("updated template validation failed: %w", err)
    }
    
    // Store updated template
    collection := cts.db.Collection("agent_templates")
    updateDoc := map[string]interface{}{
        "name":         updatedTemplate.Name,
        "description":  updatedTemplate.Description,
        "version":      updatedTemplate.Version,
        "configuration": updatedTemplate.Configuration,
        "resource_requirements": updatedTemplate.ResourceRequirements,
        "environment_variables": updatedTemplate.EnvironmentVariables,
        "security_settings": updatedTemplate.SecuritySettings,
        "updated_at":   updatedTemplate.UpdatedAt,
        "tags":         updatedTemplate.Tags,
    }
    
    _, err = collection.UpdateDocument(ctx, templateID, updateDoc)
    if err != nil {
        return fmt.Errorf("failed to update template: %w", err)
    }
    
    cts.logger.Info("agent template updated",
        zap.String("template_id", templateID),
        zap.String("new_version", newVersion))
    
    return nil
}

func (cts *ConfigurationTemplateService) DeleteTemplate(
    ctx context.Context, templateID string) error {
    
    // Check if template is in use
    inUse, err := cts.isTemplateInUse(ctx, templateID)
    if err != nil {
        return fmt.Errorf("failed to check template usage: %w", err)
    }
    
    if inUse {
        return ErrTemplateInUse
    }
    
    // Delete template
    collection := cts.db.Collection("agent_templates")
    _, err = collection.RemoveDocument(ctx, templateID)
    if err != nil {
        return fmt.Errorf("failed to delete template: %w", err)
    }
    
    cts.logger.Info("agent template deleted",
        zap.String("template_id", templateID))
    
    return nil
}

func (cts *ConfigurationTemplateService) isTemplateInUse(
    ctx context.Context, templateID string) (bool, error) {
    
    // Check if any agents are using this template
    query := `
        FOR agent IN agents
        FILTER agent.template_id == @templateId
        LIMIT 1
        RETURN agent
    `
    
    cursor, err := cts.db.Query(ctx, query, map[string]interface{}{
        "templateId": templateID,
    })
    if err != nil {
        return false, err
    }
    defer cursor.Close()
    
    return cursor.HasMore(), nil
}
```

### 2.2 Health Monitoring and Auto-Scaling

#### Health Monitoring Service
```go
type HealthMonitorService struct {
    kubernetesClient kubernetes.Interface
    metricsClient   metricsv1beta1.MetricsV1beta1Interface
    coordinator     *DataCoordinator
    alertManager    *AlertManager
    checkInterval   time.Duration
    logger          *zap.Logger
}

func (hms *HealthMonitorService) StartMonitoring(agent *Agent) {
    go hms.monitorAgent(agent)
}

func (hms *HealthMonitorService) monitorAgent(agent *Agent) {
    ticker := time.NewTicker(hms.checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            health, err := hms.checkAgentHealth(agent)
            if err != nil {
                hms.logger.Error("health check failed",
                    zap.String("agent_id", agent.ID),
                    zap.Error(err))
                continue
            }
            
            if health.Status != HealthStatusHealthy {
                hms.handleUnhealthyAgent(agent, health)
            }
            
            // Update agent health in coordination service
            if err := hms.coordinator.UpdateAgentHealth(agent.ID, health); err != nil {
                hms.logger.Error("failed to update agent health",
                    zap.String("agent_id", agent.ID),
                    zap.Error(err))
            }
        }
    }
}

func (hms *HealthMonitorService) checkAgentHealth(agent *Agent) (*HealthStatus, error) {
    // Check Kubernetes pod status
    pod, err := hms.kubernetesClient.CoreV1().
        Pods(agent.Config.Namespace).
        Get(context.TODO(), agent.PodName, metav1.GetOptions{})
    
    if err != nil {
        return &HealthStatus{
            Status:    HealthStatusUnhealthy,
            Message:   fmt.Sprintf("pod not found: %v", err),
            Timestamp: time.Now(),
        }, nil
    }
    
    // Check pod readiness
    if pod.Status.Phase != corev1.PodRunning {
        return &HealthStatus{
            Status:    HealthStatusUnhealthy,
            Message:   fmt.Sprintf("pod not running: %s", pod.Status.Phase),
            Timestamp: time.Now(),
        }, nil
    }
    
    // Check resource usage
    metrics, err := hms.metricsClient.PodMetricses(agent.Config.Namespace).
        Get(context.TODO(), agent.PodName, metav1.GetOptions{})
    
    if err == nil {
        cpuUsage := metrics.Containers[0].Usage[corev1.ResourceCPU]
        memoryUsage := metrics.Containers[0].Usage[corev1.ResourceMemory]
        
        if hms.isResourceUsageExcessive(cpuUsage, memoryUsage, agent.Config) {
            return &HealthStatus{
                Status:    HealthStatusDegraded,
                Message:   "high resource usage detected",
                Timestamp: time.Now(),
                Metrics: map[string]interface{}{
                    "cpu_usage":    cpuUsage.String(),
                    "memory_usage": memoryUsage.String(),
                },
            }, nil
        }
    }
    
    return &HealthStatus{
        Status:    HealthStatusHealthy,
        Message:   "agent healthy",
        Timestamp: time.Now(),
    }, nil
}
```

### 2.3 Agent Communication System

#### Database-Driven Messaging Architecture

CodeValdCortex implements a database-driven agent communication system using ArangoDB as the central coordination layer. This approach provides persistent, auditable, and scalable inter-agent communication with support for both direct messaging and publish/subscribe patterns.

**Communication Patterns**:
- **Direct Messaging**: Point-to-point message delivery between specific agents
- **Publish/Subscribe**: Broadcast events with subscription-based filtering
- **Event Sourcing**: Complete audit trail of all agent communications
- **Polling-Based Delivery**: Agents poll for messages at configurable intervals

**Key Benefits**:
- **Persistence**: All messages stored in database for reliability and audit
- **Scalability**: Database handles message routing and delivery coordination
- **Flexibility**: Support for multiple communication patterns
- **Observability**: Complete visibility into message flows and delivery status

#### Message Service Architecture

```go
type AgentCommunicationService struct {
    messageService *MessageService
    pubsubService  *PubSubService
    repository     *CommunicationRepository
    poller         *MessagePoller
    logger         *zap.Logger
}

// Direct messaging: Agent A → Agent B
func (acs *AgentCommunicationService) SendMessage(
    ctx context.Context, 
    fromAgentID string,
    toAgentID string,
    messageType MessageType,
    payload map[string]interface{},
    options *MessageOptions) (string, error) {
    
    message := &Message{
        FromAgentID: fromAgentID,
        ToAgentID:   toAgentID,
        MessageType: messageType,
        Payload:     payload,
        Priority:    options.Priority,
        TTL:         options.TTL,
    }
    
    // Validate message
    if err := acs.validateMessage(message); err != nil {
        return "", fmt.Errorf("invalid message: %w", err)
    }
    
    // Store message in ArangoDB agent_messages collection
    messageID, err := acs.messageService.SendMessage(ctx, message)
    if err != nil {
        return "", fmt.Errorf("failed to send message: %w", err)
    }
    
    acs.logger.Info("message sent",
        zap.String("from", fromAgentID),
        zap.String("to", toAgentID),
        zap.String("message_id", messageID))
    
    return messageID, nil
}

// Publish/Subscribe: Agent publishes event
func (acs *AgentCommunicationService) Publish(
    ctx context.Context,
    publisherID string,
    eventName string,
    payload map[string]interface{},
    options *PublicationOptions) (string, error) {
    
    publication := &Publication{
        PublisherAgentID: publisherID,
        EventName:        eventName,
        PublicationType:  options.Type,
        Payload:          payload,
        TTLSeconds:       options.TTLSeconds,
    }
    
    // Store publication in ArangoDB agent_publications collection
    pubID, err := acs.pubsubService.Publish(ctx, publication)
    if err != nil {
        return "", fmt.Errorf("failed to publish: %w", err)
    }
    
    acs.logger.Info("event published",
        zap.String("publisher", publisherID),
        zap.String("event", eventName),
        zap.String("pub_id", pubID))
    
    return pubID, nil
}

// Subscribe: Agent subscribes to events
func (acs *AgentCommunicationService) Subscribe(
    ctx context.Context,
    subscriberID string,
    eventPattern string,
    filters *SubscriptionFilters) (string, error) {
    
    subscription := &Subscription{
        SubscriberAgentID: subscriberID,
        EventPattern:      eventPattern,
        PublisherAgentID:  filters.PublisherID,
        PublicationTypes:  filters.Types,
        FilterConditions:  filters.Conditions,
    }
    
    // Store subscription in ArangoDB agent_subscriptions collection
    subID, err := acs.pubsubService.Subscribe(ctx, subscription)
    if err != nil {
        return "", fmt.Errorf("failed to subscribe: %w", err)
    }
    
    acs.logger.Info("subscription created",
        zap.String("subscriber", subscriberID),
        zap.String("pattern", eventPattern),
        zap.String("sub_id", subID))
    
    return subID, nil
}
```

#### ArangoDB Collections for Communication

**agent_messages Collection**:
```javascript
{
  "_key": "msg-uuid",
  "from_agent_id": "agent-123",
  "to_agent_id": "agent-456",
  "message_type": "task_request",
  "payload": { /* flexible JSON */ },
  "status": "pending",
  "priority": 5,
  "created_at": "2025-10-20T10:00:00Z",
  "delivered_at": null,
  "expires_at": "2025-10-20T11:00:00Z"
}
```

**agent_publications Collection**:
```javascript
{
  "_key": "pub-uuid",
  "publisher_agent_id": "agent-123",
  "publication_type": "status_change",
  "event_name": "state.changed",
  "payload": {
    "old_state": "running",
    "new_state": "paused"
  },
  "published_at": "2025-10-20T10:00:00Z",
  "ttl_seconds": 3600
}
```

**agent_subscriptions Collection**:
```javascript
{
  "_key": "sub-uuid",
  "subscriber_agent_id": "agent-456",
  "publisher_agent_id": "agent-123",
  "event_pattern": "state.*",
  "active": true,
  "created_at": "2025-10-20T09:00:00Z"
}
```

#### Message Polling Mechanism

```go
type MessagePoller struct {
    agentID   string
    interval  time.Duration
    service   *MessageService
    handler   MessageHandler
}

func (mp *MessagePoller) Start(ctx context.Context) {
    ticker := time.NewTicker(mp.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Poll for pending messages
            messages, err := mp.service.GetPendingMessages(ctx, mp.agentID, 100)
            if err != nil {
                log.Error("failed to poll messages", "error", err)
                continue
            }
            
            // Process each message
            for _, msg := range messages {
                if err := mp.handler(msg); err != nil {
                    log.Error("message handling failed", "msg_id", msg.ID, "error", err)
                    continue
                }
                
                // Mark as delivered
                mp.service.MarkDelivered(ctx, msg.ID)
            }
            
        case <-ctx.Done():
            return
        }
    }
}
```

**Polling Configuration**:
- **High Priority Agents**: 1-2 second intervals
- **Normal Priority**: 5 second intervals
- **Background Agents**: 10-30 second intervals
- **Adaptive Polling**: Adjust based on message volume

**Performance Optimizations**:
- Database indexes on `to_agent_id`, `status`, `priority`
- Batch message retrieval (up to 100 per poll)
- Message expiration and automatic cleanup
- Optional: ArangoDB change streams for push-based delivery (future enhancement)

### 2.4 Agent Task Execution System

#### Task Management Architecture

The agent task execution system provides a comprehensive framework for scheduling, executing, and monitoring tasks across the agent ecosystem. The system is built with priority-based scheduling, retry mechanisms, and persistent storage.

**Core Components**:
- **Task Scheduler**: Priority queue-based scheduling with worker pool management
- **Task Executor**: Handler registry system for pluggable task execution
- **Task Manager**: High-level orchestration combining scheduler and executor
- **Task Repository**: ArangoDB persistence layer for tasks and results
- **Built-in Handlers**: Common task types (Echo, HTTP, Delay, Error)

#### Task Scheduler Service

```go
type TaskScheduler struct {
    tasks       *TaskQueue          // Priority queue implementation
    workers     []*TaskWorker       // Worker pool
    executor    TaskExecutor        // Task execution engine
    repository  TaskRepository      // Persistence layer
    minWorkers  int                 // Minimum worker count
    maxWorkers  int                 // Maximum worker count
    metrics     *TaskMetrics        // Performance metrics
    running     atomic.Bool         // Service state
    ctx         context.Context
    cancel      context.CancelFunc
    wg          sync.WaitGroup
}

func (ts *TaskScheduler) Submit(task *Task) error {
    // Store task in ArangoDB for persistence
    if err := ts.repository.StoreTask(context.Background(), task); err != nil {
        return fmt.Errorf("failed to store task: %w", err)
    }
    
    // Add to priority queue
    ts.tasks.Push(task)
    
    // Scale workers if needed
    ts.scaleWorkers()
    
    return nil
}
```

#### Task Executor and Handler Registry

```go
type TaskExecutor struct {
    handlers   map[string]TaskHandler  // Registered task handlers
    repository TaskRepository          // For result persistence
    metrics    *TaskMetrics           // Performance tracking
    timeout    time.Duration          // Default task timeout
    mutex      sync.RWMutex
}

func (te *TaskExecutor) RegisterHandler(taskType string, handler TaskHandler) {
    te.mutex.Lock()
    defer te.mutex.Unlock()
    te.handlers[taskType] = handler
}

func (te *TaskExecutor) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
    // Get handler for task type
    handler, exists := te.getHandler(task.Type)
    if !exists {
        return nil, fmt.Errorf("no handler registered for task type: %s", task.Type)
    }
    
    // Create execution context with timeout
    execCtx, cancel := context.WithTimeout(ctx, te.timeout)
    defer cancel()
    
    // Execute task with handler
    result, err := handler.Execute(execCtx, task)
    if err != nil {
        result = &TaskResult{
            TaskID:      task.ID,
            AgentID:     task.AgentID,
            Status:      TaskStatusFailed,
            Error:       err.Error(),
            CompletedAt: time.Now(),
        }
    }
    
    // Store result in ArangoDB
    if storeErr := te.repository.StoreTaskResult(ctx, result); storeErr != nil {
        log.Error("failed to store task result", "error", storeErr)
    }
    
    return result, err
}
```

#### Built-in Task Handlers

**Echo Handler** - Simple task for testing and validation:
```go
type EchoHandler struct{}

func (h *EchoHandler) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
    message, ok := task.Parameters["message"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid 'message' parameter")
    }
    
    return &TaskResult{
        TaskID:      task.ID,
        AgentID:     task.AgentID,
        Status:      TaskStatusCompleted,
        Result:      map[string]interface{}{"echo": message},
        CompletedAt: time.Now(),
    }, nil
}
```

**HTTP Request Handler** - External API interactions:
```go
type HTTPHandler struct {
    client *http.Client
}

func (h *HTTPHandler) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
    url, ok := task.Parameters["url"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid 'url' parameter")
    }
    
    method := "GET"
    if m, exists := task.Parameters["method"].(string); exists {
        method = m
    }
    
    req, err := http.NewRequestWithContext(ctx, method, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    resp, err := h.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    return &TaskResult{
        TaskID:      task.ID,
        AgentID:     task.AgentID,
        Status:      TaskStatusCompleted,
        Result: map[string]interface{}{
            "status_code": resp.StatusCode,
            "headers":     resp.Header,
        },
        CompletedAt: time.Now(),
    }, nil
}
```

#### ArangoDB Collections for Task Execution

**agent_tasks Collection**:
```javascript
{
  "_key": "uuid-v4-string",          // Google UUID v4 format
  "agent_id": "agent-uuid-123",
  "type": "http_request",
  "parameters": {
    "url": "https://api.example.com",
    "method": "GET"
  },
  "priority": 5,
  "status": "pending",
  "retry_count": 0,
  "max_retries": 3,
  "timeout_seconds": 30,
  "created_at": "2025-10-20T10:00:00Z",
  "scheduled_at": "2025-10-20T10:00:00Z",
  "started_at": null,
  "completed_at": null
}
```

**agent_task_results Collection**:
```javascript
{
  "_key": "result-uuid-v4",          // Google UUID v4 format
  "task_id": "uuid-v4-string",
  "agent_id": "agent-uuid-123",
  "status": "completed",
  "result": {
    "status_code": 200,
    "response_time_ms": 150
  },
  "error": null,
  "execution_time_ms": 145,
  "completed_at": "2025-10-20T10:00:15Z"
}
```

**agent_task_metrics Collection**:
```javascript
{
  "_key": "metrics-uuid-v4",         // Google UUID v4 format
  "task_type": "http_request",
  "agent_id": "agent-uuid-123",
  "total_count": 100,
  "success_count": 95,
  "failure_count": 5,
  "avg_execution_time_ms": 200,
  "last_updated": "2025-10-20T10:00:00Z"
}
```

**UUID Requirements**:
All task-related IDs (`_key` fields) must use Google UUID v4 format for uniqueness and consistency across the distributed system. This ensures proper task tracking, prevents ID collisions, and maintains referential integrity between collections.

**Performance Optimizations**:
- Database indexes on `agent_id`, `status`, `priority`, `scheduled_at`
- Compound indexes for efficient task querying and filtering
- Automatic cleanup of completed tasks older than configurable retention period
- Metrics aggregation for performance monitoring and optimization

### 2.5 Agent Pool Management System

#### Pool Management Architecture
The Agent Pool Management System provides sophisticated grouping, load balancing, and resource allocation capabilities for organizing and managing groups of agents efficiently. This system enables horizontal scaling, fault tolerance, and optimized resource utilization across agent clusters.

**Core Components**:
- **AgentPool**: Core pool structure managing agent membership and configuration
- **LoadBalancer**: Pluggable load balancing strategies for request distribution  
- **ResourceManager**: Resource allocation and monitoring for pool members
- **PoolManager**: Orchestration and lifecycle management for multiple pools
- **PoolRepository**: ArangoDB-based persistence for pool configurations and state

#### Agent Pool Structure
```go
type AgentPool struct {
    ID         string
    Config     PoolConfig
    Status     PoolStatus
    Members    map[string]*AgentPoolMember
    LoadBalancer LoadBalancer
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type PoolConfig struct {
    Name                  string
    Description           string
    LoadBalancingStrategy LoadBalancingStrategy
    MinAgents            int
    MaxAgents            int
    HealthCheckInterval  time.Duration
    ResourceLimits       ResourceLimits
    AutoScaling          AutoScalingConfig
}

type AgentPoolMember struct {
    Agent             *agent.Agent
    Weight            int
    JoinedAt          time.Time
    ActiveConnections int
    LastHealthCheck   time.Time
    Healthy           bool
}
```

#### Load Balancing Strategies
The system implements multiple load balancing algorithms to optimize request distribution:

**Round Robin Load Balancing**:
- Sequential distribution across healthy agents
- Fair distribution regardless of agent capacity
- Simple and predictable request routing

**Least Connections Load Balancing**:
- Routes requests to agents with fewest active connections
- Optimizes for balanced workload distribution
- Ideal for long-running or variable-duration tasks

**Weighted Load Balancing**:
- Agent weights determine proportional request distribution
- Supports heterogeneous agent capabilities
- Configurable weights based on agent performance metrics

**Random Load Balancing**:
- Probabilistic distribution across healthy agents
- Simple implementation with good distribution properties
- Suitable for stateless request handling

#### Resource Management
```go
type ResourceLimits struct {
    TotalCPU           int    // millicores across all agents
    TotalMemory        int    // megabytes across all agents
    MaxConcurrentTasks int    // across all agents
}

type ResourceUtilization struct {
    CPUUsage    float64  // percentage (0-100)
    MemoryUsage float64  // percentage (0-100)
    TaskLoad    float64  // percentage (0-100)
}
```

**Resource Allocation Features**:
- Aggregate resource limit enforcement across pool members
- Real-time resource utilization monitoring and reporting
- Automatic resource allocation balancing between agents
- Integration with auto-scaling for dynamic resource management

#### Pool Management Operations
**Pool Lifecycle Management**:
```bash
# Create new agent pool
POST /api/v1/pools
{
    "name": "web-processing-pool",
    "description": "Pool for web content processing agents",
    "loadBalancingStrategy": "least_connection", 
    "minAgents": 2,
    "maxAgents": 10,
    "healthCheckInterval": "30s",
    "resourceLimits": {
        "totalCPU": 4000,
        "totalMemory": 8192,
        "maxConcurrentTasks": 100
    }
}

# Add agent to pool
POST /api/v1/pools/{pool_id}/agents
{
    "agentId": "uuid-agent-id",
    "weight": 1
}

# Remove agent from pool
DELETE /api/v1/pools/{pool_id}/agents/{agent_id}

# Get pool metrics
GET /api/v1/pools/{pool_id}/metrics
```

#### ArangoDB Pool Collections

**Pools Collection** (`agent_pools`):
```json
{
    "_key": "pool_uuid",
    "name": "web-processing-pool",
    "description": "Pool for web content processing agents",
    "config": {
        "loadBalancingStrategy": "least_connection",
        "minAgents": 2,
        "maxAgents": 10,
        "healthCheckInterval": "30s",
        "resourceLimits": { "totalCPU": 4000, "totalMemory": 8192 },
        "autoScaling": { "enabled": true, "scaleUpThreshold": 80.0 }
    },
    "status": "active",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Pool Memberships Collection** (`pool_memberships`):
```json
{
    "_key": "membership_uuid",
    "poolId": "pool_uuid",
    "agentId": "agent_uuid", 
    "weight": 1,
    "joinedAt": "2024-01-01T00:00:00Z",
    "healthy": true,
    "activeConnections": 3,
    "lastHealthCheck": "2024-01-01T00:05:00Z"
}
```

**Pool Metrics Collection** (`pool_metrics`):
```json
{
    "_key": "metrics_uuid",
    "poolId": "pool_uuid",
    "timestamp": "2024-01-01T00:00:00Z",
    "totalAgents": 5,
    "healthyAgents": 5,
    "totalRequests": 1000,
    "activeRequests": 25,
    "failedRequests": 5,
    "averageResponseTime": 150.5,
    "resourceUtilization": {
        "cpuUsage": 65.2,
        "memoryUsage": 72.8,
        "taskLoad": 45.0
    }
}
```

#### Auto-Scaling Configuration
```go
type AutoScalingConfig struct {
    Enabled              bool
    ScaleUpThreshold     float64   // CPU percentage to trigger scale up
    ScaleDownThreshold   float64   // CPU percentage to trigger scale down
    ScaleUpCooldown      time.Duration
    ScaleDownCooldown    time.Duration
    MaxScaleUpSteps      int       // Maximum agents to add at once
    MaxScaleDownSteps    int       // Maximum agents to remove at once
}
```

**Auto-Scaling Behavior**:
- CPU/Memory threshold monitoring for scale decisions
- Cooldown periods prevent oscillation
- Gradual scaling respects min/max agent limits
- Integration with agent lifecycle for seamless scaling

#### Pool Health Monitoring
**Health Check Strategy**:
- Periodic health checks for all pool members
- Configurable health check intervals per pool
- Automatic unhealthy agent exclusion from load balancing
- Health recovery detection and re-inclusion

**Metrics Collection**:
- Real-time pool performance metrics
- Resource utilization tracking across all agents
- Request success/failure rates and response times
- Historical metrics for trend analysis and capacity planning

#### Integration with Agent Lifecycle
**Lifecycle Event Handling**:
- Automatic pool membership updates on agent state changes
- Graceful agent removal during shutdown or failure
- Pool rebalancing on agent additions or removals
- Coordination with agent health monitoring systems

**Pool-Aware Agent Management**:
- Pool membership consideration in agent placement decisions
- Resource allocation coordination between pools
- Conflict resolution for agents belonging to multiple pools
- Pool-scoped agent configuration and policy enforcement

#### Performance Considerations
**Load Balancing Performance**:
- O(1) agent selection for round-robin and random strategies
- O(n) agent selection for least-connections (where n = healthy agents)
- Cached health status for fast exclusion of unhealthy agents
- Lock-free read operations for high-throughput scenarios

**Resource Monitoring Efficiency**:
- Batched metrics collection to reduce overhead
- Configurable monitoring intervals based on pool requirements
- Efficient aggregation for pool-wide resource calculations
- Historical metrics retention with configurable cleanup policies

**Scalability Optimizations**:
- Database indexes on `pool_id`, `agent_id`, `status`, `timestamp`
- Compound indexes for efficient pool membership queries
- Connection pooling for database operations
- Caching of frequently accessed pool configurations
