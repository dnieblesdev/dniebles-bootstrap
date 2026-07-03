# Design: First Go Planning Slice

## Technical Approach

Create the first Go slice as a pure, format-agnostic planning core. The package accepts already-decoded domain values, expands profiles and bundles, orders dependencies deterministically, and returns a `Plan` plus structured attention/error information. No TOML parsing, CLI, installer, dotfiles, OS probing, or command execution enters this slice.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Package boundary | `internal/planning` with domain types, validation, and builder | Split domain/application early or use `internal/core` | One small package keeps the first slice understandable; `planning` names the behavior being proven and can split later when real use cases/adapters appear. |
| Input model | In-memory `Catalog` containing `Profiles`, `Bundles`, and `Resources` maps/slices | TOML-shaped structs | Prevents parser/schema leakage and satisfies domain-only planning requirements. |
| Builder API | Pure function: `BuildPlan(catalog Catalog, request PlanRequest, facts EnvironmentFacts, state ConfigState) PlanResult` | Stateful planner service | A pure function is easier to table-test and makes side effects impossible by construction. |
| Ordering | Stable graph expansion with deterministic sorting before topological dependency ordering | Preserve declaration order only | Stable topological ordering keeps identical inputs reproducible while still honoring dependencies. |
| Missing config | Emit attention-required step/result metadata and continue unrelated planning | Hard fail planning | Specs require missing config to stay visible without blocking valid resources. |

## Data Flow

```text
Catalog + PlanRequest + EnvironmentFacts + ConfigState
        │
        ├─→ validate refs and collect diagnostics
        ├─→ expand Profile → Bundle → ResourceRef → Resource
        ├─→ filter by EnvironmentFacts / existing state
        ├─→ dependency-sort resources
        └─→ Plan{Steps[]} + PlanStepResult diagnostics
```

`PlanStep` describes intended work only. `PlanStepResult` describes planning-time statuses such as planned, skipped, attention-required, or error; execution-time installed/failed statuses remain future work.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `go.mod` | Create | Initialize the Go module for testable core code. |
| `internal/planning/types.go` | Create | Domain entities/value objects: catalog, profile, bundle, resource, refs, config, facts, plan, statuses. |
| `internal/planning/builder.go` | Create | Pure plan builder and deterministic expansion/ordering. |
| `internal/planning/builder_test.go` | Create | Table-driven tests for planning behavior. |
| `openspec/changes/first-go-planning-slice/design.md` | Create | This design artifact. |

## Interfaces / Contracts

Core concepts:
- `Catalog`: in-memory collection of `Profile`, `Bundle`, and `Resource` definitions.
- `Profile`: named install scope referencing bundles and/or resources.
- `Bundle`: reusable group of `ResourceRef` entries.
- `Resource`: installable desired item with `ResourceKind`, dependencies, config policy, and optional environment conditions.
- `ResourceRef`: typed stable reference to a bundle/resource.
- `ResourceKind`: tool, runtime, package, bundle-like grouping where needed by planning.
- `ConfigPolicy`: required/optional config expectations.
- `ConfigState`: caller-supplied data describing whether required config is present; missing required config produces attention.
- `EnvironmentFacts`: caller-supplied OS/arch/distro/WSL facts; no probing.
- `Plan`: ordered `PlanStep` list and diagnostics.
- `PlanStep`: desired action for one resource/ref, including dependencies and attention markers.
- `PlanStepResult` / `PlanStepStatus`: structured planning outcome (`planned`, `skipped`, `attention_required`, `error`).

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Profile expansion | Table-driven cases asserting exact plan steps. |
| Unit | Bundle expansion | Nested/shared refs dedupe deterministically. |
| Unit | Dependencies | Dependencies precede dependents across kinds. |
| Unit | Missing config | Required missing config returns attention without stopping valid steps. |
| Unit | Environment conditions | Synthetic facts include/skip resources predictably. |
| Unit | No side effects | Tests use only in-memory values; no commands, files, env probes, or dotfiles calls. |

No integration or E2E tests are in scope because no adapters or interface exist.

## Migration / Rollout

No migration required. This creates the first Go module and pure tests only.

## Out of Scope

TOML schema/loader, catalog files, CLI/TUI, installers, command runner, first-run wrapper, git/dotfiles operations, persisted state, and real environment detection.

## Risks / Tradeoffs

- A single `internal/planning` package may need splitting later; acceptable until boundaries are proven by real adapters.
- Deferring TOML means schema assumptions remain unvalidated; next slice should add the adapter against this domain contract.
- Planning-time statuses intentionally differ from execution results; names must stay clear to avoid pretending installation happened.

## Open Questions

- [ ] Exact Go module path should be confirmed during implementation if repository remote/module naming differs.
