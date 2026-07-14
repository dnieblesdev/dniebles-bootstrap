# Review Ledger: Dotfiles Prerequisite Failure Diagnostics

## Target

- Phase: design
- Base: `e576669dcc4f565b9d474c66b9a53db2b882713e`
- Design SHA-256: `b9fea57350401dd10fd96eccbe18395872986f754d6f772704598db33f08e79b`
- Judgment Day round: 0
- Corrected design SHA-256: `3ab93237fabfb2da3782486475c8718f8422d0854a02c6dcaa4ab7f741066709`
- Current cleaned design SHA-256: `de7c6070fde07348cce1b433f3323c590d74093fe70f51461e753983487fabe5`
- State: approved

## Findings

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-001 | judgment-day | `openspec/changes/dotfiles-prerequisite-failure-diagnostics/design.md:45-53` | CRITICAL | verified | Both blind judges confirmed that the corrected design now captures an explicit attempted runner/module candidate before validation, transports it separately from typed causes, renders it as attempted/unvalidated, and covers missing and escaping paths with zero runner calls. |

## Correction Scope

Amend only the design as needed to define explicit safe prerequisite-candidate ownership, transport, rendering, and focused tests. Do not expand into provider redesign, legacy compatibility, planning, or `PlanStep.AttentionReasons`.

## Scoped Re-judgment

Both blind judges returned no findings for the correction scope. JD-001 is verified and the design is approved for task planning.

## Apply Review — Round 0

- Native lineage: `review-f0264cd1b8e63fe5`
- Base: `e576669dcc4f565b9d474c66b9a53db2b882713e`
- Projection: workspace
- Tracked implementation/test diff SHA-256: `0511eca3b94cbebf0719346cda578e28eb732e030911e90e40a32a4af0fcec35`

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-002 | judgment-day | `cmd/dbootstrap/render.go:115-116,163-165,182-216` | CRITICAL | verified | Round 1 added focused oversized prerequisite/base candidate RED tests and bounds each implicated rendered diagnostic field to 4096 bytes after terminal escaping with `...[truncated]`. Both blind full-slice re-judges returned no findings; focused and full verification passed. |
| JD-003 | judgment-day | `internal/execution/dotfiles_provider.go:177-180` | WARNING | info | One judge reported that `filepath.Join` cleans traversal segments before invalid-module validation. This was not independently confirmed and does not drive correction in this round. |

### Round 1 Correction Scope

Bound every newly rendered diagnostic field while preserving terminal sanitization and existing stderr behavior. Add focused oversized-candidate tests. Do not change provider candidate construction or any unconfirmed/suspect behavior.

### Round 1 Full-slice Re-judgment

Both blind judges returned no findings for the complete corrected 378-line code/test unit. JD-002 is verified; JD-003 remains uncorroborated informational context. Artifact cleanup changed only documentation claims and produced the current cleaned design hash above.

## Native Full-4R Review — Round 1

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| R3-001 | reliability | `cmd/dbootstrap/render_test.go` | BLOCKER | verified | Added focused external-renderer assertions for human-visible `phase: command-execution` plus `cause: dotlink command failed`, and `phase: report-validation` plus `cause: invalid dotlink report`; focused and required package suites pass with production hashes unchanged. Scoped reliability re-review returned no findings. |
| R4-001 | resilience | `internal/execution/dotfiles_installer.go:45-54` | WARNING | info | Prerequisite diagnostics do not add explicit retry guidance. Non-blocking first-pass context only; it does not drive correction or re-review. |

### Round 1 Correction Scope

Add focused externally observable renderer coverage for command-execution and report-validation phase/cause output. Do not change production behavior, recovery copy, prerequisite transport, provider behavior, or unrelated tests.

### Round 1 Scoped Re-review

The reliability reviewer returned no findings when restricted to R3-001 and the fix-touched test lines. R3-001 is verified; production bytes remained unchanged.
