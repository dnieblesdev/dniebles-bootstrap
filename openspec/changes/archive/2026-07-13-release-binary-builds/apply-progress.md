# Apply Progress: Release Binary Builds

## Task Status

### Phase 1: Version Contract (TDD)
- [x] 1.1 In `cmd/dbootstrap/main_test.go`, add failing `--version` coverage asserting `dev\n`, empty stderr, and success; include the injected-version case using the existing test seam.
- [x] 1.2 Create `internal/version/version.go` with linker-overridable `Version = "dev"`; update `cmd/dbootstrap/main.go` to print it and exit successfully for top-level `--version`.
- [x] 1.3 Run focused Go tests, then build with `-ldflags -X github.com/dnieblesdev/dniebles-bootstrap/internal/version.Version=v1.2.3` and verify `--version` reports `v1.2.3`.

### Phase 2: Workflow and Packaging
- [x] 2.1 Create `.github/workflows/release-build.yml` with only `workflow_dispatch`, optional `version`, `contents: read`, pinned checkout/setup/upload actions, and the supported `linux/amd64`, `linux/arm64`, and `windows/amd64` matrix.
- [x] 2.2 Implement version resolution, `CGO_ENABLED=0` cross-builds, per-target staging with the executable plus `catalog/bootstrap.toml`, target archive formats/names, and adjacent SHA-256 files.
- [x] 2.3 Gate one final upload job on every matrix build and upload exactly the three archives plus three checksum files; include no release, tag, package, or external publication step.

### Phase 3: Verification and Evidence
- [x] 3.1 Review YAML triggers, permissions, target matrix, archive layout, checksum commands, and failure gating; run `go test ./...`, `go vet ./...`, and the relevant builds.
- [x] 3.2 Manually dispatch the workflow with `v1.2.3`, record the successful run URL, download the artifact bundle, and verify all archive formats, filenames, catalog contents, executables, and checksum matches.
- [x] 3.3 Inspect the manual run side effects and record evidence that no release, tag, package-manager channel, or other publication was created; retain evidence in the verification report.

### Phase 4: Cleanup
- [x] 4.1 Confirm the existing validation workflow remains unchanged and document deferred Windows/ARM runtime testing and artifact-retention rollback behavior.

### Phase 5: Review Fixes (R4-001/002/003/004)
- [x] R4-001 Add a `quality` job to `.github/workflows/release-build.yml` that runs `go test ./...`, `go vet ./...`, and `go build ./...`; gate `build` and `upload` jobs on it.
- [x] R4-002 Add safe version validation for `workflow_dispatch.inputs.version` before it is used in shell commands, filenames, or ldflags; add Go validation function with tests and invoke it from the workflow.
- [x] R4-003 Ensure `workflow_dispatch.inputs.version` never interpolates into shell source before validation; pass it to every workflow shell step via an environment variable and quote the shell reference.
- [x] R4-004 Add tested Git version normalization to safe `[A-Za-z0-9._-]`: coalesce invalid runs (including `/`) to `-`, trim edge separators, fall back to `dev` if empty; preserve original Git metadata separately and use the sanitized value only for filesystem paths and artifact names. Cover normal tags, slash branches, spaces, invalid characters, and empty input.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1 | `cmd/dbootstrap/main_test.go` | Unit | ✅ baseline passed | ✅ Written | ✅ Passed | ✅ 2 cases (default dev + injected v1.2.3) | ✅ Clean |
| 1.2 | `cmd/dbootstrap/main_test.go` | Unit | N/A (new package) | ✅ Referenced | ✅ Passed | ✅ Covered by 1.1 cases | ✅ Clean |
| 1.3 | `cmd/dbootstrap/main_test.go` / manual | Integration | N/A | N/A (verification step) | ✅ Verified | ✅ ldflags build reports v1.2.3 | ➖ None needed |
| 2.1–2.3 | `.github/workflows/release-build.yml` | Workflow | N/A (new file) | N/A (structural/config) | ✅ YAML valid | ➖ Single structural output | ➖ None needed |
| 3.1 | `cmd/dbootstrap/main_test.go` / local simulation | Integration | ✅ full suite green | N/A | ✅ Verified | ✅ Linux tar.gz contents + checksums + Windows PE binary | ➖ None needed |
| 3.2 | GitHub Actions runtime | Integration | ✅ workflow on remote main | N/A | ✅ Verified | ✅ Real dispatch produced all target artifacts + consolidated bundle | ✅ Clean |
| 3.3 | GitHub API checks | Integration | ✅ no publish steps in workflow | N/A | ✅ Verified | ✅ No tags, releases, or packages created | ✅ Clean |
| R4-001 | `.github/workflows/release-build.yml` | Workflow | ✅ full suite green | N/A (structural/config) | ✅ YAML valid | ➖ Single structural output | ➖ None needed |
| R4-002 | `internal/version/validate_test.go`, `internal/version/cmd_validate_test.go` | Unit / Integration | ✅ full suite green | ✅ Written | ✅ Passed | ✅ 20 cases (valid/invalid/empty/length/charset) | ✅ Extracted `maxVersionLength` constant |
| R4-003 | `.github/workflows/release-build.yml` | Workflow | ✅ YAML valid | N/A (structural/config) | ✅ YAML valid | ✅ No `${{ inputs.version }}` inside `run:` scripts | ✅ Clean |
| R4-004 | `internal/version/normalize_test.go`, `internal/version/cmd_normalize_test.go` | Unit / Integration | ✅ full suite green | ✅ Written | ✅ Passed | ✅ 13 cases (tags, slash branches, spaces, invalid chars, empty, edge trims, unicode) | ✅ Clean |
| R4-004 workflow | `.github/workflows/release-build.yml` | Workflow | ✅ YAML valid | N/A (structural/config) | ✅ YAML valid | ✅ `version` output preserved for ldflags; `safe_version` used for paths/names | ✅ Clean |

## Test Summary

- **Total tests written**: 5 (`TestRunVersionFlag` with 2 sub-cases; `TestValidate` with 20 sub-cases; `TestValidateCmd` with 3 sub-cases; `TestNormalizeGitVersion` with 13 sub-cases; `TestNormalizeCmd` with 3 sub-cases)
- **Total tests passing**: 5
- **Layers used**: Unit (3), Integration (2)
- **Approval tests**: None — no refactoring tasks
- **Pure functions created**: 2 (`version.Validate`, `version.NormalizeGitVersion`)
- **Full suite result**: `go test ./...` passed (all packages)
- **Static analysis**: `go vet ./...` passed
- **Build verification**: `go build ./...` passed

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/version/version.go` | Created | Linker-overridable `Version = "dev"` with package documentation. |
| `cmd/dbootstrap/main.go` | Modified | Added top-level `--version` branch that prints `version.Version` and exits successfully. |
| `cmd/dbootstrap/main_test.go` | Modified | Added `TestRunVersionFlag` covering default `dev` output and injected `v1.2.3` output. |
| `internal/version/validate.go` | Created | Safe version validation function with charset and length rules. |
| `internal/version/validate_test.go` | Created | Table-driven tests for valid versions, invalid characters, injection attempts, and length limits. |
| `internal/version/cmd_validate_test.go` | Created | Integration test that runs `go run ./cmd/validate --version ...` for valid/invalid/empty inputs. |
| `internal/version/cmd/validate/main.go` | Created | Tiny CLI wrapper that invokes `version.Validate` and exits non-zero on failure. |
| `.github/workflows/release-build.yml` | Modified | Added `quality` job (test/vet/build), gated `build` and `upload` on it, and added a `go run ./internal/version/cmd/validate` step in the `version` job. |
| `.github/workflows/release-build.yml` | Modified (R4-003) | Replaced all `${{ inputs.version }}` interpolations inside `run:` scripts with `env.INPUT_VERSION`; quoted shell references to prevent injection before validation. |
| `internal/version/normalize.go` | Created | `NormalizeGitVersion` function: safe charset `[A-Za-z0-9._-]`, collapse invalid runs to `-`, trim edge separators, fall back to `dev` if empty. |
| `internal/version/normalize_test.go` | Created | Table-driven tests covering tags, slash branches, spaces, invalid characters, empty input, edge trims, and unicode. |
| `internal/version/cmd/normalize/main.go` | Created | Tiny CLI wrapper that invokes `version.NormalizeGitVersion` and prints the sanitized version. |
| `internal/version/cmd_normalize_test.go` | Created | Integration test that runs `go run ./cmd/normalize --version ...` for normal, slash-branch, and empty inputs. |
| `.github/workflows/release-build.yml` | Modified (R4-004) | Added `safe_version` job output; preserve `version` for ldflags injection; use `safe_version` for staging directories, archive names, checksum files, and artifact names. |
| `openspec/changes/release-binary-builds/verify-report.md` | Created | Hybrid verification evidence for workflow dispatch, artifact inspection, and no-publication checks. |

## Local Build Verification

Simulated the workflow packaging commands locally with `VERSION=v1.2.3`:

- `linux/amd64` archive created: `dbootstrap_v1.2.3_linux_amd64.tar.gz`
- `linux/arm64` archive created: `dbootstrap_v1.2.3_linux_arm64.tar.gz`
- `windows/amd64` executable cross-compiled: valid `PE32+ executable for MS Windows 6.01 (console), x86-64`
- Archive root contains `dbootstrap` and `catalog/bootstrap.toml`
- SHA-256 checksums generated and verified with `sha256sum -c`
- Injected binary reports `v1.2.3` for `--version`

Version validation verified locally:

- `go run ./internal/version/cmd/validate --version "v1.2.3"` exits 0
- `go run ./internal/version/cmd/validate --version "v1.2.3; rm -rf /"` exits 1 with "contains invalid characters"
- `go run ./internal/version/cmd/validate --version ""` exits 0 (empty means use default/git-derived version)

Version normalization verified locally:

- `go run ./internal/version/cmd/normalize --version "v1.2.3"` prints `v1.2.3`
- `go run ./internal/version/cmd/normalize --version "feature/new-thing"` prints `feature-new-thing`
- `go run ./internal/version/cmd/normalize --version "v1.2.3; rm -rf /"` prints `v1.2.3-rm--rf`
- `go run ./internal/version/cmd/normalize --version ""` prints `dev`
- Local packaging simulation with `VERSION=feature/test-branch` produced `dbootstrap_feature-test-branch_linux_amd64.tar.gz` while the binary reported the original `feature/test-branch` for `--version`.

Workflow YAML validated with `python3 -c "import yaml; yaml.safe_load(...)`" and passes.

## Remote Workflow Dispatch Verification

Dispatched `release-build` on `main` with version `v1.2.3`:

- **Run URL**: https://github.com/dnieblesdev/dniebles-bootstrap/actions/runs/29233218426
- **Run ID**: `29233218426`
- **Trigger**: `workflow_dispatch` with `version=v1.2.3`
- **Conclusion**: success
- **Jobs**:
  - `version` in 13s (ID 86761880126) — validated input and resolved `version=v1.2.3`, `safe_version=v1.2.3`
  - `quality` in 12s (ID 86761880152) — `go test ./...`, `go vet ./...`, `go build ./...` all passed
  - `build (windows, amd64, .exe, zip)` in 14s (ID 86761926665) — produced `dbootstrap_v1.2.3_windows_amd64.zip` + `.sha256`
  - `build (linux, amd64, tar.gz)` in 13s (ID 86761926666) — produced `dbootstrap_v1.2.3_linux_amd64.tar.gz` + `.sha256`
  - `build (linux, arm64, tar.gz)` in 16s (ID 86761926683) — produced `dbootstrap_v1.2.3_linux_arm64.tar.gz` + `.sha256`
  - `upload` in 8s (ID 86761979446) — downloaded matrix artifacts and uploaded consolidated `dbootstrap-artifacts-v1.2.3`

Downloaded artifacts:

| Artifact | Contents |
|----------|----------|
| `dbootstrap-linux-amd64` | `dbootstrap_v1.2.3_linux_amd64.tar.gz` + `.sha256` |
| `dbootstrap-linux-arm64` | `dbootstrap_v1.2.3_linux_arm64.tar.gz` + `.sha256` |
| `dbootstrap-windows-amd64` | `dbootstrap_v1.2.3_windows_amd64.zip` + `.sha256` |
| `dbootstrap-artifacts-v1.2.3` | All six files above consolidated |

Artifact inspection results:

- All three `.sha256` files verified with `sha256sum -c`.
- `linux/amd64` archive extracts to `./dbootstrap` + `./catalog/bootstrap.toml`; binary is `ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked`; `./dbootstrap --version` prints `v1.2.3`.
- `linux/arm64` archive extracts to `./dbootstrap` + `./catalog/bootstrap.toml`; binary is `ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), statically linked`.
- `windows/amd64` archive extracts to `dbootstrap.exe` + `catalog/bootstrap.toml`; binary is `PE32+ executable for MS Windows 6.01 (console), x86-64`.
- `catalog/bootstrap.toml` matches the source catalog file.

SHA-256 values of downloaded archives:

```text
421cc9ac9b6cfae6ddd06e7bc411b40d36ee652c044e88083ad91eef333c8f6a  dbootstrap_v1.2.3_linux_amd64.tar.gz
39cdb63b072860639783ef939f220e9f0164c7f252528389af360029dedf0c6c  dbootstrap_v1.2.3_linux_arm64.tar.gz
ebc1711ff841443c6dee15e29f9f760863669ac0e2c1049ec38d4e8b3b27c8ad  dbootstrap_v1.2.3_windows_amd64.zip
```

## No-Publication Verification

Checked GitHub side effects after the successful run:

| Check | Command | Result |
|-------|---------|--------|
| Tags | `gh api repos/dnieblesdev/dniebles-bootstrap/git/refs/tags` | 404 — no tags exist |
| Releases | `gh release list --repo dnieblesdev/dniebles-bootstrap` | empty list |
| Packages | `gh api repos/dnieblesdev/dniebles-bootstrap/packages` | 404 — no packages exist |

The workflow file contains no `actions/create-release`, `softprops/action-gh-release`, `gh release`, `docker push`, npm publish, or any other publication step. `permissions: contents: read` prevents write access to contents/releases.

## Deviations from Design

None — implementation matches design.

## Issues Found

- Windows/ARM runtime testing remains deferred (out of scope per proposal).
- Node.js 20 deprecation annotation for `actions/checkout@v4`, `actions/setup-go@v5`, `actions/upload-artifact@v4`, and `actions/download-artifact@v4` (forced to Node.js 24 by GitHub). Non-blocking; does not affect artifact correctness.

## Remaining Tasks

None — all tasks complete.

## Workload / PR Boundary

- Mode: single PR
- Current work unit: Release binary builds slice including version contract, workflow, review fixes, and remote dispatch evidence
- Boundary: starts at the existing local release-build workflow and ends after real remote dispatch verification with artifact inspection and no-publication evidence
- Estimated review budget impact: within original 180–260 line forecast plus ~80–130 line review fixes; still within single-PR budget

## Rollback / Deferred Notes

- The existing `.github/workflows/build.yml` validation workflow remains unchanged.
- Reverting this change removes the new workflow and version CLI behavior; previously uploaded workflow artifacts expire under GitHub retention.
- Windows/ARM runtime testing and artifact-retention rollback behavior are documented as deferred.
