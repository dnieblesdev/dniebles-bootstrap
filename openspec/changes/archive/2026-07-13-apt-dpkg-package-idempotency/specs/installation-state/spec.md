# Delta for installation-state

## ADDED Requirements

### Requirement: Conservative injectable APT detection

On confirmed Linux, the system MUST use injected, read-only `dpkg-query --show --showformat=${Status} <package>`. A well-formed three-field status MUST be installed iff its error field is `ok` and its package-status field is `installed`, including `hold ok installed`. A known well-formed non-installed status MUST be absent only when it is a valid definitive non-installed state; partial states such as `unpacked` or `half-configured` MUST NOT be installed and MUST dispatch normally. The exact absent signature is exit 1, stderr `dpkg-query: no packages found matching <package>`, and no contradictory stdout. Every other non-zero, missing-command, timeout, runner-error, empty, malformed, or ambiguous result MUST be unknown. No `sudo`, `apt-get`, fallback, or retry is permitted.

#### Scenario: Held installed status skips
- GIVEN an eligible package on confirmed Linux
- WHEN the injected query returns `hold ok installed`
- THEN the result is installed and confirmed execution skips the installer

#### Scenario: Partial status is not installed
- GIVEN an eligible package on confirmed Linux
- WHEN the injected query returns `install ok unpacked` or `install ok half-configured`
- THEN the result is not installed and the normal APT installer remains eligible

#### Scenario: Definitive absence dispatches
- GIVEN an eligible package on confirmed Linux
- WHEN the query returns a valid definitive non-installed status or the exact provider not-found signature
- THEN the result is absent and the normal APT installer remains eligible

#### Scenario: Ambiguous evidence is unknown
- GIVEN the query has any other failure or empty, malformed, or ambiguous output
- WHEN detection runs
- THEN the result is unknown and no presence is reported

## MODIFIED Requirements

### Requirement: Idempotency uses reliable presence

The system MUST use presence only for reliable tool/runtime commands or positively confirmed eligible APT packages whose status satisfies the `ok` plus `installed` predicate. It MUST NOT perform version, virtual-package, multi-architecture, configuration, dotfile-link, retry, or fallback checks.
(Previously: idempotency used only reliable tool and runtime command presence and prohibited package-manager probes.)

#### Scenario: Reliable presence skips
- GIVEN a supported command is found or an APT package has a well-formed `ok installed` status
- WHEN confirmed execution runs
- THEN the resource is unchanged without installer dispatch
