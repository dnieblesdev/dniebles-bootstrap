# Delta for Operational README

## MODIFIED Requirements

### Requirement: README documents the command workflow and Homebrew summary

The README MUST document `plan`, `apply`, and `bootstrap`, distinguishing planning, execution, and advisory behavior. It MUST also document direct install/uninstall, supported architectures, paths, PATH, force, catalog location, and the Linux/WSL Homebrew channel at a summary level only: tap command, install/uninstall/upgrade lifecycle, stable-release prerequisite, installed catalog path, and macOS exclusion. Detailed publication evidence, archive hashes, and operational proof MUST live in the tap README, not the main README.
(Previously: The README documented the command workflow and direct-install guidance but not the Homebrew stable channel lifecycle and boundary.)

#### Scenario: A new operator can identify commands and first install

- GIVEN an operator reads the operational README
- WHEN they look for the primary workflow
- THEN the README describes `plan` for inspecting selected work
- AND `apply` for reporting or performing the supported execution modes
- AND `bootstrap` with its actual execution semantics and safety boundary
- AND it provides install, PATH, catalog, force, and uninstall guidance

#### Scenario: Unsupported platforms are not promised

- GIVEN an operator reads the installation guidance
- WHEN their host is macOS, Windows, or an unsupported architecture
- THEN the README states that direct binary installation is unavailable

#### Scenario: Homebrew lifecycle and boundary are actionable

- GIVEN an operator wants the supported Homebrew installation path
- WHEN they read the README
- THEN it explains tap, install, reinstall or controlled upgrade, uninstall, and Linux/WSL `amd64`/`arm64` support
- AND it states that stable publication requires a qualifying release, identifies the installed catalog path, and excludes macOS
- AND it links to the tap README for detailed publication evidence, hashes, and operational proof
