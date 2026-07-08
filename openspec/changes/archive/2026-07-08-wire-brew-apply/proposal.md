# Proposal: Wire Brew Apply

## Intent

Turn `dbootstrap apply --yes` from a confirmed-future noop into the first real, narrowly gated mutation path for Homebrew-backed tool/package installs, while preserving default apply and `--dry-run` as non-mutating safety paths.

## Scope

### In Scope
- Register `HomebrewInstaller` for brew-backed `tool` and `package` plan steps only in confirmed `--yes` mode.
- Use `OSCommandRunner` plus `BrewCommandExists` to execute `brew install <package>` through existing explicit command seams.
- Keep default apply and `--dry-run` on noop installers and report clearly when confirmed mode may execute real brew installs.
- If `brew` is missing, fail/report guidance through existing Homebrew bootstrap behavior; never install Homebrew.

### Out of Scope
- Installing Homebrew, remote scripts, raw command metadata, `sh -c`, pipelines, dotfiles execution, or bootstrap entrypoints.
- Non-Homebrew providers, broad provider routing, shell-first commands, retries, concurrency, clones, sparse checkout, or dotlink.
- Changing the default catalog’s package providers unless needed only as test fixture data.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `apply-command-dry-run`: `--yes` becomes the only accepted apply mode that may mutate for brew-backed tool/package installs.
- `brew-package-installer`: Installer is no longer isolated; it may be registered by the CLI composition root under confirmed mode.
- `execution-contracts`: Apply may use real execution only for the confirmed Homebrew tool/package path; noop remains required elsewhere.
- `homebrew-bootstrap-provider`: Missing `brew` continues to produce advisory/manual guidance and prevents package installation attempts.

## Approach

Branch runner composition on `applyMode`: default and `--dry-run` keep `NoopForKind`; confirmed mode registers `NewHomebrewInstaller` for `tool` and `package` with `NewOSCommandRunner()` and `brewCommandExists`, leaving runtime/dotfile/non-brew work noop or structured failure. Update rendering/mode labels to warn that confirmed mode may run real `brew install` commands.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modified | Compose confirmed vs noop runners. |
| `cmd/dbootstrap/render.go` | Modified | Clarify confirmed execution reporting. |
| `internal/execution/` | Modified | Reuse Homebrew installer/runner seams; add focused tests if needed. |
| `openspec/changes/wire-brew-apply/` | New | Proposal and follow-on SDD artifacts. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Accidental mutation outside `--yes` | Medium | Preserve noop runner for default and dry-run; test each mode. |
| Overbroad kind registration | Medium | Register only `tool` and `package`; rely on provider metadata validation. |
| Missing brew semantics regress | Low | Keep `BrewCommandExists` check before command execution and advisory bootstrap reporting. |

## Rollback Plan

Revert the runner-composition/render changes and this change folder. `HomebrewInstaller` remains available as an isolated component; apply returns to all-noop behavior.

## Dependencies

- Existing `HomebrewInstaller`, `OSCommandRunner`, `BrewCommandExists`, execution contracts, and structured install metadata.

## Success Criteria

- [ ] Default apply and `--dry-run` remain non-mutating/noop.
- [ ] `apply --yes` can execute `brew install <package>` only for brew-backed tool/package steps.
- [ ] Missing `brew` reports failure/guidance without installing Homebrew or target packages.
