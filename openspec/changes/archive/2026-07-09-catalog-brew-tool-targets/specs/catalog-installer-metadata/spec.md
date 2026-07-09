# Delta for catalog-installer-metadata

## ADDED Requirements

### Requirement: Default catalog includes a brew-backed tool target

The default catalog MUST declare `tool:git` as a brew-backed tool target using structured install metadata.
The system MUST preserve `install.package = "git"` and `presence.kind = "command_exists"` with `presence.name = "git"`.
The default catalog MUST preserve existing resources, bundles, and profiles, and MUST NOT introduce new resources.
The default catalog MUST preserve existing `package:ripgrep` brew-backed install metadata.
The default catalog MUST NOT require multi-provider or fallback install metadata for this resource set.

#### Scenario: tool:git is brew-backed

- GIVEN the default catalog is loaded
- WHEN `tool:git` is decoded
- THEN its install provider is `brew`
- AND its install package is `git`
- AND its presence kind is `command_exists`
- AND its presence name is `git`

#### Scenario: existing default catalog shape remains unchanged

- GIVEN the default catalog is loaded
- WHEN the catalog resources, bundles, and profiles are inspected
- THEN the existing resource set is preserved
- AND `bundle:cli` still includes `tool:git` and `package:ripgrep`
- AND `profile:dev` remains unchanged

#### Scenario: package:ripgrep remains brew-backed

- GIVEN the default catalog is loaded
- WHEN `package:ripgrep` is decoded
- THEN its install provider is `brew`
- AND its install package is `ripgrep`

#### Scenario: no multi-provider metadata is introduced

- GIVEN the default catalog is loaded
- WHEN install metadata is inspected
- THEN no fallback provider list is required
- AND no multi-provider selection metadata is introduced
