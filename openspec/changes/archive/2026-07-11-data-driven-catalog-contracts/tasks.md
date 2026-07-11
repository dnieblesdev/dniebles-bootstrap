# Tasks: Data-Driven Catalog Contracts

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | ~800 |
| 400-line budget risk | High |
| Chained PRs recommended | No |
| Suggested split | One direct commit/PR to `main` with accepted size exception |
| Delivery strategy | exception-ok |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | Complete test/spec-only contract refactor | PR 1 | Direct `main` commit/push; include focused tests and final evidence-backed verification |

## Phase 1: Independent Catalog Oracle

- [x] 1.1 In `internal/catalog/toml/catalog_test.go`, add test-local raw TOML structs and helpers for section identities, references, Brew install/presence metadata, and profile-root closure.
- [x] 1.2 Replace the named default inventory snapshot with raw-derived identity/count, reference-resolution, metadata, orphan-rejection, and decoded profile-plan closure checks; never use point selection for membership.

## Phase 2: Planner and CLI Contract Refactor

- [x] 2.1 In `internal/planning/builder_test.go`, retain/add synthetic closure, repeatability, deterministic ordering, and dependency-before-dependent assertions independent of the default catalog.
- [x] 2.2 In `cmd/dbootstrap/main_test.go`, replace default exact plan/apply snapshots with minimal `t.TempDir()` TOML fixtures covering rendering, safe modes, provider dispatch, bootstrap guidance, reports, and ordering.
- [x] 2.3 Preserve a single derived default-catalog smoke check that sorts and iterates every declared profile, independently builds each production plan, and asserts every planned step renders exactly once; preserve existing provider, safety, reporting, manual-action, and order behavioral coverage without named inventory expectations.

## Phase 3: Canonical Specification

- [x] 3.1 Update `openspec/specs/catalog-installer-metadata/spec.md` to generic declared Brew metadata, raw profile-root reachability, orphan failure, independent invariants, deterministic plans, active canonical/development non-enumeration, and immutable historical archive wording.
- [x] 3.2 Confirm no archived OpenSpec artifact, `catalog/bootstrap.toml`, production code, schema, provider, or runtime/default behavior is modified.

## Phase 4: Verification and Review Evidence

- [x] 4.1 Run one fresh evidence-backed verification after the confirmed review corrections: focused Go tests, full `go test ./...`, `go vet ./...`, formatting/diff/scope inspection, and evidence capture for every spec scenario; do not repeat unless a failure requires a corrective edit.
- [x] 4.2 Confirm the final diff remains test/spec-only, review size is documented as the accepted `size:exception`, and the single direct commit/push-to-`main` delivery plan is honored.
