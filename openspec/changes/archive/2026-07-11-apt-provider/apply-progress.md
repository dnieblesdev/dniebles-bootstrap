# Apply Progress: APT Provider

## Status

All 13 implementation tasks are complete. Delivery used the approved single cohesive `size:exception` work unit (800-line budget).

## Corrective Apply Scope

- Corrected confirmed-mode output and `--yes` help so it names eligible Homebrew, Linux APT, and selected dotfile mutation paths; unsupported, non-provider-backed, and unselected work is the only work described as non-mutating or not supported yet.
- Extended behavior-focused CLI coverage for missing `apt-get`, missing `sudo`, command failure, timeout rendering, and confirmed non-zero outcomes.
- Replaced the checked-in APT catalog fixture with the design-approved `t.TempDir()` custom catalog helper. The default catalog remains unchanged.

## Fresh Evidence

The following sequence ran once after the final corrective code/test change and passed:

1. `go test ./cmd/dbootstrap ./internal/execution -run 'TestRenderExecutionReportFramesConfirmedModeMutability|TestRunApplyAptFixtureContracts|TestAptInstaller|TestBrewOrAptInstaller'`
2. `go test ./...`
3. `go vet ./...`
4. `gofmt -d` on all changed Go files produced no output.
5. `git diff --check` passed for tracked changes; `git status --short` identified untracked implementation and OpenSpec files, and `git diff --no-index --check /dev/null <untracked-file>` checked each of them. Per-file `git diff --no-index --stat` recorded their size. `git diff --stat` was used only for tracked changes and was not treated as untracked-file evidence.

## Final Verification Warning Correction

- General apply help now limits APT disclosure to eligible Linux APT installs, states direct `apt-get` use with `--yes`, and states `sudo apt-get` is available only with explicit `--yes --sudo`.
- Exact-output coverage for both general-help usage-error paths was written first and failed against the stale text, then passed after the help-only production copy change.
- Focused TDD evidence: baseline `go test ./cmd/dbootstrap -run '^TestRunUsageErrors$'` passed; RED failed after the new exact expectations; GREEN passed after the help-copy correction.
- Final evidence sequence ran once and passed: `go test ./...`; `go vet ./...`; `gofmt -d` across changed and untracked Go files produced no output; `git diff --check` passed for tracked files; and `git diff --no-index --check /dev/null <untracked-file>` completed for each untracked file without whitespace-error output.

## TDD Cycle Evidence

| Task | RED | GREEN | REFACTOR |
|---|---|---|---|
| 1.1–1.3 | Flag and command-contract tests preceded the original implementation | Focused command and CLI tests passed | Preserved explicit sudo confirmed mode and parser compatibility |
| 2.1–2.3 | Installer and provider-router tests preceded the original contracts | Table-driven execution tests passed | Reused command seams and kind-based Runner dispatch |
| 3.1–3.3 | Original fixture-backed scenarios covered direct, sudo, and non-Linux paths; corrective renderer test failed against stale disclosure | Focused CLI tests passed after truthful copy | Custom catalog now uses `t.TempDir()` as designed |
| 4.1–4.3 | Corrective CLI cases specified missing executables, failed/timed-out commands, rendered failures, and non-zero exit behavior | Focused and full suites passed | Kept fakes table-driven; no production-contract expansion |
| 4.4 | `TestRunUsageErrors` exact output was updated first and failed for the missing-command and unknown-command help paths | Focused test passed after the two-line general-help copy correction | Two distinct usage-error cases prove the shared help output; no refactor needed |

## Non-goals Preserved

No shell invocation, automatic privilege fallback, bootstrap/update/repository changes, retries, presence detection, rollback, or provider-registry redesign was introduced.
