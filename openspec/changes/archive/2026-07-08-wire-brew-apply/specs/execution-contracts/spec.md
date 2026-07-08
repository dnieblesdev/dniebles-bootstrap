# Delta for execution-contracts

## MODIFIED Requirements

### Requirement: Execution contracts remain non-mutating for apply

`internal/execution` MUST remain a safe boundary used by `apply`.
The command MUST use noop execution contracts by default and in `--dry-run`, MUST allow real execution only for confirmed brew-backed tool/package steps, MUST surface Homebrew bootstrap reporting as advisory data only, and MUST NOT introduce real execution outside that narrow path.
(Previously: execution contracts were entirely noop for apply.)

#### Scenario: Apply uses noop execution contracts by default

- GIVEN the `apply` command runs without `--yes`
- WHEN execution is dispatched
- THEN only noop results are produced

#### Scenario: Confirmed brew steps may execute

- GIVEN `apply --yes` and a brew-backed tool/package step
- WHEN execution is dispatched
- THEN real brew execution is allowed for that step only

#### Scenario: Side effects remain absent outside confirmed brew steps

- GIVEN execution contracts are present
- WHEN `apply` is reviewed end-to-end
- THEN no real execution or production mutation occurs outside confirmed brew-backed steps

#### Scenario: Bootstrap data stays advisory

- GIVEN Homebrew bootstrap need data is attached
- WHEN execution contracts report it
- THEN the data remains non-mutating and reviewable

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
