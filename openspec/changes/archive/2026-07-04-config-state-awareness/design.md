# Design: Config State Awareness

## Technical Approach

Add `internal/config` as a read-only adapter that mirrors `internal/environment` and `internal/state`: domain data stays in `internal/planning`, detectors live at the edge, and `cmd/dbootstrap` is the composition root. The detector derives required keys from catalog resources, maps each key to a convention-based path under a configurable dotfiles config base, checks existence only, and returns `planning.ConfigState` for `BuildPlan`.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Detector package | Create `internal/config` with `Detect(catalog)` and `Detector.Detect(catalog)`. | Extend `internal/state`; probe inside `internal/planning`. | Config presence is distinct from installation state, and planning must remain pure/caller-driven. |
| Filesystem seams | Inject `Exists PathExists` and `PathForKey KeyPathResolver`; default `Exists` uses `os.Stat`. | Use `os.Stat` directly; parse dotfiles metadata. | Seams keep tests deterministic and preserve the established adapter pattern. |
| Key path convention | Default base is `$HOME/.dotfiles/config`; key segments split on `.`: `go.env` → `$HOME/.dotfiles/config/go/env`. Reject absolute, empty, or `..`-escaping keys by treating them absent. | Add catalog `config_path`; understand dotfiles module layout. | No schema change, no dotfiles internals ownership, and the convention can be replaced through injection or a future slice. |
| CLI wiring | Add `detectConfigState = config.Detect`; call it after catalog load and before `BuildPlan`. | Keep `planning.ConfigState{}`; hide detection in catalog load. | The CLI already wires environment and installation state; catalog loading should stay a decode adapter, not host probing. |

## Data Flow

```text
catalog/bootstrap.toml ──→ catalogtoml.LoadFile ──→ planning.Catalog
                                      │
                                      ├─→ environment.Detect() ──→ EnvironmentFacts
                                      ├─→ state.Detect(catalog) ─→ InstallationState
                                      └─→ config.Detect(catalog) ─→ ConfigState

Catalog + facts + config + installation ──→ planning.BuildPlan ──→ renderer
```

Planning remains unchanged: `missingConfigReasons` reads only `ConfigState.PresentKeys` supplied by the caller.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/config/detector.go` | Create | Read-only config detector, default base/path convention, injected seams. |
| `internal/config/detector_test.go` | Create | Table tests for key collection, path mapping, missing/present state, invalid keys, and deterministic fixture seams. |
| `cmd/dbootstrap/main.go` | Modify | Import `internal/config`, add `detectConfigState`, and pass detected state to `BuildPlan`. |
| `cmd/dbootstrap/main_test.go` | Modify | Stub config detector and assert present config changes `runtime:go` from missing-config attention to planned/already-installed as applicable. |
| `internal/planning/builder_test.go` | Modify | Add/strengthen caller-supplied config-state cases if coverage is missing. |
| `README.md` | Modify | Optional current-status cleanup only; no new dotfiles ownership language. |

## Interfaces / Contracts

```go
type PathExists func(path string) bool
type KeyPathResolver func(basePath, key string) (string, bool)

type Detector struct {
    BasePath   string
    Exists     PathExists
    PathForKey KeyPathResolver
}

func Detect(catalog planning.Catalog) planning.ConfigState
func (d Detector) Detect(catalog planning.Catalog) planning.ConfigState
```

Only `ConfigPolicy.RequiredKeys` are inspected. The detector returns `PresentKeys` entries only for keys whose mapped path exists. Missing paths, invalid keys, and stat errors are absence, not diagnostics.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Config detector mapping and read-only behavior | Inject fake resolver/existence maps; assert stable `ConfigState` and no host dependency. |
| Unit | Planner purity | Existing `BuildPlan` tests continue proving planning only uses supplied state. |
| CLI | Composition-root wiring | Stub environment, installation, and config detectors; assert catalog-load failures skip detection and success forwards config state. |
| E2E | Real host dotfiles behavior | Out of scope for this slice. |

## Migration / Rollout

No migration required. This is read-only detection behind existing `ConfigState` semantics.

## Open Questions

- [ ] None.
