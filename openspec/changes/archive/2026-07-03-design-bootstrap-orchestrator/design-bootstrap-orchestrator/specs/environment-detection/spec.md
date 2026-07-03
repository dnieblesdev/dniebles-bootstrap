# Delta for environment-detection

## ADDED Requirements

### Requirement: Environment facts before planning

The system MUST detect OS, distro, WSL status, and CPU architecture before resolving install plans.

#### Scenario: Platform facts feed planning

- GIVEN a profile or point install request
- WHEN planning begins
- THEN environment facts are collected first
- AND plan resolution uses those facts

#### Scenario: Unsupported platform is visible

- GIVEN an unsupported OS, distro, WSL mode, or architecture
- WHEN planning runs
- THEN the result reports the unsupported fact
- AND no ambiguous installer choice is made

### Requirement: Environment facts are reportable

The system SHOULD include detected environment facts in plan/report output for CLI and future TUI use.

#### Scenario: User can inspect detected facts

- GIVEN a generated plan or result
- WHEN the user reviews operational output
- THEN OS, distro, WSL status, and architecture are visible
- AND troubleshooting does not require rerunning detection manually
