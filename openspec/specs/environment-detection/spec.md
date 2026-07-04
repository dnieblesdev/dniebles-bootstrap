# Delta for environment-detection

## ADDED Requirements

### Requirement: Host facts adapter

The system MUST detect OS, architecture, distro, and WSL into `planning.EnvironmentFacts` or compatible domain data.
The detector MUST live outside `internal/planning`.

#### Scenario: Detect supported host facts

- GIVEN runtime OS and arch are available
- WHEN the adapter runs
- THEN it returns facts for OS, architecture, distro, and WSL status
- AND planning receives only domain facts

#### Scenario: Planning stays pure

- GIVEN planning is building a plan
- WHEN facts are supplied
- THEN `internal/planning` does not probe the host or import the adapter

### Requirement: Host-independent detection seams

The system MUST support injected seams for runtime OS/arch, env vars, file reads, and kernel/proc text.
Tests MUST be deterministic and host-independent.

#### Scenario: Deterministic detection test

- GIVEN fake runtime, env, and file providers
- WHEN detection is exercised
- THEN the result matches the fixture data

#### Scenario: Missing optional data

- GIVEN an optional file or env key is absent
- WHEN detection runs
- THEN it does not hard fail

### Requirement: Conservative distro and WSL fallback

The system MUST detect distro from `/etc/os-release` style data with conservative fallback.
It MUST use multiple WSL signals, including env vars and kernel/proc text, and MAY return unknown or false when signals are insufficient.

#### Scenario: Distro from os-release

- GIVEN os-release data contains a clear distro identifier
- WHEN detection runs
- THEN the distro fact is populated conservatively

#### Scenario: WSL signal fallback

- GIVEN one WSL signal is present and another is absent
- WHEN detection runs
- THEN WSL detection is deterministic and reflects the supported signal set

### Requirement: CLI plan consumes detected facts

The `plan` command MUST use detected facts instead of hardcoded values.
It MUST still avoid installer, dotfiles runtime, command-runner, apply/install, and TUI side effects.

#### Scenario: Plan uses detected host facts

- GIVEN detection succeeds
- WHEN `plan` runs
- THEN the plan is built with detected facts
- AND rendered output shows the facts used

#### Scenario: Plan avoids side effects

- GIVEN `plan` is invoked
- WHEN facts are detected
- THEN no installer or runtime mutation occurs

## REMOVED Requirements

### Requirement: Hardcoded static planning facts in CLI

(Reason: CLI facts must come from host detection.)
(Migration: Replace call sites with detected facts.)
