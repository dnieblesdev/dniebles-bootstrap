# Tasks: CLI Plan Command

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 220-360 |
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
| 1 | Ship `dbootstrap plan` wiring, deterministic rendering, tests, and usage docs | PR 1 | Base: main; keep code, tests, and README together as one reviewable work unit. |

## Phase 1: Foundation / Entry Point

- [x] 1.1 Create `cmd/dbootstrap/main.go` with `main()` -> `run(args, stdout, stderr)` and `plan` subcommand dispatch.
- [x] 1.2 Add stdlib `flag` parsing for `plan --profile <name>` plus optional `--catalog <path>` defaulting to `catalog/bootstrap.toml`.
- [x] 1.3 Define static CLI `EnvironmentFacts` and empty `ConfigState` inputs; do not add any OS probing.

## Phase 2: Core Command Behavior

- [x] 2.1 Call `internal/catalog/toml.LoadFile` with the resolved catalog path and pass the catalog into `planning.BuildPlan`.
- [x] 2.2 Render `planning.PlanResult` deterministically in `cmd/dbootstrap`, including planned, skipped, attention, and diagnostic/error cases.
- [x] 2.3 Map usage, load, and planning failures to stable exit codes and human-readable stderr messages.

## Phase 3: Testing / Verification

- [x] 3.1 Add table-driven tests in `cmd/dbootstrap/*_test.go` for success, missing `--profile`, unknown profile, and invalid catalog path/input.
- [x] 3.2 Assert exact stdout/stderr text using `bytes.Buffer`, including exit codes and deterministic ordering.
- [x] 3.3 Verify the command boundary calls the adapter and planner without duplicating planning rules in CLI code.

## Phase 4: Cleanup / Documentation

- [x] 4.1 Update README minimal status/usage text for `dbootstrap plan --profile <name>` if the current docs do not mention it.
- [x] 4.2 Keep generated strings/comments in English and remove any temporary debug scaffolding before finishing.
