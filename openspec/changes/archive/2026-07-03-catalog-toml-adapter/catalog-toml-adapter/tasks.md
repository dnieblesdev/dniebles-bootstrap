# Tasks: Catalog TOML Adapter

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 280-380 |
| Size exception status | Approved for this change |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | exception-ok |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
Size exception status: Approved for this change

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Add TOML adapter package and sample catalog | PR 1 | Base branch = main; include parser, schema, validation, fixture. |
| 2 | Prove decode-to-plan behavior | PR 1 | Same PR; include tests for decode, validation errors, and BuildPlan integration. |

## Phase 1: Foundation / Infrastructure

- [x] 1.1 Create `internal/catalog/toml/catalog.go`, `schema.go`, and `validate.go` with `LoadFile`/`Decode` APIs returning `planning.Catalog`.
- [x] 1.2 Add a minimal TOML dependency in `go.mod`/`go.sum` and keep all TOML DTOs isolated from `internal/planning`.
- [x] 1.3 Add `catalog/bootstrap.toml` or equivalent fixture source matching the initial schema in the design.

## Phase 2: Core Implementation

- [x] 2.1 Map TOML tables for profiles, bundles, resources, config policy, dependencies, and environment conditions into `planning.Catalog`.
- [x] 2.2 Implement shallow validation for required IDs/names, duplicate IDs, supported kinds, malformed refs, and basic unknown local references.
- [x] 2.3 Keep planner semantics out of the adapter; pass only structurally valid decoded data into `planning.BuildPlan`.

## Phase 3: Testing / Verification

- [x] 3.1 Add table-driven tests in `internal/catalog/toml/catalog_test.go` for valid decode mappings from fixture and inline TOML.
- [x] 3.2 Add explicit error tests for invalid TOML syntax, missing required fields, duplicate IDs, and bad/unknown local refs.
- [x] 3.3 Add an integration test that decodes the fixture and runs `planning.BuildPlan`, asserting the adapter stays side-effect free.
- [x] 3.4 Verify with `go test ./...` after the adapter and tests are in place.

## Phase 4: Cleanup / Documentation

- [x] 4.1 Update `README.md` or minimal status docs if the new catalog authoring surface needs one short mention.
- [x] 4.2 Keep comments and fixture names aligned with the initial TOML schema so future adapters can extend without leaking TOML into planning.
- [x] 4.3 Prepare a work-unit commit with code + tests + fixture together so the PR stays reviewable as one slice.
