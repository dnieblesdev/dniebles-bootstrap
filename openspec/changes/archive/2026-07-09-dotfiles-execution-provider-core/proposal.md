# Proposal: dotfiles-execution-provider-core

## Intent
Add the core local dotfiles execution provider under `internal/execution` without wiring it into `cmd/dbootstrap` yet. This first chained slice establishes safe path resolution, repository-shape validation, command-runner seams, timeout behavior, and installer mapping for selected `dotfile:<name>` plan steps so a later slice can wire confirmed `apply --yes` with a small review surface.

This proposal supersedes the broader active planning in `dotfiles-provider-execution` for the first implementation PR. CLI composition, render/output changes, and user-facing confirmed apply behavior are deferred to `wire-dotfiles-apply-yes`.

## Scope

### In scope
- `internal/execution` provider core only.
- Dotfiles base path resolver/validator using `DBOOTSTRAP_DOTFILES_DIR` when set and non-empty, failing if set empty, otherwise `~/.dotfiles`.
- Canonical symlink resolution with `EvalSymlinks` before safety validation.
- Validation that the canonical base path is an existing safe directory, not `/`, not the user's home directory, and has no silent fallback, including no fallback when the env var is set but empty.
- Local dotfiles repository shape validation for selected modules:
  - `bin/dotlink` exists under the canonical repository.
  - selected module directories exist under the canonical repository.
  - module names pass a strict allowlist: `[A-Za-z0-9._-]+`, do not start with `-`, are not `.`, `..`, absolute paths, or path traversal, and contain no path separators.
  - dotlink and module paths do not escape the canonical repository.
- CommandRunner seam only; no direct process execution from the provider.
- Dotlink args contract is explicit: `dotlink link <module...>` preserving selected module order; empty module lists fail before runner invocation.
- Bounded dotlink timeout; timeout maps to failed dotfile execution.
- `DotfilesInstaller` maps selected `dotfile:<name>` plan steps to module `<name>` only.
- Tests using fake runner/filesystem seams only; no real external commands.
- Source-safety tests preventing direct `exec.Command` and clone/pull/submodule/remote acquisition behavior.
- Preserve `internal/dotfiles` as read-only detection/advisory code.

### Out of scope
- No `cmd/dbootstrap` wiring.
- No `apply --yes` behavior change.
- No render/copy/reporting changes.
- No actual CLI execution of dotlink.
- No bootstrap entrypoint.
- No clone, pull, submodule, sparse checkout, remote acquisition, or repository synchronization.
- No symlink rollback, repair, durable tracking, or undo behavior.
- No mutation in `internal/dotfiles`.

## Affected areas
- `internal/execution` contracts and tests.
- New execution-owned dotfiles resolver/provider/installer files.
- Source-safety regression tests for `internal/execution` and `internal/dotfiles`.
- OpenSpec deltas for `execution-contracts` and `dotfiles-provider` only.

## Risks
- Dotlink can mutate user files once wired in a later slice, so this core must be explicit about command shape, cwd, timeout, and validated paths before any CLI integration.
- Symlinked dotfiles paths or dotlink files could escape the intended repository unless canonicalized and checked.
- Resource names could be abused as path traversal unless module names are defensively validated.
- Overbuilding CLI/reporting now would increase review risk and duplicate the planned second slice.

## Rollback
- Revert the new `internal/execution` dotfiles resolver/provider/installer and tests.
- Existing noop execution behavior remains unchanged because this slice does not wire the provider into `cmd/dbootstrap`.
- No host cleanup is required from this slice because tests use fakes and no CLI path invokes dotlink.

## Success criteria
- `go test ./...` passes.
- Dotfiles base path resolution is deterministic: non-empty env override or `~/.dotfiles` only when env is unset, canonicalized with `EvalSymlinks`, with no silent fallback, including no fallback when the env var is set but empty.
- Unsafe/missing base paths, missing dotlink, missing modules, invalid module names, command failure, and timeout fail safely.
- The provider issues dotlink only through `CommandRunner` with bounded timeout, canonical `Dir`, and expected executable/args.
- Installer maps `dotfile:<name>` to module `<name>` only.
- Tests prove no real commands, direct `exec.Command`, clone/pull/submodule, or remote acquisition behavior is introduced.
- `internal/dotfiles` remains read-only.
