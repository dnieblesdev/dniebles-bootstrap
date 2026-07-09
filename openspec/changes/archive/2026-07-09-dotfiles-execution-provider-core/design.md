# Design: dotfiles-execution-provider-core

## Overview

Add the dotfiles execution core under `internal/execution` only. This slice creates safe, testable building blocks for a future `apply --yes` wiring slice, but it does not change CLI behavior.

The core flow is:

1. Resolve a raw base path from `DBOOTSTRAP_DOTFILES_DIR` when set and non-empty; if it is set empty, fail safely with no fallback; if unset, use `$HOME/.dotfiles`.
2. Canonicalize the chosen path with `EvalSymlinks`.
3. Validate the canonical base path as a safe existing directory.
4. Validate local repository shape for selected modules: `bin/dotlink` and module directories are under the canonical base.
5. Map selected `dotfile:<name>` plan steps to module `<name>`.
6. Build a bounded-timeout dotlink `CommandRequest` and send it through `CommandRunner` only.

No `cmd/dbootstrap` composition, reporting, or confirmed apply behavior changes are included here.

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Package boundary | Implement resolver/provider/installer in `internal/execution`. | Keeps mutation-capable behavior out of read-only `internal/dotfiles`. |
| Base path source | `DBOOTSTRAP_DOTFILES_DIR` when set and non-empty; set-empty fails safely with no fallback; unset uses `$HOME/.dotfiles`. | Supports explicit override and the existing convention without hiding misconfiguration. |
| Canonicalization | Use `EvalSymlinks` before validating or building command paths. | Lets `~/.dotfiles` be a symlink while ensuring execution uses the real safe path. |
| Unsafe paths | Reject empty/unresolved/relative/non-directory paths, `/`, and the home directory itself. | Avoids ambiguous or overly broad mutation roots. |
| Repository shape | Require dotlink under `<base>/bin/dotlink` and each selected module as a directory under `<base>/<module>`. | Ensures the local repo is sufficient before command execution; no acquisition fallback. |
| Command seam | Use injected `execution.CommandRunner`; never direct `exec.Command` or shell execution. | Keeps tests hermetic and centralizes process execution. |
| Timeout | Every dotlink request has a bounded timeout; timeout becomes failure. | Prevents indefinite hangs. |
| Module mapping | `planning.PlanStep.Ref.Name` for `ResourceKindDotfile` only. | Prevents catalog metadata or shell text from becoming command input. |
| CLI wiring | Deferred to `wire-dotfiles-apply-yes`. | Keeps this PR within the approved core-only review slice. |

## Components

### Dotfiles base path resolver

Add an execution-owned resolver such as `ResolveDotfilesBasePath` with small injectable seams for environment/home/stat/symlink behavior.

Required behavior:
- If `DBOOTSTRAP_DOTFILES_DIR` is set and non-empty, resolve only that path.
- If `DBOOTSTRAP_DOTFILES_DIR` is set but empty, fail safely and do not fallback to `$HOME/.dotfiles`.
- If unset, resolve exactly `$HOME/.dotfiles`.
- Do not try alternate paths after a failure.
- Canonicalize with `EvalSymlinks`.
- Require a clean absolute existing directory.
- Reject `/` and the home directory itself.
- Return a resolved value that records the canonical path and source (`env` or `home`).

### Local dotfiles provider

Add a provider responsible for local validation and command construction.

Required behavior:
- Validate module names before path joining: module names MUST match `[A-Za-z0-9._-]+`, MUST NOT start with `-`, and MUST reject empty names, absolute paths, `.`, `..`, path separators, and traversal.
- Validate `<canonical-base>/bin/dotlink` exists and resolves within the canonical base.
- Validate each selected module directory exists and resolves within the canonical base.
- Build exactly one dotlink request for requested modules:
  - `Executable: <canonical-base>/bin/dotlink`
  - `Dir: <canonical-base>`
  - `Args: []string{"link", module1, module2, ...}` exactly, preserving selected module order; empty module list fails before runner invocation
  - bounded timeout
- Convert command failure, timeout, or missing runner into typed/clear errors.
- Never clone, pull, fetch, sync, or acquire anything remotely.

### Dotfiles installer

Add an `Installer` implementation for dotfile plan steps.

Required behavior:
- Accept only `planning.ResourceKindDotfile` steps.
- Map `dotfile:<name>` to module `<name>`.
- Call the provider for local validation and dotlink execution.
- Return installed/succeeded status on success and failed status with clear text on validation/command/timeout errors.
- Do not inspect or execute catalog metadata.

## Safety and test strategy

Strict test command: `go test ./...`.

Tests must use fake filesystem and fake `CommandRunner` seams only. No test should invoke real dotlink or any external command.

Focused coverage:
1. Base path resolution and validation for env override, home fallback, symlink canonicalization, missing path, non-directory path, root, home directory, and no silent fallback, including no fallback when the env var is set but empty.
2. Provider validation for dotlink/module presence, path containment, invalid module names including leading `-` and names outside `[A-Za-z0-9._-]+`, missing runner, command failure, and timeout.
3. Command shape assertions for executable, args exactly `link <module...>` preserving selected module order, dir, and timeout; empty module lists must fail before runner invocation.
4. Installer mapping from selected `dotfile:<name>` to module `<name>`.
5. Source-safety checks that the core does not directly use `exec.Command` and contains no clone/pull/submodule/remote acquisition path.
6. Regression checks that `internal/dotfiles` remains read-only.

## Deferred to `wire-dotfiles-apply-yes`

- `cmd/dbootstrap` composition seams.
- Confirmed `apply --yes` provider wiring.
- User-facing reporting of canonical base path and selected modules.
- Render/copy changes describing dotfile mutation scope.
- CLI tests for confirmed apply behavior.
