# Framework Topology Visualizer - Testing & Delivery

> **Part 5 of 5**: Testing strategy, config validation & versioning, performance telemetry, canonical type registry, delivery plan, MVP scope, success metrics

**Document ID**: ARCH-VIZ-001  
**Component**: Framework Topology Visualizer  
**Version**: 1.0  
**Date**: October 24, 2025  
**Status**: Design Specification for INFRA-017

---

**Navigation**: [README](README.md) | [Overview & Architecture](01-overview-architecture.md) | [API Contracts & Real-Time](02-api-realtime.md) | [Expression Language & Rendering](03-expression-rendering.md) | [Security & Accessibility](04-security-accessibility.md) | **[Testing & Delivery]**

---

## Testing Strategy

### 1. Golden Image Tests (Layout Determinism)

**Purpose**: Ensure same data → same visual output (catch layout regressions)

```javascript
describe('Force-Directed Layout', () => {
  it('produces identical output with same seed', async () => {
    const seed = 42;
    const topology1 = await renderTopology(testData, {layout: {seed}});
    const topology2 = await renderTopology(testData, {layout: {seed}});
    
    const diff = pixelDiff(topology1.screenshot, topology2.screenshot);
    expect(diff.pixelsDifferent).toBe(0);
  });
  
  it('matches golden image', async () => {
    const screenshot = await renderTopology(testData, config);
    const golden = loadGoldenImage('force-directed-27-agents.png');
    
    const diff = pixelDiff(screenshot, golden);
    expect(diff.pixelsDifferent).toBeLessThan(100); // tolerance for AA
  });
});
```

**Tools**: Playwright + pixelmatch library

### 2. Performance Tests

**Render Time**:
```javascript
test('Canvas renderer handles 500 nodes in <100ms', async () => {
  const start = performance.now();
  const viz = new CanvasRenderer(generate500Nodes());
  viz.render();
  const duration = performance.now() - start;
  
  expect(duration).toBeLessThan(100);
});
```

**Frame Rate**:
```javascript
test('Maintains 30 FPS with 1000 animated nodes', async () => {
  const viz = new CanvasRenderer(generate1000AnimatedNodes());
  const fps = await measureFPS(viz, duration: 5000);
  
  expect(fps).toBeGreaterThan(30);
});
```

**Memory**:
```javascript
test('Memory usage stays under 100MB for 5000 nodes', async () => {
  const initialMemory = performance.memory.usedJSHeapSize;
  const viz = new CanvasRenderer(generate5000Nodes());
  viz.render();
  const finalMemory = performance.memory.usedJSHeapSize;
  
  const memoryUsed = (finalMemory - initialMemory) / 1024 / 1024;
  expect(memoryUsed).toBeLessThan(100);
});
```

### 3. Accessibility Tests

```javascript
const { AxePuppeteer } = require('@axe-core/puppeteer');

test('Visualization is WCAG 2.2 AA compliant', async () => {
  await page.goto('http://localhost:8083/visualization/water-network');
  const results = await new AxePuppeteer(page).analyze();
  
  expect(results.violations).toHaveLength(0);
});

test('Keyboard navigation works', async () => {
  await page.keyboard.press('Tab');
  const focused = await page.evaluate(() => document.activeElement.getAttribute('aria-label'));
  expect(focused).toContain('PUMP-001');
  
  await page.keyboard.press('Enter');
  const detailsVisible = await page.isVisible('.agent-details-panel');
  expect(detailsVisible).toBe(true);
});
```

### 4. WebSocket Reconnection Tests

```javascript
test('Reconnects after network interruption', async () => {
  const client = new VisualizationWSClient(url);
  await client.connect();
  
  // Simulate network failure
  await client.socket.close();
  
  // Should reconnect automatically
  await wait(2000);
  expect(client.state).toBe('CATCHING_UP');
  
  // Should replay missed events
  await wait(1000);
  expect(client.state).toBe('ONLINE');
  expect(client.missedEvents).toBe(0);
});

test('Handles replay window overflow', async () => {
  const client = new VisualizationWSClient(url);
  await client.connect();
  client.lastSeq = 100;
  
  // Server's replay window: seq 1000-2000 (client missed too much)
  await client.socket.close();
  await wait(10000); // long disconnect
  
  await client.reconnect();
  
  // Should trigger full resync
  expect(client.state).toBe('RESYNCING');
  expect(client.httpFetchCalled).toBe(true);
});
```

### 5. Load Tests (Synthetic Data)

```bash
# Generate test topology
go run scripts/generate_test_topology.go \
  --nodes=5000 \
  --edge-probability=0.02 \
  --output=test/fixtures/topology-5000nodes.json

# Run load test
k6 run test/load/visualization_load_test.js \
  --vus=50 \
  --duration=5m
```

```javascript
// k6 script
export default function() {
  const res = http.get('http://localhost:8083/visualization/water-network');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'load time < 1s': (r) => r.timings.duration < 1000,
  });
}
```

### 6. Edge Inference Tests (Table-Driven)

See [01-overview-architecture.md](01-overview-architecture.md) for complete table-driven test examples.

### 7. Config Schema Validation Tests

```go
func TestConfigValidator_ValidConfigs(t *testing.T) {
    validConfigs := []string{
        "testdata/configs/water-network.json",
        "testdata/configs/vehicle-tracking.json",
        "testdata/configs/warehouse.json",
    }
    
    for _, path := range validConfigs {
        t.Run(path, func(t *testing.T) {
            config, err := LoadConfig(path)
            require.NoError(t, err, "valid config should load without error")
            assert.NotNil(t, config)
        })
    }
}

func TestConfigValidator_InvalidConfigs(t *testing.T) {
    tests := []struct {
        name        string
        config      string
        expectedErr string
    }{
        {
            name: "Missing schemaVersion",
            config: `{"visualization": {"id": "test"}}`,
            expectedErr: "schemaVersion is mandatory",
        },
        {
            name: "Missing $schema",
            config: `{"schemaVersion": "1.0.0", "visualization": {"id": "test"}}`,
            expectedErr: "$schema is required",
        },
        {
            name: "Invalid expression - recursive descent",
            config: `{
                "schemaVersion": "1.0.0",
                "$schema": "...",
                "visualization": {
                    "id": "test",
                    "connections": {
                        "strategy": "metadata",
                        "match": "$..password"
                    }
                }
            }`,
            expectedErr: "recursive descent (..) forbidden",
        },
        {
            name: "Invalid CRS",
            config: `{
                "schemaVersion": "1.0.0",
                "$schema": "...",
                "crs": {"geographic": "INVALID"},
                "visualization": {"id": "test"}
            }`,
            expectedErr: "invalid CRS code",
        },
        {
            name: "Expression depth exceeded",
            config: `{
                "schemaVersion": "1.0.0",
                "$schema": "...",
                "visualization": {
                    "id": "test",
                    "connections": {
                        "match": "$.a.b.c.d.e.f.g.h.i.j.k.l.m"
                    }
                }
            }`,
            expectedErr: "expression depth 13 exceeds max 10",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := LoadConfig(strings.NewReader(tt.config))
            require.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedErr)
        })
    }
}
```

## Config Validation & Versioning Enforcement

See [README.md](README.md) for complete schema versioning and migration specifications.

## Performance Telemetry (Runtime Observability)

See [README.md](README.md) for complete telemetry and SLO tracking implementation.

## Canonical Relationship Type Registry

See [README.md](README.md) for complete canonical type taxonomy and CI validation pipeline.

## Delivery Plan (Revised - Realistic)

### Phase 0: Design Validation (3 days)

**Tasks**:
1. Create JSON Schema for visualization config
2. Implement config validator in Go with **mandatory version check**
3. Create unit tests for validation (including version rejection tests)
4. Implement migration mechanism (v1.0 → v1.1 example)
5. Performance-test renderer thresholds (100, 500, 1000, 5000 nodes)
6. Validate graph theory model with sample data
7. Create edge inference unit tests (table-driven)

**Deliverables**:
- ✅ `internal/web/visualization/config-schema.json`
- ✅ `internal/web/visualization/config_validator.go` (with version enforcement)
- ✅ `internal/web/visualization/config_validator_test.go` (version rejection tests)
- ✅ `internal/web/visualization/migrations.go` (migration registry)
- ✅ `internal/web/visualization/edge_inference_test.go` (table-driven tests)
- ✅ Performance benchmark results (thresholds validated)
- ✅ Sample configs for 3 use cases (all with schemaVersion)

### Phase 1: Core Rendering (5 days)

**Tasks**:
1. Agent API client with pagination & ETag support
2. JSONPath expression evaluator
3. Edge inference engine (connection_rules)
4. Canvas renderer (baseline)
5. Force-directed layout
6. Basic interaction (pan, zoom, select)

**Deliverables**:
- ✅ `static/js/visualization/agent-data-source.js` (with pagination)
- ✅ `static/js/visualization/expression-evaluator.js` (JSONPath)
- ✅ `static/js/visualization/edge-inference.js`
- ✅ `static/js/visualization/renderers/canvas-renderer.js`
- ✅ `static/js/visualization/layouts/force-directed.js`
- ✅ Works with 500-node synthetic dataset

### Phase 2: Real-Time Updates (3 days)

**Tasks**:
1. WebSocket client with reconnection
2. Differential update handling (JSON Patch ops)
3. Animation/interpolation
4. Backpressure/rate limiting

**Deliverables**:
- ✅ `static/js/visualization/websocket-client.js` (state machine)
- ✅ Real-time agent updates reflected in <1s
- ✅ Graceful handling of disconnect/reconnect

### Phase 3: UC-INFRA-001 Integration (4 days)

**Tasks**:
1. Geographic layout (lat/lon → screen coordinates)
2. Basemap integration (MapTiler or OSM)
3. Water network config file
4. Security: role-based visibility
5. A11y: keyboard navigation & ARIA
6. i18n: en-KE and sw-KE locales

**Deliverables**:
- ✅ `usecases/UC-INFRA-001-tumaini/config/visualization.json`
- ✅ Working water network visualization
- ✅ WCAG 2.2 AA compliant
- ✅ English & Swahili UI

### Phase 4: Testing & Documentation (3 days)

**Tasks**:
1. Golden image tests (5 scenarios)
2. Performance tests (render time, FPS, memory)
3. A11y tests (axe-core)
4. Load tests (k6)
5. Developer documentation
6. User guide
7. Demo video

**Deliverables**:
- ✅ `test/visualization/golden-images/` (5 reference screenshots)
- ✅ `test/visualization/performance_test.js`
- ✅ `test/visualization/accessibility_test.js`
- ✅ `documents/3-SofwareDevelopment/visualization-framework-guide.md`
- ✅ Demo video showcasing UC-INFRA-001

**Total Timeline**: 18 days (3.5 weeks)

## Razor-Thin MVP Cut (Fastest Path to Production)

**Goal**: Minimum viable implementation that delivers core value.

**Scope** (MVP-only features):

### Included ✅
1. **Canvas Renderer** (only)
   - No SVG, no WebGL
   - Handles 500 nodes reliably
   - Simpler lifecycle, faster implementation

2. **Layouts** (2 only)
   - **Geographic**: lat/lon → screen XY (for UC-INFRA-001, UC-TRACK-001)
   - **Force-Directed**: physics-based (for UC-COMM-001)
   - Deterministic with seed

3. **Data Source**: HTTP Agent API
   - `GET /api/v1/agents` with pagination
   - ETag support for conditional requests
   - **Polling only** (no WebSocket real-time yet)

4. **Update Mechanism**: JSON Patch
   - Differential updates (not full replacement)
   - Preserve selection/animation state
   - 30s polling interval

5. **Connection Rules Inference**
   - Strategy 2: role `connection_rules`
   - JSONPath expressions (sandboxed)
   - Canonical relationship taxonomy

6. **RBAC** (Server-Side)
   - Row-level filtering in AQL
   - Field masking (coarse GPS)
   - Edge type filtering

7. **Config Validation**
   - Mandatory `schemaVersion`
   - Expression validation (whitelist, depth, timeout)
   - JSON Schema validation

8. **Canonical Types**
   - `canonical_types.json` registry
   - CI lint validation
   - Default styles per category

9. **Keyboard Navigation**
   - Tab/Arrow keys to select agents
   - Enter to show details
   - Esc to clear selection

10. **Telemetry**
    - FPS, render time, dropped frames
    - Console metrics (no overlay yet)
    - SLO tracking (p95 initial render, update)

### Deferred 🔄 (Post-MVP)
- ❌ WebSocket real-time updates (use polling first)
- ❌ SVG renderer (Canvas sufficient)
- ❌ WebGL renderer (not needed for <1000 nodes)
- ❌ Advanced layouts (hierarchical, circular, timeline, grid)
- ❌ Basemap integration (plain background + grid)
- ❌ Animations (pulse, flow)
- ❌ Indoor coordinates (focus on geographic first)
- ❌ Multiple layers (single layer MVP)
- ❌ Search/filter UI (select by clicking only)
- ❌ Export (PNG, SVG, PDF)
- ❌ Performance overlay (console metrics sufficient)

### MVP Delivery Plan (Compressed)

**Phase 0: Foundation** (2 days)
- Config validator with version enforcement
- Edge inference engine with table-driven tests
- Canonical types registry + CI validation

**Phase 1: Core Rendering** (3 days)
- Canvas renderer (lifecycle contract)
- Geographic layout (lat/lon projection)
- Force-directed layout (seeded RNG)
- Agent API client (pagination, ETag)
- JSONPath expression evaluator (sandboxed)

**Phase 2: Interactions** (2 days)
- Pan/zoom controls
- Keyboard navigation (Tab/Arrow/Enter/Esc)
- Agent selection + details panel
- Textual summary toggle (A11y)

**Phase 3: UC-INFRA-001 Integration** (2 days)
- Water network config (`visualization.json`)
- RBAC enforcement (server-side)
- Polling updates (30s interval, JSON Patch)
- Dashboard integration

**Phase 4: Testing** (2 days)
- Golden image tests (deterministic layout)
- RBAC integration tests
- Performance benchmarks (500 nodes)
- A11y tests (axe-core, keyboard nav)

**MVP Timeline**: **11 days** (2 weeks)

**Post-MVP** (Phase 5+):
- WebSocket real-time (Week 3)
- Basemap integration (Week 4)
- Advanced layouts + WebGL (Week 5-6)

**Rationale**: 
- MVP proves core concept (topology viz + RBAC + config-driven)
- Canvas handles most use cases (<1000 nodes)
- Polling is sufficient for non-real-time use cases
- Can iterate based on user feedback before building real-time/WebGL

## Bundle Size (Realistic Estimate)

**Core Libraries** (minified + gzipped):
- `topology-visualizer.js`: ~25 KB
- `jsonpath-plus`: ~8 KB
- `d3-force` (layout): ~15 KB
- `maplibre-gl` (basemap): ~180 KB
- **Total**: ~228 KB

**Optimization Strategies**:
1. **Code Splitting**: Load basemap only when `layout.algorithm === 'geographic'`
   - Without basemap: ~48 KB
   - With basemap: ~228 KB
2. **Tree Shaking**: Import only needed D3 modules
3. **Lazy Loading**: Load WebGL renderer only when needed
4. **CDN**: Serve maplibre-gl from CDN (user's browser may cache)

**Revised Target**: **< 50 KB core + < 200 KB geographic** (realistic for production)

## Success Metrics

### Performance SLOs (Service Level Objectives)

**Production Targets** (wire to telemetry):

| Metric | SLO Target | Measurement | Telemetry Key |
|--------|------------|-------------|---------------|
| **Initial Render** | p95 < 800ms @ 500 nodes | Time from data load to first paint | `render.initial.duration` |
| **Incremental Update** | p95 < 50ms | Time to apply JSON Patch and re-render | `render.update.duration` |
| **Frame Rate** | p95 ≥ 30 FPS | Measured over 10s window | `render.fps` |
| **Dropped Frames** | < 5% of total frames | Count frames > 33ms (30 FPS threshold) | `render.droppedFrames` |
| **WebSocket Reconnect** | p95 < 2s | Time from disconnect to ONLINE state | `ws.reconnect.duration` |
| **API Response Time** | p95 < 300ms | GET /api/v1/agents latency | `api.agents.latency` |
| **Memory Footprint** | < 150MB @ 1000 nodes | Heap size after initial render | `memory.heapSize` |

See [README.md](README.md) for complete telemetry implementation.

### Framework Metrics
- ✅ Config-driven: 0 custom code for basic visualizations
- ✅ Reusable: 5+ use cases using same component
- ✅ Performance: 30+ FPS with 1000 agents (Canvas), 60+ FPS with 5000 agents (WebGL)
- ✅ Update Latency: < 500ms for agent state changes (HTTP), < 100ms (WebSocket)
- ✅ Bundle Size: < 50KB core, < 250KB with basemap (realistic)
- ✅ **SLOs**: p95 initial render < 800ms, p95 update < 50ms @ 500 nodes

### UC-INFRA-001 Metrics
- ✅ Render 27 agents with 0 lag
- ✅ Real-time updates within 1 second
- ✅ Interactive: Click → agent details < 50ms
- ✅ Responsive: Works on mobile devices
- ✅ Accessible: Keyboard navigation supported

## Future Enhancements

1. **3D Visualization**: WebGL-based 3D topology for buildings, terrain
2. **Time Travel**: Replay historical agent states
3. **Heatmaps**: Density visualization for activity patterns
4. **Clustering**: Automatic grouping of related agents
5. **Collision Detection**: Visual alerts for spatial conflicts
6. **Export**: PNG, SVG, PDF export functionality
7. **Collaborative**: Multi-user synchronized view
8. **AR/VR**: Immersive visualization for spatial use cases

## References

- **INFRA-017**: Network Topology Visualizer (This Implementation)
- **INFRA-016**: Framework Web UI (Base Dashboard)
- **UC-INFRA-001**: Water Distribution Network (Primary Use Case)
- **UC-TRACK-001**: Safiri Salama (Real-time Tracking)
- **UC-WMS-001**: Warehouse Management (Grid Layout)

---

**Status**: Design specification complete, ready for implementation in INFRA-017  
**Target Completion**: 2 weeks from start  
**Effort**: High (Framework component creation)  
**Impact**: High (Enables visualization for all use cases)

---

**Navigation**: [README](README.md) | [Overview & Architecture](01-overview-architecture.md) | [API Contracts & Real-Time](02-api-realtime.md) | [Expression Language & Rendering](03-expression-rendering.md) | [Security & Accessibility](04-security-accessibility.md) | **[Testing & Delivery]**
