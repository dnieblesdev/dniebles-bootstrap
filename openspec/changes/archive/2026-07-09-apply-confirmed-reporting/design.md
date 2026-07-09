# Design: Apply Confirmed Reporting

## Context

`dbootstrap apply` already separates planning from execution and selects one of three apply modes in `cmd/dbootstrap/main.go`: default non-mutating, `--dry-run`, and confirmed `--yes`. Confirmed mode is already constrained to real Homebrew execution for brew-backed `tool` and `package` steps; runtimes, dotfiles, non-brew resources, and missing Homebrew paths remain noop/advisory. This slice is therefore rendering-focused and must not change provider behavior or add mutation paths.

## Goals

- Render a Summary section for apply execution in default, dry-run, and confirmed modes when selected execution results exist.
- Use stable user-facing summary/step categories: `changed`, `unchanged`, `not supported yet`, and `failed`.
- Preserve internal execution statuses and report model fields unless implementation proves impossible without model changes.
- Make confirmed-mode copy explicit that only brew-backed `tool` and `package` steps are eligible to mutate the host.
- Render an explicit empty selected-plan state instead of a zero-count summary.

## Non-goals

- No apt provider, dotfiles execution, runtime execution, extra catalog targets, retries, concurrency, or new mutation paths.
- No changes to Homebrew installer/provider behavior.
- No changes to planning semantics or execution report fields expected for this slice.

## Design

### Rendering location

Keep the feature in `cmd/dbootstrap/render.go`, centered on `renderExecutionReport`. `cmd/dbootstrap/main.go` should not need changes unless RED tests reveal that the renderer lacks required context. The existing `execution.ExecutionReport.Results` shape is enough for summary counts and per-step labels.

Recommended renderer flow:

1. Print the existing `Execution Report` title and `Mode:` line.
2. In confirmed mode, print non-ambiguous mutability framing, for example: `Confirmed mode: only brew-backed tool/package steps may change this machine; runtime, dotfile, non-brew, and unsupported steps remain non-mutating.`
3. If `len(report.Results) == 0`, print an explicit empty state such as `No actionable steps were selected; nothing to apply.`, skip the Summary, and continue to `Manual Actions`.
4. Otherwise render `Summary:` in fixed category order: `changed`, `unchanged`, `not supported yet`, `failed`.
5. Render `Steps:` with user-facing category labels instead of raw internal statuses.
6. Render `Manual Actions:` unchanged.

### Status/category mapping

Add small unexported helpers in `render.go`; do not rename `internal/execution` statuses.

| Internal `execution.StepStatus` | User-facing category | Rationale |
| --- | --- | --- |
| `installed` | `changed` | Real installer reported success/host mutation. |
| `skipped` | `unchanged` | The step did not mutate the host; the message/manual action explains why. |
| `not_implemented` | `not supported yet` | Meets user-facing wording while preserving internal vocabulary. |
| `failed` | `failed` | Error outcome remains visible. |
| unknown/future status | conservative `failed` or explicit tested fallback | Avoid overclaiming success/change for unrecognized outcomes. |

Suggested helper names:

- `executionSummaryCategory(status execution.StepStatus) string`
- `executionSummaryCounts(results []execution.StepResult)` using a small fixed-field struct or fixed-order map
- `renderExecutionSummary(w io.Writer, results []execution.StepResult)`
- `renderExecutionStepStatus(status execution.StepStatus) string`

### Per-step output

Per-step output should preserve the current resource/message shape but display the category in brackets, for example:

- `tool:fd [changed] installed fd with Homebrew`
- `runtime:go [not supported yet] noop installer does not perform real installation`
- `package:ripgrep [unchanged] skipped because Homebrew must be installed manually before brew-backed resources can be applied`
- `tool:fd [failed] brew install fd failed ...`

This satisfies the spec without changing provider return values or `execution.StepStatusNotImplemented`.

### Empty selected-plan behavior

When the execution report has zero results, render a clear sentence and do not render a zero-count Summary table. Keep this in `renderExecutionReport` so `runApply` remains a thin composition root. Existing planning-error behavior remains unchanged: planning errors render plan diagnostics and do not render an execution report.

### Confirmed-mode mutability framing

Replace the current warning with wording that does not imply all selected resources mutate. It must explicitly say only brew-backed `tool` and `package` steps may change the machine, while runtime, dotfile, non-brew, and unsupported work remains non-mutating/advisory or `not supported yet`.

## Contracts

- `execution.ExecutionReport` and `execution.StepResult` remain unchanged.
- Internal execution status constants remain unchanged.
- CLI output contract changes are limited to apply execution rendering.
- Manual actions remain appended/rendered as before.
- No provider behavior, catalog semantics, or mutation eligibility changes.
- Later implementation should use idiomatic small Go helpers, fixed output ordering, and direct table-driven tests.

## RED/GREEN test strategy

Start with failing tests before implementation:

1. `cmd/dbootstrap/render_test.go`
   - Summary renders all four categories in fixed order for mixed results.
   - `not_implemented` renders as `[not supported yet]` in step output.
   - Confirmed mode renders the new mutability framing.
   - Empty report renders the explicit empty-state sentence and no zero-count Summary.
2. `cmd/dbootstrap/main_test.go`
   - Default, `--dry-run`, and `--yes` apply cases include `Summary:` when execution results exist.
   - Confirmed missing-brew case counts skipped as `unchanged` and not-implemented as `not supported yet`.
   - Confirmed brew-present case counts installed as `changed` while non-brew/runtime work remains `not supported yet`.
3. Leave provider and execution unit tests unchanged unless exact CLI-message assertions need updates due rendering-only wording.

Run focused tests first with `go test ./cmd/dbootstrap`, then the full strict runner `go test ./...`. Warn during implementation if exact-output fixture updates become review-heavy, though this should remain a small single-PR change.

## Rollout and rollback

Rollout is rendering/test/spec-only with no host mutation impact. Rollback is a straight revert of `cmd/dbootstrap/render.go`, related tests, and OpenSpec wording for this change.

## Risks and mitigations

- Category drift from internal statuses: centralize the mapping in unexported render helpers with direct tests.
- Users may infer all `--yes` work changed the host: confirmed preamble must state the brew-backed `tool`/`package` limit.
- `skipped` may be ambiguous: categorize as `unchanged` and rely on the message/manual action for the reason.
- Future statuses could be miscategorized: require explicit tests and conservative handling when new statuses are added.
