# Tasks: Brew Package Installer

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 220-320 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|------|
| 1 | Add isolated brew installer component | PR 1 | Base on main; include component tests and zero-call cases. |
| 2 | Lock regression around unwired CLI | PR 1 | Same PR; verify `cmd/dbootstrap/main.go` and apply stay noop. |

## Phase 1: Foundation / Contracts

- [x] 1.1 Add `internal/execution/homebrew_installer.go` with `HomebrewInstaller`, constructor, and `SupportedKind()`.
- [x] 1.2 Define validation helpers for `Install.Provider == "brew"` and non-empty `Install.Package`.

## Phase 2: Core Implementation

- [x] 2.1 Implement `Install(context.Context, planning.PlanStep)` to return structured non-success for unsupported provider, missing package, or missing `brew`.
- [x] 2.2 Build only `CommandRequest{Executable:"brew", Args:[]string{"install", package}}` and map command success/failure to installer results.
- [x] 2.3 Keep execution behind injected `CommandRunner` and `CommandExists`; do not use shell strings, pipelines, or raw command metadata.

## Phase 3: Testing / Verification

- [x] 3.1 Add `internal/execution/homebrew_installer_test.go` coverage for success and exact argv shape with a fake runner.
- [x] 3.2 Add tests for command failure, unsupported provider, missing metadata/package, and missing `brew` with zero runner calls.
- [x] 3.3 Add regression coverage that `cmd/dbootstrap/main.go` and apply remain unwired/noop for brew installation.

## Phase 4: Cleanup / Documentation

- [x] 4.1 Confirm file-level comments and names describe the isolated installer behavior without promising CLI wiring.
- [x] 4.2 Remove any temporary test helpers or debug scaffolding used to prove no shell or host brew invocation.
