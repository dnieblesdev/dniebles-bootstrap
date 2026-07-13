# Proposal: Homebrew Installation Channel

## Intent

Define the Linux/WSL Homebrew technical contract for `dbootstrap` without weakening the existing direct/XDG installation contract. `dbootstrap plan` must locate its installed catalog from any working directory when the formula is present.

**Status: COMPLETED technical slice.** The resolver/formula approach, Linux/WSL `amd64`/`arm64` contract, package/share catalog path, and nine-case resolver evidence (9/9 PASS) are implemented. Stable publication, lifecycle evidence, and README documentation have moved to [`publish-homebrew-stable-channel`](openspec/changes/publish-homebrew-stable-channel/proposal.md).

## Scope

### In Scope (completed)
- Define the standalone tap `dnieblesdev/homebrew-dniebles-bootstrap` formula contract for Linux/WSL `amd64` and `arm64`.
- Add a last-resort Homebrew-prefix catalog fallback after `--catalog`, XDG, and `$HOME/.local/share` resolution, preserving their precedence.
- Reject macOS before download through the formula contract.
- Produce and validate the resolver precedence with nine table-driven test cases.

### Out of Scope
- Creating a real tap repository, publishing a formula, or collecting lifecycle/Homebrew acceptance evidence.
- Stable-release selection, SHA-256 verification against a public release, and README/tap README documentation.
- Formula-update automation, Scoop, macOS support, and release/publication-pipeline changes.
- Changes to `install.sh`, XDG/direct-install ownership behavior, package-install Homebrew integration, or catalog contents.

## Capabilities

### New Capabilities
- `homebrew-installation-channel`: Linux/WSL Homebrew resolver fallback and formula/catalog contract.

### Modified Capabilities
- `direct-binary-installation`: Default catalog discovery gains a lower-priority Homebrew-prefix fallback while retaining existing direct/XDG behavior.

## Approach

Use a manually maintained formula in the standalone tap, pinned to a non-prerelease release URL and its amd64/arm64 SHA-256. Install the catalog under `pkgshare`; resolve it only when explicit, XDG, and home-local candidates do not apply. No writes occur outside the Homebrew prefix.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `cmd/dbootstrap/main.go` | Modified | CWD-agnostic Homebrew catalog fallback |
| `cmd/dbootstrap/main_test.go` | Modified | Resolution precedence and fallback coverage |
| `dnieblesdev/homebrew-dniebles-bootstrap/Formula/dbootstrap.rb` | Contract only | Formula approach and platform/sha contract defined here; physical creation and pinning moved to `publish-homebrew-stable-channel` |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| `HOMEBREW_PREFIX` unavailable | Medium | Preserve `--catalog` and existing fallbacks |
| Incorrect or changed artifact digest | Low | Verify release URL, asset names, and SHA-256 before any future publication (`publish-homebrew-stable-channel`) |

## Rollback Plan

Revert the additive resolver fallback. Existing direct/XDG installs remain untouched. Formula publication rollback is owned by `publish-homebrew-stable-channel`.

## Dependencies

- `publish-homebrew-stable-channel` owns the stable GitHub Release gate, tap creation, formula pinning, lifecycle evidence, and documentation.
- The completed resolver evidence is a dependency of the publication change.

## Success Criteria

- [x] The catalog resolver preserves `--catalog`, XDG, and `$HOME/.local/share` precedence and adds the lower-priority Homebrew-prefix fallback.
- [x] Nine table-driven unit tests verify explicit, XDG, home-local, Homebrew fallback, higher-priority wins, absent `HOMEBREW_PREFIX`, and no-existing-candidate behavior.
- [x] The formula contract defines Linux Intel/ARM branches, `pkgshare` catalog installation, and pre-download macOS rejection.
