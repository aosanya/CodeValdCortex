# Rendering and Layouts

## Renderer selection heuristic

- Heuristic inputs: node count, edge count, geographic coords presence, device GPU capability, memory, config override.
- Defaults:
  - < 500 nodes: SVG or Canvas
  - 500–5k: Canvas
  - > 5k: WebGL

## Renderer lifecycle

- Interface: init(container, config) -> render(model) -> update(patch) -> destroy()
- Renderers must be idempotent on update() and honor deterministic ordering of model events.

## Layouts

- Geographic: Convert CRS -> WGS84 -> mercator for tile rendering. For indoor/local systems, use local-XY and overlay transformation.
- Force-Directed: Seeded RNG for determinism; allow anchor nodes by coordinates.
- Hierarchical/Topo: Useful for supply-chain and tree-like relationships.

## Basemap & fallback

- Basemap assets (tiles or vector) are optional. App must gracefully degrade to a plain background or grid when basemap fails.
- Basemap loading: lazy-load MapLibre-GL when config.basemap.enabled=true and device supports it.

## Interaction basics (MVP)

- Pan/zoom, click node->open detail panel, hover tooltip, keyboard focus navigation.
- Accessibility: focus outlines, readable labels, and a textual summary panel describing visible nodes/edges.

