# MVP-015: Management Dashboard Specification

## Overview

Build a comprehensive web-based management dashboard using **Templ + HTMX + Alpine.js** stack for agent monitoring, real-time updates, and control operations. This approach provides a React-like component architecture while staying in Go, with full HTML debuggability.

## Technology Stack Rationale

### Why Templ + HTMX + Alpine.js?

1. **Full Go Integration**
   - Type-safe components compiled at build time
   - Share types and logic between backend and frontend
   - Single language across the entire stack
   - Leverages existing Go REST API (MVP-013)

2. **Superior Debugging Experience**
   - Generates real HTML (not Virtual DOM)
   - Full browser DevTools support
   - Network tab shows all HTMX requests clearly
   - No WASM debugging complexity
   - Can inspect HTML source directly

3. **Component-Based Architecture**
   - Templ components are composable Go functions
   - Alpine.js for client-side interactivity
   - HTMX for server-driven updates
   - Similar development experience to React

4. **Performance & Simplicity**
   - Server-side rendering for fast initial load
   - Minimal JavaScript bundle (Alpine.js ~15KB)
   - Progressive enhancement approach
   - No build tools required (just `templ generate`)

## Architecture

### Component Structure

```
internal/web/
├── components/           # Reusable Templ components
│   ├── agent_card.templ         # Agent display card
│   ├── agent_grid.templ         # Agent grid layout
│   ├── agent_details.templ      # Agent detail view
│   ├── agent_metrics.templ      # Metrics visualization
│   ├── agent_logs.templ         # Log viewer
│   ├── stats_card.templ         # Statistics card
│   ├── create_agent_form.templ  # Agent creation form
│   ├── config_editor.templ      # Configuration editor
│   └── layout.templ             # Base layout
├── pages/               # Full page components
│   ├── dashboard.templ          # Main dashboard
│   ├── agent_detail.templ       # Agent detail page
│   ├── agent_list.templ         # Agent list page
│   ├── pool_management.templ    # Pool management
│   └── settings.templ           # Settings page
├── handlers/            # HTTP handlers for web routes
│   ├── dashboard_handler.go     # Dashboard routes
│   ├── agent_web_handler.go     # Agent web routes
│   └── htmx_handler.go          # HTMX-specific routes
└── static/              # Static assets
    ├── css/
    │   └── styles.css           # Tailwind CSS
    ├── js/
    │   ├── alpine-components.js # Alpine.js components
    │   └── htmx-config.js       # HTMX configuration
    └── img/
        └── logo.svg

static/
└── [linked to internal/web/static/]
```

### Data Flow

```
Browser Request
    ↓
Gin Router (/dashboard)
    ↓
DashboardHandler
    ↓
Fetch data from Runtime Manager + REST API
    ↓
Render Templ Component (server-side)
    ↓
Return HTML to browser
    ↓
HTMX triggers updates (polling/events)
    ↓
Partial HTML updates (no full page reload)
```

## Core Features

### 1. Dashboard Overview
**Route**: `/dashboard`

**Components**:
- Stats overview (total agents, running, stopped, unhealthy)
- Recent activity feed
- Agent grid with status indicators
- System health metrics

**Interactions**:
- Real-time updates via HTMX polling (every 2s)
- Click agent card → navigate to detail view
- Quick actions: start, stop, restart agents
- Filter agents by status/type
- Search agents by name

**Templ Component Example**:
```go
templ Dashboard(agents []agent.Agent, stats DashboardStats) {
    @Layout("Dashboard") {
        <div x-data="dashboard()" class="container">
            <!-- Stats Cards -->
            <div class="stats-grid">
                @StatsCard("Total Agents", stats.Total, "users")
                @StatsCard("Running", stats.Running, "play-circle")
                @StatsCard("Stopped", stats.Stopped, "stop-circle")
                @StatsCard("Unhealthy", stats.Unhealthy, "alert-circle")
            </div>
            
            <!-- Agent Grid with Auto-Update -->
            <div 
                id="agents-grid"
                hx-get="/api/web/agents/live"
                hx-trigger="every 2s"
                hx-swap="innerHTML"
            >
                @AgentGrid(agents)
            </div>
        </div>
    }
}
```

### 2. Agent Card Component
**Reusable component showing agent status**

**Features**:
- Agent name, ID, type
- Status badge (running/stopped/error)
- Health indicator
- Quick action buttons
- Expandable details section

**Component Example**:
```go
templ AgentCard(a agent.Agent) {
    <div 
        class="agent-card"
        x-data="{ expanded: false }"
        data-agent-id={ a.ID }
    >
        <div class="card-header">
            <h3>{ a.Name }</h3>
            @StatusBadge(a.State)
            @HealthIndicator(a.IsHealthy())
        </div>
        
        <div class="card-body">
            <p><strong>Type:</strong> { a.Type }</p>
            <p><strong>Uptime:</strong> { formatDuration(time.Since(a.CreatedAt)) }</p>
        </div>
        
        <div class="card-actions">
            if a.State == agent.StateStopped {
                @ActionButton("Start", "/agents/" + a.ID + "/start", "success")
            } else if a.State == agent.StateRunning {
                @ActionButton("Stop", "/agents/" + a.ID + "/stop", "danger")
                @ActionButton("Restart", "/agents/" + a.ID + "/restart", "warning")
            }
            
            <button 
                @click="expanded = !expanded"
                class="btn btn-info"
            >
                <span x-show="!expanded">Details</span>
                <span x-show="expanded">Hide</span>
            </button>
        </div>
        
        <div x-show="expanded" x-transition class="card-details">
            @AgentDetails(a)
        </div>
    </div>
}

templ ActionButton(text, url, style string) {
    <button 
        hx-post={ url }
        hx-target="closest .agent-card"
        hx-swap="outerHTML"
        class={ "btn btn-" + style }
    >
        { text }
    </button>
}
```

### 3. Agent Detail Page
**Route**: `/dashboard/agents/:id`

**Features**:
- Complete agent information
- Real-time metrics charts
- Live log streaming
- Configuration viewer/editor
- Memory state viewer
- Task list

**Tabs**:
- Overview: Status, config, metadata
- Metrics: CPU, memory, task counts (Chart.js)
- Logs: Live log stream with filtering
- Memory: Agent memory state viewer
- Tasks: Active and completed tasks
- Configuration: Edit agent config

**Component Example**:
```go
templ AgentDetailPage(a agent.Agent) {
    @Layout(a.Name) {
        <div class="agent-detail" x-data="{ tab: 'overview' }">
            @TabNav([]string{"overview", "metrics", "logs", "memory", "tasks", "config"})
            
            <div x-show="tab === 'overview'">
                @AgentOverview(a)
            </div>
            
            <div x-show="tab === 'metrics'">
                @AgentMetricsChart(a.ID)
            </div>
            
            <div x-show="tab === 'logs'">
                @AgentLogViewer(a.ID)
            </div>
            
            <div x-show="tab === 'memory'">
                @AgentMemoryViewer(a.ID)
            </div>
            
            <div x-show="tab === 'tasks'">
                @AgentTaskList(a.ID)
            </div>
            
            <div x-show="tab === 'config'">
                @ConfigEditor(a)
            </div>
        </div>
    }
}
```

### 4. Real-Time Metrics Component
**Live metrics visualization with Chart.js**

**Features**:
- CPU usage over time
- Memory usage trend
- Task completion rate
- Auto-updates every 2 seconds

**Component Example**:
```go
templ AgentMetricsChart(agentID string) {
    <div 
        x-data="metricsChart()"
        x-init="init('{ agentID }')"
        class="metrics-container"
    >
        <canvas id={ "chart-" + agentID } width="800" height="400"></canvas>
        
        <!-- Auto-update metrics -->
        <div 
            hx-get={ "/api/v1/agents/" + agentID + "/metrics" }
            hx-trigger="every 2s"
            hx-swap="none"
            hx-on::after-request="window.updateChart($event.detail.xhr.response)"
        ></div>
    </div>
}
```

**Alpine.js Component** (`static/js/alpine-components.js`):
```javascript
function metricsChart() {
    return {
        chart: null,
        agentId: null,
        
        init(agentId) {
            this.agentId = agentId;
            const ctx = document.getElementById(`chart-${agentId}`);
            
            this.chart = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: 'CPU Usage (%)',
                        data: [],
                        borderColor: 'rgb(75, 192, 192)',
                        tension: 0.1
                    }, {
                        label: 'Memory (MB)',
                        data: [],
                        borderColor: 'rgb(255, 99, 132)',
                        tension: 0.1
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: true
                        }
                    }
                }
            });
            
            // Store for HTMX to access
            window.updateChart = (data) => this.updateData(JSON.parse(data));
        },
        
        updateData(metrics) {
            const now = new Date().toLocaleTimeString();
            this.chart.data.labels.push(now);
            this.chart.data.datasets[0].data.push(metrics.cpu);
            this.chart.data.datasets[1].data.push(metrics.memory);
            
            // Keep only last 20 data points
            if (this.chart.data.labels.length > 20) {
                this.chart.data.labels.shift();
                this.chart.data.datasets[0].data.shift();
                this.chart.data.datasets[1].data.shift();
            }
            
            this.chart.update('none'); // Update without animation
        }
    }
}
```

### 5. Live Log Viewer
**Real-time log streaming**

**Features**:
- Auto-scroll to latest logs
- Filter by log level
- Search logs
- Download logs

**Component Example**:
```go
templ AgentLogViewer(agentID string) {
    <div class="log-viewer" x-data="logViewer()">
        <!-- Filters -->
        <div class="log-filters">
            <select x-model="level" @change="filterLogs()">
                <option value="all">All Levels</option>
                <option value="error">Error</option>
                <option value="warn">Warning</option>
                <option value="info">Info</option>
                <option value="debug">Debug</option>
            </select>
            
            <input 
                type="text" 
                x-model="search"
                @input="filterLogs()"
                placeholder="Search logs..."
            />
            
            <label>
                <input type="checkbox" x-model="autoScroll" />
                Auto-scroll
            </label>
        </div>
        
        <!-- Logs -->
        <div 
            id="logs-container"
            class="logs"
            hx-get={ "/api/v1/agents/" + agentID + "/logs" }
            hx-trigger="every 1s"
            hx-swap="beforeend"
        >
            <!-- Log entries appended here -->
        </div>
    </div>
}
```

### 6. Create Agent Form
**Modal form for creating new agents**

**Features**:
- Agent name, type selection
- Configuration options based on type
- Template selection
- Validation
- Submit via HTMX

**Component Example**:
```go
templ CreateAgentModal() {
    <div 
        x-data="{ 
            open: false, 
            agentType: 'worker',
            formData: { name: '', type: 'worker', config: {} }
        }"
        x-show="open"
        @open-create-modal.window="open = true"
        @keydown.escape.window="open = false"
        class="modal"
    >
        <div class="modal-overlay" @click="open = false"></div>
        
        <div class="modal-content">
            <h2>Create New Agent</h2>
            
            <form 
                hx-post="/api/v1/agents"
                hx-target="#agents-grid"
                hx-swap="afterbegin"
                @htmx:afterRequest="open = false; formData = {}"
            >
                <div class="form-group">
                    <label for="name">Agent Name *</label>
                    <input 
                        type="text" 
                        id="name" 
                        name="name"
                        x-model="formData.name"
                        required
                        placeholder="my-worker-agent"
                    />
                </div>
                
                <div class="form-group">
                    <label for="type">Agent Type *</label>
                    <select 
                        id="type" 
                        name="type"
                        x-model="agentType"
                        required
                    >
                        <option value="worker">Worker Agent</option>
                        <option value="coordinator">Coordinator Agent</option>
                        <option value="monitor">Monitor Agent</option>
                    </select>
                </div>
                
                <!-- Dynamic config fields based on type -->
                <div x-show="agentType === 'worker'">
                    @WorkerConfigFields()
                </div>
                
                <div x-show="agentType === 'coordinator'">
                    @CoordinatorConfigFields()
                </div>
                
                <div x-show="agentType === 'monitor'">
                    @MonitorConfigFields()
                </div>
                
                <div class="form-actions">
                    <button type="submit" class="btn btn-primary">
                        Create Agent
                    </button>
                    <button 
                        type="button" 
                        @click="open = false" 
                        class="btn btn-secondary"
                    >
                        Cancel
                    </button>
                </div>
            </form>
        </div>
    </div>
}
```

### 7. Agent Pool Management
**Route**: `/dashboard/pools`

**Features**:
- List all agent pools
- Create new pools
- Assign agents to pools
- Pool-level operations (start all, stop all)
- Resource allocation per pool

## Backend Handlers

### Dashboard Handler
**File**: `internal/web/handlers/dashboard_handler.go`

```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/aosanya/CodeValdCortex/internal/runtime"
    "github.com/aosanya/CodeValdCortex/internal/web/pages"
    "github.com/aosanya/CodeValdCortex/internal/web/components"
)

type DashboardHandler struct {
    runtime *runtime.Manager
}

func NewDashboardHandler(runtime *runtime.Manager) *DashboardHandler {
    return &DashboardHandler{runtime: runtime}
}

// ShowDashboard renders the main dashboard page
func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
    agents := h.runtime.ListAgents()
    stats := h.calculateStats(agents)
    
    // Render Templ component
    pages.Dashboard(agents, stats).Render(c.Request.Context(), c.Writer)
}

// GetAgentsLive returns just the agents grid for HTMX updates
func (h *DashboardHandler) GetAgentsLive(c *gin.Context) {
    agents := h.runtime.ListAgents()
    
    // Return only the agent cards (partial HTML)
    c.Writer.Header().Set("Content-Type", "text/html")
    for _, agent := range agents {
        components.AgentCard(agent).Render(c.Request.Context(), c.Writer)
    }
}

// ShowAgentDetail renders the agent detail page
func (h *DashboardHandler) ShowAgentDetail(c *gin.Context) {
    agentID := c.Param("id")
    agent := h.runtime.GetAgent(agentID)
    
    if agent == nil {
        c.String(http.StatusNotFound, "Agent not found")
        return
    }
    
    pages.AgentDetailPage(*agent).Render(c.Request.Context(), c.Writer)
}

// HandleAgentAction handles start/stop/restart actions via HTMX
func (h *DashboardHandler) HandleAgentAction(c *gin.Context) {
    agentID := c.Param("id")
    action := c.Param("action")
    
    agent := h.runtime.GetAgent(agentID)
    if agent == nil {
        c.String(http.StatusNotFound, "Agent not found")
        return
    }
    
    switch action {
    case "start":
        h.runtime.StartAgent(agentID)
    case "stop":
        h.runtime.StopAgent(agentID)
    case "restart":
        h.runtime.RestartAgent(agentID)
    case "pause":
        h.runtime.PauseAgent(agentID)
    case "resume":
        h.runtime.ResumeAgent(agentID)
    default:
        c.String(http.StatusBadRequest, "Unknown action")
        return
    }
    
    // Return updated agent card
    agent = h.runtime.GetAgent(agentID)
    components.AgentCard(*agent).Render(c.Request.Context(), c.Writer)
}

func (h *DashboardHandler) calculateStats(agents []*agent.Agent) DashboardStats {
    stats := DashboardStats{}
    stats.Total = len(agents)
    
    for _, a := range agents {
        if a.State == agent.StateRunning {
            stats.Running++
        } else if a.State == agent.StateStopped {
            stats.Stopped++
        }
        
        if !a.IsHealthy() {
            stats.Unhealthy++
        }
    }
    
    return stats
}

type DashboardStats struct {
    Total     int
    Running   int
    Stopped   int
    Unhealthy int
}
```

## Routes Configuration

Update `internal/api/server.go` to add web routes:

```go
func (s *Server) setupRoutes() {
    // ... existing API routes ...
    
    // Web UI routes
    s.setupWebRoutes()
}

func (s *Server) setupWebRoutes() {
    // Serve static files
    s.router.Static("/static", "./static")
    
    // Dashboard routes
    dashboard := s.router.Group("/dashboard")
    {
        dashboard.GET("", s.dashboardHandler.ShowDashboard)
        dashboard.GET("/agents/:id", s.dashboardHandler.ShowAgentDetail)
        dashboard.GET("/pools", s.dashboardHandler.ShowPools)
    }
    
    // HTMX API routes (return HTML fragments)
    webAPI := s.router.Group("/api/web")
    {
        webAPI.GET("/agents/live", s.dashboardHandler.GetAgentsLive)
        webAPI.POST("/agents/:id/:action", s.dashboardHandler.HandleAgentAction)
    }
    
    // Root redirect to dashboard
    s.router.GET("/", func(c *gin.Context) {
        c.Redirect(http.StatusFound, "/dashboard")
    })
}
```

## Development Workflow

### Setup

1. **Install Templ CLI**:
```bash
go install github.com/a-h/templ/cmd/templ@latest
```

2. **Install Air for hot reload**:
```bash
go install github.com/cosmtrek/air@latest
```

3. **Create `.air.toml`**:
```toml
[build]
  cmd = "templ generate && go build -o ./bin/server ./cmd/main.go"
  bin = "./bin/server"
  include_ext = ["go", "templ"]
  exclude_dir = ["tmp", "vendor", "node_modules"]
  delay = 1000
```

4. **Development mode**:
```bash
# Terminal 1: Watch templates
templ generate --watch

# Terminal 2: Run with hot reload
air

# Browser: http://localhost:8080/dashboard
```

### Debugging

1. **Browser DevTools**:
   - Inspect HTML elements normally
   - Network tab shows HTMX requests
   - Console for Alpine.js debugging

2. **HTMX Debugging**:
```javascript
// In browser console
htmx.logAll(); // Enable verbose logging
```

3. **Alpine.js DevTools**:
   - Install Alpine.js DevTools extension
   - Inspect component state: `$el.__x.$data`

4. **Go Debugging**:
   - Set breakpoints in handlers
   - Use `dlv debug` or VS Code debugger
   - Inspect component rendering

## Testing Strategy

### Component Testing
```go
func TestAgentCard(t *testing.T) {
    agent := agent.Agent{
        ID:    "test-123",
        Name:  "Test Agent",
        State: agent.StateRunning,
    }
    
    var buf bytes.Buffer
    err := components.AgentCard(agent).Render(context.Background(), &buf)
    
    assert.NoError(t, err)
    html := buf.String()
    
    assert.Contains(t, html, "Test Agent")
    assert.Contains(t, html, "test-123")
    assert.Contains(t, html, "Running")
}
```

### Integration Testing
- Test HTMX endpoints return correct HTML
- Test form submissions
- Test real-time updates

### E2E Testing
- Use Playwright or Cypress
- Test full user workflows

## Dependencies

### Go Dependencies
```bash
go get github.com/a-h/templ
# (Already have Gin, ArangoDB, etc.)
```

### Frontend Assets (Self-Hosted - No CDN)

**Important**: CodeValdCortex must work in air-gapped environments. All frontend assets must be downloaded and self-hosted.

#### Required Assets

```
static/
├── css/
│   ├── tailwind.min.css          # Built with Tailwind CLI (~10KB)
│   └── styles.css                # Custom styles
├── js/
│   ├── htmx.min.js               # HTMX v1.9.10+ (~14KB)
│   ├── alpine.min.js             # Alpine.js v3.13.3+ (~15KB)
│   ├── chart.min.js              # Chart.js v4.4.1+ (~60KB)
│   └── alpine-components.js      # Custom components
└── img/
    └── logo.svg
```

#### Download Script

Create `scripts/download-assets.sh`:

```bash
#!/bin/bash
set -e

echo "Downloading frontend assets..."
mkdir -p static/{css,js,img}

# Download HTMX
curl -L https://unpkg.com/htmx.org@1.9.10/dist/htmx.min.js \
  -o static/js/htmx.min.js

# Download Alpine.js
curl -L https://unpkg.com/alpinejs@3.13.3/dist/cdn.min.js \
  -o static/js/alpine.min.js

# Download Chart.js
curl -L https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js \
  -o static/js/chart.min.js

echo "✓ All assets downloaded"
```

#### Tailwind CSS Setup

```bash
# Install Tailwind CLI
npm install -D tailwindcss

# Build CSS
npx tailwindcss -i ./input.css -o ./static/css/tailwind.min.css --minify
```

#### Asset Verification

Before deployment, verify all assets:

```bash
# Create verification script
cat > scripts/verify-assets.sh << 'EOF'
#!/bin/bash
REQUIRED=(
    "static/css/tailwind.min.css"
    "static/js/htmx.min.js"
    "static/js/alpine.min.js"
    "static/js/chart.min.js"
)

for asset in "${REQUIRED[@]}"; do
    [ -f "$asset" ] && echo "✓ $asset" || echo "✗ $asset MISSING"
done
EOF

chmod +x scripts/verify-assets.sh
./scripts/verify-assets.sh
```

## Implementation Phases

### Phase 1: Foundation (2-3 hours)
- [ ] Install Templ, setup project structure
- [ ] Download and setup static assets (run download-assets.sh)
- [ ] Build Tailwind CSS
- [ ] Create base layout component (with self-hosted assets)
- [ ] Setup static file serving
- [ ] Create dashboard handler
- [ ] Basic routing

### Phase 2: Core Components (3-4 hours)
- [ ] Agent card component
- [ ] Agent grid component
- [ ] Stats cards
- [ ] Dashboard page
- [ ] Real-time updates with HTMX

### Phase 3: Agent Details (2-3 hours)
- [ ] Agent detail page
- [ ] Tabs navigation
- [ ] Metrics chart
- [ ] Log viewer

### Phase 4: Forms & Actions (2-3 hours)
- [ ] Create agent modal
- [ ] Config editor
- [ ] Agent actions (start/stop/restart)
- [ ] Form validation

### Phase 5: Polish & Testing (2-3 hours)
- [ ] Styling with Tailwind
- [ ] Error handling
- [ ] Loading states
- [ ] Component tests
- [ ] E2E tests
- [ ] Documentation

**Total Estimated Effort**: 11-16 hours (Medium complexity)

## Success Criteria

- [ ] Dashboard displays all agents with real-time updates
- [ ] Can create, start, stop, restart agents from UI
- [ ] Agent detail page shows comprehensive information
- [ ] Metrics update in real-time
- [ ] Logs stream live
- [ ] All components are debuggable in browser DevTools
- [ ] Responsive design works on mobile/tablet
- [ ] No JavaScript errors in console
- [ ] All HTMX requests return proper HTML
- [ ] Form validation works correctly

## Future Enhancements (Post-MVP)

- WebSocket for instant updates (replace polling)
- Dark mode toggle
- User preferences persistence
- Advanced filtering and sorting
- Bulk operations
- Agent templates library
- Drag-and-drop workflow builder
- Export/import configurations
- Alert notifications
- Multi-user support with authentication

---

**Note**: This specification aligns with the existing REST API from MVP-013 and leverages the full Go stack for a unified development experience with excellent debugging capabilities.
