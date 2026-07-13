# Tasks: Minimize Release Workflow Permissions

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 1 workflow line; ~20 validation lines reviewed |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR: permission deletion plus validation evidence |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | Remove unused global write permission and prove the barrier remains intact | PR 1 | Single focused PR; no release dispatch or tag creation |

## Phase 1: Workflow Permission Change

- [x] 1.1 Modify `.github/workflows/release-publish.yml` by deleting only the global `actions: write` entry; preserve `contents: read`, triggers, inputs, jobs, and job-level permissions.
- [x] 1.2 Review the resulting effective scopes: `validate` and reusable `build` remain read-only; `publish` alone retains `contents: write` and `actions: read`.

## Phase 2: Static Validation

- [x] 2.1 Parse `.github/workflows/release-publish.yml` with the repository's available YAML/CI validation tooling and confirm the workflow is syntactically valid.
- [x] 2.2 Verify the invalid-version barrier: `validate` still runs `go run ./internal/version/cmd/validate --release`, while both `build` and `publish` retain `needs` dependencies that prevent them from starting after invalid or unprefixed input (`1.2.3`, `v1`, invalid SemVer).
- [x] 2.3 Confirm no release dispatch, tag creation, asset publication, or changes to `release-build.yml` are introduced; this slice is validation-only for the no-release path.

## Phase 3: Verification

- [x] 3.1 Run focused repository checks for YAML/workflow validation and the existing release-tag validator tests; record failures as blockers.
- [x] 3.2 Run `go test ./...` per `openspec/config.yaml` when the focused checks pass, then inspect the final diff to ensure it contains only the approved one-line workflow deletion and task artifact.
