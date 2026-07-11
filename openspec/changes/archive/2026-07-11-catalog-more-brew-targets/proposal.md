# Proposal: Catalog More Brew Targets

## Intent

Add `package:jq` to the default catalog: a conservative, broadly useful JSON-processing CLI target supported by the existing Homebrew package installer. This expands the catalog without changing runtime, dotfile, provider, runner, or apply behavior.

## Scope

### In Scope
- Add Homebrew-backed `package:jq` with package `jq` and `command_exists: jq` presence metadata.
- Include `package:jq` in `bundle:cli`, making it available through the existing `dev` profile.
- Update the default-catalog contract and focused catalog fixture/plan assertions.

### Out of Scope
- Changes to `runtime:go`, `dotfile:bash`, providers, command execution, or apply semantics.
- New bundles/profiles, fallback providers, platform-specific selection, or Homebrew bootstrap changes.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `catalog-installer-metadata`: Permit the default catalog to add `package:jq` and require its Homebrew and presence metadata plus `bundle:cli` membership.

## Approach

Extend the existing `[[packages]]` convention in `catalog/bootstrap.toml`; reuse the established Homebrew metadata shape used by `package:ripgrep`. Update only the fixture contract that decodes and plans the default catalog. Existing Homebrew package handling consumes the metadata unchanged.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `catalog/bootstrap.toml` | Modified | Add `package:jq`; add it to `bundle:cli`. |
| `internal/catalog/toml/catalog_test.go` | Modified | Assert catalog metadata, bundle membership, and planned resource set. |
| `openspec/specs/catalog-installer-metadata/spec.md` | Modified | Replace the no-new-resources contract with the approved jq target. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Default `dev` plans include one extra package | Low | Make bundle membership and expected plan order explicit in tests. |
| Metadata drifts from Homebrew capability | Low | Use existing `package` + `brew` convention; no execution changes. |

## Rollback Plan

Revert the `package:jq` catalog and `bundle:cli` entries, then restore the matching tests and catalog capability requirement. No migration or runtime state is introduced.

## Dependencies

- Existing Homebrew package installer capability; no new dependency.

## Success Criteria

- [ ] The default catalog declares `package:jq` with `brew/jq` and `command_exists: jq` metadata.
- [ ] `bundle:cli` and the `dev` profile plan include `package:jq` through existing selection behavior.
- [ ] Focused catalog tests and the full Go suite pass without runtime or provider changes.
