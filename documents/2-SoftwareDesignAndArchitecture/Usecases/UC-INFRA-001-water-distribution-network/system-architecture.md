# System Architecture - Water Distribution Network Management

## Architecture Overview

The Water Distribution Network Management system follows a hierarchical edge-to-cloud architecture with autonomous agents distributed across multiple deployment tiers.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Central Control (Cloud/On-Premise)              │
│  ┌──────────────────┐  ┌──────────────────┐  ┌────────────────────┐   │
│  │  Control Center  │  │   Dashboard      │  │   Analytics        │   │
│  │  Agent           │  │   (MVP-015)      │  │   Engine           │   │
│  └──────────────────┘  └──────────────────┘  └────────────────────┘   │
│  ┌──────────────────┐  ┌──────────────────┐  ┌────────────────────┐   │
│  │  API Gateway     │  │   ML Models      │  │   External         │   │
│  │                  │  │                  │  │   Integrations     │   │
│  └──────────────────┘  └──────────────────┘  └────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ (HTTPS/WebSocket)
                                    │
┌─────────────────────────────────────────────────────────────────────────┐
│                       Regional Servers (Data Centers)                    │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │            CodeValdCortex Runtime Manager                        │  │
│  │  ┌────────────────┐  ┌────────────────┐  ┌──────────────────┐  │  │
│  │  │ Agent Registry │  │ Communication  │  │  Task System     │  │  │
│  │  └────────────────┘  └────────────────┘  └──────────────────┘  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────┐  ┌──────────────────┐  ┌────────────────────┐   │
│  │  Message Broker  │  │   Time-Series    │  │   Configuration    │   │
│  │  (Redis)         │  │   Database       │  │   Service          │   │
│  └──────────────────┘  └──────────────────┘  └────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ (MQTT/TCP)
                                    │
┌─────────────────────────────────────────────────────────────────────────┐
│                    Field Gateways (Local Agent Clusters)                 │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │               Zone Coordinator Agents                            │  │
│  │  ┌────────────────┐  ┌────────────────┐  ┌──────────────────┐  │  │
│  │  │ Reservoir      │  │ Pump           │  │  Network         │  │  │
│  │  │ Agent          │  │ Agent          │  │  Optimizer       │  │  │
│  │  └────────────────┘  └────────────────┘  └──────────────────┘  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │               Infrastructure Agents                              │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌──────────┐  │  │
│  │  │ Pipe       │  │ Valve      │  │ Hydrant    │  │  Meter   │  │  │
│  │  │ Agents     │  │ Agents     │  │ Agents     │  │  Agents  │  │  │
│  │  └────────────┘  └────────────┘  └────────────┘  └──────────┘  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │               Sensor Agents                                      │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌──────────┐  │  │
│  │  │ Pressure   │  │ Flow       │  │ Quality    │  │  Temp    │  │  │
│  │  │ Sensors    │  │ Sensors    │  │ Sensors    │  │  Sensors │  │  │
│  │  └────────────┘  └────────────┘  └────────────┘  └──────────┘  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌──────────────────┐  ┌──────────────────┐                            │
│  │  Local Cache     │  │   Edge Database  │                            │
│  │  (Redis)         │  │   (PostgreSQL)   │                            │
│  └──────────────────┘  └──────────────────┘                            │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ (MQTT/Modbus/OPC UA)
                                    │
┌─────────────────────────────────────────────────────────────────────────┐
│                        Edge Devices (IoT Sensors)                        │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌──────────────────┐ │
│  │ Pressure   │  │ Flow       │  │ Quality    │  │  Temperature     │ │
│  │ Sensors    │  │ Meters     │  │ Monitors   │  │  Sensors         │ │
│  └────────────┘  └────────────┘  └────────────┘  └──────────────────┘ │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌──────────────────┐ │
│  │ Smart      │  │ Valve      │  │ Pump       │  │  Level           │ │
│  │ Meters     │  │ Actuators  │  │ Controllers│  │  Sensors         │ │
│  └────────────┘  └────────────┘  └────────────┘  └──────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

## Component Descriptions

### Central Control Layer

**Purpose**: System-wide monitoring, control, and analytics

**Components**:
- **Control Center Agent**: Master coordination agent for the entire network
- **Dashboard (MVP-015)**: Web-based visualization and control interface
- **Analytics Engine**: Historical analysis, reporting, and ML model training
- **API Gateway**: External API access for integrations
- **ML Models**: Predictive maintenance, leak detection, optimization models
- **External Integrations**: SCADA, GIS, CRM, Emergency Services

**Deployment**: Cloud (AWS/Azure/GCP) or on-premise data center

**Scalability**: Load-balanced, horizontally scalable services

### Regional Server Layer

**Purpose**: Agent runtime management and data aggregation for geographic regions

**Components**:
- **CodeValdCortex Runtime Manager**: Agent lifecycle management
- **Agent Registry**: Central registry of all agents and relationships
- **Communication System**: Message routing and pub/sub
- **Task System**: Scheduled tasks, maintenance windows
- **Message Broker (Redis)**: High-performance message queuing
- **Time-Series Database (TimescaleDB)**: Sensor data, metrics, trends
- **Configuration Service**: Agent parameters, thresholds, rules

**Deployment**: Regional data centers for low-latency access

**Scalability**: Partitioned by geographic zone

### Field Gateway Layer

**Purpose**: Local agent hosting at the network edge with offline capability

**Components**:
- **Zone Coordinator Agents**: Manage local network segment
- **Infrastructure Agents**: Pipe, Valve, Hydrant, Meter agents
- **Sensor Agents**: Process IoT sensor data locally
- **Reservoir/Pump Agents**: Control critical infrastructure
- **Network Optimizer**: Local pressure/flow optimization
- **Local Cache**: Redis for fast agent communication
- **Edge Database**: PostgreSQL for local state persistence

**Deployment**: Industrial PCs at pumping stations, reservoirs, or key network locations

**Scalability**: One gateway per network zone (typically 500-2000 devices)

**Resilience**: Continues operation if regional server connection is lost

### Edge Device Layer

**Purpose**: Physical sensors and actuators in the field

**Components**:
- **IoT Sensors**: Pressure, flow, quality, temperature, level sensors
- **Smart Meters**: Customer consumption monitoring
- **Valve Actuators**: Remote valve control
- **Pump Controllers**: Pump speed and operation control

**Deployment**: Distributed throughout water distribution network

**Protocols**: MQTT, Modbus RTU/TCP, OPC UA, LoRaWAN

**Power**: Battery, solar, or line-powered depending on location

## Data Flow Patterns

### 1. Real-Time Monitoring Flow

```
Sensor Device → Field Gateway (Sensor Agent) → Regional Server (Agent Registry)
                                             → Control Center (Dashboard)
                                             → ML Models (Anomaly Detection)
```

**Frequency**: 1-60 second intervals  
**Data**: Pressure, flow, quality readings  
**Protocol**: MQTT (QoS 1)

### 2. Alert Flow

```
Infrastructure Agent (detects anomaly) → Zone Coordinator → Regional Server
                                                          → Control Center
                                                          → Notification Service
                                                          → Operators/Maintenance
```

**Latency**: <1 second for critical alerts  
**Protocol**: Redis Pub/Sub + WebSocket  
**Priority**: Emergency > Critical > Warning > Info

### 3. Control Command Flow

```
Control Center/Dashboard → Regional Server → Field Gateway → Valve/Pump Actuator
                                                           → Sensor Agent (verify)
                                                           ← Acknowledgment
```

**Latency**: <2 seconds end-to-end  
**Protocol**: HTTPS/WebSocket → TCP → Modbus/OPC UA  
**Safety**: Confirmation required, timeout handling

### 4. Predictive Maintenance Flow

```
Infrastructure Agents (collect metrics) → Time-Series Database → ML Models
                                                               → Predictions
                                       ← Maintenance Schedule ← Task System
```

**Frequency**: Hourly analysis, daily predictions  
**Data**: Vibration, temperature, efficiency, operation hours  
**Output**: Maintenance work orders with priority and timing

## Scalability Design

### Horizontal Scaling

- **Regional Servers**: Add servers per geographic region
- **Field Gateways**: Add gateways for network expansion
- **Agent Instances**: Agents scale independently based on infrastructure elements

### Vertical Scaling

- **Database Partitioning**: Time-series data partitioned by date/zone
- **Message Broker Clustering**: Redis Cluster for high-throughput messaging
- **Load Balancing**: API Gateway distributes requests across servers

### Performance Targets

| Metric | Target | Scale |
|--------|--------|-------|
| Agents per Gateway | 500-2000 | Per zone |
| Gateways per Regional Server | 50-100 | Per region |
| Total Agents per System | 100K-1M | System-wide |
| Message Throughput | 10K msg/sec | Per regional server |
| Database Time-Series Writes | 100K points/sec | Per database |
| Agent Response Time | <100ms | P99 |
| Alert Propagation | <1s | Critical alerts |

## Resilience and Fault Tolerance

### Field Gateway Resilience

- **Offline Operation**: Gateways continue local control if regional server unavailable
- **Local Cache**: Critical state cached in Redis
- **Automatic Reconnection**: Exponential backoff retry to regional server
- **Data Buffering**: Queue messages during outages, sync when reconnected

### Agent Resilience

- **Health Monitoring**: Each agent reports health status
- **Automatic Restart**: Failed agents automatically restarted by Runtime Manager
- **State Persistence**: Agent state persisted to database
- **Graceful Degradation**: Reduced functionality if dependent services unavailable

### Infrastructure Resilience

- **Database Replication**: PostgreSQL streaming replication (primary + standby)
- **Message Broker HA**: Redis Sentinel for automatic failover
- **Multi-Region Deployment**: Active-active or active-passive regional servers
- **Backup Power**: Field gateways on UPS for continuous operation

## Security Architecture

### Network Security

- **TLS/SSL**: All communication encrypted in transit
- **VPN**: Field gateways connect to regional servers via VPN
- **Firewall Rules**: Strict ingress/egress rules per layer
- **Network Segmentation**: IoT devices on isolated VLAN

### Authentication & Authorization

- **Agent Authentication**: Mutual TLS certificates for agent-to-server
- **API Authentication**: OAuth 2.0 for external integrations
- **Role-Based Access Control (RBAC)**: Operators, engineers, administrators
- **Audit Logging**: All control commands logged with user attribution

### Data Security

- **Encryption at Rest**: Database encryption (AES-256)
- **Sensitive Data Protection**: Customer data encrypted and access-controlled
- **Backup Encryption**: Encrypted backups stored securely
- **Data Retention**: Time-series data retention policies (e.g., raw data 90 days, aggregated 7 years)

## Monitoring and Observability

### System Monitoring

- **Agent Health**: Each agent reports liveness and readiness
- **Infrastructure Metrics**: CPU, memory, disk, network per server
- **Application Metrics**: Message throughput, API latency, database performance
- **Business Metrics**: Water loss rate, maintenance efficiency, alert response time

### Logging

- **Structured Logging**: JSON format with correlation IDs
- **Centralized Log Aggregation**: ELK Stack or Grafana Loki
- **Log Levels**: DEBUG, INFO, WARN, ERROR, CRITICAL
- **Log Retention**: 30 days online, 1 year archived

### Alerting

- **System Alerts**: Infrastructure failures, service degradation
- **Operational Alerts**: Leak detection, pressure anomalies, equipment failures
- **Alert Channels**: Dashboard, email, SMS, PagerDuty
- **Alert Priority**: Automatic escalation for critical unacknowledged alerts

## Deployment Strategy

### Phased Rollout

1. **Pilot Zone** (Month 1-3): Deploy in small network zone, validate functionality
2. **Regional Expansion** (Month 4-6): Expand to full region, tune performance
3. **Multi-Region** (Month 7-9): Deploy across multiple regions
4. **Full Network** (Month 10-12): Complete network coverage

### Blue-Green Deployment

- Maintain parallel "blue" (current) and "green" (new) environments
- Switch traffic to green after validation
- Rollback capability by switching back to blue

### Database Migration Strategy

- **Schema Changes**: Use migration tools (e.g., Flyway, Liquibase)
- **Data Migration**: Gradual migration with validation
- **Zero-Downtime**: Online schema changes, read replicas during migration

## Technology Decisions

### Why CodeValdCortex Framework?

- Native agent runtime with lifecycle management
- Built-in communication system for agent messaging
- Task scheduling for maintenance and data collection
- Memory service for state persistence
- Configuration management for dynamic agent parameters

### Why Redis for Message Broker?

- Sub-millisecond latency for real-time messaging
- Pub/Sub pattern perfect for agent-to-agent communication
- High throughput (100K+ messages/sec)
- Simple deployment and operation

### Why TimescaleDB for Time-Series?

- PostgreSQL extension = familiar SQL interface
- Automatic data partitioning by time
- High compression rates (90%+) for sensor data
- Excellent query performance for time-range queries

### Why MQTT for IoT?

- Lightweight protocol for constrained devices
- Built-in Quality of Service (QoS) levels
- Efficient bandwidth usage
- Wide industry adoption

## Future Enhancements

### Phase 2 Features

- **ML-Based Optimization**: AI-driven pressure/flow optimization
- **Digital Twin**: Real-time 3D visualization of network
- **Mobile Apps**: Field technician apps for maintenance
- **Customer Portal**: Consumer water usage dashboard

### Phase 3 Features

- **Blockchain Integration**: Immutable audit trail for critical operations
- **Edge AI**: Run ML models on field gateways for ultra-low latency
- **Advanced Analytics**: Predictive demand forecasting, what-if scenarios
- **Integration Hub**: Pre-built connectors for common SCADA/GIS systems

## Related Documents

- [Agent Design](./agent-design.md)
- [Communication Patterns](./communication-patterns.md)
- [Deployment Architecture](./deployment-architecture.md)
