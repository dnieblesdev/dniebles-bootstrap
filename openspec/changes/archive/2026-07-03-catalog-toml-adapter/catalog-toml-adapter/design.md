# Design: Catalog TOML Adapter

## Technical Approach

Add an isolated `internal/catalog/toml` adapter that loads repo-local TOML and maps it into `internal/planning.Catalog`. `internal/planning` remains unchanged and format-agnostic; adapter validation stays structural so `planning.BuildPlan` keeps ownership of expansion, ordering, filtering, and semantic diagnostics.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Package boundary | Create `internal/catalog/toml` | Put parser in `internal/planning`; generic `internal/catalog` only | Keeps TOML DTOs and schema tags away from the pure planning core while leaving room for future adapters. |
| TOML parser | Use a minimal maintained TOML module such as `github.com/pelletier/go-toml/v2`; defer exact pin to apply | Go stdlib; heavier config frameworks | TOML is not in Go stdlib. A focused TOML decoder minimizes dependency surface and avoids config-framework behavior. |
| Schema style | Typed resource tables grouped by kind (`tools`, `runtimes`, `packages`) plus refs as `kind:name` strings | Fully generic resources only | Matches existing `planning.ResourceKind` constants and keeps fixture readable; a generic fallback can wait until new kinds are needed. |
| Validation boundary | Parse, required fields, duplicate IDs, supported kinds, malformed refs, basic unknown refs | Full planner validation | Structural errors should fail close to authoring; dependency ordering, environment filtering, bundle expansion, and config attention stay in planning. |

## Data Flow

```text
catalog/bootstrap.toml
  -> internal/catalog/toml DTO structs
  -> shallow validate + parse refs
  -> planning.Catalog
  -> planning.BuildPlan(request, facts, config)
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/catalog/toml/catalog.go` | Create | Public load/decode API returning `planning.Catalog`. |
| `internal/catalog/toml/schema.go` | Create | TOML DTO structs with decoder tags; unexported unless tests require package-level access. |
| `internal/catalog/toml/validate.go` | Create | Required-field, duplicate-ID, ref parsing, and basic unknown-ref checks. |
| `internal/catalog/toml/catalog_test.go` | Create | Table-driven decode, mapping, validation, and integration tests. |
| `internal/catalog/toml/testdata/*.toml` | Create | Valid and invalid focused fixtures. |
| `catalog/bootstrap.toml` | Create | Small repo-local sample catalog proving intended authoring shape. |
| `go.mod`, `go.sum` | Modify | Add selected TOML parser dependency during apply. |
| `internal/planning/*.go` | No change | Planning remains adapter-free. |

## Interfaces / Contracts

```go
package toml

func LoadFile(path string) (planning.Catalog, error)
func Decode(r io.Reader) (planning.Catalog, error)
```

Initial TOML shape:

```toml
schema = "dniebles.catalog"
version = 1

[[tools]]
id = "git"
description = "Version control"
depends_on = []
config_required = []
os = ["linux", "darwin"]

[[runtimes]]
id = "go"
depends_on = ["tool:git"]
config_required = ["go.env"]

[[packages]]
id = "ripgrep"
depends_on = ["tool:git"]

[[bundles]]
id = "cli"
resources = ["tool:git", "package:ripgrep"]

[[profiles]]
id = "dev"
bundles = ["cli"]
resources = ["runtime:go"]
```

Mapping rules: `id` becomes map key and `ResourceRef.Name`; table kind becomes `ResourceRef.Kind`; `depends_on`, bundle resources, and profile resources parse `kind:name`; `config_required` maps to `ConfigPolicy.RequiredKeys`; `os`, `arch`, `distro`, `wsl` map to `EnvironmentConditions`.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Valid TOML maps to exact `planning.Catalog` | Table-driven `Decode(strings.NewReader(...))`. |
| Unit | Malformed TOML, missing IDs, duplicate IDs, bad/unknown refs | Table-driven error tests matching stable substrings, not whole messages. |
| Integration | Fixture decodes and works with `planning.BuildPlan` | Load `catalog/bootstrap.toml` or copied `testdata` fixture; assert planned refs/statuses. |
| Side effects | Decode does not probe host or execute commands | Keep API limited to reader/file input; tests use `t.TempDir()` only for file loading. |

## Migration / Rollout

No migration required. The adapter adds a new package, fixture, and dependency only.

## Risks / Tradeoffs

- `kind:name` refs are simple but stringly typed; validation must produce clear errors.
- Grouped tables are readable now but less extensible than generic resources; acceptable until new kinds appear.
- Unknown refs checked in the adapter may overlap planner diagnostics; keep checks limited to refs available in the decoded file.

## Out of Scope

- CLI/TUI commands, installers, command runner, OS probing, git/dotfile runtime, remote catalogs, and planning-core changes.

## Open Questions

- [ ] Exact TOML library/version should be selected and pinned during apply.
