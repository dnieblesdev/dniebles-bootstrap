# Proposal: CLI Plan Command

## Intent

Add the first executable slice for `dniebles-bootstrap`: a minimal `dbootstrap plan --profile dev` command that proves the existing TOML catalog adapter and pure planning core can produce a deterministic human-readable plan without runtime side effects.

## Scope

### In Scope
- Add a tiny `cmd/dbootstrap` CLI entrypoint with stdlib `flag` parsing for `plan --profile <id>`.
- Load `catalog/bootstrap.toml` through `internal/catalog/toml.LoadFile` and call `planning.BuildPlan`.
- Print deterministic human text output for planned, skipped, attention, and error results.
- Use caller-supplied static `planning.EnvironmentFacts` and empty/static `planning.ConfigState` for this slice.
- Add focused Go tests for command behavior/output through the smallest testable boundary.

### Out of Scope
- `apply`/install commands, installers, command runner, TUI, or CLI framework adoption.
- Git/dotfiles/dotlink runtime behavior.
- OS probing/environment detection adapter.
- Remote catalog loading or alternate catalog locations beyond the repo-local default.

## Capabilities

### New Capabilities
- `cli-plan-command`: Minimal executable planning command and deterministic human plan rendering.

### Modified Capabilities
- None. `planning-core` and `catalog-toml-adapter` are reused without spec-level behavior changes.

## Approach

Use stdlib `flag` to keep the first CLI dependency-free. Keep `main.go` as wiring only; put command execution/rendering behind a small testable function if needed. The command loads `catalog/bootstrap.toml`, builds a profile plan with static facts, sorts/prints output deterministically, and returns non-zero on load/planning errors.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/dbootstrap` | New | CLI entrypoint and `plan` command wiring. |
| `cmd/dbootstrap/*_test.go` | New | Deterministic command/output tests. |
| `internal/catalog/toml` | Reused | Loads repo-local TOML catalog. |
| `internal/planning` | Reused | Builds pure plan from decoded catalog. |
| `catalog/bootstrap.toml` | Reused | Default catalog input for `--profile dev`. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| CLI concerns leak into planning | Med | Keep parsing/rendering at command boundary only. |
| Output becomes flaky | Med | Sort rendered refs/results and test exact text. |
| Static facts surprise users | Low | Document this as first-slice behavior; no host probing yet. |

## Rollback Plan

Revert `cmd/dbootstrap`, related tests, and this change artifact. No installers, migrations, external commands, or host state are introduced.

## Dependencies

- Existing `internal/catalog/toml.LoadFile` and `internal/planning.BuildPlan`.
- No new third-party CLI dependency planned.

## Success Criteria

- [ ] `dbootstrap plan --profile dev` reads `catalog/bootstrap.toml` and exits successfully when planning has no errors.
- [ ] Output is deterministic and includes planned/attention/skipped/error statuses as applicable.
- [ ] Tests cover success, missing/unknown profile or catalog failure, and output ordering.
- [ ] No OS probing, install/apply execution, or CLI framework is introduced.
