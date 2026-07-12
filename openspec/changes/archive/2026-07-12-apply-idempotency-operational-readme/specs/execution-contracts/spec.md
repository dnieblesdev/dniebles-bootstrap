# Delta for execution-contracts

## ADDED Requirements

### Requirement: Confirmed execution honors already-installed plan steps

The apply execution boundary MUST treat a plan step with status `already_installed` as unchanged and MUST NOT dispatch an installer or invoke its command runner. This guard MUST be based only on the planning status for that step; it MUST NOT be inferred from prior results, unsupported status, missing configuration, package metadata, or dotfile module presence.

#### Scenario: Confirmed present tool is not dispatched

- GIVEN a confirmed `apply --yes` plan contains an eligible tool or runtime step with status `already_installed`
- WHEN execution processes the plan
- THEN the step produces an unchanged result
- AND the installer and command runner are not called for that step
- AND the report states that no mutation was attempted

#### Scenario: Absent eligible step remains executable

- GIVEN a confirmed plan contains an eligible absent tool or runtime step with an executable status
- WHEN execution processes the plan
- THEN the installer remains eligible for dispatch
- AND the step's existing execution result is preserved

#### Scenario: No-op and dry-run modes remain non-mutating

- GIVEN a plan contains an `already_installed` step
- WHEN default apply or `apply --dry-run` processes the plan
- THEN the mode-specific noop/dry-run result remains unchanged
- AND no command runner is called

### Requirement: Execution results preserve plan order and status outcomes

Confirmed, default, and dry-run reports MUST contain results in the original plan order. Skipped `already_installed` results MUST occupy their original positions. Unsupported and failed results MUST remain unsupported and failed, respectively, and MUST NOT be converted to unchanged merely because another step was detected present. Processing MUST continue according to existing execution semantics after a non-terminal step failure.

#### Scenario: Mixed plan retains order and outcomes

- GIVEN a plan ordered as present eligible, absent eligible, unsupported, and failed
- WHEN execution processes the plan
- THEN the report contains four results in that same order
- AND the present step is unchanged with no mutation attempted
- AND the absent eligible step is dispatched
- AND the unsupported step remains not supported yet
- AND the failed step remains failed

#### Scenario: Failed step does not rewrite other results

- GIVEN a plan contains a detected present step and an executable step whose command fails
- WHEN confirmed execution runs
- THEN the detected step remains unchanged
- AND the failing step remains failed with its existing failure information
- AND later steps follow the existing continued-execution behavior

### Requirement: Bootstrap uses the same apply execution semantics

The `bootstrap` command MUST use the same planning-status guard, result ordering, reporting categories, confirmed/no-op/dry-run mode rules, and failure/unsupported preservation as `apply`. Bootstrap MUST NOT acquire Homebrew or any other dependency as part of this slice; bootstrap reporting remains advisory where existing behavior provides it.

#### Scenario: Bootstrap skips a reliably present resource

- GIVEN `bootstrap` produces an `already_installed` tool or runtime step
- WHEN bootstrap executes in confirmed mode
- THEN the step is reported unchanged with explicit no-mutation wording
- AND no installer command is invoked for that step

#### Scenario: Bootstrap preserves unsupported and failure results

- GIVEN bootstrap contains unsupported or failed steps
- WHEN bootstrap executes
- THEN those results remain not supported yet or failed
- AND result ordering and existing exit behavior are preserved

#### Scenario: Bootstrap does not acquire missing tooling

- GIVEN bootstrap reports a missing provider or bootstrap need
- WHEN bootstrap runs in any supported mode
- THEN it reports the advisory/bootstrap information
- AND it does not clone, fetch, install, retry, or otherwise acquire that dependency
