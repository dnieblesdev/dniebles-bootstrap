# Delta for execution-contracts

## ADDED Requirements

### Requirement: Execution reporting renders phase-specific diagnostic context once

Execution reporting MUST render truthful operation, selected modules, phase, attempted candidate or executable, and safe concrete cause for resolution, prerequisite, command-execution, and report-validation failures. It MUST deduplicate shared base context without hiding a distinct attempted candidate or executable.

#### Scenario: Resolution and prerequisite distinction
- GIVEN a rejected base candidate or a missing `bin/dotlink` prerequisite
- WHEN the failed result is rendered
- THEN the output names the respective resolution or prerequisite/repository-validation phase
- AND it never labels the rejected candidate as canonical or validated

#### Scenario: Command execution detail
- GIVEN a command was invoked and failed
- WHEN the result is rendered
- THEN the output identifies command-execution phase, executable/runner/command, exit status, and bounded terminal-safe stderr
- AND shared base context appears once

#### Scenario: Report validation detail
- GIVEN command execution completed but its stdout report is missing or invalid
- WHEN the result is rendered
- THEN the output identifies report-validation phase and a safe concrete cause
- AND it does not reinterpret stderr or invent link outcomes

### Requirement: Diagnostic safety and existing execution contracts are preserved

Diagnostic output MUST retain existing statuses, command semantics, valid failed-report translation, and independently unwrap-able typed causes. Stderr MUST remain bounded and sanitized without exposing unsafe terminal control sequences. Tests MUST prove zero command-runner calls for prerequisite rejection and preserve `errors.Is`/`errors.As` behavior.

#### Scenario: Missing-runner acceptance anchor
- GIVEN a confirmed apply selects `dotfile:bash`/`bash` and the repository lacks `bin/dotlink`
- WHEN the apply report is rendered
- THEN it is non-zero, names prerequisite/repository-validation, shows the attempted unvalidated runner candidate and missing-path cause
- AND the command runner call count is zero

#### Scenario: Candidate remains distinct after deduplication
- GIVEN a result contains base context plus a different attempted executable candidate
- WHEN it is rendered
- THEN common base fields appear once
- AND the distinct candidate remains visible and unvalidated

#### Scenario: Scope boundary
- GIVEN this diagnostic change is implemented
- WHEN contracts and behavior are reviewed
- THEN it does not require legacy `DotfilesBaseReporter`, provider redesign, parser redesign, new statuses, planning/configuration changes, monolith cleanup, or `PlanStep.AttentionReasons` propagation to `StepResult`
