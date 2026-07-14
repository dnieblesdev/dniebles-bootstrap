# Delta for dotfiles-provider

## ADDED Requirements

### Requirement: Dotfiles failures preserve truthful phase context

The provider MUST preserve the operation, selected module resources and module names, failure phase, attempted candidate or executable, and a safe concrete cause across resolution, prerequisite validation, command execution, and report validation. An attempted candidate MUST NOT be presented as validated or canonical.

#### Scenario: Resolution failure
- GIVEN base resolution rejects an environment or home candidate for a selected `dotfile:bash`
- WHEN the provider returns the failure
- THEN it identifies `dotfile:bash` and `bash`, the resolution phase, attempted candidate, and safe cause
- AND no canonical base or executable is claimed

#### Scenario: Missing executable prerequisite
- GIVEN the canonical repository is present but `bin/dotlink` is missing
- WHEN confirmed apply requests `dotfile:bash`
- THEN apply is non-zero and reports `dotfile:bash`/`bash`, the prerequisite or repository-validation phase, the attempted runner candidate as unvalidated, and a concrete missing-path cause
- AND the command runner receives zero calls

#### Scenario: Module prerequisite failure
- GIVEN the base and runner are valid but a selected module directory is absent or unsafe
- WHEN prerequisites are validated
- THEN the failure identifies the selected module, prerequisite phase, attempted relevant path, and safe cause
- AND no command is requested

### Requirement: Existing execution and report contracts survive diagnostic completion

The provider MUST preserve existing statuses, command arguments and ordering, typed `errors.Is`/`errors.As` classification, valid failed-report translation, and the rule that stdout is the only structured report source. It MUST NOT infer link outcomes from a prerequisite rejection.

#### Scenario: Command failure with valid failed report
- GIVEN confirmed execution returns a non-zero command result and a valid failed report
- WHEN the provider composes the result
- THEN the failed report and execution error remain available, statuses and command semantics are unchanged, and both typed causes remain classifiable

#### Scenario: Invalid report after command execution
- GIVEN stdout is absent, malformed, contradictory, or invalid
- WHEN report validation fails
- THEN the result is safely failed with a report-validation phase and concrete safe cause
- AND stderr is not treated as report input

#### Scenario: Prerequisite rejection has no inferred links
- GIVEN prerequisite validation rejects execution before a command runs
- WHEN the result is translated
- THEN no changed, unchanged, failed, or rolled-back link is inferred from the selected module
