# Delta for point-install-planning

## ADDED Requirements

### Requirement: CLI accepts repeatable resource targets

The `plan` command MUST accept repeatable `--resource kind:name` targets.
It MUST validate the `kind:name` shape and reject unsupported or malformed refs with clear errors.

#### Scenario: Resource-only planning input is accepted

- GIVEN the user passes one or more valid `--resource` values
- WHEN `dbootstrap plan` runs
- THEN the command accepts the targets for planning

#### Scenario: Malformed resource ref is rejected

- GIVEN the user passes `--resource git` or another invalid ref
- WHEN `dbootstrap plan` runs
- THEN the command fails with a clear validation error

### Requirement: Plan requires a target profile or resource

The `plan` command MUST require at least one of `--profile` or `--resource`.
It MUST reject invocations that provide neither target type.

#### Scenario: Missing target is rejected

- GIVEN the user provides no `--profile` and no `--resource`
- WHEN `dbootstrap plan` runs
- THEN the command fails with a clear target-required error

#### Scenario: Profile-only planning remains valid

- GIVEN the user provides `--profile dev`
- WHEN `dbootstrap plan` runs
- THEN the command proceeds with profile planning

### Requirement: Profile and resource targets MAY be unioned

The `plan` command MUST allow `--profile` and `--resource` together.
When both are present, the command MUST pass both targets into planning so the existing domain can union them.

#### Scenario: Profile plus resource is accepted

- GIVEN the user provides `--profile dev --resource tool:git`
- WHEN `dbootstrap plan` runs
- THEN the plan uses both targets
- AND the result reflects the combined scope

#### Scenario: Existing profile behavior is preserved

- GIVEN the user provides only `--profile dev`
- WHEN `dbootstrap plan` runs
- THEN the existing profile plan path remains unchanged

### Requirement: Resource-only plans render resource-oriented headers

When no profile is supplied, the rendered plan header MUST describe the selected resources instead of a profile.
This behavior MUST be read-only and MUST NOT imply apply, install, mutation, or runtime execution.

#### Scenario: Resource-only header is shown

- GIVEN the user provides only `--resource tool:git`
- WHEN the plan is rendered
- THEN the header identifies the resource-based plan

#### Scenario: No runtime side effects are introduced

- GIVEN a resource-only plan is requested
- WHEN planning completes
- THEN no installer, apply, mutation, or runtime execution occurs

### Requirement: Existing pure planning domain support is reused

The CLI MUST reuse existing planning domain support for explicit resources.
It MUST NOT require planner or domain model changes unless a test proves the current domain inputs cannot express the requested plan.

#### Scenario: CLI forwards explicit resources to planning

- GIVEN the caller supplies valid `--resource` values
- WHEN `dbootstrap plan` constructs the request
- THEN the explicit resources are forwarded into the existing planning request

#### Scenario: Domain changes remain unnecessary

- GIVEN the current planning request already supports profile and resources
- WHEN this change is implemented
- THEN no planner or domain model change is required

## REMOVED Requirements

### Requirement: None

(Reason: This change adds CLI behavior without removing existing plan behavior.)
(Migration: None)
