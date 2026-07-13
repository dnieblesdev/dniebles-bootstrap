# Apply Progress: GitHub Release Publishing

## Mode
Strict TDD (strict_tdd: true; go test runner available)

## Work Unit
PR 3 — Remote prerelease and failure-barrier evidence. Builds on PR 2. One disposable prerelease tag created; no commits.

## Completed Tasks

- [x] 1.1 Add RED table tests in `internal/version/validate_test.go` for stable, prerelease, build metadata, unprefixed, partial, leading-zero, and malformed SemVer inputs.
- [x] 1.2 Add `ValidateReleaseTag` and prerelease classification in `internal/version/validate.go`; preserve permissive `Validate` behavior and expose release mode in `internal/version/cmd/validate/main.go`.
- [x] 1.3 Modify `.github/workflows/release-build.yml` for typed `workflow_call` version/output inputs, consolidated artifact output, retained manual dispatch, and `contents: read` only.
- [x] 1.4 Verify `go test ./...`, `go vet ./...`, and direct/manual workflow behavior remains artifact-only.
- [x] 2.1 Create `.github/workflows/release-publish.yml` with manual dispatch, strict validation before side effects, reusable build invocation, and read-only defaults.
- [x] 2.2 Add publish-job artifact download and exact six-file allowlist for the three archives plus matching `.sha256` files derived from `safe_version`.
- [x] 2.3 Add strict checksum verification and fail-fast barriers for missing/extra/tampered assets and unsuccessful called builds; prove no release upload occurs on failure.
- [x] 2.4 Add remote tag/release existence guards, `contents: write` only on publish, prerelease flagging, and explicit `gh release create` asset paths.
- [x] 3.1 Review workflow permissions and YAML contracts; run the full Go suite and capture focused validator output.
- [x] 3.2 Dispatch a unique `vX.Y.Z-rc.N`, record run URL, input tag, prerelease state, and exactly six remote assets as success evidence.
- [x] 3.3 Exercise invalid/unprefixed input barrier; record failed run proving no tag/release/upload mutation.
- [x] 3.4 Re-dispatch the successful prerelease tag; record existing-tag/release failure and unchanged remote state.

## TDD Cycle Evidence

| Task | Test File / Evidence | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|----------------------|-------|------------|-----|-------|-------------|----------|
| 1.1 | `internal/version/validate_test.go` | Unit | ✅ baseline `go test ./internal/version/...` passed | ✅ Written (undefined `ValidateReleaseTag`) | ✅ Passed | ✅ 22 table cases | ✅ Clean |
| 1.2 | `internal/version/cmd_validate_test.go` | Unit | ✅ 1.1 tests still green | ✅ Written (`-release` flag undefined) | ✅ Passed | ✅ 5 table cases | ✅ Clean |
| 1.3 | N/A — structural YAML contract | N/A | N/A | ➖ Structural | ✅ YAML valid | ➖ Skipped (single output contract) | ➖ None needed |
| 1.4 | Existing + new suites | Unit/Integration | ✅ All prior tests green | ➖ Verification | ✅ `go test ./...` and `go vet ./...` pass | ➖ N/A | ➖ N/A |
| 2.1 | N/A — structural YAML workflow | N/A | ✅ Existing Go suite still green | ➖ Structural | ✅ YAML valid; workflow dispatch + reusable call reviewed | ➖ Skipped (structural contract) | ➖ None needed |
| 2.2 | N/A — structural YAML workflow | N/A | ✅ Existing Go suite still green | ➖ Structural | ✅ Allowlist + download contract reviewed | ➖ Skipped (structural contract) | ➖ None needed |
| 2.3 | N/A — structural YAML workflow | N/A | ✅ Existing Go suite still green | ➖ Structural | ✅ Checksum + fail-fast barriers reviewed | ➖ Skipped (structural contract) | ➖ None needed |
| 2.4 | N/A — structural YAML workflow | N/A | ✅ Existing Go suite still green | ➖ Structural | ✅ Existence guards + prerelease + explicit asset paths reviewed | ➖ Skipped (structural contract) | ➖ None needed |
| 3.1 | `go test ./...`, `go vet ./...`, `go run ./internal/version/cmd/validate` | Verification | ✅ Existing Go suite green before dispatch | ✅ Scenarios defined in spec as acceptance criteria | ✅ All pass; validator accepts `v0.0.0-rc.1` and rejects `1.2.3` | ➖ Inspection-based (no branching logic) | ➖ None needed |
| 3.2 | Remote workflow dispatch + `gh release view v0.0.0-rc.1` | E2E / Remote | ✅ Tag/release did not exist before dispatch | ✅ Expected success scenario from spec | ✅ Run 29236008116 succeeded; six assets uploaded; prerelease=true | ➖ Single success path | ➖ None needed |
| 3.3 | Remote workflow dispatch with `version=1.2.3` | E2E / Remote | ✅ No existing release/tag for invalid input | ✅ Expected failure scenario from spec | ✅ Run 29235983110 failed at `Validate release tag`; no tag/release created | ➖ Single failure path | ➖ None needed |
| 3.4 | Remote re-dispatch with `version=v0.0.0-rc.1` | E2E / Remote | ✅ Release/tag created in 3.2 | ✅ Expected non-overwrite scenario from spec | ✅ Run 29236121713 failed at `Guard existing tag and release`; original release unchanged | ➖ Single failure path | ➖ None needed |

### Test Summary
- **Total tests written this batch**: 0 (remote verification only)
- **Total tests passing**: 27 (existing Go suite)
- **Layers used this batch**: E2E/Remote (3 workflow dispatches), Verification (Go suite + validator CLI)
- **Approval tests**: None — no refactoring of existing behavior
- **Pure functions created this batch**: None

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `.github/workflows/release-publish.yml` | Created (PR 2) | Manual-dispatch publish workflow: strict tag validation, reusable `release-build` call, exact six-file allowlist, SHA-256 verification, tag/release non-overwrite guards, prerelease flagging, and `contents: write` scoped to the publish job only. |
| `openspec/changes/github-release-publishing/tasks.md` | Modified | Marked Phase 3 tasks 3.1–3.4 complete. |
| `openspec/changes/github-release-publishing/apply-progress.md` | Created | Merged PR 1–PR 2 progress with PR 3 remote evidence. |
| `openspec/changes/github-release-publishing/verify-report.md` | Created | Detailed verification evidence for Go suite, validator, and three remote dispatches. |

## Deviations from Design

None — implementation matches design. The publish workflow uses `gh release create` (GitHub CLI) as the Release API choice per the design decision table, with an explicit `--target` and empty `--notes` to avoid changelog generation.

## Issues Found

None.

## Remaining Tasks

None. All 12 tasks complete.

## Workload / PR Boundary

- **Mode**: stacked-to-main (auto-chain)
- **Current work unit**: PR 3 — Remote evidence
- **Boundary**: No product code changes; only SDD evidence artifacts updated. One disposable prerelease `v0.0.0-rc.1` was created and left in place as evidence.
- **Estimated review budget impact**: ~0 changed production lines; only markdown evidence files added.

## Status
12/12 tasks complete. Ready for sdd-verify.
