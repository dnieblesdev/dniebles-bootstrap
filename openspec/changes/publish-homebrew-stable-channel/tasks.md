# Tasks: Publish Homebrew Stable Channel

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 400–550 |
| 400-line budget risk | Medium |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 README summary + tap skeleton; PR 2 stable gate + pinned formula + evidence |
| Delivery strategy | auto-chain |
| Chain strategy | pending orchestrator resolution |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: Medium

## Change Status

**OPEN and BLOCKED.** This change owns the eight tasks moved from `homebrew-installation-channel` (previously 2.1–4.2). It cannot leave BLOCKED status until a real GitHub Release is public, not draft, not prerelease, and provides Linux `amd64` and `arm64` archives with matching SHA-256 files. A prerelease may be used only for technical validation and MUST NOT seed the stable formula.

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | README summary and tap README structure | PR 1 | `README.md` summary only; tap README owns evidence; no formula yet |
| 2 | Stable gate, formula creation, and evidence | PR 2 | Depends on PR 1; formula created from scratch after gate passes |

## Phase 1: Documentation Boundary

- [ ] 1.1 Update `README.md` with summary Linux/WSL Homebrew install/use instructions only; link to the tap README for detailed publication evidence, hashes, and operational proof.
- [ ] 1.2 In standalone repository `dnieblesdev/homebrew-dniebles-bootstrap`, create `README.md` documenting tap/install/upgrade/uninstall commands, ownership, stable-release prerequisite, installed catalog path, macOS exclusion, and where detailed evidence/hashes will be recorded.

## Phase 2: Stable Release Gate and Formula Creation

- [ ] 2.1 Qualify a named stable release: run `gh release view <tag> --json isDraft,isPrerelease,tagName,assets`; require public, non-draft, non-prerelease status; require Linux `amd64` and `arm64` archives plus matching `.sha256` assets; download and validate each `.sha256` content against the archive bytes.
- [ ] 2.2 After the gate passes, create `dnieblesdev/homebrew-dniebles-bootstrap/Formula/dbootstrap.rb` from scratch with Linux Intel/ARM branches, `pkgshare.install "catalog/bootstrap.toml"`, macOS pre-download `odie`, no unsupported-CPU fallback, and literal pinned version/URL/SHA-256 values.
- [ ] 2.3 On Linux/WSL `amd64` and `arm64`, capture `brew tap`, install, `dbootstrap --version`, strict audit/style, reinstall/upgrade, uninstall, payload cleanup, unrelated-file preservation, and arbitrary-CWD `plan --profile dev` proof.
- [ ] 2.4 Capture macOS formula output proving clear rejection before download; record release URL, version, asset names/digests, installed catalog path, and command output in the tap README.

## Phase 3: Verification

- [ ] 3.1 Run focused `cmd/dbootstrap` tests followed by `go test ./...`; report test files, scenarios, and any skipped external Homebrew integration.
- [ ] 3.2 Review the diff against `publish-homebrew-stable-channel` proposal/spec/design and approved `review-ledger.md`; verify only docs, standalone tap, and evidence boundary changed, and that no prerelease was published as stable.
