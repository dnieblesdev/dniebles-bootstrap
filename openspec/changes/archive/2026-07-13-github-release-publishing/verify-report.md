# Verification Report: github-release-publishing

**Change**: github-release-publishing
**Version**: v0.0.0-rc.1 (disposable prerelease evidence tag)
**Mode**: Strict TDD (strict_tdd: true; go test runner available)
**Date**: 2026-07-13
**Persistence**: hybrid (openspec file + Engram)
**Verifier evidence basis**: SDD specs + tasks + design + current remote GitHub evidence (re-queried live this run)

## Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 12 (1.1–1.4, 2.1–2.4, 3.1–3.4) |
| Tasks complete | 12 |
| Tasks incomplete | 0 |

All 12 tasks checked in `tasks.md` and `apply-progress.md`. No unchecked implementation tasks.

## Build & Tests Execution

**Build**: ✅ Passed
```text
$ go build ./...
(exits 0; full suite compiled via `go test ./...` and `go vet ./...`)
```

**Tests**: ✅ 0 failed / 0 skipped — fresh execution (`-count=1`, not cached)
```text
$ go test -count=1 ./...
ok  	github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/config
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/diameter
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/environment
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/execution
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/planning
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/state
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/version   1.488s
?   	github.com/dnieblesdev/dniebles-bootstrap/internal/version/cmd/normalize  [no test files]
?   	github.com/dnieblesdev/dniebles-bootstrap/internal/version/cmd/validate   [no test files]

$ go vet ./...   → rc=0
```

**Coverage**: `internal/version` package 91.4% of statements → ✅ Above threshold
```text
$ go test -count=1 -coverprofile=/tmp/cover.out ./internal/version/...
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/version	1.669s	coverage: 91.4% of statements
```

**Validator CLI (focused)** — re-executed this run:
```text
$ go run ./internal/version/cmd/validate --release --version v0.0.0-rc.1   → prerelease=true   rc=0
$ go run ./internal/version/cmd/validate --release --version v1.2.3       → prerelease=false  rc=0
$ go run ./internal/version/cmd/validate --release --version v1.2.3+build.123 → prerelease=false rc=0
$ go run ./internal/version/cmd/validate --release --version 1.2.3        → "release tag \"1.2.3\" is not a valid v-prefixed SemVer tag"  rc=1
$ go run ./internal/version/cmd/validate --release --version v1.2        → "release tag \"v1.2\" is not a valid v-prefixed SemVer tag"    rc=1
```

## Spec Compliance Matrix

Scenarios sourced from `specs/github-release-publishing/spec.md` (new capability) and `specs/release-binary-builds/spec.md` (delta). Compliance = covering runtime evidence + source inspection.

| # | Requirement | Scenario | Covering evidence (runtime + source) | Result |
|---|-------------|----------|--------------------------------------|--------|
| 1 | Validate the release tag | Valid stable version (`v1.2.3`) | CLI `v1.2.3`→`prerelease=false` (live); `TestValidateReleaseTag/stable`, `TestValidateCmdRelease/stable`; workflow passes same value to both `workflow_call` `version` input and `gh release create "${VERSION}"` | ✅ COMPLIANT |
| 2 | Validate the release tag | Invalid or unprefixed version | Run 29235983110 failed at step `validate > Validate release tag` (`release tag "1.2.3" is not a valid v-prefixed SemVer tag`, exit 1); `build`+`publish` jobs SKIPPED (live). `TestValidateReleaseTag/{unprefixed,partial*,leading-zero*,empty,too-long,invalid-chars,...}`, `TestValidateCmdRelease/{unprefixed,partial}` | ✅ COMPLIANT |
| 3 | Publish the called build outputs | Verified assets are published (exactly 3 archives + 3 checksums) | Run 29236008116 success; Release `v0.0.0-rc.1` has exactly 6 assets (live `gh release view`): 3 archives + 3 `.sha256`, names/sizes match `safe_version` derivation | ✅ COMPLIANT |
| 4 | Publish the called build outputs | Verification fails → no release created | Checksum step `sha256sum --check --strict` (release-publish.yml L97) runs BEFORE `gh release create` (L123) under `set -euo pipefail`; the fail-fast mechanism is proven at runtime (run 29236121713: a `publish`-job step failure prevented any release mutation). The exact tampered-checksum path was not injected in a remote run. | ⚠️ PARTIAL |
| 5 | Restrict publication authority | Permissions are inspected → only publish job has `contents: write` | Source inspection: `release-build.yml` `permissions: contents: read` (L27-28); `release-publish.yml` top-level `contents: read` (L11-13), `validate` job `contents: read` (L18-19), `publish` job `contents: write`+`actions: read` (L48-50). The scenario explicitly defines inspection as the test. | ✅ COMPLIANT |
| 6 | Prevent overwrites / capture evidence | Prerelease evidence | Release `v0.0.0-rc.1` `isPrerelease: true` (live); validate job outputs `prerelease` and publish job sets `--prerelease` conditionally | ✅ COMPLIANT |
| 7 | Prevent overwrites / capture evidence | Existing release is protected | Run 29236121713 failed at `publish > Guard existing tag and release` (live log shows the `gh api .../git/refs/tags/${VERSION}` + `gh release view` guard); release `v0.0.0-rc.1` unchanged (`publishedAt` 2026-07-13T08:37:17Z, same 6 assets, same target commit `9d78c38`) | ✅ COMPLIANT |
| 8 | Preserve scope boundaries | No scope creep → only GitHub Release created | Source inspection: neither workflow signs, generates changelogs, publishes to package managers, or auto-triggers; `release-publish.yml` only calls `gh release create` with `--notes ""` | ✅ COMPLIANT |
| 9 | (delta) Support reusable verified builds | Publish workflow calls the build → 3 archives + 3 checksums | Run 29236008116: reusable `build` job (version, quality, 3 matrix builds, upload) succeeded; release exposes the 6 assets | ✅ COMPLIANT |
| 10 | (delta) Support reusable verified builds | Direct manual behavior remains unchanged (artifacts only) | `release-build.yml` retains `workflow_dispatch`, no `gh release`/tag step; only `actions/upload-artifact` with `if-no-files-found: error` | ✅ COMPLIANT |
| 11 | (delta) Support reusable verified builds | Called build fails → no successful complete bundle | `upload` job `needs: [version, build, quality]`; `upload-artifact` `if-no-files-found: error` (L131, L149); run 29235983110 shows build is skipped when validate fails | ✅ COMPLIANT |
| 12 | (delta) Exclude release publishing (MODIFIED) | Manual build does not publish | `release-build.yml` has no release/tag steps (source inspection) | ✅ COMPLIANT |
| 13 | (delta) Exclude release publishing (MODIFIED) | Reusable build does not publish | Run 29236008116: the release was created by the `publish` job, not the called `build` job; `build` only uploads artifacts | ✅ COMPLIANT |

**Compliance summary**: 12/13 scenarios COMPLIANT, 1 PARTIAL (scenario 4 — checksum-rejection path structurally enforced and fail-fast mechanism runtime-proven, but no remote run injects a tampered checksum).

## Correctness (Static Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| Strict `v`-SemVer validation before side effects | ✅ Implemented | `ValidateReleaseTag` (validate.go L42) + CLI `-release` mode; permissive `Validate` preserved unchanged (TestValidate still green) |
| Reusable build via `workflow_call` | ✅ Implemented | release-build.yml L10-25 typed input/outputs match design contract verbatim |
| Consolidated artifact consumed by name | ✅ Implemented | release-publish.yml downloads `${{ needs.build.outputs.artifact_name }}` |
| Exact six-file allowlist (no missing/extra) | ✅ Implemented | release-publish.yml L68-91 enumerates 6 expected files and asserts exact count |
| Checksum verification before release | ✅ Implemented | `sha256sum --check --strict` (L97) precedes `gh release create` (L123); both under `set -euo pipefail` |
| Tag + release existence guards | ✅ Implemented | release-publish.yml L103-111 checks both refs/tags and `gh release view` |
| Prerelease flagging | ✅ Implemented | L118-121 conditional `--prerelease` from validate job output |
| `contents: write` scoped to publish job | ✅ Implemented | Only `publish` job declares `contents: write` (L48-50) |
| No package/sign/changelog/auto-trigger | ✅ Implemented | Confirmed absent in both workflows |

## Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| Strict validator beside permissive `Validate` | ✅ Yes | `ValidateReleaseTag` added without altering `Validate`; TestValidate confirms permissive contract intact |
| Local reusable workflow via `workflow_call` | ✅ Yes | release-build.yml exposes `workflow_call`; release-publish.yml `uses: ./.github/workflows/release-build.yml` |
| Consume named consolidated artifact + outputs | ✅ Yes | `artifact_name`/`safe_version` outputs consumed; expected names derived from `safe_version` |
| `gh release create` after explicit guards | ✅ Yes | `gh release create` with `--target "${GITHUB_SHA}"` and `--notes ""` after tag/release guards (matches design's "no changelog generation") |
| `workflow_call` interface contract (yaml) | ✅ Yes | Outputs `version`/`safe_version`/`artifact_name` match design L43-53 verbatim |
| `ValidateReleaseTag(v) (isPrerelease bool, err error)` signature | ✅ Yes | validate.go L42 matches design L56 verbatim |
| `release-publish.yml` "defaults to read access" | ⚠️ Partial | Top-level declares `contents: read` (matches) **plus** `actions: write` (L13) not contemplated by the design's "read-only defaults". Inert in practice — both jobs override `permissions`, so `actions: write` is never inherited — but it deviates from the stated default. See WARNING. |

## TDD Compliance

| Check | Result | Details |
|-------|--------|---------|
| TDD Evidence reported | ✅ | `apply-progress.md` "TDD Cycle Evidence" table present (12 rows) |
| All tasks have tests | ✅ | Logic tasks (1.1, 1.2) have Go test files; structural YAML tasks (1.3, 2.1–2.4) explicitly marked structural; verification tasks (3.x) backed by remote E2E |
| RED confirmed (tests exist) | ✅ | `validate_test.go` (TestValidateReleaseTag) and `cmd_validate_test.go` (TestValidateCmdRelease) exist and compile |
| GREEN confirmed (tests pass) | ✅ | Fresh `go test -count=1 ./internal/version/...` PASS (1.488s) |
| Triangulation adequate | ✅ | 21 release-tag table cases + 5 cmd-release cases; covers stable/prerelease/build/unprefixed/partial/leading-zero/malformed. Minor: progress reported "22 table cases" for 1.1; actual is 21 (see SUGGESTION) |
| Safety Net for modified files | ✅ | `validate.go` modified; pre-existing `TestValidate` (19 cases) re-run green, preserving the permissive contract |

**TDD Compliance**: 6/6 checks passed

## Test Layer Distribution

| Layer | Tests | Files | Tools |
|-------|-------|-------|-------|
| Unit | 40 | 1 (`internal/version/validate_test.go`) | Go `testing` (in-process) |
| Integration | 8 | 1 (`internal/version/cmd_validate_test.go`) | Go `testing` via `exec.Command` (subprocess CLI) |
| E2E / Remote | 3 dispatches | 2 workflow files | GitHub Actions (`release-publish` runs) |
| **Total** | **48 Go cases + 3 remote dispatches** | **4** | |

## Changed File Coverage

| File | Line % | Branch % | Uncovered Lines | Rating |
|------|--------|----------|-----------------|--------|
| `internal/version/validate.go` | 91.4% (pkg) | — | `validateSemVerPrerelease` 77.8% (empty-id L74-76, invalid-char L77-79 error branches), `validateSemVerBuildMetadata` 66.7% (empty-id L92-94, invalid-char L95-97 error branches) | ✅ Excellent (file); ⚠️ two helpers <80% |
| `internal/version/cmd/validate/main.go` | ➖ | ➖ | Not measurable in-package — exercised via `exec.Command` subprocess in `cmd_validate_test.go`; the in-package coverprofile does not attribute subprocess execution | ➖ Tooling limitation, not a real gap |
| `.github/workflows/release-build.yml` | ➖ | ➖ | YAML structural contract; covered by remote dispatch + source inspection | ➖ N/A |
| `.github/workflows/release-publish.yml` | ➖ | ➖ | YAML structural contract; covered by 3 remote dispatches + source inspection | ➖ N/A |

**Average changed-Go-file coverage**: 91.4% (validate.go). Coverage analysis of `main.go` skipped for in-package attribution — subprocess-tested.

## Assertion Quality

Scanned `validate_test.go` and `cmd_validate_test.go`:

- No tautologies (`expect(true).toBe(true)`-style) — all assertions call `ValidateReleaseTag`/`Validate` or `exec.Command` and assert real error/prerelease values.
- No ghost loops; no type-only assertions used alone; no implementation-detail coupling (asserts on returned `error`/`bool`, not internals).
- Triangulation has real variance: assertions expect DIFFERENT outcomes (error vs nil, prerelease true vs false) across 21+5 cases.

**Assertion quality**: ✅ All assertions verify real behavior. 0 CRITICAL, 0 WARNING.

## Quality Metrics

**Linter / vet**: ✅ No errors — `go vet ./...` rc=0.
**Type Checker**: ➖ Not applicable (Go is statically typed; `go build ./...` compiles clean).

## Remote Evidence — Current State (re-queried live this run, 2026-07-13)

### Release `v0.0.0-rc.1` (live `gh release view`)
- `tagName`: v0.0.0-rc.1
- `isPrerelease`: **true**
- `publishedAt`: 2026-07-13T08:37:17Z
- `targetCommitish`: 9d78c3866ec4f4762a3ff0af709d5c42ba45ced5
- **Exactly 6 assets** (names + sizes match the persisted report byte-for-byte):

| Asset | Size |
|-------|------|
| dbootstrap_v0.0.0-rc.1_linux_amd64.tar.gz | 2,723,855 |
| dbootstrap_v0.0.0-rc.1_linux_amd64.tar.gz.sha256 | 108 |
| dbootstrap_v0.0.0-rc.1_linux_arm64.tar.gz | 2,490,275 |
| dbootstrap_v0.0.0-rc.1_linux_arm64.tar.gz.sha256 | 108 |
| dbootstrap_v0.0.0-rc.1_windows_amd64.zip | 2,766,608 |
| dbootstrap_v0.0.0-rc.1_windows_amd64.zip.sha256 | 107 |

### Workflow runs (live `gh run view`)
| Run ID | Scenario | Conclusion | Failed job | Failed step |
|--------|----------|------------|------------|-------------|
| 29235983110 | invalid input `1.2.3` | failure | `validate` | `Validate release tag` (`build`+`publish` skipped) |
| 29236008116 | success `v0.0.0-rc.1` | success | — | — (all incl. `publish` succeeded) |
| 29236121713 | duplicate tag `v0.0.0-rc.1` | failure | `publish` | `Guard existing tag and release` (release unchanged) |

All three runs are `workflow_dispatch` on `main`, `release-publish` workflow, `status: completed`.

## Issues Found

**CRITICAL**: None

**WARNING**:
1. **Top-level `permissions: actions: write` in `release-publish.yml` (L13) deviates from the design's "release-publish.yml defaults to read access".** Inert in practice — both jobs (`validate`, `publish`) declare their own `permissions`, so the top-level `actions: write` is never inherited by any job; effective job permissions remain least-privilege and the spec's "only publish job has `contents: write`" holds. Flagged for design-coherence: the stated default was read-only. Recommendation (for orchestrator/user): drop `actions: write` from the top level (or set `actions: read`) to match the design intent. No spec violation.

**SUGGESTION**:
1. **Scenario 4 ("Verification fails") is PARTIAL**: the checksum-rejection path is structurally enforced and ordered before release creation, and the fail-fast mechanism is runtime-proven by the analogous duplicate-tag failure, but no remote run injects a tampered/mismatched checksum to exercise the rejection end-to-end. Consider a dedicated tampered-checksum barrier run (e.g., corrupt one `.sha256` before the publish job) for full E2E confidence.
2. **Self-reported triangulation count off-by-one**: `apply-progress.md` TDD table reports "✅ 22 table cases" for task 1.1; `TestValidateReleaseTag` actually has 21 table cases. Triangulation is still excellent; correct the count for accuracy.
3. **Two validator helpers below 80% line coverage**: `validateSemVerPrerelease` (77.8%) and `validateSemVerBuildMetadata` (66.7%) — their error branches (empty identifier, invalid characters within an identifier) are not exercised because malformed inputs fail the regex before reaching these helpers. Consider cases like `v1.2.3-rc.1..2` or `v1.2.3+build@123` if the regex is loosened, or accept that the regex front-loads these rejections.

## Verdict

**PASS WITH WARNINGS**

All 12 tasks complete; Go suite + vet pass fresh (91.4% coverage on the changed package); the validator CLI behaves correctly for stable/prerelease/invalid input; and all three remote workflow runs plus the live release `v0.0.0-rc.1` (prerelease=true, exactly six assets, unchanged after the duplicate-tag failure) confirm the spec scenarios at runtime. 12/13 spec scenarios are COMPLIANT. The single WARNING is an inert top-level `actions: write` permission that deviates from the design's "read-only defaults" without breaking any spec; the PARTIAL checksum-rejection scenario and minor coverage/count notes are non-blocking SUGGESTIONs.
