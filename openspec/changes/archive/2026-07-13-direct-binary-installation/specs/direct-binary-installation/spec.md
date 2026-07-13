# Direct Binary Installation Specification

## Requirements

### Requirement: Install a supported release safely

The installer MUST support only Linux/WSL `amd64` and `arm64`, default to the latest stable GitHub Release, and accept `vX.Y.Z`. Prereleases MUST require `--allow-prerelease`. Archive and SHA-256 assets MUST come from one release, with verification before extraction or replacement.

#### Scenario: Stable supported install

- GIVEN a supported host and no requested version
- WHEN the operator runs the installer
- THEN the latest stable release is selected, verified, and staged for installation

#### Scenario: Unsupported or untrusted selection

- GIVEN an unsupported tuple, unapproved prerelease, or checksum mismatch
- WHEN installation is attempted
- THEN it fails clearly and leaves existing managed files unchanged

### Requirement: Install and validate managed payloads

The installer MUST place the binary at `${XDG_BIN_HOME:-$HOME/.local/bin}/dbootstrap` and catalog at `${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/catalog/bootstrap.toml`. It MUST validate the payload, atomically replace managed files, and verify `plan --profile dev` reads the catalog from another working directory.

#### Scenario: First install works outside the repository

- GIVEN a verified release and writable paths
- WHEN installation completes and `plan --profile dev` runs from another directory
- THEN the command reads the installed catalog successfully

#### Scenario: Existing files are protected

- GIVEN an unmanaged file exists at a managed path
- WHEN installation runs without an authorized force reinstall
- THEN it aborts without overwriting that file or leaving a partial installation

### Requirement: Control reinstall and uninstall ownership

The installer MUST require `--force` for a matching managed installation, record paths and digests in state, report PATH setup without editing startup files, and uninstall only manifest-owned files whose digests still match. `--force` MAY upgrade or downgrade.

#### Scenario: Safe uninstall

- GIVEN unmodified manifest-owned files
- WHEN the operator runs `--uninstall`
- THEN both managed files and state are removed

#### Scenario: Modified managed file is preserved

- GIVEN a manifest-owned file changed after installation
- WHEN uninstall is requested
- THEN uninstall aborts without deleting any managed file
