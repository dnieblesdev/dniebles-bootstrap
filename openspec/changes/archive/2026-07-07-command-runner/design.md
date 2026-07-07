# Design: Command Runner

## Technical Approach

Add a small argv-style process execution primitive inside `internal/execution` without connecting it to `apply` or real installers. The existing plan `Runner` and noop installers remain unchanged; future installer slices can depend on the new `CommandRunner` contract when real mutation is explicitly introduced. This matches the proposal and preserves the existing execution-contracts rule that `apply` is noop-only today.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|----------|--------|--------------------------|-----------|
| Runner shape | Add `CommandRunner` beside existing `Runner` | Reuse existing `Runner`; place under `internal/system` | Existing `Runner` dispatches plan steps, so a separate name avoids semantic drift while keeping execution infrastructure colocated. |
| Command model | `Executable string` plus `Args []string` | Shell string; `sh -c`; pipeline support | Argv data prevents accidental shell interpretation and keeps catalog metadata structured/inert. |
| Result model | Return captured stdout/stderr, status, exit code, duration, and error | Return only `error`; stream output directly | Installers need auditable data and tests need deterministic assertions. |
| Dry-run | Provide deterministic noop command runner | Let callers skip execution manually | A noop runner gives future dry-run paths the same contract without host mutation. |
| Tests | Use fake exec seam and Go test helper process only | Depend on `/bin/echo`, `false`, or package managers | Tests must be portable and must not mutate or require arbitrary host tools. |

## Data Flow

```text
Future installer ── CommandRequest ──> CommandRunner
      │                                  │
      │                                  ├─ Real: exec.CommandContext(argv)
      │                                  └─ Noop: deterministic dry-run result
      └──────────── CommandResult <──────┘

Current apply ── existing execution.Runner ── NoopForKind installers only
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/execution/command.go` | Create | Define `CommandRequest`, `CommandResult`, `CommandStatus`, `CommandRunner`, and validation helpers. |
| `internal/execution/os_command_runner.go` | Create | Implement process execution with `exec.CommandContext`, optional cwd/env, timeout wrapping, stdout/stderr capture, exit-code extraction, and duration measurement. |
| `internal/execution/noop_command_runner.go` | Create | Return a deterministic dry-run `CommandResult` without invoking host commands. |
| `internal/execution/command_runner_test.go` | Create | Cover success, non-zero exit, stderr capture, missing executable, timeout/cancel, validation, and noop determinism. |
| `internal/execution/types.go` | Modify | Add command status vocabulary only if not kept in `command.go`; existing step statuses remain separate. |
| `cmd/dbootstrap/main.go` | Unchanged | Keep `apply` wired only to `NoopForKind`; no command runner integration in this slice. |
| `catalog/bootstrap.toml` | Unchanged | Do not add raw command metadata. |

## Interfaces / Contracts

```go
type CommandRequest struct {
    Executable string
    Args       []string
    Dir        string
    Env        []string
    Timeout    time.Duration
}

type CommandStatus string

const (
    CommandStatusSucceeded CommandStatus = "succeeded"
    CommandStatusFailed    CommandStatus = "failed"
    CommandStatusTimedOut  CommandStatus = "timed_out"
    CommandStatusNotRun    CommandStatus = "not_run"
)

type CommandResult struct {
    Request  CommandRequest
    Status   CommandStatus
    ExitCode int
    Stdout   string
    Stderr   string
    Duration time.Duration
    Err      error
}

type CommandRunner interface {
    RunCommand(context.Context, CommandRequest) CommandResult
}
```

`ExitCode` should be the process exit code when available and `-1` when no process exit occurred. Empty `Executable` is validation failure with `not_run`.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Request validation and noop result | Table tests in `internal/execution`. |
| Unit | OS runner success/failure/output/status | Go helper process pattern via `os.Args[0]` and test env vars. |
| Unit | Timeout/context cancellation | Helper process that blocks until context kills it; assert `timed_out` or canceled status data. |
| Integration | Current `apply` remains noop-only | Existing CLI/execution tests plus regression assertion that `cmd/dbootstrap/main.go` is not changed to use `CommandRunner`. |
| E2E | None in this slice | Real installer behavior is explicitly out of scope. |

## Migration / Rollout

No migration required. This slice adds unused internal infrastructure and tests only; real installer wiring must happen in a later SDD change.

## Open Questions

- [ ] None.
