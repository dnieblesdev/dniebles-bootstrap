# Apply Progress: Bootstrap Entrypoint Primary Delivery Record

**Status:** Apply complete; the delivery record is ready for the independent verify phase.

This delivery-record-only change does not modify implementation, tests, or the historical `openspec/changes/bootstrap-entrypoint/` record.

## Completed focused tasks

- [x] **2.1 Current CLI and test inspection**
- [x] **2.2 Focused bootstrap entrypoint test execution**
- [x] **2.3 Discrepancy follow-up assessment** — no discrepancy observed.

## Inspection evidence

| Recorded requirement | Current implementation and test evidence |
| --- | --- |
| Command discovery and standalone help | `run` dispatches `bootstrap` to `runApplyLike("bootstrap", ...)`; `runApplyLike` handles only bootstrap `-h` and `--help` before parsing or runtime work. `TestRunBootstrapHelp` checks root discovery and both help aliases with a fatal environment-detection seam. |
| Shared validation and apply parity | `runApply` and bootstrap use `runApplyLike`; `parseApplyLikeFlags` validates syntax before `buildPlan`. Tests cover syntactic rejection, unknown profile/resource semantic parity, prerequisite parity, and ordered partial failure. |
| Non-mutating default and dry-run modes | `buildApplyRunner` returns a noop runner when the mode is not confirmed. `TestRunBootstrapMatchesApplyAcrossSafetyModes` asserts zero command-runner calls for default and `--dry-run`, and parity for confirmed and confirmed-sudo modes. |
| Historical-record boundary | `openspec/changes/bootstrap-entrypoint/` remains unmodified and untracked; no source or test file was edited in this delivery-record change. |

The current delivery record includes the authoritative change-local `specs/bootstrap-entrypoint/spec.md` specification. Inspection used that specification together with its proposal, design, and tasks, the applicable `openspec/specs/apply-command-dry-run/spec.md` contract, and the historical change only as evidence.

## Focused runtime evidence

```text
go test ./cmd/dbootstrap -run '^(TestRunBootstrapHelp|TestRunApplyHelpRetainsParserUsageFailure|TestRunBootstrapMatchesApplyAcrossSafetyModes|TestRunApplyLikeRejectsSyntacticInputBeforeProbing|TestRunBootstrapMatchesApplyForUnknownProfile|TestRunBootstrapMatchesApplyForUnknownResource|TestRunBootstrapMatchesApplyForPrerequisites|TestRunBootstrapMatchesApplyForPartialFailure)$'
ok   github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap	0.163s
```

## Archive-readiness evidence

- [x] **3.1 Requirement coverage** — the focused test matrix above covers help discovery and side-effect freedom, syntactic validation before probes, explicit-target planning, apply-equivalent semantic and prerequisite failures, default and dry-run non-mutation, confirmed-mode parity, and ordered partial-failure reporting. The full repository suite below passed. No discrepancy or follow-up was identified.
- [x] **3.2 Scope boundary** — `git status --short -- openspec/changes/bootstrap-entrypoint-primary-record/ openspec/changes/bootstrap-entrypoint/` showed both change directories as untracked. The primary record contains only its five delivery-record artifacts; `git diff --name-only` and `git diff --cached --name-only` scoped to `openspec/changes/bootstrap-entrypoint/` were empty. This apply work did not modify the historical record. Pre-existing modifications to `cmd/dbootstrap/main.go` and `cmd/dbootstrap/main_test.go` remain outside this record and were not changed.
- [x] **3.3 Archive preparation** — all apply tasks are complete. No `verify-report.md` or archive artifact was created; the next lifecycle step is independent verification.

## Full verification

```text
go test -count=1 ./...
ok   github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap  0.340s
ok   github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml  0.012s
ok   github.com/dnieblesdev/dniebles-bootstrap/internal/config  0.005s
ok   github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles  0.004s
ok   github.com/dnieblesdev/dniebles-bootstrap/internal/environment  0.004s
ok   github.com/dnieblesdev/dniebles-bootstrap/internal/execution  0.221s
ok   github.com/dnieblesdev/dniebles-bootstrap/internal/planning  0.007s
ok   github.com/dnieblesdev/dniebles-bootstrap/internal/state  0.109s
```

## TDD status

Strict TDD was not activated for this delivery-record-only apply task. No implementation or test behavior was added; existing focused and full test suites passed.
