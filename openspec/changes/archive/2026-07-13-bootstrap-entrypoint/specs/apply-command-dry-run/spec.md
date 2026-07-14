# Delta for apply-command-dry-run

## ADDED Requirements

### Requirement: Bootstrap has full apply command parity

The `bootstrap` command MUST expose the apply command's target, catalog,
safety, planning, provider, reporting, and exit contracts while requiring an
explicit `--profile` and/or repeatable `--resource`. Root help MUST list and
describe `bootstrap`, and command-specific help MUST provide its usage.
`apply` MUST remain backward-compatible and MUST NOT gain an implicit target
or changed default behavior. Both commands MUST use one shared orchestration
path.

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

## MODIFIED Requirements

### Requirement: None

(Reason: Existing apply requirements remain authoritative; this delta adds bootstrap parity without replacing apply behavior.)
(Migration: None)
