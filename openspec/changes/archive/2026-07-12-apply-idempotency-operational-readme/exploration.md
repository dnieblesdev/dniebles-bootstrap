# Exploration: apply-idempotency-operational-readme

## Current state

The repository has a pure planning core and a CLI composition root. The relevant flow is:

```text
runApplyLike -> parseApplyLikeFlags -> buildPlan
             -> state/config/dotfiles detection
             -> planning.BuildPlan
             -> buildApplyRunner -> execution.Runner.Run
             -> appendApplyBootstrap -> renderExecutionReport
```

`apply` and `bootstrap` already share `runApplyLike`; `--yes` is the only confirmed mode, `--sudo` requires `--yes`, and default/`--dry-run` paths compose no-op installers. Confirmed execution is limited to Homebrew-backed tool/package steps, Linux APT-backed tool/package steps, and selected dotfile steps. Existing reports distinguish internal statuses (`installed`, `skipped`, `not_implemented`, `failed`) from user-facing categories (`changed`, `unchanged`, `not supported yet`, `failed`).

## Idempotency findings

- `internal/planning/builder.go` already gives `already_installed` precedence after environment matching and keeps the resource in the plan. Missing required config does not override that status.
- `internal/state/detector.go` currently detects only `tool` and `runtime` resources and uses `ref.Name` for PATH lookup. It does not use `Resource.Presence.Name`; this is a correctness gap for catalog entries whose command differs from the resource ID.
- Package presence is intentionally not detected. `catalog/bootstrap.toml` has package presence metadata, but the current detector ignores packages. Adding package-manager queries or a broader presence model would exceed the safe slice.
- `internal/dotfiles/detector.go` can mark a dotfile module present based on local module state, but `execution.DotfilesInstaller` can still invoke dotlink when the plan step is already installed. A module directory is not proof that links are current, so this must not be generalized into full dotfile convergence.
- `internal/execution.Runner` executes every `PlanStep` sequentially and has no access to planning statuses. Therefore, merely reporting `already_installed` does not prevent a confirmed installer call.
- `internal/execution/homebrew_installer.go` and `apt_installer.go` always issue install commands when invoked. Neither installer currently performs a second presence check.

## Minimal safe implementation seam

Keep planning unchanged as the source of `already_installed`. At the CLI/application boundary, carry the planning statuses into apply execution and prevent installer dispatch for steps whose detected planning result is `already_installed`. Preserve an execution result for those steps as unchanged/skipped with explicit user-facing text such as `already installed; no mutation attempted`, so reports remain complete and ordered. Execute all other steps through the existing Runner and preserve continued execution, confirmed failure exit behavior, and no-op behavior.

This should be limited to detected state that the existing detector can safely establish (currently command-presence-backed tools/runtimes, plus any explicitly reliable dotfile detector result only if the active spec confirms it). Do not add package-manager detection, version reconciliation, persistent receipts, retries, or link-content convergence. Correcting the detector to honor `Presence.Name` is a small compatible fix and should be covered because it determines whether the idempotency guard can identify the installed command; otherwise the existing `git`-style fixture can hide the defect.

The implementation should avoid teaching individual installers about planning state. A small execution/apply orchestration helper or filtered execution plan plus synthetic unchanged results is preferable to adding host probes to `internal/execution`. It must retain plan order and not treat unsupported or failed prior outcomes as already installed.

## Documentation findings

`README.md` describes only the current `plan` command and still says runtime execution/apply/install commands are outside the implemented slice, which is stale relative to `cmd/dbootstrap` and the current OpenSpec contracts. It mentions a first-run bootstrap concept but does not explain the actual `bootstrap` command, its explicit target requirement, or safety flags.

The operational README should lead with a quick path and then define:

- `dbootstrap plan`: read-only planning and state inspection.
- `dbootstrap apply`: the shared execution workflow; requires `--profile` or `--resource` and is non-mutating without `--yes`.
- `dbootstrap bootstrap`: thin alias over the same apply workflow, not an implicit default-profile or acquisition flow.
- `--yes`: explicit confirmation boundary; only eligible brew, Linux APT, and selected dotfile work may mutate; unsupported resources remain unsupported.
- `--sudo`: valid only with `--yes`; enables sudo-backed Linux APT and may require an interactive credential-capable environment; it does not make other resources executable.
- `--dry-run`: explicit non-mutating mode and mutually exclusive with `--yes`.
- “already installed”: a detected state result, not a version or configuration guarantee. On a confirmed re-run, detected resources are reported unchanged and no redundant installer mutation is attempted. Package presence is not currently detected; dotfile module presence does not prove links are current.
- Recovery/partial execution: confirmed apply continues through steps, can partially succeed, reports failures, and should be rerun after prerequisites are fixed. No automatic sudo, acquisition, retry, or rollback should be implied.

Use a compact command table and examples, followed by boundaries/limitations. Keep dotfiles ownership external and do not document shell-wrapper orchestration beyond acquiring/launching the binary.

## Acceptance-test seams

- `cmd/dbootstrap/main_test.go`: existing seams stub environment, installation, config, and dotfiles detectors; command existence, installer factories, and command runners are injectable. Add a confirmed re-run/present-state test that proves zero command-runner calls for a detected eligible resource while retaining an `unchanged` report result.
- Add coverage for a non-default `Presence.Name` to prove state detection probes the catalog presence target rather than the resource ID.
- Cover mixed plans: detected already-installed step is unchanged, absent eligible step still invokes the installer, unsupported runtime remains not supported, and order is preserved.
- Preserve existing parser tests for `--yes`, `--sudo`, `--dry-run`, target requirements, and alias parity.
- README acceptance can be review-oriented: commands/flags and “already installed” language must match actual output and the active apply/execution specs; do not add a code-generated CLI reference unless the repository introduces one.
- Run focused tests first and the configured full runner `go test ./...`; strict TDD is active.

## Scope boundaries and risks

- Do not redesign installation detection for packages or add version checks.
- Do not claim that a successful prior apply is globally idempotent; only avoid mutation where current detection proves presence.
- Do not skip a step solely because it has a missing-config attention reason; existing `already_installed` precedence is the relevant condition.
- Keep advisory Homebrew bootstrap reporting non-mutating.
- The main risk is conflating a resource ID, presence target, and package name. The detector and tests should make those contracts explicit before applying the execution guard.

## Recommendation

Implement one narrow apply orchestration change that honors `planning.PlanStepStatusAlreadyInstalled` during confirmed execution, fixes/locks the `Presence.Name` lookup contract, and adds deterministic tests for no redundant command invocation and mixed-plan reporting. Update only the operational sections of `README.md` to reflect the existing command surface and precise safety/idempotency limits. Do not change catalog schema, installer provider behavior, package detection, dotfiles semantics, or bootstrap acquisition.
