# Design: Minimize Release Workflow Permissions

## Technical Approach

Make the approved one-line deletion in `.github/workflows/release-publish.yml`:
remove workflow-level `actions: write`, retaining global `contents: read` and
every job definition unchanged. This implements the least-privilege delta while
preserving the existing validation → reusable build → publication flow.

## Architecture Decisions

| Decision | Alternatives considered | Rationale |
|---|---|---|
| Remove only global `actions: write` | Add per-job declarations; alter reusable build permissions | The line is unused by `validate` and the same-repository `build`; `publish` already explicitly declares its required `contents: write` and `actions: read`. A single deletion minimizes behavioral risk. |
| Prove behavior by static and workflow validation | Change Go validation code; add product tests | `ValidateReleaseTag` and the `needs` graph already enforce the invalid-version barrier. This slice changes neither; validation must demonstrate that fact rather than duplicate it. |

## Data Flow

```
workflow_dispatch(version)
        │
        ▼
validate (contents: read) ──invalid──► fail; build/publish skipped
        │ valid
        ▼
build reusable workflow (read-only) ──► artifact
        │
        ▼
publish (contents: write, actions: read) ──► checksum verification → GitHub Release
```

The workflow default becomes `contents: read` only. Job-level permissions remain

## File Changes

| File | Action | Description |
|---|---|---|
| `.github/workflows/release-publish.yml` | Modify | Delete only the workflow-level `actions: write` entry. |
| `openspec/changes/minimize-release-workflow-permissions/design.md` | Create | Record this implementation design. |

## Interfaces / Contracts

No interface, workflow trigger, input, reusable-workflow output, artifact, or
release contract changes. The permissions contract after the edit is:

```yaml
permissions:
  contents: read
# publish only:
#   contents: write
#   actions: read
```

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Static | YAML structure and exact permission scopes | Parse/inspect `release-publish.yml`; assert global mapping is only `contents: read`, and publish retains its two explicit scopes. |
| Focused | Invalid-version barrier | Run the existing release-tag validator with invalid/unprefixed values; inspect `build` and `publish` both retain `needs: validate`. |
| Workflow | Valid release behavior | If GitHub runtime validation is available, dispatch a controlled valid release workflow and verify build, artifact download, checksum verification, and publication complete with reduced scope. |

## Migration / Rollout

No migration required. Roll back by restoring the single line only if GitHub
runtime evidence proves an operation requires it.

## Open Questions

None.
