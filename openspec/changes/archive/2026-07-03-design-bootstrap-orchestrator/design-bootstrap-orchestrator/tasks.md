# Tasks: Bootstrap Orchestrator

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 180-260 |
| Estimated changed files | 10-12 |
| 800-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr-default |
| Chain strategy | none |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: none
800-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Normalize OpenSpec docs and update repo orientation | PR 1 | Include specs, README, AGENT, and verification notes in one reviewable slice. |

## Phase 1: OpenSpec Alignment

- [x] 1.1 Update `openspec/changes/design-bootstrap-orchestrator/proposal.md`, `design.md`, and all `specs/*.md` to match the accepted alignment around first-run bootstrap, dotfiles boundary, catalog-in-repo, and CLI-now/TUI-later direction.
- [x] 1.2 Ensure the `bootstrap-entrypoint`, `environment-detection`, `catalog-planning`, `dotfiles-integration`, `bootstrap-orchestration`, and `repository-guidance` specs use consistent terms and no conflicting scope.

## Phase 2: Repository Orientation Docs

- [x] 2.1 Rewrite `README.md` to cover purpose, current status, goals/non-goals, profile and point install flows, domain-first direction, first-run bootstrap entrypoint concept, dotfiles boundary, catalog direction, and CLI now / TUI later.
- [x] 2.2 Add `AGENT.md` with operating rules: no implementation before specs/design, generated artifacts in English, `.atl/` stays local/ignored, dotfiles boundary, SDD/OpenSpec + Engram workflow, Bash first-run wrapper boundary, and one-core/two-thin-interfaces guidance.

## Phase 3: Verification

- [x] 3.1 Review the updated OpenSpec artifacts against `openspec/changes/design-bootstrap-orchestrator/design.md` and confirm requirements, scenarios, and terminology still align.
- [x] 3.2 Review `README.md` and `AGENT.md` for scanability, explicit non-goals, and consistency with the approved design and spec bundle.
- [x] 3.3 Verify the change remains documentation/spec-only and does not introduce application code, runtime wiring, or catalog implementation files.

## Phase 4: Work-Unit Commit Guidance

- [x] 4.1 Keep documentation edits in a single Conventional Commit sized as one reviewable work unit; do not split by file type alone.
- [x] 4.2 If line count grows unexpectedly, regroup by outcome so specs, README, and AGENT still tell one clear story in review.
