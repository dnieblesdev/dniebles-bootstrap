# Exploration: Homebrew Installation Channel

## Current State

`dniebles-bootstrap` ships `dbootstrap` through two channels today:

1. **Direct binary installation** (`install.sh`): Downloads a versioned GitHub Release archive, verifies SHA-256, and atomically installs the binary and `bootstrap.toml` catalog into XDG paths (`$HOME/.local/bin/dbootstrap`, `$HOME/.local/share/dbootstrap/catalog/bootstrap.toml`). State is tracked in `install-state.toml` for safe reinstall/upgrade/downgrade/uninstall.

2. **Go source** (`go run ./cmd/dbootstrap`): Runs from the repository root using the local `catalog/bootstrap.toml`.

**No Homebrew formula or tap exists.** The GitHub Release pipeline (`release-publish.yml`) produces and publishes these assets per release:
- `dbootstrap_{safe_version}_linux_amd64.tar.gz` + `.sha256`
- `dbootstrap_{safe_version}_linux_arm64.tar.gz` + `.sha256`
- `dbootstrap_{safe_version}_windows_amd64.zip` + `.sha256`

Each archive contains the static binary and `catalog/bootstrap.toml`. The only published release is `v0.0.0-rc.1` (a **pre-release**). No stable release exists yet.

### Existing Homebrew-aware code (adjacent but separate)

These components already exist and are NOT the target of this change:

| Component | Role | Relevant? |
|---|---|---|
| `HomebrewInstaller` (`internal/execution/homebrew_installer.go`) | Installs third-party packages via `brew install <pkg>` | No — installs packages, not `dbootstrap` itself |
| `BrewFormulaDetector` (`internal/state/brew_formula_detector.go`) | Probes `brew list --formula` for presence | No — detects installed formulas |
| `AppendHomebrewBootstrap` (`internal/execution/homebrew_bootstrap.go`) | Reports advisory "install Homebrew" guidance | No — targets missing `brew`, not `dbootstrap` |
| `homebrew-bootstrap-provider` spec | Defines detection/reporting of missing Homebrew | No — provider bootstrap, not tool bootstrap |

### Catalog resolution (CWD-agnostic)

`cmd/dbootstrap/main.go` resolves the default catalog path via `catalogPathResolver`:

```go
// 1. $XDG_DATA_HOME/dbootstrap/catalog/bootstrap.toml, or
// 2. $HOME/.local/share/dbootstrap/catalog/bootstrap.toml
```

A `--catalog <path>` CLI flag overrides both. This resolution is already CWD-agnostic — the catalog is found from environment/home, never from working directory.

## Affected Areas

This change introduces a **new standalone artifact** — a Homebrew tap repository with a formula — plus a potential catalog-resolution enhancement in the existing Go binary.

### New artifacts (must create)

- **Homebrew tap repository**: `dnieblesdev/homebrew-dniebles-bootstrap` on GitHub, containing `Formula/dbootstrap.rb`
- **`Formula/dbootstrap.rb`**: The Homebrew formula that defines install, uninstall, and platform constraints

### Existing code that MAY need minor changes

- **`cmd/dbootstrap/main.go`** — `catalogPathResolver`: May need a third fallback for Homebrew-installed catalog paths (e.g., reading `HOMEBREW_PREFIX` at runtime) so `dbootstrap plan --profile dev` works from any CWD without `--catalog`
- **`catalog/bootstrap.toml`**: No changes; packaged as-is
- **`openspec/specs/direct-binary-installation/spec.md`**: May need a note that Homebrew is an alternative channel (informational, not contractual)
- **`openspec/specs/homebrew-bootstrap-provider/spec.md`**: No change — this spec governs `dbootstrap` detecting missing Homebrew for package installs, not installing `dbootstrap` itself

### Files explicitly NOT affected
- `.github/workflows/*` — no publishing-pipeline changes (scope constraint)
- `install.sh` / `install_test.sh` — direct binary channel remains unchanged
- `internal/execution/*` — brew package installer untouched
- `internal/planning/*` — planning core untouched

## Approaches

### 1. Pure Formula (No Go Code Changes)

The formula installs everything inside the Homebrew prefix. The catalog goes to `#{pkgshare}/catalog/bootstrap.toml`. Users must use `--catalog` or set `XDG_DATA_HOME` to point to the Homebrew-installed catalog. The formula's `caveats` block prints instructions.

```ruby
# Formula/dbootstrap.rb
class Dbootstrap < Formula
  desc "..."
  homepage "https://github.com/dnieblesdev/dniebles-bootstrap"
  url "https://github.com/dnieblesdev/dniebles-bootstrap/releases/download/v0.1.0/dbootstrap_v0.1.0_linux_amd64.tar.gz"
  sha256 "..."

  on_linux do
    if Hardware::CPU.arm?
      url "...arm64.tar.gz"
      sha256 "..."
    end
  end

  on_macos do
    odie "dbootstrap does not yet provide macOS binaries. Use Linux or WSL."
  end

  def install
    bin.install "dbootstrap"
    pkgshare.install "catalog"
  end

  def caveats
    <<~EOS
      The catalog is installed at #{pkgshare}/catalog/bootstrap.toml.
      To use dbootstrap from any directory without --catalog, set:
        export XDG_DATA_HOME="#{pkgshare}"
      Or add to your shell profile.
    EOS
  end

  test do
    system "#{bin}/dbootstrap", "--version"
  end
end
```

| Pros | Cons | Complexity |
|------|------|------------|
| Zero Go code changes | Bad UX: `dbootstrap plan` fails by default until user configures XDG_DATA_HOME or uses `--catalog` | Low |
| Formula is self-contained | Caveats are easy to miss | |
| Follows Homebrew conventions strictly | XDG_DATA_HOME override is semantically awkward (it sets the entire data home, not just the catalog path) | |
| Easy to review and test | | |

### 2. Formula + Post-Install Symlink

The formula installs to `#{pkgshare}` and then, during `post_install`, creates a symlink from `$HOME/.local/share/dbootstrap/catalog/bootstrap.toml` → `#{pkgshare}/catalog/bootstrap.toml`. The `post_install` block only creates the target directory if it does not already exist (to avoid overwriting a user's existing catalog).

```ruby
def post_install
  target_dir = "#{Dir.home}/.local/share/dbootstrap/catalog"
  target_file = "#{target_dir}/bootstrap.toml"
  unless File.exist?(target_file)
    mkdir_p target_dir
    ln_s "#{pkgshare}/catalog/bootstrap.toml", target_file
  end
end
```

| Pros | Cons | Complexity |
|------|------|------------|
| No Go code changes | Homebrew discourages modifying `$HOME` during install | Low |
| `dbootstrap plan` works from any CWD out of the box | Symlink not cleaned on uninstall (dangling symlink) | |
| Simple, auditable | If user already has direct-binary install, post_install won't overwrite (good) but also won't link (confusing) | |
| | Formula loses self-containment purity | |

### 3. Formula + Go Catalog Resolution Fallback (Recommended)

The formula installs to `#{pkgshare}` conventionally. The Go code's `catalogPathResolver` gains a third fallback: probe `HOMEBREW_PREFIX` at runtime and check `$HOMEBREW_PREFIX/share/dbootstrap/catalog/bootstrap.toml`. This is looked up **after** XDG_DATA_HOME and `$HOME/.local/share` fallbacks, so user overrides always win.

```go
// In catalogPathResolver.Resolve(), after existing fallbacks:
if prefix, ok := lookupEnv("HOMEBREW_PREFIX"); ok && prefix != "" {
    candidate := filepath.Join(prefix, "share", "dbootstrap", "catalog", "bootstrap.toml")
    if _, err := os.Stat(candidate); err == nil {
        return candidate
    }
}
```

The formula is clean:
- Binary → `bin/dbootstrap`
- Catalog → `pkgshare/catalog/bootstrap.toml`
- `HOMEBREW_PREFIX` is automatically set by Homebrew in the install environment and available to spawned processes

`caveats` become informational only:
```ruby
def caveats
  <<~EOS
    Run `dbootstrap plan --profile dev` to verify the catalog loads.
    The catalog is installed at #{pkgshare}/catalog/bootstrap.toml.
  EOS
end
```

| Pros | Cons | Complexity |
|------|------|------------|
| `dbootstrap plan` works from any CWD without configuration | Requires a small Go code change (∼15 lines) | Medium |
| Formula stays clean and purely Homebrew-conventional | Must test on systems without Homebrew to ensure no regression | |
| HOMEBREW_PREFIX is a reliable, standard Homebrew env var | `os.Stat` probe adds a filesystem access to catalog resolution (currently pure env/home lookup) | |
| No $HOME modification | | |
| Uninstall is clean — Homebrew removes keg, dangling env var is harmless | | |

### 4. Formula + `DBOOTSTRAP_CATALOG` Env Var

The Go code reads `DBOOTSTRAP_CATALOG` as a dedicated env-var override (separate from `--catalog`). The formula sets this env var via shell profile integration. Rejected: Homebrew formula must not modify shell profiles.

| Pros | Cons | Complexity |
|------|------|------------|
| Explicit, no magic | Formula cannot set env vars persistently | Medium |
| | User must configure manually (worse than Approach 3) | |

## Recommendation

**Approach 3 (Formula + Go Catalog Resolution Fallback)** is the recommended path.

**Why**:
1. The formula stays clean and follows Homebrew conventions exactly — no `$HOME` modification, no post-install symlink hacks
2. The Go code change is minimal (∼15 lines in `catalogPathResolver.Resolve()`), purely additive, and introduces no breaking changes
3. `HOMEBREW_PREFIX` is set by Homebrew at formula install time and persists in the shell environment — it's the standard way Homebrew-aware tools find their prefix
4. The fallback is the **last** resort in the resolution chain — user-configured `XDG_DATA_HOME` and `$HOME/.local/share` take priority, then `--catalog` overrides everything
5. Uninstall leaves no residue beyond a dangling env-var reference that a future `dbootstrap` run (from any install method) will gracefully skip
6. The `os.Stat` call only fires when `HOMEBREW_PREFIX` is set, keeping the hot path unchanged for non-Homebrew users

## Detailed Design Notes

### Tap and Formula Naming

- **Tap**: `dnieblesdev/homebrew-dniebles-bootstrap` (separate GitHub repo)
- **Formula**: `dbootstrap.rb` in the tap's `Formula/` directory
- **Install command**: `brew install dnieblesdev/dniebles-bootstrap/dbootstrap`
- **Short form** (after tap): `brew install dbootstrap` (if tap is pinned or unique)

### Stable Release Selection

The formula uses a hardcoded version, URL, and SHA-256 per release. When a new stable release is published, the formula is manually updated. This avoids:
- Formula auto-update complexity (out of scope)
- Accidentally serving a prerelease as stable
- GitHub API rate-limit issues

**Release evidence**: The formula version and SHA-256 MUST correspond to a GitHub Release that is **not** a prerelease. The current `v0.0.0-rc.1` is a prerelease and MUST NOT be used as the formula's stable target. The first stable release (e.g., `v0.1.0`) would be the formula's initial version.

### Linux-Only (macOS Rejection)

The formula uses Homebrew's `on_macos` / `on_linux` blocks:

```ruby
on_macos do
  odie "dbootstrap does not provide macOS binaries. Use Linux or WSL."
end

on_linux do
  url "https://github.com/dnieblesdev/dniebles-bootstrap/releases/download/v#{version}/dbootstrap_v#{version}_linux_#{Hardware::CPU.arch == :arm64 ? "arm64" : "amd64"}.tar.gz"
  # ...
end
```

This rejects macOS at formula evaluation time (before download), with a clear error message.

### Architecture Selection

```ruby
if Hardware::CPU.arm?
  url "..._linux_arm64.tar.gz"
  sha256 "ARM64_SHA256"
else
  url "..._linux_amd64.tar.gz"
  sha256 "AMD64_SHA256"
end
```

### Catalog Installation

```ruby
def install
  bin.install "dbootstrap"
  (pkgshare/"catalog").install "catalog/bootstrap.toml"
end
```

Results in:
- `/home/linuxbrew/.linuxbrew/bin/dbootstrap` → symlink to Cellar
- `/home/linuxbrew/.linuxbrew/share/dbootstrap/catalog/bootstrap.toml`

### Uninstall Behavior

`brew uninstall dbootstrap`:
- Removes `bin/dbootstrap` symlink
- Removes Cellar directory (including `pkgshare/catalog/`)
- No residue in `$HOME` or XDG paths
- `HOMEBREW_PREFIX` env var remains (system-level, not formula-specific)

### Controlled Upgrade

`brew upgrade dbootstrap`:
- Downloads new release archive
- Verifies new SHA-256
- Replaces Cellar contents
- Updates `bin/dbootstrap` symlink
- Overwrites `pkgshare/catalog/bootstrap.toml`

Reinstall (same version): `brew reinstall dbootstrap` reinstalls from the same cached bottle/source.

### Test Block

```ruby
test do
  assert_match "dbootstrap", shell_output("#{bin}/dbootstrap --version")
  system "#{bin}/dbootstrap", "plan", "--profile", "dev", "--catalog", "#{pkgshare}/catalog/bootstrap.toml"
end
```

Verifies the binary runs and the catalog loads successfully.

## Evidence Plan

The following evidence MUST be captured during implementation:

### 1. Tap Discovery
- `brew tap dnieblesdev/dniebles-bootstrap` → exits 0
- `brew tap-info dnieblesdev/dniebles-bootstrap` → shows the tap with at least one formula

### 2. Formula Installation
- `brew install dnieblesdev/dniebles-bootstrap/dbootstrap` → exits 0
- `which dbootstrap` → points to Homebrew-managed binary
- `dbootstrap --version` → reports the installed version

### 3. Catalog Loading from Arbitrary CWD
- `cd /tmp && dbootstrap plan --profile dev` → reads catalog and produces plan (no `--catalog` flag needed)
- `dbootstrap plan --profile dev --catalog /nonexistent/path` → fails with clear catalog-load error (proves `--catalog` override works)

### 4. Reinstall / Controlled Upgrade
- `brew reinstall dbootstrap` → exits 0, catalog and binary replaced
- (Post-v0.2.0): `brew upgrade dbootstrap` → exits 0, new version active

### 5. Uninstall (No Managed-File Residue)
- `brew uninstall dbootstrap` → exits 0
- `which dbootstrap` → not found
- `ls $(brew --prefix)/share/dbootstrap/` → directory does not exist (or is empty)
- `ls $HOME/.local/share/dbootstrap/` → no files created by Homebrew formula

### 6. macOS Rejection
- On macOS: `brew install dnieblesdev/dniebles-bootstrap/dbootstrap` → fails with `odie` message about missing Darwin binaries, no download attempted

### 7. Formula Integrity
- `brew audit --strict dnieblesdev/dniebles-bootstrap/dbootstrap` → passes
- `brew style dnieblesdev/dniebles-bootstrap/dbootstrap` → passes

## Risks

- **Risk 1 — No stable release exists**: The formula needs a non-prerelease GitHub Release. Until `v0.1.0` (or similar) is published, the formula cannot be functional. The implementation should be ready to ship with the first stable release. **Mitigation**: The formula can be tested against `v0.0.0-rc.1` during development (as a prerelease test), but the committed version must target stable.

- **Risk 2 — SHA-256 drift**: Formula SHA-256 values are hardcoded. If a release is re-uploaded with different binaries (GitHub allows this), the formula fails. **Mitigation**: The publish workflow already prevents release overwrites (`release-publish.yml` line 102-110 guards existing tags/releases). This is a pre-existing invariant.

- **Risk 3 — HOMEBREW_PREFIX variability**: On Linux, Homebrew can be installed at `/home/linuxbrew/.linuxbrew` (default), `/opt/homebrew`, or custom paths. The formula itself uses `#{prefix}` which Homebrew resolves correctly, but the Go runtime fallback reads `HOMEBREW_PREFIX` from the environment. If the user runs `dbootstrap` from a shell where Homebrew's `bin/brew shellenv` hasn't been sourced, `HOMEBREW_PREFIX` won't be set and the fallback won't trigger. **Mitigation**: This is acceptable — users who install via Homebrew but don't have it in their shell env are unlikely. The `--catalog` flag remains available as a manual override.

- **Risk 4 — New external dependency**: The tap repository is a new GitHub repo that must be created and maintained. Formula updates are manual. **Mitigation**: The formula only changes when a new release is cut (infrequent), and the update is a single URL+SHA change. No automation is required per scope.

## Ready for Proposal

**Yes**. All major design decisions (tap structure, formula approach, catalog resolution strategy, macOS rejection, evidence plan) have concrete answers. The following contracts must be closed in the design phase:

1. **Formula version source**: Defaults to the latest stable GitHub Release tag; manual formula update per release (no automation).
2. **Catalog resolution fallback**: `HOMEBREW_PREFIX` read at runtime, checked only after existing XDG/$HOME fallbacks, guarded by `os.Stat`.
3. **macOS rejection**: `on_macos { odie "..." }` — explicit, early, no download.
4. **Prerelease protection**: Formula is hardcoded to a specific stable release; prereleases cannot be installed through the channel.
5. **Uninstall contract**: No files outside Homebrew prefix. No `$HOME` or XDG modifications by formula.
