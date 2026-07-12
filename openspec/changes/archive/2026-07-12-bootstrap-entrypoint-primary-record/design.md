# Bootstrap Entrypoint Design

## Decision

`dbootstrap bootstrap` is an explicit-target entrypoint that reuses the existing apply pipeline. This change records that delivered design; it does not add a second planner, runner, or provider path.

## Help and discovery

Root help exposes `bootstrap` as an explicit-selection entrypoint to the safe apply workflow. The bootstrap command handles both `bootstrap -h` and `bootstrap --help` before flag parsing or runtime work; its command usage lists the explicit-target options (`--profile` and repeatable `--resource <kind:name>`) and safety options (`--catalog`, `--dry-run`, `--yes`, and `--sudo`). These help paths do not load the catalog, detect the environment, create a runner, or start execution.

## Execution path

1. The root dispatcher routes `bootstrap` to `runApplyLike("bootstrap", ...)`.
2. Bootstrap help is handled before flag parsing or runtime work.
3. `parseApplyLikeFlags` validates and normalizes the target and safety options into a `planning.PlanRequest`, catalog path, and `applyMode`.
4. `buildPlan` loads the catalog, collects read-only environment and state facts, and invokes the shared planner.
5. Planning errors render the shared plan diagnostics and fail before a runner is created.
6. `buildApplyRunner` selects either a non-mutating runner or the confirmed provider-aware runner.
7. The shared runner produces an execution report, which is rendered and classified using the existing apply rules.

This gives `bootstrap` the same catalog loading, detection, planning, execution, reporting, and exit behavior as `apply` for the same request and flags.

## Argument validation boundary

The command boundary is `parseApplyLikeFlags`. It owns only syntactic input validation and does so before `buildPlan` is called:

| Input rule | Rejection behavior |
| --- | --- |
| At least one `--profile` or repeatable `--resource <kind:name>` is required. | Render bootstrap usage and return the usage exit code. |
| Positional arguments are not accepted. | Render usage and return the usage exit code. |
| Resources must use a supported `kind:name` reference. | Render usage and return the usage exit code. |
| `--dry-run` and `--yes` cannot be combined. | Render usage and return the usage exit code. |
| `--sudo` requires `--yes`. | Render usage and return the usage exit code. |

A valid `--profile` and one or more repeatable `--resource` values may be supplied together. Valid resources are deduplicated, and the profile plus deduplicated resources are forwarded together as one `planning.PlanRequest` selection. Catalog membership and other semantic selection failures remain planner concerns: they occur through `buildPlan`, render normal diagnostics, and do not start execution.

## Safety modes

`applyMode` is selected only after argument validation:

| Mode | Flags | Behavior |
| --- | --- | --- |
| Default non-mutating | no `--yes` or `--dry-run` | Uses a noop runner. |
| Dry run | `--dry-run` | Uses a noop runner. |
| Confirmed | `--yes` | Builds provider-aware installers for eligible work. |
| Confirmed sudo | `--yes --sudo` | Uses the confirmed path and enables sudo only for eligible APT work. |

Non-mutating modes do not create the OS command runner. Confirmed mode creates it lazily only when selected work requires an eligible Brew, Linux APT, or dotfiles installer. Unsupported, unavailable, and non-provider-backed work remains represented by the shared runner rather than being made executable by bootstrap.

The execution report preserves ordered results. A failed result in confirmed mode produces a failure exit; there is no transaction or rollback guarantee for the command as a whole.

## Composition root

`cmd/dbootstrap/main.go` is the composition root. It wires concrete catalog loading, environment and state detection, installers, command execution, planning, and rendering at the CLI edge. The root keeps two responsibilities separate:

- `buildPlan` composes read-only catalog and detection dependencies before calling `planning.BuildPlan`.
- `buildApplyRunner` composes noop or provider-aware execution dependencies after a valid plan is available.

The `bootstrap` dispatcher branch changes only the command label passed to this existing orchestration. The label supplies bootstrap-specific help and usage while preserving the same underlying request, plan, runner, report, and exit classification as `apply`.

## Scope boundaries

This is a delivery-record design for existing behavior. It does not require or authorize source-code, test, catalog, provider, runtime, or behavior changes. It also does not modify or reclassify `openspec/changes/bootstrap-entrypoint/`, which remains historical evidence.

## Review checklist

- [ ] Root help identifies bootstrap as the explicit-selection safe-apply entrypoint.
- [ ] `bootstrap -h` and `bootstrap --help` list explicit-target and safety options without runtime side effects.
- [ ] Bootstrap reaches `runApplyLike` rather than a distinct execution pipeline.
- [ ] `--profile` and repeatable `--resource` values may be combined into one shared planner selection.
- [ ] Syntactic failures return before catalog loading, detection, runner creation, or provider work.
- [ ] Default and dry-run modes use non-mutating execution.
- [ ] Confirmed execution retains existing provider eligibility, reporting, and failure semantics.
- [ ] This current record remains independent of the older OpenSpec change.
