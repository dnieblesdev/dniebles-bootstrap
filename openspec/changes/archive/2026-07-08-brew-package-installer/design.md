# Design: Brew Package Installer

## Technical Approach

Add an isolated Homebrew installer in `internal/execution` that implements the existing `Installer` contract but is not registered by `cmd/dbootstrap/main.go`. The component validates structured `planning.InstallMetadata`, checks for `brew` through the existing `CommandExists` seam, and delegates the only mutating action to an injected `CommandRunner` with `CommandRequest{Executable:"brew", Args:[]string{"install", package}}`. This satisfies the `brew-package-installer` delta while preserving the active apply safety spec: `apply` remains noop, non-mutating, and unwired.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|----------|--------|-------------------------|-----------|
| Installer shape | Create `HomebrewInstaller` implementing `Installer` for one selected `planning.ResourceKind` at construction time, with helpers for tool/package instances. | Broaden `Runner` dispatch by provider or register in CLI now. | Existing `Runner` dispatches by resource kind; provider-wide dispatch belongs in a future wiring slice. |
| Execution seam | Inject `CommandRunner` and `CommandExists`; never call `exec.Command`, shell, or `BrewCommandExists` directly in tests. | Use `OSCommandRunner` or host `brew` in component tests. | Keeps the component testable and proves no real brew invocation occurs. |
| Missing brew | Return `StepStatusFailed` with a clear missing-brew message and do not run the command. | Add `ManualAction` from installer or trigger bootstrap reporter. | The bootstrap provider owns advisory guidance; this installer returns a structured non-success execution result. |
| CLI wiring | Do not modify `cmd/dbootstrap/main.go`; keep `NoopForKind` registrations. | Register the installer behind `--yes`. | The active apply specs require no real mutation in this slice. |

## Data Flow

```text
planning.PlanStep
  └─ HomebrewInstaller.Install(ctx, step)
       ├─ validate step.Resource.Install provider/package
       ├─ exists("brew")
       └─ runner.RunCommand(ctx, CommandRequest{Executable:"brew", Args:["install", package]})
            └─ map CommandResult → StepResult
```

If metadata is unsupported/incomplete or `brew` is missing, the flow stops before `RunCommand`.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/execution/homebrew_installer.go` | Create | Defines `HomebrewInstaller`, constructor(s), validation, presence check, command construction, and command-result mapping. |
| `internal/execution/homebrew_installer_test.go` | Create | Uses fake `CommandRunner` and fake `CommandExists` to cover success, command failure, unsupported metadata, missing package, missing brew, and exact argv shape. |
| `cmd/dbootstrap/main.go` | Unchanged | Must continue registering only `NoopForKind(...)` installers. |
| `openspec/changes/brew-package-installer/design.md` | Create | This design artifact. |

## Interfaces / Contracts

```go
type HomebrewInstaller struct {
    kind   planning.ResourceKind
    runner CommandRunner
    exists CommandExists
}

func NewHomebrewInstaller(kind planning.ResourceKind, runner CommandRunner, exists CommandExists) *HomebrewInstaller
func (i *HomebrewInstaller) SupportedKind() planning.ResourceKind
func (i *HomebrewInstaller) Install(context.Context, planning.PlanStep) StepResult
```

Contract details:
- Accept only `step.Resource.Install.Provider == "brew"` and non-empty `Package`.
- Check only `exists("brew")` before execution.
- Build only `CommandRequest{Executable:"brew", Args:[]string{"install", package}}`.
- Map `CommandStatusSucceeded` to `StepStatusInstalled`; all other command statuses map to `StepStatusFailed` with command outcome summarized in `Message`/`Err`.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Metadata validation and missing brew | Fake presence seam; assert fake runner has zero calls. |
| Unit | Explicit command construction | Fake runner records request; assert executable `brew` and args exactly `install`, package. |
| Unit | Command success/failure mapping | Fake runner returns `CommandResult`; assert `StepResult` status/message/error. |
| Regression | CLI remains unwired | Existing apply tests should continue proving noop results; no real installer registration. |
| E2E | None for this slice | No real brew invocation or host mutation is allowed. |

## Migration / Rollout

No migration required. Rollout is component-only; a later SDD slice may wire provider/kind dispatch into apply behind explicit safety gates.

## Open Questions

None.
