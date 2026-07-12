# Delta for apply-command-dry-run

## ADDED Requirements

### Requirement: Apply reports idempotent no-mutation results

When confirmed apply receives a plan step marked `already_installed` for a reliably command-presence-detected tool or runtime, it MUST report that step in its original plan position as `unchanged` and MUST explicitly state that no mutation was attempted. It MUST NOT dispatch an installer or command runner for the step.

#### Scenario: Confirmed apply skips a detected tool

- GIVEN `apply --yes` has a selected eligible tool or runtime step with status `already_installed`
- WHEN execution reporting completes
- THEN the result is `unchanged`
- AND the result appears in the step's original plan position
- AND the output says that no mutation was attempted
- AND no installer command-runner call occurred for that step

#### Scenario: Dry-run reports mode-specific non-mutation

- GIVEN `apply --dry-run` has a selected step marked `already_installed`
- WHEN execution reporting completes
- THEN the existing dry-run/not-supported-yet behavior is preserved
- AND the output does not claim that a confirmed mutation was skipped
- AND no command runner is called

#### Scenario: Default apply remains a safe noop

- GIVEN default `apply` has a selected step marked `already_installed`
- WHEN execution reporting completes
- THEN the existing default non-mutating noop result is preserved
- AND no command runner is called

### Requirement: Apply preserves mixed-plan ordering and outcomes

Apply MUST preserve the original plan order in every mode. Idempotency handling MUST change only eligible command-presence-detected steps marked `already_installed`; absent eligible steps MUST retain dispatch eligibility, and unsupported, failed, and other non-matching statuses MUST retain their existing user-facing categories, diagnostics, and exit behavior.

#### Scenario: Mixed confirmed plan is reported in order

- GIVEN a plan contains, in order, a detected present step, an absent eligible step, an unsupported step, and a failed step
- WHEN `apply --yes` executes
- THEN results are rendered in that same order
- AND the detected step is unchanged with no mutation attempted
- AND the absent eligible step is dispatched
- AND the unsupported step is reported as not supported yet
- AND the failed step is reported as failed

#### Scenario: Failure and unsupported status are not masked

- GIVEN one selected step fails while another selected step is unsupported
- WHEN confirmed apply renders its report
- THEN both original outcomes remain visible
- AND the confirmed command retains its existing non-zero behavior for an eligible failure

### Requirement: Apply safety boundaries exclude broader convergence

The idempotency promise MUST be limited to reliable command-presence detection for eligible tools and runtimes. Apply MUST NOT use package presence, package version, configuration state, dotfile-link content, retries, rollback, or bootstrap acquisition to decide that a step is unchanged or to make it converge.

#### Scenario: Dotfile module presence is not link convergence

- GIVEN a dotfile module directory is detected as present
- WHEN apply plans or executes the selection
- THEN apply does not claim that dotfile links are current
- AND this idempotency guard does not skip dotfile link convergence

#### Scenario: No retry or rollback is implied

- GIVEN an eligible installer fails during confirmed apply
- WHEN the result is reported
- THEN the failure remains failed
- AND no automatic retry or rollback is attempted or claimed

#### Scenario: Missing bootstrap dependency remains advisory

- GIVEN a provider reports that bootstrap is needed
- WHEN apply runs in default, dry-run, or confirmed mode
- THEN bootstrap guidance remains advisory according to existing behavior
- AND apply does not acquire or install the bootstrap dependency
