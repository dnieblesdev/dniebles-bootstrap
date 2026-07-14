# Design: Bootstrap CLI Entrypoint

## Technical Approach

Expose `dbootstrap bootstrap` as a second command name for the current apply
orchestration, not as a second workflow. The command dispatcher and help text
remain name-aware; every operational step is executed by one shared function.
This preserves the contracts in `bootstrap-entrypoint` and the
`apply-command-dry-run` delta while retaining the current `apply` behavior.

## Architecture Decisions

| Decision | Options / tradeoff | Choice and rationale |
|---|---|---|
| Share at the CLI boundary | Duplicate `runApply`; add a command package; extract one small helper | Add a small shared apply-like runner in `cmd/dbootstrap/main.go`, with thin `apply`/`bootstrap` dispatch wrappers if needed for existing tests. It is the narrowest seam that covers parsing through exit mapping without a new layer. |
| Make bootstrap discoverable without a new workflow | Hidden command; duplicate help; root listing plus command help | Add `bootstrap` and its concise description to root usage, then parameterize command usage by name. The shared flag parser remains one implementation while `bootstrap --help` provides focused guidance. |
| Retain injected dependencies | Introduce an application container; use package globals | Reuse the existing detector and execution factory variables. Tests already restore these seams with `t.Cleanup`; a container would broaden an alias-only change. |
| Preserve error semantics | Reclassify apply errors; share existing paths | Propagate current errors unchanged: parser/argument errors return `exitUsage`; catalog and planning failures render their current errors and return `exitFailure`; only confirmed reports containing failed results return `exitFailure`. |

## Data Flow

```text
apply | bootstrap dispatch
        │ (name for usage only)
        v
shared parse + validation ──invalid──> usage / exit 2
        │
        v
catalog load → environment + installation + dotfiles + config detection
        │                                      │
        v                                      v
planning.BuildPlan → planning diagnostics → provider-aware Runner composition
        │                                            │
        └── planning failure → render / exit 1       v
                                      Runner.Run → AppendHomebrewBootstrap
                                                     │
                                                     v
                                           renderExecutionReport → exit map
```

`--profile` and repeatable `--resource` are parsed and deduplicated before any
catalog, detector, runner, or provider work. No explicit target, malformed
resource syntax, unexpected positionals, `--dry-run --yes`, and `--sudo`
without `--yes` are usage failures. Syntactically valid unknown profiles or
resources are catalog-dependent: they continue through the existing shared
catalog/detection/config/planning path, then use apply's diagnostic, report,
and failure exit behavior. Neither command pre-validates them differently.

## File Changes

| File | Action | Description |
|---|---|---|
| `cmd/dbootstrap/main.go` | Modify | Dispatch `bootstrap`; list it in root help; factor the complete current `runApply` body into one name-aware shared path; make parser/command usage name-aware while retaining current `apply` behavior. |
| `cmd/dbootstrap/main_test.go` | Modify | Add table-driven root-help, command-help, syntactic-validation, semantic-failure, and command-parity cases using existing seams. Preserve existing apply cases. |
| `openspec/changes/bootstrap-entrypoint/design.md` | Create | This technical design. |

`cmd/dbootstrap/render.go` is intentionally unchanged: both names call the
existing renderer, so report ordering and wording cannot drift.

## Interfaces / Contracts

No exported domain interface or package is added. The private shape is limited
to a command-name argument at the shared CLI boundary, conceptually:

```go
func runApplyLike(command string, args []string, stdout, stderr io.Writer) int
func parseApplyFlags(command string, args []string, stderr io.Writer) (...)
```

The command string is presentation metadata only. It MUST NOT alter the
`planning.PlanRequest`, selected catalog, `applyMode`, detectors, runner
composition, report, or exit classification. `apply` retains its flags,
defaults, output, and exit behavior; `bootstrap` requires the same explicit
target already required by the shared parser.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| CLI unit | Root-help listing/description, `bootstrap --help`, syntactic target and mode validation | Table-drive `run` with buffers; assert exit code/output and zero catalog/detector/runner calls for help and usage cases. |
| CLI parity | Default, dry-run, yes, and yes+sudo | Run identical explicit profile/resource arguments under both names with stubbed facts, states, command existence, factories, and recording runner; compare plan/report, exit status, and command requests. |
| Failure parity | Unknown catalog target, missing catalog/config/environment, planning diagnostics, and a later confirmed failure after an earlier success | Inject each failure through existing seams; assert apply/bootstrap have identical semantic report/exit behavior, no command on prerequisites, ordered partial report, and no rollback claim. |
| Regression | Existing apply suite | Keep current direct apply/parser tests intact; add bootstrap rows rather than replacing them. |

Use `t.TempDir()` for catalog fixtures and `t.Cleanup` to restore package seams.
The execution tests stay host-independent; no real home directory, package
manager, or sudo invocation is permitted.

## Migration / Rollout

No migration required. The change is additive. Rollback removes only the
`bootstrap` dispatch/help/tests; the shared path remains the existing apply
behavior.

## Open Questions

None.
