# Tasks: Direct Binary Installation

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 650–900 (installer, shell fixtures, CLI tests, README) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 installer/transaction; PR 2 catalog wiring; PR 3 docs and end-to-end evidence |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | Secure acquisition and managed-file transaction | PR 1 | Standalone `install.sh` plus shell tests; base main |
| 2 | Installed catalog resolution and real CLI read | PR 2 | Base PR 1; Go tests and integration evidence |
| 3 | Operator documentation and final acceptance evidence | PR 3 | Base PR 2; README, full suite, captured runtime results |

## Phase 1: Installer Foundation

- [x] 1.1 Create `install.sh` with strict Bash mode, Linux/WSL amd64/arm64 detection, argument parsing, XDG paths, and clear unsupported-host/path errors; never invoke sudo or a package manager.
- [x] 1.2 Implement one-release API resolution in `install.sh`: stable latest by default, exact `vX.Y.Z`, prerelease rejection unless `--allow-prerelease`, and exact archive/checksum pairing from that release object.
- [x] 1.3 Add staged archive validation for only `dbootstrap` and `catalog/bootstrap.toml`, SHA-256 verification before extraction/mutation, per-file atomic replacement, backups, and transaction recovery.

## Phase 2: Ownership and Lifecycle

- [x] 2.1 Implement `install-state.toml` recording release, exact managed paths, and digests; refuse unmanaged/incompatible files, require `--force` for matching managed installs, and permit explicit upgrades/downgrades.
- [x] 2.2 Implement `--uninstall` to validate manifest ownership and current digests, remove only unmodified binary/catalog/state, preserve modified files, and report PATH export without editing shell startup files.

## Phase 3: CLI Integration

- [x] 3.1 Modify `cmd/dbootstrap/main.go` to default catalog resolution to `${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/catalog/bootstrap.toml`, preserving `--catalog` and repository-local explicit paths.
- [x] 3.2 Add strict-TDD tests in `cmd/dbootstrap/main_test.go` for XDG precedence, fallback paths, `--catalog` override, non-repository CWD, and `plan --profile dev` against an installed catalog.

## Phase 4: Verification and Evidence

- [x] 4.1 Create `install_test.sh` using real temporary homes, local HTTP/archive fixtures, `tar`, checksum tools, and subprocess exit/output assertions for tuple/version/prerelease/asset-pair/checksum-before-extract/rollback/force/unmanaged/safe-uninstall scenarios; no live release calls.
- [x] 4.2 Capture real evidence: `bash -n install.sh`, `bash install_test.sh`, focused Go tests, `go test ./...`, `go vet ./...`, and a built release-like binary running `plan --profile dev` from a different CWD; record exit codes and relevant output.

## Phase 5: Documentation

- [x] 5.1 Update `README.md` with local-script and manual download/verify/install commands, supported scope, XDG paths, PATH export, `--force`, managed uninstall, catalog behavior, and all exclusions/privilege boundaries.
- [x] 5.2 Review the final diff against proposal/spec/design, verify `git diff --check`, and ensure evidence proves existing files remain untouched on every pre-mutation failure.
