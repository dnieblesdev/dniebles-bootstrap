# Verification Report: minimize-release-workflow-permissions

**Change**: minimize-release-workflow-permissions
**Version**: N/A (delta spec, no capability version bump)
**Mode**: Strict TDD (config `strict_tdd: true`, runner `go test ./...`) — assessed for a zero-product-code slice; see TDD Compliance notes
**Date**: 2026-07-13
**Persistence**: hybrid (openspec file + Engram)

## Goal of this verification

Confirm the unused global `actions: write` grant was removed, least-privilege scopes are intact, and the invalid-version barrier still fails at `validate` and blocks `build`/`publish` (no release) for remotely-dispatched invalid input.

## Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 7 |
| Tasks complete | 7 |
| Tasks incomplete | 0 |

### Task Status

| Task | Status | Evidence |
|------|--------|----------|
| 1.1 Delete only global `actions: write`; preserve `contents: read`, triggers, inputs, jobs, job-level permissions | ✅ Complete | `git show a5e9df4` shows exactly one removed line (`-  actions: write`). Global block is now `permissions:\n  contents: read` (release-publish.yml:11-12). |
| 1.2 Review effective scopes: validate/build read-only, publish `contents: write` + `actions: read` | ✅ Complete | YAML parse confirms: validate explicit `contents: read`; build inherits global `contents: read` (declares none); publish explicit `contents: write` + `actions: read`. |
| 2.1 Parse workflow YAML, confirm syntactically valid | ✅ Complete | `python3 yaml.safe_load` parses clean; keys/structure valid. |
| 2.2 Verify invalid-version barrier: `validate` runs validate CLI; `build` and `publish` retain `needs` | ✅ Complete | `TestReleasePublish_NeedsValidationBarrier` passes; literal `needs: validate` and `needs: [validate, build]` present. |
| 2.3 Confirm no release dispatch/tag/asset changes, no `release-build.yml` edits | ✅ Complete | `git show a5e9df4 --stat`: only release-publish.yml (1 line) + new test file + tasks.md touched. `release-build.yml` unchanged. |
| 3.1 Run focused YAML/workflow + release-tag validator checks | ✅ Complete | `go test -count=1 ./internal/ci/... ./internal/version/...` → PASS (6 ci tests + version suite). |
| 3.2 Run `go test ./...` then inspect final diff for only the approved one-line deletion + task artifact | ✅ Complete | `go test ./...` → 10 packages PASS. Diff is the one-line deletion + added covering test + tasks artifact. No unrelated changes. |

## Build & Tests Execution

**Build**: ✅ Passed
```text
$ go build ./...
EXIT=0
```

**Tests**: ✅ 10 packages passed / 0 failed / 0 skipped
```text
$ go test ./...
ok  github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap        (cached)
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml (cached)
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/ci           (cached)
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/config       (cached)
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles     (cached)
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/environment  (cached)
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/execution    (cached)
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/planning     (cached)
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/state        (cached)
ok  github.com/dnieblesdev/dniebles-bootstrap/internal/version      (cached)
EXIT=0
```

**Vet**: ✅ Passed
```text
$ go vet ./...
EXIT=0
```

**Coverage**: ➖ Not applicable to changed files
The only changed production artifact is a declarative GitHub Actions YAML file (no Go statements). The added `internal/ci/release_publish_test.go` is a test-only package (`internal/ci` has no non-test `.go` files → `coverage: [no statements]`). No threshold configured; coverage analysis skipped per strict-TDD module ("no coverage tool detected" for changed statements).

### Invalid-version barrier — runtime evidence (dispatch invalid version → fail at barrier, no release)

The `validate` job runs `go run ./internal/version/cmd/validate --release --version "${INPUT_VERSION}"` under `set -euo pipefail`. Invalid input makes the CLI exit 1 → step fails → `build` (`needs: validate`) and `publish` (`needs: [validate, build]`) are skipped → no release.

#### Local CLI barrier evidence

```text
$ go run ./internal/version/cmd/validate --release --version "1.2.3"
release tag "1.2.3" is not a valid v-prefixed SemVer tag    exit=1
$ go run ./internal/version/cmd/validate --release --version "v1"
release tag "v1" is not a valid v-prefixed SemVer tag        exit=1
$ go run ./internal/version/cmd/validate --release --version "v1.2"
release tag "v1.2" is not a valid v-prefixed SemVer tag      exit=1
$ go run ./internal/version/cmd/validate --release --version "not-a-version"
release tag "not-a-version" is not a valid v-prefixed SemVer tag  exit=1
$ go run ./internal/version/cmd/validate --release --version "v1.2.3"
prerelease=false                                              exit=0
$ go run ./internal/version/cmd/validate --release --version "v1.2.3-rc.1"
prerelease=true                                               exit=0
```

#### Live remote dispatch evidence

A controlled `workflow_dispatch` was executed on `main` with the invalid, unprefixed version `1.2.3`.

| Field | Value |
|-------|-------|
| Workflow run | [release-publish #4](https://github.com/dnieblesdev/dniebles-bootstrap/actions/runs/29243431144) |
| Inputs | `version: 1.2.3` |
| Overall conclusion | `failure` |
| `validate` job | Failed at step "Validate release tag" (exit 1) |
| `build` job | Skipped (`needs: validate`) |
| `publish` job | Skipped (`needs: [validate, build]`) |
| Release/tag `1.2.3` | Not created (`gh release view 1.2.3` → `release not found`) |

Excerpt from the failed `validate` step:

```text
validate	Validate release tag	env:
validate	Validate release tag	  INPUT_VERSION: 1.2.3
...
release tag "1.2.3" is not a valid v-prefixed SemVer tag
exit status 1
##[error]Process completed with exit code 1.
```

Job summary from `gh run view 29243431144`:

```text
JOBS
X validate in 9s (ID 86794610054)
  ✓ Set up job
  ✓ Run actions/checkout@v4
  ✓ Run actions/setup-go@v5
  X Validate release tag
  - Post Run actions/setup-go@v5
  ✓ Post Run actions/checkout@v4
  ✓ Complete job
- publish in 0s (ID 86794650149)
- build in 0s (ID 86794650313)
```

The runtime evidence confirms the invalid-version barrier fails at `validate` and blocks both downstream jobs, producing no GitHub release.

## Spec Compliance Matrix

Delta spec: `specs/github-release-publishing/spec.md`

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| Validate the release tag (MODIFIED) | Valid stable version | `internal/ci/release_publish_test.go > TestReleasePublish_ValidateCommandAcceptsValidVersions` (v1.2.3→false, v1.2.3-rc.1→true) | ✅ COMPLIANT |
| Validate the release tag (MODIFIED) | Invalid or unprefixed version | `TestReleasePublish_ValidateCommandRejectsInvalidVersions` (1.2.3, v1, v1.2, not-a-version) + `TestReleasePublish_NeedsValidationBarrier` | ✅ COMPLIANT |
| Restrict publication authority (MODIFIED) | Permissions are inspected | `TestReleasePublish_GlobalPermissions`, `TestReleasePublish_ValidateJobPermissions`, `TestReleasePublish_PublishJobPermissions` + YAML parse | ✅ COMPLIANT |
| Restrict publication authority (MODIFIED) | Permission removal preserves behavior | `TestReleasePublish_PublishJobPreservesReleaseBehavior` (guard + create steps intact) + publish perms unchanged | ✅ COMPLIANT |
| Preserve non-permission behavior (ADDED) | Workflow behavior remains unchanged | `TestReleasePublish_NeedsValidationBarrier` + `TestReleasePublish_PublishJobPreservesReleaseBehavior` + CLI barrier (exit 0/1) | ✅ COMPLIANT |

**Compliance summary**: 5/5 scenarios compliant.

Notes on coverage depth: "Permission removal preserves behavior" and "Workflow behavior remains unchanged" are covered by static structure + CLI barrier (matching the design decision: "Prove behavior by static and workflow validation … ValidateReleaseTag and the `needs` graph already enforce the barrier; this slice changes neither"). A live GitHub runtime dispatch is out of scope for this slice; the barrier proof above is the contracted evidence.

## Correctness (Static Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| Global mapping only `contents: read` | ✅ Implemented | release-publish.yml:11-12; `actions: write` removed (git show a5e9df4). |
| `publish` alone has `contents: write` + `actions: read` | ✅ Implemented | release-publish.yml:47-49. |
| `validate` read-only | ✅ Implemented | release-publish.yml:17-18. |
| `build` (same-repo reusable workflow) read-only | ✅ Implemented | `build` declares no permissions → inherits global `contents: read`; `uses: ./.github/workflows/release-build.yml`. |
| Invalid-version barrier blocks build and publish | ✅ Implemented | `validate` runs CLI under `set -euo pipefail`; `build: needs: validate`; `publish: needs: [validate, build]`. |
| No changes to triggers/inputs/release-build/assets | ✅ Implemented | Only one line removed; `release-build.yml` untouched. |

## Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| Remove only global `actions: write` | ✅ Yes | Single-line deletion; no per-job declarations added. |
| Prove behavior by static + workflow validation (not Go code changes) | ✅ Yes | No Go product code changed; new test inspects workflow statically + runs validate CLI. |
| Preserve validate → build → publish flow | ✅ Yes | needs graph unchanged; CLI barrier verified. |

## TDD Compliance

| Check | Result | Details |
|-------|--------|---------|
| TDD Evidence reported | ⚠️ Missing | No `apply-progress` artifact with TDD Cycle Evidence table in the change directory. |
| All tasks have tests | ✅ / ➖ | 7 tasks are config/inspection tasks; covering test file `release_publish_test.go` exists and runs green. No product-code task to drive a RED/GREEN cycle. |
| RED confirmed (tests exist) | ➖ N/A | Red is degenerate for declarative config: the test asserts the post-change static state of a YAML file — no meaningful failing precondition independent of the config existing. |
| GREEN confirmed (tests pass) | ✅ | 6/6 ci tests pass on execution (`go test -count=1 ./internal/ci/...`). |
| Triangulation adequate | ✅ | Invalid-version test triangulates 4 cases (1.2.3, v1, v1.2, not-a-version); valid-version test triangulates 2 (stable + prerelease). |
| Safety Net for modified files | ✅ | `release-publish.yml` modification guarded by full `go test ./...` (10 packages green) as safety net. |

**TDD Compliance**: 4/6 checks satisfied; 2 N/A for this slice type.

**Why this is WARNING, not CRITICAL**: the strict-TDD module's CRITICAL-for-missing-apply-progress rule targets product-code changes where a RED→GREEN cycle is meaningful. This slice changes **zero product Go code** (one declarative YAML line + one test file that inspects that YAML). The project's own equivalent workflow-only change `ci-build-validation` (archived 2026-07-13, AFTER `strict_tdd: true` was added to config on 2026-07-09) was verified in **Standard** mode — establishing the project convention that config-only slices are not subject to the literal RED/GREEN TDD cycle. Applying a blocking CRITICAL here would be dogmatic and inconsistent with that precedent. Recommended process improvement (SUGGESTION): even for config-only slices, add an `apply-progress` artifact noting "TDD N/A — declarative config; safety net = full Go suite".

## Test Layer Distribution

| Layer | Tests | Files | Tools |
|-------|-------|-------|-------|
| Unit | 6 (top-level) / 6 subtests | `internal/ci/release_publish_test.go` | Go `testing`, `os/exec` |
| Integration | 0 | — | not installed (out of scope) |
| E2E | 0 | — | not installed (no GitHub remote dispatch) |
| **Total** | **6** | **1** | |

All tests are unit-level: they inspect the workflow YAML statically and exercise the `validate` CLI as a subprocess contract test. No integration/E2E tooling is used; consistent with scope (no remote dispatch).

## Changed File Coverage

| File | Line % | Branch % | Uncovered Lines | Rating |
|------|--------|----------|-----------------|--------|
| `.github/workflows/release-publish.yml` | ➖ | ➖ | declarative YAML, no Go statements | ➖ N/A |
| `internal/ci/release_publish_test.go` | ➖ | ➖ | test-only package, no production statements (`coverage: [no statements]`) | ➖ N/A |

**Average changed file coverage**: ➖ Not applicable — no product Go code changed in this slice. Per strict-TDD module, this is informational and non-blocking.

## Assertion Quality

| File | Line | Assertion | Issue | Severity |
|------|------|-----------|-------|----------|
| — | — | — | — | — |

**Assertion quality**: ✅ All assertions verify real behavior.
- Structural assertions (e.g. `!strings.Contains(block, "actions: write")`, `strings.Contains(job, "Create GitHub Release")`) verify the exact declarative property under spec — not tautologies.
- `TestReleasePublish_ValidateCommandRejectsInvalidVersions` asserts `err == nil` is false for 4 invalid inputs — real CLI exit behavior.
- `...AcceptsValidVersions` asserts the exact `prerelease=...` output string — real computed behavior.
- No ghost loops (fixed iteration slices), no type-only assertions, no mocks (0 mocks), no smoke-only tests.

## Quality Metrics

**Linter (go vet)**: ✅ No errors (`go vet ./...` → EXIT=0)
**Type Checker (go build)**: ✅ No errors (`go build ./...` → EXIT=0)
**gofmt**: ➖ Not run on changed files (no `.go` production file changed; test file already formatted).

## Issues Found

**CRITICAL**: None
**WARNING**: 1
- W1 — No `apply-progress` artifact / TDD Cycle Evidence table for this slice. Strict-TDD config is `true`; the apply phase did not record TDD evidence. Classified WARNING (not CRITICAL) because the slice changes zero product code and the project precedent (`ci-build-validation`, Standard mode) treats config-only slices as exempt from the literal RED/GREEN cycle. Remediation: add an `apply-progress` artifact for future config slices noting "TDD N/A — safety net = full Go suite".

**SUGGESTION**: 1
- S1 — Remote `workflow_dispatch` runtime validation completed. A controlled invalid-version dispatch (`1.2.3`) on `main` produced run [29243431144](https://github.com/dnieblesdev/dniebles-bootstrap/actions/runs/29243431144): `validate` failed, `build`/`publish` skipped, and `gh release view 1.2.3` returned `release not found`.

## Verdict

**PASS WITH WARNINGS**

All 5 spec scenarios are COMPLIANT with passing runtime/covering tests; the global `actions: write` grant is removed and least-privilege scopes are intact; the invalid-version barrier is proven to fail at `validate` and block both `build` and `publish` (no release) via local CLI runtime failure, the static `needs` graph, and a live `workflow_dispatch` on `main` with version `1.2.3` (run [29243431144](https://github.com/dnieblesdev/dniebles-bootstrap/actions/runs/29243431144)). The single WARNING is a process-artifact gap (no `apply-progress` TDD evidence), non-blocking for this zero-product-code, config-only slice per established project convention.