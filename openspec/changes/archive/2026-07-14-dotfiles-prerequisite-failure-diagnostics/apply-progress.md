# Apply Progress: Dotfiles Prerequisite Failure Diagnostics

## Outcome

Completed one autonomous strict-TDD work unit. The execution contract now carries lexical prerequisite candidates separately from typed causes; confirmed missing-runner apply failures are non-zero, preserve `dotfile:bash` and `bash`, render a bounded terminal-safe diagnostic, and make zero runner calls.

## Completed Tasks

- [x] 1.1–1.2 Typed phase, prerequisite target, and multi-cause unwrap contract.
- [x] 2.1–2.4 Prerequisite candidate capture, no-runner validation, and phase retention across command/report outcomes.
- [x] 3.1–3.3 Curated rendering, deterministic base snapshot comparison, and confirmed missing-runner apply acceptance.
- [x] 4.1–4.2 Focused/full verification and isolated in-process CLI runtime harness.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| 1.1–1.2 | `internal/execution/types_test.go` | Unit | `go test ./internal/execution ./cmd/dbootstrap` passed | Missing target/phase symbols failed to compile | Target, phase, and multi-error unwrap passed | Runner target plus independent prerequisite/execution/parse causes | Clean |
| 2.1–2.4 | `internal/execution/{dotfiles_provider,dotfiles_installer}_test.go` | Unit | Focused baseline passed | Missing runner/module target transport failed to compile | Missing, escaping, installer no-link, and command/report phase tests passed | Missing runner, missing module, escaping module, malformed/absent/contradictory report cases | Clean |
| 3.1–3.2 | `cmd/dbootstrap/render_test.go` | Unit | Focused baseline passed | Curated prerequisite output and differing base causes failed | Phase/target/cause rendering and cause-aware snapshot comparison passed | Runner target, control sanitization, equal and unequal snapshots | Clean |
| 3.3 | `cmd/dbootstrap/main_test.go` | Integration | Focused baseline passed | Expected `prerequisite validation` label failed | Confirmed missing-runner CLI harness passed | Existing missing base/module/execution/report regression matrix remained green | Clean |
| 4.1–4.2 | Existing focused suites | Verification | N/A | N/A | Full suite, vet, build, and formatting passed | N/A | None needed |

## Work Unit Evidence

| Evidence | Result |
|---|---|
| Focused tests | `go test ./internal/execution ./cmd/dbootstrap` — exit 0; both packages passed. |
| Full tests | `go test ./...` — exit 0; all tested packages passed. |
| Static/build/format | `go vet ./...`, `go build ./...`, and `gofmt -w` on eight changed Go files — all exit 0. |
| Runtime harness | `go test ./cmd/dbootstrap -run TestRunApplyConfirmedMissingDotlinkRendersPrerequisiteDiagnostics` — exit 0; exercises confirmed `apply --yes --resource dotfile:bash` against a temp missing-runner fixture, verifies non-zero result, diagnostic facts, and zero runner calls. |
| Rollback boundary | Revert only `internal/execution/types.go`, `internal/execution/dotfiles_provider.go`, `cmd/dbootstrap/render.go`, and their focused tests. Existing prerequisite rejection and command/report behavior remain otherwise unchanged. |

## Judgment Day Round 1 Correction Evidence

| Evidence | Result |
|---|---|
| Scope | JD-002 only; no JD-003/provider/planning/configuration changes. |
| RED | Added oversized prerequisite and base attempted-candidate renderer tests before the bound; the unbounded base-candidate test failed as expected. |
| GREEN | `go test ./cmd/dbootstrap -run TestRenderLinkDetailsBoundsOversized` — exit 0; both focused oversized-candidate tests passed. |
| Required verification | `go test ./cmd/dbootstrap` — exit 0; `go test ./internal/execution ./cmd/dbootstrap` — exit 0. |
| Bound | Newly rendered prerequisite/base diagnostic fields are terminal-escaped, then deterministically limited to 4096 bytes with `...[truncated]`; existing stderr rendering is unchanged. |
| Rollback boundary | Revert this round's changes in `cmd/dbootstrap/render.go` and `cmd/dbootstrap/render_test.go`; prior prerequisite diagnostics remain intact. |

## Native R3-001 Correction Evidence

| Evidence | Result |
|---|---|
| Scope | R3-001 only; test and evidence changes with no production, recovery, provider/installer, planning/configuration, AttentionReasons, JD-003, or R4-001 changes. |
| RED | `go test ./cmd/dbootstrap -run TestRenderLinkDetailsRendersCommandAndReportFailureDiagnostics -count=1` — exit 1; report-validation output was `"   phase: report-validation\n"` and missed `cause: invalid dotlink report` while the typed parse-error fixture was deliberately absent. |
| GREEN | `go test ./cmd/dbootstrap -run TestRenderLinkDetailsRendersCommandAndReportFailureDiagnostics -count=1` — exit 0; command-execution/command-failure and report-validation/invalid-report labels passed. |
| Required verification | `go test ./cmd/dbootstrap -run TestRenderLinkDetailsRendersCommandAndReportFailureDiagnostics`, `go test ./cmd/dbootstrap`, and `go test ./internal/execution ./cmd/dbootstrap` — all exit 0. |
| Production hash check | Every tracked non-test Go file matched its pre-correction SHA-256 hash. |
| Correction budget | This correction adds 48 changed lines, taking the native correction delta from 94 to 142 of 200. |
| Runtime harness | N/A: this proof-only correction exercises the shared external renderer directly and changes no runtime behavior. |
| Rollback boundary | Revert only the R3-001 test in `cmd/dbootstrap/render_test.go` and this evidence block; production behavior remains unchanged. |

## Scope and Delivery

- Delivery: the original corrected code/tests totaled 378 additions+deletions; this R3-001 proof adds 31 test lines, while the separate native correction delta remains 142 of 200.
- Excluded: legacy `DotfilesBaseReporter` work, provider/parser redesign, statuses, planning/configuration, monolith cleanup, and `PlanStep.AttentionReasons -> StepResult`.
