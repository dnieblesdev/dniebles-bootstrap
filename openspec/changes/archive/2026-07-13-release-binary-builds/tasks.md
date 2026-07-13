# Tasks: Release Binary Builds

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 180–260 |
| 800-line budget risk | Low |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR with behavior, tests, workflow, and evidence |
| Delivery strategy | auto-chain |
| Chain strategy | pending (single PR; no chain needed) |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | Versioned CLI behavior and release workflow | PR 1 | Include focused tests and local ldflags verification; single-PR scope |
| 2 | Manual runtime evidence and artifact inspection | PR 1 | Verify successful dispatch outputs and no publishing side effects |

## Phase 1: Version Contract (TDD)

- [x] 1.1 In `cmd/dbootstrap/main_test.go`, add failing `--version` coverage asserting `dev\n`, empty stderr, and success; include the injected-version case using the existing test seam.
- [x] 1.2 Create `internal/version/version.go` with linker-overridable `Version = "dev"`; update `cmd/dbootstrap/main.go` to print it and exit successfully for top-level `--version`.
- [x] 1.3 Run focused Go tests, then build with `-ldflags -X github.com/dnieblesdev/dniebles-bootstrap/internal/version.Version=v1.2.3` and verify `--version` reports `v1.2.3`.

## Phase 2: Workflow and Packaging

- [x] 2.1 Create `.github/workflows/release-build.yml` with only `workflow_dispatch`, optional `version`, `contents: read`, pinned checkout/setup/upload actions, and the supported `linux/amd64`, `linux/arm64`, and `windows/amd64` matrix.
- [x] 2.2 Implement version resolution, `CGO_ENABLED=0` cross-builds, per-target staging with the executable plus `catalog/bootstrap.toml`, target archive formats/names, and adjacent SHA-256 files.
- [x] 2.3 Gate one final upload job on every matrix build and upload exactly the three archives plus three checksum files; include no release, tag, package, or external publication step.

## Phase 3: Verification and Evidence

- [x] 3.1 Review YAML triggers, permissions, target matrix, archive layout, checksum commands, and failure gating; run `go test ./...`, `go vet ./...`, and the relevant builds.
- [x] 3.2 Manually dispatch the workflow with `v1.2.3`, record the successful run URL, download the artifact bundle, and verify all archive formats, filenames, catalog contents, executables, and checksum matches.
- [x] 3.3 Inspect the manual run side effects and record evidence that no release, tag, package-manager channel, or other publication was created; retain evidence in the verification report.

## Phase 4: Cleanup

- [x] 4.1 Confirm the existing validation workflow remains unchanged and document deferred Windows/ARM runtime testing and artifact-retention rollback behavior.

## Phase 5: Review Fixes (R4-001/002/003/004)

- [x] R4-001 Add a `quality` job to `.github/workflows/release-build.yml` that runs `go test ./...`, `go vet ./...`, and `go build ./...`; gate `build` and `upload` jobs on it.
- [x] R4-002 Add safe version validation for `workflow_dispatch.inputs.version` before it is used in shell commands, filenames, or ldflags; add Go validation function with tests and invoke it from the workflow.
- [x] R4-003 Ensure `workflow_dispatch.inputs.version` never interpolates into shell source before validation; pass it to every workflow shell step via an environment variable and quote the shell reference.
- [x] R4-004 Add tested Git version normalization to safe `[A-Za-z0-9._-]`: coalesce invalid runs (including `/`) to `-`, trim edge separators, fall back to `dev` if empty; preserve original Git metadata separately and use the sanitized value only for filesystem paths and artifact names. Cover normal tags, slash branches, spaces, invalid characters, and empty input.
