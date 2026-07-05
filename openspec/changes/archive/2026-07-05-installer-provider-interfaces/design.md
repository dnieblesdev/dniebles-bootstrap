# Design: Installer Provider Interfaces

## Technical Approach

Add a new `internal/execution` package that defines execution contracts beside, not inside, the pure planning domain. The package will accept `planning.Plan`/`planning.PlanStep` as input data, dispatch steps sequentially by `planning.ResourceKind`, and return execution-only results. This slice is contracts-only: noop implementations compile and test the seams while performing no command execution, dotlink invocation, filesystem mutation, concurrency, retries, or CLI wiring.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|----------|--------|--------------------------|-----------|
| Package boundary | Create one `internal/execution` package. | Split into `internal/runner`, `internal/installer`, `internal/provider`; place contracts in `internal/planning`. | One package matches the existing focused `internal/planning` style and keeps review load low. Planning must remain pure and unchanged. |
| Status model | Define execution statuses separately from `planning.PlanStepStatus`. | Reuse planning statuses. | Planning statuses describe intended work; execution statuses describe runtime outcome. Separate Go types prevent semantic drift, especially around `skipped`. |
| Runner behavior | Use a concrete sequential `Runner` with kind-to-installer dispatch. | Add concurrency/retry/options now; return a runner interface. | Sequential behavior is enough for the current contract slice. Interfaces can be extracted when there is a second runner implementation. |
| Noop behavior | Provide noop installer/provider stubs returning `not_implemented` results or nil-safe no-op errors. | Panic, return generic failure, or omit stubs. | Noops prove the contracts are safe and non-mutating while making future apply wiring explicit. |
| Dotfiles execution | Keep `DotfilesProvider` separate from the existing read-only `internal/dotfiles.Detector`. | Replace detector or model dotfiles as only an `Installer`. | Detection reports local availability; provider operations are future mutation-capable execution seams and must stay isolated. |

## Data Flow

```text
planning.Plan ──→ execution.Runner
                     │
                     ├─ step.Ref.Kind ──→ Installer.Install(ctx, step)
                     │
                     └─ append StepResult ──→ ExecutionReport
```

The runner reads plan steps in order and never changes `internal/planning` production code or plan data.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/execution/types.go` | Create | Define `StepStatus`, `StepResult`, and `ExecutionReport`. |
| `internal/execution/installer.go` | Create | Define small `Installer` contract scoped by `planning.ResourceKind`. |
| `internal/execution/runner.go` | Create | Implement sequential runner and kind-based installer lookup. |
| `internal/execution/provider.go` | Create | Define high-level `DotfilesProvider` execution contract. |
| `internal/execution/noop.go` | Create | Add safe noop installer/provider helpers returning `not_implemented` without mutation. |
| `internal/execution/*_test.go` | Create | Add table-driven tests for statuses, noops, sequential dispatch, missing installers, and provider stubs. |
| `internal/planning/*` | Unchanged | Referenced only as input data. Production files stay unchanged. |
| `cmd/dbootstrap/*` | Unchanged | No apply command or CLI wiring in this slice. |

## Interfaces / Contracts

```go
type StepStatus string

const (
    StepStatusInstalled      StepStatus = "installed"
    StepStatusFailed         StepStatus = "failed"
    StepStatusSkipped        StepStatus = "skipped"
    StepStatusNotImplemented StepStatus = "not_implemented"
)

type StepResult struct {
    Ref     planning.ResourceRef
    Status  StepStatus
    Message string
    Err     error
}

type ExecutionReport struct { Results []StepResult }

type Installer interface {
    SupportedKind() planning.ResourceKind
    Install(context.Context, planning.PlanStep) StepResult
}

type DotfilesProvider interface {
    EnsureModules(context.Context, []string) error
    RunDotlink(context.Context, []string) error
}
```

`Runner.Run(ctx, plan)` should stop neither on `not_implemented` nor failed status unless a later slice explicitly adds failure policy. Missing installer returns a `not_implemented` result for that step.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Execution status constants and result shape | Table tests in `internal/execution`. |
| Unit | Noop installer/provider safety | Assert `not_implemented` results and no mutation-capable seams are called. |
| Unit | Runner sequential dispatch | Fake installers record call order and received `PlanStep` values. |
| Unit | Missing installer behavior | Plan step with unsupported kind yields `not_implemented`. |
| Regression | Planning boundary | No production changes under `internal/planning`; existing planning tests remain authoritative. |

## Migration / Rollout

No migration required. The change is additive and not wired to the CLI. Rollback removes `internal/execution/` and its tests.

## Open Questions

- [x] None.
