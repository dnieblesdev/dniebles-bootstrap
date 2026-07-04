# Delta for installation-state

## ADDED Requirements

### Requirement: Planning accepts explicit installation state

The system MUST treat installation state as caller-supplied input to planning.
Planning MUST remain pure and MUST NOT probe the host for installation state.

#### Scenario: Empty state preserves current behavior

- GIVEN planning is called with an empty installation state
- WHEN the plan is built
- THEN existing planned, attention_required, and skipped behavior is unchanged

#### Scenario: State is provided by the caller

- GIVEN the caller supplies installation state data
- WHEN the plan is built
- THEN planning uses that data without host access

### Requirement: Host-independent state detection seams

The system MUST detect tool and runtime presence through PATH lookup behind injectable seams.
Tests MUST be deterministic and MUST NOT depend on the real host environment.

#### Scenario: Tool presence is detected through injected lookup

- GIVEN a fake PATH lookup reports a tool present
- WHEN state detection runs
- THEN the tool is reported as present

#### Scenario: Tests avoid host dependence

- GIVEN a test fixture with injected lookup behavior
- WHEN state detection runs
- THEN the result depends only on the fixture

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

### Requirement: Status precedence is deterministic

`already_installed` MUST take precedence after environment matching succeeds, including when the resource would otherwise require attention for missing config.
Resources excluded by environment mismatch MUST remain skipped and MUST NOT become `already_installed`.
(Previously: no installation-state precedence existed.)

#### Scenario: Present resource beats missing config

- GIVEN a matching resource is present in installation state and also lacks required config
- WHEN the plan is built
- THEN the step status is `already_installed`

#### Scenario: Environment mismatch still skips

- GIVEN a resource does not match the environment facts
- WHEN the plan is built
- THEN the resource remains skipped
- AND installation state does not change that outcome

## REMOVED Requirements

### Requirement: None

(Reason: No requirement is removed in this change.)
(Migration: None)
