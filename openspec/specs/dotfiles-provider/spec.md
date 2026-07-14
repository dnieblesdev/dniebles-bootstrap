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

### Requirement: Dotlink JSON reports are the only execution source of truth

For confirmed dotfile execution, the system MUST request `dotlink link --report=json MODULE...` and treat stdout as the only candidate structured outcome source. Stderr MUST NOT be parsed as report input or as a fallback.

When stdout is present, the system MUST parse and validate it regardless of `CommandStatus`. Parsing MUST first reject duplicate JSON object keys at every object level—including top-level report, entry, cause, failure, and rollback objects—before decoding the domain/wire report. `json.Decoder.DisallowUnknownFields` alone is insufficient. The parser MUST also reject malformed JSON, unknown fields, trailing documents, unsupported schema versions, missing/unknown modules, invalid outcomes, incomplete reports, and semantic contradictions. No invalid-report path may parse human-readable stdout or stderr as a compatibility fallback.

A valid JSON report with `status: failed` and non-success command status MUST preserve its validated entries, causes, and rollback details and produce a failed Bootstrap result. When command status is non-success and stdout is absent, malformed, or invalid, the system MUST produce a generic safe failed result. A report with `status: success` and non-success command status, or `status: failed` and success command status, MUST be treated as inconsistent and fail safely.

#### Scenario: Duplicate keys fail before domain decode

- GIVEN stdout contains a duplicate top-level JSON key or a duplicate key in an entry, cause, failure, or rollback object
- WHEN the report is consumed
- THEN the report is rejected as invalid
- AND execution produces a safe failed result
- AND no human-output parsing fallback is used

#### Scenario: Valid failed report survives a non-zero process exit

- GIVEN confirmed Dotlink execution exits non-zero
- AND stdout contains a valid schema v1 report with `status: failed`
- WHEN the provider consumes the command result
- THEN it preserves validated entries, safe causes, and rollback detail
- AND the aggregate Bootstrap result is failed

#### Scenario: Missing or malformed report after command failure fails safely

- GIVEN confirmed Dotlink execution exits non-zero
- AND stdout is absent, malformed, or fails validation
- WHEN the provider consumes the command result
- THEN it returns a generic safe failed result
- AND no human-output parsing fallback is used

#### Scenario: Command and report success states must agree

- GIVEN stdout contains a valid report
- AND the report status and command status disagree
- WHEN the provider consumes the result
- THEN it treats the result as inconsistent and failed safely

### Requirement: Dotlink report outcomes are rendered per entry

The system MUST render each validated per-link detail from the execution result, rather than infer every upstream outcome from a module `StepStatus`.
A successful aggregate with all unchanged entries maps to a skipped module summary; a successful aggregate with one or more changed entries and no failures maps to an installed module summary. Any failed or rolled_back entry, or an aggregate failed report, maps to a failed module summary and a non-zero confirmed apply exit.

Rendering MUST show each entry’s `changed`, `unchanged`, `failed`, or `rolled_back` outcome with source, target, safe cause, and rollback detail when supplied.

#### Scenario: Mixed changed and unchanged entries are reported truthfully

- GIVEN a valid successful JSON report includes one `changed` entry and one `unchanged` entry
- WHEN the report is rendered
- THEN the module summary is installed
- AND the output shows changed and unchanged outcomes for the respective links

#### Scenario: Failed and rolled_back entries preserve diagnostics

- GIVEN a valid failed JSON report includes `failed` or `rolled_back` entries
- WHEN the report is rendered
- THEN the module summary is failed
- AND the output shows each failed or rolled_back entry with available safe cause or rollback detail

### Requirement: TOML dotfiles catalog support

The catalog MUST support `[[dotfiles]]` entries and map them into dotfile resources.
It MUST validate dotfile entries and their dependencies using existing catalog rules.

#### Scenario: Dotfiles entries load into resources

- GIVEN a catalog contains a valid `[[dotfiles]]` entry
- WHEN the catalog is loaded
- THEN the entry is available as a dotfile resource

#### Scenario: Invalid dotfiles entries fail validation

- GIVEN a catalog contains an invalid `[[dotfiles]]` entry
- WHEN the catalog is loaded
- THEN validation fails

### Requirement: Read-only dotfiles repo and module detection

The system MUST detect dotfiles repository and module directory presence through injected filesystem seams.
Detection MUST be read-only and MUST NOT clone, apply, install, or mutate dotfiles.

#### Scenario: Present module is detected

- GIVEN injected seams report the dotfiles repo and a module directory exist
- WHEN detection runs
- THEN the module is reported present

#### Scenario: Missing module is absent without side effects

- GIVEN injected seams report the module directory is missing
- WHEN detection runs
- THEN the module is reported absent
- AND no dotfiles mutation occurs

### Requirement: CLI wiring merges present dotfiles into installation state

The `plan` command MUST merge detected present dotfile modules into existing `InstallationState.PresentResources` before planning.
The CLI MUST NOT duplicate planner semantics or own dotfiles runtime behavior.

#### Scenario: Present dotfile module reaches planning

- GIVEN catalog loading succeeds and dotfiles detection reports a present module
- WHEN `dbootstrap plan` runs
- THEN the module is added to `InstallationState.PresentResources`
- AND planning sees the module as already installed

#### Scenario: Detection is skipped on catalog failure

- GIVEN the catalog cannot be loaded
- WHEN `dbootstrap plan` runs
- THEN dotfiles detection is not attempted

### Requirement: Planner remains pure and caller-driven for dotfiles

`internal/planning` MUST remain free of dotfiles filesystem probing and ownership logic.
`BuildPlan` SHOULD keep its signature unchanged unless a test proves that dotfiles state cannot be supplied through existing caller inputs.

#### Scenario: Existing inputs carry dotfiles presence

- GIVEN the caller supplies installation state with dotfile resources present
- WHEN the plan is built
- THEN planning uses the supplied state only

#### Scenario: Signature expansion is avoided

- GIVEN dotfiles presence can be merged into existing installation state
- WHEN the slice is implemented
- THEN `BuildPlan` is not expanded solely for dotfiles

### Requirement: Dotfile module availability semantics

The system MUST treat a dotfile module as available when its local module directory exists under the configured dotfiles base path.
Availability MUST remain a presence signal only and MUST NOT imply applied, cloned, or symlinked state.

#### Scenario: Existing directory means available

- GIVEN a module directory exists under the dotfiles base path
- WHEN availability is evaluated
- THEN the module is available

#### Scenario: Presence does not imply mutation

- GIVEN a module is available
- WHEN planning completes
- THEN no apply, install, clone, or symlink action is performed

### Requirement: Local dotfiles execution core is separate from read-only detection

The dotfiles provider capability MUST keep read-only detection behavior separate from local execution behavior.
Execution-capable code MUST live under `internal/execution` and MUST NOT add mutation behavior to `internal/dotfiles`.
Local execution core MUST NOT change planning semantics or command-line behavior in this slice.

#### Scenario: Detection remains read-only

- GIVEN dotfiles presence is being checked through `internal/dotfiles`
- WHEN detection runs
- THEN no dotlink, clone, pull, submodule, fetch, or remote acquisition is attempted

#### Scenario: Execution core remains outside detection package

- GIVEN local dotfiles execution support is implemented
- WHEN package boundaries are reviewed
- THEN dotlink command construction and command-runner use are in `internal/execution`
- AND `internal/dotfiles` remains read-only

### Requirement: Local dotfiles execution requires explicit safe prerequisites

The local dotfiles execution core MUST resolve the base path from `DBOOTSTRAP_DOTFILES_DIR` when set and non-empty, MUST fail safely with no fallback when it is set empty, and MUST resolve `~/.dotfiles` only when the env var is unset.
It MUST canonicalize symlinks with `EvalSymlinks` before validation.
It MUST fail safely when the canonical base path is missing, unresolved, unsafe, not an existing directory, `/`, or the user's home directory itself.
It MUST fail safely when `bin/dotlink` is missing or resolves outside the canonical dotfiles repository.
It MUST validate module names before path joining or command construction: names MUST match `[A-Za-z0-9._-]+`, MUST NOT start with `-`, MUST NOT be empty, `.`, or `..`, and MUST NOT contain path separators, traversal segments, or absolute paths.
It MUST fail safely when a selected module directory is missing or resolves outside the canonical dotfiles repository.
It MUST NOT attempt clone, pull, submodule, fetch, or other remote acquisition.

On resolution failure, it MUST report the base-resolution source, attempted source/candidate, selected modules, and safe cause. It MUST NOT label the attempted candidate as `canonical base`. It MAY render `canonical base` only after successful canonicalization and safety validation.

#### Scenario: Base resolution failure does not claim a false canonical base

- GIVEN base resolution fails for an env or home-convention candidate
- WHEN the failure is rendered
- THEN the attempted source/candidate, selected modules, and safe cause remain visible
- AND the unresolved candidate is not labeled as canonical base

#### Scenario: Safe prerequisites allow provider execution

- GIVEN a selected dotfile module name
- AND a safe canonical local base path is available
- AND `bin/dotlink` exists under the canonical repository
- AND the selected module directory exists under the canonical repository
- WHEN local execution runs through the provider
- THEN dotlink may be requested through the injected command runner

#### Scenario: Empty env base path fails safely without fallback

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` is set to an empty value
- WHEN the provider resolves the base path
- THEN resolution fails safely
- AND no home fallback is attempted

#### Scenario: Missing base path fails safely

- GIVEN a selected dotfile module name
- AND no explicit or home-convention safe local base path is available
- WHEN local execution runs through the provider
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Missing dotlink fails safely

- GIVEN a safe canonical local base path is available
- AND `bin/dotlink` is not available under the canonical dotfiles repository
- WHEN local execution runs through the provider
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Missing module fails safely

- GIVEN a safe canonical local base path and dotlink are available
- AND a selected module directory is missing
- WHEN local execution runs through the provider
- THEN execution fails with a clear failure result or error
- AND the command runner is not called

#### Scenario: Dotfiles repo symlink resolves safely

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` or `~/.dotfiles` points through a symlink
- WHEN the resolver validates the base path
- THEN validation uses the canonical destination
- AND dotlink and selected modules must remain inside that canonical repository

### Requirement: Base resolution exposes safe typed context

The provider MUST carry base source, attempted candidate, selected modules, and a safe resolution cause as structured context. A canonical base MUST be populated only after filesystem validation succeeds; failed resolution MUST leave it empty. Filesystem failures MUST preserve identity for `errors.Is` and `errors.As` classification.

#### Scenario: Valid base has canonical identity

- GIVEN a configured base candidate resolves to an existing safe directory
- WHEN the provider validates the base
- THEN the context records its source and attempted candidate
- AND the canonical base contains the resolved filesystem identity

#### Scenario: Failed base retains attempted identity only

- GIVEN an empty, missing, unsafe, or non-directory candidate
- WHEN base resolution fails
- THEN the context records the source, attempted candidate, selected modules, and safe cause
- AND canonical base is empty

#### Scenario: Wrapped filesystem errors remain classifiable

- GIVEN a filesystem operation returns a wrapped missing or invalid-path error
- WHEN the failure is propagated
- THEN callers can classify the underlying cause with `errors.Is` or `errors.As`

### Requirement: Dotlink executable requires a validated canonical base

The provider MUST NOT construct or expose the dotlink executable path until a canonical base has been resolved and validated. Once valid, the executable context MUST be derived from that canonical base and remain associated with the same filesystem identity.

#### Scenario: Invalid base omits executable context

- GIVEN base canonicalization or validation fails
- WHEN provider prerequisites are evaluated
- THEN no dotlink executable path is derived
- AND the command runner is not called

#### Scenario: Valid base derives executable context

- GIVEN a canonical validated base exists
- WHEN provider prerequisites are evaluated
- THEN executable context is derived beneath that canonical base

### Requirement: Dotfiles installer maps selected dotfile resources to module names only

The `internal/execution` dotfiles installer MUST map a selected plan step with kind `dotfile` and name `<name>` to the single module `<name>`.
It MUST NOT derive command arguments from catalog descriptions, install metadata, dependency text, shell strings, or any other field.
It MUST reject or fail non-dotfile steps when used directly.

#### Scenario: Dotfile resource name becomes module name

- GIVEN a selected plan step for `dotfile:bash`
- WHEN the dotfiles installer handles the step
- THEN it requests module `bash` from the provider

#### Scenario: Catalog metadata is ignored for command input

- GIVEN a selected dotfile step has catalog metadata or descriptions
- WHEN the installer builds provider input
- THEN only the resource name is used as the module name

#### Scenario: Non-dotfile step is not accepted

- GIVEN a tool, package, runtime, or other non-dotfile step
- WHEN the dotfiles installer is invoked directly
- THEN it returns an unsupported or failed result
- AND no dotlink command is requested

#### Scenario: Unsafe module names fail safely

- GIVEN a selected dotfile resource name starts with `-`, is `.`, is `..`, contains a path separator, is absolute, contains traversal, or contains characters outside `[A-Za-z0-9._-]`
- WHEN local execution validates modules
- THEN execution fails safely before invoking dotlink

#### Scenario: Dotlink args are explicit

- GIVEN selected modules `bash` and `nvim`
- WHEN local execution builds the dotlink command
- THEN command args are exactly `link bash nvim` in that order

#### Scenario: Empty module list fails safely

- GIVEN no selected dotfile modules are provided
- WHEN local execution is requested
- THEN execution fails before invoking dotlink

## MODIFIED Requirements

### Requirement: Planned resources reflect installation state

Resources that match environment facts and are marked present in installation state MUST remain in plan steps and MUST be reported with `already_installed` status.
Dotfile resources supplied through installation state MUST use the same presence semantics.
Resources that are not present MUST keep existing planning semantics.
(Previously: matching resources were always marked planned or attention_required.)

#### Scenario: Present resource is already installed

- GIVEN a tool, runtime, or dotfile resource matches the environment and is present in installation state
- WHEN the plan is built
- THEN the step is included
- AND the step status is `already_installed`

#### Scenario: Absent resource keeps existing semantics

- GIVEN a matching resource is not present in installation state
- WHEN the plan is built
- THEN the step status remains planned or attention_required as before

## REMOVED Requirements

### Requirement: None

(Reason: No requirement is removed in this change.)
(Migration: None)
