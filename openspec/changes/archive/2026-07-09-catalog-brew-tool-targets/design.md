# Design: catalog-brew-tool-targets

## Overview

This micro-slice is a catalog metadata-only change. The default catalog will describe `tool:git` as Homebrew-backed by changing only `catalog/bootstrap.toml` from `install.provider = "apt"` to `install.provider = "brew"` for `tool:git`.

The change must preserve the default catalog shape exactly: the same resource ids, bundle/profile definitions, dependencies, OS/arch metadata, descriptions, presence metadata, and package names. No provider execution behavior, fallback provider model, platform selector, resource shape, bundle/profile definition, or mutation path changes are part of this design.

## Current State

- `catalog/bootstrap.toml` currently defines:
  - `tool:git` with `provider = "apt"`, `package = "git"`, description `Version control`, presence `command_exists/git`, OS `linux,darwin`, no dependencies, and no arch metadata.
  - `runtime:go` with `provider = "asdf"`, `package = "golang"`, description `Go toolchain`, dependency `tool:git`, config key `go.env`, presence `command_exists/go`, OS `linux,darwin`, and arch `amd64,arm64`.
  - `package:ripgrep` with `provider = "brew"`, package `ripgrep`, description `Fast text search`, dependency `tool:git`, presence `command_exists/rg`, and no OS/arch metadata.
  - `dotfile:bash` with description `Bash dotfiles`, no dependencies, no install metadata, no presence metadata, and no OS/arch metadata.
  - `bundle:cli` containing `tool:git` and `package:ripgrep` in that order.
  - `profile:dev` containing bundle `cli` and resource `runtime:go` in that order.
- `internal/catalog/toml` decodes install and presence metadata as inert planning-domain data. It does not special-case provider values beyond requiring non-empty `provider` and `package`.
- `cmd/dbootstrap` execution output already reacts to brew-backed metadata for manual Homebrew bootstrap guidance and confirmed-mode brew-only routing. This slice does not change that behavior; default-catalog tests that assert exact apply output may need expectation updates only where the catalog metadata change affects existing behavior.

## Decisions

1. Make a single catalog edit: set `tool:git` install provider to `brew`.
2. Preserve `install.package = "git"` and `presence.kind/name = "command_exists"/"git"` exactly.
3. Preserve the decoded default catalog shape exactly except for the one provider value.
4. Do not introduce multi-provider metadata, fallback provider lists, platform-specific provider selection, or new schema fields.
5. Keep decoder behavior generic: no new validation restricting or interpreting provider values.
6. Add focused default-fixture coverage that loads `catalog/bootstrap.toml` and asserts the exact decoded catalog shape, not only the plan subset.
7. Update CLI exact-output tests only where existing behavior changes because `tool:git` is now brew-backed; do not change execution routing or mutation behavior.

## Data Flow

1. `catalog/bootstrap.toml` is loaded by `internal/catalog/toml.LoadFile`.
2. `Decode` maps TOML entries into `planning.Catalog` resources, bundles, and profiles.
3. `planning.BuildPlan` carries `planning.InstallMetadata{Provider: "brew", Package: "git"}` for `tool:git` through plan steps as inert metadata.
4. Plan rendering remains unchanged because current plan output does not print install provider/package metadata.
5. Apply rendering may change for default-catalog exact-output tests because existing apply code already treats brew-backed planned resources as requiring Homebrew bootstrap guidance or brew-only confirmed-mode handling.

## File Changes

### `catalog/bootstrap.toml`

Change only this field under `[[tools]] id = "git"`:

```toml
[tools.install]
provider = "brew"
package = "git"
```

Everything else in the file must remain unchanged.

### `internal/catalog/toml/catalog_test.go`

Add or refactor focused fixture coverage so it loads `../../../catalog/bootstrap.toml` and asserts the exact decoded default catalog shape before planning. This test is mandatory for this change and must fail if any extra resource is introduced or any existing default metadata changes unexpectedly.

The fixture assertion must compare the loaded catalog to an explicit expected `planning.Catalog` containing exactly:

- Resource refs, with no additions or omissions:
  - `tool:git`
  - `runtime:go`
  - `package:ripgrep`
  - `dotfile:bash`
- `tool:git`:
  - description `Version control`
  - `DependsOn: []`
  - `Install: &planning.InstallMetadata{Provider: "brew", Package: "git"}`
  - `Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "git"}`
  - `Conditions.OS: []string{"linux", "darwin"}`
  - no arch metadata
- `runtime:go`:
  - description `Go toolchain`
  - `DependsOn: []planning.ResourceRef{toolGit}`
  - `ConfigPolicy.RequiredKeys: []string{"go.env"}`
  - `Install: &planning.InstallMetadata{Provider: "asdf", Package: "golang"}`
  - `Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "go"}`
  - `Conditions.OS: []string{"linux", "darwin"}`
  - `Conditions.Arch: []string{"amd64", "arm64"}`
- `package:ripgrep`:
  - description `Fast text search`
  - `DependsOn: []planning.ResourceRef{toolGit}`
  - `Install: &planning.InstallMetadata{Provider: "brew", Package: "ripgrep"}`
  - `Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "rg"}`
  - unchanged absence of OS/arch metadata
- `dotfile:bash`:
  - description `Bash dotfiles`
  - `DependsOn: []`
  - unchanged absence of install, presence, OS, and arch metadata
- Bundles, exactly:
  - `bundle:cli` / map key `"cli"` with resources `[]planning.ResourceRef{toolGit, packageRipgrep}` in that order
- Profiles, exactly:
  - `profile:dev` / map key `"dev"` with bundles `[]string{"cli"}` and resources `[]planning.ResourceRef{runtimeGo}` in that order

This exact-catalog fixture assertion may be a new test such as `TestLoadDefaultFixturePreservesExactCatalogShape` or may be folded into `TestLoadFileAndBuildPlanFromFixture` before the plan assertions. The important requirement is a full decoded catalog equality check, not only spot checks of planned steps.

Keep or update the existing plan fixture assertions:

- `tool:git` expected install metadata becomes `&planning.InstallMetadata{Provider: "brew", Package: "git"}`.
- Keep presence expectation `&planning.PresenceMetadata{Kind: "command_exists", Name: "git"}`.
- Keep expected plan steps `tool:git`, `package:ripgrep`, `runtime:go`.
- Keep status expectations unchanged.

Do not change parser-only tests unless required by failing assertions. In particular, `TestDecodePreservesMetadata` may continue proving arbitrary provider strings are preserved, because decoder semantics are not changing.

### `cmd/dbootstrap/main_test.go`

Update exact-output expectations only where default-catalog apply behavior changes as a direct consequence of `tool:git` becoming brew-backed:

- Confirmed apply for default profile with missing Homebrew:
  - `tool:git` becomes `[unchanged] skipped because Homebrew must be installed manually before brew-backed resources can be applied`.
  - summary changes from `unchanged: 1` / `not supported yet: 2` to `unchanged: 2` / `not supported yet: 1`.
  - `package:ripgrep` remains the same skipped/unchanged brew-backed step.
  - `runtime:go` remains not supported.
- Default non-mutating apply for `--resource tool:git`:
  - step remains `[not supported yet] noop installer does not perform real installation`.
  - manual actions change from `- none` to the existing Homebrew bootstrap guidance because the selected resource is now brew-backed and the test stubs missing Homebrew.

Plan command exact output should not need updates because provider metadata is not rendered in plan output.

## Contracts

- Default metadata contract: `tool:git.install.provider == "brew"`, `tool:git.install.package == "git"`, `tool:git.presence == command_exists/git`.
- Exact shape preservation contract: loading `catalog/bootstrap.toml` must produce exactly the four existing resources (`tool:git`, `runtime:go`, `package:ripgrep`, `dotfile:bash`), exactly `bundle:cli`, and exactly `profile:dev`; no new resources, bundles, or profiles are introduced.
- Metadata preservation contract: descriptions, presence metadata, dependencies, OS metadata, arch metadata, package names, config requirements, and absence/presence of optional metadata remain unchanged except for `tool:git.install.provider`.
- Execution contract: existing apply behavior may observe the changed metadata, but implementation must not add provider behavior or modify execution routing.
- Schema contract: no new TOML fields or planning-domain fields.

## RED/GREEN Test Path

1. Before implementation, add the exact default-catalog fixture assertion that loads `catalog/bootstrap.toml`; it should fail while `tool:git` remains apt-backed if expecting the new brew provider.
2. Apply the single catalog metadata edit and focused expectation updates.
3. Re-run focused tests:
   - `go test ./internal/catalog/toml`
   - `go test ./cmd/dbootstrap`
4. Run the required full suite:
   - `go test ./...`

## Rollout and Rollback

- Rollout is a normal code review containing only the catalog edit and focused test expectation updates.
- Rollback is to change `tool:git` provider back to `apt` and revert associated test expectation updates.

## Risks

- Exact CLI apply tests can fail because current execution already treats any brew-backed resource as requiring Homebrew bootstrap guidance.
- Users may infer broader Homebrew provider support from the catalog metadata. The change intentionally remains metadata-only and does not alter provider capabilities.
- The default catalog will describe `git` as brew-backed for both Linux and Darwin because platform-specific provider selection is out of scope.

## Reliability Review Correction

This revision explicitly fixes the prior design gap by requiring focused fixture coverage that loads `catalog/bootstrap.toml` and asserts exact default catalog shape, including resource refs, bundles/profiles, `bundle:cli` members, `profile:dev` members, dependencies, OS/arch metadata, descriptions, presence metadata, package metadata, and the absence of additional resources.
