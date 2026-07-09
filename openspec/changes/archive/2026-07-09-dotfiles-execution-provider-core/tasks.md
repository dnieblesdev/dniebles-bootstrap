# dotfiles-execution-provider-core Tasks

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 250-400 |
| 400-line budget risk | Medium |
| Chained PRs recommended | Already approved |
| This slice | `internal/execution` resolver/provider/installer + tests |
| Deferred slice | `wire-dotfiles-apply-yes`: CLI wiring/report copy/tests |
| Delivery strategy | first chained PR |

This task list supersedes the broader `dotfiles-provider-execution` task plan for the first implementation slice.

## Tasks

- [x] **RED — add base path resolver tests in `internal/execution`**
   - Cover `DBOOTSTRAP_DOTFILES_DIR` when set and non-empty, set but empty (must fail with no home fallback), and `$HOME/.dotfiles` when env is unset.
   - Cover canonical symlink resolution with `EvalSymlinks`.
   - Cover unsafe failures: empty/unresolved path, relative path, missing path, non-directory, `/`, home directory itself, and no silent fallback, including no fallback when the env var is set but empty.
   - Use fake home/stat/eval seams only.

- [x] **RED — add local provider tests with fake filesystem and fake runner**
   - Cover successful validation of canonical base, contained `bin/dotlink`, and selected module directories.
   - Assert command shape directly: executable `<base>/bin/dotlink`, args exactly `link <module...>` preserving selected module order, `Dir: <base>`, bounded timeout; empty module list fails before runner invocation.
   - Cover missing dotlink, dotlink escaping base, missing module, module escaping base, invalid module names including leading `-` and names outside `[A-Za-z0-9._-]+`, missing runner, command failure, and timeout.
   - Assert validation failures do not call the runner.

- [x] **RED — add installer mapping tests**
   - Cover selected `dotfile:<name>` maps to module `<name>` only.
   - Cover non-dotfile plan steps are rejected/unsupported by this installer.
   - Cover provider success maps to installed/succeeded result and provider error maps to failed result.
   - Assert catalog metadata is not used as command input.

- [x] **RED — add source-safety/regression tests**
   - Assert the new dotfiles execution core does not directly call `exec.Command`.
   - Assert execution source/tests do not introduce clone/pull/submodule/fetch/remote acquisition behavior.
   - Assert `internal/dotfiles` remains read-only and does not contain dotlink execution behavior.

- [x] **GREEN — implement minimal resolver/provider/installer**
   - Add new `internal/execution` files for base path resolution, local provider, dotlink command construction, and installer.
   - Keep all process execution behind `CommandRunner`.
   - Validate base path, dotlink, module names, path containment, and timeout behavior before returning success.
   - Do not touch `cmd/dbootstrap`, renderers, or apply mode selection in this slice.

- [x] **TRIANGULATE — run focused and full tests**
   - Run focused package tests for `internal/execution` and `internal/dotfiles`.
   - Run strict suite: `go test ./...`.
   - If the diff approaches the review budget, trim helper duplication or wording without adding CLI scope.

## Explicitly deferred

- `cmd/dbootstrap` composition seams and provider wiring.
- `apply --yes` behavior change.
- User-facing render/report/copy changes for canonical base path/modules.
- Actual CLI execution of dotlink.
