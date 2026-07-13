# Verify Report: Brew Package Presence Idempotency

## Status: FAIL

Independent verification completed on 2026-07-12 under strict TDD. Focused and full Go tests are green, but archive readiness is blocked by an invalidated review receipt, an unformatted changed Go test, and incomplete evidence for checked strict-TDD work units.

## Structured Status and Action Context

- Change: `brew-package-presence-idempotency`
- Native artifact store: `openspec` (authoritative); project configuration is hybrid (`both`).
- Action context: `repo-local`; workspace and allowed edit root: `/home/dniebles/dniebles-bootstrap`.
- Task progress: 10/10 complete; no unchecked implementation markers matching `^\s*- \[ \]` remain.
- Native status reported `nextRecommended: review` and blocked verification because no review receipt is linked in its SDD artifacts.
- The supplied receipt `.git/gentle-ai/reviews/compact-v2/review-fd50b8bf60a1c988/review-receipt.json` exists and has `terminal_state: approved`, but live validation failed: its staged candidate cannot be derived because the staged tree does not exactly match the complete reviewed candidate.

## Spec Coverage

| Requirement area | Result | Evidence |
|---|---|---|
| Exact read-only formula query and metadata identity | PASS | Detector builds `brew list --formula <InstallMetadata.Package>` through `CommandRunner`; focused tests assert argv and timeout. |
| Conservative installed/absent/unknown classification | PASS (implementation) | Unknown suppresses dispatch; only failed exit 1 with `No such keg` is absent. |
| Ordered execution and no mutation for installed/unknown | PASS (implementation) | Runner gates eligible Brew packages before installer dispatch. |
| Confirmed apply/bootstrap-only composition; safe modes non-probing | PARTIAL | Shared `runApplyLike` implements the boundary, but required bootstrap and package-safe-mode test evidence is absent. |
| No broader provider/convergence scope | PASS | Diff is limited to the planned detector, execution guard, CLI wiring, tests, README, and artifacts. |

## Task Completion

No unchecked implementation task lines remain in `tasks.md`.

However, these checked strict-TDD tasks lack the required test evidence:

- **CRITICAL:** Task 7 claims confirmed `apply` **and `bootstrap`** composition plus default/dry-run package no-probe tests. `cmd/dbootstrap/main_test.go` adds only `TestConfirmedApplyChecksBrewFormulaBeforeInstall`; no corresponding Brew-package `bootstrap` test or default/dry-run Brew-package lookup/runner assertion was found.
- **CRITICAL:** Task 9 claims end-to-end coverage of mixed ordering, missing Brew, runner error, unclassified non-zero, `apply`/`bootstrap`, and scope regressions. The new CLI table covers only installed, explicit absent, and timeout for `apply`; the listed scenarios are not substantiated by the changed tests.
- **CRITICAL:** Task 10 claims final formatting. `gofmt -l` reports `cmd/dbootstrap/main_test.go`.

These checked-but-unproven work units are completeness blockers despite no unchecked markers.

## Strict TDD Compliance: FAIL

- Strict TDD is active in `openspec/config.yaml` and the apply progress.
- `apply-progress.md` contains a `TDD Cycle Evidence` table.
- Referenced tests exist: `internal/state/brew_formula_detector_test.go`, `internal/execution/runner_test.go`, and `cmd/dbootstrap/main_test.go`.
- GREEN remains true for focused and full tests.
- Assertion-quality audit: behavior assertions are generally concrete (argv, order, output, dispatch count); no tautologies, ghost loops, type-only assertions, CSS assertions, or smoke-only-only tests were found in the added tests.
- **CRITICAL:** The TDD evidence is incomplete for the checked CLI/bootstrap and safe-mode/mixed-error work units above. It cannot prove all claimed RED/GREEN/triangulation coverage.

## Validation Commands

| Command | Result |
|---|---|
| `go test -count=1 ./internal/state ./internal/planning ./internal/execution ./cmd/dbootstrap` | PASS |
| `go test -count=1 ./...` | PASS |
| `go vet ./...` | PASS |
| `gofmt -l cmd/dbootstrap/main.go cmd/dbootstrap/main_test.go internal/execution/runner.go internal/execution/runner_test.go internal/planning/types.go internal/state/brew_formula_detector.go internal/state/brew_formula_detector_test.go` | FAIL: `cmd/dbootstrap/main_test.go` |
| `git diff --check` | PASS |
| `gentle-ai review validate --gate pre-commit --cwd /home/dniebles/dniebles-bootstrap` | FAIL: `current repository target cannot be derived: staged tree does not exactly match the complete reviewed candidate` |

`gofmt -d cmd/dbootstrap/main_test.go` shows formatting changes in the new `TestConfirmedApplyChecksBrewFormulaBeforeInstall` table and assertions.

## Review Workload / PR Boundary

- Forecast: single PR; chained PRs not recommended. The implemented source/test slice matches that boundary.
- The current implementation diff is below the 400-line forecast when source/test changes are counted, but the approved receipt records a different high-risk, 1,134-line reviewed target.
- **CRITICAL:** The approved receipt cannot validate against the current staged tree. Do not claim the reviewed PR boundary remains approved until a maintainer resolves the invalidated receipt/scope relationship.
- **WARNING:** Untracked sibling artifact directory `openspec/changes/package-presence-idempotency/` is outside this change; keep it out of this PR unless explicitly assigned.

## Exact Blockers

1. `gofmt` failure: `cmd/dbootstrap/main_test.go` is not formatted.
2. Strict-TDD task evidence is incomplete for checked Tasks 7, 9, and 10.
3. Review receipt `review-fd50b8bf60a1c988` is approved on disk but invalidated by the live pre-commit gate because the staged tree does not exactly match the reviewed candidate.
4. Native SDD status still requires the bounded review linkage before independent final verification/archival.

Archive is not ready.
