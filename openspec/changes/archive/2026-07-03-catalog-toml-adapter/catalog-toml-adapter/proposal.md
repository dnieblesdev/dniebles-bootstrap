# Proposal: Catalog TOML Adapter

## Intent

Introduce the first repository-local catalog authoring surface by decoding TOML into the existing `internal/planning` domain model. This lets the project prove catalog schema direction and file-to-plan integration while keeping planning pure, format-agnostic, and free of runtime side effects.

## Scope

### In Scope
- Add isolated `internal/catalog/toml` decoding and shallow structural validation.
- Map repo-local TOML catalog data into `planning.Catalog` inputs.
- Add one small `catalog/*.toml` fixture if it proves the schema and tests.
- Add Go tests for decode behavior and `planning.BuildPlan` integration.

### Out of Scope
- CLI commands, installers, command runner, Bubble Tea TUI.
- Git, dotfiles, dotlink runtime, OS probing, environment detection adapters.
- Remote catalog loading or deep validation duplicated from `internal/planning`.
- Planning core changes, unless a tiny boundary fix is unavoidable and justified.

## Capabilities

### New Capabilities
- `catalog-toml-adapter`: Repository-local TOML catalog decoding into planning domain inputs.

### Modified Capabilities
- `planning-core`: Integration only; requirements stay format-agnostic and should not absorb TOML concerns.

## Approach

Create TOML-shaped adapter structs inside `internal/catalog/toml`, decode a minimal schema for profiles, bundles, resources, dependencies, config keys, and environment conditions, then translate into `planning.Catalog`. The adapter owns syntax, required fields, duplicate/empty names, and basic malformed-reference checks. Unknown refs and planning semantics remain planner concerns. A fixture catalog should mirror real repo intent without invoking installers or probing the host.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/catalog/toml` | New | TOML decoder, adapter structs, shallow validation, mapping. |
| `catalog/*.toml` | New | Small sample/fixture catalog if useful for integration. |
| `internal/catalog/toml/*_test.go` | New | Table-driven decode and file-to-`BuildPlan` tests. |
| `internal/planning` | Unchanged | Boundary target; only tiny justified fixes allowed. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Over-validation duplicates planner behavior | Med | Keep adapter validation structural only. |
| Under-validation creates confusing fixture failures | Med | Validate required fields, names, kinds, and malformed refs. |
| TOML concerns leak into planning | Med | Translate at adapter boundary; do not add parser/schema types to core. |

## Rollback Plan

Revert the adapter package, fixture catalog, tests, and this SDD change. No migrations, commands, installers, or external state are introduced.

## Dependencies

- Existing `internal/planning` domain types and `BuildPlan` API.
- A TOML decoding library choice during implementation.

## Success Criteria

- [ ] Fixture TOML decodes successfully into `planning.Catalog`.
- [ ] Decoded profile/resources map correctly to planning inputs.
- [ ] Integration test runs decoded catalog through `planning.BuildPlan`.
- [ ] Tests prove no runtime side effects, OS probes, or command execution.
