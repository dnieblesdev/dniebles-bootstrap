# Delta for catalog-installer-metadata

## ADDED Requirements

### Requirement: Structured install metadata

Catalog resources MUST allow inert install metadata to be stored as structured data for downstream provider selection.
The system SHALL preserve provider/package information without inferring execution behavior.
Only the supported providers `brew` and `apt` are accepted; other providers are rejected at decode time.

#### Scenario: Provider and package metadata are accepted

- GIVEN a catalog resource with install metadata containing a supported provider and package name
- WHEN the resource is decoded into the catalog model
- THEN the install metadata is preserved as structured data
- AND no shell command is required

#### Scenario: Unsupported install provider is rejected

- GIVEN a catalog resource with install metadata using provider `asdf`
- WHEN the resource is decoded
- THEN decoding fails with a clear unsupported-provider error
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

### Requirement: Catalog contracts remain inventory-independent

Active canonical and development contracts going forward MUST derive default-catalog expectations from declared data without enumerating current resource names. Exact CLI behavior contracts MUST use minimal custom catalogs, while retaining a derived smoke check for the default catalog. Archived historical artifacts are immutable and MAY truthfully retain prior resource enumerations.

#### Scenario: CLI behavior is isolated from default inventory growth

- GIVEN a CLI test asserts formatting, mode, safety, provider, ordering, or reporting behavior
- WHEN the test executes
- THEN it uses a minimal custom catalog with only the resources needed by that assertion
- AND adding a default catalog resource does not require exact-output edits

#### Scenario: Existing runtime contracts remain covered

- GIVEN provider selection, safe execution modes, manual-action handling, or report rendering is under test
- WHEN the contract is exercised with a minimal fixture
- THEN the expected provider, safety, and reporting behavior remains asserted
- AND no default inventory names are required

#### Scenario: Historical artifacts remain immutable

- GIVEN the change updates active canonical specifications
- WHEN the change is archived
- THEN archived specifications and prior change artifacts remain unchanged

### Requirement: Default catalog declares generic reachable Brew-backed metadata

The default catalog MUST remain the runtime source of truth for its declared resources, bundles, profiles, dependencies, and structured metadata. Every declared Brew-backed tool or package intended for default user workflows MUST have non-empty Brew package metadata and a `command_exists` presence check, and MUST be reachable from a declared profile root through profile resources, profile bundles, bundle resources, and transitive resource dependencies. This workflow-membership closure MUST be derived from raw declarations independently of decoded planning maps. Point-resource selection MUST NOT satisfy this membership invariant, although separate contracts MAY test explicit selection behavior. Generic active canonical and development contracts going forward MUST validate section identity, reference resolution, profile-plan closure, deterministic planning, and dependency-before-dependent ordering without naming individual resources. No runtime behavior, schema, default catalog declaration, provider capability, or archive history is changed by this refactor.

#### Scenario: Generic Brew metadata is present

- GIVEN the default catalog is loaded
- WHEN each declared Brew-backed tool or package is inspected
- THEN its provider and package metadata are non-empty
- AND its presence metadata is a `command_exists` check with a non-empty command

#### Scenario: Default Brew targets are reachable from declared profile roots

- GIVEN raw default resources, bundles, profiles, and dependencies are declared
- WHEN the contract traverses each declared profile root through its direct resources, referenced bundles, bundle resources, and transitive dependencies
- THEN every cross-section reference resolves
- AND every declared Brew-backed tool or package is a member of that raw workflow closure
- AND an orphaned Brew-backed resource fails the contract
- AND point-resource selection is not used to establish workflow membership

#### Scenario: Profile plans reflect the declared workflow closure

- GIVEN the raw profile-root closures are valid
- WHEN each declared profile is planned through the decoded catalog
- THEN the resulting plan contains its independently derived raw closure
- AND no point-resource request is required to prove default workflow reachability

#### Scenario: Planning is complete and deterministic

- GIVEN a valid declared profile selection
- WHEN its plan is built repeatedly
- THEN the plan contains the complete derived dependency closure
- AND its order is deterministic with every dependency before its dependent

#### Scenario: Invariants are independently derived

- GIVEN raw catalog sections and decoded typed maps are available
- WHEN catalog contracts validate them
- THEN section identities and counts are compared independently against raw declarations
- AND expected workflow reachability and profile-plan closure are derived from raw declarations or schema rules rather than decoded planning maps

#### Scenario: Default catalog and runtime behavior are unchanged

- GIVEN this contract refactor is applied
- WHEN the default catalog and runtime flows are inspected
- THEN no catalog declaration, schema, planning, provider, CLI, safety, execution, or reporting behavior changes
- AND active canonical and development specifications going forward do not enumerate current resource names
- AND archived historical artifacts remain immutable and MAY truthfully retain prior resource enumerations
