# Apply Progress: dotfiles-execution-provider-core

## Structured status consumed

- change: `dotfiles-execution-provider-core`
- artifactStore: `both` (OpenSpec + Engram); OpenSpec artifacts read directly, Engram read attempted but provider was unavailable.
- applyState: ready/proceeding for first chained slice based on parent preflight and OpenSpec artifacts.
- actionContext: auto mode, workspace `/home/dniebles/dniebles-bootstrap`; allowed scope limited to `internal/execution`, source-safety tests for `internal/dotfiles`, and SDD artifacts.
- strictTDD: active; test runner `go test ./...`.
- workload gate: chained strategy approved; first chained PR only; 400-line budget risk Medium; no local line limit per parent, scope kept to execution core.

## TDD Cycle Evidence

| Task | RED evidence | GREEN evidence | TRIANGULATE/REFACTOR evidence |
|---|---|---|---|
| Base path resolver tests | `go test ./internal/execution ./internal/dotfiles` failed to compile with undefined `DotfilesBaseResolver`, `DotfilesBaseSource`, etc. | Added `dotfiles_base.go` with env/home resolution, symlink canonicalization, and base safety validation. | Focused and full suites pass. |
| Local provider tests | Same RED run failed with undefined `LocalDotfilesProvider`, `DefaultDotlinkTimeout`. | Added `dotfiles_provider.go` with module validation, path containment, dotlink request construction, bounded timeout, and runner-only execution. | Focused and full suites pass. |
| Installer tests | Same RED run failed with undefined `NewDotfilesInstaller`. | Added `dotfiles_installer.go` mapping `dotfile:<name>` to the single module name and returning installed/failed results. | Focused and full suites pass. |
| Source-safety/regression tests | Initial source-safety test caught existing comment text in `internal/dotfiles`; test was refined to scan uncommented source for behavior tokens. | Added source-safety assertions for new dotfiles core and read-only `internal/dotfiles` boundary. | Focused and full suites pass. |

## Completed tasks and persisted checkbox updates

All persisted tasks in `openspec/changes/dotfiles-execution-provider-core/tasks.md` are marked `- [x]`:

- [x] RED — add base path resolver tests in `internal/execution`
- [x] RED — add local provider tests with fake filesystem and fake runner
- [x] RED — add installer mapping tests
- [x] RED — add source-safety/regression tests
- [x] GREEN — implement minimal resolver/provider/installer
- [x] TRIANGULATE — run focused and full tests

## Files changed

- `internal/execution/dotfiles_base.go`
- `internal/execution/dotfiles_base_test.go`
- `internal/execution/dotfiles_provider.go`
- `internal/execution/dotfiles_provider_test.go`
- `internal/execution/dotfiles_installer.go`
- `internal/execution/dotfiles_installer_test.go`
- `internal/execution/dotfiles_source_safety_test.go`
- `openspec/changes/dotfiles-execution-provider-core/tasks.md`
- `openspec/changes/dotfiles-execution-provider-core/apply-progress.md`

## Test commands run

1. RED: `go test ./internal/execution ./internal/dotfiles` — failed as expected on undefined resolver/provider/installer symbols.
2. GREEN/focused: `go test ./internal/execution ./internal/dotfiles` — passed.
3. Full strict suite: `go test ./...` — passed.
4. Final confirmation: `go test ./internal/execution ./internal/dotfiles && go test ./...` — passed.

## Deviations from design

- No product-scope deviations. The slice stayed under `internal/execution` for implementation and did not touch CLI wiring, renderers, apply mode selection, or `internal/dotfiles` production code.
- Source-safety test strips full-line comments before scanning `internal/dotfiles`, because existing read-only package comments mention forbidden acquisition words as negative guarantees.

## Remaining tasks

No unchecked implementation tasks remain in this slice.

Deferred by design:

- `cmd/dbootstrap` composition seams and provider wiring.
- `apply --yes` behavior change.
- User-facing render/report/copy changes for canonical dotfiles base path/modules.
- Actual CLI execution of dotlink.

## Workload / PR boundary

First chained PR boundary only: core resolver/provider/installer plus tests and source-safety regressions. No CLI behavior change, no real dotlink invocation from CLI, no repository acquisition behavior.

## Persistence notes

- OpenSpec tasks and apply progress were written locally.
- Engram reads were attempted before implementation but failed because the Engram HTTP server was unavailable at `http://127.0.0.1:7437`; Engram persistence will be attempted before return and reported accurately.

## Post-apply review fixes

Fresh post-apply review found two safety gaps and one test gap; all were fixed before verify:

- Provider-held `ResolvedDotfilesBase` values are now canonicalized and revalidated before command construction, so direct base injection cannot bypass resolver safety checks.
- Home-directory rejection now compares against a canonicalized home path; tests cover a selected base resolving to the canonical home while the raw home path differs.
- Provider tests now cover unsafe injected bases (`/`, home, alias-to-home, relative, missing, non-directory) failing before runner invocation.
- Provider tests now cover a safe injected alias resolving to a canonical repo and assert the runner receives canonical executable and `Dir` paths.

Additional verification after these fixes:

- `go test ./internal/execution ./internal/dotfiles` — passed.
- `go test ./...` — passed.
- Fresh reliability re-review of the injected-base/canonicalization fix — no findings.
