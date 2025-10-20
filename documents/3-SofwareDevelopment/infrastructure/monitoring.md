# CodeValdCortex - Monitoring and Observability

## Overview

The monitoring and observability stack provides comprehensive insights into CodeValdCortex's multi-agent orchestration platform. Built on Prometheus, Grafana, and Jaeger, it delivers real-time metrics, alerting, and distributed tracing for enterprise-grade operational visibility.

## 1. Prometheus Configuration

### Production Prometheus Deployment

```yaml
# Prometheus Configuration
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: pweza-core-prometheus
  namespace: pweza-core-monitoring
spec:
  serviceAccountName: prometheus
  
  replicas: 2
  retention: 30d
  retentionSize: 50GB
  
  resources:
    requests:
      memory: 2Gi
      cpu: 1000m
    limits:
      memory: 4Gi
      cpu: 2000m
  
  storage:
    volumeClaimTemplate:
      spec:
        storageClassName: fast-ssd
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 100Gi
  
  serviceMonitorSelector:
    matchLabels:
      app: pweza-core
  
  ruleSelector:
    matchLabels:
      app: pweza-core
      type: alerting-rules
  
  additionalScrapeConfigs:
    name: pweza-core-scrape-config
    key: additional-scrape-configs.yaml
  
  alerting:
    alertmanagers:
    - namespace: pweza-core-monitoring
      name: pweza-core-alertmanager
      port: web
```

### Service Monitor Configuration

```yaml
# Service Monitor for Agent Metrics
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: pweza-core-agents
  namespace: pweza-core-monitoring
  labels:
    app: pweza-core
spec:
  selector:
    matchLabels:
      app: pweza-core-agent
  endpoints:
  - port: metrics
    interval: 10s
    path: /metrics
    scheme: http
  namespaceSelector:
    matchNames:
    - pweza-core-agents
    - pweza-core-system
---
# Service Monitor for Core Manager
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: pweza-core-manager
  namespace: pweza-core-monitoring
  labels:
    app: pweza-core
spec:
  selector:
    matchLabels:
      app: pweza-core-manager
  endpoints:
  - port: metrics
    interval: 15s
    path: /metrics
    scheme: http
  namespaceSelector:
    matchNames:
    - pweza-core-system
```

## 2. Grafana Dashboards

### Agent Management Dashboard

```json
{
  "dashboard": {
    "id": null,
    "title": "CodeValdCortex - Agent Management",
    "tags": ["pweza-core", "agents"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Agent Status Overview",
        "type": "stat",
        "targets": [
          {
            "expr": "count by (status) (pweza_agent_status)",
            "legendFormat": "{{status}}"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "palette-classic"
            },
            "mappings": [
              {
                "options": {
                  "running": {
                    "color": "green",
                    "index": 0
                  },
                  "failed": {
                    "color": "red",
                    "index": 1
                  },
                  "pending": {
                    "color": "yellow",
                    "index": 2
                  }
                },
                "type": "value"
              }
            ]
          }
        }
      },
      {
        "id": 2,
        "title": "Agent CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(pweza_agent_cpu_usage_seconds_total[5m]) * 100",
            "legendFormat": "{{agent_id}}"
          }
        ],
        "yAxes": [
          {
            "label": "CPU Usage (%)",
            "max": 100,
            "min": 0
          }
        ]
      },
      {
        "id": 3,
        "title": "Agent Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "pweza_agent_memory_usage_bytes / 1024 / 1024",
            "legendFormat": "{{agent_id}}"
          }
        ],
        "yAxes": [
          {
            "label": "Memory Usage (MB)",
            "min": 0
          }
        ]
      },
      {
        "id": 4,
        "title": "Message Throughput",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(pweza_agent_messages_total[5m])",
            "legendFormat": "Messages/sec"
          }
        ]
      },
      {
        "id": 5,
        "title": "Workflow Executions",
        "type": "table",
        "targets": [
          {
            "expr": "pweza_workflow_execution_status",
            "format": "table",
            "instant": true
          }
        ],
        "transformations": [
          {
            "id": "organize",
            "options": {
              "excludeByName": {
                "__name__": true,
                "Time": true
              },
              "indexByName": {
                "workflow_id": 0,
                "status": 1,
                "duration": 2,
                "tasks_total": 3,
                "tasks_completed": 4
              }
            }
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "10s"
  }
}
```

### System Performance Dashboard

```json
{
  "dashboard": {
    "id": null,
    "title": "CodeValdCortex - System Performance",
    "tags": ["pweza-core", "performance"],
    "panels": [
      {
        "id": 1,
        "title": "API Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(pweza_api_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "id": 2,
        "title": "API Response Times",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(pweza_api_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          },
          {
            "expr": "histogram_quantile(0.95, rate(pweza_api_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.99, rate(pweza_api_request_duration_seconds_bucket[5m]))",
            "legendFormat": "99th percentile"
          }
        ]
      },
      {
        "id": 3,
        "title": "Database Operations",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(pweza_database_operations_total[5m])",
            "legendFormat": "{{operation}}"
          }
        ]
      },
      {
        "id": 4,
        "title": "Error Rates",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(pweza_api_requests_total{status=~\"5..\"}[5m]) / rate(pweza_api_requests_total[5m]) * 100",
            "legendFormat": "API Errors (%)"
          },
          {
            "expr": "rate(pweza_agent_errors_total[5m])",
            "legendFormat": "Agent Errors/sec"
          }
        ]
      }
    ]
  }
}
```

## 3. Alerting Rules

### Critical System Alerts

```yaml
# Prometheus Alerting Rules
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: pweza-core-alerts
  namespace: pweza-core-monitoring
  labels:
    app: pweza-core
    type: alerting-rules
spec:
  groups:
  - name: pweza-core.agent-health
    interval: 30s
    rules:
    - alert: AgentDown
      expr: up{job="pweza-core-agents"} == 0
      for: 1m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Agent {{ $labels.instance }} is down"
        description: "Agent {{ $labels.instance }} has been down for more than 1 minute"
        runbook_url: "https://docs.pweza-core.com/runbooks/agent-down"
    
    - alert: HighAgentCPUUsage
      expr: rate(pweza_agent_cpu_usage_seconds_total[5m]) * 100 > 90
      for: 5m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "High CPU usage on agent {{ $labels.agent_id }}"
        description: "Agent {{ $labels.agent_id }} has been using >90% CPU for 5 minutes"
    
    - alert: HighAgentMemoryUsage
      expr: pweza_agent_memory_usage_bytes / pweza_agent_memory_limit_bytes > 0.9
      for: 3m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "High memory usage on agent {{ $labels.agent_id }}"
        description: "Agent {{ $labels.agent_id }} is using >90% of allocated memory"
    
    - alert: AgentMessageQueueFull
      expr: pweza_agent_message_queue_size > 1000
      for: 2m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Agent message queue full on {{ $labels.agent_id }}"
        description: "Agent {{ $labels.agent_id }} message queue has >1000 pending messages"
  
  - name: pweza-core.workflow-health
    interval: 60s
    rules:
    - alert: WorkflowExecutionFailed
      expr: increase(pweza_workflow_execution_failures_total[5m]) > 0
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Workflow execution failures detected"
        description: "{{ $value }} workflow executions have failed in the last 5 minutes"
    
    - alert: LongRunningWorkflow
      expr: pweza_workflow_execution_duration_seconds > 3600
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "Long-running workflow detected"
        description: "Workflow {{ $labels.workflow_id }} has been running for over 1 hour"
  
  - name: pweza-core.system-health
    interval: 30s
    rules:
    - alert: DatabaseConnectionFailure
      expr: pweza_database_connection_errors_total > 0
      for: 1m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Database connection failures"
        description: "ArangoDB connection failures detected"
    
    - alert: HighAPILatency
      expr: histogram_quantile(0.95, rate(pweza_api_request_duration_seconds_bucket[5m])) > 1
      for: 3m
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "High API latency"
        description: "95th percentile API latency is >1 second"
```

## 4. Distributed Tracing with Jaeger

### Jaeger Deployment

```yaml
# Jaeger All-in-One Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jaeger
  namespace: pweza-core-monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: jaeger
  template:
    metadata:
      labels:
        app: jaeger
    spec:
      containers:
      - name: jaeger
        image: jaegertracing/all-in-one:1.50
        ports:
        - containerPort: 16686
          name: query
        - containerPort: 14268
          name: collector
        - containerPort: 14250
          name: grpc
        env:
        - name: COLLECTOR_OTLP_ENABLED
          value: "true"
        - name: SPAN_STORAGE_TYPE
          value: "elasticsearch"
        - name: ES_SERVER_URLS
          value: "http://elasticsearch:9200"
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
---
apiVersion: v1
kind: Service
metadata:
  name: jaeger
  namespace: pweza-core-monitoring
spec:
  selector:
    app: jaeger
  ports:
  - name: query
    port: 16686
    targetPort: 16686
  - name: collector
    port: 14268
    targetPort: 14268
  - name: grpc
    port: 14250
    targetPort: 14250
```

### OpenTelemetry Configuration

```go
// OpenTelemetry setup for distributed tracing
package tracing

import (
    "context"
    "fmt"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func InitTracing(serviceName, jaegerEndpoint string) (*trace.TracerProvider, error) {
    // Create Jaeger exporter
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
    if err != nil {
        return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
    }
    
    // Create resource with service information
    res, err := resource.New(context.Background(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
            semconv.ServiceVersionKey.String("1.0.0"),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create resource: %w", err)
    }
    
    // Create tracer provider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(res),
        trace.WithSampler(trace.AlwaysSample()),
    )
    
    // Set global tracer provider
    otel.SetTracerProvider(tp)
    
    return tp, nil
}
```

## 5. Log Aggregation

### Fluentd Configuration

```yaml
# Fluentd DaemonSet for log collection
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
  namespace: pweza-core-monitoring
spec:
  selector:
    matchLabels:
      app: fluentd
  template:
    metadata:
      labels:
        app: fluentd
    spec:
      serviceAccountName: fluentd
      containers:
      - name: fluentd
        image: fluent/fluentd-kubernetes-daemonset:v1.16-debian-elasticsearch7-1
        env:
        - name: FLUENT_ELASTICSEARCH_HOST
          value: "elasticsearch"
        - name: FLUENT_ELASTICSEARCH_PORT
          value: "9200"
        - name: FLUENT_ELASTICSEARCH_SCHEME
          value: "http"
        - name: FLUENT_ELASTICSEARCH_USER
          valueFrom:
            secretKeyRef:
              name: elasticsearch-credentials
              key: username
        - name: FLUENT_ELASTICSEARCH_PASSWORD
          valueFrom:
            secretKeyRef:
              name: elasticsearch-credentials
              key: password
        resources:
          requests:
            cpu: 100m
            memory: 200Mi
          limits:
            cpu: 500m
            memory: 500Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
        - name: fluentd-config
          mountPath: /fluentd/etc
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: fluentd-config
        configMap:
          name: fluentd-config
```

## 6. Performance Metrics

### Go Application Metrics

```go
// Prometheus metrics for CodeValdCortex components
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Agent metrics
    AgentStatus = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "pweza_agent_status",
            Help: "Current status of agents (0=stopped, 1=running, 2=failed)",
        },
        []string{"agent_id", "agent_type", "status"},
    )
    
    AgentCPUUsage = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "pweza_agent_cpu_usage_seconds_total",
            Help: "Total CPU time consumed by agent",
        },
        []string{"agent_id"},
    )
    
    AgentMemoryUsage = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "pweza_agent_memory_usage_bytes",
            Help: "Current memory usage by agent",
        },
        []string{"agent_id"},
    )
    
    // Workflow metrics
    WorkflowExecutions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "pweza_workflow_executions_total",
            Help: "Total number of workflow executions",
        },
        []string{"workflow_id", "status"},
    )
    
    WorkflowDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "pweza_workflow_execution_duration_seconds",
            Help:    "Duration of workflow executions",
            Buckets: prometheus.DefBuckets,
        },
        []string{"workflow_id"},
    )
    
    // API metrics
    APIRequests = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "pweza_api_requests_total",
            Help: "Total number of API requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    APIRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "pweza_api_request_duration_seconds",
            Help:    "Duration of API requests",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint"},
    )
)
```

This monitoring and observability infrastructure provides comprehensive visibility into CodeValdCortex's multi-agent orchestration platform with enterprise-grade alerting and performance tracking.