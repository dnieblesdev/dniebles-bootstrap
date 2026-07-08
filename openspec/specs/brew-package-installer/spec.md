# Delta for brew-package-installer

## ADDED Requirements

### Requirement: Brew package installation is provider-gated

The system MUST only install packages when structured install metadata declares `Install.Provider == "brew"` and `Install.Package` is non-empty.
Unsupported providers or missing package data MUST return a non-success result.

#### Scenario: Brew install is accepted

- GIVEN a resource with `Install.Provider == "brew"` and a non-empty package name
- WHEN the installer handles the resource
- THEN the resource is eligible for brew installation

#### Scenario: Unsupported or incomplete metadata is rejected

- GIVEN a resource with another provider or an empty package name
- WHEN the installer handles the resource
- THEN the installer returns a structured failure or unsupported result

### Requirement: Brew installation uses explicit command requests only

The system MUST build `CommandRequest{Executable:"brew", Args:["install", package]}` and MUST invoke it only through the injected `CommandRunner` seam.
The system MUST NOT use raw command metadata, shell-first fields, `sh -c`, pipelines, dotfiles execution, or the bootstrap entrypoint.

#### Scenario: Brew install request is constructed explicitly

- GIVEN a valid brew-backed resource
- WHEN installation begins
- THEN the command request targets `brew install <package>`
- AND the request is sent through `CommandRunner`

#### Scenario: Shell-based execution is not allowed

- GIVEN a valid brew-backed resource
- WHEN the installer prepares execution
- THEN no shell string or pipeline representation is used

### Requirement: Missing brew is reported as a structured failure

The system MUST detect when `brew` is unavailable and MUST return a structured non-success result without attempting installation.

#### Scenario: Brew is missing

- GIVEN a brew-backed resource and no `brew` executable on PATH
- WHEN the installer runs
- THEN the result reports missing brew
- AND no install command is executed

### Requirement: Command execution outcomes are surfaced

The system MUST surface command success and command failure as structured execution results.

#### Scenario: Command success is reported

- GIVEN `brew install <package>` completes successfully
- WHEN the installer receives the command result
- THEN the installation result reports success

#### Scenario: Command failure is reported

- GIVEN `brew install <package>` exits non-zero
- WHEN the installer receives the command result
- THEN the installation result reports failure
- AND the command outcome remains visible

### Requirement: Apply remains noop and unwired to brew installation

The system MUST NOT wire brew installation into `dbootstrap apply` in this slice.
Apply MUST remain noop and non-mutating, and the new installer MUST remain available only as an isolated component.

#### Scenario: Apply does not trigger brew installation

- GIVEN `dbootstrap apply` runs with a brew-backed resource
- WHEN execution is evaluated
- THEN no brew install command is issued

#### Scenario: Installer remains isolated from CLI mutation

- GIVEN the installer component exists
- WHEN the CLI composition root is inspected
- THEN apply does not register the installer for mutation
