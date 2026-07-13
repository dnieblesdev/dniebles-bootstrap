# Apply Progress: Brew Package Presence Idempotency

## Status

Completed under strict TDD. All ten persisted implementation tasks are marked `- [x]` in `tasks.md`.

## Remediation (2026-07-13)

Independent verification identified a functional-evidence gap in bootstrap composition coverage for eligible Brew packages. This remediation adds the minimum test coverage to close that gap; no production code was changed.

### Bootstrap composition coverage added

- `TestConfirmedCommandsCheckBrewFormulaBeforeInstall` expanded with table-driven bootstrap cases:
  - `bootstrap explicitly absent`: proves the `brew list --formula jq` query precedes `brew install jq` and the package is reported as installed.
  - `bootstrap timed out`: proves unknown presence suppresses the installer and reports failure.
  - `bootstrap runner error`: proves a runner failure is treated as unknown and suppresses the installer.
  - `bootstrap unclassified non-zero`: proves an unrecognized non-zero exit is treated as unknown and suppresses the installer.
- New `TestRunBootstrapConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller`: proves confirmed `bootstrap --yes` with missing Brew reports unknown failure and does not instantiate the OS command runner, Homebrew installer, or dotfiles installer.
- `TestRunBootstrapDefaultAndDryRunDoNotProbeBrew` extended with `stubExecutionFactories` to prove default and `--dry-run` bootstrap do not instantiate OS command runners, Homebrew installers, or dotfiles installers for package presence.

### Production code changes

None. The shared `runApplyLike` path already treats `apply` and `bootstrap` identically for confirmed-mode Brew formula presence. The remediation only adds behavioral evidence for the bootstrap branch.

## Preserved blocked-attempt history

The initial apply attempt was blocked before implementation because the task artifact had no persisted Markdown checkboxes. No production or test files changed then. The task artifact defect was repaired before this run; authoritative native status on 2026-07-12 then reported `applyState: ready`, `nextRecommended: apply`, `artifactStore: openspec`, and repo-local edit authority rooted at `/home/dniebles/dniebles-bootstrap`.

## Completed work

- Added transient Brew formula presence states and a read-only, injected detector using exact `brew list --formula <InstallMetadata.Package>` argv with a fixed 30-second timeout.
- Installed formulas produce an ordered unchanged result and no installer dispatch; explicit `No such keg` absence retains ordinary Brew installation eligibility.
- Missing Brew, nil runner, timeout, runner failure, malformed/ambiguous results become an ordered failed result with `Homebrew formula presence could not be determined; no mutation attempted` and suppress the installer.
- Wired detection only for confirmed `apply`/`bootstrap`; default and dry-run continue without presence probing.
- Updated confirmed-rerun documentation and adapted the established missing-Brew confirmed expectation to the required conservative failure behavior.

## Persisted task checkbox updates

- [x] 1. **RED — Planning transient state and detector contract**
- [x] 2. **GREEN — Implement planning state and conservative Brew detector**
- [x] 3. **TRIANGULATE — Detector edge cases and no-host-mutation proof**
- [x] 4. **RED — Execution pre-dispatch behavior**
- [x] 5. **GREEN — Add the execution guard without changing unrelated dispatch**
- [x] 6. **TRIANGULATE — Mixed-plan order and mutation safety at the runner boundary**
- [x] 7. **RED — Confirmed CLI composition and safe-mode boundaries**
- [x] 8. **GREEN — Wire detector only into confirmed apply/bootstrap**
- [x] 9. **TRIANGULATE — End-to-end acceptance and scope regression tests** *(refreshed with additional bootstrap composition evidence)*
- [x] 10. **REFACTOR — Documentation and review-ready cleanup**

## TDD Cycle Evidence

| Cycle | RED evidence | GREEN / triangulation evidence |
|---|---|---|
| Detector | `go test ./internal/state ./internal/planning` failed on absent `BrewFormulaDetector`, `PackagePresence`, and decorator symbols. | Added fake-only detector cases for metadata identity, exact argv/timeout, explicit absence, and unknown conditions; `go test -count=1 ./internal/state ./internal/planning` passed. |
| Runner | `go test ./internal/execution` failed because installed Brew presence dispatched the installer. | Added installed/absent/unknown and malformed-boundary tests; `go test ./internal/execution` and `go vet ./internal/execution` passed. |
| CLI | `go test ./cmd/dbootstrap` failed because the first command was `brew install jq`, not the presence query. | Confirmed-mode composition now queries first, skips installed formulas, and preserves safe modes; `go test ./cmd/dbootstrap` and `go vet ./cmd/dbootstrap` passed. |
| Refactor | N/A; cleanup followed passing behavior. | `gofmt`, `go test -count=1 ./...`, `go vet ./...`, and `git diff --check` passed. |
| Bootstrap coverage remediation | Coverage remediation for existing behavior; new table cases reference existing production behavior and pass immediately because `runApplyLike` is shared between `apply` and `bootstrap`. | Added bootstrap explicit-absent, timeout, runner-error, and unclassified-non-zero cases; added missing-Brew and safe-mode no-instantiation tests; focused and full suites pass. |

## Files changed

### Original implementation

- `README.md`
- `cmd/dbootstrap/main.go`
- `cmd/dbootstrap/main_test.go`
- `internal/execution/runner.go`
- `internal/execution/runner_test.go`
- `internal/planning/types.go`
- `internal/state/brew_formula_detector.go`
- `internal/state/brew_formula_detector_test.go`
- `openspec/changes/brew-package-presence-idempotency/tasks.md`
- `openspec/changes/brew-package-presence-idempotency/apply-progress.md`

### Remediation only

- `cmd/dbootstrap/main_test.go` (+70 lines; table-driven bootstrap coverage and missing-Brew/safe-mode instantiation proofs)
- `openspec/changes/brew-package-presence-idempotency/apply-progress.md`

## Verification

### Original implementation

- `go test ./internal/state ./internal/planning` (RED: expected failure)
- `go test ./internal/execution` (RED: expected failure)
- `go test ./cmd/dbootstrap` (RED: expected failure)
- `go test -count=1 ./internal/state ./internal/planning ./internal/execution ./cmd/dbootstrap` (pass)
- `go test -count=1 ./...` (pass)
- `go vet ./...` (pass)
- `gofmt -l` on all touched Go files (clean)
- `git diff --check` (pass)

### Remediation verification (2026-07-13)

- `go test -count=1 -run 'TestConfirmedCommandsCheckBrewFormulaBeforeInstall|TestRunBootstrapConfirmedMissingBrewReportsUnknownWithoutInstantiatingHomebrewInstaller|TestRunBootstrapDefaultAndDryRunDoNotProbeBrew' ./cmd/dbootstrap -v` (pass)
- `go test -count=1 ./cmd/dbootstrap` (pass)
- `go test -count=1 ./...` (pass)
- `go vet ./...` (pass)
- `gofmt -l cmd/dbootstrap/main_test.go` (clean)
- `git diff --check` (pass)

## Workload / PR boundary

Single PR work unit. Original implementation estimated changed-line count was 356 including new untracked detector/test files, below the 400-line budget. Remediation adds 70 lines in `cmd/dbootstrap/main_test.go` only; cumulative diff remains well below 400 lines.

## Deviations and remaining work

No scope deviations. No production code changes were required for the remediation. The absent-result recognizer accepts only failed exit 1 with the exact supported `No such keg` diagnostic fragment; all other non-success outcomes remain unknown. No unchecked implementation tasks remain. Next phase: independent SDD verification re-run.

## Structured status consumed

- `changeName`: `brew-package-presence-idempotency`
- `artifactStore`: `openspec` (authoritative); this apply run also mirrors progress to Engram.
- `applyState`: `ready`
- `actionContext`: `repo-local`, workspace and allowed edit root `/home/dniebles/dniebles-bootstrap`
- `actionContext warnings`: none
