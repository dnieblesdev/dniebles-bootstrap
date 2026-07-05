# Proposal: Installer Provider Interfaces

## Intent

Introduce execution-layer contracts so future apply work has a safe boundary for installers, runners, and dotfiles operations without changing pure planning or mutating the host.

## Scope

### In Scope
- Add `internal/execution/` with `Installer`, `Runner`, and `DotfilesProvider` contracts.
- Add execution-only status, result, and report types distinct from planning statuses.
- Add noop stubs that return safe `not_implemented` results without mutation.
- Add table-driven tests for contracts, noop behavior, and kind-based runner dispatch.

### Out of Scope
- No `apply` command or CLI wiring.
- No real command execution, host mutation, installers, dotlink invocation, clone, sparse checkout, concurrency, or retry behavior.
- No changes to `internal/planning` production code or planning purity.

## Capabilities

### New Capabilities
- `execution-contracts`: Execution boundary for installer dispatch, dotfiles execution-provider operations, execution statuses, results, reports, and safe noop behavior.

### Modified Capabilities
- None.

## Approach

Create a single `internal/execution/` package. The runner sequentially accepts a planning `Plan`, dispatches planned steps to kind-scoped installers, and returns an execution report. Status vocabulary remains separate from `planning.PlanStepStatus`. `DotfilesProvider` is a high-level execution contract, not a replacement for the existing read-only dotfiles detector.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/execution/` | New | Contracts, execution result/report types, noop stubs, and tests. |
| `internal/planning/` | Unchanged | Referenced as input data only; production code remains pure. |
| `internal/dotfiles/` | Unchanged | Existing detector remains read-only and separate. |
| `cmd/dbootstrap/` | Unchanged | Apply command wiring is deferred. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Planning/execution status confusion | Med | Use separate Go types and document execution semantics in tests/specs. |
| Runner API overfitting future execution | Med | Keep the first runner sequential and minimal; defer options for retries/concurrency. |
| Dotfiles provider scope creep | Med | Keep provider operations high-level and noop-only in this slice. |
| No end-to-end execution proof | Low | Expected for contracts-only work; table tests prove dispatch and safety. |

## Rollback Plan

Remove `internal/execution/` and its tests. No migrations or production call-site reversions are needed because this slice does not wire execution into CLI or planning.

## Dependencies

- Existing `internal/planning` domain types.
- No third-party dependencies.

## Success Criteria

- [x] `internal/execution` compiles with contracts and noop stubs.
- [x] Tests prove noop results are non-mutating and report `not_implemented`.
- [x] Tests prove runner dispatches sequentially by resource kind.
- [x] Planning production code remains unchanged.
