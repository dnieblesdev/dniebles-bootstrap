# Proposal: Apply Command Dry Run

## Intent

Add a safe `apply` CLI bridge that proves the existing planning-to-execution path end-to-end without mutating the host. Users should see execution-style results from the current plan using noop execution contracts.

## Scope

### In Scope
- Add `dbootstrap apply` inline in `cmd/dbootstrap/main.go`, following the existing `plan` command pattern.
- Support `--profile`, `--resource`, and `--catalog` with the same planning validation behavior as `plan`.
- Run planning first; planning errors fail before any execution report is produced.
- Execute the resulting plan with noop execution dependencies and render a distinct execution report.
- Remove or replace obsolete `TestNoApplyCommandInCLI`.
- Add `execution.NoopForKind(kind)` if kind-aware noop dispatch is needed.

### Out of Scope
- Real installers, real command execution, host mutation, dotlink invocation, clone/sparse checkout.
- Retry, concurrency, or installer orchestration beyond the existing sequential runner.
- Adding `--dry-run`; this command is dry-run-only for this slice.

## Capabilities

### New Capabilities
- `apply-command-dry-run`: CLI `apply` behavior that reuses planning and produces noop execution reports safely.

### Modified Capabilities
- `execution-contracts`: remove the previous no-apply boundary and clarify noop/kind-aware execution support remains non-mutating.

## Approach

Wire `case "apply"` to `runApply()` in `cmd/dbootstrap/main.go`. `runApply()` reuses catalog loading, detector calls, and `planning.BuildPlan()`, then constructs `execution.Runner` with noop installers per resource kind, runs the plan, and calls `renderExecutionReport()`. Keep planning and execution status vocabularies visually distinct.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modified | Add `apply`, usage text, and `runApply()` composition. |
| `cmd/dbootstrap/render.go` | Modified | Render execution reports separately from plan output. |
| `cmd/dbootstrap/*_test.go` | Modified | Add apply CLI/render coverage and update usage/error expectations. |
| `internal/execution/noop.go` | Modified | Optional `NoopForKind` helper for safe dispatch. |
| `internal/execution/regression_test.go` | Modified | Remove/replace obsolete no-apply regression. |
| `openspec/specs/execution-contracts/spec.md` | Modified | Remove no-apply requirement during archive. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Users confuse dry-run apply with real apply | Medium | Output must state noop/not implemented execution clearly. |
| Planning vs execution statuses blur | Medium | Use explicit “Execution Report” rendering. |
| Obsolete regression removal looks unsafe | Low | Replace with functional apply command tests. |

## Rollback Plan

Remove the `apply` switch case, `runApply()`, execution rendering, apply tests, and any `NoopForKind` helper; restore the no-apply regression requirement if this slice is reverted.

## Dependencies

- Existing `internal/planning.BuildPlan()` pipeline.
- Existing `internal/execution.Runner`, noop contracts, and report types.

## Success Criteria

- [ ] `dbootstrap apply` accepts the same target/catalog flags as `plan`.
- [ ] Planning failures exit before execution reporting.
- [ ] Successful apply dry-run renders execution results with `not_implemented` outcomes and no mutation.
- [ ] Obsolete no-apply regression is removed or replaced.
