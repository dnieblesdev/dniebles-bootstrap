# Proposal: Catalog Brew Provider Targets

## Intent

Exercise the completed brew-backed `apply --yes` path from the default catalog with one safe, common package target. Today the default catalog keeps `package:ripgrep` on `apt`, so confirmed brew execution is only reachable through custom catalog data.

## Scope

### In Scope
- Change the default catalog metadata for `package:ripgrep` from `apt` to `brew`.
- Keep `package:ripgrep` presence detection as `command_exists: rg`.
- Add planning/spec coverage for the default catalog containing a brew-backed package target.

### Out of Scope
- Multi-provider catalog entries or fallback providers.
- Changes to command runner, installers, apply wiring, Homebrew bootstrap, dotfiles, or bootstrap entrypoint.
- Changing `tool:git` or `runtime:go`.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `catalog-installer-metadata`: default catalog package metadata must include a brew-backed `package:ripgrep` target.

## Approach

Make the smallest catalog-only behavior change: update `catalog/bootstrap.toml` so `package:ripgrep` declares `install.provider = "brew"` and `install.package = "ripgrep"`. Existing metadata propagation and confirmed-mode brew execution should handle the rest.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `catalog/bootstrap.toml` | Modified | `package:ripgrep` provider changes from `apt` to `brew`. |
| `internal/catalog/toml/catalog_test.go` or existing catalog coverage | Modified | Assert default catalog decodes ripgrep as brew-backed and preserves `rg` presence. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Linux default catalog now points to brew metadata | Med | Keep execution limited to `apply --yes`; default and dry-run remain non-mutating with advisory bootstrap behavior. |
| Accidentally expanding provider scope | Low | Only edit ripgrep metadata and tests; no installer or runner changes. |

## Rollback Plan

Revert `catalog/bootstrap.toml` ripgrep install provider/package metadata and related test/spec updates.

## Dependencies

- Completed `wire-brew-apply` behavior for brew-backed tool/package execution under `apply --yes`.

## Success Criteria

- [ ] Default catalog decodes `package:ripgrep` with provider `brew` and package `ripgrep`.
- [ ] Presence detection for ripgrep remains `command_exists: rg`.
- [ ] No execution wiring, installer, bootstrap, dotfile, or entrypoint files change.
