# Delta for dotfiles-provider

## ADDED Requirements

### Requirement: Local dotfiles execution core is separate from read-only detection

The dotfiles provider capability MUST keep read-only detection behavior separate from local execution behavior.
Execution-capable code MUST live under `internal/execution` and MUST NOT add mutation behavior to `internal/dotfiles`.
Local execution core MUST NOT change planning semantics or command-line behavior in this slice.

#### Scenario: Detection remains read-only

- GIVEN dotfiles presence is being checked through `internal/dotfiles`
- WHEN detection runs
- THEN no dotlink, clone, pull, submodule, fetch, or remote acquisition is attempted

#### Scenario: Execution core remains outside detection package

- GIVEN local dotfiles execution support is implemented
- WHEN package boundaries are reviewed
- THEN dotlink command construction and command-runner use are in `internal/execution`
- AND `internal/dotfiles` remains read-only

### Requirement: Local dotfiles execution requires explicit safe prerequisites

The local dotfiles execution core MUST resolve the base path from `DBOOTSTRAP_DOTFILES_DIR` when set and non-empty, MUST fail safely with no fallback when it is set empty, and MUST resolve `~/.dotfiles` only when the env var is unset.
It MUST canonicalize symlinks with `EvalSymlinks` before validation.
It MUST fail safely when the canonical base path is missing, unresolved, unsafe, not an existing directory, `/`, or the user's home directory itself.
It MUST fail safely when `bin/dotlink` is missing or resolves outside the canonical dotfiles repository.
It MUST validate module names before path joining or command construction: names MUST match `[A-Za-z0-9._-]+`, MUST NOT start with `-`, MUST NOT be empty, `.`, or `..`, and MUST NOT contain path separators, traversal segments, or absolute paths.
It MUST fail safely when a selected module directory is missing or resolves outside the canonical dotfiles repository.
It MUST NOT attempt clone, pull, submodule, fetch, or other remote acquisition.

#### Scenario: Safe prerequisites allow provider execution

- GIVEN a selected dotfile module name
- AND a safe canonical local base path is available
- AND `bin/dotlink` exists under the canonical repository
- AND the selected module directory exists under the canonical repository
- WHEN local execution runs through the provider
- THEN dotlink may be requested through the injected command runner

#### Scenario: Empty env base path fails safely without fallback

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` is set to an empty value
- WHEN the provider resolves the base path
- THEN resolution fails safely
- AND no home fallback is attempted

#### Scenario: Missing base path fails safely

- GIVEN a selected dotfile module name
- AND no explicit or home-convention safe local base path is available
- WHEN local execution runs through the provider
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Missing dotlink fails safely

- GIVEN a safe canonical local base path is available
- AND `bin/dotlink` is not available under the canonical dotfiles repository
- WHEN local execution runs through the provider
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Missing module fails safely

- GIVEN a safe canonical local base path and dotlink are available
- AND a selected module directory is missing
- WHEN local execution runs through the provider
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Dotfiles repo symlink resolves safely

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` or `~/.dotfiles` points through a symlink
- WHEN the resolver validates the base path
- THEN validation uses the canonical destination
- AND dotlink and selected modules must remain inside that canonical repository

### Requirement: Dotfiles installer maps selected dotfile resources to module names only

The `internal/execution` dotfiles installer MUST map a selected plan step with kind `dotfile` and name `<name>` to the single module `<name>`.
It MUST NOT derive command arguments from catalog descriptions, install metadata, dependency text, shell strings, or any other field.
It MUST reject or fail non-dotfile steps when used directly.

#### Scenario: Dotfile resource name becomes module name

- GIVEN a selected plan step for `dotfile:bash`
- WHEN the dotfiles installer handles the step
- THEN it requests module `bash` from the provider

#### Scenario: Catalog metadata is ignored for command input

- GIVEN a selected dotfile step has catalog metadata or descriptions
- WHEN the installer builds provider input
- THEN only the resource name is used as the module name

#### Scenario: Non-dotfile step is not accepted

- GIVEN a tool, package, runtime, or other non-dotfile step
- WHEN the dotfiles installer is invoked directly
- THEN it returns an unsupported or failed result
- AND no dotlink command is requested

#### Scenario: Unsafe module names fail safely

- GIVEN a selected dotfile resource name starts with `-`, is `.`, is `..`, contains a path separator, is absolute, contains traversal, or contains characters outside `[A-Za-z0-9._-]`
- WHEN local execution validates modules
- THEN execution fails safely before invoking dotlink

#### Scenario: Dotlink args are explicit

- GIVEN selected modules `bash` and `nvim`
- WHEN local execution builds the dotlink command
- THEN command args are exactly `link bash nvim` in that order

#### Scenario: Empty module list fails safely

- GIVEN no selected dotfile modules are provided
- WHEN local execution is requested
- THEN execution fails before invoking dotlink
