# Delta for execution-contracts

## ADDED Requirements

### Requirement: Confirmed Brew package presence is checked before installer dispatch

Confirmed `apply` and `bootstrap` MUST evaluate eligible Brew-backed package presence before dispatching that package's installer. A positively installed formula MUST produce an unchanged result with explicit no-mutation wording. An explicitly absent formula MUST retain eligibility for the existing installer. An unknown result MUST produce an attention/failure outcome and MUST suppress the installer for that package.

#### Scenario: Installed package is skipped in order

- GIVEN a confirmed plan contains a Brew-backed package whose formula query is positively installed
- WHEN the execution runner processes the plan
- THEN the package result remains in its original plan position
- AND the result is unchanged/already installed
- AND the report explicitly states that no mutation was attempted
- AND the package installer and mutation runner are not called

#### Scenario: Absent package remains eligible

- GIVEN a confirmed plan contains a Brew-backed package whose query is explicitly classified as formula absent
- WHEN the execution runner processes the plan
- THEN the existing Brew installer remains eligible
- AND the package is not reported as already installed solely because detection ran

#### Scenario: Unknown package suppresses mutation

- GIVEN a confirmed plan contains a Brew-backed package whose query is unavailable, failed, timed out, non-zero but unclassified, malformed, or unsupported
- WHEN the execution runner processes the plan
- THEN the package result is visibly attention/failure
- AND the package installer is not called
- AND no retry or fallback query is attempted

### Requirement: Brew presence handling preserves mixed-plan execution

Brew presence detection MUST change only the eligible Brew package step it evaluates. Confirmed `apply` and `bootstrap` MUST preserve plan order, existing outcomes for tools, runtimes, dotfiles, unsupported resources, and unrelated failures, and existing continued-execution behavior for later steps.

#### Scenario: Mixed plan remains ordered

- GIVEN a confirmed plan contains an installed Brew package, an absent Brew package, and an unrelated step in that order
- WHEN execution processes the plan
- THEN results are reported in the same order
- AND the installed package is unchanged without mutation
- AND the absent package follows existing installer behavior
- AND the unrelated step retains its existing semantics

#### Scenario: Bootstrap uses the same conservative guard

- GIVEN `bootstrap` runs in confirmed mode with an eligible Brew package
- WHEN the package query is positively installed or unknown
- THEN bootstrap applies the same unchanged/no-mutation or attention/no-install outcome as `apply`
- AND bootstrap does not acquire Brew or perform fallback detection

## MODIFIED Requirements

### Requirement: Confirmed execution honors already-installed plan steps

The apply execution boundary MUST treat a plan step with status `already_installed` as unchanged and MUST NOT dispatch an installer or invoke its command runner. This guard MUST be based only on the planning status for that step; it MUST NOT be inferred from prior results, unsupported status, missing configuration, package versions, or dotfile module presence. For an eligible Brew-backed package, confirmed-mode presence detection MAY establish that planning status only after a successful read-only query using `InstallMetadata.Package`.
(Previously: `already_installed` was honored only when supplied by existing planning state, without a confirmed Brew package-presence source.)

#### Scenario: Confirmed present tool remains undispatched

- GIVEN a confirmed `apply --yes` plan contains an eligible tool or runtime step with status `already_installed`
- WHEN execution processes the plan
- THEN the step produces an unchanged result
- AND the installer and command runner are not called for that step
- AND the report states that no mutation was attempted

#### Scenario: Confirmed present Brew package is undispatched

- GIVEN a confirmed `apply --yes` or `bootstrap` plan contains an eligible Brew package
- AND its exact formula query succeeds
- WHEN execution reaches the package step
- THEN the step is marked `already_installed` before installer dispatch
- AND the step produces an unchanged result in plan order
- AND no installer or mutation command is invoked

#### Scenario: Uncertain Brew package is not undispatched as installed

- GIVEN a confirmed eligible Brew package query is unknown
- WHEN execution reaches the package step
- THEN the step is not marked `already_installed`
- AND its installer is not dispatched
- AND the package result communicates attention/failure

### Requirement: No-op and dry-run modes remain non-mutating

Default and `--dry-run` modes MUST remain non-probing and non-mutating for Brew package presence. They MUST NOT resolve or execute `brew`, invoke a Brew presence lookup, or use a detected package state to claim that a confirmed mutation was skipped. Their existing noop/dry-run result semantics MUST remain unchanged.
(Previously: no-op and dry-run modes were non-mutating, but the requirement did not explicitly prohibit Brew presence probing.)

#### Scenario: Default mode does not probe Brew

- GIVEN default `apply` or `bootstrap` selects a Brew-backed package
- WHEN execution runs
- THEN no Brew lookup or Brew command runner call occurs
- AND no host mutation occurs

#### Scenario: Dry-run does not probe Brew

- GIVEN `apply --dry-run` or `bootstrap --dry-run` selects a Brew-backed package
- WHEN execution runs
- THEN no Brew lookup or Brew command runner call occurs
- AND the existing dry-run/non-mutating result is preserved

## REMOVED Requirements

### Requirement: None

(Reason: No requirement is removed in this change.)
(Migration: None)
