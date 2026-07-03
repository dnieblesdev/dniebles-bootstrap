# Tasks: First Go Planning Slice

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~220-320 |
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
| 1 | Module + pure planning core scaffold | PR 1 | Base = main; include `go.mod`, domain types, builder skeleton, and minimal docs note. |
| 2 | Deterministic expansion and dependency ordering | PR 1 | Base = main; keep pure logic and table-driven tests together. |
| 3 | Config attention + invalid reference reporting | PR 1 | Base = main; verify status/result semantics and no side effects. |

## Phase 1: Foundation / Module Setup

- [x] 1.1 Create `go.mod` with the repository module path and Go version, so the planning package can compile and tests can run.
- [x] 1.2 Create `internal/planning/types.go` with `Catalog`, `Profile`, `Bundle`, `Resource`, `ResourceRef`, `ConfigPolicy`, `EnvironmentFacts`, `Plan`, `PlanStep`, and `PlanStepResult` value types.
- [x] 1.3 Add any minimal README/OpenSpec status note needed to state this slice only covers the pure planning core.

## Phase 2: Pure Planning Core

- [x] 2.1 Implement `internal/planning/builder.go` with `BuildPlan(catalog, request, facts, state) PlanResult` as a pure function using only decoded domain inputs.
- [x] 2.2 Add deterministic profile/bundle/resource expansion in `internal/planning`, ensuring duplicate refs are deduped and dependencies are topologically ordered.
- [x] 2.3 Encode missing-config attention handling in `PlanStepResult` without blocking unrelated valid resources.
- [x] 2.4 Preserve environment-fact filtering in the core using caller-supplied `EnvironmentFacts` only; no OS probing or adapter calls.

## Phase 3: Testing / Verification

- [x] 3.1 Add table-driven tests in `internal/planning/builder_test.go` for ordering, bundle expansion, and stable results from identical inputs.
- [x] 3.2 Add table-driven tests for invalid bundle/resource refs and missing-config attention semantics, asserting safe partial planning.
- [x] 3.3 Verify tests are pure and deterministic; acceptance: `go test ./...` passes after the module and planning package exist.

## Phase 4: Cleanup / Documentation

- [x] 4.1 Update only minimal docs/status text if needed to note TOML loader, CLI/TUI, installers, and OS probing remain deferred.
- [x] 4.2 Review the final diff for work-unit boundaries; prepare work-unit commit messages that keep tests with each behavior slice.
