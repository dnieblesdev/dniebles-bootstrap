# Design: Apply Command Dry Run

## Technical Approach

Add `dbootstrap apply` as a safe dry-run bridge in the CLI composition root. The command reuses the existing `plan` target parsing, catalog loading, detection, dotfile-state merge, and `planning.BuildPlan()` flow, then sends only an error-free plan to `execution.Runner` configured with noop installers. The output is a separate execution report so planning statuses (`planned`, `attention_required`) do not masquerade as execution outcomes (`not_implemented`).

## Architecture Decisions

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Extract shared `parsePlanRequest(command, args, stderr)` vs duplicate `runPlan` flags | Extraction reduces drift but touches the CLI path once. Duplication would make `apply` validation regress easily. | Extract shared flag/parser helper and keep command-specific usage labels. |
| Add `execution.NoopForKind(kind)` vs rely on missing installer fallback | Missing fallback is safe but reports “no installer registered,” not deliberate dry-run support. | Add kind-aware noop installers for `tool`, `runtime`, `package`, and `dotfile` so dry-run reports intentional noop execution. |
| Render execution in `render.go` vs inline in `runApply` | Renderer keeps CLI orchestration readable and matches existing plan rendering tests. | Add `renderExecutionReport()` to `cmd/dbootstrap/render.go`. |
| Execute plans with planning errors | Would produce confusing partial reports. | Planning errors return `exitFailure` after plan diagnostics and before runner construction. |

## Data Flow

```text
run(args)
  └─ apply ─→ runApply
       ├─ parse shared target flags
       ├─ load catalog + detect environment/config/install/dotfiles state
       ├─ planning.BuildPlan(...)
       ├─ if planning error: render plan + diagnostics, exit 1
       └─ execution.NewRunner(NoopForKind(...)).Run(context.Background(), plan)
            └─ renderExecutionReport(stdout, report)
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modify | Add `case "apply"`, `runApply()`, `printApplyUsage()`, and shared parsing/build-plan helper used by both `plan` and `apply`. Import `context` and `internal/execution`. |
| `cmd/dbootstrap/render.go` | Modify | Add `renderExecutionReport()` using execution status/message fields and an explicit “Execution Report” heading. |
| `cmd/dbootstrap/main_test.go` | Modify | Add apply CLI coverage for success, resource flags, validation parity, catalog failures, and planning error short-circuit. Update unknown-command/usage expectations. |
| `cmd/dbootstrap/render_test.go` | Modify | Add execution report rendering tests separate from plan rendering. |
| `internal/execution/noop.go` | Modify | Add a small kind-aware noop installer wrapper and `NoopForKind(kind planning.ResourceKind) Installer`. |
| `internal/execution/noop_test.go` | Modify | Cover `NoopForKind` supported kind and non-mutating `not_implemented` result. |
| `internal/execution/regression_test.go` | Modify | Remove `TestNoApplyCommandInCLI`; replace only if needed with a safety regression proving noop helpers stay non-mutating. |

## Interfaces / Contracts

```go
func NoopForKind(kind planning.ResourceKind) Installer
func renderExecutionReport(w io.Writer, report execution.ExecutionReport)
```

`NoopForKind` must only wrap `NoopInstaller.Install`; it must not invoke command execution, dotlink, clone, retry, concurrency, or host mutation. `runApply` must use the same planning request semantics as `runPlan`, including deduped repeated `--resource` values.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `NoopForKind` kind dispatch and noop result | Extend `internal/execution/noop_test.go`. |
| Unit | Execution report text | Add exact-output tests in `cmd/dbootstrap/render_test.go`. |
| CLI integration-style | `apply` accepts `--profile`, repeated `--resource`, `--catalog`; shares validation with `plan` | Extend table-driven `run()` tests with stubs. |
| Regression | Planning errors short-circuit execution | Assert stdout contains plan diagnostics but no “Execution Report”. |
| Safety | Obsolete no-apply boundary removed | Delete/replace `TestNoApplyCommandInCLI` with functional apply dry-run coverage. |

## Migration / Rollout

No migration required. This slice exposes `apply` as dry-run-only and intentionally returns noop `not_implemented` execution outcomes.

## Open Questions

- [ ] None.
