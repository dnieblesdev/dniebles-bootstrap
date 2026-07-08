# Design: Wire Brew Apply

## Technical Approach

Wire real Homebrew package execution only at the CLI composition root and only when `dbootstrap apply --yes` is selected. Default apply and `--dry-run` continue to build the plan, run kind noops, and render advisory Homebrew bootstrap guidance without using `HomebrewInstaller` or `OSCommandRunner`.

This also fixes Homebrew bootstrap guidance before any real mutation path is wired: rendered guidance must remove the executable remote-script one-liner and point users to official Homebrew documentation/manual review instead.

## Architecture Decisions

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Branch runner composition by `applyMode` | Keeps the safety gate centralized; requires focused CLI tests. | Use noop installers for default/dry-run, and real brew-capable installers only for confirmed mode. |
| Register `HomebrewInstaller` directly for `tool`/`package` | Unsupported providers fail, which avoids false success but makes non-brew resources look like errors. | Add a small provider-aware adapter that delegates brew metadata to `HomebrewInstaller` and otherwise returns `not_implemented`. |
| Reuse `OSCommandRunner` and `BrewCommandExists` | Real process execution exists only behind `--yes`; no shell path is introduced. | Confirmed brew installs use `NewOSCommandRunner()` and the existing `brewCommandExists` seam; no raw command metadata, `sh -c`, or pipelines. |
| Check missing brew before real installers | Adds one composition branch but avoids misreporting prerequisite absence as package failure. | If `apply --yes` selects brew-backed steps and `brew` is missing, skip target package installation, do not invoke `HomebrewInstaller`, and make Homebrew bootstrap guidance the primary report. |
| Replace bootstrap one-liner | Less copy-paste convenience; materially safer for first-run guidance. | `AppendHomebrewBootstrap` renders official docs/manual review wording only, with no executable remote-script command. |

## Data Flow

Default / dry-run:

    runApply ──→ buildPlan ──→ NewRunner(NoopForKind...) ──→ renderExecutionReport
                         └──→ AppendHomebrewBootstrap(advisory only)

Confirmed `--yes`:

    runApply ──→ buildPlan ──→ confirmed runner
                              ├── brew missing + brew-backed steps ──→ bootstrap guidance + skipped target installs
                              ├── brew present + brew-backed tool/package ──→ HomebrewInstaller ──→ OSCommandRunner
                              └── non-brew/other kinds ──→ not_implemented noop result

The missing-brew branch must happen before registering/running `HomebrewInstaller` for brew-backed `tool`/`package` steps, or must otherwise prove `HomebrewInstaller.Install` is not invoked. The report should describe the Homebrew prerequisite first, not a failed target package install.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modify | Rename confirmed mode label/flag help, introduce runner construction by mode, and wire `HomebrewInstaller` only for `--yes`. |
| `cmd/dbootstrap/render.go` | Modify | Render confirmed-mode wording that warns real `brew install` may run; keep noop wording clear for safe modes. |
| `cmd/dbootstrap/main_test.go` | Modify | Cover default/dry-run noops, confirmed brew wiring with fake seams, missing brew advisory-first behavior, and non-brew provider behavior. |
| `cmd/dbootstrap/render_test.go` | Modify | Prove rendered bootstrap guidance contains official docs/manual review wording and no remote-script one-liner. |
| `internal/execution/homebrew_bootstrap.go` | Modify | Replace the current executable Homebrew install instruction with non-executable official docs/manual review guidance. |
| `internal/execution/homebrew_bootstrap_test.go` | Modify | Assert bootstrap actions are advisory-only and never include `/bin/bash`, `curl`, `sh -c`, pipes, or raw install commands. |
| `internal/execution/provider_aware_installer.go` | Create | Adapter that supports a kind, delegates only when `Install.Provider == "brew"`, otherwise returns `not_implemented`. |
| `internal/execution/provider_aware_installer_test.go` | Create | Unit tests for brew delegation and non-brew/no-metadata noop behavior. |

## Interfaces / Contracts

No planning contract changes. The new execution adapter implements the existing `Installer` interface:

```go
func BrewOnlyInstaller(kind planning.ResourceKind, brew Installer) Installer
```

The adapter MUST NOT execute commands itself. It only routes provider metadata:
- `Install.Provider == "brew"` delegates to the supplied brew installer.
- missing or non-brew metadata returns `StepStatusNotImplemented`, not success.

Bootstrap guidance contract:
- Render `https://brew.sh/` or official Homebrew documentation wording plus manual review instructions.
- Do not render copy-paste remote script commands, shell snippets, raw command metadata, `sh -c`, or pipelines.
- When brew is missing, target brew installs are skipped/not invoked and the manual Homebrew prerequisite is the primary report item.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Provider-aware adapter routing | Fake installer proves brew delegation; non-brew metadata returns `not_implemented`. |
| Unit | Bootstrap guidance safety | Assert official-docs/manual wording and absence of executable remote-script instructions. |
| CLI wiring | Safety mode composition | Inject fake command runner/existence seams or factory seam so default/dry-run cannot call real execution, missing brew does not invoke `HomebrewInstaller`, and `--yes` can call only brew-backed steps when brew exists. |
| Integration | Existing command runner behavior | Keep direct OS process tests outside CLI wiring; do not invoke real `brew` from CLI tests. |

## Migration / Rollout

No data migration required. Rollout is gated by the existing `--yes` flag; rollback reverts runner composition and the adapter while leaving isolated Homebrew installer code intact.

## Open Questions

- None.
