# Use Case Specification: Smart Logistics Service Platform

**Use Case ID**: UC-LOG-001  
**Use Case Name**: Smart Logistics Service Summoning Platform  
**Version**: 1.0  
**Date**: October 22, 2025  
**Author**: System Architect

## Executive Summary

A logistics platform where shippers (businesses/individuals needing transport), drivers/trucks (service providers), and facilities (warehouses, depots, terminals) are all registered agents that autonomously negotiate, reconcile conflicts, and optimize deliveries through intelligent agent reconciliation.

## Use Case Classification

- **Category**: Logistics & Transportation
- **Priority**: High
- **Complexity**: High
- **Status**: Proposed

## Actors

### Primary Actors
1. **Shipper Agent** (Business/Individual)
   - Requests transport services
   - Monitors shipment progress
   - Provides feedback and ratings

2. **Driver/Truck Agent** (Service Provider)
   - Responds to shipment requests
   - Executes deliveries
   - Updates real-time status

3. **Facility Agent** (Warehouse/Depot/Terminal)
   - Manages dock appointments
   - Coordinates loading/unloading
   - Tracks inventory and capacity

### Secondary Actors
4. **Platform Agent** (System)
   - Broadcasts shipment requests
   - Orchestrates agent reconciliation
   - Enforces business rules

5. **Route Optimizer Agent**
   - Calculates optimal routes
   - Considers traffic and road conditions
   - Optimizes multi-stop deliveries

6. **Payment Agent**
   - Processes transactions
   - Manages invoicing
   - Handles disputes

## Preconditions

1. All actors must be registered and verified on the platform
2. Drivers must have valid licenses and insurance
3. Vehicles must pass safety inspections
4. Shippers must have verified business/identity credentials
5. Facilities must have operational capacity management systems
6. GPS tracking must be functional on all vehicles

## Main Success Scenario

### 1. Registration Phase

**A. Shipper Registration**
1. Shipper accesses registration portal (web/mobile)
2. System presents registration form based on shipper type:
   - Business Shipper: Full profile with company details
   - Individual/SME: Simplified profile
3. Shipper provides:
   - Identity/Business verification documents
   - Contact information
   - Pickup locations (with GPS coordinates)
   - Shipping profile (cargo types, volumes, special requirements)
   - Payment information
4. System validates information
5. System performs KYC/KYB verification
6. System generates Shipper ID (e.g., `SHP-NAI-45782`)
7. System activates shipper account

**B. Driver/Truck Agent Registration**
1. Driver accesses registration portal
2. System presents driver/vehicle registration form
3. Driver provides:
   - Personal identification (ID, passport, driver's license)
   - Vehicle information (registration, type, capacity, features)
   - Insurance and inspection certificates
   - GPS tracker details
   - Operating profile (service area, hours, specializations)
   - Rate structure
4. System validates documents
5. System performs background check
6. System verifies vehicle inspection and insurance
7. System generates Driver ID (e.g., `DRV-NAI-89234`)
8. System generates Vehicle ID (e.g., `TRK-KBZ-123X`)
9. System activates driver/vehicle agent

**C. Facility Agent Registration**
1. Facility manager accesses registration portal
2. System presents facility registration form
3. Facility provides:
   - Facility details (name, type, location)
   - Capacity and infrastructure information
   - Loading dock specifications
   - Services offered
   - Dock appointment scheduling system parameters
   - Rate structure
4. System validates information
5. System performs facility verification
6. System generates Facility ID (e.g., `FAC-NAI-W0234`)
7. System integrates with facility's scheduling system
8. System activates facility agent

### 2. Service Summoning Phase

**A. Shipment Request Creation**
1. Shipper logs into platform (mobile app/web portal/API)
2. Shipper initiates "New Shipment Request"
3. System presents shipment request form
4. Shipper provides:
   - Pickup details (location, contact, ready time, loading bay)
   - Delivery details (destination, contact, required delivery time)
   - Cargo details (type, dimensions, weight, value, special requirements)
   - Service level (Standard/Express/Economy)
5. Shipper submits request
6. System validates request completeness
7. System creates Shipment ID (e.g., `SHP-54782-001`)
8. System broadcasts request to eligible truck agents

**B. Request Broadcast (0-30 seconds)**
1. Platform Agent analyzes shipment requirements
2. Platform Agent determines eligibility criteria:
   - Truck capacity ≥ cargo weight + 15% safety margin
   - Vehicle type matches cargo requirements
   - Service area includes pickup and delivery locations
   - Driver has required licenses/certifications
   - Vehicle availability within required timeframe
3. Platform Agent broadcasts to eligible Truck Agents
4. System sends push notifications to eligible drivers

### 3. Agent Reconciliation Phase

**A. Truck Agent Evaluation (30-120 seconds)**
1. Each eligible Truck Agent receives broadcast
2. Truck Agent evaluates shipment:
   - **Capacity Analysis**: Can vehicle handle cargo?
   - **Route Analysis**: Calculate distance to pickup, driving time, delivery timeline
   - **Schedule Compatibility**: Check current commitments, availability
   - **Cost Calculation**: Fuel, crew, tolls, wear-tear, profit margin
   - **Return Optimization**: Query for return cargo opportunities
   - **Risk Assessment**: Weather, road conditions, cargo value
   - **Confidence Score**: Calculate based on experience, vehicle condition, driver rating
3. Truck Agent makes decision:
   - **Accept**: Generate bid with price and timeline
   - **Decline**: Provide reason (capacity, timing, route, vehicle mismatch)
4. Truck Agent submits bid to Platform Agent

**B. Bid Collection and Ranking (120-300 seconds)**
1. Platform Agent collects all bids
2. Platform Agent ranks bids using multi-criteria scoring:
   ```
   Score = (0.3 × Price Score) + (0.3 × Time Score) + (0.2 × Reliability Score) + 
           (0.1 × Experience Score) + (0.1 × Vehicle Quality Score)
   
   Where:
   - Price Score: Normalized inverse of bid price (lower is better)
   - Time Score: Based on pickup ETA and delivery ETA
   - Reliability Score: Historical on-time delivery rate
   - Experience Score: Number of similar shipments completed
   - Vehicle Quality Score: Vehicle age, condition, features
   ```
3. Platform Agent applies shipper preferences:
   - Preferred carriers (boost score by 10%)
   - Blacklisted carriers (exclude)
   - Sustainability preference (prioritize fuel-efficient vehicles)
4. Platform Agent generates ranked list of bids
5. System presents bids to Shipper

**C. Shipper Selection (0-15 minutes)**
1. Shipper receives bid notifications
2. System displays bid comparison dashboard:
   - Price comparison
   - Estimated pickup and delivery times
   - Driver/vehicle details and ratings
   - Historical performance metrics
   - Vehicle features and photos
3. Shipper reviews bids
4. Shipper selects preferred bid OR sets auto-accept criteria
5. System confirms selection with Shipper

**D. Automatic Matching (if configured)**
1. If Shipper has auto-accept enabled:
   - System automatically selects highest-ranked bid
   - System confirms with both parties
   - System proceeds to confirmation phase
2. If no bids received after 5 minutes:
   - System expands search radius
   - System adjusts eligibility criteria
   - System re-broadcasts to wider network
3. If still no bids after 15 minutes:
   - System notifies Shipper
   - System suggests adjustments (budget, timeline, requirements)

### 4. Confirmation and Execution Phase

**A. Booking Confirmation**
1. Platform Agent sends confirmation to selected Truck Agent
2. Truck Agent accepts booking (30-second window)
3. If no response:
   - System moves to next-ranked bid
   - System notifies original agent of timeout
4. Upon acceptance:
   - System creates binding contract
   - System locks in price and timeline
   - System reserves truck capacity
   - System generates booking reference

**B. Facility Coordination (if applicable)**
1. System checks if pickup location requires dock appointment
2. If yes:
   - Facility Agent receives booking request
   - Facility Agent checks dock availability
   - Facility Agent proposes available slots
   - Truck Agent confirms preferred slot
   - System finalizes dock appointment
3. System sends appointment details to all parties

**C. Pre-Pickup Preparation**
1. System sends pickup reminder to Truck Agent (1 hour before)
2. System sends preparation reminder to Shipper (30 minutes before)
3. Truck Agent updates status: "En Route to Pickup"
4. System tracks GPS location in real-time
5. System calculates dynamic ETA
6. System notifies Shipper of truck approach (10 minutes out)

**D. Pickup Execution**
1. Truck Agent arrives at pickup location
2. Truck Agent updates status: "Arrived at Pickup"
3. Shipper and Driver verify:
   - Cargo matches description
   - Quantity is correct
   - Special handling requirements understood
4. Loading commences
5. Facility Agent (if applicable) monitors loading
6. Driver and Shipper inspect loaded cargo
7. Driver captures photos of loaded cargo
8. Shipper signs digital Bill of Lading
9. Truck Agent updates status: "Loaded, En Route to Delivery"
10. System triggers payment hold (if prepaid)

**E. In-Transit Monitoring**
1. System tracks GPS location continuously
2. System monitors for:
   - Route deviations
   - Unexpected stops (>15 minutes)
   - Speed violations
   - Geofence breaches
3. System calculates dynamic delivery ETA
4. System sends ETA updates to consignee
5. Truck Agent can update status manually:
   - "Rest Stop" (planned breaks)
   - "Delay" (with reason: traffic, weather, mechanical)
6. System alerts stakeholders of significant delays (>30 minutes)

**F. Delivery Execution**
1. System sends arrival notification to consignee (30 minutes out)
2. Truck Agent updates status: "Approaching Delivery"
3. If facility delivery:
   - Facility Agent confirms dock assignment
   - System provides dock directions
4. Truck Agent arrives at delivery location
5. Truck Agent updates status: "Arrived at Delivery"
6. Unloading commences
7. Consignee inspects cargo
8. Consignee verifies:
   - Cargo condition (damage check)
   - Quantity matches Bill of Lading
   - Special handling requirements met
9. Consignee signs digital Proof of Delivery (POD)
10. Driver captures photos of delivered cargo
11. Truck Agent updates status: "Delivered"
12. System records delivery timestamp

**G. Post-Delivery**
1. System triggers payment release to Driver
2. System sends delivery confirmation to Shipper
3. System requests feedback from both parties:
   - Shipper rates Driver (1-5 stars)
   - Driver rates Shipper (1-5 stars)
   - Facility rates both (if applicable)
4. System updates agent reputation scores
5. System archives shipment data
6. System generates invoice and receipt
7. If issues reported:
   - System initiates dispute resolution workflow
   - Payment Agent holds funds pending resolution

### 5. Advanced Features

**A. Route Optimization for Multi-Stop Deliveries**
1. Shipper creates multi-stop shipment
2. Route Optimizer Agent calculates optimal sequence
3. System considers:
   - Delivery time windows
   - Cargo compatibility (can't mix hazmat with food)
   - Vehicle capacity at each stop
   - Traffic patterns
4. System generates optimized route
5. Truck Agent follows route with turn-by-turn navigation

**B. Return Cargo Optimization**
1. After delivery, Truck Agent queries for return opportunities
2. Platform Agent searches for:
   - Shipments originating near current location
   - Destination near Driver's base or next planned location
   - Cargo compatible with vehicle type
3. System presents return cargo options to Driver
4. Driver accepts return cargo
5. System updates route and earnings projection

**C. Emergency Re-routing**
1. System monitors for disruptions:
   - Severe weather
   - Road closures
   - Accidents
   - Vehicle breakdown
2. If disruption detected:
   - Route Optimizer Agent calculates alternative route
   - System notifies Truck Agent
   - System updates ETA for consignee
3. If vehicle breakdown:
   - System broadcasts emergency re-assignment
   - Nearby Truck Agents receive priority notification
   - System coordinates cargo transfer
   - Original driver assists with transfer

## Alternative Flows

### A1: No Bids Received
1. After 5 minutes, if no bids:
   - System analyzes request for potential issues:
     - Price too low
     - Timeline too tight
     - Route unpopular
     - Cargo requirements too restrictive
2. System suggests adjustments to Shipper
3. Shipper modifies request
4. System re-broadcasts with adjusted parameters
5. If still no bids after 15 minutes:
   - System escalates to human dispatcher
   - Dispatcher manually reaches out to drivers

### A2: Bid Accepted but Driver Cancels
1. Driver cancels after acceptance (penalty applies)
2. System immediately moves to next-ranked bid
3. System sends urgent notification to backup driver
4. If backup accepts:
   - System updates booking
   - System notifies Shipper of driver change
5. If no backup available:
   - System re-broadcasts as urgent request
   - System offers bonus incentive

### A3: Cargo Does Not Match Description
1. Driver arrives at pickup
2. Driver finds cargo mismatch (size, weight, or type)
3. Driver captures photos and documents discrepancy
4. System notifies Shipper and Platform Agent
5. Platform Agent evaluates:
   - Can current vehicle still handle cargo? → Proceed with price adjustment
   - Cargo requires different vehicle? → Cancel and re-broadcast
6. If cancellation required:
   - System charges Shipper cancellation fee
   - System compensates Driver for wasted trip
   - System creates new shipment request with correct details

### A4: Delivery Refused by Consignee
1. Consignee refuses delivery (wrong items, damaged, etc.)
2. Driver documents refusal with photos and signature
3. System notifies Shipper
4. Shipper provides instructions:
   - Return to origin (Driver charges return trip)
   - Deliver to alternate location
   - Dispose/donate (with written authorization)
5. System updates booking with new instructions
6. Payment Agent holds funds pending resolution

### A5: Driver Stuck in Traffic (Major Delay)
1. GPS tracking detects vehicle stopped for >30 minutes
2. System analyzes traffic data
3. If major delay confirmed:
   - System notifies Shipper and consignee
   - System calculates new ETA
   - Route Optimizer suggests alternate route
4. If delay exceeds delivery window:
   - System offers consignee options:
     - Accept delayed delivery
     - Reschedule delivery
     - Cancel with full refund
5. If cancellation chosen:
   - System compensates Driver for distance traveled
   - System re-broadcasts shipment

## Exception Flows

### E1: Vehicle Breakdown During Transit
1. Driver reports breakdown via app
2. System marks vehicle as "BREAKDOWN"
3. System dispatches roadside assistance (if subscribed)
4. Platform Agent broadcasts emergency re-assignment:
   - Nearby vehicles within 50km
   - Same or larger capacity
   - Urgent priority (premium pay)
5. Backup driver accepts
6. System coordinates:
   - Meeting location for cargo transfer
   - Transfer documentation
   - Payment split between drivers
7. System updates consignee with new ETA

### E2: Accident During Transit
1. Driver reports accident via app emergency button
2. System alerts emergency services (if enabled)
3. System notifies Shipper, consignee, and insurance provider
4. System locks GPS location and records
5. Platform Agent assesses:
   - Driver/vehicle safety status
   - Cargo condition
6. If cargo intact and driveable:
   - Proceed after clearance
7. If cargo transfer needed:
   - Follow breakdown protocol (E1)
8. Insurance Agent initiates claim process

### E3: Theft or Cargo Loss
1. Driver reports theft via app
2. System immediately:
   - Captures last known GPS location
   - Notifies law enforcement
   - Alerts insurance provider
   - Locks all shipment data as evidence
3. System notifies Shipper
4. Investigation Agent reviews:
   - GPS tracking history
   - Route deviations
   - Stop duration anomalies
   - Driver history
5. Insurance Agent processes claim
6. System flags driver for review (if suspicious activity)

### E4: Payment Dispute
1. Either party disputes payment amount
2. Payment Agent holds funds
3. System retrieves shipment data:
   - Original bid and acceptance
   - Actual distance traveled
   - Delivery proof and timestamps
   - Any modifications or delays
4. Dispute Agent reviews evidence
5. If dispute valid:
   - System adjusts payment
   - System compensates affected party
6. If dispute invalid:
   - System releases original payment
   - System warns disputing party (penalty for false claims)

## Postconditions

### Success Postconditions
1. Cargo delivered to destination on time
2. All parties satisfied (ratings ≥ 4 stars)
3. Payment processed successfully
4. GPS tracking data archived
5. All documentation stored (BOL, POD, photos)
6. Agent reputation scores updated
7. Shipment marked as "COMPLETED"

### Failure Postconditions
1. Shipment cancelled with appropriate refunds
2. All parties notified of cancellation reason
3. Penalties applied per contract terms
4. Incident report generated (if applicable)
5. Lessons learned documented for system improvement
6. Shipment marked as "CANCELLED" or "FAILED"

## Business Rules

### BR1: Pricing Rules
- Minimum charge: 1000 KSH (local), 5000 KSH (intercity)
- Surge pricing during peak hours: Up to 1.5× base rate
- Volume discounts for corporate shippers: 10-30% based on monthly volume
- Fuel surcharge: Adjusts weekly based on fuel price index
- Waiting time: 500 KSH per hour after 30-minute grace period

### BR2: Capacity Rules
- Vehicles must not be loaded beyond 85% of rated capacity (safety margin)
- Mixed cargo loads require compatibility check (e.g., no chemicals with food)
- High-value cargo (>1M KSH) requires GPS tracking and insurance
- Hazmat requires special license and dedicated vehicle (no mixed loads)

### BR3: Timing Rules
- Pickup window: ±30 minutes of scheduled time
- Delivery window: ±2 hours for intercity, ±30 minutes for local
- Driver must accept booking within 30 seconds or system moves to next bid
- Shipper must load cargo within stated timeframe or pay waiting charges
- Cancellation penalty: 20% if <2 hours before pickup, 50% if after driver en route

### BR4: Rating and Reputation Rules
- Minimum acceptable rating: 3.5 stars (below triggers review)
- Drivers with <3.5 stars for 5 consecutive shipments suspended
- Shippers with <3.5 stars flagged as "difficult" (drivers can decline)
- Facility delays (>30 minutes) count against facility rating
- Perfect delivery streak (10+) earns "Gold Driver" badge (priority matching)

### BR5: Geographic Rules
- Service areas defined by postal codes or GPS polygons
- Cross-border shipments require customs documentation
- Restricted areas (military, private) require special permits
- Some routes require convoy travel (security concerns)

### BR6: Insurance and Liability Rules
- Platform provides basic coverage: 100K KSH per shipment
- Shippers can purchase additional coverage for high-value cargo
- Driver liable for cargo loss/damage if due to negligence
- Force majeure events (natural disasters, war) void liability
- Proof of Delivery required for payment release

## Special Requirements

### SR1: Real-Time Tracking
- GPS updates every 30 seconds during transit
- System stores complete route history for 90 days
- Shipper and consignee can view live map
- Geofencing alerts for route deviations
- ETA updates every 5 minutes

### SR2: Mobile-First Design
- Driver app works offline (queue actions for sync)
- Low-bandwidth mode for rural areas (<100 KB per update)
- Voice navigation in local languages (English, Swahili, etc.)
- One-tap status updates (no typing while driving)
- Emergency button accessible from lock screen

### SR3: Multi-Language Support
- Platform available in: English, Swahili, Arabic, French
- Driver app voice commands in native languages
- Consignee SMS notifications in preferred language
- Customer support chatbot multilingual

### SR4: Accessibility
- Voice-guided interface for visually impaired drivers
- High-contrast mode for outdoor visibility
- Large touch targets (minimum 44×44 pixels)
- Screen reader compatible

### SR5: Scalability
- System must handle 10,000+ concurrent shipments
- Bid reconciliation must complete within 5 minutes
- GPS tracking for 50,000+ vehicles simultaneously
- Database must store 5 years of shipment history

### SR6: Security and Privacy
- End-to-end encryption for all communications
- Anonymized GPS data (no exact home addresses exposed)
- Driver background checks required
- Payment card data PCI-DSS compliant
- GDPR-compliant data retention and deletion

### SR7: Integration Requirements
- REST API for corporate shipper integration
- Webhook notifications for shipment status updates
- Integration with M-Pesa, PayPal, Stripe
- Integration with Google Maps, Waze for traffic data
- Integration with national logistics database (KRA customs)

## Assumptions and Dependencies

### Assumptions
1. All vehicles have functioning GPS trackers
2. Drivers have smartphones with data connectivity
3. Pickup and delivery locations have road access
4. Shippers provide accurate cargo descriptions
5. Payment systems are operational
6. Weather conditions allow for safe driving

### Dependencies
1. GPS/cellular network coverage
2. Third-party payment gateways (M-Pesa, banks)
3. Mapping services (Google Maps API)
4. SMS gateway for notifications (Twilio)
5. Cloud infrastructure (AWS, Azure)
6. Government regulatory compliance (NTSA, KRA)

## Performance Requirements

| Metric | Target |
|--------|--------|
| Bid response time | <5 minutes for 90% of requests |
| GPS update frequency | Every 30 seconds |
| System uptime | 99.5% (excluding planned maintenance) |
| API response time | <500ms for 95% of requests |
| Mobile app launch time | <3 seconds |
| Concurrent users | 50,000+ |
| Database query time | <200ms for 95% of queries |

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Driver no-show | Medium | High | Automated backup driver system, penalties |
| GPS failure | Low | High | Fallback to manual check-ins, cached route data |
| Payment fraud | Medium | High | Multi-factor authentication, transaction limits |
| Cargo theft | Low | Critical | GPS tracking, insurance, vetted drivers |
| System downtime | Low | High | Redundant servers, auto-failover, offline mode |
| Data breach | Low | Critical | Encryption, regular security audits, compliance |

## Success Metrics

1. **Operational Metrics**
   - Bid fill rate: >90% (percentage of requests receiving bids)
   - On-time pickup rate: >85%
   - On-time delivery rate: >90%
   - Average bid response time: <3 minutes
   - Driver acceptance rate: >70%

2. **Quality Metrics**
   - Average shipper rating: >4.2 stars
   - Average driver rating: >4.2 stars
   - Cargo damage rate: <1%
   - Dispute rate: <5%
   - Successful deliveries: >95%

3. **Business Metrics**
   - Monthly active shippers: Growth target 20% MoM
   - Monthly active drivers: Growth target 15% MoM
   - Gross Merchandise Value (GMV): Platform commission revenue
   - Customer retention rate: >80%
   - Driver retention rate: >75%

## Related Documents

- [Use Case: Ride-Hailing Platform](./UC-RIDE-001-ride-hailing-platform.md)
- [Use Case: Charity Distribution Network](./UC-CHAR-001-charity-distribution-network.md)
- [System Architecture](../../2-SoftwareDesignAndArchitecture/backend-architecture.md)
- [Agent Design Patterns](../../2-SoftwareDesignAndArchitecture/Usecases/UsecaseDeisgn.md)

## Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Product Owner | | | |
| Technical Lead | | | |
| Business Analyst | | | |
| Stakeholder | | | |

---

**Change History**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-22 | System Architect | Initial version |
