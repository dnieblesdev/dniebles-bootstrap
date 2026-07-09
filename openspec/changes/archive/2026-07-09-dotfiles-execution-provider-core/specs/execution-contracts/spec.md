# Delta for execution-contracts

## ADDED Requirements

### Requirement: Dotfiles execution core uses an injectable command runner

The dotfiles execution core in `internal/execution` MUST route dotlink invocation through an injected `CommandRunner` seam.
The core MUST NOT invoke `exec.Command`, a shell, or any real external command directly.
Tests MUST be able to substitute a fake runner and MUST NOT require real external commands.
Dotlink command requests MUST include a bounded timeout, and timeout results MUST become failed dotfile execution results.

#### Scenario: Tests use a fake runner

- GIVEN the dotfiles execution provider is under test
- WHEN a fake command runner is injected
- THEN the test can assert the requested executable, arguments, working directory, and timeout
- AND no real command is executed

#### Scenario: Direct process execution is absent

- GIVEN the dotfiles execution core source is reviewed
- WHEN command invocation behavior is inspected
- THEN the core does not call `exec.Command` or a shell directly
- AND all dotlink execution goes through `CommandRunner`

#### Scenario: Dotlink timeout fails safely

- GIVEN the provider invokes dotlink through the command runner
- WHEN the command result indicates timeout
- THEN the dotfile execution result is failed
- AND no retry, fallback acquisition, or second command is attempted

### Requirement: Dotfiles execution core validates local prerequisites only

The dotfiles execution core MUST resolve the local base path from `DBOOTSTRAP_DOTFILES_DIR` when set and non-empty, MUST fail safely with no fallback when it is set empty, and MUST resolve `~/.dotfiles` only when the env var is unset.
It MUST canonicalize the selected path with `EvalSymlinks` before validating or constructing command paths.
It MUST fail safely when the canonical base path is missing, unresolved, relative, not an existing directory, `/`, the home directory itself, or otherwise unsafe.
It MUST validate that `bin/dotlink` and selected module directories exist under the canonical repository.
It MUST validate module names before path joining or command construction: names MUST match `[A-Za-z0-9._-]+`, MUST NOT start with `-`, MUST NOT be empty, `.`, or `..`, and MUST NOT contain path separators, traversal segments, or absolute paths.
It MUST NOT silently fallback to another path after the selected source fails validation.
It MUST NOT attempt clone, pull, submodule, fetch, or other remote acquisition.

#### Scenario: Environment path is canonicalized and validated

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` is set to a dotfiles path
- WHEN the base path resolver runs
- THEN it resolves symlinks with `EvalSymlinks`
- AND validation uses the canonical directory
- AND no home fallback is attempted if validation fails

#### Scenario: Empty environment path does not fallback

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` is set to an empty value
- WHEN the base path resolver runs
- THEN resolution fails safely
- AND `$HOME/.dotfiles` is not used as fallback

#### Scenario: Home convention is used when environment is unset

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` is unset
- AND the user's home directory is known
- WHEN the base path resolver runs
- THEN it resolves exactly `$HOME/.dotfiles`
- AND validation uses the canonical directory

#### Scenario: Unsafe base path fails safely

- GIVEN the selected base path resolves to `/`, the home directory itself, a missing path, or a non-directory
- WHEN the provider validates prerequisites
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Repository shape is required before command execution

- GIVEN a safe canonical base path exists
- BUT `bin/dotlink` or a selected module directory is missing or resolves outside the canonical repository
- WHEN the provider validates prerequisites
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Unsafe module names fail safely

- GIVEN a selected module name is empty, starts with `-`, contains a path separator, is `.`, is `..`, is absolute, contains traversal, or contains characters outside `[A-Za-z0-9._-]`
- WHEN the provider validates prerequisites
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Remote acquisition is not attempted

- GIVEN dotfiles execution is invoked
- WHEN prerequisites are missing or command execution fails
- THEN no clone, pull, submodule, fetch, remote URL, or other acquisition command is requested

## MODIFIED Requirements

### Requirement: Execution contracts remain non-mutating unless explicitly wired by a caller

`internal/execution` MUST provide testable execution contracts without changing command-line behavior by itself.
This slice MAY add a dotfiles installer/provider implementation, but MUST NOT wire it into `cmd/dbootstrap` or change default, dry-run, or confirmed apply behavior.
Existing noop behavior remains available for callers that have not explicitly selected a real installer.

#### Scenario: Core provider is dormant until composed

- GIVEN the dotfiles provider and installer exist in `internal/execution`
- WHEN no caller composes them into an execution runner
- THEN no dotlink execution is possible through the CLI
- AND existing noop execution behavior is unchanged
