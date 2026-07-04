# Tasks: Wire Installation State into CLI Plan

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~60-90 |
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
| 1 | Wire CLI state detection and keep tests deterministic | PR 1 | Base on main; include verification and README only if needed |

## Phase 1: CLI Wiring / Foundation

- [x] 1.1 Add `internal/state` import in `cmd/dbootstrap/main.go` and define `detectInstallationState = state.Detect` beside the existing environment seam.
- [x] 1.2 Call installation-state detection in `runPlan` after catalog load and before `planning.BuildPlan`.
- [x] 1.3 Pass the detected `planning.InstallationState` into `planning.BuildPlan` instead of `planning.InstallationState{}`.

## Phase 2: Test Seam / Output Coverage

- [x] 2.1 Add `stubInstallationState(t, planning.InstallationState)` in `cmd/dbootstrap/main_test.go` mirroring the environment-facts seam.
- [x] 2.2 Update existing plan tests to stub empty installation state so current success and failure output stays host-independent.
- [x] 2.3 Add an exact-output plan test that stubs `tool:git` as present and asserts `already_installed` appears in stdout.

## Phase 3: Verification

- [x] 3.1 Run `go test ./... -count=1` and confirm the CLI output tests and package tests pass together.
- [x] 3.2 Verify catalog-load failures still skip detection by keeping existing load-error coverage unchanged.

## Phase 4: Cleanup / Documentation

- [x] 4.1 Update README only if the CLI plan behavior documentation needs a minimal note about detected installation state.
