# Apply Progress: Point Install Planning

## Status

All tasks completed. No blockers.

## Completed Tasks

- [x] 1.1 Add repeatable `--resource` flag handling in `cmd/dbootstrap/main.go` and update plan usage text.
- [x] 1.2 Add CLI-local `kind:name` parsing for `tool`, `runtime`, `package`, and `dotfile` refs with clear validation errors.
- [x] 1.3 Enforce `--profile` or at least one `--resource`; allow profile+resource union and pass both into `planning.PlanRequest.Resources`.
- [x] 2.1 Update `cmd/dbootstrap/render.go` to print a resource-oriented header when no profile is supplied.
- [x] 2.2 Keep profile-only output unchanged and preserve read-only plan rendering semantics.
- [x] 3.1 Extend `cmd/dbootstrap/main_test.go` for resource-only, mixed profile+resource, repeated flags, malformed refs, unsupported kinds, and missing-target validation.
- [x] 3.2 Extend `cmd/dbootstrap/render_test.go` for resource-only header output and profile-header regression coverage.
- [x] 3.3 Run focused Go tests for `cmd/dbootstrap` and verify no planner/domain changes are required.
- [x] 4.1 Confirm help text, error messages, and test fixtures match the spec scenarios exactly.
- [x] 4.2 Update OpenSpec and Engram task artifacts for the finalized implementation scope.

## Files Changed

| File | Action | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modified | Added `--resource` repeatable flag, `resourceFlag` accumulator, `parseResourceRef`, `parseResourceRefs`, `dedupeResourceRefs`, target-required validation, and `PlanRequest.Resources` wiring. Updated plan usage text. |
| `cmd/dbootstrap/render.go` | Modified | Made `renderPlanResult` header conditional: profile header when profile is set, `Plan resources: ...` header when profile is empty. |
| `cmd/dbootstrap/main_test.go` | Modified | Added parser unit tests, resource-only plan test, profile+resource union test, repeated-resource deduplication test, malformed/unsupported ref tests, and updated missing-target usage assertion. Added `dotfilesState` to test struct to keep union test deterministic. |
| `cmd/dbootstrap/render_test.go` | Modified | Updated existing call to new signature; added `TestRenderPlanResultResourceOnlyHeader`. |

## Deviations from Design

None — implementation matches design.

## Issues Found

None.

## Remediation

Formal verify reported a `gofmt -l cmd/dbootstrap/` warning for `cmd/dbootstrap/main_test.go`.
Ran `gofmt -w cmd/dbootstrap/main_test.go`, then re-ran `gofmt -l cmd/dbootstrap/` (clean) and focused `go test ./cmd/dbootstrap/... -count=1` (passed).

## Tests Run

- `go test ./cmd/dbootstrap/...` — passed
- `go test ./...` — all packages passed

## Workload / PR Boundary

- Mode: single PR with maintainer-approved size exception
- Chain strategy: size-exception
- Estimated review budget impact: 281 insertions, 10 deletions across 4 files; within the approved size exception.

## Next Recommended

`sdd-verify`
