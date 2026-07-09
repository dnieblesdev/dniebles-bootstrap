## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 40-90 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | single-pr |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Tasks

- [x] **RED — exact default catalog fixture coverage**
   - Update `internal/catalog/toml/catalog_test.go` to load `../../../catalog/bootstrap.toml` and assert the full decoded `planning.Catalog` shape, not only plan output.
   - Cover exact preservation of resources, bundles, profiles, dependency order, OS/arch metadata, descriptions, presence metadata, and package names.
   - Make the expected default catalog require `tool:git` install provider `brew` while keeping `install.package = "git"` and `presence = command_exists/git`.
   - Keep `package:ripgrep` brew-backed and `profile:dev` / `bundle:cli` unchanged.
   - Verification: `go test ./internal/catalog/toml` should fail before the catalog edit and pass after.

- [x] **RED — refresh CLI/apply expectations that depend on brew-backed `tool:git`**
   - Update only the affected exact-output assertions in `cmd/dbootstrap/main_test.go` where default apply behavior changes because `tool:git` is now brew-backed.
   - Keep plan output assertions unchanged unless a failure proves they are affected.
   - Focus on the confirmed/default apply path and the `--resource tool:git` path if the manual Homebrew guidance or step status text changes.
   - Verification: `go test ./cmd/dbootstrap` should fail before the catalog edit and pass after the expectation updates.

- [x] **GREEN — change the default catalog metadata only**
   - Edit `catalog/bootstrap.toml` so `[[tools]] id = "git"` uses `[tools.install] provider = "brew"` and keeps the existing package, presence, dependencies, resources, bundles, profiles, OS/arch metadata, and descriptions intact.
   - Do not change provider behavior, execution wiring, apt provider code, dotfiles execution, mutation paths, or schema shape.
   - Verification: re-run the focused tests from tasks 1-2 and confirm only the intended metadata change is required.

- [x] **TRIANGULATE — run the required test sequence**
   - Run `go test ./internal/catalog/toml`.
   - Run `go test ./cmd/dbootstrap`.
   - Run `go test ./...`.
   - Record any unexpected deltas; if a broader package fails, inspect only the failing assertions and keep the slice metadata-only.

- [x] **REFACTOR — only if tests expose cleanup needs**
   - If the focused suite exposes brittle assertions or duplicated catalog-shape expectations, factor them in `internal/catalog/toml/catalog_test.go` without changing behavior.
   - Keep the final diff limited to the catalog metadata edit and the smallest necessary test expectation updates.
