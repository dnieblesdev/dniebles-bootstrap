# Proposal: Installation State Detector

## Intent

Teach planning to distinguish resources that are already present on the host from resources that still need installation, without coupling `internal/planning` to host probing or installers.

## Scope

### In Scope
- Add `InstallationState` and `PlanStepStatusAlreadyInstalled` to `internal/planning`.
- Update `BuildPlan` to consume caller-supplied installation state after environment matching and before normal planned/attention status selection.
- Add `internal/state` with PATH-based detection for `tool` and `runtime` resources via injectable seams.
- Add deterministic unit coverage for planning state semantics and host-independent detection.

### Out of Scope
- Installers, package-manager integration, shell command runner, dotfiles runtime, apply/install command, and TUI.
- Package and dotfile detection implementations; keep future seams only if they clarify the adapter boundary.
- CLI state detector wiring and already-installed rendering; defer to a follow-up slice unless design finds a blocking reason.

## Capabilities

### New Capabilities
- `installation-state`: Planning-time installed-resource state, already-installed plan status, and host-independent state detection seams.

### Modified Capabilities
- None.

## Approach

Use the same boundary as `EnvironmentFacts` and `ConfigState`: callers provide `InstallationState`, and planning stays pure. `internal/state` detects `tool` and `runtime` presence with `exec.LookPath` behind an injectable lookup seam. `BuildPlan` keeps present resources in `Plan.Steps` and records `already_installed` in `Results`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/planning/types.go` | Modified | Add state type and status constant. |
| `internal/planning/builder.go` | Modified | Consume installation state during status selection. |
| `internal/planning/builder_test.go` | Modified | Cover present, absent, mixed, and zero-state behavior. |
| `internal/state/` | New | Add PATH-based detector and tests with injectable seams. |
| `cmd/dbootstrap` | Deferred | Update BuildPlan call sites with empty state only if signature requires it; no CLI detector wiring. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `BuildPlan` signature churn touches call sites | Med | Keep change mechanical; pass zero `InstallationState` where CLI wiring is deferred. |
| PATH presence overstates readiness | Med | Document it as a planning-time heuristic, not version or health validation. |
| `already_installed` conflicts with attention-required semantics | Med | Define precedence in specs and tests before implementation. |

## Rollback Plan

Revert the planning type/status additions, `BuildPlan` state parameter, `internal/state`, and related tests. Existing planning behavior returns to `planned`, `attention_required`, `skipped`, and `error` only.

## Dependencies

- Existing catalog resource kinds and `ResourceRef` naming.
- Go standard library PATH lookup semantics via `exec.LookPath`.

## Success Criteria

- [ ] Planning marks state-present `tool`/`runtime` resources as `already_installed` while preserving step ordering.
- [ ] Empty installation state preserves current plan output semantics.
- [ ] State detection tests are deterministic and do not depend on the real host PATH.
