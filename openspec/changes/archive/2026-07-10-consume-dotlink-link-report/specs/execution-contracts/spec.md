# Delta for execution-contracts

## ADDED Requirements

### Requirement: Module summaries and per-link outcomes are distinct

The execution layer MUST represent default and `--dry-run` noop work as `not_implemented`.
For confirmed dotfile execution, `StepResult` (or an equivalent execution-owned result) MUST contain both an aggregate module status and ordered per-link details. Per-link detail MUST preserve the Dotlink outcome (`changed`, `unchanged`, `failed`, or `rolled_back`), source, target, safe cause when supplied, and rollback detail when supplied. Legacy `StepStatus` MUST NOT be required to represent every upstream entry outcome.

The aggregate module status MUST be:
- `skipped` when all reported entries are `unchanged` and the aggregate report is successful;
- `installed` when one or more entries are `changed`, no entry failed or rolled back, and the aggregate report is successful;
- `failed` when any entry is `failed` or `rolled_back`, when the aggregate report is failed, or when report/command reconciliation fails.

#### Scenario: Mixed successful entries retain their own outcomes

- GIVEN a confirmed dotfile report has one `changed` entry and one `unchanged` entry
- WHEN the execution layer records the result
- THEN the module aggregate status is `installed`
- AND the ordered per-link details retain one `changed` and one `unchanged` outcome

#### Scenario: Failed aggregate does not erase entry detail

- GIVEN a valid failed report contains changed, failed, or rolled_back entries
- WHEN the execution layer records the result
- THEN the module aggregate status is `failed`
- AND each available per-link outcome, cause, and rollback detail remains available

#### Scenario: Noop apply stays not_implemented

- GIVEN apply runs without confirmed execution
- WHEN the execution layer returns its report
- THEN the status is `not_implemented`
