# Bootstrap Entrypoint Specification

## Purpose

`dbootstrap bootstrap` is the current explicit-target entrypoint for the existing safe apply workflow. This specification records delivered behavior only; it introduces no implementation changes.

## Requirements

### Requirement: Bootstrap requires an explicit target

`bootstrap` SHALL require at least one `--profile` or repeatable `--resource <kind:name>` target. It SHALL accept both target forms together and SHALL pass the resulting selection to the shared planner.

#### Scenario: Profile and resources select scope

- **GIVEN** a valid profile, one or more valid resources, or both
- **WHEN** `dbootstrap bootstrap` runs
- **THEN** the selected scope is planned through the shared apply workflow

#### Scenario: Missing target is rejected

- **GIVEN** `dbootstrap bootstrap` is invoked without `--profile` or `--resource`
- **WHEN** arguments are validated
- **THEN** it SHALL render bootstrap usage guidance and exit with a usage failure
- **AND** it SHALL not load the catalog, detect the environment, or start execution

### Requirement: Bootstrap is apply-equivalent

For the same explicit target, catalog, and safety flags, `bootstrap` SHALL use the same flag parsing, catalog loading, detection, planning, execution, reporting, and exit classification as `apply`. It SHALL not create a separate execution pipeline.

#### Scenario: Valid invocations have parity

- **GIVEN** identical valid explicit targets, catalog paths, and safety flags
- **WHEN** `apply` and `bootstrap` run
- **THEN** they SHALL produce equivalent reports and exit statuses

#### Scenario: Semantic target failure has parity

- **GIVEN** a syntactically valid profile or resource that is absent from the catalog
- **WHEN** `bootstrap` runs
- **THEN** it SHALL use the shared planning path
- **AND** it SHALL report the same diagnostic and failure outcome as `apply` without execution

### Requirement: Bootstrap preserves the shared safety contract

`bootstrap` SHALL support `--catalog`, `--dry-run`, `--yes`, and `--sudo` under the existing apply safety contract. Default mode and `--dry-run` SHALL be non-mutating. Only `--yes` SHALL permit eligible work to execute, and `--sudo` SHALL require `--yes` and apply only to eligible APT work.

Confirmed execution SHALL retain the existing provider eligibility and failure behavior: eligible Brew-backed tool and package work, eligible Linux APT-backed tool and package work, and selected dotfile work may execute; unsupported or ineligible work remains non-mutating or is reported as unsupported. A failed confirmed step SHALL produce a failure exit, while the report SHALL retain ordered completed and failed results. The command SHALL make no transaction or rollback guarantee.

#### Scenario: Default and dry-run modes do not mutate

- **GIVEN** a valid explicit target
- **WHEN** `bootstrap` runs in default mode or with `--dry-run`
- **THEN** it SHALL render the corresponding non-mutating execution report
- **AND** it SHALL not instantiate or invoke mutating command execution

#### Scenario: Confirmed modes preserve limits

- **GIVEN** a valid explicit target and `--yes`, optionally with `--sudo`
- **WHEN** eligible provider-backed work is selected
- **THEN** `bootstrap` SHALL apply the same provider, timeout, failure, and sudo limits as `apply`

#### Scenario: Partial confirmed execution is reported honestly

- **GIVEN** one eligible confirmed step succeeds and a later eligible step fails
- **WHEN** `bootstrap` renders the execution report
- **THEN** it SHALL retain the ordered successful and failed results
- **AND** it SHALL exit with failure without claiming atomicity or rollback

### Requirement: Invalid input is rejected before probes or mutation

`bootstrap` SHALL reject unexpected positional arguments, malformed or unsupported `--resource` values, `--dry-run` combined with `--yes`, and `--sudo` without `--yes`. For each such syntactic input failure, it SHALL render usage guidance, exit with a usage failure, and SHALL not start catalog loading, detection, runner, or provider work.

#### Scenario: Conflicting safety flags are rejected

- **GIVEN** `bootstrap` receives both `--dry-run` and `--yes`, or receives `--sudo` without `--yes`
- **WHEN** arguments are validated
- **THEN** it SHALL render usage guidance and exit with a usage failure
- **AND** it SHALL not start detection or execution

#### Scenario: Invalid resource or positional argument is rejected

- **GIVEN** `bootstrap` receives a malformed or unsupported resource reference, or an unexpected positional argument
- **WHEN** arguments are validated
- **THEN** it SHALL render usage guidance and exit with a usage failure
- **AND** it SHALL not start detection or execution

### Requirement: Help is discoverable and side-effect free

Root help SHALL list `bootstrap` as an explicit-selection entrypoint to the safe apply workflow. `bootstrap --help` and `bootstrap -h` SHALL render command usage that includes explicit target and safety-flag options without loading a catalog, detecting the environment, or starting execution.

#### Scenario: Help guides bootstrap use without work

- **GIVEN** the user requests root help or bootstrap help
- **WHEN** help renders
- **THEN** it SHALL expose bootstrap usage and its supported target and safety options
- **AND** it SHALL not perform detection or execution

## Scope Boundary

This specification is an authoritative record of existing behavior. It SHALL NOT require source-code, test, catalog, provider, runtime, or historical `openspec/changes/bootstrap-entrypoint/` changes.
