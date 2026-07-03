# Delta for catalog-adapter

## ADDED Requirements

### Requirement: TOML Catalog Decode
The adapter MUST decode repo-local TOML catalog files into `internal/planning` domain inputs without exposing TOML-specific types to planning.

#### Scenario: Decode fixture catalog
- GIVEN a valid `catalog/*.toml` fixture
- WHEN the adapter loads the file
- THEN it returns a `planning.Catalog`
- AND no TOML struct types escape into planning APIs.

#### Scenario: Reject malformed TOML
- GIVEN a catalog file with invalid TOML syntax
- WHEN the adapter loads the file
- THEN it returns a parse error
- AND no plan is built.

### Requirement: Initial Catalog Schema
The TOML schema MUST support profiles, bundles, resources, config policy, dependencies, and environment constraints for the initial fixture.

#### Scenario: Map supported sections
- GIVEN a TOML catalog containing all supported sections
- WHEN the adapter decodes it
- THEN the resulting planning input contains those sections
- AND the fixture proves the repo-local catalog direction.

#### Scenario: Missing required field
- GIVEN a TOML entry missing a required identifier or name
- WHEN the adapter decodes it
- THEN it returns a structural validation error.

### Requirement: Adapter Isolation
TOML-specific structs, helpers, and validation details MUST remain isolated outside `internal/planning`.

#### Scenario: Planning core stays format-agnostic
- GIVEN the planning package
- WHEN it is reviewed for the catalog slice
- THEN it contains no TOML schema types or parsing logic.

### Requirement: Structural Validation Only
The adapter MUST perform shallow structural validation for parse errors, required fields, duplicate IDs, and basic unknown references when they can be checked locally.

#### Scenario: Duplicate IDs are rejected
- GIVEN a catalog with duplicate profile or resource IDs
- WHEN the adapter decodes it
- THEN it returns a duplicate-ID error.

#### Scenario: Unknown local reference is rejected
- GIVEN a catalog with a reference to a missing local dependency or bundle member
- WHEN the adapter decodes it
- THEN it returns a basic unknown-reference error.

### Requirement: No Planner Semantics Duplication
The adapter MUST NOT duplicate planner validation, dependency resolution, ordering, environment filtering, or execution semantics.

#### Scenario: Delegate deeper validation to planning
- GIVEN a structurally valid catalog that still violates planner rules
- WHEN it is passed to `planning.BuildPlan`
- THEN the adapter succeeds
- AND planning reports the semantic issue.

### Requirement: File-to-Plan Integration Coverage
The repo MUST include a `catalog/` fixture and tests proving decode plus `planning.BuildPlan` integration.

#### Scenario: Build plan from decoded fixture
- GIVEN the fixture catalog and a plan request
- WHEN the adapter decodes the file and calls `planning.BuildPlan`
- THEN a plan is produced or a planning error is returned only from planner rules
- AND the adapter adds no runtime side effects.

### Requirement: No Runtime Side Effects
The adapter MUST NOT perform CLI, TUI, installer, dotfile, git, OS probing, environment detection, or other runtime side effects.

#### Scenario: Pure decode path
- GIVEN a catalog decode test
- WHEN it runs in a temporary test environment
- THEN it uses only file input and in-memory transformation
- AND it performs no external commands or host inspection.
