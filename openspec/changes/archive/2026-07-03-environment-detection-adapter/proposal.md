# Proposal: Environment Detection Adapter

## Intent

Replace `dbootstrap plan` static `linux/amd64` facts with a thin, testable host detection adapter while preserving `internal/planning` as a pure domain package.

## Scope

### In Scope
- Add `internal/environment` to detect OS, architecture, distro, and WSL status.
- Use injected runtime/env/file readers so tests do not depend on the host machine.
- Wire `cmd/dbootstrap plan` to pass detected facts into `planning.BuildPlan` if detection succeeds.
- Keep rendered plan output showing the facts used for planning.

### Out of Scope
- Installers, command runners, apply/install command, or dotfiles runtime.
- TUI, remote loading, user override flags, or runtime mutation.

## Capabilities

### New Capabilities
- `environment-detection`: Detects reportable OS, architecture, distro, and WSL facts before planning.

### Modified Capabilities
- `cli-plan`: Replaces static planning facts with detected facts at the CLI boundary.

## Approach

Create a small `internal/environment` adapter returning `planning.EnvironmentFacts`. Default detection should use `runtime.GOOS/GOARCH`, environment variables, and safe Linux file reads such as `/etc/os-release` and kernel/proc signals for WSL. Tests should inject fake runtime values, env lookups, and file contents. `internal/planning` must remain unchanged unless a type gap is discovered.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/environment` | New | Host fact detector and test seams. |
| `cmd/dbootstrap/main.go` | Modified | Use detected facts instead of `staticEnvironmentFacts`. |
| `cmd/dbootstrap/*_test.go` | Modified | Inject deterministic detection for CLI tests. |
| `internal/planning` | Unchanged | Continues consuming caller-supplied facts only. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Distro or WSL heuristics are incomplete | Med | Prefer conservative detection, multiple signals, and unknown/empty values over false certainty. |
| CLI tests become host-coupled | Med | Require injected providers and fixture-driven tests. |
| Adapter leaks into planning | Low | Keep detector outside planning and return existing domain facts. |

## Rollback Plan

Revert `internal/environment` and restore `cmd/dbootstrap` to caller-supplied static facts; planning and catalog behavior remain unaffected.

## Dependencies

- Existing `planning.EnvironmentFacts` contract.
- Go standard library only.

## Success Criteria

- [ ] `dbootstrap plan` uses detected facts at the CLI boundary.
- [ ] Environment detection tests cover Linux distro, WSL true/false, OS, arch, and missing-file cases with no host dependency.
- [ ] Planning package remains free of OS probes and adapter imports.
