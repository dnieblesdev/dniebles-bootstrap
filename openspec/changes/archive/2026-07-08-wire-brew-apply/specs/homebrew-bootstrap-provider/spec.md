# Delta for homebrew-bootstrap-provider

## MODIFIED Requirements

### Requirement: Homebrew bootstrap is detected as a provider-owned need

The system MUST detect when a Homebrew-backed resource requires `brew` but Homebrew is missing.
The detection result MUST be reported as a bootstrap need, not as package installation failure, and confirmed `apply --yes` MUST stop before installation and surface the guidance.
(Previously: missing brew only produced a bootstrap need and no package install was attempted.)

#### Scenario: Missing brew is detected

- GIVEN a Homebrew-backed resource is requested
- WHEN `brew` is unavailable on the host
- THEN the system reports a Homebrew bootstrap need
- AND no package install is attempted

#### Scenario: Brew present does not trigger bootstrap

- GIVEN a Homebrew-backed resource is requested
- WHEN `brew` is available on the host
- THEN no bootstrap need is reported

#### Scenario: Confirmed apply stops on missing brew

- GIVEN `dbootstrap apply --yes` and a Homebrew-backed step
- WHEN `brew` is unavailable on the host
- THEN installation does not proceed
- AND bootstrap guidance is reported

#### Scenario: Missing brew does not attempt package installation

- GIVEN `dbootstrap apply --yes` and a Homebrew-backed package step
- WHEN `brew` is unavailable on the host
- THEN the target package is not installed
- AND bootstrap guidance is the primary outcome

### Requirement: Bootstrap reporting provides explicit manual guidance

The system MUST report a provider-owned bootstrap action with clear manual official install guidance.
The report MUST be reviewable, MUST use official Homebrew website or docs guidance, and MUST NOT include an executable remote-script install command.

#### Scenario: Bootstrap guidance is rendered

- GIVEN Homebrew is missing for a Homebrew-backed need
- WHEN the result is rendered
- THEN the output includes an explicit bootstrap action
- AND the official Homebrew website or docs URL is shown as text

#### Scenario: Guidance remains non-executable

- GIVEN bootstrap guidance is present
- WHEN the report is inspected
- THEN it contains no executable remote script step
- AND it does not include a copy-paste install one-liner
