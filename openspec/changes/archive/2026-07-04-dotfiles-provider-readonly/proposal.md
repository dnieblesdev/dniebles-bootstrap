# Proposal: Dotfiles Provider Readonly

## Intent

Make dotfile modules visible to `dbootstrap plan` without applying or owning dotfiles runtime behavior. The slice adds catalog support for `[[dotfiles]]`, detects local module directories under `$HOME/.dotfiles`, and feeds that read-only presence into existing planning state.

## Scope

### In Scope
- TOML `[[dotfiles]]` catalog parsing, mapping, dependency validation, and fixture coverage.
- New `internal/dotfiles` read-only detector/provider with injectable filesystem seams.
- CLI composition-root wiring that merges present dotfile modules into `InstallationState.PresentResources` before `BuildPlan`.

### Out of Scope
- Applying, installing, cloning, or mutating dotfiles.
- Invoking dotlink, managing symlinks, or checking symlink drift.
- Claiming dotfiles runtime ownership or changing `BuildPlan` signature unless unavoidable.

## Capabilities

### New Capabilities
- `dotfiles-provider`: TOML dotfile resources plus read-only repo/module presence detection and CLI wiring.

### Modified Capabilities
- `installation-state`: dotfile resource presence may be supplied through existing `InstallationState.PresentResources` semantics.

## Approach

Implement a separate `internal/dotfiles` adapter following the existing detector pattern. Default `BasePath` to `$HOME/.dotfiles`; detect repo presence and module directory presence through injected `PathExists`/`ReadDir` seams. Keep planning pure by merging detected `dotfile:<module>` resources into installation state at `cmd/dbootstrap` before calling `BuildPlan`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/catalog/toml/*` | Modified | Add `[[dotfiles]]` schema, mapping, validation, and tests. |
| `internal/dotfiles/` | New | Read-only detector/provider with host-independent tests. |
| `cmd/dbootstrap/main.go` | Modified | Inject and run dotfiles detection, then merge into installation state. |
| `cmd/dbootstrap/main_test.go` | Modified | Stub dotfiles detection for deterministic CLI tests. |
| `internal/planning/*` | Modified | Add/adjust tests only if needed for dotfile presence semantics. |
| `catalog/bootstrap.toml` | Modified | Add a small sample dotfile resource if it improves fixture coverage. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Directory presence is confused with installed symlinks | Med | Document/read tests as module-availability only; keep dotlink out of scope. |
| Dotfiles boundary drifts into apply/runtime ownership | Med | Keep adapter read-only and limit CLI wiring to state merge. |
| Planner signature expands unnecessarily | Low | Prefer composition-root merge; change signature only if tests prove required. |

## Rollback Plan

Remove `internal/dotfiles`, revert TOML dotfiles support and fixture entries, and restore CLI wiring to the previous environment/config/installation detection flow. No host data migration is required because the slice is read-only.

## Dependencies

- Existing planning `ResourceKindDotfile` and `InstallationState.PresentResources` support.
- Local convention: dotfile resource names map to `$HOME/.dotfiles/<name>/` directories.

## Success Criteria

- [ ] `[[dotfiles]]` entries load, validate, and map into catalog resources.
- [ ] Dotfiles detector reports repo/module presence through injected seams with no host mutation.
- [ ] `dbootstrap plan` can show present dotfile modules via existing `already_installed` semantics.
- [ ] No dotlink, symlink, clone, apply, or install behavior is introduced.
