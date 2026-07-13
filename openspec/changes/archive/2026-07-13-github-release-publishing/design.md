# Design: GitHub Release Publishing

## Technical Approach

Keep builds and publication separate. Extend `release-build.yml` with a typed `workflow_call` interface that accepts an explicit version and returns the resolved version, normalized asset version, and consolidated artifact name. A new manually dispatched `release-publish.yml` validates strict release SemVer, calls that workflow, downloads its one artifact bundle, verifies the exact six assets, then creates one GitHub Release.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Version validation | Add a strict release-tag validator beside `Validate` | Shell regex; tightening `Validate` | Direct builds intentionally accept git descriptions and empty input. A separately unit-tested pure validator preserves that contract while requiring `vMAJOR.MINOR.PATCH`, SemVer prerelease/build grammar, and no leading-zero numeric identifiers. |
| Build reuse | Local reusable workflow via `workflow_call` | Duplicate packaging in publish workflow | The caller gets the same tested packaging and checksum logic; there is no second implementation that can drift. Direct dispatch remains artifact-only. |
| Asset provenance | Consume the called workflow's named consolidated artifact and its outputs | Rebuild; glob all artifacts | The publish job can require exactly the three expected archives and their checksums for the called build's normalized version before GitHub state changes. |
| Release API | `gh release create` after explicit guards | Release action; creating first then validating | GitHub-hosted runners provide `gh`; it supports a target commit, prerelease flag, and explicit asset paths. Validation and checksum failure happen before release creation. |

## Data Flow

    workflow_dispatch(version)
             │
             ▼
    validate job ──strict tag──► reusable release-build (contents: read)
             │                         │
             │                    artifact + outputs
             ▼                         ▼
    publish job ◄── download exact bundle ── sha256sum --check
             │
    tag/release existence guards ──► gh release create + six explicit assets

`release-build.yml` retains `contents: read`. `release-publish.yml` defaults to read access; only its publish job declares `contents: write`.

## File Changes

| File | Action | Description |
|---|---|---|
| `.github/workflows/release-build.yml` | Modify | Add `workflow_call` input/output contract and expose consolidated artifact identity without publishing. |
| `.github/workflows/release-publish.yml` | Create | Manual validate, call, verify, non-overwrite guard, and release workflow. |
| `internal/version/validate.go` | Modify | Add strict release-tag validation and prerelease classification without changing `Validate`. |
| `internal/version/validate_test.go` | Modify | Table-driven acceptance/rejection coverage for strict SemVer. |
| `internal/version/cmd/validate/main.go` | Modify | Add an explicit release-validation mode for workflow use. |

## Interfaces / Contracts

```yaml
# release-build.yml
on:
  workflow_call:
    inputs:
      version: { required: true, type: string }
    outputs:
      version: { value: ${{ jobs.version.outputs.version }} }
      safe_version: { value: ${{ jobs.version.outputs.safe_version }} }
      artifact_name: { value: ${{ jobs.upload.outputs.artifact_name }} }
```

```go
// ValidateReleaseTag accepts only v-prefixed SemVer and reports prerelease state.
func ValidateReleaseTag(v string) (isPrerelease bool, err error)
```

The publish input is required and becomes both called-build `version` and release tag. Its job derives the expected names from `safe_version`: Linux amd64/arm64 `.tar.gz`, Windows amd64 `.zip`, plus each `.sha256`. It rejects any missing or extra downloaded file, runs `sha256sum --check --strict`, checks both remote tag and release absence, then passes only those six explicit paths to `gh release create`. A prerelease is identified from the validated tag and supplied with `--prerelease`.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Strict SemVer grammar and prerelease state | Go table tests: stable/prerelease/build accepted; unprefixed, partial, leading-zero, malformed identifiers rejected. |
| Integration | Existing build remains valid | `go test ./...`, `go vet ./...`, and manual reusable/direct workflow review. |
| E2E | Immutable prerelease publication | Dispatch a unique `vX.Y.Z-rc.N`, inspect six release assets and prerelease evidence, then re-dispatch the same tag and require failure without mutation. |

## Migration / Rollout

No migration required. First remote evidence uses a unique prerelease. Roll back by reverting both workflow changes and the strict-validator addition; delete only an explicitly identified test release/tag if necessary.

## Open Questions

- [ ] None.
