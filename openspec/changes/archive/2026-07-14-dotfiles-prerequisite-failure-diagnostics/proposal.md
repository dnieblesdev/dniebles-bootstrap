# Proposal: Dotfiles Prerequisite Failure Diagnostics

## Intent

Make dotfiles failures actionable without weakening safety or claiming validation that did not occur. The acceptance anchor is a repository missing `bin/dotlink`.

## Scope

### In Scope
- Preserve operation, module(s), phase, safe candidate, and safe cause across resolution, prerequisite validation, command execution, and report validation.
- Add the minimal execution-owned failure field needed for prerequisite identity; retain `DotfilesBaseDiagnostic`, `DotfilesFailure`, typed errors, statuses, report translation, and command semantics.
- Render one human-readable, terminal-safe diagnostic without duplicate base context.
- Prove a missing runner yields non-zero apply, `dotfile:bash`/`bash`, prerequisite phase, runner candidate (not validated), missing-path cause, and zero runner calls.

### Out of Scope
- The unused legacy `DotfilesBaseReporter`/legacy provider compatibility seam; it will not be ported.
- `PlanStep.AttentionReasons` to `StepResult`, planning/configuration/provider redesign, parser redesign, new statuses, and unrelated refactors.
- Monolith cleanup; perform it separately after a functional-equivalence audit.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `dotfiles-provider`: prerequisite and report-validation failures retain truthful, safe diagnostic context.
- `execution-contracts`: execution reporting renders phase-specific causes once, without duplicate base context.

## Approach

Use bounded diagnostic completion on current contracts: keep resolution facts in `DotfilesBaseDiagnostic`, retain existing command/report failure transport, and add a narrowly scoped prerequisite carrier only if necessary. Render phase labels and sanitized causes; never invoke the runner or label a rejected candidate as validated.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/execution/{types,dotfiles_provider}.go` | Modified | Preserve failure identity and typed causes. |
| `cmd/dbootstrap/render.go` | Modified | Safe, deduplicated diagnostic rendering. |
| Focused execution and CLI tests | Modified | Cover all failure paths and zero-call prerequisite guard. |
| `openspec/specs/{dotfiles-provider,execution-contracts}/spec.md` | Modified | Contract deltas only. |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Candidate implied validated | Low | Separate attempted and canonical identities in tests/rendering. |
| Duplicate or unsafe output | Medium | Reuse bounded sanitization; assert one base context. |
| Scope expands from monolith | Medium | Retain contracts; forecast 150–300 authored lines, below 800. |

## Rollback Plan

Revert this isolated diagnostic slice and its tests/spec deltas; existing prerequisite rejection, non-zero exit, and zero runner-call behavior remain intact.

## Dependencies

- Current base-resolution and dotlink report contracts at `e576669`.

## Success Criteria

- [ ] Resolution, prerequisite, command, and report-validation failures show truthful operation, modules, phase, and safe concrete cause.
- [ ] The missing-`bin/dotlink` confirmed apply fails non-zero, identifies `dotfile:bash`/`bash` and the prerequisite phase, shows an unvalidated candidate and missing-path cause, and makes zero runner calls.
- [ ] Rendering is terminal-safe and emits base context at most once; existing statuses, typed errors, report translation, and command semantics remain unchanged.
