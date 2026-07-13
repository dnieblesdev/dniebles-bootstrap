# Proposal: APT/dpkg Package Idempotency

## Intent

Avoid redundant confirmed APT installs while allowing demonstrably absent packages to continue to normal APT installation. Ambiguous evidence must never cause a skip or dispatch.

## Scope

### In Scope
- Detect eligible Linux APT packages through injected, read-only `dpkg-query --show --showformat=${Status} <package>`.
- Apply a three-state contract: parse only well-formed three-field dpkg status output. Any status with error field `ok` and installation-status field `installed` is present and skips, including `hold ok installed`; partial states such as `unpacked` and `half-configured` are absent and dispatch. The definitive provider not-found signature (exit 1, matching `no packages found matching <package>` stderr, and no contradictory stdout) is also absent. Ambiguous/error/malformed/unavailable/timeout evidence is unknown and blocks dispatch.
- In confirmed Linux `apply` and `bootstrap`, decorate only the execution-plan copy; runner guards preserve ordering and prevent installer calls for installed or unknown packages.
- Add deterministic detector, runner, and CLI composition tests for all states and safety gates.

### Out of Scope
- Version, virtual-package, multi-architecture, or general package reconciliation.
- Detection through `sudo`, `apt-get`, `apt-cache`, `apt list`, retries, or fallback commands.
- Catalog, planner, `AptInstaller`, Homebrew, and dotfiles changes.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `installation-state`: add injectable APT detection with installed, absent, and unknown outcomes without changing planning semantics.
- `execution-contracts`: preserve ordered APT dispatch for absent packages and prevent dispatch for installed or unknown packages.
- `apply-command-dry-run`: compose detection only for confirmed Linux apply/bootstrap; safe modes remain non-probing and non-mutating.

## Approach

Mirror Brew's state-detector → execution-plan decoration → runner-guard pattern. `AptPackageDetector` owns provider-specific classification; `runner.go` independently revalidates package kind, provider, and non-empty package; `main.go` composes it after Brew only for confirmed Linux execution. Detection never escalates privileges or enters the installer path.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/state/apt_package_detector.go` | New | Read-only three-field status parser and classifier. |
| `internal/execution/runner.go` | Modified | Three-state APT guards. |
| `cmd/dbootstrap/main.go` | Modified | Confirmed Linux composition. |
| Related `*_test.go` | Modified/New | Deterministic state and safety coverage. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Unexpected `dpkg-query` result | Low | Classify unknown; do not dispatch. |
| False absence signature | Low | Require exit, matching stderr, and non-contradictory stdout. |
| Provider cross-talk | Low | Revalidate eligibility in runner guards. |

## Rollback Plan

Revert the detector, plan decoration, and guards together; confirmed APT resumes existing installer behavior without catalog changes.

## Dependencies

- `dpkg-query` is required only for confirmed Linux detection.

## Success Criteria

- [ ] Well-formed statuses with `ok` and `installed`, including `hold ok installed`, skip without APT dispatch.
- [ ] Well-formed partial/non-installed statuses, including `unpacked` and `half-configured`, and the definitive not-found signature dispatch normal APT installation.
- [ ] Unknown evidence makes no `apt-get` or `sudo` call; default, dry-run, and non-Linux paths do not probe.
