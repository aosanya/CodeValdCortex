# Architecture

## Layers

- Data Layer: Agent API, optional materialized edges (DB view). Server enforces row-level RBAC here.
- Inference Layer: Rules engine or AQL/SQL view that materializes connection_rules into edges. Use deterministic hashing for edge ids.
- API Layer: REST endpoints (/api/v1/agents, /api/v1/agents/{id}, /api/v1/visualization/config-schema) and WebSocket hub for diffs and replays.
- Client Layer: Visualization module under /static/js/visualization. Responsibilities: AgentDataSource, TopologyModel, Layouts, Renderers, Interactions, and Telemetry.

## Component responsibilities

- AgentDataSource: Fetch agents, materialized edges, provide snapshots and patches, support replay and resync.
- TopologyModel: Apply JSON Patch diffs, keep seq, debounce UI updates, maintain deterministic ordering.
- Layouts: Geographic (lat/lon -> mercator), Force-Directed (seeded RNG), Hierarchical (for tree-like topologies).
- Renderers: CanvasRenderer (MVP), SVGRenderer, WebGLRenderer (optional for large graphs).
- Security: Server enforces RBAC and field-level masking; client performs least-privilege UI rendering but does not trust client-side filters.

## Files & suggested module layout

- /internal/web/visualization/
  - config_validator.go — validates schemaVersion and JSON Schema
  - handler_visualization.go — serves schema and config artifacts
- /static/js/visualization/
  - agent-data-source.js
  - topology-model.js
  - layouts/
  - renderers/
  - ui/

## Determinism contract

- Edge IDs: SHA256(sorted(nodeA, nodeB) + canonical_type + label) -> hex -> short id
- Layout RNG: seeded per config (seed is part of config and used to generate deterministic positions when using force layouts)
- Sorting: stable sort by deterministic id across all rendering steps

