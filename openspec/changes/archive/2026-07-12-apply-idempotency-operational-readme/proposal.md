# Apply Idempotency and Operational README

## Intent

Make confirmed `apply` and `bootstrap` runs avoid redundant installer mutations for resources the existing state detector has already established as present. Align the README with the implemented command surface and its safety boundaries.

## Problem

Planning can report `already_installed`, but confirmed execution still dispatches installers for those steps. State detection also probes a resource ID instead of an explicitly configured `Presence.Name`. The README remains stale about apply/bootstrap availability and operational behavior.

## Goals

- Preserve planning as the source of `already_installed` and honor that status at the apply orchestration boundary.
- Report detected present resources in their original plan position as unchanged/skipped, with clear no-mutation messaging.
- Correct state detection to probe `Resource.Presence.Name` when configured.
- Retain execution for absent eligible resources, unsupported reporting, ordered output, continued execution, failure exit behavior, and no-op/dry-run behavior.
- Update the operational README for `plan`, `apply`, `bootstrap`, `--yes`, `--sudo`, `--dry-run`, idempotency limits, and partial-failure recovery.
- Follow strict TDD: add focused failing tests before implementation and run `go test ./...` after focused coverage passes.

## Non-goals

- Package-manager presence detection, version/configuration reconciliation, retries, persistent receipts, or rollback.
- Installer-level planning-state probes or changes to package provider behavior.
- Dotfile link-content convergence; module presence is not proof that links are current.
- Catalog schema changes, bootstrap acquisition/default-profile behavior, or shell-wrapper orchestration.

## Affected areas

- Apply orchestration and execution-result reporting in `cmd/dbootstrap` and execution integration seams.
- Command-presence lookup in `internal/state`.
- Focused command, state-detection, and mixed-plan tests.
- Operational sections of `README.md`.

## Acceptance criteria

- A confirmed run with a detected eligible tool/runtime makes no installer command-runner call for that step and reports it unchanged with explicit no-mutation wording.
- A mixed plan preserves order: detected steps are unchanged, absent eligible steps still dispatch, and unsupported resources remain not supported.
- Presence detection uses configured `Presence.Name`, proven with a resource whose presence target differs from its ID.
- Existing target validation, `--yes`/`--sudo`/`--dry-run` rules, alias parity, continued execution, and no-op behavior remain covered and unchanged.
- README command and flag guidance matches actual behavior and states that detected presence is neither package/version/configuration proof nor dotfile-link convergence.
- Focused tests and `go test ./...` pass.

## Risks and mitigations

- **Incorrect identity mapping:** resource ID, presence target, and package name can diverge. Use `Presence.Name` explicitly and cover a non-default fixture.
- **Over-skipping work:** only skip steps with planning status `already_installed`; do not infer it from missing config, unsupported status, or prior outcomes.
- **Misleading operations guidance:** document partial success and rerun-after-fix recovery without implying automatic retry, sudo, or rollback.

## Rollback

Revert the orchestration guard, presence-name lookup, and README update together. This restores prior dispatch behavior without data migration or persistent state cleanup.

## Proposal question round

This proposal assumes the first slice is limited to currently reliable detection and that users value avoiding known redundant mutations over broader convergence. Before final approval, product owners may want to confirm:

1. Is “already installed; no mutation attempted” the desired operator-facing wording, or should reports use existing terminology only?
2. Should reliable dotfile detector results participate in this slice, given that module presence does not prove links are current?
3. Is the intended operational promise strictly “avoid mutation when current detection proves presence,” rather than general apply idempotency?

## Success criteria

Confirmed reruns are observably non-mutating for reliably detected present resources, while all other execution semantics remain intact and the README accurately communicates the supported workflow and limits.
