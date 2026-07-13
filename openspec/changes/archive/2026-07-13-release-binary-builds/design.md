# Design: Release Binary Builds

## Technical Approach

Add one manually dispatched, Ubuntu-hosted GitHub Actions workflow. It resolves a version, cross-compiles static binaries, stages a self-contained per-target layout, archives it, generates SHA-256 files, and uploads all final files only after every target succeeds. The existing validation workflow remains unchanged and produces no artifacts.

## Architecture Decisions

### Decision: Native workflow and shell packaging

| Option | Tradeoff | Decision |
|---|---|---|
| GoReleaser or custom tooling | More features but new configuration and dependency | Reject |
| GitHub Actions matrix plus standard `tar`, `zip`, and `sha256sum` | Small, auditable, runner-dependent scripts | Choose |

**Rationale**: Three fixed targets need no release framework. An Ubuntu runner can cross-compile with `CGO_ENABLED=0` and create both archive formats.

### Decision: Version package with linker injection

| Option | Tradeoff | Decision |
|---|---|---|
| Read version from Git metadata at runtime | Requires repository metadata in distributed binaries | Reject |
| `internal/version.Version` initialized to `dev`, overridden with `-X` | Explicit build-time contract | Choose |

**Rationale**: `cmd/dbootstrap/main.go` remains a thin CLI adapter: its top-level `--version` branch prints `version.Version` and exits successfully, while ordinary local builds preserve `dev`. The workflow accepts optional `workflow_dispatch.inputs.version`; when empty it derives an immutable `git describe --always --dirty`-free commit/tag value from full checkout history before passing `-ldflags "-X github.com/dnieblesdev/dniebles-bootstrap/internal/version.Version=$VERSION"`.

## Data Flow

    workflow_dispatch(version?)
             │
             ▼
    checkout (contents: read) ──→ resolve VERSION
             │                         │
             ▼                         ▼
    target matrix ──→ go build ──→ staging/<name>/
                                    ├── dbootstrap[.exe]
                                    └── catalog/bootstrap.toml
                                             │
                                             ▼
                              tar.gz / zip ──→ SHA-256 files
                                             │
                                             ▼
                              single upload-artifact bundle

## File Changes

| File | Action | Description |
|---|---|---|
| `.github/workflows/release-build.yml` | Create | Manual, least-privilege release-build workflow. |
| `internal/version/version.go` | Create | Linker-overridable `Version = "dev"`. |
| `cmd/dbootstrap/main.go` | Modify | Recognize top-level `--version`. |
| `cmd/dbootstrap/main_test.go` | Modify | Assert version output and success exit code. |

## Interfaces / Contracts

```go
// internal/version/version.go
package version

var Version = "dev"
```

Dispatch exposes an optional string `version`. The matrix targets `linux/amd64`, `linux/arm64`, and `windows/amd64`; Linux uses `dbootstrap_<version>_linux_<arch>.tar.gz`, Windows uses `dbootstrap_<version>_windows_amd64.zip`. Each archive root contains `dbootstrap` (or `dbootstrap.exe`) and `catalog/bootstrap.toml`. Its adjacent `<archive>.sha256` contains the archive hash and filename. A final `dbootstrap-artifacts-<version>` workflow artifact contains exactly the three archives and three checksum files.

The workflow declares `permissions: { contents: read }`; it requests neither write nor identity-token permissions. `actions/checkout@v4`, `actions/setup-go@v5`, and `actions/upload-artifact@v4` are pinned to current major versions. A separate final upload job `needs` every matrix build, so any failed target prevents a complete bundle upload.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Default version CLI behavior | Add table case calling `run([]string{"--version"}, ...)`; assert `dev\n`, empty stderr, success. |
| Integration | Injected value | Build locally with the documented `-ldflags -X` path, run `--version`, assert the injected value. |
| Workflow | Targets, layout, checksums, permissions | Review YAML; manually dispatch and inspect/download the artifact bundle. |

## Migration / Rollout

No migration required. The workflow has no automatic triggers and no publishing side effects.

## Open Questions

None.
