# Delta for apply-command-dry-run

## MODIFIED Requirements

### Requirement: Apply execution summary is always rendered

The apply command MUST render a Summary section in default, `--dry-run`, and `--yes` modes when execution results exist. Confirmed dotfile summaries MUST use aggregate module categories: `changed` for installed, `unchanged` for skipped, and `failed` for failed aggregates. A failed aggregate with rolled-back entries MAY include a `rolled_back` breakdown, but it MUST remain failed for exit status. `not supported yet` remains reserved for `not_implemented` noop results in default and dry-run modes.

#### Scenario: Confirmed mixed report has a truthful summary

- GIVEN confirmed execution receives a valid successful report containing changed and unchanged links
- WHEN the execution report is rendered
- THEN the module summary is changed
- AND each link is rendered with its own changed or unchanged outcome

### Requirement: Apply execution mode-specific reporting

The apply command MUST render execution reporting separately from plan rendering. Default and `--dry-run` execution MUST remain noop and report `not_implemented` as `not supported yet`. Only confirmed `--yes` may run selected dotfile resources through the configured dotfiles provider.

When dotfile execution is attempted, output MUST render the aggregate module result and every available validated per-link detail. A resolution failure MUST show attempted source/candidate, selected modules, and safe cause; it MUST show `canonical base` only when canonicalization and validation succeeded.

#### Scenario: Confirmed dotfile report is rendered per link

- GIVEN a selected plan contains `dotfile:bash`
- AND `dbootstrap apply --yes --resource dotfile:bash` receives a valid Dotlink report
- WHEN execution reporting is rendered
- THEN the output shows the aggregate module result
- AND it shows each link’s outcome, source, target, and available safe cause or rollback detail

#### Scenario: Resolution failure is not mislabeled

- GIVEN a confirmed selected dotfile step fails before base validation
- WHEN execution reporting is rendered
- THEN attempted source/candidate, selected modules, and safe cause are shown
- AND no attempted candidate is labeled canonical base

### Requirement: Apply remains strictly non-mutating except confirmed eligible execution

The apply command MUST NOT perform real execution, host mutation, Dotlink invocation, clone, sparse checkout, retry, concurrency, or remote acquisition in default mode or `--dry-run` mode. Confirmed `--yes` mutation remains limited to eligible brew steps and selected dotfile resources.

#### Scenario: Default and dry-run do not execute Dotlink

- GIVEN a selected dotfile resource
- WHEN apply runs without `--yes` or with `--dry-run`
- THEN no command runner is used
- AND the result is `not_implemented`

### Requirement: Confirmed apply exits non-zero for failed eligible execution

When `apply --yes` attempts eligible execution and an eligible step has aggregate status `failed`, the CLI MUST render the execution report and return non-zero. This includes failed/rolled-back report entries, aggregate failed reports, missing prerequisites, missing/malformed/duplicate-key/inconsistent reports, and command failures. A valid failed report from non-success Dotlink execution MUST retain and render detail before the non-zero exit. Default apply and `--dry-run` remain non-mutating and MUST NOT imply real execution.

#### Scenario: Valid failed report from a failing command exits non-zero with detail

- GIVEN a selected dotfile resource is eligible under `apply --yes`
- AND Dotlink exits non-zero with a valid `status: failed` report on stdout
- WHEN apply renders the execution report
- THEN the step is reported as failed with validated entry/cause/rollback detail
- AND the CLI exits non-zero

#### Scenario: Inconsistent command/report states fail safely

- GIVEN Dotlink returns a success report with non-success status or a failed report with success status
- WHEN confirmed apply consumes the result
- THEN the step is reported as failed safely
- AND the CLI exits non-zero

### Requirement: Apply mode is explicit and safe by default

The apply command MUST treat the default mode and `--dry-run` as non-mutating and `--yes` as the only confirmed mode that may mutate eligible resources. It MUST reject `--dry-run --yes` with a usage error and no execution result.

#### Scenario: Yes is the only confirmed mutating mode

- GIVEN the user runs `dbootstrap apply --yes` for a selected dotfile step
- WHEN the command executes
- THEN only configured eligible execution may be attempted

## REMOVED Requirements

### Requirement: No apply command is introduced

(Reason: `apply` is intentionally a functional bridge; its default and dry-run paths remain safe noop modes.)
