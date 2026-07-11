## Exploration: catalog-more-brew-targets

### Current State
The default catalog contains exactly four resources: `tool:git`, `package:ripgrep`, `runtime:go`, and `dotfile:bash`. The two installable tool/package targets are already Homebrew-backed (`git` and `ripgrep`) from the prior slices. `runtime:go` still uses `asdf` and carries `go.env`, OS, and architecture constraints; `dotfile:bash` has no install metadata and is owned by the separate dotfiles provider.

The existing confirmed apply path supports Homebrew only for `tool` and `package` resources. Runtime resources are routed to a noop installer, while dotfiles use their own provider boundary. Catalog decoding and planning preserve provider/package metadata as inert data, but changing runtime metadata to `brew` would not make it executable and would create misleading default metadata.

### Affected Areas
- `catalog/bootstrap.toml` — the only source of default catalog target metadata; no currently eligible additional tool/package target remains.
- `internal/catalog/toml/catalog_test.go` — exact default-catalog assertions already cover all four resources and their provider metadata.
- `cmd/dbootstrap/main.go` — confirms the existing provider capability boundary: brew installers are wired only for tools/packages; no change is proposed.
- `openspec/specs/catalog-installer-metadata/spec.md` — current contract already records `git` and `ripgrep` as brew-backed and preserves the four-resource shape.
- `openspec/changes/archive/2026-07-09-catalog-brew-tool-targets/` — prior slice moved `tool:git` to brew and explicitly excluded new resources/provider architecture.

### Approaches
1. **Change `runtime:go` to Homebrew** — would be a metadata-only edit, but it is not a safe coherent slice: confirmed apply does not support brew-backed runtimes, and the runtime's existing asdf/config/condition semantics would be obscured.
   - Pros: Reuses an existing catalog resource and Homebrew package naming could be defined.
   - Cons: Misleading provider metadata; changes an established asdf contract; likely changes apply bootstrap/report behavior without enabling runtime installation.
   - Effort: High risk / not recommended

2. **Change `dotfile:bash` to Homebrew** — not applicable. Dotfiles have no package-install semantics and are intentionally handled by the separate dotfiles provider.
   - Pros: None within the requested scope.
   - Cons: Violates resource-kind/provider boundaries and would require architecture changes or invalid metadata.
   - Effort: High risk / out of scope

3. **Do not create another target slice** — retain the catalog as-is because all eligible default tool/package targets have already moved to brew.
   - Pros: Smallest and safest outcome; preserves provider boundaries, runtime behavior, and exact catalog shape.
   - Cons: Does not produce the requested 1–2 additional targets; adding one would require introducing a new catalog resource, which is outside the stated metadata-only preservation scope.
   - Effort: Low

### Recommendation
Do not proceed with a production change under this name. There are no safe additional default catalog targets: `tool:git` and `package:ripgrep` are already brew-backed, while `runtime:go` and `dotfile:bash` are not eligible without expanding provider or mutation architecture. The smallest coherent slice is to stop here and, if another brew target is desired later, first approve adding a new tool/package resource plus its catalog/test contract as a separate change.

### Risks
- Changing `runtime:go` to `brew` would advertise support that `cmd/dbootstrap` does not provide and could alter Homebrew bootstrap/report output without installing Go.
- Adding a new default target would change the exact four-resource catalog shape and bundle/profile behavior, contradicting the prior slice's preservation boundary.
- Treating the dotfile as a brew target would cross the explicit external dotfiles ownership boundary.

### Ready for Proposal
No. The requested additional-target slice has no safe candidate in the current catalog. The orchestrator should tell the user that no implementation is recommended, or request approval for a broader follow-up that adds a new tool/package catalog resource.
