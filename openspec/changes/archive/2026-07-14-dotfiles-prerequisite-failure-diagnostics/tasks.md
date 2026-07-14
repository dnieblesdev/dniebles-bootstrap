# Tasks: Dotfiles Prerequisite Failure Diagnostics

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 220–320 additions+deletions |
| 800-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | One work unit / single PR |
| Delivery strategy | auto-forecast |
| Chain strategy | pending |

### Suggested Work Units

| Unit | Goal | Likely PR | Focused test command | Runtime harness | Rollback boundary |
|---|---|---|---|---|---|
| 1 | Add and prove end-to-end truthful dotfiles diagnostics | Single PR | `go test ./internal/execution ./cmd/dbootstrap` | `go run ./cmd/dbootstrap apply --yes --resource dotfile:bash` with isolated missing-runner fixture | Revert the listed execution/renderer changes and tests only |

## Phase 1: Typed Diagnostic Contract (strict TDD)

- [x] 1.1 RED: In `internal/execution/types_test.go`, specify typed `DotfilesPrerequisiteTarget`, stable phases, and `DotfilesFailure.Unwrap()` preserving independent prerequisite, execution, and parse causes through `errors.Is`/`errors.As`.
- [x] 1.2 GREEN: In `internal/execution/types.go`, add the smallest target/phase/cause carrier and multi-error unwrap; preserve existing statuses and command/report fields.

## Phase 2: Prerequisite and Execution Transport (strict TDD)

- [x] 2.1 RED: In `internal/execution/dotfiles_provider_test.go`, add table cases for resolution rejection, missing runner, missing module, and escaping module; assert original attempted candidate, safe typed cause, stable phase, zero runner calls, and no inferred links.
- [x] 2.2 GREEN: In `internal/execution/dotfiles_provider.go`, capture runner/module candidates before validation and return typed failures without invoking the runner or claiming rejected paths are canonical.
- [x] 2.3 RED: In `internal/execution/dotfiles_installer_test.go`, cover valid failed reports, absent/malformed/contradictory stdout, typed cause classification, stable execution/report-validation phases, and unchanged command semantics.
- [x] 2.4 GREEN: Verify the existing installer translation retains typed causes and phase-specific failure context while translating prerequisite rejection to a failed result with no links; no installer production change is required.

## Phase 3: Safe Rendering and Acceptance

- [x] 3.1 RED: In `cmd/dbootstrap/render_test.go`, specify operation/modules/phase/attempted runner-or-module target, curated causes, bounded terminal sanitization, deterministic equal-snapshot dedup, distinct-target retention, and no raw wrapped error text.
- [x] 3.2 GREEN: In `cmd/dbootstrap/render.go`, render curated typed causes and phase labels once, retain distinct candidates/executables, and preserve bounded stderr sanitization.
- [x] 3.3 RED/GREEN: In `cmd/dbootstrap/main_test.go`, prove confirmed missing-`bin/dotlink` apply is non-zero, reports `dotfile:bash`/`bash`, prerequisite/repository-validation, unvalidated candidate, missing-path cause, and zero runner calls; preserve existing regression contracts.

## Phase 4: Verification

- [x] 4.1 Run `go test ./internal/execution ./cmd/dbootstrap`, then `go test ./...`, `go vet ./...`, `go build ./...`, and `gofmt` on changed Go files; record results.
- [x] 4.2 Record the isolated runtime harness result and confirm scope excludes legacy provider, monolith cleanup, and `PlanStep.AttentionReasons -> StepResult`.
