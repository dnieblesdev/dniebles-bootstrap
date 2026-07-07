# Delta for homebrew-bootstrap-provider

## ADDED Requirements

### Requirement: Homebrew bootstrap is detected as a provider-owned need

The system MUST detect when a Homebrew-backed resource requires `brew` but Homebrew is missing.
The detection result MUST be reported as a bootstrap need, not as package installation failure.

#### Scenario: Missing brew is detected

- GIVEN a Homebrew-backed resource is requested
- WHEN `brew` is unavailable on the host
- THEN the system reports a Homebrew bootstrap need
- AND no package install is attempted

#### Scenario: Brew present does not trigger bootstrap

- GIVEN a Homebrew-backed resource is requested
- WHEN `brew` is available on the host
- THEN no bootstrap need is reported

### Requirement: Bootstrap reporting provides explicit manual guidance

The system MUST report a provider-owned bootstrap action with clear manual official install guidance.
The report MUST be reviewable and MUST NOT require remote script execution to be described.

#### Scenario: Bootstrap guidance is rendered

- GIVEN Homebrew is missing for a Homebrew-backed need
- WHEN the result is rendered
- THEN the output includes an explicit bootstrap action
- AND the official manual install instruction is shown as text

#### Scenario: Guidance remains non-executable

- GIVEN bootstrap guidance is present
- WHEN the report is inspected
- THEN it contains no executable remote script step
