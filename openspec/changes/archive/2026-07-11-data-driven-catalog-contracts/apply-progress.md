# Apply Progress: Data-Driven Catalog Contracts

## Status

All 9 tasks are complete. Strict TDD mode was active; this test/spec-only refactor introduced no production runtime changes.

## Completed Tasks

- [x] 1.1 Raw TOML oracle structs and helpers
- [x] 1.2 Raw-derived default integrity, orphan rejection, and profile-plan checks
- [x] 2.1 Synthetic planner closure, repeatability, and dependency ordering contract
- [x] 2.2 Fixture-sized plan/apply output contracts
- [x] 2.3 All-profile derived default-catalog CLI smoke check with exactly-once rendered-step assertions
- [x] 3.1 Generic active canonical/development catalog-installer-metadata requirement with immutable archive wording
- [x] 3.2 Test/spec-only scope confirmation
- [x] 4.1 Fresh evidence-backed verification after confirmed apply-review corrections
- [x] 4.2 Accepted `size:exception` scope and delivery-plan confirmation

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| 1.1 | `internal/catalog/toml/catalog_test.go` | Unit | `go test ./internal/catalog/toml` passed | Raw oracle test referenced missing test-local types/helpers; compile failed as expected | Focused raw-oracle tests passed | Default catalog plus orphan/point-selection fixture | Extracted raw graph helpers; focused test passed after formatting |
| 1.2 | `internal/catalog/toml/catalog_test.go` | Unit | `go test ./internal/catalog/toml` passed | Same raw integrity contract failed before helper implementation | Focused raw-oracle tests passed | Valid default graph and invalid orphan graph | None needed |
| 2.1 | `internal/planning/builder_test.go` | Unit | `go test ./internal/planning` passed | New closure-order test referenced missing assertion helper; compile failed as expected | Focused planner test passed | Existing profile/point-selection table plus three-node transitive graph | Extracted dependency-order assertion; focused test passed after formatting |
| 2.2 | `cmd/dbootstrap/main_test.go` | Integration | `go test ./cmd/dbootstrap` passed | Plan/apply contract tests referenced missing custom fixture/path helper; compile failed as expected | Focused plan/apply tests passed | Plan and apply exact-output modes use the temporary catalog | Reused existing fixture writer pattern; focused test passed after formatting |
| 2.3 | `cmd/dbootstrap/main_test.go` | Integration | `go test ./cmd/dbootstrap -run '^TestRunPlanDefaultCatalogSmokeIsDerived$'` passed | Strengthened all-profile, exactly-once assertions were written before the focused execution; existing production behavior already satisfied the corrected test-only contract | Focused smoke test passed | All declared profiles, independently built plans, and every rendered planned step | Sorted profile traversal and exact rendered-step counting; focused test remained green |
| 3.1 | `openspec/specs/catalog-installer-metadata/spec.md` | Documentation | N/A | N/A: canonical-spec-only wording correction | N/A | Active-contract rule and immutable-archive exception stated explicitly | Scoped non-enumeration to active canonical/development contracts; did not edit archives |
| 3.2 | Diff scope | Inspection | N/A | N/A: inspection-only task | `git diff --check` passed | Triangulation skipped: one final scoped diff is definitive | None needed |
| 4.1 | Focused/full Go suite | Verification | Focused smoke safety net passed | N/A: verification task | One fresh focused/full/vet/format/diff verification passed | Focused packages plus full suite and vet | None needed |
| 4.2 | Diff scope | Inspection | N/A | N/A: delivery confirmation task | Final diff/stat passed | Triangulation skipped: single accepted delivery boundary | None needed |

## Verification Evidence

Executed once after implementation:

```text
go test ./internal/catalog/toml ./internal/planning ./cmd/dbootstrap
ok   github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml  0.003s
ok   github.com/dnieblesdev/dniebles-bootstrap/internal/planning      0.002s
ok   github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap         0.011s

go test ./...
all packages passed

go vet ./...
passed

git diff --check
passed
```

`git diff --stat` reported 506 additions and 118 deletions across four tracked test/spec files. The change directory is an active, untracked OpenSpec artifact; no archive was modified. Changed tracked files are limited to `cmd/dbootstrap/main_test.go`, `internal/catalog/toml/catalog_test.go`, `internal/planning/builder_test.go`, and `openspec/specs/catalog-installer-metadata/spec.md`.

## Spec Scenario Coverage

- Minimal custom catalog behavior: plan/apply exact-output contracts now use `t.TempDir()` fixture catalogs.
- Provider, safe modes, APT/sudo, bootstrap guidance, reporting, and dotfile detail ordering: retained in existing focused fixture tests.
- Generic Brew metadata, raw references, profile-root reachability, orphan rejection, and point-selection exclusion: covered by the independent raw TOML oracle.
- Profile closure and deterministic dependency-first planning: compared against raw closure and reinforced with a synthetic transitive planner graph.
- Unchanged runtime/default catalog/archive behavior: verified by diff scope; no production or catalog files changed.

## Delivery Boundary

- Mode: `size:exception` / `exception-ok`
- Review impact: 624 changed lines, above the 400-line budget and within the approved ~800-line exception.
- Boundary: test/spec-only contract refactor. No commit, push, branch, issue, or PR was performed in this apply run.

## Confirmed Apply-Review Corrections

- Task 2.3: Strengthened the default-catalog CLI smoke check. It now sorts and iterates every declared profile, independently builds each profile plan through `buildPlan`, and requires every planned rendered-step line exactly once. No inventory names were introduced.
- Task 3.1: Scoped the non-enumeration rule to active canonical and development contracts going forward. Archived historical artifacts remain immutable and MAY truthfully retain prior resource enumerations; no archive was modified.
- Task 4.1: Fresh verification ran once after these corrections: `go test ./internal/catalog/toml ./internal/planning ./cmd/dbootstrap`, `go test ./...`, `go vet ./...`, `gofmt -w cmd/dbootstrap/main_test.go`, and `git diff --check` all passed.
- Scope/status: tracked modifications remain the four pre-existing test/spec files; the active OpenSpec change artifacts remain untracked. No production runtime, catalog, or archive files changed.
