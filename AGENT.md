# Repository Operating Guide

This repository keeps bootstrap orchestration domain-first and must not move dotfiles ownership into this project. This guide only documents repo-local policies; general SDD/OpenSpec workflow comes from the active agent prompts and OpenSpec artifacts.

## Working rules

| Rule | Requirement |
|------|-------------|
| Local agent state | `.atl/` is local skill-registry state and must remain ignored by git. |
| Repo policy only | Keep this file limited to local boundaries, architecture, safety rules, and mutation constraints. Do not duplicate generic SDD/OpenSpec instructions here. |
| Slice scope | Respect the active change scope. If a slice is docs/spec-only, do not create Go source files, catalog runtime files, installers, or CLI wiring. |
| Apply safety | Do not make `dbootstrap apply` mutate the host unless the active change explicitly defines the safety contract and opt-in behavior. |

## Local SDD persistence

When automated SDD artifacts are saved to Engram, use `capture_prompt: false` for generated phase artifacts such as apply progress, verification reports, and archive reports.

## Dotfiles boundary

`dniebles-bootstrap` owns bootstrap orchestration. `~/.dotfiles` owns dotfiles internals.

Bootstrap may:

- request dotfiles modules from an external provider;
- use partial clone and sparse checkout strategies for requested scopes;
- invoke `dotlink` as a provider operation;
- report missing dotfiles configuration as attention-required.

Bootstrap must not:

- define dotfiles module internals;
- own symlink lifecycle or asset layout;
- duplicate dotfiles validations or configuration semantics;
- move declarative dotfiles profile ownership into this repository.

## First-run wrapper boundary

A Bash first-run wrapper may exist only to make `dbootstrap` available and hand control to it.

Allowed wrapper responsibilities:

- download/install a compatible released `dbootstrap` binary;
- install or use Go to compile/run from this repository;
- launch `dbootstrap`.

Forbidden wrapper responsibilities:

- catalog resolution;
- dotfiles integration;
- installer selection;
- dependency ordering;
- plan execution;
- operational reporting.

## Architecture guardrail

Future code should preserve one core with thin interfaces:

- Domain/core owns profiles, resources, catalog concepts, plans, plan steps, validation, and structured statuses.
- Application/use cases own profile planning, point planning, execution orchestration, and result aggregation.
- Infrastructure owns adapters such as TOML catalog loading, installers, command execution, git sparse checkout, dotfiles provider calls, and first-run acquisition.
- CLI is the first interface; a future Bubble Tea TUI should be a thin presenter/controller over the same use cases.

If a future change starts duplicating planning rules in CLI or TUI code, stop and redesign before implementation.
