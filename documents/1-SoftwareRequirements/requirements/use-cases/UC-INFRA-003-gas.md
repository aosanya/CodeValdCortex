# Use Case: CodeValdGas - Natural Gas Distribution Network Management

**Use Case ID**: UC-INFRA-003  
**Use Case Name**: Natural Gas Distribution Network Agent System  
**System**: CodeValdGas  
**Created**: October 24, 2025  
**Status**: Example/Reference Implementation

## Overview

CodeValdGas is an example agentic system built on the CodeValdCortex framework that demonstrates how natural gas distribution infrastructure can be represented and managed as autonomous agents. This use case focuses on a natural gas distribution network where pipelines, pressure regulators, metering stations, odorization systems, and safety monitoring equipment are modeled as intelligent agents that monitor, communicate, and respond to network conditions.

**Note**: *"Gas" means "gas" in Swahili, reflecting the system's focus on natural gas distribution infrastructure.*

## System Context

### Domain
Municipal and industrial natural gas distribution infrastructure management and safety monitoring

### Business Problem
Traditional natural gas distribution systems suffer from:
- Reactive leak detection (finding problems after incidents occur)
- Limited real-time visibility into pressure and flow conditions
- Manual pressure regulation and load balancing
- Delayed response to safety hazards and gas leaks
- Poor coordination between infrastructure components
- Manual meter reading and billing inefficiencies
- Difficulty predicting equipment failures and pressure anomalies
- Safety risks from undetected leaks and pressure excursions
- Inefficient gas distribution causing waste and economic losses
- Slow emergency response coordination

### Proposed Solution
An agentic system where each gas infrastructure element is an autonomous agent that:
- Monitors its own state, pressure, flow, and safety metrics
- Communicates with neighboring network agents
- Detects leaks, pressure anomalies, and safety hazards in real-time
- Coordinates responses to network events and demand changes
- Provides real-time data for decision-making and load balancing
- Enables predictive maintenance and leak prevention
- Optimizes gas distribution and reduces losses
- Automates emergency response and safety protocols
- Integrates with odorization and gas quality monitoring

## Agent Types

### 1. Gas Pipeline Agent
**Represents**: Natural gas pipelines in the distribution network

**Attributes**:
- `pipeline_id`: Unique identifier (e.g., "PIPE-G-001")
- `pipeline_type`: transmission, distribution, service_line, industrial_feeder
- `material`: steel, polyethylene (PE), PVC, cast_iron
- `diameter`: Pipe diameter in millimeters
- `length`: Pipeline length in meters
- `pressure_rating`: Maximum operating pressure (PSI or bar)
- `installation_date`: Date installed
- `location`: GPS coordinates (start and end points)
- `coating_type`: Cathodic protection, epoxy, bare
- `design_pressure`: Design maximum allowable operating pressure (MAOP)
- `current_pressure`: Real-time pressure (PSI)
- `flow_rate`: Current gas flow (cubic meters/hour)
- `temperature`: Gas temperature (°C)
- `connected_nodes`: List of connected regulators, meters, and sensor agents
- `condition_score`: Health score (0-100)
- `leak_detection_sensors`: References to leak detection equipment
- `cathodic_protection`: CP status and readings

**Capabilities**:
- Monitor pressure, flow rate, and temperature continuously
- Detect pressure drops indicating leaks
- Calculate flow efficiency and detect anomalies
- Track corrosion and degradation over time
- Communicate with adjacent pipeline segments
- Report maintenance needs based on condition
- Predict remaining lifespan based on usage patterns
- Detect third-party damage and ground movement
- Coordinate with pressure regulators for balancing
- Integrate with SCADA for remote monitoring

**State Machine**:
- `Operational` - Normal gas flow
- `High_Pressure` - Pressure above normal operating range
- `Low_Pressure` - Pressure below minimum threshold
- `Leak_Suspected` - Anomaly suggesting leak
- `Leak_Confirmed` - Leak detected and confirmed
- `Isolated` - Section isolated for safety
- `Maintenance` - Under repair

**Example Behavior**:
```
IF pressure_drop > threshold AND flow_rate > 0
  THEN raise_alert("Possible leak detected")
  AND notify_adjacent_pipelines()
  AND notify_control_center()
  AND calculate_leak_location()
  
IF pressure > MAOP
  THEN raise_alert("Overpressure condition - CRITICAL")
  AND notify_pressure_regulators("Reduce pressure immediately")
  AND prepare_for_emergency_shutdown()
```

### 2. Pressure Regulator Agent
**Represents**: Pressure reducing and regulating stations

**Attributes**:
- `regulator_id`: Unique identifier (e.g., "REG-001")
- `regulator_type`: district, service, industrial, master
- `location`: GPS coordinates
- `inlet_pressure`: Upstream pressure (PSI)
- `outlet_pressure`: Downstream pressure (PSI)
- `set_point`: Target outlet pressure
- `capacity`: Maximum flow capacity (cubic meters/hour)
- `current_flow`: Current gas flow
- `valve_position`: Current valve opening (0-100%)
- `upstream_pipeline`: Reference to upstream pipeline agent
- `downstream_pipeline`: Reference to downstream pipeline agent
- `pressure_sensors`: References to pressure monitoring
- `safety_relief_valve`: Over-pressure protection status
- `operation_count`: Number of regulation cycles
- `last_maintained`: Date of last maintenance
- `condition_score`: Health score based on performance

**Capabilities**:
- Regulate downstream pressure automatically
- Monitor inlet and outlet pressure continuously
- Detect regulator malfunction or failure
- Coordinate with other regulators for load balancing
- Predict maintenance needs based on cycle count
- Activate safety relief in overpressure conditions
- Optimize pressure for demand patterns
- Report performance degradation
- Calculate flow capacity and efficiency
- Emergency shutdown capability

**State Machine**:
- `Regulating` - Active pressure control
- `Steady_State` - Minimal regulation needed
- `High_Flow` - Operating near capacity
- `Malfunction` - Performance degraded
- `Safety_Relief_Active` - Overpressure relief engaged
- `Locked_Out` - Manual isolation for maintenance
- `Emergency_Shutdown` - Critical safety response

**Example Behavior**:
```
IF outlet_pressure > set_point + tolerance
  THEN reduce_valve_position()
  AND monitor_pressure_response()
  
IF outlet_pressure < set_point - tolerance
  THEN increase_valve_position()
  AND check_upstream_pressure_availability()
  
IF inlet_pressure_loss_detected
  THEN notify_downstream_customers("Low pressure warning")
  AND notify_control_center("Supply pressure issue")
```

### 3. Gas Meter Agent
**Represents**: Gas consumption meters at customer endpoints

**Attributes**:
- `meter_id`: Unique identifier (e.g., "MTR-G-001")
- `customer_id`: Associated customer account
- `location`: GPS coordinates or address
- `meter_type`: residential, commercial, industrial
- `technology`: diaphragm, rotary, turbine, ultrasonic
- `connected_pipeline`: Reference to service line agent
- `current_flow`: Real-time gas consumption (cubic meters/hour)
- `cumulative_volume`: Total gas consumed (cubic meters)
- `pressure`: Line pressure at meter (PSI)
- `temperature`: Gas temperature at meter (°C)
- `installation_date`: Date installed
- `last_reading_date`: Last meter reading timestamp
- `billing_cycle`: Monthly, bi-monthly
- `average_consumption`: Historical average
- `leak_detection`: Continuous flow detection capability
- `communication_status`: AMR/AMI connectivity

**Capabilities**:
- Continuous consumption monitoring
- Automatic meter reading (AMR/AMI)
- Leak detection at customer premises
- Usage pattern analysis
- Billing data generation
- Tamper detection
- Demand forecasting
- Customer alerts for unusual consumption
- Pressure monitoring for service quality
- Remote valve control (smart meters)

**State Machine**:
- `Active` - Normal operation
- `No_Flow` - No consumption detected
- `High_Flow` - Above normal consumption
- `Possible_Leak` - Continuous low flow detected
- `Low_Pressure` - Supply pressure issue
- `Tampered` - Tampering detected
- `Offline` - Communication lost
- `Maintenance` - Under service

**Example Behavior**:
```
IF continuous_flow > 0 AND time_period > 24_hours
  THEN raise_alert("Possible leak at customer premises")
  AND notify_customer()
  AND log_anomaly()
  
IF pressure < minimum_service_pressure
  THEN notify_customer("Low pressure condition")
  AND notify_upstream_regulator("Service pressure issue")
```

### 4. Leak Detection Agent
**Represents**: Leak detection sensors and monitoring systems

**Attributes**:
- `sensor_id`: Unique identifier (e.g., "LEAK-001")
- `sensor_type`: combustible_gas, acoustic, optical, ground_movement
- `location`: GPS coordinates
- `attached_to`: Reference to pipeline or facility
- `detection_method`: point_sensor, distributed_fiber_optic, mobile_survey
- `sensitivity`: Detection threshold (PPM or percentage LEL)
- `current_reading`: Real-time gas concentration or acoustic signature
- `baseline_reading`: Normal background level
- `battery_level`: For wireless sensors (percentage)
- `last_calibration`: Date of last calibration
- `alert_thresholds`: Warning and critical levels
- `communication_status`: Connected, intermittent, offline

**Capabilities**:
- Continuous gas leak monitoring
- Real-time leak detection and alerting
- Acoustic signature analysis for underground leaks
- Fiber optic distributed sensing along pipelines
- Trend analysis for early leak indication
- Automatic calibration verification
- Battery monitoring and alerts
- Correlation with pressure drop data
- Geographic leak location pinpointing
- Integration with GIS for leak mapping

**State Machine**:
- `Monitoring` - Normal operation
- `Warning` - Elevated reading detected
- `Leak_Detected` - Confirmed gas leak
- `Calibrating` - Running calibration routine
- `Low_Battery` - Battery replacement needed
- `Error` - Malfunction detected
- `Offline` - Communication lost

**Example Behavior**:
```
IF gas_concentration > warning_threshold
  THEN increase_sampling_rate()
  AND notify_pipeline_agent()
  AND log_detection_event()
  
IF gas_concentration > critical_threshold
  THEN raise_emergency_alert()
  AND notify_emergency_response("Gas leak confirmed")
  AND activate_nearby_leak_sensors()
  AND calculate_leak_severity()
```

### 5. Odorization System Agent
**Represents**: Gas odorization equipment adding mercaptan for leak detection

**Attributes**:
- `odorant_id`: Unique identifier (e.g., "ODOR-001")
- `location`: GPS coordinates and facility
- `odorant_type`: mercaptan, THT (tetrahydrothiophene), blend
- `injection_rate`: Odorant injection rate (PPM)
- `tank_level`: Odorant tank level (liters or percentage)
- `target_concentration`: Required odor intensity (PPM)
- `current_concentration`: Measured downstream concentration
- `pump_status`: Running, standby, failed
- `flow_rate`: Gas flow through odorizer
- `tank_capacity`: Total odorant capacity
- `refill_threshold`: Minimum level for refill alert
- `last_refilled`: Date of last odorant refill
- `calibration_status`: Last calibration verification

**Capabilities**:
- Maintain consistent odorant concentration
- Monitor odorant tank levels
- Adjust injection rate based on gas flow
- Detect odorant pump failures
- Alert on low odorant supply
- Verify downstream odor intensity
- Log injection rates and consumption
- Predict refill schedule
- Emergency odorization boost capability
- Integration with safety systems

**State Machine**:
- `Operating` - Normal odorization
- `Low_Tank` - Odorant level below threshold
- `Pump_Failure` - Injection pump malfunction
- `Insufficient_Odor` - Downstream odor below target
- `Refilling` - Odorant tank being refilled
- `Calibrating` - System calibration in progress
- `Offline` - Not operational

**Example Behavior**:
```
IF tank_level < refill_threshold
  THEN notify_operations("Odorant refill required")
  AND schedule_refill_delivery()
  AND estimate_days_until_empty()
  
IF downstream_concentration < target_concentration
  THEN increase_injection_rate()
  AND verify_pump_operation()
  AND notify_maintenance_if_persistent()
```

### 6. Compressor Station Agent
**Represents**: Gas compression facilities maintaining transmission pressure

**Attributes**:
- `compressor_id`: Unique identifier (e.g., "COMP-001")
- `location`: GPS coordinates and facility name
- `compressor_type`: reciprocating, centrifugal, screw
- `rated_capacity`: Maximum compression capacity (cubic meters/hour)
- `current_throughput`: Current gas flow
- `suction_pressure`: Inlet pressure (PSI)
- `discharge_pressure`: Outlet pressure (PSI)
- `compression_ratio`: Current pressure ratio
- `power_consumption`: Energy usage (kW)
- `efficiency`: Operating efficiency (%)
- `operating_hours`: Total runtime hours
- `vibration_level`: Vibration monitoring
- `temperature`: Compressor temperature (°C)
- `cooling_system`: Cooling system status
- `maintenance_interval`: Hours between maintenance

**Capabilities**:
- Compress gas to maintain transmission pressure
- Monitor performance and efficiency
- Detect abnormal vibration and temperature
- Predict maintenance needs
- Optimize energy consumption
- Coordinate with other compressors
- Emergency shutdown on critical conditions
- Load balancing across multiple units
- Performance trending and analysis
- Integration with SCADA systems

**State Machine**:
- `Running` - Operating normally
- `Loaded` - High throughput operation
- `Unloaded` - Low demand, minimal compression
- `Starting` - Startup sequence
- `Stopping` - Shutdown sequence
- `Warning` - Abnormal condition detected
- `Emergency_Stop` - Critical failure
- `Maintenance` - Under repair

**Example Behavior**:
```
IF discharge_pressure < target_pressure
  THEN increase_compression_ratio()
  AND monitor_power_consumption()
  
IF vibration > critical_threshold
  THEN emergency_shutdown()
  AND notify_maintenance("Excessive vibration detected")
  AND activate_backup_compressor()
```

### 7. SCADA Control Agent
**Represents**: Supervisory control and data acquisition system coordination

**Attributes**:
- `scada_id`: Unique identifier (e.g., "SCADA-001")
- `coverage_area`: Geographic region monitored
- `connected_assets`: List of all monitored infrastructure
- `active_alarms`: Current system alerts
- `control_mode`: automatic, manual, semi_automatic
- `operator_stations`: Connected operator workstations
- `data_historian`: Time-series data storage
- `communication_protocol`: DNP3, Modbus, OPC-UA
- `redundancy_status`: Primary/backup system status

**Capabilities**:
- Monitor entire gas distribution network
- Aggregate data from all field agents
- Coordinate emergency responses
- Optimize pressure and flow across network
- Manage operator interfaces
- Log all events and alarms
- Generate performance reports
- Coordinate with multiple utility systems
- Provide situational awareness
- Enable remote control operations

**State Machine**:
- `Normal_Operation` - All systems functioning
- `Alert_Active` - One or more alarms raised
- `Emergency_Mode` - Critical situation management
- `Manual_Override` - Operator manual control
- `System_Degraded` - Partial system failure
- `Maintenance_Mode` - Scheduled maintenance

**Example Behavior**:
```
IF multiple_leak_alerts_detected
  THEN coordinate_emergency_response()
  AND notify_all_operators()
  AND prepare_isolation_procedures()
  AND alert_emergency_services()
  
IF network_pressure_trending_low
  THEN analyze_demand_patterns()
  AND optimize_compressor_operations()
  AND balance_regulator_settings()
```

## Agent Interaction Scenarios

### Scenario 1: Gas Leak Detection and Emergency Response

**Trigger**: Leak detection sensor detects natural gas concentration

**Agent Interaction Flow**:

1. **Leak Detection Agent (LEAK-045)** detects gas presence
   ```
   State: Monitoring → Leak_Detected
   Gas concentration: 15% LEL (Lower Explosive Limit)
   Location: GPS coordinates with 5-meter accuracy
   Trend: Increasing concentration
   ```

2. **Leak Detection Agent** raises emergency alert
   ```
   → Pipeline Agent (PIPE-G-123): "CRITICAL: Gas leak detected"
   → SCADA Control Agent: "Emergency - leak at GPS coordinates"
   → Adjacent Leak Sensors: "Increase monitoring frequency"
   → Emergency Response Team: "Gas leak response required"
   ```

3. **Pipeline Agent (PIPE-G-123)** confirms leak
   ```
   State: Operational → Leak_Confirmed
   Pressure analysis: 8 PSI drop over 15 minutes
   Flow analysis: Unaccounted gas loss detected
   Estimated leak size: 50 cubic meters/hour
   Action: Prepare for isolation
   ```

4. **SCADA Control Agent** coordinates response
   ```
   Priority: EMERGENCY
   Action: Initiate leak isolation protocol
   Affected area: 500-meter radius
   Customers impacted: 350 residential, 12 commercial
   Emergency services notified: Fire department, police
   ```

5. **Pressure Regulator Agents** isolate section
   ```
   REG-044 (upstream): Close isolation valve
   REG-045 (downstream): Close isolation valve
   Isolation time: 3 minutes
   Venting: Controlled depressurization initiated
   ```

6. **Gas Meter Agents** detect service interruption
   ```
   350 meters: Pressure = 0 PSI (service interrupted)
   Automatic notifications: Sent to affected customers
   Message: "Temporary gas service interruption - safety maintenance"
   Estimated restoration: 6-8 hours
   ```

7. **Odorization System Agent** verifies odor adequacy
   ```
   Analysis: Odorant concentration adequate for leak detection
   Public reports: Multiple calls reporting gas odor
   Coverage: Leak was detectable by smell (safety system effective)
   ```

8. **Emergency Response Team** dispatched
   ```
   Crew arrival: 12 minutes from alert
   Evacuation: 500-meter radius evacuated
   Leak confirmation: Visual and instrument verification
   Repair initiated: Excavation and pipe replacement
   Repair duration: 6 hours
   ```

9. **SCADA Control Agent** manages restoration
   ```
   Repair verified: Pressure test passed
   Re-pressurization: Gradual pressure increase over 30 minutes
   Meter relighting: Crews dispatched to relight pilot lights
   Service restored: All 350 customers online
   Post-incident report: Generated with timeline and lessons learned
   ```

### Scenario 2: Predictive Maintenance and Pressure Optimization

**Trigger**: Pressure regulator agent detects performance degradation

**Agent Interaction Flow**:

1. **Pressure Regulator Agent (REG-078)** analyzes trends
   ```
   Performance trending: Valve response time increased 25% over 60 days
   Pressure oscillation: Hunting behavior (±3 PSI swings)
   Cycle count: 45,000 cycles (maintenance interval: 50,000)
   Efficiency: Decreased from 98% to 92%
   Spring compression: 10% loss detected
   ```

2. **Pressure Regulator Agent** predicts failure
   ```
   ML Model prediction: 65% probability of failure in next 30 days
   Failure mode: Diaphragm fatigue or spring degradation
   Impact: 450 customers, critical commercial accounts
   Recommendation: Schedule preventive maintenance within 2 weeks
   ```

3. **SCADA Control Agent** receives alert
   ```
   Priority: MEDIUM (preventive, critical location)
   Action: Schedule maintenance during low-demand period
   Backup plan: Activate parallel regulator (REG-078-B)
   Customer notification: Not required (no service interruption)
   ```

4. **Adjacent Regulator Agent (REG-078-B)** prepares
   ```
   State: Standby → Pre-Activation
   Self-diagnostic: All systems operational
   Capacity check: Can handle 100% of REG-078 load
   Warm-up: Begin gradual pressure equalization
   ```

5. **Pipeline Agents** prepare for load shift
   ```
   Upstream pipeline: Increase pressure 5 PSI to support both regulators
   Downstream pipeline: Monitor pressure during transition
   Flow balancing: Optimize distribution across network
   ```

6. **Maintenance coordination** scheduled
   ```
   Work order: PM-2025-1024-078
   Scheduled: Low-demand period (Sunday 2:00 AM - 6:00 AM)
   Duration: 4 hours (diaphragm replacement, spring replacement, calibration)
   Parts: Pre-positioned (diaphragm, springs, seals)
   Crew: Regulator specialists Team Charlie
   ```

7. **Cutover execution** seamless
   ```
   REG-078: Gradually close over 15 minutes
   REG-078-B: Simultaneously open to maintain pressure
   Pressure variation: <1 PSI (customers unaware)
   REG-078: Isolated for maintenance
   ```

8. **Maintenance completion** and restoration
   ```
   Work completed: 3.5 hours (under estimated time)
   Testing: Passed all performance tests
   Performance: Restored to 99% efficiency
   REG-078: Returned to service as primary
   REG-078-B: Returned to standby
   Success: Zero customer impact, prevented potential failure
   ```

### Scenario 3: Demand Surge Management and Load Balancing

**Trigger**: Cold weather drives sudden increase in gas demand

**Agent Interaction Flow**:

1. **Gas Meter Agents** detect demand surge
   ```
   Time: 6:00 AM (morning heating startup)
   Residential demand: Increased 350% from overnight baseline
   Commercial demand: Increased 280% (building heating)
   Total network demand: 85% of system capacity (was 25%)
   Trend: Still increasing
   ```

2. **Pipeline Agents** detect pressure drop
   ```
   Distribution mains: Pressure dropping 2 PSI per 10 minutes
   Normal pressure: 60 PSI
   Current pressure: 52 PSI (borderline low)
   Forecast: Will reach 45 PSI (minimum) in 30 minutes
   ```

3. **Pressure Regulator Agents** fully open
   ```
   Multiple regulators: Valve positions at 95-100% open
   Capacity: Operating at maximum throughput
   Inlet pressure: Drawing down from transmission system
   Status: Cannot provide additional capacity
   ```

4. **SCADA Control Agent** initiates demand management
   ```
   Analysis: Demand exceeds supply capacity
   Weather: -10°C (extreme cold, continued high demand expected)
   Duration: Likely 8-12 hours until temperatures moderate
   Action: Activate demand management protocol
   ```

5. **Compressor Station Agent (COMP-003)** increases output
   ```
   State: Unloaded → Loaded
   Compression ratio: Increased from 1.2:1 to 1.5:1
   Throughput: Increased from 60% to 95% capacity
   Discharge pressure: Increased from 90 PSI to 110 PSI
   Additional supply: +15% system capacity
   ```

6. **Backup Compressor (COMP-004)** activated
   ```
   State: Standby → Starting → Running
   Startup time: 8 minutes
   Additional capacity: +25% system capacity
   Combined compression: Both compressors at 80% capacity
   Network pressure: Stabilizing at 55 PSI
   ```

7. **Industrial Customer Coordination**
   ```
   Interruptible contracts: 5 industrial customers notified
   Curtailment request: Reduce consumption by 40%
   Response time: Within 30 minutes
   Compensation: Per contract rates ($50,000 incentive)
   Capacity freed: 15% of total system demand
   ```

8. **SCADA Control Agent** confirms stabilization
   ```
   Network pressure: Restored to 58 PSI (safe operating level)
   Demand met: 100% of firm customers served
   Industrial curtailment: Temporary (6 hours)
   Cost avoidance: Prevented emergency supply purchases ($200,000)
   Weather forecast: Temperatures moderating by evening
   ```

## Technical Architecture

### Agent Communication Patterns

1. **Direct Agent-to-Agent** (via Internal Message Bus):
   - Adjacent infrastructure elements (pipeline-to-regulator, regulator-to-meter)
   - Ultra-low latency for safety-critical alerts (<100ms)
   - Use cases: Leak detection, pressure anomalies, emergency isolation
   - Protocol: Internal message bus with safety priority queuing

2. **Publish-Subscribe** (via Message Queue):
   - Safety alerts broadcast to all stakeholders
   - Topics: leak_alerts, pressure_events, maintenance_scheduled, odor_complaints
   - Use cases: Emergency notifications, public safety alerts
   - Protocol: MQTT, Redis Pub/Sub with guaranteed delivery

3. **Hierarchical Reporting** (via SCADA/REST API):
   - Field agents → Regional SCADA → Central control → Corporate systems
   - Aggregated data, performance metrics, compliance reporting
   - Use cases: Operations monitoring, regulatory compliance
   - Protocol: DNP3, Modbus, OPC-UA, REST API

4. **Event-Driven** (via Event Stream):
   - Leak detections, pressure excursions, equipment failures
   - Safety system activations
   - Use cases: Real-time safety monitoring, incident response
   - Protocol: Event stream with audit trail

5. **Safety System Integration**:
   - Emergency shutdown systems (ESD)
   - Fire and gas detection systems
   - Public alerting systems
   - Protocol: Hardwired interlocks + digital supervisory

### Data Flow

```
Leak Detection Sensors & Gas Meters (IoT/AMI)
  ↓ (real-time measurements, leak alerts)
Pipeline, Regulator, Compressor Agents
  ↓ (pressure, flow, status changes)
SCADA Control Agent
  ↓ (aggregated network status, alarms)
Control Center / Operations
  ↓ (dispatch commands, emergency procedures)
Emergency Response / Public Safety
  ↓ (incident management, public notifications)
Regulatory Compliance Systems
```

### Agent Deployment Model

**CodeValdCortex Framework Components Used**:

1. **Runtime Manager**: Manages lifecycle of all gas infrastructure agents
2. **Agent Registry**: Tracks pipelines, regulators, meters, sensors, and connectivity
3. **Task System**: Schedules meter readings, leak surveys, maintenance inspections
4. **Memory Service**: Stores pressure history, leak incidents, maintenance records
5. **Communication System**: Enables multi-channel safety-critical messaging
6. **Configuration Service**: Manages pressure setpoints, alarm thresholds, safety limits
7. **Health Monitor**: Tracks agent communication health and system reliability
8. **Event System**: Publishes and routes safety events with audit trails

**Deployment Architecture**:

```
Field Devices (Leak Sensors, Pressure Transmitters, Flow Meters, AMI)
  ↓ (4G/LoRaWAN/Hardwired)
Edge Gateways / RTUs (Regional Substations)
  ├─ Pipeline Agents
  ├─ Pressure Regulator Agents
  ├─ Leak Detection Agents
  ├─ Gas Meter Agents
  └─ Safety interlocks (hardwired)
  ↓ (Fiber/Microwave/Cellular)
Regional SCADA Servers (CodeValdCortex Runtime)
  ├─ Agent Runtime Manager
  ├─ SCADA Control Agents
  ├─ Compressor Station Agents
  ├─ Odorization System Agents
  ├─ Message Broker (MQTT/RabbitMQ)
  ├─ Time-Series Database (InfluxDB)
  ├─ ML Models (leak prediction, demand forecasting)
  └─ SCADA Gateway (DNP3, Modbus)
  ↓
Central Control (Cloud/On-Premise - High Availability)
  ├─ Control Center Dashboard
  ├─ Emergency Response Coordination
  ├─ Analytics Engine (predictive maintenance)
  ├─ GIS Integration (leak mapping, pipeline routes)
  ├─ Regulatory Compliance Reporting
  └─ Public Alerting Integration
  ↓
External Integrations
  ├─ Emergency Services (911, Fire, Police)
  ├─ Weather Services
  ├─ Public Notification Systems
  ├─ Regulatory Agencies
  └─ Customer Information Systems
```

**Visualization**: This use case uses the Framework Topology Visualizer 
(schema version 1.0.0) with MapLibre-GL rendering for geographic pipeline 
routes and Force-Directed layout for network topology. Relationships follow 
the canonical taxonomy using `supply` (gas flow), `route` (pipeline connections), 
`command` (control relationships), `observe` (sensor monitoring), and `depends_on` 
(backup/redundancy) edge types. See visualization configuration in 
`/usecases/UC-INFRA-003-gesi/viz-config.json`.

## Integration Points

### 1. SCADA Systems
- Bidirectional integration with existing gas SCADA infrastructure
- Real-time telemetry from RTUs and field devices
- Control commands to valves, regulators, and compressors
- Protocol: DNP3, Modbus TCP/RTU, OPC-UA

### 2. Advanced Metering Infrastructure (AMI)
- Automated meter reading for billing
- Real-time consumption and pressure data
- Leak detection at customer premises
- Integration: AMI head-end systems, MDMS

### 3. GIS (Geographic Information Systems)
- Pipeline routes and infrastructure locations
- Spatial analysis for leak location
- Proximity analysis for public safety
- Integration: ESRI ArcGIS, pipeline GIS databases

### 4. Emergency Services Integration
- 911 dispatch systems
- Fire department pre-planning
- Police for evacuations and road closures
- Integration: CAD (Computer-Aided Dispatch) systems

### 5. Public Notification Systems
- Reverse 911 for emergency alerts
- SMS alerts for affected customers
- Social media integration for public awareness
- Integration: Emergency mass notification platforms

### 6. Weather Services
- Temperature forecasts for demand planning
- Severe weather alerts for emergency preparedness
- Wind data for leak dispersion modeling
- Integration: NOAA, proprietary weather services

### 7. Regulatory Compliance Systems
- Pipeline safety reporting (PHMSA)
- Environmental compliance (EPA)
- State utility commission reporting
- Integration: Regulatory reporting portals, XML/API submissions

### 8. Customer Information System (CIS)
- Customer account linkage
- Billing integration
- Service request management
- Integration: CIS API, SAP IS-U

### 9. Leak Survey Systems
- Mobile leak detection equipment
- Periodic pipeline surveys
- Corrosion monitoring
- Integration: Leak survey data upload, GIS integration

## Benefits Demonstrated

### 1. Safety and Emergency Response
- **Before**: Leak detection delayed, manual emergency response
- **With Agents**: Real-time leak detection, automated emergency protocols
- **Metric**: 90% faster leak detection and isolation (minutes vs hours)

### 2. System Visibility
- **Before**: Limited pressure/flow visibility, manual monitoring
- **With Agents**: Complete real-time network monitoring
- **Metric**: 100% infrastructure monitoring coverage

### 3. Gas Loss Reduction
- **Before**: Unaccounted gas loss 2-3% (leaks and measurement errors)
- **With Agents**: Early leak detection, accurate metering
- **Metric**: Losses reduced to <0.5%, saving $1.5M annually

### 4. Predictive Maintenance
- **Before**: Time-based maintenance, reactive equipment replacement
- **With Agents**: Condition-based predictive maintenance
- **Metric**: 50% reduction in emergency repairs, 40% maintenance cost savings

### 5. Demand Management
- **Before**: Manual load balancing, reactive capacity management
- **With Agents**: Automated demand response, optimized compression
- **Metric**: 20% better peak demand handling without infrastructure expansion

### 6. Public Safety
- **Before**: Customer-reported leaks, delayed evacuations
- **With Agents**: Proactive leak detection, coordinated emergency response
- **Metric**: Zero public safety incidents from gas leaks

### 7. Regulatory Compliance
- **Before**: Manual reporting, compliance challenges
- **With Agents**: Automated compliance tracking and reporting
- **Metric**: 100% on-time regulatory reporting, zero violations

### 8. Operational Efficiency
- **Before**: Manual meter reading, truck rolls for inspections
- **With Agents**: Automated readings, targeted inspections
- **Metric**: 70% reduction in field operations costs

### 9. Pressure Optimization
- **Before**: Conservative pressure settings, energy waste
- **With Agents**: Dynamic pressure optimization
- **Metric**: 15% reduction in compression energy costs

### 10. Customer Service
- **Before**: Estimated billing, service complaints
- **With Agents**: Accurate consumption data, proactive notifications
- **Metric**: 85% reduction in billing disputes

## Implementation Phases

### Phase 1: Safety and Monitoring Foundation (Months 1-4)
- Deploy leak detection sensors on critical pipelines
- Implement Pipeline and Leak Detection agents
- Establish SCADA integration
- Deploy emergency notification systems
- **Deliverable**: Real-time leak detection and safety monitoring

### Phase 2: Metering and Pressure Control (Months 5-8)
- Deploy AMI smart meters
- Implement Gas Meter and Pressure Regulator agents
- Add automated pressure optimization
- Integrate with billing systems
- **Deliverable**: Automated meter reading and pressure management

### Phase 3: Advanced Operations (Months 9-12)
- Implement Compressor Station and Odorization agents
- Add SCADA Control agent coordination
- Deploy predictive maintenance ML models
- Integrate GIS for spatial analysis
- **Deliverable**: Autonomous network optimization

### Phase 4: Intelligence and Scale (Months 13-16)
- Scale to full distribution network coverage
- Deploy demand forecasting models
- Implement advanced leak detection (fiber optic)
- Add public safety integration
- **Deliverable**: Intelligent, self-optimizing gas network

## Success Criteria

### Technical Metrics
- ✅ 99.99% safety system uptime
- ✅ <1 minute leak detection and isolation time
- ✅ 100% pipeline monitoring coverage
- ✅ <0.1% false positive rate on leak alerts

### Safety Metrics
- ✅ Zero public safety incidents from gas leaks
- ✅ 100% compliance with pipeline safety regulations
- ✅ <5 minute emergency response coordination time
- ✅ 90% faster leak detection vs manual surveys

### Operational Metrics
- ✅ 50% reduction in unaccounted gas losses
- ✅ 40% reduction in maintenance costs
- ✅ 70% reduction in field operation costs
- ✅ 15% energy savings in compression

### Business Metrics
- ✅ ROI within 30 months
- ✅ $2.5M annual operational savings
- ✅ 90% customer satisfaction
- ✅ Zero regulatory violations

## Conclusion

CodeValdGas demonstrates the transformative power of the CodeValdCortex agent framework applied to natural gas distribution infrastructure. By treating gas infrastructure elements as intelligent, autonomous agents, the system achieves:

- **Safety**: Dramatically improved public safety through real-time leak detection and emergency response
- **Efficiency**: Optimized gas distribution with reduced losses and energy consumption
- **Intelligence**: Predictive maintenance, demand forecasting, and autonomous optimization
- **Compliance**: Automated regulatory compliance and reporting
- **Resilience**: Self-monitoring network with coordinated emergency response

This use case serves as a reference implementation for applying agentic principles to other hazardous material distribution systems such as petroleum pipelines, chemical distribution, hydrogen networks, and industrial gas systems.

---

**Related Documents**:
- System Architecture: `documents/2-SoftwareDesignAndArchitecture/`
- Framework Topology Visualizer: `documents/2-SoftwareDesignAndArchitecture/framework-topology-visualizer/`
- Agent Implementation: `internal/agent/`
- Communication System: `internal/communication/`
- Standard Use Case Definition: `documents/1-SoftwareRequirements/requirements/use-cases/standardusecasedefinition.md`

**Related Use Cases**:
- [UC-INFRA-001]: Water Distribution Network (CodeValdMaji)
- [UC-INFRA-002]: Electric Power Distribution (CodeValdStima)
- [UC-INFRA-004]: Telecommunications Network Management
- [UC-INFRA-005]: Smart City Infrastructure Coordination
