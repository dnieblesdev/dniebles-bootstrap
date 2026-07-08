# Tasks: Catalog Brew Provider Targets

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 20-60 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | exception-ok |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Catalog metadata plus focused regression coverage | PR 1 | Single micro-slice; no chaining needed. |

## Phase 1: Catalog Metadata

- [x] 1.1 In `catalog/bootstrap.toml`, change only `package:ripgrep` install provider from `apt` to `brew`.
- [x] 1.2 Keep `package:ripgrep` package as `ripgrep` and presence as `command_exists: rg`.
- [x] 1.3 Confirm `tool:git` remains `apt` and `runtime:go` remains `asdf`.

## Phase 2: Regression Coverage

- [x] 2.1 In `internal/catalog/toml/catalog_test.go`, assert default catalog ripgrep install metadata is `brew`/`ripgrep`.
- [x] 2.2 Assert ripgrep presence metadata remains `command_exists` with name `rg`.
- [x] 2.3 Assert git and Go install providers remain unchanged.

## Phase 3: Verification

- [x] 3.1 Run the focused catalog test package covering TOML decode behavior.
- [x] 3.2 Run the minimal relevant Go test command if focused coverage touches shared helpers.
- [x] 3.3 Inspect the diff to verify no command runner, installer, apply wiring, bootstrap, dotfile, or entrypoint files changed.
