# Delta for catalog-installer-metadata

## ADDED Requirements

### Requirement: Default catalog includes a brew-backed package target

The default catalog MUST declare `package:ripgrep` as a brew-backed package target using structured install metadata.
The system SHALL preserve its existing `command_exists` presence metadata for `rg`.
This requirement MUST NOT introduce multi-provider metadata or change unrelated default catalog resources.

#### Scenario: Ripgrep is brew-backed in the default catalog

- GIVEN the default catalog is loaded
- WHEN `package:ripgrep` is decoded
- THEN its install provider is `brew`
- AND its install package is `ripgrep`

#### Scenario: Ripgrep presence remains command-based

- GIVEN the default catalog is loaded
- WHEN `package:ripgrep` presence metadata is decoded
- THEN its presence kind is `command_exists`
- AND its presence name is `rg`

#### Scenario: Other default resources remain unchanged

- GIVEN the default catalog is loaded
- WHEN `tool:git` and `runtime:go` are decoded
- THEN `tool:git` keeps provider `apt`
- AND `runtime:go` keeps provider `asdf`

#### Scenario: No multi-provider metadata is introduced

- GIVEN the default catalog is inspected
- WHEN package install metadata is evaluated
- THEN `package:ripgrep` has exactly one provider value
- AND no fallback provider list is required
