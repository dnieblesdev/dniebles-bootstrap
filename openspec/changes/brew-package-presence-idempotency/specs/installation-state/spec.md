# Delta for installation-state

## ADDED Requirements

### Requirement: Confirmed Brew formula presence detection is read-only

In confirmed `apply` and `bootstrap`, the system MUST be able to query presence only for a Brew-backed package resource whose non-empty formula authority is `InstallMetadata.Package`. The query MUST use the injected command boundary with the exact executable-and-argument vector `brew list --formula <InstallMetadata.Package>`, MUST be bounded by the existing timeout mechanism, and MUST NOT mutate the host.

#### Scenario: Eligible formula uses package metadata

- GIVEN a confirmed plan contains a Brew-backed package with `InstallMetadata.Package` `jq`
- AND its resource ID and `Presence.Name` differ from `jq`
- WHEN presence detection runs
- THEN the injected runner receives exactly executable `brew` and arguments `list`, `--formula`, `jq`
- AND no package installation command is requested

#### Scenario: Detection is not shell-based

- GIVEN an eligible Brew package is queried
- WHEN the query is executed
- THEN it is represented as an executable plus argument vector
- AND no shell string, pipeline, or shell interpolation is used

### Requirement: Brew query results are classified conservatively

The system MUST treat only a successful, supported Brew formula query as proof of installed state. A query result MUST be classified as `installed` only when the runner reports successful completion for the exact read-only query. A completed query MUST be classified as `absent` only when the runner or Brew adapter explicitly identifies the formula as absent. Missing or unavailable `brew`, an unclassified non-zero result, runner failure, timeout, malformed command result, or unsupported metadata MUST be classified as `unknown`.

#### Scenario: Successful query proves installed

- GIVEN the exact Brew formula query completes successfully
- WHEN the result is classified
- THEN the package state is `installed`
- AND it is eligible to become `already_installed`

#### Scenario: Explicit absent result remains installable

- GIVEN the exact Brew formula query completes without runner error
- AND the result is explicitly classified as formula absent
- WHEN the result is classified
- THEN the package state is `absent`
- AND the package remains eligible for the existing Brew installer

#### Scenario: Operational non-zero is not absence

- GIVEN the exact query exits non-zero
- AND the result is not explicitly classified as formula absent
- WHEN the result is classified
- THEN the package state is `unknown`
- AND it is not reported as `already_installed` or absent

#### Scenario: Unavailable Brew is unknown

- GIVEN `brew` cannot be resolved or cannot be invoked
- WHEN a confirmed Brew package is evaluated
- THEN the package state is `unknown`
- AND no installer is invoked for that package

#### Scenario: Timeout or runner failure is unknown

- GIVEN the Brew query times out or the injected runner reports an execution failure
- WHEN the package state is evaluated
- THEN the package state is `unknown`
- AND no retry, fallback query, or installer invocation occurs for that package

#### Scenario: Metadata cannot authorize a probe

- GIVEN a package is not Brew-backed, has empty or unsupported install metadata, or lacks a valid `InstallMetadata.Package`
- WHEN confirmed execution evaluates the plan
- THEN no Brew query is attempted
- AND the package is not marked `already_installed` from Brew presence

### Requirement: Confirmed Brew presence affects execution state only after a positive result

A positively installed eligible Brew formula MUST be represented as `already_installed` for the affected plan step before installer dispatch. An unknown result MUST produce a visible attention/failure outcome for that package and MUST NOT authorize installer dispatch. Other plan steps MUST retain their existing ordered continued-execution behavior.

#### Scenario: Installed formula occupies its original position

- GIVEN a confirmed plan contains an eligible Brew package whose query is classified as installed
- WHEN the plan is executed
- THEN that step remains in its original order position
- AND its result is `already_installed`/unchanged
- AND no installer or mutation command is invoked for it

#### Scenario: Unknown package does not become false absence

- GIVEN a confirmed plan contains a Brew package whose query is unknown
- WHEN execution processes the plan
- THEN the package receives an attention/failure outcome
- AND it is not reported as installed or absent
- AND its installer is not invoked
- AND unrelated later steps follow existing order and continuation semantics

## MODIFIED Requirements

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
