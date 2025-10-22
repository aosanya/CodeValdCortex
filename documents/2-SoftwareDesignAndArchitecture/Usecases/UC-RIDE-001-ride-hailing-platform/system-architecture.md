# UC-RIDE-001 System Architecture: RideLink Ride-Hailing Platform

**Version**: 1.0  
**Last Updated**: October 22, 2025  
**Status**: Design Phase

## Architecture Overview

RideLink follows a **real-time, event-driven architecture** optimized for sub-second responsiveness and high-frequency ride matching. The system emphasizes:

- **Ultra-Low Latency**: <10-second ride matching, <100ms GPS updates
- **Horizontal Scalability**: Support 50,000+ concurrent rides
- **Real-Time Intelligence**: Dynamic surge pricing, demand prediction, route optimization
- **Safety-First**: Continuous monitoring, instant emergency response

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                       CLIENT LAYER                               │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐              ┌──────────────┐                 │
│  │    Rider     │              │    Driver    │                 │
│  │  Mobile App  │              │  Mobile App  │                 │
│  │(React Native)│              │(React Native)│                 │
│  └──────┬───────┘              └──────┬───────┘                 │
│         │                             │                          │
│         └─────────────┬───────────────┘                          │
│                       │                                          │
│           WebSocket (Real-time) + HTTPS (REST)                   │
└───────────────────────┼──────────────────────────────────────────┘
                        │
┌───────────────────────┼──────────────────────────────────────────┐
│                  API GATEWAY LAYER                                │
├───────────────────────┼──────────────────────────────────────────┤
│    ┌──────────────────┴──────────────────┐                       │
│    │   API Gateway (Kong) + Load Balancer│                       │
│    │   • Rate Limiting  • Auth  • TLS    │                       │
│    └──────────┬────────────────┬─────────┘                       │
│               │                │                                  │
│      ┌────────┴────┐    ┌─────┴──────┐                          │
│      │  REST API   │    │ WebSocket  │                          │
│      │  (Go/Gin)   │    │  Server    │                          │
│      └─────────────┘    └────────────┘                          │
└───────────────────────┼──────────────────────────────────────────┘
                        │
┌───────────────────────┼──────────────────────────────────────────┐
│               AGENT RUNTIME LAYER                                 │
├───────────────────────┼──────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────┐          │
│  │        CodeValdCortex Runtime Manager              │          │
│  └────────────────────────────────────────────────────┘          │
│          │                    │                    │              │
│  ┌───────▼───┐       ┌───────▼───┐      ┌────────▼───┐         │
│  │  Rider    │       │  Driver   │      │  Platform  │         │
│  │  Agents   │       │  Agents   │      │   Agent    │         │
│  │           │       │           │      │  (Matcher) │         │
│  └───────────┘       └───────────┘      └────────────┘         │
│                                                                   │
│  ┌───────────┐  ┌────────────┐  ┌───────────┐ ┌──────────┐    │
│  │   Route   │  │   Surge    │  │  Safety   │ │ Payment  │    │
│  │ Optimizer │  │  Pricing   │  │   Agent   │ │  Agent   │    │
│  └───────────┘  └────────────┘  └───────────┘ └──────────┘    │
│                                                                   │
│  ┌──────────────────────────────────────────────────────┐       │
│  │     Message Bus (Redis Pub/Sub + Streams)             │       │
│  └──────────────────────────────────────────────────────┘       │
└───────────────────────┼──────────────────────────────────────────┘
                        │
┌───────────────────────┼──────────────────────────────────────────┐
│                    DATA LAYER                                     │
├───────────────────────┼──────────────────────────────────────────┤
│  ┌──────────┐  ┌──────▼────┐  ┌─────────────┐  ┌────────────┐  │
│  │PostgreSQL│  │   Redis   │  │ TimescaleDB │  │  AWS S3    │  │
│  │          │  │           │  │             │  │            │  │
│  │• Users   │  │• Real-time│  │• GPS History│  │• Receipts  │  │
│  │• Trips   │  │  Matching │  │• Analytics  │  │• Photos    │  │
│  │• Ratings │  │• Cache    │  │• Predictions│  │            │  │
│  └──────────┘  └───────────┘  └─────────────┘  └────────────┘  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                  EXTERNAL INTEGRATIONS                            │
├─────────────────────────────────────────────────────────────────┤
│ Google Maps │ Twilio SMS │ Firebase │ M-Pesa │ Stripe │ AWS Rek│
└─────────────────────────────────────────────────────────────────┘
```

## Key Components

### 1. Real-Time Matching Engine
**Purpose**: Match riders with drivers in <10 seconds

**Algorithm**:
```
1. Rider request received
2. Geospatial query: Find drivers within 2km radius
3. Filter by: online, available, vehicle type match, rating ≥4.5
4. Broadcast to top 5 closest drivers (sorted by proximity)
5. First driver to accept wins
6. If no accept in 15s: Expand radius to 5km, retry
7. If still no accept: Offer surge pricing or scheduled ride
```

**Technology**: Redis geospatial indexes, WebSocket for instant notifications

### 2. Surge Pricing Engine
**Purpose**: Balance supply and demand through dynamic pricing

**Algorithm**:
```
Every 2 minutes per geographic zone:
1. Count ride requests (demand)
2. Count available drivers (supply)
3. Calculate ratio: demand / supply
4. If ratio > 1.5: Apply surge
   - 1.5-2.0 → 1.2× multiplier
   - 2.0-3.0 → 1.5× multiplier
   - 3.0+    → 2.0× multiplier
5. Notify riders of surge before booking
6. Higher fares incentivize drivers to move to surge zones
```

### 3. Safety Monitoring System
**Purpose**: Continuous trip monitoring and emergency response

**Features**:
- GPS deviation detection (>500m off route)
- Speed monitoring (>20 km/h over limit)
- Unexpected stops (>5 minutes without movement)
- Emergency SOS with <3-second response
- Automatic trusted contact notifications (night rides)

### 4. Route Optimization
**Purpose**: Calculate fastest routes considering real-time traffic

**Features**:
- A* algorithm with traffic weights from Google Maps
- Dynamic re-routing when faster route available
- ETA updates every 30 seconds
- Shared ride route optimization (multi-stop TSP)

## Data Flow: Ride Request to Completion

```
[Rider App] 
    │
    │ 1. POST /api/v1/rides (pickup, destination, ride_type)
    ▼
[REST API] → Create Ride entity (state: REQUESTED)
    │
    │ 2. Publish "ride.requested" event
    ▼
[Platform Agent]
    │
    │ 3. Geospatial query for eligible drivers
    │    Redis: GEORADIUS key longitude latitude 2000m
    ▼
[Platform Agent] → Filter by availability, vehicle_type, rating
    │
    │ 4. Broadcast to top 5 drivers via WebSocket
    ├───────┬───────┬───────┬───────┐
    ▼       ▼       ▼       ▼       ▼
[Driver1][Driver2][Driver3][Driver4][Driver5]
    │
    │ 5. First driver taps "Accept" (e.g., Driver2, 7 seconds)
    ▼
[Platform Agent] → Match confirmed
    │
    ├──────────────┬────────────────┐
    │              │                │
    ▼              ▼                ▼
[Rider App]  [Driver App]  [Payment Agent]
(Driver info) (Route nav)   (Hold payment)
    │
    │ 6. Real-time GPS tracking (every 5s)
    ▼
[WebSocket Server] → Broadcast location to rider
    │
    │ 7. Trip completion
    ▼
[Payment Agent] → Process payment
    │
    │ 8. Rating exchange
    ▼
[Analytics Agent] → Update KPIs
```

## Scalability Design

### Horizontal Scaling Targets

| Component | Initial | Scale Target | Scaling Trigger |
|-----------|---------|--------------|-----------------|
| API Servers | 3 pods | 30 pods | CPU > 70% |
| WebSocket Servers | 2 pods | 20 pods | Connections > 10,000/pod |
| Platform Agents | 5 pods | 50 pods | Matching queue > 50 |
| Surge Pricing Agents | 3 pods | 10 pods | Zone calculations > 100 |

### Database Optimization

**PostgreSQL**:
- Connection pooling (max 100 connections per instance)
- Read replicas for analytics queries
- Partitioning: Trips table by month, indexed on (rider_id, created_at)

**Redis**:
- Cluster mode with 6 nodes (3 masters, 3 replicas)
- Geospatial indexes for driver locations (TTL: 30 seconds)
- Matching queue as Redis Streams (FIFO, consumer groups)

**TimescaleDB**:
- GPS tracking compressed after 7 days (90% space savings)
- Continuous aggregates for analytics (pre-computed hourly/daily stats)

## Resilience & Fault Tolerance

### High Availability

**Multi-Region Setup**:
- Primary: Nairobi (AWS af-south-1)
- Secondary: Europe (AWS eu-west-1) for redundancy
- Active-active for API/WebSocket (latency-based routing)
- Async database replication (RPO: 5 seconds)

### Circuit Breakers

External service failures gracefully handled:
- **Google Maps down**: Use cached routes + historical traffic data
- **Payment gateway timeout**: Queue transaction for retry (up to 3 attempts)
- **SMS service failure**: Fallback to push notification only

### Disaster Recovery

- Automated backups: PostgreSQL (hourly), Redis snapshots (every 6 hours)
- Point-in-time recovery: 7-day retention
- RTO (Recovery Time Objective): < 15 minutes
- RPO (Recovery Point Objective): < 5 minutes

## Security Architecture

### Authentication
- Phone number + SMS OTP for registration
- JWT tokens (access: 15 min, refresh: 7 days)
- Driver facial recognition before each shift (AWS Rekognition)

### Data Protection
- TLS 1.3 for all communication
- AES-256 encryption at rest for PII
- GPS data anonymized after 90 days (aggregated only)

### Safety
- Driver background checks (criminal, driving record)
- Real-time trip monitoring (speed, route, stops)
- Emergency SOS with automatic authority notification
- Two-way rating system (low-rated users flagged)

## Monitoring & Observability

### Key Metrics (Prometheus)

```
# Business Metrics
rides_requested_total
rides_matched_total{outcome="success|timeout|cancelled"}
rides_completed_total{on_time="true|false"}
surge_pricing_active{zone}

# Technical Metrics
matching_duration_seconds{percentile="p50|p95|p99"}
gps_update_latency_milliseconds
payment_success_rate
api_response_time_seconds{endpoint}

# Safety Metrics
emergency_sos_activated_total
trip_anomalies_detected_total{type="speed|deviation|stop"}
```

### Alerting

**Critical** (PagerDuty):
- Service down > 2 minutes
- Matching success rate < 90%
- Payment failure rate > 5%
- Emergency SOS not processed within 3 seconds

**Warning** (Slack):
- API latency P95 > 500ms
- GPS update latency > 200ms
- Database connection pool > 80%

## Technology Decisions

### Why Go?
- **Concurrency**: Goroutines ideal for 50,000+ concurrent rides
- **Performance**: Low latency for real-time matching
- **Deployment**: Single binary, fast startup

### Why Redis for Matching?
- **Geospatial Indexing**: Built-in GEORADIUS for proximity queries
- **Speed**: Sub-millisecond lookups
- **Pub/Sub**: Real-time driver notifications

### Why React Native?
- **Code Sharing**: 80% code shared iOS/Android
- **Performance**: Near-native UI responsiveness
- **Maps Integration**: Excellent Google Maps SDK support

## Future Enhancements

### Phase 2: Advanced Features
1. **AI-Powered Demand Prediction**: Pre-position drivers in predicted high-demand zones
2. **Carpool/Ride-Sharing**: Multi-passenger rides with optimized routing
3. **Electric Vehicle Support**: EV charging station navigation, range anxiety mitigation
4. **Accessibility Features**: Wheelchair-accessible vehicles, hearing/vision impaired support

### Phase 3: Ecosystem
1. **Food Delivery Integration**: Leverage driver network for last-mile delivery
2. **Package Delivery**: Small parcel delivery using ride-hailing network
3. **Corporate Accounts**: Business travel with invoicing and reporting
4. **API Platform**: Third-party integrations (hotels, airlines, events)

## Related Documents

- [README](./README.md) - Use case overview
- [Agent Design](./agent-design.md) - Detailed agent specifications
- [Use Case Specification](../../../1-SoftwareRequirements/requirements/use-cases/UC-RIDE-001-ride-hailing-platform.md)

---

**Maintained by**: Architecture Team  
**Review Cadence**: Weekly during design phase  
**Next Review**: October 29, 2025
