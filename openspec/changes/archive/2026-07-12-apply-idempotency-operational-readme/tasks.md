# Tasks: Apply Idempotency and Operational README

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 220–340 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR: detector/planning/execution behavior, focused tests, and README guidance |
| Delivery strategy | single-pr |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

The change touches nine implementation/test/docs files and requires strict-TDD evidence, but it has no migration, generated artifact, new dependency, or cross-cutting schema work. Keep each behavior's tests with its code and deliver as one rollback-safe PR. Reassess the line count before apply if command-level fixtures or README replacement expands materially.

## Scope Guard

Eligibility for the idempotency skip is deliberately exact: **only** a `tool` or `runtime` resource with non-nil presence metadata where `Presence.Kind == "command_exists"` and `Presence.Name` is non-empty, and whose detector lookup succeeds. Every other step retains current runner behavior, including all `package` and `dotfile` resources, missing/empty/other presence metadata, non-`already_installed` statuses, unsupported steps, and failures.

## Ordered Implementation Tasks

### Phase 1 — RED: establish failing contracts

- [x] **1.1 RED — detector eligibility and configured-name lookup**
  - File: `internal/state/detector_test.go`; target `Detector.Detect` and existing injected `PathLookup` seam.
  - Add table-driven cases using `t.Run` for: a tool whose resource ID differs from `Presence.Name` and finds the configured command; a runtime with `Presence.Kind == "command_exists"`; missing presence metadata; empty `Presence.Name`; a non-`command_exists` presence kind; package metadata; and dotfile resources.
  - Assert only eligible tool/runtime command-presence records call lookup with the configured non-empty name; all other cases remain absent/current unsupported behavior and never fall back to resource ID, package, configuration, or dotfile state.
  - Run `go test ./internal/state`; preserve the failing output as RED evidence before production edits.

- [x] **1.2 RED — planning status propagation**
  - File: `internal/planning/builder_test.go`; target `planBuilder.appendOrderedSteps` / `BuildPlan` and `planning.PlanStep`.
  - Add failing assertions that the ordered `PlanStep.Status` equals the status already emitted in its corresponding `PlanStepResult`, including `PlanStepStatusAlreadyInstalled` for an eligible detected tool/runtime and existing statuses for planned/attention steps.
  - Verify plan order is unchanged and no status is inferred from package or dotfile metadata.
  - Run `go test ./internal/planning`; preserve the failing output as RED evidence.

- [x] **1.3 RED — runner guard, exclusions, and order**
  - File: `internal/execution/runner_test.go`; target `execution.Runner.Run`, installer registry/factory, and recording command runner.
  - Add table-driven failing cases proving the runner guard requires all of: `step.Status == PlanStepStatusAlreadyInstalled`, resource kind `tool` or `runtime`, non-nil `step.Resource.Ref.Presence`, `Presence.Kind == "command_exists"`, and non-empty `Presence.Name`; an eligible step returns one `StepStatusSkipped`/unchanged result with `already installed; no mutation attempted`.
      - Assert installer lookup and command runner are not called only for that fully eligible case; manually constructed `already_installed` tool/runtime steps with nil, empty, or non-command presence metadata are not skipped and retain the existing runner path. Also prove `already_installed` package and dotfile steps retain the existing runner path, while planned/attention/unsupported/failed statuses retain current behavior.
  - Add a mixed-plan case in original order covering fully eligible present, eligible-but-invalid-presence, absent eligible, unsupported, failed, followed by a later step to prove continued execution after failure.
  - Run `go test ./internal/execution`; preserve the failing output as RED evidence.

- [x] **1.4 RED — apply/bootstrap command behavior**
  - File: `cmd/dbootstrap/main_test.go`; target `runApplyLike`, `buildPlan`, `buildApplyRunner`, `runApply`/`runBootstrap`, and existing detector/runner injection seams.
  - Add failing confirmed-mode cases for both `apply --yes` and `bootstrap`: detected eligible tool/runtime has no installer command call, is rendered unchanged with explicit no-mutation wording, and remains in plan order; absent eligible resources still run; unsupported and failures retain output and exit behavior.
  - Add regression assertions that default apply and `--dry-run` retain their existing non-mutating output and do not claim a confirmed idempotency skip; retain `--yes`/`--dry-run` incompatibility, `--sudo` validation, and alias parity coverage.
  - Run `go test ./cmd/dbootstrap`; preserve the failing output as RED evidence.

### Phase 2 — GREEN: smallest implementation

- [x] **2.1 GREEN — carry planning status on ordered steps**
  - File: `internal/planning/types.go`; target `PlanStep`.
  - Add `Status PlanStepStatus` with a comment identifying it as the planning-time status for the ordered executable step; preserve zero-value compatibility for manually constructed steps.
  - File: `internal/planning/builder.go`; target `planBuilder.appendOrderedSteps`.
  - Assign the single computed `status` to both `PlanStep.Status` and `PlanStepResult.Status`; do not reorder/filter steps or add execution-side inference.
  - Make `go test ./internal/planning` pass, then run `gofmt` on changed Go files.

- [x] **2.2 GREEN — constrain detector to reliable command presence**
  - File: `internal/state/detector.go`; target `Detector.Detect`.
  - Gate lookup and present-state recording on resource kind `tool` or `runtime`, non-nil `Presence`, exact `Presence.Kind == "command_exists"`, and non-empty `Presence.Name`.
  - Call injected/default `LookPath` with `Presence.Name`, never `ref.Name`; leave packages, dotfiles, package/version/configuration checks, retries, rollback, and acquisition untouched.
  - Make `go test ./internal/state` pass.

- [x] **2.3 GREEN — skip only eligible already-installed steps at the runner seam**
  - File: `internal/execution/runner.go`; target `Runner.Run` loop before installer lookup/dispatch.
  - Define the skip predicate as the conjunction of `step.Status == PlanStepStatusAlreadyInstalled`, resource kind `tool` or `runtime`, non-nil `step.Resource.Ref.Presence`, `step.Resource.Ref.Presence.Kind == "command_exists"`, and non-empty `step.Resource.Ref.Presence.Name`. Only when every condition is true, append exactly one skipped result with the explicit message `already installed; no mutation attempted`, and continue without resolving an installer or invoking the command runner.
  - Ensure manually constructed steps with nil/empty/non-command presence metadata, zero/non-matching status, package/dotfile resources, unsupported resources, and failures retain the existing path and diagnostics. Preserve one result per plan step and original ordering.
  - Make `go test ./internal/execution` pass.

- [x] **2.4 GREEN — satisfy command-level contracts without changing mode semantics**
  - File: `cmd/dbootstrap/main_test.go` (and only production files if the focused tests prove an existing seam needs the planned wiring); target the shared apply/bootstrap path.
  - Make confirmed apply and bootstrap tests pass while preserving renderer categories, advisory bootstrap output, continued execution, and exit semantics. Do not add flags, installers, acquisition, or a second execution path.
  - Make `go test ./cmd/dbootstrap` pass.

### Phase 3 — TRIANGULATE: prove cross-boundary behavior

- [x] **3.1 TRIANGULATE — focused suite and boundary matrix**
  - Run `go test ./internal/state`, `go test ./internal/planning`, `go test ./internal/execution`, and `go test ./cmd/dbootstrap` after GREEN.
  - Confirm evidence covers the exact eligibility matrix at both detector and runner seams: only `already_installed` tool/runtime steps with non-nil `Presence`, `Presence.Kind == "command_exists"`, and non-empty `Presence.Name` skip; manually constructed tool/runtime steps missing or carrying invalid presence metadata do not skip; package and dotfile resources do not skip; absent fully eligible resources dispatch; unsupported and failed steps remain unchanged.
  - Confirm both `apply --yes` and `bootstrap --yes` share the runner behavior, while default and dry-run remain their existing noop/dry-run behavior.

- [x] **3.2 TRIANGULATE — full regression and formatting**
  - Run the configured strict-TDD runner `go test ./...`.
  - Run `gofmt -l` against changed Go files and require no output; if formatting is needed, apply `gofmt` and rerun the focused tests plus `go test ./...`.
  - Record command results and any skipped external-command scope in apply progress; do not claim success without captured evidence.

### Phase 4 — REFACTOR: documentation and review-safe cleanup

- [x] **4.1 REFACTOR — operational README**
  - File: `README.md`; target the stale command/status/CLI sections.
  - Add an outcome-first quick path for `plan`, default/dry-run inspection, and confirmed `apply --yes`; document `bootstrap` as sharing apply execution semantics while remaining advisory/non-acquiring.
  - Document `--profile`, repeatable `--resource`, `--catalog`, `--yes`, `--sudo`, and `--dry-run`, including `--dry-run` + `--yes` incompatibility and `--sudo` requiring confirmed mode where supported.
  - Explain the narrow promise exactly: only an already-installed `tool` or `runtime` with valid command-presence metadata (`Presence.Kind == command_exists` and non-empty `Presence.Name`) is unchanged with no mutation attempted after successful configured-command detection. State that package and dotfile resources retain normal runner behavior.
  - Document ordered `changed`, `unchanged`, `not supported yet`, and `failed` results; continued processing; non-zero confirmed failures; fix the cause and deliberately rerun; no automatic retry or rollback.
  - State command presence is not package/version/configuration proof or dotfile-link convergence, and missing bootstrap dependencies are manual/advisory—not cloned, fetched, installed, retried, or acquired automatically.

- [x] **4.2 REFACTOR — final scope and rollback inspection**
  - Inspect the diff and changed-line count for only these intended files: `internal/planning/types.go`, `internal/planning/builder.go`, `internal/state/detector.go`, `internal/execution/runner.go`, `internal/state/detector_test.go`, `internal/planning/builder_test.go`, `internal/execution/runner_test.go`, `cmd/dbootstrap/main_test.go`, and `README.md`.
  - Confirm no catalog schema, command flags, installer interface, provider acquisition, package detection, dotfile convergence, retry, rollback, generated artifact, or unrelated documentation changes slipped in.
  - Verify rollback is one PR revert of status propagation, constrained detector lookup, runner guard, tests, and README; no migration or cleanup is required.

## Delivery Boundary

Deliver as one PR containing the detector/planning/execution behavior, its focused regression tests, command-level apply/bootstrap coverage, and README guidance. Keep tests in the same work unit as the behavior they verify. If the implementation exceeds the forecast and approaches 400 changed lines, stop before opening the PR and re-forecast rather than silently expanding the review unit.
