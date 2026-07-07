# Delta for apply-command-dry-run

## ADDED Requirements

### Requirement: Apply mode is explicit and safe by default

The `apply` command MUST treat the default mode as non-mutating, MUST treat `--dry-run` as explicit non-mutating, and MUST treat `--yes` as a reserved future confirmed-mode opt-in.

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

### Requirement: Conflicting safety flags are rejected

The `apply` command MUST reject `--dry-run --yes` as invalid input and MUST return a clear usage error.

#### Scenario: Dry-run and yes cannot be combined

- GIVEN the user runs `dbootstrap apply --dry-run --yes`
- WHEN the command validates flags
- THEN the command fails with a usage error
- AND no execution result is produced

### Requirement: Confirmed mode is reserved but not wired

The `apply` command MUST accept `--yes` as a future mutation opt-in marker, but this slice MUST NOT wire real installer, CommandRunner, Homebrew bootstrap, or remote script mutation behind it.

#### Scenario: Yes is accepted without mutation wiring

- GIVEN the user runs `dbootstrap apply --yes`
- WHEN the command executes
- THEN the selected mode is reported as confirmed future mode
- AND execution remains non-mutating

#### Scenario: No real mutation surfaces are active

- GIVEN the apply command completes in any accepted mode
- WHEN the execution path is observed
- THEN no real installer, CommandRunner, Homebrew, or remote script mutation is invoked
- AND raw command metadata is not exposed
