# Proposal: APT Provider

## Intent

Enable the smallest safe APT-backed install path for confirmed Linux execution. Today provider metadata can express `apt`, but only Homebrew-backed tool/package steps can execute.

## Scope

### In Scope
- Add an APT installer/provider composition parallel to Homebrew for supported `tool` and `package` kinds.
- On Linux, wire it only for `apply --yes` when injected `apt-get` presence confirms availability; allow sudo only through explicit `apply --yes --sudo`.
- Validate trimmed package metadata (non-empty and not beginning with `-`), then send exactly `apt-get install -y -- <package>` or `sudo apt-get install -y -- <package>` through `CommandRunner` with a ten-minute timeout.
- Prove the path with an opt-in fixture/custom catalog; preserve the default catalog unchanged.

### Out of Scope
- Shells, automatic `sudo`/`pkexec` escalation, APT bootstrap/update/repositories, retries, fallback selection, or provider-registry redesign.
- Package-presence detection, rollback/transaction orchestration, and default-catalog migration.

## Capabilities

### New Capabilities
- `apt-package-installer`: Provider-gated direct APT installation through the existing command seam.

### Modified Capabilities
- `apply-command-dry-run`: Permit eligible APT execution only in confirmed Linux `--yes` mode; keep default and dry-run non-mutating.
- `execution-contracts`: Extend confirmed execution composition without changing kind-based Runner dispatch or noop safety.

## Approach

Add an `AptInstaller` matching the Homebrew installer pattern: accept only `Install.Provider == "apt"` and a trimmed, non-empty package not starting with `-`. Require injected `apt-get` availability, then issue `CommandRequest{Executable: "apt-get", Args: ["install", "-y", "--", package], Timeout: 10 * time.Minute}` for direct confirmed mode, or `CommandRequest{Executable: "sudo", Args: ["apt-get", "install", "-y", "--", package], Timeout: 10 * time.Minute}` only for explicit `--yes --sudo`. The delimiter prevents option injection from custom catalog metadata; it is not shell escaping. Use provider-aware wrappers so APT and brew metadata cannot reach the wrong delegate, while retaining Homebrew's existing cross-platform execution behavior. The CLI composition root receives detected facts and wires APT only for Linux; non-Linux selected APT is failed without probes or commands. Ten minutes bounds package-manager lock waits through the existing request seam; timeout is a structured failure with no retry or rollback claim. Fixture/custom catalog tests supply an explicit APT target rather than altering defaults.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/execution/` | Modified/New | APT installer and provider gate beside Homebrew |
| `cmd/dbootstrap/main.go` | Modified | Linux confirmed-mode composition and injectable APT seam |
| `cmd/dbootstrap/*_test.go` | Modified | Confirmed/default/dry-run composition contracts |
| `internal/execution/*_test.go` | New/Modified | Explicit command and failure contracts |
| `testdata/` or custom catalog fixture | New | Opt-in APT proof target |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| APT may partially mutate before failure | Med | Report failure; explicitly do not claim rollback |
| Process lacks privileges or lock wait times out | High | Explicit direct/sudo mode only; ten-minute bound; no fallback, retry, rollback, or bootstrap |
| Existing package is reattempted | Med | Defer presence detection; document limitation |

## Rollback Plan

Revert the APT installer, confirmed composition, and opt-in fixture. The default catalog and default/dry-run behavior remain unchanged.

## Dependencies

- Linux facts and injected `apt-get` presence seam already available to CLI composition.

## Success Criteria

- [ ] Linux `apply --yes` requests only `apt-get install -y -- <package>`; explicit `apply --yes --sudo` requests only `sudo apt-get install -y -- <package>`, through `CommandRunner`.
- [ ] Default and `--dry-run` neither probe nor execute APT.
- [ ] Non-Linux APT is `StepStatusFailed` with a non-zero confirmed result and zero apt/sudo probes or commands; invalid metadata, missing commands, and ten-minute timeouts are structured failures without escalation, retries, or rollback claims.
