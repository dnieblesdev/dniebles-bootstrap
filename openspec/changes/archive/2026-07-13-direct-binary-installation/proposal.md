# Proposal: Direct Binary Installation

## Intent

Provide a zero-package-manager first install for `dbootstrap` on Linux and WSL, so a fresh host can enter the existing bootstrap workflow before Homebrew exists.

## Scope

### In Scope
- An auditable Linux/WSL `install.sh` supporting only amd64 and arm64, with explicit platform/architecture detection and failure for every other tuple.
- Download a selected GitHub Release archive and matching SHA-256 file to staging; verify before extraction or replacement.
- Install `dbootstrap` at `$XDG_BIN_HOME/dbootstrap` or `~/.local/bin/dbootstrap`; install the bundled catalog at `$XDG_DATA_HOME/dbootstrap/catalog/bootstrap.toml` or `~/.local/share/dbootstrap/catalog/bootstrap.toml`.
- Verify the installed binary reads the installed catalog in a real command path; document install, PATH setup, reinstall protection, and uninstall of both managed paths.

### Out of Scope
- Brew, Scoop, Windows, macOS, package-manager dependencies, signing, or a `dbootstrap install` command.
- `homebrew-installation-channel`; it remains a later convenience channel for hosts that already have Homebrew.

## Capabilities

### New Capabilities
- `direct-binary-installation`: Secure, package-manager-free GitHub Release installation and removal for supported Linux/WSL hosts.

### Modified Capabilities
- `operational-readme`: Document the direct-install and uninstall lifecycle and catalog location.

## Approach

Add a minimal shell first-run wrapper plus injectable Go acquisition primitives for download, SHA-256 verification, and archive extraction. Reuse the established release asset naming and host facts; keep the wrapper outside catalog resolution, planning, and execution. Stage and validate all content before atomically replacing managed files.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `install.sh` | New | Detect, download, verify, stage, and install. |
| `internal/acquisition/` | New | Testable release acquisition primitives. |
| `cmd/`, `internal/` | Modified | Resolve the installed catalog and prove a real read. |
| `README.md` | Modified | Install, PATH, and uninstall guidance. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Tampered or partial asset | Low | SHA-256 check before extraction/replacement. |
| Unsupported host or unwritable path | Med | Explicit tuple/path errors; never fall back to sudo. |
| Existing install overwritten | Med | Require explicit force and stage atomically. |

## Rollback Plan

Remove the installed binary and catalog from the documented managed paths; revert the script and acquisition code. No release assets or package-manager state are changed.

## Dependencies

- Existing GitHub Releases, archive layout, and SHA-256 assets.
- Standard Linux/WSL download, checksum, and tar utilities.

## Success Criteria

- [ ] Linux and WSL amd64/arm64 install without Brew or another package manager; unsupported tuples fail clearly.
- [ ] A checksum mismatch leaves existing managed files untouched.
- [ ] The installed binary successfully reads the installed catalog; documented uninstall removes both managed paths.
