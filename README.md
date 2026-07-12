# dniebles-bootstrap

`dniebles-bootstrap` is the planning home for a personal development-environment bootstrapper. It will become a domain-first Go orchestrator that plans and runs profile installs and point installs while keeping dotfiles ownership external.

## Current status

This repository provides a domain-first Go CLI for deterministic planning and explicitly confirmed execution. Planning remains pure; host probing is performed by read-only adapters at the CLI composition root.

- Go application code includes planning, catalog decoding, execution reporting, provider-aware installers, and table-driven unit/integration tests.
- A repository-local TOML catalog fixture exists at `catalog/bootstrap.toml`; it decodes into planning inputs while planner-owned semantics remain in `internal/planning`.
- The accepted direction is captured under `openspec/changes/archive/2026-07-03-design-bootstrap-orchestrator/`, `openspec/changes/first-go-planning-slice/`, and `openspec/changes/catalog-toml-adapter/`.

## CLI usage

Run the CLI from the repository root. The default catalog is `catalog/bootstrap.toml`; use `--catalog <path>` for another local catalog.

## Operational workflow

### Quick path

1. Inspect selected work: `go run ./cmd/dbootstrap plan --profile dev`.
2. Review non-mutating reporting: `go run ./cmd/dbootstrap apply --profile dev` or add `--dry-run`.
3. Confirm eligible execution deliberately: `go run ./cmd/dbootstrap apply --profile dev --yes`.

Select targets with `--profile <name>`, repeatable `--resource <kind:name>`, and `--catalog <path>`. `bootstrap` accepts the same target and safety flags and uses the same execution workflow as `apply`.

### Commands and safety modes

| Command or flag | Behavior |
|---|---|
| `plan` | Inspects the selected work and renders planning statuses; it does not mutate the host. |
| `apply` | Reports execution results by default; only `--yes` confirms eligible execution. |
| `bootstrap` | Uses the same apply execution semantics for an explicit selection; provider/bootstrap needs remain advisory. |
| `--dry-run` | Reports the dry-run mode without mutation. It cannot be combined with `--yes`. |
| `--yes` | Explicitly confirms supported eligible execution. Default and dry-run modes do not mutate the host. |
| `--sudo` | Is meaningful only with confirmed `--yes` where the provider supports it; it does not independently enable mutation. |

### Confirmed reruns

A confirmed `apply --yes` or `bootstrap --yes` avoids installer mutation only when planning has marked an eligible `tool` or `runtime` as `already_installed` after reliable configured-command detection. The resource must have non-nil presence metadata with `Presence.Kind == "command_exists"` and a non-empty `Presence.Name`. The result is reported as `unchanged`: `already installed; no mutation attempted`.

This is intentionally narrow. Package and dotfile resources keep their normal runner behavior. Command presence is not proof of package installation details, package version, configuration correctness, or dotfile-link convergence; dotfile module presence does not prove links are current.

### Results and recovery

Reports keep the original plan order, including mixed results:

| Result | Meaning |
|---|---|
| `changed` | The eligible confirmed action completed. |
| `unchanged` | No action was needed or mutation was not attempted. |
| `not supported yet` | The selected action has no supported execution path in this mode. |
| `failed` | The action failed; confirmed eligible failures produce a non-zero result. |

Execution continues according to existing behavior after a non-terminal step failure. Fix the reported cause, then rerun deliberately. This workflow performs no automatic retry or rollback.

### Advisory bootstrap boundary

When a required provider or bootstrap dependency is missing, bootstrap output is manual/advisory guidance only. This workflow does not clone, fetch, install, retry, or otherwise acquire that dependency automatically.

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
| Automatic convergence | Confirmed execution is explicit and limited; it does not provide package/version/configuration reconciliation or general idempotency. |

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
