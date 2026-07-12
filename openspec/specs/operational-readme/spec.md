# Operational README Specification

## Purpose

The operational README MUST describe the supported planning and execution commands accurately, including their safety modes, idempotency boundary, bootstrap behavior, and recovery from partial failure.

## Requirements

### Requirement: README documents the command workflow

The README MUST document `plan`, `apply`, and `bootstrap` as available command surfaces and MUST distinguish planning, execution reporting, and bootstrap/advisory behavior.

#### Scenario: A new operator can identify the commands

- GIVEN an operator reads the operational README
- WHEN they look for the primary workflow
- THEN the README describes `plan` for inspecting selected work
- AND `apply` for reporting or performing the supported execution modes
- AND `bootstrap` with its actual execution semantics and safety boundary

### Requirement: README documents target and safety flags

The README MUST document the applicable `--profile`, repeatable `--resource`, and `--catalog` target flags, and MUST explain `--yes`, `--sudo`, and `--dry-run` according to the command's actual validation and mutation behavior. It MUST state that `--dry-run` and `--yes` are incompatible, and that `--sudo` is meaningful only with confirmed `--yes` where supported.

#### Scenario: Flag guidance matches command behavior

- GIVEN an operator follows the README flag guidance
- WHEN they compare it with the command surface
- THEN target selection and safety-mode descriptions match actual behavior
- AND the README does not imply that default or dry-run apply mutates the host
- AND the README does not imply that `--sudo` independently enables mutation

### Requirement: README states the narrow idempotency promise

The README MUST state that confirmed apply and bootstrap avoid installer mutation only for eligible tool/runtime resources whose configured command presence was reliably detected and whose plan status is `already_installed`. It MUST state that the result is reported as unchanged and that no mutation was attempted.

#### Scenario: README explains a confirmed rerun

- GIVEN a resource is command-present and planned as `already_installed`
- WHEN an operator reads the README before a confirmed rerun
- THEN the README explains that the step is reported unchanged
- AND it explicitly says no mutation is attempted
- AND it identifies that `bootstrap` shares this apply behavior

### Requirement: README states idempotency limits and exclusions

The README MUST explicitly state that detected command presence is not proof of package installation details, package version, configuration correctness, or dotfile-link convergence. It MUST state that this slice does not perform retries, rollback, or bootstrap acquisition, and does not promise general idempotency for dotfiles.

#### Scenario: README prevents overclaiming

- GIVEN an operator reads the idempotency and limitations guidance
- WHEN they assess whether a rerun reconciles the machine
- THEN the README says package/version/configuration state is not verified by command presence
- AND the README says dotfile module presence does not prove links are current
- AND the README does not promise retry, rollback, or acquisition

### Requirement: README documents reporting and partial-failure recovery

The README MUST describe ordered result reporting with `changed`, `unchanged`, `not supported yet`, and `failed` categories. It MUST explain that mixed plans retain their original order, execution continues according to existing behavior after a step failure, and confirmed eligible failures produce a non-zero outcome. Recovery guidance MUST instruct the operator to fix the reported cause and rerun deliberately; it MUST NOT imply automatic retry or rollback.

#### Scenario: Operator can recover from a partial failure

- GIVEN a confirmed run reports changed, unchanged, unsupported, and failed results
- WHEN the operator consults the README
- THEN they can identify each result in plan order
- AND they are instructed to fix the failed cause before a deliberate rerun
- AND the documentation does not claim that failed work was automatically retried or rolled back

### Requirement: README documents bootstrap acquisition boundaries

The README MUST describe bootstrap or provider-need output as advisory guidance when the required bootstrap dependency is absent. It MUST state that this workflow does not acquire, clone, fetch, install, retry, or otherwise bootstrap that dependency automatically.

#### Scenario: Missing provider guidance is understood

- GIVEN a run reports a missing bootstrap dependency
- WHEN an operator reads the README
- THEN the output is understood as manual/advisory guidance
- AND the operator is not led to believe that the command acquired or installed the dependency
