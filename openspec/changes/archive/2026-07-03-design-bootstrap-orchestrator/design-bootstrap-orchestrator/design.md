# Design: Bootstrap Orchestrator

## Technical Approach

This documentation-first change defines `dniebles-bootstrap` as a domain-first Go bootstrap orchestrator without implementing runtime code. A tiny first-run wrapper may only make `dbootstrap` available, then hand control to the Go binary/application. The core detects environment facts, plans profile and point installs from an in-repo catalog, orders dependencies, delegates execution, and reports results for CLI now and future TUI later.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Core shape | Domain/core concepts first: Profile, Tool, Runtime, Package, Bundle, Catalog, Plan, PlanStep, Resource, ResourceRef, ConfigPolicy | Command-first scripts | Planning, validation, retries, and future TUI need stable concepts independent of interfaces. |
| Layers | Domain → application/use-case orchestration → infrastructure adapters → interfaces | Put logic in CLI commands | Keeps install ordering, missing config policy, and provider boundaries reusable by CLI now and TUI later. |
| Catalog | In-repo TOML first; domain model remains format-agnostic and schema-versioned | YAML/JSON or dotfiles-owned catalog | TOML is readable and Go-friendly; schema/versioning prevents early format lock-in. |
| Dotfiles | DotfilesProvider uses partial clone `--filter=blob:none` and sparse checkout for requested DotfilesModule scopes | Vendor or model dotfiles modules in bootstrap | Bootstrap may request modules and invoke `dotlink`, but module internals, assets, symlinks, validations, and configs stay external. |
| Missing config | Tool install proceeds; Plan/PlanStep result marks missing config as attention-required | Hard fail installs | Specs require installation continuity while preserving visible follow-up work. |
| First run | FirstRunEntrypoint/BootstrapEntrypoint is interface/infrastructure: download released `dbootstrap` when available, or install/use Go to compile/run from repo | Shell orchestrator | The wrapper solves availability only; catalog, dotfiles, installers, dependency ordering, plan execution, and reporting stay in Go. |
| Reporting | PlanStepResult uses installed, already-installed, skipped, failed, attention-required | Free-form logs only | Structured results support CLI output, future TUI views, and auditable logs. |

## Data Flow

```text
first-run wrapper ─→ dbootstrap binary ─→ CLI request ─→ Application use case
                                                    ├─→ EnvironmentDetector
                                                    ├─→ Catalog adapter ─→ Resolver
                                                    └─→ Runner ─→ Installer/DotfilesProvider
```

First startup flow: wrapper checks for a released `dbootstrap` binary and installs it when desired/available; otherwise it installs or uses Go to compile/run the bootstrapper from this repo. After that, the Go application owns all orchestration. Binary path is faster and more reproducible; Go compile/run path keeps the bootstrapper usable before releases or when source execution is preferred.

```text
CLI request ─→ Application use case ─→ EnvironmentDetector(OS/distro/WSL/arch)
                     │
                     ├─→ Catalog(TOML adapter) ─→ Domain resolver
                     │                         └─→ dependency-ordered Plan[PlanStep]
                     └─→ Runner ─→ Installer adapters ─→ PlanStepResult
                               └─→ DotfilesProvider ─→ git sparse checkout + dotlink
```

Profile install lifecycle: detect environment → resolve Profile → expand Bundle and ResourceRef entries → validate Catalog and ConfigPolicy → dependency-sort Tool/Runtime/Package/Bundle actions → install resources → request needed DotfilesModule scopes → partial clone or update `~/.dotfiles` with sparse checkout → invoke `dotlink` as provider operation → report installed, already-installed, skipped, failed, and attention-required results.

Point install lifecycle: detect environment → resolve one Tool/Runtime/Package/Bundle target → avoid unrelated resources → check existing state → install if needed → request only target DotfilesModule scopes → run scoped `dotlink` → report result and any missing config attention item.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `openspec/changes/design-bootstrap-orchestrator/design.md` | Update | Technical design for documentation-first architecture direction. |
| `openspec/changes/design-bootstrap-orchestrator/specs/` | Update | Delta specs for orchestration, catalog planning, environment detection, first-run entrypoint, dotfiles integration, profile/point planning, and repository guidance. |
| `README.md` | Modify | Project orientation: purpose, current status, goals/non-goals, flows, architecture direction, dotfiles boundary, catalog direction, and CLI-now/TUI-later path. |
| `AGENT.md` | Create | Agent/contributor guide: no code before specs, SDD workflow, `.atl/` ignored, English docs/specs, dotfiles boundary, first-run wrapper boundary, and one-core/two-thin-interfaces guidance. |
| `catalog/` | Future create | Repository-local TOML catalog, outside this documentation-only slice. |

## Interfaces / Contracts

Core concepts are format-independent. `Catalog` owns declared Profile, Bundle, and Resource records. `Resource` covers Tool, Runtime, and Package variants and may hold dependencies, `ConfigPolicy`, and `DotfilesModule` references. `EnvironmentDetector` supplies OS/distro/WSL/architecture facts before resolution. `Plan` is a dependency-ordered collection of `PlanStep` entries. `Runner` executes steps through `Installer` adapters and returns structured `PlanStepResult` values: installed, already-installed, skipped, failed, attention-required. `DotfilesProvider` accepts `DotfilesModule` requests and returns provider results without exposing module internals. `FirstRunEntrypoint` only acquires or runs `dbootstrap` and never owns catalog, dotfiles integration, installers, or plan execution.

Layer boundaries:
- Domain/core: entities, relationships, validation semantics, plan state, attention-required status.
- Application/use cases: profile install planning, point install planning, execution orchestration.
- Infrastructure: first-run acquisition wrapper, TOML catalog adapter, git partial clone/sparse checkout, command runner, installers, dotfiles provider.
- Interfaces: CLI now; Bubble Tea TUI later as a thin presenter/controller over use cases.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Documentation | README/AGENT/design/spec alignment | Manual inspection during apply/verify. |
| Future Unit | Catalog expansion, ConfigPolicy, plan ordering | Table-driven Go tests once code exists. |
| Future Integration | Provider sparse checkout and installer boundaries | Adapter tests with fake Runner/DotfilesProvider. |
| E2E | Profile and point install flows | Deferred until CLI/runtime exists. |

## Migration / Rollout

No migration required. This change only creates design/spec/docs artifacts. Future implementation should start with catalog schema and pure planning before installers.

## Open Questions

- [ ] Exact TOML file layout and schema version field name will be decided in the implementation proposal.
- [ ] Exact CLI command names are intentionally deferred.
