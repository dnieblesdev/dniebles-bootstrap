# Design: Environment Detection Adapter

## Technical Approach

Add `internal/environment` as a host-probing adapter at the CLI boundary. It detects runtime OS/arch, Linux distro, and WSL status through injectable sources, then maps the result into `planning.EnvironmentFacts` before `planning.BuildPlan`. `internal/planning` remains pure and receives facts only.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Package boundary | Create `internal/environment` | Put detection in `cmd/dbootstrap` or `internal/planning` | A reusable adapter keeps CLI thin and preserves planning as a domain package. |
| Planning mapping | `internal/environment` imports `internal/planning` and returns `planning.EnvironmentFacts` | Return duplicate local facts plus CLI mapper | Import direction is acceptable because adapter depends inward on domain types; planning does not import adapter. Avoiding a duplicate type is simpler until another consumer needs adapter-neutral facts. |
| Test seam | `Detector` with injected sources and `Detect()` default helper | Package globals or host reads in tests | Keeps tests deterministic and avoids host-dependent assertions. |
| Fallbacks | Missing optional env/files produce empty distro and `WSL=false`, not fatal errors | Fail plan on partial detection | Planning can still run with OS/arch; conservative unknowns are safer than false certainty. |

## Data Flow

```text
dbootstrap plan
  ├─ environment.Detect()
  │    ├─ runtime.GOOS / runtime.GOARCH
  │    ├─ env lookup
  │    └─ optional file reads: /etc/os-release, /proc/version, /proc/sys/kernel/osrelease
  └─ planning.BuildPlan(catalog, request, facts, state)
       └─ renderPlanResult(..., facts, result)
```

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/environment/detector.go` | Create | Defines `Detector`, source interfaces/functions, default sources, `Detect()`, and mapping to `planning.EnvironmentFacts`. |
| `internal/environment/osrelease.go` | Create | Parses `/etc/os-release` style content and returns conservative distro ID. |
| `internal/environment/detector_test.go` | Create | Table-driven fake-source tests for OS/arch, distro, WSL signals, and fallback behavior. |
| `cmd/dbootstrap/main.go` | Modify | Replace `staticEnvironmentFacts` with a package-level detection function used by `runPlan`. |
| `cmd/dbootstrap/main_test.go` | Modify | Override detection function in tests so expected output stays deterministic. |
| `internal/planning/*` | No change | Planning continues accepting caller-supplied `EnvironmentFacts` only. |

## Interfaces / Contracts

```go
package environment

type RuntimeSource func() (goos, goarch string)
type EnvSource func(key string) (string, bool)
type FileSource func(path string) (string, error)

type Detector struct {
    Runtime RuntimeSource
    Env     EnvSource
    ReadFile FileSource
}

func Detect() planning.EnvironmentFacts
func (d Detector) Detect() planning.EnvironmentFacts
```

Default runtime source returns `runtime.GOOS` and `runtime.GOARCH`. Default file source reads text from known optional paths. Optional read errors are swallowed. If runtime source returns blanks, the detector returns blanks rather than inventing values.

`/etc/os-release` parsing should prefer `ID`, trim quotes, ignore malformed lines, and leave distro empty when absent. WSL detection checks supported signals in deterministic order: env keys such as `WSL_DISTRO_NAME`/`WSL_INTEROP`, then kernel/proc text containing `microsoft` or `wsl` case-insensitively. Any positive signal sets `WSL=true`; absent/unreadable signals keep `false`.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Detector maps fake runtime/env/files into facts | Table-driven tests with fake functions; no real `runtime`, env, or filesystem assertions. |
| Unit | os-release parser handles quoting, comments, missing ID, malformed data | Pure parser tests. |
| CLI | `plan` uses detected facts in planner/render output | Override package-level detect function and assert stable stdout. |
| Integration/E2E | None for this slice | Host probing and installer behavior stay out of scope. |

## Migration / Rollout

No migration required. Rollback is deleting `internal/environment` and restoring the static CLI facts.

## Out of Scope

- User override flags for OS/arch/distro/WSL.
- Install/apply commands, command runners, package managers, dotfile runtime, TUI, or remote execution.
- Exhaustive distro normalization beyond conservative `ID` extraction.

## Open Questions

None.
