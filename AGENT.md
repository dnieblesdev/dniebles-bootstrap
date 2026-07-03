# Repository Operating Guide

This repository uses SDD/OpenSpec planning before implementation. Agents and contributors must keep the bootstrapper domain-first and must not move dotfiles ownership into this project.

## Working rules

| Rule | Requirement |
|------|-------------|
| Specs before code | Do not implement runtime behavior before the relevant proposal, specs, design, and tasks exist. |
| English artifacts | Generated technical artifacts, docs, specs, code comments, and user-facing strings default to English. |
| Local agent state | `.atl/` is local skill-registry state and must remain ignored by git. |
| Documentation-only slices | When a change is declared docs/spec-only, do not create Go source files, catalog runtime files, installers, or CLI wiring. |
| Reviewable units | Keep specs and docs that describe one outcome together; do not split work only by file type. |

## SDD workflow

Use the active OpenSpec change folder and Engram artifacts together when the session requests both backends.

1. Read the proposal, delta specs, design, and tasks before editing.
2. Apply only the tasks assigned for the current slice.
3. Keep OpenSpec files and Engram SDD topics aligned.
4. Mark completed tasks in `tasks.md` as soon as they are done.
5. Save apply progress to Engram with `capture_prompt: false` for automated SDD artifacts.

Primary planning path for the current change:

- `openspec/changes/design-bootstrap-orchestrator/proposal.md`
- `openspec/changes/design-bootstrap-orchestrator/design.md`
- `openspec/changes/design-bootstrap-orchestrator/specs/`
- `openspec/changes/design-bootstrap-orchestrator/tasks.md`

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
