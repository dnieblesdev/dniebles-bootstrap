# Delta for catalog-installer-metadata

## ADDED Requirements

### Requirement: Structured install metadata

Catalog resources MUST allow inert install metadata to be stored as structured data for downstream provider selection.
The system SHALL preserve provider/package information without inferring execution behavior.

#### Scenario: Provider and package metadata are accepted

- GIVEN a catalog resource with install metadata containing a provider and package name
- WHEN the resource is decoded into the catalog model
- THEN the install metadata is preserved as structured data
- AND no shell command is required

#### Scenario: Missing install metadata remains valid

- GIVEN a catalog resource without install metadata
- WHEN the resource is decoded
- THEN the resource remains valid
- AND no default install command is synthesized

### Requirement: Structured presence metadata

Catalog resources MUST allow inert presence metadata to describe how existing tools are detected.
The system SHALL support presence checks as data using check kinds such as path or command_exists.

#### Scenario: Presence check metadata is preserved

- GIVEN a catalog resource with presence metadata using kind path or command_exists
- WHEN the resource is decoded and mapped into planning types
- THEN the presence metadata is preserved unchanged
- AND it is not executed during planning

#### Scenario: Presence metadata is absent

- GIVEN a catalog resource without presence metadata
- WHEN the resource is processed
- THEN planning continues normally
- AND no presence check is inferred

### Requirement: Inert metadata propagation

The planning model MUST carry catalog install and presence metadata forward as inert resource data.
The system SHALL NOT change planning decisions, ordering, or execution semantics because metadata is present.

#### Scenario: Metadata survives plan creation

- GIVEN a resource with structured install and presence metadata
- WHEN a plan is built
- THEN the resulting plan resource contains the same metadata
- AND existing plan behavior is unchanged

#### Scenario: Metadata does not alter planning outcome

- GIVEN the same catalog inputs with and without metadata
- WHEN plans are built for both inputs
- THEN the planner produces equivalent behavior for existing steps
- AND metadata only affects preserved resource data
