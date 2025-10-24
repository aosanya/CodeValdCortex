# Security, Accessibility, and Testing

## Security model

- RBAC enforced server-side. The API filters rows (agents/edges) by role and field-level masking.
- denyEdges[]: Server-level filter used in AQL or SQL queries to exclude edges from materialization for specific roles.
- Audit events for config changes, schema reads, and access violations. Config updates require schemaVersion bump when incompatible.

## Expression sandboxing

- Expression language: JSONPath (restricted dialect)
- Guardrails: whitelist operators, no network/IO, CPU/time limit per-eval (e.g., 50ms), recursion depth limit, max result size.
- Server validates expressions on config upload; client re-validates before runtime.

## Testing strategy

- Unit tests for config validation and schema enforcement (Go tests).
- Golden-image visual tests comparing deterministic layouts (node positions within epsilon) stored under testdata/visualization/golden.
- Integration tests: WS replay + patch application across reconnects.
- Performance tests: load tests for up to 10k nodes using WebGL renderer.

## CI checks

- Lint schema and canonical_types.json.
- Validator CI job fails on missing schemaVersion or incompatible deprecations.

