# Delta for apply-command-dry-run

## ADDED Requirements

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

## MODIFIED Requirements

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

## REMOVED Requirements

### Requirement: None

(Reason: No requirement is removed in this change.)
(Migration: None)
