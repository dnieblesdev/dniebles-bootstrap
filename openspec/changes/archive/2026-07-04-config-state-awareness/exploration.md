# Exploration: config-state-awareness

## Current State

`dbootstrap plan` has three caller-supplied inputs to `BuildPlan(..., facts, state, installation)`:

| Input | Source | Status |
|-------|--------|--------|
| `EnvironmentFacts` | `environment.Detect()` — OS, arch, distro, WSL | ✅ Real probing via injected seams |
| `InstallationState` | `state.Detect(catalog)` — PATH lookup for tools/runtimes | ✅ Real probing via injected seams |
| `ConfigState` | `planning.ConfigState{}` — always empty | ❌ Hardcoded empty |

The planner uses `ConfigState.PresentKeys` to check `ConfigPolicy.RequiredKeys` via `missingConfigReasons()`. When a key is absent from `PresentKeys`, the resource gets `attention_required` status. Since `ConfigState` is always empty, **every resource with `config_required` keys is permanently stuck as `attention_required`**, even when the actual configuration exists on the host.

Example: `runtime:go` in `catalog/bootstrap.toml` requires `go.env`. The plan always reports `attention_required: missing required config "go.env"` regardless of whether that config file actually exists in the user's dotfiles.

## Affected Areas

- `cmd/dbootstrap/main.go:86` — `planning.ConfigState{}` call site, the only line that needs wiring
- `internal/planning/types.go:54-56` — `ConfigState` type (already correct, no changes needed)
- `internal/planning/builder.go:9,30,203-212` — `BuildPlan` and `missingConfigReasons` (already correct, no changes needed)
- `cmd/dbootstrap/main_test.go` — needs test seam for config-state detector
- `cmd/dbootstrap/render.go` — already handles `attention_required` status correctly
- `README.md:7` — stale "without real environment probing" claim
- `openspec/specs/` — needs a new delta spec for config-state detection

## Approaches

### 1. New `internal/config` package with injected seams (RECOMMENDED)

Create `internal/config/detector.go` following the exact same pattern as `internal/environment` and `internal/state`:
- Struct with injectable `ReadFile`/`LookupConfig` seam
- Package-level `Detect()` function using defaults
- Method receiver for test injection
- Returns `planning.ConfigState` with `PresentKeys`
- Detector inspects catalog resources' `config_required` keys and checks filesystem for corresponding config files

Wiring in `cmd/dbootstrap`:
```go
config := detectConfigState(catalog)  // new detector var like detectEnvironmentFacts
result := planning.BuildPlan(catalog, request, facts, config, installation)
```

**Pros:**
- Follows established project pattern (`internal/environment`, `internal/state`)
- Deterministic, host-independent tests via injected seams
- Keeps planning pure — no filesystem access in `internal/planning`
- Clean separation of concerns: one package per concern
- Makes `ConfigState` real the same way `EnvironmentFacts` and `InstallationState` were made real

**Cons:**
- Adds a new package (~50-100 lines including tests)
- Must define the mapping from config key names to filesystem paths

**Effort:** Low-Medium

### 2. Inline config probing in `cmd/dbootstrap`

Add filesystem probing directly in `runPlan()` without a separate package.

**Pros:**
- Fewer files

**Cons:**
- Violates project pattern of keeping detection outside CLI
- Makes CLI tests host-dependent or harder to isolate
- Mixes infrastructure concerns with composition root
- Goes against the established architecture (every other detector has its own package)

**Effort:** Low (but creates technical debt)

### 3. Extend `internal/state` detector to also detect config

Merge config key detection into the existing `internal/state` detector.

**Pros:**
- No new package

**Cons:**
- Violates single responsibility — installation state and config state are different concerns
- Breaks the 1:1 pattern established with `environment` (facts) and `state` (installation)
- Makes `state.Detect` signature more complex or confusing (adds `ConfigState` return)
- Future evolution would require modifying two unrelated concerns in one file

**Effort:** Low (but creates maintainability issues)

## Key Architectural Decision: Config Key → Filesystem Path Mapping

The `config_required` keys in the catalog (e.g., `"go.env"`) are abstract identifiers. The detector must map them to actual filesystem paths. This is the critical design decision:

| Mapping Strategy | Description | Tradeoff |
|-----------------|-------------|----------|
| Convention-based subpath | `go.env` → `~/.dotfiles/env/go/config.env` or similar | No catalog changes; path logic lives in detector; fragile if dotfiles structure changes |
| Explicit path in catalog | Add `config_path` field to `resourceEntry` next to `config_required` | Most flexible; requires TOML schema change; keeps mapping close to the resource definition |
| Dotfiles module name | `go.env` means "module `go` has env config" → probe `~/.dotfiles/modules/go/` | Follows dotfiles conventions naturally; requires understanding of dotfiles module structure |

**Recommendation**: Start with a convention-based approach in the detector that maps keys to a known dotfiles config base path (e.g., `~/.dotfiles/config/`). This keeps the catalog schema unchanged and the dotfiles boundary respected. The base path can be injected as a seam for testability. If the convention proves insufficient, a future slice can add explicit path mapping to the catalog.

## README Cleanup

Line 7 of `README.md` currently reads:
> This repository has its first pure Go planning-core slice (...) without (...) real environment probing.

This is stale because:
1. `environment.Detect()` performs real OS/arch/distro/WSL probing (slice: `environment-detection-adapter`)
2. `state.Detect()` performs real PATH-based tool/runtime probing (slice: `installation-state-detector`)

The cleanup should remove the "without real environment probing" clause and update the "Current status" section to reflect that the CLI now uses detected environment facts and installation state.

## Findings from Code Inspection

### Testing capability
```
ok  cmd/dbootstrap               0.003s  coverage: 87.4%
ok  internal/catalog/toml        0.003s  coverage: 86.0%
ok  internal/environment         0.003s  coverage: 82.4%
ok  internal/planning            0.003s  coverage: 92.2%
ok  internal/state               0.065s  coverage: 100.0%
```
All tests pass. Coverage is strong across all packages. The `strict_tdd false` in `sdd-init` baseline is outdated — tests exist and are comprehensive.

### Pattern consistency
The project has established a clear pattern for detectors:
1. Domain types in `internal/planning/types.go` (pure data, no side effects)
2. Detector in its own `internal/{domain}/` package with injectable seams
3. Package-level `Detect()` function using real defaults
4. CLI composition root wires detector functions via package-level vars
5. Tests stub those vars for host-independent testing

Adding `internal/config/` follows this exact proven pattern.

### Scope boundary
The dotfiles boundary is explicit in `README.md` and `AGENT.md`: `~/.dotfiles` owns modules, configs, assets, symlinks, validations, and `dotlink` semantics. The config-state detector MUST read filesystem paths but MUST NOT:
- Own or validate dotfiles configuration semantics
- Mutate any files
- Invoke `dotlink` or any dotfiles runtime

## Recommendation

**Create `internal/config/detector.go`** as a new package following the established detector pattern. The detector reads the catalog's `config_required` keys and checks whether corresponding configuration exists in the dotfiles directory via injectable filesystem seams.

The key mapping from config key names (like `"go.env"`) to filesystem paths should use a convention-based approach — derive the path from the key — with the base config path as an injectable parameter for testability. This avoids catalog schema changes and respects the dotfiles boundary.

Wire it into `cmd/dbootstrap` exactly as `environment.Detect` and `state.Detect` are wired: via a package-level function variable that tests can stub.

The README cleanup is a small, independent change that can be folded into this slice as optional scope.

## Risks

- **Unknown dotfiles layout**: The exact mapping from config keys to filesystem paths is not yet defined in the codebase. This exploration recommends convention-based mapping, but the actual convention depends on the user's dotfiles structure. Mitigation: make the base path injectable.
- **Dotfiles boundary tension**: Probing dotfiles directories for config presence walks close to the boundary line. The detector must limit itself to existence checks only, never parsing or validating dotfiles internals. Mitigation: strict read-only contract, no dotfiles runtime calls.
- **Empty catalog resources**: If no resources have `config_required` keys, the detector returns an empty `ConfigState`, which is identical to current behavior — zero risk.

## Ready for Proposal

Yes. The architecture is well-understood, the pattern is established, and the gap is a single hardcoded `ConfigState{}` call site. Proceed to `sdd-propose` with:
- New `internal/config` detector package
- CLI wiring in `cmd/dbootstrap/main.go`
- Optional README cleanup
