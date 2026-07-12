# Delta for installation-state

## ADDED Requirements

### Requirement: Presence detection uses the configured command name

For tool and runtime resources whose presence detector is command-based, the system MUST probe `Resource.Presence.Name` when that value is configured. It MUST NOT substitute the resource ID, package name, or another catalog field when a configured presence name exists.

#### Scenario: Configured presence name differs from resource ID

- GIVEN a tool or runtime has resource ID `editor` and `Presence.Name` `vim`
- AND the injected PATH lookup reports `vim` present but does not report `editor` present
- WHEN installation-state detection runs
- THEN the resource is reported present

#### Scenario: Missing presence name is not guessed

- GIVEN a tool or runtime has no configured command presence name
- WHEN installation-state detection runs
- THEN the detector preserves its existing unsupported/absent behavior
- AND it does not infer a command name from package metadata or configuration

### Requirement: Idempotency detection is limited to reliable command presence

The system MUST use detected presence for apply idempotency only for tool and runtime resources whose command presence was reliably detected. Presence detection MUST NOT perform package-manager, package-version, configuration, or dotfile-link convergence checks.

#### Scenario: Command presence is sufficient for the first slice

- GIVEN a supported tool or runtime command is found through the injected PATH lookup
- WHEN planning and confirmed execution run
- THEN the plan marks the resource `already_installed`
- AND confirmed execution treats that step as unchanged without mutation

#### Scenario: Dotfile presence does not enable idempotency skipping

- GIVEN a dotfile module directory is present
- BUT the current slice cannot prove that its links are current
- WHEN planning and confirmed execution run
- THEN dotfile link convergence is not inferred from module presence
- AND the dotfile step is not skipped by this command-presence idempotency guard

#### Scenario: Broader reconciliation is not attempted

- GIVEN a resource is selected for planning or execution
- WHEN detection runs
- THEN no package/version/configuration probe, retry, rollback, or bootstrap acquisition is attempted
