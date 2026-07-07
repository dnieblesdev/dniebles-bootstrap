# Proposal: Catalog Installer Metadata

## Intent

Add structured install and presence metadata to catalog resources so future real installers can choose providers and detect existing tools without hardcoding package decisions or normalizing shell snippets.

## Scope

### In Scope
- Extend catalog metadata for tools, runtimes, and packages with structured install provider/package data.
- Add structured presence/check metadata for existing-resource detection.
- Map metadata into format-agnostic planning/resource types without executing it.
- Update the sample catalog to demonstrate safe metadata.

### Out of Scope
- Command runner, raw shell commands, `curl | sh`, retries, concurrency, or command execution.
- Real installers, real `apply` mutation, dotfiles execution, or bootstrap entrypoint wiring.
- Provider-specific install behavior beyond representing metadata.

## Capabilities

### New Capabilities
- `catalog-installer-metadata`: Catalog resources expose safe, structured install and presence metadata for downstream installer/provider selection.

### Modified Capabilities
- None

## Approach

Introduce nested structured metadata such as `[install] provider = "brew", package = "ripgrep"` and `[presence] kind = "path", name = "rg"` or `[check] kind = "command_exists", name = "rg"`. Decode and validate these fields in the TOML adapter, then carry them through the planning domain as inert data. Do not add a default `command = "..."` model; any raw-command escape hatch remains deferred and must be explicitly gated later.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `catalog/bootstrap.toml` | Modified | Add representative safe metadata examples. |
| `internal/catalog/toml/schema.go` | Modified | Decode nested install/presence metadata. |
| `internal/catalog/toml/catalog.go` | Modified | Map TOML metadata into planning resources. |
| `internal/catalog/toml/validate.go` | Modified | Validate provider/check shapes without shell execution. |
| `internal/planning/types.go` | Modified | Add inert metadata fields to `Resource`. |
| `internal/planning/builder.go` | Modified | Preserve metadata in `PlanStep.Resource`; no planning semantics change. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Metadata becomes shell-first | Med | Require provider/check structs and defer raw commands. |
| Planner semantics accidentally change | Low | Keep metadata inert and verify existing plan behavior. |
| Provider names overfit one platform | Med | Treat provider/package as data, not execution policy. |

## Rollback Plan

Remove the metadata structs, TOML fields, validation, sample catalog entries, and tests. Existing catalog fields and dry-run `apply` remain unchanged because no execution behavior is introduced.

## Dependencies

- Existing catalog decoding, planning resource model, and noop execution boundary.

## Success Criteria

- [ ] Catalog metadata is decoded, validated, and preserved in plan resources.
- [ ] Sample metadata uses structured provider/presence fields only.
- [ ] No command execution, real installer, mutation, or bootstrap entrypoint behavior is added.
