# Delta for installation-state

## ADDED Requirements

### Requirement: Planning accepts explicit installation state

The system MUST treat installation state as caller-supplied input to planning.
Planning MUST remain pure and MUST NOT probe the host for installation state.

#### Scenario: Empty state preserves current behavior

- GIVEN planning is called with an empty installation state
- WHEN the plan is built
- THEN existing planned, attention_required, and skipped behavior is unchanged

#### Scenario: State is provided by the caller

- GIVEN the caller supplies installation state data
- WHEN the plan is built
- THEN planning uses that data without host access

### Requirement: Host-independent state detection seams

The system MUST detect tool and runtime presence through PATH lookup behind injectable seams.
Tests MUST be deterministic and MUST NOT depend on the real host environment.

#### Scenario: Tool presence is detected through injected lookup

- GIVEN a fake PATH lookup reports a tool present
- WHEN state detection runs
- THEN the tool is reported as present

#### Scenario: Tests avoid host dependence

- GIVEN a test fixture with injected lookup behavior
- WHEN state detection runs
- THEN the result depends only on the fixture

### Requirement: CLI plan detects installation state before planning

The plan command MUST detect installation state after catalog load and before BuildPlan.

#### Scenario: Detection runs before planning

- GIVEN the catalog is loaded successfully
- WHEN `dbootstrap plan` runs
- THEN installation state is detected before planning begins
- AND the detected state is available to the planner call

#### Scenario: Catalog load failure prevents detection

- GIVEN the catalog cannot be loaded
- WHEN `dbootstrap plan` runs
- THEN installation state detection is not attempted
- AND the command returns the existing catalog-load failure

### Requirement: CLI passes detected state to planning without duplicated logic

The plan command MUST pass the detected installation state to BuildPlan through the CLI composition root.
The command MUST NOT duplicate planner rules or detector rules in CLI code.

#### Scenario: Detected state is forwarded intact

- GIVEN a detector returns installation state with present resources
- WHEN `dbootstrap plan` builds the plan
- THEN BuildPlan receives that exact detected state

#### Scenario: CLI does not reimplement selection logic

- GIVEN a resource is present or absent in installation state
- WHEN `dbootstrap plan` runs
- THEN the CLI does not decide step status itself
- AND the planner remains the source of plan semantics

### Requirement: CLI tests use an injected detector seam

The plan command tests MUST be host-independent and MUST inject installation-state detection.

#### Scenario: Present-state test is deterministic

- GIVEN a test-supplied detector reports a present resource
- WHEN the plan command runs in tests
- THEN the output is deterministic
- AND the result does not depend on the host PATH

#### Scenario: Empty-state baseline is deterministic

- GIVEN a test-supplied detector returns empty installation state
- WHEN the plan command runs in tests
- THEN the baseline output remains stable

## MODIFIED Requirements

### Requirement: Confirmed Brew formula presence detection is read-only

In confirmed `apply` and `bootstrap`, the system MUST be able to query presence only for a Brew-backed package resource whose non-empty formula authority is `InstallMetadata.Package`. The query MUST use the injected command boundary with the exact executable-and-argument vector `brew list --formula <InstallMetadata.Package>`, MUST be bounded by the existing timeout mechanism, and MUST NOT mutate the host.

#### Scenario: Eligible formula uses package metadata

- **GIVEN** a confirmed plan contains a Brew-backed package with `InstallMetadata.Package` `jq`
- **AND** its resource ID and `Presence.Name` differ from `jq`
- **WHEN** presence detection runs
- **THEN** the injected runner receives exactly executable `brew` and arguments `list`, `--formula`, `jq`
- **AND** no package installation command is requested

#### Scenario: Detection is not shell-based

- **GIVEN** an eligible Brew package is queried
- **WHEN** the query is executed
- **THEN** it is represented as an executable plus argument vector
- **AND** no shell string, pipeline, or shell interpolation is used

### Requirement: Brew query results are classified conservatively

The system MUST treat only a successful, supported Brew formula query as proof of installed state. A query result MUST be classified as `installed` only when the runner reports successful completion for the exact read-only query. A completed query MUST be classified as `absent` only when the runner or Brew adapter explicitly identifies the formula as absent. Missing or unavailable `brew`, an unclassified non-zero result, runner failure, timeout, malformed command result, or unsupported metadata MUST be classified as `unknown`.

#### Scenario: Successful query proves installed

- **GIVEN** the exact Brew formula query completes successfully
- **WHEN** the result is classified
- **THEN** the package state is `installed`
- **AND** it is eligible to become `already_installed`

#### Scenario: Explicit absent result remains installable

- **GIVEN** the exact Brew formula query completes without runner error
- **AND** the result is explicitly classified as formula absent
- **WHEN** the result is classified
- **THEN** the package state is `absent`
- **AND** the package remains eligible for the existing Brew installer

#### Scenario: Operational non-zero is not absence

- **GIVEN** the exact query exits non-zero
- **AND** the result is not explicitly classified as formula absent
- **WHEN** the result is classified
- **THEN** the package state is `unknown`
- **AND** it is not reported as `already_installed` or absent

#### Scenario: Unavailable Brew is unknown

- **GIVEN** `brew` cannot be resolved or cannot be invoked
- **WHEN** a confirmed Brew package is evaluated
- **THEN** the package state is `unknown`
- **AND** no installer is invoked for that package

#### Scenario: Timeout or runner failure is unknown

- **GIVEN** the Brew query times out or the injected runner reports an execution failure
- **WHEN** the package state is evaluated
- **THEN** the package state is `unknown`
- **AND** no retry, fallback query, or installer invocation occurs for that package

#### Scenario: Metadata cannot authorize a probe

- **GIVEN** a package is not Brew-backed, has empty or unsupported install metadata, or lacks a valid `InstallMetadata.Package`
- **WHEN** confirmed execution evaluates the plan
- **THEN** no Brew query is attempted
- **AND** the package is not marked `already_installed` from Brew presence

### Requirement: Confirmed Brew presence affects execution state only after a positive result

A positively installed eligible Brew formula MUST be represented as `already_installed` for the affected plan step before installer dispatch. An unknown result MUST produce a visible attention/failure outcome for that package and MUST NOT authorize installer dispatch. Other plan steps MUST retain their existing ordered continued-execution behavior.

#### Scenario: Installed formula occupies its original position

- **GIVEN** a confirmed plan contains an eligible Brew package whose query is classified as installed
- **WHEN** the plan is executed
- **THEN** that step remains in its original order position
- **AND** its result is `already_installed`/unchanged
- **AND** no installer or mutation command is invoked for it

#### Scenario: Unknown package does not become false absence

- **GIVEN** a confirmed plan contains a Brew package whose query is unknown
- **WHEN** execution processes the plan
- **THEN** the package receives an attention/failure outcome
- **AND** it is not reported as installed or absent
- **AND** its installer is not invoked
- **AND** unrelated later steps follow existing order and continuation semantics

### Requirement: Planned resources reflect installation state

Resources that match environment facts and are marked present in installation state MUST remain in plan steps and MUST be reported with `already_installed` status.
Resources that are not present MUST keep existing planning semantics.
(Previously: matching resources were always marked planned or attention_required.)

#### Scenario: Present resource is already installed

- GIVEN a tool or runtime resource matches the environment and is present in installation state
- WHEN the plan is built
- THEN the step is included
- AND the step status is `already_installed`

#### Scenario: Absent resource keeps existing semantics

- GIVEN a matching resource is not present in installation state
- WHEN the plan is built
- THEN the step status remains planned or attention_required as before

### Requirement: Detector failures remain future scope

The plan command MUST use the current installation-state detector contract, which returns installation state without an error value.
Detector failure diagnostics are deferred until the detector contract can represent failures explicitly.

#### Scenario: Current detector contract is used unchanged

- GIVEN the installation-state detector returns installation state without an error
- WHEN `dbootstrap plan` runs
- THEN the command passes that state to planning
- AND it does not invent detector-failure diagnostics

#### Scenario: Future detector failures are not implemented in this slice

- GIVEN future detector contracts may report errors
- WHEN this slice is applied
- THEN no detector error branch is added
- AND no detector, planner, or renderer contract is expanded for failure diagnostics

### Requirement: Status precedence is deterministic

`already_installed` MUST take precedence after environment matching succeeds, including when the resource would otherwise require attention for missing config.
Resources excluded by environment mismatch MUST remain skipped and MUST NOT become `already_installed`.
(Previously: no installation-state precedence existed.)

#### Scenario: Present resource beats missing config

- GIVEN a matching resource is present in installation state and also lacks required config
- WHEN the plan is built
- THEN the step status is `already_installed`

#### Scenario: Environment mismatch still skips

- GIVEN a resource does not match the environment facts
- WHEN the plan is built
- THEN the resource remains skipped
- AND installation state does not change that outcome

### Requirement: Presence detection uses the configured command name

For tool and runtime resources whose presence detector is command-based, the system MUST probe `Resource.Presence.Name` when that value is configured. It MUST NOT substitute the resource ID, package name, or another catalog field when a configured presence name exists.

#### Scenario: Configured presence name differs from resource ID

- GIVEN a tool or runtime has resource ID `editor` and `Presence.Name` `vim`
- AND the injected PATH lookup reports `vim` present but does not report `editor` present
- WHEN installation-state detection runs
- THEN the resource is reported present

#### Scenario: Missing presence name is not guessed

- GIVEN a tool or runtime has no configured command presence name
- WHEN installation-state detection runs
- THEN the detector preserves its existing unsupported/absent behavior
- AND it does not infer a command name from package metadata or configuration

### Requirement: Conservative injectable APT detection

On confirmed Linux, the system MUST use injected, read-only `dpkg-query --show --showformat=${Status} <package>`. A well-formed three-field status MUST be installed iff its error field is `ok` and its package-status field is `installed`, including `hold ok installed`. A known well-formed non-installed status MUST be absent only when it is a valid definitive non-installed state; partial states such as `unpacked` or `half-configured` MUST NOT be installed and MUST dispatch normally. The exact absent signature is exit 1, stderr `dpkg-query: no packages found matching <package>`, and no contradictory stdout. Every other non-zero, missing-command, timeout, runner-error, empty, malformed, or ambiguous result MUST be unknown. No `sudo`, `apt-get`, fallback, or retry is permitted.

#### Scenario: Held installed status skips
- GIVEN an eligible package on confirmed Linux
- WHEN the injected query returns `hold ok installed`
- THEN the result is installed and confirmed execution skips the installer

#### Scenario: Partial status is not installed
- GIVEN an eligible package on confirmed Linux
- WHEN the injected query returns `install ok unpacked` or `install ok half-configured`
- THEN the result is not installed and the normal APT installer remains eligible

#### Scenario: Definitive absence dispatches
- GIVEN an eligible package on confirmed Linux
- WHEN the query returns a valid definitive non-installed status or the exact provider not-found signature
- THEN the result is absent and the normal APT installer remains eligible

#### Scenario: Ambiguous evidence is unknown
- GIVEN the query has any other failure or empty, malformed, or ambiguous output
- WHEN detection runs
- THEN the result is unknown and no presence is reported

### Requirement: Idempotency detection is limited to reliable command or Brew formula presence

The system MUST use detected presence for apply idempotency only for tool and runtime resources whose command presence was reliably detected, or for eligible Brew-backed package resources whose formula presence was positively proven by the read-only query defined by this change. Presence detection MUST NOT perform package-version, configuration, or dotfile-link convergence checks. Brew formula presence MUST NOT be treated as evidence of a required version, executable health, PATH/link/configuration correctness, or dotfile convergence.
(Previously: idempotency detection was limited to reliable command presence for tools and runtimes.)

#### Scenario: Reliable command presence remains sufficient

- GIVEN a supported tool or runtime command is found through the injected PATH lookup
- WHEN planning and confirmed execution run
- THEN the resource remains governed by the existing command-presence idempotency behavior

#### Scenario: Positive Brew formula presence enables idempotency

- GIVEN a confirmed eligible Brew-backed package query is classified as installed
- WHEN the package step reaches execution
- THEN the step is treated as already installed and unchanged
- AND no installer mutation is attempted

#### Scenario: Broader reconciliation is not attempted

- GIVEN a Brew-backed package is selected for detection
- WHEN detection runs
- THEN no version, executable, configuration, dotfile, retry, fallback, or bootstrap query is attempted

## REMOVED Requirements

### Requirement: None

(Reason: No requirement is removed in this change.)
(Migration: None)
