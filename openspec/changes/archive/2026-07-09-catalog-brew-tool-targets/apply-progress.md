# Apply Progress: catalog-brew-tool-targets

## Structured Status Consumed

- change: `catalog-brew-tool-targets`
- artifactStore: `both` (OpenSpec + Engram)
- applyState: ready (derived from parent preflight and confirmed OpenSpec/Engram artifacts)
- executionMode: `auto`
- actionContext.mode: implementation in workspace `/home/dniebles/dniebles-bootstrap`
- allowedEditRoots: not provided; edits stayed within the authoritative workspace root
- strictTDD: active, runner `go test ./...`
- workload: single PR default; forecast 40-90 lines, low risk; no decision needed before apply

## Completed Tasks

- [x] RED — exact default catalog fixture coverage
  - Persisted checkbox updated in `openspec/changes/catalog-brew-tool-targets/tasks.md`.
  - Added exact decoded `planning.Catalog` fixture assertion for `catalog/bootstrap.toml`, including resources, bundles, profiles, dependency order, OS/arch metadata, descriptions, presence metadata, install package names, and absence of optional metadata.
- [x] RED — refresh CLI/apply expectations that depend on brew-backed `tool:git`
  - Persisted checkbox updated in `openspec/changes/catalog-brew-tool-targets/tasks.md`.
  - Updated exact apply output assertions for confirmed default profile and default non-mutating `--resource tool:git`.
- [x] GREEN — change the default catalog metadata only
  - Persisted checkbox updated in `openspec/changes/catalog-brew-tool-targets/tasks.md`.
  - Changed only `tool:git` install provider from `apt` to `brew`; package, presence, OS metadata, descriptions, dependencies, bundles, and profiles remain unchanged.
- [x] TRIANGULATE — run the required test sequence
  - Persisted checkbox updated in `openspec/changes/catalog-brew-tool-targets/tasks.md`.
- [x] REFACTOR — only if tests expose cleanup needs
  - Persisted checkbox updated in `openspec/changes/catalog-brew-tool-targets/tasks.md`.
  - No production refactor was needed; final diff stayed catalog-only plus tests and SDD artifacts.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1 | `internal/catalog/toml/catalog_test.go` | Unit/fixture | ✅ `go test ./internal/catalog/toml` passed before edits | ✅ Exact default catalog shape expectation failed while `tool:git` was still apt-backed | ✅ Passed after catalog provider edit | ✅ Covered resources, bundles, profiles, metadata, and plan metadata with `go test ./internal/catalog/toml` | ➖ None needed |
| 2 | `cmd/dbootstrap/main_test.go` | CLI integration-style unit | ✅ `go test ./cmd/dbootstrap` passed before edits | ✅ Apply output expectations failed while `tool:git` was still apt-backed | ✅ Passed after catalog provider edit | ✅ Confirmed default profile and `--resource tool:git` paths with `go test ./cmd/dbootstrap` | ➖ None needed |
| 3 | `catalog/bootstrap.toml` | Fixture metadata | Covered by task 1/2 safety nets | ✅ Tests expected `brew` before catalog edit | ✅ Single metadata edit satisfied focused tests | ✅ Full suite `go test ./...` passed | ➖ None needed |

## Test Commands Run

1. `go test ./internal/catalog/toml` — passed before edits (safety net).
2. `go test ./cmd/dbootstrap` — passed before edits (safety net).
3. `gofmt -w internal/catalog/toml/catalog_test.go cmd/dbootstrap/main_test.go && go test ./internal/catalog/toml` — failed as expected in RED before catalog edit because `tool:git` was still `apt`.
4. `go test ./cmd/dbootstrap` — failed as expected in RED before catalog edit because apply output still reflected apt-backed `tool:git`.
5. `go test ./internal/catalog/toml` — passed after catalog metadata edit.
6. `go test ./cmd/dbootstrap` — passed after catalog metadata edit.
7. `go test ./...` — passed.

## Files Changed

- `catalog/bootstrap.toml`
- `internal/catalog/toml/catalog_test.go`
- `cmd/dbootstrap/main_test.go`
- `openspec/changes/catalog-brew-tool-targets/tasks.md`
- `openspec/changes/catalog-brew-tool-targets/apply-progress.md`

## Deviations from Design

None. No provider behavior, execution wiring, apt provider code, dotfiles execution, mutation path, schema shape, resources, bundles, or profile membership changed.

## Remaining Tasks

None. No unchecked `- [ ]` task lines remain in `openspec/changes/catalog-brew-tool-targets/tasks.md`.

## Workload / PR Boundary

Single PR boundary retained. Final implementation is limited to catalog metadata plus focused tests and SDD progress/task artifacts.

## Important Discoveries / Decisions

- Exact plan output did not require updates because provider metadata is not rendered by plan output.
- Default non-mutating apply for a selected brew-backed `tool:git` reports Homebrew bootstrap guidance when Homebrew is missing, while the execution step remains `not supported yet` because default mode is non-mutating.
- Confirmed apply now treats both `tool:git` and `package:ripgrep` as brew-backed and skips both as unchanged when Homebrew is missing.
