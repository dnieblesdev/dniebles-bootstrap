# Proposal: Config State Awareness

## Intent

Make `dbootstrap plan` aware of existing local configuration state instead of passing an empty `planning.ConfigState{}`. The slice is read-only: it detects whether required config keys appear to exist and lets the pure planner keep deciding status. It must not install, apply, mutate files, invoke dotfiles tooling, or claim dotfiles ownership.

## Scope

### In Scope
- Add an `internal/config` detector with injectable filesystem/path seams.
- Wire detected config state into `cmd/dbootstrap plan` after catalog load and before `BuildPlan`.
- Define the config key → filesystem path mapping as convention-based, read-only, and testable.
- Add deterministic tests for detector behavior and CLI wiring.
- Optionally clean stale README wording about “real environment probing”.

### Out of Scope
- Applying/installing resources or mutating dotfiles.
- Parsing, validating, owning, or invoking dotfiles runtime semantics.
- Changing catalog schema for explicit config paths.
- Adding detector failure diagnostics beyond the current no-error detector pattern.

## Capabilities

### New Capabilities
- `config-state-awareness`: Detect required configuration key presence from local filesystem conventions and pass it to planning without side effects.

### Modified Capabilities
- None.

## Approach

Create `internal/config` following `internal/environment` and `internal/state`: a package-level `Detect(catalog)` backed by a `Detector` with injectable seams. The detector derives required config keys from catalog resources, maps keys to paths using an injectable convention/base path, performs existence checks only, and returns `planning.ConfigState{PresentKeys: ...}`. `cmd/dbootstrap` adds a package-level detector var for host-independent tests and replaces the current hardcoded empty config state.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/` | New | Read-only config-state detector and tests. |
| `cmd/dbootstrap/main.go` | Modified | Wire detected `ConfigState` into `BuildPlan`. |
| `cmd/dbootstrap/main_test.go` | Modified | Stub config detector seam. |
| `internal/planning/` | Unchanged | Remains pure and consumes caller-supplied config state. |
| `README.md` | Modified | Optional stale current-status wording cleanup. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Dotfiles layout convention is wrong | Med | Keep mapping injectable and isolated for later replacement. |
| Boundary drift into dotfiles ownership | Med | Limit detector to existence checks; no parsing, validation, mutation, or runtime calls. |
| Host-dependent tests | Low | Use injected filesystem/path seams and CLI detector var. |

## Rollback Plan

Remove `internal/config`, restore `planning.ConfigState{}` at the CLI call site, and revert related tests/README wording. No data migration is needed because the change is read-only.

## Dependencies

- Existing catalog `config_required` keys.
- Existing planner `ConfigState` semantics.
- Local dotfiles path convention, injected for tests.

## Success Criteria

- [ ] `dbootstrap plan` passes detected config state instead of an empty state.
- [ ] Missing config still renders as `attention_required`; present config no longer triggers missing-key attention.
- [ ] Planning remains pure and host probing stays outside `internal/planning`.
- [ ] Tests prove detector and CLI wiring without depending on the real host.
