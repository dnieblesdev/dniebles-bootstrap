# Tasks: CI Build Validation

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 25–40 |
| 800-line budget risk | Low |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | auto-chain |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | Add and verify the complete CI validation gate | PR 1 | Single self-contained workflow change; tests are static/CI verification. |

## Phase 1: Workflow Foundation

- [x] 1.1 Create `.github/workflows/build.yml` with `push` and `pull_request` triggers limited to `main`, using `ubuntu-latest`.
- [x] 1.2 Add checkout and `actions/setup-go@v5` with `go-version-file: go.mod`; keep one sequential `build` job.

## Phase 2: Validation Implementation

- [x] 2.1 Add ordered fail-fast run steps for `go test ./...`, `go vet ./...`, and `go build ./...`.
- [x] 2.2 Confirm the workflow contains no artifact upload, publication, signing, release, caching, matrix, or packaging behavior.

## Phase 3: Testing / Verification

- [x] 3.1 Static-review `.github/workflows/build.yml` against trigger filters, Go-version sourcing, command order, and artifact restrictions.
- [x] 3.2 Open or update a PR targeting `main` and verify the GitHub job runs all three commands successfully in order.
- [x] 3.3 Verify a failed command produces a failed check and prevents later validation steps; confirm no artifacts are uploaded or published.
