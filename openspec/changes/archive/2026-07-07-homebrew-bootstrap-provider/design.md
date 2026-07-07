# Design: Homebrew Bootstrap Provider

## Technical Approach

Add a non-mutating Homebrew bootstrap reporter in `internal/execution` that inspects the already-built `planning.Plan` for resources whose structured install metadata has `provider = "brew"`. If such resources exist and `brew` is not present, `apply` appends a manual bootstrap action to the execution report. The action displays the official install command as human guidance only; no code path executes it in default, `--dry-run`, or `--yes` modes.

This design maps to the active delta spec set split by capability: `homebrew-bootstrap-provider` adds missing-Homebrew bootstrap detection and manual guidance, `apply-command-dry-run` keeps apply reporting non-mutating across default, `--dry-run`, and `--yes`, and `execution-contracts` keeps bootstrap data advisory within noop execution contracts.

## Architecture Decisions

| Option | Tradeoff | Decision |
|---|---|---|
| Extend `ExecutionReport` with manual actions | Keeps plan steps and provider guidance in one apply output; requires renderer/test updates | Chosen: add a small `ManualAction`/`ManualInstruction` report model |
| Implement Homebrew as an `Installer` | Fits `Runner`, but implies install execution and confuses bootstrap with package installation | Rejected: this slice reports bootstrap only |
| Use `CommandRunner` or shell command for detection | Reuses execution abstractions, but risks hidden process execution | Rejected: use a safe `CommandExists` seam backed by `exec.LookPath` only |
| Add raw command/catalog install fields | Easy to render, but violates structured metadata and shell-safety specs | Rejected: inspect existing `Install.Provider` only |

## Data Flow

```text
catalog TOML -> planning.BuildPlan -> Runner(noop installers) -> ExecutionReport
                                      plan metadata -> Homebrew reporter ┘
                                                          |
                                                     render manual action
```

The Homebrew reporter reads `PlanStep.Resource.Install.Provider`; it does not alter planning, target package installation, or `Runner` dispatch.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/execution/types.go` | Modify | Add manual action fields to `ExecutionReport` without changing `StepResult` semantics. |
| `internal/execution/homebrew_bootstrap.go` | Create | Detect brew-backed plan needs and produce a manual bootstrap action when `brew` is absent. |
| `internal/execution/homebrew_bootstrap_test.go` | Create | Unit tests for no brew-backed resources, brew present, brew missing, and command text not executed. |
| `cmd/dbootstrap/main.go` | Modify | After noop `Runner.Run`, enrich the report with Homebrew manual actions before rendering. |
| `cmd/dbootstrap/render.go` | Modify | Render a clear `Manual Actions` section with official Homebrew instruction and non-mutating wording. |
| `cmd/dbootstrap/main_test.go` | Modify | Assert default, `--dry-run`, and `--yes` remain noop while reporting manual bootstrap when applicable. |
| `cmd/dbootstrap/render_test.go` | Modify | Cover manual action rendering separately from step rendering. |
| `catalog/bootstrap.toml` | Keep | No change until the catalog supports multi-provider metadata; tests use fixtures for brew-backed resources. |

## Interfaces / Contracts

```go
type ManualAction struct {
    ID           string
    Title        string
    Reason       string
    Instructions []string
}

type CommandExists func(name string) bool

func AppendHomebrewBootstrap(report ExecutionReport, plan planning.Plan, exists CommandExists) ExecutionReport
```

`AppendHomebrewBootstrap` MUST only read plan metadata and call `exists("brew")`. The official Homebrew command is a string inside `Instructions`; it MUST NOT be converted to `CommandRequest`, `Installer`, or `CommandRunner` input.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Homebrew reporter behavior | Table tests with synthetic plans and fake `CommandExists`. |
| Unit | Renderer output | Snapshot-style string assertions for `Manual Actions`. |
| CLI integration | Apply safety modes | `run()` tests using fixture brew catalog; assert noop step statuses and manual instruction in all accepted modes. |
| Regression | No mutation wiring | Extend existing regression checks to forbid Homebrew command execution through `CommandRunner` or shell strings. |

## Migration / Rollout

No migration required. This is additive reporting only and can be removed without changing catalog data or host state.

## Open Questions

- [ ] None.
