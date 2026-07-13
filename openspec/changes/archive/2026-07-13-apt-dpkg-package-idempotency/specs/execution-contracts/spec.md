# Delta for execution-contracts

## ADDED Requirements

### Requirement: Conservative confirmed-Linux APT guards

Execution MUST preserve order. A well-formed status MUST skip installation iff its error field is `ok` and package-status field is `installed`, including `hold ok installed`. A valid definitive non-installed status, or the exact provider-specific not-found signature (exit 1, stderr `dpkg-query: no packages found matching <package>`, and no contradictory stdout), MUST dispatch the normal APT installer. Partial states such as `unpacked` or `half-configured` MUST NOT skip and MUST dispatch. Unknown MUST fail without installer, `apt-get`, or `sudo`. Detection MUST be injected, read-only, and free of retries or fallbacks.

#### Scenario: Installed skips; absent dispatches
- GIVEN confirmed execution has an APT step classified installed or absent
- WHEN the runner processes it
- THEN installed is unchanged and absent dispatches the normal installer
- AND the step remains in its original position

#### Scenario: Held installed skips
- GIVEN confirmed execution has an APT step with status `hold ok installed`
- WHEN the runner processes it
- THEN it reports unchanged and makes no installer or command call

#### Scenario: Partial state does not skip
- GIVEN confirmed execution has an APT step with status `install ok unpacked` or `install ok half-configured`
- WHEN the runner processes it
- THEN it dispatches the normal APT installer

#### Scenario: Not-found dispatches
- GIVEN the query exits 1 with matching `no packages found matching <package>` stderr and no contradictory stdout
- WHEN the runner processes the step
- THEN it dispatches the normal APT installer without retry or fallback

#### Scenario: Unknown fails safely
- GIVEN the APT result is unknown
- WHEN the runner processes the plan
- THEN it reports failure and makes no installer, `apt-get`, or `sudo` call
