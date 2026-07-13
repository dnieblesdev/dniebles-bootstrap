# Tasks: GitHub Release Publishing

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | ~800 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 foundation → PR 2 publishing → PR 3 remote evidence |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | Strict tag validation and reusable build contract | PR 1 | Foundation; verify Go tests and workflow syntax. |
| 2 | Isolated checksum-verified GitHub publishing | PR 2 | Base on PR 1; no release side effects before barriers. |
| 3 | Remote prerelease and failure-barrier evidence | PR 3 | Base on PR 2; use a unique disposable prerelease tag. |

## Phase 1: Validation and Build Contract

- [x] 1.1 Add RED table tests in `internal/version/validate_test.go` for stable, prerelease, build metadata, unprefixed, partial, leading-zero, and malformed SemVer inputs.
- [x] 1.2 Add `ValidateReleaseTag` and prerelease classification in `internal/version/validate.go`; preserve permissive `Validate` behavior and expose release mode in `internal/version/cmd/validate/main.go`.
- [x] 1.3 Modify `.github/workflows/release-build.yml` for typed `workflow_call` version/output inputs, consolidated artifact output, retained manual dispatch, and `contents: read` only.
- [x] 1.4 Verify `go test ./...`, `go vet ./...`, and direct/manual workflow behavior remains artifact-only.

## Phase 2: Publish Workflow and Barriers

- [x] 2.1 Create `.github/workflows/release-publish.yml` with manual dispatch, strict validation before side effects, reusable build invocation, and read-only defaults.
- [x] 2.2 Add publish-job artifact download and exact six-file allowlist for the three archives plus matching `.sha256` files derived from `safe_version`.
- [x] 2.3 Add strict checksum verification and fail-fast barriers for missing/extra/tampered assets and unsuccessful called builds; prove no release upload occurs on failure.
- [x] 2.4 Add remote tag/release existence guards, `contents: write` only on publish, prerelease flagging, and explicit `gh release create` asset paths.

## Phase 3: Verification and Evidence

- [ ] 3.1 Review workflow permissions and YAML contracts; run the full Go suite and capture focused validator output.
- [ ] 3.2 Dispatch a unique `vX.Y.Z-rc.N`, record run URL, input tag, prerelease state, and exactly six remote assets as success evidence.
- [ ] 3.3 Exercise invalid/unprefixed input and checksum/build failure barriers; record failed runs proving no tag/release/upload mutation.
- [ ] 3.4 Re-dispatch the successful prerelease tag; record existing-tag/release failure and unchanged remote state.
