# Data Source and Inference

## Data Source

The visualizer uses the Agent API as the canonical data source. Endpoints:
- GET /api/v1/agents — paginated list of agents (include ETag)
- GET /api/v1/agents/{id} — single agent
- GET /api/v1/visualization/edges?since={seq} — optional materialized edges view / or inferred on server
- WS /api/v1/visualization/ws — seq-numbered JSON Patch diffs (RFC6902) for efficient updates

### Server guarantees
- Responses include seq numbers and ETags for deterministic replay.
- Server enforces pagination and provides snapshot endpoints for resync.

## Edge inference

- Primary source: agent.connection_rules (matches on other agents via JSONPath-like match expressions)
- Secondary: materialized edges based on message history or DB views (AQL snippet provided in full document)

## Edge canonicalization

- connection_rules specify canonical_type from canonical_types registry
- canonical_type maps to a taxonomy (supply, observe, route, command, host, depends_on)
- Each canonical_type has a default weight and display style

## JSON Patch diff semantics

- Server emits patches with seq numbers and compact diffs to minimize bandwidth.
- Client applies patches in order; if a gap > replay window occurs, client re-fetches snapshot.
- Patches include metadata for affected node ids, revision, and optional causal timestamp.

## Replay & Backpressure

- Server stores recent N patches (ReplayWindowSize). If client falls behind beyond this window, the client must perform a snapshot resync.
- Each client has a MaxClientBufferSize; once exceeded, server may drop oldest buffered patches or ask client to resync.

