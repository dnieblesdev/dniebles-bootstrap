# Apply Progress: Direct Binary Installation — Final

**Change**: direct-binary-installation  
**Mode**: Strict TDD  
**Work units**: 1 (Secure acquisition + XDG resolver), 2 (Real installed-binary catalog read), 3 (Operator docs + final evidence)  
**Chain strategy**: stacked-to-main (base `main`)  
**Date**: 2026-07-13

## Completed Tasks

- [x] 1.1 Create `install.sh` with strict Bash mode, Linux/WSL amd64/arm64 detection, argument parsing, XDG paths, and clear unsupported-host/path errors; never invoke sudo or a package manager.
- [x] 1.2 Implement one-release API resolution in `install.sh`: stable latest by default, exact `vX.Y.Z`, prerelease rejection unless `--allow-prerelease`, and exact archive/checksum pairing from that release object.
- [x] 1.3 Add staged archive validation for only `dbootstrap` and `catalog/bootstrap.toml`, SHA-256 verification before extraction/mutation, per-file atomic replacement, backups, and transaction recovery.
- [x] 2.1 Implement `install-state.toml` recording release, exact managed paths, and digests; refuse unmanaged/incompatible files, require `--force` for matching managed installs, and permit explicit upgrades/downgrades.
- [x] 2.2 Implement `--uninstall` to validate manifest ownership and current digests, remove only unmodified binary/catalog/state, preserve modified files, and report PATH export without editing shell startup files.
- [x] 3.1 Modify `cmd/dbootstrap/main.go` to default catalog resolution to `${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/catalog/bootstrap.toml`, preserving `--catalog` and repository-local explicit paths.
- [x] 3.2 Add strict-TDD tests in `cmd/dbootstrap/main_test.go` for XDG precedence, fallback paths, `--catalog` override, non-repository CWD, and `plan --profile dev` against an installed catalog.
- [x] 4.1 Create `install_test.sh` using real temporary homes, local HTTP/archive fixtures, `tar`, checksum tools, and subprocess exit/output assertions for tuple/version/prerelease/asset-pair/checksum-before-extract/rollback/force/unmanaged/safe-uninstall scenarios; no live release calls.
- [x] 4.2 Capture real evidence: `bash -n install.sh`, `bash install_test.sh`, focused Go tests, `go test ./...`, `go vet ./...`, and a built release-like binary running `plan --profile dev` from a different CWD; record exit codes and relevant output.
- [x] 5.1 Update `README.md` with local-script and manual download/verify/install commands, supported scope, XDG paths, PATH export, `--force`, managed uninstall, catalog behavior, and all exclusions/privilege boundaries.
- [x] 5.2 Review the final diff against proposal/spec/design, verify `git diff --check`, and ensure evidence proves existing files remain untouched on every pre-mutation failure.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `install.sh` | Created | Bash installer/uninstaller with platform detection, release resolution, SHA-256 verification, atomic install, state manifest, and uninstall. |
| `install_test.sh` | Created | Fixture-driven shell tests covering tuple/version/prerelease/checksum/rollback/force/unmanaged/uninstall scenarios. |
| `cmd/dbootstrap/main.go` | Modified | Replaced CWD-relative catalog default with XDG-data resolver; `--catalog` and explicit paths preserved. |
| `cmd/dbootstrap/main_test.go` | Modified | Added `TestResolveDefaultCatalogPath`, `TestRunPlanDefaultCatalogFromXDGDataHome`, `TestRunPlanDefaultCatalogFromOutsideRepository`. |
| `README.md` | Modified | Added Direct binary installation section with local-script and manual workflows, managed paths, PATH guidance, `--force`, uninstall, catalog behavior, and privilege/scope boundaries. |

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1-1.3 | `install_test.sh` | Shell/Integration | N/A (new) | Written (installer missing) | Passed | 10 scenarios | Removed jq dep; global work_dir for trap cleanup |
| 2.1-2.2 | `install_test.sh` | Shell/Integration | N/A (new) | Written | Passed | Safe uninstall + modified preservation | Extracted `state_digest_for_path` |
| 3.1-3.2 | `cmd/dbootstrap/main_test.go` | Unit + Integration | 38 cmd tests passing | Written (compile failed) | Passed | 4 resolver cases + 2 integration cases | Made `defaultCatalogPath` injectable variable |

## Test Summary

- **Total tests written**: 3 Go tests + 10 shell scenarios
- **Total tests passing**: All
- **Layers used**: Unit (Go), Integration (Go + shell subprocess)
- **Approval tests**: None — no refactoring tasks
- **Pure functions created**: `catalogPathResolver.Resolve`, `safe_version_from_tag`

## Evidence Captured

```text
$ bash -n install.sh
exit 0

$ bash install_test.sh
All install tests passed.

$ go test ./cmd/dbootstrap/... -count=1
ok  	github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap	0.119s

$ go test ./... -count=1
ok  	github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap	0.160s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml	0.010s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/ci	0.966s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/config	0.007s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles	0.005s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/environment	0.006s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/execution	0.266s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/planning	0.012s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/state	0.012s
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/version	1.595s

$ go vet ./...
(no output)
```

### Built release-like binary reading installed catalog from arbitrary CWD

```text
$ CGO_ENABLED=0 go build -ldflags "-X github.com/dnieblesdev/dniebles-bootstrap/internal/version.Version=v0.0.0-test" -o /tmp/dbootstrap-test ./cmd/dbootstrap
$ /tmp/dbootstrap-test --version
v0.0.0-test

$ TEST_HOME=$(mktemp -d)
$ mkdir -p "$TEST_HOME/.local/share/dbootstrap/catalog"
$ cp catalog/bootstrap.toml "$TEST_HOME/.local/share/dbootstrap/catalog/bootstrap.toml"
$ WORK_DIR=$(mktemp -d)
$ cd "$WORK_DIR"
$ HOME="$TEST_HOME" /tmp/dbootstrap-test plan --profile dev
Plan profile: dev
Catalog: /tmp/tmp.6LgDgOdWfT/.local/share/dbootstrap/catalog/bootstrap.toml
Environment: os=linux arch=amd64 distro=ubuntu wsl=true

Steps:
1. tool:git [already_installed] Version control
   depends_on: none
   attention: none
2. package:jq [planned] JSON processor
   depends_on: tool:git
   attention: none
3. package:ripgrep [planned] Fast text search
   depends_on: tool:git
   attention: none
4. runtime:go [already_installed] Go toolchain
   depends_on: tool:git
   attention: missing required config "go.env"

Results:
- package:jq: planned
- package:ripgrep: planned
- runtime:go: already_installed
  reason: missing required config "go.env"
- tool:git: already_installed
exit code: 0
```

### Controlled failure preservation: checksum mismatch

```text
=== Controlled failure: checksum mismatch ===
exit code: 1
output contains 'checksum': 1
binary preserved (missing): yes
catalog preserved (missing): yes
```

### PATH guidance when bin_dir not on PATH

```text
=== PATH guidance when bin_dir not on PATH ===

Add /tmp/tmp.h57TSKFu8a/.local/bin to your PATH, for example:
  export PATH="/tmp/tmp.h57TSKFu8a/.local/bin:$PATH"
Installed dbootstrap v1.2.3:
  binary:  /tmp/tmp.h57TSKFu8a/.local/bin/dbootstrap
  catalog: /tmp/tmp.h57TSKFu8a/.local/share/dbootstrap/catalog/bootstrap.toml
```

## Deviations from Design

1. **JSON parsing dependency**: Design did not prescribe a JSON parser. Implementation uses `python3` for GitHub API response parsing because `jq` is unavailable on the target bootstrap host and not listed as a dependency. This keeps the installer self-contained on Linux/WSL systems where Python is commonly present.
2. **Work unit boundary**: The user requested XDG catalog resolver foundations in the first work unit, so Phase 3.1/3.2 were implemented there rather than in PR 2. PR 2 focused on real installed-binary catalog read and integration evidence; PR 3 added README documentation and final diff review.

## Issues Found

None.

## Remaining Tasks

None.

## Workload / PR Boundary

- **Mode**: stacked-to-main
- **Work units completed**: 1, 2, 3
- **PR 1**: Secure acquisition + managed-file transaction + XDG catalog resolver foundations → targets `main`
- **PR 2**: Real installed-binary catalog read evidence → targets PR 1 branch
- **PR 3**: README documentation + final diff review → targets PR 2 branch
- **Estimated review budget impact**: ~290 lines product code + ~215 lines tests/fixtures + ~95 lines README; total across stacked PRs exceeds single-PR budget, justifying the chained split.

## Final Review

- `git diff --check`: clean (no output)
- All proposal success criteria met:
  - Linux/WSL amd64/arm64 install without Brew or package manager; unsupported tuples fail clearly.
  - Checksum mismatch leaves existing managed files untouched.
  - Installed binary successfully reads installed catalog; documented uninstall removes both managed paths.

## Status

11/11 tasks complete. Ready for verify.
