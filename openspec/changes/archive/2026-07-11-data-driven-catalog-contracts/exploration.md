## Exploration: data-driven-catalog-contracts

### Current State
Runtime catalog behavior is already data-driven. `catalog/bootstrap.toml` declares resources, bundle membership, profiles, dependencies, and structured install/presence metadata. `internal/catalog/toml` decodes every declared resource by section; `internal/planning` expands profiles and bundles, resolves dependencies, and computes deterministic topological order without knowing the default inventory. CLI composition selects Homebrew behavior from `Resource.Install.Provider == "brew"` and the resource kind, so a new eligible Brew target does not require a runtime switch.

The duplication is in development contracts:

- `internal/catalog/toml/catalog_test.go:254-358` reconstructs the entire default catalog as a hardcoded `planning.Catalog`, then separately hardcodes the expected `profile:dev` plan order and metadata cases. Adding a resource requires editing the fixture and several switch/assertion branches.
- `cmd/dbootstrap/main_test.go:26-274` and `:514-684` use exact CLI output for the real default catalog and hardcode its four-step inventory, statuses, and ordering. The same inventory is repeated for default, dry-run, and confirmed apply output.
- `internal/planning/builder_test.go:41-49`, `:126-129`, and other cases intentionally hardcode small synthetic ordering expectations. These test planner ordering semantics, not the default catalog, and should remain independent.
- `openspec/specs/catalog-installer-metadata/spec.md:62-101` names `tool:git`, `package:ripgrep`, and `package:jq`, their metadata, exact bundle membership, and a prohibition on new resources. This is a historical inventory contract rather than a reusable rule for future declared Brew targets.
- Archived change artifacts under `openspec/changes/archive/` are audit history and must not be rewritten.

The exact-output tests are valuable behavioral snapshots, but the default-catalog snapshots currently conflate renderer/CLI contracts with catalog inventory. The renderer tests in `cmd/dbootstrap/render_test.go` already use synthetic data and are appropriately inventory-independent.

### Affected Areas
- `internal/catalog/toml/catalog_test.go` — replace the full expected default model with generic fixture contracts and retain focused decoder/metadata tests.
- `cmd/dbootstrap/main_test.go` — move exact rendering cases to minimal temporary catalogs; keep a derived default-catalog smoke assertion rather than a four-step snapshot.
- `openspec/specs/catalog-installer-metadata/spec.md` — replace named inventory requirements with canonical data-driven requirements covering every declared default Brew target and its bundle/profile reachability.
- `catalog/bootstrap.toml` — no change is needed for this refactor; it remains the single source of default resource declarations.
- `internal/planning/builder_test.go` — preserve synthetic ordering tests; only consider adding a generic dependency-before-dependent assertion if the default fixture contract needs it.
- `openspec/changes/data-driven-catalog-contracts/` — subsequent proposal/spec/design phases will describe the approved test/spec refactor.

### Approaches
1. **Generic default-catalog contract plus minimal CLI fixtures** — load the real TOML once, iterate declared resources/bundles/profiles, validate map completeness and reference reachability, validate structured Brew metadata for every Brew-backed tool/package, and derive the selected plan from catalog data. Use small temporary catalogs for exact CLI output tests.
   - Pros: smallest safe change; new Brew resources are automatically covered; preserves exact renderer/CLI behavior checks without inventory churn; keeps runtime untouched.
   - Cons: default-catalog output no longer snapshots every label and line; generic assertions must be explicit enough to avoid becoming tautological.
   - Effort: Low/Medium

2. **Generate expected Go fixtures or snapshots from TOML** — decode the catalog and serialize/compare a generated expected model or golden output.
   - Pros: broad structural coverage and low manual maintenance after setup.
   - Cons: can hide incorrect mappings by deriving both sides from the same decoder; golden updates can normalize unintended inventory changes; larger tooling surface than needed.
   - Effort: Medium/High

3. **Leave named contracts and add more named cases** — retain the current exact inventory assertions and append each future target manually.
   - Pros: no refactor risk and very explicit current behavior.
   - Cons: directly violates the objective; every target remains duplicated across Go tests, CLI snapshots, and canonical specs.
   - Effort: Low initially / High ongoing

### Recommendation
Choose Approach 1. Keep `catalog/bootstrap.toml` as the only default inventory declaration. Refactor the catalog integration test to assert generic, non-tautological invariants: every declared resource is decoded with its section kind; every bundle/profile reference resolves; every default Brew-backed `tool`/`package` has non-empty provider/package metadata and `command_exists` presence metadata; every such target is reachable through the intended bundle/profile selection; and the plan contains the derived dependency closure in deterministic order with dependencies before dependents. Assert that the selected plan and result set are internally complete, but do not enumerate resource names.

Retain exact-output tests for usage errors, renderer formatting, statuses, manual actions, and ordering semantics. For tests whose purpose is CLI formatting or execution-mode behavior, use purpose-built temporary catalogs with stable one- or two-resource inventories. Add at most one default-catalog smoke test that derives expected step lines from the loaded plan and verifies each selected step is rendered once; it should not hardcode resource names or count.

Update the canonical requirement to describe the rule over all declared default Brew targets and their declared bundle/profile reachability. Preserve separate canonical requirements for inert metadata propagation, provider selection, exact command vectors, safe modes, and execution boundaries; those are runtime contracts and must not be weakened by this refactor.

### Risks
- A generic assertion that only reuses the decoder can pass while both catalog parsing and the test share the same omission. Mitigate by checking raw section counts/IDs against decoded typed maps and validating cross-section references and plan closure independently.
- Removing the default exact-output snapshot could miss an unintended description or ordering change. Mitigate with minimal-fixture exact-output tests, a derived default smoke check, and the existing planner ordering tests.
- Requiring every Brew-backed target to be reachable from `profile:dev` could over-constrain future catalog design. Scope the contract to the declared default target bundle/profile policy (or explicitly define that policy in the canonical spec) rather than assuming all future profiles use `dev`.
- Canonical spec edits must be made as a full `MODIFIED` requirement delta in the later spec phase; archived specs remain immutable audit history.

### Ready for Proposal
Yes. The proposal should state that this is a development-contract refactor only: no production Go/runtime behavior, catalog schema, provider capability, CLI semantics, or default resource declarations change. It should name `internal/catalog/toml/catalog_test.go`, `cmd/dbootstrap/main_test.go`, and the canonical `catalog-installer-metadata` requirement as the minimal scope.
