# Delta for dotfiles-integration

## ADDED Requirements

### Requirement: External dotfiles provider boundary

The system MUST treat dotfiles as an external provider and MUST NOT own dotfiles internals.

#### Scenario: Provider operations stay external

- GIVEN a plan requires dotfiles work
- WHEN the system resolves the provider boundary
- THEN dotfiles internals remain outside bootstrap ownership
- AND provider actions are expressed as external requests

#### Scenario: Dotlink is not modeled as bootstrap internals

- GIVEN a requested dotfiles action
- WHEN bootstrap delegates the action
- THEN `dotlink` is invoked only as an external provider operation
- AND bootstrap does not define dotfiles asset semantics

### Requirement: Sparse checkout and partial clone planning

The system SHOULD support sparse checkout and partial clone as provider-level operations when dotfiles modules are needed.

#### Scenario: Requested modules are planned for provider fetch

- GIVEN a profile or point install needs dotfiles modules
- WHEN the plan is built
- THEN the provider fetch scope includes only the requested modules
- AND fetch strategy is recorded as provider responsibility

#### Scenario: Missing module scope does not leak internals

- GIVEN a module is absent from the requested scope
- WHEN planning runs
- THEN bootstrap reports the missing provider scope
- AND it does not invent dotfiles internals
