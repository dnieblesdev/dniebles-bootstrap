# Design: Dotfiles Provider Readonly

## Technical Approach

Add dotfile resources as catalog data and detect local module availability at the CLI edge. `internal/planning` already supports `ResourceKindDotfile` and `InstallationState.PresentResources`, so the slice keeps `BuildPlan` pure: `cmd/dbootstrap` loads the catalog, runs existing detectors plus the new dotfiles detector, merges present resources, then calls `BuildPlan`.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|----------|--------|--------------------------|-----------|
| Dotfiles model | Add `Dotfiles []resourceEntry` with TOML tag `dotfiles`; map to `ResourceKindDotfile`. | New resource entry type. | Reuses existing resource fields (`depends_on`, config policy, conditions) and validation style. |
| Detection package | Create `internal/dotfiles` with `Detector{BasePath, Exists, ReadDir}` and package `Detect(catalog)`. | Extend `internal/state`. | Dotfiles are directory/module presence, not executable PATH lookup; separate adapter keeps responsibilities clear. |
| Default base path | Default to `$HOME/.dotfiles`. | Config flag/env var now. | Matches project convention and keeps this read-only slice small. |
| Planner boundary | Merge into `InstallationState.PresentResources` before planning. | Expand `BuildPlan` signature or probe from planner. | Existing input already expresses resource presence; planner remains deterministic and caller-driven. |
| Mutation boundary | No dotlink, clone, symlink drift, apply, install, or file writes. | Partial dotlink integration. | This slice only exposes availability to plan output. |

## Data Flow

```text
TOML [[dotfiles]] ──→ catalog.Resources[dotfile:name]
                              │
                              ▼
cmd/dbootstrap loads catalog ──→ state.Detect + dotfiles.Detect
                              │              │
                              └──── merge PresentResources ──→ planning.BuildPlan
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/catalog/toml/schema.go` | Modify | Add `Dotfiles []resourceEntry \`toml:"dotfiles"\`` to the private TOML schema. |
| `internal/catalog/toml/validate.go` | Modify | Collect dotfile refs, allow `dotfile:` in `supportedKind`, validate dotfile dependencies and bundle/profile refs. |
| `internal/catalog/toml/catalog.go` | Modify | Include dotfiles in resource capacity and `mapResources`. |
| `internal/catalog/toml/catalog_test.go` | Modify | Cover valid `[[dotfiles]]`, dotfile refs, dependency validation, and fixture loading. |
| `internal/dotfiles/detector.go` | Create | Read-only detector/provider with injectable filesystem seams and default `$HOME/.dotfiles`. |
| `internal/dotfiles/detector_test.go` | Create | Deterministic tests for repo missing, module present/absent, non-dotfile resources ignored, and nil defaults. |
| `cmd/dbootstrap/main.go` | Modify | Add `detectDotfilesState` seam, run after catalog load, merge present resources before `BuildPlan`. |
| `cmd/dbootstrap/main_test.go` | Modify | Stub dotfiles detection and assert catalog-load failures skip it. |
| `internal/planning/builder_test.go` | Modify | Add/adjust coverage proving supplied dotfile presence becomes `already_installed` without filesystem probing. |
| `catalog/bootstrap.toml` | Modify | Add a minimal dotfile resource/profile reference only if needed to exercise fixture coverage. |

## Interfaces / Contracts

```go
package dotfiles

type PathExists func(path string) bool
type ReadDir func(path string) ([]os.DirEntry, error)

type Detector struct {
    BasePath string // default: filepath.Join($HOME, ".dotfiles")
    Exists   PathExists
    ReadDir  ReadDir
}

func Detect(catalog planning.Catalog) planning.InstallationState
func (d Detector) Detect(catalog planning.Catalog) planning.InstallationState
```

Detection reports only catalog refs whose kind is `dotfile` and whose module name appears as a directory under `BasePath`. Repo absence or read errors return an empty state. Directory presence means “module available”, not “applied”.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | TOML schema/mapping/validation | Table-driven decoder tests. |
| Unit | Dotfiles detector | Inject `Exists`/`ReadDir`; no host-dependent filesystem assumptions except a default-missing smoke test. |
| Unit | Planner purity | Existing `InstallationState` tests include dotfile refs. |
| CLI | Composition-root merge and skip-on-load-error | Stub detector functions in `cmd/dbootstrap/main_test.go`. |
| E2E | Not needed in this slice | CLI unit tests cover the full in-process plan path. |

## Migration / Rollout

No migration required. The change is additive and read-only.

## Open Questions

- [ ] None.
