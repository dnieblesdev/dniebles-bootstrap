# Delta for Direct Binary Installation

## MODIFIED Requirements

### Requirement: Install and validate managed payloads

The installer MUST place the binary at `${XDG_BIN_HOME:-$HOME/.local/bin}/dbootstrap` and catalog at `${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/catalog/bootstrap.toml`. It MUST validate the payload, atomically replace managed files, and verify `plan --profile dev` reads the catalog from another working directory. Discovery MUST preserve this precedence: explicit `--catalog`, XDG data, `$HOME/.local/share`, then the Homebrew-prefix package-share fallback; the fallback MUST be CWD-independent and skipped when `HOMEBREW_PREFIX` is unavailable.
(Previously: Discovery lacked a lower-priority Homebrew-prefix fallback.)

#### Scenario: First install works outside the repository

- GIVEN a verified release and writable paths
- WHEN installation completes and `plan --profile dev` runs from another directory
- THEN the command reads the installed catalog successfully

#### Scenario: Existing files are protected

- GIVEN an unmanaged file exists at a managed path
- WHEN installation runs without an authorized force reinstall
- THEN it aborts without overwriting that file or leaving a partial installation

#### Scenario: Homebrew catalog is the last-resort default

- GIVEN no explicit, XDG, or `$HOME/.local/share` catalog exists and `HOMEBREW_PREFIX` identifies an installed formula catalog
- WHEN `dbootstrap plan --profile dev` runs from an arbitrary working directory without `--catalog`
- THEN it loads the Homebrew catalog

#### Scenario: Higher-priority catalog wins

- GIVEN explicit, XDG, home-local, and Homebrew-prefix catalogs are available
- WHEN `dbootstrap plan --profile dev` runs
- THEN it loads the first candidate in that precedence order
