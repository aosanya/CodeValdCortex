# UC-LOG-001 System Architecture: Smart Logistics Service Platform

**Version**: 1.0  
**Last Updated**: October 22, 2025  
**Status**: Design Phase

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Component Descriptions](#component-descriptions)
4. [Data Flow Patterns](#data-flow-patterns)
5. [Agent Deployment Model](#agent-deployment-model)
6. [Scalability Design](#scalability-design)
7. [Resilience and Fault Tolerance](#resilience-and-fault-tolerance)
8. [Security Architecture](#security-architecture)
9. [Monitoring and Observability](#monitoring-and-observability)
10. [Deployment Strategy](#deployment-strategy)
11. [Technology Decisions](#technology-decisions)
12. [Future Enhancements](#future-enhancements)

## Architecture Overview

CodeValdUsafiri follows a **4-tier agent-oriented architecture**:

1. **Client Layer** - Mobile/web applications for shippers, drivers, and facility managers
2. **API Gateway Layer** - REST API endpoints, WebSocket servers, authentication
3. **Agent Runtime Layer** - Core autonomous agents executing business logic
4. **Data Layer** - Persistent storage, caching, message queuing

The architecture emphasizes:
- **Agent Autonomy**: Each agent independently evaluates situations and makes decisions
- **Asynchronous Communication**: Agents communicate via message passing (pub/sub, direct messaging)
- **Event-Driven**: State changes trigger events that cascade through the system
- **Horizontal Scalability**: Stateless agents can be replicated across multiple instances
- **Real-Time Responsiveness**: Sub-second updates for GPS tracking and status changes

### Key Architectural Patterns

- **Agent-Oriented Architecture**: Domain entities as autonomous agents
- **Event Sourcing**: State changes captured as immutable events
- **CQRS (Command Query Responsibility Segregation)**: Separate read/write paths for performance
- **Circuit Breaker**: Graceful degradation when external services fail
- **Saga Pattern**: Distributed transactions across agent interactions

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CLIENT LAYER                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                  │
│  │   Shipper    │    │    Driver    │    │   Facility   │                  │
│  │  Mobile App  │    │  Mobile App  │    │  Web Portal  │                  │
│  │ (React Nat.) │    │ (React Nat.) │    │   (React)    │                  │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘                  │
│         │                   │                   │                            │
│         └───────────────────┴───────────────────┘                            │
│                             │                                                │
│                  HTTPS / WebSocket (TLS 1.3)                                 │
│                             │                                                │
└─────────────────────────────┼────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────────────────────┐
│                    API GATEWAY LAYER                                          │
├─────────────────────────────┼────────────────────────────────────────────────┤
│                             │                                                │
│         ┌───────────────────┴───────────────────┐                           │
│         │        API Gateway (Kong/Nginx)        │                           │
│         │  • Rate Limiting • Auth • TLS Term.   │                           │
│         └───────────┬───────────────┬───────────┘                           │
│                     │               │                                        │
│          ┌──────────┴────┐    ┌────┴─────────┐                             │
│          │  REST API     │    │  WebSocket   │                             │
│          │  Endpoints    │    │    Server    │                             │
│          │  (Go/Gin)     │    │  (Gorilla)   │                             │
│          └──────┬────────┘    └────┬─────────┘                             │
│                 │                  │                                        │
│                 └──────────┬───────┘                                        │
│                            │                                                │
└────────────────────────────┼────────────────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────────────────┐
│                     AGENT RUNTIME LAYER                                      │
├────────────────────────────┼────────────────────────────────────────────────┤
│                            │                                                │
│  ┌────────────────────────────────────────────────────────────────┐        │
│  │              CodeValdCortex Runtime Manager                     │        │
│  │   • Agent Lifecycle  • Registry  • Message Router              │        │
│  └────────────────────────────────────────────────────────────────┘        │
│                            │                                                │
│         ┌──────────────────┼──────────────────┐                            │
│         │                  │                  │                            │
│  ┌──────▼──────┐    ┌──────▼──────┐   ┌──────▼──────┐                    │
│  │  Shipper    │    │   Driver/   │   │  Facility   │                    │
│  │   Agents    │    │ Truck Agents│   │   Agents    │                    │
│  │             │    │             │   │             │                    │
│  │ • Register  │    │ • Register  │   │ • Manage    │                    │
│  │ • Request   │    │ • Evaluate  │   │   capacity  │                    │
│  │ • Monitor   │    │ • Bid       │   │ • Schedule  │                    │
│  │ • Rate      │    │ • Execute   │   │   docks     │                    │
│  └──────┬──────┘    └──────┬──────┘   └──────┬──────┘                    │
│         │                  │                  │                            │
│         └──────────────────┼──────────────────┘                            │
│                            │                                                │
│         ┌──────────────────┼──────────────────┐                            │
│         │                  │                  │                            │
│  ┌──────▼──────┐    ┌──────▼──────┐   ┌──────▼──────┐                    │
│  │  Platform   │    │    Route    │   │   Payment   │                    │
│  │   Agent     │    │  Optimizer  │   │    Agent    │                    │
│  │             │    │    Agent    │   │             │                    │
│  │ • Broadcast │    │ • Calculate │   │ • Escrow    │                    │
│  │ • Reconcile │    │   routes    │   │ • Release   │                    │
│  │ • Match     │    │ • Traffic   │   │ • Dispute   │                    │
│  └──────┬──────┘    └──────┬──────┘   └──────┬──────┘                    │
│         │                  │                  │                            │
│         └──────────────────┼──────────────────┘                            │
│                            │                                                │
│         ┌──────────────────┼──────────────────┐                            │
│         │                  │                  │                            │
│  ┌──────▼──────┐    ┌──────▼──────┐   ┌──────▼──────┐                    │
│  │  Conflict   │    │  Analytics  │   │   Health    │                    │
│  │  Resolver   │    │    Agent    │   │   Monitor   │                    │
│  │             │    │             │   │    Agent    │                    │
│  │ • Mediate   │    │ • Track KPI │   │ • Heartbeat │                    │
│  │ • Re-route  │    │ • Predict   │   │ • Alert     │                    │
│  │ • Emergency │    │ • Report    │   │ • Recover   │                    │
│  └──────┬──────┘    └──────┬──────┘   └──────┬──────┘                    │
│         │                  │                  │                            │
│         └──────────────────┴──────────────────┘                            │
│                            │                                                │
│  ┌─────────────────────────┴────────────────────────────────────┐         │
│  │                  Message Bus (Redis Pub/Sub)                  │         │
│  │  • Event Streaming  • Direct Messaging  • Task Queue         │         │
│  └───────────────────────────────────────────────────────────────┘         │
│                            │                                                │
└────────────────────────────┼────────────────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────────────────┐
│                         DATA LAYER                                           │
├────────────────────────────┼────────────────────────────────────────────────┤
│                            │                                                │
│  ┌──────────────┐   ┌──────▼──────┐   ┌──────────────┐                   │
│  │  PostgreSQL  │   │    Redis    │   │ TimescaleDB  │                   │
│  │              │   │             │   │              │                   │
│  │ • Agents     │   │ • Cache     │   │ • GPS Track  │                   │
│  │ • Bookings   │   │ • Sessions  │   │ • Metrics    │                   │
│  │ • Users      │   │ • Real-time │   │ • Analytics  │                   │
│  │ • Payments   │   │   state     │   │              │                   │
│  └──────────────┘   └─────────────┘   └──────────────┘                   │
│                                                                             │
│  ┌──────────────┐   ┌─────────────┐                                       │
│  │    AWS S3    │   │ Elasticsearch│                                      │
│  │              │   │              │                                       │
│  │ • Documents  │   │ • Search     │                                      │
│  │ • Photos     │   │ • Logs       │                                      │
│  │ • Invoices   │   │ • Audit      │                                      │
│  └──────────────┘   └──────────────┘                                      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                     EXTERNAL INTEGRATIONS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Google Maps  │  │    Twilio    │  │   SendGrid   │  │   M-Pesa/    │  │
│  │   Platform   │  │     SMS      │  │    Email     │  │   Stripe     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Component Descriptions

### Client Layer

#### 1. Shipper Mobile App (React Native)
**Purpose**: Enable shippers to request shipments, track deliveries, manage payments

**Key Features**:
- Shipment request creation with pickup/delivery details
- Real-time bid comparison and selection
- Live GPS tracking of active shipments
- Digital document signing (BoL, PoD)
- Payment and invoice management
- Driver ratings and feedback

**Technology**: React Native (iOS/Android), Redux for state management, React Navigation

#### 2. Driver Mobile App (React Native)
**Purpose**: Enable drivers to receive requests, navigate routes, update statuses

**Key Features**:
- Real-time request notifications
- Bid submission with custom pricing
- Turn-by-turn navigation with traffic updates
- Photo capture for cargo verification
- Digital signature capture
- Earnings and performance dashboard
- Return cargo opportunity alerts

**Technology**: React Native, Google Maps SDK, Camera API, Background geolocation

#### 3. Facility Web Portal (React)
**Purpose**: Enable facility managers to coordinate dock scheduling and monitor operations

**Key Features**:
- Dock capacity calendar management
- Incoming shipment notifications
- Loading/unloading coordination
- Digital check-in/check-out
- Performance analytics
- Issue reporting and resolution

**Technology**: React, Material-UI, WebSocket for real-time updates

### API Gateway Layer

#### 4. API Gateway (Kong/Nginx)
**Purpose**: Single entry point for all client requests with cross-cutting concerns

**Responsibilities**:
- **Authentication**: Validate JWT tokens, refresh tokens
- **Rate Limiting**: Prevent abuse (100 req/min per user, 1000 req/min per API key)
- **TLS Termination**: Handle HTTPS encryption
- **Request Routing**: Route to appropriate backend services
- **Load Balancing**: Distribute traffic across API server instances
- **CORS Handling**: Manage cross-origin requests
- **Analytics**: Log all API traffic for monitoring

**Technology**: Kong API Gateway with PostgreSQL backend, or Nginx with Lua scripts

#### 5. REST API Server (Go/Gin)
**Purpose**: Handle synchronous requests for CRUD operations

**Endpoints**:
- `/api/v1/shipments` - Create, retrieve, update shipments
- `/api/v1/bids` - Submit and manage bids
- `/api/v1/agents` - Agent registration and profiles
- `/api/v1/payments` - Payment processing
- `/api/v1/facilities` - Facility management

**Technology**: Go 1.21+, Gin web framework, JWT authentication middleware

#### 6. WebSocket Server (Gorilla)
**Purpose**: Handle real-time bidirectional communication

**Use Cases**:
- GPS location streaming from drivers
- Real-time bid notifications to shippers
- Status update broadcasts
- Live tracking for shippers

**Technology**: Go with Gorilla WebSocket, Redis for pub/sub

### Agent Runtime Layer

#### 7. CodeValdCortex Runtime Manager
**Purpose**: Core framework managing agent lifecycle and communication

**Components**:
- **Agent Registry**: Track all active agents, their capabilities, and locations
- **Lifecycle Manager**: Start, stop, restart agents; handle crashes
- **Message Router**: Route messages between agents based on subscriptions
- **Task Scheduler**: Schedule periodic tasks (health checks, analytics)
- **Configuration Service**: Manage agent configuration dynamically
- **Health Monitor**: Track agent health, detect failures

**Technology**: Go, PostgreSQL for agent state, Redis for messaging

#### 8. Shipper Agents
**Purpose**: Represent shippers and manage their shipment requests

**State**: Registered, RequestPending, BidsReceived, Booked, InTransit, Delivered, Disputed

**Key Operations**:
- Create and broadcast shipment requests
- Receive and evaluate bids
- Select winning bid
- Monitor shipment progress
- Provide feedback and ratings

**Communication**:
- Publishes: `shipment.requested`, `bid.selected`
- Subscribes: `bid.submitted`, `shipment.status_changed`

#### 9. Driver/Truck Agents
**Purpose**: Represent drivers/vehicles and execute deliveries

**State**: Offline, Available, Evaluating, Committed, EnRoute, Loading, InTransit, Unloading, ReturnSearch

**Key Operations**:
- Receive shipment broadcasts
- Evaluate feasibility and profitability
- Submit competitive bids
- Execute pickup and delivery
- Update real-time location and status
- Search for return cargo

**Communication**:
- Publishes: `bid.submitted`, `location.updated`, `status.changed`
- Subscribes: `shipment.requested`, `booking.confirmed`, `route.updated`

#### 10. Facility Agents
**Purpose**: Manage warehouse/depot operations and dock scheduling

**State**: Operational, AtCapacity, Maintenance, Closed

**Key Operations**:
- Monitor dock availability
- Schedule appointments
- Coordinate loading/unloading
- Verify cargo
- Report delays or issues

**Communication**:
- Publishes: `dock.scheduled`, `loading.completed`, `delay.reported`
- Subscribes: `booking.confirmed`, `driver.arriving`

#### 11. Platform Agent
**Purpose**: Orchestrate the entire booking lifecycle

**Key Operations**:
- Broadcast shipment requests to eligible drivers
- Collect and rank bids using multi-criteria scoring
- Match shippers with drivers
- Enforce business rules and policies
- Handle escalations

**Algorithm (Bid Ranking)**:
```
Score = (0.3 × PriceScore) + (0.3 × TimeScore) + (0.2 × ReliabilityScore) +
        (0.1 × ExperienceScore) + (0.1 × VehicleQualityScore)

PriceScore = 1 - (BidPrice - MinBidPrice) / (MaxBidPrice - MinBidPrice)
TimeScore = 1 - (EstimatedTime - MinTime) / (MaxTime - MinTime)
ReliabilityScore = OnTimeDeliveryRate (0-1)
ExperienceScore = min(SimilarShipments / 100, 1.0)
VehicleQualityScore = (VehicleAge < 5) ? 1.0 : (VehicleAge < 10) ? 0.7 : 0.5
```

#### 12. Route Optimizer Agent
**Purpose**: Calculate optimal routes and provide navigation

**Key Operations**:
- Calculate shortest/fastest routes using traffic data
- Optimize multi-stop delivery sequences (TSP solver)
- Monitor traffic and suggest re-routing
- Calculate accurate ETAs
- Identify return cargo opportunities en route

**Algorithms**:
- **Single Route**: A* algorithm with traffic weights
- **Multi-Stop**: Greedy nearest-neighbor + 2-opt optimization
- **ETA Prediction**: Historical data + current traffic conditions

**Technology**: Google Maps Directions API, custom Go algorithms, Redis for caching routes

#### 13. Payment Agent
**Purpose**: Manage all financial transactions

**State**: Pending, Held, Released, Refunded, Disputed

**Key Operations**:
- Process payment holds when shipment booked
- Release payment to driver after delivery confirmation
- Handle refunds for cancellations
- Manage dispute escrow
- Generate invoices and receipts

**Integration**: Stripe API for card payments, M-Pesa API for mobile money

#### 14. Conflict Resolver Agent
**Purpose**: Handle exceptions, disputes, and emergencies

**Key Operations**:
- Mediate disputes between shippers and drivers
- Coordinate emergency re-routing (weather, breakdowns, accidents)
- Facilitate cargo transfers to alternate drivers
- Escalate to human support when needed

**Decision Logic**:
- Auto-resolve simple issues (minor delays with documentation)
- Escalate complex disputes to human reviewers
- Trigger emergency protocols for safety issues

#### 15. Analytics Agent
**Purpose**: Track performance, identify patterns, provide insights

**Key Operations**:
- Calculate KPIs (on-time rate, average bid time, platform utilization)
- Identify high-performing drivers and facilities
- Predict demand patterns (time of day, day of week, seasonality)
- Generate reports for stakeholders
- Detect anomalies (fraud, abuse, system issues)

**Technology**: TimescaleDB for time-series queries, Python for ML models, Grafana for dashboards

## Data Flow Patterns

### Pattern 1: Shipment Request to Booking

```
[Shipper App] 
    │
    │ 1. POST /api/v1/shipments
    ▼
[REST API Server]
    │
    │ 2. Validate & Create Shipment Entity
    ▼
[PostgreSQL] ← Shipment stored
    │
    │ 3. Publish "shipment.requested" event
    ▼
[Redis Pub/Sub]
    │
    │ 4. Broadcast to eligible drivers
    ▼
[Platform Agent] ← Determines eligibility (capacity, location, type)
    │
    │ 5. Filter & Notify
    ├─────────┬─────────┬─────────┐
    ▼         ▼         ▼         ▼
[Driver 1] [Driver 2] [Driver 3] ... [Driver N]
    │         │         │
    │ 6. Evaluate (30-120s)
    │         │         │
    │ 7. Submit Bids    │
    ▼         ▼         ▼
[Platform Agent] ← Collect all bids
    │
    │ 8. Rank using multi-criteria scoring
    │
    │ 9. Present top 5 bids
    ▼
[Shipper App] ← Real-time notification
    │
    │ 10. Shipper selects bid
    ▼
[Platform Agent]
    │
    │ 11. Confirm booking
    ├────────────┬────────────┐
    ▼            ▼            ▼
[Driver]    [Shipper]   [Payment Agent]
            (Notified)   (Hold payment)
```

### Pattern 2: Real-Time GPS Tracking

```
[Driver App] ← Background location service (every 5 seconds)
    │
    │ GPS coordinates + heading + speed
    ▼
[WebSocket Connection]
    │
    │ Binary message (lat, lng, heading, speed, timestamp)
    ▼
[WebSocket Server]
    │
    ├─────────────────┬─────────────────┐
    │                 │                 │
    │ 1. Update Cache │ 2. Broadcast    │ 3. Store History
    ▼                 ▼                 ▼
[Redis Cache]   [WebSocket]     [TimescaleDB]
(TTL: 60s)      (to Shipper)    (permanent)
                    │
                    ▼
            [Shipper App] ← Live tracking map
```

### Pattern 3: Emergency Re-routing

```
[Route Optimizer Agent] ← Polls traffic API every 2 minutes
    │
    │ Detects: Road closure on planned route
    ▼
[Conflict Resolver Agent]
    │
    │ 1. Calculate alternate route
    ▼
[Route Optimizer Agent]
    │
    │ New route: +45 minutes, +20 km
    │
    │ 2. Notify all stakeholders
    ├────────────────┬────────────────┐
    ▼                ▼                ▼
[Driver App]    [Shipper App]   [Facility Agent]
(Updated nav)   (Delay notice)  (Reschedule dock)
```

## Agent Deployment Model

### Deployment Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Kubernetes Cluster                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌───────────────────────────────────────────────────────┐      │
│  │              Namespace: codevald-api                   │      │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │      │
│  │  │ API Server  │  │ API Server  │  │ API Server  │   │      │
│  │  │  Pod (x3)   │  │  Pod (x3)   │  │  Pod (x3)   │   │      │
│  │  └─────────────┘  └─────────────┘  └─────────────┘   │      │
│  │                                                        │      │
│  │  ┌─────────────┐  ┌─────────────┐                    │      │
│  │  │  WebSocket  │  │  WebSocket  │                    │      │
│  │  │ Server (x2) │  │ Server (x2) │                    │      │
│  │  └─────────────┘  └─────────────┘                    │      │
│  └───────────────────────────────────────────────────────┘      │
│                                                                   │
│  ┌───────────────────────────────────────────────────────┐      │
│  │            Namespace: codevald-agents                  │      │
│  │  ┌─────────────────────────────────────────────┐      │      │
│  │  │  Platform Agent Pool (x5 replicas)          │      │      │
│  │  │  • Stateless • Auto-scaling                 │      │      │
│  │  └─────────────────────────────────────────────┘      │      │
│  │                                                        │      │
│  │  ┌─────────────────────────────────────────────┐      │      │
│  │  │  Route Optimizer Pool (x3 replicas)         │      │      │
│  │  │  • CPU-intensive • Cached results           │      │      │
│  │  └─────────────────────────────────────────────┘      │      │
│  │                                                        │      │
│  │  ┌─────────────────────────────────────────────┐      │      │
│  │  │  Domain Agent Managers (Shipper, Driver, Facility) │      │
│  │  │  • Stateful Sets • Manage agent instances    │      │      │
│  │  │  • 1 Manager per 1000 domain agents          │      │      │
│  │  └─────────────────────────────────────────────┘      │      │
│  │                                                        │      │
│  │  ┌─────────────────────────────────────────────┐      │      │
│  │  │  Payment Agent (x2 replicas)                │      │      │
│  │  │  • Leader-follower for consistency          │      │      │
│  │  └─────────────────────────────────────────────┘      │      │
│  │                                                        │      │
│  │  ┌─────────────────────────────────────────────┐      │      │
│  │  │  Analytics Agent (x1, Cron Jobs)            │      │      │
│  │  │  • Scheduled batch processing               │      │      │
│  │  └─────────────────────────────────────────────┘      │      │
│  └───────────────────────────────────────────────────────┘      │
│                                                                   │
│  ┌───────────────────────────────────────────────────────┐      │
│  │               Namespace: codevald-data                 │      │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐            │      │
│  │  │PostgreSQL│  │  Redis   │  │Timescale │            │      │
│  │  │StatefulS.│  │Deployment│  │StatefulS.│            │      │
│  │  └──────────┘  └──────────┘  └──────────┘            │      │
│  └───────────────────────────────────────────────────────┘      │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Agent Instance Management

**Domain Agents (Shipper, Driver, Facility)**: 
- Not running 24/7; instantiated on-demand when users interact
- Managed by Agent Manager pods
- Lightweight: 10-20 MB memory per agent
- Can support 10,000+ concurrent agents per cluster

**System Agents (Platform, Route Optimizer, Payment)**:
- Always running, replicated for availability
- Stateless (state in PostgreSQL/Redis)
- Horizontally scalable

**Agent Manager Pattern**:
```go
type AgentManager struct {
    activeAgents map[string]*Agent  // agent_id -> Agent instance
    maxAgents    int                // e.g., 1000
    idleTimeout  time.Duration      // e.g., 30 minutes
}

func (m *AgentManager) GetOrCreateAgent(agentID string) *Agent {
    if agent, exists := m.activeAgents[agentID]; exists {
        agent.ResetIdleTimer()
        return agent
    }
    
    // Load agent state from database
    agent := m.LoadAgentFromDB(agentID)
    m.activeAgents[agentID] = agent
    
    // Start idle timer
    go m.MonitorIdleTimeout(agent)
    
    return agent
}

func (m *AgentManager) MonitorIdleTimeout(agent *Agent) {
    timer := time.NewTimer(m.idleTimeout)
    <-timer.C
    
    // Save state and remove from memory
    m.SaveAgentToDB(agent)
    delete(m.activeAgents, agent.ID)
}
```

## Scalability Design

### Horizontal Scaling Targets

| Component | Initial | Scale Target | Scaling Trigger |
|-----------|---------|--------------|-----------------|
| API Servers | 3 pods | 20 pods | CPU > 70% |
| WebSocket Servers | 2 pods | 10 pods | Connection count > 8000/pod |
| Platform Agents | 5 pods | 30 pods | Queue depth > 100 |
| Route Optimizers | 3 pods | 15 pods | CPU > 80% |
| Agent Managers | 5 pods | 50 pods | Active agents > 800/pod |

### Database Scaling

**PostgreSQL**:
- **Vertical**: Start with 16 GB RAM, scale to 64 GB
- **Read Replicas**: 2 replicas for read-heavy queries (tracking, analytics)
- **Partitioning**: Partition shipments table by date (monthly)
- **Indexing**: Composite indexes on (shipper_id, status, created_at)

**Redis**:
- **Cluster Mode**: 3 master nodes, 3 replica nodes
- **Sharding**: By agent_id for distributed cache
- **Persistence**: AOF (Append-Only File) every 1 second

**TimescaleDB**:
- **Hypertables**: Automatically partition by time (1-day chunks)
- **Retention**: Keep detailed GPS data for 90 days, then downsample to 1-minute intervals
- **Compression**: Compress chunks older than 7 days

### Performance Optimization

**Caching Strategy**:
```
1. API Response Caching (Redis):
   - Agent profiles: TTL 5 minutes
   - Facility availability: TTL 30 seconds
   - Historical routes: TTL 1 hour

2. Database Query Caching:
   - Frequent queries cached at application level
   - Use prepared statements to reduce parsing overhead

3. WebSocket Connection Pooling:
   - Reuse connections for multiple messages
   - Heartbeat every 30 seconds to keep alive
```

**Load Balancing**:
- **API Gateway**: Round-robin across API servers
- **WebSocket**: Sticky sessions (same client always to same server)
- **Agent Managers**: Hash-based routing by agent_id

## Resilience and Fault Tolerance

### Circuit Breaker Pattern

For external service calls (Google Maps, payment gateways):

```go
type CircuitBreaker struct {
    maxFailures   int
    failureCount  int
    state         string  // "closed", "open", "half-open"
    resetTimeout  time.Duration
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == "open" {
        return errors.New("Circuit breaker is OPEN")
    }
    
    err := fn()
    if err != nil {
        cb.failureCount++
        if cb.failureCount >= cb.maxFailures {
            cb.state = "open"
            go cb.AutoReset()
        }
        return err
    }
    
    cb.failureCount = 0
    return nil
}

func (cb *CircuitBreaker) AutoReset() {
    time.Sleep(cb.resetTimeout)
    cb.state = "half-open"
    // Try one request; if succeeds, close circuit
}
```

### Graceful Degradation

**Scenario 1: Google Maps API Down**
- Fallback: Use last known traffic data + historical averages
- Notify users: "ETA may be less accurate"

**Scenario 2: Payment Gateway Timeout**
- Fallback: Mark payment as "pending" and retry asynchronously
- Notify stakeholders: "Payment processing delayed"

**Scenario 3: Database Connection Loss**
- Fallback: Buffer writes in Redis, replay when database recovers
- Read-only mode: Serve cached data for reads

### High Availability

**Multi-Region Deployment**:
```
Primary Region: AWS eu-west-1 (Europe)
Secondary Region: AWS af-south-1 (Africa)

Active-Active:
- Both regions serve traffic
- Users routed to nearest region (latency-based DNS)
- Cross-region database replication (5-second lag)

Failover:
- Health checks every 10 seconds
- Automatic DNS failover if region unavailable (TTL: 60 seconds)
- Manual intervention required for split-brain scenarios
```

## Security Architecture

### Authentication Flow

```
1. User Login:
   [Client] → POST /api/v1/auth/login {phone, password}
   [API] → Validate credentials
   [API] → Generate JWT (access: 15 min, refresh: 7 days)
   [API] ← Return {access_token, refresh_token}

2. Authenticated Request:
   [Client] → GET /api/v1/shipments (Header: Authorization: Bearer <access_token>)
   [API Gateway] → Validate JWT signature
   [API Gateway] → Check expiration
   [API Gateway] → Extract user_id from claims
   [API Server] → Process request with user context

3. Token Refresh:
   [Client] → POST /api/v1/auth/refresh {refresh_token}
   [API] → Validate refresh token (check if revoked)
   [API] → Generate new access_token
   [API] ← Return {access_token}
```

### Data Encryption

**At Rest**:
- PostgreSQL: Transparent Data Encryption (TDE)
- S3: Server-side encryption with AWS KMS
- Redis: Encryption enabled (AES-256)

**In Transit**:
- TLS 1.3 for all external communication
- mTLS (mutual TLS) for inter-service communication within cluster

### Access Control

**Role-Based Access Control (RBAC)**:
```
Roles:
- Shipper: Can create shipments, view own shipments, rate drivers
- Driver: Can view broadcasts, submit bids, update GPS, upload photos
- Facility: Can manage docks, view scheduled arrivals
- Admin: Full access to all resources
- Support: Read-only access + ability to mediate disputes

Permissions enforced at API layer:
- Check JWT role claim
- Validate resource ownership (e.g., shipper can only view own shipments)
- Audit all access to sensitive data (PII, payment info)
```

## Monitoring and Observability

### Metrics (Prometheus)

**Application Metrics**:
```
# Request metrics
http_requests_total{method, endpoint, status}
http_request_duration_seconds{method, endpoint}

# Agent metrics
agent_active_count{type}
agent_message_processing_duration_seconds{agent_type, message_type}
agent_state_transitions_total{agent_type, from_state, to_state}

# Business metrics
shipments_created_total{service_level}
bids_submitted_total{outcome}  # outcome: accepted, rejected, timeout
deliveries_completed_total{on_time}
payment_transactions_total{status}
```

**Infrastructure Metrics**:
```
# System metrics (node-exporter)
node_cpu_usage
node_memory_usage
node_disk_io

# Database metrics
postgres_connections_active
postgres_query_duration_seconds
redis_connected_clients
redis_memory_usage_bytes
```

### Logging (ELK Stack)

**Structured Logging**:
```json
{
  "timestamp": "2025-10-22T14:32:15Z",
  "level": "INFO",
  "agent_type": "Driver",
  "agent_id": "DRV-NAI-89234",
  "message": "Bid submitted",
  "shipment_id": "SHP-54782-001",
  "bid_amount": 15000,
  "estimated_time": 480
}
```

**Log Aggregation**:
- All logs shipped to Elasticsearch
- Kibana dashboards for visualization
- Alerts configured for ERROR/CRITICAL logs

### Tracing (Jaeger)

Distributed tracing for request flows:
```
Trace ID: abc123
├─ API Gateway (2ms)
├─ REST API Server (15ms)
│  ├─ Platform Agent (50ms)
│  │  ├─ Database Query (20ms)
│  │  └─ Redis Publish (5ms)
│  └─ Response Serialization (3ms)
└─ Total: 95ms
```

### Alerting

**Critical Alerts** (PagerDuty):
- Service down (health check fails for 2+ minutes)
- Error rate > 5%
- P99 latency > 3 seconds
- Database connection pool exhausted
- Payment processing failures > 10 in 5 minutes

**Warning Alerts** (Slack):
- Error rate > 1%
- CPU usage > 80%
- Memory usage > 85%
- API rate limit approaching (80% of limit)

## Deployment Strategy

### CI/CD Pipeline

```
┌──────────────┐
│  Git Push    │
│  to GitHub   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ GitHub Actions│
│ CI Workflow  │
│              │
│ 1. Unit Tests│
│ 2. Lint      │
│ 3. Build     │
│ 4. Integration│
│    Tests     │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Build Docker │
│ Image        │
│ Push to ECR  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Deploy to    │
│ Staging      │
│ (Kubernetes) │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Automated    │
│ E2E Tests    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Manual       │
│ Approval     │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Deploy to    │
│ Production   │
│ (Blue-Green) │
└──────────────┘
```

### Blue-Green Deployment

```
Production Environment:

┌─────────────────┐     ┌─────────────────┐
│  Blue (v1.2.3)  │     │ Green (v1.2.4)  │
│  [Active]       │     │  [Standby]      │
│                 │     │                 │
│  50% traffic    │◄───►│  0% traffic     │
└────────▲────────┘     └────────▲────────┘
         │                       │
         │                       │
    ┌────┴───────────────────────┴────┐
    │     Load Balancer (Kong)        │
    │   • Health checks every 10s     │
    │   • Gradual traffic shift       │
    └─────────────────────────────────┘

Deployment Steps:
1. Deploy v1.2.4 to Green environment
2. Run health checks (5 minutes)
3. Shift 10% traffic to Green
4. Monitor metrics (15 minutes)
5. If OK: Shift 50%, then 100% traffic
6. If Issues: Instant rollback to Blue
7. After 24 hours: Decommission Blue
```

### Rollback Strategy

**Automatic Rollback Triggers**:
- Error rate > 10% within 5 minutes of deployment
- P99 latency > 5 seconds
- More than 5 consecutive health check failures

**Manual Rollback**:
```bash
# Instant rollback via kubectl
kubectl set image deployment/api-server api-server=codevald/api:v1.2.3

# Expected rollback time: < 30 seconds
```

## Technology Decisions

### Why Go for Backend?

**Decision**: Use Go as primary backend language

**Rationale**:
- **Concurrency**: Goroutines ideal for handling thousands of agent instances
- **Performance**: Compiled language with low latency and memory footprint
- **Ecosystem**: Excellent libraries (Gin, Gorilla, gRPC)
- **Deployment**: Single binary deployment simplifies operations
- **Team Expertise**: Team has strong Go experience

**Trade-offs**:
- **Learning Curve**: Steeper than Python for some team members
- **ML Libraries**: Less mature than Python; need to use microservices for ML

### Why PostgreSQL?

**Decision**: PostgreSQL as primary relational database

**Rationale**:
- **ACID Compliance**: Critical for financial transactions
- **JSON Support**: Flexible schema for agent attributes
- **Mature**: Battle-tested at scale
- **Extensions**: PostGIS for geospatial queries, pg_cron for scheduled tasks
- **Open Source**: No licensing costs

### Why Redis?

**Decision**: Redis for caching and message broker

**Rationale**:
- **Speed**: Sub-millisecond latency for cache hits
- **Pub/Sub**: Built-in pub/sub for agent messaging
- **Data Structures**: Lists, sets, sorted sets useful for queue management
- **Persistence**: Optional persistence for important cache data

### Why TimescaleDB?

**Decision**: TimescaleDB for GPS tracking and time-series data

**Rationale**:
- **Time-Series Optimized**: Automatic partitioning by time
- **PostgreSQL Compatible**: Use same client libraries
- **Compression**: 90% space savings for historical data
- **Continuous Aggregates**: Pre-computed analytics queries

### Why React Native?

**Decision**: React Native for mobile apps

**Rationale**:
- **Code Sharing**: 70-80% code shared between iOS and Android
- **Performance**: Near-native performance for UI
- **Ecosystem**: Mature libraries (React Navigation, Maps)
- **Developer Velocity**: Faster iterations than native development

**Trade-offs**:
- **Platform-Specific Code**: Still need native modules for GPS, camera
- **App Size**: Slightly larger than pure native apps

## Future Enhancements

### Phase 2: Advanced Intelligence

1. **Predictive Demand Forecasting**
   - ML model predicts shipment volume by route/time
   - Incentivize drivers to be available at predicted high-demand times
   - Dynamic pricing based on predicted demand

2. **Collaborative Routing**
   - Multiple trucks share route information
   - Convoy formation for long-distance travel (fuel savings, safety)
   - Dynamic load balancing (transfer cargo between trucks en route)

3. **Autonomous Vehicle Integration**
   - Support for self-driving trucks
   - Remote monitoring and intervention
   - Regulatory compliance tracking

### Phase 3: Ecosystem Expansion

1. **Multi-Modal Transport**
   - Integration with rail, air, sea freight
   - Automated mode selection based on cost/time trade-offs
   - Container tracking across modalities

2. **International Expansion**
   - Cross-border shipments with customs automation
   - Multi-currency support
   - Localization (languages, units, regulations)

3. **API Marketplace**
   - Public API for third-party integrations
   - White-label solutions for enterprise clients
   - Partner ecosystem (insurance, financing, maintenance)

### Phase 4: Sustainability

1. **Carbon Footprint Tracking**
   - Calculate emissions per shipment
   - Carbon offset purchasing
   - Green routing (prefer fuel-efficient routes even if slightly slower)

2. **Electric Vehicle Incentives**
   - Higher priority for EV trucks in bid ranking
   - Charging station network integration
   - Battery range optimization

## Related Documents

- [README](./README.md) - Use case overview and quick reference
- [Agent Design](./agent-design.md) - Detailed agent specifications
- [Use Case Specification](../../../1-SoftwareRequirements/requirements/use-cases/UC-LOG-001-smart-logistics-platform.md) - Functional requirements

---

**Maintained by**: Architecture Team  
**Review Cadence**: Weekly during design phase  
**Next Review**: October 29, 2025
