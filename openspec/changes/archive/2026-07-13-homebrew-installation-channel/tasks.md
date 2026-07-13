# Tasks: Homebrew Installation Channel

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | ~150 (resolver only) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR — resolver foundation |
| Delivery strategy | single-pr |
| Chain strategy | N/A |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: N/A
400-line budget risk: Low

## Change Status

**COMPLETED technical slice.** Phase 1 tasks 1.1–1.3 are complete. Tasks 2.1–4.2 (documentation, tap formula creation, stable-release gate, lifecycle evidence, and final verification) have been moved to [`publish-homebrew-stable-channel`](openspec/changes/publish-homebrew-stable-channel/tasks.md). Do not add new implementation tasks to this change.

## Phase 1: Resolver TDD Foundation

- [x] 1.1 RED: extend `cmd/dbootstrap/main_test.go` table cases for explicit, XDG, home-local, and `<HOMEBREW_PREFIX>/share/dbootstrap/catalog/bootstrap.toml` precedence, absent prefix, and no existing candidates using `t.TempDir()` seams.
- [x] 1.2 GREEN: modify `cmd/dbootstrap/main.go` `catalogPathResolver` with `PathExists`, environment/home defaults, and the last-resort Homebrew candidate; preserve existing missing-catalog diagnostics and CLI flag precedence.
- [x] 1.3 REFACTOR: simplify resolver helpers/comments in `cmd/dbootstrap/main.go`, run focused tests, then `go test ./...`.
