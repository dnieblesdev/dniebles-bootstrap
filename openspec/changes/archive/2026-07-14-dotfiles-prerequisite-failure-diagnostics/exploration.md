## Exploration: dotfiles prerequisite failure diagnostics

### Current State

Current main (`e576669dcc4f565b9d474c66b9a53db2b882713e`) already has the
execution-owned dotfiles base diagnostic and the validated dotlink report
pipeline. The confirmed apply composition is in `cmd/dbootstrap/main.go` and
uses `LocalDotfilesProvider` through `DotfilesInstaller`.

The minimum scenario (a valid repository with no `bin/dotlink`) currently
behaves as follows:

1. `LocalDotfilesProvider.validateRepo` rejects the missing runner path before
   invoking `CommandRunner`, returning only a wrapped filesystem error.
2. `DotfilesInstaller.Install` marks the step failed and retains the base
   diagnostic, but does not retain a structured prerequisite failure carrier.
3. `renderExecutionReport` prints the generic module failure and canonical base
   context, but not the attempted operation, failing phase, executable/candidate,
   or concrete missing-runner cause. The exit remains non-zero and the runner
   call count remains zero.

The current paths have different observability contracts:

| Path | Current useful facts | Current gap | Needed in this slice |
|---|---|---|---|
| Base resolution | source, attempted candidate, modules, safe cause; canonical path only after validation | no phase/operation label in the human failure line | Preserve the existing safe base diagnostic and label resolution context without false canonical identity |
| Prerequisite/repository validation | failed status, base identity, wrapped cause internally | missing `bin/dotlink` cause is not transported/rendered as a structured failure | Add the smallest structured carrier needed for operation, modules, phase, executable candidate, and cause; do not invoke the runner |
| Command execution | `DotfilesFailure` already carries executable, command, exit code, bounded sanitized stderr, and independent execution/report errors; valid failed reports remain translated | human output does not consistently label the phase/cause for report-validation failures | Preserve the current carrier and render its existing independent causes with an explicit phase |
| Report validation | `ParseErr` retains `ErrInvalidDotlinkReport` and typed parser causes; invalid reports are rejected without stderr fallback | parse cause is not directly rendered as an operator-facing diagnostic | Render the safe parse/validation cause; do not change parser semantics |

The complete dirty monolith comparison against current main was reviewed, not
only remembered filenames. Its 27-file diff is broad (630 insertions and 1264
deletions, including archived/spec cleanup). The actionable diagnostic
remainder is concentrated in `internal/execution/types.go`,
`internal/execution/dotfiles_base.go`, `internal/execution/dotfiles_provider.go`,
`internal/execution/dotfiles_installer.go`, `cmd/dbootstrap/render.go`, and
their focused tests. The monolith's report-consumption and rollback changes
are already represented in current main and are not part of this slice.

### Affected Areas

- `internal/execution/types.go` — smallest execution failure carrier change, if needed, for phase/operation/modules and preserved typed causes.
- `internal/execution/dotfiles_provider.go` — classify resolution, repository/prerequisite, runner, command, and report-validation failures without changing validation safety or command semantics.
- `internal/execution/dotfiles_base.go` — retain existing safe resolution diagnostics; only adapt failure transport if required by the bounded carrier.
- `internal/execution/dotfiles_installer.go` — preserve structured failures and the selected module while translating provider errors; do not infer links from prerequisite failures.
- `cmd/dbootstrap/render.go` — render operation, modules, phase, and concrete safe cause with terminal sanitization, without duplicate base context.
- `internal/execution/dotfiles_provider_test.go` — table-driven provider contracts for missing runner path, resolution/prerequisite failures, command failure, and invalid/inconsistent reports; assert zero runner calls where applicable.
- `internal/execution/dotfiles_installer_test.go` — assert carrier/error transport and no inferred link details for prerequisite failures.
- `cmd/dbootstrap/render_test.go` — assert terminal-safe human output and one base context.
- `cmd/dbootstrap/main_test.go` — one end-to-end confirmed-apply scenario for a repository missing `bin/dotlink`, including failed exit, attempted operation, module, phase, concrete cause, and zero command calls.
- `openspec/specs/dotfiles-provider/spec.md` and `openspec/specs/execution-contracts/spec.md` — modify only the existing prerequisite/failure-rendering requirements; no new status or planning requirement.

The legacy `DotfilesBaseReporter` plus `baseContext` formatting seam is
explicitly rejected. The complete current production search has no caller for
that interface; it is only a compatibility-oriented provider shape in the
dirty monolith. The existing execution-context and base-diagnostic seams are
the live boundaries.

### Approaches

1. **Port the monolith failure model wholesale** — replace the current
   `DotfilesFailure` shape, rework provider error composition, and replace
   renderer/base-context handling.
   - Pros: direct monolith parity and one apparent carrier for every failure.
   - Cons: reopens already integrated report/execution behavior, changes
     established error fields, increases regression surface, and mixes
     prerequisite diagnostics with provider redesign.
   - Effort: High

2. **Bounded diagnostic completion on current contracts** — retain the current
   base diagnostic, report model, statuses, and command behavior; add only the
   missing prerequisite failure identity and render existing command/parser
   causes with explicit phase labels.
   - Pros: fixes the missing-runner acceptance case, covers resolution,
     execution, and report-validation observability, preserves typed error
     identity, and stays comfortably below the 800-line review budget.
   - Cons: requires carefully defining phase names and avoiding duplicate base
     output across the base diagnostic and failure carrier.
   - Effort: Medium

3. **Render raw provider errors directly** — append `error.Error()` to the
   generic step message without a structured carrier.
   - Pros: smallest code diff.
   - Cons: loses stable operation/module/phase contracts, risks exposing
     unbounded or terminal-unsafe text, does not compose independent report and
     execution causes, and is not suitable for the required contract.
   - Effort: Low

### Recommendation

Choose the bounded diagnostic completion. Keep resolution facts in
`DotfilesBaseDiagnostic`; keep validated report/link facts unchanged; and use
one execution-owned failure carrier only where a prerequisite, command, or
report-validation failure needs additional identity. The rendered failure
should explicitly identify the attempted dotlink operation, selected module(s),
failing phase, and a sanitized concrete cause. For the minimum case the
observable contract should distinguish `validate repository` (or an equally
stable prerequisite phase), `dotfile:bash`/`bash`, the candidate
`<canonical-base>/bin/dotlink`, and the underlying missing-path error, while
proving that `CommandRunner` was not called.

Preserve the existing `errors.Is`/`errors.As` behavior and the valid-failed
report transport. Report-validation failures need only gain safe rendering of
the existing parse cause; they do not need parser redesign. Do not port the
legacy provider seam, `PlanStep.AttentionReasons -> StepResult`, provider
redesign, configuration mutation, new statuses, planning changes, or unrelated
refactors. A focused production/test/spec slice should remain well below the
800 changed-line budget (expected authored scope: roughly 150–300 lines).

### Risks

- A rejected prerequisite must never be presented as a validated canonical
  executable or trigger the command runner.
- Base diagnostics and failure carriers can describe the same base; rendering
  must deduplicate them deterministically without hiding a distinct failure
  candidate.
- Parser and command errors may contain control characters or large stderr;
  reuse the existing bounded/sanitized transport and terminal renderer.
- The monolith contains unrelated deletions and compatibility shape changes;
  cherry-picking its diff would exceed the independent slice and risk regressing
  already integrated report behavior.

### Ready for Proposal

Yes. The proposal should name this as a bounded diagnostic-completion change,
use the missing `bin/dotlink` scenario as the acceptance anchor, define stable
observable fields for resolution/prerequisite, execution, and report-validation
failures, and explicitly record the rejected `DotfilesBaseReporter` seam and
all out-of-scope planning/provider changes.
