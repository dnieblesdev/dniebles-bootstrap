# Tasks: Catalog More Brew Targets

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 60–110 |
| 400-line budget risk | Low |
| 800-line budget assessment | Low; well below the review ceiling |
| Chained PRs recommended | No |
| Suggested split | Single PR: catalog metadata, contract tests, verification |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low
800-line budget risk: Low

Maintainer-approved size exception: deliver as one PR; keep the catalog change and its contract tests in the same review unit.

## Phase 1: Catalog Foundation

- [x] 1.1 Modify `catalog/bootstrap.toml` to add the existing-schema `package:jq` record with `provider = "brew"`, package `jq`, and `command_exists` presence name `jq`.
- [x] 1.2 Append `package:jq` to the existing `bundle:cli`; preserve all other resources, bundles, profiles, and ordering semantics.

## Phase 2: Contract Tests

- [x] 2.1 Update `internal/catalog/toml/catalog_test.go` and its expected catalog model with the jq resource, including brew install and command-presence metadata.
- [x] 2.2 Extend `TestLoadFileAndBuildPlanFromFixture` assertions for `bundle:cli`, `profile:dev` selection, deterministic plan order/status, and metadata retention without command execution.
- [x] 2.3 Assert the pre-existing catalog remains intact and no fallback or multi-provider metadata is introduced.

## Phase 3: Verification and Rollback Readiness

- [x] 3.1 Run `go test ./internal/catalog/toml` and confirm the focused real-catalog contract scenarios pass without Homebrew or apply execution.
- [x] 3.2 Run `go test ./...`, `go vet ./...`, and formatting checks required by `openspec/config.yaml`; record results and any skipped external-command scope.
- [x] 3.3 Review the diff against the approved out-of-scope boundaries: no runtime, dotfile, provider, runner, command-execution, apply, bootstrap, fallback, platform-selection, bundle/profile, or migration changes.
- [x] 3.4 Confirm rollback removes the jq catalog and bundle entries together with their matching assertions; persist this task artifact to Engram key `sdd/catalog-more-brew-targets/tasks` and retain the OpenSpec file.

## Phase 4: Artifact Handling

- [x] 4.1 Keep `openspec/changes/catalog-more-brew-targets/tasks.md` and Engram `sdd/catalog-more-brew-targets/tasks` synchronized in English; do not implement code, commit, or create a PR during task planning.
