# Proposal: Brew Package Installer

## Intent

Add the first real installer component: a Homebrew package installer that assumes `brew` is already installed, consumes structured catalog install metadata, and executes only explicit executable-plus-args commands through `CommandRunner`.

## Scope

### In Scope
- Implement a testable Homebrew installer component for brew-backed tool/package plan steps.
- Use `Install.Provider == "brew"` and `Install.Package` to build `brew install <package>` requests.
- Return structured execution results for success, failure, unsupported metadata, missing package, and missing `brew` cases.

### Out of Scope
- Installing Homebrew, remote scripts, raw command fields, `sh -c`, pipelines, or dotfiles execution.
- Wiring the installer into `dbootstrap apply`; apply remains noop/non-mutating for this slice.
- Broad multi-provider support beyond Homebrew.

## Capabilities

### New Capabilities
- `brew-package-installer`: Executes Homebrew package installs from structured metadata through the command-runner seam.

### Modified Capabilities
- None.

## Approach

Create an `internal/execution` installer that supports the relevant package/tool resource kind(s), validates brew metadata before execution, checks/handles missing `brew` as a structured attention/failure result, and delegates mutation to injected `CommandRunner` using `Executable: "brew"`, `Args: ["install", package]`. Cover behavior with component/unit tests using fake command runners and command-existence seams. Keep CLI composition on `NoopForKind` until a future `wire-real-apply` slice.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/execution/` | New | Homebrew installer and tests. |
| `openspec/changes/brew-package-installer/` | New | Proposal and future delta specs/design/tasks. |
| `cmd/dbootstrap/main.go` | Unchanged | No real apply wiring in this slice. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Accidental CLI mutation | Medium | Keep installer unregistered from apply. |
| Shell-first regression | Low | Build only `CommandRequest{Executable, Args}` and test it. |
| Missing brew semantics conflict with bootstrap guidance | Medium | Return structured failure/attention; bootstrap reporter remains advisory. |

## Rollback Plan

Delete the new installer, tests, and change artifacts. Since the component is not wired into apply, rollback has no user-visible CLI behavior change.

## Dependencies

- Existing `CommandRunner`, structured install metadata, execution result types, and Homebrew bootstrap reporting.

## Success Criteria

- [ ] Brew-backed resources produce `brew install <package>` command requests via a fake runner.
- [ ] Missing/unsupported metadata and missing `brew` return structured non-success results.
- [ ] `dbootstrap apply` remains noop/non-mutating after this slice.
