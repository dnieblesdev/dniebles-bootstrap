# apt-package-installer Specification

## Purpose

Define explicit, provider-gated APT installation through the existing command boundary.

## Requirements

### Requirement: APT installation is explicitly provider-gated

The system MUST accept only `Install.Provider == "apt"` with a trimmed package that is non-empty and does not begin with `-` for APT installation. Other providers or invalid metadata MUST return structured non-success results.

#### Scenario: Valid APT metadata is accepted

- GIVEN a `tool` or `package` resource has provider `apt` and a non-empty package that does not begin with `-`
- WHEN the APT installer handles it
- THEN the resource is eligible for APT installation

#### Scenario: Other metadata is rejected

- GIVEN metadata selects another provider, is absent, has an empty package after trimming, or has a package beginning with `-`
- WHEN the APT installer handles it
- THEN it returns structured non-success and does not run a command

### Requirement: APT uses an explicit privilege command vector

For confirmed Linux execution, the system MUST send exactly one vector: `apt-get install -y -- <package>` for `apply --yes`, or `sudo apt-get install -y -- <package>` for explicit `apply --yes --sudo`. The `--` delimiter prevents option injection through custom catalog metadata; it is not shell escaping. The system MUST set `CommandRequest.Timeout` to ten minutes through the existing CLI composition seam. It MUST NOT construct a shell string, automatically escalate, retry, select a fallback, bootstrap/update APT, change repositories, detect package presence, or orchestrate rollback.

#### Scenario: Direct install request is explicit

- GIVEN valid APT metadata and an already-privileged confirmed process
- WHEN installation begins without `--sudo`
- THEN `CommandRunner` receives `apt-get install -y -- <package>` as separate executable and argument values with a ten-minute timeout

#### Scenario: Sudo installation request is explicit

- GIVEN valid APT metadata and confirmed `--sudo` mode
- WHEN installation begins
- THEN `CommandRunner` receives `sudo apt-get install -y -- <package>` as separate executable and argument values with a ten-minute timeout

#### Scenario: Command success is reported

- GIVEN the selected APT command vector succeeds
- WHEN the installer records the command result
- THEN the resource result is structured success

#### Scenario: Command failure remains non-success

- GIVEN the command runner reports failure or timeout
- WHEN the installer records the outcome
- THEN the result is `StepStatusFailed`, exposes the command outcome, and makes no retry or rollback claim

### Requirement: APT availability and privilege mode are required

The system MUST validate that `--sudo` is used only with confirmed `--yes` execution. It MUST use the injected `apt-get` availability seam before either command vector and the injected `sudo` availability seam before the sudo vector. Missing executables MUST produce structured non-success without invoking `CommandRunner`.

#### Scenario: APT is unavailable

- GIVEN an eligible APT resource and no available `apt-get`
- WHEN installation is attempted
- THEN the result reports structured non-success and no command runs

#### Scenario: Sudo is unavailable

- GIVEN confirmed `--yes --sudo` mode and no available `sudo`
- WHEN installation is attempted
- THEN the result reports structured non-success and no command runs

#### Scenario: Sudo is invalid outside confirmed mode

- GIVEN `--sudo` is supplied without `--yes` or with `--dry-run`
- WHEN flags are validated
- THEN the command returns a usage error and performs no probe or command

### Requirement: APT proof is opt-in

Tests MUST prove both privilege vectors, flag validation, Linux gating, missing executable behavior, command outcomes, and non-mutating default/dry-run behavior with explicit fixtures or custom catalog targets. Tests MUST leave the default catalog unchanged.

#### Scenario: Default catalog is preserved

- GIVEN the default catalog is loaded
- WHEN the APT slice is exercised
- THEN no default APT target is added or migrated

#### Scenario: Test proof covers the safety boundary

- GIVEN table-driven installer tests and a temporary custom catalog
- WHEN the APT contract suite runs
- THEN it proves exact vectors and all rejected/non-mutating paths without a real external command
