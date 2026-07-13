# Exploration: apt-dpkg-package-idempotency

## Current State

### Brew formula detection (reference pattern — already implemented)

The system already implements provider-specific package-presence idempotency for Homebrew:

| Layer | File | What it does |
|-------|------|--------------|
| State detection | `internal/state/brew_formula_detector.go` | Probes eligible Brew packages via `brew list --formula <InstallMetadata.Package>` through injected `CommandExists` and `CommandRunner` seams. Classifies results as `installed` (exit 0), `absent` (exit 1 + `No such keg`), or `unknown` (everything else). |
| Execution guard | `internal/execution/runner.go:72-85` | `isInstalledBrewFormulaStep` skips dispatch for `PackagePresenceInstalled`; `isUnknownBrewFormulaStep` returns `StepStatusFailed` without installer call. |
| CLI composition | `cmd/dbootstrap/main.go:150-161` | In confirmed modes, runs `BrewFormulaDetector.Detect()` before `buildApplyRunner()`, decorates an execution-plan copy via `ApplyBrewFormulaPresence`. |
| Planning types | `internal/planning/types.go:104-112` | `PackagePresence` is a transient `PlanStep` field with values `Unchecked`, `Installed`, `Absent`, `Unknown`. |

### APT installer (exists — no idempotency guard)

| Layer | File | Current behavior |
|-------|------|-----------------|
| APT installer | `internal/execution/apt_installer.go` | `AptInstaller.Install()` validates metadata, checks `apt-get`/`sudo` availability, then dispatches `apt-get install -y -- <pkg>` (or `sudo apt-get install -y -- <pkg>`). NO pre-flight check exists. Every confirmed eligible APT step always reaches `apt-get install`. |
| Runner | `internal/execution/runner.go` | `isEligibleBrewFormulaStep` guards apply ONLY for `Provider == "brew"`. APT packages with `PackagePresenceInstalled` are **not** recognized by the runner: they fall through to the installer dispatch path (confirmed by `TestRunnerIgnoresPackagePresenceForInvalidBrewPackage` at line 188-202). |
| CLI composition | `cmd/dbootstrap/main.go:276-284` | `planHasEligibleBrewFormulaPackage` checks only `Provider == "brew"`. There is no `planHasEligibleAptPackage` function. APT packages never trigger pre-flight detection. |
| Non-Linux guard | `internal/execution/apt_installer.go:94-108` | `nonLinuxAptInstaller` returns `StepStatusFailed` with `AptExecutionUnsupportedOS` without running commands. |

### Command execution seams already available for reuse

- `CommandRunner` interface (`internal/execution/command.go:59-61`): `RunCommand(context.Context, CommandRequest) CommandResult` — already used by both `AptInstaller` and `BrewFormulaDetector`.
- `CommandExists` type (`internal/execution/homebrew_bootstrap.go:12`): `func(name string) bool` — already injected into `AptInstaller`.
- `OSCommandRunner` (`internal/execution/os_command_runner.go`): real process execution honoring `CommandRequest.Timeout`.
- `aptCommandExistsOnPath` (`cmd/dbootstrap/main.go:59-62`): CLI-level `exec.LookPath` wrapper for `apt-get`/`sudo`/`dpkg-query`.

### Test patterns already established

- `recordingCommandRunner` in `internal/state/brew_formula_detector_test.go`: records requests, returns configured results.
- `sequenceCommandRunner` in `cmd/dbootstrap/main_test.go`: records requests, returns results in order.
- CLI stub functions: `stubEnvironmentFacts`, `stubBrewCommandExists`, `stubExecutionFactories`.
- Runner tests in `internal/execution/runner_test.go`: prove installed-skip, unknown-fail, absent-dispatch, and cross-provider non-interference.

## Affected Areas

- **`internal/state/apt_package_detector.go`** (NEW) — Debian package-presence detector using `dpkg-query` through injected seams.
- **`internal/state/apt_package_detector_test.go`** (NEW) — Table-driven unit tests covering eligibility, argv/timeout, all classifications, and no-probe cases.
- **`internal/execution/runner.go:72-85`** — Add `isInstalledAptPackageStep` and `isUnknownAptPackageStep` guards; extend `isEligibleAptPackageStep` to validate `Provider == "apt"` and trimmed package metadata.
- **`internal/execution/runner_test.go`** — Prove installed skip, unknown failed/no dispatch, absent dispatch, revalidation (e.g., manually injected `PackagePresenceInstalled` on a non-package, non-apt, or blank-package step must not skip), and continued ordered execution.
- **`cmd/dbootstrap/main.go`** — Add `planHasEligibleAptPackage()` check; add APT package detection block AFTER the existing Brew detection block (only for confirmed Linux execution); decorate execution plan with detected APT presence.
- **`cmd/dbootstrap/main_test.go`** — CLI composition tests for `apply`/`bootstrap` confirmed-mode APT detection, safe-mode non-probing, mixed-plan order, Linux gating, and zero install calls for installed/unknown.
- **`openspec/changes/apt-dpkg-package-idempotency/specs/`** — Delta specs for `apt-package-installer`, `installation-state`, and `execution-contracts` as needed.
- **`internal/planning/types.go`** — No change needed. The `PackagePresence` field and constants (`Unchecked`, `Installed`, `Absent`, `Unknown`) are already generic. The comment "Brew formula presence result" should be broadened to reflect dual-provider use.
- **`internal/execution/apt_installer.go`** — No change needed. Detection is a separate concern, not an installer responsibility.

Files NOT affected: catalog schema, TOML parser, planner, `AptInstaller` implementation, `provider_aware_installer.go`, dotfiles, Homebrew installer.

## Approaches

### Approach 1: Mirror Brew Formula Detection Pattern (RECOMMENDED)

Create `AptPackageDetector` in `internal/state/` mirroring `BrewFormulaDetector`, reuse the existing `PackagePresence` transient field, add `isInstalledAptPackageStep` / `isUnknownAptPackageStep` guards in the runner, and wire detection in `cmd/dbootstrap/main.go` after the existing Brew detection block.

| Pros | Cons |
|------|------|
| Consistent with existing Brew implementation — familiar pattern for reviewers | Requires coordination between state detector, runner guards, and CLI wiring |
| Reuses `PackagePresence` type — no new types or fields | `PackagePresence` field doc comment is Brew-biased (trivial fix) |
| Detection is a separate, testable unit before execution dispatch | |
| Clear separation: state package owns detection, execution owns guard | |
| Runner guard revalidates eligibility (provider, kind, package) — defense against malformed manually constructed plans | |

### Approach 2: Inline pre-flight in AptInstaller

Add `dpkg-query` check inside `AptInstaller.Install()` before attempting `apt-get install`. Use the existing `exists` seam to check `dpkg-query` availability, and use the existing `runner` seam to query.

| Pros | Cons |
|------|------|
| No runner changes needed | Detection happens at install time — requires constructing the installer (and runner) even for a "skip" |
| Self-contained change | Violates the Brew precedent of detection before dispatch |
| | Harder to test independently from the installer |
| | Mixes detection concern with installation concern |
| | The runner would still invoke `inst.Install()` for an APT package with `PackagePresenceInstalled` — the skip happens inside the installer, not at the runner gate |
| | Inconsistent with the Brew pattern: two different code paths for the same semantic outcome |

### Approach 3: Broaden `isEligibleBrewFormulaStep` to `isEligibleProviderPackageStep`

Rename and extend the existing Brew eligibility/guard functions to handle both `brew` and `apt` providers with a shared dispatch table.

| Pros | Cons |
|------|------|
| Less code duplication in the runner | Couples Brew and APT guard logic — a change to one provider's eligibility affects the other's code path |
| Fewer guard functions | Detection logic is provider-specific (`brew list --formula` vs `dpkg-query --show`); a unified function would still branch internally |
| | The Brew detector already uses separate eligibility (`isEligibleBrewFormulaStep`) and classification (`classifyBrewFormulaResult`) — unifying guards would require restructuring both |
| | Higher regression risk for existing Brew behavior |

### Complexity Comparison

| Approach | New files | Modified files | Changed lines (est.) | Test surface |
|----------|-----------|---------------|---------------------|-------------|
| 1 (Mirror Brew) | 2 | 5 | 150–250 | Full coverage of all classification states, Linux gating, cross-provider isolation |
| 2 (Inline in installer) | 0 | 2 | 80–130 | Only installer-level tests; no runner or CLI composition coverage |
| 3 (Unify guards) | 1 | 4 | 180–300 | Must re-prove all Brew states plus new APT states; higher regression surface |

## Recommendation

**Approach 1: Mirror Brew Formula Detection Pattern.** It's the cleanest, most consistent, and follows the existing architecture without adding unnecessary complexity. The Brew precedent already established the contract: detection in `internal/state`, classification as `installed`/`absent`/`unknown`, plan decoration, and runner guards. APT should follow the same contract.

### Specific design decisions

1. **`dpkg-query` command vector**: `dpkg-query --show --showformat=${Status} <package>` — executable `dpkg-query`, args `["--show", "--showformat=${Status}", packageName]`. The `--showformat` output for an installed package produces the exact string `install ok installed`; for a known-not-installed package it produces `unknown ok not-installed` or `deinstall ok config-files`; for an unknown package, stderr contains `no packages found matching`. This gives us a structured stable output to parse.

2. **Classifier logic**:
   - **Installed**: `CommandStatusSucceeded` AND exit code 0 AND stdout contains `install ok installed` for the exact package line.
   - **Absent**: `CommandStatusSucceeded` AND exit code 0 AND stdout does NOT contain `install ok installed` — `dpkg-query` exits 0 for both installed and known-but-not-installed packages; only an unknown package exits non-zero. This means we must parse stdout, not rely on exit code alone.
   - **Unknown**: everything else — missing `dpkg-query`, nil runner, timeout, runner error, exit code non-zero (unknown package, `no packages found matching`), malformed success result, empty stdout.

3. **Eligibility**: Only `ResourceKindPackage` with `Install.Provider == "apt"` and a trimmed, non-empty `Install.Package`. No `tool` or `runtime` resources. No `Presence.Name` substitution.

4. **Timeouts**: Use a fixed constant `aptPackagePresenceTimeout` (e.g., `30 * time.Second`) in `internal/state`. No retry, no fallback.

5. **Linux gating**: Detection runs only when `facts.OS == "linux"`. The `nonLinuxAptInstaller` already handles non-Linux APT steps; detection on non-Linux would be pointless and adds an unnecessary command existence check. Confirmed mode on non-Linux should skip APT detection entirely.

6. **`dpkg-query` availability**: Check via injected `CommandExists("dpkg-query")`. If unavailable, classify as `unknown` — do not fall back to `apt-cache`, `apt list`, or any other command.

7. **No `sudo` for detection**: `dpkg-query --show` is read-only and does not require privileges. The `--sudo` flag must not affect detection.

8. **Runner guard revalidation**: `isEligibleAptPackageStep` must independently revalidate provider, kind, and package (not just trust `PackagePresence`). This prevents a malformed manually constructed plan from gaining a false skip.

## Risks

- **`dpkg-query` output format stability**: `dpkg-query --showformat` output has been stable across Debian/Ubuntu releases for decades, but must be tested against the specific format: the status field uses three space-separated words (`install ok installed`). If a future release changes this format, detection safely becomes `unknown` (the classifier requires an exact string match).
- **Virtual/provided packages**: `dpkg-query` only reports on explicitly installed packages. A virtual package (e.g., `awk` provided by `gawk`) would not be detected as installed. This is correct behavior for this slice: the catalog `Install.Package` is an install target, not a virtual name.
- **Package name injection**: The package name is passed as a separate argument in the `CommandRequest.Args` vector, not interpolated into a shell string. The `--showformat` flag value is hardcoded. No shell metacharacter risk.
- **Cross-provider interference**: Must ensure APT guards don't accidentally skip Brew packages or vice versa. The eligibility checks (`Provider == "apt"` vs `Provider == "brew"`) are mutually exclusive by design.
- **Runner guard ordering**: `isInstalledAptPackageStep` must be checked AFTER `isAlreadyInstalledCommandStep` and the Brew guards, preserving the existing skip precedence. The guard must use `||` (any matching guard skips) for `isInstalledBrewFormulaStep` / `isInstalledAptPackageStep`, and `||` for `isUnknownBrewFormulaStep` / `isUnknownAptPackageStep` — consistent with the current `step.PackagePresence`-based routing.

## Ready for Proposal

**Yes.** The architecture is well-understood, the Brew precedent provides a clear pattern, and the `dpkg-query` command seam is a safe, read-only detection mechanism. Proceed to `sdd-propose` with the following scope boundary:

**In scope**: `AptPackageDetector`, `dpkg-query` classifier, runner guards, CLI wiring, strict-TDD unit and composition tests.
**Out of scope**: Version reconciliation, virtual packages, multi-arch resolution, `apt-cache` fallback, `apt-get` for detection, `sudo` for detection, retries, cask/APT conflation, catalog schema changes, installer changes.
