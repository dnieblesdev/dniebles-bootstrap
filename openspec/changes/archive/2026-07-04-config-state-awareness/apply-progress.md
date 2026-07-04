# Apply Progress: config-state-awareness

## Status

All tasks complete.

## Completed Tasks

- [x] 1.1 Create `internal/config/detector.go` with `Detector`, `Detect`, `PathExists`, and `KeyPathResolver` seams mirroring `internal/environment`/`internal/state`.
- [x] 1.2 Add path convention logic for `$HOME/.dotfiles/config/<key parts>` and treat empty, absolute, or escaping keys as absent.
- [x] 1.3 Add `internal/config/detector_test.go` table cases for present, missing, invalid-key, and deterministic fixture behavior using injected seams.
- [x] 2.1 Wire `internal/config` into `cmd/dbootstrap/main.go` with a package-level `detectConfigState` var and pass the detected state to `planning.BuildPlan`.
- [x] 2.2 Update `cmd/dbootstrap/main_test.go` to stub config detection, prove catalog-load failures skip detection, and prove present config changes runtime planning output.
- [x] 2.3 Review `internal/planning/builder_test.go` for any missing caller-supplied config-state coverage and add a focused case if needed.
- [x] 3.1 Clean `README.md` wording that still says the plan command uses empty configuration state or avoids real environment probing.
- [x] 3.2 Verify the catalog TOML adapter still maps `config_required` into `planning.ConfigPolicy.RequiredKeys` without schema changes.
- [x] 4.1 Run focused Go tests for `internal/config`, `cmd/dbootstrap`, `internal/planning`, and `internal/catalog/toml`.
- [x] 4.2 Confirm `dbootstrap plan --profile dev` behavior still shows missing-config attention when absent and no missing-config attention when config is reported present.
- [x] 4.3 Re-run the full relevant test slice after README and wiring updates to prove no host-dependent behavior leaked in.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/config/detector.go` | Created | Read-only config detector with injectable `PathExists`, `KeyPathResolver`, and `BasePath` seams; default convention maps keys to `$HOME/.dotfiles/config/<parts>`. |
| `internal/config/detector_test.go` | Created | Table-driven tests for present/missing/invalid keys, deterministic behavior, default seam path, catalog immutability, and error-as-absence. |
| `cmd/dbootstrap/main.go` | Modified | Imported `internal/config`, added `detectConfigState` package-level var, and passed detected `ConfigState` to `planning.BuildPlan`. |
| `cmd/dbootstrap/main_test.go` | Modified | Added `stubConfigState` helper; updated existing cases to stub empty state; added present-config case; added test proving catalog-load errors skip config detection. |
| `internal/planning/builder_test.go` | Modified | Added `TestBuildPlanConfigState` covering missing-config attention, present-config avoidance, and empty `PresentKeys` map preservation. |
| `README.md` | Modified | Updated current-status and CLI-usage wording to reflect read-only config-state detection and pure planning boundary. |

## Deviations from Design

None — implementation matches design.

## Issues Found

None.

## Tests Run

- `go test ./internal/config/... -v`
- `go test ./cmd/dbootstrap/... -v`
- `go test ./internal/planning/... -v`
- `go test ./internal/catalog/toml/... -v`
- `go test ./... -v`
- `go test -race ./...`
- `go vet ./...`
- `gofmt -l .` (clean)
- Manual: `go run ./cmd/dbootstrap plan --profile dev` with and without `$HOME/.dotfiles/config/go/env` fixture.

## Workload / PR Boundary

- Mode: single PR with maintainer-approved size exception
- Current work unit: Unit 1 (detector + CLI wiring + coverage)
- Boundary: complete config-state-awareness feature from detector through CLI wiring, tests, and docs
- Estimated review budget impact: within 220-360 line forecast; size exception recorded

## Next Recommended

`sdd-verify`
