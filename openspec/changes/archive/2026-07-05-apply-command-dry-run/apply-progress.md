# Apply Progress: apply-command-dry-run

## Status

All tasks complete. Implementation ready for verify.

## Completed Tasks

- [x] 1.1 Extract a shared plan/apply request parser in `cmd/dbootstrap/main.go` for `--profile`, repeatable `--resource`, and `--catalog` validation.
- [x] 1.2 Add a shared build-plan helper so `plan` and `apply` reuse the same catalog load, detection, and `planning.BuildPlan()` path.
- [x] 1.3 Add `execution.NoopForKind(kind)` in `internal/execution/noop.go` for kind-aware noop dispatch without side effects.
- [x] 2.1 Wire `case "apply"` in `cmd/dbootstrap/main.go` to `runApply()` and add apply-specific usage text.
- [x] 2.2 Implement `runApply()` to execute the shared build-plan helper, short-circuit on planning errors, and run the plan through noop execution only.
- [x] 2.3 Add `renderExecutionReport()` in `cmd/dbootstrap/render.go` with an explicit execution heading and `not_implemented` outcomes.
- [x] 2.4 Ensure apply remains non-mutating and does not introduce retry, concurrency, dotlink, clone, or sparse-checkout behavior.
- [x] 3.1 Update CLI usage and command help so `apply` mirrors `plan` target flags and validation messages.
- [x] 3.2 Replace the obsolete no-apply boundary with functional apply dry-run coverage in `internal/execution/regression_test.go`.
- [x] 3.3 Update OpenSpec artifacts so execution-contracts no-apply wording is removed/replaced consistently.
- [x] 4.1 Add CLI tests in `cmd/dbootstrap/main_test.go` for apply success, invalid targets, catalog failures, and planning-error short circuiting.
- [x] 4.2 Add renderer tests in `cmd/dbootstrap/render_test.go` to prove execution output is distinct from plan rendering.
- [x] 4.3 Add execution tests in `internal/execution/noop_test.go` for `NoopForKind` dispatch and `not_implemented` reporting.
- [x] 4.4 Verify no real execution, host mutation, dotlink, clone, retry, or concurrency is reachable from the apply path.
- [x] 5.1 Remove or replace `TestNoApplyCommandInCLI` so the regression matches the new dry-run apply contract.
- [x] 5.2 Refresh comments and usage strings to clearly label apply as dry-run-only and execution-report oriented.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `cmd/dbootstrap/main.go` | Modified | Added `apply` command wiring, shared `parsePlanFlags`/`buildPlan` helpers, `runApply()`, and apply usage text. |
| `cmd/dbootstrap/render.go` | Modified | Added `renderExecutionReport()` with explicit "Execution Report" heading and `not_implemented` output. |
| `cmd/dbootstrap/main_test.go` | Modified | Added apply CLI tests; updated usage-error expectations to include `apply`; replaced unknown command example. |
| `cmd/dbootstrap/render_test.go` | Modified | Added execution report rendering tests separate from plan rendering. |
| `internal/execution/noop.go` | Modified | Added `NoopForKind(kind)` kind-aware noop installer wrapper. |
| `internal/execution/noop_test.go` | Modified | Added `NoopForKind` dispatch and `not_implemented` reporting coverage. |
| `internal/execution/regression_test.go` | Modified | Removed `TestNoApplyCommandInCLI`; added `TestNoopExecutionRemainsNonMutating`. |
| `openspec/specs/execution-contracts/spec.md` | Modified | Replaced no-apply requirement with apply uses noop execution contracts only. |
| `openspec/changes/apply-command-dry-run/tasks.md` | Modified | Marked all tasks complete. |
| `openspec/changes/apply-command-dry-run/apply-progress.md` | Created | This apply-progress artifact. |

## Deviations from Design

None — implementation matches design.

## Issues Found

None.

## Tests Run

- `go test ./...` — all packages pass.
- `go vet ./...` — clean.

## Workload / PR Boundary

- Mode: single PR with maintainer-approved size exception.
- Current work unit: full apply-command-dry-run slice.
- Boundary: complete dry-run apply command, rendering, tests, and regression cleanup.
- Estimated review budget impact: ~334 insertions / ~60 deletions across 8 files; within accepted size-exception scope.

## Next Recommended

`sdd-verify`
