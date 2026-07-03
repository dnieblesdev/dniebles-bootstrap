# Exploration: first Go planning slice

### Current State
The repository is still documentation/specification-only. The archived design says the bootstrapper is a domain-first Go orchestrator, but implementation is deferred. No Go source files exist yet, and the current architecture explicitly pushes future implementation to start with catalog schema and pure planning before installers or execution.

### Affected Areas
- `README.md` — confirms the repo is not yet implemented and states the domain-first, CLI-first/TUI-later direction.
- `AGENT.md` — defines the no-code-before-specs rule and the one-core/thin-interfaces architecture guardrails.
- `openspec/changes/archive/2026-07-03-design-bootstrap-orchestrator/design-bootstrap-orchestrator/design.md` — identifies the domain concepts, layering, and the “catalog schema and pure planning first” rollout note.
- `openspec/changes/archive/2026-07-03-design-bootstrap-orchestrator/design-bootstrap-orchestrator/specs/*.md` — establishes the required planning behaviors, environment facts, and attention-required handling.

### Approaches
1. **Minimal planning core slice** — create `go.mod`, domain entities, plan builder, and table-driven tests; keep TOML loading out of the first slice except as a thin adapter boundary.
   - Pros: smallest meaningful Go increment; preserves architecture; easiest to test deeply.
   - Cons: no real catalog file parsing yet.
   - Effort: Low

2. **Planning plus TOML adapter slice** — include `go.mod`, domain entities, catalog loader, TOML schema, plan builder, and tests.
   - Pros: end-to-end from file to plan; validates the in-repo catalog format early.
   - Cons: larger first slice; risks leaking TOML concerns into domain design too early.
   - Effort: Medium

### Recommendation
Start with the minimal planning core slice, but define the catalog loader/TOML schema as an explicit next slice. The first meaningful Go increment should prove the domain model and planning rules with tests before any installer, execution, or CLI wiring.

### Risks
- Overloading the first slice with schema/adapter concerns could blur the domain boundary.
- Deferring a real catalog file means some integration assumptions remain unvalidated until the next slice.

### Ready for Proposal
Yes — the next step should be a proposal/spec for a first Go slice centered on domain entities + pure planning, with TOML adapter work explicitly deferred or isolated.
