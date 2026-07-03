# dniebles-bootstrap

`dniebles-bootstrap` is the planning home for a personal development-environment bootstrapper. It will become a domain-first Go orchestrator that plans and runs profile installs and point installs while keeping dotfiles ownership external.

## Current status

This repository has its first pure Go planning-core slice, an isolated TOML catalog adapter, and a minimal CLI plan command. It accepts decoded domain inputs and builds deterministic plans without installers, command execution, dotfiles runtime integration, TUI wiring, or real environment probing.

- Go application code currently includes `internal/planning` domain/value types, a pure plan builder, the `internal/catalog/toml` adapter, `cmd/dbootstrap` plan command wiring, and table-driven unit/integration tests.
- A repository-local TOML catalog fixture exists at `catalog/bootstrap.toml`; it decodes into planning inputs while planner-owned semantics remain in `internal/planning`.
- The accepted direction is captured under `openspec/changes/archive/2026-07-03-design-bootstrap-orchestrator/`, `openspec/changes/first-go-planning-slice/`, and `openspec/changes/catalog-toml-adapter/`.

## CLI usage

Build or run the current CLI from the repository root:

```sh
go run ./cmd/dbootstrap plan --profile dev
```

The command loads `catalog/bootstrap.toml` by default. Use `--catalog <path>` to point at another local TOML catalog file. This slice only plans with static environment facts (`linux/amd64`) and empty configuration state; it does not probe the host or apply/install anything.

## Goals and non-goals

| Goal | Decision |
|------|----------|
| Fresh-machine bootstrap | Provide a path to make `dbootstrap` available and then let the Go application own orchestration. |
| Profile installs | Plan and execute named environment profiles made of bundles, tools, runtimes, packages, and dotfiles requests. |
| Point installs | Install or reconcile one requested tool, runtime, package, bundle, or capability without pulling unrelated scope. |
| Domain-first core | Keep planning, dependency ordering, execution, and reporting in one shared core. |
| CLI now, TUI later | Start with a CLI interface and preserve a future Bubble Tea TUI as a thin interface over the same core. |

| Non-goal | Boundary |
|----------|----------|
| Dotfiles internals | `~/.dotfiles` owns modules, configs, assets, symlinks, validations, and `dotlink` semantics. |
| Shell orchestration | A shell wrapper may acquire `dbootstrap`, but it must not resolve catalogs, run installers, or own reporting. |
| Runtime execution in this slice | This change does not add installers, CLI commands, command runners, or runtime OS probing. |

## Install flows

### Profile install

1. Detect environment facts: OS, distro, WSL status, and CPU architecture.
2. Resolve the requested profile from the repository catalog.
3. Expand bundles, tools, runtimes, packages, and dotfiles module requests.
4. Build a dependency-ordered plan.
5. Execute through installer and dotfiles provider adapters.
6. Report structured results: installed, already-installed, skipped, failed, or attention-required.

Missing expected dotfiles configuration should not block tool installation. It must remain visible as an attention-required result.

### Point install

1. Detect environment facts.
2. Resolve only the requested point target.
3. Avoid unrelated catalog resources.
4. Install or skip based on existing state.
5. Request only the dotfiles modules needed for that point target.
6. Report the result and any missing configuration attention item.

## First-run bootstrap entrypoint

The first-run entrypoint is intentionally small. Its job is to make `dbootstrap` available, then hand control to the Go application.

Supported entrypoint paths:

- Download and install a compatible released `dbootstrap` binary when available.
- Install or use Go to compile/run `dbootstrap` from this repository when a binary is unavailable or source execution is preferred.

After `dbootstrap` starts, the Go application owns catalog resolution, dotfiles integration, installer selection, dependency ordering, plan execution, and operational reporting.

## Architecture direction

Future implementation should preserve these layers:

| Layer | Responsibility |
|-------|----------------|
| Domain/core | Profiles, catalog concepts, plans, plan steps, validation semantics, and structured statuses. |
| Application/use cases | Profile planning, point planning, execution orchestration, and result aggregation. |
| Infrastructure | TOML catalog adapter, installers, command runner, first-run acquisition, git sparse checkout, and dotfiles provider. |
| Interfaces | CLI first; future TUI as a thin presenter/controller over the same use cases. |

## Dotfiles boundary

`dniebles-bootstrap` integrates with dotfiles as an external provider. It may request modules, use partial clone and sparse checkout strategies, and invoke `dotlink` as a provider operation.

It must not own or duplicate dotfiles module internals, declarative profile semantics, symlink lifecycle, asset layout, validations, or configuration files.

## Catalog direction

The catalog belongs in this repository. TOML is the first implemented authoring format because it is readable and maps well to Go structs, but the domain model remains format-agnostic and schema-versioned so the format can evolve later.

The TOML adapter lives in `internal/catalog/toml`, and the initial fixture lives in `catalog/bootstrap.toml`. Adapter validation is intentionally shallow: TOML syntax, required fields, duplicate IDs, supported refs, and basic local references stay in the adapter; dependency expansion, environment filtering, missing config attention, and other planner semantics stay in `internal/planning`.

## Project guidance

- Use SDD/OpenSpec artifacts before implementation.
- Keep generated technical artifacts in English.
- Keep `.atl/` local and ignored.
- See `AGENT.md` for repository operating rules.
