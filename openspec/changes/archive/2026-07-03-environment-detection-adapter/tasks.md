# Tasks: Environment Detection Adapter

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 260-380 |
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
| 1 | Build host-detection adapter, parser, CLI wiring, tests, and README note | PR 1 | Base on main; keep adapter, CLI integration, tests, and docs together as one reviewable work unit. |

## Phase 1: Foundation / Infrastructure

- [x] 1.1 Create `internal/environment/detector.go` with injectable runtime, env, and file sources returning `planning.EnvironmentFacts`.
- [x] 1.2 Create `internal/environment/osrelease.go` to parse `os-release` content conservatively and extract distro ID only.
- [x] 1.3 Add package-level defaults for runtime/OS-release/WSL signal reads, with unreadable optional inputs falling back to empty facts.

## Phase 2: Core Implementation

- [x] 2.1 Implement WSL detection in `internal/environment/detector.go` from env signals and `/proc`/kernel text, setting `WSL` only on positive evidence.
- [x] 2.2 Map detected OS, arch, distro, and WSL into `planning.EnvironmentFacts` without inventing missing values.
- [x] 2.3 Replace any hardcoded planning facts in `cmd/dbootstrap/main.go` with a detection seam used by `runPlan`.

## Phase 3: Testing / Verification

- [x] 3.1 Add table-driven tests in `internal/environment/detector_test.go` for OS/arch mapping, distro parsing, WSL true/false, and missing-file fallback.
- [x] 3.2 Add host-independent CLI tests in `cmd/dbootstrap/main_test.go` that override detection and assert plan output uses detected facts.
- [x] 3.3 Run `go test ./... -count=1` and verify no tests depend on the current host environment.

## Phase 4: Cleanup / Documentation

- [x] 4.1 Update README wording only if needed to reflect that `plan` now uses detected host facts.
- [x] 4.2 Remove any temporary helpers or test-only scaffolding added for the detection seam.
