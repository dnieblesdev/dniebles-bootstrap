# Review Ledger: Homebrew Installation Channel

## Judgment Day — Design Phase, Round 1

**Verdict:** `JUDGMENT: APPROVED`

Blind Judge A approved the design. Blind Judge B identified the following blocker. The judges contradicted, so no automatic fix was applied until the user confirmed the package-share layout.

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-B-001 | judgment-day | `openspec/changes/homebrew-installation-channel/design.md:13-14,25,27,37,40,53,59-63` | BLOCKER | verified | Confirmed decision applied: the formula preserves the release archive's `catalog/bootstrap.toml` structure through `pkgshare.install "catalog/bootstrap.toml"`, and every resolver, data-flow, ownership, testing, and evidence statement now uses `<HOMEBREW_PREFIX>/share/dbootstrap/catalog/bootstrap.toml`. Both scoped re-judges verified the correction. |

## Judge Results

- Judge A: no real user-impacting inconsistency found.
- Judge B: `JD-B-001` BLOCKER; recommends preserving the `catalog/` segment in the resolver and design unless the formula explicitly stages/renames the archive payload.

## Resolved Decision

The user confirmed that the formula and resolver must use `share/dbootstrap/catalog/bootstrap.toml`. The correction preserves `catalog/bootstrap.toml` from the published asset; it does not stage or rename it.

## Judgment Day — Apply Phase 1

**Verdict:** `JUDGMENT: APPROVED`

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-A-001 | judgment-day | `cmd/dbootstrap/main.go:85-88` | WARNING | info | `fileExists` accepts an existing directory at a catalog-file path, which can prevent fallback to a valid lower-priority catalog. |
| JD-B-001 | judgment-day | `cmd/dbootstrap/main_test.go:745-861` | WARNING | info | Missing triangulation for a configured-but-absent XDG candidate falling through to a lower existing candidate. |
| JD-B-002 | judgment-day | `cmd/dbootstrap/main_test.go` home-resolution-error case | WARNING | info | Missing direct coverage of home-directory failure while an existing Homebrew catalog is available. |

Warnings are non-blocking and were not auto-fixed under the review policy.

## Judgment Day — Scope Split Review

**Verdict:** `JUDGMENT: APPROVED`

The narrowed slice preserves nine passing resolver cases, moves all eight publication tasks to `publish-homebrew-stable-channel`, and makes no stable-publication claim.

| id | lens | location | severity | status | evidence |
|---|---|---|---|---|---|
| JD-B-101 | judgment-day | `tasks.md`, `design.md`, `apply-progress.md` | WARNING | info | Artifacts mention `t.TempDir()` although resolver tests use an injected `PathExists` map. |
| JD-B-102 | judgment-day | `apply-progress.md` | WARNING | info | Remaining-task echo numbering differs from the destination change's task numbering. |
