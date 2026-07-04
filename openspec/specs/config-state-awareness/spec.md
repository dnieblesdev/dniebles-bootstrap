# Delta for config-state-awareness

## ADDED Requirements

### Requirement: Read-only config-state detection

The system MUST detect required config key presence from local filesystem conventions only.
It MUST NOT mutate dotfiles, invoke runtime ownership logic, or perform config installation.

#### Scenario: Present config is detected

- GIVEN a required config key maps to an existing path
- WHEN the detector runs
- THEN the key is reported present
- AND no files are changed

#### Scenario: Missing config is detected without side effects

- GIVEN a required config key maps to a missing path
- WHEN the detector runs
- THEN the key is reported absent
- AND dotfiles ownership is not claimed

### Requirement: Deterministic injected seams

The detector MUST use injected filesystem and path-convention seams.
Tests MUST be host-independent and deterministic.

#### Scenario: Fixture-driven detection is stable

- GIVEN injected seams that return fixed paths and existence checks
- WHEN detection runs twice
- THEN both results are identical

#### Scenario: Host filesystem is not required

- GIVEN a test fixture with no real host dependency
- WHEN the detector runs
- THEN the result depends only on the fixture

### Requirement: CLI wiring supplies detected config state

The `plan` command MUST detect config state after catalog load and before `BuildPlan`.
It MUST replace the empty `planning.ConfigState{}` call-site with detected state.

#### Scenario: Detected state is forwarded to planning

- GIVEN catalog load succeeds
- WHEN `dbootstrap plan` runs
- THEN config state detection occurs before planning
- AND `BuildPlan` receives the detected state

#### Scenario: Catalog load failure skips detection

- GIVEN the catalog cannot be loaded
- WHEN `dbootstrap plan` runs
- THEN config state detection is not attempted

### Requirement: Planner remains pure and caller-driven

`internal/planning` MUST remain free of filesystem probing and detector ownership.
Planning MUST consume caller-supplied config state only.

#### Scenario: Planning uses supplied state only

- GIVEN planning receives config state from the caller
- WHEN the plan is built
- THEN no filesystem access occurs inside planning

#### Scenario: Empty config state preserves planner behavior

- GIVEN an empty config state
- WHEN the plan is built
- THEN missing config still yields `attention_required` where required

### Requirement: Status behavior depends on config presence

Resources requiring config MUST be `attention_required` when required keys are absent.
Resources with required keys present MUST not be marked missing solely for config.

#### Scenario: Missing config yields attention required

- GIVEN a matching resource requires a config key that is absent
- WHEN the plan is built
- THEN the step status is `attention_required`
- AND the missing key is listed as a reason

#### Scenario: Present config avoids missing-config attention

- GIVEN a matching resource requires a config key that is present
- WHEN the plan is built
- THEN the step is not marked `attention_required` for that key

### Requirement: No dotfiles mutation or runtime ownership

The config-state feature MUST remain read-only and MUST NOT claim ownership of dotfiles runtime behavior.
Docs MAY be cleaned up, but docs wording MUST NOT expand runtime scope.

#### Scenario: Detector does not own dotfiles runtime

- GIVEN config-state detection runs
- WHEN it completes
- THEN it has not installed, applied, or mutated dotfiles

#### Scenario: Documentation cleanup stays non-normative

- GIVEN README wording is stale
- WHEN docs are updated
- THEN the domain requirements remain read-only and unchanged
