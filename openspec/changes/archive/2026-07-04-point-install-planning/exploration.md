## Exploration: point-install-planning

### Current State

`dniebles-bootstrap` has a fully working profile-level planning pipeline driven by `dbootstrap plan --profile dev`. The planning domain (`internal/planning`) already defines `PlanRequest` with both `Profile string` and `Resources []ResourceRef` fields, and `BuildPlan` already handles both expansion paths — the `expandRequest()` method in `builder.go` expands profiles (into bundles + resources) AND includes enumerated resource refs directly. The test "point resources include dependencies only once" in `builder_test.go` proves this works at the domain level.

However, the CLI (`cmd/dbootstrap/main.go`, `runPlan`) mandates `--profile` as a required flag and only ever constructs `planning.PlanRequest{Profile: *profile}` — the `Resources` field is never populated. There is no CLI path to plan a specific resource without a profile.

The full planning pipeline (environment detection, installation state detection, dotfiles availability merge, config state detection) runs before `BuildPlan` in the CLI and is agnostic to whether resources are profile-expanded or point-specified. The renderer (`render.go`) accepts `PlanResult` generically.

The TOML catalog adapter already supports all resource kinds (tool, runtime, package, dotfile), and the `parseRef` function validates `kind:name` format with a list of supported kinds. The catalog fixture (`catalog/bootstrap.toml`) has resources of all four kinds.

### Affected Areas

- `cmd/dbootstrap/main.go:56-101` — `runPlan` function: flags parsing, validation logic (`--profile` required), `PlanRequest` construction. Must add `--resource` flag, make `--profile` optional when resources are given, and pass resources to `PlanRequest.Resources`.
- `cmd/dbootstrap/main_test.go` — test suite: must cover point-only planning (no profile), combined profile+resource planning, `--resource` flag parsing errors, and ensure existing profile-only tests still pass.
- `cmd/dbootstrap/render.go:11-14` — `renderPlanResult` header: currently prints `"Plan profile: %s"`. Should adapt when only resources are specified (no profile).
- `internal/planning/types.go` — No change needed. `PlanRequest.Resources` already exists and `Builder` already handles it.
- `internal/planning/builder.go` — No change needed. `expandRequest` already loops over `request.Resources`.
- `catalog/bootstrap.toml` — No change needed for functionality, but adding a second profile or documenting point usage in this fixture improves test coverage. Optional.

### Approaches

1. **`--resource` repeatable flag + profile optionality** — Add a `--resource` CLI flag accepting `kind:name` strings. Make `--profile` optional when at least one `--resource` is given; keep it required otherwise. Pass parsed refs to `PlanRequest.Resources`. Mixing both (profile + resources) produces a union plan — already the planner's behavior.
   - Pros: Minimal change; domain already models it; backward-compatible; no new subcommands; renders naturally.
   - Cons: `kind:name` parsing at CLI edge duplicates `parseRef` logic from the TOML adapter (addressable by extracting shared validation or accepting a small duplication).
   - Effort: Low

2. **Separate `point` subcommand** — Add `dbootstrap point --resource tool:git`. Profile remains under `plan`. 
   - Pros: Explicit command-level separation; profile and point clearly distinct.
   - Cons: More CLI code; shares nearly identical pipeline; may invite code duplication; adds subcommand the user doesn't need to learn separately.
   - Effort: Medium

3. **Positional arguments on `plan`** — `dbootstrap plan tool:git runtime:go` (no `--profile`).
   - Pros: Concise UX.
   - Cons: Ambiguous — how to distinguish positional resources from a positional profile name? `--profile` would need a separate flag; breaks existing `plan` invocation convention. Harder error messaging.
   - Effort: Low (code), High (UX risk)

### Recommendation

**Approach 1** — `--resource` repeatable flag with profile optionality. The domain already supports this path; the CLI just needs to expose it. The change is ~50-80 lines in `main.go` plus tests (~100-150 lines). All existing profile-based tests pass unchanged. The `--resource` flag accepts `kind:name` values validated at parse time; unsupported kinds or malformed values produce clear usage errors. When only resources are given (no profile), the renderer header adapts from `"Plan profile: %s"` to `"Plan resources: ..."`.

Design decisions:
- `--resource` is repeatable: `dbootstrap plan --resource tool:git --resource runtime:go`
- `--profile` remains supported and becomes optional when `--resource` is given
- Mixing both is allowed (union behavior, already the planner's semantics)
- At least one of `--profile` or `--resource` is required
- Resource refs use the same `kind:name` format as the catalog (tool:git, runtime:go, package:ripgrep, dotfile:bash)
- The `kind:name` parser at CLI level mirrors the TOML adapter's `parseRef` and `supportedKind` — either extract to a shared `internal/ref` package or accept the small duplication given the TOML adapter's functions are unexported

### Risks

- `kind:name` validation duplication between CLI and TOML adapter: low risk if extracted to a small shared package; medium risk if duplicated and diverges later.
- Renderer header format when only resources are specified: needs to be clear without making the output contract too different from profile output (reviewer consistency).
- No runtime mutation risk: this is planning-only; the planner is pure; no installers or apply commands are touched.
- Mixing `--profile` and `--resource` semantics: the planner does union — make sure this is documented in CLI help and rendered output so users aren't confused.

### Ready for Proposal

Yes — the domain already models point planning. The change is scoped to CLI wiring, flag parsing, and test coverage. All tests pass clean, and the planner requires zero changes. Proceed to `sdd-propose`.
