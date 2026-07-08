# Verification Report: wire-brew-apply

## Verdict

PASS

`wire-brew-apply` satisfies the proposal, all four delta specs, design constraints, and completed tasks. Runtime tests passed without invoking real Homebrew from apply tests. Re-verification after non-behavioral comment cleanup in `internal/execution/homebrew_installer.go` confirms the stale comment is resolved and no behavior changed. The only validation gap is environmental: the `openspec` CLI is not installed in this shell, so `openspec validate wire-brew-apply --strict` could not run.

## Mode

| Field | Value |
|-------|-------|
| Verification mode | Standard SDD verify |
| Strict TDD | Inactive |
| Artifact store | Hybrid: OpenSpec + Engram |
| Change | `wire-brew-apply` |
| Scope reviewed | Proposal, four delta specs, design, tasks, Engram apply progress, implementation, tests, comment-cleanup diff |

## Completeness

| Artifact | Status | Evidence |
|----------|--------|----------|
| Proposal | PASS | `openspec/changes/wire-brew-apply/proposal.md` reviewed. |
| Specs | PASS | Reviewed deltas for `apply-command-dry-run`, `brew-package-installer`, `execution-contracts`, and `homebrew-bootstrap-provider`. |
| Design | PASS | `design.md` reviewed against implementation. |
| Tasks | PASS | All tasks 1.1 through 4.2 are checked. |
| Engram apply progress | PASS | Observation `#2275`, topic `sdd/wire-brew-apply/apply-progress`, found and reviewed. |

## Execution Evidence

| Command | Result | Notes |
|---------|--------|-------|
| `go test ./...` | PASS | Initial full suite passed; output was cached. |
| `go test ./...` | PASS | Re-run after comment cleanup; output was cached. |
| `go test -count=1 ./...` | PASS | Full uncached suite passed. |
| `go test -count=1 ./cmd/dbootstrap -run 'TestRunApply|TestRenderExecutionReport'` | PASS | Focused CLI/render apply tests passed. |
| `go test -count=1 ./internal/execution -run 'TestBrewOnlyInstaller|TestAppendHomebrewBootstrap|TestHomebrewInstaller'` | PASS | Focused execution/bootstrap tests passed. |
| `go vet ./...` | PASS | No diagnostics. |
| `test -z "$(gofmt -l .)"` | PASS | No formatting drift. |
| `openspec validate wire-brew-apply --strict` | SKIPPED | `openspec` command not found in this environment. |

## Behavioral Compliance Matrix

| Spec | Scenario coverage | Status |
|------|-------------------|--------|
| `apply-command-dry-run` | Default and dry-run non-mutating, `--yes` confirmed mode, mode-specific reporting, conflicting flags, no orchestration features. Covered by `TestRunApplyCommand`, `TestRunApplySafeModesDoNotInstantiateRealExecution`, `TestRunApplyConfirmedBrewPresentUsesInjectedRunnerForBrewOnly`, `TestRunApplyConfirmedMissingBrewDoesNotInstantiateHomebrewInstaller`, render tests, and regression tests. | PASS |
| `brew-package-installer` | Brew metadata gating, explicit `CommandRequest{Executable:"brew", Args:["install", package]}`, shell rejection, missing brew, command success/failure. Covered by existing `HomebrewInstaller` tests plus new CLI confirmed/missing-brew tests and `BrewOnlyInstaller` tests. | PASS |
| `execution-contracts` | Default noop contracts, confirmed brew-only execution, no side effects outside confirmed brew steps, advisory bootstrap data, runner sequential/kind dispatch. Covered by CLI apply tests, provider-aware adapter tests, bootstrap tests, and existing runner tests. | PASS |
| `homebrew-bootstrap-provider` | Missing brew detection, brew-present branch, confirmed apply stop, no package install on missing brew, official/manual advisory guidance, no copy-paste remote script. Covered by `TestRunApplyHomebrewBootstrap`, `TestRunApplyConfirmedMissingBrewDoesNotInstantiateHomebrewInstaller`, `TestAppendHomebrewBootstrap*`, and render tests. | PASS |

## Correctness Findings

| Criterion | Status | Evidence |
|-----------|--------|----------|
| `apply --yes` only may execute brew installs for brew-backed tool/package steps | PASS | `buildApplyRunner` only constructs `newOSCommandRunner` and `newHomebrewInstaller` in `applyModeConfirmed`; tool/package paths are wrapped with `BrewOnlyInstaller`; runtime/dotfile stay noop. |
| Default and `--dry-run` avoid real execution construction | PASS | `mode != applyModeConfirmed` returns `newNoopApplyRunner`; `TestRunApplySafeModesDoNotInstantiateRealExecution` fails if OS runner/Homebrew installer factories are called. |
| Missing brew under `--yes` is advisory-first and skips target install | PASS | Missing-brew branch returns `missingHomebrewInstaller` skipped results and `AppendHomebrewBootstrap` guidance; test fails if OS runner/Homebrew installer is instantiated. |
| Bootstrap guidance is official/manual and non-executable | PASS | `AppendHomebrewBootstrap` renders official Homebrew URL/manual review wording only; tests forbid `/bin/bash`, `curl`, `sh -c`, pipes, and `install.sh`. |
| Provider-aware adapter returns `not_implemented` for non-brew/no metadata | PASS | `BrewOnlyInstaller` returns `StepStatusNotImplemented` before delegation; tests prove delegate is not called. |
| No raw shell, remote script, dotfiles execution, bootstrap entrypoint, retry/concurrency/clone/sparse checkout introduced | PASS | Source inspection and static checks found no such behavior in apply wiring. Dotfiles remain `NoopForKind`. |
| Tests avoid real brew invocation | PASS | CLI tests use fake factories/recording runners and stub `brewCommandExists`; package tests use fake runners. |
| Comment cleanup is non-behavioral | PASS | `git diff -- internal/execution/homebrew_installer.go` shows only the stale `HomebrewInstaller` comment changed; implementation code is unchanged. |

## Design Coherence

| Design decision | Status | Evidence |
|-----------------|--------|----------|
| Centralize safety gate by `applyMode` | PASS | `buildApplyRunner` branches by mode before constructing real execution seams. |
| Use provider-aware adapter instead of direct broad Homebrew registration | PASS | `BrewOnlyInstaller` delegates only `Install.Provider == "brew"`; unsupported/missing metadata is `not_implemented`. |
| Reuse `OSCommandRunner`/`BrewCommandExists` without shell paths | PASS | Confirmed path uses `newOSCommandRunner`, `newHomebrewInstaller`, and explicit command requests. |
| Check missing brew before real installers | PASS | `brewCommandExists("brew")` is checked before OS runner/Homebrew installer construction. |
| Replace executable bootstrap one-liner | PASS | Guidance now points to `https://brew.sh/` and manual review wording only. |

## Issues

### CRITICAL

None.

### WARNING

- `openspec validate wire-brew-apply --strict` could not run because `openspec` is not installed in the verification shell.

### SUGGESTION

None.

## Final Verdict

PASS. The implementation is archive-ready from the code/test/spec verification perspective.
