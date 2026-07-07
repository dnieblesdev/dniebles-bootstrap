# Proposal: Apply Safety Contract

## Intent

Make `dbootstrap apply` safe-by-default before real installers or Homebrew bootstrap are wired. The CLI must preserve today's non-mutating behavior while reserving `--yes` as the future explicit opt-in required before host mutation can exist.

## Scope

### In Scope
- Define apply mode semantics: default apply is non-mutating, `--dry-run` is explicit non-mutating, and `--yes` is future mutation opt-in.
- Reject ambiguous flag combinations such as `--dry-run --yes` with a clear usage error.
- Report the selected apply mode so users can tell dry-run/noop output from future confirmed execution.

### Out of Scope
- Wiring real installers, `CommandRunner`, Homebrew bootstrap, remote scripts, dotlink, clones, or host mutation.
- Adding catalog raw commands or shell-first execution metadata.
- Implementing actual confirmed mutation behind `--yes` in this slice.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `apply-command-dry-run`: strengthen apply requirements with explicit dry-run/default-safe/confirmed-mode flag semantics and conflict validation.

## Approach

Add apply-specific flag parsing around the existing plan target surface. Keep execution on kind-aware noop installers only. Introduce an apply mode concept for rendering and validation, with `--yes` accepted as an explicit confirmed mode marker but still non-mutating until a later installer-wiring change defines and implements real execution.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/dbootstrap/main.go` | Modified | Parse apply safety flags and reject unsafe/ambiguous combinations. |
| `cmd/dbootstrap/render.go` | Modified | Show selected apply mode in execution output. |
| `internal/execution/` | Modified | Preserve noop-only apply execution boundary; no real runner connection. |
| `openspec/specs/apply-command-dry-run/spec.md` | Modified | Source capability for safety-contract deltas. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Users misread apply as real install | Med | Render mode explicitly and keep results `not_implemented`. |
| `--yes` appears to mutate now | Med | Document/report confirmed mode as accepted but not wired to real installers in this slice. |
| Future Homebrew path bypasses opt-in | High | Make explicit confirmation semantics a spec gate before provider work. |

## Rollback Plan

Revert the change folder and any follow-up apply flag/rendering changes; current apply behavior returns to the existing noop execution report with no host mutation.

## Dependencies

- Archived `catalog-installer-metadata` and `command-runner` slices.
- Existing noop execution contracts and apply dry-run spec.

## Success Criteria

- [ ] `dbootstrap apply` remains non-mutating without flags.
- [ ] `dbootstrap apply --dry-run` is explicit non-mutating.
- [ ] `dbootstrap apply --dry-run --yes` fails clearly.
- [ ] No Homebrew bootstrap, remote script, raw command, or real `CommandRunner` mutation is wired.
