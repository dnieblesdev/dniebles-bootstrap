# Design: APT Provider

## Technical Approach

Add `AptInstaller` beside `HomebrewInstaller`, selected through a fixed brew-or-APT kind adapter. `execution.Runner` dispatches by kind; the adapter dispatches provider metadata. Confirmed Linux creates `apt-get install -y -- <package>` for `apply --yes`, or `sudo apt-get install -y -- <package>` only for explicit `apply --yes --sudo`, with a ten-minute `CommandRequest` timeout. Default and dry-run compose no APT path or probes.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| APT boundary | `AptInstaller` for `tool` and `package` kinds | Shell commands; planner changes | Matches `HomebrewInstaller` and keeps planning pure. |
| Privilege mode | Parse `--sudo` only with `--yes`; pass it to APT composition | Automatic sudo fallback; hidden escalation | Makes the mutating vector chosen and auditable by the user. |
| Package safety | Trim, reject empty and `-`-prefixed package metadata; insert `--` before the package | Accept arbitrary metadata; shell escaping | `--` prevents APT option injection from a custom catalog. It is not shell escaping; execution remains executable-plus-args. |
| Timeout | `CommandRequest.Timeout = 10 * time.Minute` for either APT vector | Unbounded command; short generic timeout | Bounds package-manager lock waits while allowing normal package transactions through the existing request/composition seam. |
| Provider composition | Fixed brew-or-APT kind adapter with APT delegate | Two Runner entries; provider registry | One installer exists per kind; avoids a registry redesign. |
| Non-Linux rejection | `buildApplyRunner` selects `nonLinuxAptInstaller` as the APT delegate when an APT step is selected and `facts.OS != "linux"` | Leaving `BrewOnlyInstaller`; probing host tools | The selected APT branch must fail, not become `not_implemented`, while making zero APT/sudo lookups or commands. |
| Failure reporting | `StepStatusFailed`, typed error, and rendered command outcome; confirmed apply exits non-zero | Retry/fallback/rollback | Preserves current report/exit mechanics and never claims mutation was undone. |
| Test proof | `t.TempDir()` custom TOML catalog | Default catalog edit; real APT integration | Proves an opt-in target through seams and preserves defaults. |

## Data Flow

```text
apply flags -> {mode, sudo} -> buildApplyRunner(options, facts, plan)
custom catalog -> planning.Plan --------^       | default/dry-run: noop, no APT seam
                                                 | confirmed Linux: provider gate -> AptInstaller
                                                 |   -> exists("apt-get") [and exists("sudo") when selected]
                                                 |   -> CommandRunner (10m timeout) -> StepResult
                                                 ` confirmed non-Linux: nonLinuxAptInstaller -> failed result
Runner (by kind) -> ExecutionReport -> confirmed failed result -> exitFailure
```

`nonLinuxAptInstaller.Install` returns `{Status: StepStatusFailed, Err: *AptExecutionError{Reason: AptExecutionUnsupportedOS, OS: facts.OS, CommandStatus: CommandStatusNotRun, ExitCode: 1}}`. It holds no `CommandExists`/`CommandRunner`, so makes no `apt-get`/`sudo` lookup or command. The non-zero outcome is the confirmed CLI result via `hasFailedExecutionResult`, not a launched process result.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/execution/apt_installer.go` | Create | Provider-gated direct/sudo vectors, availability checks, typed APT errors, and non-Linux rejecting adapter. |
| `internal/execution/apt_installer_test.go` | Create | Table-driven vector, validation, availability, rejection, and command-outcome coverage. |
| `internal/execution/provider_aware_installer.go` | Modify | Add fixed brew-or-APT routing; retain kind dispatch and brew-only behavior where APT is not composed. |
| `internal/execution/provider_aware_installer_test.go` | Modify | Prove provider routing and no delegate call on rejection. |
| `cmd/dbootstrap/main.go` | Modify | Parse `--sudo`, reject it outside `--yes`, and inject facts/APT/sudo seams into confirmed composition. |
| `cmd/dbootstrap/main_test.go` | Modify | Prove custom-catalog vectors, CLI non-zero failures, and non-probing modes. |
| `catalog/bootstrap.toml` | No change | Default catalog remains untouched. |

## Interfaces / Contracts

```go
type AptExecutionReason string
const AptExecutionUnsupportedOS AptExecutionReason = "unsupported_os"
type AptExecutionError struct {
    Reason AptExecutionReason
    OS string
    CommandStatus CommandStatus
    ExitCode int
}

const aptCommandTimeout = 10 * time.Minute
// Direct: CommandRequest{"apt-get", []string{"install", "-y", "--", package}, Timeout: aptCommandTimeout}
// Sudo:   CommandRequest{"sudo", []string{"apt-get", "install", "-y", "--", package}, Timeout: aptCommandTimeout}
```

`AptInstaller` accepts only `metadata.Provider == "apt"` and a trimmed package that is non-empty and does not begin with `-`. It checks `apt-get` after validation and `sudo` only for the sudo vector. Missing executables, invalid metadata, failed/timed-out commands, and non-Linux rejection yield `StepStatusFailed`; timeout preserves `CommandStatusTimedOut`, triggers no retry, and makes no rollback claim. No path uses `PresenceMetadata`.

CLI seams mirror the existing Homebrew seams: an `aptCommandExists` function, an APT-installer factory, and the shared injectable `newOSCommandRunner`. The runner is lazily constructed only when confirmed composition has an eligible brew, APT, or dotfile path. No new public provider registry is introduced.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Direct/sudo vectors include `-y --`; empty and `-`-prefixed metadata; missing `apt-get`/`sudo` | Table-driven fakes record calls and requests; assert no command after rejected validation/probe. |
| Unit | Non-Linux adapter | Assert `StepStatusFailed`, `*AptExecutionError`, `CommandStatusNotRun`, exit code 1, and zero availability/runner calls. |
| Unit | Ten-minute request timeout; command failure/timeout and partial mutation | Return failed/timed-out command results; assert the bounded request, structured failure, no retry, and no rollback claim. |
| CLI integration | `--yes`, `--yes --sudo`, invalid sudo flag, and non-Linux exit | `t.TempDir()` catalog plus stubbed facts, availability, installer factory, and recording runner; non-Linux asserts failed/non-zero and zero apt/sudo probe/command calls. |
| Safety regression | Default and `--dry-run` | Failing APT/sudo probe and runner seams prove no probe/no-op; default catalog remains unchanged. |

## Migration / Rollout

No migration required. The change is additive and reachable only through an explicit custom catalog APT target with `apply --yes` (and optionally `--sudo`). Reverting the adapter and composition removes the path; no default catalog data changes.

## Open Questions

- [ ] None.
