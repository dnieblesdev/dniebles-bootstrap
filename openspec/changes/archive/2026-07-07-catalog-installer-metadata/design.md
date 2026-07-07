# Design: Catalog Installer Metadata

## Technical Approach

Extend the catalog-to-planning data path with inert structured metadata. The TOML adapter will decode nested `install` and `presence` tables, validate their shape, and map them into format-agnostic `planning.Resource` fields. `BuildPlan()` already copies full `Resource` values into `PlanStep.Resource`; no planner branching, command execution, installer dispatch, or apply mutation is added.

This design follows the delta spec in `openspec/changes/catalog-installer-metadata/specs/catalog-installer-metadata/spec.md`: structured install metadata, structured presence metadata, and inert metadata propagation through planning without changing execution semantics.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Structured install metadata | Add `InstallMetadata{Provider, Package}` to `planning.Resource`, backed by `[install]` in TOML. | Raw `command`; provider-specific execution config. | Keeps the model safe and provider-selectable without making the catalog shell-first. |
| Structured presence metadata | Add `PresenceMetadata{Kind, Name}` to `planning.Resource`, backed by `[presence]` in TOML. | Reusing `InstallationState`; shell checks. | Presence rules describe how future detectors may probe, while current planning state remains caller-supplied. |
| Adapter validation only | Validate non-empty paired fields and known presence kinds in `internal/catalog/toml/validate.go`. | Planner-time validation. | Existing semantic catalog checks live in the TOML adapter; planning stays format-agnostic and pure. |
| Preserve planning semantics | Do not change `BuildPlan()` logic. Add tests proving metadata survives in `PlanStep.Resource`. | Introduce metadata-aware statuses. | The slice is preparatory; metadata must be inert for later execution slices. |

## Data Flow

```text
catalog/bootstrap.toml
  └─ TOML resourceEntry {install,presence}
      └─ Decode() / validate()
          └─ mapResources()
              └─ planning.Resource metadata
                  └─ BuildPlan() copies Resource into PlanStep.Resource
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `catalog/bootstrap.toml` | Modify | Add representative `[install]` and `[presence]` examples for tool/runtime/package entries. |
| `internal/catalog/toml/schema.go` | Modify | Add private `installEntry` and `presenceEntry` structs and pointer fields on `resourceEntry`. |
| `internal/catalog/toml/catalog.go` | Modify | Map TOML metadata into planning metadata structs, cloning values without side effects. |
| `internal/catalog/toml/validate.go` | Modify | Validate install provider/package and presence kind/name shapes; reject empty partial metadata. |
| `internal/catalog/toml/catalog_test.go` | Modify | Assert decode, fixture load, and validation errors for metadata. |
| `internal/planning/types.go` | Modify | Add inert metadata structs and optional fields to `Resource`. |
| `internal/planning/builder_test.go` | Modify | Assert planning preserves metadata and remains pure/stable. |

## Interfaces / Contracts

```go
type Resource struct {
    Ref ResourceRef
    Description string
    DependsOn []ResourceRef
    ConfigPolicy ConfigPolicy
    Conditions EnvironmentConditions
    Install *InstallMetadata
    Presence *PresenceMetadata
}

type InstallMetadata struct {
    Provider string // e.g. "brew", "apt", "asdf", "go"
    Package  string // provider package name
}

type PresenceMetadata struct {
    Kind string // initially "path" or "command_exists"
    Name string // binary/path/check target name
}
```

The metadata is desired-state data only. Consumers must not interpret it as authorization to execute commands.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | TOML decoding maps install/presence metadata. | Extend `TestDecodeValidCatalog`. |
| Unit | Validation rejects partial/unknown metadata. | Add cases to `TestDecodeValidationErrors`. |
| Unit | `BuildPlan()` preserves metadata and stays pure. | Extend builder tests and clone helper. |
| Integration | Fixture catalog still builds the same plan. | Extend `TestLoadFileAndBuildPlanFromFixture` to inspect metadata without changing expected refs/statuses. |

## Migration / Rollout

No migration required. Existing catalogs without metadata remain valid; new fields are optional and inert.

## Open Questions

- [ ] None.
