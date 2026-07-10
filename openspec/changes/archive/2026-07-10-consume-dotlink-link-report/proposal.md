# Proposal: Consume Dotlink Link Report

## Intent

Make confirmed Bootstrap dotfile linking report actual per-link outcomes from Dotlink’s structured JSON v1 report, rather than inferring results from human-readable stdout. This gives users truthful per-link results while preserving Dotlink’s ownership of linking and rollback.

## Scope

- Invoke the selected external Dotlink provider with `link --report=json MODULE...` through the existing `CommandRunner` seam.
- Accept stdout as a candidate JSON report whenever it is present, regardless of command exit status; stderr is non-contractual and is never parsed.
- Strictly parse one JSON v1 document, rejecting unknown fields, duplicate object keys at every object depth, trailing data, malformed data, unsupported versions, and semantic inconsistencies.
- Preserve valid failed/rollback reports—including entries, safe causes, and rollback details—even when Dotlink exits non-zero.
- Treat an absent, malformed, or inconsistent report with command failure as a generic safe failed result; reject a success report with failed exit or a failed report with successful exit as inconsistent failures.
- Add execution-owned per-link details to the module `StepResult`; module status remains an aggregate, not a lossy representation of every entry.
- Render each link’s changed, unchanged, failed, or rolled-back detail truthfully.
- Incorporate base-resolution diagnostics: source, attempted candidate, selected modules, and safe cause. Call a path `canonical base` only after canonicalization and validation succeed.
- Preserve confirmation-only execution, safe default/dry-run behavior, non-zero confirmed failure exits, and no remote/acquisition behavior.

## Non-goals

- Dotfiles repository or `dotlink` changes.
- Human stdout/stderr parsing fallback.
- Clone, pull, sparse checkout, submodules, remotes, or acquisition.
- Bootstrap-owned rollback, repair, or link lifecycle.
- Mutation in plan, default apply, or dry-run modes.

## Success criteria

1. Confirmed invocation uses exact `dotlink link --report=json MODULE...` argv.
2. The report parser rejects duplicate keys at top-level and every nested object level before domain decoding; no invalid report can fall back to human output.
3. Present stdout is parsed and validated independently of command status. Valid failed reports retain detail; missing, malformed, or status-inconsistent reports fail safely.
4. `StepResult` has an aggregate module status plus ordered per-link details that preserve outcome, source, target, safe cause, and rollback detail.
5. Aggregate status is skipped when all entries are unchanged, installed when one or more entries changed and none failed, and failed when any entry failed/rolled back or the aggregate report failed; confirmed apply exits non-zero for failed aggregates.
6. Base-resolution failures show an attempted candidate, never a falsely labeled canonical base.
7. Default apply and dry-run remain non-mutating; no remote/acquisition behavior is introduced; fake-runner fixtures and `go test ./...` cover the contract.

## Risks and safeguards

| Risk | Safeguard |
|---|---|
| Duplicate JSON keys alter decoded meaning | Detect duplicates recursively before decoding typed wire/domain values and fail closed. |
| Non-zero process exit discards useful failure report | Parse present stdout first; reconcile command and report status under explicit precedence rules. |
| Module summary conceals mixed entries | Preserve and render execution-owned per-link details separately from aggregate `StepStatus`. |
| Unresolved path is presented as trusted | Render it only as an attempted candidate until successful canonicalization and validation. |
| Scope broadens beyond explicit confirmation | Keep default/dry-run noop routing and prohibit fallback, acquisition, retry, and Bootstrap rollback. |

## Rollback

Revert the consumer and rendering changes. Bootstrap returns to prior provider behavior without changing Dotlink or filesystem state. No remote, acquisition, or Bootstrap-owned recovery exists.

## Delivery

One cohesive PR is expected. If parser/fixture coverage exceeds the review budget, split parser/translator tests from provider/CLI integration only at that behavioral boundary.
