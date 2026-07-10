# Apply Progress: wire-dotfiles-apply-yes

## Structured status consumed

- change: `wire-dotfiles-apply-yes`
- artifactStore: both (OpenSpec + Engram)
- execution mode: auto
- actionContext: repository workspace `/home/dniebles/dniebles-bootstrap`; delegated scope allowed `cmd/dbootstrap`, tiny provider-core fixes, and SDD artifacts.
- applyState/readiness: proceeded after required OpenSpec artifacts and Engram references were confirmed present. OpenSpec has design/proposal/tasks and spec deltas under `specs/` rather than a top-level `spec.md`.
- review workload gate: tasks forecast is moderate and prompt supplied resolved delivery path as second chained slice after `dotfiles-execution-provider-core`, no local line limit, keep scope tight.
- strict TDD: active; runner `go test ./...`.

## Workload / PR boundary

Second chained slice after committed `dotfiles-execution-provider-core`. Scope stayed within CLI composition/reporting/tests and tiny provider-core reporting/accessor changes. No acquisition, rollback, bootstrap entrypoint, apt provider, or symlink tracking was added.

## TDD Cycle Evidence

| Cycle | RED evidence | GREEN evidence | TRIANGULATE / REFACTOR evidence |
|---|---|---|---|
| Dotfiles apply safety and confirmed execution | Added CLI tests for safe modes, plan non-composition, confirmed dotlink invocation, prerequisite failures, runner failure, timeout, and forbidden acquisition commands. Initial focused run failed to build on missing `newDotfilesInstaller` seam: `go test ./cmd/dbootstrap ./internal/execution`. | Added confirmed-only dotfiles installer composition seam and provider wiring; focused tests passed after updating confirmed-mode output expectations. | Added explicit safe-mode assertion that dotfile resources remain `not supported yet`; reran focused tests successfully. |
| Reporting / exit behavior | Failure tests expected non-zero confirmed apply exit and failed dotfile step rendering. | Added confirmed failed-result exit rule, dotfiles base/source/modules reporting, and updated confirmed-mode copy. | Full repository suite `go test ./...` passed. |

## Completed tasks and persisted checkbox updates

All tasks in `openspec/changes/wire-dotfiles-apply-yes/tasks.md` were marked `- [x]`:

- 1. Add failing apply safety tests for dotfile resources.
- 2. Add failing confirmed dotfiles execution tests with fake seams.
- 3. Add failing safe-failure tests.
- 4. Add failing guard tests for excluded acquisition/provider behavior.
- 5. Add or refine CLI composition seams.
- 6. Wire confirmed apply dotfiles execution.
- 7. Update reporting.
- 8. Verify and refactor.

## Files changed

- `cmd/dbootstrap/main.go`
  - Added `newDotfilesInstaller` test seam.
  - Composed dotfiles installer only for confirmed plans containing selected dotfile steps.
  - Kept default and dry-run on noop execution.
  - Preserved brew-backed tool/package behavior and missing-brew manual guidance.
  - Returned non-zero from confirmed apply when any execution result fails.
- `cmd/dbootstrap/render.go`
  - Updated confirmed-mode safety copy to mention brew-backed tool/package and selected dotfile resources may have changed.
- `cmd/dbootstrap/main_test.go`
  - Added strict-TDD coverage for safe modes, confirmed dotlink execution with fake runner/base resolver, missing base/dotlink/module, runner failure/timeout, no acquisition commands, and unselected resources.
- `cmd/dbootstrap/render_test.go`
  - Updated confirmed-mode copy expectation.
- `internal/execution/dotfiles_installer.go`
  - Added optional base-context reporting for providers that expose resolved base metadata.
- `internal/execution/dotfiles_provider.go`
  - Exposed resolved base metadata through `DotfilesBase()` for reporting.
- `openspec/changes/wire-dotfiles-apply-yes/tasks.md`
  - Marked completed tasks.
- `openspec/changes/wire-dotfiles-apply-yes/apply-progress.md`
  - Recorded this progress.

## Verification commands run

1. `go test ./cmd/dbootstrap ./internal/execution` — RED build failure after tests: missing `newDotfilesInstaller` seam.
2. `gofmt -w cmd/dbootstrap/main.go cmd/dbootstrap/render.go cmd/dbootstrap/main_test.go internal/execution/dotfiles_installer.go internal/execution/dotfiles_provider.go && go test ./cmd/dbootstrap ./internal/execution` — failed on old confirmed-mode output expectations.
3. `gofmt -w cmd/dbootstrap/main_test.go cmd/dbootstrap/render_test.go && go test ./cmd/dbootstrap ./internal/execution` — passed.
4. `gofmt -w cmd/dbootstrap/main_test.go && go test ./cmd/dbootstrap ./internal/execution` — passed after triangulation assertion.
5. `go test ./...` — passed.
6. `gofmt -w cmd/dbootstrap/main.go cmd/dbootstrap/main_test.go && go test ./cmd/dbootstrap ./internal/execution && go test ./...` — passed after updating command/flag help copy to mention selected dotfiles.

## Design deviations

- No functional deviation from the approved design.
- Tiny provider-core additions were made only to expose resolved dotfiles base metadata for CLI result reporting (`DotfilesBaseReporter` and `LocalDotfilesProvider.DotfilesBase`). No acquisition/rollback/repair/tracking behavior was added.

## Remaining tasks

None. Persisted task artifact was re-read after checkbox update and all implementation tasks are visibly checked.

## Risks / notes

- Diff is test-heavy as forecasted (CLI matrix coverage); implementation remains limited to composition/reporting plus provider metadata exposure.
- Dotfiles execution uses the existing `CommandRunner` abstraction only; tests do not run real `dotlink`.
- Dotfiles apply path requests only `<base>/bin/dotlink link <module>` in tests and guards against clone/pull/submodule/fetch/remote/sparse/apt requests.

## Task-level strict TDD evidence

| Task | Test files / functions | Safety net before change | RED evidence | GREEN evidence | Triangulation / final evidence |
|---|---|---|---|---|---|
| 1. Safe apply modes | `cmd/dbootstrap/main_test.go`: `TestRunApplySafeModesDoNotInstantiateRealExecution` | Existing `go test ./cmd/dbootstrap ./internal/execution` passed before new CLI tests. | New safe-mode cases were added before the composition seam existed; the focused package run initially failed during compilation on the missing dotfiles-installer seam. | After confirmed-only composition was added, default apply, dry-run, and plan cases asserted no dotfiles installer construction and `not supported yet` behavior. | `go test -count=1 ./cmd/dbootstrap ./internal/execution` and `go test -count=1 ./...` pass. |
| 2. Confirmed success | `cmd/dbootstrap/main_test.go`: `TestRunApplyConfirmedDotfilesUsesInjectedRunner` | Existing focused suite passed before the new test. | The test referenced the missing `newDotfilesInstaller` composition seam and failed to compile. | The new seam plus confirmed-only runner composition made the fake runner receive exactly the selected `bash` module and report changed output with base/source/module context. | Focused/full uncached suites pass; fake runner prevents real dotlink execution. |
| 3. Confirmed failures / exit status | `cmd/dbootstrap/main_test.go`: `TestRunApplyConfirmedDotfilesFailuresExitNonZero` | Existing focused suite passed before new failure cases. | Missing base, missing dotlink, missing module, fake runner failure, and fake timeout cases were written before confirmed wiring and could not compile without the seam. | Confirmed apply now renders failed results and returns non-zero for all five cases; validation failures make zero runner calls and runner failures/timeouts make one deterministic fake call. | Focused/full uncached suites pass; test asserts no changed status and no retry/fallback acquisition. |
| 4. Excluded acquisition / selection guard | `cmd/dbootstrap/main_test.go`: confirmed dotfiles runner cases and forbidden-command assertions | Existing focused suite passed before guard assertions. | Guard assertions were introduced with the confirmed execution tests before fake composition existed. | Fake runner command recording proves only `<base>/bin/dotlink link bash` is requested; unselected/non-dotfile resources are excluded. | Full suite passes; source-safety/core tests prohibit direct process execution and acquisition behavior. |
| 5. Composition seams | `cmd/dbootstrap/main.go`, `cmd/dbootstrap/main_test.go` confirmed/safe-mode tests | Existing apply tests passed before seam creation. | Compile failure explicitly identified missing `newDotfilesInstaller`. | Added `newDotfilesInstaller` seam with production construction used only in confirmed mode and injected fakes in tests. | Safe-mode and confirmed tests pass together in focused/full suites. |
| 6. Confirmed provider wiring | `cmd/dbootstrap/main_test.go`: confirmed success/failure tests | Existing brew confirmed tests passed before dotfiles wiring. | Confirmed dotfiles tests failed/compiled red until the dotfiles installer was registered only in the confirmed runner. | Confirmed selected dotfiles dispatch through the existing provider/installer; brew-backed tool/package behavior remains covered by existing confirmed tests. | Focused/full uncached suites pass. |
| 7. Reporting | `cmd/dbootstrap/render_test.go` plus confirmed CLI output tests | Existing renderer tests passed before copy/context updates. | Focused run failed on stale confirmed-mode output expectations after behavior/tests were added. | Updated renderer copy and provider base-context reporting; tests assert canonical base, source, module, and changed/failed status text. | Focused/full uncached suites pass. |
| 8. Verify / refactor | all files above | Focused tests served as the initial safety net. | RED failures are recorded above before implementation. | GREEN implementation passed focused tests after formatting. | `go test -count=1 ./cmd/dbootstrap ./internal/execution`, `go test -count=1 ./...`, and `go vet ./...` pass; no scope expansion beyond wiring/reporting plus minimal provider metadata. |

## Strict TDD status ledger

| Task | RED status | GREEN status | TRIANGULATE status | Evidence reference |
|---|---|---|---|---|
| 1. Safe apply modes | ✅ Written — safe-mode cases added before composition seam | ✅ Passed — no installer construction / noop assertions pass | ✅ Passed | `TestRunApplySafeModesDoNotInstantiateRealExecution`; focused and full uncached test commands. |
| 2. Confirmed success | ✅ Written — fake-seam confirmed success test added before seam | ✅ Passed — selected `bash` fake runner request and changed output pass | ✅ Passed | `TestRunApplyConfirmedDotfilesUsesInjectedRunner`; focused and full uncached test commands. |
| 3. Confirmed failures / exit status | ✅ Written — missing prerequisite, runner failure, and timeout cases added before wiring | ✅ Passed — failures render failed and return non-zero | ✅ Passed | `TestRunApplyConfirmedDotfilesFailuresExitNonZero`; focused and full uncached test commands. |
| 4. Excluded acquisition / selection guard | ✅ Written — forbidden acquisition and selected-module assertions added before wiring | ✅ Passed — fake runner observes only allowed dotlink request | ✅ Passed | confirmed dotfiles test helpers/assertions; focused and full uncached test commands. |
| 5. Composition seams | ✅ Written — tests referenced missing `newDotfilesInstaller` seam | ✅ Passed — injected seam works in confirmed tests and is absent from safe modes | ✅ Passed | initial RED compile failure and safe/confirmed CLI test functions. |
| 6. Confirmed provider wiring | ✅ Written — selected-dotfile confirmed execution tests added before registration | ✅ Passed — confirmed runner dispatches dotfiles only under `--yes` | ✅ Passed | confirmed success/failure CLI test functions; focused and full uncached test commands. |
| 7. Reporting | ✅ Written — output assertions added before renderer/context updates | ✅ Passed — base/source/module and confirmed-copy assertions pass | ✅ Passed | `render_test.go` and confirmed CLI output assertions; focused and full uncached test commands. |
| 8. Verify / refactor | ✅ Written — test-first cycle recorded for every preceding task | ✅ Passed — focused suite green after implementation/refactor | ✅ Passed | `go test -count=1 ./cmd/dbootstrap ./internal/execution`; `go test -count=1 ./...`; `go vet ./...`. |

## Corrective verification blocker — 2026-07-09

- Native authoritative status was consumed before remediation. It reports `applyState: all_done`, `taskProgress: 34/34`, and `nextRecommended: verify`.
- `verify-report.md` identifies one CRITICAL gap: `DotfilesInstaller.Install` collapses missing-base, missing-`bin/dotlink`, and missing-module provider errors into `dotfile module <name> failed`.
- Per the apply-state guard, no corrective code or test edits were made while the authoritative apply state is `all_done`; no existing completed task checkbox was reopened or changed.
- Required follow-up: route this as a verification remediation (or reopen an explicit corrective apply task) before changing `internal/execution/dotfiles_installer.go` and its tests.

### Corrective TDD evidence

| Cycle | RED | GREEN | TRIANGULATE / REFACTOR |
|---|---|---|---|
| Failure-message clarity blocker | Blocked before RED: authoritative status is `all_done`; no test was added or changed. | Not started. | Not started. |

### Corrective files changed

- `openspec/changes/wire-dotfiles-apply-yes/apply-progress.md` only (status/progress record; no implementation change).

### Corrective remaining work

- [ ] Add RED assertions for distinct understandable missing-base, missing-`bin/dotlink`, and missing-module messages plus non-zero confirmed exits.
- [ ] Preserve or safely map provider error detail into `DotfilesInstaller.Install` failure `StepResult.Message`.
- [ ] Run focused and full strict-TDD validation, then rerun verification.

## Corrective remediation completed — 2026-07-09

### Structured status consumed

- Native authoritative status: `applyState: ready`, `taskProgress: 34/35`, `nextRecommended: apply`, no blocked reasons.
- `actionContext`: `repo-local`, workspace and allowed edit root `/home/dniebles/dniebles-bootstrap`; all edits are within this root.
- Strict TDD is active; configured runner: `go test ./...`.

### Completed task and persisted checkbox update

- [x] 9. **RED/GREEN — preserve understandable dotfiles prerequisite failures.**
  - Updated `openspec/changes/wire-dotfiles-apply-yes/tasks.md` immediately after GREEN and full validation.

### Corrective TDD Cycle Evidence

| Task | Test file / layer | Safety net | RED | GREEN | TRIANGULATE / REFACTOR |
|---|---|---|---|---|---|
| 9. Preserve understandable prerequisite failures | `cmd/dbootstrap/main_test.go` / in-process CLI unit test | `go test ./internal/execution -run 'TestDotfilesInstaller'` and `go test ./cmd/dbootstrap -run '^TestRunApplyConfirmedDotfilesFailuresExitNonZero$'` passed before edits. | Changed the three confirmed prerequisite cases to require distinct user-facing provider causes: `resolve dotfiles base`, `validate dotlink`, and `validate module \"zsh\"`; focused CLI run failed with each missing assertion while retaining non-zero exits and zero runner calls. | `DotfilesInstaller.Install` now includes the provider error text in its failed `StepResult.Message`; focused CLI and installer tests pass. | The three distinct prerequisite inputs force the generic message to preserve each separate provider cause. No further refactor was needed beyond the minimal message formatting change. |

### Corrective files changed

- `cmd/dbootstrap/main_test.go` — requires distinct rendered messages for missing base, missing `bin/dotlink`, and missing selected module, while retaining non-zero exits and zero runner calls.
- `internal/execution/dotfiles_installer.go` — preserves provider failure detail in the dotfile failed-step message.
- `openspec/changes/wire-dotfiles-apply-yes/tasks.md` — marked task 9 complete.
- `openspec/changes/wire-dotfiles-apply-yes/apply-progress.md` — merged this corrective evidence.

### Corrective verification commands

1. `go test ./internal/execution -run 'TestDotfilesInstaller'` — PASS (safety net).
2. `go test ./cmd/dbootstrap -run '^TestRunApplyConfirmedDotfilesFailuresExitNonZero$'` — PASS (safety net), then FAIL (RED: all three distinct cause assertions missing), then PASS (GREEN).
3. `go test ./internal/execution -run '^TestDotfilesInstaller'` — PASS after GREEN.
4. `go test ./cmd/dbootstrap ./internal/execution` — PASS.
5. `go test ./...` — PASS.
6. `git diff --check` — PASS.

### Corrective scope and remaining work

- No design deviation: the change remains confirmed-only reporting/error propagation.
- No acquisition, clone, pull, submodule, fetch, remote, sparse, apt, bootstrap, rollback, or safe-mode behavior changed.
- Workload / PR boundary: corrective task 9 only; two production/test files plus SDD artifacts. No chaining or size exception required.
- Remaining implementation tasks: none. The verification report must be rerun to clear its previously recorded CRITICAL finding.
