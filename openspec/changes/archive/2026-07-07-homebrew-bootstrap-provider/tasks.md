# Tasks: Homebrew Bootstrap Provider

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 260-380 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Add safe Homebrew detection and bootstrap planning | PR 1 | Base: main; tests included |
| 2 | Render manual guidance and preserve non-mutating apply paths | PR 1 | Same PR; verify no mutation wiring |

## Phase 1: Foundation / Contracts

- [x] 1.1 Add a Homebrew presence seam using `exec.LookPath("brew")`-style detection in the bootstrap provider path; do not invoke `CommandRunner` or spawn shell commands.
- [x] 1.2 Define the Homebrew bootstrap resource/planning shape in the provider layer without adding raw command fields or shell-first metadata.

## Phase 2: Core Implementation

- [x] 2.1 Build bootstrap planning for brew-backed resources from synthetic fixtures so the provider can distinguish: no brew-backed resources, brew present, and brew missing.
- [x] 2.2 Emit manual-only bootstrap guidance using the official Homebrew install command as text, and never execute it or install target packages.
- [x] 2.3 Ensure apply/default, `--dry-run`, and `--yes` all remain non-mutating for Homebrew bootstrap steps, including no-op result wiring.

## Phase 3: Integration / Wiring

- [x] 3.1 Wire the provider into apply/report rendering so Homebrew bootstrap status appears in plan output and manual action rendering.
- [x] 3.2 Keep `catalog/bootstrap.toml` untouched; route all Homebrew scenarios through fixtures/synthetic plans and existing execution contracts.

## Phase 4: Testing / Verification

- [x] 4.1 Add tests for planning when there are no brew-backed resources and when brew is missing, asserting manual guidance is rendered and no mutating work is scheduled.
- [x] 4.2 Add tests for brew-present detection and apply modes (`default`, `--dry-run`, `--yes`) to prove the provider stays non-mutating.
- [x] 4.3 Add tests for no-mutation wiring so Homebrew bootstrap never reaches `CommandRunner` and never executes shell commands.

## Phase 5: Cleanup / Documentation

- [x] 5.1 Update inline comments and any task-linked docs so the Homebrew bootstrap slice stays explicit about detection-only behavior and manual operator steps.
