# Design: CLI Plan Command

## Technical Approach

Add the first executable boundary as `cmd/dbootstrap`. Keep `main` as process wiring and put testable command behavior in the same package (`run`) until more commands justify an internal CLI package. `plan` loads `catalog/bootstrap.toml` by default through `internal/catalog/toml.LoadFile`, calls `planning.BuildPlan`, and renders deterministic text from `planning.PlanResult`. This satisfies the cli-plan spec while preserving the current pure planning and adapter boundaries.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Package layout | Create only `cmd/dbootstrap`; no new internal CLI package yet. | `internal/cli` from day one. | The repo has no `cmd/` today and only one command; an internal package would add indirection before reuse exists. Move later when command count or shared CLI behavior grows. |
| CLI parsing | Use stdlib `flag` with manual `plan` subcommand dispatch. | Cobra/urfave/kingpin. | Proposal requires a dependency-free thin boundary; manual dispatch is enough for `plan --profile`. |
| Testable process shape | `main()` calls `os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))`; tests call `run(args, stdout, stderr)`. | Calling `os.Exit` inside command logic. | Keeps exit behavior real for users while avoiding process exits in tests. |
| Catalog path | Default `catalog/bootstrap.toml`; add optional `--catalog <path>` on `plan` for tests and explicit local use. | Hard-code only. | Spec requires default repo-local catalog; an override keeps error-case tests small without temp chdir or fixture mutation. |
| Environment facts | Use static facts constant/var in CLI, e.g. `EnvironmentFacts{OS:"linux", Arch:"amd64"}` and empty `ConfigState`. | OS probing via runtime or shell. | This slice explicitly proves wiring without host probing; facts are caller-supplied planning inputs. |
| Rendering | Deterministic renderer in `cmd/dbootstrap` over `PlanResult`; no planning text in `internal/planning`. | Add `String()`/rendering to planning. | Planning stays pure structured data; CLI owns human output. |

## Data Flow

```text
argv ──→ run/flag parsing ──→ toml.LoadFile(catalog path)
                         └──→ planning.BuildPlan(profile, static facts, empty state)
                                      └──→ renderPlanResult(stdout/stderr)
```

## File Changes

| File | Action | Description |
|---|---|---|
| `cmd/dbootstrap/main.go` | Create | CLI entrypoint, `run`, flag parsing, default catalog/facts, exit-code mapping. |
| `cmd/dbootstrap/render.go` | Create | Deterministic rendering helpers for steps, results, and diagnostics. |
| `cmd/dbootstrap/main_test.go` | Create | Command/run tests with buffer-backed stdout/stderr and exact output/error assertions. |
| `cmd/dbootstrap/render_test.go` | Create | Renderer ordering and status/diagnostic output tests if main tests become too dense. |

## Interfaces / Contracts

```go
const defaultCatalogPath = "catalog/bootstrap.toml"

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
func run(args []string, stdout, stderr io.Writer) int
func runPlan(args []string, stdout, stderr io.Writer) int
func renderPlanResult(w io.Writer, profile string, catalogPath string, facts planning.EnvironmentFacts, result planning.PlanResult)
func renderDiagnostics(w io.Writer, result planning.PlanResult)
```

Exit codes: `0` success; `2` usage errors such as missing subcommand/profile; `1` catalog load/decode or planning error results. Planning `PlanStepStatusError` results are rendered to stderr and make the command fail.

Output contract: stdout lists plan steps in `result.Plan.Steps` order with stable `kind:name`, status, description, dependencies, and attention reasons. stderr lists diagnostics/errors in `result.Results` order after filtering error/diagnostic entries. No map iteration in rendering.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Renderer formats planned, skipped, attention, and error/diagnostic results deterministically. | Table-driven tests over synthetic `planning.PlanResult`; exact string assertions. |
| Command | `plan --profile dev`, missing profile flag, unknown profile, invalid catalog path/input. | Call `run`/`runPlan` with `bytes.Buffer`; use `t.TempDir()` for catalog overrides; assert exit code/stdout/stderr exactly. |
| Integration | Repo fixture loads through real adapter and planner. | Normal Go test using default path from package working directory or `--catalog ../../catalog/bootstrap.toml`; no external commands. |
| E2E | Not included. | No install/apply/runtime behavior exists in this slice. |

## Migration / Rollout

No migration required. This adds a new executable package only; no persisted state, installers, dotfiles runtime, command runner, OS probing, apply/install flow, or TUI.

## Risks / Tradeoffs

- Manual `flag` dispatch will become noisy if many subcommands arrive; acceptable for one command.
- Static facts may skip OS-conditioned catalog entries unexpectedly; document as slice behavior and replace with a probing adapter later.
- Exact text tests are intentionally strict; they protect deterministic UX but require deliberate updates when wording changes.

## Open Questions

None.
