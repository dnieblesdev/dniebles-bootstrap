# Design: APT/dpkg Package Idempotency

## Technical Approach

Mirror the existing Brew pre-execution pattern without changing planning or `AptInstaller`. Confirmed Linux `apply --yes` and `bootstrap` probe each eligible APT package once through injected, read-only `dpkg-query --show --showformat=${Status} <package>`, then decorate only an execution-plan copy. Parse the three dpkg fields: a well-formed status is installed iff error is `ok` and package status is `installed`—so `hold ok installed` skips. Definitive well-formed non-installed states, including `unpacked` and `half-configured`, are absent and dispatch normally; the definitive exit-1 not-found signature is also absent. All other evidence is unknown and fails before dispatch.

## Architecture Decisions

| Decision | Options / tradeoff | Choice and rationale |
|---|---|---|
| Probe boundary | Inspect in `AptInstaller` vs. a state adapter | Add `internal/state/AptPackageDetector`, parallel to `BrewFormulaDetector`. Detection stays injectable and read-only; the installer remains the mutation boundary. |
| Three-state classifier | Compare a complete status literal vs. parse fields | Parse non-empty, well-formed three-field successful output. It is installed iff error=`ok` and package-status=`installed`; the desired field does not change that predicate, so `hold ok installed` is installed. Definitive non-installed states—including `install ok unpacked` and `install ok half-configured`—are absent. Separately classify only exit 1 with matching `no packages found matching <package>` stderr and no contradictory stdout as absent. Missing command, nil runner, timeout, runner error, empty/malformed/ambiguous output, and every other non-zero result are unknown. This permits dispatch only on reliable absence evidence. |
| Guard isolation | Share provider guards vs. APT-specific guards | Add separately named APT eligibility/installed/unknown guards beside Brew guards. Each revalidates package kind, `apt` provider, and trimmed package, preventing cross-provider or malformed-plan effects. |
| Composition | Probe all executions vs. confirmed Linux only | Compose after existing Brew decoration only for confirmed modes, `facts.OS == "linux"`, and eligible APT steps. Plan/default/dry-run and non-Linux flows do not probe; the existing non-Linux installer failure remains intact. |

## Data Flow

```text
validated result.Plan -> confirmed Linux APT gate -> AptPackageDetector
                                              |       `dpkg-query` only
                                              v
                           ApplyAptPackagePresence(execution-plan copy)
                                              |
  fields: * ok installed ------------------> installed -> ordered skip
  fields: definitive non-installed --------> absent ----> normal APT installer
    (including unpacked / half-configured)
  exit 1 + matching stderr + no stdout ----> absent ----> normal APT installer
  every other result ----------------------> unknown ---> ordered failure
                                              |              (no dispatch)
                                              v
                                            Runner
```

`runApplyLike` preserves its sequence: validate the original plan, decorate its execution copy after Brew, build the provider-aware runner, execute every step in order, then append bootstrap guidance and render. No detection calls `sudo`, `apt-get`, fallback probe, or retry.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/state/apt_package_detector.go` | Create | Injectable eligibility, read-only probe, strict three-state classifier, and immutable copy decorator. |
| `internal/state/apt_package_detector_test.go` | Create | Table-driven classifier, command-vector, no-probe, and copy-isolation tests. |
| `internal/planning/types.go` | Modify | Generalize the `PackagePresence` comment from Brew-only to provider-specific transient presence. |
| `internal/execution/runner.go` | Modify | Add isolated APT installed-skip and unknown-fail guards; absent falls through unchanged. |
| `internal/execution/runner_test.go` | Modify | Add ordered APT skip/fail/absent-dispatch and provider-isolation cases. |
| `cmd/dbootstrap/main.go` | Modify | Compose the injected APT detector after Brew only for confirmed Linux eligible plans. |
| `cmd/dbootstrap/main_test.go` | Modify | Cover apply/bootstrap composition and safe/non-Linux non-probing. |

## Interfaces / Contracts

```go
type AptPackageDetector struct {
    CommandExists execution.CommandExists
    Runner        execution.CommandRunner
    Timeout       time.Duration
}

func (d AptPackageDetector) Detect(context.Context, planning.Plan) map[planning.ResourceRef]planning.PackagePresence
func ApplyAptPackagePresence(planning.Plan, map[planning.ResourceRef]planning.PackagePresence) planning.Plan
```

The sole request is `CommandRequest{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", packageName}, Timeout: ...}`. The classifier parses desired, error, and package-status fields; installed is exactly `error == "ok" && packageStatus == "installed"`. The exit-1 absence path is admitted only when stderr identifies that exact package as not found and stdout contains no status or other contradictory data. It reuses `PackagePresence` (`installed`, `absent`, `unknown`); no catalog, planner, or APT-installer contract changes.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Field predicate: `hold ok installed` and other `* ok installed` skip; `unpacked`/`half-configured` dispatch; definitive exit-1 not-found; contradictory stdout; empty/malformed success; unexpected non-zero; missing command/nil runner/error/timeout | Table-driven detector tests with recording seams; assert one exact request at most, only definitive absence dispatches, and every other failure is unknown. |
| Unit | Installed skip, unknown fail/no dispatch, field-derived and exit-1 absence dispatch, ordering, and Brew/non-APT isolation | Extend runner recording-installer tests; assert result position and calls. |
| CLI integration | Confirmed Linux `apply`/`bootstrap` composition for held installed, partial states, and definitive not-found; default/dry-run/plan/non-Linux non-probing | Existing CLI stubs and sequence runner; assert no `sudo`/`apt-get` for installed or unknown and normal installer dispatch for absent. |
| Regression | Affected packages and repository suite | Run focused `go test` for state, execution, and `cmd/dbootstrap`, then `go test ./...`. |

## Migration / Rollout

No migration required. Revert detector, composition, and runner guards together to restore prior confirmed APT behavior.

## Open Questions

None.
