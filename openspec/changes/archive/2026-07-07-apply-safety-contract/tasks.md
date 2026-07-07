# Tasks: Apply Safety Contract

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~180-260 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | exception-ok |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Lock apply safety semantics in CLI and renderer | PR 1 | Base: current branch; include behavior tests and output checks. |

## Phase 1: CLI Safety Contract Foundation

- [x] 1.1 Add apply mode parsing in `cmd/dbootstrap/main.go` for default, `--dry-run`, and reserved `--yes`.
- [x] 1.2 Reject `--dry-run --yes` in `cmd/dbootstrap/main.go` with a clear usage error before planning or execution.
- [x] 1.3 Keep plan validation and existing noop apply flow unchanged when flags are accepted.

## Phase 2: Execution Output Reporting

- [x] 2.1 Update `cmd/dbootstrap/render.go` to print the selected apply mode in the execution report.
- [x] 2.2 Ensure the mode line distinguishes default non-mutating, dry-run, and confirmed-future noop states.
- [x] 2.3 Preserve current no-op execution behavior for all accepted apply modes; do not wire real mutation paths.

## Phase 3: Tests and Verification

- [x] 3.1 Extend `cmd/dbootstrap/main_test.go` for default apply, `--dry-run`, `--yes`, and `--dry-run --yes` conflict rejection.
- [x] 3.2 Assert conflict rejection produces no execution report and no execution path is reached.
- [x] 3.3 Update `cmd/dbootstrap/render_test.go` to verify the mode line is rendered for each accepted mode.

## Phase 4: Safety Cleanup

- [x] 4.1 Confirm no changes are introduced under `internal/execution/` beyond existing noop contracts.
- [x] 4.2 Verify the change does not add real installers, `CommandRunner` mutation, Homebrew bootstrap, remote scripts, raw command metadata, dotfiles execution, or bootstrap entrypoint wiring.
