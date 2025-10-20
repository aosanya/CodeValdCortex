# CodeValdCortex - ArangoDB Database Infrastructure

## Overview

ArangoDB serves as CodeValdCortex's multi-model database, providing document, graph, and key-value storage capabilities for agent state management, workflow coordination, and real-time analytics. The database is configured for high availability, automatic scaling, and enterprise security.

## 1. ArangoDB Cluster Configuration

### Production Cluster Setup

```yaml
# ArangoDB Cluster Configuration
apiVersion: database.arangodb.com/v1
kind: ArangoDeployment
metadata:
  name: pweza-core-arango
  namespace: pweza-core-data
spec:
  mode: Cluster
  image: arangodb/arangodb:3.11.5
  
  auth:
    jwtSecretName: pweza-arango-jwt
  
  tls:
    caSecretName: pweza-arango-ca
    certSecretName: pweza-arango-cert
  
  agents:
    count: 3
    resources:
      requests:
        cpu: 250m
        memory: 512Mi
      limits:
        cpu: 500m
        memory: 1Gi
    storageClassName: fast-ssd
    storage: 20Gi
  
  dbservers:
    count: 3
    resources:
      requests:
        cpu: 1000m
        memory: 2Gi
      limits:
        cpu: 2000m
        memory: 4Gi
    storageClassName: fast-ssd
    storage: 100Gi
  
  coordinators:
    count: 3
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        cpu: 1000m
        memory: 2Gi
  
  monitoring:
    enabled: true
  
  backup:
    enabled: true
    schedule: "0 2 * * *"
    retention: "30d"
    
  upgrade:
    autoUpgrade: false
    upgradeStrategy: "recreate"
```

### Backup and Recovery Configuration

```yaml
# ArangoDB Backup Configuration
apiVersion: backup.arangodb.com/v1
kind: ArangoBackup
metadata:
  name: pweza-core-backup
  namespace: pweza-core-data
spec:
  deployment:
    name: pweza-core-arango
  
  # Backup schedule (daily at 2 AM)
  schedule: "0 2 * * *"
  
  # Retention policy
  retention:
    full: "30d"
    incremental: "7d"
  
  # Storage configuration
  storage:
    type: s3
    s3:
      bucket: pweza-core-backups
      region: us-east-1
      endpoint: s3.amazonaws.com
      credentialsSecret: aws-s3-credentials
      encryption:
        enabled: true
        kmsKeyId: alias/pweza-core-backup-key
  
  # Backup verification
  verification:
    enabled: true
    schedule: "0 4 * * 0" # Weekly verification on Sundays
```

## 2. Database Schema Design

### Agent Data Model

```javascript
// Agent Collection Schema
{
  "_key": "agent-12345",
  "_id": "agents/agent-12345",
  "name": "processing-agent-001",
  "type": "data-processor",
  "status": "running",
  "configuration": {
    "resourceLimits": {
      "cpu": "1000m",
      "memory": "2Gi"
    },
    "environmentVariables": {
      "LOG_LEVEL": "info",
      "BATCH_SIZE": "1000"
    },
    "securityContext": {
      "runAsUser": 1000,
      "runAsGroup": 2000
    }
  },
  "deployment": {
    "clusterId": "production-cluster",
    "namespace": "pweza-core-agents",
    "podName": "processing-agent-001-pod",
    "image": "pweza/agent:v1.0.0"
  },
  "metrics": {
    "cpuUsage": 0.65,
    "memoryUsage": 0.72,
    "throughput": 1250.5,
    "errorRate": 0.001,
    "lastUpdated": "2025-10-20T14:30:00Z"
  },
  "state": {
    "version": 42,
    "lastModified": "2025-10-20T14:30:00Z",
    "data": {
      "processingQueue": 150,
      "activeConnections": 8,
      "customState": {}
    }
  },
  "createdAt": "2025-10-20T10:00:00Z",
  "updatedAt": "2025-10-20T14:30:00Z"
}
```

### Workflow Execution Model

```javascript
// Workflow Executions Collection Schema
{
  "_key": "exec-67890",
  "_id": "workflow_executions/exec-67890",
  "workflowId": "workflow-data-pipeline",
  "status": "running",
  "startTime": "2025-10-20T14:00:00Z",
  "endTime": null,
  "duration": null,
  "context": {
    "inputData": "/data/batch-2025-10-20",
    "outputLocation": "/results/batch-2025-10-20",
    "batchSize": 10000
  },
  "tasks": [
    {
      "id": "data-ingestion",
      "status": "completed",
      "agentId": "agent-12345",
      "startTime": "2025-10-20T14:00:00Z",
      "endTime": "2025-10-20T14:15:00Z",
      "result": {
        "recordsProcessed": 10000,
        "errors": 0
      }
    },
    {
      "id": "data-transformation",
      "status": "running",
      "agentId": "agent-12346",
      "startTime": "2025-10-20T14:15:00Z",
      "endTime": null,
      "progress": 0.65
    }
  ],
  "metrics": {
    "totalTasks": 4,
    "completedTasks": 1,
    "failedTasks": 0,
    "totalAgents": 3
  }
}
```

### Agent Communication Graph

```javascript
// Agent Communications (Edge Collection)
{
  "_key": "comm-12345-67890",
  "_id": "agent_communications/comm-12345-67890",
  "_from": "agents/agent-12345",
  "_to": "agents/agent-67890",
  "messageType": "coordination",
  "timestamp": "2025-10-20T14:30:00Z",
  "payload": {
    "action": "task_delegation",
    "taskId": "data-transformation-001",
    "priority": "high"
  },
  "status": "delivered",
  "deliveryTime": "2025-10-20T14:30:01Z",
  "ttl": "2025-10-27T14:30:00Z"
}
```

## 3. Data Access Layer

### Go Database Client Configuration

```go
// ArangoDB Connection and Client Setup
package database

import (
    "context"
    "crypto/tls"
    "time"
    
    "github.com/arangodb/go-driver"
    "github.com/arangodb/go-driver/http"
)

type DatabaseConfig struct {
    Endpoints    []string          `json:"endpoints"`
    Database     string            `json:"database"`
    Username     string            `json:"username"`
    Password     string            `json:"password"`
    TLS          *TLSConfig        `json:"tls,omitempty"`
    Connection   *ConnectionConfig `json:"connection"`
}

type TLSConfig struct {
    Enabled            bool   `json:"enabled"`
    CertificatePath    string `json:"certificatePath"`
    PrivateKeyPath     string `json:"privateKeyPath"`
    CACertificatePath  string `json:"caCertificatePath"`
    InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

type ConnectionConfig struct {
    MaxConnections    int           `json:"maxConnections"`
    ConnectionTimeout time.Duration `json:"connectionTimeout"`
    RequestTimeout    time.Duration `json:"requestTimeout"`
    RetryAttempts     int           `json:"retryAttempts"`
    RetryDelay        time.Duration `json:"retryDelay"`
}

func NewDatabaseClient(config DatabaseConfig) (driver.Database, error) {
    // Configure HTTP transport with TLS
    transport := &http.Transport{
        MaxIdleConns:        config.Connection.MaxConnections,
        MaxIdleConnsPerHost: config.Connection.MaxConnections / len(config.Endpoints),
        IdleConnTimeout:     30 * time.Second,
    }
    
    if config.TLS != nil && config.TLS.Enabled {
        tlsConfig := &tls.Config{
            InsecureSkipVerify: config.TLS.InsecureSkipVerify,
        }
        
        if config.TLS.CertificatePath != "" && config.TLS.PrivateKeyPath != "" {
            cert, err := tls.LoadX509KeyPair(config.TLS.CertificatePath, config.TLS.PrivateKeyPath)
            if err != nil {
                return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
            }
            tlsConfig.Certificates = []tls.Certificate{cert}
        }
        
        transport.TLSClientConfig = tlsConfig
    }
    
    // Create HTTP connection
    conn, err := http.NewConnection(http.ConnectionConfig{
        Endpoints: config.Endpoints,
        Transport: transport,
        Timeout:   config.Connection.ConnectionTimeout,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create ArangoDB connection: %w", err)
    }
    
    // Create authenticated client
    client, err := driver.NewClient(driver.ClientConfig{
        Connection:     conn,
        Authentication: driver.BasicAuthentication(config.Username, config.Password),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create ArangoDB client: %w", err)
    }
    
    // Access or create database
    ctx, cancel := context.WithTimeout(context.Background(), config.Connection.RequestTimeout)
    defer cancel()
    
    db, err := client.Database(ctx, config.Database)
    if driver.IsNotFound(err) {
        // Create database if it doesn't exist
        db, err = client.CreateDatabase(ctx, config.Database, nil)
        if err != nil {
            return nil, fmt.Errorf("failed to create database: %w", err)
        }
    } else if err != nil {
        return nil, fmt.Errorf("failed to access database: %w", err)
    }
    
    // Initialize collections and indexes
    if err := initializeCollections(ctx, db); err != nil {
        return nil, fmt.Errorf("failed to initialize collections: %w", err)
    }
    
    return db, nil
}
```

### Collection Initialization

```go
func initializeCollections(ctx context.Context, db driver.Database) error {
    collections := []struct {
        name    string
        options *driver.CreateCollectionOptions
        indexes []driver.EnsureIndexOptions
    }{
        {
            name: "agents",
            options: &driver.CreateCollectionOptions{
                Type: driver.CollectionTypeDocument,
            },
            indexes: []driver.EnsureIndexOptions{
                {
                    Type:   driver.HashIndex,
                    Fields: []string{"status"},
                },
                {
                    Type:   driver.HashIndex,
                    Fields: []string{"type"},
                },
                {
                    Type:   driver.SkipListIndex,
                    Fields: []string{"createdAt"},
                },
                {
                    Type:   driver.GeoIndex,
                    Fields: []string{"deployment.location"},
                },
            },
        },
        {
            name: "workflow_executions",
            options: &driver.CreateCollectionOptions{
                Type: driver.CollectionTypeDocument,
            },
            indexes: []driver.EnsureIndexOptions{
                {
                    Type:   driver.HashIndex,
                    Fields: []string{"workflowId"},
                },
                {
                    Type:   driver.HashIndex,
                    Fields: []string{"status"},
                },
                {
                    Type:   driver.SkipListIndex,
                    Fields: []string{"startTime"},
                },
            },
        },
        {
            name: "agent_communications",
            options: &driver.CreateCollectionOptions{
                Type: driver.CollectionTypeEdge,
            },
            indexes: []driver.EnsureIndexOptions{
                {
                    Type:   driver.SkipListIndex,
                    Fields: []string{"timestamp"},
                },
                {
                    Type:   driver.TTLIndex,
                    Fields: []string{"ttl"},
                    ExpireAfter: 0, // Use document TTL field
                },
            },
        },
    }
    
    for _, collDef := range collections {
        // Create collection if it doesn't exist
        coll, err := db.Collection(ctx, collDef.name)
        if driver.IsNotFound(err) {
            coll, err = db.CreateCollection(ctx, collDef.name, collDef.options)
            if err != nil {
                return fmt.Errorf("failed to create collection %s: %w", collDef.name, err)
            }
        } else if err != nil {
            return fmt.Errorf("failed to access collection %s: %w", collDef.name, err)
        }
        
        // Ensure indexes
        for _, indexOpts := range collDef.indexes {
            _, _, err := coll.EnsureIndex(ctx, indexOpts)
            if err != nil {
                return fmt.Errorf("failed to create index on collection %s: %w", collDef.name, err)
            }
        }
    }
    
    return nil
}
```

## 4. Performance Optimization

### Query Optimization

```go
// Optimized agent queries with proper indexing
type AgentRepository struct {
    db     driver.Database
    agents driver.Collection
}

func (ar *AgentRepository) FindActiveAgentsByType(ctx context.Context, agentType string) ([]*Agent, error) {
    query := `
        FOR agent IN agents
        FILTER agent.type == @type AND agent.status == @status
        SORT agent.createdAt DESC
        RETURN agent
    `
    
    bindVars := map[string]interface{}{
        "type":   agentType,
        "status": "running",
    }
    
    cursor, err := ar.db.Query(ctx, query, bindVars)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer cursor.Close()
    
    var agents []*Agent
    for {
        var agent Agent
        if _, err := cursor.ReadDocument(ctx, &agent); err != nil {
            if driver.IsNoMoreDocuments(err) {
                break
            }
            return nil, fmt.Errorf("failed to read document: %w", err)
        }
        agents = append(agents, &agent)
    }
    
    return agents, nil
}

// Graph traversal for agent communication patterns
func (ar *AgentRepository) GetCommunicationGraph(ctx context.Context, agentID string, depth int) (*CommunicationGraph, error) {
    query := `
        FOR vertex, edge, path IN 1..@depth ANY @startAgent agent_communications
        RETURN {
            "vertex": vertex,
            "edge": edge,
            "path": path
        }
    `
    
    bindVars := map[string]interface{}{
        "startAgent": fmt.Sprintf("agents/%s", agentID),
        "depth":      depth,
    }
    
    cursor, err := ar.db.Query(ctx, query, bindVars)
    if err != nil {
        return nil, fmt.Errorf("graph traversal failed: %w", err)
    }
    defer cursor.Close()
    
    graph := &CommunicationGraph{
        Nodes: make(map[string]*Agent),
        Edges: make([]*Communication, 0),
    }
    
    for {
        var result struct {
            Vertex json.RawMessage `json:"vertex"`
            Edge   json.RawMessage `json:"edge"`
            Path   json.RawMessage `json:"path"`
        }
        
        if _, err := cursor.ReadDocument(ctx, &result); err != nil {
            if driver.IsNoMoreDocuments(err) {
                break
            }
            return nil, fmt.Errorf("failed to read graph result: %w", err)
        }
        
        // Process vertex and edge data
        // ... implementation details
    }
    
    return graph, nil
}
```

## 5. Security Configuration

### Authentication and Authorization

```yaml
# ArangoDB Security Configuration
apiVersion: v1
kind: Secret
metadata:
  name: pweza-arango-auth
  namespace: pweza-core-data
type: Opaque
data:
  username: cHdlemEtY29yZQ==  # base64: pweza-core
  password: # generated secure password
  jwt-secret: # JWT signing secret
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: arango-security-config
  namespace: pweza-core-data
data:
  arangod.conf: |
    [server]
    authentication = true
    
    [ssl]
    keyfile = /etc/ssl/arangodb/tls.key
    cafile = /etc/ssl/arangodb/ca.crt
    
    [log]
    level = info
    
    [database]
    maximal-journal-size = 33554432
    
    [cluster]
    agency-size = 3
    
    [foxx]
    enable = false
```

## 6. Monitoring and Metrics

### ArangoDB Exporter Configuration

```yaml
# Prometheus ArangoDB Exporter
apiVersion: apps/v1
kind: Deployment
metadata:
  name: arangodb-exporter
  namespace: pweza-core-data
spec:
  replicas: 1
  selector:
    matchLabels:
      app: arangodb-exporter
  template:
    metadata:
      labels:
        app: arangodb-exporter
    spec:
      containers:
      - name: exporter
        image: arangodb/arangodb-exporter:0.1.6
        ports:
        - containerPort: 9101
          name: metrics
        env:
        - name: ARANGO_SERVER
          value: "http://pweza-core-arango:8529"
        - name: ARANGO_USER
          valueFrom:
            secretKeyRef:
              name: pweza-arango-auth
              key: username
        - name: ARANGO_PASSWORD
          valueFrom:
            secretKeyRef:
              name: pweza-arango-auth
              key: password
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
```

This ArangoDB infrastructure provides a robust, scalable foundation for CodeValdCortex's multi-model data requirements with enterprise-grade security and performance optimization.