# Delta for bootstrap-orchestration

## ADDED Requirements

### Requirement: Domain-first orchestration core

The system MUST define bootstrap as a domain-first orchestration core that plans installs from profiles and point targets.

#### Scenario: Profile install is planned through the core

- GIVEN a requested profile install
- WHEN the system builds a plan
- THEN the plan is derived by the orchestration core
- AND the interface does not contain duplicate planning logic

#### Scenario: Point install is planned through the same core

- GIVEN a requested point install
- WHEN the system builds a plan
- THEN the same core produces the plan
- AND the plan remains specific to the requested point scope

### Requirement: Interface parity

CLI and future TUI interfaces MUST be thin views over the same planning and execution core.

#### Scenario: CLI and TUI share behavior

- GIVEN the same install request
- WHEN it is submitted through CLI or a future TUI
- THEN the resulting plan semantics are equivalent
- AND the interfaces do not diverge in business rules

#### Scenario: No duplicated workflow rules

- GIVEN a workflow rule such as install ordering
- WHEN the system evolves
- THEN the rule is defined once in the core
- AND interfaces only invoke it

### Requirement: Operational result reporting

The system MUST report plan and execution outcomes using structured statuses: installed, already-installed, skipped, failed, and attention-required.

#### Scenario: Results are structured for interfaces

- GIVEN a plan step completes
- WHEN the result is recorded
- THEN it has one structured status
- AND CLI and future TUI can render it without parsing logs

#### Scenario: Logs preserve troubleshooting detail

- GIVEN a step fails or requires attention
- WHEN reporting is produced
- THEN the status is visible in the result
- AND logs retain enough detail to diagnose the cause
