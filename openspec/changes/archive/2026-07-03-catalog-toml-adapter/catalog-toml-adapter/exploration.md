# Exploration: catalog-toml-adapter

### Current State
The repo already has a pure, format-agnostic planning core in `internal/planning`. `BuildPlan` consumes an in-memory `Catalog`, `PlanRequest`, `EnvironmentFacts`, and `ConfigState`; it handles profile/bundle/resource expansion, dependency ordering, environment filtering, and missing-config attention without any TOML knowledge. The archived design/specs explicitly say the catalog should live in-repo and TOML is the preferred first authoring format, but the core must stay format-agnostic.

### Affected Areas
- `internal/planning/*.go` — already defines the domain contract; should remain pure and untouched for this slice.
- `internal/catalog` or `internal/catalog/toml` — best home for decoding/validation so TOML concerns stay out of planning.
- `catalog/*.toml` — likely the first in-repo authoring surface and fixture source.
- `openspec/changes/catalog-toml-adapter/exploration.md` — exploration artifact for the new slice.

### Approaches
1. **Minimal decode-only adapter** — add `internal/catalog/toml` with structs and a decoder that maps one TOML catalog file into `planning.Catalog`.
   - Pros: smallest meaningful slice; preserves boundary cleanly; enables end-to-end tests from file to plan.
   - Cons: validation stays shallow at first.
   - Effort: Low

2. **Schema-plus-validation adapter** — add the same adapter, but also validate unknown refs, duplicate names, and malformed dependency/config references before handing off to planning.
   - Pros: catches bad catalog data earlier; better authoring feedback.
   - Cons: more code and decisions now; risks duplicating planning-level checks.
   - Effort: Medium

### Recommendation
Start with `internal/catalog/toml` as a thin adapter that decodes TOML into `planning.Catalog` and performs only structural validation needed for safe decoding. Keep `internal/planning` unchanged. Include a small fixture catalog and an integration test that decodes the fixture and runs `planning.BuildPlan` to prove the boundary end-to-end.

### Risks
- If validation is too light, bad catalog data may reach planning and produce noisy diagnostics later.
- If the adapter becomes too smart, TOML/schema details will leak into the domain core.
- Without fixtures, the slice may remain theoretical and not prove the repository-local catalog direction.

### Ready for Proposal
Yes — the next step should propose the adapter/schema slice with `internal/catalog/toml`, an initial TOML shape, and fixture-driven tests while explicitly keeping CLI/installers/dotfiles runtime out of scope.
