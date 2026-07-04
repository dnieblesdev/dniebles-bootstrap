# Delta for cli-installation-state

## ADDED Requirements

### Requirement: CLI plan detects installation state before planning

The plan command MUST detect installation state after catalog load and before BuildPlan.

#### Scenario: Detection runs before planning

- GIVEN the catalog is loaded successfully
- WHEN `dbootstrap plan` runs
- THEN installation state is detected before planning begins
- AND the detected state is available to the planner call

#### Scenario: Catalog load failure prevents detection

- GIVEN the catalog cannot be loaded
- WHEN `dbootstrap plan` runs
- THEN installation state detection is not attempted
- AND the command returns the existing catalog-load failure

### Requirement: CLI passes detected state to planning without duplicated logic

The plan command MUST pass the detected installation state to BuildPlan through the CLI composition root.
The command MUST NOT duplicate planner rules or detector rules in CLI code.

#### Scenario: Detected state is forwarded intact

- GIVEN a detector returns installation state with present resources
- WHEN `dbootstrap plan` builds the plan
- THEN BuildPlan receives that exact detected state

#### Scenario: CLI does not reimplement selection logic

- GIVEN a resource is present or absent in installation state
- WHEN `dbootstrap plan` runs
- THEN the CLI does not decide step status itself
- AND the planner remains the source of plan semantics

### Requirement: CLI tests use an injected detector seam

The plan command tests MUST be host-independent and MUST inject installation-state detection.

#### Scenario: Present-state test is deterministic

- GIVEN a test-supplied detector reports a present resource
- WHEN the plan command runs in tests
- THEN the output is deterministic
- AND the result does not depend on the host PATH

#### Scenario: Empty-state baseline is deterministic

- GIVEN a test-supplied detector returns empty installation state
- WHEN the plan command runs in tests
- THEN the baseline output remains stable

## MODIFIED Requirements

### Requirement: Planned resources reflect installation state

Resources that match environment facts and are marked present in installation state MUST remain in plan steps and MUST be reported with `already_installed` status.
Resources that are not present MUST keep existing planning semantics.
(Previously: matching resources were always marked planned or attention_required.)

#### Scenario: Present resource is already installed

- GIVEN a tool or runtime resource matches the environment and is present in installation state
- WHEN the plan is built
- THEN the step is included
- AND the step status is `already_installed`

#### Scenario: Absent resource keeps existing semantics

- GIVEN a matching resource is not present in installation state
- WHEN the plan is built
- THEN the step status remains planned or attention_required as before

### Requirement: Detector failures remain future scope

The plan command MUST use the current installation-state detector contract, which returns installation state without an error value.
Detector failure diagnostics are deferred until the detector contract can represent failures explicitly.

#### Scenario: Current detector contract is used unchanged

- GIVEN the installation-state detector returns installation state without an error
- WHEN `dbootstrap plan` runs
- THEN the command passes that state to planning
- AND it does not invent detector-failure diagnostics

#### Scenario: Future detector failures are not implemented in this slice

- GIVEN future detector contracts may report errors
- WHEN this slice is applied
- THEN no detector error branch is added
- AND no detector, planner, or renderer contract is expanded for failure diagnostics

## REMOVED Requirements

### Requirement: None

(Reason: No requirement is removed in this change.)
(Migration: None)
