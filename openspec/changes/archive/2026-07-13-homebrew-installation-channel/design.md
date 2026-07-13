# Design: Homebrew Installation Channel

## Technical Approach

Make the CLI discover its packaged Homebrew catalog only after the current explicit, XDG, and home-local locations, and define the standalone tap formula contract for Linux/WSL `amd64`/`arm64`. This implements the resolver delta spec and the formula/catalog contract without changing the release pipeline, direct installer, or creating a real tap repository.

**Status: COMPLETED technical slice.** Phase 1 (resolver) is implemented and verified with 9/9 table-driven cases. Stable publication, lifecycle evidence, README updates, and physical tap creation moved to `publish-homebrew-stable-channel`.

## Architecture Decisions

| Decision | Options / tradeoff | Decision and rationale |
|---|---|---|
| Catalog resolver | Read `HOMEBREW_PREFIX` at call sites; inject environment/home/filesystem seams | Extend the existing `catalogPathResolver` with `PathExists func(string) bool`; retain nil defaults to `os.LookupEnv`, `os.UserHomeDir`, and `os.Stat`. Its last candidate is `<HOMEBREW_PREFIX>/share/dbootstrap/catalog/bootstrap.toml`, preserving the release archive's `catalog/` directory and deterministic precedence. |
| Package layout | Put catalog beside `bin`; use formula package share | Formula uses `bin.install "dbootstrap"` and `pkgshare.install "catalog/bootstrap.toml"`. `pkgshare` is `<HOMEBREW_PREFIX>/share/dbootstrap`, so Homebrew preserves the archive path at `<prefix>/share/dbootstrap/catalog/bootstrap.toml`. It is conventional, version-link-safe, and stays in the prefix without staging or renaming. |
| Formula platforms | One universal URL; platform/architecture branches | Use `on_linux` with explicit `Hardware::CPU.intel?` and `Hardware::CPU.arm?` URL/SHA pairs; unsupported Linux CPU calls `odie`. `on_macos { odie "dbootstrap supports Linux/WSL only; macOS assets are unavailable" }` contains no URL, failing before download. |
| Release source | Latest release lookup; pinned stable metadata | A maintainer selects a named release only after `gh release view <tag> --json isDraft,isPrerelease,tagName,assets` proves stable and asset names/digests match. The committed formula has literal version, URL, and SHA-256 values; it never resolves "latest." This gate and the physical formula commit are owned by `publish-homebrew-stable-channel`. |
| Delivery boundary | Separate tap change; alter publishing workflow | The tap repository and formula are created/published by `publish-homebrew-stable-channel`. This change only defines the contract and resolver behavior. Do not change release/publish workflows: existing releases already emit the required Linux archives and `.sha256` files; availability is an external gate, not a contract blocker. |

## Data Flow

```text
Qualified GitHub stable release ── pinned URL + SHA ──> tap Formula (owned by publish-homebrew-stable-channel)
                                                    └──> brew install
                                                             ├── bin/dbootstrap
                                                             └── share/dbootstrap/catalog/bootstrap.toml

CLI: --catalog → XDG candidate → ~/.local candidate → HOMEBREW_PREFIX/share/dbootstrap/catalog/bootstrap.toml
                                       (first existing candidate wins)
```

`Resolve` builds these candidates using injected seams. It checks existence in precedence order; if none exist, it returns the highest-priority configured conventional candidate so current missing-catalog diagnostics remain meaningful. An unavailable home directory or `HOMEBREW_PREFIX` simply omits that candidate. `parsePlanFlags` and `parseApplyLikeFlags` keep calling `defaultCatalogPath`, so explicit `--catalog` remains untouched.

## File Changes

| File | Action | Description |
|---|---|---|
| `cmd/dbootstrap/main.go` | Modified | Add existence seam and lower-priority `<HOMEBREW_PREFIX>/share/dbootstrap/catalog/bootstrap.toml` candidate. |
| `cmd/dbootstrap/main_test.go` | Modified | Table-driven resolver precedence and missing-environment tests. |

## Interfaces / Contracts

```go
type catalogPathResolver struct {
    LookupEnv  func(string) (string, bool)
    HomeDir    func() (string, error)
    PathExists func(string) bool
}
```

Formula contract: archives are named `dbootstrap_<safe-version>_linux_{amd64,arm64}.tar.gz`; each selected branch supplies its own `sha256`. Formula installation owns only `bin/dbootstrap` and `pkgshare/catalog/bootstrap.toml`; Homebrew owns lifecycle cleanup. Physical creation and pinning of `dnieblesdev/homebrew-dniebles-bootstrap/Formula/dbootstrap.rb` are owned by `publish-homebrew-stable-channel`.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Explicit/XDG/home/Homebrew precedence, the exact package-share candidate, absent prefix, no existing candidates | Table-driven `catalogPathResolver` tests; use `t.TempDir()` paths and injected `LookupEnv`, `HomeDir`, and `PathExists`, asserting `<prefix>/share/dbootstrap/catalog/bootstrap.toml`. |

Run focused Go tests, then `go test ./...`. Lifecycle/Homebrew acceptance tests and evidence recording are owned by `publish-homebrew-stable-channel`.

## Migration / Rollout

No migration required. The resolver fallback is safe to merge independently. Roll back by reverting the resolver fallback; direct/XDG installations remain unchanged. Formula publication rollback is owned by `publish-homebrew-stable-channel`.

## Traceability

- Completed technical slice: `homebrew-installation-channel` (this change).
- Stable publication boundary: `publish-homebrew-stable-channel`.
- Moved pending tasks: old tasks 2.1–4.2 now live under `publish-homebrew-stable-channel/tasks.md`.

## Open Questions

None.
