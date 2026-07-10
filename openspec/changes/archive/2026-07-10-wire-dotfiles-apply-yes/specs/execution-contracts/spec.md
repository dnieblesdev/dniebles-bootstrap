# Delta for execution-contracts

## ADDED Requirements

### Requirement: CLI composition uses injectable execution seams

The CLI apply composition root MUST allow tests to inject a fake dotfiles `CommandRunner` and fake base-resolution/filesystem dependencies.
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

## MODIFIED Requirements

### Requirement: Execution contracts remain non-mutating for apply

`internal/execution` MUST remain a safe boundary used by `apply`.
The command MUST use noop execution contracts by default and in `--dry-run`, MUST allow real execution only for confirmed brew-backed tool/package steps and confirmed selected dotfile steps, MUST surface Homebrew bootstrap reporting as advisory data only, and MUST NOT introduce real execution outside those narrow paths.
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
- THEN no real execution or production mutation occurs outside confirmed brew-backed steps and confirmed selected dotfile steps

#### Scenario: Bootstrap data stays advisory

- GIVEN Homebrew bootstrap need data is attached
- WHEN execution contracts report it
- THEN the data remains non-mutating and reviewable

#### Scenario: Core provider remains dormant unless composed

- GIVEN the dotfiles provider and installer exist in `internal/execution`
- WHEN no caller composes them into the confirmed apply runner
- THEN no dotlink execution is possible through the CLI
- AND existing noop execution behavior is unchanged


## REMOVED Requirements

### Requirement: None

(Reason: This change composes the dormant dotfiles provider for one confirmed apply path without removing existing execution contracts.)
(Migration: None.)
