# Proposal: Bootstrap CLI Entrypoint

## Intent

Provide a discoverable `dbootstrap bootstrap` front door without creating a second installation workflow. It must be safe for first use: users select scope explicitly and must explicitly confirm any mutation.

## Scope

### In Scope
- Add `bootstrap` as a thin CLI alias/front door over one shared apply pipeline: parsing, environment detection, catalog loading, planning, execution, reporting, and exit handling.
- List and describe `bootstrap` in root help; provide command-specific help for explicit targets and safety modes.
- Require `--profile` and/or repeatable `--resource`; reject only syntactic target/mode errors before catalog or host work. Preserve shared catalog-dependent semantic validation for unknown profiles/resources.
- Preserve apply safety and failure behavior: default/dry-run are non-mutating; only `--yes` executes eligible work; only `--yes --sudo` enables APT sudo.
- Add parity-focused CLI tests using existing injectable seams.

### Out of Scope
- Default profile/catalog changes, TUI, providers, package-presence redesign, automatic escalation, downloads/shell orchestration, retries, transactions, or rollback guarantees.

## Capabilities

### New Capabilities
- `bootstrap-entrypoint`: Explicit-target bootstrap command with apply-equivalent orchestration and user-visible outcomes.

### Modified Capabilities
- `apply-command-dry-run`: Extend the command-surface safety and reporting contract to require parity for `bootstrap`, while retaining `apply` behavior.

## Approach

Route both commands through one application function (or have `bootstrap` delegate to the existing apply path). Reuse `buildPlan`, provider eligibility/wiring, renderers, and exit logic; do not duplicate detector, Runner, or provider composition.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `cmd/dbootstrap/main.go` | Modified | Shared dispatch, target/mode parsing, and application flow |
| `cmd/dbootstrap/main_test.go` | Modified | Alias parity and safety/exit coverage |
| `openspec/specs/apply-command-dry-run/spec.md` | Modified | Bootstrap parity requirement |

## Compatibility & Migration

`dbootstrap apply` remains supported and behaviorally unchanged. `bootstrap` is additive; callers must provide an explicit target, so no migration or implicit default selection is introduced.

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Alias drifts from apply | Med | One shared pipeline and parity tests |
| Implicit scope/mutation | Low | Require profile or resource; retain confirmation gates |
| Partial confirmed execution | Med | Keep ordered reports and rerun-oriented recovery; make no transaction claim |

## Rollback Plan

Remove the `bootstrap` dispatch/help surface and its tests. `apply` and all existing provider behavior remain intact because the entrypoint adds no catalog, domain, or provider changes.

## Dependencies

- Existing apply pipeline, catalog, planning, execution, and rendering contracts.

## Success Criteria

- [ ] `bootstrap` and `apply` produce equivalent plan, report, and exit outcomes for the same explicit flags.
- [ ] Root help lists and describes `bootstrap`; `bootstrap --help` provides command-specific usage without probing or executing.
- [ ] Missing target, malformed resource, and invalid flag combinations return usage errors without probing or mutating; syntactically valid unknown targets follow apply's shared semantic failure path.
- [ ] Default and `--dry-run` never execute commands; `--yes`/`--yes --sudo` retain existing Brew, APT, and dotfiles limits and failures.
