# Delta for command-runner

## ADDED Requirements

### Requirement: Explicit command representation

The system MUST execute commands as executable-plus-args data and MUST NOT require shell strings.

#### Scenario: Run an explicit command

- GIVEN an executable and argument list
- WHEN the runner executes the command
- THEN the command is invoked without shell interpretation
- AND the result records the executable and args

#### Scenario: Reject shell-first input

- GIVEN a shell string or pipeline-only representation
- WHEN the runner is asked to execute it
- THEN the request is rejected or translated only through an explicit executable-plus-args contract

### Requirement: Structured execution results

The system MUST return stdout, stderr, exit status, and exit code as structured result data.

#### Scenario: Successful command completes

- GIVEN a command that exits successfully
- WHEN the runner completes execution
- THEN stdout and stderr are captured separately
- AND the result includes a success status and exit code

#### Scenario: Failing command reports failure details

- GIVEN a command that exits non-zero
- WHEN the runner completes execution
- THEN the result includes failure status and exit code
- AND stdout and stderr remain available for inspection

### Requirement: Context-aware cancellation

The system MUST support context or timeout cancellation and MUST surface cancellation as execution failure data.

#### Scenario: Timeout cancels execution

- GIVEN a command that exceeds its timeout
- WHEN the timeout is reached
- THEN execution is canceled
- AND the result indicates cancellation

#### Scenario: External context cancellation stops execution

- GIVEN a live command and a canceled context
- WHEN the runner observes cancellation
- THEN execution stops promptly
- AND no success result is reported

### Requirement: Deterministic no-op dry run

The system MUST support deterministic dry-run behavior that does not mutate state or perform real execution.

#### Scenario: Dry run returns planned result

- GIVEN dry-run mode and an executable-plus-args command
- WHEN the runner is invoked
- THEN no process is started
- AND the result is deterministic

#### Scenario: Dry run preserves non-mutating behavior

- GIVEN dry-run mode during apply-related orchestration
- WHEN the runner is used
- THEN no installer action is performed
- AND no host mutation occurs
