# Design: Direct Binary Installation

## Technical Approach

Add a repository-owned, auditable Bash `install.sh` for first installs on Linux/WSL `amd64` and `arm64`. It resolves one GitHub Release, obtains that release's archive and checksum asset, verifies SHA-256 before extraction, stages both payload files, then replaces managed files safely. The CLI will resolve its default catalog from the XDG data location rather than the working directory; `--catalog` remains an explicit override.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Version source | Default to GitHub's latest **stable** release; accept `--version vX.Y.Z`; require `--allow-prerelease` for any prerelease tag. | Git-derived version; prereleases by default. | A first-install path must be predictable; prereleases require deliberate consent. |
| Release integrity | Resolve archive and `.sha256` from one REST release document; require exactly the expected Linux asset pair. Verify SHA-256 before tar extraction. | Construct independent `/latest/download` URLs; verify after extraction. | The API object binds both assets to one release and validation precedes every managed-path mutation. |
| Installer boundary | Keep acquisition in `install.sh`; do not add an unused Go acquisition package in this slice. | Go acquisition primitives. | A fresh host cannot execute Go code before the binary exists; duplicating security logic would weaken the boundary. |
| Catalog resolution | Default to `${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/catalog/bootstrap.toml`; `--catalog` overrides it. | Repository-relative default; infer from CWD. | Installed behavior is CWD-independent and uses the approved user path. |
| Existing files | A state manifest marks both managed paths. Any unmanaged/incompatible existing file aborts. A matching managed install requires `--force`; force permits reinstall, upgrade, or downgrade. | Silent overwrite; version-only checks. | Ownership, not filename, protects user files while retaining explicit lifecycle control. |
| Interface and docs | Curated local `install.sh` is primary; README also provides reviewed manual download/verify/install commands. Docs never make `curl | sh` the only path. | Remote pipe-to-shell only; manual-only workflow. | The script is repeatable, while manual commands remain inspectable and usable in restricted environments. |
| PATH | Report the required PATH export when needed; do not edit shell rc files. | Mutate `.bashrc`/`.zshrc`. | Shell startup ownership remains with the user. |

## Data Flow

    install.sh --version/--allow-prerelease
        -> GitHub Release API (one release object)
        -> archive + matching checksum -> SHA-256 verification
        -> temporary extraction and payload validation
        -> atomic per-file replace + state manifest
        -> installed `dbootstrap plan --profile dev` default-catalog read

The script creates a private temporary directory, validates the archive contains only the expected `dbootstrap` and `catalog/bootstrap.toml` payload paths, and preserves backups before replacement. It atomically renames each staged file and manifest; any detected replacement or real-read failure restores all prior managed files and removes newly created paths. Thus expected failures leave no partial installation; interrupted-process recovery is handled by retained transaction backups on the next invocation.

## File Changes

| File | Action | Description |
|---|---|---|
| `install.sh` | Create | Bash installer/uninstaller, release selection, verification, staged transaction, PATH report. |
| `cmd/dbootstrap/main.go` | Modify | Replace the CWD-relative default with an XDG/home catalog resolver while preserving `--catalog`. |
| `cmd/dbootstrap/main_test.go` | Modify | Cover resolver precedence and a real installed-path `plan` catalog read. |
| `install_test.sh` | Create | Fixture-driven shell tests for tuple/version/asset/checksum/rollback/managed-file cases. |
| `README.md` | Modify | Local-script and manual installation, PATH reporting, `--force`, and managed-only uninstall guidance. |

## Interfaces / Contracts

```text
./install.sh [--version vX.Y.Z] [--allow-prerelease] [--force] [--uninstall]

binary:  ${XDG_BIN_HOME:-$HOME/.local/bin}/dbootstrap
catalog: ${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/catalog/bootstrap.toml
state:   ${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/install-state.toml
```

`install-state.toml` records the release tag and exact managed binary/catalog paths and digests. Uninstall validates that manifest and current digests before removing only those files; a mismatch aborts without deletion. The installer uses injectable command/HTTP/path variables solely for shell tests, never as undocumented user options.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | XDG catalog resolver and override precedence | Go table tests, including non-repository CWD. |
| Integration | Installed binary reads installed catalog | Invoke `plan --profile dev` from a different CWD. |
| Shell | Stable/prerelease selection, asset pairing, checksum-before-extract, rollback, force, unmanaged refusal, safe uninstall | Local HTTP/archive fixtures and temp homes; no live release calls. |

## Migration / Rollout

No migration required. Existing repository invocations must pass `--catalog catalog/bootstrap.toml` when they need the checked-out catalog; README will make this explicit. Existing unmarked installations are intentionally not adopted or removed automatically.

## Open Questions

None.
