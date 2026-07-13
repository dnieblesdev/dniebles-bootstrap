# Tasks: APT/dpkg Package Idempotency

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 550–750 (implementation, tests, and wiring) |
| Requested 800-line review budget risk | Medium |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 detector; PR 2 runner guards; PR 3 CLI composition |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Add injectable APT detector and classifier | PR 1 | Detector plus its unit tests; independently verifiable |
| 2 | Add execution guards for APT presence states | PR 2 | Targets main after PR 1; runner tests included |
| 3 | Wire confirmed-Linux apply/bootstrap composition | PR 3 | Targets main after PR 2; CLI tests and full regression |

## Phase 1: Detector Contract (RED)

- [x] 1.1 Create `internal/state/apt_package_detector_test.go` with table-driven failing cases for `* ok installed` (including `hold ok installed`), `unpacked`, `half-configured`, exact exit-1 not-found, contradictory stdout, malformed/empty output, unavailable command, runner error, and timeout.
- [x] 1.2 Add failing tests for exact `dpkg-query` request arguments, one probe per eligible APT step, ineligible/provider-isolated steps, and immutable plan-copy decoration.

## Phase 2: Detector Implementation (GREEN/REFACTOR)

- [x] 2.1 Create `internal/state/apt_package_detector.go` with injected `CommandExists`, `CommandRunner`, timeout default, strict three-field classifier, exact absence signature, and `Detect`/`ApplyAptPackagePresence` APIs.
- [x] 2.2 Refactor detector seams and classifier helpers for clear provider eligibility, no retries/fallbacks, and deterministic `installed`/`absent`/`unknown` outcomes; run `gofmt` and focused state tests.
- [x] 2.3 Update the `PackagePresence` comment in `internal/planning/types.go` to describe provider-specific transient presence.
- [x] 2.4 Fix CRITICAL R3-001: validate parsed dpkg status fields against allowed definitive states so malformed/ambiguous three-field output is `unknown`, never `absent`/dispatch; add detector test cases.

## Phase 3: Ordered Execution (RED → GREEN)

- [x] 3.1 Extend `internal/execution/runner_test.go` with failing tests proving installed/held skips, partial and definitive absence dispatch, unknown failure without installer calls, original ordering, and Brew/non-APT isolation.
- [x] 3.2 Modify `internal/execution/runner.go` with isolated APT eligibility, installed-skip, and unknown-fail guards; keep absent states on the normal `AptInstaller` path.

## Phase 4: CLI Wiring (RED → GREEN)

- [x] 4.1 Add failing `cmd/dbootstrap/main_test.go` coverage for confirmed Linux `apply --yes` and `bootstrap`, plus safe/default, dry-run, planning-only, and non-Linux no-probe guarantees.
- [x] 4.2 Modify `cmd/dbootstrap/main.go` to compose APT detection after Brew only for confirmed Linux eligible plans and preserve execution-plan copy isolation.

## Phase 5: Verification

- [x] 5.1 Run focused `go test ./internal/state ./internal/execution ./cmd/dbootstrap`, then `go test ./...`, `go vet ./...`, and inspect the diff against the requested 800-line review budget.
- [x] 5.2 Verify no `sudo`, `apt-get`, fallback, retry, or mutation occurs during detection; record any skipped external-command integration coverage.
