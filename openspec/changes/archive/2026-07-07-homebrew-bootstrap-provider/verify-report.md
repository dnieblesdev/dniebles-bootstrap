# Verify Report: Homebrew Bootstrap Provider

**Change**: homebrew-bootstrap-provider
**Version**: N/A (delta spec set)
**Mode**: Standard (Strict TDD not active per `sdd-init/dniebles-bootstrap` baseline)
**Persistence**: Hybrid (OpenSpec + Engram)
**Delivery**: Single PR with maintainer-approved size exception

## Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 11 |
| Tasks complete | 11 |
| Tasks incomplete | 0 |

All tasks in `tasks.md` are marked `[x]`. No unchecked implementation or cleanup tasks remain.

## Build & Tests Execution

**Build / vet**: ✅ Passed
```text
$ go vet ./...
(no output — clean)

$ gofmt -l cmd/dbootstrap/ internal/execution/
(no output — all files formatted)
```

**Tests**: ✅ 43 subtests passed / 0 failed / 0 skipped (across `internal/execution` and `cmd/dbootstrap`); all 8 packages `ok`.
```text
$ go test -count=1 ./...
ok  github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap          0.007s
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml   0.003s
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/config         0.003s
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles       0.003s
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/environment    0.002s
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/execution      0.187s
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/planning       0.005s
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/state          0.083s
```

**Coverage**: ➖ Not available (no coverage threshold configured for this slice; tests are host-safe and exercise behavior at unit + CLI integration layers).

## Spec Compliance Matrix

### `homebrew-bootstrap-provider` delta

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| Homebrew bootstrap is detected as a provider-owned need | Missing brew is detected | `internal/execution/homebrew_bootstrap_test.go > TestAppendHomebrewBootstrapBrewMissing`; `cmd/dbootstrap/main_test.go > TestRunApplyHomebrewBootstrap/default apply reports manual bootstrap when brew is missing` | ✅ COMPLIANT |
| Homebrew bootstrap is detected as a provider-owned need | Brew present does not trigger bootstrap | `homebrew_bootstrap_test.go > TestAppendHomebrewBootstrapBrewPresent`; `main_test.go > TestRunApplyHomebrewBootstrap/brew present does not trigger bootstrap` | ✅ COMPLIANT |
| Bootstrap reporting provides explicit manual guidance | Bootstrap guidance is rendered | `cmd/dbootstrap/render_test.go > TestRenderExecutionReportRendersManualActions`; `main_test.go > TestRunApplyHomebrewBootstrap` (default/dry-run/yes) | ✅ COMPLIANT |
| Bootstrap reporting provides explicit manual guidance | Guidance remains non-executable | `homebrew_bootstrap_test.go > TestAppendHomebrewBootstrapDoesNotExecuteInstruction`; `regression_test.go > TestHomebrewBootstrapDoesNotUseCommandRunner` | ✅ COMPLIANT |

### `apply-command-dry-run` delta (MODIFIED)

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| Apply command exists with plan-style target flags | Apply accepts the same targets as plan | `main_test.go > TestRunApplyCommand` (profile + resource cases) | ✅ COMPLIANT |
| Apply command exists with plan-style target flags | Invalid target input is rejected | `main_test.go > TestRunApplyCommand` (malformed / unsupported / combined flags) | ✅ COMPLIANT |
| Apply command exists with plan-style target flags | Missing brew is reported without mutation | `main_test.go > TestRunApplyHomebrewBootstrap` (default, dry-run, yes — steps stay `not_implemented`) | ✅ COMPLIANT |
| Apply renders a noop execution report | Dry-run execution reports not_implemented | `main_test.go > TestRunApplyCommand/default apply` + `render_test.go > TestRenderExecutionReportIsDistinctFromPlanRendering` | ✅ COMPLIANT |
| Apply renders a noop execution report | Execution rendering is distinct from plan rendering | `render_test.go > TestRenderExecutionReportIsDistinctFromPlanRendering` | ✅ COMPLIANT |
| Apply renders a noop execution report | Bootstrap reporting does not become execution | `main_test.go > TestRunApplyHomebrewBootstrap` (manual action rendered while all step statuses remain `not_implemented`) | ✅ COMPLIANT |
| Apply mode is explicit and safe by default | Default apply is non-mutating | `main_test.go > TestRunApplyCommand/default apply` | ✅ COMPLIANT |
| Apply mode is explicit and safe by default | Dry-run is explicit non-mutating | `main_test.go > TestRunApplyCommand/explicit dry run` | ✅ COMPLIANT |
| Apply mode is explicit and safe by default | Bootstrap guidance remains non-mutating under yes | `main_test.go > TestRunApplyHomebrewBootstrap/yes mode reports manual bootstrap` | ✅ COMPLIANT |

### `execution-contracts` delta (MODIFIED)

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| Execution contracts remain non-mutating for apply | Apply uses noop execution contracts only | `main_test.go > TestRunApplyCommand`; `regression_test.go > TestNoopExecutionRemainsNonMutating` | ✅ COMPLIANT |
| Execution contracts remain non-mutating for apply | Side effects remain absent | `regression_test.go > TestApplyRemainsNoopOnlyAndUnwiredFromCommandRunner`; `TestHomebrewBootstrapDoesNotUseCommandRunner` | ✅ COMPLIANT |
| Execution contracts remain non-mutating for apply | Bootstrap data stays advisory | `render_test.go > TestRenderExecutionReportRendersManualActions` (manual action is text-only; no `CommandRequest`/`CommandRunner` wiring) | ✅ COMPLIANT |

**Compliance summary**: 16/16 scenarios compliant across all three delta specs.

## Correctness (Static Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| Missing brew reported as advisory/manual bootstrap action | ✅ Implemented | `AppendHomebrewBootstrap` appends a `ManualAction` with `ID "homebrew:bootstrap"` when `planNeedsHomebrew` and `exists("brew")==false`. |
| Brew present does not trigger bootstrap | ✅ Implemented | Early `return report` when `exists("brew")` is true; verified by `TestAppendHomebrewBootstrapBrewPresent`. |
| Official Homebrew command rendered only as text/manual instruction, never executed | ✅ Implemented | `homebrewInstallInstruction` is a string constant placed in `ManualAction.Instructions`; `TestHomebrewBootstrapDoesNotUseCommandRunner` forbids `CommandRunner`/`RunCommand`/`CommandRequest` references in the provider file. |
| All apply modes remain non-mutating, including `--yes` | ✅ Implemented | `runApply` uses `NoopForKind` installers for all kinds; `--yes` maps to `applyModeConfirmedFuture` (noop). `TestRunApplyHomebrewBootstrap` covers all three modes with brew missing — steps stay `not_implemented`. |
| No target package install | ✅ Implemented | No `Installer` implementation added; only noop installers wired in `main.go`. |
| No catalog raw command fields / no shell-first metadata | ✅ Implemented | Provider reads existing `Install.Provider == "brew"` only; `catalog/bootstrap.toml` unchanged (`git diff` empty). |
| No CommandRunner mutation / no real installer | ✅ Implemented | `main.go` still passes `TestApplyRemainsNoopOnlyAndUnwiredFromCommandRunner`; `homebrew_bootstrap.go` passes `TestHomebrewBootstrapDoesNotUseCommandRunner`. |
| No remote script execution | ✅ Implemented | Detection uses `exec.LookPath("brew")` only; install command is a string literal never fed to a runner. |
| No apply safety bypass | ✅ Implemented | `--dry-run` + `--yes` still rejected; mode selection unchanged. |
| No dotfiles execution / no bootstrap entrypoint | ✅ Implemented | Dotfiles provider untouched; no new CLI entrypoint added. |
| `catalog/bootstrap.toml` unchanged | ✅ Verified | `git diff --stat catalog/bootstrap.toml` produces no output. |
| Detection seam uses `exec.LookPath` only | ✅ Implemented | `BrewCommandExists` calls `exec.LookPath(name)` and returns `err == nil`; `TestAppendHomebrewBootstrapLookupOnlyBrew` asserts only `"brew"` is probed. |

## Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| Extend `ExecutionReport` with `ManualAction`/`ManualInstructions` | ✅ Yes | `types.go` adds `ManualAction{ID,Title,Reason,Instructions []string}` and `ManualActions []ManualAction` on `ExecutionReport` without changing `StepResult`. |
| Do NOT implement Homebrew as an `Installer` | ✅ Yes | No `Installer` for brew; `AppendHomebrewBootstrap` is a pure report enricher. |
| Use `CommandExists` seam backed by `exec.LookPath` only, not `CommandRunner` | ✅ Yes | `CommandExists func(name string) bool` + `BrewCommandExists`; regression test forbids `CommandRunner`/`CommandRequest` in provider file. |
| No raw command/catalog install fields | ✅ Yes | Only existing `Install.Provider` is inspected; catalog untouched. |
| Wiring: `AppendHomebrewBootstrap` after `Runner.Run`, before render | ✅ Yes | `main.go` line 127: `report = execution.AppendHomebrewBootstrap(report, result.Plan, brewCommandExists)`. `brewCommandExists` is an overridable package var for tests. |
| `catalog/bootstrap.toml` kept; scenarios via fixtures | ✅ Yes | `TestRunApplyHomebrewBootstrap` writes a temp catalog fixture with a `brew`-backed tool; catalog file unchanged. |

No design deviations detected. Apply-progress reported "None"; source inspection confirms.

## Issues Found

**CRITICAL**: None
**WARNING**: None
**SUGGESTION**: None

## Safety Verification (explicit criteria from verification request)

| Criterion | Result |
|-----------|--------|
| Missing brew reported as advisory/manual bootstrap action | ✅ Confirmed |
| Brew present does not trigger bootstrap guidance | ✅ Confirmed |
| Official Homebrew command rendered only as text, never executed | ✅ Confirmed |
| All apply modes non-mutating including `--yes` | ✅ Confirmed |
| No target package install | ✅ Confirmed |
| No catalog raw command fields / no shell-first metadata | ✅ Confirmed |
| No CommandRunner mutation / no real installer / no remote script execution / no apply safety bypass | ✅ Confirmed |
| No dotfiles execution / no bootstrap entrypoint | ✅ Confirmed |
| `catalog/bootstrap.toml` unchanged | ✅ Confirmed |
| Tests meaningful and host-safe | ✅ Confirmed (table-driven unit tests + CLI integration via stubs; no real shell/process execution; `t.TempDir()` for fixtures) |

## Workload / PR Boundary

- Mode: single PR with maintainer-approved size exception.
- Diff stat: 6 modified files (234 insertions / 11 deletions) + 2 new files (198 lines) ≈ 245 added lines. Within the accepted size exception.
- `gofmt` clean; `go vet` clean.

## Verdict

**PASS**

All 11 tasks complete; all 16 scenarios across the three delta specs are covered by passing runtime tests; design decisions are followed with no deviations; `go test ./...`, `go vet ./...`, and `gofmt` are clean; `catalog/bootstrap.toml` is untouched; and every explicit safety criterion in the verification request is satisfied.
