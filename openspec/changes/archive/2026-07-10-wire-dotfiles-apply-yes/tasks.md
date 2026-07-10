# Tasks: wire-dotfiles-apply-yes

## RED

- [x] 1. Add failing apply safety tests for dotfile resources.
  - [x] 1.1 `dbootstrap apply --resource dotfile:bash` remains noop/not supported and does not use a dotfiles command runner.
  - [x] 1.2 `dbootstrap apply --dry-run --resource dotfile:bash` remains noop/not supported and does not use a dotfiles command runner.
  - [x] 1.3 `dbootstrap plan --resource dotfile:bash` remains non-mutating and does not compose dotfiles execution.

- [x] 2. Add failing confirmed dotfiles execution tests with fake seams.
  - [x] 2.1 `dbootstrap apply --yes --resource dotfile:bash` uses a fake base/filesystem seam and fake `CommandRunner`.
  - [x] 2.2 Assert the fake runner receives dotlink for selected module `bash` only.
  - [x] 2.3 Assert the result reports changed when the fake runner succeeds.
  - [x] 2.4 Assert output includes canonical base path, base source, and selected module names.

- [x] 3. Add failing safe-failure tests.
  - [x] 3.1 Missing dotfiles base reports a failed dotfile step with understandable text, no command invocation, and non-zero CLI exit.
  - [x] 3.2 Missing `bin/dotlink` reports a failed dotfile step with understandable text, no command invocation, and non-zero CLI exit.
  - [x] 3.3 Missing selected module reports a failed dotfile step with understandable text, no command invocation, and non-zero CLI exit.
  - [x] 3.4 Fake command-runner failure reports a failed dotfile step, does not report the step as changed, performs no retry/fallback acquisition, and exits non-zero.
  - [x] 3.5 Fake command-runner timeout reports a failed dotfile step, does not report the step as changed, performs no retry/fallback acquisition, and exits non-zero.

- [x] 4. Add failing guard tests for excluded acquisition/provider behavior.
  - [x] 4.1 Assert no clone, pull, submodule, fetch, remote URL, sparse checkout, or apt command is requested by the dotfiles apply path.
  - [x] 4.2 Assert unselected or non-dotfile resources are not passed to the dotfiles installer.

## GREEN

- [x] 5. Add or refine CLI composition seams.
  - [x] 5.1 Allow apply command tests to inject fake dotfiles `CommandRunner` and fake base resolver/filesystem/prerequisite dependencies.
  - [x] 5.2 Keep production real runner construction gated behind `apply --yes` only.
  - [x] 5.3 Avoid broad provider-core changes; make only tiny fixes if RED tests prove the existing core cannot be composed safely.

- [x] 6. Wire confirmed apply dotfiles execution.
  - [x] 6.1 In `--yes`, dispatch selected dotfile plan steps through the existing dotfiles installer/provider.
  - [x] 6.2 Keep default apply and `--dry-run` on noop execution for dotfile resources.
  - [x] 6.3 Preserve existing confirmed brew-backed tool/package behavior.

- [x] 7. Update reporting.
  - [x] 7.1 Render canonical dotfiles base path/source and selected modules when dotfiles execution is eligible or attempted.
  - [x] 7.2 Render missing base/dotlink/module as failed dotfile steps with understandable text.
  - [x] 7.3 Update confirmed-mode copy to mention brew-backed tool/package steps and selected dotfile resources may have changed.

- [x] 8. Verify and refactor.
  - [x] 8.1 Run focused tests for apply command, execution contracts, and dotfiles provider composition.
  - [x] 8.2 Run the repository test suite if practical.
  - [x] 8.3 Refactor only after tests pass, preserving confirmed-only gating and injected seams.

## Corrective remediation

- [x] 9. **RED/GREEN — preserve understandable dotfiles prerequisite failures.**
  - Add focused failing assertions proving confirmed apply renders distinct understandable causes for missing base path, missing `bin/dotlink`, and missing selected module while returning non-zero and avoiding runner invocation.
  - Preserve or safely map provider error detail into the dotfile `StepResult.Message` with the minimal production change.
  - Verify: `go test ./cmd/dbootstrap ./internal/execution && go test ./...`.

## Review workload forecast

Moderate review workload. Expect several CLI tests and small composition/reporting edits. There is no local line limit for this slice, but warn reviewers if implementation expands beyond CLI wiring/reporting plus tiny provider-core fixes, because that likely indicates scope creep into dotfiles acquisition, rollback, or provider redesign.
