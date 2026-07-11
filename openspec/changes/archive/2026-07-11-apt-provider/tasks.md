# Tasks: APT Provider

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 650–780 |
| 400-line budget risk | High |
| 800-line risk | Medium |
| Chained PRs recommended | No — approved single-delivery size exception |
| Suggested split | One PR: implementation, tests, fixture, and reporting |
| Delivery strategy | exception-ok |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: High
800-line risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | Complete the approved APT slice | Single PR | Include verification; maintainer-approved size exception; no catalog migration |

## Phase 1: Contracts and Flag Foundation

- [x] 1.1 Add APT command/error contracts in `internal/execution/apt_installer.go`, including the ten-minute timeout, typed non-Linux outcome, and direct/sudo request shapes.
- [x] 1.2 RED: add table-driven `--sudo` parsing tests in `cmd/dbootstrap/main_test.go`; reject `--sudo` unless `--yes` is present and preserve existing mode parsing.
- [x] 1.3 GREEN: update `cmd/dbootstrap/main.go` and usage text to carry explicit sudo state without changing default catalog or non-APT flags.

## Phase 2: APT Installer and Provider Routing

- [x] 2.1 RED: create `internal/execution/apt_installer_test.go` covering provider gating, trimming, empty/`-`-prefixed validation, missing `apt-get`/`sudo`, exact vectors, `-y --`, ten-minute timeout, failures, and timeouts.
- [x] 2.2 GREEN: implement `AptInstaller` with injected availability/runner seams; never use shell strings, fallback, retry, rollback, bootstrap, update, repository changes, or presence detection.
- [x] 2.3 RED/GREEN: extend `internal/execution/provider_aware_installer.go` and `provider_aware_installer_test.go` with fixed brew-or-APT routing for tool/package kinds and a non-Linux rejecting delegate that performs zero probes/calls.

## Phase 3: Confirmed Composition and Reporting

- [x] 3.1 RED/GREEN: wire Linux-only APT composition in `cmd/dbootstrap/main.go`, keeping Homebrew cross-platform behavior and lazy command-runner construction intact.
- [x] 3.2 Use a `t.TempDir()` opt-in custom catalog fixture for APT CLI coverage; leave `catalog/bootstrap.toml` unchanged.
- [x] 3.3 Update `cmd/dbootstrap/render.go` and tests so APT failures, timeout status, unsupported OS, and confirmed non-zero outcomes are visible without claiming rollback.

## Phase 4: Verification and Cleanup

- [x] 4.1 Add CLI tests for fixture-backed direct/sudo execution, invalid sudo, default/dry-run zero probing, non-Linux zero calls, missing `apt-get`/`sudo`, command failure, timeout rendering, and confirmed non-zero outcomes.
- [x] 4.2 REFACTOR: keep fakes table-driven, use a `t.TempDir()` custom catalog, preserve existing Homebrew/noop regressions, and document the explicit non-goals in relevant test comments.
- [x] 4.3 Run focused Go tests, then one fresh evidence-backed verification after the corrective code/test changes, including tracked and untracked diff hygiene.
- [x] 4.4 RED/GREEN: correct general apply help and exact-output tests to describe eligible Linux APT direct `--yes` and explicit `--yes --sudo` execution without changing semantics.
