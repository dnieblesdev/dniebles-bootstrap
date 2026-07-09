# Proposal: Apply Confirmed Reporting

## Intent

Improve `dbootstrap apply --yes` reporting so users can immediately tell what may have changed the machine and what is still advisory or unsupported. This is a reporting-only slice that reduces confusion before adding more mutation providers.

## Scope

### In Scope
- Render a clearer execution summary using user-facing categories: `changed`, `unchanged`, `not supported yet`, and `failed`.
- Make confirmed-mode framing explicit: only brew-backed `tool` and `package` steps may mutate the machine; runtime, dotfile, non-brew, and other unsupported steps remain non-mutating.
- Use `not supported yet` in per-step user-facing output while preserving internal execution statuses unless a later phase finds a concrete need for model changes.
- Render a special empty-state sentence when the selected plan has no actionable/selected steps instead of a zero-count summary table.
- Update rendering-focused tests and the `apply-command-dry-run` capability wording as needed.

### Out of Scope
- Adding new provider behavior or mutation paths.
- Adding an apt provider, dotfiles execution, dotlink, clone, sparse checkout, retry, or concurrency behavior.
- Adding catalog targets or changing catalog semantics.
- Changing execution report model fields unless implementation proves rendering cannot meet the requirement safely.

## Affected Areas

| Area | Impact | Notes |
|------|--------|-------|
| `cmd/dbootstrap/render.go` | Modified | Clarify execution report summary, confirmed-mode preamble, per-step wording, and empty-state output. |
| `cmd/dbootstrap/main.go` | Possibly modified | Only if apply wiring needs small presenter/context adjustments; no provider behavior changes. |
| `cmd/dbootstrap/*_test.go` | Modified | Cover confirmed summary categories, unsupported wording, and empty selected-plan output. |
| `openspec/specs/apply-command-dry-run/spec.md` | Modified | Capture mode-specific reporting expectations for the apply command. |

## Risks

| Risk | Mitigation |
|------|------------|
| Users may infer `--yes` mutates every selected resource. | Confirmed-mode copy must state that only brew-backed `tool`/`package` steps are eligible to change the host. |
| User-facing categories may drift from internal statuses. | Keep the mapping in rendering/tests; do not rename internal statuses unless necessary. |
| A reporting slice may accidentally expand execution scope. | Limit implementation to rendering, tests, and spec wording; reject apt/dotfiles/new provider work. |
| Empty output could hide that no work was selected. | Use an explicit empty-state sentence for zero actionable/selected steps. |

## Rollback

Revert the rendering/test/spec changes. Since this slice introduces no host mutation, catalog changes, data migrations, or new providers, rollback restores the current execution report text only.

## Success Criteria

- `apply --yes` output clearly distinguishes `changed`, `unchanged`, `not supported yet`, and `failed` work.
- Confirmed-mode output explicitly says only brew-backed `tool`/`package` steps may have changed the machine.
- Unsupported/non-mutating work is described as `not supported yet` in user-facing output.
- A selected plan with zero actionable/selected steps renders a clear empty-state sentence instead of a zero-count table.
- The main product outcome is met: users are less likely to misunderstand what `--yes` actually changed.
