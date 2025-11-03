# Framework Topology Visualizer — Overview

## Executive Summary

This repository/feature provides a canonical topology visualizer that reads agent data from the Agent API, infers networks, and renders them in the browser. The visualizer focuses on: reuse across multiple use-cases, deterministic rendering, server-side enforcement of access rules, and strong schema/versioning guarantees.

## Goals and Non-goals

Goals:
- Single source of truth: Agent API.
- Canonical relationship taxonomy for cross-domain reuse.
- Deterministic IDs and seeded layout for stable visualizations.
- Production safety: schemaVersion mandatory, expression sandbox, server-side RBAC.

Non-goals:
- Replace domain-specific GIS stacks. Use MapLibre-GL or basemap libs as optional plug-ins.

## When to use this visualizer

- When multiple roles and relationship types need to be visualized on a single, interactive topology map.
- When you need a cross-domain standard for relationships (supply, observe, route, command, host, depends_on).

## Key concepts

- Nodes = agents (id, agentType, coordinates, metadata)
- Edges = inferred or materialized connections with canonical_type
- Config = visualization-config.schema.json with schemaVersion
- Realtime updates = JSON Patch diffs with seq numbers and replay
- Renderer lifecycle = init -> render -> update -> destroy

## Razor-thin MVP (recommended)

- Canvas-only renderer
- Geographic + seeded Force-Directed layout
- Agent API polling (or WS if already available) with JSON Patch diffs
- Server-side config validator to require schemaVersion
- RBAC enforced server-side
- Minimal interactions: pan, zoom, node hover, side-panel details

(See implementation plan in 05-delivery-mvp.md)
