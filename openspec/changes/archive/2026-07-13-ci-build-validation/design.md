# Design: CI Build Validation

## Technical Approach

Add one GitHub Actions workflow at `.github/workflows/build.yml`. It runs one Ubuntu job for pushes to `main` and pull requests targeting `main`, checks out the repository, configures Go from `go.mod`, then executes the three specified Go commands as ordered, fail-fast steps. This implements both requirements in the `github-actions-build-validation` delta spec without adding release behavior.

## Architecture Decisions

| Decision | Options / tradeoff | Choice and rationale |
|---|---|---|
| One sequential job | Separate jobs improve isolated reporting but do not guarantee the required command order and add workflow complexity. | Use one `build` job with ordered steps so a non-zero command stops later validation and the result is unambiguous. |
| Derive Go version from module metadata | Hard-coding `1.26` is explicit but drifts when `go.mod` changes. | Configure `actions/setup-go` with `go-version-file: go.mod`, making the module declaration the single version source. |
| Validation only | Uploading binaries could aid inspection but expands into packaging, retention, and release policy. | Do not create artifact directories or use artifact, release, signing, or publishing actions; `go build ./...` remains a compile check. |

## Data Flow

    push main / PR → GitHub Actions event
                         ↓
                  ubuntu-latest build job
                         ↓
        checkout → setup-go (go.mod) → test → vet → build
                                                   ↓
                                      pass/fail GitHub check

Each command runs only after its predecessor succeeds. GitHub Actions marks the job failed on any non-zero exit, so no subsequent check executes after a failure.

## File Changes

| File | Action | Description |
|---|---|---|
| `.github/workflows/build.yml` | Create | Defines the main-branch GitHub Actions Go validation workflow. |

## Interfaces / Contracts

```yaml
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go test ./...
      - run: go vet ./...
      - run: go build ./...
```

The workflow's public contract is one GitHub check that runs for the stated events and reports success only if all three commands complete successfully. It does not expose artifacts.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Static review | Trigger filters, action configuration, and command order | Inspect workflow YAML against the delta spec. |
| CI integration | Go commands run with the declared Go version | Open or update a PR targeting `main` and confirm the GitHub job output. |
| Failure behavior | A non-zero validation command fails the check | Rely on GitHub Actions step failure semantics; no application test changes are required. |

## Migration / Rollout

No migration required. Merge the workflow; GitHub Actions evaluates it on the next push to `main` or pull request targeting `main`. Rollback is deleting or reverting this single workflow file.

## Open Questions

None.
