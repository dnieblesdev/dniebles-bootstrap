# Proposal: Release Binary Builds

## Intent

Make distributable `dbootstrap` binaries available from a manually initiated GitHub Actions run, so maintainers can validate packaged builds before release publishing exists.

## Scope

### In Scope
- Add a native `workflow_dispatch` build workflow.
- Build static `dbootstrap` archives for `darwin/amd64`, `linux/amd64`, and `windows/amd64` (tar.gz for Unix; zip for Windows).
- Bundle `catalog/bootstrap.toml`, generate SHA-256 checksums, and upload the build output as a workflow artifact.
- Add an ldflags-injectable version package and `--version` CLI output.

### Out of Scope
- GitHub Releases, tags, publishing, signing, changelogs, or release notes.
- Package-manager installation channels and installer documentation.
- Runtime execution tests on Windows or ARM hardware.

## Capabilities

### New Capabilities
- `release-binary-builds`: Manually produce versioned, checksummed, self-contained binary archives as GitHub Actions artifacts.

### Modified Capabilities
None. The existing `github-actions-build-validation` capability remains validation-only and does not publish artifacts.

## Approach

Add `.github/workflows/release-build.yml` using a Go build matrix with `CGO_ENABLED=0`, `GOOS`, `GOARCH`, and ldflags. Resolve an optional dispatch version or `git describe`, package each binary with the catalog, checksum archives, then upload one cataloged artifact bundle. Keep all archive logic native to GitHub Actions and shell tooling; do not add GoReleaser.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `.github/workflows/release-build.yml` | New | Manual cross-platform artifact build workflow. |
| `internal/version/version.go` | New | ldflags-overridable `Version` default. |
| `cmd/dbootstrap/main.go` | Modified | `--version` handling. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Catalog and binary mismatch | Low | Archive the catalog beside every binary. |
| Version lookup has no tag | Medium | Fetch history; fall back to a commit-derived value. |
| Cross-compiled Windows binary is unexecuted | Medium | Validate compilation and defer platform runtime tests. |

## Rollback Plan

Revert the workflow and version CLI changes; previously uploaded workflow artifacts remain non-published and expire under GitHub retention.

## Dependencies

- GitHub Actions, `actions/checkout`, `actions/setup-go`, and artifact upload support.

## Success Criteria

- [ ] A manual run uploads checksummed archives for all three target/format pairs.
- [ ] Each archive contains its executable and `catalog/bootstrap.toml`.
- [ ] `dbootstrap --version` reports the injected version; local builds retain `dev`.
- [ ] No workflow step creates a release, tag, publication, or install channel.
