# Delta for profile-install-planning

## ADDED Requirements

### Requirement: Profile install planning

The system MUST produce a validated plan for profile installs that expands bundles, tools, runtimes, packages, and dotfiles requests.

#### Scenario: Full profile planning succeeds

- GIVEN a valid profile and environment facts
- WHEN planning is requested
- THEN the system emits an ordered profile plan
- AND the plan includes all declared resource groups

#### Scenario: Profile install with dependency gaps is reported

- GIVEN the profile requires unavailable dependencies
- WHEN planning runs
- THEN the missing dependency is reported in the plan result
- AND the remaining plan can still be inspected

### Requirement: Missing dotfiles config is reported, not blocked

The system MUST allow profile install planning and tool installation to proceed when expected dotfiles configuration is missing, and MUST report that missing configuration requires attention.

#### Scenario: Installation continues with missing config

- GIVEN the required dotfiles configuration is absent
- WHEN the profile install is planned or executed
- THEN tool installation continues
- AND the plan/result reports missing configuration requiring attention

#### Scenario: Attention signal remains visible

- GIVEN a completed profile install without dotfiles config
- WHEN the user reviews the result
- THEN the missing configuration is still surfaced
- AND it is distinguishable from successful configuration
