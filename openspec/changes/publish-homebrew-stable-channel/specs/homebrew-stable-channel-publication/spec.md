# Homebrew Stable Channel Publication Specification

## Purpose

Define the release gate and evidence for Linux/WSL Homebrew publication. Direct resolver behavior is out of scope.

## Requirements

### Requirement: Qualify the stable release before publication

The change MUST remain OPEN/BLOCKED until a GitHub Release is public, non-draft, non-prerelease, and contains Linux `amd64` and `arm64` archives with matching SHA-256 assets. A prerelease MAY support technical validation only; it MUST NOT publish the stable channel or be marked stable. Qualification MUST validate each `.sha256` file's content against the corresponding archive bytes, not merely confirm asset presence.

#### Scenario: Qualifying release unlocks publication

- GIVEN release metadata proves public, non-draft, and non-prerelease status
- AND both archives and matching SHA-256 assets are present
- AND each SHA-256 file content matches the digest of the downloaded archive bytes
- WHEN the stable gate is evaluated
- THEN the gate is satisfied and publication MAY proceed

#### Scenario: Missing or prerelease release blocks publication

- GIVEN the release is draft, prerelease, unavailable, or missing either archive/digest
- WHEN publication is evaluated
- THEN the change remains OPEN/BLOCKED and no stable channel is published

#### Scenario: Checksum content mismatch blocks publication

- GIVEN a non-prerelease release with the required asset names
- WHEN a `.sha256` file's content does not match the digest of its archive bytes
- THEN the change remains OPEN/BLOCKED and no stable channel is published

### Requirement: Pin verified stable formula assets

The published formula MUST pin the qualified version, archive URL/name, and matching SHA-256 digest for `amd64` and `arm64`. It MUST NOT use a prerelease, “latest” lookup, or unresolved placeholder.

#### Scenario: Formula metadata matches the release

- GIVEN a qualifying release is selected
- WHEN the formula is reviewed
- THEN both branches identify its exact version, URL, asset name, and digest

### Requirement: Prove supported Linux/WSL lifecycle acceptance

Evidence MUST include separate Linux/WSL `amd64` and `arm64` runs covering tap, install, version, strict audit/style, reinstall or controlled upgrade, uninstall, managed-payload removal, unrelated-file preservation, and installed catalog path.

#### Scenario: Both supported architectures pass

- GIVEN the pinned formula is tested on both architectures
- WHEN lifecycle and integrity checks pass
- THEN the channel has publishable technical acceptance evidence for both architectures

### Requirement: Prove the macOS platform boundary

The channel MUST reject macOS with a clear unsupported-platform result before any release download is attempted.

#### Scenario: macOS is rejected before network access

- GIVEN the host platform is macOS
- WHEN formula installation is attempted
- THEN installation fails clearly and evidence shows no download occurred

### Requirement: Complete final channel verification

Final verification MUST record release metadata, architecture results, macOS rejection, documentation, and scope. It MUST confirm no prerelease was published as stable.

#### Scenario: Evidence supports the final decision

- GIVEN evidence and documentation are available
- WHEN final verification is reviewed
- THEN availability is proven, or OPEN/BLOCKED status names the missing gate
