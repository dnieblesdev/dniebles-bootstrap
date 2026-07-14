# Proposal: Dotfiles Base Resolution Context

## Intent

Make local dotfiles base failures diagnosable without presenting unvalidated paths as canonical or attempting dotlink before a safe base exists.

## Scope

### In Scope
- Typed base diagnostic transport: source, attempted candidate, canonical path only after validation, selected modules, and safe cause.
- Resolve and validate empty, missing, unsafe, and non-directory bases; classify filesystem failures with `errors.Is` and `errors.As`.
- Gate executable derivation on a canonical, validated base; render base diagnostics consistently.

### Out of Scope
- Dotlink execution, report parsing, stderr handling, multi-cause handling, deduplication, and execution-result behavior.
- Any dependency on the current monolithic artifact.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `dotfiles-provider`: distinguish attempted and canonical base paths; preserve typed base-resolution diagnostics and prevent executable construction before validation.
- `execution-contracts`: carry and render base-resolution diagnostics without changing dotlink execution/report semantics.

## Approach

Centralize source selection, canonicalization, validation, and filesystem-error classification in the base resolver. Return a typed resolution context to provider and renderer boundaries. Derive `<canonical>/bin/dotlink` only after that context contains a validated canonical base.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/execution/dotfiles_base.go` | Modified | Base resolution and typed validation failures |
| `internal/execution/dotfiles_provider.go` | Modified | Validated-base transport and executable gate |
| `internal/execution/types.go` | Modified | Base diagnostic representation |
| `cmd/dbootstrap/render.go` | Modified | Safe base-diagnostic rendering |
| `internal/execution/*dotfiles*_test.go` | Modified | Resolution and error-classification coverage |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Diagnostic contract leaks an unvalidated path | Low | Assert canonical field is empty on every failed resolution |
| Filesystem errors lose identity | Medium | Test wrapped errors with `errors.Is`/`errors.As` |

## Rollback Plan

Revert this isolated change; it introduces no migration or persisted state.

## Dependencies

- No current monolithic artifact.
- Successor `dotlink-execution-failure-context` depends on this typed base transport baseline.

## Success Criteria

- [ ] Empty, missing, unsafe, and non-directory bases produce typed, safe diagnostics.
- [ ] No executable is derived before base canonicalization and validation succeed.
- [ ] Implementation forecast remains at or below 800 changed lines.
