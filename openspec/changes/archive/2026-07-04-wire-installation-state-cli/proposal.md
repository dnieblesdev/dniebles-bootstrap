# Proposal: Wire Installation State into CLI Plan

## Intent

`dbootstrap plan` still passes an empty `InstallationState`, so detected tools/runtimes never affect CLI output. Wire the existing detector into the CLI composition root while keeping planning pure, rendering mechanical, and tests deterministic.

## Scope

### In Scope
- Call `internal/state` detection from `cmd/dbootstrap plan` after catalog load.
- Pass detected `planning.InstallationState` to `planning.BuildPlan`.
- Add a package-level CLI test seam and deterministic tests for empty state and `already_installed` output.

### Out of Scope
- Planning or detector behavior changes.
- Installers, package-manager checks, command runner, apply/install, dotfiles runtime, or TUI.
- Renderer changes unless implementation proves current generic status rendering is insufficient.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `installation-state`: `plan` command consumes detected installation state so present tool/runtime resources render as `already_installed`.

## Approach

Mirror the existing environment-facts seam: add `detectInstallationState = state.Detect`, call it with the loaded catalog, and pass the result to `BuildPlan`. Tests should stub this seam to avoid real PATH dependence. Keep `render.go` unchanged because it already prints arbitrary step statuses from planner results.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modified | Compose `internal/state` detector into `plan` and pass installation state to planner. |
| `cmd/dbootstrap/main_test.go` | Modified | Add seam helper and output tests for empty and present-resource states. |
| `cmd/dbootstrap/render.go` | Unchanged | Existing generic `[status]` rendering should cover `already_installed`. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| CLI tests become host-dependent through real PATH lookup | Med | Stub installation state in CLI tests, including empty-state baseline. |
| Scope creep into package manager or installers | Low | Keep detector/planner/installers untouched; only wire composition root. |
| Existing exact-output test changes unexpectedly | Med | Preserve baseline with empty stub and add separate `already_installed` case. |

## Rollback Plan

Revert the `cmd/dbootstrap` wiring and tests; `planning.BuildPlan` can keep receiving `planning.InstallationState{}` as before. No data migration or runtime state cleanup is required.

## Dependencies

- Existing `internal/state.Detect` and archived `installation-state` spec.

## Success Criteria

- [ ] `dbootstrap plan` passes detected installation state to `planning.BuildPlan`.
- [ ] Empty installation state preserves current planned/attention output.
- [ ] Stubbed present tool/runtime renders `already_installed` without renderer changes.
