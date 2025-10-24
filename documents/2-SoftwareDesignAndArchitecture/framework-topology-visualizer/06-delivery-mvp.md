# Delivery Plan and Razor-Thin MVP

## Razor-thin MVP (11 days recommended)

Scope:
- Canvas-only renderer
- Geographic + seeded force-directed layout
- Agent API polling with JSON Patch diffs
- Server-side config validator that requires schemaVersion
- RBAC enforcement for agent/edge queries
- Minimal interactions: pan/zoom, hover, side panel

## Phases

Phase 0 (2 days): Config schema + validator, canonical_types.json, migration map, CI schema lint.
Phase 1 (4 days): AgentDataSource, TopologyModel, Canvas renderer, seeded force layout.
Phase 2 (3 days): Interactions, RBAC enforcement wiring, golden-image tests for determinism.
Phase 3 (2 days): Observability + telemetry (FPS, WS reconnects, patch rejects), bundle optimizations.

## Acceptance criteria

- Configs with missing schemaVersion are rejected by the server.
- Deterministic layouts reproduce stable positions across runs with same seed (within epsilon).
- WS replay works for clients that reconnect within the ReplayWindowSize.
- RBAC denies edges and nodes correctly in a role-restricted test matrix.

