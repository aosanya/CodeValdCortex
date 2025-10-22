# UC-RIDE-001 Agent Design: RideLink Ride-Hailing Platform

**Version**: 1.0  
**Last Updated**: October 22, 2025  
**Status**: Design Phase

## Overview

RideLink employs **8 autonomous agent types** that collaborate to deliver sub-10-second ride matching, real-time navigation, dynamic pricing, and comprehensive safety monitoring. Each agent operates independently while coordinating through high-frequency message passing.

## Agent Type Summary

| Agent Type | Count | Lifecycle | Primary Responsibility |
|------------|-------|-----------|------------------------|
| **Rider Agent** | 100,000+ | On-demand (per session) | Request rides, track progress, make payments |
| **Driver Agent** | 20,000+ | On-demand (when online) | Accept rides, navigate, update GPS location |
| **Platform Agent** | 10-20 | Persistent (replicated) | Match riders with drivers using geospatial indexing |
| **Route Optimizer Agent** | 5-10 | Persistent (replicated) | Calculate routes, monitor traffic, update ETAs |
| **Surge Pricing Agent** | 5-10 | Persistent (per zone) | Monitor supply/demand, calculate pricing multipliers |
| **Safety Agent** | 3-5 | Persistent (24/7) | Monitor trips for anomalies, handle emergencies |
| **Payment Agent** | 2-3 | Persistent (leader-follower) | Process payments, handle refunds, manage disputes |
| **Analytics Agent** | 1-2 | Persistent (batch + real-time) | Track KPIs, predict demand, optimize operations |

---

## 1. Rider Agent

**Purpose**: Represents a passenger requesting and tracking rides

**Agent ID Format**: `RDR-{CITY}-{SEQUENTIAL}` (e.g., `RDR-NAI-78945`)

### Attributes

```go
type RiderAgent struct {
    agent_id          string
    user_id           string
    full_name         string
    phone             string
    email             string
    payment_methods   []PaymentMethod  // M-Pesa, cards, cash
    
    // Preferences
    default_pickup    GeoPoint
    saved_locations   map[string]GeoPoint  // "Home", "Work"
    preferred_language string
    
    // Reputation
    rating            float64  // 1.0-5.0 from drivers
    total_rides       int
    cancellation_rate float64
    no_show_count     int
    
    // Current State
    current_state     string  // Idle, Requesting, Matched, InRide
    active_ride_id    string
}
```

### State Machine

```
Idle → Requesting (create ride request)
Requesting → Matched (driver accepts)
Requesting → Idle (timeout/cancel)
Matched → InRide (driver starts trip)
Matched → Idle (cancel before pickup)
InRide → Idle (trip completed)
```

### Key Capabilities

- **Create Ride Requests**: Specify pickup, destination, ride type
- **Track Driver Approach**: Monitor real-time ETA
- **Monitor Trip Progress**: Live GPS tracking
- **Make Payments**: Auto-debit or confirm cash payment
- **Rate Driver**: Provide feedback after trip

### Communication

**Publishes**: `ride.requested`, `ride.cancelled`, `rating.submitted`  
**Subscribes**: `driver.matched`, `driver.location`, `trip.started`, `trip.completed`

---

## 2. Driver Agent

**Purpose**: Represents a driver providing ride services

**Agent ID Format**: `DRV-{CITY}-{SEQUENTIAL}` (e.g., `DRV-NAI-34521`)

### Attributes

```go
type DriverAgent struct {
    agent_id          string
    user_id           string
    full_name         string
    phone             string
    license_number    string
    
    // Vehicle
    vehicle_id        string
    vehicle_type      string  // "economy", "comfort", "xl", "premium"
    make_model        string
    year              int
    license_plate     string
    color             string
    
    // Current Status
    current_location  GeoPoint
    availability      string  // "offline", "available", "on_trip"
    heading           float64 // Compass direction (0-360)
    speed_kmh         float64
    
    // Reputation
    rating            float64  // 1.0-5.0 from riders
    total_trips       int
    acceptance_rate   float64
    cancellation_rate float64
    
    // State
    current_state     string  // Offline, Available, EnRoute, OnTrip
    active_ride_id    string
}
```

### State Machine

```
Offline → Available (driver goes online)
Available → EnRoute (accepts ride request)
Available → Offline (driver goes offline)
EnRoute → OnTrip (picks up rider, starts trip)
EnRoute → Available (rider cancels)
OnTrip → Available (trip completed)
```

### Key Capabilities

- **Receive Ride Requests**: Get notified of nearby opportunities
- **Accept/Decline Rides**: Autonomous decision-making (3-15 seconds)
- **Navigate to Pickup**: Follow turn-by-turn directions
- **Stream GPS Location**: Update every 5 seconds during trips
- **Execute Trip**: Pick up, navigate, drop off
- **Collect Payment**: For cash rides
- **Rate Rider**: Provide feedback

### Example Behavior

```go
// Driver decides whether to accept ride request
func (d *DriverAgent) EvaluateRideRequest(req RideRequest) bool {
    // Check distance to pickup
    distanceToPickup := d.CalculateDistance(d.current_location, req.PickupLocation)
    if distanceToPickup > 5.0 {  // > 5 km
        return false  // Too far
    }
    
    // Check trip profitability
    estimatedFare := req.EstimatedFare
    estimatedTime := req.EstimatedDuration
    minimumAcceptableFare := 200.0  // KSH
    
    if estimatedFare < minimumAcceptableFare {
        return false  // Not profitable
    }
    
    // Check destination desirability
    // (prefer trips toward home at end of shift)
    if d.IsEndOfShift() {
        distanceToHome := d.CalculateDistance(req.DropoffLocation, d.home_location)
        if distanceToHome > 10.0 {  // > 10 km from home
            return false  // Going wrong direction
        }
    }
    
    // Check rider rating
    if req.RiderRating < 4.0 {
        return false  // Low-rated rider
    }
    
    return true  // Accept ride
}
```

### Communication

**Publishes**: `ride.accepted`, `location.updated`, `trip.started`, `trip.completed`  
**Subscribes**: `ride.broadcast`, `route.updated`, `surge.active`

---

## 3. Platform Agent

**Purpose**: Orchestrate ride matching using geospatial indexing

**Agent ID**: `PLAT-{INSTANCE}` (e.g., `PLAT-001`)

### Key Capabilities

- **Geospatial Matching**: Redis GEORADIUS for proximity queries
- **Broadcast Ride Requests**: Send to top 5 closest drivers
- **Handle Timeouts**: Expand search radius if no acceptance
- **Coordinate Cancellations**: Free up drivers, refund riders

### Matching Algorithm

```go
func (p *PlatformAgent) MatchRiderWithDriver(req RideRequest) (*Driver, error) {
    // Phase 1: Initial broadcast (2 km radius)
    drivers := p.FindEligibleDrivers(req.PickupLocation, 2.0, req.VehicleType)
    if len(drivers) == 0 {
        return nil, errors.New("No drivers nearby")
    }
    
    // Filter by rating and availability
    eligibleDrivers := p.FilterDrivers(drivers, rating >= 4.5, acceptance_rate >= 0.7)
    
    // Sort by proximity and broadcast to top 5
    topDrivers := p.SortByProximity(eligibleDrivers)[:min(5, len(eligibleDrivers))]
    p.BroadcastRideRequest(req, topDrivers)
    
    // Wait for acceptance (15 second timeout)
    driver, err := p.WaitForAcceptance(req.ID, 15*time.Second)
    if err == nil {
        return driver, nil
    }
    
    // Phase 2: Expand radius to 5 km
    drivers = p.FindEligibleDrivers(req.PickupLocation, 5.0, req.VehicleType)
    // ... repeat process
    
    // Phase 3: Suggest surge pricing or scheduled ride
    return nil, errors.New("No drivers available")
}
```

### Communication

**Publishes**: `ride.broadcast`, `match.confirmed`, `match.failed`  
**Subscribes**: `ride.requested`, `ride.accepted`, `driver.location`

---

## 4. Route Optimizer Agent

**Purpose**: Calculate optimal routes and provide real-time navigation

**Agent ID**: `ROUTE-{INSTANCE}` (e.g., `ROUTE-001`)

### Key Capabilities

- **Route Calculation**: A* algorithm + Google Maps traffic data
- **ETA Updates**: Recalculate every 30 seconds during trips
- **Dynamic Re-routing**: Suggest faster routes when traffic changes
- **Shared Ride Optimization**: Multi-stop TSP for ride-sharing

### Algorithm

```go
func (r *RouteOptimizerAgent) CalculateRoute(origin, destination GeoPoint) Route {
    // Query Google Maps Directions API
    response := r.googleMaps.Directions(origin, destination, "driving", "best_guess")
    
    // Parse response
    route := Route{
        Distance:       response.Routes[0].Legs[0].Distance.Value,  // meters
        Duration:       response.Routes[0].Legs[0].Duration.Value,  // seconds
        Polyline:       response.Routes[0].OverviewPolyline.Points,
        Steps:          r.ParseSteps(response.Routes[0].Legs[0].Steps),
        TrafficFactor:  response.Routes[0].Legs[0].DurationInTraffic / Duration,
    }
    
    // Cache route (TTL: 10 minutes)
    r.CacheRoute(origin, destination, route)
    
    return route
}

// Update ETA based on current location
func (r *RouteOptimizerAgent) UpdateETA(currentLocation, destination GeoPoint) time.Duration {
    remainingRoute := r.CalculateRoute(currentLocation, destination)
    return time.Duration(remainingRoute.Duration) * time.Second
}
```

### Communication

**Publishes**: `route.calculated`, `eta.updated`, `route.reroute_suggested`  
**Subscribes**: `trip.started`, `driver.location`

---

## 5. Surge Pricing Agent

**Purpose**: Monitor supply/demand and calculate dynamic pricing

**Agent ID**: `SURGE-{ZONE}` (e.g., `SURGE-NAI-CBD`)

### Key Capabilities

- **Demand Monitoring**: Track ride requests per zone per minute
- **Supply Monitoring**: Count available drivers per zone
- **Multiplier Calculation**: Apply surge when demand > supply
- **Driver Incentives**: Notify drivers of surge zones (heat map)

### Algorithm

```go
func (s *SurgePricingAgent) CalculateSurgeMultiplier(zone string) float64 {
    // Count ride requests in last 2 minutes
    demand := s.CountRideRequests(zone, 2*time.Minute)
    
    // Count available drivers in zone
    supply := s.CountAvailableDrivers(zone)
    
    // Calculate ratio
    ratio := float64(demand) / float64(supply)
    
    // Determine multiplier
    var multiplier float64
    switch {
    case ratio < 1.2:
        multiplier = 1.0  // No surge
    case ratio < 2.0:
        multiplier = 1.2  // Light surge
    case ratio < 3.0:
        multiplier = 1.5  // Moderate surge
    default:
        multiplier = 2.0  // High surge (capped at 2.0×)
    }
    
    // Smooth transitions (don't jump directly)
    currentMultiplier := s.GetCurrentMultiplier(zone)
    if math.Abs(multiplier - currentMultiplier) > 0.2 {
        // Gradual change (0.1 step per 2-minute interval)
        if multiplier > currentMultiplier {
            multiplier = currentMultiplier + 0.1
        } else {
            multiplier = currentMultiplier - 0.1
        }
    }
    
    s.SetMultiplier(zone, multiplier)
    return multiplier
}
```

### Communication

**Publishes**: `surge.active`, `surge.ended`, `heat_map.updated`  
**Subscribes**: `ride.requested`, `ride.completed`, `driver.location`

---

## 6. Safety Agent

**Purpose**: Monitor trips for anomalies and handle emergencies

**Agent ID**: `SAFETY-{INSTANCE}` (e.g., `SAFETY-001`)

### Key Capabilities

- **Anomaly Detection**: Monitor speed, route deviation, unexpected stops
- **Emergency Response**: Handle SOS activations (<3-second response)
- **Trusted Contact Notification**: Auto-notify for night rides
- **Incident Management**: Coordinate with authorities and support team

### Monitoring Rules

```go
func (s *SafetyAgent) MonitorTrip(trip Trip) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Check route deviation
            deviation := s.CalculateRouteDeviation(trip)
            if deviation > 0.5 {  // > 500 meters off route
                s.RaiseAlert(trip, "Route deviation detected", "MEDIUM")
            }
            
            // Check speed
            if trip.CurrentSpeed > trip.SpeedLimit + 20 {  // 20 km/h over limit
                s.RaiseAlert(trip, "Excessive speed", "HIGH")
            }
            
            // Check for unexpected stops
            if trip.CurrentSpeed == 0 && trip.ElapsedTime > 5*time.Minute {
                timeStopped := time.Since(trip.LastMovementTime)
                if timeStopped > 5*time.Minute {
                    s.RaiseAlert(trip, "Prolonged stop detected", "MEDIUM")
                }
            }
            
            // Exit monitoring when trip completes
            if trip.Status == "COMPLETED" {
                return
            }
        }
    }
}
```

### Emergency SOS Flow

```go
func (s *SafetyAgent) HandleEmergencySOS(trip Trip, activatedBy string) {
    // 1. Record event immediately
    s.RecordSOSEvent(trip, activatedBy, time.Now())
    
    // 2. Capture current state
    snapshot := s.CaptureTrip Snapshot(trip)  // GPS, speed, driver info
    
    // 3. Notify trusted contacts (if configured)
    if len(trip.TrustedContacts) > 0 {
        s.NotifyTrustedContacts(trip, snapshot)
    }
    
    // 4. Alert safety team (24/7 hotline)
    s.AlertSafetyTeam(trip, snapshot, "CRITICAL")
    
    // 5. Start audio recording (if permission granted)
    s.StartAudioRecording(trip)
    
    // 6. Display reassurance message to user
    s.SendMessageToApp(activatedBy, "Help is on the way. We've alerted authorities.")
    
    // 7. If confirmed emergency (not accidental tap):
    if s.ConfirmEmergency(trip, 30*time.Second) {
        // Contact local authorities
        s.AlertAuthorities(trip, snapshot)
        
        // Lock driver account pending investigation
        s.SuspendDriverAccount(trip.DriverID)
    }
    
    // 8. Flag trip for review
    s.FlagForHumanReview(trip, "SOS_ACTIVATED")
}
```

### Communication

**Publishes**: `alert.raised`, `sos.activated`, `incident.reported`  
**Subscribes**: `trip.started`, `driver.location`, `sos.button_pressed`

---

## 7. Payment Agent

**Purpose**: Process payments, manage refunds, handle disputes

**Agent ID**: `PAY-{INSTANCE}` (e.g., `PAY-001`)

### Key Capabilities

- **Payment Processing**: M-Pesa, Stripe (cards), cash
- **Automatic Charging**: Charge rider after trip completion
- **Driver Payouts**: Weekly transfers to driver accounts
- **Refund Management**: Handle cancellations and disputes
- **Fraud Detection**: Anomaly detection for suspicious transactions

### Payment Flow

```go
func (p *PaymentAgent) ProcessPayment(trip Trip) error {
    // Calculate final fare
    fare := p.CalculateFare(trip)
    
    // Apply promo codes
    if trip.PromoCode != "" {
        discount := p.ApplyPromoCode(trip.PromoCode)
        fare -= discount
    }
    
    // Process payment based on method
    switch trip.PaymentMethod {
    case "m-pesa":
        return p.ChargeMPesa(trip.RiderID, fare, trip.ID)
    case "card":
        return p.ChargeCard(trip.RiderID, fare, trip.ID)
    case "cash":
        return p.ConfirmCashPayment(trip.DriverID, fare, trip.ID)
    }
    
    return nil
}

// Driver payout calculation
func (p *PaymentAgent) CalculateDriverPayout(trip Trip) float64 {
    fare := trip.FinalFare
    platformCommission := fare * 0.25  // 25%
    driverEarnings := fare - platformCommission
    return driverEarnings
}
```

### Communication

**Publishes**: `payment.success`, `payment.failed`, `payout.processed`  
**Subscribes**: `trip.completed`, `trip.cancelled`, `dispute.raised`

---

## 8. Analytics Agent

**Purpose**: Track performance, predict demand, optimize operations

**Agent ID**: `ANALYTICS-001`

### Key Capabilities

- **KPI Tracking**: Rides/day, matching time, on-time rate
- **Demand Prediction**: ML models for hourly/daily forecasting
- **Driver Performance**: Identify top performers and issues
- **Anomaly Detection**: Unusual patterns (fraud, abuse)
- **Reporting**: Executive dashboards, operational insights

### Key Metrics

```go
// Real-time metrics
rides_requested_per_minute
rides_matched_per_minute
average_matching_time_seconds
surge_zones_active_count
drivers_online_count

// Business metrics
daily_active_riders
daily_active_drivers
revenue_per_ride
driver_utilization_rate
rider_retention_rate
```

### Communication

**Publishes**: `report.generated`, `anomaly.detected`, `prediction.updated`  
**Subscribes**: All events (comprehensive data collection)

---

## Agent Relationship Diagram

```
           [Rider Agent]
                 │
                 │ 1. ride.requested
                 ▼
          [Platform Agent] ←──── 2. Geospatial query
                 │                   (Redis GEORADIUS)
                 │ 3. ride.broadcast
                 │
        ┌────────┼────────┐
        │                 │
        ▼                 ▼
   [Driver 1]        [Driver N]
        │                 │
        │ 4. ride.accepted (First wins)
        ▼                 │
  [Platform Agent] ←──────┘
        │
        ├─────────────┬─────────────┬─────────────┐
        ▼             ▼             ▼             ▼
  [Route Opt.]  [Surge Pricing] [Safety Agt] [Payment Agt]
  (Navigation)  (Check surge)   (Monitor)    (Hold payment)
        │
        │ 5. GPS updates (every 5s)
        ▼
   [Safety Agent] ← Continuous monitoring
        │
        │ 6. Trip completion
        ▼
  [Payment Agent] → Process payment
        │
        ▼
  [Analytics Agent] → Update KPIs
```

## Agent Lifecycle

- **Domain Agents** (Rider, Driver): On-demand, hibernate after 30 min idle
- **System Agents** (Platform, Route, Surge, Safety, Payment): Persistent, replicated
- **Analytics Agent**: Batch processing + real-time stream processing

---

## Related Documents

- [README](./README.md) - Use case overview
- [System Architecture](./system-architecture.md) - Overall system design
- [Use Case Specification](../../../1-SoftwareRequirements/requirements/use-cases/UC-RIDE-001-ride-hailing-platform.md)

---

**Maintained by**: Architecture Team  
**Review Cadence**: Weekly  
**Next Review**: October 29, 2025
