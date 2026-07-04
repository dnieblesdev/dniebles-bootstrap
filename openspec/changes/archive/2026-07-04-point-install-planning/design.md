# Design: Point Install Planning

## Technical Approach

Expose existing `planning.PlanRequest.Resources` through the `dbootstrap plan` CLI only. The CLI will parse repeatable `--resource kind:name` flags, validate shape and supported kinds before catalog loading, require at least one target (`--profile` or `--resource`), then call `planning.BuildPlan` with both `Profile` and `Resources`. The planner remains unchanged because it already expands profile targets and explicit point resources into one deterministic union.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|----------|--------|-------------------------|-----------|
| Parser location | Add a small CLI-local parser in `cmd/dbootstrap/main.go` with the same accepted shape/kinds as catalog refs. | Export/reuse `internal/catalog/toml.parseRef`; create a new shared package. | The existing parser is unexported and TOML-adapter scoped. A shared package would be cleaner long-term, but for one CLI edge it adds package churn. Keep duplication minimal, table-tested, and revisit if another caller needs parsing. |
| Request model | Populate existing `planning.PlanRequest{Profile: profile, Resources: resources}`. | Add planner API or new domain types. | `BuildPlan.expandRequest` already unions profile-expanded resources and explicit resources, de-duplicates selections, and orders output deterministically. |
| Profile optionality | Require `profile != "" || len(resources) > 0`. | Keep profile mandatory and make resources additive only. | Specs require resource-only planning; CLI validation is the correct boundary for target-required errors. |
| Rendering | Keep renderer read-only and make the first header line conditional: profile header when profile exists, resource header when profile is empty. | Always print empty `Plan profile:`; pass a new renderer request struct. | Conditional output is the smallest compatible change. A new struct is unnecessary unless rendering gains more target metadata later. |

## Data Flow

```text
argv --resource flags ──→ CLI parser/validation ──→ []planning.ResourceRef
            --profile ────────────────────────────→ planning.PlanRequest
catalog/state/facts ───────────────────────────────→ planning.BuildPlan
PlanResult ────────────────────────────────────────→ renderPlanResult
```

`BuildPlan` remains a pure planner: no installer, apply, mutation, or runtime execution path is introduced.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modify | Add repeatable `--resource`; parse/validate refs; update target-required validation; pass resources into `PlanRequest`; update plan usage/help. |
| `cmd/dbootstrap/main_test.go` | Modify | Add CLI cases for resource-only, profile+resource union, repeated resources, malformed refs, unsupported kinds, missing target, and unchanged profile-only output. |
| `cmd/dbootstrap/render.go` | Modify | Render `Plan resources: kind:name, ...` when profile is empty; preserve existing profile header otherwise. |
| `cmd/dbootstrap/render_test.go` | Modify | Cover resource-only header and unchanged profile header. |
| `internal/planning/*` | Unchanged | Existing `PlanRequest.Resources` and `BuildPlan` union behavior are reused. |

## Interfaces / Contracts

```go
// CLI-only helper shape; exact names may vary.
type resourceRefsFlag []string

func parseResourceRef(value string) (planning.ResourceRef, error)
func parseResourceRefs(values []string) ([]planning.ResourceRef, error)
```

Validation contract:
- accepted format: exactly `kind:name`
- both parts must be non-empty
- supported kinds: `tool`, `runtime`, `package`, `dotfile`
- malformed or unsupported refs return `exitUsage` before read-only detection/planning begins

Union contract: profile resources, profile bundle resources, explicit resources, and dependencies are combined by `BuildPlan`; duplicate refs are selected once.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | CLI resource parser and repeatable flag accumulation | Table tests in `cmd/dbootstrap/main_test.go` or focused helper tests. |
| Integration-ish CLI | Resource-only, profile+resource union, profile-only regression, missing target, invalid refs | Existing `run(...)` buffer tests with fixture catalogs and stubbed detectors. |
| Renderer | Header behavior | Existing render tests for profile header plus new resource-only expectation. |
| Domain | Planner remains unchanged | Rely on existing `internal/planning` tests that already cover point resources and purity. |

## Migration / Rollout

No migration required. Rollout is a CLI-only additive flag. Rollback removes the flag path and restores profile-required validation/header behavior.

## Open Questions

- [ ] None.
