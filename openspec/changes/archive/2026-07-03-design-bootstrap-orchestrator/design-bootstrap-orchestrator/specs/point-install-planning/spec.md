# Delta for point-install-planning

## ADDED Requirements

### Requirement: Point install planning

The system MUST produce a validated plan for point installs scoped to the requested resource or capability.

#### Scenario: Point scope stays narrow

- GIVEN a point install request
- WHEN the plan is built
- THEN only the requested scope is included
- AND unrelated catalog resources are excluded

#### Scenario: Point install resolves existing state

- GIVEN the requested point is already installed
- WHEN planning runs
- THEN the result identifies the existing state
- AND redundant installation is avoided

### Requirement: Point-scoped dotfiles handling

The system MUST request only the dotfiles modules needed for the point scope.

#### Scenario: Single point module is requested

- GIVEN one point resource depends on a dotfiles module
- WHEN the plan is built
- THEN only that module is requested from the provider
- AND dotlink is limited to the requested scope

#### Scenario: Point request with missing module config

- GIVEN the needed dotfiles config is missing
- WHEN the point install is planned or executed
- THEN the install can proceed
- AND the missing configuration is reported for attention
