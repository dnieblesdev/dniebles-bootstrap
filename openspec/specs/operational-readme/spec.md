# Operational README Specification

## Purpose

The operational README MUST describe the supported planning and execution commands accurately, including their safety modes, idempotency boundary, bootstrap behavior, and recovery from partial failure.

## Requirements

### Requirement: README documents the command workflow

The README MUST document `plan`, `apply`, and `bootstrap`, distinguishing planning, execution, and advisory behavior. It MUST also document direct install/uninstall, supported architectures, paths, PATH, force, and catalog location.

#### Scenario: A new operator can identify commands and first install

- GIVEN an operator reads the operational README
- WHEN they look for the primary workflow
- THEN the README describes `plan` for inspecting selected work
- AND `apply` for reporting or performing the supported execution modes
- AND `bootstrap` with its actual execution semantics and safety boundary
- AND it provides install, PATH, catalog, force, and uninstall guidance

#### Scenario: Unsupported platforms are not promised

- GIVEN an operator reads the installation guidance
- WHEN their host is macOS, Windows, or an unsupported architecture
- THEN the README states that direct binary installation is unavailable

### Requirement: README documents target and safety flags

The README MUST document `--profile`, repeatable `--resource`, `--catalog`, `--yes`, `--sudo`, and `--dry-run` accurately. It MUST state that dry-run and yes conflict, sudo requires confirmed yes where supported, and direct install never falls back to sudo or package managers.

#### Scenario: Flag guidance matches command behavior

- GIVEN an operator follows the README flag guidance
- WHEN they compare it with the command surface
- THEN target selection and safety-mode descriptions match actual behavior
- AND the README does not imply that default or dry-run apply mutates the host
- AND the README does not imply that `--sudo` independently enables mutation

### Requirement: README states idempotency limits and exclusions

The README MUST retain idempotency limits and state that direct installation verifies checksums, protects unmanaged files, and removes only unmodified manifest-owned files. It MUST NOT promise signing, package managers, macOS, or automatic `dbootstrap install` acquisition.

#### Scenario: README prevents overclaiming

- GIVEN an operator assesses installation or rerun behavior
- WHEN they read the lifecycle guidance
- THEN checksum-before-mutation, force protection, managed uninstall, and unsupported scope are explicit

### Requirement: README states the narrow idempotency promise

The README MUST state that confirmed apply and bootstrap avoid installer mutation only for eligible tool/runtime resources whose configured command presence was reliably detected and whose plan status is `already_installed`. It MUST state that the result is reported as unchanged and that no mutation was attempted.

#### Scenario: README explains a confirmed rerun

- GIVEN a resource is command-present and planned as `already_installed`
- WHEN an operator reads the README before a confirmed rerun
- THEN the README explains that the step is reported unchanged
- AND it explicitly says no mutation is attempted
- AND it identifies that `bootstrap` shares this apply behavior

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
