# Design: wire dotfiles into confirmed apply

## Overview

This slice composes the existing dormant dotfiles execution core into `dbootstrap apply --yes` only. Planning stays unchanged and remains non-mutating. Default apply and dry-run keep using noop execution for dotfile resources. Confirmed apply gains one additional eligible execution path: selected `dotfile` plan steps may be dispatched to the existing dotfiles installer/provider, which validates the local repository and invokes `dotlink` through `CommandRunner`.

## Key decisions

1. **Confirmed-only composition**
   - The real dotfiles provider is installed into the apply execution graph only when `--yes` is set.
   - Default apply and `--dry-run` keep their current noop/non-mutating behavior.

2. **Selected plan steps are the only input**
   - Dotlink modules come from selected `dotfile:<name>` plan steps.
   - The CLI must not derive modules from unrelated catalog metadata, descriptions, dependencies, shell snippets, or unselected resources.

3. **Provider core remains owner of local validation**
   - The existing provider/base resolver validates canonical base path, source, `bin/dotlink`, and module directories.
   - Missing base, missing dotlink, and missing module become failed dotfile execution results.

4. **CLI composition seams are explicit**
   - Add or extend apply command construction so tests can inject:
     - fake `CommandRunner`;
     - fake base resolver/filesystem/prerequisite seams as supported by the provider core;
     - existing fake planning/execution dependencies.
   - Production construction can use the real local runner only in confirmed mode.

5. **Reporting explains eligibility and context**
   - When dotfiles execution is eligible or attempted, output includes:
     - canonical base path;
     - base source (`DBOOTSTRAP_DOTFILES_DIR` or home convention);
     - selected module names;
     - dotfile result status and understandable failure text.
   - Confirmed-mode safety copy must say brew-backed tool/package steps and selected dotfile resources may have changed.

## Flow

1. CLI parses `apply` flags and rejects conflicting safety flags as today.
2. CLI loads catalog and builds the plan through the existing planning pipeline.
3. Apply mode is selected:
   - default: noop execution/reporting;
   - dry-run: noop execution/reporting;
   - yes: real brew-backed tool/package path plus real selected-dotfile path.
4. In `--yes`, selected plan steps with kind `dotfile` are dispatched through the existing dotfiles installer/provider.
5. Provider validates local prerequisites and either:
   - calls `CommandRunner` with dotlink arguments for selected modules; or
   - returns a failed result before command execution.
6. Reporter renders summary plus dotfiles context when the dotfiles execution path is used.

## Failure handling

- Missing/unsafe base path: failed dotfile result with base-source/path context if available; no command; confirmed apply exits non-zero.
- Missing `bin/dotlink`: failed dotfile result; no command; confirmed apply exits non-zero.
- Missing selected module: failed dotfile result naming the missing module; no command; confirmed apply exits non-zero.
- Runner error/timeout: failed dotfile result using existing provider-core semantics; no retry or fallback acquisition; confirmed apply exits non-zero and does not report the step as changed.

## Non-goals and guardrails

- Do not implement clone, pull, submodule, fetch, remote acquisition, sparse checkout, apt, bootstrap entrypoint, rollback, repair, or symlink tracking.
- Do not change planning semantics.
- Do not make `internal/dotfiles` mutating.
- Do not call `exec.Command` or a shell from tests or provider logic outside the existing runner abstraction.

## Test approach

- Start with RED CLI tests proving default apply and dry-run with dotfile resources do not use the fake/real runner and remain non-mutating/not supported.
- Add RED confirmed apply test for `--yes --resource dotfile:bash` with fake seams, asserting dotlink is requested through fake `CommandRunner`, output includes canonical base/source/modules, and result reports changed.
- Add RED missing-prerequisite tests for missing base, missing dotlink, and missing module; assert failed result, no runner invocation, and non-zero CLI exit.
- Add RED runner failure and timeout tests; assert failed result, no changed status, no retry/fallback acquisition, and non-zero CLI exit.
- Add guard tests/assertions that no clone/pull/submodule/fetch/remote acquisition commands are requested.

## Review workload forecast

Review workload is moderate. The desired code change should be mostly CLI composition and reporting tests, but the test matrix crosses safety modes, output text, and failure states. Reviewers should pay close attention to the confirmed-only gating and to ensuring the provider core remains unchanged except for test-proven tiny fixes.


## Exit status rule

Confirmed `apply --yes` MUST return a non-zero exit status if any eligible real execution step reports `failed`, including dotfiles prerequisite failures, runner failures, and timeouts. The command should still render the execution report before returning. Default apply and `--dry-run` remain non-mutating and keep existing behavior.
