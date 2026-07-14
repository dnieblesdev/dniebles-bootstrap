# Proposal: Dotlink Execution Failure Context

## Intent

Make confirmed dotlink failures actionable while preserving safe command, report, and error identities. This change consumes the merged validated canonical-base context; it does not resolve or validate bases.

## Scope

### In Scope
- Derive the executable only from the validated canonical base; report a missing runner without executing.
- Preserve command failures with exit code and sanitized, bounded stderr.
- Classify unavailable, invalid, and inconsistent reports; retain a valid failed report alongside a command failure.
- Compose independent execution and report causes so `errors.Is` and `errors.As` can discover each; render execution facts without duplicating base context.

### Out of Scope
- Base resolution, validation, source, attempted-candidate, or canonical-path behavior (merged dependency).
- Remote acquisition, rollback behavior, dotlink semantics, package managers, new capabilities, and success or dry-run behavior.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `dotfiles-provider`: preserve dotlink executable, runner, command, report, and multi-cause failure context.
- `execution-contracts`: expose and render execution failure detail without duplicating base diagnostics.

## Approach

Extend the existing execution-owned failure transport after prerequisite validation. Keep command execution behind `CommandRunner`; translate its failure and report-validation failure into explicit independently unwrap-able causes. Bound and sanitize stderr before storing it. Reuse the existing base diagnostic solely as input/context and render it once.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/execution/dotfiles_provider.go` | Modified | Command/report failure composition |
| `internal/execution/dotlink_report.go` | Modified | Invalid-report classification |
| `internal/execution/types.go` | Modified | Typed execution diagnostic transport |
| `internal/execution/dotfiles_installer.go` | Modified | Failed result translation |
| `cmd/dbootstrap/render.go` | Modified | Sanitized execution detail rendering |
| `internal/execution/*_test.go` | Modified | Failure-identity and bounded-output coverage |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Error identity is lost during composition | Medium | Assert `errors.Is`/`errors.As` for both causes |
| Diagnostics leak or flood terminal output | Medium | Bound and sanitize stderr; renderer tests |
| Base context is repeated | Low | Keep base rendering separate and assert one presentation |

## Rollback Plan

Revert this isolated change. It adds no migration or persisted state; the merged base-resolution contract remains intact.

## Dependencies

- Merged `dotfiles-base-resolution-context` at `068c315` supplies validated canonical-base context.

## Success Criteria

- [ ] Missing runner, execution failure, invalid report, and valid failed report produce safe failed results.
- [ ] Combined execution/report failures preserve both causes for `errors.Is` and `errors.As`.
- [ ] Renderer shows bounded sanitized execution detail once and does not duplicate base context.
- [ ] Estimated implementation is 703 changed lines; normal hard budget remains at or below 800.
