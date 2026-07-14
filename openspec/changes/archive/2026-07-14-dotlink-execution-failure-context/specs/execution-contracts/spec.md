# Delta for execution-contracts

## ADDED Requirements

### Requirement: Transport and presentation of failed dotlink context

The execution boundary MUST transport executable, runner, command, exit code, bounded sanitized stderr, report status, and independently unwrap-able causes. The installer MUST return a valid failed report with its execution error. Presentation MUST render execution facts without duplicate base context.

#### Scenario: Installer transport
- GIVEN command failure with a valid failed report
- WHEN the installer translates it
- THEN the result is failed, includes the validated report, and retains the execution cause

#### Scenario: Single base rendering
- GIVEN canonical-base context and execution detail in a failed result
- WHEN it is rendered
- THEN base context appears once and executable, runner, command, exit code, and sanitized stderr appear as execution detail

#### Scenario: Separate test structures
- GIVEN failure context is tested
- WHEN tests run
- THEN structural tests prove fields, report/error transport, and `errors.Is`/`errors.As`; presentation tests separately prove bounded sanitization and single base rendering

### Requirement: Existing contracts remain unchanged

The system MUST preserve success, default, and dry-run behavior and consume merged base-resolution context without resolving, validating, or relabeling it. Acquisition, rollback, dotlink semantics, packages, and capabilities are excluded.

#### Scenario: Success and dry-run
- GIVEN successful confirmed execution or default/dry-run execution
- WHEN processed
- THEN existing result and non-mutating behavior remain unchanged

#### Scenario: Base identity
- GIVEN validated canonical context or safe resolution failure from the merged resolver
- WHEN failure detail is attached
- THEN supplied base identity and classification remain unchanged
