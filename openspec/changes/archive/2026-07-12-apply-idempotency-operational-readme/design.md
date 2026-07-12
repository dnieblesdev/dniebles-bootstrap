# Design: Apply Idempotency and Operational README

## Decision

Carry the planning status on each ordered `planning.PlanStep`, then let `execution.Runner.Run` make the only mutation decision from that status. `already_installed` is an execution skip **only** when the step is a `tool` or `runtime`; those are the only kinds that `internal/state.Detector` can mark present, and it will do so only after a configured command-presence lookup succeeds. This preserves the existing runner as the single sequential reporting seam used by both `apply` and `bootstrap`.

No new package, persistence layer, installer interface, command flag, or catalog schema is needed.

## Anchored seams and file changes

| File | Symbol / change |
|---|---|
| `internal/planning/types.go` | Add `Status PlanStepStatus` to `PlanStep`; update its comment to state that it carries the planning-time status for the ordered executable step. |
| `internal/planning/builder.go` | In `planBuilder.appendOrderedSteps`, assign the already-computed `status` to both the `PlanStep` and its `PlanStepResult`. This maintains one planning decision and avoids execution-side inference. |
| `internal/state/detector.go` | In `Detector.Detect`, inspect only `tool` and `runtime` resources with non-nil command-presence metadata and call the injected/default `LookPath` with `resource.Presence.Name`, never `ref.Name`. Missing/unsupported presence metadata remains absent; no fallback to ID, package, config, or host/provider probe is introduced. |
| `internal/execution/runner.go` | Before installer lookup/dispatch, translate an eligible `already_installed` step to `StepStatusSkipped` with an explicit `"already installed; no mutation attempted"` message and append it. All other steps retain the current installer lookup and dispatch path. Keep the existing one-result-per-step loop. |
| `internal/state/detector_test.go` | Add configured-name-differs-from-ID and missing-presence tests; retain package/dotfile exclusion coverage. |
| `internal/planning/builder_test.go` | Assert that `PlanStep.Status` matches the existing planning result status for command-present tool/runtime fixtures. |
| `internal/execution/runner_test.go` | Add table-driven guard tests: eligible present tool/runtime is skipped without installer calls; planned/attention/failed/unsupported equivalents still follow current dispatch; `already_installed` package and dotfile steps are not skipped. Add a mixed-plan order/continued-failure case. |
| `cmd/dbootstrap/main_test.go` | Add confirmed `apply` and `bootstrap` integration cases with injected installation state and recording runner: a detected tool/runtime is unchanged, includes no-mutation wording, and produces no installer command; an absent eligible resource still runs; unsupported and failure results remain ordered and preserve exit behavior. Retain safe-mode and alias-parity tests, adding assertions that default/dry-run do not claim a confirmed skip. |
| `README.md` | Replace stale “minimal plan command/no execution” status and CLI material with an operational workflow aligned to the existing command parser and renderer. |

## Data and result flow

1. `cmd/dbootstrap.buildPlan` composes read-only `state.Detect` output with catalog, environment, and configuration facts.
2. `state.Detector` checks `Resource.Presence.Name` through its injected `PathLookup` only for tool/runtime command-presence metadata. A successful lookup adds that resource ref to `InstallationState.PresentResources`.
3. `planBuilder.appendOrderedSteps` computes the status once. Present refs become `PlanStepStatusAlreadyInstalled`; it records the same value in `PlanStep.Status` and `PlanStepResult.Status` while retaining topological plan order.
4. `runApplyLike` passes that plan unchanged to `buildApplyRunner` and `Runner.Run`; therefore `apply` and `bootstrap` share behavior.
5. `Runner.Run` iterates the plan once, in slice order. It appends a skipped/unchanged result directly for only `(tool|runtime, already_installed)`; it does not select an installer or call a command runner for that step. It otherwise follows the existing installer registry path, including unsupported, failed, and continued execution behavior.
6. `renderExecutionReport` already maps `StepStatusSkipped` to `unchanged`, counts results in report order, and emits each result in order. The skip message provides the explicit no-mutation statement without renderer API changes.
7. `appendApplyBootstrap` remains advisory-only and continues after report creation; it neither acquires Homebrew nor changes step results.

### Status-to-execution mapping

| Planning step status / kind | Execution action | Result category |
|---|---|---|
| `already_installed` + tool or runtime | Do not resolve/dispatch installer; append direct result | `skipped` → `unchanged`, message explicitly says no mutation attempted |
| `already_installed` + package or dotfile | Do not use this guard; use existing dispatch behavior | Existing result |
| `planned`, `attention_required`, or other non-matching status | Use existing installer lookup and dispatch | Existing installed/skipped/not-implemented/failed result |
| Missing installer | Preserve existing direct unsupported result | `not_implemented` → `not supported yet` |
| Installer failure | Preserve result and continue later steps | `failed`; confirmed mode remains non-zero |

The kind check is intentional defense in depth: status is necessary but not sufficient. Reliable command presence is established exclusively by the detector's constrained tool/runtime lookup, not by package metadata, dotfile state, configuration state, or installer probing.

## API impact

`planning.PlanStep` gains an exported `Status planning.PlanStepStatus` field. No function signature changes are required: `execution.Runner.Run(context.Context, planning.Plan)` consumes the status already embedded in the ordered plan. Existing manually constructed plan steps retain the zero value and therefore follow normal dispatch.

## Preserve-order strategy

Do not filter the plan before execution and do not post-merge skipped results. `Runner.Run` appends exactly one result during each iteration of `plan.Steps`; the direct skip occupies the same iteration where installer dispatch would otherwise occur. This preserves the original plan order for every execution mode and cannot let one status rewrite another result.

## Strict-TDD test sequence

1. Add failing `internal/state` table cases for configured `Presence.Name` different from the resource ID, and for absent/missing presence metadata not being guessed. Run `go test ./internal/state`.
2. Add failing `internal/planning` assertions that the ordered step carries its computed status. Run `go test ./internal/planning`.
3. Add failing `internal/execution` runner tests for direct eligible skips, excluded kinds, mixed ordering, and continued failure. Run `go test ./internal/execution`.
4. Add failing command-level `apply`/`bootstrap` cases using existing injected state/factory seams and the recording command runner. Prove the confirmed call count excludes the detected step, output contains `unchanged` and `no mutation attempted`, and existing dry-run/default/alias/failure behavior remains. Run `go test ./cmd/dbootstrap`.
5. Implement the smallest changes required to make each focused package green, format Go files, then run the mandated complete suite: `go test ./...`.

## README structure

Use an outcome-first operational section after the project summary:

1. **Quick path:** inspect with `plan`, then use default/dry-run reporting, then explicitly confirm with `apply --yes`; show profile/resource/catalog targeting.
2. **Commands and safety modes:** compact table for `plan`, `apply`, and `bootstrap`; describe default non-mutation, `--dry-run`, `--yes`, `--sudo`, and validation (`--dry-run` conflicts with `--yes`; `--sudo` requires `--yes`).
3. **Confirmed reruns:** explain the narrow command-presence/`already_installed` guard, unchanged result, explicit no mutation, and shared bootstrap semantics.
4. **Results and recovery:** table for changed/unchanged/not supported yet/failed, ordered mixed results, continued processing, confirmed non-zero failure, and fix-then-deliberately-rerun recovery.
5. **Limits and advisory bootstrap:** command presence is not package/version/configuration/link convergence proof; no automatic retry, rollback, clone/fetch, or provider acquisition. Missing-provider bootstrap output is manual advice only.

Keep existing broader architecture and dotfiles boundary material where still accurate; remove claims that execution commands or runners do not exist.

## Non-goals

- No package-manager presence, package/version/configuration reconciliation, retries, receipts, rollback, or host/provider probing by installers.
- No skipping package or dotfile steps through this guard; dotfile module presence never establishes link convergence.
- No catalog schema/default-profile/bootstrap-acquisition/shell-wrapper change.
- No changed flag semantics, target validation, rendering categories, continued-execution policy, or command aliases beyond the explicitly described status guard.

## Risks and rollout

The key risk is over-skipping when resource identifiers diverge from command names. The configured-name test and the status-plus-kind guard constrain the change. Roll out as one small change with focused tests first and `go test ./...` last; revert the `PlanStep` status propagation, detector lookup, runner guard, and README together if behavior must be restored.
