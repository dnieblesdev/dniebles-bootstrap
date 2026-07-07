# Delta for execution-contracts

## MODIFIED Requirements

### Requirement: Execution contracts remain non-mutating for apply

`internal/execution` MUST remain a safe, non-mutating boundary used by `apply`.
The command MUST use noop execution contracts only, MUST surface Homebrew bootstrap reporting as advisory data only, and MUST NOT introduce real execution, host mutation, installers with side effects, or planning production changes.
(Previously: The execution slice prohibited any apply command or CLI wiring.)

#### Scenario: Apply uses noop execution contracts only

- GIVEN the `apply` command runs
- WHEN execution is dispatched
- THEN only noop results are produced

#### Scenario: Side effects remain absent

- GIVEN execution contracts are present
- WHEN `apply` is reviewed end-to-end
- THEN no real execution or production mutation occurs

#### Scenario: Bootstrap data stays advisory

- GIVEN Homebrew bootstrap need data is attached
- WHEN execution contracts report it
- THEN the data remains non-mutating and reviewable
