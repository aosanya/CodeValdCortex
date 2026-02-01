# Framework Topology Visualizer - Security & Accessibility

> **Part 4 of 5**: RBAC, row-level filtering, field masking, edge filtering, WCAG 2.2 AA compliance, i18n

**Document ID**: ARCH-VIZ-001  
**Component**: Framework Topology Visualizer  
**Version**: 1.0  
**Date**: October 24, 2025  
**Status**: Design Specification for INFRA-017

---

**Navigation**: [README](README.md) | [Overview & Architecture](01-overview-architecture.md) | [API Contracts & Real-Time](02-api-realtime.md) | [Expression Language & Rendering](03-expression-rendering.md) | **[Security & Accessibility]** | [Testing & Delivery](05-testing-delivery.md)

---

## Security Model

### Role-Based Access Control (RBAC)

**Per-Visualization Permissions**:
```json
{
  "security": {
    "minRole": "viewer",
    "allowedRoles": ["viewer", "operator", "admin"],
    "capabilities": {
      "viewer": ["read"],
      "operator": ["read", "control"],
      "admin": ["read", "control", "configure"]
    }
  }
}
```

### Row-Level Filtering

**Filter agents by ownership**:
```json
{
  "security": {
    "rowLevelFilter": "$.metadata.orgId == $user.orgId"
  }
}
```

**Multi-tenant isolation**:
```json
{
  "security": {
    "rowLevelFilter": "$.usecase_id == $user.allowed_usecases[*]"
  }
}
```

### Field-Level Masking

**PII Protection** (coarse GPS coordinates):
```json
{
  "security": {
    "mask": {
      "fields": ["metadata.location"],
      "mode": "coarse",
      "precision": 3
    }
  }
}
```

Result: `{"lat": -1.292, "lon": 36.822}` instead of `{"lat": -1.292134, "lon": 36.821956}`

**Conditional Masking**:
```json
{
  "security": {
    "mask": {
      "fields": ["metadata.driver_name"],
      "when": "$user.role != 'admin'",
      "mode": "redact",
      "replacement": "[REDACTED]"
    }
  }
}
```

### Edge Type Filtering

**Hide sensitive relationships**:
```json
{
  "security": {
    "denyEdges": ["social:messages", "hierarchical:reports_to"]
  }
}
```

Only users with `admin` role can see who reports to whom.

### Audit Logging

```json
{
  "type": "visualization_access",
  "user_id": "user_42",
  "visualization_id": "water-network-topology",
  "action": "view",
  "agents_viewed": ["PUMP-001", "SENSOR-023"],
  "timestamp": "2025-10-24T10:32:17Z",
  "ip": "41.90.X.X"
}
```

### RBAC Enforcement (Server-Side - CRITICAL)

**Problem**: Client-side filtering alone is insufficient for multi-tenant security.

**Solution**: **Server MUST enforce all security rules before sending data.**

**Backend Implementation** (`internal/api/handlers/agents.go`):
```go
func (h *AgentHandler) GetAgents(c *gin.Context) {
    user := c.MustGet("user").(*User)
    
    // Load visualization config
    vizID := c.Query("visualization_id")
    config, err := h.vizService.GetConfig(vizID)
    if err != nil {
        c.JSON(404, gin.H{"error": "visualization not found"})
        return
    }
    
    // Check user has minimum role
    if !user.HasRole(config.Security.MinRole) {
        c.JSON(403, gin.H{"error": "insufficient permissions"})
        return
    }
    
    // Build AQL query with row-level filter
    filter := h.buildRowLevelFilter(config.Security.RowLevelFilter, user)
    
    query := fmt.Sprintf(`
        FOR agent IN agents
        FILTER agent.usecase_id == @usecase_id
        FILTER %s
        RETURN agent
    `, filter)
    
    // Execute query
    agents, err := h.db.Query(query, map[string]any{
        "usecase_id": config.UseCaseID,
        "user_org": user.OrgID,
        "user_id": user.ID,
    })
    
    // Apply field-level masking
    for _, agent := range agents {
        h.applyFieldMasking(agent, config.Security.Mask, user)
    }
    
    // Filter edges by type
    allowedEdges := h.filterEdgeTypes(agents, config.Security.DenyEdges, user)
    
    c.JSON(200, gin.H{
        "agents": agents,
        "edges": allowedEdges,
    })
}

func (h *AgentHandler) buildRowLevelFilter(expr string, user *User) string {
    if expr == "" {
        return "true" // No filter
    }
    
    // Replace user context variables
    expr = strings.ReplaceAll(expr, "$user.orgId", fmt.Sprintf("\"%s\"", user.OrgID))
    expr = strings.ReplaceAll(expr, "$user.id", fmt.Sprintf("\"%s\"", user.ID))
    expr = strings.ReplaceAll(expr, "$user.role", fmt.Sprintf("\"%s\"", user.Role))
    
    // Convert JSONPath to AQL
    // "$.metadata.orgId == $user.orgId" → "agent.metadata.orgId == \"org_123\""
    aqlFilter := convertJSONPathToAQL(expr)
    
    return aqlFilter
}

func (h *AgentHandler) applyFieldMasking(agent *Agent, mask MaskConfig, user *User) {
    if mask.When != "" && !evaluateCondition(mask.When, user) {
        return // Condition not met, no masking
    }
    
    for _, field := range mask.Fields {
        switch mask.Mode {
        case "coarse":
            // Reduce GPS precision
            if location, ok := getNestedField(agent, field).(*Location); ok {
                location.Lat = roundToPrecision(location.Lat, mask.Precision)
                location.Lon = roundToPrecision(location.Lon, mask.Precision)
            }
        case "redact":
            setNestedField(agent, field, mask.Replacement)
        case "hash":
            val := getNestedField(agent, field)
            setNestedField(agent, field, hashValue(val))
        }
    }
}
```

### Edge Filtering Enforcement (Database Level)

**Problem**: `denyEdges` enforcement must happen in DB query, not client-side.

**Solution**: Filter edges in AQL query before materialization.

```go
func (h *AgentHandler) buildEdgeQuery(config *VisualizationConfig, user *User) string {
    // Base query
    query := `
        FOR agent IN agents
        FILTER agent.usecase_id == @usecase_id
    `
    
    // Apply row-level filter
    query += fmt.Sprintf("FILTER %s\n", h.buildRowLevelFilter(config.Security.RowLevelFilter, user))
    
    // Materialize edges from connection_rules
    query += `
        LET edges = (
            FOR rule IN @connection_rules
            FOR targetId IN agent.metadata[rule.match_field]
            FOR target IN agents
            FILTER target.id == targetId
    `
    
    // CRITICAL: Filter denied edge types at DB level
    if len(config.Security.DenyEdges) > 0 {
        denyPattern := strings.Join(config.Security.DenyEdges, "|")
        query += fmt.Sprintf(`
            FILTER !REGEX_TEST(rule.canonical_type, "^(%s)$")
        `, denyPattern)
    }
    
    // Check user can see target agent (row-level security on edges)
    query += fmt.Sprintf(`
            FILTER %s
            RETURN {from: agent.id, to: target.id, type: rule.canonical_type}
        )
    `, h.buildRowLevelFilter(config.Security.RowLevelFilter, user))
    
    query += `
        RETURN {agent: agent, edges: edges}
    `
    
    return query
}
```

**Test: Edge filtering enforced**:
```go
func TestAgentAPI_EdgeTypeFiltering(t *testing.T) {
    // Create agents with social connections
    createAgent("USER-001", map[string]any{
        "type": "member",
        "metadata": map[string]any{
            "messages": []string{"USER-002"},
        },
    })
    
    // Config denies social:messages edges
    config := &VisualizationConfig{
        Security: SecurityConfig{
            DenyEdges: []string{"social:messages"},
        },
    }
    saveConfig("social-network", config)
    
    // Viewer should not see message edges
    viewerToken := loginAs("viewer@example.com")
    resp := httpGet("/api/v1/agents?visualization_id=social-network", viewerToken)
    result := parseResult(resp.Body)
    
    // Agents returned, but no edges of type "social:messages"
    assert.Len(t, result.Agents, 2)
    for _, edge := range result.Edges {
        assert.NotContains(t, edge.Type, "social:messages")
    }
}
```

### Audit Events for Config Access

```go
func (h *ConfigHandler) GetConfig(c *gin.Context) {
    user := c.MustGet("user").(*User)
    vizID := c.Param("id")
    
    config, err := h.configService.Get(vizID)
    if err != nil {
        c.JSON(404, gin.H{"error": "not found"})
        return
    }
    
    // Audit log
    h.auditLog.Record(AuditEvent{
        Type: "config_read",
        Actor: user.ID,
        Resource: vizID,
        Timestamp: time.Now(),
        IP: c.ClientIP(),
        UserAgent: c.Request.UserAgent(),
    })
    
    c.JSON(200, config)
}

func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
    user := c.MustGet("user").(*User)
    vizID := c.Param("id")
    
    var newConfig VisualizationConfig
    if err := c.BindJSON(&newConfig); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    oldConfig, _ := h.configService.Get(vizID)
    
    // Update
    if err := h.configService.Update(vizID, &newConfig); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // Audit log with diff
    h.auditLog.Record(AuditEvent{
        Type: "config_changed",
        Actor: user.ID,
        Resource: vizID,
        Timestamp: time.Now(),
        Changes: computeDiff(oldConfig, &newConfig),
        IP: c.ClientIP(),
    })
    
    c.JSON(200, gin.H{"success": true})
}
```

**Integration Test** (verify server enforcement):
```go
func TestAgentAPI_RBACEnforcement(t *testing.T) {
    // User with viewer role
    viewerToken := loginAs("viewer@example.com")
    
    // Config requires operator role
    config := &VisualizationConfig{
        Security: SecurityConfig{
            MinRole: "operator",
        },
    }
    saveConfig("water-network", config)
    
    // Request should be denied
    resp := httpGet("/api/v1/agents?visualization_id=water-network", viewerToken)
    assert.Equal(t, 403, resp.StatusCode)
    assert.Contains(t, resp.Body, "insufficient permissions")
}

func TestAgentAPI_RowLevelFiltering(t *testing.T) {
    // Create agents in different orgs
    createAgent("PUMP-001", "org_a")
    createAgent("PUMP-002", "org_b")
    
    // Config filters by org
    config := &VisualizationConfig{
        Security: SecurityConfig{
            RowLevelFilter: "$.metadata.orgId == $user.orgId",
        },
    }
    saveConfig("water-network", config)
    
    // User from org_a
    tokenOrgA := loginAs("user@org-a.com") // user.orgId = "org_a"
    
    // Should only see org_a agents
    resp := httpGet("/api/v1/agents?visualization_id=water-network", tokenOrgA)
    agents := parseAgents(resp.Body)
    
    assert.Len(t, agents, 1)
    assert.Equal(t, "PUMP-001", agents[0].ID)
}

func TestAgentAPI_FieldMasking(t *testing.T) {
    // Create agent with precise location
    createAgent("PUMP-001", map[string]any{
        "metadata": map[string]any{
            "location": map[string]float64{
                "lat": -1.292134,
                "lon": 36.821956,
            },
        },
    })
    
    // Config masks location for non-admins
    config := &VisualizationConfig{
        Security: SecurityConfig{
            Mask: MaskConfig{
                Fields: []string{"metadata.location"},
                When: "$user.role != 'admin'",
                Mode: "coarse",
                Precision: 3,
            },
        },
    }
    saveConfig("water-network", config)
    
    // Viewer sees coarse location
    viewerToken := loginAs("viewer@example.com")
    resp := httpGet("/api/v1/agents?visualization_id=water-network", viewerToken)
    agents := parseAgents(resp.Body)
    
    assert.Equal(t, -1.292, agents[0].Metadata["location"].(map[string]any)["lat"])
    assert.Equal(t, 36.822, agents[0].Metadata["location"].(map[string]any)["lon"])
    
    // Admin sees full precision
    adminToken := loginAs("admin@example.com")
    resp = httpGet("/api/v1/agents?visualization_id=water-network", adminToken)
    agents = parseAgents(resp.Body)
    
    assert.Equal(t, -1.292134, agents[0].Metadata["location"].(map[string]any)["lat"])
}
```

## Accessibility (A11y) & Internationalization (i18n)

### WCAG 2.2 AA Compliance

**Requirements**:
1. **Keyboard Navigation**:
   - `Tab` → select next agent
   - `Shift+Tab` → select previous agent
   - `Enter/Space` → activate agent (show details)
   - `Arrow keys` → pan viewport
   - `+/-` → zoom in/out
   - `Esc` → clear selection

2. **Screen Reader Support**:
```html
<div role="img" aria-label="Water distribution network with 27 agents">
  <div role="button" tabindex="0" aria-label="Pump PUMP-001, status: running, pressure: 125 kPa">
    <!-- Visual pump icon -->
  </div>
</div>
```

3. **Color Contrast**:
   - All status colors: contrast ratio ≥ 4.5:1
   - Edge colors on basemap: ≥ 3:1

4. **Focus Indicators**:
```css
.agent:focus {
  outline: 3px solid #3273dc;
  outline-offset: 2px;
}
```

5. **Reduced Motion**:
```javascript
const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
if (prefersReducedMotion) {
  config.animation.enabled = false;
}
```

6. **Text Alternatives**:
   - All icons have `aria-label`
   - Complex visualizations have textual summary

### Internationalization

**Locale Resource Bundles**:
```json
// locales/en-KE.json
{
  "visualization.title": "Water Distribution Network",
  "agent.pump.label": "Pump",
  "status.running": "Running",
  "status.stopped": "Stopped",
  "units.pressure": "kPa",
  "units.flow_rate": "L/min"
}

// locales/sw-KE.json
{
  "visualization.title": "Mtandao wa Usambazaji wa Maji",
  "agent.pump.label": "Pampu",
  "status.running": "Inafanya kazi",
  "status.stopped": "Imesimama",
  "units.pressure": "kPa",
  "units.flow_rate": "L/dak"
}
```

**Date/Time Formatting**:
```javascript
const formatter = new Intl.DateTimeFormat(config.locale, {
  timeZone: config.time.timezone,
  dateStyle: 'medium',
  timeStyle: 'short'
});

formatter.format(new Date(agent.updated_at));
// en-KE: "Oct 24, 2025, 10:32 AM"
// sw-KE: "24 Okt 2025, 10:32"
```

**Number Formatting**:
```javascript
const numberFormatter = new Intl.NumberFormat(config.locale, {
  style: 'unit',
  unit: 'liter-per-minute',
  maximumFractionDigits: 1
});

numberFormatter.format(125.7);
// en-KE: "125.7 L/min"
// sw-KE: "125.7 L/dak"
```

**RTL Support** (future):
```json
{
  "locale": "ar-EG",
  "dir": "rtl"
}
```

### Accessibility on Maps (Basemap Contrast)

**Problem**: Edge colors may have poor contrast against varying basemap colors (dark tiles, satellite imagery).

**Solution**: Adaptive edge rendering with halos.

**Edge Halo Technique**:
```javascript
class CanvasRenderer {
  drawEdge(edge: Edge, style: EdgeStyle) {
    const ctx = this.ctx;
    
    // Check contrast against basemap (sample pixels under edge path)
    const needsHalo = this.detectLowContrast(edge.path, style.stroke);
    
    if (needsHalo) {
      // Draw white halo first (provides contrast on dark backgrounds)
      ctx.strokeStyle = '#ffffff';
      ctx.lineWidth = style.strokeWidth + 4;
      ctx.globalAlpha = 0.8;
      this.strokePath(edge.path);
      
      // Then draw actual edge on top
      ctx.strokeStyle = style.stroke;
      ctx.lineWidth = style.strokeWidth;
      ctx.globalAlpha = 1.0;
      this.strokePath(edge.path);
    } else {
      // Direct rendering (sufficient contrast)
      ctx.strokeStyle = style.stroke;
      ctx.lineWidth = style.strokeWidth;
      this.strokePath(edge.path);
    }
  }
  
  detectLowContrast(path: Path, color: string): boolean {
    // Sample 5 points along path
    const samples = this.samplePathPixels(path, 5);
    
    // Convert edge color to luminance
    const edgeLuminance = this.getLuminance(color);
    
    // Check contrast ratio (WCAG 2.2)
    for (const pixel of samples) {
      const bgLuminance = this.getLuminance(pixel);
      const ratio = this.contrastRatio(edgeLuminance, bgLuminance);
      
      if (ratio < 3.0) {
        return true; // Low contrast detected
      }
    }
    
    return false;
  }
  
  contrastRatio(L1: number, L2: number): number {
    const lighter = Math.max(L1, L2);
    const darker = Math.min(L1, L2);
    return (lighter + 0.05) / (darker + 0.05);
  }
  
  getLuminance(color: string): number {
    // Convert RGB to relative luminance (WCAG formula)
    const rgb = this.parseColor(color);
    const [r, g, b] = rgb.map(c => {
      c = c / 255;
      return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
    });
    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
  }
}
```

**Alternative: Adaptive Blending Mode**:
```javascript
// Use CSS blend mode for automatic contrast
ctx.globalCompositeOperation = 'difference'; // Inverts against background
ctx.strokeStyle = style.stroke;
this.strokePath(edge.path);
ctx.globalCompositeOperation = 'source-over'; // Reset
```

**Textual Topology Summary** (screen reader alternative):
```html
<div role="region" aria-label="Network topology visualization">
  <button id="toggle-text-summary" aria-label="Show text summary">
    Text Summary
  </button>
  
  <div id="text-summary" hidden aria-live="polite">
    <h2>Water Distribution Network - Text Summary</h2>
    <p>27 agents in network: 5 pumps, 15 pipes, 7 sensors</p>
    
    <h3>Agents</h3>
    <ul>
      <li>PUMP-001 (status: running, pressure: 125 kPa)</li>
      <li>SENSOR-023 (status: active, temperature: 22°C)</li>
      <!-- ... -->
    </ul>
    
    <h3>Connections</h3>
    <ul>
      <li>PUMP-001 supplies PIPE-015 (flow: 450 L/min)</li>
      <li>SENSOR-023 monitors PIPE-015</li>
      <!-- ... -->
    </ul>
    
    <h3>Active Alerts</h3>
    <ul>
      <li>SENSOR-023: Pressure above threshold (125 kPa > 120 kPa)</li>
    </ul>
  </div>
</div>

<script>
document.getElementById('toggle-text-summary').addEventListener('click', () => {
  const summary = document.getElementById('text-summary');
  summary.hidden = !summary.hidden;
});
</script>
```

**Auto-Generated from Data**:
```javascript
function generateTextSummary(agents: Agent[], edges: Edge[]): string {
  const typeCounts = countByType(agents);
  const alertAgents = agents.filter(a => a.status === 'alert');
  
  return `
    <h2>${config.visualization.title} - Text Summary</h2>
    <p>${agents.length} agents: ${formatTypeCounts(typeCounts)}</p>
    
    <h3>Agents</h3>
    <ul>
      ${agents.map(a => `<li>${formatAgent(a)}</li>`).join('')}
    </ul>
    
    <h3>Connections</h3>
    <ul>
      ${edges.map(e => `<li>${formatEdge(e)}</li>`).join('')}
    </ul>
    
    ${alertAgents.length > 0 ? `
      <h3>Active Alerts</h3>
      <ul>
        ${alertAgents.map(a => `<li>${formatAlert(a)}</li>`).join('')}
      </ul>
    ` : ''}
  `;
}
```

---

**Navigation**: [README](README.md) | [Overview & Architecture](01-overview-architecture.md) | [API Contracts & Real-Time](02-api-realtime.md) | [Expression Language & Rendering](03-expression-rendering.md) | **[Security & Accessibility]** | [Testing & Delivery](05-testing-delivery.md)
