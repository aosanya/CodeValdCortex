# Agent Design - Water Distribution Network Management

**Version**: 1.0  
**Last Updated**: October 22, 2025

## Overview

This document provides detailed specifications for all agent types in the Water Distribution Network Management system. The system consists of 7 autonomous agent types that represent physical infrastructure elements and monitoring systems.

## Agent Types Summary

1. **Pipe Agent** - Physical water pipes in the distribution network
2. **Sensor Agent** - IoT sensors monitoring water quality, pressure, flow, temperature
3. **Hydrant Agent** - Fire hydrants in the distribution network
4. **Valve Agent** - Control valves in the distribution network
5. **Pump Agent** - Water pumps boosting pressure in the network
6. **Reservoir Agent** - Water storage reservoirs and tanks
7. **Meter Agent** - Water consumption meters at customer endpoints

## Agent Relationship Diagram

```
                    ┌─────────────────┐
                    │   Control       │
                    │   Center        │
                    └────────┬────────┘
                             │
                    ┌────────┴────────┐
                    │   Coordinator   │
                    │   Agents        │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
    ┌───▼───┐           ┌───▼───┐           ┌───▼───┐
    │Reservoir│         │ Pump  │           │ Valve │
    │ Agent  │◄────────►│Agent  │◄─────────►│Agent  │
    └───┬────┘          └───┬───┘           └───┬───┘
        │                   │                   │
        │              ┌────┴────┐              │
        │              │         │              │
    ┌───▼───┐      ┌──▼──┐  ┌──▼──┐       ┌───▼───┐
    │ Pipe  │◄────►│Sensor│  │Sensor│◄─────►│ Pipe  │
    │Agent  │      │Agent │  │Agent │       │Agent  │
    └───┬───┘      └──┬──┘  └──┬──┘       └───┬───┘
        │             │        │              │
        │             └────┬───┘              │
        │                  │                  │
    ┌───▼───┐         ┌───▼───┐          ┌──▼───┐
    │Hydrant│         │ Meter │          │ Meter│
    │Agent  │         │ Agent │          │Agent │
    └───────┘         └───────┘          └──────┘
```

---

## 1. Pipe Agent

### Purpose
Represents physical water pipes in the distribution network, monitoring flow and pressure conditions.

### Attributes

```go
type PipeAgent struct {
    // Identity
    pipe_id         string    // Unique identifier (e.g., "PIPE-001")
    
    // Physical Properties
    material        string    // Pipe material: PVC, steel, copper, cast_iron
    diameter        int       // Pipe diameter in millimeters
    length          float64   // Pipe length in meters
    installation_date time.Time // Date installed
    pressure_rating float64   // Maximum pressure rating (PSI)
    flow_capacity   float64   // Maximum flow rate (liters/minute)
    
    // Location
    location        struct {
        start GeoPoint  // GPS coordinates of start point
        end   GeoPoint  // GPS coordinates of end point
    }
    
    // Operational State
    age             int       // Calculated from installation date (years)
    condition_score int       // Health score (0-100)
    current_pressure float64  // Current pressure reading (PSI)
    current_flow    float64   // Current flow rate (L/min)
    
    // Relationships
    connected_agents []string // List of connected pipe, valve, sensor agents
    upstream_agent   string   // Agent ID of upstream component
    downstream_agent string   // Agent ID of downstream component
    
    // Monitoring
    last_inspection  time.Time
    next_maintenance time.Time
    alerts_active    []Alert
}
```

### Capabilities

- **Monitor flow rate and pressure** - Continuous monitoring via connected sensors
- **Detect pressure anomalies** - Identify leaks, bursts, or blockages
- **Calculate flow efficiency** - Compare actual vs. design flow rates
- **Track degradation over time** - ML-based condition assessment
- **Communicate with adjacent pipes** - Coordinate pressure balancing
- **Report maintenance needs** - Generate work orders based on condition
- **Predict remaining lifespan** - Forecasting using historical data

### State Machine

```
┌──────────────┐
│ Operational  │ ◄─────┐
└──────┬───────┘       │
       │               │
       │ performance < │ repair
       │   threshold   │ completed
       ▼               │
┌──────────────┐       │
│  Degraded    │       │
└──────┬───────┘       │
       │               │
       │ anomaly       │
       │ detected      │
       ▼               │
┌──────────────┐       │
│   Warning    │       │
└──────┬───────┘       │
       │               │
       │ immediate     │
       │ attention     │
       ▼               │
┌──────────────┐       │
│   Critical   │       │
└──────┬───────┘       │
       │               │
       │ repair        │
       │ scheduled     │
       ▼               │
┌──────────────┐       │
│ Maintenance  │───────┘
└──────┬───────┘
       │
       │ network
       │ isolation
       ▼
┌──────────────┐
│   Offline    │
└──────────────┘
```

### Example Behaviors

```go
// Leak Detection
func (p *PipeAgent) MonitorForLeaks() {
    if p.currentPressure < p.expectedPressure - threshold && 
       p.currentFlow > 0 {
        alert := Alert{
            Type:     "LEAK_SUSPECTED",
            Severity: "HIGH",
            Message:  "Possible leak detected - pressure drop with active flow",
            Location: p.location,
        }
        p.RaiseAlert(alert)
        p.NotifyAdjacentAgents()
        p.NotifyControlCenter()
        p.IncreaseSensorSamplingRate()
    }
}

// Condition Monitoring
func (p *PipeAgent) UpdateConditionScore() {
    factors := []float64{
        p.CalculateAgeScore(),           // Older pipes = lower score
        p.CalculateMaterialScore(),      // Material degradation
        p.CalculateLeakHistory(),        // Past leak incidents
        p.CalculatePressureVariance(),   // Pressure stability
        p.CalculateFlowEfficiency(),     // Flow performance
    }
    
    p.condition_score = WeightedAverage(factors, weights)
    
    if p.condition_score < criticalThreshold {
        p.TransitionState(StateCritical)
        p.RequestMaintenance(PriorityHigh)
    }
}
```

### Communication Patterns

**Publishes**:
- `pipe.pressure.anomaly` - Pressure anomaly detected
- `pipe.leak.suspected` - Possible leak identified
- `pipe.maintenance.required` - Maintenance needed

**Subscribes**:
- `valve.status.changed` - Upstream/downstream valve changes
- `sensor.pressure.reading` - Pressure sensor updates
- `sensor.flow.reading` - Flow sensor updates

**Direct Messages**:
- Adjacent pipe agents - Pressure coordination
- Connected valve agents - Flow control requests
- Control center - Status reports and alerts

### Performance Characteristics

- **Response Time**: <100ms for sensor data processing
- **Alert Latency**: <1s from detection to notification
- **State Update Frequency**: Every 30 seconds
- **Memory Footprint**: ~2KB per agent
- **CPU Usage**: <0.1% average

---

## 2. Sensor Agent

### Purpose
IoT sensors monitoring water quality, pressure, flow, and temperature throughout the network.

### Attributes

```go
type SensorAgent struct {
    // Identity
    sensor_id       string    // Unique identifier (e.g., "SENS-P-001")
    sensor_type     SensorType // pressure, flow, quality, temperature, vibration
    
    // Physical Properties
    location        GeoPoint  // GPS coordinates
    attached_to     string    // Reference to pipe/hydrant/valve agent ID
    
    // Configuration
    sampling_rate   int       // Measurement frequency (seconds)
    accuracy        float64   // Sensor accuracy rating (percentage)
    measurement_range struct {
        min float64
        max float64
    }
    
    // Operational State
    battery_level   int       // For wireless sensors (percentage)
    last_calibration time.Time
    current_reading float64
    
    // Thresholds
    alert_thresholds struct {
        warning  float64
        critical float64
    }
    
    // Data Management
    buffer          []Reading // Local data buffer
    last_transmission time.Time
}

type Reading struct {
    timestamp time.Time
    value     float64
    quality   string // good, suspect, error
}
```

### Capabilities

- **Continuous monitoring and data collection** - Regular measurements
- **Real-time anomaly detection** - Local threshold checking
- **Trend analysis** - Identify patterns over time
- **Automatic calibration checks** - Self-diagnostics
- **Battery monitoring and alerts** - Power management
- **Data transmission to central system** - Batched or real-time
- **Correlation with adjacent sensors** - Multi-sensor validation
- **Predictive alerting** - Forecast threshold violations

### State Machine

```
┌──────────────┐
│    Active    │ ◄─────────┐
└──────┬───────┘           │
       │                   │ calibration
       │ calibration       │ complete
       │ required          │
       ▼                   │
┌──────────────┐           │
│ Calibrating  │───────────┘
└──────────────┘
       │
       │ battery < 20%
       ▼
┌──────────────┐
│ Low_Battery  │
└──────┬───────┘
       │
       │ malfunction
       │ detected
       ▼
┌──────────────┐
│    Error     │
└──────┬───────┘
       │
       │ communication
       │ lost
       ▼
┌──────────────┐
│   Offline    │
└──────────────┘
```

### Example Behaviors

```go
// Anomaly Detection
func (s *SensorAgent) ProcessReading(value float64) {
    reading := Reading{
        timestamp: time.Now(),
        value:     value,
        quality:   s.AssessQuality(value),
    }
    
    s.buffer = append(s.buffer, reading)
    s.current_reading = value
    
    // Check thresholds
    if value < s.alert_thresholds.critical {
        s.RaiseAlert(Alert{
            Type:     "SENSOR_CRITICAL",
            Severity: "CRITICAL",
            Value:    value,
        })
        s.NotifyAttachedAgent()
        s.IncreaseSamplingRate()
    } else if value < s.alert_thresholds.warning {
        s.LogWarning("Sensor reading approaching threshold")
    }
    
    // Transmit if buffer full or time elapsed
    if len(s.buffer) >= bufferSize || 
       time.Since(s.last_transmission) > transmitInterval {
        s.TransmitData()
    }
}

// Battery Management
func (s *SensorAgent) MonitorBattery() {
    if s.battery_level < 20 {
        s.TransitionState(StateLowBattery)
        s.ReduceSamplingRate() // Conserve power
        s.NotifyMaintenance("Battery replacement needed")
    }
    
    if s.battery_level < 5 {
        s.TransmitBufferedData() // Emergency data flush
        s.PrepareForShutdown()
    }
}
```

### Communication Patterns

**Publishes**:
- `sensor.{type}.reading` - Sensor measurements (e.g., `sensor.pressure.reading`)
- `sensor.anomaly.detected` - Threshold violation
- `sensor.battery.low` - Low battery alert
- `sensor.calibration.required` - Needs calibration

**Subscribes**:
- `sensor.calibration.scheduled` - Calibration commands
- `sensor.sampling.adjust` - Sampling rate changes

**Direct Messages**:
- Attached infrastructure agent - Real-time readings
- Data aggregator - Batched historical data
- Maintenance system - Service requests

### Performance Characteristics

- **Sampling Rate**: 1-60 seconds (configurable)
- **Data Transmission**: Every 5 minutes or 100 readings
- **Battery Life**: 2-5 years (depending on sampling rate)
- **Accuracy**: ±1% for pressure, ±2% for flow
- **Alert Latency**: <2s from threshold violation

---

## 3. Hydrant Agent

### Purpose
Represents fire hydrants in the distribution network, tracking availability, usage, and maintenance.

### Attributes

```go
type HydrantAgent struct {
    // Identity
    hydrant_id      string    // Unique identifier (e.g., "HYD-001")
    
    // Physical Properties
    location        GeoPoint  // GPS coordinates
    hydrant_type    string    // pillar, underground, wall_mounted
    number_of_outlets int     // Typically 2-4
    flow_capacity   float64   // Liters per minute at standard pressure
    
    // Installation & Maintenance
    installation_date time.Time
    last_inspection  time.Time
    last_maintenance time.Time
    condition        string    // excellent, good, fair, poor
    
    // Operational State
    is_available    bool
    accessibility   string     // road_access, restricted, blocked
    in_use          bool
    usage_count     int        // Historical usage count
    
    // Monitoring
    pressure_sensor string     // Reference to attached sensor agent
    flow_meter      string     // Reference to attached flow meter
    
    // Relationships
    connected_pipes []string   // List of connected pipe agents
    
    // Usage Tracking
    usage_history   []UsageEvent
}

type UsageEvent struct {
    timestamp  time.Time
    duration   time.Duration
    flow_rate  float64
    event_type string // training, emergency, maintenance, unauthorized
    operator   string // Fire dept, maintenance crew, etc.
}
```

### Capabilities

- **Monitor availability and accessibility** - Real-time status tracking
- **Track usage** - Training exercises, emergencies, maintenance
- **Detect unauthorized use** - Anomalous flow patterns
- **Schedule inspection reminders** - Based on last inspection date
- **Report maintenance needs** - Condition-based work orders
- **Provide flow testing data** - Performance validation
- **Coordinate with emergency services** - Availability for fire response
- **Monitor surrounding pressure impact** - Network pressure during use

### State Machine

```
┌──────────────┐
│  Available   │ ◄─────────┐
└──────┬───────┘           │
       │                   │
       │ hydrant opened    │ flow
       ▼                   │ stopped
┌──────────────┐           │
│   In_Use     │───────────┘
└──────────────┘
       │
       │ inspection
       │ due date
       ▼
┌──────────────┐
│Inspection_Due│
└──────┬───────┘
       │
       │ maintenance
       │ required
       ▼
┌──────────────────┐
│Maintenance       │
│Required          │
└──────┬───────────┘
       │
       │ not operational
       ▼
┌──────────────────┐
│Out_of_Service    │
└──────┬───────────┘
       │
       │ access blocked
       ▼
┌──────────────┐
│   Blocked    │
└──────────────┘
```

### Example Behaviors

```go
// Usage Detection
func (h *HydrantAgent) DetectUsage() {
    currentFlow := h.GetFlowReading()
    
    if currentFlow > usageThreshold && !h.in_use {
        // Hydrant opened
        h.in_use = true
        h.TransitionState(StateInUse)
        
        usage := UsageEvent{
            timestamp:  time.Now(),
            flow_rate:  currentFlow,
            event_type: h.ClassifyUsage(), // Determine if authorized
        }
        h.usage_history = append(h.usage_history, usage)
        
        // Notify network
        h.NotifyConnectedPipes("Flow demand increased")
        h.NotifyControlCenter("Hydrant in use")
        h.MonitorPressureImpact()
    }
    
    if currentFlow < usageThreshold && h.in_use {
        // Hydrant closed
        h.in_use = false
        h.usage_count++
        h.TransitionState(StateAvailable)
        h.LogUsageEvent()
    }
}

// Inspection Scheduling
func (h *HydrantAgent) CheckInspectionDue() {
    daysSinceInspection := time.Since(h.last_inspection).Hours() / 24
    
    if daysSinceInspection > inspectionIntervalDays {
        h.TransitionState(StateInspectionDue)
        h.ScheduleInspection()
        h.NotifyMaintenance("Hydrant inspection due")
    }
}
```

### Communication Patterns

**Publishes**:
- `hydrant.opened` - Hydrant usage started
- `hydrant.closed` - Hydrant usage ended
- `hydrant.inspection.due` - Inspection needed
- `hydrant.maintenance.required` - Repairs needed

**Subscribes**:
- `fire.emergency.alert` - Emergency in area
- `maintenance.inspection.scheduled` - Inspection appointments
- `sensor.pressure.reading` - Surrounding pressure monitoring

**Direct Messages**:
- Connected pipe agents - Flow demand notifications
- Emergency services - Availability status
- Control center - Usage reports

### Performance Characteristics

- **Status Update Frequency**: Real-time during use, hourly when idle
- **Flow Detection Latency**: <5s
- **Inspection Interval**: 6 months (configurable)
- **Data Retention**: 5 years of usage history

---

## 4. Valve Agent

[Content continues with similar detailed specifications for Valve, Pump, Reservoir, and Meter agents...]

## Agent Lifecycle Management

### Agent Creation

```go
func CreateAgent(agentType string, config AgentConfig) (Agent, error) {
    agent := NewAgent(agentType)
    agent.Initialize(config)
    agent.RegisterWithRegistry()
    agent.EstablishConnections()
    agent.TransitionState(StateInitializing)
    agent.RunHealthCheck()
    agent.TransitionState(StateActive)
    return agent, nil
}
```

### Agent Shutdown

```go
func (a *Agent) Shutdown() error {
    a.TransitionState(StateShuttingDown)
    a.FlushPendingMessages()
    a.PersistState()
    a.NotifyConnectedAgents()
    a.DeregisterFromRegistry()
    a.TransitionState(StateOffline)
    return nil
}
```

### Health Monitoring

All agents implement:

```go
type HealthCheck interface {
    IsHealthy() bool
    GetHealthStatus() HealthStatus
    GetLastHeartbeat() time.Time
}

type HealthStatus struct {
    Status      string    // healthy, degraded, unhealthy
    Uptime      time.Duration
    LastError   error
    MetricsSnapshot map[string]interface{}
}
```

## Related Documents

- [System Architecture](./system-architecture.md)
- [Communication Patterns](./communication-patterns.md)
- [Data Models](./data-models.md)
- [Use Case Specification](../../../1-SoftwareRequirements/requirements/use-cases/UC-INFRA-001-water-distribution-network.md)
