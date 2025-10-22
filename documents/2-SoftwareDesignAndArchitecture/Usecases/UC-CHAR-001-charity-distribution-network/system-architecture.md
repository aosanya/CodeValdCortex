# System Architecture - Charity Distribution Network (DignityNet)

**Version**: 1.0  
**Last Updated**: October 22, 2025

## Architecture Overview

DignityNet is designed to connect donors with recipients while preserving dignity, ensuring transparency, and optimizing logistics for charity distribution at scale.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Client Applications                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │ Donor Web    │  │ Recipient    │  │  Volunteer   │                 │
│  │ Portal       │  │  Mobile App  │  │   Tablet     │                 │
│  │ (React)      │  │(React Native)│  │    App       │                 │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                 │
└─────────┼──────────────────┼──────────────────┼───────────────────────┘
          │                  │                  │
          └──────────────────┴──────────────────┘
                             │ (HTTPS/WebSocket)
┌─────────────────────────────────────────────────────────────────────────┐
│                         API Gateway Layer                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │  API         │  │  WebSocket   │  │  Payment     │                 │
│  │  Gateway     │  │  Gateway     │  │  Gateway     │                 │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                 │
└─────────┼──────────────────┼──────────────────┼───────────────────────┘
          │                  │                  │
          └──────────────────┴──────────────────┘
                             │
┌─────────────────────────────────────────────────────────────────────────┐
│                    Application Layer (CodeValdCortex)                    │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                  Agent Runtime Manager                            │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────────┐    │
│  │  Donor     │ │ Recipient  │ │   Item     │ │   Volunteer    │    │
│  │  Agents    │ │  Agents    │ │  Agents    │ │    Agents      │    │
│  └────────────┘ └────────────┘ └────────────┘ └────────────────┘    │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────────┐    │
│  │ Logistics  │ │  Storage   │ │    Need    │ │    Impact      │    │
│  │Coordinator │ │  Facility  │ │  Matcher   │ │    Tracker     │    │
│  └────────────┘ └────────────┘ └────────────┘ └────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                             │
┌─────────────────────────────────────────────────────────────────────────┐
│                         Service Layer                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │ Communication│  │  Task        │  │  Memory      │                 │
│  │ Service      │  │  System      │  │  Service     │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │Configuration │  │  Matching    │  │  Routing     │                 │
│  │Service       │  │  Engine      │  │  Engine      │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
└─────────────────────────────────────────────────────────────────────────┘
                             │
┌─────────────────────────────────────────────────────────────────────────┐
│                         Data Layer                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │ PostgreSQL   │  │    Neo4j     │  │    Redis     │                 │
│  │ (Donors,     │  │  (Donation   │  │  (Pub/Sub,   │                 │
│  │  Items)      │  │   Network)   │  │   Cache)     │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │    S3        │  │  TimescaleDB │  │  Blockchain  │                 │
│  │ (Images,     │  │  (Impact     │  │ (Transaction │                 │
│  │  Documents)  │  │   Metrics)   │  │    Ledger)   │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
└─────────────────────────────────────────────────────────────────────────┘
                             │
┌─────────────────────────────────────────────────────────────────────────┐
│                    External Integrations                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │   Payment    │  │     SMS      │  │   Identity   │                 │
│  │ Processing   │  │Notification  │  │Verification  │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │   Mapping    │  │    Tax       │  │   Social     │                 │
│  │   Services   │  │ Receipt API  │  │   Services   │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
└─────────────────────────────────────────────────────────────────────────┘
```

## Component Descriptions

### Client Layer

**Donor Web Portal (React)**
- Donation management
- Impact dashboard
- Tax receipt downloads
- Recurring donation setup
- Organization profiles

**Recipient Mobile App (React Native)**
- Simple, accessible UI (multilingual, low-literacy design)
- Browse available items
- Request items anonymously
- Schedule pickup
- Provide feedback

**Volunteer Tablet App (React Native)**
- Inventory management
- Check-in/check-out
- Quality verification
- Route optimization
- Delivery confirmation

### API Gateway Layer

**API Gateway**
- RESTful endpoints for CRUD operations
- GraphQL for complex queries (donor impact reports)
- JWT authentication
- Rate limiting per client type

**WebSocket Gateway**
- Real-time inventory updates
- Live donation matching notifications
- Delivery status updates
- Volunteer coordination

**Payment Gateway**
- PCI-compliant payment processing
- Multiple payment methods (credit card, PayPal, crypto)
- Recurring donations
- Tax receipt generation

### Application Layer (CodeValdCortex)

**Agent Runtime Manager**
- Manages lifecycle of 8 agent types
- Dynamic scaling based on donation volume
- Resource allocation
- Health monitoring

**Agent Types** (8 types - see agent-design.md for details):
1. Donor Agent
2. Recipient Agent
3. Item Agent
4. Volunteer Agent
5. Logistics Coordinator Agent
6. Storage Facility Agent
7. Need Matcher Agent
8. Impact Tracker Agent

### Service Layer

**Communication Service**
- Redis Pub/Sub for agent communication
- Event broadcasting (new donation, item matched)
- SMS notifications for recipients (Twilio)
- Email notifications for donors

**Task System**
- Scheduled matching runs (every 15 minutes)
- Recurring donation processing
- Impact report generation (monthly)
- Inventory audits (weekly)

**Matching Engine**
- Multi-criteria matching algorithm:
  - Need urgency (1-10 scale)
  - Item condition (new, like-new, good, fair)
  - Geographic proximity (<10 miles preferred)
  - Dietary restrictions (for food)
  - Size/quantity requirements
  - Cultural preferences
- Machine learning model for optimization
- Priority given to urgent needs (medical, food insecurity)

**Routing Engine**
- Vehicle routing problem (VRP) solver
- Pickup and delivery optimization
- Time window constraints
- Volunteer availability
- Traffic-aware (Google Maps API)
- Generates efficient routes for multiple stops

**Configuration Service**
- Matching criteria weights
- Urgency thresholds
- Geographic boundaries
- Operating hours per facility

### Data Layer

**PostgreSQL**
- Donor profiles and payment info
- Recipient profiles (anonymized)
- Item catalog
- Donation transactions
- Volunteer records
- Storage facility inventory

**Neo4j (Graph Database)**
- Donation network (donor → item → recipient)
- Impact chains (who helped whom)
- Volunteer networks
- Facility relationships
- Pattern detection (frequent donors, bottlenecks)

**Redis**
- Real-time inventory cache
- Active delivery tracking
- Session management
- Pub/Sub for agent communication
- Geospatial indexing (GEORADIUS for proximity matching)

**S3 (Object Storage)**
- Item photos
- Proof of delivery (signed receipts)
- Tax receipt PDFs
- Impact report visualizations

**TimescaleDB**
- Donation volumes over time
- Impact metrics (meals served, families housed)
- Facility utilization
- Volunteer hours
- Response times (need → fulfillment)

**Blockchain (Hyperledger Fabric)**
- Immutable donation ledger
- Transparent fund allocation
- Proof of delivery
- Audit trail for compliance
- Smart contracts for recurring donations

### External Integrations

**Payment Processing**
- Stripe for credit/debit cards
- PayPal for alternative payments
- Cryptocurrency wallets (Bitcoin, Ethereum)

**SMS Notification (Twilio)**
- Recipient notifications (item available, pickup ready)
- Volunteer dispatch alerts
- Donor impact updates (optional)

**Identity Verification**
- Government ID scan (optional, for high-value items)
- Phone number verification (required for recipients)
- Email verification (required for donors)

**Mapping Services**
- Google Maps API for geocoding
- Distance matrix for proximity calculations
- Traffic data for route optimization

**Tax Receipt API**
- IRS-compliant receipt generation
- Fair market value estimation
- Donation aggregation (annual summary)

**Social Services Integration**
- SNAP/EBT eligibility verification (optional)
- Homeless shelter databases
- Food bank networks
- Government assistance programs

## Data Flow Patterns

### 1. Donation Flow

```
Donor creates donation → Donor Agent created
→ Item Agents created (one per item)
→ Photos uploaded to S3 → Quality verification
→ Storage Facility Agent assigns location
→ Item status: AVAILABLE
→ Need Matcher Agent triggered
→ Matching algorithm runs → Recipient found
→ Recipient Agent notified → Recipient accepts
→ Logistics Coordinator plans pickup
→ Volunteer assigned → Delivery scheduled
→ Delivery completed → Proof uploaded
→ Impact Tracker records → Donor notified
→ Blockchain ledger updated
```

**Latency Target**: Need fulfillment within 48 hours (urgent), 7 days (standard)

### 2. Recipient Request Flow

```
Recipient creates need → Recipient Agent created
→ Urgency score calculated → Need Matcher searches
→ If match found: Recipient notified immediately
→ If no match: Need queued → Donor Agent notified (matching profile)
→ Donor responds → Donation flow initiated
```

**Latency Target**: Match notification within 15 minutes if item available

### 3. Matching Algorithm

```
Need Matcher triggers every 15 minutes
→ Fetch unmatched recipients (sorted by urgency)
→ For each recipient:
   - Fetch available items (filter by category, location)
   - Score each item (urgency × proximity × condition)
   - Select best match (threshold: 70%)
   - Create tentative match
→ Notify Recipient Agent (24-hour acceptance window)
→ If accepted: Mark item RESERVED, trigger logistics
→ If declined/expired: Return item to pool, try next recipient
```

**Performance**: Process 1000 matches in <60 seconds

### 4. Logistics Routing

```
Logistics Coordinator fetches today's deliveries
→ Group by geographic region (grid-based)
→ Assign to available volunteers (capacity, vehicle type)
→ For each volunteer:
   - Solve VRP (pickup → recipient locations)
   - Generate optimized route (minimize time/distance)
   - Consider time windows (facility hours, recipient availability)
→ Send route to Volunteer Agent → Navigate with turn-by-turn
→ Real-time updates (traffic, delays)
→ Delivery confirmed → Update all related agents
```

**Optimization Goal**: Minimize volunteer driving time, maximize deliveries per trip

## Scalability Design

### Horizontal Scaling

**Stateless Components** (scale out):
- API Gateway (add instances behind load balancer)
- Agent Runtime (add worker nodes)
- Matching Engine (parallel processing by region)

**Stateful Components** (partitioned):
- PostgreSQL: Shard by geographic region
- Neo4j: Shard by subgraph (region-based)
- Redis: Redis Cluster with hash slots

### Partitioning Strategy

**Geographic Partitioning**:
- Divide service area into regions (e.g., 10-mile grid)
- Each region has dedicated:
  - Database partition
  - Agent pool
  - Storage facilities
- Cross-region transfers only for rare items

**Time-Based Partitioning**:
- Active donations: Hot database
- Completed donations (>6 months): Cold storage

### Performance Targets

| Metric | Target | Scale |
|--------|--------|-------|
| Concurrent Donors | 10K+ | Platform-wide |
| Concurrent Recipients | 50K+ | Platform-wide |
| Donations per Day | 100K+ | Peak (holidays) |
| Matching Latency | <15 min | Per matching cycle |
| API Response Time | <300ms | P95 |
| Database Queries | <100ms | P95 |
| Delivery Route Calculation | <30s | 20-stop route |

## Resilience and Fault Tolerance

### High Availability

**Database Replication**:
- PostgreSQL: Primary + 2 replicas (async streaming)
- Neo4j: 3-node cluster (leader election)
- Redis: Redis Sentinel (automatic failover)

**Application Redundancy**:
- Kubernetes with 3+ replicas per service
- Multi-AZ deployment
- Health checks and auto-restart

### Disaster Recovery

**Backup Strategy**:
- PostgreSQL: WAL archiving + daily full backups
- Neo4j: Daily graph snapshots
- Blockchain: Replicated ledger (inherent redundancy)

**Recovery Time Objective (RTO)**: 2 hours  
**Recovery Point Objective (RPO)**: 15 minutes

### Graceful Degradation

**If Matching Engine fails**:
- Manual matching by staff (temporary)
- Queue requests for auto-matching when service recovers

**If Blockchain fails**:
- Continue operations (write to PostgreSQL)
- Sync to blockchain when service recovers
- Maintain audit trail in traditional database

**If Payment Gateway fails**:
- Queue donations (process when service recovers)
- Notify donors of delay (email)

## Security Architecture

### Authentication & Authorization

**Donor Authentication**:
- Email/password with 2FA
- OAuth (Google, Facebook)
- Session timeout: 30 days

**Recipient Authentication**:
- Phone number verification (SMS code)
- Anonymous ID generation (privacy protection)
- No email required (accessibility)

**Volunteer Authentication**:
- Background check required
- Biometric login (fingerprint) for tablets
- Active session monitoring

**Authorization Roles**:
- Donor: View own donations, impact reports
- Recipient: Browse items, make requests
- Volunteer: Delivery tasks, inventory management
- Facility Manager: Local inventory, volunteer coordination
- System Admin: Full access

### Data Protection

**Encryption**:
- In Transit: TLS 1.3 for all connections
- At Rest: AES-256 for databases
- End-to-End: Recipient identity encrypted

**Privacy (Recipient Protection)**:
- Anonymous IDs (no real names exposed to donors)
- Geolocation obscured (zip code only)
- Photo opt-in (for impact stories)
- Data retention: 90 days post-delivery, then anonymized

**PCI Compliance**:
- No card data stored (tokenized via Stripe)
- PCI-DSS Level 1 compliance
- Annual security audits

### Content Security

**Fraud Prevention**:
- Donor: Verify email, monitor for duplicate accounts
- Recipient: Phone verification, rate limiting (1 request per item type per 30 days)
- Volunteer: Background check, delivery photo required

**Quality Control**:
- Item photos required
- Facility staff verification
- Recipient feedback (quality rating)

## Monitoring and Observability

### Metrics (Prometheus)

**System Metrics**:
- API latency, throughput
- Database connection pool, query time
- Cache hit rate
- Agent instance count

**Operational Metrics**:
- Donations received (count, value)
- Match rate (% of needs fulfilled)
- Delivery success rate
- Volunteer utilization
- Facility capacity

**Impact Metrics**:
- Families served
- Meals provided
- Tons of CO2 saved (reuse vs landfill)
- Economic value distributed

### Alerting

**Critical Alerts** (PagerDuty):
- Payment gateway down
- Database primary failure
- Matching engine stopped

**Warning Alerts** (Slack):
- Match rate below 80%
- Delivery delayed > 24 hours
- Facility over capacity

**Impact Alerts** (Email to leadership):
- Weekly impact summary
- Low inventory warning (specific categories)
- New high-value donor

## Deployment Strategy

### Kubernetes Architecture

```yaml
Namespaces:
  - production
  - staging

Deployments:
  - api-gateway (3 replicas)
  - agent-runtime (5 replicas)
  - matching-engine (2 replicas)
  - routing-engine (2 replicas)

StatefulSets:
  - postgresql (1 primary + 2 replicas)
  - neo4j (3 replicas)
  - redis-cluster (6 replicas)

CronJobs:
  - matching-job (every 15 minutes)
  - impact-report (daily at midnight)
  - inventory-audit (weekly)
```

### CI/CD Pipeline

```
Code Push → GitHub
  ↓
Tests (unit + integration)
  ↓
Build Docker images
  ↓
Deploy to Staging
  ↓
E2E tests (automated matching, donation flow)
  ↓
Manual approval (product manager)
  ↓
Deploy to Production (canary: 10% → 50% → 100%)
  ↓
Monitor key metrics (30 minutes)
  ↓
Rollback if match rate drops or errors spike
```

## Technology Decisions

### Why Neo4j over ArangoDB?
- Superior graph query performance for impact chains
- Better community support for non-profit use cases
- Visual browser for stakeholder demos

### Why Blockchain?
- Transparency for donors (see where money goes)
- Immutable audit trail (regulatory compliance)
- Trust building for new non-profit
- Hyperledger Fabric chosen for private consortium (not public)

### Why Multilingual Support?
- Recipient app in 10+ languages (Spanish, Arabic, Chinese, etc.)
- Low-literacy mode (icons, voice navigation)
- Cultural sensitivity (dietary restrictions, modesty)

### Why Dignity-Centered Design?
- Anonymous recipient IDs (no exposure to donors)
- Choice-based system (recipients choose, not assigned)
- Feedback loop (recipients rate quality, inform system)
- No stigma (looks like regular shopping app)

## Future Enhancements

### Phase 2 (Months 13-18)
- AI-powered need prediction (seasonal trends, economic indicators)
- Barcode/QR scanning for faster item entry
- Integration with corporate donation programs
- Volunteer mobile app (iOS/Android native)

### Phase 3 (Months 19-24)
- Drone delivery for urgent items (pilot program)
- Blockchain-based donor loyalty program
- Predictive matching (notify donors before items donated)
- Impact stories (automated generation from data)

## Related Documents

- [Agent Design](./agent-design.md)
- [Matching Algorithm](./matching-algorithm.md)
- [Data Models](./data-models.md)
- [Security Design](./security-design.md)
- [Use Case Specification](../../../1-SoftwareRequirements/requirements/use-cases/UC-CHAR-001-charity-distribution-network.md)
