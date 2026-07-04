# Delta for dotfiles-provider

## ADDED Requirements

### Requirement: TOML dotfiles catalog support

The catalog MUST support `[[dotfiles]]` entries and map them into dotfile resources.
It MUST validate dotfile entries and their dependencies using existing catalog rules.

#### Scenario: Dotfiles entries load into resources

- GIVEN a catalog contains a valid `[[dotfiles]]` entry
- WHEN the catalog is loaded
- THEN the entry is available as a dotfile resource

#### Scenario: Invalid dotfiles entries fail validation

- GIVEN a catalog contains an invalid `[[dotfiles]]` entry
- WHEN the catalog is loaded
- THEN validation fails

### Requirement: Read-only dotfiles repo and module detection

The system MUST detect dotfiles repository and module directory presence through injected filesystem seams.
Detection MUST be read-only and MUST NOT clone, apply, install, or mutate dotfiles.

#### Scenario: Present module is detected

- GIVEN injected seams report the dotfiles repo and a module directory exist
- WHEN detection runs
- THEN the module is reported present

#### Scenario: Missing module is absent without side effects

- GIVEN injected seams report the module directory is missing
- WHEN detection runs
- THEN the module is reported absent
- AND no dotfiles mutation occurs

### Requirement: CLI wiring merges present dotfiles into installation state

The `plan` command MUST merge detected present dotfile modules into existing `InstallationState.PresentResources` before planning.
The CLI MUST NOT duplicate planner semantics or own dotfiles runtime behavior.

#### Scenario: Present dotfile module reaches planning

- GIVEN catalog loading succeeds and dotfiles detection reports a present module
- WHEN `dbootstrap plan` runs
- THEN the module is added to `InstallationState.PresentResources`
- AND planning sees the module as already installed

#### Scenario: Detection is skipped on catalog failure

- GIVEN the catalog cannot be loaded
- WHEN `dbootstrap plan` runs
- THEN dotfiles detection is not attempted

### Requirement: Planner remains pure and caller-driven for dotfiles

`internal/planning` MUST remain free of dotfiles filesystem probing and ownership logic.
`BuildPlan` SHOULD keep its signature unchanged unless a test proves that dotfiles state cannot be supplied through existing caller inputs.

#### Scenario: Existing inputs carry dotfiles presence

- GIVEN the caller supplies installation state with dotfile resources present
- WHEN the plan is built
- THEN planning uses the supplied state only

#### Scenario: Signature expansion is avoided

- GIVEN dotfiles presence can be merged into existing installation state
- WHEN the slice is implemented
- THEN `BuildPlan` is not expanded solely for dotfiles

### Requirement: Dotfile module availability semantics

The system MUST treat a dotfile module as available when its local module directory exists under the configured dotfiles base path.
Availability MUST remain a presence signal only and MUST NOT imply applied, cloned, or symlinked state.

#### Scenario: Existing directory means available

- GIVEN a module directory exists under the dotfiles base path
- WHEN availability is evaluated
- THEN the module is available

#### Scenario: Presence does not imply mutation

- GIVEN a module is available
- WHEN planning completes
- THEN no apply, install, clone, or symlink action is performed

## MODIFIED Requirements

### Requirement: Planned resources reflect installation state

Resources that match environment facts and are marked present in installation state MUST remain in plan steps and MUST be reported with `already_installed` status.
Dotfile resources supplied through installation state MUST use the same presence semantics.
Resources that are not present MUST keep existing planning semantics.
(Previously: matching resources were always marked planned or attention_required.)

#### Scenario: Present resource is already installed

- GIVEN a tool, runtime, or dotfile resource matches the environment and is present in installation state
- WHEN the plan is built
- THEN the step is included
- AND the step status is `already_installed`

#### Scenario: Absent resource keeps existing semantics

- GIVEN a matching resource is not present in installation state
- WHEN the plan is built
- THEN the step status remains planned or attention_required as before

## REMOVED Requirements

### Requirement: None

(Reason: No requirement is removed in this change.)
(Migration: None)
