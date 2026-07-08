# Delta for apply-command-dry-run

## MODIFIED Requirements

### Requirement: Apply mode is explicit and safe by default

The `apply` command MUST treat the default mode as non-mutating, MUST treat `--dry-run` as explicit non-mutating, and MUST treat `--yes` as the only confirmed mode that may mutate for brew-backed tool/package steps.
The command MUST keep Homebrew bootstrap reporting non-mutating in default and `--dry-run` modes.
(Previously: `--yes` was reserved future mode and no real mutation wiring was allowed.)

#### Scenario: Default apply is non-mutating

- GIVEN the user runs `dbootstrap apply` with no safety flags
- WHEN the command executes
- THEN the selected mode is reported as non-mutating default
- AND no host mutation is performed

#### Scenario: Dry-run is explicit non-mutating

- GIVEN the user runs `dbootstrap apply --dry-run`
- WHEN the command executes
- THEN the selected mode is reported as dry-run
- AND no host mutation is performed

#### Scenario: Yes is the only confirmed mutating mode

- GIVEN the user runs `dbootstrap apply --yes` with a brew-backed tool or package step
- WHEN the command executes
- THEN the selected mode is reported as confirmed mode
- AND real brew installation may be attempted only for that step

### Requirement: Apply renders execution mode-specific reporting

The `apply` command MUST render execution reporting separate from plan rendering.
Successful dry-run execution MUST report `not_implemented` results, while confirmed mode MAY report real brew execution for brew-backed tool/package steps only.
Homebrew bootstrap reporting MUST remain advisory and non-mutating in default and `--dry-run` modes.
(Previously: apply always rendered noop execution results and did not wire real execution.)

#### Scenario: Dry-run execution reports not_implemented

- GIVEN a valid plan is produced
- WHEN `dbootstrap apply` runs the execution phase in dry-run mode
- THEN each step is reported as `not_implemented`

#### Scenario: Confirmed brew steps can report real execution

- GIVEN a brew-backed tool or package step is present
- WHEN `dbootstrap apply --yes` runs the execution phase
- THEN the step may report real brew execution
- AND other step kinds remain non-mutating or unsupported

### Requirement: Conflicting safety flags are rejected

The `apply` command MUST reject `--dry-run --yes` as invalid input and MUST return a clear usage error.

#### Scenario: Dry-run and yes cannot be combined

- GIVEN the user runs `dbootstrap apply --dry-run --yes`
- WHEN the command validates flags
- THEN the command fails with a usage error
- AND no execution result is produced

### Requirement: Default apply remains non-mutating

The `apply` command MUST NOT perform real execution, host mutation, dotlink, clone, sparse checkout, retry, or concurrency behavior in default mode.
It MUST remain a safe noop bridge over the existing plan.

#### Scenario: No host mutation occurs

- GIVEN `dbootstrap apply` runs successfully
- WHEN the command completes
- THEN no filesystem or host state is mutated

#### Scenario: No orchestration features are introduced

- GIVEN `dbootstrap apply` runs
- WHEN the execution path is reviewed
- THEN no retry, concurrency, dotlink, clone, or sparse checkout behavior is present

## ADDED Requirements

### Requirement: Confirmed mode only wires brew-backed installs

The `apply` command MUST wire real execution only for brew-backed `tool` and `package` steps when `--yes` is set.
Runtime, dotfile, and non-brew steps MUST remain noop or unsupported.

#### Scenario: Brew-backed steps may execute under yes

- GIVEN a brew-backed `tool` step and `--yes`
- WHEN apply executes
- THEN the step is eligible for real brew installation

#### Scenario: Non-brew steps stay non-mutating

- GIVEN a runtime or dotfile step and `--yes`
- WHEN apply executes
- THEN the step remains noop or returns unsupported
