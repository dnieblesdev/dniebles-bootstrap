# Tasks: Dotfiles Provider Readonly

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 250-450 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | Single PR with maintainer-approved size exception |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Catalog + detector readiness | PR 1 | Schema, mapping, validation, detector package, fixtures, tests |
| 2 | CLI merge wiring | PR 1 | Composition-root merge, skip-on-load-error, planner assertions |

## Phase 1: Catalog + Schema Foundation

- [x] 1.1 Add `Dotfiles []resourceEntry \`toml:"dotfiles"\`` to `internal/catalog/toml/schema.go` and keep `ResourceKindDotfile` mapping explicit.
- [x] 1.2 Update `internal/catalog/toml/catalog.go` to map dotfile resources into `planning.Catalog.Resources` and size the map accordingly.
- [x] 1.3 Extend `internal/catalog/toml/validate.go` to accept `dotfile` refs, validate `[[dotfiles]]`, and preserve bundle/profile dependency checks.

## Phase 2: Read-Only Dotfiles Detector

- [x] 2.1 Create `internal/dotfiles/detector.go` with `Detector{BasePath, Exists, ReadDir}` and package-level `Detect(catalog)`.
- [x] 2.2 Implement read-only module presence detection under `$HOME/.dotfiles`, ignoring non-dotfile resources and returning empty state on missing repo/path errors.
- [x] 2.3 Add `internal/dotfiles/detector_test.go` for repo missing, module present/absent, nil-seam defaults, and no-mutation behavior.

## Phase 3: CLI Wiring + Planner Integration

- [x] 3.1 Wire `detectDotfilesState` into `cmd/dbootstrap/main.go`, merge present dotfile refs into `InstallationState.PresentResources`, and keep `BuildPlan` signature unchanged.
- [x] 3.2 Update `cmd/dbootstrap/main_test.go` to stub dotfiles detection, verify present modules reach planning, and confirm detection is skipped when catalog loading fails.
- [x] 3.3 Add/adjust `internal/planning/builder_test.go` so supplied `dotfile:*` presence is reported as `already_installed` without planner filesystem probing.

## Phase 4: Fixture Coverage + Verification

- [x] 4.1 Add a minimal dotfile entry to `catalog/bootstrap.toml` if needed to exercise catalog/fixture coverage for `dotfile:bash`-style resources.
- [x] 4.2 Extend `internal/catalog/toml/catalog_test.go` with valid dotfile decoding, validation failures, and fixture load coverage.
- [x] 4.3 Run focused Go tests for `internal/catalog/toml`, `internal/dotfiles`, `internal/planning`, and `cmd/dbootstrap`; then rerun the broader suite after the fresh verifier pass.

## Phase 5: Cleanup / Artifact Updates

- [x] 5.1 Update any affected comments/docs to state dotfiles are read-only availability signals only.
- [x] 5.2 Remove temporary test seams or fixture scaffolding that is not needed after verification.
