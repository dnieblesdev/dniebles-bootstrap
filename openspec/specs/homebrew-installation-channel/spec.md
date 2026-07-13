# Homebrew Installation Channel Specification

## Purpose

Define the Linux/WSL Homebrew resolver and formula contract for `dbootstrap`. Stable publication, lifecycle evidence, and README documentation are owned by `publish-homebrew-stable-channel`.

## Requirements

### Requirement: Provide a Homebrew-prefix catalog fallback

Default catalog discovery MUST preserve the existing precedence `--catalog` > XDG data > `$HOME/.local/share`, then add a lower-priority Homebrew-prefix candidate. The fallback MUST be CWD-independent and skipped when `HOMEBREW_PREFIX` is unavailable.

#### Scenario: Homebrew catalog is the last-resort default

- GIVEN no explicit, XDG, or `$HOME/.local/share` catalog exists and `HOMEBREW_PREFIX` identifies an installed formula catalog
- WHEN `dbootstrap plan --profile dev` runs from an arbitrary working directory without `--catalog`
- THEN it loads the Homebrew catalog at `<HOMEBREW_PREFIX>/share/dbootstrap/catalog/bootstrap.toml`

#### Scenario: Higher-priority catalog wins

- GIVEN explicit, XDG, home-local, and Homebrew-prefix catalogs are available
- WHEN `dbootstrap plan --profile dev` runs
- THEN it loads the first candidate in that precedence order

#### Scenario: Absent Homebrew prefix omits the fallback

- GIVEN `HOMEBREW_PREFIX` is unset or empty
- WHEN default catalog resolution runs
- THEN no Homebrew candidate is considered and existing fallbacks behave as before

### Requirement: Define the pinned formula contract

The formula design MUST target one public, non-prerelease GitHub Release and pin its Linux `amd64` and `arm64` archive URLs with matching SHA-256 values. It MUST support Linux and WSL on those architectures only; release availability and publication are owned by `publish-homebrew-stable-channel`, and prerelease assets MUST NOT be substituted.

#### Scenario: Supported Linux or WSL installation

- GIVEN a stable release with pinned Linux assets and matching digests
- WHEN Homebrew installs the formula on Linux/WSL `amd64` or `arm64`
- THEN the selected archive is verified and the binary and catalog are installed

#### Scenario: Missing stable release evidence

- GIVEN no publicly available stable release contains the required immutable assets
- WHEN the tap formula is prepared for publication
- THEN publication is blocked and no prerelease release is presented as stable

### Requirement: Define install/uninstall within the Homebrew prefix

The Homebrew channel MUST install the executable and catalog in Homebrew's conventional prefix locations, with the catalog under the formula package share location. It MUST NOT write outside the prefix, and uninstall MUST remove formula-managed payloads without removing unrelated files.

#### Scenario: Reinstall and clean uninstall

- GIVEN the formula is installed and its managed files are present
- WHEN the operator reinstalls and then uninstalls it
- THEN the payload remains usable after reinstall and all formula-managed files are removed

#### Scenario: Unrelated files are preserved

- GIVEN unrelated files exist in or around the Homebrew prefix
- WHEN the formula is uninstalled
- THEN those files remain unchanged

### Requirement: Reject unsupported macOS before download

The channel MUST explicitly reject macOS before attempting any release download or installation because Darwin assets are out of scope.

#### Scenario: macOS installation is blocked early

- GIVEN the host is macOS
- WHEN Homebrew attempts to install the formula
- THEN it fails with a clear unsupported-platform message and makes no download attempt

### Requirement: Provide evidence for the completed resolver contract

Acceptance evidence for this change MUST cover the nine table-driven resolver cases and the formula/catalog contract. Lifecycle evidence (`brew tap`, install, `--version`, strict audit/style, reinstall, uninstall) and stable release metadata evidence are owned by `publish-homebrew-stable-channel`.

#### Scenario: Resolver evidence is complete

- GIVEN the resolver unit tests run
- WHEN the change is reviewed
- THEN nine passing cases prove explicit, XDG, home-local, Homebrew fallback, higher-priority wins, absent `HOMEBREW_PREFIX`, and no-existing-candidate behavior
