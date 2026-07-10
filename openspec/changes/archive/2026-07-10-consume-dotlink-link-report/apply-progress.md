# Apply Progress: Consume Dotlink Link Report

## PR 1 — parser and provider reconciliation

Completed the first chained work unit only. No commit was created.

### Completed tasks

- Task 1 fixture and fake-runner safeguards are checked in `tasks.md`.
- Tasks 2, 3, and 4 are checked in `tasks.md`.
- Added strict JSON v1 parsing with recursive duplicate-key detection at every object depth, strict typed decoding, semantic validation, and safe errors.
- Added `RunDotlinkReport` as an additive provider boundary; the legacy `RunDotlink` method consumes the report, safely fails a validated failed aggregate, and does not translate it into execution/renderer detail.
- Provider invokes exactly `dotlink link --report=json MODULE...` once, parses stdout before reconciliation, ignores stderr, preserves valid failed reports from non-success command status, and fails closed for absent, invalid, or inconsistent reports.

### Files changed

- `internal/execution/dotlink_report.go`
- `internal/execution/dotlink_report_test.go`
- `internal/execution/dotfiles_provider.go`
- `internal/execution/dotfiles_provider_test.go`
- `internal/execution/provider.go`
- `internal/execution/testdata/dotlink-report/*.json`
- `cmd/dbootstrap/main_test.go` (existing fake-runner confirmed-path test updated only for JSON stdout and exact argv)
- `openspec/changes/consume-dotlink-link-report/tasks.md`
- `openspec/changes/consume-dotlink-link-report/apply-progress.md`

### TDD Cycle Evidence

| Cycle | RED evidence | GREEN / triangulation evidence | Refactor |
|---|---|---|---|
| Parser boundary | `go test ./internal/execution -run 'TestParseDotlinkLinkReport'` failed because parser symbols did not exist. | Table tests cover valid changed/unchanged/mixed/failed/rolled-back reports plus malformed, unknown, schema mismatch, trailing data, duplicate keys at each documented depth, incomplete coverage, duplicate selected modules, missing cause, and rollback contradiction. | Scanner, wire decode, and semantic validation are separate functions. |
| Provider reconciliation | `go test ./internal/execution -run 'TestLocalDotfilesProvider(BuildsExactCommand|ReconcilesCommandAndReport)'` failed because `RunDotlinkReport` and inconsistency errors did not exist; legacy failed-report safety was also added RED-first. | Fake runner tests prove exact argv, one call, stdout-first behavior, ignored stderr, valid failed non-zero report retention, safe legacy aggregate failure, and both mismatch directions. | Kept report access additive through `DotlinkReportProvider`; no execution result translation was added. |
| Integration regression | Existing confirmed CLI test failed after the provider correctly required JSON stdout. | Updated its injected fake result to supply a valid report and assert `link --report=json bash`; `go test ./...` passes. | No CLI production/reporting behavior changed. |

### Verification

- `go test ./internal/execution` — PASS
- `go test ./internal/execution ./cmd/dbootstrap` — PASS
- `go test ./...` — PASS
- `go vet ./...` — PASS
- `gofmt` applied to changed Go files.
- `git diff --check` — PASS.

### Deviations and scope control

- The specified PR 1 boundary is preserved. No `StepResult` outcome/detail types, installer translation, renderer/CLI production reporting, base-diagnostic presentation, archive/supersession cleanup, retry, acquisition, or real-dotlink execution was added.
- `CommandResult` already supplied stdout and command-status metadata, so it needed no structural expansion.
- The two Task 1 checklist lines that explicitly require renderer/safe-mode RED coverage remain unchecked because those behaviors are PR 2 and explicitly out of this work-unit scope.

### Remaining tasks

- [ ] Add table-driven tests in `internal/execution/dotlink_report_test.go`, `internal/execution/dotfiles_provider_test.go`, `internal/execution/dotfiles_installer_test.go`, and focused CLI tests under `cmd/dbootstrap/` for the approved scenarios before implementing the new behavior.
- [ ] Verify the RED suite fails for the missing parser, report details, reconciliation, rendering, and safe-mode behavior, then record the focused command used: `go test ./internal/execution ./cmd/dbootstrap`.
- [ ] Extend `internal/execution/types.go` with the per-link outcome enum/detail, ordered `StepResult` detail collection, safe aggregate failure, rollback detail, and `DotfilesBaseDiagnostic` fields while preserving legacy aggregate `StepStatus` behavior for ordinary installers.
- [ ] Update `internal/execution/dotfiles_installer.go` to translate validated reports into `StepResult`: all unchanged → skipped; changed plus unchanged with no failures → installed; failed/rolled_back or aggregate failure → failed; generic/provider errors → failed with no inferred links.
- [ ] Update `internal/execution/noop.go` and `internal/execution/provider_aware_installer.go` as needed so default/dry-run results remain `not_implemented` and non-dotfile installers remain compatible.
- [ ] Add RED→GREEN table tests in `internal/execution/dotfiles_installer_test.go` covering changed-only, unchanged-only, mixed, failed, rolled_back, aggregate failed, ordered details, causes, rollback details, and generic failure behavior.
- [ ] Update `internal/execution/dotfiles_provider.go` and `internal/execution/dotfiles_installer.go` to populate source, attempted candidate, selected modules, safe cause, and canonical path only after successful symlink canonicalization and safety validation.
- [ ] Update `cmd/dbootstrap/render.go` to render the aggregate module result first, then every validated link with outcome/source/target and available safe cause or rollback detail; preserve failed aggregate classification and summary counts.
- [ ] Ensure unresolved candidates are labeled only as `attempted candidate`, never `canonical base`; validated paths may use `canonical base`.
- [ ] Add focused renderer tests for changed, unchanged, failed, rolled_back, rollback breakdown, base resolution failure, and canonical success wording. Use deterministic output assertions or golden files only through the repository’s update workflow.
- [ ] Update `cmd/dbootstrap/main.go` and relevant execution/CLI tests so only confirmed `--yes` reaches the configured dotfiles provider; default and `--dry-run` remain non-mutating, make zero runner calls, and return `not_implemented`.
- [ ] Ensure confirmed failed aggregates render their detail before returning non-zero, including failed reports, rolled_back entries, command failures, invalid reports, reconciliation failures, and base/prerequisite failures.
- [ ] Test `--dry-run --yes` usage rejection and verify no acquisition, retry, remote access, Bootstrap rollback, or human-output parsing was introduced.
- [ ] Run `go test ./internal/execution ./cmd/dbootstrap`, then `go test ./...`, `go test -cover ./...`, and `go vet ./...`; format changed Go files with `gofmt` and review the diff for accidental scope expansion.
- [ ] Confirm all required fixtures and fake-runner tests are present, deterministic, and do not require a real Dotlink binary or home directory.
- [ ] Treat `openspec/changes/dotfiles-base-failure-context/` as superseded planning input: do not implement its work or merge its requirements; if the directory exists, leave cleanup/movement to the archive phase and report any untracked contents there rather than deleting them during apply.
- [ ] Before merge, verify rollback boundaries: reverting PR 1 removes report consumption without changing Dotlink/filesystem ownership; reverting PR 2 removes translation/rendering integration while preserving the parser/provider contract.

### Workload / PR boundary

Delivery path consumed: **auto-chain, PR 1 local work unit**. This slice is independently testable and revertible: reverting it removes report parsing/reconciliation and restores the prior provider contract without touching Dotlink ownership or PR 2 translation/rendering work.

### Structured status consumed

```yaml
changeName: consume-dotlink-link-report
artifactStore: openspec
applyState: ready
actionContext:
  mode: repo-local
  workspaceRoot: /home/dniebles/dniebles-bootstrap
  allowedEditRoots:
    - /home/dniebles/dniebles-bootstrap
warnings: []
nextRecommended: apply-next-pr2-slice
```

## Corrective PR 1 — command completion reconciliation

### Completed task

- Task 4 corrective PR 1 task is checked in `tasks.md`: failed JSON reports are trusted only when `CommandStatusFailed` confirms a completed failed command.

### TDD Cycle Evidence

| Cycle | RED evidence | GREEN / triangulation evidence | Refactor |
|---|---|---|---|
| Timed-out/not-run failed report reconciliation | Added table cases for `CommandStatusTimedOut` and `CommandStatusNotRun` with parseable `failed.json`; `go test ./internal/execution -run '^TestLocalDotfilesProviderReconcilesCommandAndReport$'` failed because both reports were returned without error. | Reconciliation now accepts successful reports only for `succeeded` and failed reports only for `failed`. The focused provider tests, `go test ./internal/execution`, and `go test ./...` pass. The existing valid `failed`/`failed` and `success`/`succeeded` cases remain covered. | Minimal condition change; no production refactor was needed. |

### Files changed

- `internal/execution/dotfiles_provider.go`
- `internal/execution/dotfiles_provider_test.go`
- `openspec/changes/consume-dotlink-link-report/tasks.md`
- `openspec/changes/consume-dotlink-link-report/apply-progress.md`

### Verification

- RED: `go test ./internal/execution -run '^TestLocalDotfilesProviderReconcilesCommandAndReport$'` — expected FAIL before the reconciliation fix.
- Focused GREEN: `go test ./internal/execution -run '^TestLocalDotfilesProvider(ReconcilesCommandAndReport|LegacyBoundaryFailsForValidatedFailedReport|CommandFailureAndTimeout)$'` — PASS.
- `go test ./internal/execution` — PASS.
- `go test ./...` — PASS.
- `gofmt -w internal/execution/dotfiles_provider.go internal/execution/dotfiles_provider_test.go` — applied.
- `git diff --check` — PASS.

### Scope and remaining work

- No PR 2 tasks (5–8) were implemented.
- The corrective slice retains the PR 1 boundary and exposes no parsed report entries when reconciliation returns an error.
- Remaining tasks are the existing unchecked PR 2 and deferred Task 1 lines in `tasks.md`.

### Workload / PR boundary

Delivery path: **corrective PR 1 slice**. This change is independently testable and limited to provider reconciliation and its focused table tests.

### Structured status consumed

```yaml
changeName: consume-dotlink-link-report
artifactStore: openspec
applyState: ready
actionContext:
  mode: repo-local
  workspaceRoot: /home/dniebles/dniebles-bootstrap
  allowedEditRoots:
    - /home/dniebles/dniebles-bootstrap
warnings: []
nextRecommended: apply-next-pr2-slice
```

## PR 2 — execution translation, diagnostics, and rendering

Completed tasks 5–8 from the clean PR 1 baseline. No commit was created.

### TDD cycle evidence

- **RED:** `go test ./internal/execution ./cmd/dbootstrap` — expected FAIL before production changes: new behavior tests referenced undefined `LinkOutcome`, `LinkDetail`, `LinkFailure`, `LinkRollback`, `DotfilesBaseDiagnostic`, and `StepResult` detail fields.
- **GREEN:** `gofmt -w internal/execution/types.go internal/execution/dotfiles_installer.go internal/execution/dotfiles_provider.go internal/execution/provider.go internal/execution/dotfiles_installer_test.go cmd/dbootstrap/render.go cmd/dbootstrap/render_test.go && go test ./internal/execution ./cmd/dbootstrap` — PASS after implementing the execution-owned translation and rendering boundary.
- **TRIANGULATE:** The focused suite covers changed-only, unchanged-only, mixed, failed, rolled-back, aggregate failure, ordered source/target preservation, safe generic provider errors without inferred links, unresolved-base wording, confirmed non-zero failure detail, and pre-existing safe-mode zero-runner / usage rejection cases.

### Implementation

- Added execution-owned ordered link details, outcomes, aggregate failure, rollback, and safe base diagnostics while retaining legacy aggregate `StepStatus` for ordinary installers.
- `DotfilesInstaller` consumes validated reports when the provider supports them: all unchanged maps to skipped; changed/mixed success maps to installed; failed, rolled-back, or aggregate-failed reports map to failed. Generic provider errors carry no inferred link details.
- `LocalDotfilesProvider` supplies attempted candidate, source, selected modules, safe cause, and canonical base only after successful resolution.
- CLI rendering prints the aggregate step first, then link details, aggregate failure, rollback data, and base diagnostics deterministically. Confirmed failed reports render details before returning non-zero.
- Default and dry-run behavior remains the existing non-mutating noop path with no command-runner construction.

### Verification

- `go test ./internal/execution ./cmd/dbootstrap` — PASS.
- `go test ./...` — PASS.
- `go test -cover ./...` — PASS.
- `go vet ./...` — PASS.
- `gofmt -w internal/execution/types.go internal/execution/dotfiles_installer.go internal/execution/dotfiles_provider.go internal/execution/provider.go internal/execution/dotfiles_installer_test.go internal/execution/dotfiles_provider_test.go cmd/dbootstrap/render.go cmd/dbootstrap/render_test.go cmd/dbootstrap/main_test.go` — applied.

### Scope and rollback boundary

- `openspec/changes/dotfiles-base-failure-context/` remains untracked and untouched as required.
- Reverting PR 2 removes only report-to-execution translation, base diagnostic presentation, and rendering integration; PR 1 parser/provider reconciliation remains intact.

## Corrective PR 2 — reliability coverage

### Completed tasks

- Added a parsed aggregate `failed` JSON report scenario with only changed/unchanged entries; after isolating the parser-required failure context, it proves aggregate status maps the dotfile module result to `StepStatusFailed`.
- Added confirmed CLI coverage using `rolled-back.json`; it asserts rollback link, aggregate failure, rollback state, and removed target render in order before the failed command returns a non-zero exit.

### TDD cycle evidence

- **RED:** `go test ./internal/execution ./cmd/dbootstrap` — expected test setup failure observed because a `status: failed` report without its required failure object is invalid under the strict parser.
- **GREEN:** Corrected the fixture to satisfy the parser contract and isolated `report.Status`; `go test ./internal/execution ./cmd/dbootstrap` — PASS. No production change was required because existing translation and rendering already satisfied both behaviors.
- **TRIANGULATE:** `go test ./...` — PASS; `go vet ./...` — PASS.

### Files changed

- `internal/execution/dotfiles_installer_test.go`
- `cmd/dbootstrap/main_test.go`
- `openspec/changes/consume-dotlink-link-report/tasks.md`
- `openspec/changes/consume-dotlink-link-report/apply-progress.md`

## Verification checkbox reconciliation

The formal verifier found two stale PR 1 Task 1 checkboxes. Both are now checked because their RED/GREEN evidence is already recorded above:

- Table-driven parser/provider/installer/CLI coverage was added before implementation across the documented test files.
- RED was observed with `go test ./internal/execution ./cmd/dbootstrap` failing on the missing parser/report detail/reconciliation symbols; the exact parser/provider RED commands and PR 2 RED command are recorded in the respective cycle sections.

Delivery metadata is also reconciled: the user selected two ordered local work-unit commits rather than a remote stacked PR strategy.
