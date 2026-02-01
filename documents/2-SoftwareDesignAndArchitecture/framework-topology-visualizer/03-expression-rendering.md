# Framework Topology Visualizer - Expression Language & Rendering

> **Part 3 of 5**: JSONPath expressions, security sandboxing, renderer selection, lifecycle contract, basemap configuration

**Document ID**: ARCH-VIZ-001  
**Component**: Framework Topology Visualizer  
**Version**: 1.0  
**Date**: October 24, 2025  
**Status**: Design Specification for INFRA-017

---

**Navigation**: [README](README.md) | [Overview & Architecture](01-overview-architecture.md) | [API Contracts & Real-Time](02-api-realtime.md) | **[Expression Language & Rendering]** | [Security & Accessibility](04-security-accessibility.md) | [Testing & Delivery](05-testing-delivery.md)

---

## Expression Language Specification

**Dialect**: JSONPath (RFC 9535)

### Match Expressions

**Simple Field Reference**:
```json
"match": "$.metadata.downstream_pipes"
```

Evaluates to array: `["PIPE-015", "PIPE-016"]`

**Conditional Match**:
```json
"match": "$.metadata.connections[?(@.type == 'supply')]"
```

Returns array of connection objects where type is 'supply'.

**Array Flattening**:
```json
"match": "$.metadata.routes[*].stops[*].id"
```

Flattens nested arrays to single list of stop IDs.

### Weight Expressions

**Numeric Field**:
```json
"weight": "$.metadata.flow_capacity"
```

Returns: `450.5` (numeric value for edge weight)

**Computed Weight**:
```json
"weight": "$.metadata.priority"
```

Returns: `1` (high), `2` (medium), `3` (low)

**Conditional Default**:
```json
"weight": "$.metadata.distance || 1.0"
```

Use distance if available, else 1.0.

### Security Sandbox (Critical)

**1. Allowed Root Paths** (Whitelist only):
```go
var allowedRootPaths = []string{
    "$.id",
    "$.type",
    "$.name", 
    "$.status",
    "$.metadata",
    "$.state",
}
```

**Forbidden patterns**:
- ❌ `$..` (recursive descent - DoS risk, information disclosure)
- ❌ `$.credentials`, `$.secrets`, `$.internal` (sensitive fields)
- ❌ `$.system`, `$.admin` (privileged data)

**2. Execution Limits**:
- **Max Path Segments**: 10 (e.g., `$.metadata.a.b.c...` up to 10 levels)
- **Max Array Expansion**: 1000 elements (prevent memory explosion)
- **Timeout**: 10ms per expression
- **Max String Length**: 10KB (prevent regex DoS on string operations)

**3. Validation at Config Load Time**:
```go
func validateExpression(expr string) error {
    // Parse expression
    path, err := jp.ParseString(expr)
    if err != nil {
        return fmt.Errorf("invalid JSONPath: %w", err)
    }
    
    // Check root path whitelist
    root := path.String()[0:strings.Index(path.String()[1:], ".")+1]
    if !slices.Contains(allowedRootPaths, root) {
        return fmt.Errorf("root path %s not in whitelist", root)
    }
    
    // Reject recursive descent
    if strings.Contains(expr, "..") {
        return fmt.Errorf("recursive descent (..) forbidden for security")
    }
    
    // Check depth
    depth := strings.Count(expr, ".")
    if depth > 10 {
        return fmt.Errorf("expression depth %d exceeds max 10", depth)
    }
    
    return nil
}
```

**4. Runtime Enforcement**:
```go
func evaluateWithTimeout(expr string, data map[string]any) (any, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
    defer cancel()
    
    resultChan := make(chan any, 1)
    errChan := make(chan error, 1)
    
    go func() {
        result := jp.Get(data, expr)
        if len(result) > 1000 {
            errChan <- fmt.Errorf("result size %d exceeds limit 1000", len(result))
            return
        }
        resultChan <- result
    }()
    
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errChan:
        return nil, err
    case <-ctx.Done():
        return nil, fmt.Errorf("expression evaluation timeout")
    }
}
```

### Expression Guardrails (Rate Limiting)

**Config Limits**:
```json
{
  "expressions": {
    "dialect": "jsonpath",
    "limits": {
      "maxExpressionsPerConfig": 100,
      "maxEvalsPerAgentPerTick": 50,
      "logRejections": true
    }
  }
}
```

**Validation at Config Load**:
```go
func validateExpressionLimits(config *VisualizationConfig) error {
    totalExpressions := 0
    
    for _, rules := range config.ConnectionRules {
        for _, rule := range rules {
            totalExpressions += 2 // match + weight expressions
        }
    }
    
    for _, entity := range config.Entities {
        if entity.Animation != nil && entity.Animation.When != "" {
            totalExpressions++
        }
    }
    
    limit := config.Expressions.Limits.MaxExpressionsPerConfig
    if totalExpressions > limit {
        return fmt.Errorf("config has %d expressions, exceeds limit %d", totalExpressions, limit)
    }
    
    return nil
}
```

**Runtime Rate Limiting** (per-agent eval cap):
```javascript
class ExpressionEvaluator {
  private evalCounts = new Map<string, number>(); // agentId → count
  private tickStart = Date.now();
  
  evaluate(agentId: string, expr: string, data: any): any {
    // Reset counter every tick (60 FPS = ~16ms ticks)
    const now = Date.now();
    if (now - this.tickStart > 16) {
      this.evalCounts.clear();
      this.tickStart = now;
    }
    
    // Check per-agent limit
    const count = this.evalCounts.get(agentId) || 0;
    if (count >= this.config.expressions.limits.maxEvalsPerAgentPerTick) {
      if (this.config.expressions.limits.logRejections) {
        console.warn(`⚠️  Agent ${agentId} exceeded expression eval limit (${count}/tick)`);
        this.metrics.expressionRejections++;
      }
      return null; // Skip evaluation, use default
    }
    
    this.evalCounts.set(agentId, count + 1);
    
    // Evaluate with timeout
    try {
      return JSONPath({path: expr, json: data, timeout: 10});
    } catch (err) {
      if (this.config.expressions.limits.logRejections) {
        console.error(`❌ Expression eval failed for ${agentId}: ${err.message}`);
        this.metrics.expressionErrors++;
      }
      return null;
    }
  }
}
```

**Telemetry Metrics**:
```javascript
{
  expressionRejections: 0,  // Count of rate-limited evals
  expressionErrors: 0,      // Count of failed evals
  expressionTimeouts: 0     // Count of timeout failures
}
```

### Implementation

**Go Library**: `github.com/ohler55/ojg` (fast JSONPath implementation)

```go
import "github.com/ohler55/ojg/jp"

func evaluateMatch(agentData map[string]any, expr string) ([]string, error) {
    path, err := jp.ParseString(expr)
    if err != nil {
        return nil, err
    }
    result := path.Get(agentData)
    return toStringArray(result), nil
}
```

**JavaScript Library**: `jsonpath-plus`

```javascript
import { JSONPath } from 'jsonpath-plus';

function evaluateWeight(agentData, expr) {
  const result = JSONPath({path: expr, json: agentData});
  return result[0] || 1.0; // default weight
}
```

## Renderer Selection Heuristic

**Problem**: Arbitrary thresholds (300/5000 nodes) are unrealistic without validation.

**Data-Driven Approach**:

```javascript
function selectRenderer(nodes, edges, config) {
  const nodeCount = nodes.length;
  const edgeCount = edges.length;
  const hasAnimation = nodes.some(n => n.animated);
  const animatedCount = nodes.filter(n => n.animated).length;
  const hasGeographic = config.layout.algorithm === 'geographic';
  
  // User override
  if (config.renderer.preferred !== 'auto') {
    return config.renderer.preferred;
  }
  
  // Performance-based selection
  if (nodeCount <= 200 && !hasAnimation) {
    return 'svg'; // High quality, simple interactions
  }
  
  if (nodeCount > 5000 || animatedCount > 100) {
    return 'webgl'; // GPU acceleration required
  }
  
  if (hasGeographic && nodeCount > 1000) {
    return 'webgl'; // Tile layers + many markers = WebGL
  }
  
  // Default: Canvas (best balance)
  return 'canvas';
}
```

**Configurable Thresholds** (validated via performance testing):
```json
{
  "renderer": {
    "preferred": "auto",
    "thresholds": {
      "svgMaxNodes": 200,
      "svgMaxAnimated": 0,
      "canvasMaxNodes": 5000,
      "canvasMaxAnimated": 100,
      "webglMinNodes": 5000
    },
    "fallback": "canvas"
  }
}
```

## Renderer Lifecycle Contract

**All renderers MUST implement this interface**:

```typescript
interface IRenderer {
  // Phase 1: Initialization
  init(container: HTMLElement, config: VisualizationConfig): Promise<void>;
  
  // Phase 2: Initial Render
  render(agents: Agent[], edges: Edge[]): Promise<void>;
  
  // Phase 3: Incremental Updates
  updateAgent(agent: Agent): void;
  updateEdge(edge: Edge): void;
  removeAgent(agentId: string): void;
  removeEdge(edgeId: string): void;
  
  // Phase 4: Interaction
  selectAgent(agentId: string): void;
  deselectAgent(agentId: string): void;
  highlightEdge(edgeId: string): void;
  
  // Phase 5: Viewport Control
  panTo(x: number, y: number, duration?: number): void;
  zoomTo(level: number, center?: {x: number, y: number}): void;
  fitBounds(bounds: BoundingBox): void;
  
  // Phase 6: Cleanup
  destroy(): void;
}
```

**Lifecycle State Machine**:
```
UNINITIALIZED → init() → INITIALIZED
INITIALIZED → render() → RENDERED
RENDERED → update*() → RENDERED (stays in rendered state)
RENDERED → destroy() → DESTROYED
```

**Implementation Example** (Canvas Renderer):
```javascript
class CanvasRenderer implements IRenderer {
  private canvas: HTMLCanvasElement;
  private ctx: CanvasRenderingContext2D;
  private state: 'uninitialized' | 'initialized' | 'rendered' | 'destroyed' = 'uninitialized';
  
  async init(container: HTMLElement, config: VisualizationConfig): Promise<void> {
    if (this.state !== 'uninitialized') {
      throw new Error(`Cannot init from state ${this.state}`);
    }
    
    this.canvas = document.createElement('canvas');
    this.canvas.width = container.clientWidth;
    this.canvas.height = container.clientHeight;
    container.appendChild(this.canvas);
    
    this.ctx = this.canvas.getContext('2d');
    this.config = config;
    
    this.state = 'initialized';
  }
  
  async render(agents: Agent[], edges: Edge[]): Promise<void> {
    if (this.state !== 'initialized' && this.state !== 'rendered') {
      throw new Error(`Cannot render from state ${this.state}`);
    }
    
    this.clear();
    
    // Render edges first (below agents)
    for (const edge of edges) {
      this.drawEdge(edge);
    }
    
    // Render agents on top
    for (const agent of agents) {
      this.drawAgent(agent);
    }
    
    this.state = 'rendered';
  }
  
  updateAgent(agent: Agent): void {
    if (this.state !== 'rendered') {
      throw new Error(`Cannot update from state ${this.state}`);
    }
    
    // Incremental update: clear agent's bounding box and redraw
    this.clearRegion(agent.bounds);
    this.drawAgent(agent);
    // Note: May need to redraw overlapping edges
  }
  
  destroy(): void {
    if (this.state === 'destroyed') {
      return;
    }
    
    this.canvas?.remove();
    this.canvas = null;
    this.ctx = null;
    this.state = 'destroyed';
  }
}
```

**WebGL Renderer Considerations**:
```javascript
class WebGLRenderer implements IRenderer {
  private gl: WebGLRenderingContext;
  private shaders: Map<string, WebGLProgram>;
  private buffers: Map<string, WebGLBuffer>;
  
  async init(container: HTMLElement, config: VisualizationConfig): Promise<void> {
    const canvas = document.createElement('canvas');
    this.gl = canvas.getContext('webgl2');
    
    if (!this.gl) {
      throw new Error('WebGL 2 not supported');
    }
    
    // Compile shaders
    await this.compileShaders();
    
    // Initialize buffers
    this.initBuffers();
    
    container.appendChild(canvas);
    this.state = 'initialized';
  }
  
  destroy(): void {
    // CRITICAL: Clean up GPU resources
    for (const buffer of this.buffers.values()) {
      this.gl.deleteBuffer(buffer);
    }
    
    for (const program of this.shaders.values()) {
      this.gl.deleteProgram(program);
    }
    
    this.gl = null;
    this.state = 'destroyed';
  }
}
```

**Renderer Switching** (e.g., when agent count changes):
```javascript
async switchRenderer(newType: 'svg' | 'canvas' | 'webgl') {
  // Destroy old renderer
  await this.currentRenderer.destroy();
  
  // Create new renderer
  this.currentRenderer = this.createRenderer(newType);
  await this.currentRenderer.init(this.container, this.config);
  
  // Re-render with new renderer
  await this.currentRenderer.render(this.agents, this.edges);
}
```

**Performance Testing** (before finalizing thresholds):
```bash
# Generate synthetic datasets
go run scripts/generate_test_topology.go --nodes=100,500,1000,5000,10000

# Benchmark renderers
npm run benchmark -- --dataset=topology-1000nodes.json --renderer=svg,canvas,webgl

# Results example:
# SVG:    1000 nodes → 12 FPS (unacceptable)
# Canvas: 1000 nodes → 45 FPS (good)
# WebGL:  1000 nodes → 58 FPS (overkill, added complexity)

# Conclusion: Canvas threshold = 5000, not 1000
```

## Basemap Configuration

**Providers**:
```json
{
  "basemap": {
    "provider": "maptiler",
    "styleUrl": "https://api.maptiler.com/maps/streets-v2/style.json",
    "apiKeyRef": "MAPTILER_API_KEY",
    "attribution": "© MapTiler © OpenStreetMap contributors",
    "maxZoom": 18,
    "minZoom": 2,
    "bounds": [[36.6, -1.5], [37.1, -1.1]]
  }
}
```

**Alternatives**:
- `"provider": "mapbox"` → Mapbox GL JS
- `"provider": "osm"` → OpenStreetMap tiles (free, rate-limited)
- `"provider": "custom"` → User-provided tile server

**Licensing**:
| Provider | License | Attribution Required | Cost |
|----------|---------|---------------------|------|
| OpenStreetMap | ODbL | Yes | Free (fair use) |
| MapTiler | Proprietary | Yes | $0-$299/mo |
| Mapbox | Proprietary | Yes | $0-$250/mo |
| Custom | User's | Depends | Self-hosted |

**API Key Management** (environment variables):
```bash
# .env file
MAPTILER_API_KEY=abc123def456
MAPBOX_ACCESS_TOKEN=pk.eyJ1...

# Config references env vars (not hardcoded)
"apiKeyRef": "MAPTILER_API_KEY"
```

**Attribution Compliance**:
```html
<!-- Automatically rendered in map corner -->
<div class="map-attribution">
  © MapTiler © OpenStreetMap contributors
  <a href="/about/data-sources">Data Sources</a>
</div>
```

### Basemap Failure Modes (Graceful Degradation)

**Scenarios**:
1. **Offline** - No network connection
2. **Invalid API Key** - 401/403 from tile server
3. **Rate Limit Exceeded** - 429 Too Many Requests
4. **Tile Server Down** - 500/503 errors

**Fallback Strategy**:
```javascript
class BasemapLoader {
  async loadBasemap(config: BasemapConfig): Promise<Basemap | null> {
    try {
      // Attempt to load configured basemap
      const map = await this.initializeMap(config);
      return map;
    } catch (err) {
      console.warn(`⚠️  Basemap failed to load: ${err.message}`);
      
      // Determine failure mode
      if (err.status === 401 || err.status === 403) {
        this.showError('Invalid basemap API key. Check environment variables.');
      } else if (err.status === 429) {
        this.showError('Basemap rate limit exceeded. Retrying in 60s...');
        setTimeout(() => this.loadBasemap(config), 60000);
      } else if (!navigator.onLine) {
        this.showError('Offline mode: Basemap unavailable');
      } else {
        this.showError('Basemap unavailable. Using fallback style.');
      }
      
      // Graceful fallback: Render without basemap
      return this.createFallbackStyle();
    }
  }
  
  createFallbackStyle(): Basemap {
    return {
      type: 'plain',
      background: '#f5f5f5',
      grid: {
        enabled: true,
        color: '#e0e0e0',
        spacing: 50
      },
      attribution: 'No basemap (offline mode)'
    };
  }
  
  showError(message: string) {
    const banner = document.createElement('div');
    banner.className = 'basemap-error-banner';
    banner.textContent = `🗺️  ${message}`;
    banner.style.cssText = `
      position: absolute;
      top: 10px;
      left: 50%;
      transform: translateX(-50%);
      background: #ff3860;
      color: white;
      padding: 8px 16px;
      border-radius: 4px;
      z-index: 1000;
    `;
    document.body.appendChild(banner);
  }
}
```

**Fallback Style** (plain background with grid):
```javascript
class FallbackRenderer {
  renderBackground(ctx: CanvasRenderingContext2D) {
    // Light gray background
    ctx.fillStyle = '#f5f5f5';
    ctx.fillRect(0, 0, this.width, this.height);
    
    // Grid overlay (helps spatial orientation)
    ctx.strokeStyle = '#e0e0e0';
    ctx.lineWidth = 1;
    
    for (let x = 0; x < this.width; x += 50) {
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, this.height);
      ctx.stroke();
    }
    
    for (let y = 0; y < this.height; y += 50) {
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(this.width, y);
      ctx.stroke();
    }
  }
}
```

**Topology Remains Usable**:
- Agents render at projected coordinates (even without tiles)
- Edges still visible
- Interactions work normally
- Only background tiles missing

---

**Navigation**: [README](README.md) | [Overview & Architecture](01-overview-architecture.md) | [API Contracts & Real-Time](02-api-realtime.md) | **[Expression Language & Rendering]** | [Security & Accessibility](04-security-accessibility.md) | [Testing & Delivery](05-testing-delivery.md)
