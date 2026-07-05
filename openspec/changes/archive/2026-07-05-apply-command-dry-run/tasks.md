# Tasks: Apply Command Dry Run

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 500-700 |
| 400-line budget risk | High |
| Chained PRs recommended | No |
| Suggested split | Single PR with maintainer-approved size exception |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Shared plan/apply target parsing and build-plan helper | PR 1 | Base CLI planning path; keep flag validation parity. |
| 2 | Apply command wiring and execution report rendering | PR 1 | Add `apply`, `NoopForKind`, and distinct execution output. |
| 3 | Test and regression cleanup | PR 1 | Replace obsolete no-apply regression and cover non-mutation. |

## Phase 1: Foundation / Infrastructure

- [x] 1.1 Extract a shared plan/apply request parser in `cmd/dbootstrap/main.go` for `--profile`, repeatable `--resource`, and `--catalog` validation.
- [x] 1.2 Add a shared build-plan helper so `plan` and `apply` reuse the same catalog load, detection, and `planning.BuildPlan()` path.
- [x] 1.3 Add `execution.NoopForKind(kind)` in `internal/execution/noop.go` for kind-aware noop dispatch without side effects.

## Phase 2: Core Implementation

- [x] 2.1 Wire `case "apply"` in `cmd/dbootstrap/main.go` to `runApply()` and add apply-specific usage text.
- [x] 2.2 Implement `runApply()` to execute the shared build-plan helper, short-circuit on planning errors, and run the plan through noop execution only.
- [x] 2.3 Add `renderExecutionReport()` in `cmd/dbootstrap/render.go` with an explicit execution heading and `not_implemented` outcomes.
- [x] 2.4 Ensure apply remains non-mutating and does not introduce retry, concurrency, dotlink, clone, or sparse-checkout behavior.

## Phase 3: Integration / Wiring

- [x] 3.1 Update CLI usage and command help so `apply` mirrors `plan` target flags and validation messages.
- [x] 3.2 Replace the obsolete no-apply boundary with functional apply dry-run coverage in `internal/execution/regression_test.go`.
- [x] 3.3 Update OpenSpec artifacts during archive so execution-contracts no-apply wording is removed/replaced consistently.

## Phase 4: Testing / Verification

- [x] 4.1 Add CLI tests in `cmd/dbootstrap/main_test.go` for apply success, invalid targets, catalog failures, and planning-error short circuiting.
- [x] 4.2 Add renderer tests in `cmd/dbootstrap/render_test.go` to prove execution output is distinct from plan rendering.
- [x] 4.3 Add execution tests in `internal/execution/noop_test.go` for `NoopForKind` dispatch and `not_implemented` reporting.
- [x] 4.4 Verify no real execution, host mutation, dotlink, clone, retry, or concurrency is reachable from the apply path.

## Phase 5: Cleanup / Documentation

- [x] 5.1 Remove or replace `TestNoApplyCommandInCLI` so the regression matches the new dry-run apply contract.
- [x] 5.2 Refresh comments and usage strings to clearly label apply as dry-run-only and execution-report oriented.
