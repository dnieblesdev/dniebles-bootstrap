# Proposal: Point Install Planning

## Intent

Expose existing point-resource planning support through `dbootstrap plan` so users can preview specific resources without selecting a full profile. The change keeps planning read-only and avoids installer, apply, mutation, or runtime execution scope.

## Scope

### In Scope
- Add repeatable `--resource kind:name` to `dbootstrap plan`.
- Keep `--profile` optional when resources are supplied; require at least one of profile or resource.
- Allow profile + resource union planning and clear validation/errors.
- Update rendered headers for resource-only plans.

### Out of Scope
- Apply/install/mutation/runtime execution.
- Planner/domain model changes unless implementation proves unavoidable.
- New point subcommands or positional resource syntax.

## Capabilities

### New Capabilities
- `point-install-planning`: CLI planning can target explicit resource refs, optionally unioned with a profile, while remaining read-only.

### Modified Capabilities
- None.

## Approach

Wire the CLI edge to the domain support that already exists. Parse repeatable `--resource` values as `kind:name`, populate `planning.PlanRequest.Resources`, preserve existing profile behavior, and reject calls with neither `--profile` nor `--resource`. Prefer a small shared resource-ref parser if it stays simple; otherwise keep CLI duplication minimal and tested.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modified | Add flag, validation, parsing, and `PlanRequest.Resources` wiring. |
| `cmd/dbootstrap/main_test.go` | Modified | Cover point-only, mixed profile/resource, invalid refs, and existing profile-only behavior. |
| `cmd/dbootstrap/render.go` | Modified | Show resource-oriented header when no profile is supplied. |
| `internal/planning/*` | Unchanged | Existing `PlanRequest.Resources` and `BuildPlan` expansion should be reused. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| CLI parser diverges from catalog parser | Med | Extract shared parser only if it reduces duplication without expanding scope. |
| Mixed profile/resource semantics confuse users | Low | Document union behavior in help, tests, and output expectations. |
| Scope creep into install/apply | Low | Keep tests and specs explicit: plan remains read-only. |

## Rollback Plan

Remove the `--resource` flag path, restore profile-required validation and old header behavior, and keep domain code untouched.

## Dependencies

- Existing catalog resource kinds: `tool`, `runtime`, `package`, `dotfile`.
- Existing `planning.PlanRequest.Resources` and `BuildPlan` expansion.

## Success Criteria

- [ ] `dbootstrap plan --resource tool:git` builds a resource-only plan.
- [ ] `--profile dev --resource runtime:go` produces union planning.
- [ ] Missing target and malformed/unsupported refs return clear errors.
- [ ] No apply/install/mutation/runtime execution is introduced.
