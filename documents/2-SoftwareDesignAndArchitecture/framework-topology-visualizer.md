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
| **UC-LIVE-001** (Mashambani) | Geographic Distribution | Owners, caretakers, animals | Rural/urban locations, connections |
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
    "crs": "EPSG:4326",
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

**Agent Update Event** (with sequence number for ordering):
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

### Security & Sandboxing

1. **No Function Execution**: Expressions are **declarative only** (no eval, no custom functions)
2. **Path Length Limit**: Max 10 segments (`$.a.b.c...`)
3. **Evaluation Timeout**: 10ms per expression
4. **Access Control**: Expressions cannot access fields outside agent's metadata (no `$..password` etc.)

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

## Delivery Plan (Revised - Realistic)

### Phase 0: Design Validation (3 days)

**Tasks**:
1. Create JSON Schema for visualization config
2. Implement config validator in Go
3. Create unit tests for validation
4. Performance-test renderer thresholds (100, 500, 1000, 5000 nodes)
5. Validate graph theory model with sample data

**Deliverables**:
- ✅ `internal/web/visualization/config-schema.json`
- ✅ `internal/web/visualization/config_validator.go`
- ✅ `internal/web/visualization/config_validator_test.go`
- ✅ Performance benchmark results (thresholds validated)
- ✅ Sample configs for 3 use cases

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

### Framework Metrics
- ✅ Config-driven: 0 custom code for basic visualizations
- ✅ Reusable: 5+ use cases using same component
- ✅ Performance: 30+ FPS with 1000 agents (Canvas), 60+ FPS with 5000 agents (WebGL)
- ✅ Update Latency: < 500ms for agent state changes (HTTP), < 100ms (WebSocket)
- ✅ Bundle Size: < 50KB core, < 250KB with basemap (realistic)

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
