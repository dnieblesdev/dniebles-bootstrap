# Proposal: wire-dotfiles-apply-yes

## Intent

Wire the existing dormant dotfiles execution core into `cmd/dbootstrap apply --yes` so a confirmed apply can invoke `dotlink` for selected `dotfile` resources in the plan. This is a narrow follow-up to `dotfiles-execution-provider-core`: it composes the already-added execution provider at the CLI boundary without expanding dotfiles ownership, acquisition, rollback, or non-confirmed mutation.

## Scope

In scope:

- Compose the existing `internal/execution` dotfiles provider/installer into confirmed `dbootstrap apply --yes` only.
- Keep default `apply`, `apply --dry-run`, and `plan` non-mutating.
- Add CLI composition seams so tests can inject a fake `CommandRunner` and fake base resolver/filesystem dependencies; tests must not execute real `dotlink`.
- Execute dotlink only for selected dotfile resources that appear in the built plan.
- Render canonical dotfiles base path/source and selected modules before or with the result when dotfiles execution is eligible or attempted.
- Update confirmed-mode copy to say brew-backed tool/package steps and selected dotfile resources may have changed.
- Render missing base, missing `bin/dotlink`, and missing module as failed dotfile execution steps with understandable text.

Out of scope:

- Provider-core redesign beyond tiny fixes proven necessary by tests.
- Clone, pull, submodule, fetch, sparse checkout, or remote acquisition.
- Bootstrap entrypoint work.
- Symlink rollback, repair, tracking, or ownership.
- Apt provider or other package-provider expansion.
- Mutation in `plan`, default `apply`, or `apply --dry-run`.

## Affected areas

- `cmd/dbootstrap` apply composition root and output/reporting.
- CLI test seams for execution dependencies.
- Existing `internal/execution` wiring usage, with provider-core behavior remaining substantially unchanged.
- OpenSpec deltas for `apply-command-dry-run` and `execution-contracts`.

## Risks

- Accidentally enabling mutation outside `--yes` would violate apply safety expectations.
- Reporting may imply bootstrap owns dotfiles internals if output is not careful about the external-provider boundary.
- CLI wiring could bypass injectable seams and make tests depend on host dotfiles state.
- Missing-prerequisite handling could be ambiguous if failures are collapsed into generic unsupported results.

## Rollback

Revert the CLI composition that installs the real dotfiles provider for `apply --yes`, leaving the noop execution path in place. The provider core remains dormant and can stay in the codebase because it is only mutating when composed with a real runner by the confirmed apply path.

## Success criteria

- `dbootstrap apply` and `dbootstrap apply --dry-run` with selected dotfile resources remain noop/non-mutating and do not use a real runner.
- `dbootstrap apply --yes --resource dotfile:bash` can run dotlink through an injected fake `CommandRunner` in tests and reports a changed dotfile result.
- Missing base, missing `bin/dotlink`, or missing selected module reports a failed dotfile step and does not invoke a real command.
- Output for the dotfiles execution path includes canonical base path/source and selected modules.
- Confirmed-mode copy mentions both brew-backed tool/package steps and selected dotfile resources may have changed.
- Tests and implementation do not introduce clone, pull, submodule, fetch, remote URL, sparse checkout, or apt execution behavior.

## Proposal assumptions

- The product intent is a small confirmed-mode bridge, not a broader reconciliation workflow.
- A dotfile resource name maps directly to a dotfiles module name using the existing execution-core mapping.
- Missing local dotfiles prerequisites are user-actionable failures, not prompts to acquire or repair the repository.
- No additional interactive proposal question round was run because this delegated task is in auto execution mode and already includes slice scope and acceptance expectations.
