# Delta for cli-plan

## ADDED Requirements

### Requirement: Plan command entrypoint

`dbootstrap plan --profile <name>` MUST load the repo-local catalog fixture from `catalog/bootstrap.toml` by default and produce a plan for the requested profile.

#### Scenario: Default repo-local catalog

- GIVEN the repo contains `catalog/bootstrap.toml`
- WHEN the user runs `dbootstrap plan --profile dev`
- THEN the command MUST load that catalog file
- AND the command MUST exit successfully when planning succeeds

#### Scenario: Missing profile

- GIVEN no profile name is provided
- WHEN the user runs `dbootstrap plan`
- THEN the command MUST exit non-zero
- AND stderr MUST explain that `--profile` is required

### Requirement: Thin command boundary

The command MUST use the catalog adapter and planning core without duplicating planning business logic in the CLI layer.

#### Scenario: Adapter-backed planning

- GIVEN a valid catalog file and profile
- WHEN the command builds a plan
- THEN it MUST call the catalog adapter and planning core
- AND it MUST not embed catalog interpretation rules in command code

### Requirement: Deterministic human output

The command MUST render deterministic human-readable output that includes planned, attention, and diagnostic information as applicable.

#### Scenario: Stable success output

- GIVEN a valid plan result
- WHEN the command prints output twice for the same inputs
- THEN the text MUST be identical
- AND it MUST include the planned items and any attention notes

#### Scenario: Diagnostics are visible

- GIVEN planning returns diagnostics
- WHEN the command prints the plan
- THEN stderr or stdout MUST include diagnostic information in a human-readable form

### Requirement: Error handling

The command MUST return a non-zero exit code and useful stderr for unknown profiles and catalog load failures.

#### Scenario: Unknown profile

- GIVEN the requested profile does not exist in the catalog
- WHEN the user runs the command
- THEN it MUST exit non-zero
- AND stderr MUST name the missing profile

#### Scenario: Invalid catalog input

- GIVEN the catalog path or contents cannot be loaded or decoded
- WHEN the user runs the command
- THEN it MUST exit non-zero
- AND stderr MUST explain the catalog load problem

### Requirement: Static environment facts only

The command MUST use caller-supplied static `EnvironmentFacts` and MUST NOT probe the host OS in this slice.

#### Scenario: No OS probing

- GIVEN the command is executed in any environment
- WHEN planning occurs
- THEN the command MUST rely only on provided environment facts
- AND it MUST not read live OS state for this slice

### Requirement: Command tests

The command MUST have tests covering the happy path, missing profile, invalid catalog path or input, and exact output or stderr where appropriate.

#### Scenario: Exact output assertions

- GIVEN a deterministic plan case
- WHEN the test runs the command
- THEN the test MUST assert the exact rendered output or stderr
- AND it MUST verify the exit status

## REMOVED Requirements

### Requirement: None

(Reason: No existing cli-plan behavior to remove.)
(Migration: None)
