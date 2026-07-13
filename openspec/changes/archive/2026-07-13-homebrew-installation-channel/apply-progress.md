# Apply Progress: Homebrew Installation Channel

## Change

- **Name**: `homebrew-installation-channel`
- **Mode**: Strict TDD (RED → GREEN → REFACTOR)
- **Artifact store**: hybrid (OpenSpec file + Engram topic)
- **Completed phases**: Phase 1 only (tasks 1.1–1.3)
- **Status**: COMPLETED technical slice

## Scope Change and Traceability

The original change included publication scope (README updates, tap formula creation, stable-release gate, lifecycle evidence, and final verification). Those eight pending tasks (2.1–4.2) have been moved to [`publish-homebrew-stable-channel`](openspec/changes/publish-homebrew-stable-channel/tasks.md). This change now owns only the completed resolver/formula-contract technical slice.

- **Stable publication gate**: owned by `publish-homebrew-stable-channel`.
- **Physical tap/formula creation**: owned by `publish-homebrew-stable-channel`.
- **README/tap README documentation**: owned by `publish-homebrew-stable-channel`.

Do not archive this change until `publish-homebrew-stable-channel` is complete, because the publication change depends on the resolver evidence recorded here.

## Completed Tasks

- [x] 1.1 RED: extend `cmd/dbootstrap/main_test.go` table cases for explicit, XDG, home-local, and `<HOMEBREW_PREFIX>/share/dbootstrap/catalog/bootstrap.toml` precedence, absent prefix, and no existing candidates using `t.TempDir()` seams.
- [x] 1.2 GREEN: modify `cmd/dbootstrap/main.go` `catalogPathResolver` with `PathExists`, environment/home defaults, and the last-resort Homebrew candidate; preserve existing missing-catalog diagnostics and CLI flag precedence.
- [x] 1.3 REFACTOR: simplify resolver helpers/comments in `cmd/dbootstrap/main.go`, run focused tests, then `go test ./...`.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `cmd/dbootstrap/main.go` | Modified | Added `PathExists func(string) bool` seam, `fileExists` helper, and last-resort `<HOMEBREW_PREFIX>/share/dbootstrap/catalog/bootstrap.toml` candidate to `catalogPathResolver.Resolve`; preserved explicit `--catalog` and existing XDG/home precedence. |
| `cmd/dbootstrap/main_test.go` | Modified | Extended `TestResolveDefaultCatalogPath` with 9 table-driven cases covering XDG precedence, HOME fallback, Homebrew fallback, higher-priority wins, absent `HOMEBREW_PREFIX`, and no-existing-candidate behavior using injected seams. |
| `openspec/changes/homebrew-installation-channel/tasks.md` | Updated | Marked tasks 1.1–1.3 complete (`[x]`); moved tasks 2.1–4.2 to `publish-homebrew-stable-channel`. |
| `openspec/changes/homebrew-installation-channel/apply-progress.md` | Updated | This cumulative apply-progress artifact; recorded scope move. |

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1 | `cmd/dbootstrap/main_test.go` | Unit | ✅ existing `go test ./cmd/dbootstrap -count=1` passed | ✅ Extended `TestResolveDefaultCatalogPath` table; tests failed because `PathExists` seam and Homebrew candidate did not exist in production code | N/A (RED task) | N/A (RED task) | N/A (RED task) |
| 1.2 | `cmd/dbootstrap/main_test.go` | Unit | ✅ existing `go test ./cmd/dbootstrap -count=1` passed | N/A (RED completed in 1.1) | ✅ `go test ./cmd/dbootstrap -run TestResolveDefaultCatalogPath -v -count=1` passed 9/9 | ✅ 9 cases force real precedence/fallback logic; no hardcoded fake-it survives | N/A (GREEN task) |
| 1.3 | `cmd/dbootstrap/main_test.go` | Unit | ✅ `go test ./cmd/dbootstrap -count=1` passed before refactor | N/A | N/A | N/A | ✅ Simplified resolver comments/helpers; `go test ./cmd/dbootstrap -count=1`, `go test ./... -count=1`, `go vet ./...`, and `gofmt -l` all clean |

## Test Summary

- **Total tests written**: 9 resolver cases in `TestResolveDefaultCatalogPath`
- **Total tests passing**: all (`cmd/dbootstrap` suite and full `go test ./...`)
- **Layers used**: Unit (1)
- **Approval tests**: None — no refactoring of existing behavior
- **Pure functions created**: `fileExists(path string) bool`

## Evidence Reconciliation

Focused re-inspection of `TestResolveDefaultCatalogPath` found **9 actual table cases**, not 10. Every previous statement in this artifact that referenced 10 cases has been corrected to 9.

### Enumerated Cases

| # | Table case name | Subtest name (verbatim) | Result |
|---|-----------------|-------------------------|--------|
| 1 | XDG_DATA_HOME takes precedence when existing | `TestResolveDefaultCatalogPath/XDG_DATA_HOME_takes_precedence_when_existing` | PASS |
| 2 | HOME wins when XDG unset and Homebrew exists | `TestResolveDefaultCatalogPath/HOME_wins_when_XDG_unset_and_Homebrew_exists` | PASS |
| 3 | Homebrew wins when higher candidates missing | `TestResolveDefaultCatalogPath/Homebrew_wins_when_higher_candidates_missing` | PASS |
| 4 | higher priority wins over Homebrew | `TestResolveDefaultCatalogPath/higher_priority_wins_over_Homebrew` | PASS |
| 5 | XDG_DATA_HOME empty falls back to HOME | `TestResolveDefaultCatalogPath/XDG_DATA_HOME_empty_falls_back_to_HOME` | PASS |
| 6 | home resolution error returns empty | `TestResolveDefaultCatalogPath/home_resolution_error_returns_empty` | PASS |
| 7 | absent HOMEBREW_PREFIX omits Homebrew candidate | `TestResolveDefaultCatalogPath/absent_HOMEBREW_PREFIX_omits_Homebrew_candidate` | PASS |
| 8 | no existing candidates returns highest priority | `TestResolveDefaultCatalogPath/no_existing_candidates_returns_highest_priority` | PASS |
| 9 | no existing candidates without XDG returns home local | `TestResolveDefaultCatalogPath/no_existing_candidates_without_XDG_returns_home_local` | PASS |

### Exact Commands and Results Executed

```text
$ go test ./cmd/dbootstrap -run TestResolveDefaultCatalogPath -v -count=1
=== RUN   TestResolveDefaultCatalogPath
=== RUN   TestResolveDefaultCatalogPath/XDG_DATA_HOME_takes_precedence_when_existing
=== RUN   TestResolveDefaultCatalogPath/HOME_wins_when_XDG_unset_and_Homebrew_exists
=== RUN   TestResolveDefaultCatalogPath/Homebrew_wins_when_higher_candidates_missing
=== RUN   TestResolveDefaultCatalogPath/higher_priority_wins_over_Homebrew
=== RUN   TestResolveDefaultCatalogPath/XDG_DATA_HOME_empty_falls_back_to_HOME
=== RUN   TestResolveDefaultCatalogPath/home_resolution_error_returns_empty
=== RUN   TestResolveDefaultCatalogPath/absent_HOMEBREW_PREFIX_omits_Homebrew_candidate
=== RUN   TestResolveDefaultCatalogPath/no_existing_candidates_returns_highest_priority
=== RUN   TestResolveDefaultCatalogPath/no_existing_candidates_without_XDG_returns_home_local
--- PASS: TestResolveDefaultCatalogPath (0.00s)
    --- PASS: TestResolveDefaultCatalogPath/XDG_DATA_HOME_takes_precedence_when_existing (0.00s)
    --- PASS: TestResolveDefaultCatalogPath/HOME_wins_when_XDG_unset_and_Homebrew_exists (0.00s)
    --- PASS: TestResolveDefaultCatalogPath/Homebrew_wins_when_higher_candidates_missing (0.00s)
    --- PASS: TestResolveDefaultCatalogPath/higher_priority_wins_over_Homebrew (0.00s)
    --- PASS: TestResolveDefaultCatalogPath/XDG_DATA_HOME_empty_falls_back_to_HOME (0.00s)
    --- PASS: TestResolveDefaultCatalogPath/home_resolution_error_returns_empty (0.00s)
    --- PASS: TestResolveDefaultCatalogPath/absent_HOMEBREW_PREFIX_omits_Homebrew_candidate (0.00s)
    --- PASS: TestResolveDefaultCatalogPath/no_existing_candidates_returns_highest_priority (0.00s)
    --- PASS: TestResolveDefaultCatalogPath/no_existing_candidates_without_XDG_returns_home_local (0.00s)
PASS
ok  	github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap	0.002s

$ go test ./... -count=1
ok  	github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap	0.157s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml	0.005s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/ci	1.090s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/config	0.005s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles	0.006s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/environment	0.005s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/execution	0.243s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/planning	0.007s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/state	0.006s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/version	1.756s
?   	github.com/dnieblesdev/dniebles-bootstrap/internal/version/cmd/normalize	[no test files]
?   	github.com/dnieblesdev/dniebles-bootstrap/internal/version/cmd/validate	[no test files]
```

## Deviations from Design

None — implementation matches the completed technical-slice design. The resolver preserves `catalog/bootstrap.toml` from the published asset (via `pkgshare.install "catalog/bootstrap.toml"`) and uses `<HOMEBREW_PREFIX>/share/dbootstrap/catalog/bootstrap.toml` as the last-resort candidate.

## Issues Found

- A prior version of this artifact incorrectly claimed 10 resolver table cases. Re-inspection of `TestResolveDefaultCatalogPath` confirms exactly 9 cases; all inconsistent references have been corrected to 9. No production code or task completion state was changed.

## Phase 1 Closure Record

- **Code scope**: `cmd/dbootstrap/main.go` (`catalogPathResolver`, `fileExists`, `defaultCatalogPath`) and `cmd/dbootstrap/main_test.go` (`TestResolveDefaultCatalogPath`).
- **Tests**: Focused `go test ./cmd/dbootstrap -run TestResolveDefaultCatalogPath -v -count=1` passes 9/9; full `go test ./... -count=1` passes.
- **Review status**: Phase 1 resolver TDD foundation is complete. No Phase 2+ work was performed in this change.
- **Publication block**: The stable Homebrew channel **cannot be published** until a real public, non-draft, non-prerelease release exists with Linux amd64 and arm64 archives plus matching SHA-256 values. A prerelease may be used only for technical validation; do **not** create a stable release merely to unblock work. Publication is owned by `publish-homebrew-stable-channel`.

## Remaining Tasks

All remaining work moved to `publish-homebrew-stable-channel`:

- [ ] 1.1 Update `README.md` with summary Linux/WSL install instructions only (detailed evidence lives in tap README).
- [ ] 1.2 In standalone repository `dnieblesdev/homebrew-dniebles-bootstrap`, create `Formula/dbootstrap.rb` from scratch after the stable gate passes, with Linux intel/arm branches, `pkgshare.install "catalog/bootstrap.toml"`, macOS pre-download `odie`, and no unsupported-CPU fallback; pin literal stable version/URL/SHA-256 values.
- [ ] 1.3 Add tap `README.md` documenting tap/install/upgrade/uninstall commands, ownership, pinned URL/version/SHA-256 fields, publication evidence, hashes, and operational proof; record that no prerelease is publishable.
- [ ] 2.1 Before creating the formula, run `gh release view <tag> --json isDraft,isPrerelease,tagName,assets`; require public non-draft/non-prerelease Linux amd64/arm64 archives and matching SHA-256 values, and validate each checksum file's content against its archive.
- [ ] 2.2 On Linux/WSL amd64 and arm64, capture `brew tap`, install, `dbootstrap --version`, strict audit/style, reinstall/upgrade, uninstall, payload cleanup, unrelated-file preservation, and arbitrary-CWD `plan --profile dev` proof.
- [ ] 2.3 Capture macOS formula output proving clear rejection before download; record release URL, version, asset names/digests, installed catalog path, and command output in the tap README. Keep publication blocked when the stable gate is absent.
- [ ] 3.1 Run focused `cmd/dbootstrap` tests followed by `go test ./...`; report test files, scenarios, and any skipped external Homebrew integration.
- [ ] 3.2 Review the diff against `publish-homebrew-stable-channel` proposal/spec/design and approved `review-ledger.md`; verify only docs, standalone tap, and evidence boundary changed.

## Workload / PR Boundary

- **Mode**: single-pr (completed slice)
- **Current work unit**: PR 1 — Add Homebrew catalog resolution (Phase 1)
- **Boundary**: This apply batch covers only the resolver TDD foundation in `cmd/dbootstrap/main.go` and `cmd/dbootstrap/main_test.go`. Documentation, tap formula, and acceptance evidence are deferred to `publish-homebrew-stable-channel`.
- **Estimated review budget impact**: Resolver-only diff is well below the 400-line budget for this slice; no size exception needed.

## Status

3 / 3 implementation tasks complete. This change is structurally ready for final verification/archive once `publish-homebrew-stable-channel` completes.
