# Design: Apply Safety Contract

## Technical Approach

Keep `apply` as a CLI-only safety contract over the existing planning pipeline and noop execution runner. Add apply-specific mode parsing in `cmd/dbootstrap/main.go`, pass the selected mode to rendering in `cmd/dbootstrap/render.go`, and leave `internal/execution` non-mutating. This implements the delta spec for default-safe, explicit dry-run, conflicting flag rejection, and reserved `--yes` behavior without wiring real installers or `CommandRunner` mutation.

## Architecture Decisions

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Add apply mode in CLI | Small surface; avoids execution-layer semantics before mutation exists. | Chosen. `cmd/dbootstrap` owns safety flag parsing and mode labels. |
| Add mode to `internal/execution.ExecutionReport` | More reusable later, but expands execution contracts now. | Rejected for this slice; rendering can receive mode separately. |
| Make `--yes` execute real installers | Matches future intent, but violates the safety contract and out-of-scope guardrails. | Rejected. `--yes` is accepted and reported as reserved confirmed mode only. |

## Data Flow

```text
args ──→ parseApplyFlags ──→ buildPlan ──→ noop execution.Runner ──→ renderExecutionReport(mode)
              │                    │                 │
              └─ reject conflicts   └─ stop on errors └─ NoopForKind only
```

Planning failures keep the existing behavior: render plan diagnostics and do not render an execution report. Accepted apply modes all reach the same noop installers.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modify | Add `applyMode`, `parseApplyFlags`, `--dry-run`, `--yes`, and conflict validation while preserving plan target validation. |
| `cmd/dbootstrap/render.go` | Modify | Render selected apply mode before execution steps so noop/default/dry-run/confirmed-future output cannot be confused. |
| `cmd/dbootstrap/main_test.go` | Modify | Cover default non-mutating apply, explicit `--dry-run`, accepted `--yes`, and rejected `--dry-run --yes`. |
| `cmd/dbootstrap/render_test.go` | Modify | Update execution report expectations to include the mode line. |
| `internal/execution/*` | No change | Keep noop installers and command runner contracts unchanged. |

## Interfaces / Contracts

```go
type applyMode string

const (
    applyModeDefaultNonMutating applyMode = "default-non-mutating"
    applyModeDryRun             applyMode = "dry-run"
    applyModeConfirmedFuture    applyMode = "confirmed-future-noop"
)
```

`parseApplyFlags(args, stderr)` returns the existing `planning.PlanRequest`, catalog path, selected `applyMode`, and `ok`. It MUST reject `--dry-run && --yes` with usage plus a clear error, e.g. `error: --dry-run and --yes cannot be combined`.

`renderExecutionReport(stdout, mode, report)` prints:

```text
Execution Report
Mode: <mode label>
```

before `Steps:`. Confirmed future mode should explicitly include that it is still noop/non-mutating.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Mode label rendering, empty reports | Update `render_test.go` exact-output tests. |
| CLI integration-style | Accepted modes and conflict rejection | Extend `TestRunApplyCommand` table in `main_test.go`; assert no execution report on conflict. |
| Execution | Non-mutating boundary | Keep existing `internal/execution` noop/runner tests unchanged; no new mutation path. |

## Migration / Rollout

No migration required. This changes CLI validation/output only and preserves noop execution for every accepted mode.

## Open Questions

None.
