# Delta for bootstrap-entrypoint

## ADDED Requirements

### Requirement: First-run entrypoint acquires dbootstrap

The system MAY provide a tiny shell wrapper only to make `dbootstrap` available and hand control to it.

#### Scenario: Released binary path

- GIVEN a compatible released `dbootstrap` binary is available
- WHEN first startup runs
- THEN the wrapper may download/install that binary
- AND execution continues in `dbootstrap`

#### Scenario: Go compile or run path

- GIVEN no binary is available or source execution is desired
- WHEN first startup runs
- THEN the wrapper may install or use Go to compile/run from the repo
- AND execution continues in `dbootstrap`

### Requirement: Wrapper does not orchestrate environments

The shell wrapper MUST NOT own catalog resolution, dotfiles integration, installer selection, dependency ordering, plan execution, or operational reporting.

#### Scenario: Wrapper boundary is enforced

- GIVEN the wrapper has started `dbootstrap`
- WHEN profile or point installation begins
- THEN the Go application/core owns orchestration
- AND the wrapper performs no development environment planning
