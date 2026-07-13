## Exploration: direct-binary-installation

### Current State

`dniebles-bootstrap` has a mature Go orchestrator, release infrastructure, and safety model — but no mechanism for a user to acquire `dbootstrap` itself without a package manager. The first-run bootstrap gap is well-understood and explicitly documented in README.md and AGENT.md.

#### What already exists (ready to reuse)

**Release infrastructure** (`.github/workflows/release-build.yml`, `release-publish.yml`):
- Builds three platform archives: `dbootstrap_{version}_linux_amd64.tar.gz`, `dbootstrap_{version}_linux_arm64.tar.gz`, `dbootstrap_{version}_windows_amd64.zip`.
- Each archive contains the `dbootstrap` binary + `catalog/bootstrap.toml`.
- SHA-256 checksums are generated per archive and verified in the publish workflow.
- GitHub Releases are created with v-prefixed SemVer tags; prereleases supported; existing tag/release overwrite prevented.
- Artifact naming is stable and predictable from version + platform tuple.

**Platform detection** (`internal/environment/detector.go`):
- `Detector` with injectable seams (`RuntimeSource`, `EnvSource`, `FileSource`).
- Returns `planning.EnvironmentFacts{OS, Arch, Distro, WSL}`.
- Detects Linux distro from `/etc/os-release`, WSL from `WSL_DISTRO_NAME`/`WSL_INTEROP` env vars and `/proc/version` kernel text.
- Tests are fully host-independent with fake providers.

**Execution model** (`internal/execution/command.go`, `command_runner_test.go`):
- `CommandRunner` interface with `CommandRequest{Executable, Args, Dir, Env, Timeout}`.
- Explicit executable-plus-args model; no shell string or shell wrapper.
- `ValidateCommandRequest` rejects shell-first input.
- `CommandResult` captures exit code, stdout, stderr, duration, and error.

**Version** (`internal/version/version.go`):
- `Version` variable set via `-ldflags -X`; defaults to `"dev"`.
- `dbootstrap --version` reports it.
- CLI validates both dispatch and release tags through `internal/version/cmd/validate`.

**Safety model** (entire codebase):
- Default and `--dry-run` are non-mutating.
- Only explicit `--yes` confirms mutation; only `--yes --sudo` enables sudo.
- Noop installers return `not_implemented` for unsupported work.
- Reports preserve plan order and structured statuses.

**Install state detection** (`internal/state/`):
- `CommandExists` function type detects tool/runtime presence via `exec.LookPath`.
- APT package presence detects via injected `dpkg-query`.
- Brew formula presence detects via `brew list --formula`.
- All detection is read-only and injectable.

**AGENT.md boundary**: Explicitly allows a "Bash first-run wrapper that may exist only to make `dbootstrap` available and hand control to it." Forbids catalog resolution, dotfiles integration, installer selection, dependency ordering, plan execution, and operational reporting — all of which remain owned by the Go application.

#### What does NOT exist (must be built)

| Missing capability | Where it would live |
|---|---|
| HTTP download of release assets | New `internal/acquisition/` adapter or similar |
| SHA-256 checksum verification in Go | New or within acquisition adapter |
| Archive extraction (tar.gz, zip) | New or within acquisition adapter |
| Binary target directory detection | New infrastructure |
| Binary placement with privilege awareness | New infrastructure |
| File-tracking for uninstall | New infrastructure |
| Install/uninstall documentation | `README.md` or dedicated doc |
| First-run delivery mechanism (shim) | Shell script or CLI subcommand |

### Affected Areas

- `internal/environment/detector.go` — Platform detection reused as-is; no change needed. The detector provides OS/arch/WSL in the exact shape needed to select a release asset URL.
- `.github/workflows/release-build.yml` — Already produces the correct archives and checksums. May need to confirm archive layout (e.g., binary at root vs. subdirectory).
- `.github/workflows/release-publish.yml` — Already publishes verified releases. No change needed for the install path to consume them.
- New `internal/acquisition/` (or similar name) — New infrastructure adapter for download/verify/extract with injectable HTTP, filesystem, and checksum seams. This layer is what future `dbootstrap install` and the Homebrew formula can both consume.
- New `install.sh` (at repository root) — Minimal curated shell script that selects the right release URL, downloads via curl/wget, verifies sha256sum, extracts, and places the binary. This is the first-run delivery mechanism.
- `README.md` — Install/uninstall documentation becomes the primary user-facing artifact for this change. Must document: platform URL pattern, verification steps, target directory conventions, privilege expectations, uninstall procedure.
- `AGENT.md` — First-run wrapper boundary already defined. May add install-script-specific constraints if needed.
- `catalog/bootstrap.toml` — No change. Catalog bundled in archives already works; `catalogtoml.LoadFile` validates on first run.
- `internal/version/version.go` — No change. Version already injectable and reportable.

### Approaches

#### 1. **Shell install script + Go acquisition infrastructure (Recommended)**

Build a tested Go `acquisition` package with injectable HTTP, checksum, and filesystem seams. Deliver the first-run path as a minimal curated `install.sh` that uses standard tools (curl/wget, sha256sum, tar) and follows the same URL pattern and verification contract that the Go infrastructure defines.

- **Pros**: Shell is the universal bootstrap mechanism — zero prerequisites beyond what ships with Linux/WSL. AGENT.md already allows this. Go infrastructure is reusable by future `dbootstrap install` subcommand and Homebrew formula. Platform detection stays in Go (script just maps `uname -sm` to archive URL). Install/uninstall docs directly reference the script.
- **Cons**: Shell scripts are harder to test comprehensively than Go. Slight duplication of platform selection logic (shell `case` mirrors Go build matrix). sha256sum vs. Go checksum implementation for verification phase.
- **Effort**: Medium

#### 2. **dbootstrap self-install subcommand**

Add `dbootstrap install` and `dbootstrap uninstall` commands. The binary knows how to curl/wget its own release, verify checksums, extract, and place itself. The first install is bootstrapped by a trivial one-liner: `curl -sSfL <url> | tar xz && ./dbootstrap install`.

- **Pros**: All logic in tested Go code. Self-contained. Uninstall is clean (`dbootstrap uninstall` removes tracked files). No shell script to maintain beyond the trivial one-liner.
- **Cons**: Chicken-and-egg for the initial download — still needs a curl pipe. The binary carries HTTP+archive code used once per machine — acceptable but worth noting. The install command must work when running from a temp directory (not yet installed), which adds state complexity.
- **Effort**: Medium-High

#### 3. **Documentation-only + manual steps**

Document the release URL pattern, manual download/verify/extract steps. No automation code — the user does it by hand.

- **Pros**: Zero code to build or maintain. Fastest to deliver. No shell vs. Go debate.
- **Cons**: Poor UX for a tool whose entire purpose is automation. Manual sha256 verification is error-prone. Doesn't move the needle on the first-run experience. Doesn't enable future Homebrew channel.
- **Effort**: Low

#### 4. **Separate `install.sh` only (no Go infrastructure)**

A curated shell script that does everything. No Go code changes.

- **Pros**: Fastest path to a working first-run experience. AGENT.md explicitly allows it. Trivial to serve from GitHub Releases.
- **Cons**: No reuse for future channels. Harder to test. Duplicates platform logic. The Go side of the house gains nothing.
- **Effort**: Low-Medium

### Recommendation

**Approach 1: Shell install script + Go acquisition infrastructure.** This is the right foundation because:

1. The shell script solves the chicken-and-egg problem instantly — users run one auditable command and get `dbootstrap` available.
2. The Go infrastructure (`internal/acquisition/`) provides tested, injectable download/verify/extract primitives that future slices (`dbootstrap install` subcommand, Homebrew formula verification step) can consume.
3. It respects AGENT.md's boundary: the shim is minimal, hands control to `dbootstrap`, and never touches catalog/planning/execution.
4. It cleanly separates `direct-binary-installation` from `homebrew-installation-channel` — the direct binary path is the foundation; Homebrew is a convenience layer on top.

The shell script should:
- Detect `$(uname -sm)` → map to `linux_amd64`, `linux_arm64`, or error
- Determine latest release tag via GitHub API or a stable `latest` URL
- Download `dbootstrap_{version}_{os}_{arch}.tar.gz` and its `.sha256`
- Verify `sha256sum --check`
- Extract to a staging location
- Place `dbootstrap` in the first writable directory from `$HOME/.local/bin`, `$XDG_BIN_HOME`, or error
- Validate the placed binary runs (`./dbootstrap --version`)
- Report success/failure with clear next steps

### How This Enables Homebrew Bootstrap

The direct binary installation channel is the **zero-dependency foundation**. Once `dbootstrap` is available on the host:

1. `dbootstrap bootstrap --profile dev --yes` runs the full orchestration pipeline
2. If Homebrew-backed resources are requested but `brew` is missing, the existing `AppendHomebrewBootstrap` produces advisory guidance with the official Homebrew docs URL
3. The user installs Homebrew manually (as currently designed), then re-runs
4. Future work: a `homebrew-bootstrap-provider` could automate Homebrew acquisition within `--yes` confirmed execution

The `homebrew-installation-channel` change (separate slice) would add `brew install dbootstrap` / `brew tap` support — a **convenience channel** for users who already have Homebrew. It does NOT replace direct binary installation; it's an alternative entrypoint for an ecosystem-specific audience.

This clean separation means:
- A user on a fresh Linux/WSL machine can install `dbootstrap` without any package manager
- A macOS user who already has Homebrew can `brew install dbootstrap` as a natural extension
- Both channels converge at the same `dbootstrap` binary and orchestration experience
- Neither channel requires the other

### Risks

- **Shell script security**: The install script must never pipe curl into sh. It must download to a temp file, verify checksum, then extract. Must validate the binary runs before suggesting PATH changes.
- **Platform detection divergence**: Shell `uname -sm` and Go `runtime.GOOS`/`runtime.GOARCH` must agree on mapping. The release-build matrix (`linux amd64`, `linux arm64`, `windows amd64`) is the source of truth. Script must map explicitly, not generically.
- **Target directory privileges**: `~/.local/bin` may not exist; `/usr/local/bin` requires sudo. The script must prefer user-writable paths and document sudo expectations explicitly.
- **Windows support**: The current release-build produces Windows archives, but the install script should focus on Linux/WSL first. Windows install is a separate concern with different conventions (PATH, zip extraction, admin privileges).
- **Catalog placement**: Archives bundle `catalog/bootstrap.toml` at `catalog/` relative to the binary. The install script must preserve this structure or document where the catalog lives. Currently `dbootstrap` uses a `--catalog` flag with a default of `catalog/bootstrap.toml` — placement matters.
- **GitHub API rate limits**: Querying the latest release via unauthenticated API may hit rate limits. Use the `latest` redirect URL (`/repos/:owner/:repo/releases/latest`) or document rate-limit fallback (manual version selection).
- **Overwrite safety**: Re-running the install script must not corrupt an existing installation. Must detect existing binary, compare versions, and require explicit `--force` to overwrite.
- **Scope creep into Homebrew**: This change must stay strictly focused on direct binary acquisition. The Homebrew bootstrap provider and `homebrew-installation-channel` are separate slices and must stay separate.

### Ready for Proposal

**Yes.** The proposal should state that `direct-binary-installation` delivers:

1. A tested, injectable Go acquisition adapter (`internal/acquisition/`) for download, checksum verification, and archive extraction.
2. A minimal, auditable `install.sh` script that uses the release asset naming convention to install `dbootstrap` on Linux/WSL amd64/arm64.
3. Install/uninstall documentation in `README.md`.
4. Explicit exclusion of Homebrew, Scoop, Windows install, `dbootstrap install` subcommand, and TUI work.

The change enables the zero-dependency bootstrap path that AGENT.md already describes, and lays the foundation for both the Homebrew installation channel and a future `dbootstrap self-install` subcommand.
