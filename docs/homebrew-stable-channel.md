# Homebrew stable channel

Install the stable `dbootstrap` formula from this repository's custom-URL tap on Linux or WSL. The supported architectures are `amd64` and `arm64`; macOS is intentionally rejected before Homebrew downloads a release asset.

## Quick path

```bash
brew tap dnieblesdev/dniebles-bootstrap https://github.com/dnieblesdev/dniebles-bootstrap.git
brew install dnieblesdev/dniebles-bootstrap/dbootstrap
```

The installed catalog is at `$(brew --prefix)/share/dbootstrap/catalog/bootstrap.toml`.

## Validation paths

| Stage | Command | Purpose |
|---|---|---|
| Pull request candidate | `HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source ./Formula/dbootstrap.rb` | Validates the checked-out candidate formula without treating a remote tap as PR evidence. |
| Merged public formula | `HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source dnieblesdev/dniebles-bootstrap/dbootstrap` | The post-merge GitHub-hosted smoke verifies the published custom-URL tap. |

The public smoke runs only after a push to `main`, checks out the triggering commit in the custom-URL tap before installation, and uploads its version, catalog, tap-repository, system, and triggering-SHA receipts. The receipt therefore identifies the formula bytes that the job installed. It does not publish releases, tags, or a separate tap.

## Stable pin

| Platform | Archive | SHA-256 |
|---|---|---|
| Linux/WSL amd64 | `dbootstrap_v0.1.0_linux_amd64.tar.gz` | `a8f21a55019ff09c08a124f30bffc6831c960be81cbd1496e43b26c92784d109` |
| Linux/WSL arm64 | `dbootstrap_v0.1.0_linux_arm64.tar.gz` | `8732f1e03ba4dc0d1a6132dd74a3291364e615aff8c52bc67727ff3f0999de6e` |

`v0.1.0` is a public non-prerelease release. The formula installs `dbootstrap` and `share/dbootstrap/catalog/bootstrap.toml`.

## Boundaries and rollback

- Linux/WSL on `amd64` and `arm64` is supported. macOS remains unsupported and rejects before release-asset download.
- Use the exact custom-URL tap command above and the qualified formula name; no standalone tap is part of this channel.
- To remove the channel, run `brew uninstall dbootstrap` and `brew untap dnieblesdev/dniebles-bootstrap`. Reverting the formula or smoke workflow does not change release publication, the direct installer, or resolver behavior.
