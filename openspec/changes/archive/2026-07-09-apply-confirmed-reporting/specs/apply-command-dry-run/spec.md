# Delta for apply-command-dry-run

## ADDED Requirements

### Requirement: Apply execution summary is always rendered

The apply command MUST render a Summary section in default, `--dry-run`, and `--yes` modes when execution results exist.
The Summary MUST use the user-facing categories `changed`, `unchanged`, `not supported yet`, and `failed`.

#### Scenario: Summary appears in default mode

- GIVEN the user runs `dbootstrap apply` successfully
- WHEN execution reporting is rendered
- THEN the output includes a Summary section
- AND the Summary uses the user-facing categories

#### Scenario: Summary appears in dry-run mode

- GIVEN the user runs `dbootstrap apply --dry-run` successfully
- WHEN execution reporting is rendered
- THEN the output includes a Summary section
- AND the Summary uses the user-facing categories

#### Scenario: Summary appears in confirmed mode

- GIVEN the user runs `dbootstrap apply --yes` successfully
- WHEN execution reporting is rendered
- THEN the output includes a Summary section
- AND the Summary uses the user-facing categories

### Requirement: Empty selected plans render an explicit empty state

The apply command MUST render a clear empty-state sentence when the selected plan has zero actionable or selected steps.
The command MUST NOT render a zero-count summary table for that case.

#### Scenario: Empty selected plan shows a sentence

- GIVEN the selected plan has no actionable or selected steps
- WHEN apply renders execution reporting
- THEN the output contains a clear empty-state sentence
- AND no zero-count summary table is shown

## MODIFIED Requirements

### Requirement: Apply renders execution mode-specific reporting

The apply command MUST render an execution report separate from plan rendering.
Successful dry-run execution MUST report `not_implemented` results, while confirmed mode MAY report real brew execution for brew-backed tool/package steps only.
Homebrew bootstrap reporting MUST remain advisory and non-mutating in default and `--dry-run` modes.
User-facing step output MUST describe internally `not_implemented` work as `not supported yet`.
Confirmed `--yes` output MUST explicitly state that only brew-backed `tool` and `package` steps may have changed the machine; unsupported or non-brew resource work remains non-mutating and `not supported yet`.
(Previously: Successful dry-run execution MUST report `not_implemented` results, while confirmed mode MAY report real brew execution for brew-backed tool/package steps only. Homebrew bootstrap reporting MUST remain advisory and non-mutating in default and `--dry-run` modes.)

#### Scenario: Dry-run execution reports not_implemented

- GIVEN a valid plan is produced
- WHEN `dbootstrap apply` runs the execution phase
- THEN each step is internally recorded as `not_implemented`
- AND user-facing output describes the step as `not supported yet`

#### Scenario: Execution rendering is distinct from plan rendering

- GIVEN both plan and apply commands are available
- WHEN their output is rendered
- THEN apply output is clearly labeled as execution reporting
- AND plan rendering remains separate

#### Scenario: Confirmed brew steps can report real execution

- GIVEN a brew-backed tool or package step is present
- WHEN `dbootstrap apply --yes` runs the execution phase
- THEN the step may report real brew execution
- AND other step kinds remain non-mutating or unsupported
