# Exploration: CI Build Validation

## Current State

The repository has **zero CI/CD infrastructure** — this is a greenfield domain:

- **No `.github/` directory exists** — no GitHub Actions workflows, no Dependabot, no CODEOWNERS.
- **No build tooling**: No `Makefile`, `Dockerfile`, or shell scripts.
- **Single branch**: Only `main`, linear history.
- **All tests pass locally** (`go test ./...` — 8 packages, 82-97% coverage), but nothing runs in CI.
- **`go build ./...` succeeds** locally but there is no automated build gate.
- **`go vet ./...` passes** locally but there is no automated static analysis gate.

### Project Characteristics

| Attribute | Value |
|---|---|
| Language | Go 1.26 |
| Module | `github.com/dnieblesdev/dniebles-bootstrap` |
| Dependencies | 1 (go-toml/v2) |
| Binary | `cmd/dbootstrap` (single entrypoint) |
| Remote | `https://github.com/dnieblesdev/dniebles-bootstrap.git` |
| SDD | Active — 30+ archived changes, 13 domain specs, strict TDD |

### What the README Promises but Doesn't Yet Deliver

The `README.md` first-run bootstrap entrypoint section states:

> "Download and install a compatible released `dbootstrap` binary when available."

No binary release mechanism exists. The operational workflow is entirely `go run ./cmd/dbootstrap ...` — no pre-built binary distribution, no versioned releases, no CI verification of the build. **This exploration explicitly excludes release delivery: that is a separate, deferred change.**

### Existing Active Changes (potential interaction)

No active change conflicts with CI work. All recent changes are archived or in exploration-only state.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `.github/workflows/build.yml` | **New** | Build validation workflow: test, vet, build on push/PR to main |
| `.gitignore` | **Modified** | None required for this slice (no artifacts generated) |

**Explicitly out of scope for this change:**
- `.goreleaser.yml` — release configuration (deferred)
- `.github/workflows/release.yml` — release publishing (deferred)
- `dist/` directory — artifact output (deferred)
- `README.md` — release mechanism references (deferred)
- Any Makefile, Dockerfile, shell script, or non-GitHub-Actions CI system

## Approaches

### 1. **Single GitHub Actions workflow** — `go test ./...`, `go vet ./...`, `go build ./...` on push/PR to `main`

One `.github/workflows/build.yml` triggered on `push` and `pull_request` (opened, synchronized, reopened) targeting the `main` branch. Uses `actions/setup-go` with `go-version-file: go.mod` to auto-detect the Go version. Runs three sequential steps: test, vet, build. Fails fast on any step.

- **Pros**:
  - Smallest possible slice (~35-45 lines of YAML)
  - No new tooling dependencies (uses built-in Go toolchain)
  - `go-version-file: go.mod` eliminates version drift
  - Immediate safety net — every push and PR gets verified automatically
  - Completely reversible — delete the file to roll back
  - Fails-fast semantics (test fails → skip vet and build)
- **Cons**:
  - No linting (`golangci-lint` not included — deferred)
  - No multi-platform testing (ubuntu-latest only — macOS-specific Homebrew behavior won't be CI-verified; acceptable: the binary is primarily a Linux dev-environment tool)
  - No caching yet (can add `actions/cache` for Go modules in a follow-up)
- **Effort**: Low

### 2. **Parallel job matrix** — separate jobs for test, vet, and build

A multi-job workflow where `test`, `vet`, and `build` run as independent jobs. Test failures don't block vet/build results from showing. Uses `actions/setup-go` with `go-version-file: go.mod`.

- **Pros**:
  - Parallel execution is faster
  - Independent failure indicators — can see that all three gates pass/fail separately
  - GitHub UI shows each job status distinctly
- **Cons**:
  - More YAML (~60-70 lines)
  - Slightly more complex to read
  - No real speed benefit for this small codebase
- **Effort**: Low-Medium

## Recommendation

**Option 1 — Single workflow with sequential steps** (`go test ./...`, `go vet ./...`, `go build ./...`).

### Why

1. **Minimal safe scope**: ~40 lines of YAML, well within the 400-line review budget. Everything else is layered on top in separate, focused changes.

2. **Fails fast, reads simple**: Sequential steps with `go test` → `go vet` → `go build` are the simplest mental model. If tests fail, vet and build are irrelevant. Parallel jobs add YAML complexity for no real speed gain on this small codebase (8 packages, single dependency).

3. **Immediate risk reduction**: The primary risk right now is that every change is tested only locally. CI eliminates the "works on my machine" failure mode.

4. **Correct boundary**: This change is scoped strictly to *build validation gates* — verifying the code compiles, passes tests, and passes vet. It does NOT touch artifacts, releases, signing, distribution, or installation channels.

5. **`go-version-file: go.mod`**: Eliminates version drift. When the repo upgrades Go, the workflow follows automatically.

### Deferred Follow-up Slices (explicitly out of scope here)

These three changes are sequenced after `ci-build-validation` ships. Their boundaries are recorded so future SDD work can reference them directly.

| # | Change Name | Boundary | What It Covers | What It Does NOT Cover |
|---|---|---|---|---|
| 1 | `release-binary-builds` | Cross-platform binary compilation | `goreleaser` configuration, multi-arch matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64), version injection via `ldflags`, `dist/` in `.gitignore` | GitHub Releases publishing, signing, provenance attestation, Homebrew tap |
| 2 | `github-release-publishing` | Publish signed release artifacts to GitHub Releases | Tag-triggered workflow, GPG signing or GitHub attestation (provenance), checksum generation, GitHub Release creation | Installation channel (curl script, Homebrew tap, package manager) |
| 3 | `release-installation-channel` | Make the released binary available to end users | Curl-based install script, Homebrew tap formula (future), version discovery, post-install verification | Binary build configuration itself, release publishing workflow |

Each slice is autonomous: `release-binary-builds` can ship and be useful even if `github-release-publishing` hasn't been built yet (you can build binaries locally with goreleaser). Similarly, `github-release-publishing` can ship without `release-installation-channel` (users can download from the Releases page manually).

## Risks

- **GitHub Actions enabled?** The repo exists on GitHub but Actions must be enabled in repo settings. If disabled by org policy, this change is blocked.
- **Single-platform testing**: GitHub Actions `ubuntu-latest` only tests Linux/amd64. macOS-specific behavior (Homebrew) won't be CI-verified. This is acceptable for now — the binary is primarily a Linux dev-environment tool.
- **No caching**: The first run will download Go modules fresh on every workflow invocation. Can add `actions/cache` in a micro-follow-up if build times become noticeable.
- **Name collision risk**: This exploration supersedes `ci-release-delivery/exploration.md` (#3499). The original exploration was factually correct but incorrectly named — the change name implied release delivery, not build validation.

## Ready for Proposal

**Yes.** The exploration confirms a greenfield CI domain with a clear minimal first step. Recommend `sdd-propose` for `ci-build-validation` with the approach:

- **Scope**: one GitHub Actions workflow (`.github/workflows/build.yml`) that runs `go test ./...`, `go vet ./...`, and `go build ./...` on `push` and `pull_request` to `main`.
- **Out of scope**: binary releases, linting (`golangci-lint`), multi-platform matrix, Dependabot, CodeQL, artifact generation, release publishing, installation channels.
- **Deferred**: `release-binary-builds`, `github-release-publishing`, `release-installation-channel` (see Deferred Follow-up Slices above).
- **Supersedes**: `ci-release-delivery` (memory #3499 and `openspec/changes/ci-release-delivery/exploration.md`) — renamed to `ci-build-validation` to correctly scope the change to validation gates only.
