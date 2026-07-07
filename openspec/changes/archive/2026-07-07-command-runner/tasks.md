# Tasks: Command Runner

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 220-360 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | exception-ok |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Add internal command runner contracts and implementations | PR 1 | Base: main; keep shell strings rejected and dry-run deterministic |
| 2 | Add safe execution tests and regression checks | PR 1 | Base: main; use helper process/fake seam, no host mutation |

## Phase 1: Foundation / Contracts

- [x] 1.1 Create `internal/execution/command.go` with `CommandRequest`, `CommandResult`, `CommandStatus`, `CommandRunner`, and validation for empty executable / shell-first rejection.
- [x] 1.2 Keep command status types in `internal/execution/command.go`; do not expand `internal/execution/types.go` unless a shared result type is required.

## Phase 2: Core Implementation

- [x] 2.1 Create `internal/execution/os_command_runner.go` to run explicit executable-plus-args commands with optional cwd/env, context/timeout cancellation, stdout/stderr capture, exit-code extraction, and duration measurement.
- [x] 2.2 Create `internal/execution/noop_command_runner.go` to return a deterministic not-run result without starting a process or mutating host state.
- [x] 2.3 Keep `cmd/dbootstrap/main.go` on existing noop-only `apply` wiring; do not connect `CommandRunner` to bootstrap or installer dispatch in this slice.

## Phase 3: Testing / Verification

- [x] 3.1 Add `internal/execution/command_runner_test.go` table tests for explicit command success, non-zero exit, stderr capture, validation rejection, and deterministic noop output.
- [x] 3.2 Verify timeout and external cancellation with a Go helper process or fake seam; avoid `/bin/sh`, package managers, and host mutation.
- [x] 3.3 Add a regression assertion that `apply` remains noop-only and `cmd/dbootstrap/main.go` stays unwired from `CommandRunner`.

## Phase 4: Cleanup / Documentation

- [x] 4.1 Update comments/docstrings in `internal/execution/*` to reflect argv-only execution and shell-first rejection.
- [x] 4.2 Confirm `catalog/bootstrap.toml` and planning types remain unchanged and contain no raw command fields.
