# Delta for planning-core

## ADDED Requirements

### Requirement: Domain-only planning inputs

The core MUST accept decoded domain inputs only, and MUST NOT depend on TOML, YAML, or JSON parsing details.

#### Scenario: Decoded inputs are accepted

- GIVEN in-memory catalog, profile, bundle, resource, and policy values
- WHEN planning runs
- THEN the core produces a plan from those domain objects

#### Scenario: File format concerns stay outside

- GIVEN a caller has file content or parser tokens
- WHEN they attempt to plan directly with them
- THEN the core rejects format-specific inputs
- AND planning remains format-agnostic

### Requirement: Deterministic dependency-aware expansion

The core MUST expand profiles, bundles, and resource refs into a deterministic plan ordered by declared dependencies across tools, runtimes, packages, bundles, and resources.

#### Scenario: Dependencies precede dependents

- GIVEN a profile that references a bundle and dependent resources
- WHEN the plan is built twice from the same inputs
- THEN both plans have the same step order
- AND dependencies appear before dependents

#### Scenario: Invalid references are reportable

- GIVEN a profile references an unknown bundle or resource
- WHEN planning runs
- THEN the result reports the invalid reference
- AND valid resources can still be planned

### Requirement: Pure state and structured results

The core MUST represent desired and existing state as pure data and MUST NOT perform command execution or other runtime side effects.

#### Scenario: State is data only

- GIVEN existing and desired state inputs
- WHEN the plan is produced
- THEN the result contains data structures only
- AND no command is run

#### Scenario: Step results remain structured

- GIVEN planned steps with success, skip, attention, or error outcomes
- WHEN the result is inspected
- THEN each PlanStepResult has a structured status
- AND callers can distinguish outcomes without parsing text

### Requirement: Attention-required config handling

The core MUST mark missing configuration as attention-required in plan/result semantics and MUST NOT block unrelated resource planning.

#### Scenario: Missing config does not halt planning

- GIVEN a profile with missing config and valid resources
- WHEN planning runs
- THEN the result marks attention required
- AND valid resource planning still completes

#### Scenario: Missing config remains visible

- GIVEN a completed plan with missing config
- WHEN the result is reviewed
- THEN the missing config is still reported
- AND it is distinguishable from satisfied config

### Requirement: EnvironmentFacts influences planning

The core MUST use EnvironmentFacts as domain input for planning decisions without requiring OS probes in this slice.

#### Scenario: Facts shape plan decisions

- GIVEN environment facts for OS or architecture
- WHEN planning runs
- THEN the resulting plan reflects those facts

#### Scenario: No probe dependency exists

- GIVEN a caller provides EnvironmentFacts directly
- WHEN planning runs
- THEN no real environment detection is required
- AND the core remains testable with synthetic facts

### Requirement: Table-driven planning tests

The Go tests MUST verify planning behavior with table-driven cases covering ordering, expansion, missing config attention, invalid references, and no side effects.

#### Scenario: Multiple cases are covered

- GIVEN a test table with several planning scenarios
- WHEN tests run
- THEN each case executes independently
- AND assertions cover plan shape and status results

#### Scenario: Side effects are absent

- GIVEN the pure planning core test suite
- WHEN it runs
- THEN no commands or external adapters are invoked
- AND tests remain deterministic
