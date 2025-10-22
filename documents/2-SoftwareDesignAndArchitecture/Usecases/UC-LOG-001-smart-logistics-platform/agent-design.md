# UC-LOG-001 Agent Design: Smart Logistics Service Platform

**Version**: 1.0  
**Last Updated**: October 22, 2025  
**Status**: Design Phase

## Table of Contents
1. [Overview](#overview)
2. [Agent Type Summary](#agent-type-summary)
3. [Agent Specifications](#agent-specifications)
4. [Agent Relationships](#agent-relationships)
5. [Agent Lifecycle Management](#agent-lifecycle-management)
6. [Communication Patterns](#communication-patterns)

## Overview

CodeValdUsafiri employs **8 autonomous agent types** that collaborate to manage the entire logistics lifecycle from shipment request to delivery completion. Each agent operates independently with local decision-making capabilities while coordinating through message passing to achieve system-wide optimization.

### Agent Philosophy

- **Autonomy**: Agents make decisions based on local state and rules without central command
- **Collaboration**: Agents achieve complex behaviors through simple message exchanges
- **Specialization**: Each agent type has a focused responsibility
- **Scalability**: Agents can be replicated horizontally without coordination overhead
- **Resilience**: Failure of individual agents doesn't cascade to system failure

## Agent Type Summary

| Agent Type | Count | Lifecycle | Primary Responsibility |
|------------|-------|-----------|------------------------|
| **Shipper Agent** | 10,000+ | On-demand (per user session) | Create shipment requests, select bids, track deliveries |
| **Driver/Truck Agent** | 5,000+ | On-demand (when driver online) | Evaluate opportunities, submit bids, execute deliveries |
| **Facility Agent** | 200+ | Persistent (24/7) | Manage dock capacity, schedule appointments, coordinate loading |
| **Platform Agent** | 5-10 | Persistent (24/7, replicated) | Broadcast requests, rank bids, match shippers with drivers |
| **Route Optimizer Agent** | 3-5 | Persistent (24/7, replicated) | Calculate optimal routes, monitor traffic, provide navigation |
| **Payment Agent** | 2 | Persistent (leader-follower) | Process payments, manage escrow, handle disputes |
| **Conflict Resolver Agent** | 2-3 | Persistent (24/7) | Mediate disputes, coordinate emergency re-routing |
| **Analytics Agent** | 1 | Persistent (batch processing) | Track KPIs, generate insights, predict patterns |

**Total Active Agents**: 15,000-20,000 at peak load

## Agent Specifications

### 1. Shipper Agent

**Purpose**: Represents a business or individual shipper requesting transport services

**Agent ID Format**: `SHP-{CITY}-{SEQUENTIAL}` (e.g., `SHP-NAI-45782`)

#### Attributes

```go
type ShipperAgent struct {
    // Identity
    agent_id          string      // Unique identifier
    user_id           string      // Associated user account
    business_name     string      // Company name (if business)
    contact_name      string      // Primary contact person
    phone             string      // Contact phone number
    email             string      // Contact email
    verified          bool        // KYC/KYB verification status
    
    // Profile
    shipper_type      string      // "business", "individual", "enterprise"
    default_locations []GeoPoint  // Frequently used pickup locations
    cargo_profile     []string    // Typical cargo types
    volume_tier       string      // "low" (<10/mo), "medium" (10-50/mo), "high" (>50/mo)
    
    // Preferences
    preferred_drivers []string    // Driver IDs of preferred partners
    auto_accept       bool        // Auto-select best bid
    price_sensitivity float64     // 0.0 (price-focused) to 1.0 (quality-focused)
    
    // Reputation
    rating            float64     // Average rating from drivers (1.0-5.0)
    total_shipments   int         // Total completed shipments
    cancellation_rate float64     // Percentage of cancellations
    
    // State
    current_state     string      // See state machine below
    active_shipments  []string    // IDs of shipments in progress
    created_at        time.Time
    last_active       time.Time
}

type GeoPoint struct {
    latitude   float64
    longitude  float64
    address    string
    name       string  // e.g., "Warehouse A"
}
```

#### Capabilities

- **Create Shipment Requests**: Define pickup, delivery, cargo details
- **Receive and Evaluate Bids**: Review submitted bids from drivers
- **Select Winning Bid**: Choose driver based on price, time, reputation
- **Monitor Shipment Progress**: Track GPS location and status updates
- **Communicate with Driver**: Send messages, call if needed
- **Verify Delivery**: Review photos, sign digital proof of delivery
- **Provide Ratings**: Rate driver performance after completion
- **Manage Payment**: Approve charges, dispute if issues

#### State Machine

```
States:
┌─────────────┐
│ Registered  │ ← Initial state after account creation
└──────┬──────┘
       │ create_shipment_request()
       ▼
┌─────────────┐
│RequestPend. │ ← Awaiting bids from drivers
└──────┬──────┘
       │ receive_bids()
       ▼
┌─────────────┐
│BidsReceived │ ← Evaluating bids, ready to select
└──────┬──────┘
       │ select_bid()
       ▼
┌─────────────┐
│   Booked    │ ← Driver assigned, awaiting pickup
└──────┬──────┘
       │ driver_loading()
       ▼
┌─────────────┐
│  InTransit  │ ← Cargo en route to destination
└──────┬──────┘
       │ delivery_completed()
       ▼
┌─────────────┐
│  Delivered  │ ← Shipment completed successfully
└──────┬──────┘
       │ dispute_raised() OR idle_timeout()
       ▼                     ▼
┌─────────────┐      ┌─────────────┐
│  Disputed   │      │ Registered  │ (back to initial)
└─────────────┘      └─────────────┘

Transitions:
- Registered → RequestPending: Shipper creates new shipment request
- RequestPending → BidsReceived: At least one bid submitted (within 120s)
- RequestPending → Registered: Timeout with no bids (15 minutes)
- BidsReceived → Booked: Shipper selects winning bid
- Booked → InTransit: Driver starts trip after loading
- Booked → Registered: Cancellation by shipper or driver
- InTransit → Delivered: Driver completes delivery, shipper signs POD
- InTransit → Disputed: Issue raised during transit
- Delivered → Disputed: Dispute raised within 24-hour window
- Disputed → Registered: Dispute resolved
```

#### Example Behavior

```go
// Shipper Agent evaluates bids and selects winner
func (s *ShipperAgent) EvaluateBids(bids []Bid) *Bid {
    if s.auto_accept {
        // Automatic selection using weighted scoring
        return s.SelectBestBid(bids)
    } else {
        // Present to shipper for manual selection
        s.NotifyShipperOfBids(bids)
        return nil  // Wait for shipper's manual selection
    }
}

func (s *ShipperAgent) SelectBestBid(bids []Bid) *Bid {
    var bestBid *Bid
    var bestScore float64 = -1.0
    
    for _, bid := range bids {
        // Calculate score based on shipper's price sensitivity
        priceScore := 1.0 - (bid.Price - minPrice) / (maxPrice - minPrice)
        timeScore := 1.0 - (bid.EstimatedTime - minTime) / (maxTime - minTime)
        reliabilityScore := bid.DriverRating / 5.0
        
        // Weight based on shipper's price sensitivity
        score := (s.price_sensitivity * priceScore) +
                 ((1 - s.price_sensitivity) * 0.5 * timeScore) +
                 ((1 - s.price_sensitivity) * 0.5 * reliabilityScore)
        
        // Boost preferred drivers by 10%
        if contains(s.preferred_drivers, bid.DriverID) {
            score *= 1.1
        }
        
        if score > bestScore {
            bestScore = score
            bestBid = &bid
        }
    }
    
    return bestBid
}

// Monitor shipment progress and alert on issues
func (s *ShipperAgent) MonitorShipment(shipmentID string) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            status := s.GetShipmentStatus(shipmentID)
            
            // Check for significant delays
            if status.ETADelay > 30*time.Minute {
                s.NotifyShipper(shipmentID, "Delivery delayed by "+status.ETADelay.String())
            }
            
            // Check for route deviations
            if status.RouteDeviation > 5.0 {  // km
                s.RaiseAlert(shipmentID, "Driver off planned route")
            }
            
            // Exit monitoring when delivered
            if status.State == "Delivered" {
                return
            }
        }
    }
}
```

#### Communication Patterns

**Publishes**:
- `shipment.requested` → Platform Agent (broadcast shipment details)
- `bid.selected` → Driver Agent (notify winning driver)
- `rating.submitted` → Analytics Agent (driver performance data)

**Subscribes**:
- `bid.submitted` ← Driver Agents (receive bids)
- `shipment.status_changed` ← Driver Agent (delivery progress updates)
- `payment.charged` ← Payment Agent (payment confirmation)

---

### 2. Driver/Truck Agent

**Purpose**: Represents a driver and vehicle providing transport services

**Agent ID Format**: `DRV-{CITY}-{SEQUENTIAL}` (e.g., `DRV-NAI-89234`)

#### Attributes

```go
type DriverTruckAgent struct {
    // Identity
    agent_id          string      // Unique identifier
    user_id           string      // Associated driver account
    full_name         string      // Driver's name
    phone             string      // Contact number
    license_number    string      // Driver's license number
    verified          bool        // Background check status
    
    // Vehicle
    vehicle_id        string      // Vehicle identifier (e.g., TRK-KBZ-123X)
    vehicle_type      string      // "pickup", "box_truck", "flatbed", "refrigerated"
    make_model        string      // e.g., "Isuzu NQR 2020"
    capacity_kg       int         // Maximum cargo weight
    capacity_m3       float64     // Volume capacity
    license_plate     string      
    year              int         // Manufacturing year
    condition         string      // "excellent", "good", "fair"
    features          []string    // ["GPS", "refrigeration", "lift_gate"]
    
    // Operating Profile
    base_location     GeoPoint    // Home base / garage location
    service_radius    float64     // Maximum distance willing to travel (km)
    operating_hours   TimeRange   // Preferred working hours
    specializations   []string    // ["hazmat", "fragile", "perishable"]
    rate_per_km       float64     // Base rate (KSH per km)
    
    // Current Status
    current_location  GeoPoint    // Real-time GPS coordinates
    availability      string      // "available", "on_trip", "offline"
    current_shipment  string      // Active shipment ID
    fuel_level        float64     // Percentage (0.0-1.0)
    
    // Reputation
    rating            float64     // Average rating from shippers (1.0-5.0)
    total_deliveries  int         // Completed deliveries
    on_time_rate      float64     // Percentage delivered on time
    acceptance_rate   float64     // Percentage of bids accepted by shippers
    cancellation_rate float64     // Percentage of cancellations
    
    // State
    current_state     string      // See state machine below
    created_at        time.Time
    last_active       time.Time
}

type TimeRange struct {
    start_hour int  // 0-23
    end_hour   int  // 0-23
}
```

#### Capabilities

- **Receive Shipment Broadcasts**: Get notified of matching opportunities
- **Evaluate Shipment Feasibility**: Check capacity, route, timing
- **Calculate Bid Price**: Factor in distance, time, fuel, competition
- **Submit Competitive Bids**: Send bid with price and estimated timeline
- **Accept Bookings**: Commit to accepted shipments
- **Navigate to Pickup**: Follow optimized route with traffic updates
- **Load Cargo**: Verify and photograph cargo at pickup
- **Execute Delivery**: Drive to destination with real-time tracking
- **Unload Cargo**: Verify delivery and capture proof of delivery
- **Search Return Cargo**: Find revenue opportunities for return trip
- **Update Real-Time Status**: Broadcast location and progress

#### State Machine

```
States:
┌─────────────┐
│   Offline   │ ← Driver not logged in or unavailable
└──────┬──────┘
       │ driver_goes_online()
       ▼
┌─────────────┐
│  Available  │ ← Ready to receive shipment requests
└──────┬──────┘
       │ receive_shipment_broadcast()
       ▼
┌─────────────┐
│ Evaluating  │ ← Analyzing shipment, preparing bid
└──────┬──────┘
       │ submit_bid() OR decline()
       ▼                  │
┌─────────────┐          │
│  Available  │ ←────────┘ (back to available if declined)
└──────┬──────┘
       │ bid_accepted_by_shipper()
       ▼
┌─────────────┐
│ Committed   │ ← Booking confirmed, preparing for pickup
└──────┬──────┘
       │ start_navigation_to_pickup()
       ▼
┌─────────────┐
│ EnRoutePickup│ ← Driving to pickup location
└──────┬──────┘
       │ arrive_at_pickup()
       ▼
┌─────────────┐
│  Loading    │ ← At pickup, loading cargo
└──────┬──────┘
       │ loading_complete()
       ▼
┌─────────────┐
│ InTransit   │ ← Delivering cargo to destination
└──────┬──────┘
       │ arrive_at_delivery()
       ▼
┌─────────────┐
│ Unloading   │ ← At destination, unloading cargo
└──────┬──────┘
       │ delivery_complete()
       ▼
┌─────────────┐
│ReturnSearch │ ← Looking for return cargo opportunity
└──────┬──────┘
       │ return_cargo_found() OR timeout(5min)
       ▼                         │
┌─────────────┐                 │
│ Committed   │ ←───────────────┘
└─────────────┘                 │
       OR                        │
       ▼                         │
┌─────────────┐                 │
│  Available  │ ←───────────────┘

Transitions:
- Offline → Available: Driver logs in and goes online
- Available → Evaluating: Receives matching shipment broadcast
- Evaluating → Available: Submits bid or declines request
- Available → Committed: Shipper accepts driver's bid
- Committed → EnRoutePickup: Driver starts navigation to pickup
- EnRoutePickup → Loading: Driver arrives at pickup location
- Loading → InTransit: Cargo loaded, BoL signed, trip started
- InTransit → Unloading: Driver arrives at delivery destination
- Unloading → ReturnSearch: Delivery complete, POD signed
- ReturnSearch → Committed: Return cargo opportunity accepted
- ReturnSearch → Available: No return cargo, driver available for new requests
- Any State → Offline: Driver goes offline or network disconnection
```

#### Example Behavior

```go
// Driver Agent evaluates shipment and decides whether to bid
func (d *DriverTruckAgent) EvaluateShipment(shipment Shipment) (*Bid, error) {
    // Check capacity
    if shipment.CargoWeight > d.capacity_kg {
        return nil, errors.New("Cargo exceeds vehicle capacity")
    }
    
    // Check if cargo type is compatible
    if shipment.RequiresRefrigeration && !contains(d.features, "refrigeration") {
        return nil, errors.New("Vehicle lacks required features")
    }
    
    // Check if within service radius
    distanceToPickup := d.CalculateDistance(d.current_location, shipment.PickupLocation)
    if distanceToPickup > d.service_radius {
        return nil, errors.New("Pickup location outside service radius")
    }
    
    // Calculate route and costs
    route := d.GetOptimalRoute(shipment.PickupLocation, shipment.DeliveryLocation)
    totalDistance := distanceToPickup + route.Distance
    estimatedTime := route.Duration
    
    // Calculate costs
    fuelCost := d.CalculateFuelCost(totalDistance)
    laborCost := d.CalculateLaborCost(estimatedTime)
    overhead := 500.0  // KSH (insurance, maintenance, tolls)
    profitMargin := 0.20  // 20%
    
    baseCost := fuelCost + laborCost + overhead
    bidPrice := baseCost * (1 + profitMargin)
    
    // Adjust for competition (bid lower if acceptance rate is low)
    if d.acceptance_rate < 0.30 {
        bidPrice *= 0.95  // 5% discount
    }
    
    // Adjust for cargo value (higher price for high-value cargo)
    if shipment.CargoValue > 1000000 {  // > 1M KSH
        bidPrice *= 1.10  // 10% premium
    }
    
    // Check profitability
    if bidPrice < baseCost * 1.05 {  // Minimum 5% margin
        return nil, errors.New("Unprofitable shipment")
    }
    
    // Create bid
    bid := &Bid{
        BidID:          generateBidID(),
        DriverID:       d.agent_id,
        ShipmentID:     shipment.ID,
        BidPrice:       bidPrice,
        EstimatedPickup: time.Now().Add(distanceToPickup / 50 * time.Hour),  // 50 km/h avg
        EstimatedDelivery: time.Now().Add(estimatedTime),
        ConfidenceScore: d.CalculateConfidence(shipment),
    }
    
    return bid, nil
}

// Real-time GPS tracking
func (d *DriverTruckAgent) StreamGPSLocation() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if d.current_state == "InTransit" || d.current_state == "EnRoutePickup" {
                location := d.GetCurrentGPSLocation()
                
                // Update cache
                d.UpdateLocationCache(location)
                
                // Broadcast to subscribers (shipper, platform)
                d.PublishMessage("location.updated", LocationUpdate{
                    DriverID:  d.agent_id,
                    Latitude:  location.latitude,
                    Longitude: location.longitude,
                    Heading:   location.heading,
                    Speed:     location.speed,
                    Timestamp: time.Now(),
                })
                
                // Check if approaching destination
                distanceToDestination := d.CalculateDistance(location, d.destination)
                if distanceToDestination < 5.0 {  // Within 5 km
                    d.NotifyApproaching()
                }
            }
        }
    }
}
```

#### Communication Patterns

**Publishes**:
- `bid.submitted` → Platform Agent, Shipper Agent
- `location.updated` → Shipper Agent, Platform Agent (real-time tracking)
- `status.changed` → Shipper Agent, Facility Agent (delivery progress)
- `return_cargo.requested` → Platform Agent (search for return loads)

**Subscribes**:
- `shipment.requested` ← Platform Agent (new opportunities)
- `booking.confirmed` ← Platform Agent (bid accepted notification)
- `route.updated` ← Route Optimizer Agent (re-routing instructions)

---

### 3. Facility Agent

**Purpose**: Represents warehouses, depots, terminals managing loading/unloading operations

**Agent ID Format**: `FAC-{CITY}-{TYPE}{SEQ}` (e.g., `FAC-NAI-W0234` for warehouse)

#### Attributes

```go
type FacilityAgent struct {
    // Identity
    agent_id        string      // Unique identifier
    facility_name   string      // e.g., "Nairobi Distribution Center"
    facility_type   string      // "warehouse", "depot", "terminal", "cross_dock"
    operator        string      // Operating company name
    
    // Location
    address         string
    location        GeoPoint
    access_points   []string    // ["Gate A", "Gate B", "Dock 1-5"]
    
    // Capacity
    total_docks     int         // Number of loading docks
    dock_types      []DockSpec  // Specifications for each dock
    max_daily_volume int        // Maximum shipments per day
    
    // Services
    services        []string    // ["loading", "unloading", "storage", "palletizing"]
    equipment       []string    // ["forklift", "pallet_jack", "overhead_crane"]
    operating_hours TimeRange   // e.g., 06:00-22:00
    
    // Current Status
    available_docks int         // Currently unoccupied docks
    scheduled_slots []DockSlot  // Upcoming appointments
    current_loads   []string    // Shipments currently being processed
    
    // Performance
    avg_loading_time   time.Duration  // Average time to load a truck
    avg_unloading_time time.Duration  // Average time to unload
    on_time_rate       float64        // Percentage of slots started on time
    
    // State
    current_state   string      // See state machine below
    created_at      time.Time
}

type DockSpec struct {
    dock_id       string    // "Dock-1"
    height        float64   // Dock height (meters)
    max_vehicle   string    // "box_truck", "semi_truck", "container"
    has_lift_gate bool
}

type DockSlot struct {
    slot_id       string
    dock_id       string
    shipment_id   string
    start_time    time.Time
    duration      time.Duration
    operation     string  // "loading" or "unloading"
    status        string  // "scheduled", "in_progress", "completed"
}
```

#### Capabilities

- **Manage Dock Availability**: Track real-time dock occupancy
- **Schedule Appointments**: Assign time slots for incoming trucks
- **Coordinate Loading/Unloading**: Direct trucks to appropriate docks
- **Monitor Operations**: Track time spent at each dock
- **Handle Delays**: Reschedule appointments when delays occur
- **Report Status**: Notify drivers and shippers of readiness
- **Optimize Utilization**: Balance load across available docks

#### State Machine

```
States:
┌─────────────┐
│Operational  │ ← Normal operations, docks available
└──────┬──────┘
       │ all_docks_occupied()
       ▼
┌─────────────┐
│ AtCapacity  │ ← All docks busy, queue forming
└──────┬──────┘
       │ dock_becomes_available()
       ▼
┌─────────────┐
│Operational  │
└──────┬──────┘
       │ scheduled_maintenance() OR emergency()
       ▼
┌─────────────┐
│ Maintenance │ ← Facility temporarily closed
└──────┬──────┘
       │ maintenance_complete()
       ▼
┌─────────────┐
│Operational  │
└──────┬──────┘
       │ end_of_day() AND outside_operating_hours()
       ▼
┌─────────────┐
│   Closed    │ ← Outside operating hours
└──────┬──────┘
       │ start_of_day()
       ▼
┌─────────────┐
│Operational  │
└─────────────┘

Transitions:
- Operational → AtCapacity: Last available dock becomes occupied
- AtCapacity → Operational: Any dock completes operation and becomes free
- Operational → Maintenance: Scheduled maintenance or emergency issue
- Maintenance → Operational: Maintenance work completed
- Operational → Closed: Outside operating hours (e.g., after 22:00)
- Closed → Operational: Operating hours start (e.g., 06:00)
```

#### Example Behavior

```go
// Facility Agent schedules dock appointment
func (f *FacilityAgent) ScheduleDockAppointment(shipment Shipment, operation string) (*DockSlot, error) {
    // Estimate operation duration based on cargo
    estimatedDuration := f.EstimateOperationTime(shipment.CargoWeight, operation)
    
    // Find best available slot
    slot := f.FindOptimalSlot(shipment.EstimatedArrival, estimatedDuration)
    if slot == nil {
        return nil, errors.New("No available slots within requested timeframe")
    }
    
    // Assign dock based on vehicle type
    dock := f.AssignDock(shipment.VehicleType)
    if dock == nil {
        return nil, errors.New("No suitable dock available")
    }
    
    // Create slot reservation
    dockSlot := &DockSlot{
        slot_id:     generateSlotID(),
        dock_id:     dock.dock_id,
        shipment_id: shipment.ID,
        start_time:  slot.start_time,
        duration:    estimatedDuration,
        operation:   operation,
        status:      "scheduled",
    }
    
    // Add to schedule
    f.scheduled_slots = append(f.scheduled_slots, *dockSlot)
    f.available_docks--
    
    // Notify relevant parties
    f.PublishMessage("dock.scheduled", DockScheduledEvent{
        FacilityID:  f.agent_id,
        ShipmentID:  shipment.ID,
        DockID:      dock.dock_id,
        StartTime:   slot.start_time,
        Instructions: f.GetDockInstructions(dock),
    })
    
    return dockSlot, nil
}

// Monitor dock utilization and optimize
func (f *FacilityAgent) OptimizeDockUtilization() {
    ticker := time.NewTicker(10 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Check for overdue operations
            for i, slot := range f.scheduled_slots {
                if slot.status == "in_progress" {
                    elapsed := time.Since(slot.start_time)
                    if elapsed > slot.duration * 1.5 {  // 50% overtime
                        // Send alert
                        f.RaiseAlert("Dock "+slot.dock_id+" operation overdue", slot)
                        
                        // Notify affected parties
                        f.NotifyDelayToUpcoming(slot.dock_id)
                    }
                }
            }
            
            // Consolidate schedule (fill gaps)
            f.ConsolidateSchedule()
            
            // Predict capacity issues
            if f.PredictCapacityIssue(time.Now().Add(2 * time.Hour)) {
                f.PublishMessage("facility.capacity_warning", CapacityWarning{
                    FacilityID: f.agent_id,
                    TimeFrame:  "Next 2 hours",
                    Action:     "Consider alternate facilities",
                })
            }
        }
    }
}
```

#### Communication Patterns

**Publishes**:
- `dock.scheduled` → Driver Agent, Shipper Agent (appointment confirmation)
- `loading.completed` → Driver Agent, Platform Agent (operation finished)
- `delay.reported` → Driver Agent, Shipper Agent (schedule changes)
- `facility.capacity_warning` → Platform Agent (capacity constraints)

**Subscribes**:
- `booking.confirmed` ← Platform Agent (new shipment requiring dock)
- `driver.arriving` ← Driver Agent (truck approaching facility)
- `shipment.cancelled` ← Shipper Agent, Platform Agent (free up dock slot)

---

### 4. Platform Agent

**Purpose**: Central orchestrator managing shipment broadcasts, bid reconciliation, and matching

**Agent ID Format**: `PLAT-{INSTANCE}` (e.g., `PLAT-001`)

*(Due to length constraints, I'll continue with the remaining agents in a more condensed format)*

#### Attributes

```go
type PlatformAgent struct {
    agent_id          string
    active_requests   map[string]ShipmentRequest  // shipment_id -> request
    bid_pools         map[string][]Bid            // shipment_id -> bids
    matching_rules    MatchingRuleSet
    eligibility_cache map[string][]string         // shipment_id -> eligible driver IDs
}
```

#### Capabilities

- Broadcast shipment requests to eligible drivers
- Collect and rank bids using multi-criteria scoring
- Match shippers with optimal drivers
- Enforce business rules and policies
- Handle timeout scenarios (no bids, no selection)
- Coordinate booking confirmations

#### Communication Patterns

**Publishes**: `shipment.broadcast`, `booking.confirmed`, `match.failed`  
**Subscribes**: `shipment.requested`, `bid.submitted`, `bid.selected`

---

### 5. Route Optimizer Agent

**Agent ID**: `ROUTE-{INSTANCE}` (e.g., `ROUTE-001`)

#### Capabilities

- Calculate optimal routes using traffic data
- Optimize multi-stop delivery sequences
- Monitor traffic and suggest re-routing
- Calculate accurate ETAs
- Identify return cargo opportunities

#### Key Algorithms

- **Single Route**: A* with traffic weights
- **Multi-Stop**: Greedy + 2-opt TSP optimization
- **ETA**: Historical data + real-time traffic

---

### 6. Payment Agent

**Agent ID**: `PAY-{INSTANCE}` (e.g., `PAY-001`)

#### Capabilities

- Process payment holds when shipment booked
- Release payment after delivery confirmation
- Handle refunds for cancellations
- Manage dispute escrow
- Generate invoices and receipts

#### State Machine

```
Pending → Held → Released
             ↓
          Refunded / Disputed
```

---

### 7. Conflict Resolver Agent

**Agent ID**: `RESOLVE-{INSTANCE}` (e.g., `RESOLVE-001`)

#### Capabilities

- Mediate disputes between parties
- Coordinate emergency re-routing
- Facilitate cargo transfers
- Escalate to human support

#### Decision Logic

- Auto-resolve: Minor delays with documentation
- Escalate: Complex disputes, safety issues

---

### 8. Analytics Agent

**Agent ID**: `ANALYTICS-001`

#### Capabilities

- Calculate KPIs (on-time rate, utilization)
- Identify high-performing drivers/facilities
- Predict demand patterns
- Generate reports
- Detect anomalies (fraud, abuse)

#### Metrics Tracked

- Platform utilization
- Average bid response time
- On-time delivery rate
- Driver/shipper retention
- Revenue per shipment

---

## Agent Relationships

```
Agent Interaction Diagram:

                    ┌──────────────────┐
                    │  Shipper Agent   │
                    └────────┬─────────┘
                             │
                1. shipment.requested
                             │
                             ▼
                    ┌──────────────────┐
                    │ Platform Agent   │ ←─── (Orchestrator)
                    └────────┬─────────┘
                             │
                2. shipment.broadcast
                             │
           ┌─────────────────┼─────────────────┐
           │                 │                 │
           ▼                 ▼                 ▼
    ┌────────────┐    ┌────────────┐    ┌────────────┐
    │ Driver #1  │    │ Driver #2  │    │ Driver #N  │
    └──────┬─────┘    └──────┬─────┘    └──────┬─────┘
           │                 │                 │
           │    3. bid.submitted (all)        │
           └─────────────────┼─────────────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Platform Agent   │
                    │ (Rank bids)      │
                    └────────┬─────────┘
                             │
                4. Present top bids
                             │
                             ▼
                    ┌──────────────────┐
                    │  Shipper Agent   │
                    │ (Select winner)  │
                    └────────┬─────────┘
                             │
                5. bid.selected
                             │
                    ┌────────┴─────────┐
                    │                  │
                    ▼                  ▼
           ┌────────────────┐  ┌──────────────┐
           │  Driver Agent  │  │Facility Agent│
           │  (Execute)     │  │(Schedule dock)│
           └────────┬───────┘  └──────────────┘
                    │
                    │
           ┌────────┼─────────┐
           │                  │
           ▼                  ▼
  ┌────────────────┐  ┌──────────────┐
  │ Route Optimizer│  │Payment Agent │
  │ (Navigation)   │  │(Hold payment)│
  └────────────────┘  └──────────────┘
```

## Agent Lifecycle Management

### Lifecycle States

1. **Uninitialized**: Agent definition exists but not instantiated
2. **Initializing**: Loading state from database, connecting to message bus
3. **Active**: Fully operational, processing messages
4. **Idle**: No activity, eligible for hibernation
5. **Hibernating**: State saved, removed from memory
6. **Terminating**: Gracefully shutting down
7. **Terminated**: Cleanup complete, resources released

### Agent Manager Responsibilities

- **Creation**: Instantiate agents on-demand
- **Monitoring**: Health checks every 30 seconds
- **Hibernation**: Hibernate after 30 minutes idle (domain agents only)
- **Recovery**: Restart crashed agents with exponential backoff
- **Cleanup**: Terminate and remove terminated agents

### Example Lifecycle Code

```go
type AgentLifecycle struct {
    agent     Agent
    state     string
    lastActive time.Time
    healthCheck *time.Ticker
}

func (lc *AgentLifecycle) Start() {
    lc.state = "initializing"
    lc.agent.LoadState()
    lc.agent.ConnectMessageBus()
    lc.state = "active"
    
    lc.healthCheck = time.NewTicker(30 * time.Second)
    go lc.MonitorHealth()
}

func (lc *AgentLifecycle) MonitorHealth() {
    for range lc.healthCheck.C {
        if time.Since(lc.lastActive) > 30*time.Minute {
            lc.Hibernate()
        }
    }
}

func (lc *AgentLifecycle) Hibernate() {
    lc.state = "hibernating"
    lc.agent.SaveState()
    lc.agent.DisconnectMessageBus()
    lc.healthCheck.Stop()
    // Agent removed from memory by garbage collector
}
```

## Communication Patterns

### Message Types

1. **Commands**: Direct requests (e.g., "Create shipment")
2. **Events**: State changes (e.g., "Shipment delivered")
3. **Queries**: Information requests (e.g., "Get driver location")

### Messaging Protocols

- **Direct Messaging**: Point-to-point via Redis lists
- **Publish-Subscribe**: Broadcast via Redis pub/sub
- **Request-Reply**: Synchronous queries via HTTP/gRPC

### Example Message Schema

```json
{
  "message_id": "msg-abc123",
  "message_type": "event",
  "event_name": "shipment.requested",
  "timestamp": "2025-10-22T14:30:00Z",
  "sender_id": "SHP-NAI-45782",
  "payload": {
    "shipment_id": "SHP-54782-001",
    "pickup_location": {"lat": -1.286, "lng": 36.817},
    "delivery_location": {"lat": -4.043, "lng": 39.668},
    "cargo_weight": 500,
    "service_level": "standard"
  }
}
```

## Related Documents

- [README](./README.md) - Use case overview
- [System Architecture](./system-architecture.md) - Overall system design
- [Use Case Specification](../../../1-SoftwareRequirements/requirements/use-cases/UC-LOG-001-smart-logistics-platform.md) - Requirements

---

**Maintained by**: Architecture Team  
**Review Cadence**: Weekly during design phase  
**Next Review**: October 29, 2025
