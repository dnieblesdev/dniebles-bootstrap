# Proposal: Design Bootstrap Orchestrator

## Intent

Design `dniebles-bootstrap` as a domain-first Go development-environment orchestrator before implementation. It must support fresh-machine bootstrap and later point installs while integrating with `dniebles-dotfiles` without owning dotfiles internals.

## Scope

### In Scope
- OpenSpec proposal/design/spec artifacts for the initial architecture.
- `README.md` orientation update and `AGENT.md` operating guide.
- Domain model framing: Profile, Catalog, Plan, Runner, Installer, EnvironmentDetector, DotfilesProvider.
- First-run entrypoint framing: a tiny wrapper may acquire `dbootstrap`, then hand control to the Go application/core.
- Catalog decision: catalog lives in this repo; TOML recommended first, with a format-agnostic domain model.

### Out of Scope
- Go application code, CLI implementation, installers, and execution runtime.
- Shell-based orchestration beyond first-run acquisition of `dbootstrap`.
- Bubble Tea TUI implementation; architecture may reserve it as a future interface.
- Dotfiles module internals, profiles, configs, symlinks, assets, validations, and `dotlink` ownership.

## Capabilities

### New Capabilities
- `bootstrap-orchestration`: Plans and coordinates profile and point installs across resources.
- `environment-detection`: Detects OS, distro, WSL, and architecture before plan resolution.
- `bootstrap-entrypoint`: Makes `dbootstrap` available on first startup via released binary or Go compile/run fallback.
- `catalog-planning`: Defines in-repo catalog-backed profiles, bundles, tools, runtimes, and packages.
- `dotfiles-integration`: Treats `~/.dotfiles` as an external provider, requests modules, supports sparse checkout/partial clone, and invokes `dotlink`.
- `repository-guidance`: Documents project scope, boundaries, SDD workflow, and future contributor expectations.

### Modified Capabilities
None.

## Approach

Choose a domain-first orchestrator core. CLI and future TUI should become thin interfaces over shared planning/execution concepts. If expected dotfiles config is missing, the plan should still install the tool but clearly report the missing configuration as requiring attention.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `openspec/changes/design-bootstrap-orchestrator/` | New | Proposal and follow-up SDD artifacts. |
| `README.md` | Modified | Purpose, goals/non-goals, flows, and dotfiles boundary. |
| `AGENT.md` | New | Repository operating guide for agents/contributors. |
| `catalog/` or equivalent | Future | In-repo TOML catalog location to be specified in design. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Dotfiles boundary drift | Med | Keep bootstrap orchestration-only and provider-driven. |
| Catalog schema lock-in | Med | Use TOML initially but keep domain model format-agnostic. |
| CLI/TUI divergence | Low | Centralize domain logic outside interfaces. |

## Rollback Plan

Revert documentation and OpenSpec artifacts for this change. No runtime migration is required because this proposal intentionally ships no Go code.

## Dependencies

- `dniebles-dotfiles` at `https://github.com/dnieblesdev/dotfiles`, local path `~/.dotfiles`.

## Success Criteria

- [ ] Proposal defines goals, non-goals, boundaries, and chosen domain-first direction.
- [ ] Follow-up specs/design can proceed without implementation code.
- [ ] README/AGENT scope is clear for the repository owner first and extensible for later contributors.
