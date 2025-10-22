# UC-LOG-001: Smart Logistics Service Platform

**Use Case ID**: UC-LOG-001  
**System Name**: CodeValdUsafiri ("Transport/Logistics" in Swahili)  
**Domain**: Logistics & Transportation  
**Version**: 1.0  
**Status**: Design Phase  
**Last Updated**: October 22, 2025

## Overview

CodeValdUsafiri is an intelligent logistics platform where shippers (businesses/individuals needing transport), drivers/trucks (service providers), and facilities (warehouses, depots, terminals) operate as autonomous agents that negotiate, reconcile conflicts, and optimize deliveries through intelligent multi-criteria decision-making.

The platform enables:
- **Dynamic Service Summoning**: Shippers broadcast needs; eligible trucks autonomously bid
- **Intelligent Reconciliation**: Platform agents rank bids using multi-criteria scoring (price, time, reliability, experience, vehicle quality)
- **Facility Coordination**: Automated dock appointment scheduling and capacity management
- **Real-Time Optimization**: Route optimization, return cargo matching, emergency re-routing
- **End-to-End Visibility**: GPS tracking, dynamic ETAs, automated status updates, digital documentation

Unlike traditional logistics marketplaces with manual matching, CodeValdUsafiri leverages agent autonomy and machine intelligence to create a self-organizing system that balances efficiency, cost, reliability, and sustainability.

## Agent Types

The system consists of **8 autonomous agent types**:

1. **Shipper Agent** - Businesses/individuals requesting transport services
2. **Driver/Truck Agent** - Service providers with vehicles executing deliveries
3. **Facility Agent** - Warehouses, depots, terminals managing loading/unloading
4. **Platform Agent** - System orchestrator managing broadcasts and reconciliation
5. **Route Optimizer Agent** - Calculates optimal routes considering traffic, multi-stop, return cargo
6. **Payment Agent** - Processes transactions, manages escrow, handles disputes
7. **Conflict Resolver Agent** - Mediates disputes, manages emergency situations
8. **Analytics Agent** - Tracks performance, identifies patterns, provides insights

Each agent operates autonomously with local decision-making while collaborating through message passing to achieve system-wide optimization.

## Key Design Principles

1. **Agent Autonomy with Intelligent Reconciliation**
   - Truck agents independently evaluate opportunities and submit bids
   - Platform agent uses multi-criteria scoring to rank options
   - Shipper maintains final authority while benefiting from intelligent recommendations

2. **Multi-Stakeholder Coordination**
   - Three-way coordination between shippers, drivers, and facilities
   - Automated dock scheduling prevents waiting time and bottlenecks
   - Real-time status synchronization across all parties

3. **Economic Optimization**
   - Return cargo matching reduces empty miles
   - Dynamic pricing reflects supply/demand
   - Route optimization minimizes fuel costs

4. **Transparency and Trust**
   - Digital documentation (Bill of Lading, Proof of Delivery)
   - Photo verification at pickup and delivery
   - Mutual rating system builds reputation

5. **Resilience and Adaptability**
   - Emergency re-routing for weather, breakdowns, accidents
   - Automatic re-assignment if driver unavailable
   - Graceful degradation (manual fallbacks available)

## Technology Stack

### Runtime & Core Framework
- **CodeValdCortex Framework** - Agent runtime, communication, lifecycle management
- **Go 1.21+** - Primary implementation language
- **PostgreSQL 15** - Agent state, bookings, transactions
- **Redis 7** - Message broker, real-time tracking cache, session management
- **TimescaleDB** - GPS tracking history, performance metrics

### External Integrations
- **Google Maps API** - Geocoding, routing, traffic data, ETA calculations
- **Twilio** - SMS notifications for critical events
- **SendGrid** - Email notifications and invoicing
- **M-Pesa/Stripe** - Payment processing
- **AWS S3** - Document and photo storage

### AI/ML Components
- **Route Optimization**: Custom algorithms (Dijkstra's + A* with traffic weights)
- **Bid Ranking**: Multi-criteria decision analysis (MCDA)
- **Demand Prediction**: Time series forecasting (ARIMA, Prophet)
- **Anomaly Detection**: GPS deviation, fraud detection

### Frontend
- **React Native** - Mobile apps (iOS/Android) for drivers and shippers
- **React** - Web portal for facilities and administrators
- **WebSocket** - Real-time GPS tracking, status updates

### Infrastructure
- **Kubernetes** - Container orchestration
- **Docker** - Containerization
- **Prometheus/Grafana** - Monitoring and alerting
- **ELK Stack** - Centralized logging

## Performance Targets

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| **Bid Collection Time** | 95% within 120 seconds | Time from broadcast to minimum 3 bids |
| **GPS Update Frequency** | Every 5 seconds during transit | WebSocket message rate |
| **ETA Accuracy** | ±10 minutes for 80% of deliveries | Predicted vs actual arrival time |
| **Platform Uptime** | 99.9% (8.76 hours/year downtime) | Uptime monitoring service |
| **API Response Time** | P95 < 500ms, P99 < 1s | Application performance monitoring |
| **Concurrent Shipments** | 10,000+ simultaneous active shipments | Load testing |
| **Driver App Performance** | < 100ms location update processing | Mobile telemetry |
| **Payment Processing** | 99.5% success rate | Transaction success monitoring |
| **Emergency Re-routing** | < 60 seconds to calculate alternate route | Route optimizer agent metrics |

## Use Case Scenarios

### Scenario 1: Standard Single-Stop Delivery
**Trigger**: Shipper creates delivery request from Nairobi warehouse to Mombasa depot

**Flow**:
1. Shipper Agent broadcasts request (pickup: 2 hours, delivery: next day)
2. 8 eligible Truck Agents evaluate and submit bids (30-120 seconds)
3. Platform Agent ranks bids; presents top 5 to shipper
4. Shipper selects bid based on price and reliability
5. Facility Agent at warehouse schedules dock appointment
6. Truck Agent arrives, loads cargo, departs (digital BoL signed)
7. Real-time tracking en route (470 km, ~8 hours)
8. Facility Agent at Mombasa schedules unloading dock
9. Delivery completed, digital PoD signed
10. Payment released to driver after 24-hour dispute window

**Outcome**: On-time delivery, 98% shipper satisfaction, driver earns return cargo opportunity

### Scenario 2: Multi-Stop Delivery with Route Optimization
**Trigger**: Shipper needs deliveries to 5 locations in Nairobi

**Flow**:
1. Route Optimizer Agent calculates optimal sequence considering:
   - Delivery time windows
   - Traffic patterns (morning rush hour)
   - Cargo compatibility
2. Platform broadcasts multi-stop request
3. Truck Agents bid with time estimates for optimized route
4. Winning driver follows turn-by-turn navigation
5. Driver completes all 5 stops in optimal order
6. System tracks completion at each stop with photo verification

**Outcome**: 40% time savings vs manual routing, all deliveries within time windows

### Scenario 3: Emergency Re-routing Due to Weather
**Trigger**: Flash flood blocks highway during active shipment

**Flow**:
1. Route Optimizer Agent detects road closure via traffic API
2. System calculates alternate route (+45 minutes)
3. Conflict Resolver Agent notifies consignee of delay
4. Driver follows updated navigation automatically
5. System updates ETA across all stakeholders
6. Delivery completed with documented delay reason

**Outcome**: Zero cancellations, preserved customer satisfaction despite force majeure

### Scenario 4: Return Cargo Matching
**Trigger**: Driver completes delivery in Mombasa, base is Nairobi

**Flow**:
1. After delivery, Truck Agent queries Platform for return opportunities
2. Platform finds shipment: Mombasa → Nairobi, compatible cargo
3. Driver accepts return cargo (earns additional 60% of base fare)
4. System updates route; driver loads return cargo
5. Driver returns to base with revenue-generating load

**Outcome**: 60% reduction in empty miles, improved driver earnings

## Security & Privacy

### Authentication & Authorization
- **Multi-Factor Authentication (MFA)** for all user types
- **OAuth 2.0** for API access
- **Role-Based Access Control (RBAC)**: Shipper, Driver, Facility Manager, Admin roles
- **JWT tokens** with short expiration (15 minutes) and refresh tokens

### Data Protection
- **Encryption at rest**: AES-256 for all sensitive data (PII, financial)
- **Encryption in transit**: TLS 1.3 for all API communication
- **GPS data anonymization**: Historical data aggregated and anonymized after 90 days
- **PII access logging**: All access to personal data logged and audited

### Payment Security
- **PCI DSS compliance** for card processing
- **Tokenization**: Credit card numbers never stored; tokens only
- **Escrow system**: Funds held until delivery confirmed
- **Fraud detection**: ML-based anomaly detection for suspicious transactions

### Operational Security
- **Background checks**: All drivers verified (criminal record, driving history)
- **Vehicle inspection**: Mandatory annual safety inspections
- **Insurance verification**: Continuous monitoring of policy validity
- **Incident response**: 24/7 security team for emergencies

## Integration Points

### External Systems
1. **Payment Gateways**
   - M-Pesa (Kenya mobile money)
   - Stripe (international cards)
   - Bank transfers (SWIFT/SEPA)

2. **Mapping & Navigation**
   - Google Maps Platform (geocoding, routing, traffic)
   - Mapbox (alternative for offline maps)

3. **Communication**
   - Twilio (SMS for critical alerts)
   - SendGrid (email notifications)
   - Firebase Cloud Messaging (push notifications)

4. **Document Management**
   - AWS S3 (BoL, PoD, photos, invoices)
   - DocuSign (digital signatures)

5. **Facility Management Systems**
   - Warehouse Management Systems (WMS) via REST APIs
   - Dock scheduling systems (custom integrations)

6. **Compliance & Regulatory**
   - KRA (Kenya Revenue Authority) for tax reporting
   - NTSA (National Transport and Safety Authority) for vehicle compliance
   - Customs systems for cross-border shipments

## Success Metrics

### Technical Metrics
- **System Uptime**: 99.9%+ (target: 99.95%)
- **API Response Time**: P95 < 500ms
- **GPS Tracking Accuracy**: ±10 meters
- **Mobile App Crash Rate**: < 0.1%

### Operational Metrics
- **Bid Fulfillment Rate**: >80% of requests receive 3+ bids within 5 minutes
- **On-Time Delivery Rate**: >90% within promised delivery window
- **Automation Rate**: >95% of bookings with zero human intervention
- **Empty Miles Reduction**: 40% reduction via return cargo matching

### User Experience Metrics
- **Shipper Satisfaction**: 4.5+ stars (average)
- **Driver Satisfaction**: 4.3+ stars
- **App Store Rating**: 4.6+ stars
- **Monthly Active Users**: 70% of registered users

### Business Metrics
- **Platform Revenue**: 15% commission on completed shipments
- **Cost per Shipment**: < $2 operational cost
- **Driver Retention**: >85% active after 6 months
- **Shipper Retention**: >75% place repeat orders within 30 days
- **ROI for Shippers**: 20-30% cost savings vs traditional brokers
- **Market Penetration**: 15% market share within 18 months

## Related Documents

### Design Documentation
- [System Architecture](./system-architecture.md) - Comprehensive architecture and component design
- [Agent Design](./agent-design.md) - Detailed agent specifications and behaviors
- [Communication Patterns](./communication-patterns.md) *(Coming Soon)* - Message schemas and protocols
- [Data Models](./data-models.md) *(Coming Soon)* - Database schemas and data structures

### Requirements
- [Use Case Specification](../../../1-SoftwareRequirements/requirements/use-cases/UC-LOG-001-smart-logistics-platform.md) - Detailed functional requirements

### Framework Documentation
- [CodeValdCortex Backend Architecture](../../backend-architecture.md)
- [Agent Runtime Documentation](../../../3-SofwareDevelopment/core-systems/)

## Project Timeline

### Phase 1: Foundation (Months 1-3)
- Core agent implementation (Shipper, Driver, Platform, Payment)
- Basic bid/accept workflow
- GPS tracking and status updates
- MVP deployment (Nairobi region)

### Phase 2: Intelligence (Months 4-6)
- Route optimization
- Return cargo matching
- Facility integration
- Multi-stop deliveries

### Phase 3: Scale (Months 7-9)
- National expansion (Kenya)
- Advanced analytics
- Predictive demand modeling
- Performance optimization

### Phase 4: Enterprise Features (Months 10-12)
- White-label options for large shippers
- API for third-party integrations
- Cross-border capabilities
- Advanced compliance features

## Questions?

For questions about this design:
- Review the [System Architecture](./system-architecture.md) for technical details
- Review the [Agent Design](./agent-design.md) for agent specifications
- Consult the [Use Case Design Guidelines](../UsecaseDeisgn.md)
- Contact the architecture team

---

**Maintained by**: Architecture Team  
**Review Cadence**: Bi-weekly during design phase  
**Next Review**: November 5, 2025
