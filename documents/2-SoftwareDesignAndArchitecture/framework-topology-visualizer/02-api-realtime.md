# Framework Topology Visualizer - API Contracts & Real-Time Updates

> **Part 2 of 5**: API specifications, WebSocket updates, differential semantics, deterministic IDs, coordinate systems

**Document ID**: ARCH-VIZ-001  
**Component**: Framework Topology Visualizer  
**Version**: 1.0  
**Date**: October 24, 2025  
**Status**: Design Specification for INFRA-017

---

**Navigation**: [README](README.md) | [Overview & Architecture](01-overview-architecture.md) | **[API Contracts & Real-Time]** | [Expression Language & Rendering](03-expression-rendering.md) | [Security & Accessibility](04-security-accessibility.md) | [Testing & Delivery](05-testing-delivery.md)

---

## API Contracts (Enhanced)

### Agent Query API

**Full Request with Pagination & Change Tracking**:
```http
GET /api/v1/agents?usecase_id=UC-INFRA-001&types=pump,pipe,sensor&limit=500&cursor=eyJpZCI6ImFnZW50XzEyMzQ1In0=&updated_since=2025-10-24T10:30:45Z&fields=id,type,status,metadata.location,metadata.downstream_pipes,metadata.flow_capacity
If-None-Match: W/"a1b2c3d4"
```

**Response**:
```json
{
  "agents": [...],
  "pagination": {
    "limit": 500,
    "next_cursor": "eyJpZCI6ImFnZW50XzY3ODkwIn0=",
    "prev_cursor": null,
    "has_more": true,
    "total_hint": 1247
  },
  "server_time": "2025-10-24T10:32:17Z"
}
```

**Response Headers**:
```
ETag: W/"revision-42-timestamp-1729765937"
Cache-Control: private, max-age=30
Vary: Accept, Authorization
```

**Conditional Request** (304 Not Modified when unchanged):
```http
GET /api/v1/agents?...
If-None-Match: W/"revision-42-timestamp-1729765937"

→ 304 Not Modified (no body, save bandwidth)
```

**Incremental Updates** (only changed agents since timestamp):
```http
GET /api/v1/agents?usecase_id=UC-INFRA-001&updated_since=2025-10-24T10:30:45Z

Response:
{
  "agents": [ /* only 3 agents changed */ ],
  "pagination": {...},
  "server_time": "2025-10-24T10:32:17Z",
  "changes_since": "2025-10-24T10:30:45Z"
}
```

**Field Selection** (reduce payload size):
```http
GET /api/v1/agents?fields=id,type,status,metadata.location

Response agents:
[
  {
    "id": "PUMP-001",
    "type": "pump",
    "status": "running",
    "metadata": {
      "location": {"lat": -1.2921, "lon": 36.8219}
    }
    // All other fields omitted
  }
]
```

### WebSocket Real-Time Updates

**Connection**:
```javascript
const ws = new WebSocket('wss://api.codevaldcortex.io/ws/agents?usecase_id=UC-INFRA-001');
```

**Agent Update Event** (JSON Patch RFC 6902):
```json
{
  "type": "agent_update",
  "seq": 142311,
  "ts": "2025-10-24T10:30:45.123Z",
  "agent_id": "SENSOR-023",
  "ops": [
    {"op": "replace", "path": "/status", "value": "alert"},
    {"op": "replace", "path": "/metadata/pressure", "value": 125.7}
  ]
}
```

**Batch Update Event** (multiple agents):
```json
{
  "type": "batch_update",
  "seq": 142312,
  "ts": "2025-10-24T10:30:46.001Z",
  "updates": [
    {"agent_id": "PUMP-001", "ops": [...]},
    {"agent_id": "SENSOR-045", "ops": [...]}
  ]
}
```

### Differential Update Semantics (Critical)

**Problem**: Full replacement breaks animations and selections.

**Solution**: JSON Patch (RFC 6902) operations with **merge semantics**.

**Supported Operations**:
```json
{"op": "replace", "path": "/status", "value": "alert"}          // Simple field
{"op": "replace", "path": "/metadata/pressure", "value": 125.7} // Nested field
{"op": "add", "path": "/metadata/tags/-", "value": "critical"}  // Append to array
{"op": "remove", "path": "/metadata/old_sensor"}                // Remove field
```

**Client Reconciliation Logic**:
```javascript
class AgentStateManager {
  applyUpdate(agentId, ops) {
    const agent = this.agents.get(agentId);
    if (!agent) {
      console.warn(`Agent ${agentId} not found for update`);
      return;
    }
    
    // Apply JSON Patch operations (merge, not replace)
    for (const op of ops) {
      this.applyOperation(agent, op);
    }
    
    // CRITICAL: Preserve UI state during update
    this.preserveSelectionState(agentId);
    this.preserveAnimationState(agentId);
    
    // Trigger incremental re-render (not full redraw)
    this.renderer.updateAgent(agent);
  }
  
  applyOperation(agent, op) {
    const pathParts = op.path.split('/').filter(p => p);
    
    switch (op.op) {
      case 'replace':
        this.setNestedValue(agent, pathParts, op.value);
        break;
      case 'add':
        if (pathParts[pathParts.length - 1] === '-') {
          // Append to array
          const arr = this.getNestedValue(agent, pathParts.slice(0, -1));
          arr.push(op.value);
        } else {
          this.setNestedValue(agent, pathParts, op.value);
        }
        break;
      case 'remove':
        this.deleteNestedValue(agent, pathParts);
        break;
    }
  }
  
  preserveSelectionState(agentId) {
    // Keep agent selected if it was selected before update
    if (this.selection.has(agentId)) {
      this.renderer.highlightAgent(agentId);
    }
  }
  
  preserveAnimationState(agentId) {
    // Don't restart animations unless animation trigger changed
    const agent = this.agents.get(agentId);
    const shouldAnimate = this.evaluateAnimationCondition(agent);
    
    if (shouldAnimate && !this.animating.has(agentId)) {
      this.renderer.startAnimation(agentId);
    } else if (!shouldAnimate && this.animating.has(agentId)) {
      this.renderer.stopAnimation(agentId);
    }
    // If already animating and should continue, don't restart
  }
}
```

**Update Conflict Resolution**:
```javascript
// If updates arrive out of order (seq 143 before seq 142)
applyUpdateWithSequencing(update) {
  if (update.seq <= this.lastAppliedSeq) {
    console.warn(`Skipping stale update seq ${update.seq}`);
    return;
  }
  
  if (update.seq > this.lastAppliedSeq + 1) {
    // Missing intermediate updates - buffer and request replay
    this.pendingUpdates.push(update);
    this.requestReplay(this.lastAppliedSeq + 1, update.seq - 1);
    return;
  }
  
  // Apply in order
  this.applyUpdate(update.agent_id, update.ops);
  this.lastAppliedSeq = update.seq;
  
  // Apply buffered updates if now in order
  this.applyBufferedUpdates();
}
```

**Edge Recomputation After Update**:
```javascript
// When agent metadata changes, recompute affected edges
applyUpdate(agentId, ops) {
  const agent = this.agents.get(agentId);
  const affectsEdges = ops.some(op => 
    op.path.startsWith('/metadata') && 
    this.config.connectionRules.some(rule => 
      op.path.includes(rule.match.replace('$.metadata.', ''))
    )
  );
  
  if (affectsEdges) {
    // Recompute edges from/to this agent
    this.edgeInference.recomputeEdgesFor(agentId);
    this.renderer.updateEdges(agentId);
  }
  
  // Regular field updates (status, etc.)
  this.renderer.updateAgent(agent);
}
```

## Deterministic IDs & Ordering (Anti-Flicker)

### Edge ID Generation (Canonical)

**Problem**: Duplicate edges across updates/replays, flickering animations.

**Solution**: Deterministic edge ID = `hash(from|to|type|configVersion)`

```javascript
import { createHash } from 'crypto';

function generateEdgeId(edge: Edge, configVersion: string): string {
  // Canonical representation (order-independent for undirected)
  const parts = edge.directed 
    ? [edge.from, edge.to, edge.type, configVersion]
    : [edge.from, edge.to].sort().concat([edge.type, configVersion]);
  
  // SHA-256 hash (first 16 chars sufficient for collision resistance)
  const canonical = parts.join('|');
  const hash = createHash('sha256').update(canonical).digest('hex').substring(0, 16);
  
  return `edge_${hash}`;
}

// Example
const edge = {from: 'PUMP-001', to: 'PIPE-015', type: 'supplies', directed: true};
const id = generateEdgeId(edge, '1.0.0');
// → "edge_a3f9c2d1e4b8f7a2"

// Same edge, same ID (idempotent)
const id2 = generateEdgeId(edge, '1.0.0');
assert(id === id2);

// Undirected edges: order doesn't matter
const undirectedA = {from: 'A', to: 'B', type: 'connects_to', directed: false};
const undirectedB = {from: 'B', to: 'A', type: 'connects_to', directed: false};
assert(generateEdgeId(undirectedA, '1.0.0') === generateEdgeId(undirectedB, '1.0.0'));
```

**Benefits**:
- No duplicate edges across WebSocket replays
- Animation state preserved (same edge ID = continue animation)
- Selection preserved across updates
- Config version included prevents stale edge reuse after config changes

### Stable Ordering (Layout Determinism)

**Problem**: Layout flickers on re-render with same data.

**Solution**: Stable sort + seeded RNG

```javascript
class ForceDirectedLayout {
  constructor(config: LayoutConfig) {
    // Use config-specified seed for reproducibility
    this.seed = config.layout.seed || 42;
    this.rng = new SeededRandom(this.seed);
  }
  
  computePositions(agents: Agent[], edges: Edge[]): Map<string, Position> {
    // CRITICAL: Stable sort by agent ID before processing
    const sortedAgents = [...agents].sort((a, b) => a.id.localeCompare(b.id));
    const sortedEdges = [...edges].sort((a, b) => a.id.localeCompare(b.id));
    
    // Initialize positions deterministically
    const positions = new Map();
    for (const agent of sortedAgents) {
      positions.set(agent.id, {
        x: this.rng.next() * this.width,
        y: this.rng.next() * this.height,
      });
    }
    
    // Run force simulation (deterministic with seeded RNG)
    for (let i = 0; i < this.iterations; i++) {
      this.applyForces(sortedAgents, sortedEdges, positions);
    }
    
    return positions;
  }
}

// Seeded RNG (Mulberry32 - fast, deterministic)
class SeededRandom {
  private state: number;
  
  constructor(seed: number) {
    this.state = seed;
  }
  
  next(): number {
    let t = this.state += 0x6D2B79F5;
    t = Math.imul(t ^ t >>> 15, t | 1);
    t ^= t + Math.imul(t ^ t >>> 7, t | 61);
    return ((t ^ t >>> 14) >>> 0) / 4294967296;
  }
}
```

**Config Declaration**:
```json
{
  "layout": {
    "algorithm": "force-directed",
    "seed": 42,
    "options": {
      "iterations": 100
    }
  }
}
```

**Golden Image Test**:
```javascript
test('Layout is deterministic with same seed', () => {
  const layout1 = new ForceDirectedLayout({seed: 42});
  const positions1 = layout1.computePositions(agents, edges);
  
  const layout2 = new ForceDirectedLayout({seed: 42});
  const positions2 = layout2.computePositions(agents, edges);
  
  // Exact same positions
  for (const [agentId, pos1] of positions1.entries()) {
    const pos2 = positions2.get(agentId);
    assert.equal(pos1.x, pos2.x);
    assert.equal(pos1.y, pos2.y);
  }
});
```

**Reconnection Strategy**:

1. **Client State Machine**:
   - `CONNECTING` → initial connection
   - `ONLINE` → receiving updates normally
   - `CATCHING_UP` → replaying missed events after reconnect
   - `OFFLINE` → disconnected, attempting reconnect

2. **Exponential Backoff**:
```javascript
const backoff = Math.min(1000 * Math.pow(2, attemptCount), 30000);
setTimeout(() => reconnect(), backoff + jitter);
```

3. **Replay Window** (catch up after disconnect):
```javascript
// On reconnect, request missed events
ws.send(JSON.stringify({
  type: 'replay',
  from_seq: lastReceivedSeq,
  usecase_id: 'UC-INFRA-001'
}));

// Server responds with buffered events
{
  "type": "replay_response",
  "events": [ /* events 142311-142450 */ ],
  "current_seq": 142450
}
```

4. **Full Resync** (if replay window exceeded):
```javascript
// If too many missed events, full resync
if (missedEvents > 1000) {
  fetchAllAgents(); // HTTP GET /api/v1/agents
}
```

5. **Heartbeat** (detect stale connections):
```json
// Server → Client every 30s
{"type": "ping", "ts": "2025-10-24T10:32:00.000Z"}

// Client → Server
{"type": "pong", "ts": "2025-10-24T10:32:00.023Z"}
```

### WebSocket Backpressure & Buffer Limits

**Server Limits** (`internal/api/websocket/agent_updates.go`):
```go
const (
    MaxBatchSize        = 100  // Max agents per batch_update event
    ReplayWindowSize    = 1000 // Keep last 1000 events for replay
    MaxClientBufferSize = 50   // Drop events if client can't keep up
)

type AgentUpdateHub struct {
    replayBuffer *RingBuffer // Size: ReplayWindowSize
    clients      map[string]*Client
}

func (hub *AgentUpdateHub) BroadcastUpdate(update AgentUpdate) {
    // Add to replay buffer
    hub.replayBuffer.Push(update)
    
    // Send to all clients
    for clientID, client := range hub.clients {
        select {
        case client.send <- update:
            // Sent successfully
        default:
            // Client buffer full - apply drop policy
            client.droppedEvents++
            
            if client.droppedEvents > 10 {
                log.Warnf("Client %s buffer overflow, closing connection", clientID)
                client.Close()
            }
        }
    }
}

func (hub *AgentUpdateHub) HandleReplayRequest(clientID string, fromSeq int) {
    client := hub.clients[clientID]
    
    // Check if replay window covers requested range
    oldestSeq := hub.replayBuffer.OldestSeq()
    currentSeq := hub.replayBuffer.CurrentSeq()
    
    if fromSeq < oldestSeq {
        // Requested events too old - force full resync
        client.send <- WebSocketMessage{
            Type: "resync_required",
            Reason: fmt.Sprintf("replay window exceeded (requested %d, oldest %d)", fromSeq, oldestSeq),
        }
        return
    }
    
    // Send buffered events
    events := hub.replayBuffer.GetRange(fromSeq, currentSeq)
    client.send <- WebSocketMessage{
        Type: "replay_response",
        Events: events,
        CurrentSeq: currentSeq,
    }
}
```

**Client Behavior on Buffer Overflow**:
```javascript
class VisualizationWSClient {
  private pendingUpdates: Queue<Update> = new Queue();
  private readonly MAX_PENDING = 100;
  
  onMessage(event: MessageEvent) {
    const msg = JSON.parse(event.data);
    
    // Add to pending queue
    if (this.pendingUpdates.size >= this.MAX_PENDING) {
      console.error('🚨 Client buffer overflow, triggering full resync');
      this.showToast('Connection overwhelmed, reloading data...', 'warning');
      this.triggerFullResync();
      this.pendingUpdates.clear();
      return;
    }
    
    this.pendingUpdates.enqueue(msg);
    this.processQueue();
  }
  
  async triggerFullResync() {
    this.state = 'RESYNCING';
    
    // Fetch all agents via HTTP
    const agents = await this.httpClient.get('/api/v1/agents?visualization_id=...');
    
    // Replace entire dataset
    this.stateManager.replaceAll(agents);
    
    // Resume WebSocket
    this.state = 'ONLINE';
    this.showToast('Data reloaded successfully', 'success');
  }
  
  showToast(message: string, severity: 'info' | 'warning' | 'error') {
    // UI notification
    const toast = document.createElement('div');
    toast.className = `toast toast-${severity}`;
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
  }
}
```

## Coordinate System Unification

### CRS Declaration (Mandatory)

**Every config MUST explicitly declare coordinate system**:
```json
{
  "crs": {
    "geographic": "EPSG:4326",
    "indoor": {
      "type": "local-xy",
      "origin": {"x": 0, "y": 0},
      "unit": "m",
      "orientation": "cartesian"
    }
  }
}
```

### Geographic Coordinates (Outdoor Use Cases)

**Standard**: WGS84 (EPSG:4326) - lat/lon in decimal degrees

**Input Formats Supported**:
```json
// Standard: lat/lon object
{"lat": -1.2921, "lon": 36.8219}

// Alternative: coordinates array [lon, lat] (GeoJSON)
{"coordinates": [36.8219, -1.2921]}

// Legacy: separate fields
{"latitude": -1.2921, "longitude": 36.8219}
```

**Normalization** (always convert to WGS84 internally):
```go
type GeographicCoordinate struct {
    Lat float64 `json:"lat"` // -90 to +90
    Lon float64 `json:"lon"` // -180 to +180
}

func NormalizeCoordinate(raw map[string]any) (*GeographicCoordinate, error) {
    // Try standard format
    if lat, ok := raw["lat"].(float64); ok {
        if lon, ok := raw["lon"].(float64); ok {
            return &GeographicCoordinate{Lat: lat, Lon: lon}, nil
        }
    }
    
    // Try GeoJSON format
    if coords, ok := raw["coordinates"].([]any); ok && len(coords) == 2 {
        lon := coords[0].(float64)
        lat := coords[1].(float64)
        return &GeographicCoordinate{Lat: lat, Lon: lon}, nil
    }
    
    // Try legacy format
    if lat, ok := raw["latitude"].(float64); ok {
        if lon, ok := raw["longitude"].(float64); ok {
            return &GeographicCoordinate{Lat: lat, Lon: lon}, nil
        }
    }
    
    return nil, fmt.Errorf("no valid geographic coordinate found")
}
```

**Validation**:
```go
func (c *GeographicCoordinate) Validate() error {
    if c.Lat < -90 || c.Lat > 90 {
        return fmt.Errorf("latitude %f out of range [-90, 90]", c.Lat)
    }
    if c.Lon < -180 || c.Lon > 180 {
        return fmt.Errorf("longitude %f out of range [-180, 180]", c.Lon)
    }
    return nil
}
```

### Indoor Coordinates (Warehouse/Facility Use Cases)

**Local XY Coordinate System**:
```json
{
  "crs": {
    "indoor": {
      "type": "local-xy",
      "origin": {"x": 0, "y": 0},
      "unit": "m",
      "orientation": "cartesian",
      "bounds": {
        "minX": 0, "maxX": 200,
        "minY": 0, "maxY": 100
      }
    }
  }
}
```

**Agent Position in Indoor Space**:
```json
{
  "metadata": {
    "position": {"x": 45.2, "y": 23.7}
  }
}
```

**Normalization** (local XY):
```go
type IndoorCoordinate struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

func NormalizeIndoorCoordinate(raw map[string]any, config IndoorCRS) (*IndoorCoordinate, error) {
    pos, ok := raw["position"].(map[string]any)
    if !ok {
        // Try legacy formats
        pos = raw
    }
    
    x, okX := pos["x"].(float64)
    y, okY := pos["y"].(float64)
    if !okX || !okY {
        return nil, fmt.Errorf("no valid indoor coordinate found")
    }
    
    coord := &IndoorCoordinate{X: x, Y: y}
    
    // Validate against bounds
    if err := coord.ValidateWithinBounds(config.Bounds); err != nil {
        return nil, err
    }
    
    return coord, nil
}
```

### Hybrid Use Cases (Mixed Indoor/Outdoor)

**Strategy**: Use role to determine CRS:
```json
{
  "entities": {
    "warehouse": {
      "coordinateType": "indoor"
    },
    "truck": {
      "coordinateType": "geographic"
    }
  }
}
```

**Renderer Handling**:
```javascript
class HybridRenderer {
  projectCoordinate(agent) {
    const entityConfig = this.config.entities[agent.type];
    
    if (entityConfig.coordinateType === 'geographic') {
      // Project lat/lon → screen XY using map projection
      return this.mapProjection.latLonToPixel(
        agent.metadata.location.lat,
        agent.metadata.location.lon
      );
    } else {
      // Direct indoor XY → screen XY with scaling
      return {
        x: agent.metadata.position.x * this.scale,
        y: agent.metadata.position.y * this.scale
      };
    }
  }
}
```

### CRS Conversion (Future)

**If non-WGS84 input** (e.g., local UTM):
```json
{
  "crs": {
    "geographic": "EPSG:32737", // UTM Zone 37S
    "transformTo": "EPSG:4326"
  }
}
```

Use `proj4js` library for conversion:
```javascript
import proj4 from 'proj4';

proj4.defs("EPSG:32737", "+proj=utm +zone=37 +south +datum=WGS84 +units=m +no_defs");

function transformCoordinate(x, y, fromCRS, toCRS) {
  return proj4(fromCRS, toCRS, [x, y]);
}
```

---

**Navigation**: [README](README.md) | [Overview & Architecture](01-overview-architecture.md) | **[API Contracts & Real-Time]** | [Expression Language & Rendering](03-expression-rendering.md) | [Security & Accessibility](04-security-accessibility.md) | [Testing & Delivery](05-testing-delivery.md)
