# Proposal: Minimize Release Workflow Permissions

## Intent

Remove an unused global GitHub Actions write grant while preserving release validation, artifact build, and publication behavior.

## Scope

### In Scope
- Delete exactly `.github/workflows/release-publish.yml:13` — `actions: write`.
- Preserve global `contents: read` and all job-level permissions unchanged.
- Validate YAML and inspect effective permissions, including the invalid-version barrier.

### Out of Scope
- Changes to release-build, release assets, validation logic, or workflow triggers.
- Adding redundant job-level permission declarations.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `github-release-publishing`: require least-privilege workflow permissions: no global `actions: write`; only `publish` retains its explicit write authority.

## Approach

Make the one-line deletion. `validate` remains read-only; same-repository `build` inherits `contents: read`; `publish` retains `contents: write` and `actions: read`. Confirm invalid versions still stop both downstream jobs before side effects.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `.github/workflows/release-publish.yml` | Modified | Remove the unused global permission only. |
| `openspec/specs/github-release-publishing/spec.md` | Modified | Archive the approved permission requirement delta. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Reusable build unexpectedly requires write access | Low | Confirm same-repository call remains read-only and inspect job scopes. |

## Rollback Plan

Restore the single `actions: write` line if GitHub runtime evidence shows a required operation fails; retain the validation and publishing job configuration.

## Dependencies

- GitHub Actions permission semantics for same-repository reusable workflows.

## Success Criteria

- [ ] The workflow-level permissions mapping contains only `contents: read`.
- [ ] `publish` alone has `contents: write`; artifact download retains `actions: read`.
- [ ] YAML validation and relevant repository checks pass; invalid input still blocks `build` and `publish`.
- [ ] On archive, the approved delta is merged into `github-release-publishing` without unrelated spec changes.
