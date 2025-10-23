# Leak Detection Scenario# INFRA-009: Leak Detection Scenario



This scenario demonstrates the multi-agent leak detection workflow in the UC-INFRA-001 Water Distribution Network.## Overview



## WorkflowThis scenario demonstrates the CodeValdCortex framework's multi-agent coordination capabilities through a realistic water distribution network leak detection and isolation workflow.



1. **Sensor detects anomaly**: Pressure sensor detects significant pressure drop## Scenario Flow

2. **Publishes alert**: Alert published to `zone.central.leak.detected` topic

3. **Pipe analyzes**: Pipe agent analyzes the situation and confirms leak### Step 1: Sensor Detects Anomaly

4. **Valve isolation**: Zone coordinator orders upstream/downstream valves to close**Agent**: `SENSOR-001` (Pressure Sensor North Main)

5. **Escalation**: Incident escalated to control room for maintenance dispatch- Detects pressure drop from expected 6.0 bar to 4.5 bar (-25% deviation)

- **Publishes** alert to pub/sub topic: `zone.north.leak.detected`

## Usage- **Sends** direct message to `PIPE-001` for immediate notification



```bash### Step 2: Pipe Agent Analyzes Alert

# Ensure CodeValdCortex framework is running with agent instances**Agent**: `PIPE-001` (North Main Pipe)

cd /workspaces/CodeValdCortex/usecases/UC-INFRA-001-water-distribution-network- Receives sensor notification

./start.sh- Analyzes pressure/flow correlation

- Confirms leak probability: 85%, estimated loss: 50 L/min

# In another terminal, run the scenario- **Publishes** confirmation to topic: `zone.north.leak.confirmed`

cd scenarios/leak_detection- **Sends** CLOSE commands to isolation valves:

go run main.go  - `VALVE-001` (upstream)

```  - `VALVE-002` (downstream)



## Expected Output### Step 3: Valve Agents Isolate Section

**Agents**: `VALVE-001`, `VALVE-002`

The scenario will demonstrate:- Receive CLOSE commands from pipe agent

- ✅ Agent discovery and validation- Execute valve closure (30-second transition)

- 📊 Pressure anomaly detection (2.1 PSI vs 4.0 PSI threshold)  - Update position from 100% open to 0% (fully closed)

- 🔧 Pipe analysis with 87% confidence- **Publish** status updates to topic: `zone.north.valve.status`

- 🚪 Valve isolation commands

- 🚨 Control room escalation### Step 4: Zone Coordinator Escalates

- 📈 Performance metrics (6 messages, 10-second response time)**Agent**: `COORDINATOR-001` (North Zone Coordinator)

- Aggregates incident data from all agents

## Dependencies- Creates escalation report with:

  - Leak probability, estimated loss

- CodeValdCortex framework running on localhost:8083  - Affected pipe and sensors

- Agent instances created (sensor, pipe, valve, zone_coordinator types)  - Isolation status

- ArangoDB for message persistence  - Maintenance requirements

- **Publishes** escalation to topic: `control.room.escalation`

## Messages Sent- Triggers maintenance crew dispatch



1. **Pub/Sub**: `zone.central.leak.detected` (pressure anomaly)## Agent Communication Patterns

2. **Direct**: Pipe → Zone Coordinator (analysis)

3. **Direct**: Zone Coordinator → Upstream Valve (close command)### Direct Messaging (Point-to-Point)

4. **Direct**: Zone Coordinator → Downstream Valve (close command)- Sensor → Pipe: Anomaly notification

5. **Pub/Sub**: `control_room.incidents.high_priority` (escalation)- Pipe → Valves: Close commands

### Publish/Subscribe (Broadcast)
- Sensor publishes: `zone.north.leak.detected`
- Pipe publishes: `zone.north.leak.confirmed`
- Valves publish: `zone.north.valve.status`
- Coordinator publishes: `control.room.escalation`

### Message Metadata
- **Priority levels**: 1-10 (8-10 for critical alerts/commands)
- **TTL**: 60s for commands, 1-2 hours for alerts
- **Correlation IDs**: Link related messages across agents
- **Custom metadata**: Zone, severity, action required

## Prerequisites

1. **ArangoDB** running and accessible
2. **Agent instances** created in database:
   - `SENSOR-001`: Pressure sensor
   - `PIPE-001`: Pipe agent
   - `VALVE-001`, `VALVE-002`: Valve agents
   - `COORDINATOR-001`: Zone coordinator

3. **Environment** configured in `.env` file (parent directory)

## Running the Scenario

### Option 1: Direct Execution
```bash
cd /workspaces/CodeValdCortex/usecases/UC-INFRA-001-water-distribution-network/scenarios/leak_detection
go run main.go
```

### Option 2: Build and Run
```bash
cd /workspaces/CodeValdCortex/usecases/UC-INFRA-001-water-distribution-network/scenarios/leak_detection
go build -o leak_detection main.go
./leak_detection
```

## Expected Output

```
=== Starting Leak Detection Scenario ===
Scenario agents: Sensor=SENSOR-001, Pipe=PIPE-001, Valves=[VALVE-001, VALVE-002], Coordinator=COORDINATOR-001

--- Step 1: Sensor Detects Anomaly ---
[SENSOR-001] Detected pressure anomaly: 4.5 bar (expected: 6.0 bar, deviation: -25.0%)
[SENSOR-001] Published alert to topic 'zone.north.leak.detected' (publication_id: pub-xxx)
[SENSOR-001] Sent direct notification to pipe PIPE-001 (message_id: msg-xxx)

--- Step 2: Pipe Agent Analyzes Alert ---
[PIPE-001] Analysis complete: Leak probability=85%, Estimated loss=50.0 L/min
[PIPE-001] Published leak confirmation to topic 'zone.north.leak.confirmed' (publication_id: pub-xxx)
[PIPE-001] Sent CLOSE command to upstream valve VALVE-001 (message_id: msg-xxx)
[PIPE-001] Sent CLOSE command to downstream valve VALVE-002 (message_id: msg-xxx)

--- Step 3: Valves Isolate Section ---
[VALVE-001] Valve closed (position: 0%, duration: 30s)
[VALVE-001] Published status update to topic 'zone.north.valve.status' (publication_id: pub-xxx)
[VALVE-002] Valve closed (position: 0%, duration: 30s)
[VALVE-002] Published status update to topic 'zone.north.valve.status' (publication_id: pub-xxx)

--- Step 4: Zone Coordinator Escalates ---
[COORDINATOR-001] Escalation report prepared:
{
  "coordinator_id": "COORDINATOR-001",
  "zone": "north",
  "incident_type": "leak_detected_and_isolated",
  ...
}
[COORDINATOR-001] Published escalation to topic 'control.room.escalation' (publication_id: pub-xxx)
✅ Incident escalated to control room for maintenance crew dispatch

=== Leak Detection Scenario Complete ===
```

## Verification

After running the scenario, check ArangoDB collections:

### 1. Messages (`agent_messages`)
```javascript
// Query for messages sent during scenario
FOR msg IN agent_messages
  FILTER msg.created_at >= "2025-10-23T00:00:00Z"
  SORT msg.created_at DESC
  RETURN {
    from: msg.from_agent_id,
    to: msg.to_agent_id,
    type: msg.message_type,
    priority: msg.priority,
    status: msg.status
  }
```

### 2. Publications (`agent_publications`)
```javascript
// Query for publications
FOR pub IN agent_publications
  FILTER pub.published_at >= "2025-10-23T00:00:00Z"
  SORT pub.published_at DESC
  RETURN {
    publisher: pub.publisher_agent_id,
    event: pub.event_name,
    type: pub.publication_type,
    metadata: pub.metadata
  }
```

### 3. Expected Results
- **4 direct messages**: Sensor→Pipe, Pipe→Valve1, Pipe→Valve2, Coordinator→ControlRoom
- **5 publications**: leak.detected, leak.confirmed, valve1.status, valve2.status, escalation
- **Message priorities**: 8-10 for critical alerts/commands
- **Publication types**: Alert, StatusChange

## Framework Features Demonstrated

✅ **Direct Messaging**
- Point-to-point communication between agents
- Message priority and TTL handling
- Command/response patterns

✅ **Publish/Subscribe**
- Topic-based event broadcasting
- Pattern matching (e.g., `zone.north.*`)
- Multiple subscribers on same topic

✅ **Message Metadata**
- Custom metadata for context
- Correlation IDs for message chains
- Priority and urgency flags

✅ **Agent Coordination**
- Multi-agent workflow orchestration
- Autonomous decision-making
- Event-driven responses

## Integration with Web UI

Once the scenario runs, you can:

1. **View Messages**: Navigate to http://localhost:8083 and check message logs
2. **Monitor Agents**: See agent states updated in real-time
3. **Track Publications**: View pub/sub events across the network
4. **Analyze Flows**: Follow message chains via correlation IDs

## Next Steps

1. **Subscribe Agents**: Implement subscription handlers so agents can react to publications automatically
2. **Agent State Updates**: Have agents update their state based on scenario events
3. **Real-time Visualization**: Build topology map showing leak location and isolation
4. **Historical Analysis**: Store scenario data for replay and analysis
5. **Automated Testing**: Create test suite to verify scenario execution

## Related Documentation

- **Use Case Requirements**: `/documents/1-SoftwareRequirements/requirements/use-cases/UC-INFRA-001-water-distribution-network.md`
- **Communication System**: `/documents/3-SofwareDevelopment/core-systems/agent-communication.md`
- **MVP Progress**: `/documents/3-SofwareDevelopment/Usecases/UC-INFRA-001-water-distribution-network/mvp.md`

## Notes

This is a **simulation scenario** that demonstrates the communication infrastructure. For a production system, agents would:
- Actually read sensor data from IoT devices
- Control real valves via SCADA/Modbus
- Update physical infrastructure state
- Integrate with work order management systems
