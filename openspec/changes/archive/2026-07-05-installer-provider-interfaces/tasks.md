# Tasks: Installer Provider Interfaces

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 250-380 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | Single PR with size exception |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Add execution contracts and noop seams | PR 1 | Base on main; keep tests with contracts |
| 2 | Add runner dispatch and regression coverage | PR 1 | Same PR; sequential dispatch, missing installer, status vocabulary |
| 3 | Lock down planning/CLI boundaries and artifact updates | PR 1 | Verify no planning production or CLI apply wiring changes |

## Phase 1: Foundation / Contracts

- [x] 1.1 Create `internal/execution/types.go` with execution-only `StepStatus`, `StepResult`, and `ExecutionReport` types.
- [x] 1.2 Create `internal/execution/installer.go` with the `Installer` interface keyed by `planning.ResourceKind`.
- [x] 1.3 Create `internal/execution/provider.go` with the `DotfilesProvider` interface and execution-only method set.

## Phase 2: Core Implementation

- [x] 2.1 Create `internal/execution/runner.go` with sequential `Runner` dispatch over `planning.Plan` steps.
- [x] 2.2 Add missing-installer handling that returns `not_implemented` without stopping later steps.
- [x] 2.3 Create `internal/execution/noop.go` with safe noop installer/provider stubs that do not mutate state.

## Phase 3: Testing / Verification

- [x] 3.1 Add table-driven tests for execution status vocabulary and result/report shape in `internal/execution/*_test.go`.
- [x] 3.2 Add tests proving noop installer/provider paths return `not_implemented` and perform no mutation-capable actions.
- [x] 3.3 Add runner tests proving sequential dispatch, kind-based installer selection, and missing-installer fallback.
- [x] 3.4 Add regression checks that planning production code and CLI surface remain unchanged, including no `apply` wiring.

## Phase 4: Cleanup / Artifact Updates

- [x] 4.1 Update OpenSpec artifact files so proposal/spec/design remain aligned with the implemented execution contract slice.
- [x] 4.2 Run formatting and package tests for `internal/execution` and adjacent planning/CLI regression scope.
