# Delta for execution-contracts

## ADDED Requirements

### Requirement: Execution contracts are separate from planning

`internal/execution` MUST define its own Installer, Runner, DotfilesProvider, status, result, and report types.
Execution types MUST NOT reuse planning statuses or mutate planning production code.

#### Scenario: Execution types remain distinct

- GIVEN planning types already exist
- WHEN execution contracts are defined
- THEN execution uses separate types and status vocabulary

#### Scenario: Planning production stays unchanged

- GIVEN the change is implemented
- WHEN planning code is inspected
- THEN planning production behavior remains unchanged

### Requirement: Noop execution is safe and non-mutating

Noop installers, runners, and providers MUST return a `not_implemented` execution result or report for unsupported work.
They MUST NOT invoke real commands, apply changes, or mutate the host.

#### Scenario: Unsupported action returns not_implemented

- GIVEN a noop execution dependency receives a request
- WHEN the request is handled
- THEN the result or report status is `not_implemented`

#### Scenario: No mutation occurs in noop mode

- GIVEN a noop execution dependency is used
- WHEN it runs
- THEN no apply, install, clone, or filesystem mutation occurs

### Requirement: Runner dispatches plan steps sequentially by kind

The Runner MUST consume a planning Plan sequentially and dispatch each step to the installer for that step's resource kind.
The Runner MUST preserve plan order and MUST return an execution report for all processed steps.

#### Scenario: Steps run in plan order

- GIVEN a plan contains multiple steps with different kinds
- WHEN the Runner executes the plan
- THEN steps are dispatched in the same order they appear in the plan

#### Scenario: Kind selects the installer

- GIVEN a plan step targets a specific resource kind
- WHEN the Runner dispatches the step
- THEN the matching installer handles that step

### Requirement: DotfilesProvider is a high-level execution boundary

DotfilesProvider MUST expose high-level execution operations for dotfiles workflow support.
It MUST remain separate from the read-only dotfiles detector and MUST NOT own planning logic.

#### Scenario: Provider is execution-only

- GIVEN a dotfiles operation is requested
- WHEN the provider is used
- THEN it serves execution-layer behavior only

#### Scenario: Provider does not own planning

- GIVEN planning is being built
- WHEN dotfiles execution support is considered
- THEN DotfilesProvider does not change planning behavior

### Requirement: Execution contracts remain non-mutating for apply

`internal/execution` MUST remain a safe, non-mutating boundary used by `apply`.
The command MUST use noop execution contracts only, and MUST NOT introduce real execution, host mutation, installers with side effects, or planning production changes.
(Previously: The execution slice prohibited any apply command or CLI wiring.)

#### Scenario: Apply uses noop execution contracts only

- GIVEN the `apply` command runs
- WHEN execution is dispatched
- THEN only noop results are produced

#### Scenario: Side effects remain absent

- GIVEN execution contracts are present
- WHEN `apply` is reviewed end-to-end
- THEN no real execution or production mutation occurs
