# Observability & Performance

This document covers metrics collection, distributed tracing, logging, and performance optimization strategies for the CodeValdCortex backend.

## 4. Observability and Monitoring

### 4.1 Metrics Collection and Monitoring

#### Prometheus Integration Service
```go
type MetricsCollectionService struct {
    prometheusClient prometheus.Client
    metricsRegistry  *prometheus.Registry
    agentMetrics     *AgentMetrics
    systemMetrics    *SystemMetrics
    businessMetrics  *BusinessMetrics
    logger           *zap.Logger
}

type AgentMetrics struct {
    AgentsDeployedTotal     prometheus.Counter
    AgentsRunningGauge     prometheus.Gauge
    AgentOperationsTotal   prometheus.CounterVec
    AgentResponseTime      prometheus.HistogramVec
    AgentResourceUsage     prometheus.GaugeVec
}

func NewMetricsCollectionService() *MetricsCollectionService {
    registry := prometheus.NewRegistry()
    
    agentMetrics := &AgentMetrics{
        AgentsDeployedTotal: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "codevaldcortex_agents_deployed_total",
            Help: "Total number of agents deployed",
        }),
        AgentsRunningGauge: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "codevaldcortex_agents_running",
            Help: "Number of currently running agents",
        }),
        AgentOperationsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "codevaldcortex_agent_operations_total",
                Help: "Total number of agent operations",
            },
            []string{"agent_id", "operation_type", "status"},
        ),
        AgentResponseTime: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "codevaldcortex_agent_response_time_seconds",
                Help:    "Agent operation response time in seconds",
                Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
            },
            []string{"agent_id", "operation_type"},
        ),
        AgentResourceUsage: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "codevaldcortex_agent_resource_usage",
                Help: "Agent resource usage (CPU/Memory)",
            },
            []string{"agent_id", "resource_type"},
        ),
    }
    
    // Register metrics
    registry.MustRegister(
        agentMetrics.AgentsDeployedTotal,
        agentMetrics.AgentsRunningGauge,
        agentMetrics.AgentOperationsTotal,
        agentMetrics.AgentResponseTime,
        agentMetrics.AgentResourceUsage,
    )
    
    return &MetricsCollectionService{
        metricsRegistry: registry,
        agentMetrics:    agentMetrics,
        logger:          zap.NewNop(),
    }
}

func (mcs *MetricsCollectionService) RecordAgentDeployment(agentID string) {
    mcs.agentMetrics.AgentsDeployedTotal.Inc()
    mcs.agentMetrics.AgentsRunningGauge.Inc()
    
    mcs.logger.Debug("recorded agent deployment metric",
        zap.String("agent_id", agentID))
}

func (mcs *MetricsCollectionService) RecordAgentOperation(
    agentID, operationType, status string, duration time.Duration) {
    
    mcs.agentMetrics.AgentOperationsTotal.WithLabelValues(
        agentID, operationType, status).Inc()
    
    mcs.agentMetrics.AgentResponseTime.WithLabelValues(
        agentID, operationType).Observe(duration.Seconds())
}

func (mcs *MetricsCollectionService) UpdateAgentResourceUsage(
    agentID string, cpuUsage, memoryUsage float64) {
    
    mcs.agentMetrics.AgentResourceUsage.WithLabelValues(
        agentID, "cpu").Set(cpuUsage)
    
    mcs.agentMetrics.AgentResourceUsage.WithLabelValues(
        agentID, "memory").Set(memoryUsage)
}

func (mcs *MetricsCollectionService) GetMetricsHandler() http.Handler {
    return promhttp.HandlerFor(mcs.metricsRegistry, promhttp.HandlerOpts{})
}
```

### 4.2 Distributed Tracing and Logging

#### Jaeger Integration
```go
type TracingService struct {
    tracer    opentracing.Tracer
    closer    io.Closer
    logger    *zap.Logger
}

func NewTracingService(serviceName string) (*TracingService, error) {
    cfg := jaegerconfig.Configuration{
        ServiceName: serviceName,
        Sampler: &jaegerconfig.SamplerConfig{
            Type:  jaeger.SamplerTypeConst,
            Param: 1, // Sample all traces
        },
        Reporter: &jaegerconfig.ReporterConfig{
            LogSpans:            true,
            BufferFlushInterval: 1 * time.Second,
            LocalAgentHostPort:  "jaeger-agent:6831",
        },
    }
    
    tracer, closer, err := cfg.NewTracer()
    if err != nil {
        return nil, fmt.Errorf("failed to create tracer: %w", err)
    }
    
    opentracing.SetGlobalTracer(tracer)
    
    return &TracingService{
        tracer: tracer,
        closer: closer,
        logger: zap.NewNop(),
    }, nil
}

func (ts *TracingService) StartSpan(
    operationName string, 
    parentCtx context.Context) (opentracing.Span, context.Context) {
    
    var span opentracing.Span
    
    if parentSpan := opentracing.SpanFromContext(parentCtx); parentSpan != nil {
        span = ts.tracer.StartSpan(operationName, 
            opentracing.ChildOf(parentSpan.Context()))
    } else {
        span = ts.tracer.StartSpan(operationName)
    }
    
    ctx := opentracing.ContextWithSpan(parentCtx, span)
    return span, ctx
}

func (ts *TracingService) TraceAgentOperation(
    ctx context.Context, agentID, operation string, 
    fn func(context.Context) error) error {
    
    span, ctx := ts.StartSpan(fmt.Sprintf("agent.%s", operation), ctx)
    defer span.Finish()
    
    span.SetTag("agent.id", agentID)
    span.SetTag("operation.type", operation)
    
    if err := fn(ctx); err != nil {
        span.SetTag("error", true)
        span.LogFields(log.Error(err))
        return err
    }
    
    span.SetTag("success", true)
    return nil
}

func (ts *TracingService) Close() error {
    return ts.closer.Close()
}
```

## 5. Performance and Scalability

### 5.1 Performance Optimization Strategies

#### Connection Pooling and Resource Management
```go
type ResourceManager struct {
    dbPool          *arangodb.ConnectionPool
    kubernetesPool  *kubernetes.ClientPool
    httpClientPool  *http.ClientPool
    goroutinePool   *ants.Pool
    logger          *zap.Logger
}

func NewResourceManager(config *ResourceConfig) (*ResourceManager, error) {
    // Database connection pool
    dbPool, err := arangodb.NewConnectionPool(arangodb.PoolConfig{
        MaxConnections: config.MaxDBConnections,
        IdleTimeout:    config.IdleTimeout,
        MaxLifetime:    config.MaxConnectionLifetime,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create DB pool: %w", err)
    }
    
    // Goroutine pool for agent operations
    goroutinePool, err := ants.NewPool(config.MaxGoroutines)
    if err != nil {
        return nil, fmt.Errorf("failed to create goroutine pool: %w", err)
    }
    
    return &ResourceManager{
        dbPool:         dbPool,
        goroutinePool:  goroutinePool,
        logger:         zap.NewNop(),
    }, nil
}

func (rm *ResourceManager) ExecuteAgentOperation(
    operation func() error) error {
    
    return rm.goroutinePool.Submit(operation)
}

func (rm *ResourceManager) GetDatabaseClient() (arangodb.Client, error) {
    return rm.dbPool.Get()
}

func (rm *ResourceManager) ReleaseDatabaseClient(client arangodb.Client) {
    rm.dbPool.Put(client)
}
```

### 5.2 Auto-Scaling and Load Management

#### Kubernetes Horizontal Pod Autoscaler Integration
```go
type AutoScalingManager struct {
    kubernetesClient kubernetes.Interface
    metricsClient   metricsv1beta1.MetricsV1beta1Interface
    coordinator     *DataCoordinationService
    config          *AutoScalingConfig
    logger          *zap.Logger
}

func (asm *AutoScalingManager) MonitorAndScale(ctx context.Context) {
    ticker := time.NewTicker(asm.config.ScalingCheckInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := asm.evaluateScaling(ctx); err != nil {
                asm.logger.Error("scaling evaluation failed", zap.Error(err))
            }
        }
    }
}

func (asm *AutoScalingManager) evaluateScaling(ctx context.Context) error {
    // Get current workload metrics
    workloadMetrics, err := asm.getWorkloadMetrics(ctx)
    if err != nil {
        return fmt.Errorf("failed to get workload metrics: %w", err)
    }
    
    // Calculate scaling decisions
    scalingDecisions := asm.calculateScalingDecisions(workloadMetrics)
    
    // Execute scaling operations
    for _, decision := range scalingDecisions {
        if err := asm.executeScalingDecision(ctx, decision); err != nil {
            asm.logger.Error("scaling decision execution failed",
                zap.String("pool_id", decision.PoolID),
                zap.Error(err))
        }
    }
    
    return nil
}
```
