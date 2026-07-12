# Bootstrap Entrypoint Specification

## Purpose

Add a discoverable `dbootstrap bootstrap` command that is an explicit-target,
thin front door to the existing apply workflow. It MUST not create a second
installation pipeline.

## Non-goals

This change MUST NOT alter default profiles or catalogs, add a TUI, providers,
package-presence redesign, automatic escalation, downloads or shell
orchestration, retries, transactions, or rollback guarantees.

## Requirements

### Requirement: Bootstrap requires an explicit target

`bootstrap` MUST require at least one `--profile` or repeatable `--resource`.
It MUST accept both together and reject syntactic input errors before probing or
mutation: no explicit target, malformed resource syntax, unexpected
positionals, or invalid safety-mode combinations. A syntactically valid unknown
profile or resource is catalog-dependent and MUST use the same shared
catalog/environment/config/planning path, diagnostic/report, and exit behavior
as `apply`.

#### Scenario: Profile and resources select scope

- GIVEN valid profile and/or repeatable `kind:name` resources
- WHEN `dbootstrap bootstrap` runs
- THEN the selected scope is passed to the existing planner

#### Scenario: Syntactic target or mode input is invalid

- GIVEN no target, a malformed resource, an unexpected positional, or an invalid mode combination
- WHEN `bootstrap` validates arguments
- THEN it renders usage guidance and exits with a usage failure
- AND catalog, environment, detector, runner, and provider work is not started

#### Scenario: Unknown target is a semantic failure

- GIVEN a syntactically valid profile or resource that is absent from the catalog
- WHEN `bootstrap` runs
- THEN it follows the existing shared catalog, environment, configuration, and planning path
- AND it produces the same semantic diagnostic, report, and failure exit as `apply` without execution

### Requirement: Bootstrap is apply-equivalent and non-duplicative

`bootstrap` MUST use the same application orchestration as `apply`, including
flag parsing, catalog loading, environment/config/installation/dotfiles
detection, planning, provider eligibility, execution, rendering, and exit
classification. The two commands MUST produce equivalent plan, report, and
exit outcomes for identical explicit targets and modes. The implementation
MUST NOT duplicate the pipeline, detector composition, Runner, or providers.

#### Scenario: Valid invocation has parity

- GIVEN the same explicit target, catalog, and safety flags
- WHEN `bootstrap` and `apply` run
- THEN their plans, user-visible reports, and exit statuses are equivalent

#### Scenario: Apply remains compatible

- GIVEN an existing `apply` invocation
- WHEN this capability is present
- THEN `apply` retains its prior flags, safety gates, outputs, failures, and exit behavior

### Requirement: Bootstrap preserves safety modes

`bootstrap` MUST support `--catalog`, `--dry-run`, `--yes`, and `--sudo` with
the existing validation. Default and `--dry-run` modes MUST be non-mutating;
only `--yes` may execute eligible work; only `--yes --sudo` may enable APT
sudo. `--dry-run --yes` and `--sudo` without `--yes` MUST be usage failures.

#### Scenario: Modes remain safe

- GIVEN a valid explicit target
- WHEN bootstrap runs in default or dry-run mode
- THEN it renders the equivalent non-mutating result and invokes no mutating command

#### Scenario: Confirmed modes retain limits

- GIVEN `--yes` or `--yes --sudo`
- WHEN eligible Brew, Linux APT, or selected dotfile work is reached
- THEN the existing apply eligibility, provider, timeout, failure, and sudo limits apply

### Requirement: Bootstrap reports failures and partial execution honestly

Bootstrap MUST report missing catalog/config/environment prerequisites with
their existing safe causes and exit behavior. It MUST render ordered results
when confirmed execution partially completes, MUST return the existing failed
outcome when an eligible step fails, and MUST make no transaction or rollback
guarantee; recovery is rerun-oriented.

#### Scenario: Prerequisite failure is safe

- GIVEN catalog loading, required configuration, or environment detection fails
- WHEN bootstrap runs
- THEN it reports the failure, exits as apply does, and performs no execution

#### Scenario: Confirmed execution is partial

- GIVEN an earlier eligible step completes and a later one fails
- WHEN bootstrap renders the result
- THEN ordered successful and failed results remain visible
- AND the command exits non-zero without claiming atomicity or rollback

### Requirement: Help discoverability and parity are testable

Root `dbootstrap --help` MUST list `bootstrap` and concisely describe its
purpose. `bootstrap --help` MUST describe the command, explicit target
requirement, supported safety flags, and usage examples or equivalent guidance.
CLI tests MUST compare bootstrap/apply parity across default, dry-run, yes,
yes+sudo, syntactic input, unknown-target semantic failure, prerequisite
failure, and partial-failure cases using injected seams; tests MUST prove no
duplicated pipeline behavior.

#### Scenario: Root help exposes bootstrap

- GIVEN the user requests root help
- WHEN `dbootstrap --help` renders
- THEN it lists `bootstrap` with a concise description
- AND no catalog, detector, runner, or provider work is started

#### Scenario: Help guides first use

- GIVEN the user requests bootstrap help
- WHEN help renders
- THEN it documents explicit target selection and safety modes without running detection or execution
