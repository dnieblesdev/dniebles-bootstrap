# Exploration: environment-detection-adapter

### Current State
`internal/planning` is already pure: `BuildPlan` only consumes caller-supplied `EnvironmentFacts` and never probes the host. Today `cmd/dbootstrap plan` still injects static `linux/amd64` facts, so the CLI is the current adapter boundary. The existing SDD spec for environment detection already expects OS, distro, WSL, and architecture facts before planning.

### Affected Areas
- `cmd/dbootstrap/main.go` — currently hardcodes environment facts; this is the first consumer of real detection.
- `internal/planning/types.go` — defines the domain facts contract and should stay adapter-free.
- `internal/planning/builder.go` — already filters by facts; no direct change needed for this slice.
- `internal/environment` or `internal/platform` — likely home for host probing adapters and test seams.
- `openspec/changes/environment-detection-adapter/*` — exploration artifact for the next slice.

### Approaches
1. **Thin environment adapter package** — add `internal/environment` (or `internal/platform`) with a small detector that returns `planning.EnvironmentFacts` and uses injectable file/env/runtime providers in tests.
   - Pros: keeps `internal/planning` pure; smallest useful slice; easy to test without host coupling.
   - Cons: needs a package boundary decision now.
   - Effort: Medium

2. **CLI-local detection wiring** — implement detection directly in `cmd/dbootstrap` and pass facts into planning.
   - Pros: fastest possible wiring; no new package.
   - Cons: leaks host-probing concerns into the CLI; harder to reuse from future TUI/app code.
   - Effort: Low

### Recommendation
Use a thin adapter package, preferably `internal/environment` unless the repo wants a broader host/platform abstraction. Detect only the facts already required by planning now: OS, architecture, distro, and WSL. Keep detection logic isolated behind injectable readers/providers so tests can assert Linux/WSL/distro behavior deterministically. Wire `cmd/dbootstrap plan` to consume detected facts, but keep override support out of this slice unless a later requirement needs it.

### Risks
- Distro detection is inherently heuristic and may misidentify containers, WSL distros, or derivative releases.
- WSL detection can be flaky if it relies on one signal; tests should cover multiple probe paths and absence cases.
- If the CLI directly probes `runtime.GOOS`/files, tests will become host-dependent and brittle.

### Ready for Proposal
Yes — propose a minimal detection adapter that returns planning facts, is host-testable, and keeps installers, dotfiles runtime, command runner, TUI, and apply/install commands out of scope.
