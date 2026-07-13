# Brew Package Presence Idempotency Tasks

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 240–340 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | Single PR: detector, execution guard, confirmed-mode composition, tests, README |
| Delivery strategy | single-pr |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Medium

## Scope and Boundary

Single PR, limited to Brew formula presence in confirmed `apply --yes` and confirmed `bootstrap`. Use `Resource.Install.Package` (`InstallMetadata.Package`) as the formula authority, never `PlanStep.Ref.Name` or `Resource.Presence.Name`. The exact read-only query is the structured argv vector `brew list --formula <InstallMetadata.Package>` through the existing injected `execution.CommandRunner`, with the fixed detector timeout and no shell.

Do not change APT, casks, versions, sudo detection, shell execution, retries, fallback queries, bootstrap acquisition, catalog/TOML schema, persisted state, or broader convergence. Do not execute real Brew or mutate the host in tests. Keep `BuildPlan` pure and unchanged in behavior.

If the implementation diff reaches 400 changed lines (`additions + deletions`), stop before adding scope, report the measured count, and reforecast/re-split for maintainer approval. Do not silently exceed the budget.

## Ordered Strict-TDD Work Units

- [x] 1. **RED — Planning transient state and detector contract**

- **Files/symbols:** `internal/planning/types.go` (`PlanStep`, new `PackagePresence` enum); new `internal/state/brew_formula_detector_test.go`; existing `internal/state/detector_test.go` seams/fakes as reusable.
- Add failing table-driven tests that define the detector contract: eligible only for `ResourceKindPackage`, `Install.Provider == "brew"`, and trimmed non-empty `Install.Package`; no probe for non-Brew, tools/runtimes, blank/invalid metadata, cask-like metadata, or unsupported resources.
- Assert the formula identity differs safely from both resource/ref name and `Presence.Name`; with package `jq`, the exact request must be executable `brew` and args `[]string{"list", "--formula", "jq"}`.
- Assert the request carries the fixed positive timeout, uses the injected runner, has no shell representation, and performs at most one query per eligible ordered step.
- Assert `BuildPlan`/planning remains OS-probe-free and starts each transient presence field as `PackagePresenceUnchecked`.
- Verify RED with: `go test ./internal/state ./internal/planning` (expected failure because the new symbols/behavior do not exist).

- [x] 2. **GREEN — Implement planning state and conservative Brew detector**

- **Files/symbols:** `internal/planning/types.go`; new `internal/state/brew_formula_detector.go`; only supporting existing command-result types in `internal/execution` if required by their current API.
- Add `PackagePresenceUnchecked`, `PackagePresenceInstalled`, `PackagePresenceAbsent`, and `PackagePresenceUnknown` plus a transient `PlanStep` field; do not add catalog or persistence fields.
- Implement `state.BrewFormulaDetector`, its injected `CommandExists` and `execution.CommandRunner` seams, fixed `brewFormulaPresenceTimeout`, eligibility predicate, exact argv construction, context/timeout propagation, and ordered detection map.
- Implement conservative classification: successful completion with exit code 0 is installed; only the supported completed Brew formula-absent diagnostic with the expected failed status/exit code is absent; missing Brew, nil runner, timeout/not-run, runner error, malformed success, unrecognized non-zero, and unsupported metadata are unknown. Exit code 1 alone is never absent. Do not expose raw query output, retry, or fallback.
- Implement an equivalent copy helper to decorate an execution plan. It must preserve order and the original plan, set only eligible classified refs, and add the stable sanitized unknown attention reason `Homebrew formula presence could not be determined; no mutation attempted` without converting unknown to absent/already-installed.
- GREEN gate: `go test ./internal/state ./internal/planning` passes; `go vet ./internal/state ./internal/planning` passes.

- [x] 3. **TRIANGULATE — Detector edge cases and no-host-mutation proof**

- **Files/symbols:** `internal/state/brew_formula_detector_test.go`; relevant `internal/execution` command-runner fake types only if existing fakes cannot record requests.
- Expand table cases for installed, explicitly absent, missing `brew`, nil runner, runner error, timeout, malformed success result, unclassified non-zero, localized/changed diagnostic, nil/invalid metadata, and duplicate/non-eligible plan steps.
- Assert exact request vector and timeout for every eligible query; assert query order follows plan order and no query is retried or followed by fallback.
- Assert zero installer calls and zero mutation requests for installed and every unknown case; assert absent is the only non-installed classification that remains eligible for the existing Brew installer.
- Assert copied-plan isolation: original `Plan` and its steps/reasons remain unchanged, and no status is marked already installed before a positive result.
- Run: `go test -count=1 ./internal/state ./internal/planning` and `go test -count=1 ./...`; retain only fake/injected command execution.

- [x] 4. **RED — Execution pre-dispatch behavior**

- **Files/symbols:** `internal/execution/runner_test.go`; existing `Runner.Run`, `isAlreadyInstalledCommandStep`, `StepResult`, and installer fakes.
- Add failing table-driven tests for an eligible Brew package with transient installed state: result remains in original position, status is the existing skipped/unchanged status, message is exactly `already installed; no mutation attempted`, and neither installer selection nor command runner is called.
- Add failing tests for transient unknown: status is failed/attention, message is exactly `Homebrew formula presence could not be determined; no mutation attempted`, a package-presence sentinel error is present, installer and runner are not called, and later steps still execute in order.
- Add failing tests for absent and unchecked: existing provider-aware Brew dispatch remains unchanged; detection alone cannot produce an idempotent skip.
- Add failing boundary-revalidation tests proving manually supplied installed/unknown state is ignored unless the step is a valid Brew package with trimmed non-empty `Install.Package`; tools/runtimes, APT, cask-like, malformed, and unsupported steps retain current semantics.
- Verify RED with: `go test ./internal/execution` (expected failure).

- [x] 5. **GREEN — Add the execution guard without changing unrelated dispatch**

- **Files/symbols:** `internal/execution/runner.go`; package-level sentinel/error definition in the existing execution error location; `internal/execution/runner_test.go`.
- Extend `Runner.Run` pre-dispatch in existing step order: retain tool/runtime command-presence behavior, then handle valid Brew package `Installed` and `Unknown`, then leave `Absent`/`Unchecked` on existing provider-aware dispatch.
- Ensure unknown is a visible failed execution outcome that contributes to confirmed command failure while preserving continued execution for later steps. Do not change installer implementations, command vectors, retries, fallback, sudo, or APT behavior.
- GREEN gate: `go test ./internal/execution` and `go vet ./internal/execution` pass.

- [x] 6. **TRIANGULATE — Mixed-plan order and mutation safety at the runner boundary**

- **Files/symbols:** `internal/execution/runner_test.go` and existing fake installer/runner types.
- Prove ordered mixed plans containing installed Brew, absent Brew, and an unrelated tool/runtime/dotfile/unsupported step: result indexes and refs remain unchanged, absent uses existing installer behavior, installed has no mutation, and unrelated/later failures do not stop continuation.
- Prove no installer or mutation runner call for unknown, including timeout, missing manager, malformed result, and unclassified non-zero outcomes; prove no retry/fallback call count.
- Prove existing already-installed tools/runtimes still report unchanged with no command runner call and that package steps cannot inherit that command-presence rule.
- Run: `go test -count=1 ./internal/execution` and `go test -count=1 ./...`.

- [x] 7. **RED — Confirmed CLI composition and safe-mode boundaries**

- **Files/symbols:** `cmd/dbootstrap/main_test.go`; `cmd/dbootstrap/main.go` seams (`runApplyLike`, `buildApplyRunner`, `isConfirmedMode`, `newOSCommandRunner`, `brewCommandExists`, installer factories); existing recording/sequence runners.
- Add failing table-driven composition tests for both `apply --yes` and `bootstrap --yes`: detector is composed only after the pure plan, the exact presence query precedes that package's install request, installed output is unchanged with explicit no-mutation wording, and no install call occurs.
- Add failing unknown cases for unavailable Brew, timeout, runner error, and unclassified non-zero: report attention/failure, return confirmed failure as existing failure aggregation requires, and make zero install calls for that package while later steps retain order/continuation.
- Add failing mixed-plan tests for installed Brew + absent Brew + unrelated step and assert exact runner call sequence.
- Add failing default and `--dry-run` tests for `apply` and `bootstrap`: no Brew lookup, no Brew presence query, no command-runner factory solely for package presence, and existing non-mutating output remains unchanged. Assert `--sudo` is not passed to detection.
- Verify RED with: `go test ./cmd/dbootstrap` (expected failure).

- [x] 8. **GREEN — Wire detector only into confirmed apply/bootstrap**

- **Files/symbols:** `cmd/dbootstrap/main.go`; `cmd/dbootstrap/main_test.go`.
- In `runApplyLike`, after plan errors and only under `isConfirmedMode(mode)`, lazily construct the existing OS runner only when the plan has an eligible Brew package; build `state.BrewFormulaDetector`, detect, decorate a copied plan, then execute it.
- Share the path between confirmed apply and bootstrap; preserve `buildPlan`, default mode, dry-run mode, advisory bootstrap behavior, and existing installer construction. Keep `--sudo` restricted to installation concerns.
- Add only a narrow detector factory/function seam if needed for tests; do not add flags, schema, provider registries, or new process abstractions.
- GREEN gate: `go test ./cmd/dbootstrap` and `go test ./...` pass.

- [x] 9. **TRIANGULATE — End-to-end acceptance and scope regression tests**

- **Files/symbols:** `cmd/dbootstrap/main_test.go`, `internal/state/brew_formula_detector_test.go`, `internal/execution/runner_test.go`; no real Brew invocation.
- Assert exact argv, timeout, query-before-install ordering, plan order, installed/absent/unknown outcomes, explicit no-mutation messages, zero host mutation, and continued execution across apply/bootstrap.
- Add regression assertions that tools/runtimes keep command-presence semantics and that APT, casks, versions, sudo detection, shell strings, retries, fallback, bootstrap acquisition, and catalog/persisted schema remain untouched and unprobed.
- Run focused suites, then mandatory `go test ./...`; also run `go vet ./...` and `gofmt -l internal/planning internal/state internal/execution cmd/dbootstrap`.
- Record changed-line count. If it reaches 400, stop and reforecast before any further edits.

- [x] 10. **REFACTOR — Documentation and review-ready cleanup**

- **Files/symbols:** `README.md` section `Confirmed reruns`; all change files listed above.
- Update the confirmed-rerun explanation to state the narrow positive Brew formula exception and conservative unknown/no-install behavior, while retaining safe-mode and broader-convergence exclusions. Do not document APT/cask/version/sudo/retry/fallback behavior as part of this change.
- Simplify helpers/comments and keep tests table-driven, scenario-named, deterministic, and focused on behavior; remove temporary scaffolding. Apply `gofmt` only to touched Go files.
- Final verification: `go test ./...`, `go vet ./...`, and `gofmt -l` on touched Go files; confirm only the design-listed files plus `tasks.md` are changed and the measured diff remains below 400 lines.
- Rollback boundary: revert the detector, transient state, execution guard, confirmed-mode wiring, tests, and README as one PR; no migration or cleanup is required.
