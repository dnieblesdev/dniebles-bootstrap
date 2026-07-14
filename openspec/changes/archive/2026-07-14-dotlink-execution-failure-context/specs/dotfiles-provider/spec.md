# Delta for dotfiles-provider

## ADDED Requirements

### Requirement: Canonical dotlink execution failure context

The provider MUST derive the executable only from the validated canonical base. A missing runner MUST fail without invoking a command. Non-zero command results MUST retain exit code and sanitized stderr bounded to 4096 bytes, without splitting UTF-8 or terminal escapes.

#### Scenario: Canonical executable
- GIVEN a validated canonical base
- WHEN execution is prepared
- THEN the executable is beneath that base, never a rejected candidate

#### Scenario: Missing runner
- GIVEN a validated base but no runner
- WHEN execution is requested
- THEN a failed result is returned and no command is invoked

#### Scenario: Bounded command failure
- GIVEN non-zero execution with Unicode and terminal escapes in stderr
- WHEN failure is transported
- THEN exit code and sanitized stderr remain, bounded without split UTF-8 or escapes

### Requirement: Command/report failure composition

For confirmed execution, stdout MUST be the only report source. The provider MUST classify unavailable, invalid, and inconsistent reports safely, preserve a valid failed report with the execution error, and retain independent causes for `errors.Is` and `errors.As`. Success, dry-run, and base-resolution contracts MUST remain unchanged.

#### Scenario: Four compositions
- GIVEN success/success, failure/valid-failed, failure/invalid-or-missing, or success/failed-or-inconsistent
- WHEN outcomes are composed
- THEN respectively they succeed, return the report plus execution failure, fail safely, or fail safely

#### Scenario: Valid failed report
- GIVEN command failure and a valid failed stdout report
- WHEN the result is returned
- THEN the failed report, validated entries, safe details, and execution error are retained

#### Scenario: Invalid report
- GIVEN stdout is missing, malformed, contradictory, or invalid
- WHEN composition completes
- THEN an invalid-report failure is returned and stderr is not parsed as a report

#### Scenario: Independent causes
- GIVEN execution and report-validation failures coexist
- WHEN the error is inspected
- THEN `errors.Is` finds both sentinels and `errors.As` finds both typed causes
