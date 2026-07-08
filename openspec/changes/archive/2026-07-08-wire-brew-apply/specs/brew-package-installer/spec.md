# Delta for brew-package-installer

## MODIFIED Requirements

### Requirement: Brew package installation is provider-gated

The system MUST only install packages when structured install metadata declares `Install.Provider == "brew"` and `Install.Package` is non-empty.
Unsupported providers or missing package data MUST return a non-success result, and confirmed `apply --yes` MAY invoke brew only for those eligible steps.

#### Scenario: Brew install is accepted

- GIVEN a resource with `Install.Provider == "brew"` and a non-empty package name
- WHEN the installer handles the resource
- THEN the resource is eligible for brew installation

#### Scenario: Unsupported or incomplete metadata is rejected

- GIVEN a resource with another provider or an empty package name
- WHEN the installer handles the resource
- THEN the installer returns a structured failure or unsupported result

#### Scenario: Confirmed apply can reach eligible brew installs

- GIVEN a brew-backed resource and `dbootstrap apply --yes`
- WHEN the installer is invoked through apply
- THEN the resource may be installed with brew

### Requirement: Brew installation uses explicit command requests only

The system MUST build `CommandRequest{Executable:"brew", Args:["install", package]}` and MUST invoke it only through the injected `CommandRunner` seam.
The system MUST NOT use raw command metadata, shell-first fields, `sh -c`, pipelines, dotfiles execution, or the bootstrap entrypoint.
(Previously: brew installation existed as an isolated component and was not wired into apply.)

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
Confirmed `apply --yes` MUST surface bootstrap guidance through Homebrew bootstrap behavior, MUST NOT attempt target package installation, and MUST NOT install Homebrew.

#### Scenario: Brew is missing

- GIVEN a brew-backed resource and no `brew` executable on PATH
- WHEN the installer runs
- THEN the result reports missing brew
- AND no install command is executed

#### Scenario: Missing brew blocks apply installation attempts

- GIVEN `dbootstrap apply --yes` and no `brew` executable on PATH
- WHEN the brew-backed step is reached
- THEN the target package installation does not proceed
- AND bootstrap guidance is reported instead

#### Scenario: Missing brew does not trigger Homebrew installation

- GIVEN `dbootstrap apply --yes` and no `brew` executable on PATH
- WHEN the brew-backed step is reached
- THEN no Homebrew installation is attempted
- AND the result remains a bootstrap advisory

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
