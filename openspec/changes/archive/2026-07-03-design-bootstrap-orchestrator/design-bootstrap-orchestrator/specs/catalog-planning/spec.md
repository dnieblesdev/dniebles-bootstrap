# Delta for catalog-planning

## ADDED Requirements

### Requirement: Declarative in-repo catalog

The system MUST keep the install catalog in this repository and SHOULD use TOML as the initial authoring format.

#### Scenario: Catalog is repository-local

- GIVEN a contributor inspects the project
- WHEN they look for catalog definitions
- THEN the authoritative catalog is found in the repo
- AND it is not owned by dotfiles

#### Scenario: Format remains domain-agnostic

- GIVEN a future format migration
- WHEN catalog data is read by the core
- THEN the domain model remains independent of the file format
- AND TOML is only a recommended starting point

### Requirement: Planning from catalog entries

The system MUST plan profile and point installs from declarative catalog entries for tools, runtimes, packages, and bundles.

#### Scenario: Profile expands catalog resources

- GIVEN a profile referencing bundles and resources
- WHEN the plan is created
- THEN catalog entries are expanded into install actions
- AND the plan preserves declared scope

#### Scenario: Unknown catalog entry is rejected

- GIVEN a requested resource that is not declared
- WHEN planning runs
- THEN the system reports the entry as unavailable
- AND no install plan is produced for that entry

### Requirement: Dependency-aware plan ordering

The system MUST order tools, runtimes, packages, and bundles according to declared dependencies before execution.

#### Scenario: Runtime dependency precedes tool

- GIVEN a tool depends on a runtime or package
- WHEN the plan is built
- THEN the dependency appears before the dependent tool
- AND execution order is deterministic

#### Scenario: Dependency issue is reportable

- GIVEN catalog dependencies cannot be resolved
- WHEN planning runs
- THEN the plan result reports the dependency issue
- AND unsafe execution is skipped or blocked for affected steps
