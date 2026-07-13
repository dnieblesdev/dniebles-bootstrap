# Delta for Operational README

## MODIFIED Requirements

### Requirement: README documents the command workflow

The README MUST document `plan`, `apply`, and `bootstrap`, distinguishing planning, execution, and advisory behavior. It MUST also document direct install/uninstall, supported architectures, paths, PATH, force, and catalog location.
(Previously: The README documented only the planning, execution, and bootstrap command workflow.)

#### Scenario: A new operator can identify commands and first install

- GIVEN an operator reads the operational README
- WHEN they look for the workflow or first-install path
- THEN the README explains `plan`, `apply`, and `bootstrap`
- AND it provides install, PATH, catalog, force, and uninstall guidance

#### Scenario: Unsupported platforms are not promised

- GIVEN an operator reads the installation guidance
- WHEN their host is macOS, Windows, or an unsupported architecture
- THEN the README states that direct binary installation is unavailable

### Requirement: README documents target and safety flags

The README MUST document `--profile`, repeatable `--resource`, `--catalog`, `--yes`, `--sudo`, and `--dry-run` accurately. It MUST state that dry-run and yes conflict, sudo requires confirmed yes where supported, and direct install never falls back to sudo or package managers.
(Previously: The README documented command target and safety flags without the direct-install privilege boundary.)

#### Scenario: Flag guidance matches behavior

- GIVEN an operator follows README flag guidance
- WHEN they compare it with the command and installer surfaces
- THEN target, safety, catalog, force, and privilege descriptions match behavior

### Requirement: README states idempotency limits and exclusions

The README MUST retain idempotency limits and state that direct installation verifies checksums, protects unmanaged files, and removes only unmodified manifest-owned files. It MUST NOT promise signing, package managers, macOS, or automatic `dbootstrap install` acquisition.
(Previously: The README stated command-presence idempotency limits and excluded acquisition without describing direct binary lifecycle guarantees.)

#### Scenario: README prevents overclaiming

- GIVEN an operator assesses installation or rerun behavior
- WHEN they read the lifecycle guidance
- THEN checksum-before-mutation, force protection, managed uninstall, and unsupported scope are explicit
