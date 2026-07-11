## Exploration: apt-provider

### Current State
The planning core is format-agnostic and carries inert `InstallMetadata{Provider, Package}` through `PlanStep`. The TOML adapter accepts any non-empty provider/package pair and does not infer execution. The current default catalog uses Homebrew for `tool:git`, `package:ripgrep`, and `package:jq`; the only other provider metadata is `asdf` for `runtime:go`.

Execution is keyed by resource kind, not provider: `Runner` registers one `Installer` per kind and dispatches steps sequentially. `HomebrewInstaller` validates `provider == "brew"`, checks `brew` through an injected `CommandExists`, then sends `CommandRequest{Executable:"brew", Args:["install", package]}` through `CommandRunner`. `BrewOnlyInstaller` prevents non-brew metadata from reaching that installer. `OSCommandRunner` uses `exec.CommandContext` without a shell, while `NoopCommandRunner` preserves the request as `not_run`.

The CLI composition root is the safety boundary. Default `apply` and `--dry-run` construct noop runners; only `apply --yes` can compose real installers. Confirmed composition currently detects whether the plan contains brew metadata, probes `brew`, and wires Homebrew plus selected dotfiles. Missing Homebrew is a failed/skipped execution condition plus an advisory, non-executable manual bootstrap action. There is no privilege escalation, APT bootstrap, package-manager selection field, or package-aware presence detection.

Environment detection already supplies `OS`, `Arch`, `Distro`, and `WSL` from injectable runtime, environment, and file seams. It does not detect executables or package managers. Installation-state detection only treats tool/runtime names as `exec.LookPath` checks; package resources are not considered already installed. Execution results expose installed/failed/skipped/not-implemented, command errors, and dotfile rollback details. The runner continues after failures. Package installation has no rollback contract because a package-manager transaction may partially mutate the host.

### Affected Areas
- `internal/execution/installer.go`, `runner.go`, `provider_aware_installer.go` — existing kind-based dispatch and provider gate; APT must fit without making the runner provider-aware.
- `internal/execution/homebrew_installer.go` and `homebrew_installer_test.go` — closest implementation and seam pattern for an additive `AptInstaller`.
- `internal/execution/command.go`, `os_command_runner.go`, `noop_command_runner.go` — explicit executable-plus-args contract, bounded `CommandRequest.Timeout`, and the mutation/dry-run boundary; APT must not use a shell.
- `cmd/dbootstrap/main.go` — confirmed-only composition, command-presence seams, plan facts propagation, and the current Homebrew bootstrap hook.
- `internal/environment/detector.go`, `internal/planning/types.go` — existing Linux/distro facts; no new package-manager fact is currently required for the smallest slice.
- `internal/state/detector.go` — package presence is not detected, so an initial APT slice must not claim idempotence for package resources.
- `internal/catalog/toml/{schema.go,validate.go,catalog.go}` and `catalog/bootstrap.toml` — structured provider metadata already migrates without schema changes; catalog changes should be opt-in and additive.
- `internal/execution/types.go`, `render.go`, `homebrew_bootstrap.go` — current failure/report/manual-action vocabulary and the explicit absence of package rollback reporting.
- Tests in `internal/execution`, `internal/environment`, `internal/catalog/toml`, and `cmd/dbootstrap` — injectable seams exist, but `parseApplyFlags` and confirmed composition have limited direct coverage.

### Approaches
1. **Additive APT installer parallel to Homebrew (recommended)** — Add `AptInstaller` for each supported kind, a provider-aware APT wrapper/factory, and confirmed CLI composition gated by Linux plus an injected `apt-get` presence check.
   - Pros: smallest blast radius; preserves kind-based `Runner`; reuses `CommandRunner`, structured metadata, noop modes, and existing error vocabulary; makes the command exact and auditable.
   - Cons: duplicate small installer validation code; package presence remains incomplete; direct `apt-get` may fail when the process lacks required privileges.
   - Effort: Medium

2. **Generalize installers into a provider registry** — Register `(resource kind, provider)` pairs and let the runner select by both metadata dimensions.
   - Pros: scales better for many providers and avoids one wrapper per provider/kind.
   - Cons: changes the stable runner contract and composition semantics; larger review surface; unnecessary before a second real provider proves the need.
   - Effort: High

3. **Use a generic command-from-metadata installer** — Store executable/arguments in catalog data and interpret them generically.
   - Pros: little provider-specific code.
   - Cons: weakens the safety model, expands catalog authority into executable behavior, and makes privilege/bootstrap mistakes easier; conflicts with the explicit Homebrew contract.
   - Effort: Medium, but unsafe

### Recommendation
Use approach 1. Introduce an explicit APT provider constant/installer that accepts only `Provider == "apt"` and a trimmed package name that is neither empty nor prefixed by `-`. It sends `CommandRequest{Executable:"apt-get", Args:[]string{"install", "-y", "--", package}, Timeout: 10 * time.Minute}` for `apply --yes`, or `CommandRequest{Executable:"sudo", Args:[]string{"apt-get", "install", "-y", "--", package}, Timeout: 10 * time.Minute}` only for `apply --yes --sudo`. `--` is an argument delimiter, not shell escaping: it prevents custom catalog metadata from being interpreted as an APT option. Do not add automatic privilege fallback, `pkexec`, shell strings, `apt update`, repository changes, or bootstrap commands. A direct command that lacks privilege fails without hidden escalation.

Pass detected environment facts into confirmed runner composition and require `facts.OS == "linux"` plus `apt-get` presence before wiring APT installers. Keep absence of `apt-get` as a structured non-success result, not an automatic install/manual bootstrap workflow. APT timeout is ten minutes through the existing `CommandRequest`/CLI composition seam: this bounds a process held by package-manager locks while allowing ordinary package transactions; a timeout is a structured failed result with no retry or rollback claim. Default and dry-run modes must continue to use noop installers and must not probe or execute mutating APT commands. Keep the Runner keyed by resource kind and use provider-aware wrappers so brew and apt metadata cannot cross delegates; preserve Homebrew's existing cross-platform execution eligibility.

The first catalog migration should be additive and opt-in: preserve all current Homebrew targets, then add or convert one explicitly Linux/Debian-family APT target only after the provider path is proven. Prefer a separate fixture/custom catalog in the first implementation slice; changing the default catalog from brew to apt would alter the default confirmed mutation surface and should be a separately approved catalog change. No planning-schema change is needed for provider/package metadata.

### Risks
- `apt-get install` can partially change the host before returning failure or timing out; the current package execution contract has no rollback mechanism. Report the structured command outcome and explicitly state rollback was not attempted rather than implying recovery.
- APT generally needs root privileges, but silently prepending `sudo` would violate the safety boundary and create an interactive/escalation side effect. Direct execution may fail by design; sudo is available only by explicit `--yes --sudo`.
- The current state detector does not detect installed packages, so an APT package target may be attempted even when already installed. Do not broaden this slice into dpkg-query/state reconciliation.
- Runner registration remains one installer per resource kind; duplicate registrations must not be introduced. Provider wrappers must reject the other provider before invoking a delegate.
- Existing CLI reporting renders status and error text but not all `CommandResult` fields, especially stderr. Preserve the current structured result contract and decide separately whether richer command diagnostics are required.

### Ready for Proposal
Yes — propose a minimal additive APT installer plus confirmed-only CLI wiring and focused contract tests. Explicit non-goals for the proposal: automatic privilege escalation/fallback, APT/Homebrew bootstrap, `apt update` or repository management, shell execution, retries/concurrency, rollback/transaction orchestration, dpkg package-presence detection, provider-registry redesign, fallback provider selection, and default-catalog migration to APT in the same first slice.
