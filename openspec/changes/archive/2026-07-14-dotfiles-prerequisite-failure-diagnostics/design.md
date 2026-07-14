# Design: Dotfiles Prerequisite Failure Diagnostics

## Technical Approach

Complete diagnostics additively at the execution/result boundary. Keep base resolution, validation, command semantics, report parsing, statuses, and `DotfilesFailure`; add explicit prerequisite identity separate from its cause, then render curated structured facts once. This implements both delta specs without changing provider, parser, planning, or legacy seams.

## Architecture Decisions

| Decision | Alternative / tradeoff | Rationale |
|---|---|---|
| Add a prerequisite target carrier beside `PrerequisiteErr` | Derive the path from an error or overload `Executable` | Validation errors can expose a resolved escaping target, not the attempted runner/module path. |
| Capture candidate before each validation call | Reconstruct after `validateRepo` fails | The provider owns the lexical candidate at the only truthful point; no runner invocation is needed. |
| Render classified causes, never `err.Error()` | Print wrapped error text | Keeps labels stable, terminal-safe, and truthful while `errors.Is`/`errors.As` remain available to callers. |
| Deduplicate complete base snapshots only | Deduplicate formatted lines | Preserves distinct prerequisite candidates and causes deterministically. |

## Data Flow

```
PlanStep.Ref + modules -> resolve base -> provider builds attempted target -> validate -> runner -> parser
                                             |                    |          |
                                             v                    v          v
                                      DotfilesFailure{target,cause} -> installer -> StepResult -> renderer
```

After a validated base, `LocalDotfilesProvider` creates the attempted runner path (`base/bin/dotlink`) before runner validation and each attempted module path (`base/<module>`) before module validation. On failure it returns `*DotfilesFailure`, carrying that target and `PrerequisiteErr`; the installer already recognizes this typed error and retains it in `StepResult`. A prerequisite rejection creates no links and makes zero `CommandRunner` calls.

## Interfaces / Contracts

```go
type DotfilesPrerequisiteTargetKind string

const (
	DotfilesPrerequisiteRunner DotfilesPrerequisiteTargetKind = "runner"
	DotfilesPrerequisiteModule DotfilesPrerequisiteTargetKind = "module"
)

type DotfilesPrerequisiteTarget struct {
	Kind               DotfilesPrerequisiteTargetKind
	AttemptedCandidate string // lexical candidate; never canonical/validated
}

type DotfilesFailure struct {
	// existing fields
	PrerequisiteTarget *DotfilesPrerequisiteTarget
	PrerequisiteErr    error
}
```

`StepResult.Ref` owns the operation; `BaseDiagnostic` owns selected modules and base facts; `DotfilesPhase` owns the lifecycle label; `PrerequisiteTarget` owns only the pre-validation runner/module candidate; and `PrerequisiteErr`, `ExecutionErr`, and `ParseErr` own independent typed causes. `DotfilesFailure.Unwrap()` includes every non-nil cause.

For a missing runner, render `attempted runner candidate`; for a missing or escaping module, render `attempted module candidate`. Neither label says executable, canonical, or validated. The renderer maps recognized sentinels/typed errors (for example missing path, escaping base, invalid module) to curated cause labels; it does not render `err.Error()`. Every rendered field is terminal-sanitized and bounded; retain the 4096-byte sanitized stderr limit. Base snapshots compare source, attempted candidate, canonical path, ordered modules, and cause, rendering one equal snapshot and both unequal snapshots.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/execution/types.go` | Modify | Add prerequisite target/cause transport and multi-unwrapping. |
| `internal/execution/dotfiles_provider.go` | Modify | Capture runner/module candidates before validation; return typed prerequisite failures. |
| `cmd/dbootstrap/render.go` | Modify | Render bounded, sanitized target and classified cause; deduplicate base snapshots. |
| `internal/execution/{dotfiles_provider,dotfiles_installer}_test.go` | Modify | Cover prerequisite transport, typed errors, and zero calls. |
| `cmd/dbootstrap/{render,main}_test.go` | Modify | Cover truthful labels, deduplication, and confirmed apply. |

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Missing runner carries runner target plus `fs.ErrNotExist`; phase is prerequisite; zero calls | Table-driven provider/installer fake. |
| Unit | Missing module and symlink-escaping module each carry the original attempted module candidate, not resolved target; retain `fs.ErrNotExist`/`ErrDotfilesPathEscapes`; zero calls | `t.TempDir()` repository fixtures and `errors.Is` assertions. |
| Unit | Renderer uses attempted runner/module labels, curated causes, no raw wrapped error text; controls/oversize fields are escaped/bounded | Exact/count buffer assertions. |
| Unit | Equal base snapshot renders once; a different target or cause remains visible | Deterministic render assertions. |
| Integration | Confirmed `apply --yes --resource dotfile:bash` with missing runner is non-zero and shows operation, module, prerequisite, target, cause, and zero calls | Existing injected-runner seam. |

## Threat Matrix

| Boundary | Applicability | Design response / planned RED test |
|---|---|---|
| Documentation-like paths | N/A — no file classification | None. |
| Git repository selection | N/A — no VCS operation | None. |
| Commit state | N/A — no commit operation | None. |
| Push state | N/A — no push operation | None. |
| PR commands | N/A — no PR automation | None. |

The executable/process boundary validates prerequisite candidates before invocation; missing runner/module and escaping module RED tests prove safe failure and zero calls.

## Migration / Rollout

No migration required. Revert this isolated transport/rendering slice and its tests; prerequisite rejection, non-zero failure, and no-runner behavior remain.

## Open Questions

None. Excluded: `DotfilesBaseReporter`, provider/parser redesign, statuses, planning/configuration work, monolith cleanup, and `PlanStep.AttentionReasons -> StepResult`.
