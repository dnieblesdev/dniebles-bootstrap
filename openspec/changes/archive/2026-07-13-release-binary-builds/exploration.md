# Exploration: Release Binary Builds

## Current State

The repository now has CI build validation (`.github/workflows/build.yml`) running `go test`, `go vet`, and `go build` on push/PR to `main`, per the `ci-build-validation` slice. That workflow explicitly prohibits artifact generation or distribution. No binary release mechanism exists yet.

### Project Build Characteristics

| Attribute | Value |
|---|---|
| Language | Go 1.26 |
| Module | `github.com/dnieblesdev/dniebles-bootstrap` |
| Dependencies | 1 (go-toml/v2, pure Go) |
| CGO usage | **None** — zero `import "C"`, `#cgo`, or build tags |
| Binary | `cmd/dbootstrap` — single `package main` |
| Build tags | None |
| CGO_ENABLED=0 | **Fully compatible** |
| Existing ldflags | **None** — no version injection contract exists |

### Real Go Main Package

`cmd/dbootstrap/main.go` — package `main`, binary name `dbootstrap`. Single entrypoint. No sub-binaries. `main()` delegates to `run()` which dispatches on `plan`, `apply`, `bootstrap`, and `help` subcommands.

### External Runtime File Dependencies

The binary reads `catalog/bootstrap.toml` at runtime via `os.Open()` in `internal/catalog/toml/catalog.go:LoadFile()`. This file is NOT embedded via `//go:embed`. The default path is `catalog/bootstrap.toml` relative to CWD, overridable via `--catalog`.

**Implication for release archives**: The tar.gz/zip MUST include `catalog/bootstrap.toml` so a downloaded binary can be used without cloning the repository. The archive layout should be:

```
dbootstrap_linux_amd64/
├── dbootstrap
└── catalog/
    └── bootstrap.toml
```

### CGO_ENABLED=0 Compatibility

**Confirmed fully compatible.** The only dependency (`github.com/pelletier/go-toml/v2`) is a pure Go TOML parser. No C imports, no CGO directives, no platform-specific build tags, no `_cgo_*` files anywhere in the codebase. Setting `CGO_ENABLED=0` produces a statically linked binary with no libc dependency.

### ldflags / Version Injection Contract

**Does not exist.** There is no version variable, no `internal/version` package, no `-ldflags` usage, and no version string displayed by any subcommand. This must be established from scratch as part of this change:

- Add `internal/version/version.go` with a `var Version = "dev"` overridable via `-ldflags="-X github.com/dnieblesdev/dniebles-bootstrap/internal/version.Version=$VERSION"`
- Wire `--version` into the CLI entrypoint (`cmd/dbootstrap/main.go`)
- The build workflow will inject `git describe --tags --always --dirty` or the workflow dispatch input

### Existing Active Changes

None that conflict. The `ci-release-delivery` exploration was superseded and split into `ci-build-validation` (completed), `release-binary-builds` (this), `github-release-publishing` (future), and `release-installation-channel` (future). The `ci-build-validation` change is archived.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/version/version.go` | **New** | Version variable for ldflags injection |
| `cmd/dbootstrap/main.go` | **Modified** | Add `--version` flag handling |
| `.github/workflows/release-build.yml` | **New** | workflow_dispatch binary build workflow |
| `openspec/specs/release-binary-builds/` | **New** | Delta spec for binary build behavior |

## Approaches

### 1. Native GitHub Actions Workflow (`go build` matrix)

A single `workflow_dispatch` workflow with a strategy matrix (`os` × `arch`) that calls `go build` with GOOS/GOARCH, archives per platform (tar.gz / zip), generates SHA-256 checksums, and uploads via `actions/upload-artifact`.

- **Pros**:
  - No external tool dependencies — uses built-in Go toolchain + GitHub Actions primitives
  - Fully transparent — every build step is explicit in the YAML
  - ~70-90 lines of YAML total, well within review budget
  - Easy to customize archive layout (include `catalog/bootstrap.toml`)
  - Version injection via `-ldflags` is explicit and user-controllable via workflow input
  - No `goreleaser` config file to maintain or version-pin
  - Works without any tag/release infrastructure
- **Cons**:
  - Manual archive/checksum logic (tar, zip, sha256sum) — more boilerplate
  - No built-in changelog, release notes, or Homebrew tap (all excluded from scope anyway)
  - Slightly more YAML to review
- **Effort**: Low-Medium

### 2. GoReleaser (goreleaser-action with snapshot mode)

A workflow that invokes `goreleaser/goreleaser-action` with `--snapshot` to produce archives and checksums without requiring a git tag. A `.goreleaser.yml` config defines the build matrix, archive format, and checksum generation.

- **Pros**:
  - Declarative build config — `.goreleaser.yml` is ~30-40 lines of known-good convention
  - Built-in archive, checksum, and naming conventions
  - Mature project with broad Go ecosystem adoption
  - Easy to graduate from snapshot to full release later (`github-release-publishing` slice)
- **Cons**:
  - Adds `goreleaser/goreleaser-action` as an external GitHub Actions dependency (pinned to a version)
  - `.goreleaser.yml` is another config file to maintain
  - GoReleaser expects a tag-triggered release model; `--snapshot` mode is a workaround for workflow_dispatch
  - `--snapshot` builds are versioned as `0.0.0-SNAPSHOT-<commit>` — version injection via ldflags requires extra config
  - Over-engineered for this scope — the native tooling is simpler and more transparent
  - GoReleaser's default archive layout puts the binary at the root without `catalog/bootstrap.toml` — requires `extra_files` or `archives.files` config
- **Effort**: Medium

## Recommendation

**Approach 1 — Native GitHub Actions Workflow**.

### Why

1. **Scope alignment**: The user explicitly excluded tags, releases, write permissions, changelogs, and install channels. GoReleaser's primary value proposition is release automation — using it purely as a build engine for manual dispatch is like renting a moving truck to go get groceries. The native approach does exactly what's needed with zero excess.

2. **Transparency and control**: A native workflow makes every step visible — GOOS, GOARCH, CGO_ENABLED, ldflags, archive command, checksum generation. There's no magic. When the `github-release-publishing` slice comes later and introduces GoReleaser, the pipeline will be easier to understand because the foundation is explicit.

3. **Simpler dependency chain**: No `goreleaser-action` to pin, no `.goreleaser.yml` to drift from GoReleaser versions. One `actions/setup-go@v5` and the Go toolchain is sufficient.

4. **Review budget**: ~80 lines of YAML vs ~70 lines (`.goreleaser.yml` + workflow invoking it). Comparable line count, but the native approach has zero hidden complexity.

5. **Archive layout control**: Including `catalog/bootstrap.toml` in the archive root is a single `cp` in the native approach. In GoReleaser, it requires `archives.files` or `extra_files` config.

### Version Injection Strategy

```go
// internal/version/version.go
package version

// Version is set at build time via ldflags:
//   go build -ldflags="-X github.com/dnieblesdev/dniebles-bootstrap/internal/version.Version=$(git describe --tags --always --dirty)"
var Version = "dev"
```

The workflow will accept an optional `version` input (defaulting to `git describe --tags --always --dirty`).

### Archive Structure

```
dbootstrap_linux_amd64.tar.gz
└── dbootstrap_linux_amd64/
    ├── dbootstrap
    ├── catalog/
    │   └── bootstrap.toml
    └── sha256sums.txt

dbootstrap_windows_amd64.zip
└── dbootstrap_windows_amd64/
    ├── dbootstrap.exe
    ├── catalog/
    │   └── bootstrap.toml
    └── sha256sums.txt
```

## Risks

- **`catalog/bootstrap.toml` drift**: The catalog file evolves with the repo. A downloaded binary with an outdated catalog may produce unexpected plans. Mitigation: include the catalog in the archive and document that users should update the catalog alongside the binary.
- **No Windows CI testing**: The build workflow will cross-compile for `windows/amd64` on `ubuntu-latest`. The binary compiles but won't be executed on Windows in CI. This is acceptable given Windows is not a primary target (Homebrew/Apt backends are Linux/macOS only), but the zip should still be built and checksummed.
- **linux/arm64 cross-compilation**: Cross-compiling for arm64 on ubuntu-latest (amd64 runner) is standard and well-supported by Go — no emulation needed.
- **Version injection fragility**: If `git describe` fails (shallow clone, no tags), the version falls back to `dev`. The workflow should fetch full history (`fetch-depth: 0`) to ensure `git describe` works. The first tag must exist or the fallback to commit hash must work gracefully.

## Ready for Proposal

**Yes.** The exploration confirms:
- Single `cmd/dbootstrap` main package with no CGO and no embedded files
- `CGO_ENABLED=0` is fully compatible
- No existing ldflags/version contract — must be established
- `catalog/bootstrap.toml` must be bundled in release archives
- Native GitHub Actions workflow (`go build` + matrix + archive + checksum) is the right fit for this scope
- Recommend `sdd-propose` for `release-binary-builds` with Approach 1 (native workflow)
