# Delta for apply-command-dry-run

## ADDED Requirements

### Requirement: APT detection is confirmed-Linux-only

The CLI MUST compose APT detection only for confirmed Linux `apply --yes` and `bootstrap` APT steps. APT presence MUST be installed only for a well-formed status whose error field is `ok` and package-status field is `installed`, including `hold ok installed`. Partial states such as `unpacked` or `half-configured` MUST remain executable rather than being skipped. Default, dry-run, planning-only, and non-Linux flows MUST NOT probe `dpkg-query`; non-Linux confirmed APT steps MUST fail without probes.

#### Scenario: Definitive not-found reaches installer
- GIVEN confirmed Linux `apply --yes` or `bootstrap` has an eligible absent package
- WHEN the query exits 1 with matching `no packages found matching <package>` stderr and no contradictory stdout
- THEN it is classified absent and dispatched through the normal APT installer
- AND no retry, fallback, or alternate probe occurs

#### Scenario: Held installed package is skipped
- GIVEN confirmed Linux `apply --yes` or `bootstrap` returns `hold ok installed`
- WHEN execution processes the package
- THEN it reports unchanged without dispatching APT

#### Scenario: Partial package state is not skipped
- GIVEN confirmed Linux `apply --yes` or `bootstrap` returns `install ok unpacked` or `install ok half-configured`
- WHEN execution processes the package
- THEN it dispatches the normal APT installer

#### Scenario: Safe or non-Linux modes do not probe
- GIVEN a safe-mode or non-Linux command has APT-backed resources
- WHEN it runs
- THEN none of `dpkg-query`, `apt-get`, or `sudo` is invoked

## MODIFIED Requirements

### Requirement: Apply excludes broader convergence

Idempotency MUST be limited to reliable command presence and three-state APT presence on confirmed Linux. Apply MUST NOT use versions, virtual packages, configuration, dotfile content, retries, rollback, or acquisition.
(Previously: apply excluded all package presence from idempotency decisions.)

#### Scenario: Installed and unknown remain safe
- GIVEN detection returns exact installed status or unknown evidence
- WHEN apply executes
- THEN installed may report unchanged, while unknown fails without installer dispatch
- AND no retry, rollback, or acquisition occurs
