# Tasks: Point Install Planning

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 250-380 |
| 400-line budget risk | High |
| Chained PRs recommended | No |
| Suggested split | Single PR with maintainer-approved size exception |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | CLI target parsing/validation and request wiring | PR 1 | Base on main; includes tests for parser, target-required rules, and PlanRequest.Resources |
| 2 | Render resource-oriented plan headers | PR 1 | Same PR; keep renderer/test updates with CLI behavior |

## Phase 1: CLI Input and Request Wiring

- [x] 1.1 Add repeatable `--resource` flag handling in `cmd/dbootstrap/main.go` and update plan usage text.
- [x] 1.2 Add CLI-local `kind:name` parsing for `tool`, `runtime`, `package`, and `dotfile` refs with clear validation errors.
- [x] 1.3 Enforce `--profile` or at least one `--resource`; allow profile+resource union and pass both into `planning.PlanRequest.Resources`.

## Phase 2: Rendering

- [x] 2.1 Update `cmd/dbootstrap/render.go` to print a resource-oriented header when no profile is supplied.
- [x] 2.2 Keep profile-only output unchanged and preserve read-only plan rendering semantics.

## Phase 3: Testing / Verification

- [x] 3.1 Extend `cmd/dbootstrap/main_test.go` for resource-only, mixed profile+resource, repeated flags, malformed refs, unsupported kinds, and missing-target validation.
- [x] 3.2 Extend `cmd/dbootstrap/render_test.go` for resource-only header output and profile-header regression coverage.
- [x] 3.3 Run focused Go tests for `cmd/dbootstrap` and verify no planner/domain changes are required.

## Phase 4: Cleanup / Artifact Updates

- [x] 4.1 Confirm help text, error messages, and test fixtures match the spec scenarios exactly.
- [x] 4.2 Update OpenSpec and Engram task artifacts for the finalized implementation scope.
