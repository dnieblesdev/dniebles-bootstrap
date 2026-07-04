# Apply Progress: Dotfiles Provider Readonly

## Status

All tasks implemented. Full test suite passes.

## Completed Tasks

- [x] 1.1 Add `Dotfiles []resourceEntry \`toml:"dotfiles"\`` to `internal/catalog/toml/schema.go` and keep `ResourceKindDotfile` mapping explicit.
- [x] 1.2 Update `internal/catalog/toml/catalog.go` to map dotfile resources into `planning.Catalog.Resources` and size the map accordingly.
- [x] 1.3 Extend `internal/catalog/toml/validate.go` to accept `dotfile` refs, validate `[[dotfiles]]`, and preserve bundle/profile dependency checks.
- [x] 2.1 Create `internal/dotfiles/detector.go` with `Detector{BasePath, Exists, ReadDir}` and package-level `Detect(catalog)`.
- [x] 2.2 Implement read-only module presence detection under `$HOME/.dotfiles`, ignoring non-dotfile resources and returning empty state on missing repo/path errors.
- [x] 2.3 Add `internal/dotfiles/detector_test.go` for repo missing, module present/absent, nil-seam defaults, and no-mutation behavior.
- [x] 3.1 Wire `detectDotfilesState` into `cmd/dbootstrap/main.go`, merge present dotfile refs into `InstallationState.PresentResources`, and keep `BuildPlan` signature unchanged.
- [x] 3.2 Update `cmd/dbootstrap/main_test.go` to stub dotfiles detection, verify present modules reach planning, and confirm detection is skipped when catalog loading fails.
- [x] 3.3 Add/adjust `internal/planning/builder_test.go` so supplied `dotfile:*` presence is reported as `already_installed` without planner filesystem probing.
- [x] 4.1 Add a minimal dotfile entry to `catalog/bootstrap.toml` if needed to exercise catalog/fixture coverage for `dotfile:bash`-style resources.
- [x] 4.2 Extend `internal/catalog/toml/catalog_test.go` with valid dotfile decoding, validation failures, and fixture load coverage.
- [x] 4.3 Run focused Go tests for `internal/catalog/toml`, `internal/dotfiles`, `internal/planning`, and `cmd/dbootstrap`; then rerun the broader suite after the fresh verifier pass.
- [x] 5.1 Update any affected comments/docs to state dotfiles are read-only availability signals only.
- [x] 5.2 Remove temporary test seams or fixture scaffolding that is not needed after verification.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/catalog/toml/schema.go` | Modified | Added `Dotfiles []resourceEntry \`toml:"dotfiles"\`` to private catalog schema. |
| `internal/catalog/toml/catalog.go` | Modified | Included dotfile resources in map capacity and `mapResources` mapping. |
| `internal/catalog/toml/validate.go` | Modified | Collected dotfile refs, validated dotfile dependencies, and added `ResourceKindDotfile` to `supportedKind`. |
| `internal/catalog/toml/catalog_test.go` | Modified | Added dotfile resource coverage, dotfile-specific validation cases, and a test proving dotfile refs work across bundles/profiles. |
| `internal/dotfiles/detector.go` | Created | Read-only detector with injectable `Exists`/`ReadDir` seams and default `$HOME/.dotfiles` base path. |
| `internal/dotfiles/detector_test.go` | Created | Deterministic tests for repo missing, module present/absent, read errors, non-dotfile ignores, nil-seam defaults, and no filesystem mutation. |
| `cmd/dbootstrap/main.go` | Modified | Added `detectDotfilesState` seam, merged detected dotfile presence into installation state, and kept `BuildPlan` signature unchanged. |
| `cmd/dbootstrap/main_test.go` | Modified | Added `stubDotfilesState` helper, test proving present dotfile modules reach planning, and extended catalog-load error test to assert dotfiles detection is skipped. |
| `internal/planning/builder_test.go` | Modified | Added test proving supplied `dotfile:*` presence reports `already_installed` without planner filesystem probing. |
| `catalog/bootstrap.toml` | Modified | Added minimal `[[dotfiles]]` entry to exercise fixture coverage. |

## Tests Run

```bash
go test ./internal/catalog/toml/... ./internal/dotfiles/... ./internal/planning/... ./cmd/dbootstrap/...
go test ./...
go vet ./...
```

All passed.

## Deviations from Design

None — implementation matches design.

## Issues Found

None.

## Remaining Work

None. Ready for verify.

## Workload / PR Boundary

- Mode: single PR with maintainer-approved size exception
- Chain strategy: size-exception / single PR
- Estimated review budget impact: high, but pre-approved exception is active
