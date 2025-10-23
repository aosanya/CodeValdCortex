# UC-TRACK-001: Safiri Salama - Safe Journey Tracking System

## Overview

**Safiri Salama** (Swahili for "Safe Journey") is a comprehensive geo-location tracking and decision support system for transportation safety and passenger convenience. The system serves two primary contexts:

1. **School Transportation**: Real-time tracking of school buses for enhanced child safety and parent peace of mind
2. **Public Transport (Matatus)**: Customer loyalty through favorite vehicle tracking and smart arrival notifications

## Key Innovation

This use case introduces **Agent Property Auto-Broadcasting** - a framework capability allowing agents to automatically publish selected properties (GPS location, status, metrics) at configurable intervals with intelligent context-aware frequency adjustment.

## Use Case Structure

```
UC-TRACK-001-safiri-salama/
├── README.md                      # This file
├── .env                           # Environment configuration
├── start.sh                       # Deployment script
└── config/
    └── agents/                    # Agent type JSON schemas
        ├── vehicle.json           # Vehicle Agent (bus/matatu)
        ├── parent.json            # Parent/Guardian Agent
        ├── passenger.json         # Passenger/Commuter Agent
        ├── route_manager.json     # Route Manager Agent
        └── fleet_operator.json    # Fleet Operator Agent
```

## Agent Types

### 1. Vehicle Agent (`vehicle.json`)
- **Purpose**: Represents physical vehicles (school buses or matatus)
- **Key Features**:
  - GPS tracking with context-aware broadcasting
  - Autonomous interval adjustment based on operational status
  - Privacy controls and geofencing
  - Passenger capacity tracking
  - Rating and loyalty support

### 2. Parent/Guardian Agent (`parent.json`)
- **Purpose**: Parents tracking their child's school bus
- **Key Features**:
  - Subscribe to specific school buses
  - Multi-channel notifications (push, SMS, email)
  - Proximity alerts when bus approaching
  - Emergency alert handling
  - Child pickup/dropoff tracking

### 3. Passenger/Commuter Agent (`passenger.json`)
- **Purpose**: Regular commuters tracking favorite matatus
- **Key Features**:
  - Mark up to 10 favorite matatus
  - Smart notifications when favorites approaching
  - Commute pattern learning
  - Trip ratings and feedback
  - Loyalty points and badges

### 4. Route Manager Agent (`route_manager.json`)
- **Purpose**: Fleet coordinator monitoring all vehicles
- **Key Features**:
  - Real-time fleet status monitoring
  - Anomaly detection (speeding, route deviation, prolonged stops)
  - Route optimization based on traffic
  - Emergency response coordination
  - Geofence management

### 5. Fleet Operator Agent (`fleet_operator.json`)
- **Purpose**: Oversight and administration (School Admin or SACCO Manager)
- **Key Features**:
  - Fleet-wide analytics and reporting
  - Compliance monitoring
  - Subscriber management
  - Loyalty program configuration (matatus)
  - Revenue analytics (matatus)
  - Incident management

## Quick Start

### Prerequisites

1. **ArangoDB** running on localhost:8529
2. **NATS** message broker running on localhost:4222
3. **Go 1.21+** installed
4. **CodeValdCortex** framework built

### Setup

1. **Configure Environment**:
   ```bash
   cd usecases/UC-TRACK-001-safiri-salama
   cp .env.example .env  # If using example
   # Edit .env with your configuration
   ```

2. **Set Required Environment Variables**:
   ```bash
   # SMS Provider (Africa's Talking)
   export SMS_API_KEY=your_api_key
   export SMS_USERNAME=your_username
   
   # Push Notifications
   export FCM_SERVER_KEY=your_fcm_key
   export APNS_KEY_ID=your_apns_key
   export APNS_TEAM_ID=your_team_id
   
   # Email
   export EMAIL_USERNAME=your_email
   export EMAIL_PASSWORD=your_password
   
   # Security
   export JWT_SECRET=your_secret_key
   
   # External Services
   export GOOGLE_MAPS_API_KEY=your_maps_key
   export TRAFFIC_API_KEY=your_traffic_key
   ```

3. **Start the System**:
   ```bash
   chmod +x start.sh
   ./start.sh
   ```

### Docker Compose Setup

```bash
# Start infrastructure services
docker-compose up -d arangodb nats

# Wait for services to be ready
sleep 10

# Start Safiri Salama
./start.sh
```

## Configuration

### Broadcasting Intervals

The system uses context-aware broadcasting with the following default intervals:

| Context              | Interval | Priority  | Use Case                           |
|----------------------|----------|-----------|------------------------------------|
| Emergency            | 3s       | Critical  | Maximum visibility during incidents |
| At Stop              | 5s       | High      | Precise updates during boarding     |
| Approaching Stop     | 10s      | High      | Alert passengers/parents            |
| En Route             | 30s      | Normal    | Regular tracking during transit     |
| Highway              | 60s      | Normal    | Efficient battery/bandwidth usage   |
| Privacy Mode         | None     | -         | Driver break periods                |

### Subscription Modes

#### School Context
- **Approval Required**: Parents must be pre-enrolled by school
- **Public Tracking**: Disabled (privacy for children)
- **Emergency Alerts**: Always notify all subscribed parents
- **Max Silence**: 5 minutes (strict safety requirement)

#### Matatu Context
- **Approval Required**: No (open subscription)
- **Public Tracking**: Enabled (anyone can track)
- **Favorite Notifications**: Enabled for priority alerts
- **Max Silence**: 30 minutes (driver break allowance)
- **Loyalty Features**: Ratings, favorites, badges

## API Endpoints

### Broadcasting Management
```
POST   /api/v1/agents/{agentId}/broadcasting/configure
POST   /api/v1/agents/{agentId}/broadcasting/start
POST   /api/v1/agents/{agentId}/broadcasting/stop
POST   /api/v1/agents/{agentId}/broadcasting/pause
PUT    /api/v1/agents/{agentId}/broadcasting/interval
GET    /api/v1/agents/{agentId}/broadcasting/metrics
```

### Subscription Management
```
POST   /api/v1/agents/{agentId}/broadcasting/subscribe
DELETE /api/v1/agents/{agentId}/broadcasting/unsubscribe
POST   /api/v1/agents/{agentId}/broadcasting/favorite
DELETE /api/v1/agents/{agentId}/broadcasting/favorite
GET    /api/v1/agents/{agentId}/broadcasting/subscribers
```

### Vehicle Operations
```
POST   /api/v1/vehicles
GET    /api/v1/vehicles/{vehicleId}
PUT    /api/v1/vehicles/{vehicleId}/location
PUT    /api/v1/vehicles/{vehicleId}/status
POST   /api/v1/vehicles/{vehicleId}/emergency
```

## Testing

### Unit Tests
```bash
go test ./internal/communication/... -v
go test ./internal/agent/... -v
```

### Integration Tests
```bash
go test ./test/integration/uc_track_001_test.go -v
```

### Load Testing
```bash
# Test with 1000 vehicles broadcasting
go test ./test/load/broadcasting_load_test.go -v -count=1
```

## Pilot Program

### School Pilot
- **Target**: 2-3 schools
- **Vehicles**: 5-10 buses per school
- **Subscribers**: 200-300 parents
- **Duration**: 4 weeks
- **Success Metrics**: >70% active usage, <5% error rate

### Matatu Pilot
- **Target**: 1-2 SACCOs
- **Vehicles**: 10-15 trackable matatus
- **Subscribers**: 500-1000 passengers
- **Duration**: 4 weeks
- **Success Metrics**: >50 favorites per vehicle, >4.0 rating

## Business Model

### School Context
- **Revenue**: Monthly subscription per school
- **Pricing**: Based on fleet size (per bus)
- **Value Prop**: Safety, compliance, parent satisfaction

### Matatu Context
- **Free Tier**: Track up to 3 favorite vehicles
- **Premium Tier**: KES 99/month for unlimited favorites + advanced features
- **SACCO Partnership**: Matatu owners pay for "Verified Trackable" badge
- **Value Prop**: Customer loyalty, competitive differentiation

## Monitoring & Analytics

### Dashboards

1. **Fleet Operations Dashboard**:
   - Real-time vehicle map
   - Fleet status overview
   - Active alerts and incidents
   - Performance metrics

2. **Loyalty Analytics Dashboard** (Matatus):
   - Top-rated vehicles
   - Favorite count distribution
   - Passenger engagement metrics
   - Revenue insights

3. **Safety Compliance Dashboard** (Schools):
   - Speed violations
   - Route adherence
   - Emergency response times
   - Parent satisfaction scores

### Metrics Collection

```yaml
# Prometheus metrics exposed on :9091/metrics
- safiri_broadcasts_total
- safiri_broadcast_interval_seconds
- safiri_subscribers_total
- safiri_favorites_total
- safiri_trips_rated_total
- safiri_emergency_alerts_total
- safiri_route_deviations_total
```

## Documentation

- **Use Case Document**: `documents/1-SoftwareRequirements/requirements/use-cases/UC-TRACK-001-safiri-salama.md`
- **Framework Documentation**: `documents/3-SofwareDevelopment/core-systems/agent-property-broadcasting.md`
- **MVP Tasks**: `documents/3-SofwareDevelopment/mvp.md` (MVP-016 to MVP-020)

## Support & Contact

For issues, questions, or pilot program inquiries:
- **Email**: support@safirisalama.com
- **Documentation**: https://docs.safirisalama.com
- **GitHub Issues**: https://github.com/aosanya/CodeValdCortex/issues

## License

See LICENSE file in project root.

---

**Status**: Concept/Planning  
**Framework Version**: 1.0.0  
**Created**: October 23, 2025  
**Next Steps**: See MVP-016 through MVP-020 in development roadmap
