## Exploration: minimize-release-workflow-permissions

### Current State

The `release-publish.yml` workflow (`.github/workflows/release-publish.yml`) has a workflow-level
`permissions` block that grants `actions: write` globally:

```yaml
permissions:
  contents: read
  actions: write          # ← global — broader than needed
```

Every job that needs elevated permissions already declares them explicitly:

| Job | Explicit Permissions | Needs |
|-----|---------------------|-------|
| `validate` | `contents: read` | `checkout`, `setup-go`, CLI run |
| `build` | none (inherits globals) | `workflow_call` to `release-build.yml` (same repo) |
| `publish` | `contents: write`, `actions: read` | `gh release create`, `download-artifact` |

The `actions: write` global scope is never consumed by any step:
- `validate` only checks out code and runs the validate CLI.
- `build` is a same-repo `workflow_call` — GitHub's reusable workflow docs confirm
  same-repository calls need only `contents: read`.
- `publish` already has its own explicit `actions: read` for artifact downloads.

The current spec (from the archived `github-release-publishing` change) requires:
> Only the publication job MAY have contents write permission.

`actions: write` is not mentioned in the spec at all — it was included unnecessarily during
implementation.

### Affected Areas

- `.github/workflows/release-publish.yml` — single-line deletion (line 13: `actions: write`)

### Approaches

1. **Remove `actions: write` from globals (Recommended)**
   - Delete line 13 (`actions: write`) from the workflow-level `permissions:` block.
   - Keep `contents: read` at global level and all job-level permissions unchanged.
   - **Pros**: Least-privilege alignment, one-line change, zero behavioral impact, no Go code
     changes, no CI impact.
   - **Cons**: None.
   - **Effort**: Low

2. **Remove `actions: write` and add explicit `contents: read` to the `build` job**
   - Add `permissions: contents: read` to the `build` job alongside removing the global line.
   - **Pros**: Even more explicit about the `build` job's boundary.
   - **Cons**: Unnecessary — `build` already inherits `contents: read` from globals; adds
     verbosity without improving security.
   - **Effort**: Low

### Permission-usage trace (post-change)

After removing `actions: write`, the effective permissions per job are:

```
workflow-level:  contents: read ──── default for all jobs
    │
    ├── validate:  contents: read   (explicit override — unchanged)
    │
    ├── build:     contents: read   (inherited from workflow-level — same repo
    │               workflow_call needs only this)
    │
    └── publish:   contents: write  (explicit — for gh release create)
                   actions: read    (explicit — for download-artifact)
```

### CI and barrier validation

**Go suite:** No Go code changes. `go test ./...` and `go vet ./...` are green and unaffected.

**YAML validity:** The change is a single-line deletion. The resulting `permissions:` block
(`contents: read`) is valid YAML and a valid GitHub Actions permissions mapping.

**Invalid-version barrier:** The `validate` job runs `go run ./internal/version/cmd/validate
--release --version "${INPUT_VERSION}"` which calls `version.ValidateReleaseTag()`. Invalid
input causes `exit 1` → `needs: validate` blocks `build` and `publish`. This barrier is
entirely permission-independent — `actions: write` was never needed for it.

### Recommendation

Approach 1: Remove `actions: write` from the global permissions block (line 13). This is a
pure least-privilege cleanup — no job ever consumed the global `actions: write` permission,
and every operation that needs elevated access already declares it at the job level.

### Risks

- None. The permission being removed is not consumed by any step. The `publish` job retains
  `contents: write` and `actions: read` explicitly. The `build` job's `workflow_call` to a
  same-repo reusable workflow needs only `contents: read`.

### Ready for Proposal

Yes — this is a well-scoped, single-line removal with zero behavioral impact. Proceed to
`sdd-propose`.
