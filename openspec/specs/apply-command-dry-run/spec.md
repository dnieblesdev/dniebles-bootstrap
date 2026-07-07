# Delta for apply-command-dry-run

## ADDED Requirements

### Requirement: Apply command exists with plan-style target flags

The `apply` command MUST exist and MUST accept `--profile`, repeatable `--resource`, and `--catalog` using the same validation rules as `plan`.

#### Scenario: Apply accepts the same targets as plan

- GIVEN the user provides valid `--profile`, `--resource`, or `--catalog` values
- WHEN `dbootstrap apply` runs
- THEN the command accepts the same target surface as `plan`

#### Scenario: Invalid target input is rejected

- GIVEN the user provides malformed or unsupported target input
- WHEN `dbootstrap apply` runs
- THEN the command fails with a clear validation error

### Requirement: Apply reuses the planning pipeline

The `apply` command MUST build its request through the existing planning pipeline before any execution report is produced.
Planning failures MUST stop the command before execution begins.

#### Scenario: Planning failure stops apply early

- GIVEN planning cannot build a valid plan
- WHEN `dbootstrap apply` runs
- THEN the command exits with the planning error
- AND no execution report is rendered

#### Scenario: Successful planning continues to execution

- GIVEN planning succeeds
- WHEN `dbootstrap apply` runs
- THEN the resulting plan is passed into execution

### Requirement: Apply renders a noop execution report

The `apply` command MUST render an execution report separate from plan rendering.
Successful dry-run execution MUST report `not_implemented` results and MUST NOT imply real work completed.

#### Scenario: Dry-run execution reports not_implemented

- GIVEN a valid plan is produced
- WHEN `dbootstrap apply` runs the execution phase
- THEN each step is reported as `not_implemented`

#### Scenario: Execution rendering is distinct from plan rendering

- GIVEN both plan and apply commands are available
- WHEN their output is rendered
- THEN apply output is clearly labeled as execution reporting
- AND plan rendering remains separate

### Requirement: Apply remains strictly non-mutating

The `apply` command MUST NOT perform real execution, host mutation, dotlink, clone, sparse checkout, retry, or concurrency behavior.
It MUST remain a safe noop bridge over the existing plan.

#### Scenario: No host mutation occurs

- GIVEN `dbootstrap apply` runs successfully
- WHEN the command completes
- THEN no filesystem or host state is mutated

#### Scenario: No orchestration features are introduced

- GIVEN `dbootstrap apply` runs
- WHEN the execution path is reviewed
- THEN no retry, concurrency, dotlink, clone, or sparse checkout behavior is present

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

## MODIFIED Requirements

### Requirement: Explicit no-apply, no-real-execution, no-mutation boundary → Execution contracts remain non-mutating for apply

`internal/execution` MUST remain a safe, non-mutating boundary used by `apply`.
The command MUST use noop execution contracts only, and MUST NOT introduce real execution, host mutation, installers with side effects, or planning production changes.
(Previously: The execution slice prohibited any apply command or CLI wiring.)

#### Scenario: Apply uses noop execution contracts only

- GIVEN the `apply` command runs
- WHEN execution is dispatched
- THEN only noop results are produced

#### Scenario: Side effects remain absent

- GIVEN execution contracts are present
- WHEN `apply` is reviewed end-to-end
- THEN no real execution or production mutation occurs

## REMOVED Requirements

### Requirement: No apply command is introduced

(Reason: `apply` is now intentionally added as a dry-run-only CLI bridge.)
(Migration: Replace this gate with functional `apply` coverage.)
