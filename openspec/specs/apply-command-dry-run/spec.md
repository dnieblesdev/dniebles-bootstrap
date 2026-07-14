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

### Requirement: APT detection is confirmed-Linux-only

The CLI MUST compose APT detection only for confirmed Linux `apply --yes` and `bootstrap` APT steps. APT presence MUST be installed only for a well-formed status whose error field is `ok` and package-status field is `installed`, including `hold ok installed`. Partial states such as `unpacked` or `half-configured` MUST remain executable rather than being skipped. Default, dry-run, planning-only, and non-Linux flows MUST NOT probe `dpkg-query`; non-Linux confirmed APT steps MUST fail without probes.

#### Scenario: Definitive not-found reaches installer
- GIVEN confirmed Linux `apply --yes` or `bootstrap` has an eligible absent package
- WHEN the query exits 1 with matching `no packages found matching <package>` stderr and no contradictory stdout
- THEN it is classified absent and dispatched through the normal APT installer
- AND no retry, fallback, or alternate probe occurs

#### Scenario: Held installed package is skipped
- GIVEN confirmed Linux `apply --yes` or `bootstrap` returns `hold ok installed`
- WHEN execution processes the package
- THEN it reports unchanged without dispatching APT

#### Scenario: Partial package state is not skipped
- GIVEN confirmed Linux `apply --yes` or `bootstrap` returns `install ok unpacked` or `install ok half-configured`
- WHEN execution processes the package
- THEN it dispatches the normal APT installer

#### Scenario: Safe or non-Linux modes do not probe
- GIVEN a safe-mode or non-Linux command has APT-backed resources
- WHEN it runs
- THEN none of `dpkg-query`, `apt-get`, or `sudo` is invoked

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
Confirmed `--yes` output MUST explicitly state that brew-backed `tool` and `package` steps with their existing cross-platform eligibility, Linux APT-backed `tool` and `package` steps, and selected dotfile resources may have changed the machine; unsupported, non-provider-backed, and unselected work remains non-mutating or `not supported yet`.
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
In confirmed `--yes` mode, mutation MUST remain limited to brew-backed tool/package execution with its existing cross-platform eligibility, Linux APT-backed tool/package execution, and selected dotfile resource execution. APT direct execution MUST use `apt-get install -y -- <package>` with a ten-minute `CommandRequest.Timeout`; only explicit `--yes --sudo` may use `sudo apt-get install -y -- <package>` with the same bound, and `--sudo` outside `--yes` MUST be rejected. APT package metadata MUST be trimmed, non-empty, and not begin with `-`; `--` prevents option injection from custom catalog metadata and is not shell escaping.

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

The `apply` command MUST treat the default mode as non-mutating, MUST treat `--dry-run` as explicit non-mutating, and MUST treat `--yes` as the only confirmed mode that may mutate for brew-backed tool/package steps with their existing cross-platform eligibility, Linux APT-backed tool/package steps, and selected dotfile resource steps. `--yes` uses direct APT execution; only explicit `--yes --sudo` uses sudo.
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

- GIVEN the user runs `dbootstrap apply --yes` with a brew-backed tool/package, Linux APT-backed tool/package, or selected dotfile step
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

The `apply` command MUST wire real execution only for brew-backed `tool` and `package` steps with their existing cross-platform eligibility, Linux APT-backed `tool` and `package` steps, and selected `dotfile` steps when `--yes` is set. APT MUST be unavailable to non-Linux confirmed execution: a selected APT step MUST be `StepStatusFailed`, cause a non-zero confirmed outcome, and make zero apt/sudo probe or command calls. Missing `apt-get` or `sudo` MUST produce failed results without command execution.
Runtime, non-provider-backed, unselected, and unsupported steps MUST remain noop or unsupported.
Dotfile execution MUST use the existing dotfiles execution provider and MUST run only through configured composition seams.

#### Scenario: Brew-backed steps may execute under yes

- GIVEN a brew-backed `tool` step and `--yes`
- WHEN apply executes
- THEN the step is eligible for real brew installation

#### Scenario: Selected dotfile steps may execute under yes

- GIVEN a selected `dotfile:bash` step and `--yes`
- WHEN apply executes
- THEN the step is eligible for dotlink execution for module `bash`

#### Scenario: Non-provider-backed steps stay non-mutating

- GIVEN a runtime, non-provider-backed, unsupported, or unselected step and `--yes`
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
This includes selected non-Linux APT, APT command timeout, and dotfiles failures caused by missing base path, missing `bin/dotlink`, missing selected module, command-runner failure, or command timeout. APT failure MUST NOT cause retry or rollback claims.
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

### Requirement: Confirmed Brew package reports explicit no-mutation idempotency

When confirmed `apply` or `bootstrap` positively proves an eligible Brew formula is installed, the command MUST report the step in its original order as unchanged/already installed and MUST explicitly state that no mutation was attempted. The command MUST NOT dispatch the Brew installer for that step.

#### Scenario: Confirmed apply reports installed formula without mutation

- GIVEN `apply --yes` selects a Brew-backed package
- AND `brew list --formula <InstallMetadata.Package>` completes successfully
- WHEN the execution report is rendered
- THEN the package is reported unchanged/already installed in plan order
- AND the output explicitly says that no mutation was attempted
- AND no Brew install command is requested

#### Scenario: Confirmed bootstrap reports installed formula without mutation

- GIVEN confirmed `bootstrap` selects an eligible Brew-backed package
- AND its exact presence query completes successfully
- WHEN the execution report is rendered
- THEN the package is reported unchanged/already installed in plan order
- AND no installer command is requested

### Requirement: Query uncertainty is visible and never authorizes installation

When a Brew presence query is unavailable, times out, fails, returns an unclassified non-zero result, or cannot be formed from supported formula metadata, confirmed `apply` and `bootstrap` MUST render an attention/failure outcome for the affected package. They MUST NOT render `already_installed` or absent for that package and MUST NOT invoke its installer. Other steps MUST retain existing report order and continued-execution behavior.

#### Scenario: Missing Brew is reported conservatively

- GIVEN a confirmed Brew-backed package is selected
- AND `brew` is unavailable
- WHEN the command renders its report
- THEN the package is visibly reported as attention/failure
- AND no Brew installer command is invoked

#### Scenario: Timeout or ambiguous result is reported conservatively

- GIVEN a confirmed Brew-backed package query times out or returns an unclassified non-zero result
- WHEN the command renders its report
- THEN the package is visibly reported as attention/failure
- AND it is not reported as already installed or absent
- AND no installer command is invoked

### Requirement: Apply safety boundaries exclude broader convergence

The idempotency promise MUST be limited to reliable command presence for eligible tools and runtimes and positive read-only formula presence for eligible Brew-backed packages. Apply MUST NOT use package versions, configuration state, dotfile-link content, retries, fallback queries, bootstrap acquisition, casks, APT, or other provider detection to decide that a step is unchanged or to make it converge. A successful Brew formula query MUST NOT imply a version, executable health, PATH/link/configuration correctness, or dotfile convergence.
(Previously: package presence was excluded from the idempotency promise entirely.)

#### Scenario: Brew formula presence is the only package exception

- GIVEN a confirmed eligible Brew-backed package has a successful exact formula presence query
- WHEN apply determines the step outcome
- THEN it MAY mark the step already installed and skip only that package's installer
- AND no version or broader convergence claim is made

#### Scenario: Non-Brew and broader checks remain excluded

- GIVEN a selected APT package, cask, tool/runtime, dotfile, or unsupported resource is evaluated
- WHEN apply determines its outcome
- THEN this Brew formula presence rule is not used
- AND no package-version, configuration, dotfile-link, retry, fallback, or bootstrap probe is introduced

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

### Requirement: Bootstrap has full apply command parity

The `bootstrap` command MUST expose the apply command's target, catalog,
safety, planning, provider, reporting, and exit contracts while requiring an
explicit `--profile` and/or repeatable `--resource`. Root help MUST list and
describe `bootstrap`, and command-specific help MUST provide its usage.
`apply` MUST remain backward-compatible and MUST NOT gain an implicit target or
changed default behavior. Both commands MUST use one shared orchestration path.

#### Scenario: Explicit target parity

- GIVEN the same valid profile/resource selection and catalog
- WHEN `apply` and `bootstrap` are invoked
- THEN planning, environment/config/dotfiles state, provider eligibility, report, and exit outcome match

#### Scenario: Safety mode parity

- GIVEN a valid explicit target
- WHEN each command runs in default, `--dry-run`, `--yes`, or `--yes --sudo` mode
- THEN both commands enforce identical mutation, Brew, APT, dotfile, and sudo behavior

#### Scenario: Syntactic input failures are consistent

- GIVEN missing targets, malformed selections, unexpected positionals, `--dry-run --yes`, or `--sudo` without `--yes`
- WHEN either command validates input
- THEN it returns a clear usage failure before catalog, environment, or execution work

#### Scenario: Semantic target failures are consistent

- GIVEN a syntactically valid profile or resource that the loaded catalog does not contain
- WHEN either command runs
- THEN both follow the shared catalog, environment, configuration, and planning path
- AND both render the same semantic diagnostic/report and failure exit without execution

#### Scenario: Failures and partial results match

- GIVEN catalog, required config, or environment failure, or a confirmed step failure after earlier work
- WHEN either command runs
- THEN it reports the same safe cause, ordered partial results, non-zero failure status, and non-transactional semantics

#### Scenario: Help and parity tests use injectable seams

- GIVEN CLI tests cover root and command help, all modes, syntactic input, unknown-target semantic failure, missing prerequisites, and partial execution
- WHEN tests compare command invocations
- THEN they assert equivalent plans, reports, exit statuses, and command calls without host-dependent side effects

## REMOVED Requirements
