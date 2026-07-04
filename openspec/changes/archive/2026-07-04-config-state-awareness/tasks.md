# Tasks: Config State Awareness

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 220-360 |
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
| 1 | Detector + CLI wiring + coverage | PR 1 | Single PR under maintainer-approved size exception; tests and docs included |

## Phase 1: Foundation

- [x] 1.1 Create `internal/config/detector.go` with `Detector`, `Detect`, `PathExists`, and `KeyPathResolver` seams mirroring `internal/environment`/`internal/state`.
- [x] 1.2 Add path convention logic for `$HOME/.dotfiles/config/<key parts>` and treat empty, absolute, or escaping keys as absent.
- [x] 1.3 Add `internal/config/detector_test.go` table cases for present, missing, invalid-key, and deterministic fixture behavior using injected seams.

## Phase 2: Core Implementation

- [x] 2.1 Wire `internal/config` into `cmd/dbootstrap/main.go` with a package-level `detectConfigState` var and pass the detected state to `planning.BuildPlan`.
- [x] 2.2 Update `cmd/dbootstrap/main_test.go` to stub config detection, prove catalog-load failures skip detection, and prove present config changes runtime planning output.
- [x] 2.3 Review `internal/planning/builder_test.go` for any missing caller-supplied config-state coverage and add a focused case if needed.

## Phase 3: Integration and Docs

- [x] 3.1 Clean `README.md` wording that still says the plan command uses empty configuration state or avoids real environment probing.
- [x] 3.2 Verify the catalog TOML adapter still maps `config_required` into `planning.ConfigPolicy.RequiredKeys` without schema changes.

## Phase 4: Verification

- [x] 4.1 Run focused Go tests for `internal/config`, `cmd/dbootstrap`, `internal/planning`, and `internal/catalog/toml`.
- [x] 4.2 Confirm `dbootstrap plan --profile dev` behavior still shows missing-config attention when absent and no missing-config attention when config is reported present.
- [x] 4.3 Re-run the full relevant test slice after README and wiring updates to prove no host-dependent behavior leaked in.
