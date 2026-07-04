# Design: Wire Installation State into CLI Plan

## Technical Approach

Wire the existing `internal/state` detector into `cmd/dbootstrap` as composition-root glue. `runPlan` will load the catalog, detect installation state from that catalog, then pass the detected `planning.InstallationState` to `planning.BuildPlan`. Planning remains the source of status semantics and `render.go` remains mechanical because it already prints arbitrary `PlanStepStatus` values, including `already_installed`.

## Architecture Decisions

| Option | Tradeoff | Decision |
|---|---|---|
| Package-level seam `detectInstallationState = state.Detect` | Matches `detectEnvironmentFacts`; adds one package variable. | Use it so CLI tests stay host-independent. |
| Direct `state.Detect(catalog)` call | Smaller production diff but tests depend on host PATH. | Reject. Deterministic tests matter more. |
| Change detector/planner/renderer contracts | Could model future detector errors or custom display. | Reject for this slice; existing contracts already support the behavior. |

## Data Flow

```text
runPlan
  ├─ catalogtoml.LoadFile(path)
  ├─ detectEnvironmentFacts()
  ├─ detectInstallationState(catalog)
  ├─ planning.BuildPlan(catalog, request, facts, ConfigState{}, installation)
  └─ renderPlanResult/renderDiagnostics
```

Detection MUST happen only after successful catalog load and before `BuildPlan`. Catalog load failures return the current load error and MUST NOT attempt installation-state detection.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modify | Import `internal/state`, add `detectInstallationState`, call it after catalog load, pass result to `BuildPlan`. |
| `cmd/dbootstrap/main_test.go` | Modify | Add `stubInstallationState`; stub empty state in existing plan tests; add present-resource case asserting `already_installed`. |
| `cmd/dbootstrap/render.go` | Unchanged | Existing `[status]` rendering covers `already_installed`. |
| `internal/state/*`, `internal/planning/*` | Unchanged | Detector and planner already expose the needed contracts. |

## Interfaces / Contracts

```go
var detectInstallationState = state.Detect

func stubInstallationState(t *testing.T, installation planning.InstallationState) {
    t.Helper()
    original := detectInstallationState
    detectInstallationState = func(planning.Catalog) planning.InstallationState {
        return installation
    }
    t.Cleanup(func() { detectInstallationState = original })
}
```

`state.Detect` currently returns `planning.InstallationState` and no error. Lookup failures are treated as absent resources inside the detector, not as CLI errors. Therefore this slice adds no detector-error branch. If a future detector contract returns `(InstallationState, error)`, CLI behavior should be non-fatal: write a diagnostic and continue with empty state. Do not change detector signature in this slice.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Existing successful plan output remains stable with empty state. | Stub environment facts and installation state to `planning.InstallationState{}`. |
| Unit | Present tool/runtime renders `already_installed`. | Stub `PresentResources` for catalog refs such as `tool:git`; assert stdout step and result status. |
| Unit | Catalog load failure skips detection. | Existing load-error tests remain; optional seam panic guard may prove detector is not called. |
| Integration/E2E | None. | No real PATH or host probing in CLI tests; `internal/state` already covers detector seams. |

## Migration / Rollout

No migration required. This is read-only CLI wiring with no persisted data and no feature flag.

## Out of Scope

- Renderer formatting changes, icons, colors, or TUI behavior.
- Package-manager, dotfile, installer, apply command, or command-runner checks.
- Planner status rules or detector lookup behavior.
- Introducing detector errors solely for this CLI slice.

## Open Questions

- [ ] The spec mentions detector error behavior, but the current detector cannot return errors; confirm the future-compatible diagnostic rule is sufficient without changing `internal/state` now.
