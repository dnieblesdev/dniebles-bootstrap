# Exploration: installation-state-detector

## Current State

The project has a working planning pipeline: `internal/catalog/toml` loads `catalog/bootstrap.toml` into a `planning.Catalog`, `internal/environment` detects OS/arch/distro/WSL into `planning.EnvironmentFacts`, and `planning.BuildPlan` merges them into a dependency-ordered `PlanResult`. The CLI `dbootstrap plan` wires these together and renders deterministic output.

However, planning currently only distinguishes three positive outcomes for included resources:

- `planned` — resource will be installed.
- `attention_required` — resource needs config keys that are missing.
- `skipped` — environment facts don't match resource conditions (the resource is excluded from steps entirely).

There is no `already_installed` status. When the plan includes `tool:git` or `runtime:go`, it always shows `planned` even when those tools already exist on the host. The original design document explicitly called for `already-installed` as a `PlanStepResult` status, but that concept has not been implemented yet.

The existing `internal/environment` package demonstrates the correct adapter pattern for this repo: injectable function-type seams (`RuntimeSource`, `EnvSource`, `FileSource`) that keep tests host-independent, and a struct-based detector that falls back to real OS calls when seams are nil.

## Affected Areas

- `internal/planning/types.go` — needs a new `PlanStepStatusAlreadyInstalled` constant and an `InstallationState` input type.
- `internal/planning/builder.go` — `BuildPlan` must accept `InstallationState` and produce `already_installed` for resources detected as present.
- `internal/planning/builder_test.go` — new test cases for already-installed semantics with and without state.
- `internal/state/` (new) — package for the state detector adapter: PATH lookup for tools/runtimes, package detection seam (no implementation yet), host-independent test fixtures.
- `cmd/dbootstrap/render.go` — may need to render `already_installed` status in output (can be deferred to a follow-up wiring slice).
- `cmd/dbootstrap/main.go` — will eventually consume the state detector and feed `InstallationState` into `BuildPlan` (can be deferred).
- `openspec/changes/installation-state-detector/` — this exploration and subsequent SDD artifacts.

## Approaches

### 1. Detection adapter only (no planning changes)

Create `internal/state` with a detector that returns `map[ResourceRef]bool`. Do not modify `BuildPlan` or add `already_installed` to planning. Return raw state to the caller.

| Pros | Cons | Complexity |
|------|------|------------|
| Zero risk to planning core | CLI/consumer must merge state with plan results manually | Low |
| Fastest slice to ship | Duplicates status logic at every call site | |
| Pure infrastructure adapter | `already_installed` is still invisible in `PlanResult` | |

### 2. Planning domain change + adapter together

Add `InstallationState` to `planning` types, add `PlanStepStatusAlreadyInstalled`, modify `BuildPlan` to accept state input and mark present resources accordingly. Create `internal/state` detector alongside.

| Pros | Cons | Complexity |
|------|------|------------|
| Status lives in the domain where it belongs | Touches planning core (moderate risk) | Medium |
| Every consumer gets `already_installed` for free | Requires careful test updates | |
| Follows existing `EnvironmentFacts`/`ConfigState` pattern | | |

### 3. Separate planning function (don't modify BuildPlan)

Add `InstallationState` type and `already_installed` status to types, but create a new `BuildPlanWithState()` function or a post-processing step that annotates a `PlanResult` with state. Keep `BuildPlan` signature unchanged.

| Pros | Cons | Complexity |
|------|------|------------|
| `BuildPlan` callers don't break | Two functions doing nearly the same thing | Medium |
| Clear separation of concerns | Future callers must remember to use the right one | |
| | `BuildPlan` becomes the "legacy" path | |

## Recommendation

Use **Approach 2** — planning domain change + adapter together.

The project already has the pattern of caller-supplied facts (`EnvironmentFacts`, `ConfigState`). Adding `InstallationState` is consistent, not a new paradigm. The planning core already sorts resources into statuses based on facts and config — adding installation state is the same kind of concern. Approach 1 would force every consumer to reimplement status logic, which violates the design goal of "one core with thin interfaces." Approach 3 creates an unnecessary fork in the API.

**Package location**: `internal/state` — clear, self-documenting, doesn't collide with `internal/environment`, and follows the convention of naming packages after the domain concept they represent (`planning`, `environment`, `catalog`).

**Detection strategy for this slice**:

| Resource Kind | Detection Method | Rationale |
|---|---|---|
| `tool` | `exec.LookPath(name)` via injectable seam | Tools like `git`, `ripgrep` are CLI binaries on PATH |
| `runtime` | `exec.LookPath(name)` via injectable seam | `go`, `node`, `python` are also PATH binaries |
| `package` | Interface/seam only; no implementation | Package manager checks (dpkg, rpm, brew) are explicitly out of scope for this slice |
| `dotfile` | Interface/seam only; no implementation | Dotfiles state detection is a future concern |

**Planning changes needed**:

1. `types.go`: Add `InstallationState` struct `{ Present map[ResourceRef]bool }` and `PlanStepStatusAlreadyInstalled`.
2. `builder.go`: Accept `InstallationState` in `BuildPlan` params; after environment-fact matching succeeds, check `state.Present[ref]` — if true, status becomes `already_installed` instead of `planned`/`attention_required`. Resource still appears in `Plan.Steps` so CLI can show it.
3. `builder_test.go`: Add test cases covering state-present resources, mixed present/absent, and regression checks that absent state behaves identically to today.

**CLI plan wiring**: Defer to the next slice. The detector and planning changes are independently testable. Wiring `dbootstrap plan` to consume state can be a focused follow-up. This keeps each slice small and reviewable.

## Risks

- Adding a parameter to `BuildPlan` changes its signature. Existing callers in tests (`builder_test.go`, `catalog_test.go`, `main_test.go`, `render_test.go`) must be updated with an empty/zero `InstallationState`. This is mechanical but affects several files.
- PATH-based detection is a heuristic. A binary on PATH doesn't guarantee it's the right version or functional. But it's the best available signal without package-manager integration, and it's good enough for "already there" detection.
- The `already_installed` status could be confused with "installation complete" after an installer runs. Naming should clearly distinguish planning-time state detection from runtime installation results. Using `already_installed` (past tense, planning-time observation) is clearer than `installed` (which could mean "just installed").

## Ready for Proposal

Yes. The next step should propose a focused slice that:

1. Adds `InstallationState` and `PlanStepStatusAlreadyInstalled` to `internal/planning`.
2. Modifies `BuildPlan` to consume state and produce the new status.
3. Creates `internal/state` with a PATH-lookup-based detector for tools and runtimes, using the same injectable-seam pattern as `internal/environment`.
4. Keeps CLI wiring, package detection, dotfiles state, and installers out of scope.
