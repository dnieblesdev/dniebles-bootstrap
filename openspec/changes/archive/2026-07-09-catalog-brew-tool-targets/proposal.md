# Change Proposal: catalog-brew-tool-targets

## Intent

Make the default catalog include one Homebrew-backed tool target by changing `tool:git` install metadata from `apt` to `brew` while preserving its package name, presence check, dependencies, bundles, and profiles.

This keeps the catalog aligned with the current provider direction without adding execution behavior or expanding the provider model.

## Scope

In scope:

- Update `catalog/bootstrap.toml` so `tool:git` uses `install.provider = "brew"`.
- Preserve `install.package = "git"`.
- Preserve `presence.kind = "command_exists"` and `presence.name = "git"`.
- Preserve existing catalog resources, bundles, profiles, dependency relationships, OS/arch metadata, and descriptions.
- Add or adjust focused tests/spec coverage that verifies the default catalog preserves the brew-backed `tool:git` metadata.

Out of scope:

- Provider execution behavior or wiring.
- Apt provider implementation or fallback behavior.
- Dotfiles execution.
- New catalog resources, bundles, profiles, or dependency changes.
- Multi-provider, platform-specific provider selection, or fallback provider models.

## Affected Areas

- `catalog/bootstrap.toml`
- Focused catalog metadata tests/specs, likely under the existing `catalog-installer-metadata` capability.
- OpenSpec delta for `catalog-installer-metadata` if needed to document that default catalog fixture metadata may include Homebrew-backed tool targets.

## Product/Behavior Expectations

- Loading the default catalog exposes `tool:git` with install provider `brew` and package `git`.
- Planning and catalog decoding continue to treat install and presence metadata as inert data.
- Existing bundle/profile behavior remains unchanged: `bundle:cli` still includes `tool:git` and `package:ripgrep`; `profile:dev` remains unchanged.
- The change does not make `dbootstrap apply` install Git via Homebrew; it only changes catalog metadata.

## Risks

- Users may infer execution support from the `brew` provider value even though provider execution remains out of scope for this slice.
- Existing tests or docs may assume `tool:git` is apt-backed and need focused updates.
- The catalog will describe `git` as brew-backed even on Linux, but this slice intentionally avoids platform-specific provider selection or fallback modeling.

## Rollback

Revert the `tool:git` provider in `catalog/bootstrap.toml` from `brew` back to `apt` and revert any focused tests/spec updates introduced for this metadata change.

## Success Criteria

- The default catalog declares `tool:git` with `install.provider = "brew"` and `install.package = "git"`.
- The `tool:git` presence check remains `command_exists: git`.
- No resources, bundles, profiles, dependencies, or provider execution behavior are added or changed.
- Focused tests/specs confirm the intended catalog metadata and continue to pass.
