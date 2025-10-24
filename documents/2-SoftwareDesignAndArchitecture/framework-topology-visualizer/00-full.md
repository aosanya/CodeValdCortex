(Full design document preserved here. This file is an exact copy of the original `framework-topology-visualizer.md` before splitting.)

Executive Summary

This document describes the Framework Topology Visualizer for CodeValdCortex. The visualizer is intended to be a reusable, canonical visualization component that can represent networks and topologies discovered or declared by agents across multiple use cases. It consumes agent data via the Agent API and infers a graph (G = (V, E)) from agent metadata, connection_rules, and optional message-history/materialized-edges. The visualizer exposes configuration-based styling, relationship canonicalization, real-time updates via JSON Patch diffs with sequencing and replay, deterministic IDs and seeded layout for stable renderings, CRS normalization, expression-sandboxed filters, server-side RBAC enforcement, telemetry, and test harnesses.

Goals

- Provide a standard, shareable visualizer across use cases that need maps or topology graphs (WMS, logistics, water networks, ride-hailing, tracking, etc.).
- Use agent API as the single source-of-truth; do not require a separate topology endpoint.
- Use a graph-theory model (G = (V, E)) with canonical relationship taxonomy so multiple domains can map onto the same renderer behavior and UI affordances.
- Ensure production readiness: schemaVersion mandatory, JSON Schema validation, expression sandbox guardrails, server-side RBAC enforcement, deterministic behavior for tests, and lightweight MVP path.

Design overview

- Nodes (V): Agents (physical/digital systems, sensors, microservices). Each agent has immutable id, agentType, metadata, coordinates, capabilities, connection_rules, and optional message-history.
- Edges (E): Inferred from connection_rules or materialized from message history/streams. Edges have source, target, canonical_type, directionality, weight, label, metadata, and stable id computed deterministically (SHA256 of ordered node ids + canonical_type + label).
- Directionality: Edges may be directed, bidirectional, or undirected. For fluid flows (water), edge directionality is derived from agent-provided flowDirection or inferred from timestamps/telemetry.
- Configuration: visualization-config.schema.json defines the schemaVersion, layout, renderer, style rules, matching expressions (JSONPath), and layout seeds.
- Real-time model: WebSocket publisher/subscriber with ordered, seq-numbered JSON Patch diffs (RFC 6902). Clients maintain replay windows; server supports replay and snapshot resync when the client misses more than the replay window.
- Security: Server-enforced RBAC and row-level filtering. denyEdges enforced in DB queries or materialized edge view. All expression evaluations validated and sandboxed.
- Determinism: Edge IDs deterministic and seeded RNG for layout jitter to ensure identical deterministic layouts across restarts when same seed is used.

(For the full original content, see the split files and the canonical 00-full.md copy.)
