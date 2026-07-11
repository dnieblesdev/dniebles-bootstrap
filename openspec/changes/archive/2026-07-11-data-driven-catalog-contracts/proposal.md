# Proposal: Data-Driven Catalog Contracts

## Intent

Remove duplicated default-catalog inventory from development contracts so adding a default Brew target normally requires only its catalog declaration and declared bundle/profile membership. This is a test/spec refactor; runtime behavior and catalog contents remain unchanged.

## Scope

### In Scope
- Replace hardcoded default inventory fixtures with generic, independent catalog, graph reachability, metadata, and derived-plan invariants.
- Use minimal temporary catalogs for exact CLI mode, safety, provider, ordering, and report assertions; retain one derived default-catalog smoke check.
- Generalize canonical catalog requirements to all declared eligible Brew targets and their declared reachability.

### Out of Scope
- Changing `catalog/bootstrap.toml`, schema, planning, providers, CLI semantics, or execution behavior.
- Updating archived OpenSpec artifacts.
- Adding providers, Brew targets, or fallback metadata.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `catalog-installer-metadata`: replace named default-resource inventory requirements with generic declared Brew-target metadata and reachability contracts.

## Approach

Treat `catalog/bootstrap.toml` as the sole default inventory. Independently derive a declared workflow graph from raw profile roots through profile resources, profile bundles, bundle resources, and transitive resource dependencies. Every default Brew-backed tool/package must belong to that raw closure; an orphan fails even if a point-resource smoke test can select it. Separately compare raw identities/references and profile-plan closure with decoded/planned results, including deterministic dependency-before-dependent order. Keep exact output and execution-mode tests fixture-sized; direct resource selection remains a separate CLI behavior contract and never proves workflow membership.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/catalog/toml/catalog_test.go` | Modified | Generic default-catalog invariants |
| `cmd/dbootstrap/main_test.go` | Modified | Minimal CLI fixtures and derived smoke check |
| `internal/planning/builder_test.go` | Modified | Retain independent ordering contracts |
| `openspec/specs/catalog-installer-metadata/spec.md` | Modified | Generic canonical requirements |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Tautological derived assertions | Med | Compare raw sections, references, and plan closure independently |
| Direct-selection false positive | Med | Derive Brew membership solely from raw profile-root closure; do not include point-resource requests in that invariant |
| Reduced default snapshot coverage | Low | Keep focused exact fixtures and a derived smoke assertion |
| Future profile over-constraint | Low | Use declared reachability policy, not a hardcoded profile |

## Rollback Plan

Revert the test and canonical-spec changes as one unit; no runtime state, catalog data, or migration is involved.

## Dependencies

- Existing catalog decoder, planner, renderer, and safety/provider contracts.

## Success Criteria

- [ ] Adding an eligible default Brew target needs no named inventory edits outside its catalog declaration and declared profile-root graph membership.
- [ ] Every default Brew-backed target is in the raw profile-root closure; direct resource selection cannot satisfy this invariant.
- [ ] Generic invariants, minimal CLI fixtures, and meaningful safety/provider/ordering/report checks remain covered.
- [ ] Canonical specs enumerate no current default Brew resource names; archived artifacts remain unchanged.
