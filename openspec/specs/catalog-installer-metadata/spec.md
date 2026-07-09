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

### Requirement: Default catalog includes a brew-backed tool target

The default catalog MUST declare `tool:git` as a brew-backed tool target using structured install metadata.
The system MUST preserve `install.package = "git"` and `presence.kind = "command_exists"` with `presence.name = "git"`.
The default catalog MUST preserve existing resources, bundles, and profiles, and MUST NOT introduce new resources.
The default catalog MUST preserve existing `package:ripgrep` brew-backed install metadata.
The default catalog MUST NOT require multi-provider or fallback install metadata for this resource set.

#### Scenario: tool:git is brew-backed

- GIVEN the default catalog is loaded
- WHEN `tool:git` is decoded
- THEN its install provider is `brew`
- AND its install package is `git`
- AND its presence kind is `command_exists`
- AND its presence name is `git`

#### Scenario: existing default catalog shape remains unchanged

- GIVEN the default catalog is loaded
- WHEN the catalog resources, bundles, and profiles are inspected
- THEN the existing resource set is preserved
- AND `bundle:cli` still includes `tool:git` and `package:ripgrep`
- AND `profile:dev` remains unchanged

#### Scenario: package:ripgrep remains brew-backed

- GIVEN the default catalog is loaded
- WHEN `package:ripgrep` is decoded
- THEN its install provider is `brew`
- AND its install package is `ripgrep`

#### Scenario: no multi-provider metadata is introduced

- GIVEN the default catalog is loaded
- WHEN install metadata is inspected
- THEN no fallback provider list is required
- AND no multi-provider selection metadata is introduced
