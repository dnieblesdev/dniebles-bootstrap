# Delta for catalog-installer-metadata

## MODIFIED Requirements

### Requirement: Default catalog includes brew-backed package and tool targets

The default catalog MUST declare `tool:git`, `package:ripgrep`, and `package:jq` as brew-backed targets using structured install metadata. The system MUST preserve each target's package and `command_exists` presence metadata: `git`/`git`, `ripgrep`/`rg`, and `jq`/`jq`, respectively. The default catalog MUST include `tool:git` and both packages in `bundle:cli`, so the existing `profile:dev` selection includes all three resources. The catalog MUST preserve all other existing resources, bundles, and profiles, and MUST NOT introduce new resources beyond `package:jq`. The system MUST NOT require multi-provider or fallback install metadata. This change MUST NOT alter runtime, dotfile, provider, runner, command-execution, or apply semantics.
(Previously: The default catalog required only `tool:git` and `package:ripgrep` as brew-backed targets and explicitly prohibited adding resources.)

#### Scenario: brew-backed metadata is present for all CLI targets

- GIVEN the default catalog is loaded
- WHEN `tool:git`, `package:ripgrep`, and `package:jq` are decoded
- THEN their install providers are `brew`
- AND their install packages are `git`, `ripgrep`, and `jq`
- AND their presence checks are `command_exists` for `git`, `rg`, and `jq`

#### Scenario: bundle:cli includes jq

- GIVEN the default catalog is loaded
- WHEN `bundle:cli` is inspected
- THEN it includes `tool:git`, `package:ripgrep`, and `package:jq`
- AND selecting the existing `profile:dev` includes `package:jq` through that bundle

#### Scenario: existing catalog behavior remains stable

- GIVEN the default catalog is loaded
- WHEN resources, bundles, and profiles are inspected
- THEN all pre-existing resources, bundles, and profiles remain present
- AND no new runtime, dotfile, provider, runner, or apply behavior is introduced

#### Scenario: metadata remains inert during planning

- GIVEN the default catalog contains the three brew-backed CLI targets
- WHEN a plan is built for `profile:dev`
- THEN `package:jq` is represented as a selected catalog resource
- AND metadata is preserved without executing a provider or presence check
- AND existing planning and execution semantics remain unchanged

#### Scenario: no multi-provider metadata is introduced

- GIVEN the default catalog's install metadata is inspected
- WHEN provider selection metadata is evaluated
- THEN no fallback provider list is required
- AND no multi-provider selection metadata is introduced
