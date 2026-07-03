# Exploration: CLI plan command

### Current State
The repo already has the two core pieces needed for a thin CLI slice: `internal/planning.BuildPlan` is pure and deterministic, and `internal/catalog/toml.LoadFile` can decode `catalog/bootstrap.toml` into planning inputs. There is no `cmd/` entrypoint yet, so the next slice is the first executable wrapper around those layers.

### Affected Areas
- `cmd/dbootstrap` — new CLI entrypoint/package for `plan`.
- `internal/catalog/toml` — reused as the repository-local catalog loader.
- `internal/planning` — reused as the pure planning engine.
- `catalog/bootstrap.toml` — default repo-local input file for planning.

### Approaches
1. **Stdlib `flag` entrypoint** — single `main` package using `flag` to parse `plan --profile dev` and delegate to a small app function.
   - Pros: smallest dependency surface, easy to keep thin, deterministic tests via direct function calls.
   - Cons: manual subcommand wiring, less ergonomic as commands grow.
   - Effort: Low

2. **CLI framework** — introduce a library like Cobra early.
   - Pros: nicer UX for future subcommands, flags, help, and nested commands.
   - Cons: extra abstraction for a slice that only needs one command; higher coupling and more test surface.
   - Effort: Medium

### Recommendation
Use stdlib `flag` for the first slice. Keep `cmd/dbootstrap/main.go` as a tiny bootstrapper and move behavior into a small internal package/function so the CLI stays thin and the core remains adapter-free.

### Risks
- CLI parsing logic can start leaking into the planning core if the boundary is not kept explicit.
- Human-readable output may become unstable if it depends on map iteration or unsorted data.

### Ready for Proposal
Yes — this slice is ready to be proposed as a minimal `plan` command that loads `catalog/bootstrap.toml`, calls `BuildPlan`, and prints deterministic text output with static environment facts.
