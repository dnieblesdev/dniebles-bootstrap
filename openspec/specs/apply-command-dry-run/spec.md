# Delta for apply-command-dry-run

## ADDED Requirements

### Requirement: Apply command exists with plan-style target flags

The `apply` command MUST exist and MUST accept `--profile`, repeatable `--resource`, and `--catalog` using the same validation rules as `plan`.
The command MUST surface Homebrew bootstrap reporting when Homebrew-backed resources are missing `brew`, while keeping every accepted mode non-mutating.

#### Scenario: Apply accepts the same targets as plan

- GIVEN the user provides valid `--profile`, `--resource`, or `--catalog` values
- WHEN `dbootstrap apply` runs
- THEN the command accepts the same target surface as `plan`

#### Scenario: Invalid target input is rejected

- GIVEN the user provides malformed or unsupported target input
- WHEN `dbootstrap apply` runs
- THEN the command fails with a clear validation error

#### Scenario: Missing brew is reported without mutation

- GIVEN a Homebrew-backed resource is selected
- WHEN `dbootstrap apply` runs on a host without `brew`
- THEN the report includes a bootstrap action
- AND no host mutation occurs

### Requirement: Apply reuses the planning pipeline

The `apply` command MUST build its request through the existing planning pipeline before any execution report is produced.
Planning failures MUST stop the command before execution begins.

#### Scenario: Planning failure stops apply early

- GIVEN planning cannot build a valid plan
- WHEN `dbootstrap apply` runs
- THEN the command exits with the planning error
- AND no execution report is rendered

#### Scenario: Successful planning continues to execution

- GIVEN planning succeeds
- WHEN `dbootstrap apply` runs
- THEN the resulting plan is passed into execution

### Requirement: Apply execution summary is always rendered

The apply command MUST render a Summary section in default, `--dry-run`, and `--yes` modes when execution results exist.
The Summary MUST use the user-facing categories `changed`, `unchanged`, `not supported yet`, and `failed`.
Confirmed dotfile summaries MAY additionally reflect aggregate module categories: `changed` for installed, `unchanged` for skipped, and `failed` for failed aggregates. A failed aggregate with rolled-back entries MAY include a `rolled_back` breakdown, but it MUST remain failed for exit status.

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

### Requirement: Apply renders execution mode-specific reporting

The apply command MUST render an execution report separate from plan rendering.
Successful dry-run execution MUST report `not_implemented` results, while confirmed mode MAY report real brew execution for brew-backed tool/package steps and MAY run selected dotfile resources through the dotfiles execution provider.
Homebrew bootstrap reporting MUST remain advisory and non-mutating in default and `--dry-run` modes.
User-facing step output MUST describe internally `not_implemented` work as `not supported yet`.
Confirmed `--yes` output MUST explicitly state that brew-backed `tool` and `package` steps and selected dotfile resources may have changed the machine; unsupported, non-brew, and unselected work remains non-mutating or `not supported yet`.
When dotfiles execution is attempted, output MUST render the aggregate module result and every validated per-link detail. A resolution failure MUST show attempted source/candidate, selected modules, and safe cause; it MUST show `canonical base` only after successful canonicalization and validation.

#### Scenario: Dry-run execution reports not_implemented

- GIVEN a valid plan is produced
- WHEN `dbootstrap apply --dry-run` runs the execution phase
- THEN each step is internally recorded as `not_implemented`
- AND user-facing output describes the step as `not supported yet`
- AND no dotfiles command runner is used

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

#### Scenario: Confirmed dotfile steps can report real execution

- GIVEN a selected plan contains `dotfile:bash`
- AND dotfiles prerequisites are valid
- WHEN `dbootstrap apply --yes --resource dotfile:bash` runs the execution phase
- THEN dotlink may be requested through the configured command runner for module `bash`
- AND the dotfile result may be reported as changed

#### Scenario: Dotfiles execution context is reported

- GIVEN `dbootstrap apply --yes` reaches the dotfiles execution path
- AND base resolution succeeds and is validated
- WHEN execution reporting is rendered
- THEN the output includes the canonical dotfiles base path
- AND the output includes whether the base came from `DBOOTSTRAP_DOTFILES_DIR` or the home convention
- AND the output includes the selected module names

#### Scenario: Resolution failure is not mislabeled

- GIVEN `dbootstrap apply --yes` reaches the dotfiles execution path
- AND base resolution fails before validation
- WHEN execution reporting is rendered
- THEN attempted source/candidate, selected modules, and safe cause are shown
- AND no attempted candidate is labeled canonical base

### Requirement: Apply remains strictly non-mutating

The `apply` command MUST NOT perform real execution, host mutation, dotlink, clone, sparse checkout, retry, or concurrency behavior in default mode or `--dry-run` mode.
It MUST remain a safe noop bridge over the existing plan unless `--yes` is explicitly provided.
In confirmed `--yes` mode, mutation MUST remain limited to brew-backed tool/package execution and selected dotfile resource execution.

#### Scenario: No host mutation occurs

- GIVEN `dbootstrap apply` runs successfully
- WHEN the command completes
- THEN no filesystem or host state is mutated

#### Scenario: Dry-run apply has no host mutation

- GIVEN `dbootstrap apply --dry-run` runs successfully
- WHEN the command completes
- THEN no filesystem or host state is mutated

#### Scenario: No acquisition or orchestration features are introduced outside confirmed eligible execution

- GIVEN `dbootstrap apply` runs in default mode, dry-run mode, or confirmed mode for selected dotfile resources
- WHEN the execution path is reviewed
- THEN no retry, clone, pull, submodule, fetch, remote acquisition, sparse checkout, or apt behavior is present
- AND dotlink may be requested only for confirmed selected dotfile resources through the configured command runner

### Requirement: Apply mode is explicit and safe by default

The `apply` command MUST treat the default mode as non-mutating, MUST treat `--dry-run` as explicit non-mutating, and MUST treat `--yes` as the only confirmed mode that may mutate for brew-backed tool/package steps and selected dotfile resource steps.
The command MUST keep Homebrew bootstrap reporting non-mutating in default and `--dry-run` modes.

#### Scenario: Default apply is non-mutating

- GIVEN the user runs `dbootstrap apply` with no safety flags
- WHEN the command executes
- THEN the selected mode is reported as non-mutating default
- AND no host mutation is performed

#### Scenario: Dry-run is explicit non-mutating

- GIVEN the user runs `dbootstrap apply --dry-run`
- WHEN the command executes
- THEN the selected mode is reported as dry-run
- AND no host mutation is performed

#### Scenario: Yes is the only confirmed mutating mode

- GIVEN the user runs `dbootstrap apply --yes` with a brew-backed tool, brew-backed package, or selected dotfile step
- WHEN the command executes
- THEN the selected mode is reported as confirmed mode
- AND real execution may be attempted only for those eligible steps

### Requirement: Conflicting safety flags are rejected

The `apply` command MUST reject `--dry-run --yes` as invalid input and MUST return a clear usage error.

#### Scenario: Dry-run and yes cannot be combined

- GIVEN the user runs `dbootstrap apply --dry-run --yes`
- WHEN the command validates flags
- THEN the command fails with a usage error
- AND no execution result is produced

### Requirement: Confirmed mode only wires eligible real execution

The `apply` command MUST wire real execution only for brew-backed `tool` and `package` steps and selected `dotfile` steps when `--yes` is set.
Runtime, non-brew, unselected, and unsupported steps MUST remain noop or unsupported.
Dotfile execution MUST use the existing dotfiles execution provider and MUST run only through configured composition seams.

#### Scenario: Brew-backed steps may execute under yes

- GIVEN a brew-backed `tool` step and `--yes`
- WHEN apply executes
- THEN the step is eligible for real brew installation

#### Scenario: Selected dotfile steps may execute under yes

- GIVEN a selected `dotfile:bash` step and `--yes`
- WHEN apply executes
- THEN the step is eligible for dotlink execution for module `bash`

#### Scenario: Non-eligible steps stay non-mutating

- GIVEN a runtime, non-brew, unsupported, or unselected step and `--yes`
- WHEN apply executes
- THEN the step remains noop or returns unsupported

#### Scenario: Missing dotfiles prerequisites fail safely

- GIVEN a selected dotfile step and `--yes`
- AND the dotfiles base, `bin/dotlink`, or selected module is missing
- WHEN apply executes
- THEN the dotfile step is reported as failed with understandable text
- AND no real command is invoked

### Requirement: Confirmed apply exits non-zero when eligible execution fails

When `apply --yes` attempts eligible real execution and any eligible step reports `failed`, the CLI MUST return a non-zero exit status after rendering the execution report.
This includes dotfiles failures caused by missing base path, missing `bin/dotlink`, missing selected module, command-runner failure, or command timeout.
Default apply and `--dry-run` MUST remain non-mutating and MUST NOT use this rule to imply real execution was attempted.

#### Scenario: Missing dotfiles prerequisite makes confirmed apply fail

- GIVEN a selected dotfile resource is eligible under `apply --yes`
- AND the dotfiles base path, dotlink executable, or selected module is missing
- WHEN apply renders the execution report
- THEN the dotfile step is reported as failed
- AND the CLI exits non-zero

#### Scenario: Dotlink runner failure makes confirmed apply fail

- GIVEN a selected dotfile resource is eligible under `apply --yes`
- AND the injected command runner reports failure or timeout
- WHEN apply renders the execution report
- THEN the dotfile step is reported as failed
- AND the CLI exits non-zero
- AND no retry or fallback acquisition is attempted

## REMOVED Requirements

### Requirement: No apply command is introduced

(Reason: `apply` is now intentionally added as a functional CLI bridge: default and `--dry-run` stay non-mutating, while `--yes` may run the narrow confirmed Homebrew path.)
(Migration: Replace this gate with functional `apply` coverage.)
