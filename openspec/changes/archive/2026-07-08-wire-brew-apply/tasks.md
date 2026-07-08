# Tasks: Wire Brew Apply

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 260-420 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | Single PR with size exception |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Wire confirmed-mode brew execution behind `apply --yes` | PR 1 | `cmd/dbootstrap/main.go`, `internal/execution/*`; keep default/`--dry-run` noop with fake seams |
| 2 | Replace bootstrap guidance with official docs/manual wording | PR 1 | `cmd/dbootstrap/render.go`, `internal/execution/homebrew_bootstrap.go`; include render safety tests |

## Phase 1: Foundation / Safety Seams

- [x] 1.1 Add a provider-aware adapter in `internal/execution/provider_aware_installer.go` that delegates only `Install.Provider == "brew"` and otherwise returns `not_implemented`.
- [x] 1.2 Add constructor/test seams in `cmd/dbootstrap/main.go` so apply mode can select noop vs confirmed runners without instantiating real commands in default or `--dry-run`.

## Phase 2: Core Implementation

- [x] 2.1 Wire `apply --yes` in `cmd/dbootstrap/main.go` to register `HomebrewInstaller` only for brew-backed `tool`/`package` steps, using `NewOSCommandRunner()` and `brewCommandExists` behind the confirmed path.
- [x] 2.2 Add the missing-brew branch before installer execution so confirmed apply skips target installs, avoids `HomebrewInstaller`, and reports bootstrap guidance first.
- [x] 2.3 Update execution/reporting flow so default and `--dry-run` remain noop/non-mutating and confirmed mode warns that real `brew install` may run.
- [x] 2.4 Replace `internal/execution/homebrew_bootstrap.go` guidance text with official Homebrew docs/manual review wording only; remove executable remote-script copy.

## Phase 3: Testing / Verification

- [x] 3.1 Add CLI tests in `cmd/dbootstrap/main_test.go` proving default and `--dry-run` never reach real execution, `--yes` can reach brew-only paths, and `--dry-run --yes` is rejected.
- [x] 3.2 Add unit tests for `internal/execution/provider_aware_installer_test.go` covering brew delegation, unsupported providers, and missing metadata returning `not_implemented`.
- [x] 3.3 Add `internal/execution/homebrew_bootstrap_test.go` and `cmd/dbootstrap/render_test.go` assertions that guidance is advisory-only and contains no `/bin/bash`, `curl`, `sh -c`, pipes, or raw install one-liners.
- [x] 3.4 Verify brew-missing confirmed-mode behavior with fakes so no real brew invocation occurs and bootstrap guidance remains the primary result.

## Phase 4: Cleanup / Documentation

- [x] 4.1 Update comments/help text in `cmd/dbootstrap/main.go` and render paths to match the new confirmed-mode semantics.
- [x] 4.2 Remove any obsolete noop-only wording that now conflicts with confirmed-mode wiring.
