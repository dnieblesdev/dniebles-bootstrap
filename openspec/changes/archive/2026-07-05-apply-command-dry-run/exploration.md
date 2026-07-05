## Exploration: apply-command-dry-run

### Current State

The `dniebles-bootstrap` codebase has a fully working planning pipeline AND execution contracts, but no CLI bridge between them:

- **`internal/planning`** — pure `BuildPlan()` produces `PlanResult` with ordered `PlanStep` slices and planning-time statuses (`planned`, `skipped`, `attention_required`, `already_installed`, `error`). All tests pass.
- **`internal/execution`** — contracts (`Installer`, `DotfilesProvider`), sequential `Runner.Run(ctx, plan) → ExecutionReport`, and noop stubs (`NoopInstaller`, `NoopDotfilesProvider`) that return `StepStatusNotImplemented` without mutation. All tests pass; `Runner` dispatches steps by `ResourceKind` and handles missing installers gracefully.
- **`cmd/dbootstrap`** — `plan` command only (no `apply`). Composition root wires catalog loading, all four detectors (environment, installation state, config state, dotfiles state), calls `planning.BuildPlan()`, and renders `PlanResult`. The CLI switch in `run()` knows only `plan`, `-h`, `--help`, and `help`. All CLI tests pass.
- **Regression gate** — `internal/execution/regression_test.go` contains `TestNoApplyCommandInCLI` that parses `cmd/dbootstrap/main.go`'s AST and asserts no `"apply"` string literal exists in the `run` function. This was the contracts-only gate from the previous slice.
- **All 8 packages pass** `go test ./...` with zero failures. Git is clean at `522d617`.

The gap is clear: `Runner` can consume any `planning.Plan` and produce an `ExecutionReport`, but no CLI surface invokes it.

### Affected Areas

- **`cmd/dbootstrap/main.go`** — MUST be modified to add `"apply"` case to the command switch and a `runApply()` function. The `run()` function currently has a regression test blocking `"apply"`; this slice intentionally adds it.
- **`cmd/dbootstrap/render.go`** — New `renderExecutionReport()` function needed. Follows the same `io.Writer`-based pattern as `renderPlanResult` but renders execution `StepResult` statuses instead of planning `PlanStepResult` statuses.
- **`cmd/dbootstrap/render_test.go`** — New tests for execution report rendering (header, step output, statuses, empty plan).
- **`cmd/dbootstrap/main_test.go`** — New end-to-end tests for `apply` command (with stubbed detectors, same pattern as plan command tests). New usage-error test for `apply` as unknown command (currently "apply" produces "unknown command" error; after this slice, it should succeed or produce proper errors).
- **`internal/execution/regression_test.go`** — `TestNoApplyCommandInCLI` MUST be removed or replaced. The whole point of this slice is to add the `apply` command, so the regression that prevents it is now invalid. Option: replace with a test that the `apply` case exists (or remove entirely and rely on functional tests in `main_test.go`).
- **`openspec/specs/execution-contracts/spec.md`** — MODIFIED. The "No apply command is introduced" requirement and scenario must be removed (they were scoped to the contracts-only slice).
- **`openspec/specs/`** — ADDED delta spec for CLI apply dry-run behavior.
- **`internal/planning/`** — Unchanged. Planning remains pure and unchanged.
- **`internal/execution/` (except `regression_test.go`)** — Unchanged. Contracts remain as-is; no new interfaces or runners needed.

### Approaches

1. **Add `apply` command to `cmd/dbootstrap/main.go` inline (same pattern as `plan`)** — Add `case "apply": return runApply(...)` to the switch, implement `runApply()` with catalog loading + detectors + `BuildPlan()` + `NewRunner(noops)` + execution report rendering. Update `printUsage()` to list the new command. Remove `TestNoApplyCommandInCLI`.

   - **Pros**: Follows established project pattern (`plan` is already fully inline in `main.go`); minimal structural change; easy to review; no new files; `runPlan` and `runApply` share the same detector stubbing infrastructure.
   - **Cons**: `main.go` grows beyond 200 lines; `TestRunUsageErrors` test must change because `"apply"` is no longer "unknown command"; the regression test removal is a deliberate gate change that needs documentation.
   - **Effort**: Low (≈ 150-250 lines, mostly in `main.go`, `render.go`, and tests)

2. **Extract shared pipeline into `cmd/dbootstrap/pipeline.go` before adding `apply`** — Create a `buildPlan()` helper that encapsulates catalog loading + detectors + `BuildPlan()`. Both `runPlan` and `runApply` call it. Keeps `main.go` thin.

   - **Pros**: Reduces `main.go` surface area; cleaner separation; extract-once pattern.
   - **Cons**: Premature extraction — the shared code between `plan` and `apply` is 12 lines (detectors + `BuildPlan` call); adds a file for what amounts to a function wrapper; changes test surface (stub injections become package-level shared state).
   - **Effort**: Medium (≈ 200-300 lines, restructured test mocks)

3. **Keep `apply` fully separate in `cmd/dbootstrap/apply.go`** — New file with `runApply()`, imported or called from `main.go`. Uses `detectEnvironmentFacts` etc. package vars from `main.go`.

   - **Pros**: Logical separation; `apply.go` is self-contained.
   - **Cons**: Duplicates ~40 lines of catalog-loading-and-detection code from `runPlan`; package-level var injection (variables in `main.go` used by `apply.go`) creates coupling that tests must handle; unnecessary for a 12-line shared preamble.
   - **Effort**: Medium (≈ 200-300 lines, plus test restructuring)

### Recommendation

**Approach 1** — add `apply` command inline in `main.go`, same pattern as `plan`. The shared code between `plan` and `apply` (catalog loading, detector calls, `BuildPlan()`) is only 12 lines. Extracting it now is premature abstraction. When a third command arrives that needs the same pipeline, extraction becomes justified.

The `apply` command composition:

```text
catalog load → detectors → BuildPlan() → [planning errors? → exit failure]
                                      → NewRunner(noop installers for all kinds) → Run() → renderExecutionReport() → exit success
```

Key design decisions:
- **No `--dry-run` flag needed.** The entire command is already dry-run because only `NoopInstaller` exists. When real installers arrive in a future slice, `apply` will gain `--dry-run` as an explicit flag.
- **Same flags as `plan`**: `--profile`, `--resource`, `--catalog`.
- **All planning errors still exit failure.** The apply command runs planning first; if planning fails (unknown profile, missing resource), the error exit is preserved — no execution happens for bad plans.
- **Exit code**: `exitSuccess` even when all steps are `not_implemented`, because that's a valid execution outcome with the current noop state. Only planning errors produce `exitFailure`.
- **Noop installer registration**: The runner is constructed with one `NoopInstaller{}` per `planning.ResourceKind` (tool, runtime, package, dotfile). Since `NoopInstaller.SupportedKind()` returns `""`, we need to register separate instances with known kinds. **IMPORTANT**: The current `NoopInstaller.SupportedKind()` returns `""` (empty string). To work with the Runner's kind-dispatch, the apply command must either create kind-aware noop wrappers or modify `NoopInstaller` to accept a configurable kind. The simplest approach: create a small `kindNoop` wrapper at the CLI composition root that wraps `NoopInstaller` with a specific `SupportedKind()`.

This reveals a design tension worth calling out: `NoopInstaller` was designed to match "no concrete kind" (empty string). The `Runner` dispatches by kind. To use noop installers in the runner, we need installers that claim a kind but still return `not_implemented`. The cleanest solution for this slice: a `noopFor(kind) Installer` helper (either in `internal/execution` or at the CLI composition root).

**Recommended**: Add `NoopForKind(kind ResourceKind) Installer` to `internal/execution/noop.go`. This is a one-line change that makes noop installers usable in the Runner without changing their safe semantics.

### Risks

- **Regression test removal**: Removing `TestNoApplyCommandInCLI` is intentional but must be documented. The original test was correct for the contracts-only slice; this slice's purpose is to wire `apply`, making the test obsolete. Replace it with a test that verifies `apply` is present and wired correctly (in `main_test.go`).
- **Status vocabulary confusion in render output**: Execution statuses (`installed`, `failed`, `skipped`, `not_implemented`) differ from planning statuses (`planned`, `skipped`, `attention_required`, `already_installed`, `error`). The `skipped` status appears in both vocabularies with different semantics. The render function must clearly distinguish execution output from planning output. Mitigation: render report header explicitly says "Execution Report" (not "Plan").
- **NoopRunnerNoKind**: `NoopInstaller.SupportedKind()` returns `""`, so `NewRunner(NoopInstaller{})` won't dispatch anything — the runner falls through to "no installer registered for kind" which ALSO returns `not_implemented`. This means even without kind-specific noops, the command produces valid `not_implemented` reports. But the output message differs ("no installer registered for kind" vs "noop installer does not perform real installation"). Both are `not_implemented` but the render function should handle both gracefully.
- **Test surface expansion**: The `apply` command tests need the same detector-stubbing infrastructure as `plan` tests. The package-level variables (`detectEnvironmentFacts`, etc.) are shared, so both command tests reset them. No new stubbing mechanism needed.
- **No real integration path**: Like the previous slice, this slice produces no real mutations. All steps will be `not_implemented`. This is expected and acceptable — the value is proving the `plan → run → report` pipeline end-to-end.

### Ready for Proposal

Yes. The planning pipeline is complete, execution contracts are in place, and the missing piece is the CLI bridge. This slice adds a safe `apply` command that produces execution-style reports without mutation. One new file (`render` changes), test updates, and one regression test gate removal. Proceed to `sdd-propose`.
