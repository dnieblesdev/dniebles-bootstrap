# Exploration: GitHub Release Publishing

## Current State

The repository has a working `release-build.yml` workflow (`workflow_dispatch` only) that produces
versioned, checksummed binary archives and uploads them as workflow artifacts. No release publishing
exists yet. Key facts:

| Attribute | State |
|---|---|
| Git tags | **Zero** — `git tag --list` returns empty |
| GitHub Releases | **Zero** — no published releases |
| `release-build.yml` triggers | `workflow_dispatch` only — no `workflow_call` |
| `release-build.yml` permissions | `contents: read` (workflow-level) |
| `ci-release-delivery/` change directory | **Exists but empty** — never populated |
| Artifact contract | `dbootstrap-artifacts-<safe_version>` bundle (3 archives + 3 checksums) |
| Version validation | **Permissive** — `version.Validate` accepts `dev`, `abc1234`, `v0.1.2-3-gabcdef`. NOT strict SemVer. |
| SemVer enforcement | **None exists** anywhere in the codebase |
| ldflags injection contract | `internal/version.Version` overridable via `-X`; validated workflow-safe since R4-003 |

### Workflow Artifact Contract (read from actual `release-build.yml`)

The workflow produces exactly these artifacts:

```
dbootstrap-artifacts-<safe_version>/          ← consolidated upload job
├── dbootstrap_<safe_version>_linux_amd64.tar.gz
├── dbootstrap_<safe_version>_linux_amd64.tar.gz.sha256
├── dbootstrap_<safe_version>_linux_arm64.tar.gz
├── dbootstrap_<safe_version>_linux_arm64.tar.gz.sha256
├── dbootstrap_<safe_version>_windows_amd64.zip
└── dbootstrap_<safe_version>_windows_amd64.zip.sha256
```

The `safe_version` is derived from `go run ./internal/version/cmd/normalize` (commits/dashes/branches normalized to
`[A-Za-z0-9._-]`). The original `version` (used for ldflags and `--version` output) is preserved unmodified.

### Version Package Contract

```go
// internal/version/version.go
var Version = "dev"
```

- **Validate** (`version.Validate`): Accepts empty, alphanumeric `[a-zA-Z0-9][a-zA-Z0-9._+-]{0,63}` — permissive, not SemVer.
- **NormalizeGitVersion** (`version.NormalizeGitVersion`): Maps arbitrary git metadata to filesystem-safe strings.
- **CLI tools**: `./internal/version/cmd/validate` (exit 0/1), `./internal/version/cmd/normalize` (prints to stdout).
- **20 validate test cases**, **13 normalize test cases** — both green.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `.github/workflows/release-build.yml` | **Modified** (Approach A) or **Modified** (Approach B) | Add `workflow_call` trigger + outputs, or add release job + SemVer input |
| `.github/workflows/release-publish.yml` | **New** (Approach A only) | New workflow: SemVer validation, tag guard, release creation, artifact attachment |
| `internal/version/validate.go` | **New function** (both approaches) | Strict SemVer validator added (new function, not replacing existing permissive one) |
| `internal/version/validate_test.go` | **Modified** | Test cases for strict SemVer function |
| `openspec/specs/github-release-publishing/` | **New** | Delta spec for release publishing behavior |

## Approaches

### 1. Safe `workflow_call` Reuse (`release-publish.yml` calls `release-build.yml`)

Add `workflow_call` trigger to `release-build.yml` so another workflow can invoke it
programmatically. Create a separate `release-publish.yml` that dispatches with a strict
SemVer tag, calls the build workflow, downloads the consolidated artifact, and publishes
a GitHub Release.

```
release-publish.yml  (workflow_dispatch, contents: write, SemVer-only)
  │
  ├─ calls release-build.yml (workflow_call)
  │    ├─ version job (resolve, validate)
  │    ├─ quality job (go test/vet/build)
  │    ├─ build job (matrix: 3 targets)
  │    └─ upload job (consolidate artifacts)
  │
  └─ publish job
       ├─ Strict SemVer validation (new function)
       ├─ Tag existence guard (non-overwrite)
       ├─ Prerelease detection (alpha/beta/rc/pre/dev)
       ├─ Create git tag + GitHub Release
       └─ Attach artifacts as release assets
```

**Artifact passing mechanism**: `release-publish.yml` passes `version` via `with: version: ${{ inputs.version }}`.
The called workflow's `upload` job creates `dbootstrap-artifacts-<safe_version>`. The caller downloads
it using `actions/download-artifact@v4` with pattern `dbootstrap-artifacts-*` (or via `workflow_call.outputs.safe_version`).

**`release-build.yml` refactoring needed**: ~15 lines.

```yaml
# ADD to existing on: block
  workflow_call:
    inputs:
      version:
        description: 'Version to embed in the binary (optional; defaults to git describe)'
        required: false
        type: string
    outputs:
      version:
        description: 'Resolved build version'
        value: ${{ jobs.version.outputs.version }}
      safe_version:
        description: 'Filesystem-safe version for artifact naming'
        value: ${{ jobs.version.outputs.safe_version }}
```

- **Pros**:
  - Clean separation of concerns: build vs publish are independent workflows
  - Build workflow retains `contents: read` — no permission escalation risk
  - Independent testability: dispatch build without risk of publishing
  - Backward compatible: existing `workflow_dispatch` on `release-build.yml` unchanged
  - Audit trail: two distinct workflow runs per release (build → publish) is clearer than one
  - Matches existing slice architecture: `release-binary-builds` (build) → `github-release-publishing` (publish)
  - Publish workflow owns SemVer enforcement exclusively — no conditional logic in build workflow
  - Fine-grained permissions: only the publish workflow needs `contents: write`
- **Cons**:
  - Refactoring required on `release-build.yml` (but ~15 lines, well-understood pattern)
  - Two workflow files to maintain instead of one
  - `workflow_call` introduces a dependency between workflows (caller must know artifact naming)
  - Slightly more total YAML (~80 publish + ~15 refactor = ~95 lines vs ~60 for inline approach)
- **Effort**: Medium

### 2. Extend `release-build.yml` Inline (conditional publish job)

Add a `publish` boolean input and a `prerelease` boolean input to the existing
`release-build.yml`. A new `release` job (conditional on `${{ inputs.publish }}`)
validates SemVer, prevents overwrite, and creates the GitHub Release.

```
release-build.yml  (workflow_dispatch, mixed permissions)
  ├─ version job
  ├─ quality job
  ├─ build job (matrix)
  ├─ upload job (consolidate)
  └─ release job (if: inputs.publish, permissions: contents: write)
       ├─ Strict SemVer validation
       ├─ Tag existence guard
       ├─ Create git tag + GitHub Release
       └─ Attach artifacts
```

- **Pros**:
  - Single workflow file — all release logic in one place
  - No `workflow_call` dependency — simpler coordination
  - Discoverability: build and publish in the same file
  - Fewer total lines: ~60 lines added vs ~95 for two-file approach
- **Cons**:
  - Monolithic — violates Single Responsibility Principle
  - Permission mixing: workflow needs `contents: read` (build) AND `contents: write` (release), even with job-level override there's design tension
  - Conditional complexity: `if: inputs.publish` on multiple steps adds cognitive overhead
  - Testing risk: dispatching with `publish: false` exercises build only; `publish: true` creates permanent side effects — no dry-run for publish path
  - Harder review: one large YAML file mixing build shell commands with release API interactions
  - Input sprawl: `version`, `publish`, and `prerelease` as three `workflow_dispatch` inputs — non-publish dispatches must remember to leave `publish: false`
  - The existing spec requirement "The workflow MUST NOT create GitHub Releases" (from `release-binary-builds`) would need a spec delta to carve out the conditional release path — awkward spec evolution
- **Effort**: Low-Medium

## Design Criteria Analysis

### 1. Artifact Contract / Workflow Reuse

| Criterion | Approach A (`workflow_call`) | Approach B (inline) |
|---|---|---|
| Artifact contract stays intact | ✅ Build workflow unchanged for non-publish dispatches | ⚠️ Build workflow gains conditional inputs |
| Artifact passing to release | ✅ Caller downloads via artifact name/pattern | ✅ Same workflow run — direct access via download-artifact |
| Reuse without duplication | ✅ Build logic runs once, caller consumes output | ✅ Single workflow, no duplication |
| Backward compatibility | ✅ Existing `workflow_dispatch` unaffected | ⚠️ New inputs visible on existing dispatch UI |

### 2. Strict Tag SemVer

Both approaches need the SAME new validation function. The current `version.Validate` is intentionally
permissive (it validates build inputs, not release tags). A release publish MUST enforce:

```
^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$
```

Required behavior:
- `v1.2.3` → ✅ valid
- `v1.2.3-alpha.1` → ✅ valid (prerelease)
- `v1.2.3+build.123` → ✅ valid (build metadata)
- `v1.2.3-alpha.1+build.123` → ✅ valid
- `1.2.3` → ❌ missing `v` prefix
- `dev` → ❌ not SemVer
- `abc1234` → ❌ not SemVer

Implementation: Add `version.ValidateStrictSemVer(version string) error` as a separate function
alongside the permissive `Validate`. The publish workflow calls the strict validator on its input.

### 3. Permissions

| Approach | Build permissions | Publish permissions | Risk |
|---|---|---|---|
| A (`workflow_call`) | `contents: read` (unchanged) | `contents: write` (release-publish.yml only) | **Low** — write is isolated to the publish workflow |
| B (inline) | `contents: read` at workflow level | `contents: write` at `release` job level via `permissions:` override | **Medium** — write permission lives in the same file as the build jobs; a future YAML mistake could broaden it |

GitHub Actions job-level `permissions:` override IS supported. Approach B can scope write to
the release job. However, Approach A isolates it at the workflow level, which is a stronger
security boundary — accidental modification of the build workflow cannot introduce write access.

### 4. Non-Overwrite Behavior

Both approaches need identical guard logic:

```bash
# Before creating tag/release
if git rev-parse "refs/tags/${TAG}" >/dev/null 2>&1; then
  echo "ERROR: Tag ${TAG} already exists — releases are immutable"
  exit 1
fi
```

Additional checks:
- Query GitHub Releases API for existing release with the same tag → fail if found
- The `softprops/action-gh-release@v2` action fails by default if the tag exists (it checks before creating)

This is independent of the approach chosen. Both implement it identically.

### 5. Testable Prerelease Evidence

Prerelease detection logic: version string contains `-(alpha|beta|rc|pre|dev)` → `prerelease: true`.

| Version | Prerelease? | Evidence |
|---|---|---|
| `v1.2.3` | No | Release page shows "Latest" badge, no "Pre-release" label |
| `v1.2.3-alpha.1` | Yes | Release page shows "Pre-release" badge |
| `v1.2.3-rc.2` | Yes | Release page shows "Pre-release" badge |

Testing procedure (same for both approaches):
1. Dispatch with `version=v1.2.3-alpha.1` → verify GitHub Release marked "Pre-release"
2. Dispatch with `version=v1.2.3` → verify GitHub Release marked "Latest" (not prerelease)
3. Re-dispatch with same `version=v1.2.3` → verify workflow **fails** with "Tag already exists" error
4. Screenshot each release page as evidence

Approach A allows testing the build portion independently before testing publish. Approach B
requires either creating a real release or skipping the release job (which doesn't test the full path).

## Recommendation

**Approach 1 — Safe `workflow_call` Reuse**

### Why

1. **Security-first**: `release-build.yml` retains `contents: read`. `release-publish.yml` is the ONLY
   workflow with `contents: write`. This is the strongest isolation boundary. Approach B's job-level
   permission override works but is a weaker separation.

2. **Spec compliance**: The existing `release-binary-builds` spec states "The workflow MUST NOT create
   GitHub Releases." Adding a publish job to `release-build.yml` would require a spec delta to carve out
   an exception — awkward spec evolution. A separate workflow naturally extends the spec without
   modifying existing requirements.

3. **Independent testability**: Build can be dispatched and verified without any risk of publishing.
   The publish workflow can be tested with a prerelease version first as a smoke test. Approach B
   always has the `publish` boolean sitting there — one accidental click creates a release.

4. **Matches project architecture**: The repo already separates `build.yml` (validation) from
   `release-build.yml` (artifact production). Adding `release-publish.yml` (publication) extends
   this pattern naturally. Three workflows, three responsibilities.

5. **Refactoring is proven and minimal**: Adding `workflow_call` to an existing workflow is a
   standard GitHub Actions pattern. The ~15 lines of YAML are well-understood, backward compatible,
   and the artifact-passing mechanism (download by pattern) is documented and tested.

6. **Cleaner audit trail**: Two workflow runs tell a clearer story: "build produced artifacts" →
   "publish created release v1.2.3". Single-run with `publish: true` conflates build success with
   release creation.

### Version Validation Strategy

Add `version.ValidateStrictSemVer(v string) error` to `internal/version/validate.go` as a SEPARATE
function. Do NOT modify the existing permissive `Validate` — it serves a different purpose (workflow
input sanitization, not SemVer enforcement). The publish workflow calls `ValidateStrictSemVer` on
its `version` input before any tag or release creation.

### Workflow Call Output Contract

```yaml
# release-build.yml additions:
workflow_call:
  inputs:
    version:
      required: false
      type: string
  outputs:
    version: ${{ jobs.version.outputs.version }}
    safe_version: ${{ jobs.version.outputs.safe_version }}
```

The caller (`release-publish.yml`) uses `${{ needs.build.outputs.safe_version }}` to construct
the artifact download name.

### Prerelease Detection

```bash
# In release-publish.yml's publish job:
if echo "${{ inputs.version }}" | grep -qP '\-(alpha|beta|rc|pre|dev)'; then
  echo "prerelease=true" >> "$GITHUB_OUTPUT"
else
  echo "prerelease=false" >> "$GITHUB_OUTPUT"
fi
```

## Risks

- **`workflow_call` artifact race**: If the called workflow's `upload` job re-uploads after the caller
  starts downloading, the caller might get a partial artifact. Mitigation: the `publish` job `needs`
  the called workflow to complete before running.
- **SemVer regex fragility**: The strict SemVer regex needs careful testing against edge cases
  (`v1.0.0-alpha.beta`, `v1.0.0+build.1-2`). Mitigation: add comprehensive Go test cases for
  `ValidateStrictSemVer`.
- **`softprops/action-gh-release@v2` version pinning**: External action dependency. Mitigation:
  pin to major version `@v2` (already standard practice in this repo).
- **No dry-run for publish**: Creating a real release creates permanent artifacts (tag, release page).
  Mitigation: test with prerelease versions first, document cleanup procedure (delete tag + release
  via GitHub UI/API if testing goes wrong).

## Ready for Proposal

**Yes.** The exploration confirms:
- `release-build.yml` needs a ~15-line `workflow_call` refactoring
- New `release-publish.yml` workflow is the right approach for publishing
- Strict SemVer validation must be added as a new Go function
- Non-overwrite is guaranteed by tag existence check + `action-gh-release` default behavior
- Prerelease evidence is testable via manual dispatch with prerelease version strings
- The `ci-release-delivery` change directory (currently empty) can be cleaned up
- Recommend `sdd-propose` for `github-release-publishing` with Approach 1 (`workflow_call` reuse)
