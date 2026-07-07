# Proposal: Command Runner

## Intent

Introduce a reusable internal process execution primitive for future installers without making catalog metadata shell-first. The runner should execute explicit executable-plus-args commands, capture outcomes, and keep `apply` non-mutating until later installer slices wire real behavior.

## Scope

### In Scope
- Add an internal command/process runner abstraction for infrastructure adapters and installers.
- Support argv-style executable and args, optional cwd/env, context/timeout handling, stdout/stderr capture, and exit status/code reporting.
- Provide deterministic dry-run/no-op behavior and tests using fake commands or safe helpers.
- Preserve existing catalog install/presence metadata as structured inert data.

### Out of Scope
- Real installers, installer dispatch wiring, or apply mutation.
- Raw shell command fields in catalog metadata.
- Dotfiles execution, bootstrap entrypoint, retries/concurrency, or `curl | sh`/shell pipeline support.

## Capabilities

### New Capabilities
- `command-runner`: Controlled process execution for future infrastructure adapters using executable-plus-args, context-aware execution, captured output, and structured command results.

### Modified Capabilities
- None

## Approach

Add a small `internal/execution` command runner contract and default implementation around OS process execution. Make shell interpretation opt-in/out of scope by default: callers provide executable and args, not `sh -c`. Model results as data: command, status, exit code, stdout, stderr, duration/error where useful. Tests should avoid unsafe host mutation by using fake executors or local test helper processes.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/execution/` | Modified/New | Add command runner types, implementation, and tests near existing execution contracts. |
| `openspec/specs/execution-contracts/spec.md` | Referenced | Existing execution boundary remains unchanged; `apply` stays noop-only until a future slice wires installers. |
| `openspec/specs/command-runner/spec.md` | New | Define command runner requirements and safety semantics. |
| `internal/planning/types.go` | Unchanged | Catalog install/presence metadata remains structured inert provider data. |
| `catalog/bootstrap.toml` | Unchanged | No raw command fields are added. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Runner becomes shell-first config by accident | Med | Specify executable-plus-args only; no default `sh -c`; no catalog command fields. |
| Tests depend on host tools | Med | Use fake executor seams or safe test helper processes. |
| Apply appears to mutate | Low | Do not wire runner into installer dispatch in this slice. |

## Rollback Plan

Delete the new command-runner code/tests and remove the OpenSpec delta. Existing noop execution, planning, catalog metadata, and CLI apply behavior should remain unchanged.

## Dependencies

- Existing `internal/execution` contracts and noop apply bridge.
- Completed `catalog-installer-metadata` slice.

## Success Criteria

- [ ] Command execution is represented as explicit executable-plus-args, not shell strings.
- [ ] Results include stdout/stderr and exit status/code.
- [ ] Context/timeout cancellation is covered.
- [ ] Dry-run/no-op behavior is deterministic.
- [ ] Catalog metadata remains structured provider/presence data only.
