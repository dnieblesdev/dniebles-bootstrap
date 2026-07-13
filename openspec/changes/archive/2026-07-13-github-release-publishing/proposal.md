# Proposal: GitHub Release Publishing

## Intent

Publish verified `dbootstrap` release assets to GitHub Releases without rebuilding, altering, or widening the existing binary-build workflow's permissions.

## Scope

### In Scope
- Reuse `release-build.yml` through `workflow_call` while retaining manual build dispatch.
- Add an isolated manual publish workflow with `contents: write` only there.
- Require a strict `v`-prefixed SemVer input as the sole release tag and build-version source.
- Attach the exact three archives and SHA-256 checksum files produced by the called build.
- Reject an existing tag or release; mark qualifying prereleases and capture dispatch evidence.

### Out of Scope
- Package-manager publishing, signing, changelog generation, or automatic release triggers.
- Changing permissive build-version validation or existing manual-build behavior.

## Capabilities

### New Capabilities
- `github-release-publishing`: Manually create immutable GitHub Releases from verified release-build artifacts.

### Modified Capabilities
- `release-binary-builds`: Permit invocation by the publish workflow while preserving manual dispatch and artifact-only behavior when run directly.

## Approach

Expose typed `workflow_call` inputs and version outputs from `release-build.yml`. A new `release-publish.yml` validates strict SemVer before side effects, calls the build, downloads its consolidated artifact bundle, verifies its checksums, guards tag/release existence, creates the tag and release, then uploads those same files. Keep `contents: read` in the build workflow and scope `contents: write` to publishing.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `.github/workflows/release-build.yml` | Modified | Add reusable-workflow interface and outputs. |
| `.github/workflows/release-publish.yml` | New | Validate, build, verify, and publish release assets. |
| `internal/version/validate.go` | Modified | Add strict `v`-SemVer validation separately. |
| `internal/version/validate_test.go` | Modified | Cover strict SemVer acceptance and rejection. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Permanent bad release | Medium | Strict validation, checksum verification, prerelease-first evidence, non-overwrite guards. |
| Artifact mismatch | Low | Consume only the called build's consolidated bundle and verify all checksums before upload. |

## Rollback Plan

Revert the publish workflow and reusable-call additions. For an erroneous test release, delete its GitHub Release and tag explicitly; existing releases remain untouched.

## Dependencies

- GitHub Actions token permitted to write repository contents.
- `softprops/action-gh-release@v2` or an equivalent release API action.

## Success Criteria

- [ ] A `v`-SemVer dispatch publishes exactly three archives and three matching checksums.
- [ ] Build artifacts pass SHA-256 verification before release upload.
- [ ] A prerelease dispatch visibly creates a prerelease; a repeated tag fails without overwrite.
