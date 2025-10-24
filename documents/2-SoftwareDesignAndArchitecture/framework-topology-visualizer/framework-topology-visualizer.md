# Framework Topology Visualizer - Standardized Component Design

**Document ID**: ARCH-VIZ-001  
**Component**: Framework Topology Visualizer  
**Version**: 1.0  
**Date**: October 24, 2025  
**Status**: Design Specification for INFRA-017

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
2. **Agent-Agnostic**: Works with any agent type, any use case
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

**Strategy 2: From Agent Type Configuration**
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

### Example (pump → pipe)

Agent type definition (pump):

```json
{
  "agent_type": "pump",
  "connection_rules": {
    "supplies": {
      "target_types": ["pipe"],
      "match": "metadata.downstream_pipes",   
      "directed": true,
      "weight": "metadata.flow_capacity",
      "label": "supplies"
    }
  }
}
```

Agent document example (instance):

```json
{
  "id": "PUMP-001",
  "type": "pump",
  "metadata": {
    "location": {"lat": -1.2921, "lon": 36.8219},
    "downstream_pipes": ["PIPE-015", "PIPE-016"],
    "flow_capacity": 1500
  }
}
```

Rendered edges for visualization: two directed edges

- {from: "PUMP-001", to: "PIPE-015", type: "supplies", weight: 1500}
- {from: "PUMP-001", to: "PIPE-016", type: "supplies", weight: 1500}

### Pseudocode: Build Graph from Agent-Type Rules (Go-like)

```go
// loadAgentTypeRules loads connection rules for each agent type
rules := LoadAgentTypeRules()

// fetch all agents once
agents := FetchAgents()
agentMap := map[string]Agent{}
for _, a := range agents { agentMap[a.ID] = a }

for _, a := range agents {
  t := a.Type
  typeRules := rules[t]
  for relName, rule := range typeRules {
    // evaluate match expression (e.g., metadata.downstream_pipes)
    targets := EvaluateMatchExpression(a, rule.Match)
    for _, targetID := range targets {
      if target, ok := agentMap[targetID]; ok {
        edge := Edge{
          From: a.ID,
          To: target.ID,
          Type: relName,
          Directed: rule.Directed,
          Weight: EvaluateWeight(a, rule.Weight, rule.DefaultWeight),
        }
        graph.AddEdge(edge)
      }
    }
  }
}
```

Notes:
- `EvaluateMatchExpression` resolves dotted paths (like `metadata.downstream_pipes`) and supports simple expressions.
- The visualizer should cache `agentMap` for fast resolution and apply differential updates on changes.

### AQL Example: Build Edges Server-side (ArangoDB)

If you prefer to materialize edges in the database, the following AQL snippet creates edges in a `topology_edges` collection by expanding `metadata.downstream_pipes`:

```aql
FOR a IN agents
  FILTER a.type == 'pump' AND HAS(a, 'metadata') AND HAS(a.metadata, 'downstream_pipes')
  FOR pid IN a.metadata.downstream_pipes
    LET p = DOCUMENT(CONCAT('agents/', pid))
    FILTER p != null
    INSERT {
      _from: a._id,
      _to: p._id,
      type: 'supplies',
      weight: a.metadata.flow_capacity || 1
    } INTO topology_edges
```

Materialized edges make some graph algorithms and queries more efficient and keep the visualizer logic simpler (visualizer reads `topology_edges` instead of inferring at runtime). However, prefer not to duplicate authoritative agent state — keep edges as derived, rebuildable structures.

### Handling Complex Matches & Rules

- Support dotted path arrays (lists of IDs), single ID fields, or relation objects (e.g., `{id: "PIPE-015", port: 1}`).
- Allow `match` to be a small mapping expression: e.g. `metadata.connections.outgoing[*].id` to support nested arrays.
- Provide a fallback order: explicit agent.metadata → agent-type rule match → message-history inference → explicit topology document collection.

### Edge Attributes & Visualizer Mapping

The visualizer expects edges with the following properties:
- `from` (agent ID)
- `to` (agent ID)
- `type` (relation label)
- `directed` (boolean)
- `weight` (number)
- `metadata` (object) optional for additional display

Map these to visualization config `connections` rules so styles (color, thickness, arrowheads) can be driven by `type` and `weight`.

### Registering New Rules (Use Case Workflow)

1. Add `connection_rules` to the use case's agent type JSON (e.g., `usecases/UC-INFRA-001-water-distribution-network/config/agent_types/pump.json`).
2. Ensure agent instances include the metadata referenced by `match` (e.g., `downstream_pipes`).
3. Optionally run a one-off AQL job to materialize `topology_edges` for performance.
4. Configure the visualization `connections` source to `agents` (inference) or `topology_edges` (materialized).

By standardizing `connection_rules` in agent-type definitions, multiple use cases can reuse the same topology visualizer without writing custom mapping code.

### Canonical Relationship Types (Standard Taxonomy)

While relation names like `"supplies"` are domain-specific, we provide a **canonical taxonomy** of relationship types that use cases should map their domain relations to. This enables the visualizer to apply consistent styling, layout hints, and graph algorithms across all use cases.

#### Standard Relationship Categories

| Category | Canonical Types | Directed? | Semantics | Visual Style | Use Cases |
|----------|----------------|-----------|-----------|--------------|-----------|
| **Spatial** | `connects_to` | No | Physical connection (bidirectional) | Solid line, no arrow | Infrastructure, Network |
| | `adjacent_to` | No | Physical proximity (symmetric) | Thin solid line | Warehouse, Facility |
| | `contains` | Yes | Containment/enclosure (parent→child) | Dashed arrow | Geographic, Zones |
| | `located_in` | Yes | Positioning (child→parent) | Dashed arrow | Geographic, Tracking |
| **Functional** | `supplies` | Yes | Resource flow (source→target) | Thick arrow, flow animation | Water, Power, Supply Chain |
| | `consumes` | Yes | Resource consumption (consumer→source) | Arrow, reverse flow | Manufacturing, Energy |
| | `produces` | Yes | Output generation (producer→product) | Arrow | Manufacturing, Processing |
| | `transforms` | Yes | Conversion (input→output) | Arrow | Chemical, Data Processing |
| | `controls` | Yes | Control authority (controller→controlled) | Bold arrow | Automation, Robotics |
| | `monitors` | No | Observation (bidirectional awareness) | Dashed line | Sensing, Supervision |
| **Communication** | `sends_to` | Yes | Message transmission (sender→receiver) | Dashed arrow, pulse | Messaging, Events |
| | `receives_from` | Yes | Message reception (receiver→sender) | Dashed arrow | Messaging, Events |
| | `publishes_to` | Yes | Broadcast (publisher→topic) | Dashed arrow | Pub/Sub, Broadcast |
| | `subscribes_to` | Yes | Subscription (subscriber→topic) | Dashed arrow | Pub/Sub, Feed |
| **Hierarchical** | `reports_to` | Yes | Organizational reporting (subordinate→superior) | Arrow, tree edge | Organizations, Command |
| | `manages` | Yes | Management (manager→managed) | Bold arrow | Organizations, Resources |
| | `owns` | Yes | Ownership (owner→owned) | Arrow | Assets, Resources |
| | `member_of` | Yes | Membership (member→group) | Arrow | Groups, Communities |
| | `parent_of` | Yes | Parent-child (parent→child) | Tree edge | Hierarchies, Genealogy |
| **Temporal** | `precedes` | Yes | Sequence (earlier→later) | Curved arrow, timeline | Workflows, Scheduling |
| | `follows` | Yes | Succession (follower→followed) | Curved arrow | Sequences, Chains |
| | `triggers` | Yes | Causation (trigger→triggered) | Bold arrow | Events, Automation |
| | `depends_on` | Yes | Dependency (dependent→dependency) | Dashed arrow | Tasks, Prerequisites |
| **Social** | `follows` | Yes | Social following (follower→followed) | Thin arrow | Social Networks |
| | `trusts` | Yes* | Trust relationship (truster→trusted) | Weighted line | Networks, Security |
| | `collaborates_with` | No | Partnership (mutual) | Thick line | Teams, Projects |
| | `competes_with` | No | Competition (symmetric) | Zigzag line | Markets, Games |

**Note**: `*` = Can be bidirectional with different weights (A trusts B ≠ B trusts A)

#### Directionality Principles

**Undirected (Symmetric) Relationships**:
- Imply bidirectional or mutual relationships
- Physical connections: `connects_to`, `adjacent_to` (if A connects to B, then B connects to A)
- Social mutuality: `collaborates_with`, `competes_with` (mutual by nature)
- Observation: `monitors` (monitoring relationship is often recorded from both perspectives)
- **Visual**: Rendered as lines without arrowheads
- **Graph**: Single edge in undirected graph or two edges in directed graph with equal weights

**Directed (Asymmetric) Relationships**:
- Imply one-way flow, hierarchy, or asymmetric relationship
- Resource flows: `supplies`, `consumes` (direction of resource movement)
- Containment: `contains`, `located_in` (parent-child asymmetry)
- Authority: `controls`, `manages`, `owns` (power direction)
- Sequence: `precedes`, `follows`, `triggers` (temporal arrow)
- Social: `follows`, `trusts` (A follows B doesn't mean B follows A)
- **Visual**: Rendered with arrowheads indicating direction
- **Graph**: Single directed edge

**Bidirectional (Mutual but Asymmetric)**:
- Some relationships can exist in both directions with different weights
- Example: `trusts` - A trusts B (weight: 0.8) and B trusts A (weight: 0.3)
- Example: `sends_to` - A sends to B and B sends to A (different message volumes)
- **Implementation**: Create two separate directed edges with potentially different weights
- **Visual**: Two arrows or single line with bidirectional arrowheads (if weights are equal)

**Use Case-Specific Override**:
```json
{
  "connection_rules": {
    "monitors": {
      "canonical_type": "functional:monitors",
      "directed": false,           // Standard: undirected
      "bidirectional": true         // Alternative: if you want explicit bidirectionality
    },
    "pipe_connection": {
      "canonical_type": "spatial:connects_to",
      "directed": true,             // Override: make directed for water flow
      "label": "water flows to"
    }
  }
}
```

**Example: Spatial Relationships in Practice**

**UC-INFRA-001 (Water Network)**:
```json
{
  "spatial_connection": {
    "canonical_type": "spatial:connects_to",
    "directed": false,              // Physical pipes are bidirectional structures
    "label": "connected"
  },
  "water_flow": {
    "canonical_type": "functional:supplies",
    "directed": true,               // Water flow has direction (pump → pipe)
    "label": "supplies water"
  }
}
```

**UC-WMS-001 (Warehouse)**:
```json
{
  "aisle_adjacency": {
    "canonical_type": "spatial:adjacent_to",
    "directed": false,              // Racks next to each other (symmetric)
    "label": "next to"
  },
  "zone_containment": {
    "canonical_type": "spatial:contains",
    "directed": true,               // Zone contains racks (parent → children)
    "label": "contains"
  }
}
```

**UC-TRACK-001 (Vehicle Tracking)**:
```json
{
  "route_sequence": {
    "canonical_type": "temporal:precedes",
    "directed": true,               // Stop A precedes Stop B (ordered sequence)
    "label": "before"
  },
  "current_location": {
    "canonical_type": "spatial:located_in",
    "directed": true,               // Vehicle is located at Stop (child → parent)
    "label": "currently at"
  }
}
```

#### Domain Mapping Examples

**UC-INFRA-001 (Water Distribution)**:
```json
{
  "connection_rules": {
    "supplies": {
      "canonical_type": "functional:supplies",
      "target_types": ["pipe"],
      "match": "metadata.downstream_pipes",
      "directed": true,
      "weight": "metadata.flow_capacity",
      "label": "supplies water"
    },
    "monitors": {
      "canonical_type": "functional:monitors",
      "target_types": ["pipe", "pump"],
      "match": "metadata.monitored_entities",
      "directed": false,
      "label": "monitors"
    },
    "connects_to": {
      "canonical_type": "spatial:connects_to",
      "target_types": ["pipe"],
      "match": "metadata.connected_pipes",
      "directed": false,
      "label": "connected"
    }
  }
}
```

**UC-TRACK-001 (Safiri Salama - Vehicle Tracking)**:
```json
{
  "agent_type": "vehicle",
  "connection_rules": {
    "follows_route": {
      "canonical_type": "temporal:follows",
      "target_types": ["route", "stop"],
      "match": "metadata.route_stops",
      "directed": true,
      "weight": "metadata.stop_sequence",
      "label": "route"
    },
    "currently_at": {
      "canonical_type": "spatial:located_in",
      "target_types": ["stop"],
      "match": "state.current_stop",
      "directed": false,
      "label": "at stop"
    }
  }
}
```

**UC-WMS-001 (Warehouse Management)**:
```json
{
  "agent_type": "robot",
  "connection_rules": {
    "transporting_to": {
      "canonical_type": "functional:supplies",
      "target_types": ["rack", "dock"],
      "match": "state.destination",
      "directed": true,
      "label": "delivering to"
    },
    "adjacent_to": {
      "canonical_type": "spatial:adjacent_to",
      "target_types": ["robot", "rack"],
      "match": "metadata.nearby_entities",
      "directed": false,
      "weight": "metadata.distance",
      "label": "near"
    }
  }
}
```

**UC-COMM-001 (DiraMoja - Social Network)**:
```json
{
  "agent_type": "member",
  "connection_rules": {
    "follows": {
      "canonical_type": "social:follows",
      "target_types": ["member", "topic"],
      "match": "metadata.following",
      "directed": true,
      "weight": "metadata.engagement_score",
      "label": "follows"
    },
    "member_of": {
      "canonical_type": "hierarchical:member_of",
      "target_types": ["group"],
      "match": "metadata.groups",
      "directed": false,
      "label": "member"
    }
  }
}
```

**UC-LOG-001 (Smart Logistics)**:
```json
{
  "agent_type": "truck",
  "connection_rules": {
    "delivers_to": {
      "canonical_type": "functional:supplies",
      "target_types": ["facility"],
      "match": "state.delivery_stops",
      "directed": true,
      "weight": "metadata.cargo_weight",
      "label": "delivers"
    },
    "precedes": {
      "canonical_type": "temporal:precedes",
      "target_types": ["truck"],
      "match": "state.next_in_sequence",
      "directed": true,
      "label": "scheduled before"
    }
  }
}
```

#### Visualization Style Mapping by Canonical Type

The visualizer applies consistent styles based on `canonical_type`:

```json
{
  "canonical_styles": {
    "functional:supplies": {
      "stroke": "#3273dc",
      "strokeWidth": 3,
      "arrowhead": "filled",
      "animation": "flow",
      "dashArray": null
    },
    "functional:monitors": {
      "stroke": "#209cee",
      "strokeWidth": 1,
      "arrowhead": "open",
      "animation": null,
      "dashArray": "5,5"
    },
    "spatial:connects_to": {
      "stroke": "#b5b5b5",
      "strokeWidth": 2,
      "arrowhead": null,
      "animation": null,
      "dashArray": null
    },
    "spatial:adjacent_to": {
      "stroke": "#dbdbdb",
      "strokeWidth": 1,
      "arrowhead": null,
      "animation": null,
      "dashArray": "2,2"
    },
    "communication:publishes_to": {
      "stroke": "#48c774",
      "strokeWidth": 2,
      "arrowhead": "filled",
      "animation": "pulse",
      "dashArray": null
    },
    "hierarchical:reports_to": {
      "stroke": "#ff3860",
      "strokeWidth": 2,
      "arrowhead": "filled",
      "animation": null,
      "dashArray": null
    },
    "temporal:precedes": {
      "stroke": "#ffdd57",
      "strokeWidth": 2,
      "arrowhead": "filled",
      "animation": "flow",
      "dashArray": null,
      "curvature": 0.3
    },
    "social:follows": {
      "stroke": "#9b59b6",
      "strokeWidth": 1,
      "arrowhead": "open",
      "animation": null,
      "dashArray": null
    }
  }
}
```

#### Layout Algorithm Hints by Canonical Type

Canonical types also guide layout selection:

- `spatial:*` → Geographic or Grid layout (preserve physical positions)
- `functional:*` → Hierarchical or Layered layout (show flow direction)
- `hierarchical:*` → Tree or Radial layout (emphasize parent-child)
- `temporal:*` → Timeline or Sequence layout (left-to-right ordering)
- `social:*` → Force-directed layout (cluster by connection strength)
- `communication:*` → Force-directed or Hub layout (show message hubs)

#### Schema with Canonical Type

Updated formal schema:

```json
{
  "connection_rules": {
    "<relation_name>": {
      "canonical_type": "<category>:<type>",        // NEW: maps to standard taxonomy
      "target_types": ["<agent_type>", ...],
      "match": "<dotted.path.or.expression>",
      "directed": true,
      "weight": "<dotted.path.or-expression>",
      "label": "<display_label>",
      "multiplicity": "one-to-many|many-to-one|many-to-many",
      "default_weight": 1.0,
      "style_override": {                           // Optional: override canonical style
        "stroke": "#custom",
        "strokeWidth": 2
      }
    }
  }
}
```

#### Benefits of Canonical Typing

1. **Consistent Visualization**: Same relation types look identical across use cases
2. **Layout Optimization**: Algorithm selection based on relationship semantics
3. **Graph Algorithms**: Run centrality, clustering on semantic edge types
4. **Reusable Queries**: AQL queries like "find all functional:supplies paths"
5. **Style Inheritance**: Use cases inherit styles but can override
6. **Documentation**: Clear semantics for each relationship

#### Registering Custom Canonical Types

Use cases can extend the taxonomy if needed:

```json
{
  "custom_canonical_types": {
    "agricultural:grazes_at": {
      "category": "spatial",
      "semantics": "Livestock grazing location",
      "style": {
        "stroke": "#8fbc8f",
        "strokeWidth": 2
      }
    }
  }
}
```

The visualizer merges custom types with the standard taxonomy.

#### Fallback Behavior

If `canonical_type` is omitted, the visualizer:
1. Attempts to infer category from relation name (e.g., `supplies` → `functional:supplies`)
2. Falls back to generic `functional:generic` with neutral styling
3. Logs a warning suggesting canonical type declaration

This ensures backward compatibility while encouraging standardization.


**Strategy 3: From Message History** (Optional)
```
GET /api/v1/communications/messages?from=PUMP-001&limit=100
```
Infer connections from recent message exchanges between agents.

**Strategy 4: Explicit Topology Document**
```json
// Store in agent metadata or separate collection
{
  "topology": {
    "connections": [
      {"from": "PUMP-001", "to": "PIPE-015", "type": "supplies"},
      {"from": "SENSOR-008", "to": "PIPE-015", "type": "monitors"}
    ]
  }
}
```

#### Real-Time Updates via Agent Events

**Option 1: Polling** (Simple, works now)
```javascript
// Visualizer polls agent list every 5 seconds
setInterval(async () => {
  const agents = await fetch('/api/v1/agents').then(r => r.json());
  visualizer.updateAgents(agents);
}, 5000);
```

**Option 2: WebSocket** (Efficient, future enhancement)
```
WS /api/v1/agents/subscribe
```

Subscribe to agent state changes:
```json
{
  "type": "subscribe",
  "filter": {
    "types": ["pump", "sensor"],
    "events": ["state_change", "status_change"]
  }
}
```

Receive updates:
```json
{
  "type": "agent_update",
  "agent_id": "SENSOR-008",
  "timestamp": "2025-10-24T10:30:45Z",
  "changes": {
    "status": "alert",
    "metadata.pressure": 85.3
  }
}
```

**Option 3: Server-Sent Events** (Middle ground)
```
GET /api/v1/agents/stream
```

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
4. **Framework Agnostic**: Works with any agent type
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
3. Add legend showing agent types
4. Performance testing with 27 agents
5. Write framework documentation
6. Create use case configuration guide
7. Record demo video
8. Write coding session document

**Deliverables**:
- ✅ Loading spinners and error messages
- ✅ Search bar to find agents by ID/name
- ✅ Legend with agent type colors
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

**Strategy**: Use agent type to determine CRS:
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

**Test Connection Rules Inference**:
```go
func TestEdgeInference_ConnectionRules(t *testing.T) {
    tests := []struct {
        name         string
        agentType    string
        agentData    map[string]any
        rule         ConnectionRule
        expectedEdges []Edge
    }{
        {
            name: "Simple match - downstream pipes",
            agentType: "pump",
            agentData: map[string]any{
                "id": "PUMP-001",
                "metadata": map[string]any{
                    "downstream_pipes": []string{"PIPE-015", "PIPE-016"},
                },
            },
            rule: ConnectionRule{
                CanonicalType: "functional:supplies",
                TargetTypes: []string{"pipe"},
                Match: "$.metadata.downstream_pipes",
                Directed: true,
            },
            expectedEdges: []Edge{
                {From: "PUMP-001", To: "PIPE-015", Type: "supplies", Directed: true},
                {From: "PUMP-001", To: "PIPE-016", Type: "supplies", Directed: true},
            },
        },
        {
            name: "Conditional match - type filter",
            agentType: "sensor",
            agentData: map[string]any{
                "id": "SENSOR-023",
                "metadata": map[string]any{
                    "connections": []map[string]any{
                        {"target": "PIPE-015", "type": "monitors"},
                        {"target": "PUMP-001", "type": "monitors"},
                        {"target": "VALVE-007", "type": "controls"}, // Should be filtered out
                    },
                },
            },
            rule: ConnectionRule{
                CanonicalType: "functional:monitors",
                TargetTypes: []string{"pipe", "pump"},
                Match: "$.metadata.connections[?(@.type == 'monitors')].target",
                Directed: false,
            },
            expectedEdges: []Edge{
                {From: "SENSOR-023", To: "PIPE-015", Type: "monitors", Directed: false},
                {From: "SENSOR-023", To: "PUMP-001", Type: "monitors", Directed: false},
                // VALVE-007 excluded (type mismatch)
            },
        },
        {
            name: "Weighted edges - flow capacity",
            agentType: "pump",
            agentData: map[string]any{
                "id": "PUMP-001",
                "metadata": map[string]any{
                    "downstream_pipes": []string{"PIPE-015"},
                    "flow_capacity": 450.5,
                },
            },
            rule: ConnectionRule{
                CanonicalType: "functional:supplies",
                TargetTypes: []string{"pipe"},
                Match: "$.metadata.downstream_pipes",
                Weight: "$.metadata.flow_capacity",
                DefaultWeight: 1.0,
                Directed: true,
            },
            expectedEdges: []Edge{
                {From: "PUMP-001", To: "PIPE-015", Type: "supplies", Weight: 450.5, Directed: true},
            },
        },
        {
            name: "No match - missing field",
            agentType: "pump",
            agentData: map[string]any{
                "id": "PUMP-002",
                "metadata": map[string]any{
                    // downstream_pipes field missing
                },
            },
            rule: ConnectionRule{
                CanonicalType: "functional:supplies",
                TargetTypes: []string{"pipe"},
                Match: "$.metadata.downstream_pipes",
                Directed: true,
            },
            expectedEdges: []Edge{}, // No edges
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            inferrer := NewEdgeInferrer()
            edges := inferrer.InferEdges(tt.agentType, tt.agentData, tt.rule)
            
            assert.Equal(t, len(tt.expectedEdges), len(edges), "edge count mismatch")
            
            for i, expected := range tt.expectedEdges {
                assert.Equal(t, expected.From, edges[i].From)
                assert.Equal(t, expected.To, edges[i].To)
                assert.Equal(t, expected.Type, edges[i].Type)
                assert.Equal(t, expected.Directed, edges[i].Directed)
                if expected.Weight != 0 {
                    assert.InDelta(t, expected.Weight, edges[i].Weight, 0.01)
                }
            }
        })
    }
}
```

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

## Canonical Relationship Type Registry

**Location**: `internal/web/visualization/canonical_types.json`

**Purpose**: Single source of truth for all canonical relationship types across all use cases.

**Format**:
```json
{
  "taxonomyVersion": "2025.10",
  "updated": "2025-10-24T00:00:00Z",
  "categories": {
    "spatial": {
      "types": {
        "connects_to": {
          "directed": false,
          "description": "Physical or logical connection between entities (symmetric)",
          "examples": ["pipe connects_to pipe", "road connects_to intersection"]
        },
        "adjacent_to": {
          "directed": false,
          "description": "Spatial adjacency or proximity (symmetric)",
          "examples": ["rack adjacent_to rack", "room adjacent_to room"]
        },
        "contains": {
          "directed": true,
          "description": "Spatial containment (parent → child)",
          "examples": ["warehouse contains robot", "building contains room"]
        },
        "located_in": {
          "directed": true,
          "description": "Entity is located within a space (child → parent)",
          "examples": ["vehicle located_in zone", "sensor located_in pipe"]
        }
      }
    },
    "functional": {
      "types": {
        "supplies": {
          "directed": true,
          "description": "Provides resource or service to target",
          "examples": ["pump supplies pipe", "warehouse supplies truck"]
        },
        "consumes": {
          "directed": true,
          "description": "Consumes resource from source",
          "examples": ["facility consumes power", "robot consumes battery"]
        },
        "monitors": {
          "directed": false,
          "description": "Observes or tracks entity state (can be bidirectional)",
          "examples": ["sensor monitors pipe", "camera monitors zone"]
        },
        "controls": {
          "directed": true,
          "description": "Actuates or commands target entity",
          "examples": ["controller controls valve", "operator controls robot"]
        }
      }
    },
    "hierarchical": {
      "types": {
        "manages": {"directed": true},
        "reports_to": {"directed": true},
        "owns": {"directed": true},
        "part_of": {"directed": true}
      }
    },
    "temporal": {
      "types": {
        "follows": {"directed": true},
        "precedes": {"directed": true},
        "triggers": {"directed": true},
        "schedules": {"directed": true}
      }
    },
    "social": {
      "types": {
        "subscribes_to": {"directed": true},
        "follows": {"directed": true},
        "messages": {"directed": true},
        "participates_in": {"directed": false}
      }
    },
    "dependency": {
      "types": {
        "depends_on": {"directed": true},
        "requires": {"directed": true},
        "blocks": {"directed": true},
        "enables": {"directed": true}
      }
    }
  }
}
```

**CI Lint Step** (validate configs against registry):
```yaml
# .github/workflows/validate-configs.yml
name: Validate Visualization Configs

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      - name: Validate configs
        run: |
          go run scripts/validate_viz_configs.go \
            --registry=internal/web/visualization/canonical_types.json \
            --configs=usecases/*/config/visualization.json
```

**Validation Script** (`scripts/validate_viz_configs.go`):
```go
func validateAgainstRegistry(config *VisualizationConfig, registry *CanonicalTypeRegistry) error {
    // Check taxonomy version
    if config.TaxonomyVersion != registry.TaxonomyVersion {
        return fmt.Errorf("taxonomy version mismatch: config uses %s, registry is %s",
            config.TaxonomyVersion, registry.TaxonomyVersion)
    }
    
    // Validate all canonical_type references
    for agentType, rules := range config.ConnectionRules {
        for ruleName, rule := range rules {
            if !registry.HasType(rule.CanonicalType) {
                return fmt.Errorf("agent %s rule %s references unknown canonical type %s",
                    agentType, ruleName, rule.CanonicalType)
            }
            
            // Validate directionality matches registry
            registryType := registry.GetType(rule.CanonicalType)
            if rule.Directed != registryType.Directed {
                return fmt.Errorf("agent %s rule %s: directed=%v but registry says %v",
                    agentType, ruleName, rule.Directed, registryType.Directed)
            }
        }
    }
    
    return nil
}
```

**Version Update Process**:
1. Update `canonical_types.json` with new types or changes
2. Increment `taxonomyVersion` (e.g., `2025.10` → `2025.11`)
3. CI validates all configs still reference valid types
4. Configs referencing old version get migration warnings

## Performance Telemetry (Runtime Observability)

**Lightweight Client Metrics** (DevTools overlay):
```javascript
class PerformanceMonitor {
  private metrics: {
    fps: number;
    droppedFrames: number;
    renderTime: number;
    wsReconnects: number;
    agentCount: number;
    edgeCount: number;
  };
  
  private overlay: HTMLElement;
  
  constructor(config: VisualizationConfig) {
    if (config.telemetry?.enabled) {
      this.createOverlay();
      this.startMonitoring();
    }
  }
  
  startMonitoring() {
    // FPS measurement
    let lastFrameTime = performance.now();
    let frameCount = 0;
    
    const measureFPS = () => {
      frameCount++;
      const now = performance.now();
      const elapsed = now - lastFrameTime;
      
      if (elapsed >= 1000) {
        this.metrics.fps = Math.round((frameCount * 1000) / elapsed);
        frameCount = 0;
        lastFrameTime = now;
        this.updateOverlay();
      }
      
      requestAnimationFrame(measureFPS);
    };
    requestAnimationFrame(measureFPS);
  }
  
  recordRenderTime(duration: number) {
    this.metrics.renderTime = Math.round(duration);
    this.updateOverlay();
  }
  
  recordDroppedFrame() {
    this.metrics.droppedFrames++;
  }
  
  recordWSReconnect() {
    this.metrics.wsReconnects++;
    this.updateOverlay();
  }
  
  createOverlay() {
    this.overlay = document.createElement('div');
    this.overlay.id = 'viz-telemetry';
    this.overlay.style.cssText = `
      position: fixed;
      top: 10px;
      right: 10px;
      background: rgba(0, 0, 0, 0.8);
      color: #0f0;
      font-family: monospace;
      font-size: 12px;
      padding: 10px;
      border-radius: 4px;
      z-index: 10000;
      pointer-events: none;
    `;
    document.body.appendChild(this.overlay);
  }
  
  updateOverlay() {
    if (!this.overlay) return;
    
    const fpsColor = this.metrics.fps >= 30 ? '#0f0' : '#f00';
    
    this.overlay.innerHTML = `
      <div><span style="color: ${fpsColor}">FPS: ${this.metrics.fps}</span></div>
      <div>Render: ${this.metrics.renderTime}ms</div>
      <div>Dropped: ${this.metrics.droppedFrames}</div>
      <div>WS Reconnects: ${this.metrics.wsReconnects}</div>
      <div>Agents: ${this.metrics.agentCount}</div>
      <div>Edges: ${this.metrics.edgeCount}</div>
    `;
  }
}

// Usage
const monitor = new PerformanceMonitor(config);

// In renderer
const startTime = performance.now();
renderer.render(agents, edges);
const duration = performance.now() - startTime;
monitor.recordRenderTime(duration);

// In WebSocket client
ws.addEventListener('close', () => {
  monitor.recordWSReconnect();
});
```

**Console Metrics** (alternative to overlay):
```javascript
// Log to console every 5 seconds
setInterval(() => {
  console.table({
    FPS: monitor.metrics.fps,
    'Render Time (ms)': monitor.metrics.renderTime,
    'Dropped Frames': monitor.metrics.droppedFrames,
    'WS Reconnects': monitor.metrics.wsReconnects,
    'Agent Count': monitor.metrics.agentCount,
    'Edge Count': monitor.metrics.edgeCount,
  });
}, 5000);
```

**Performance Budget Alerts**:
```javascript
if (monitor.metrics.fps < 20) {
  console.warn('⚠️ FPS below threshold (20), consider switching to WebGL renderer');
}

if (monitor.metrics.renderTime > 100) {
  console.warn('⚠️ Render time > 100ms, consider reducing node count or enabling LOD');
}

if (monitor.metrics.droppedFrames > 100) {
  console.error('🚨 100+ dropped frames, serious performance issue');
}
```

## Bundler Configuration (Code Splitting)

**Tool**: esbuild (fast, modern)

**Config** (`build.config.js`):
```javascript
const esbuild = require('esbuild');

esbuild.build({
  entryPoints: {
    'topology-visualizer': 'src/topology-visualizer.js',
    'basemap': 'src/basemap-loader.js', // Lazy loaded
    'webgl-renderer': 'src/renderers/webgl-renderer.js', // Lazy loaded
  },
  bundle: true,
  splitting: true, // Enable code splitting
  format: 'esm',
  outdir: 'static/js/visualization/dist',
  minify: true,
  sourcemap: true,
  target: ['es2020'],
  external: [
    // Don't bundle these, load from CDN
    'maplibre-gl',
  ],
  define: {
    'process.env.NODE_ENV': '"production"',
  },
  treeShaking: true,
  metafile: true, // Generate bundle analysis
}).then(result => {
  // Analyze bundle sizes
  const analysis = require('esbuild-visualizer');
  analysis(result.metafile, {
    filename: 'bundle-analysis.html',
  });
  
  console.log('✅ Build complete');
  console.log('📦 Bundle sizes:');
  for (const [name, output] of Object.entries(result.metafile.outputs)) {
    const sizeKB = (output.bytes / 1024).toFixed(2);
    console.log(`  ${name}: ${sizeKB} KB`);
  }
});
```

**Lazy Loading** (basemap):
```javascript
async function loadBasemap() {
  if (config.layout.algorithm !== 'geographic') {
    return null; // Don't load basemap if not needed
  }
  
  // Dynamic import (code splitting)
  const { MapLibreGL } = await import('./basemap-loader.js');
  
  // Load from CDN (cached by browser)
  await loadScript('https://unpkg.com/maplibre-gl@3.6.2/dist/maplibre-gl.js');
  await loadCSS('https://unpkg.com/maplibre-gl@3.6.2/dist/maplibre-gl.css');
  
  return new MapLibreGL(config.basemap);
}
```

**Bundle Size Targets** (validated):
```javascript
// package.json
{
  "scripts": {
    "build": "node build.config.js",
    "analyze": "open bundle-analysis.html",
    "check-size": "node scripts/check-bundle-size.js"
  }
}

// scripts/check-bundle-size.js
const fs = require('fs');
const path = require('path');

const limits = {
  'topology-visualizer': 50 * 1024, // 50 KB
  'basemap': 200 * 1024, // 200 KB
  'webgl-renderer': 80 * 1024, // 80 KB
};

let failed = false;

for (const [bundle, limit] of Object.entries(limits)) {
  const filePath = path.join(__dirname, `../static/js/visualization/dist/${bundle}.js`);
  const stats = fs.statSync(filePath);
  const sizeKB = (stats.size / 1024).toFixed(2);
  const limitKB = (limit / 1024).toFixed(0);
  
  if (stats.size > limit) {
    console.error(`❌ ${bundle}.js (${sizeKB} KB) exceeds limit (${limitKB} KB)`);
    failed = true;
  } else {
    console.log(`✅ ${bundle}.js (${sizeKB} KB) within limit (${limitKB} KB)`);
  }
}

if (failed) {
  process.exit(1);
}
```

**CI Integration**:
```yaml
# .github/workflows/build.yml
- name: Build bundles
  run: npm run build

- name: Check bundle sizes
  run: npm run check-size

- name: Upload bundle analysis
  uses: actions/upload-artifact@v3
  with:
    name: bundle-analysis
    path: bundle-analysis.html
```

## Delivery Plan (Revised - Realistic)

## Config Validation & Versioning Enforcement

### Mandatory Fields

**Every config MUST declare**:
```json
{
  "$schema": "https://codevaldcortex.io/schemas/visualization/v1.0.0.json",
  "schemaVersion": "1.0.0"
}
```

Configs without `schemaVersion` are **rejected at load time**.

### Schema URL Resolvable (Served by App)

**URL Pattern**: `https://<your-domain>/schemas/visualization/v{version}.json`

**Backend Route** (`internal/api/handlers/schema.go`):
```go
func (h *SchemaHandler) ServeVisualizationSchema(c *gin.Context) {
    version := c.Param("version") // e.g., "v1.0.0"
    
    schemaPath := fmt.Sprintf("internal/web/visualization/schemas/%s.json", version)
    schema, err := os.ReadFile(schemaPath)
    if err != nil {
        c.JSON(404, gin.H{"error": "schema version not found"})
        return
    }
    
    c.Header("Content-Type", "application/schema+json")
    c.Header("Cache-Control", "public, max-age=86400") // Cache 24h
    c.Data(200, "application/schema+json", schema)
}

// Register route
router.GET("/schemas/visualization/:version", h.ServeVisualizationSchema)
```

**Schema File** (`internal/web/visualization/schemas/v1.0.0.json`):
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://codevaldcortex.io/schemas/visualization/v1.0.0.json",
  "title": "CodeValdCortex Visualization Configuration",
  "type": "object",
  "required": ["$schema", "schemaVersion", "visualization"],
  "properties": {
    "$schema": {
      "type": "string",
      "pattern": "^https://codevaldcortex.io/schemas/visualization/v[0-9]+\\.[0-9]+\\.[0-9]+\\.json$"
    },
    "schemaVersion": {
      "type": "string",
      "pattern": "^[0-9]+\\.[0-9]+\\.[0-9]+$",
      "description": "Semantic version of the config schema"
    },
    "visualization": {
      "type": "object",
      "required": ["id", "title", "type"],
      "properties": {
        "id": {"type": "string", "pattern": "^[a-z0-9-]+$"},
        "title": {"type": "string"},
        "type": {"enum": ["network", "map", "tree", "grid"]},
        "locale": {"type": "string", "pattern": "^[a-z]{2}-[A-Z]{2}$"}
      }
    },
    "crs": {
      "type": "object",
      "required": ["geographic"],
      "properties": {
        "geographic": {"type": "string", "pattern": "^EPSG:[0-9]+$"}
      }
    }
  },
  "additionalProperties": false
}
```

### Deprecations & Migrations (Config Evolution)

**Deprecations Section** (in config):
```json
{
  "schemaVersion": "1.1.0",
  "deprecations": [
    {
      "field": "dataSource.polling.interval",
      "deprecated_in": "1.1.0",
      "removed_in": "2.0.0",
      "replacement": "dataSource.realtime.enabled",
      "message": "Use realtime WebSocket instead of polling"
    }
  ]
}
```

**Migrations Registry** (`internal/web/visualization/migrations.json`):
```json
{
  "migrations": {
    "1.0.0->1.1.0": {
      "description": "Add CRS indoor support, rename oldField",
      "changes": [
        {"op": "rename", "from": "oldField", "to": "newField"},
        {"op": "add", "path": "crs.indoor", "default": null}
      ]
    },
    "1.1.0->1.2.0": {
      "description": "Remove polling, enforce WebSocket",
      "changes": [
        {"op": "remove", "path": "dataSource.polling"},
        {"op": "require", "path": "dataSource.realtime"}
      ]
    }
  },
  "unknown_field_policy": "reject"
}
```

**CI Validation** (fail on unknown fields):
```go
func validateConfig(config map[string]any, schema *JSONSchema) error {
    // Validate against JSON Schema
    if err := schema.Validate(config); err != nil {
        return err
    }
    
    // Check for unknown fields (strict mode)
    knownFields := extractFieldPaths(schema)
    actualFields := extractFieldPaths(config)
    
    unknownFields := difference(actualFields, knownFields)
    if len(unknownFields) > 0 {
        if migrationRegistry.UnknownFieldPolicy == "reject" {
            return fmt.Errorf("unknown fields detected: %v", unknownFields)
        } else {
            log.Warn("Unknown fields (will be ignored): %v", unknownFields)
        }
    }
    
    return nil
}
```

**Deprecation Warnings**:
```go
func checkDeprecations(config *VisualizationConfig) {
    for _, dep := range config.Deprecations {
        if fieldExists(config, dep.Field) {
            log.Warnf("⚠️  Deprecated field '%s' (since %s, removed in %s): %s",
                dep.Field, dep.DeprecatedIn, dep.RemovedIn, dep.Message)
        }
    }
}
```

### Version Migration Mechanism

**Version History**:
```go
const (
    VisualizationConfigV1_0_0 = "1.0.0" // Initial release
    VisualizationConfigV1_1_0 = "1.1.0" // Future: add temporal layers
)

var supportedVersions = []string{
    VisualizationConfigV1_0_0,
    VisualizationConfigV1_1_0,
}
```

**Migration Registry**:
```go
type ConfigMigration func(config map[string]any) (map[string]any, error)

var migrations = map[string]ConfigMigration{
    "1.0.0->1.1.0": migrateV1_0_to_V1_1,
}

func migrateV1_0_to_V1_1(config map[string]any) (map[string]any, error) {
    // Example: Rename field
    if val, ok := config["oldField"]; ok {
        config["newField"] = val
        delete(config, "oldField")
    }
    
    // Update version
    config["schemaVersion"] = "1.1.0"
    return config, nil
}
```

**Auto-Migration at Load**:
```go
func LoadConfig(path string) (*VisualizationConfig, error) {
    raw := loadJSON(path)
    
    // CRITICAL: Reject if no version
    version, ok := raw["schemaVersion"].(string)
    if !ok || version == "" {
        return nil, fmt.Errorf("REJECTED: schemaVersion is mandatory")
    }
    
    // Check if current version
    if version != VisualizationConfigV1_1_0 {
        // Attempt migration
        migrated, err := migrateConfig(raw, version, VisualizationConfigV1_1_0)
        if err != nil {
            return nil, fmt.Errorf("migration failed: %w", err)
        }
        raw = migrated
        log.Printf("Migrated config from %s to %s", version, VisualizationConfigV1_1_0)
    }
    
    // Validate against JSON Schema
    if err := validateAgainstSchema(raw); err != nil {
        return nil, fmt.Errorf("schema validation failed: %w", err)
    }
    
    return parseConfig(raw)
}
```

### Validation Test Suite

**Test: Reject configs without version**:
```go
func TestConfigValidator_RejectMissingVersion(t *testing.T) {
    config := `{
        "visualization": {
            "id": "test",
            "type": "network"
        }
    }`
    
    _, err := LoadConfig(strings.NewReader(config))
    require.Error(t, err)
    assert.Contains(t, err.Error(), "schemaVersion is mandatory")
}
```

**Test: Reject configs without $schema**:
```go
func TestConfigValidator_RejectMissingSchema(t *testing.T) {
    config := `{
        "schemaVersion": "1.0.0",
        "visualization": {"id": "test"}
    }`
    
    _, err := LoadConfig(strings.NewReader(config))
    require.Error(t, err)
    assert.Contains(t, err.Error(), "$schema is required")
}
```

**Test: Migration v1.0 → v1.1**:
```go
func TestConfigMigration_V1_0_to_V1_1(t *testing.T) {
    configV1_0 := map[string]any{
        "schemaVersion": "1.0.0",
        "oldField": "value",
    }
    
    migrated, err := migrateV1_0_to_V1_1(configV1_0)
    require.NoError(t, err)
    
    assert.Equal(t, "1.1.0", migrated["schemaVersion"])
    assert.Equal(t, "value", migrated["newField"])
    assert.NotContains(t, migrated, "oldField")
}
```

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
   - Strategy 2: agent type `connection_rules`
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

## Performance SLOs (Service Level Objectives)

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

**Telemetry Integration**:
```javascript
class PerformanceMonitor {
  private metrics: {
    render: {
      initial: {duration: number[]},
      update: {duration: number[]},
      fps: number[],
      droppedFrames: number
    },
    ws: {
      reconnect: {duration: number[]}
    },
    api: {
      agents: {latency: number[]}
    },
    memory: {
      heapSize: number
    }
  };
  
  recordInitialRender(duration: number) {
    this.metrics.render.initial.duration.push(duration);
    
    // Check SLO
    if (duration > 800) {
      console.warn(`⚠️  SLO violation: Initial render ${duration}ms > 800ms target`);
      this.sloViolations.push({metric: 'render.initial', value: duration, threshold: 800});
    }
  }
  
  recordUpdate(duration: number) {
    this.metrics.render.update.duration.push(duration);
    
    if (duration > 50) {
      console.warn(`⚠️  SLO violation: Update ${duration}ms > 50ms target`);
    }
  }
  
  computeP95(values: number[]): number {
    const sorted = [...values].sort((a, b) => a - b);
    const index = Math.ceil(sorted.length * 0.95) - 1;
    return sorted[index];
  }
  
  getSLOReport(): SLOReport {
    return {
      'render.initial.p95': this.computeP95(this.metrics.render.initial.duration),
      'render.update.p95': this.computeP95(this.metrics.render.update.duration),
      'render.fps.p95': this.computeP95(this.metrics.render.fps),
      'ws.reconnect.p95': this.computeP95(this.metrics.ws.reconnect.duration),
      'violations': this.sloViolations.length
    };
  }
}

// Log SLO report every minute
setInterval(() => {
  const report = monitor.getSLOReport();
  console.table(report);
  
  // Send to monitoring backend
  fetch('/api/v1/telemetry', {
    method: 'POST',
    body: JSON.stringify({
      timestamp: Date.now(),
      visualization_id: config.visualization.id,
      slo_report: report
    })
  });
}, 60000);
```

**SLO Dashboard** (optional Grafana/Prometheus):
```yaml
# prometheus.yml
- job_name: 'visualization_slos'
  metrics_path: '/api/v1/metrics'
  static_configs:
    - targets: ['localhost:8083']
```

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
