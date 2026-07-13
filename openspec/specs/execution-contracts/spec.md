# Delta for execution-contracts

## ADDED Requirements

### Requirement: Execution contracts are separate from planning

`internal/execution` MUST define its own Installer, Runner, DotfilesProvider, status, result, and report types.
Execution types MUST NOT reuse planning statuses or mutate planning production code.

#### Scenario: Execution types remain distinct

- GIVEN planning types already exist
- WHEN execution contracts are defined
- THEN execution uses separate types and status vocabulary

#### Scenario: Planning production stays unchanged

- GIVEN the change is implemented
- WHEN planning code is inspected
- THEN planning production behavior remains unchanged

### Requirement: Conservative confirmed-Linux APT guards

Execution MUST preserve order. A well-formed status MUST skip installation iff its error field is `ok` and package-status field is `installed`, including `hold ok installed`. A valid definitive non-installed status, or the exact provider-specific not-found signature (exit 1, stderr `dpkg-query: no packages found matching <package>`, and no contradictory stdout), MUST dispatch the normal APT installer. Partial states such as `unpacked` or `half-configured` MUST NOT skip and MUST dispatch. Unknown MUST fail without installer, `apt-get`, or `sudo`. Detection MUST be injected, read-only, and free of retries or fallbacks.

#### Scenario: Installed skips; absent dispatches
- GIVEN confirmed execution has an APT step classified installed or absent
- WHEN the runner processes it
- THEN installed is unchanged and absent dispatches the normal installer
- AND the step remains in its original position

#### Scenario: Held installed skips
- GIVEN confirmed execution has an APT step with status `hold ok installed`
- WHEN the runner processes it
- THEN it reports unchanged and makes no installer or command call

#### Scenario: Partial state does not skip
- GIVEN an APT step has status `install ok unpacked` or `install ok half-configured`
- WHEN the runner processes it
- THEN it dispatches the normal APT installer

#### Scenario: Not-found dispatches
- GIVEN the query exits 1 with matching `no packages found matching <package>` stderr and no contradictory stdout
- WHEN the runner processes the step
- THEN it dispatches the normal APT installer without retry or fallback

#### Scenario: Unknown fails safely
- GIVEN the APT result is unknown
- WHEN the runner processes the plan
- THEN it reports failure and makes no installer, `apt-get`, or `sudo` call

### Requirement: Noop execution is safe and non-mutating

Noop installers, runners, and providers MUST return a `not_implemented` execution result or report for unsupported work.
They MUST NOT invoke real commands, apply changes, or mutate the host.

#### Scenario: Unsupported action returns not_implemented

- GIVEN a noop execution dependency receives a request
- WHEN the request is handled
- THEN the result or report status is `not_implemented`

#### Scenario: No mutation occurs in noop mode

- GIVEN a noop execution dependency is used
- WHEN it runs
- THEN no apply, install, clone, or filesystem mutation occurs

### Requirement: Runner dispatches plan steps sequentially by kind

The Runner MUST consume a planning Plan sequentially and dispatch each step to the installer for that step's resource kind.
The Runner MUST preserve plan order and MUST return an execution report for all processed steps.

#### Scenario: Steps run in plan order

- GIVEN a plan contains multiple steps with different kinds
- WHEN the Runner executes the plan
- THEN steps are dispatched in the same order they appear in the plan

#### Scenario: Kind selects the installer

- GIVEN a plan step targets a specific resource kind
- WHEN the Runner dispatches the step
- THEN the matching installer handles that step

### Requirement: DotfilesProvider is a high-level execution boundary

DotfilesProvider MUST expose high-level execution operations for dotfiles workflow support.
It MUST remain separate from the read-only dotfiles detector and MUST NOT own planning logic.

#### Scenario: Provider is execution-only

- GIVEN a dotfiles operation is requested
- WHEN the provider is used
- THEN it serves execution-layer behavior only

#### Scenario: Provider does not own planning

- GIVEN planning is being built
- WHEN dotfiles execution support is considered
- THEN DotfilesProvider does not change planning behavior

### Requirement: Execution contracts remain non-mutating for apply

`internal/execution` MUST remain a safe, non-mutating boundary used by `apply`.
The command MUST use noop execution contracts by default and in `--dry-run`, MUST allow real execution only for confirmed brew-backed tool/package steps with their existing cross-platform eligibility, confirmed Linux APT-backed tool/package steps, and confirmed selected dotfile steps, MUST surface Homebrew bootstrap reporting as advisory data only, and MUST NOT introduce real execution outside that narrow path. APT MUST use direct `apt-get install -y -- <package>` for `--yes` and explicit `sudo apt-get install -y -- <package>` for `--yes --sudo`, with a ten-minute `CommandRequest.Timeout`; APT metadata MUST be trimmed, non-empty, and not start with `-`.
Dotfiles execution MUST remain dormant unless the confirmed apply composition root explicitly wires the provider with configured seams.

#### Scenario: Apply uses noop execution contracts by default

- GIVEN the `apply` command runs without `--yes`
- WHEN execution is dispatched
- THEN only noop results are produced
- AND the dotfiles command runner is not used

#### Scenario: Apply dry-run uses noop execution contracts

- GIVEN the `apply` command runs with `--dry-run`
- WHEN execution is dispatched
- THEN only noop results are produced
- AND the dotfiles command runner is not used

#### Scenario: Confirmed brew steps may execute

- GIVEN `apply --yes` and a brew-backed tool/package step
- WHEN execution is dispatched
- THEN real brew execution is allowed for that step only

#### Scenario: Confirmed selected dotfile steps may execute

- GIVEN `apply --yes` and a selected dotfile plan step
- WHEN execution is dispatched
- THEN the CLI may compose the existing dotfiles provider for that step
- AND dotlink execution is requested only through the injected command runner

#### Scenario: Side effects remain absent outside confirmed eligible steps

- GIVEN execution contracts are present
- WHEN `apply` is reviewed end-to-end
- THEN no real execution or production mutation occurs outside confirmed brew-backed steps with their existing eligibility, confirmed Linux APT-backed steps, and confirmed selected dotfile steps

#### Scenario: Bootstrap data stays advisory

- GIVEN Homebrew bootstrap need data is attached
- WHEN execution contracts report it
- THEN the data remains non-mutating and reviewable

#### Scenario: Core provider is dormant until composed

- GIVEN the dotfiles provider and installer exist in `internal/execution`
- WHEN no caller composes them into the confirmed apply runner
- THEN no dotlink execution is possible through the CLI
- AND existing noop execution behavior is unchanged

### Requirement: CLI composition uses injectable execution seams

The CLI apply composition root MUST allow tests to inject a fake dotfiles `CommandRunner`, Linux facts, `apt-get` and `sudo` availability seams, and fake base-resolution/filesystem/prerequisite dependencies.
Production confirmed apply MAY use the real local command runner and resolver, but tests MUST NOT require or invoke real `dotlink`, host dotfiles state, clone, pull, submodule, fetch, or remote acquisition.

#### Scenario: Tests inject fake dotfiles execution dependencies

- GIVEN apply command tests exercise `apply --yes --resource dotfile:bash`
- WHEN the CLI composes execution dependencies
- THEN the test can provide a fake command runner and fake dotfiles prerequisite seams
- AND no real external command is executed

#### Scenario: Production composition remains confirmed-only

- GIVEN production apply dependencies are used
- WHEN the user runs default apply or `apply --dry-run`
- THEN the dotfiles provider is not composed with a mutating runner for execution

#### Scenario: Acquisition commands are absent from composition

- GIVEN confirmed dotfiles execution is composed
- WHEN command requests are inspected
- THEN no clone, pull, submodule, fetch, remote URL, sparse checkout, or apt command is requested

### Requirement: Dotfiles execution core uses an injectable command runner

The dotfiles execution core in `internal/execution` MUST route dotlink invocation through an injected `CommandRunner` seam.
The core MUST NOT invoke `exec.Command`, a shell, or any real external command directly.
Tests MUST be able to substitute a fake runner and MUST NOT require real external commands.
Dotlink command requests MUST include a bounded timeout, and timeout results MUST become failed dotfile execution results.

#### Scenario: Tests use a fake runner

- GIVEN the dotfiles execution provider is under test
- WHEN a fake command runner is injected
- THEN the test can assert the requested executable, arguments, working directory, and timeout
- AND no real command is executed

#### Scenario: Direct process execution is absent

- GIVEN the dotfiles execution core source is reviewed
- WHEN command invocation behavior is inspected
- THEN the core does not call `exec.Command` or a shell directly
- AND all dotlink execution goes through `CommandRunner`

#### Scenario: Dotlink timeout fails safely

- GIVEN the provider invokes dotlink through the command runner
- WHEN the command result indicates timeout
- THEN the dotfile execution result is failed
- AND no retry, fallback acquisition, or second command is attempted

### Requirement: Dotfiles execution core validates local prerequisites only

The dotfiles execution core MUST resolve the local base path from `DBOOTSTRAP_DOTFILES_DIR` when set and non-empty, MUST fail safely with no fallback when it is set empty, and MUST resolve `~/.dotfiles` only when the env var is unset.
It MUST canonicalize the selected path with `EvalSymlinks` before validating or constructing command paths.
It MUST fail safely when the canonical base path is missing, unresolved, relative, not an existing directory, `/`, the home directory itself, or otherwise unsafe.
It MUST validate that `bin/dotlink` and selected module directories exist under the canonical repository.
It MUST validate module names before path joining or command construction: names MUST match `[A-Za-z0-9._-]+`, MUST NOT start with `-`, MUST NOT be empty, `.`, or `..`, and MUST NOT contain path separators, traversal segments, or absolute paths.
It MUST NOT silently fallback to another path after the selected source fails validation.
It MUST NOT attempt clone, pull, submodule, fetch, or other remote acquisition.

#### Scenario: Environment path is canonicalized and validated

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` is set to a dotfiles path
- WHEN the base path resolver runs
- THEN it resolves symlinks with `EvalSymlinks`
- AND validation uses the canonical directory
- AND no home fallback is attempted if validation fails

#### Scenario: Empty environment path does not fallback

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` is set to an empty value
- WHEN the base path resolver runs
- THEN resolution fails safely
- AND `$HOME/.dotfiles` is not used as fallback

#### Scenario: Home convention is used when environment is unset

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` is unset
- AND the user's home directory is known
- WHEN the base path resolver runs
- THEN it resolves exactly `$HOME/.dotfiles`
- AND validation uses the canonical directory

#### Scenario: Unsafe base path fails safely

- GIVEN the selected base path resolves to `/`, the home directory itself, a missing path, or a non-directory
- WHEN the provider validates prerequisites
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Repository shape is required before command execution

- GIVEN a safe canonical base path exists
- BUT `bin/dotlink` or a selected module directory is missing or resolves outside the canonical repository
- WHEN the provider validates prerequisites
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Unsafe module names fail safely

- GIVEN a selected module name is empty, starts with `-`, contains a path separator, is `.`, is `..`, is absolute, contains traversal, or contains characters outside `[A-Za-z0-9._-]`
- WHEN the provider validates prerequisites
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Remote acquisition is not attempted

- GIVEN dotfiles execution is invoked
- WHEN prerequisites are missing or command execution fails
- THEN no clone, pull, submodule, fetch, remote URL, or other acquisition command is requested

### Requirement: Module summaries and per-link outcomes are distinct

The execution layer MUST represent default and `--dry-run` noop work as `not_implemented`.
For confirmed dotfile execution, `StepResult` (or an equivalent execution-owned result) MUST contain both an aggregate module status and ordered per-link details. Per-link detail MUST preserve the Dotlink outcome (`changed`, `unchanged`, `failed`, or `rolled_back`), source, target, safe cause when supplied, and rollback detail when supplied. Legacy `StepStatus` MUST NOT be required to represent every upstream entry outcome.

The aggregate module status MUST be:
- `skipped` when all reported entries are `unchanged` and the aggregate report is successful;
- `installed` when one or more entries are `changed`, no entry failed or rolled back, and the aggregate report is successful;
- `failed` when any entry is `failed` or `rolled_back`, when the aggregate report is failed, or when report/command reconciliation fails.

#### Scenario: Mixed successful entries retain their own outcomes

- GIVEN a confirmed dotfile report has one `changed` entry and one `unchanged` entry
- WHEN the execution layer records the result
- THEN the module aggregate status is `installed`
- AND the ordered per-link details retain one `changed` and one `unchanged` outcome

#### Scenario: Failed aggregate does not erase entry detail

- GIVEN a valid failed report contains changed, failed, or rolled_back entries
- WHEN the execution layer records the result
- THEN the module aggregate status is `failed`
- AND each available per-link outcome, cause, and rollback detail remains available

#### Scenario: Noop apply stays not_implemented

- GIVEN apply runs without confirmed execution
- WHEN the execution layer returns its report
- THEN the status is `not_implemented`

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
