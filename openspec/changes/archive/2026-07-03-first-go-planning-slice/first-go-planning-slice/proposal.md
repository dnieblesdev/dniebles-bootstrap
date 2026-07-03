# Proposal: First Go Planning Slice

## Intent

Establish the first executable and testable Go foundation for `dniebles-bootstrap` without coupling the domain to adapters, interfaces, installers, or dotfiles operations. This slice proves the pure planning core before TOML loading, CLI/TUI workflows, or execution concerns exist.

## Scope

### In Scope
- Add `go.mod` and pure core packages.
- Model domain concepts: Catalog-like in-memory model, Profile, Bundle, Resource, ResourceRef, ConfigPolicy, EnvironmentFacts, Plan, PlanStep, and PlanStepResult.
- Build deterministic, dependency-aware plans from already-decoded domain inputs.
- Add table-driven tests for ordering, scope expansion, missing config attention, and invalid references.

### Out of Scope
- TOML loader, catalog file schema, and infrastructure adapter wiring.
- CLI, Bubble Tea TUI, installers, command runner, git/dotfiles/dotlink operations.
- Runtime execution, first-run acquisition, and persisted catalog files.

## Capabilities

### New Capabilities
- `planning-core`: Pure Go domain model and plan builder for profile/point planning from in-memory inputs.

### Modified Capabilities
- None; no active `openspec/specs/` source specs exist yet.

## Approach

Create a format-agnostic core that accepts decoded domain objects, validates references and dependencies, expands profiles/bundles/resources, and emits a deterministic `Plan` with ordered `PlanStep` entries and reportable `PlanStepResult` values. Keep TOML and all side effects outside the core boundary.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `go.mod` | New | Initialize the Go module. |
| `internal/core` or equivalent | New | Domain entities, value objects, and plan builder. |
| `internal/core/*_test.go` | New | Table-driven tests for core planning behavior. |
| `openspec/changes/first-go-planning-slice/` | Modified | SDD artifacts for this implementation slice. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Core accidentally absorbs adapter/TOML concerns | Med | Accept only domain inputs; defer loader/schema to a later slice. |
| First slice grows beyond review comfort | Med | Size exception is approved, but keep commits as coherent work units with tests beside behavior. |
| Deferred TOML leaves integration assumptions unvalidated | Med | Treat adapter/schema as the next explicit SDD change. |

## Rollback Plan

Revert the change commit(s) for `go.mod`, core package files, tests, and this OpenSpec change. No migrations, runtime side effects, or external state changes are introduced.

## Dependencies

- Go toolchain available for development and tests.
- Existing architecture/spec guidance from the archived `design-bootstrap-orchestrator` change.

## Success Criteria

- [ ] Planning is deterministic for identical inputs.
- [ ] Dependencies are ordered before dependents.
- [ ] Missing config is surfaced as attention-required without blocking plan creation.
- [ ] Invalid or missing references are reported safely.
- [ ] Table-driven Go tests cover the pure core behaviors.
