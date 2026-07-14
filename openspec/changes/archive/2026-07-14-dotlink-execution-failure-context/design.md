# Design: Dotlink Execution Failure Context

## Technical Approach

Extend the validated-report boundary with execution-owned `DotfilesFailure`. The provider consumes the merged target's validated `DotfilesExecutionContext`, derives `<canonical>/bin/dotlink`; never resolves, validates, or relabels a base. Success, dry-run, dotlink semantics remain unchanged.

## Architecture Decisions

| Decision | Alternative / tradeoff | Rationale |
|---|---|---|
| Primary base source is `StepResult.BaseDiagnostic` | Reconstruct from error text | Result/report context is stable and presentation-owned. |
| Failure carries an optional structured base snapshot | Omit failure-time comparison data | Supports transport and semantic deduplication without copying text into messages. |
| Two joined cause fields | One linear cause | Command and parser identities remain independently discoverable. |
| Sanitize before transport | Store raw stderr | Limits terminal risk and diagnostic size. |

## Data Flow

```
validated context -> canonical executable -> CommandRunner
  -> stdout parser -> (report, DotfilesFailure) -> installer StepResult -> renderer
```

| Command / report | Report | Failure fields |
|---|---|---|
| succeeded / valid success | retained | neither |
| failed / valid failed | retained | `ExecutionErr` |
| failed / missing or invalid | discarded | `ExecutionErr`, `ParseErr` |
| succeeded / failed or inconsistent | discarded | `ParseErr` |

Missing runner creates `ExecutionErr` and invokes no command. Stderr is never a report source. Failure fields retain runner, request, canonical executable, optional exit code, report status, escaped bounded stderr, and an optional base snapshot.

## Interfaces / Contracts

```go
type DotfilesFailure struct {
    Executable, Runner string
    Command            CommandRequest
    ExitCode           *int
    Stderr             string
    ReportStatus       DotlinkReportStatus
    BaseSnapshot       *DotfilesBaseDiagnostic
    ExecutionErr, ParseErr error
}
func (f *DotfilesFailure) Unwrap() []error // non-nil fields only
```

`StepResult` gains `DotfilesFailure *DotfilesFailure`; its `BaseDiagnostic` remains the primary report/result-owned base presentation source. The installer translates a returned valid failed report even with an error, attaches the failure, and marks the step failed. It MUST stop appending `baseContext` to `StepResult.Message`: messages are short module summaries. The renderer MUST render structured fields, never blindly append `err.Error()` context.

For a command failure, construct `ExecutionErr = errors.Join(ErrDotlinkCommandFailed, result.Err)` (omit nil runner cause); therefore one top-level failure supports both `errors.Is(err, ErrDotlinkCommandFailed)` and `errors.As(err, *exec.ExitError)`. For parser failure, preserve the concrete parser error and construct `ParseErr = errors.Join(ErrInvalidDotlinkReport, parserErr)`; thus the same top-level failure supports `errors.Is(err, ErrInvalidDotlinkReport)` and `errors.As` (for example, `*json.SyntaxError`). Failed command plus invalid report sets both; each other table row sets only its applicable field.

Sanitization walks complete runes, escapes controls as complete tokens (such as `\x1b`), and appends only whole UTF-8 or escape tokens until escaped output is <=4096 bytes. `ExitCode` is nil without a meaningful process exit.

## Presentation Ownership

Renderer comparison uses `Source`, `AttemptedCandidate`, `CanonicalPath`, and ordered `Modules`, not formatted strings or `err.Error()`. An identical `BaseDiagnostic` and `BaseSnapshot` render once from `StepResult.BaseDiagnostic`. Different snapshots render both as `report base context` and `failure base context`; executable, runner, and command remain separately labeled execution facts.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/execution/types.go` | Modify | Failure transport, multi-unwrapper, result attachment. |
| `internal/execution/dotfiles_provider.go` | Modify | Canonical command and outcome/cause composition. |
| `internal/execution/dotlink_report.go` | Modify | Preserve concrete parser error for `ParseErr`. |
| `internal/execution/dotfiles_installer.go` | Modify | Report-plus-error transport; short messages. |
| `cmd/dbootstrap/render.go` | Modify | Structured execution output and semantic base comparison. |
| `internal/execution/dotfiles_provider_test.go` | Modify | Composition, identity, runner, and sanitizer structure tests. |
| `internal/execution/dotfiles_installer_test.go` | Modify | Valid failed report plus error transport. |
| `cmd/dbootstrap/render_test.go` | Modify | Base deduplication and labeled differing snapshots. |

## Testing Strategy

| Layer | RED tests | Approach |
|---|---|---|
| Structure | `TestLocalDotfilesProviderFailedCommandInvalidReportPreservesExecutionAndParserIdentity` asserts the *same returned error* `Is` both sentinels and `As` both `*exec.ExitError` and `*json.SyntaxError`; `TestDotfilesInstallerPreservesFailedReportAndExecutionError` asserts the same `StepResult.Err` and retained report. | Table-driven fake runner. |
| Structure | `TestLocalDotfilesProviderUsesCanonicalExecutable`, `TestLocalDotfilesProviderMissingRunnerDoesNotRun`, and bounded Unicode/control stderr cases. | Request/call assertions. |
| Presentation | `TestRenderLinkDetailsDeduplicatesIdenticalBaseSnapshots` and `TestRenderLinkDetailsLabelsDifferentBaseSnapshots`. | Assert one primary render or both explicit labels; no terminal controls. |
| Regression | Existing success, default, dry-run, report validation, and base-identity tests. | Focused packages then `go test ./...`. |

## Threat Matrix

`CommandRunner` remains executable-plus-args, never shell. Applicable RED tests are missing-runner no-call, canonical executable, and escaped <=4096 stderr.

| Boundary | Applicability | Response |
|---|---|---|
| Documentation-like paths | N/A — no classification | No change. |
| Git repository selection | N/A — no VCS | No selector. |
| Commit state | N/A — no commits | No mutation. |
| Push state | N/A — no pushes | No ref resolution. |
| PR commands | N/A — no automation | No composition. |

## Migration / Rollout

No migration required; rollback reverts this isolated in-memory transport change.

## Open Questions

None.
