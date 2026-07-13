# Proposal: CI Build Validation

## Intent

Establish a minimal GitHub Actions gate so changes to `main` are automatically tested, vetted, and compiled rather than relying solely on local validation.

## Scope

### In Scope
- Add `.github/workflows/build.yml`.
- Trigger on pushes to `main` and pull requests targeting `main`.
- Run `go test ./...`, `go vet ./...`, then `go build ./...` using the Go version declared in `go.mod`.

### Out of Scope
- Artifacts, binary packaging, release publishing, signing, provenance, or installation channels.
- Linting, caching, matrices, Dependabot, CodeQL, and non-GitHub-Actions CI.

## Capabilities

### New Capabilities
- `github-actions-build-validation`: Validate the Go repository on GitHub Actions for pushes and pull requests affecting `main`.

### Modified Capabilities
None. Existing specs do not define CI behavior.

## Approach

Create one Ubuntu-based workflow with a single sequential job. Configure Go from `go.mod` via `actions/setup-go`, then fail fast through test, vet, and build steps in that order.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `.github/workflows/build.yml` | New | Main-branch Go validation workflow. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Actions are disabled by repository policy | Low | Enable Actions or resolve policy before merge. |
| Ubuntu-only validation misses OS-specific behavior | Medium | Keep this slice scoped; add platform coverage separately if needed. |

## Rollback Plan

Revert or delete `.github/workflows/build.yml`; this removes the CI gate without changing application behavior or generated state.

## Dependencies

- GitHub Actions must be enabled for the repository.
- `actions/setup-go` and GitHub-hosted Ubuntu runners must be available.

## Success Criteria

- [ ] A push to `main` runs test, vet, and build successfully.
- [ ] A pull request targeting `main` runs the same three checks.
- [ ] No workflow creates, uploads, publishes, signs, or releases artifacts.
