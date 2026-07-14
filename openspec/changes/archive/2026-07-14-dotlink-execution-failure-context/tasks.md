# Tasks: Dotlink Execution Failure Context

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 703 total, including 5 OpenSpec artifacts (proposal, 2 specs, design, tasks) in the forecast |
| 800-line budget risk | Low |
| 400-line budget risk | Low (the approved hard budget is 800) |
| Chained PRs recommended | No |
| Suggested split | One sequential independent target / single PR |
| Delivery strategy | single-pr |
| Chain strategy | size:exception (accepted) |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size:exception
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Focused test command | Runtime harness | Rollback boundary |
|---|---|---|---|---|---|
| 1 | Failure transport, provider composition, installer retention, renderer presentation | Single PR | `go test ./internal/execution ./cmd/dbootstrap` | `go test ./...` (no separate runtime harness; command runner is mocked) | Revert the listed implementation/tests and preserve merged base-resolution code |

## Phase 1: Contracts and Structural RED/GREEN

- [x] 1.1 RED: In `internal/execution/dotfiles_provider_test.go`, add table-driven tests for `DotfilesFailure` fields, nil-filtered multi-unwrap, joined concrete `*exec.ExitError`/`*json.SyntaxError`, canonical executable, missing-runner no-call, and the four command/report compositions.
- [x] 1.2 GREEN: In `internal/execution/types.go`, add `DotfilesFailure`, `StepResult.DotfilesFailure`, and `Unwrap()`; preserve `BaseDiagnostic` as primary and keep messages base-free.
- [x] 1.3 RED: Add parser-error identity coverage in `internal/execution/dotlink_report_test.go` (or the existing report test file) for concrete malformed stdout errors.
- [x] 1.4 GREEN: In `internal/execution/dotlink_report.go`, preserve concrete parser errors while retaining safe invalid-report classification.

## Phase 2: Provider Execution and Composition

- [x] 2.1 RED: Add bounded Unicode/control/terminal-escape stderr cases, exit-code assertions, and stdout-only report-source tests in `internal/execution/dotfiles_provider_test.go`.
- [x] 2.2 GREEN: In `internal/execution/dotfiles_provider.go`, derive `<canonical>/bin/dotlink`, create canonical execution failures, sanitize stderr to ≤4096 bytes without split runes/escape tokens, and compose execution/parser causes for unavailable, invalid, inconsistent, and valid-failed reports.

## Phase 3: Installer Transport and Presentation RED/GREEN

- [x] 3.1 RED: In `internal/execution/dotfiles_installer_test.go`, assert a failed `StepResult` retains the valid failed report, execution error, structured failure, unchanged base identity, and short message without base text.
- [x] 3.2 GREEN: In `internal/execution/dotfiles_installer.go`, translate report-plus-error outcomes, retain both, and stop appending `baseContext` to `StepResult.Message`.
- [x] 3.3 RED: In `cmd/dbootstrap/render_test.go`, separately test structural fields versus presentation: identical snapshots deduplicate semantically; differing snapshots receive explicit labels; execution facts are labeled and controls absent.
- [x] 3.4 GREEN: In `cmd/dbootstrap/render.go`, render structured executable/runner/command/exit/stderr facts and compare base snapshots by fields and ordered modules, never formatted text.

## Phase 4: Regression Verification

- [x] 4.1 RED/GREEN: Extend focused regressions for success, default, dry-run, report validation, and merged base identity; confirm no first-target base resolution or out-of-scope domain changes.
- [x] 4.2 Run `gofmt`, focused package tests, `go vet ./...`, and `go test ./...`; inspect the diff against the 703-line estimate and verify the five OpenSpec artifacts are counted.
