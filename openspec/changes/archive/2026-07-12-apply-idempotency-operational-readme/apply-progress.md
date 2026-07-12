# Apply Progress: Apply Idempotency and Operational README

## Status

Completed all 12 implementation tasks. The authoritative OpenSpec status consumed before work was `applyState: ready`, `nextRecommended: apply`, with repo-local action context rooted at `/home/dniebles/dniebles-bootstrap`.

## Completed tasks

- Marked tasks 1.1 through 4.2 as complete in `tasks.md`.
- Added `PlanStep.Status` and propagated the computed planning status into ordered steps.
- Constrained command presence detection to tool/runtime resources with non-nil `Presence`, `Kind == "command_exists"`, and non-empty `Name`; detection now looks up `Presence.Name`.
- Added the runner skip guard requiring `already_installed` plus the exact valid tool/runtime presence metadata. Eligible steps produce `skipped` / `unchanged` with `already installed; no mutation attempted` without installer dispatch.
- Kept default and dry-run apply reports on their existing non-mutating path; confirmed apply/bootstrap share the direct skip behavior.
- Added focused detector, planning, runner, and apply/bootstrap coverage, and updated the operational README.

## Files changed

- `README.md`
- `cmd/dbootstrap/main.go`
- `cmd/dbootstrap/main_test.go`
- `internal/execution/runner.go`
- `internal/execution/runner_test.go`
- `internal/planning/builder.go`
- `internal/planning/builder_test.go`
- `internal/planning/types.go`
- `internal/state/detector.go`
- `internal/state/detector_test.go`

## TDD Cycle Evidence

| Cycle | RED evidence | GREEN / verification |
|---|---|---|
| Detector | `go test ./internal/state` failed: configured `vim` presence target was not recognized because the prior detector looked up the resource ID. | Constrained metadata guard and `Presence.Name` lookup; `go test ./internal/state` passed. |
| Planning | `go test ./internal/planning` failed to compile because `PlanStep.Status` did not exist. | Added and propagated `PlanStep.Status`; `go test ./internal/planning` passed. |
| Runner | `go test ./internal/execution` failed to compile because the step status field did not exist. | Added exact status/kind/presence guard and direct skipped result; `go test ./internal/execution` passed. |
| Command modes | `go test ./cmd/dbootstrap` failed because default and dry-run rendered the confirmed skip wording. | Cleared execution-only status in non-confirmed modes, preserving no-op/dry-run output; `go test ./cmd/dbootstrap` passed. |
| Triangulate | Focused suites passed together. | `go test ./...` passed; `gofmt -l` on changed Go files produced no output; `git diff --check` passed. |

## Commands run

- `go test ./internal/state` — RED fail, then pass.
- `go test ./internal/planning` — RED fail, then pass.
- `go test ./internal/execution` — RED fail, then pass.
- `go test ./cmd/dbootstrap` — RED fail for safe-mode regression, then pass.
- `go test ./internal/state && go test ./internal/planning && go test ./internal/execution && go test ./cmd/dbootstrap` — pass.
- `go test ./...` — pass.
- `gofmt -l` on all changed Go files — no output.
- `git diff --check` — pass.

## Scope, workload, and deviations

- No deviation from approved behavior. `cmd/dbootstrap/main.go` was changed only to preserve the approved default/dry-run behavior while the runner owns confirmed-mode skip handling.
- Changed paths are the nine approved implementation/test/documentation paths plus the permitted apply-path wiring file `cmd/dbootstrap/main.go` listed by the design for the shared seam.
- Diff workload: 255 changed lines (241 additions, 14 deletions), below the 400-line single-PR budget.
- No external command integration was skipped beyond the injected command-runner seams already used by focused tests.

## Remaining tasks

None. All persisted implementation task checkboxes are marked `- [x]`.

## Next action

Run independent SDD verification, then start or validate the bounded review receipt before commit/PR.
