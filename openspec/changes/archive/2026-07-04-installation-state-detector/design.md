# Design: Installation State Detector

## Technical Approach

Extend planning with caller-supplied `InstallationState`, mirroring existing `EnvironmentFacts` and `ConfigState`: adapters probe, planning stays pure. `internal/state` will inspect catalog resources and mark `tool`/`runtime` refs present through an injectable PATH lookup seam. This slice updates the `BuildPlan` API and tests, but defers CLI detector wiring and user-facing rendering changes.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Planning state ownership | Add `planning.InstallationState{PresentResources map[ResourceRef]bool}` and `PlanStepStatusAlreadyInstalled = "already_installed"`. | Post-process `PlanResult`; separate `BuildPlanWithState`. | Status belongs in the domain planner; a second API would split semantics and invite drift. |
| Status precedence | Environment mismatch still skips; otherwise `already_installed` wins over `attention_required`, while missing-config reasons remain attached as metadata. | Make missing config produce `attention_required` even for present resources; drop attention metadata. | Presence means no install action is needed, so the primary status should be `already_installed`. Keeping reasons preserves visibility for follow-up setup without misclassifying the resource as installable work. |
| Detection seam | `internal/state.Detector{LookPath PathLookup}` where `PathLookup func(string) (string, error)` defaults to `exec.LookPath`. | Execute `--version`; inspect package managers. | PATH lookup is deterministic behind a seam and does not execute commands; version/health/package checks are out of scope. |
| Package layout | Create `internal/state` with detector and tests. | Put host probing in `internal/planning` or `internal/environment`. | Planning must stay pure; environment detects OS facts, while installation state is a separate adapter concern. |

## Data Flow

```text
catalog + env facts + config state + installation state
        └──────────────→ planning.BuildPlan ──→ PlanResult

future CLI wiring:
catalog ──→ state.Detector ──→ planning.InstallationState
```

Planning order: expand request → filter environment mismatches → topological order → compute missing config reasons → if ref is present, status `already_installed`; else `attention_required` when reasons exist, otherwise `planned`.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/planning/types.go` | Modify | Add `InstallationState` and `PlanStepStatusAlreadyInstalled`. |
| `internal/planning/builder.go` | Modify | Change `BuildPlan(catalog, request, facts, config, installation)` and apply state during status selection. |
| `internal/planning/builder_test.go` | Modify | Add already-installed, mixed state, missing-config precedence, env-skip precedence, and purity regression cases. |
| `internal/state/detector.go` | Create | Public detector API and PATH lookup implementation for `tool`/`runtime`. |
| `internal/state/detector_test.go` | Create | Host-independent seam tests. |
| `cmd/dbootstrap/main.go` | Modify | Mechanical signature update only: pass `planning.InstallationState{}`. No detector wiring. |
| `internal/catalog/toml/catalog_test.go` | Modify | Mechanical test call-site updates with empty installation state. |

## Interfaces / Contracts

```go
type InstallationState struct {
    PresentResources map[ResourceRef]bool
}

type PathLookup func(name string) (string, error)

type Detector struct { LookPath PathLookup }
func Detect(catalog planning.Catalog) planning.InstallationState
func (d Detector) Detect(catalog planning.Catalog) planning.InstallationState
```

`internal/state` maps `ResourceKindTool` and `ResourceKindRuntime` resource names directly to PATH lookup names. `package` and `dotfile` refs are ignored in this slice.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Planner status precedence and empty-state compatibility | Table-driven `internal/planning` tests; assert `Results`, `Steps`, reasons, and no mutation. |
| Unit | Detector seam | Inject fake `LookPath`; assert present/absent tool/runtime refs; never depend on host PATH. |
| Integration | Existing catalog fixture call sites | Update current tests to pass empty state and preserve output. |
| E2E | CLI state wiring/rendering | Out of scope; existing CLI tests should remain unchanged except signature fallout. |

## Migration / Rollout

No data migration required. This is an API-breaking internal change; update all `BuildPlan` call sites mechanically with `planning.InstallationState{}` until CLI wiring is proposed.

## Out of Scope

- CLI detector wiring, flags, and changed plan output.
- Dedicated rendering UX for `already_installed` beyond generic status printing.
- Package manager, dotfile, installer, command execution, version, or health checks.

## Risks / Tradeoffs

- PATH presence is only a readiness heuristic, not proof of correct version.
- `BuildPlan` signature churn touches tests and `cmd/dbootstrap`.
- Keeping attention metadata under `already_installed` is richer but requires tests so consumers do not assume reasons only appear with `attention_required`.

## Open Questions

- [ ] None.
