# Design: Catalog Brew Provider Targets

## Technical Approach

Make a catalog-only slice. `catalog/bootstrap.toml` already carries structured install metadata into `planning.Resource.Install`; `apply --yes` already routes brew-backed `tool`/`package` steps through provider-gated brew installers. This change only gives the default catalog one safe brew-backed package target.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|----------|--------|-------------------------|-----------|
| Target resource | Change only `package:ripgrep` to `brew` | Change `tool:git`, `runtime:go`, or add a new package | Ripgrep is common, presence is already `rg`, and package install semantics are simpler than tool/runtime changes. |
| Provider model | Keep a single `provider` string | Add multi-provider/fallback support | Existing execution is provider-gated; fallback behavior is a larger product decision. |
| Execution wiring | No changes | Modify runner/installers/apply composition | `wire-brew-apply` already supports confirmed brew-backed package execution. |

## Data Flow

```text
catalog/bootstrap.toml
  -> internal/catalog/toml decode
  -> planning.Resource.Install{Provider:"brew", Package:"ripgrep"}
  -> planning.PlanStep for package:ripgrep
  -> apply --yes provider-aware package installer
  -> HomebrewInstaller builds brew install ripgrep
```

Default apply and `--dry-run` still use noop/non-mutating execution paths.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `catalog/bootstrap.toml` | Modify | Change `packages.install.provider` for ripgrep from `apt` to `brew`; keep package and presence unchanged. |
| `internal/catalog/toml/catalog_test.go` | Modify | Add or update default catalog coverage for ripgrep brew install metadata and unchanged `rg` presence. |

## Interfaces / Contracts

No new interfaces. Existing `InstallMetadata{Provider, Package}` contract is sufficient.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|--------------|----------|
| Catalog/unit | Default catalog decodes ripgrep as brew-backed | Load `catalog/bootstrap.toml` and assert install metadata. |
| Catalog/unit | Ripgrep presence stays `command_exists: rg` | Extend the same catalog assertion. |
| Regression | Git and Go providers stay unchanged | Assert `tool:git=apt` and `runtime:go=asdf`. |

## Migration / Rollout

No migration required. This is a default catalog metadata change.

## Open Questions

- None
