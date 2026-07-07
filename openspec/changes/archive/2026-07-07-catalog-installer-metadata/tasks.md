# Tasks: Catalog Installer Metadata

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 220-320 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Add inert install/presence metadata plumbing end-to-end | PR 1 | Single PR size-exception; include catalog, planning, adapter, and tests together |

## Phase 1: Planning Model Foundation

- [x] 1.1 Add `InstallMetadata` and `PresenceMetadata` structs to `internal/planning/types.go`, and extend `Resource` with optional pointers for inert metadata.
- [x] 1.2 Update `internal/planning/builder.go` to preserve metadata in `PlanStep.Resource` without changing ordering, status, or diagnostics.
- [x] 1.3 Extend `internal/planning/builder_test.go` to prove metadata survives plan creation and does not alter existing plan results.

## Phase 2: TOML Schema and Mapping

- [x] 2.1 Add private `installEntry` and `presenceEntry` shapes plus pointer fields on `internal/catalog/toml/schema.go` for nested `[install]` and `[presence]` tables.
- [x] 2.2 Map TOML metadata into `planning.Resource` in `internal/catalog/toml/catalog.go`, cloning values and leaving absent metadata nil.
- [x] 2.3 Extend `internal/catalog/toml/validate.go` to reject partial/empty metadata and unknown presence kinds while keeping metadata optional.

## Phase 3: Catalog Fixtures and Verification

- [x] 3.1 Update `catalog/bootstrap.toml` with safe representative `[install]` and `[presence]` metadata for tool/runtime/package examples only.
- [x] 3.2 Extend `internal/catalog/toml/catalog_test.go` with decode coverage for metadata preservation and absent-metadata no-op behavior.
- [x] 3.3 Add validation tests for malformed install metadata, malformed presence metadata, and unsupported presence kinds in `internal/catalog/toml/catalog_test.go`.
- [x] 3.4 Verify the fixture still builds the same plan and preserves existing statuses when metadata is present or absent.

## Phase 4: Cleanup / Review Readiness

- [x] 4.1 Review comments and test names for English-only, review-friendly wording that emphasizes metadata is inert and optional.
- [x] 4.2 Confirm no command runner, raw shell execution, installer dispatch, or bootstrap entrypoint wiring was introduced.
