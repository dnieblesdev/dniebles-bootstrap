# Delta for apply-command-dry-run

## MODIFIED Requirements

### Requirement: Apply command exists with plan-style target flags

The `apply` command MUST exist and MUST accept `--profile`, repeatable `--resource`, and `--catalog` using the same validation rules as `plan`.
The command MUST surface Homebrew bootstrap reporting when Homebrew-backed resources are missing `brew`, while keeping every accepted mode non-mutating.
(Previously: Apply accepted the same target surface as plan and rendered a noop execution report.)

#### Scenario: Apply accepts the same targets as plan

- GIVEN the user provides valid `--profile`, `--resource`, or `--catalog` values
- WHEN `dbootstrap apply` runs
- THEN the command accepts the same target surface as `plan`

#### Scenario: Invalid target input is rejected

- GIVEN the user provides malformed or unsupported target input
- WHEN `dbootstrap apply` runs
- THEN the command fails with a clear validation error

#### Scenario: Missing brew is reported without mutation

- GIVEN a Homebrew-backed resource is selected
- WHEN `dbootstrap apply` runs on a host without `brew`
- THEN the report includes a bootstrap action
- AND no host mutation occurs

### Requirement: Apply renders a noop execution report

The `apply` command MUST render an execution report separate from plan rendering.
Successful dry-run execution MUST report `not_implemented` results and MUST NOT imply real work completed.
Homebrew bootstrap reporting MUST remain advisory and non-mutating in every accepted mode.
(Previously: Dry-run execution reported not_implemented results and did not imply real work completed.)

#### Scenario: Dry-run execution reports not_implemented

- GIVEN a valid plan is produced
- WHEN `dbootstrap apply` runs the execution phase
- THEN each step is reported as `not_implemented`

#### Scenario: Execution rendering is distinct from plan rendering

- GIVEN both plan and apply commands are available
- WHEN their output is rendered
- THEN apply output is clearly labeled as execution reporting
- AND plan rendering remains separate

#### Scenario: Bootstrap reporting does not become execution

- GIVEN Homebrew bootstrap guidance is shown
- WHEN the apply report is rendered
- THEN the guidance is advisory only
- AND no execution step is treated as completed

### Requirement: Apply mode is explicit and safe by default

The `apply` command MUST treat the default mode as non-mutating, MUST treat `--dry-run` as explicit non-mutating, and MUST treat `--yes` as a reserved future confirmed-mode opt-in.
The command MUST keep Homebrew bootstrap reporting non-mutating in all modes.
(Previously: Default, dry-run, and `--yes` mode selection only described non-mutating safety.)

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

#### Scenario: Bootstrap guidance remains non-mutating under yes

- GIVEN the user runs `dbootstrap apply --yes` on a host without `brew`
- WHEN the command executes
- THEN bootstrap guidance may be reported
- AND no host mutation is performed
