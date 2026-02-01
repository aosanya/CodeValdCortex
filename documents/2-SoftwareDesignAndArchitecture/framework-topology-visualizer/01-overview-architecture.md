# Framework Topology Visualizer - Overview & Architecture

> **Part 1 of 5**: Overview, Motivation, Design Philosophy, Component Structure, Use Case Implementations, Implementation Plan, and Extensibility

**Document ID**: ARCH-VIZ-001  
**Component**: Framework Topology Visualizer  
**Version**: 1.0  
**Date**: October 24, 2025  
**Status**: Design Specification for INFRA-017

---

**Navigation**: [README](README.md) | **[Overview & Architecture]** | [API Contracts & Real-Time](02-api-realtime.md) | [Expression Language & Rendering](03-expression-rendering.md) | [Security & Accessibility](04-security-accessibility.md) | [Testing & Delivery](05-testing-delivery.md)

---

## Executive Summary

The Framework Topology Visualizer is a **reusable, generic component** in the CodeValdCortex framework that provides standardized visualization capabilities for any use case requiring spatial or network topology representation. This component abstracts the common patterns across multiple use cases (infrastructure networks, transportation routes, logistics networks, facility layouts, geographic tracking) into a flexible, configuration-driven visualization system.

## Motivation

Analysis of CodeValdCortex use cases reveals consistent visualization needs:

| Use Case | Visualization Type | Key Elements | Spatial Characteristics |
|----------|-------------------|--------------|-------------------------|
| **UC-INFRA-001** (Water) | Infrastructure Network | Pipes, sensors, pumps, valves | Fixed topology, directional flow |
| **UC-TRACK-001** (Safiri Salama) | Real-time Tracking Map | Vehicles, routes, stops | Moving agents, geographic paths |
| **UC-RIDE-001** (RideLink) | Live Location Map | Riders, drivers, routes | Dynamic matching, real-time positions |
| **UC-LOG-001** (Logistics) | Route & Facility Network | Trucks, warehouses, routes | Geographic + facility layout |
| **UC-WMS-001** (Warehouse) | Facility Layout | Robots, racks, docks, zones | Indoor spatial layout, grid-based |
| **UC-AGRO-001** (Mashambani) | Geographic Distribution | Owners, caretakers, animals | Rural/urban locations, connections |
| **UC-COMM-001** (DiraMoja) | Social Network Graph | Members, topics, connections | Relationship-based, non-spatial |

**Common Patterns**:
1. **Entities** (agents) with visual representations (nodes/icons)
2. **Relationships** between entities (edges/connections)
3. **Status** indicators (color-coding, animations)
4. **Real-time updates** (agent state changes)
5. **Interactivity** (click, hover, select)
6. **Layers** (different information densities)

## Design Philosophy

### Core Principles

1. **Configuration-Driven**: Use cases configure visualizer through JSON, not custom code
2. **Agent-Agnostic**: Works with any role, any use case
3. **Render-Flexible**: Supports SVG, Canvas, and WebGL backends
4. **Update-Efficient**: Optimized for real-time agent state updates
5. **Style-Customizable**: Use case-specific themes and visual languages
6. **Layout-Pluggable**: Multiple layout algorithms (geographic, force-directed, hierarchical, grid)
7. **Interaction-Extensible**: Standard interactions + custom use case behaviors

### Architecture Layers

```
┌─────────────────────────────────────────────────────────────────┐
│ Use Case Specific Configuration Layer                           │
│ - Agent type → icon mappings                                    │
│ - Status → color/style rules                                    │
│ - Layout algorithm selection                                    │
│ - Position field mapping (metadata.location, coordinates, etc)  │
│ - Connection inference rules                                    │
│ - Interaction handlers                                          │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ Framework Topology Visualizer Core                              │
│ - Agent data fetching (GET /api/v1/agents)                     │
│ - Generic entity/relationship rendering                         │
│ - Real-time update management (polling/WebSocket)               │
│ - Layout computation engine                                     │
│ - Interaction event system                                      │
│ - Viewport/zoom/pan controls                                    │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ Rendering Backend (Pluggable)                                   │
│ - SVG: Static/semi-static, high quality                        │
│ - Canvas: Dynamic, medium agent counts (< 1000)                │
│ - WebGL: High performance, large agent counts (> 1000)         │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ Framework Agent API (Existing)                                  │
│ - GET /api/v1/agents - List all agents                         │
│ - GET /api/v1/agents/{id} - Get agent details                  │
│ - GET /api/v1/agents/{id}/state - Get agent state              │
│ - GET /api/v1/agents?type=X - Filter by type                   │
└─────────────────────────────────────────────────────────────────┘
```

## Component Structure

### 1. Core Framework Module

**Location**: `/internal/web/visualization/` (framework)

**Files**:
```
/internal/web/visualization/
├── config.go                # Visualization configuration loading
├── layout_algorithms.go     # Layout computation engines
└── handlers.go              # Serve visualization HTML pages

/static/js/visualization/
├── topology-visualizer.js   # Main JavaScript component
├── agent-data-source.js     # Fetch data from /api/v1/agents
├── renderers/
│   ├── svg-renderer.js      # SVG rendering engine
│   ├── canvas-renderer.js   # Canvas rendering engine
│   └── webgl-renderer.js    # WebGL rendering engine (future)
├── layouts/
│   ├── geographic.js        # Geographic coordinate layout
│   ├── force-directed.js    # Physics-based network layout
│   ├── hierarchical.js      # Tree/hierarchy layout
│   └── grid.js              # Grid-based layout (warehouse)
├── interactions/
│   ├── pan-zoom.js          # Viewport controls
│   ├── selection.js         # Entity selection
│   └── tooltips.js          # Hover information
└── themes/
    ├── default.js           # Default visual theme
    └── theme-loader.js      # Dynamic theme loading

/internal/web/templates/
└── visualization.templ      # Visualization page template
```

### 2. Configuration Schema

**Visualization Config** (JSON):
```json
{
  "$schema": "https://codevaldcortex.io/schemas/visualization/v1.0.0.json",
  "schemaVersion": "1.0.0",
  "visualization": {
    "id": "water-network-topology",
    "title": "Water Distribution Network",
    "type": "network",
    "locale": "en-KE",
    "crs": {
      "geographic": "EPSG:4326",
      "indoor": {
        "type": "local-xy",
        "origin": {"x": 0, "y": 0},
        "unit": "m",
        "orientation": "cartesian"
      }
    },
    "units": {
      "distance": "m",
      "pressure": "kPa",
      "flow_rate": "L/min",
      "temperature": "C"
    },
    "time": {
      "timezone": "Africa/Nairobi"
    },
    "taxonomyVersion": "2025.10",
    "expressions": {
      "dialect": "jsonpath"
    },
    "renderer": {
      "preferred": "auto",
      "thresholds": {
        "svgMaxNodes": 300,
        "canvasMaxNodes": 5000
      }
    },
    
    "dataSource": {
      "type": "agents",
      "endpoint": "/api/v1/agents",
      "filter": {
        "types": ["pipe", "sensor", "pump", "valve"],
        "exclude_status": ["deleted", "archived"]
      },
      "polling": {
        "enabled": true,
        "interval": 5000
      }
    },
    
    "mapping": {
      "position": {
        "source": "metadata.location",
        "type": "geographic",
        "fallback": "metadata.coordinates"
      },
      "status": {
        "source": "status"
      },
      "label": {
        "source": "name"
      }
    },
    
    "connections": {
      "strategy": "metadata",
      "source": "metadata.connected_to",
      "bidirectional": false
    },
    
    "layout": {
      "algorithm": "geographic",
      "options": {
        "center": [36.8219, -1.2921],
        "zoom": 12,
        "projection": "mercator"
      }
    },
    
    "entities": {
      "pipe": {
        "icon": "line",
        "style": {
          "stroke": "#3273dc",
          "strokeWidth": 3,
          "opacity": 0.8
        },
        "statusColors": {
          "operational": "#48c774",
          "degraded": "#ffdd57",
          "failed": "#f14668"
        },
        "label": {
          "show": false,
          "field": "id"
        }
      },
      "sensor": {
        "icon": "circle",
        "size": 8,
        "style": {
          "fill": "#209cee",
          "stroke": "#ffffff",
          "strokeWidth": 2
        },
        "statusColors": {
          "active": "#48c774",
          "inactive": "#b5b5b5",
          "error": "#f14668"
        },
        "label": {
          "show": true,
          "field": "name",
          "position": "top"
        },
        "tooltip": {
          "fields": ["metadata.pressure", "metadata.flow_rate", "metadata.temperature"]
        }
      },
      "pump": {
        "icon": "square",
        "size": 12,
        "style": {
          "fill": "#ff3860",
          "stroke": "#ffffff"
        },
        "animation": {
          "when": "status == 'running'",
          "type": "pulse",
          "duration": 2000
        },
        "tooltip": {
          "fields": ["metadata.capacity", "metadata.efficiency", "metadata.uptime"]
        }
      }
    },
    
    "layers": [
      {
        "id": "basemap",
        "type": "tile",
        "source": "osm",
        "opacity": 0.6
      },
      {
        "id": "network",
        "type": "agents",
        "filter": "type in ['pipe', 'sensor', 'pump', 'valve']"
      },
      {
        "id": "alerts",
        "type": "agents",
        "filter": "status == 'alert'",
        "zIndex": 100
      }
    ],
    
    "interactions": {
      "pan": true,
      "zoom": true,
      "select": "single",
      "hover": true,
      "contextMenu": true,
      "onClick": {
        "action": "showDetails",
        "fetchState": true
      }
    },
    
    "controls": {
      "search": true,
      "filters": true,
      "layerToggle": true,
      "legend": true,
      "refresh": true
    }
  }
}
```

### 3. Data Source Strategy - Agent-Based

**Core Principle**: Visualizer fetches data **directly from agents** using existing framework APIs, not through specialized visualization endpoints.

#### Existing Framework APIs (Already Implemented)

**List All Agents**:
```
GET /api/v1/agents
```

Response:
```json
{
  "agents": [
    {
      "id": "PUMP-001",
      "type": "pump",
      "name": "Main Pump Station",
      "status": "running",
      "created_at": "2025-10-23T10:00:00Z",
      "updated_at": "2025-10-24T10:30:45Z",
      "metadata": {
        "capacity": 5000,
        "efficiency": 92.3,
        "uptime": 168,
        "location": {"lat": -1.2921, "lon": 36.8219}
      }
    }
  ]
}
```

**Get Single Agent**:
```
GET /api/v1/agents/{agent_id}
```

**Get Agent State** (from memory service):
```
GET /api/v1/agents/{agent_id}/state
```

Response:
```json
{
  "agent_id": "SENSOR-008",
  "state": {
    "pressure": 85.3,
    "flow_rate": 1250,
    "temperature": 22.5,
    "last_reading": "2025-10-24T10:30:45Z"
  }
}
```

**List Agents by Type**:
```
GET /api/v1/agents?type=sensor
GET /api/v1/agents?type=pump&status=running
```

#### Visualization Configuration Maps Agent Data

The visualization config tells the visualizer **how to interpret agent data**:

```json
{
  "visualization": {
    "dataSource": {
      "type": "agents",
      "endpoint": "/api/v1/agents",
      "filter": {
        "types": ["pipe", "sensor", "pump", "valve"],
        "status": ["running", "operational", "active"]
      },
      "polling": {
        "enabled": true,
        "interval": 5000
      }
    },
    "mapping": {
      "position": {
        "source": "metadata.location",
        "type": "geographic"
      },
      "status": {
        "source": "status",
        "field": "status"
      },
      "label": {
        "source": "name"
      }
    },
    "connections": {
      "source": "relationships",
      "types": ["supplies", "monitors", "controls"]
    }
  }
}
```

#### Connection Inference Strategies

Since agents may not explicitly store connections, the visualizer can infer them:

**Strategy 1: From Agent Metadata**
```json
// Agent metadata includes connected agent IDs
{
  "id": "PUMP-001",
  "metadata": {
    "connected_to": ["PIPE-015", "PIPE-016"],
    "monitored_by": ["SENSOR-003"]
  }
}
```

**Strategy 2: From Role Configuration**
```json
// Standardized graph-theory based agent-type connection rules
{
  "agent_type": "pump",
  "connection_rules": {
    "supplies": {
      "target_types": ["pipe"],
      "match": "metadata.downstream_pipes",
      "directed": true,
      "weight": "metadata.flow_capacity",        // optional numeric field or expression
      "label": "supplies"
    }
  }
}
```

### Standard: Agent-Type Connection Rules → Graph Model

To make topology inference robust and reusable across use cases we standardize `connection_rules` with an explicit mapping to graph concepts:

- Node: an agent instance. Every agent document returned by `/api/v1/agents` is a graph node.
- Edge: a connection relationship between two nodes inferred from agent metadata, agent-type rules, message history, or explicit topology documents.
- Directed: boolean flag indicating whether the edge is directed (true) or undirected (false).
- Weight: optional numeric attribute to represent capacity, strength, distance, or other edge metric.
- Label/Type: string describing semantic relation (e.g., "supplies", "monitors", "controls").

This standard enables the visualizer and downstream analytic components to treat all topologies as graphs G = (V, E) and to run graph algorithms (centrality, shortest path, clustering) consistently.

### Formal JSON Schema (excerpt)

```json
{
  "connection_rules": {
    "<relation_name>": {
      "target_types": ["<agent_type>", ...],      // types this rule connects to
      "match": "<dotted.path.or.expression>",    // path in agent metadata or expression to list target agent IDs
      "directed": true,                            // default: true
      "weight": "<dotted.path.or-expression>",   // optional numeric field to use as weight
      "label": "<relation_label>",
      "multiplicity": "one-to-many|many-to-one|many-to-many", // optional
      "default_weight": 1.0                        // fallback weight
    }
  }
}
```

*(Continued with full canonical relationship taxonomy, pseudocode examples, and AQL materialization strategies...)*

For the complete connection rules specification, see lines 192-535 of the original document.

#### Visualization Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Visualizer Initialization                                │
│    - Load visualization config                              │
│    - Fetch all agents: GET /api/v1/agents                  │
│    - Parse agent metadata for positions                     │
│    - Infer connections from agent relationships             │
│    - Render initial topology                                │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Real-Time Updates (Polling)                              │
│    - Every N seconds: GET /api/v1/agents                   │
│    - Compare with cached state                              │
│    - Identify changes (status, position, metadata)          │
│    - Apply differential updates to visualization            │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. User Interaction                                         │
│    - Click entity: GET /api/v1/agents/{id}/state           │
│    - Show detailed agent state in popup                     │
│    - Display recent messages/events for agent               │
└─────────────────────────────────────────────────────────────┘
```

#### Benefits of Agent-Based Approach

1. **No Duplicate Data**: Single source of truth (agents themselves)
2. **No Custom Endpoints**: Reuse existing REST API
3. **Consistent Data Model**: Same data structure everywhere
4. **Framework Agnostic**: Works with any role
5. **State Coherence**: Agent state = visualization state
6. **Simplified Testing**: Test against existing agent API

## Use Case Implementations

### UC-INFRA-001: Water Distribution Network

**Config Location**: `usecases/UC-INFRA-001-water-distribution-network/config/visualization.json`

**Specific Customizations**:
- **Layout**: Geographic (pipes follow street layout)
- **Entity Icons**: Infrastructure-specific (pipe segments, sensors, pumps)
- **Directionality**: Show water flow direction
- **Status Colors**: Operational health (green/yellow/red)
- **Interactions**: Click pipe → show flow data; Click sensor → show readings
- **Layers**: Basemap + Network + Alerts + Maintenance zones

### UC-TRACK-001: Safiri Salama (Vehicle Tracking)

**Config Location**: `usecases/UC-TRACK-001-safiri-salama/config/visualization.json`

**Specific Customizations**:
- **Layout**: Geographic (real map with routes)
- **Entity Icons**: Vehicle icons (bus, matatu), stop markers
- **Movement**: Animate vehicle position changes
- **Trails**: Show recent path history
- **Status**: Vehicle states (on-route, delayed, stopped)
- **Interactions**: Click vehicle → passenger count, ETA; Click stop → waiting passengers
- **Layers**: Map + Routes + Vehicles + Stops + Traffic

### UC-WMS-001: Warehouse Management

**Config Location**: `usecases/UC-WMS-001-warehouse-management/config/visualization.json`

**Specific Customizations**:
- **Layout**: Grid (warehouse floor plan)
- **Entity Icons**: Robot icons, rack grids, dock bays
- **Movement**: Animate robot navigation
- **Occupancy**: Color racks by occupancy level
- **Status**: Robot states (idle, moving, picking)
- **Interactions**: Click robot → current task; Click rack → inventory
- **Layers**: Floor plan + Robots + Racks + Docks + Zones

### UC-COMM-001: DiraMoja (Social Network)

**Config Location**: `usecases/UC-COMM-001-diramoja/config/visualization.json`

**Specific Customizations**:
- **Layout**: Force-directed (relationship network)
- **Entity Icons**: User avatars, topic nodes
- **Connections**: Relationship strength (edge thickness)
- **Status**: Engagement level (node size)
- **Clustering**: Group by topics or location
- **Interactions**: Click member → profile; Click topic → discussions
- **Layers**: Members + Topics + Connections + Communities

## Implementation Plan for INFRA-017

### Phase 1: Configuration & Setup (Days 1-2)

**Tasks**:
1. Create `/internal/web/visualization/` module structure
2. Define visualization configuration Go types
3. Create configuration loader (reads JSON from use case config dir)
4. Add visualization page route to web server
5. Create basic HTML template for visualization page

**Deliverables**:
- ✅ `internal/web/visualization/config.go` - Config types and loader
- ✅ `internal/web/visualization/handlers.go` - HTTP handlers
- ✅ `internal/web/templates/visualization.templ` - Page template
- ✅ Route: `/visualization/{config_name}`
- ✅ Unit tests for config loading

### Phase 2: JavaScript Core & Agent Data Fetching (Days 3-5)

**Tasks**:
1. Create `/static/js/visualization/` directory structure
2. Implement `AgentDataSource` class
   - Fetch agents from `/api/v1/agents`
   - Apply filters from config
   - Map agent data to visualization entities
   - Handle polling for updates
3. Create `TopologyVisualizer` main class
4. Implement basic entity rendering
5. Add pan/zoom controls

**Deliverables**:
- ✅ `agent-data-source.js` - Fetches from existing agent API
- ✅ `topology-visualizer.js` - Main visualizer class
- ✅ Canvas renderer with entity rendering
- ✅ Pan/zoom controller
- ✅ Integration with agent API (no new endpoints needed)

### Phase 3: Layout Algorithms (Days 6-7)

**Tasks**:
1. Implement geographic layout (lat/lon → screen coordinates)
2. Add connection inference from agent metadata
3. Implement entity positioning based on config mapping
4. Add connection rendering (lines between agents)
5. Handle missing position data gracefully

**Deliverables**:
- ✅ `layouts/geographic.js` - Mercator projection
- ✅ Connection inference from `metadata.connected_to`
- ✅ Fallback positioning strategies
- ✅ Connection rendering with directionality

### Phase 4: UC-INFRA-001 Implementation (Days 8-10)

**Tasks**:
1. Create `visualization.json` for water network
2. Add `location` field to agent instances metadata:
   ```json
   {
     "metadata": {
       "location": {"lat": -1.2921, "lon": 36.8219},
       "connected_to": ["PIPE-015", "PIPE-016"]
     }
   }
   ```
3. Define pipe/sensor/pump/valve visual styles
4. Add status color mapping
5. Implement interaction handlers (click → agent details)
6. Add new navigation item to dashboard

**Deliverables**:
- ✅ `usecases/UC-INFRA-001-*/config/visualization.json`
- ✅ Agent metadata updated with locations
- ✅ New dashboard page: "Network Topology"
- ✅ Real-time agent status updates (polling)
- ✅ Interactive agent selection with details panel

### Phase 5: Polish & Documentation (Days 11-12)

**Tasks**:
1. Add loading states and error handling
2. Implement search/filter controls
3. Add legend showing roles
4. Performance testing with 27 agents
5. Write framework documentation
6. Create use case configuration guide
7. Record demo video
8. Write coding session document

**Deliverables**:
- ✅ Loading spinners and error messages
- ✅ Search bar to find agents by ID/name
- ✅ Legend with role colors
- ✅ Performance benchmarks (< 100ms render)
- ✅ Framework documentation
- ✅ Configuration guide for other use cases
- ✅ Demo video
- ✅ Coding session document

**Key Simplification**: No new backend endpoints needed! Visualizer fetches directly from `/api/v1/agents`.

## Extensibility for Future Use Cases

### Adding a New Use Case Visualization

**Step 1**: Create visualization config
```bash
# Create use case visualization config
cat > usecases/UC-XXX-name/config/visualization.json << 'EOF'
{
  "visualization": {
    "id": "my-use-case",
    "title": "My Use Case Topology",
    "type": "network",
    "renderer": "canvas",
    "layout": {
      "algorithm": "geographic"
    },
    "entities": {
      "my_agent_type": {
        "icon": "circle",
        "size": 10,
        "statusColors": {
          "active": "#48c774",
          "inactive": "#b5b5b5"
        }
      }
    }
  }
}
EOF
```

**Step 2**: Register visualization with framework
```go
// In use case initialization
visualizationConfig := loadVisualizationConfig("./config/visualization.json")
app.RegisterVisualization(visualizationConfig)
```

**Step 3**: Access via dashboard
```
http://localhost:8083/visualization/my-use-case
```

**No custom code needed** - everything is configuration-driven!

### Custom Layout Algorithms

If a use case needs a unique layout:

1. Implement `LayoutAlgorithm` interface:
```javascript
class CustomLayout {
  computePositions(entities, connections, options) {
    // Return {entityId: {x, y}}
  }
}
```

2. Register with visualizer:
```javascript
TopologyVisualizer.registerLayout('custom', CustomLayout);
```

3. Use in config:
```json
"layout": {
  "algorithm": "custom",
  "options": {...}
}
```

## Performance Considerations

| Agent Count | Recommended Renderer | Update Strategy | Expected FPS |
|-------------|---------------------|-----------------|--------------|
| < 100 | SVG | Full re-render | 60 |
| 100-1000 | Canvas | Differential updates | 30-60 |
| 1000-10000 | Canvas + LOD | Spatial indexing | 20-30 |
| > 10000 | WebGL | GPU acceleration | 30-60 |

**Optimization Techniques**:
- **Spatial Indexing**: Only render entities in viewport
- **Level of Detail (LOD)**: Simplify distant entities
- **Batch Rendering**: Group similar entities
- **Differential Updates**: Only redraw changed entities
- **Viewport Culling**: Skip offscreen entities
- **Animation Throttling**: Limit animation frame rate

---

**Navigation**: [README](README.md) | **[Overview & Architecture]** | [API Contracts & Real-Time](02-api-realtime.md) | [Expression Language & Rendering](03-expression-rendering.md) | [Security & Accessibility](04-security-accessibility.md) | [Testing & Delivery](05-testing-delivery.md)
