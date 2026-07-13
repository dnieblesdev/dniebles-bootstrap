# Design: Publish Homebrew Stable Channel

## Technical Approach

This change is a publication boundary, not a second implementation of `homebrew-installation-channel`. It first qualifies a named GitHub Release; only a passing gate may resolve formula metadata, publish the tap formula, and add availability documentation. It reuses Phase 1's completed nine-case resolver evidence and its installed catalog contract: `<HOMEBREW_PREFIX>/share/dbootstrap/catalog/bootstrap.toml`.

**Current status: OPEN/BLOCKED.** `v0.0.0-rc.1` is public and has both Linux archives/checksums, but is a prerelease; it is validation-only and cannot seed stable publication.

## Architecture Decisions

| Decision | Options / tradeoff | Decision and rationale |
|---|---|---|
| Release gate precedes all publication | Start tap work from the prerelease; qualify immutable stable metadata first | Query a named release with `gh release view <tag> --json tagName,isDraft,isPrerelease,publishedAt,assets,url`. Require public, non-draft, non-prerelease status plus exact Linux `amd64`/`arm64` archives and matching `.sha256` assets. Download each archive and its `.sha256` file, then verify the checksum file content matches the archive bytes. This prevents a prerelease, placeholder, or mismatched digest from becoming public stable state. |
| Formula metadata is post-gate | “latest” lookup/placeholders; literal verified values | Only after the gate passes, create `dnieblesdev/homebrew-dniebles-bootstrap/Formula/dbootstrap.rb` from scratch with literal version, archive URL/name, and SHA-256 for each architecture. Commit only these pinned values; no `latest`, prerelease, placeholder path, or unresolved field. |
| Reuse the completed technical contract | Rework resolver/archive layout; consume Phase 1 | Keep the existing resolver and archive layout unchanged. Formula installation uses `bin.install "dbootstrap"` and `pkgshare.install "catalog/bootstrap.toml"`; it matches the already-tested fallback path and preserves direct/XDG precedence. |
| Explicit platform boundary | Implicit unsupported behavior; explicit formula guards | The formula has Linux Intel and ARM branches with independent URL/SHA pairs. Its macOS branch calls `odie` before defining a URL/download. This creates auditable pre-download rejection rather than accidental macOS support. |

## Data Flow

```text
named GitHub Release ──> metadata + asset gate ──fail──> OPEN/BLOCKED
         │ pass
         v
archive + .sha256 pairs ──> pinned tap formula ──> brew lifecycle evidence
                                                     ├─ Linux/WSL amd64
                                                     ├─ Linux/WSL arm64
                                                     └─ macOS rejects pre-download
                                                           │
README + tap README <── final evidence and scope review ┘
```

## File Changes

| File | Action | Description |
|---|---|---|
| `openspec/changes/publish-homebrew-stable-channel/design.md` | Create | This gated publication design. |
| `README.md` | Modify after gate | Add summary Linux/WSL tap install/use instructions, supported architectures, catalog path, stable prerequisite, and macOS exclusion only. |
| `dnieblesdev/homebrew-dniebles-bootstrap/Formula/dbootstrap.rb` | Create after gate | Pinned Linux-only formula and pre-download macOS rejection; created from scratch with literal stable metadata. |
| `dnieblesdev/homebrew-dniebles-bootstrap/README.md` | Create after gate | Tap commands, ownership, release metadata, archive hashes, operational proof, and detailed publication evidence record. |
| `.github/workflows/release-publish.yml` | No change | Existing workflow already emits and verifies the required Linux archives/checksums. |

## Interfaces / Contracts

```text
release qualification = !isDraft && !isPrerelease && publishedAt != null
  && assets include {dbootstrap_<version>_linux_amd64.tar.gz, .sha256,
                     dbootstrap_<version>_linux_arm64.tar.gz, .sha256}
  && each .sha256 file content equals sha256(download(archive))
```

Formula ownership is limited to `bin/dbootstrap` and `share/dbootstrap/catalog/bootstrap.toml`. Release, source, workflow, resolver, archive-layout, and formula automation changes are out of scope.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Release gate | Metadata, exact assets, checksum-to-archive matching | Save command output and checksum verification before any formula metadata is resolved. |
| Linux/WSL acceptance | Tap, install, version, strict audit/style, reinstall or controlled upgrade, uninstall, catalog path, cleanup, unrelated-file preservation | Run separately on `amd64` and `arm64`; retain per-command evidence. |
| Platform boundary | No macOS download | Capture formula failure output and network/download absence for an attempted macOS install. |
| Final verification | Scope, docs, release status, all evidence | Review against proposal/spec; run focused `cmd/dbootstrap` tests then `go test ./...` as regression evidence. |

## Migration / Rollout

No migration required. Do not publish, create a release, create a formula, or change a workflow while the gate fails. Once evidence qualifies, publish the pinned tap and documentation together; rollback removes/reverts the tap publication and docs without touching the completed resolver.

## Open Questions

- [ ] Which named stable release will satisfy the gate? **Blocking:** no qualifying release currently exists.
