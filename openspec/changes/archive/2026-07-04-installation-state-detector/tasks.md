# Tasks: Installation State Detector

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 250-380 |
| Size exception status | Approved for this change |
| Chained PRs recommended | No |
| Suggested split | Single PR with reviewable work-unit commit |
| Delivery strategy | exception-ok |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
Size exception status: Approved for this change

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Add planning state, PATH detector seam, call-site updates, and tests | PR 1 | Keep planning, detector, call sites, tests, and any minimal docs together as one reviewable work unit. |

## Phase 1: Planning domain foundation

- [x] 1.1 Add `InstallationState` and `PlanStepStatusAlreadyInstalled` to `internal/planning/types.go` with `PresentResources map[ResourceRef]bool`.
- [x] 1.2 Update `internal/planning/builder.go` to accept `InstallationState` in `BuildPlan` and preserve pure planning semantics.

## Phase 2: Planner status precedence

- [x] 2.1 Implement `already_installed` selection after environment matching and before `planned` / `attention_required` in `internal/planning/builder.go`.
- [x] 2.2 Keep missing-config reasons attached when status becomes `already_installed`, and keep environment mismatch as `skipped`.

## Phase 3: Host-independent state detector

- [x] 3.1 Create `internal/state/detector.go` with `PathLookup` seam and `Detector{LookPath}` defaulting to `exec.LookPath`.
- [x] 3.2 Map `tool` and `runtime` refs to PATH presence only; ignore `package` and `dotfile` in this slice.

## Phase 4: Tests and call-site updates

- [x] 4.1 Extend `internal/planning/builder_test.go` for empty-state compatibility, present-resource precedence, mixed state, and purity regression.
- [x] 4.2 Add `internal/state/detector_test.go` with injected lookup fixtures proving host-independent present/absent behavior.
- [x] 4.3 Update `cmd/dbootstrap/main.go` and `internal/catalog/toml/catalog_test.go` to pass `planning.InstallationState{}` to `BuildPlan`.
- [x] 4.4 Run `go test ./... -count=1` and update `README.md` only if a minimal note is required.
