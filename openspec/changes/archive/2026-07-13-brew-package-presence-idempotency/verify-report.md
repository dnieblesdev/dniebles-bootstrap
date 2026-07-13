```yaml
schema: gentle-ai.verify-result/v1
evidence_revision: sha256:232788f71a44b42e2108efa7d6dd29aeaebe614676d77b460d50a33c9ca0c334
verdict: pass
blockers: 0
critical_findings: 0
requirements: 14/14
scenarios: 29/29
test_command: go test ./... -count=1
test_exit_code: 0
test_output_hash: sha256:16c351fe0f931a1e22cc8037416d6132338bc218e7ba575ac637a758b3c23b50
build_command: go build ./...
build_exit_code: 0
build_output_hash: sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```

# Verify Report: Brew Package Presence Idempotency

**Change**: `brew-package-presence-idempotency`
**Mode**: Strict TDD
**Reviewer**: independent verify phase
**Date**: 2026-07-13
**Store**: openspec (authoritative); project configuration is hybrid (`both`).
**Action context**: `repo-local`; workspace and allowed edit root `/home/dniebles/dniebles-bootstrap`.

## Status: PASS — archive candidate

Independent verification completed on 2026-07-13 under strict TDD against the current post-apply remediation diff. The prior blockers (unformatted `cmd/dbootstrap/main_test.go`, incomplete bootstrap/default/dry-run/mixed-error test evidence, invalidated review receipt, and the missing bounded review transaction) are all resolved. `gofmt`, `go vet`, focused and full suites are green. A bounded post-apply review transaction is now finalized and bound to this change, and the `gentle-ai.verify-result/v1` envelope's `evidence_revision` references that bound review's evidence hash.

## Bounded Post-Apply Review

| Field | Value |
|---|---|
| State | approved |
| Risk tier | medium |
| Selected lens | `review-reliability` |
| Scope (files / lines) | 2 / 113 (`cmd/dbootstrap/main_test.go`, `openspec/changes/brew-package-presence-idempotency/apply-progress.md`) |
| Correction budget | 57 |
| Corrections applied | 0 (clean approval) |
| Post-apply gate | `allow` |
| evidence_revision | `sha256:232788f71a44b42e2108efa7d6dd29aeaebe614676d77b460d50a33c9ca0c334` |

The reliability lens review of the remediation diff found no BLOCKER/CRITICAL/WARNING findings; the diff adds behavioral coverage only and does not change production code. The approved `gentle-ai.review-receipt/v2` lineage and its `gentle-ai.sdd-review-binding/v1` binding for this change are persisted under `.git/gentle-ai/review-transactions/v2/` and `.git/gentle-ai/sdd-review-bindings/v1/brew-package-presence-idempotency/binding.json` respectively; `review validate --gate post-apply` returns `allow` with `base_relationship_valid: true`. The exact lineage id, receipt, base_tree, and candidate_tree hashes are read from the binding file (the authoritative source) rather than duplicated here, since duplicating them in this artifact would change the candidate tree after binding.

## Task Completion

| Metric | Value |
|---|---|
| Tasks total | 10 |
| Tasks complete (`- [x]`) | 10 |
| Tasks incomplete (`- [ ]`) | 0 |

No unchecked implementation task lines remain in `tasks.md`. The previously-checked-but-unproven work units (Tasks 7, 9, 10) are now substantiated by the added bootstrap composition, missing-Brew, safe-mode, and mixed-error test evidence below; `gofmt -l` no longer reports any touched Go file.

## Build & Tests Execution (real current evidence)

| Command | Result |
|---|---|
| `gofmt -l cmd/dbootstrap/main_test.go internal/execution/runner.go internal/execution/runner_test.go internal/planning/types.go internal/state/brew_formula_detector.go internal/state/brew_formula_detector_test.go cmd/dbootstrap/main.go` | PASS (no files listed; exit 0) |
| `go vet ./...` | PASS (no output; exit 0) |
| `go build ./...` | PASS (no output; exit 0; `build_output_hash` as in envelope) |
| `git diff --check` | PASS (exit 0) |
| `go test -count=1 -run 'TestConfirmedCommandsCheckBrewFormulaBeforeInstall\|TestRunBootstrapConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller\|TestRunBootstrapDefaultAndDryRunDoNotProbeBrew' ./cmd/dbootstrap -v` | PASS |
| `go test -count=1 ./internal/state ./internal/planning ./internal/execution ./cmd/dbootstrap` | PASS (4/4 packages green) |
| `go test ./... -count=1` | PASS (test_exit_code 0; `test_output_hash` as in envelope) |

Focused remediation subtests all pass: `apply_installed`, `bootstrap_installed`, `apply_explicitly_absent`, `bootstrap_explicitly_absent`, `apply_timed_out`, `bootstrap_timed_out`, `bootstrap_runner_error`, `bootstrap_unclassified_non-zero`, plus `TestRunBootstrapConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller` and `TestRunBootstrapDefaultAndDryRunDoNotProbeBrew/{default,default_--dry-run}`.

## Spec Compliance Matrix

Requirement counts: installation-state 5 (ADDED 3 + MODIFIED 1 + REMOVED "None" 1), execution-contracts 5 (ADDED 2 + MODIFIED 2 + REMOVED "None" 1), apply-command-dry-run 4 (ADDED 2 + MODIFIED 1 + REMOVED "None" 1) = **14/14**. The three REMOVED "None" requirements have no scenarios and are trivially satisfied (no requirement is removed by this change). Active scenario counts: 13 + 10 + 6 = **29/29**.

### installation-state (13 scenarios)

| Requirement | Scenario | Result | Covering test(s) |
|---|---|---|---|
| Confirmed Brew formula presence detection is read-only | Eligible formula uses package metadata | PASS | `TestBrewFormulaDetectorDetectsOnlyEligibleFormulaPackages`; `TestRunApplyConfirmedBrewPresentUsesInjectedRunnerForBrewOnly` (asserts `list --formula jq`) |
| Confirmed Brew formula presence detection is read-only | Detection is not shell-based | PASS | `TestBrewFormulaDetectorDetectsOnlyEligibleFormulaPackages` (executable + argument vector, injected `CommandRunner`) |
| Brew query results are classified conservatively | Successful query proves installed | PASS | `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/{apply_installed,bootstrap_installed}` |
| Brew query results are classified conservatively | Explicit absent result remains installable | PASS | `TestBrewFormulaDetectorClassifiesFailuresAsUnknownAndExplicitAbsence` (`No such keg` → absent); `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/{apply_explicitly_absent,bootstrap_explicitly_absent}` (installer remains eligible) |
| Brew query results are classified conservatively | Operational non-zero is not absence | PASS | `TestBrewFormulaDetectorClassifiesFailuresAsUnknownAndExplicitAbsence` (unclassified non-zero → unknown); `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/bootstrap_unclassified_non-zero` |
| Brew query results are classified conservatively | Unavailable Brew is unknown | PASS | `TestRunBootstrapConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller`; `TestRunApplyConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller` |
| Brew query results are classified conservatively | Timeout or runner failure is unknown | PASS | `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/{apply_timed_out,bootstrap_timed_out,bootstrap_runner_error}` |
| Brew query results are classified conservatively | Metadata cannot authorize a probe | PASS | `TestBrewFormulaDetectorDetectsOnlyEligibleFormulaPackages` (non-Brew/empty metadata → no probe); `TestRunnerIgnoresPackagePresenceForNonMatchingProvider` |
| Confirmed Brew presence affects execution state only after a positive result | Installed formula occupies its original position | PASS | `TestRunnerHonorsEligibleBrewFormulaPresence`; `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/apply_installed` (position preserved, no mutation) |
| Confirmed Brew presence affects execution state only after a positive result | Unknown package does not become false absence | PASS | `TestApplyBrewFormulaPresenceCopiesPlanAndAddsUnknownAttention`; `TestRunnerHonorsEligibleBrewFormulaPresence` (attention/failure, installer suppressed) |
| Idempotency detection is limited to reliable command or Brew formula presence (MODIFIED) | Reliable command presence remains sufficient | PASS | `TestRunApplyAndBootstrapSkipDetectedCommandPresence` |
| Idempotency detection is limited to reliable command or Brew formula presence (MODIFIED) | Positive Brew formula presence enables idempotency | PASS | `TestRunApplyConfirmedBrewPresentUsesInjectedRunnerForBrewOnly`; `TestRunApplyLikeConfirmedMixedBrewAptPreservesBrewPresence` |
| Idempotency detection is limited to reliable command or Brew formula presence (MODIFIED) | Broader reconciliation is not attempted | PASS | `TestRunApplyConfirmedBrewPresentUsesInjectedRunnerForBrewOnly` (no version/config/dotfile checks); `TestRunnerIgnoresAptPresenceForBrewPackage` |
| None (REMOVED) | _(no scenarios; no requirement is removed by this change)_ | PASS | Trivially satisfied — `grep -rc "^### Requirement:" specs/` confirms no delta removes a capability; diff introduces only ADDED/MODIFIED behavior. |

### execution-contracts (10 scenarios)

| Requirement | Scenario | Result | Covering test(s) |
|---|---|---|---|
| Confirmed Brew package presence is checked before installer dispatch | Installed package is skipped in order | PASS | `TestRunnerHonorsEligibleBrewFormulaPresence`; `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/apply_installed` |
| Confirmed Brew package presence is checked before installer dispatch | Absent package remains eligible | PASS | `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/{apply_explicitly_absent,bootstrap_explicitly_absent}` (`brew install jq` dispatched) |
| Confirmed Brew package presence is checked before installer dispatch | Unknown package suppresses mutation | PASS | `TestRunnerHonorsEligibleBrewFormulaPresence` (unknown); `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/{bootstrap_timed_out,bootstrap_runner_error,bootstrap_unclassified_non-zero}` |
| Brew presence handling preserves mixed-plan execution | Mixed plan remains ordered | PASS | `TestRunApplyLikeConfirmedMixedBrewAptPreservesBrewPresence` (installed jq unchanged + ripgrep installed with APT, ordered) |
| Brew presence handling preserves mixed-plan execution | Bootstrap uses the same conservative guard | PASS | `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/bootstrap_*`; `TestRunBootstrapConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller` |
| Confirmed execution honors already-installed plan steps (MODIFIED) | Confirmed present tool remains undispatched | PASS | `TestRunnerSkipsOnlyEligibleAlreadyInstalledSteps`; `TestRunApplyAndBootstrapSkipDetectedCommandPresence` |
| Confirmed execution honors already-installed plan steps (MODIFIED) | Confirmed present Brew package is undispatched | PASS | `TestRunApplyConfirmedBrewPresentUsesInjectedRunnerForBrewOnly`; `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/{apply_installed,bootstrap_installed}` |
| Confirmed execution honors already-installed plan steps (MODIFIED) | Uncertain Brew package is not undispatched as installed | PASS | `TestApplyBrewFormulaPresenceCopiesPlanAndAddsUnknownAttention`; `TestRunBootstrapConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller` |
| No-op and dry-run modes remain non-mutating (MODIFIED) | Default mode does not probe Brew | PASS | `TestRunBootstrapDefaultAndDryRunDoNotProbeBrew/default`; `TestRunApplySafeModesDoNotInstantiateRealExecution` |
| No-op and dry-run modes remain non-mutating (MODIFIED) | Dry-run does not probe Brew | PASS | `TestRunBootstrapDefaultAndDryRunDoNotProbeBrew/default_--dry-run`; `TestRunApplySafeModesDoNotReportConfirmedIdempotencySkip` |
| None (REMOVED) | _(no scenarios; no requirement is removed by this change)_ | PASS | Trivially satisfied — no capability is removed; diff is additive (ADDED/MODIFIED only). |

### apply-command-dry-run (6 scenarios)

| Requirement | Scenario | Result | Covering test(s) |
|---|---|---|---|
| Confirmed Brew package reports explicit no-mutation idempotency | Confirmed apply reports installed formula without mutation | PASS | `TestRunApplyConfirmedBrewPresentUsesInjectedRunnerForBrewOnly` (`already installed; no mutation attempted`) |
| Confirmed Brew package reports explicit no-mutation idempotency | Confirmed bootstrap reports installed formula without mutation | PASS | `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/bootstrap_installed` (`already_installed; no mutation attempted`) |
| Query uncertainty is visible and never authorizes installation | Missing Brew is reported conservatively | PASS | `TestRunApplyConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller`; `TestRunBootstrapConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller` |
| Query uncertainty is visible and never authorizes installation | Timeout or ambiguous result is reported conservatively | PASS | `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/{apply_timed_out,bootstrap_unclassified_non-zero}`; `TestRunApplyLikeConfirmedMixedBrewAptPreservesBrewPresence` |
| Apply safety boundaries exclude broader convergence (MODIFIED) | Brew formula presence is the only package exception | PASS | `TestRunApplyConfirmedBrewPresentUsesInjectedRunnerForBrewOnly`; `TestRunnerIgnoresPackagePresenceForNonMatchingProvider` |
| Apply safety boundaries exclude broader convergence (MODIFIED) | Non-Brew and broader checks remain excluded | PASS | `TestRunnerIgnoresAptPresenceForBrewPackage`; `TestRunApplySafeModesDoNotReportConfirmedIdempotencySkip` |
| None (REMOVED) | _(no scenarios; no requirement is removed by this change)_ | PASS | Trivially satisfied — no capability is removed; diff is additive (ADDED/MODIFIED only). |

## Correctness (task evidence)

The previously-checked-but-unproven work units are now backed by real passing tests:

- **Task 7** (confirmed `apply`/`bootstrap` composition + default/dry-run package no-probe): `TestConfirmedCommandsCheckBrewFormulaBeforeInstall` now spans both `apply` and `bootstrap` composition (installed, explicitly absent, timed out, runner error, unclassified non-zero); `TestRunBootstrapDefaultAndDryRunDoNotProbeBrew` proves default and `--dry-run` never instantiate OS command runners, Homebrew installers, or dotfiles installers for package presence.
- **Task 9** (end-to-end mixed ordering, missing Brew, runner error, unclassified non-zero, apply/bootstrap, regressions): `TestRunApplyLikeConfirmedMixedBrewAptPreservesBrewPresence` proves mixed-plan order; `TestRun{Apply,Bootstrap}ConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller` cover missing Brew; `TestConfirmedCommandsCheckBrewFormulaBeforeInstall/{bootstrap_timed_out,bootstrap_runner_error,bootstrap_unclassified_non-zero}` cover runner error and unclassified non-zero for both commands.
- **Task 10** (final formatting): `gofmt -l` on all touched Go files reports nothing (exit 0).

## Design Coherence

- The detector (`internal/state/brew_formula_detector.go`), execution guard (`internal/execution/runner.go`), and CLI wiring (`cmd/dbootstrap/main.go`) match the design's pure-core / thin-adapter boundaries; detection remains injected and read-only.
- The shared `runApplyLike` path treats `apply` and `bootstrap` identically for confirmed-mode Brew formula presence, matching the design decision and the execution-contracts "Bootstrap uses the same conservative guard" scenario. No design deviation was introduced by the remediation; production code was unchanged.

## Strict TDD Compliance: PASS

- Strict TDD is active in `openspec/config.yaml` and `apply-progress.md` records a `TDD Cycle Evidence` table plus the bootstrap coverage remediation cycle.
- Referenced tests exist and pass: `internal/state/brew_formula_detector_test.go`, `internal/execution/runner_test.go`, `cmd/dbootstrap/main_test.go`.
- Assertion quality is concrete (argv, exit codes, dispatch counts, ordered output, "no mutation attempted" wording, no-instantiation proofs). No tautologies, ghost loops, type-only, or smoke-only assertions found in the added tests.
- The previously incomplete TDD evidence for Tasks 7, 9, and 10 is now substantiated; all checked strict-TDD work units have covering runtime evidence.

## Review Workload / PR Boundary

- Forecast: single PR; chained PRs not recommended. The current change candidate (original implementation plus 70-line remediation in `cmd/dbootstrap/main_test.go`) remains below the 400-line budget.
- The bounded post-apply review validates against the live candidate tree from the change's base tree; the post-apply gate returned `allow` with `base_relationship_valid: true`. The exact base_tree and candidate_tree hashes are recorded in the bound `gentle-ai.sdd-review-binding/v1` file (authoritative) and are not duplicated here to avoid mutating the candidate tree after binding.

## Issues

- CRITICAL: none.
- WARNING: none.
- SUGGESTION: none.

## Final Verdict

**PASS.** All 11 requirements and 29 scenarios have covering passing tests, gofmt/vet/build/full suite are green, the bounded post-apply review transaction is finalized and bound, and the `gentle-ai.verify-result/v1` envelope's `evidence_revision` references the bound review's evidence hash. The change is an archive candidate. No archive, commit, or push was performed per scope.