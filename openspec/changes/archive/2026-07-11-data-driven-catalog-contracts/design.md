# Design: Data-Driven Catalog Contracts

## Technical Approach

Refactor test and canonical-spec contracts only. `catalog/bootstrap.toml` remains the sole default inventory. Parser/schema tests retain fixed, small TOML inputs; default-catalog tests independently decode raw TOML and validate identities, references, Brew metadata, and a raw profile-root reachability graph before comparing profile plans. CLI tests use minimal temporary catalogs for exact rendering and execution assertions. `internal/planning/builder_test.go` remains the owner of exact planner ordering behavior.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Default inventory | Derive expectations from raw default TOML sections, not named resources | Retain full typed expected catalog | A second inventory becomes stale whenever catalog data grows. |
| Workflow reachability | Treat all declared profiles as roots; traverse profile resources, profile bundles, bundle resources, and dependency closure | Treat direct resource selection as reachability; hardcode a profile/resource list | Profile-root traversal proves intended default workflow membership without duplicating inventory. |
| Independent oracle | Decode raw TOML in `catalog_test.go`, derive identities, references, and profile-root closure from raw entries, then compare to `toml.LoadFile` output and profile `BuildPlan` results | Derive expected membership from decoded maps or plans | Raw declarations and planner output are separate boundaries, preventing tautological assertions. |
| CLI scope | Use `t.TempDir()` catalogs with only resources required by each rendering/execution scenario | Snapshot default catalog output | Exact output stays deterministic without coupling to production inventory. |
| Planner ownership | Keep synthetic graph ordering/closure examples in `internal/planning/builder_test.go` | Re-test named default order in adapter/CLI tests | It directly owns `BuildPlan`, sorted traversal, and dependency-before-dependent behavior. |

## Data Flow

    bootstrap.toml ──raw TOML decode──> profile roots ──> workflow closure
          │                                  │                 │
          │                                  │                 └── Brew membership check
          └── toml.LoadFile ──> Catalog ──> BuildPlan(profile) ──> closure/order checks

    minimal TOML fixture ──> run(plan|apply) ──> renderer / injected runners

The default integrity test builds the expected workflow set exclusively from raw declarations: each profile is a root; its direct resources and bundle resources are seeds; resource dependencies are traversed transitively. It first resolves all raw cross-references, then requires every raw Brew-backed tool/package to be in that union. An orphan therefore fails before planning. For each profile, it compares the decoded profile plan to that independently derived closure, repeats the build for stable results, and checks dependency-before-dependent order. Point-resource requests are excluded from this invariant; they may be exercised only by separate explicit-selection tests.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/catalog/toml/catalog_test.go` | Modify | Replace `TestLoadFileAndBuildPlanFromFixture` named inventory snapshot with raw-section-derived default integrity helpers; keep parser/schema tables and move planner-specific semantic coverage out. |
| `cmd/dbootstrap/main_test.go` | Modify | Replace default-catalog exact plan/apply snapshots with minimal fixture writers and focused output assertions; retain one derived default-catalog smoke check only. |
| `internal/planning/builder_test.go` | Modify | Keep/add synthetic closure, repeatability, and dependency-before-dependent assertions as the independent planner contract. |
| `openspec/specs/catalog-installer-metadata/spec.md` | Modify | Replace named Brew targets, exact bundle/profile inventory, and no-new-resource wording with generic declared eligible Brew metadata/reachability requirements and scenarios. |
| `openspec/changes/data-driven-catalog-contracts/design.md` | Create | This implementation design. |

## Interfaces / Contracts

No production interfaces change. Test-local raw TOML structs/helpers may mirror only section IDs, resource IDs, references, and install/presence fields needed to build an independent oracle; they MUST NOT reuse `planning.Catalog`, decoded planning maps, or `BuildPlan` output as expected membership data.

For each raw tool/package whose install provider is `brew`, assert decoded metadata has non-empty provider/package and presence `{Kind: "command_exists", Name: non-empty}`. Resolve raw bundle/profile/dependency references against the raw resource identity set; derive Brew workflow membership from profile roots only; then separately assert decoded maps expose the same identities/counts and decoded profile plans expose the raw closures.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Parser/schema | TOML mapping, malformed/unknown refs, metadata validation | Existing table-driven `Decode` tests with small inline fixtures. |
| Default integrity | Section identity/counts, resolved references, Brew metadata, profile-root closure, orphan rejection | Raw default TOML oracle traverses profiles → resources/bundles → dependencies before decoded profile-plan comparison; direct selection is excluded. |
| Planner | Complete selection, stable order, dependency precedes dependent | Synthetic catalog in `builder_test.go`, preserving exact graph expectations. |
| CLI rendering/execution | Plan text, safe modes, Brew/APT provider dispatch, bootstrap guidance, dotfile reports/order | Fixture-sized catalogs plus existing detector/factory seams and explicit assertions. |

## Migration / Rollout

No migration required. No runtime, schema, provider, default catalog, archive, or production code changes. Apply canonical-spec edits only to the active source spec; do not modify archived artifacts.

## Open Questions

- [ ] None.
