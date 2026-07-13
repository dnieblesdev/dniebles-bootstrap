# Exploration: package-presence-idempotency

## Recommendation

**Split, then proceed with a narrow provider-by-provider slice.** Package-presence detection is valuable: the current confirmed idempotency guard intentionally skips only command-present tools/runtimes, while package resources with catalog presence metadata are always eligible for an install attempt. That leaves repeated Brew/APT package installs noisy and needlessly dependent on installer mutation paths.

Do not implement Brew and APT package detection as one undifferentiated change. Start with a shared, read-only detection contract plus one provider (Brew is the lower-risk first slice), then add APT after its Debian-state semantics and privilege boundaries are separately accepted.

## Current gap and evidence

- `planning.PresenceMetadata` already carries `Kind` and `Name`; the catalog fixture declares `command_exists` presence for `ripgrep`/`jq`, but `internal/state.Detector` intentionally detects only `tool` and `runtime` kinds.
- `BuildPlan` already accepts caller-supplied `InstallationState` and gives `already_installed` precedence after environment matching. `PlanStep.Status` now carries that planning result into execution.
- Confirmed execution skips only `already_installed` tools/runtimes with valid command-presence metadata. Package and dotfile steps continue through normal installer dispatch by design.
- `cmd/dbootstrap` is the composition boundary: default and `--dry-run` use no-op execution and clear planning status; only confirmed modes may compose real installers. Detection is read-only and occurs before planning.
- The current catalog separates provider package metadata (`InstallMetadata.Package`) from presence metadata (`Presence.Name`). This is the key distinction: a package-manager query must use the provider package/formula identity, not the resource ID or configured executable name.
- Archived apply-idempotency exploration and verification explicitly excluded package-manager queries, version reconciliation, and package/dotfile convergence. Archived APT-provider contracts explicitly prohibit package presence detection in that provider slice. Those historical boundaries should remain immutable; this is new follow-up scope.

## Candidate mechanisms

### Homebrew

Use a read-only query through the existing `CommandRunner` seam, with an explicit argv vector such as:

```text
brew list --formula <formula>
```

A successful exit means the formula is installed in the current Homebrew prefix; non-zero means absent or not queryable. A version-aware alternative (`brew list --formula --versions <formula>`) should not be used for the first slice because it adds parsing and still does not define a desired version contract. If casks are eventually supported, they need a separate metadata kind/query; do not silently treat a formula as a cask.

The detector should validate provider metadata before querying, check `brew` availability through an injected lookup, and represent manager-unavailable/query-failed as absent or an explicit attention/error state only after the status contract is designed. It must never invoke `brew install`, `brew update`, or shell text.

### APT/Debian

Use a read-only `dpkg-query` query, not `apt-get`, for installed state. A practical vector is:

```text
dpkg-query --show --showformat=${Status}\n <package>
```

Then require the output to contain the exact state `install ok installed` (prefer a structured parser over exit-code-only detection). An alternative is `dpkg-query -W -f=${Status} <package>`; either form must be fixed in code and tested as an argv vector, not assembled as a shell command.

`apt-cache policy` is not an installed-state authority and should not be used as the primary detector. `apt-get` must never be used for detection because it can resolve/update package state and carries mutation risk. The detector should first check `dpkg-query` availability through an injected lookup. `sudo` is not needed for a local `dpkg-query` read; do not prepend it, prompt for credentials, or infer that `--sudo` should affect detection.

## Host-safety and testing seams

- Keep planning pure. Detection remains an infrastructure adapter called from the existing CLI composition root.
- Add a small command-query seam (or a provider-specific read-only detector interface) that accepts `context.Context`, `CommandRequest`, and `CommandResult`. Reuse `CommandRunner`; do not add shell strings or direct `exec.Command` in detector code.
- Use bounded, non-mutating query timeouts and preserve `CommandStatus`/stdout/stderr for diagnostics if the state model is expanded. A query failure must not be treated as proof of absence without an explicit product decision, because that can cause a mutating reinstall.
- Default and dry-run modes must not probe package managers, just as current safe-mode tests assert they do not instantiate real runners or installers. Confirmed mode may perform read-only package queries before deciding whether an install command is needed.
- Unit tests should use fake command runners and fake executable lookups, table-driven by scenario. Assert exact executable, argument vector, timeout, query order, and zero install calls when presence is confirmed.
- CLI tests should inject detection and command seams and verify confirmed apply/bootstrap behavior, mixed-plan order, and safe-mode non-probing. Do not run `brew`, `dpkg-query`, `apt-get`, or `sudo` against the host. Integration tests, if eventually needed, must be opt-in/skipped under `testing.Short()` and still must not mutate the host.
- Strict TDD is active with `go test ./...`; focused package tests should precede the full suite.

## Semantic traps

| Trap | Required handling |
|---|---|
| Package name vs formula | Query `InstallMetadata.Package`; never use resource ID or `Presence.Name`. Brew formula and cask are different namespaces. |
| Installed but broken | Package-manager installed state does not prove the executable works, PATH visibility, configuration, or links. Do not replace command detection for tools/runtimes. |
| Version state | “Installed” must not imply desired/latest/version-compatible. Defer version constraints and parsing. |
| Manager unavailable | Do not silently classify an unavailable manager as absent and trigger mutation without an explicit failure/unknown policy. Prefer an explicit non-success/attention state or a conservative no-skip result. |
| Query failure/timeout | Distinguish absent from unknown/query failure; never claim idempotency from a failed query. No retry or fallback. |
| Sudo | Package detection is read-only and must not invoke sudo. `--sudo` remains an explicit install-mode choice only. |
| Cross-provider metadata | A Brew query must reject APT metadata and vice versa. Preserve provider-aware installer routing. |
| Package-manager scope | Brew prefixes, multi-arch/Multi-Homebrew setups, dpkg package states, and virtual/provided packages need explicit policy; keep the first slice narrow. |

## Scope options

1. **Brew-only package presence (recommended first slice):** add a provider-aware read-only Brew detector, use formula metadata, skip confirmed package mutation only on a successful query, and leave APT unchanged. Lowest semantic and platform risk.
2. **APT-only package presence:** add a Debian/dpkg detector with exact `install ok installed` parsing and Linux gating. More fragile because package states, manager availability, and output parsing need explicit policy.
3. **Shared contract plus Brew and APT:** coherent user outcome but larger cross-platform surface, more failure states, and higher review burden. Do not choose without a separate review budget/split.
4. **Generic manager registry/version reconciliation:** defer. It would redesign provider dispatch and turn a narrow idempotency slice into a package state model.

## Estimated impact and review risk

A provider-specific first slice is approximately **5–8 files and 150–260 changed lines**: detector/query adapter, focused tests, CLI composition tests/wiring, and a small active OpenSpec delta; README changes only if user-visible wording becomes stale. Risk is **standard reliability/resilience**, not low: command vectors, manager failure handling, and accidental mutation are the dominant concerns. A combined Brew+APT slice is approximately **8–12 files and 280–450 lines**, with a material chance of crossing the 400-line review threshold and requiring a split or exception.

Likely touch points are `internal/state`, `internal/execution/command.go`/runner seams, `cmd/dbootstrap/main.go` and tests, plus active `openspec/specs/installation-state` and `apply-command-dry-run` deltas. Avoid changing catalog schema unless a new presence kind is required; existing `InstallMetadata.Package` is sufficient for the narrow provider queries.

## Acceptance criteria if proceeding

- Package presence is read-only, provider-specific, and uses the provider package/formula field.
- Confirmed apply/bootstrap skips a package only after a successful provider query, reports `unchanged` with explicit no-mutation wording, preserves plan order, and makes no install command call.
- Absent packages remain eligible for the existing installer; query failure, timeout, or unavailable manager never produces a false already-installed result and has an explicit stable report outcome.
- Brew and APT query vectors are exact executable-plus-arguments requests; no shell-string execution, `apt-get` detection, `sudo` detection for reads, `apt update`, retries, fallback, or version reconciliation is introduced.
- Default and dry-run modes do not probe package managers and remain host-non-mutating.
- Tests use injected lookup/runner seams, cover installed/absent/unknown/manager-unavailable and mixed-plan cases, and never mutate the host.
- Focused tests pass, then `go test ./...` passes; the final diff stays within the approved slice and review budget.

## Final decision

**Split.** Package-presence idempotency is a valuable next capability, but the safest next SDD slice is Brew-only detection with a deliberately conservative unknown/error policy. Follow with an independently specified APT/dpkg slice. This preserves the archived command-presence boundary, avoids conflating formula and Debian package semantics, and keeps each review small enough to prove that read-only detection cannot accidentally authorize or execute host mutation.
