# Apply Progress: Catalog More Brew Targets

## Status

All 10 tasks remain complete in the single maintainer-approved `size:exception` work unit. This corrective apply pass ran the configured repository-wide formatter and re-ran verification; it made no production-scope changes. No commit, push, or pull request was created.

## Completed Tasks

- [x] 1.1 Add the brew-backed `package:jq` catalog record.
- [x] 1.2 Add `package:jq` to `bundle:cli`.
- [x] 2.1 Add jq to the default catalog contract model with install and presence metadata.
- [x] 2.2 Assert bundle membership, `profile:dev` selection, deterministic plan status/order, and inert metadata retention.
- [x] 2.3 Preserve existing catalog contents without fallback or multi-provider metadata.
- [x] 3.1 Run focused catalog tests without Homebrew or apply execution.
- [x] 3.2 Run formatting, the full Go suite, and `go vet`.
- [x] 3.3 Review the diff against all approved out-of-scope boundaries.
- [x] 3.4 Confirm rollback removes the jq catalog and bundle entries with their assertions.
- [x] 4.1 Synchronize OpenSpec and Engram task state in English.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|---|---|---|---|
| 1.1–1.2, 2.1–2.3 | `internal/catalog/toml/catalog_test.go` | Catalog contract / planning integration | `go test ./internal/catalog/toml` passed | `TestLoadFileAndBuildPlanFromFixture` was extended before catalog data changed; it failed because jq and CLI membership were absent | Focused test passed after the catalog entry and membership were added | Existing fixture asserts two outcomes from the same real catalog: exact decoded metadata/model and deterministic `profile:dev` plan/status; no alternate production path exists for this declarative entry | None needed — data-only extension |
| 3.1–3.3 | `cmd/dbootstrap/main_test.go` | CLI contract | Full suite revealed exact-output expectations that still modeled the old catalog | Existing CLI contract tests failed after the new default selection was introduced | Updated expectations passed, confirming jq appears in plan and non-mutating apply reports without changing execution code | Existing table covers default plan, installed-tool plan, configured-runtime plan, profile-plus-resource union, default apply, dry-run, and confirmed apply | None needed — snapshot updates only |

## Verification Evidence

### Previous implementation evidence retained

| Command | Result |
|---|---|
| `go test ./internal/catalog/toml` (baseline) | Passed |
| `go test ./internal/catalog/toml -run '^TestLoadFileAndBuildPlanFromFixture$'` before catalog edit | Failed as expected: jq was absent from the catalog/model |
| `go test ./internal/catalog/toml -run '^TestLoadFileAndBuildPlanFromFixture$'` after catalog edit | Passed |
| `go test ./cmd/dbootstrap -run '^(TestRunPlanCommand|TestRunApplyCommand)$'` | Passed |

### Corrective AUTO GATE evidence — 2026-07-11

All commands below were run from the repository root. Empty output means the command exited successfully without stdout/stderr.

#### Required formatting command

```text
$ gofmt -w .
(no output; exit 0)
```

The configured command completed successfully. Its post-format diff inspection found no formatter-only changes; the existing approved production diff remained 66 additions and 20 deletions across the three files below.

```text
$ git diff --stat
 catalog/bootstrap.toml                | 13 ++++++++-
 cmd/dbootstrap/main_test.go           | 53 ++++++++++++++++++++++++-----------
 internal/catalog/toml/catalog_test.go | 20 +++++++++++--
 3 files changed, 66 insertions(+), 20 deletions(-)
```

#### Focused catalog suite

```text
$ go test ./internal/catalog/toml
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml	(cached)
```

#### Full suite

```text
$ go test ./...
ok  	github.com/dnieblesdev/dniebles-bootstrap/cmd/dbootstrap	(cached)
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/catalog/toml	(cached)
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/config	(cached)
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/dotfiles	(cached)
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/environment	(cached)
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/execution	(cached)
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/planning	(cached)
ok  	github.com/dnieblesdev/dniebles-bootstrap/internal/state	(cached)
```

#### Vet

```text
$ go vet ./...
(no output; exit 0)
```

#### Diff check

```text
$ git diff --check
(no output; exit 0)
```

No Homebrew command, `apply` command, bootstrap operation, or external command execution was performed.

## Scope Review and Rollback

The post-format diff is limited to the jq catalog record, existing CLI bundle membership, and derived catalog/CLI output contracts in `catalog/bootstrap.toml`, `internal/catalog/toml/catalog_test.go`, and `cmd/dbootstrap/main_test.go`. It does not alter `runtime:go`, `dotfile:bash`, providers, runner, command execution, apply semantics, bootstrap behavior, fallback metadata, profiles, or unrelated production source.

Rollback removes the `[[packages]]` jq record and `package:jq` from `bundle:cli`, then reverts the matching contract assertions in `internal/catalog/toml/catalog_test.go` and `cmd/dbootstrap/main_test.go`.
