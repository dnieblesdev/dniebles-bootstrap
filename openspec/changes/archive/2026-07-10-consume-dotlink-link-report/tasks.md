# Tasks: Consume Dotlink Link Report

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 550–750 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: strict report parser, provider reconciliation, fixtures, and focused tests → PR 2: execution translation, rendering, mode/exit integration, and full-suite tests |
| Delivery strategy | local commits |
| Chain strategy | two ordered local work units |

Decision needed before apply: No — user selected local commits
Chained PRs recommended: Yes
Chain strategy: two ordered local work units
400-line budget risk: High

The estimate includes seven production areas, parser/provider/installer/CLI integration, table-driven tests, and JSON fixtures. Keep each PR independently testable and revertible; do not split the parser contract from its provider boundary tests.

## Implementation Tasks

### 1. Establish RED coverage and fixture inventory

- [x] Add table-driven tests in `internal/execution/dotlink_report_test.go`, `internal/execution/dotfiles_provider_test.go`, `internal/execution/dotfiles_installer_test.go`, and focused CLI tests under `cmd/dbootstrap/` for the approved scenarios before implementing the new behavior.
- [x] Add deterministic fixtures under `internal/execution/testdata/dotlink-report/` for: all changed, all unchanged, mixed changed/unchanged, failed, rolled_back, duplicate top-level key, duplicate nested entry/cause/failure/rollback keys, malformed JSON, unknown/schema-mismatch fields, trailing data, status/exit mismatch, and base-resolution failure context.
- [x] Use only injected fake `CommandRunner` implementations and `t.TempDir()` for filesystem setup; tests MUST never invoke real `dotlink`, use a real home directory, or parse stderr/human output.
- [x] Verify the RED suite fails for the missing parser, report details, reconciliation, rendering, and safe-mode behavior, then record the focused command used: `go test ./internal/execution ./cmd/dbootstrap`.

### 2. Implement the strict JSON v1 boundary (PR 1)

- [x] Create `internal/execution/dotlink_report.go` with `ParseDotlinkLinkReport(stdout []byte, selected []string) (DotlinkLinkReport, error)` and safe invalid-report errors.
- [x] Implement a recursive token/object scanner that rejects duplicate keys at every object depth, validates one top-level JSON value, and rejects malformed/trailing input before typed translation.
- [x] Decode with `json.Decoder.DisallowUnknownFields`, require schema version 1 and EOF, and validate selected-module order/count, module membership, entry identity/order constraints, required source/target/cause data, known statuses/outcomes, rollback invariants, and success/failed semantic contradictions.
- [x] Preserve ordered entries and validated safe failure/rollback data in concrete domain values; do not expose raw stdout or stderr in safe errors.
- [x] Run GREEN parser tests against every required fixture, including duplicate top-level and nested keys independently.

### 3. Reconcile command status with validated reports (PR 1)

- [x] Update `internal/execution/command.go`, `internal/execution/provider.go`, and `internal/execution/dotfiles_provider.go` so the runner exposes the required stdout and command-status metadata without expanding the seam beyond what tests need.
- [x] In `internal/execution/dotfiles_provider.go`, invoke exactly `dotlink link --report=json MODULE...` once, parse present stdout regardless of exit status, preserve valid failed reports from non-zero exits, and return generic safe failures for absent/invalid reports.
- [x] Reject success-report/non-success-exit and failed-report/success-exit combinations as safe inconsistency failures; never parse stderr or human-readable stdout as fallback.
- [x] Preserve prerequisite and base-resolution errors separately from report-consumption errors so later translation can attach safe context.
- [x] Run provider tests with fake runners and assert exact argv, runner call count, stdout-first behavior, stderr non-use, and all status/exit reconciliation cases.

### 4. TRIANGULATE parser/provider behavior and harden the boundary (PR 1)

- [x] Add adversarial table cases for duplicate keys nested in each documented object, schema/type mismatch, unknown fields, trailing documents, unselected/duplicate modules, incomplete entry coverage, missing causes, and contradictory rollback state.
- [x] Assert invalid reports cannot produce per-link details and cannot trigger any compatibility parser; assert valid failed reports retain all validated details even with command failure.
- [x] Run `go test ./internal/execution` and `go vet ./...`; inspect errors for sensitive raw-output leakage and accidental command retries/acquisition.
- [x] Refactor only after the above passes: keep scanner, wire decoding, semantic validation, and reconciliation responsibilities separated and idiomatic Go error wrapping intact.
- [x] Corrective PR 1: reject `timed_out` and `not_run` command statuses even when stdout contains parseable failed JSON; only a completed `failed` command status may reconcile with a failed report.

### 5. Add execution-owned outcome types and translation (PR 2)

- [x] Extend `internal/execution/types.go` with the per-link outcome enum/detail, ordered `StepResult` detail collection, safe aggregate failure, rollback detail, and `DotfilesBaseDiagnostic` fields while preserving legacy aggregate `StepStatus` behavior for ordinary installers.
- [x] Update `internal/execution/dotfiles_installer.go` to translate validated reports into `StepResult`: all unchanged → skipped; changed plus unchanged with no failures → installed; failed/rolled_back or aggregate failure → failed; generic/provider errors → failed with no inferred links.
- [x] Update `internal/execution/noop.go` and `internal/execution/provider_aware_installer.go` as needed so default/dry-run results remain `not_implemented` and non-dotfile installers remain compatible.
- [x] Add RED→GREEN table tests in `internal/execution/dotfiles_installer_test.go` covering changed-only, unchanged-only, mixed, failed, rolled_back, aggregate failed, ordered details, causes, rollback details, and generic failure behavior.

### 6. Carry and render base diagnostics and per-link details (PR 2)

- [x] Update `internal/execution/dotfiles_provider.go` and `internal/execution/dotfiles_installer.go` to populate source, attempted candidate, selected modules, safe cause, and canonical path only after successful symlink canonicalization and safety validation.
- [x] Update `cmd/dbootstrap/render.go` to render the aggregate module result first, then every validated link with outcome/source/target and available safe cause or rollback detail; preserve failed aggregate classification and summary counts.
- [x] Ensure unresolved candidates are labeled only as `attempted candidate`, never `canonical base`; validated paths may use `canonical base`.
- [x] Add focused renderer tests for changed, unchanged, failed, rolled_back, rollback breakdown, base resolution failure, and canonical success wording. Use deterministic output assertions or golden files only through the repository’s update workflow.

### 7. Preserve safe modes and confirmed exit behavior (PR 2)

- [x] Update `cmd/dbootstrap/main.go` and relevant execution/CLI tests so only confirmed `--yes` reaches the configured dotfiles provider; default and `--dry-run` remain non-mutating, make zero runner calls, and return `not_implemented`.
- [x] Ensure confirmed failed aggregates render their detail before returning non-zero, including failed reports, rolled_back entries, command failures, invalid reports, reconciliation failures, and base/prerequisite failures.
- [x] Test `--dry-run --yes` usage rejection and verify no acquisition, retry, remote access, Bootstrap rollback, or human-output parsing was introduced.

### Corrective PR 2 reliability coverage

- [x] Add RED→GREEN coverage proving aggregate report failure overrides changed/unchanged entries and maps the module result to failed.
- [x] Add confirmed CLI coverage using `rolled-back.json` proving ordered rollback details render before non-zero exit.

### 8. Full verification and planning-folder supersession (PR 2 / archive boundary)

- [x] Run `go test ./internal/execution ./cmd/dbootstrap`, then `go test ./...`, `go test -cover ./...`, and `go vet ./...`; format changed Go files with `gofmt` and review the diff for accidental scope expansion.
- [x] Confirm all required fixtures and fake-runner tests are present, deterministic, and do not require a real Dotlink binary or home directory.
- [x] Treat `openspec/changes/dotfiles-base-failure-context/` as superseded planning input: do not implement its work or merge its requirements; if the directory exists, leave cleanup/movement to the archive phase and report any untracked contents there rather than deleting them during apply.
- [x] Before merge, verify rollback boundaries: reverting PR 1 removes report consumption without changing Dotlink/filesystem ownership; reverting PR 2 removes translation/rendering integration while preserving the parser/provider contract.
