# Use Case: CodeValdStima - Electric Power Distribution Network Management

**Use Case ID**: UC-INFRA-002  
**Use Case Name**: Electric Power Distribution Network Agent System  
**System**: CodeValdStima  
**Created**: October 24, 2025  
**Status**: Example/Reference Implementation

## Overview

CodeValdStima is an example agentic system built on the CodeValdCortex framework that demonstrates how electric power distribution infrastructure can be represented and managed as autonomous agents. This use case focuses on an electric power distribution network where transformers, power lines, substations, smart meters, circuit breakers, and other grid components are modeled as intelligent agents that monitor, communicate, and respond to network conditions.

**Note**: *"Stima" means "electricity" or "electric power" in Swahili, reflecting the system's focus on electrical energy distribution.*

## System Context

### Domain
Municipal electric power distribution infrastructure management and smart grid operations

### Business Problem
Traditional electric power distribution systems suffer from:
- Reactive maintenance (fixing outages after they occur)
- Limited real-time visibility into grid status and load distribution
- Inefficient load balancing and energy routing
- Delayed fault detection and response
- Poor coordination between grid components during demand spikes
- Manual inspection and meter reading requirements
- Difficulty predicting equipment failures and power quality issues
- Energy losses due to inefficient distribution
- Slow integration of renewable energy sources

### Proposed Solution
An agentic system where each grid infrastructure element is an autonomous agent that:
- Monitors its own state, load, and performance metrics
- Communicates with neighboring grid agents
- Detects faults, voltage anomalies, and power quality issues
- Coordinates responses to grid events and demand changes
- Provides real-time data for decision-making and load balancing
- Enables predictive maintenance and fault prevention
- Optimizes energy distribution and reduces losses
- Facilitates renewable energy integration

## Agent Types

### 1. Power Line Agent
**Represents**: Electrical transmission and distribution lines (feeders) in the power grid

**Attributes**:
- `line_id`: Unique identifier (e.g., "LINE-001")
- `line_type`: overhead, underground, high_voltage, medium_voltage, low_voltage
- `voltage_rating`: Nominal voltage (kV)
- `conductor_type`: Material (aluminum, copper, ACSR)
- `length`: Line length in meters
- `ampacity`: Maximum current capacity (amperes)
- `installation_date`: Date installed
- `coordinates`: GPS coordinates array [[start_lat, start_lon], [end_lat, end_lon]]
- `location`: GPS coordinates (start and end points) - human readable
- `phase_configuration`: single_phase, three_phase
- `connected_nodes`: List of connected substation, transformer, and sensor agents
- `impedance`: Line resistance and reactance values
- `temperature_rating`: Maximum operating temperature (°C)
- `condition_score`: Health score (0-100)
- `sag_level`: Current conductor sag measurement
- `connection_rules`: Array of connection definitions with canonical_type
  - `{ target_type: "power_line", canonical_type: "route", directionality: "directed" }`
  - `{ target_type: "transformer", canonical_type: "supply", directionality: "directed" }`
  - `{ target_type: "circuit_breaker", canonical_type: "depends_on", directionality: "directed" }`
- `visualization_metadata`: Display properties for topology visualizer
  - `color`: Dynamic based on voltage_rating (red: high, yellow: medium, blue: low)
  - `width`: Dynamic based on voltage_rating
  - `style`: Solid for overhead, dashed for underground
  - `power_flow_animation`: Boolean indicating power flow visualization

**Capabilities**:
- Monitor current flow, voltage, and power factor
- Detect overcurrent, undervoltage, and overvoltage conditions
- Calculate line losses and efficiency
- Track thermal loading and temperature
- Communicate with adjacent line segments
- Report maintenance needs based on condition
- Predict remaining lifespan based on loading patterns
- Detect ground faults and short circuits

**State Machine**:
- `Operational` - Normal power transmission
- `Loaded` - Operating near capacity
- `Overloaded` - Exceeding safe capacity
- `Faulted` - Short circuit or ground fault detected
- `De-energized` - Offline for maintenance
- `Emergency` - Critical condition requiring immediate action

**Example Behavior**:
```
IF current > ampacity_threshold AND duration > 5_minutes
  THEN raise_alert("Line overload detected")
  AND notify_connected_transformers("Reduce load")
  AND notify_breaker_agents("Prepare for trip")
  AND calculate_load_transfer_options()
```

### 2. Transformer Agent
**Represents**: Distribution transformers stepping down voltage for customer use

**Attributes**:
- `transformer_id`: Unique identifier (e.g., "XFMR-001")
- `transformer_type`: pole_mounted, pad_mounted, substation, distribution
- `coordinates`: GPS coordinates [latitude, longitude]
- `location`: GPS coordinates - human readable
- `primary_voltage`: High voltage side (kV)
- `secondary_voltage`: Low voltage side (volts)
- `rated_capacity`: kVA rating
- `current_load`: Current power delivery (kW)
- `load_percentage`: Percentage of rated capacity
- `installation_date`: Date installed
- `cooling_type`: ONAN, ONAF, oil_filled, dry_type
- `temperature`: Oil/winding temperature (°C)
- `voltage_sensors`: References to voltage monitoring sensors
- `current_sensors`: References to current sensors
- `upstream_line`: Reference to feeding line agent
- `downstream_meters`: List of connected meter agents
- `tap_position`: Current tap changer position
- `efficiency`: Current operating efficiency (%)
- `connection_rules`: Array of connection definitions
  - `{ target_id: upstream_line, canonical_type: "depends_on", directionality: "directed" }`
  - `{ target_id: downstream_meters, canonical_type: "supply", directionality: "directed" }`
- `visualization_metadata`: Display properties
  - `icon`: Transformer type specific icon
  - `color`: Load-based (green: <80%, yellow: 80-90%, red: >90%)
  - `size`: Based on rated_capacity

**Capabilities**:
- Monitor load, voltage, and temperature continuously
- Regulate voltage through tap changing
- Detect overloading and overheating
- Calculate remaining capacity
- Coordinate with other transformers for load balancing
- Predict failures based on temperature and load patterns
- Schedule maintenance based on condition
- Optimize voltage regulation for power quality

**State Machine**:
- `Normal_Operation` - Operating within specifications
- `High_Load` - Load >80% of capacity
- `Overloaded` - Exceeding rated capacity
- `Overheating` - Temperature above safe threshold
- `Voltage_Regulation` - Active tap changing
- `Faulted` - Internal fault detected
- `Offline` - De-energized for maintenance

**Example Behavior**:
```
IF load_percentage > 90
  THEN notify_upstream_substation("Capacity constraint")
  AND notify_adjacent_transformers("Load transfer request")
  AND monitor_temperature_increase()
  
IF temperature > critical_threshold
  THEN reduce_load_shedding()
  AND emergency_shutdown_if_exceeded()
  AND notify_maintenance_team()
```

### 3. Smart Meter Agent
**Represents**: Advanced metering infrastructure (AMI) at customer premises

**Attributes**:
- `meter_id`: Unique identifier (e.g., "MTR-001")
- `customer_id`: Associated customer account
- `coordinates`: GPS coordinates [latitude, longitude]
- `location`: GPS coordinates or address - human readable
- `meter_type`: residential, commercial, industrial
- `phase_type`: single_phase, three_phase
- `connected_transformer`: Reference to transformer agent
- `current_demand`: Real-time power consumption (kW)
- `voltage`: Line voltage measurement (volts)
- `current`: Line current measurement (amperes)
- `power_factor`: Current power factor
- `energy_consumed`: Cumulative energy (kWh)
- `peak_demand`: Maximum demand in billing period
- `installation_date`: Date installed
- `last_reading_date`: Last successful communication
- `billing_cycle`: Monthly, bi-monthly
- `tariff_plan`: TOU (time-of-use), flat_rate, tiered
- `net_metering`: Boolean for solar/renewable connection
- `connection_rules`: Array of connection definitions
  - `{ target_id: connected_transformer, canonical_type: "depends_on", directionality: "directed" }`
  - `{ target_type: "renewable_energy_agent", canonical_type: "supply", directionality: "bidirectional" }`
- `visualization_metadata`: Display properties
  - `icon`: Meter type specific icon
  - `color`: Status-based (green: active, yellow: high consumption, red: outage)
  - `cluster`: Group residential meters for performance

**Capabilities**:
- Real-time energy consumption monitoring
- Automatic meter reading (AMR/AMI)
- Power quality monitoring (voltage sags, swells)
- Tamper detection and revenue protection
- Outage detection and reporting
- Demand response participation
- Time-of-use billing data generation
- Customer usage analytics and forecasting
- Support for distributed energy resources (DER)

**State Machine**:
- `Active` - Normal operation
- `High_Consumption` - Above typical usage pattern
- `Low_Voltage` - Voltage below acceptable range
- `Power_Quality_Issue` - Harmonics or distortion detected
- `Outage` - No power detected
- `Tampered` - Tampering detected
- `Communication_Loss` - Unable to communicate

**Example Behavior**:
```
IF voltage < minimum_threshold
  THEN notify_transformer_agent("Low voltage at customer")
  AND log_power_quality_event()
  AND notify_customer_if_prolonged()
  
IF consumption_pattern_unusual
  THEN analyze_for_theft_or_malfunction()
  AND alert_utility_if_tampering_suspected()
```

### 4. Circuit Breaker Agent
**Represents**: Protective devices that interrupt fault currents

**Attributes**:
- `breaker_id`: Unique identifier (e.g., "CB-001")
- `breaker_type`: vacuum, SF6, oil, air_blast, miniature (MCB), molded_case (MCCB)
- `location`: GPS coordinates
- `rated_voltage`: Voltage rating (kV)
- `rated_current`: Current rating (amperes)
- `interrupting_capacity`: Short circuit breaking capacity (kA)
- `position`: open, closed
- `trip_count`: Number of trip operations
- `last_trip_date`: Timestamp of last trip
- `trip_reason`: overcurrent, ground_fault, manual, remote
- `protected_line`: Reference to line/feeder agent
- `coordination_group`: Group of breakers for selective coordination
- `operating_mechanism`: spring, solenoid, motor
- `maintenance_interval`: Days between maintenance
- `condition_score`: Health score based on operations

**Capabilities**:
- Detect and clear fault currents
- Coordinate with other breakers for selective tripping
- Remote control (open/close operations)
- Monitor contact wear and operating mechanism health
- Automatic reclosing for transient faults
- Arc flash protection
- Load shedding during emergencies
- Predictive maintenance based on trip history

**State Machine**:
- `Closed_Normal` - Conducting current normally
- `Closed_High_Current` - Near trip threshold
- `Tripped` - Opened due to fault
- `Open_Manual` - Manually opened
- `Reclosing` - Automatic reclose sequence active
- `Locked_Out` - Multiple trips, manual intervention required
- `Maintenance` - Offline for service

**Example Behavior**:
```
IF fault_current_detected
  THEN trip_breaker(fault_type)
  AND notify_upstream_breaker("Fault cleared downstream")
  AND log_fault_event(location, magnitude, duration)
  AND initiate_reclose_sequence_if_transient()
  
IF trip_count > threshold_in_period
  THEN lock_out()
  AND notify_operators("Persistent fault condition")
```

### 5. Substation Agent
**Represents**: Electrical substations managing power distribution to zones

**Attributes**:
- `substation_id`: Unique identifier (e.g., "SUB-001")
- `location`: GPS coordinates
- `substation_type`: transmission, distribution, collector
- `primary_voltage`: Incoming voltage level (kV)
- `secondary_voltage`: Outgoing voltage level (kV)
- `rated_capacity`: Total MVA capacity
- `current_load`: Current total load (MW)
- `load_percentage`: Percentage of capacity
- `num_transformers`: Number of main transformers
- `num_feeders`: Number of outgoing feeders
- `connected_lines`: List of incoming line agents
- `feeder_agents`: List of outgoing feeder/line agents
- `transformer_agents`: References to transformer agents
- `breaker_agents`: References to breaker agents
- `voltage_control_mode`: manual, automatic, remote
- `reactive_power_capability`: MVAR range
- `SCADA_connected`: Boolean for SCADA integration

**Capabilities**:
- Manage load distribution across feeders
- Voltage regulation and reactive power control
- Coordinate feeder switching and load transfer
- Monitor substation equipment health
- Detect and isolate faults within substation
- Balance load across transformers
- Optimize power factor and minimize losses
- Coordinate with upstream transmission system
- Support islanding operations during emergencies
- Provide backup power routing

**State Machine**:
- `Normal_Operation` - All systems functioning
- `High_Load` - Operating near capacity
- `Voltage_Regulation_Active` - Active voltage control
- `Feeder_Fault` - Fault on one or more feeders
- `Transformer_Fault` - Main transformer issue
- `Emergency_Mode` - Critical condition, load shedding
- `Maintenance_Mode` - Partial shutdown for service

**Example Behavior**:
```
IF total_load > 90% of capacity
  THEN coordinate_load_transfer_to_adjacent_substations()
  AND notify_dispatch_center("Capacity constraint")
  AND prepare_load_shedding_plans()
  
IF feeder_fault_detected
  THEN isolate_faulted_feeder()
  AND reroute_load_to_healthy_feeders()
  AND notify_affected_customers()
```

### 6. Capacitor Bank Agent
**Represents**: Capacitor banks for reactive power compensation and voltage support

**Attributes**:
- `capacitor_id`: Unique identifier (e.g., "CAP-001")
- `location`: GPS coordinates
- `rated_capacity`: Reactive power capacity (MVAR)
- `voltage_level`: Operating voltage (kV)
- `num_steps`: Number of switchable steps
- `current_steps_active`: Steps currently energized
- `control_mode`: voltage, time, temperature, var, manual
- `switching_device`: breaker, contactor, vacuum_switch
- `connected_bus`: Reference to bus/node agent
- `voltage_sensor`: Reference to voltage monitoring
- `switch_count`: Total switching operations
- `last_switched`: Timestamp of last operation
- `installation_date`: Date installed
- `maintenance_interval`: Days between maintenance
- `condition_score`: Health score based on operations

**Capabilities**:
- Provide reactive power compensation
- Regulate voltage on distribution feeders
- Improve power factor
- Reduce line losses
- Respond to voltage fluctuations
- Coordinate with other voltage control devices
- Time-based switching for load patterns
- Temperature-based switching for seasonal adjustment
- Prevent excessive switching (minimize operations)
- Predictive maintenance based on switch count

**State Machine**:
- `All_Steps_Off` - Fully de-energized
- `Partial_Energization` - Some steps active
- `Fully_Energized` - All steps active
- `Switching` - Operation in progress
- `Locked_Out` - Disabled due to fault
- `Maintenance` - Offline for service

**Example Behavior**:
```
IF voltage < voltage_setpoint - deadband
  THEN energize_next_step()
  AND monitor_voltage_response()
  AND log_switching_operation()
  
IF switch_count_excessive_in_period
  THEN extend_deadband()
  AND notify_operators("Excessive switching detected")
```

### 7. Renewable Energy Agent (Solar/Wind)
**Represents**: Distributed energy resources (DER) integrated into the grid

**Attributes**:
- `der_id`: Unique identifier (e.g., "SOLAR-001", "WIND-001")
- `resource_type`: solar_pv, wind_turbine, battery_storage, microgrid
- `location`: GPS coordinates
- `rated_capacity`: Maximum generation capacity (kW)
- `current_generation`: Real-time power output (kW)
- `capacity_factor`: Current percentage of rated capacity
- `connected_meter`: Reference to interconnection meter agent
- `connected_transformer`: Reference to transformer agent
- `inverter_status`: online, offline, fault, curtailed
- `grid_connection_status`: connected, islanded, disconnected
- `forecast_generation`: Predicted output (next 24h)
- `weather_data`: Current irradiance/wind speed
- `energy_exported`: Cumulative energy to grid (kWh)
- `energy_imported`: Cumulative energy from grid (kWh)
- `power_factor`: Current power factor
- `reactive_power_mode`: fixed, volt-var, volt-watt
- `installation_date`: Date commissioned

**Capabilities**:
- Real-time generation monitoring
- Weather-based generation forecasting
- Grid synchronization and anti-islanding protection
- Reactive power support (volt-var optimization)
- Curtailment response during grid stress
- Net metering and bidirectional power flow
- Coordinate with energy storage for smoothing
- Voltage regulation through smart inverters
- Provide grid services (frequency response)
- Communicate outages and faults

**State Machine**:
- `Generating` - Producing power to grid
- `Standby` - Online but not generating
- `Curtailed` - Reduced output per grid request
- `Islanded` - Operating independently
- `Faulted` - Fault condition detected
- `Offline` - Disconnected from grid
- `Maintenance` - Under service

**Example Behavior**:
```
IF grid_voltage > max_threshold
  THEN reduce_active_power_output()
  AND absorb_reactive_power()
  AND notify_utility("Voltage support provided")
  
IF grid_fault_detected
  THEN disconnect_from_grid(anti_islanding)
  AND wait_for_grid_restoration()
  AND attempt_reconnection_after_delay()
```

## Agent Interaction Scenarios

### Scenario 1: Fault Detection and Isolation

**Trigger**: Line agent detects short circuit fault

**Agent Interaction Flow**:

1. **Power Line Agent (LINE-045)** detects fault condition
   ```
   State: Operational → Faulted
   Action: Analyze current surge and voltage collapse
   Decision: Short circuit detected at 2.3km from substation
   Fault current: 8,500A (exceeds normal 450A by 1889%)
   ```

2. **Line Agent** broadcasts fault alert
   ```
   → Upstream Circuit Breaker (CB-045): "URGENT: Fault current detected"
   → Downstream Transformers (XFMR-101, XFMR-102): "Prepare for power loss"
   → Adjacent Lines (LINE-044, LINE-046): "Standby for load transfer"
   → Substation Agent (SUB-003): "Feeder 045 fault condition"
   ```

3. **Circuit Breaker Agent (CB-045)** trips to clear fault
   ```
   State: Closed_Normal → Tripped
   Action: Open contacts in 3 cycles (50ms @ 60Hz)
   Trip reason: Overcurrent protection (87% differential)
   Fault cleared: Confirmed
   ```

4. **Substation Agent (SUB-003)** coordinates response
   ```
   Priority: CRITICAL
   Analysis: Fault location isolated, 1,250 customers affected
   Action: Initiate load transfer sequence
   Notification: Dispatch crew to fault location (GPS: 1.2345°N, 36.7890°E)
   ```

5. **Adjacent Line Agents** accept transferred load
   ```
   LINE-044: Increase load from 65% to 82% capacity
   LINE-046: Increase load from 58% to 75% capacity
   Breakers adjusted to new configuration
   Voltage regulation: Active
   ```

6. **Circuit Breaker Agent** initiates reclose sequence
   ```
   Wait time: 30 seconds (check for transient fault clearance)
   Reclose attempt 1: SUCCESS
   Fault status: Cleared (transient tree contact)
   State: Tripped → Closed_Normal
   Load restoration: Gradual ramp over 60 seconds
   ```

7. **Smart Meter Agents** confirm service restoration
   ```
   1,250 meters: Power restored
   Outage duration: 45 seconds average
   Customer notifications: Auto-sent via app
   Outage report: Generated for utility database
   ```

### Scenario 2: Peak Demand Management and Load Balancing

**Trigger**: High load during evening peak hours

**Agent Interaction Flow**:

1. **Substation Agent (SUB-007)** detects approaching capacity limit
   ```
   Current load: 28.5 MW (95% of 30 MW capacity)
   Trend: Increasing at 0.5 MW/hour
   Forecast: Will exceed capacity in 60 minutes
   Temperature: 35°C (high cooling demand)
   ```

2. **Substation Agent** initiates demand management
   ```
   → Adjacent Substations (SUB-006, SUB-008): "Load transfer request"
   → Transformer Agents: "Optimize tap positions for efficiency"
   → Capacitor Banks: "Maximize reactive power support"
   → Demand Response Coordinator: "Activate DR programs"
   ```

3. **Transformer Agents** optimize voltage
   ```
   XFMR-701: Adjust tap from position 5 to position 3
   Secondary voltage: Reduced from 242V to 235V (conservation voltage reduction)
   Load reduction: ~2% (0.57 MW saved)
   Customer impact: Minimal (within ±5% standard)
   ```

4. **Capacitor Bank Agents** improve power factor
   ```
   CAP-071: Energize 2 additional steps (4 MVAR total)
   CAP-072: Fully energize (6 MVAR)
   Power factor: Improved from 0.89 to 0.96
   Line losses: Reduced by 1.2% (0.34 MW freed)
   ```

5. **Smart Meter Agents** (DR participants) respond
   ```
   500 commercial meters: Curtail HVAC by 15%
   200 industrial meters: Shift non-critical loads by 2 hours
   Total demand reduction: 1.8 MW
   Incentive credits: Calculated and logged
   ```

6. **Adjacent Substation Agent (SUB-006)** accepts load transfer
   ```
   Current load: 22 MW (73% of capacity)
   Available capacity: 8 MW
   Accept transfer: 2.5 MW via tie-line reconfiguration
   Breakers adjusted: TIE-67 closed, reconfigure mesh
   ```

7. **Renewable Energy Agents** maximize contribution
   ```
   SOLAR-045 to SOLAR-089: 45 rooftop systems
   Current generation: 680 kW (cloud coverage limiting)
   Battery storage: Discharge 320 kW from aggregated systems
   Net grid support: 1.0 MW local generation
   ```

8. **Control Center** confirms stabilization
   ```
   Substation load: 26.8 MW (89% of capacity - safe level)
   Actions taken: Conservation voltage, capacitors, DR, transfer
   Peak shaved: 3.7 MW (13% reduction)
   Grid stability: Maintained
   Cost avoidance: $15,000 (avoided emergency purchases)
   ```

### Scenario 3: Predictive Maintenance and Equipment Health

**Trigger**: Transformer agent detects gradual performance degradation

**Agent Interaction Flow**:

1. **Transformer Agent (XFMR-234)** analyzes health trends
   ```
   Oil temperature: Trending up 2°C/month over 6 months
   Current temperature: 78°C (normal operating max: 75°C)
   Load: 85% (normal for this unit)
   Dissolved gas analysis (virtual sensor): Elevated acetylene
   Cooling fans: Running constantly (previously cycled)
   Efficiency: Decreased from 98.2% to 97.5%
   Operating hours: 87,500 (maintenance interval: 90,000)
   ```

2. **Transformer Agent** calculates failure probability
   ```
   ML Model prediction: 45% chance of failure in next 60 days
   Failure mode: Likely winding insulation breakdown
   Impact if failed: Critical (serves hospital and 3,500 customers)
   Recommendation: Schedule preventive maintenance within 30 days
   ```

3. **Transformer Agent** notifies maintenance system
   ```
   Priority: HIGH (preventive, critical load)
   Recommended action: Oil analysis, thermal imaging, tap changer inspection
   Preferred window: Planned outage (Sunday 2 AM - 8 AM)
   Redundancy available: Partial (70% load transferable to XFMR-235)
   Customer notification: Required for hospital and 30% of customers
   ```

4. **Adjacent Transformer Agent (XFMR-235)** prepares backup
   ```
   Current load: 65% (3.9 MVA of 6 MVA capacity)
   Available capacity: 2.1 MVA
   Required transfer from XFMR-234: 3.5 MVA (current load at 85%)
   Assessment: Can handle 60% of load (2.1 MVA)
   Gap: 1.4 MVA requires load curtailment or second backup
   ```

5. **Substation Agent (SUB-012)** coordinates backup plan
   ```
   Plan A: Transfer 2.1 MVA to XFMR-235
   Plan B: Transfer 1.0 MVA to adjacent substation SUB-013 via tie
   Plan C: Temporary load shedding for 30% non-critical customers
   Hospital circuit: Protected (dedicated backup via XFMR-236)
   Total coverage: 100% with multi-source backup
   ```

6. **Circuit Breaker Agents** prepare switching sequence
   ```
   CB-234-PRIMARY: Will open to de-energize XFMR-234
   CB-235-TIE: Will close to connect XFMR-235 to additional feeders
   CB-TIE-SUB013: Will close for inter-substation transfer
   Switching sequence: Validated via simulation
   Estimated switching time: 12 minutes total
   ```

7. **Smart Meter Agents** prepare customer notifications
   ```
   Hospital meters: Priority notification 72 hours advance
   3,500 customer meters: Notification 48 hours advance
   Message: Planned maintenance, estimated duration 6 hours
   Non-critical customers (30%): Possible brief interruption
   Actual outage window: Sunday 2:00 AM - 8:00 AM
   ```

8. **Maintenance Coordination Agent** schedules work
   ```
   Work order: PM-2025-1024-234
   Scheduled: Sunday, October 27, 2025, 2:00 AM
   Duration: 6 hours (thermal imaging, oil sampling, tap changer service)
   Crew: Team Bravo (transformer specialists)
   Parts pre-positioned: Cooling oil (500L), gaskets, filters
   Testing equipment: Oil analysis kit, thermal camera, turns-ratio tester
   Success criteria: Temperature reduction to <70°C, pass all tests
   ```

## Technical Architecture

### Agent Communication Patterns

1. **Direct Agent-to-Agent** (via Internal Message Bus):
   - Adjacent infrastructure elements (line-to-transformer, transformer-to-meter)
   - Ultra-low latency, high frequency updates (millisecond response)
   - Use cases: Fault propagation, voltage regulation coordination
   - Protocol: Internal message bus with priority queuing

2. **Publish-Subscribe** (via Message Queue):
   - Alerts and events broadcast to interested agents
   - Topics: fault_alerts, load_changes, voltage_events, maintenance_scheduled
   - Use cases: System-wide alerts, demand response, weather impacts
   - Protocol: MQTT, Redis Pub/Sub, or RabbitMQ

3. **Hierarchical Reporting** (via REST API/WebSocket):
   - Field agents → Zone controllers → Regional SCADA → Central EMS
   - Aggregated data, summary reports, historical trends
   - Use cases: Operations monitoring, analytics, regulatory reporting
   - Protocol: REST API, WebSocket for real-time streaming

4. **Mesh Coordination** (via Distributed Consensus):
   - Substations coordinating load balancing
   - Capacitor banks optimizing reactive power collectively
   - Distributed energy resources (DER) for grid support
   - Protocol: Consensus algorithms, peer-to-peer communication

5. **SCADA/IEC 61850 Integration**:
   - Legacy grid equipment integration
   - Standards-based substation automation
   - Use cases: Breaker control, transformer monitoring, protection coordination
   - Protocol: DNP3, Modbus, IEC 61850, IEC 60870-5-104

### Data Flow

```
Smart Meters (AMI)
  ↓ (consumption, voltage, power quality data)
Distribution Infrastructure Agents (lines, transformers, breakers)
  ↓ (load, faults, status changes)
Substation Agents
  ↓ (aggregated load, voltage profiles, fault reports)
Zone Coordinator / SCADA Agents
  ↓ (regional status, optimization recommendations)
Energy Management System (EMS) / Control Center
  ↓ (dispatch commands, configuration updates)
SCADA Dashboard / Operator Interface
  ↓ (manual control, approvals, overrides)
Grid Operators / Automated Systems
```

### Agent Deployment Model

**CodeValdCortex Framework Components Used**:

1. **Runtime Manager**: Manages lifecycle of all grid agent instances
2. **Agent Registry**: Tracks all infrastructure agents, their topology, and connectivity
3. **Task System**: Schedules meter readings, maintenance inspections, demand response events
4. **Memory Service**: Stores agent state, historical load profiles, learned patterns, fault history
5. **Communication System**: Enables agent-to-agent messaging with priority handling for faults
6. **Configuration Service**: Manages agent parameters, protection settings, voltage thresholds
7. **Health Monitor**: Tracks agent health, communication status, and performance metrics
8. **Event System**: Publishes and routes grid events (faults, load changes, DR activations)

**Deployment Architecture**:

```
Field Devices (Smart Meters, Sensors, RTUs)
  ↓
Edge Gateways (Substation Controllers, Field Agent Clusters)
  ├─ Smart Meter Agents (AMI)
  ├─ Infrastructure Agents (lines, transformers, breakers, capacitors)
  ├─ Renewable Energy Agents (solar, wind, storage)
  └─ Substation Agents (coordination, protection)
  ↓
Regional SCADA Servers (CodeValdCortex Runtime)
  ├─ Agent Runtime Manager
  ├─ Message Broker (MQTT/RabbitMQ)
  ├─ Time-Series Database (InfluxDB, TimescaleDB)
  ├─ ML Models (fault prediction, load forecasting)
  └─ SCADA Gateway (DNP3, Modbus, IEC 61850)
  ↓
Central Energy Management System (Cloud/On-Premise)
  ├─ Control Center Agent (dispatch, optimization)
  ├─ Dashboard (MVP-015 + Topology Visualizer)
  ├─ Analytics Engine (predictive maintenance, demand forecasting)
  ├─ GIS Integration (network topology, outage maps)
  └─ External Integrations (weather, markets, DR aggregators)
```

**Visualization**: This use case uses the Framework Topology Visualizer 
(schema version 1.0.0) with MapLibre-GL rendering for geographic display 
and Force-Directed layout for electrical network topology. Relationships 
follow the canonical taxonomy using `supply` (power flow), `route` (electrical 
connections), `command` (control relationships), and `depends_on` (backup/redundancy) 
edge types. See visualization configuration in 
`/usecases/UC-INFRA-002-stima/viz-config.json`.

## Integration Points

### 1. SCADA Systems
- Bidirectional integration with existing distribution SCADA infrastructure
- Agents read from RTUs, IEDs, and protective relays
- Agents send control commands to breakers, reclosers, and voltage regulators
- Real-time telemetry and control
- Protocol: DNP3, Modbus TCP/RTU, IEC 61850, IEC 60870-5-104

### 2. Advanced Metering Infrastructure (AMI)
- Smart meter data collection and commands
- Real-time consumption, voltage, and power quality data
- Remote connect/disconnect capabilities
- Demand response activation and confirmation
- Integration: Head-End System (HES) API, MDMS integration

### 3. GIS (Geographic Information Systems)
- Agent locations and network topology stored in GIS
- Spatial queries for network connectivity and fault location
- Outage mapping and crew dispatch optimization
- Visualization of real-time grid status on geographic maps
- Integration: ESRI ArcGIS, GeoJSON API, WMS/WFS services

### 4. Weather Services
- Real-time weather data for renewable energy forecasting
- Severe weather alerts for grid preparation
- Temperature forecasts for load forecasting
- Lightning detection for fault correlation
- Integration: NOAA API, Weather Underground, proprietary weather services

### 5. Energy Markets and ISO/RTO
- Real-time electricity pricing (LMP - Locational Marginal Pricing)
- Grid frequency and ancillary services signals
- Demand response program coordination
- Renewable energy curtailment signals
- Integration: ISO/RTO API (CAISO, PJM, ERCOT, etc.), OpenADR protocol

### 6. Customer Information System (CIS)
- Meter-to-customer account linking
- Billing data generation from consumption
- Outage notifications to customers
- Customer service request integration
- Integration: CIS API, SAP IS-U, Oracle Utilities

### 7. Outage Management System (OMS)
- Automated outage detection from meter and SCADA data
- Crew dispatch and work order management
- Estimated restoration times
- Customer communication
- Integration: OMS API, GE Smallworld, Oracle NMS

### 8. Distributed Energy Resource Management System (DERMS)
- Aggregation and control of distributed solar, wind, and storage
- Volt-var optimization across DERs
- Virtual power plant coordination
- Grid services from aggregated resources
- Integration: IEEE 2030.5, SunSpec Modbus, proprietary DERMS APIs

### 9. Asset Management Systems
- Equipment lifecycle tracking
- Maintenance scheduling and work history
- Parts inventory and procurement
- Condition-based maintenance triggers
- Integration: IBM Maximo, SAP EAM, Infor EAM

## Benefits Demonstrated

### 1. Fault Response and Reliability
- **Before**: Manual fault location, 2-4 hour average outage duration
- **With Agents**: Automated fault detection, isolation, and restoration
- **Metric**: 75% reduction in outage duration (30-60 minutes average)

### 2. Grid Visibility and Situational Awareness
- **Before**: Limited visibility, periodic manual readings, delayed fault awareness
- **With Agents**: Real-time monitoring of all grid assets and customer endpoints
- **Metric**: 100% infrastructure monitoring coverage, <1 second fault detection

### 3. Energy Efficiency and Loss Reduction
- **Before**: Distribution losses 6-8% (industry average)
- **With Agents**: Optimized voltage, power factor correction, load balancing
- **Metric**: Distribution losses reduced to <4%, saving $800K annually

### 4. Predictive Maintenance
- **Before**: Time-based maintenance, reactive equipment replacements
- **With Agents**: Condition-based predictive maintenance, health monitoring
- **Metric**: 50% reduction in equipment failures, 35% reduction in maintenance costs

### 5. Peak Demand Management
- **Before**: Manual load shedding, limited demand response, high peak demand charges
- **With Agents**: Automated demand response, intelligent load balancing, voltage optimization
- **Metric**: 12% peak demand reduction, $1.2M annual savings on capacity charges

### 6. Renewable Energy Integration
- **Before**: Limited solar/wind integration, manual curtailment, power quality issues
- **With Agents**: Intelligent DER management, automated volt-var control, grid services
- **Metric**: 300% increase in renewable hosting capacity without infrastructure upgrades

### 7. Customer Service and Satisfaction
- **Before**: Customer-reported outages, estimated restoration times, billing disputes
- **With Agents**: Proactive outage detection and notification, accurate usage data
- **Metric**: 90% customer satisfaction, 80% reduction in billing inquiries

### 8. Operational Efficiency
- **Before**: Manual switching, truck rolls for meter reading, reactive crew dispatch
- **With Agents**: Remote operations, automated meter reading, optimized dispatch
- **Metric**: 60% reduction in truck rolls, $500K annual labor savings

### 9. Power Quality
- **Before**: Reactive voltage regulation, poor power factor, customer complaints
- **With Agents**: Continuous voltage optimization, coordinated VAR support
- **Metric**: Voltage within ±3% (was ±8%), 95% reduction in power quality complaints

### 10. Grid Resilience
- **Before**: Single point failures, slow recovery from major events
- **With Agents**: Self-healing grid, automatic reconfiguration, islanding capability
- **Metric**: 90% of outages auto-restored, 45% faster recovery from major storms

## Implementation Phases

### Phase 1: Smart Metering and Data Foundation (Months 1-4)
- Deploy AMI smart meters across service territory
- Implement Smart Meter Agents with basic monitoring
- Establish data collection and time-series storage infrastructure
- Integrate with existing CIS and billing systems
- **Deliverable**: Real-time consumption monitoring dashboard

### Phase 2: Distribution Infrastructure Agents (Months 5-8)
- Implement Power Line, Transformer, and Circuit Breaker agents
- Deploy sensors for voltage, current, and power quality monitoring
- Connect agents to SCADA and field devices
- Establish agent-to-agent communication patterns
- **Deliverable**: Automated fault detection and alerting system

### Phase 3: Substation Automation and Control (Months 9-12)
- Implement Substation and Capacitor Bank agents
- Add automated voltage regulation and power factor optimization
- Integrate with legacy SCADA via DNP3/IEC 61850
- Deploy topology visualization (Framework Topology Visualizer)
- **Deliverable**: Automated voltage/var control and load balancing

### Phase 4: Advanced Grid Intelligence (Months 13-16)
- Deploy ML models for fault prediction and load forecasting
- Implement predictive maintenance capabilities
- Develop self-healing grid automation (auto-reconfiguration)
- Integrate Renewable Energy Agents (DER management)
- **Deliverable**: Predictive analytics and autonomous grid optimization

### Phase 5: Demand Response and Grid Services (Months 17-20)
- Implement demand response coordination through meter agents
- Enable time-of-use and dynamic pricing programs
- Coordinate DERs for grid services (frequency response, voltage support)
- Deploy customer engagement portal
- **Deliverable**: Active demand response program with 15% enrollment

## Success Criteria

### Technical Metrics
- ✅ 99.9% agent uptime and availability
- ✅ <100ms agent response time for critical faults
- ✅ 100% distribution infrastructure monitoring coverage
- ✅ <2% false positive rate on fault alerts
- ✅ <500ms end-to-end latency for SCADA commands
- ✅ Support for 100,000+ concurrent smart meter agents

### Operational Metrics
- ✅ 75% reduction in average outage duration
- ✅ 50% reduction in equipment failure rates
- ✅ 35% reduction in maintenance costs
- ✅ 12% peak demand reduction through optimization
- ✅ <4% distribution system losses (from 6-8%)
- ✅ 90% automated fault isolation and restoration

### Business Metrics
- ✅ ROI within 24 months
- ✅ $3.5M annual operational savings
- ✅ 90% customer satisfaction score
- ✅ 80% reduction in billing disputes
- ✅ $1.2M annual savings from demand management
- ✅ 300% increase in renewable energy hosting capacity

### Reliability Metrics
- ✅ SAIDI (System Average Interruption Duration Index) <90 minutes
- ✅ SAIFI (System Average Interruption Frequency Index) <1.2 events/year
- ✅ CAIDI (Customer Average Interruption Duration Index) <75 minutes
- ✅ 99.98% system reliability (excluding major events)

### Environmental and Sustainability Metrics
- ✅ 20% reduction in carbon emissions through efficiency gains
- ✅ 300 MW renewable energy integrated (was <100 MW)
- ✅ 15% participation in demand response programs
- ✅ $800K annual energy savings from loss reduction

## Conclusion

CodeValdStima demonstrates the power of the CodeValdCortex agent framework applied to electric power distribution infrastructure. By treating grid infrastructure elements as intelligent, autonomous agents, the system achieves:

- **Resilience**: Self-healing grid that automatically detects, isolates, and restores faults
- **Efficiency**: Optimized energy distribution with reduced losses and improved power quality
- **Intelligence**: Predictive maintenance, demand forecasting, and autonomous decision-making
- **Scalability**: Easily expandable to additional grid equipment and renewable energy resources
- **Integration**: Seamless interoperability with legacy SCADA, AMI, and enterprise systems
- **Sustainability**: Enhanced renewable energy integration and demand response capabilities

This use case serves as a reference implementation for applying agentic principles to other critical infrastructure domains such as water distribution, gas pipelines, telecommunications networks, transportation systems, and smart cities.

---

**Related Documents**:
- System Architecture: `documents/2-SoftwareDesignAndArchitecture/`
- Framework Topology Visualizer: `documents/2-SoftwareDesignAndArchitecture/framework-topology-visualizer/`
- Agent Implementation: `internal/agent/`
- Communication System: `internal/communication/`
- API Documentation: `documents/4-QA/`
- Dashboard: MVP-015 Management Dashboard
- Standard Use Case Definition: `documents/1-SoftwareRequirements/requirements/use-cases/standardusecasedefinition.md`

**Related Use Cases**:
- [UC-INFRA-001]: Water Distribution Network (CodeValdMaji)
- [UC-INFRA-003]: Natural Gas Distribution Network
- [UC-INFRA-004]: Telecommunications Network Management
- [UC-INFRA-005]: Smart City Infrastructure Coordination
