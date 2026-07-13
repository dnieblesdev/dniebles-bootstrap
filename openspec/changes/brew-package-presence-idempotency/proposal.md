# Brew Package Presence Idempotency

## Intent

Add the first provider-specific package-presence slice: in confirmed execution, avoid a redundant Homebrew formula installation only when a read-only Brew query positively proves the configured formula is installed. This extends the archived command-presence boundary without changing it: tools and runtimes retain their existing command-presence behavior, while package detection is explicitly limited to Brew formulas.

## Problem

Brew-backed package resources are currently eligible for installer dispatch on every confirmed rerun because the existing idempotency guard intentionally covers only reliable command presence for tools and runtimes. This produces unnecessary mutation attempts and noisy reruns even when a Brew formula is already installed.

## Goals

- Detect installed Brew **formulas** through an injected, read-only command query using `brew list --formula <InstallMetadata.Package>`.
- Use the provider package/formula identity, never the resource ID or `Presence.Name`.
- In confirmed `apply` and `bootstrap`, mark a positively detected Brew package as `already_installed`, preserve plan order, report it unchanged with explicit no-mutation wording, and make no install call for that step.
- Keep planning pure and preserve default and `--dry-run` modes as non-probing and non-mutating.
- Cover installed, absent, unavailable-manager, query-failure/timeout, and mixed-plan behavior with injected lookup and command-runner seams under strict TDD (`go test ./...`).

## Explicit query-failure policy

Only a successful Brew query is proof of installed state. A missing `brew` executable, non-success query, timeout, or runner error is **unknown**, not absent and not installed. The affected package must be reported as an attention/failure outcome and must not invoke its installer during that run. This avoids authorizing a host mutation from an inconclusive read. Other plan steps retain their existing ordered, continued-execution behavior.

A successful non-zero Brew result that is explicitly classified as “formula absent” remains eligible for the existing installer. The design/spec phase must define the command-result classification precisely so operational query failures cannot be misclassified as absence.

## Non-goals

- APT/dpkg or any other provider detection.
- Casks, Brew prefix discovery, multi-Homebrew handling, virtual packages, or generic provider registries.
- Version/latest reconciliation, executable health, PATH/link/configuration verification, or dotfile convergence.
- Catalog schema redesign; the existing provider package metadata is the formula authority.
- Retries, fallback queries, shell-string execution, `brew update`, `brew install` during detection, sudo detection, or bootstrap acquisition.
- Rewriting archived idempotency or APT-provider history.

## Safety constraints

- Detection is read-only, bounded by a timeout, and uses the existing injected command boundary with an executable plus argument vector; it must not use a shell.
- Detection runs only in confirmed modes before installer dispatch. Default and `--dry-run` must not look up or execute Brew.
- Only Brew-backed package resources with valid formula metadata are eligible. Formula and cask namespaces must not be conflated.
- An installed formula is not evidence of a required version, a working executable, configuration correctness, or dotfile convergence.
- `--sudo` remains an installation-mode concern and must not affect detection.

## Affected areas

- `internal/state` package-presence detection and state representation.
- Existing command lookup/runner seams used for a bounded Brew query.
- `cmd/dbootstrap` confirmed-mode composition and ordered reporting for `apply`/`bootstrap`.
- Focused state, command, and CLI composition tests.
- Additive OpenSpec deltas for installation state, apply safety, and execution behavior as required.

## Acceptance criteria

- Confirmed execution queries an eligible Brew package with exactly `brew list --formula <InstallMetadata.Package>` through an injected runner.
- A successful installed query causes an ordered `already_installed`/unchanged result, explicitly states that no mutation was attempted, and makes no installer call.
- A positively absent formula remains eligible for the existing Brew installer.
- Missing Brew, query failure, timeout, or ambiguous result is visibly unknown/attention, never `already_installed` or absent, and makes no installer call for that package.
- Default and `--dry-run` modes do not probe Brew and preserve current non-mutating behavior.
- Non-Brew packages, tools/runtimes, dotfiles, casks, and unsupported resources retain their current semantics.
- Tests use fakes, assert exact argv, timeout, query ordering, and zero host mutation; focused tests and `go test ./...` pass.

## Risks

- **False absence from an operational failure:** mitigated by the explicit unknown/no-install policy and exact result classification.
- **Wrong formula identity:** mitigated by requiring `InstallMetadata.Package` and tests where it differs from resource and command names.
- **Accidental expansion into convergence:** mitigated by provider/formula-only eligibility and explicit exclusions.
- **Safety-mode regression:** mitigated by composition tests proving no Brew lookup or runner use outside confirmed modes.

## Rollback

Revert the Brew presence detector, confirmed-mode wiring, and its additive specs/tests together. No migration, catalog transformation, or persistent state cleanup is required; prior installer-dispatch behavior resumes.

## Success criteria

A confirmed rerun skips only Brew package installs that it can positively prove are installed, while inconclusive checks never cause either a false idempotency claim or an untrusted installer mutation. Existing safe-mode and non-package behavior remains unchanged.

## Delivery forecast

- Estimated impact: 5–8 files and 150–260 changed lines.
- Review risk: standard reliability/resilience; one focused review lens is expected.
- Review budget: within the 400-line budget; no chained PR forecast.
- Delivery: one additive provider-specific slice, with APT and all broader reconciliation deferred.
