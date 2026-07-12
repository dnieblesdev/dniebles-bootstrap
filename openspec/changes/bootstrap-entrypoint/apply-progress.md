# Apply Progress: Bootstrap CLI Entrypoint

**Status:** Complete — 9/9 tasks
**Mode:** Strict TDD
**Delivery:** `exception-ok` / accepted `size:exception` (420 changed lines; no commit, push, branch, issue, or PR)

## Completed Tasks

- [x] 1.1–1.3 Shared `apply`/`bootstrap` dispatch, command-aware usage, and unchanged renderer contract.
- [x] 2.1–2.3 Shared syntactic and semantic planning paths, safety modes, and partial-failure behavior.
- [x] 3.1–3.3 Help/no-probe, mode, semantic-target, prerequisite, and partial-failure parity coverage.

## Correction Applied

- Restored historical `apply -h` and `apply --help` behavior: both remain flag-parser usage failures written to stderr with `exitUsage`.
- Kept successful standalone command help exclusive to `bootstrap -h` and `bootstrap --help`.
- Added missing parity coverage for a syntactically valid unknown resource and for catalog, configuration, and environment prerequisite paths. The comparisons assert output, exit status, detector probes, and command-runner calls.
- Strengthened the confirmed partial-failure parity test to explicitly assert
  `package:first [changed]` before `package:second [failed]` in both `apply`
  and `bootstrap` reports.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| 1.2 | `cmd/dbootstrap/main_test.go` | CLI unit | `go test ./cmd/dbootstrap` passed before correction | `TestRunApplyHelpRetainsParserUsageFailure` failed: `apply -h/--help` returned 0 to stdout | Guarded standalone help to `bootstrap` only; focused test passed | Both `-h` and `--help` aliases | One shared runner retained |
| 2.2 | `cmd/dbootstrap/main_test.go` | CLI parity | Same baseline | `TestRunBootstrapMatchesApplyForUnknownResource` written before validation | Passed against the existing shared semantic path | Unknown profile and unknown resource | No production change needed |
| 3.3 | `cmd/dbootstrap/main_test.go` | CLI parity | Same baseline | Prerequisite parity cases written before validation | Passed after correcting the test fixture to use the catalog's top-level `os` field | Catalog load, missing config, environment mismatch, and existing ordered partial failure | Reused injected seams; no pipeline/provider changes |
| 3.3 follow-up | `cmd/dbootstrap/main_test.go` | CLI parity | Existing focused suite was green | Added explicit per-command relative report-order assertions | Passed with no production change; existing shared runner preserves order | `apply` and `bootstrap` both require changed-before-failed output | None needed |

## Verification Evidence

- Focused: `go test ./cmd/dbootstrap` — passed.
- Full: `go test ./...` — passed.
- Static analysis: `go vet ./...` — passed.
- Formatting: `gofmt -w cmd/dbootstrap/main_test.go`.
- Diff validation: `git diff --check` — passed.

## Files Changed

- `cmd/dbootstrap/main.go` — preserves `apply` parser help semantics while leaving `bootstrap` command help successful.
- `cmd/dbootstrap/main_test.go` — adds alias regressions and missing semantic/prerequisite parity cases; explicitly verifies partial-failure report order for both commands.
- `openspec/changes/bootstrap-entrypoint/tasks.md` — adds truthful completion evidence.
- `openspec/changes/bootstrap-entrypoint/apply-progress.md` — records cumulative hybrid implementation evidence.

## Deviations / Risks

None. `cmd/dbootstrap/render.go` remains unchanged. The two command names use the same pipeline; partial execution remains rerun-oriented and non-transactional.
