# Design: Consume Dotlink Link Report

## Context and decision

Bootstrap consumes Dotlink JSON v1 as the sole confirmed-execution outcome contract. Dotlink owns link creation and rollback; Bootstrap only validates, translates, renders, and selects the confirmed apply exit status. No remote acquisition, repair, retry, shell parsing, or mutation outside the existing confirmed `apply --yes` route is added.

The provider treats command status and stdout report status as separate signals. Present stdout is always validated before final reconciliation so a valid failure report is not discarded merely because Dotlink exited non-zero.

## Upstream contract consumed

`dotlink link --report=json MODULE...` writes one JSON report to stdout and human diagnostics to stderr. The v1 report contains `schema_version`, ordered `modules`, aggregate `status` (`success` or `failed`), `entries`, nullable `failure`, and `rollback`. Entries contain module, source, target, outcome (`changed`, `unchanged`, `failed`, or `rolled_back`), and optional cause; failure and rollback include their documented nested objects/fields.

Stdout is JSON-only when supplied. Stderr is non-contractual and is never parsed. A failure before an entry can still be represented by a valid failed report with selected modules and failure context.

## Data flow

```text
confirmed apply --yes, selected dotfile plan step
  -> DotfilesInstaller.Install(step)
  -> LocalDotfilesProvider.RunDotlink(ctx, []string{step.Ref.Name})
     -> resolve/validate local base and module
     -> CommandRunner.RunCommand(dotlink, [link, --report=json, module])
     -> if stdout present: duplicate-key scan -> strict typed decode -> semantic validation
     -> reconcile report aggregate status with CommandStatus
  -> typed report or typed safe failure
  -> one aggregate StepResult plus ordered per-link execution details
  -> renderer displays aggregate and every available per-link detail
  -> confirmed failure check returns non-zero for aggregate failed
```

Production preserves the one-plan-step/one-module invariant, while the parser accepts an ordered selected-module list for unit coverage. It rejects duplicate requested modules and a report `modules` list that differs in value, count, or order. Entries and `failure.module` must not name an unselected module.

## Strict JSON boundary

`internal/execution/dotlink_report.go` owns the parse-and-validate boundary:

```go
ParseDotlinkLinkReport(stdout []byte, selected []string) (DotlinkLinkReport, error)
```

It never receives stderr. The boundary has two deliberate stages:

1. **Duplicate-key structural scan.** Use a token-based recursive JSON object/array scanner over the original bytes. For every `{...}`, maintain a per-object key set and reject a repeated key before values are decoded. Recursion covers the root report and all nested objects, including entries, causes, failure, and rollback objects. The scan also validates JSON token structure and one top-level JSON value.
2. **Strict typed decode and semantic validation.** Re-read the same bytes through `json.Decoder` with `DisallowUnknownFields`, decode the wire struct, and require EOF. This stage enforces wire types and unknown-field rejection; it is not relied on for duplicate-key detection. Convert only the validated wire value to domain values.

The parser rejects empty stdout, malformed JSON, duplicate keys, unknown fields, trailing documents, unsupported schema, invalid statuses/outcomes, missing or reordered modules, unknown modules, duplicate/ambiguous entries, missing required source/target/cause data, and report contradictions. It returns `ErrInvalidDotlinkReport` (wrapped with non-sensitive context) so callers can safely distinguish invalid reports from runner/prerequisite errors without exposing raw stdout.

Semantic rules include:

- `schema_version == 1`; status is exactly `success` or `failed`; report modules exactly equal selected modules.
- Each entry names a selected module, has non-empty source/target, has a known outcome, and has a unique `(module, source, target)` identity.
- Success reports have no failure object and no failed/rolled-back entries; their rollback object is non-attempted, non-completed, and has empty removal lists.
- Failed reports have a non-nil failure with safe non-empty cause code/message. Failed/rolled-back entries have safe non-empty causes; rolled-back entries require `rollback.attempted == true`. Rollback `completed` cannot be true unless attempted.
- Each selected module has at least one entry, except an explicit failed aggregate may identify its module solely through `failure.module`. No report may leave a selected module terminal state indeterminate.

## Command/report reconciliation

`LocalDotfilesProvider.RunDotlink` validates local prerequisites, builds `[]string{"link", "--report=json", module...}`, and executes once through `CommandRunner`.

After execution:

| Command status | stdout/report | Result |
|---|---|---|
| success | valid `success` report | return validated report |
| non-success | valid `failed` report | return validated report with command failure metadata; preserve entries, causes, and rollback |
| non-success | stdout absent, malformed, duplicate-keyed, or otherwise invalid | return a generic safe `ErrDotlinkCommandFailed`; no report detail and no fallback |
| success | valid `failed` report | return safe inconsistency failure |
| non-success | valid `success` report | return safe inconsistency failure |
| success | stdout absent or invalid | return safe report-consumption failure |

A valid report is authoritative for link detail only when its aggregate status coheres with command status. The provider never infers outcomes from exit status, stderr, or human-readable output. Generic failures contain only safe command/report classification, not raw stdout/stderr.

## Execution model and translation

Keep `StepStatus` as the aggregate module outcome for compatibility with existing execution code. Add execution-owned detail to `StepResult`, for example:

- `DotfileLinks []DotfileLinkResult`, ordered as validated report entries;
- `DotfileLinkResult` with `Module`, `Source`, `Target`, `Outcome` enum (`changed`, `unchanged`, `failed`, `rolled_back`), optional safe cause, and rollback detail/reference;
- optional aggregate safe failure and base diagnostic context.

The detail enum is not `StepStatus`; it preserves upstream entry outcomes even for a mixed module. Ordinary installers leave it zero-valued.

For the selected module:

| Valid report condition | Aggregate `StepStatus` | Per-link detail |
|---|---|---|
| aggregate success; all entries unchanged | `skipped` | each `unchanged` |
| aggregate success; one or more changed; no failed/rolled_back | `installed` | each changed/unchanged outcome |
| aggregate failed; or any failed/rolled_back entry | `failed` | retain all available entries, causes, and rollback data |
| prerequisite, command-without-valid-report, parser, validation, or reconciliation error | `failed` | no inferred links; include only safe failure/base context |

`hasFailedExecutionResult` continues to use the aggregate status. Thus every failed/rolled-back aggregate is rendered before confirmed apply exits non-zero.

## Base-resolution diagnostic context

Introduce `DotfilesBaseDiagnostic` populated before resolution with `Source`, `AttemptedCandidate`, `SelectedModules`, and safe `Cause`. A typed resolution error carries it through provider and installer.

`AttemptedCandidate` is the raw env candidate or derived `~/.dotfiles` candidate. `CanonicalPath` is populated and labeled `canonical base` only after `EvalSymlinks` and all safety checks succeed. On resolution failure, rendering must say `attempted candidate`, selected modules, source, and cause; it must never label the unresolved candidate canonical. Post-resolution failures may render the validated canonical base alongside the same selected-module context.

## Rendering and safety behavior

`cmd/dbootstrap/render.go` renders the module aggregate first, then each `DotfileLinkResult` with its truthful outcome and source-to-target detail. It renders safe causes and rollback attempted/completed/removal details only when validated and supplied. Summary counts use aggregate module results; a failed module with rollback details remains a failure, with a rollback breakdown where applicable.

`newNoopApplyRunner` remains the route for default and dry-run modes, so no `CommandRunner` is reachable there. Confirmed mode constructs the local provider only for selected dotfile plan steps. No fallback, retry, clone/pull/fetch, or Bootstrap rollback is introduced.

## Planned file changes

| File | Design change |
|---|---|
| `internal/execution/dotlink_report.go` | Recursive duplicate-key scanner, strict decode, semantic validation, domain report/detail values, and safe errors. |
| `internal/execution/dotfiles_provider.go` | Exact argv, stdout-first parse/reconciliation, command metadata, and base diagnostics. |
| `internal/execution/provider.go` | Provider returns typed report or typed safe error. |
| `internal/execution/dotfiles_installer.go` | Translate report to aggregate `StepResult` plus per-link details. |
| `internal/execution/types.go` | Per-link outcome enum/detail and base diagnostic fields; preserve aggregate status compatibility. |
| `cmd/dbootstrap/render.go` | Aggregate/per-link rendering, rollback breakdown, and truthful base-failure wording. |
| focused execution/CLI tests and `internal/execution/testdata/dotlink-report/*.json` | Fake-runner fixtures and contract coverage only. |
| OpenSpec proposal/spec/design artifacts | Corrected planning contract; tasks are intentionally deferred. |

## Test strategy

Use table-driven parser tests and fake `CommandRunner` tests; never invoke Dotlink or a real home directory.

1. Parser fixtures: valid successful all-changed, all-unchanged, mixed changed/unchanged, valid failed, and rolled-back reports; malformed JSON, unknown field, trailing JSON, human stdout, unsupported schema, selection/entry contradictions, and missing causes.
2. Duplicate-key fixtures: duplicate top-level `status`; duplicate nested keys independently in entry, cause, failure, and rollback objects. Assert rejection occurs before domain translation and no fallback is attempted.
3. Provider tests: exact argv; success + success report; non-success + valid failed report preserving detail; non-success + absent/malformed/duplicate-key report yields generic failure; both command/report status inconsistencies fail safely; stderr is ignored.
4. Installer tests: aggregate mapping for all unchanged, mixed changed/unchanged, failed, rolled_back, aggregate failed, and generic errors; per-link details remain ordered and truthful.
5. Base diagnostics/rendering: resolution failures show attempted candidate/source/modules/cause without canonical label; validated paths render as canonical; per-link and rollback output are truthful.
6. Mode/exit tests: default/dry-run make zero runner calls and remain `not_implemented`; confirmed aggregate failures render and exit non-zero.
7. Run `go test ./internal/execution ./cmd/dbootstrap`, then `go test ./...`.

## Rollout and rollback

The consumer intentionally fails closed on upstream contract drift. Roll back by reverting the consumer/rendering change; no Dotlink state is migrated or repaired. `dotfiles-base-failure-context` remains superseded and its untracked planning directory is handled only during archive work if present.
