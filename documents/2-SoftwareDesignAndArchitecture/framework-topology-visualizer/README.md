Framework Topology Visualizer (split)

The large single-file design `framework-topology-visualizer.md` has been split into smaller files for easier review and implementation.

Files in this folder:
- 00-full.md — full preserved original document (canonical reference)
- 01-overview.md — Executive summary, goals, and MVP cut
- 02-architecture.md — Component architecture, determinism contract
- 03-data-source-and-inference.md — Agent API, edge inference, JSON Patch semantics
- 04-rendering.md — Renderer selection, layouts, basemap behavior
- 05-security-and-testing.md — RBAC, expression sandboxing, testing strategy
- 06-delivery-mvp.md — Delivery phases and acceptance criteria

Next steps:
- Review split files for accuracy and request further fine-grained splits if desired.
- Create `visualization-config.schema.json` artifact and add `config_validator.go` under `/internal/web/visualization`.
- When approved, I'll commit these files on the feature branch `feature/INFRA-017_network-topology-visualizer` (or do the commit now if you want).
