# Standard implementation review policy

## Binding

| Field | Value |
| --- | --- |
| Target | `bootstrap-entrypoint` OpenSpec change |
| Scope | The uncommitted implementation in `cmd/dbootstrap/main.go` and `cmd/dbootstrap/main_test.go`, plus this change's approved OpenSpec artifacts |
| Review type | Standard implementation review |
| Lineage | To be assigned by native `gentle-ai review-start` |

## Budget

- One exhaustive review sweep per required lens.
- One general refutation pass for BLOCKER or CRITICAL candidates only.
- At most two fix rounds; this transaction does not authorize fixes.

## Evidence policy

- Reuse the approved Judgment Day and verification evidence already recorded for this change.
- Do not rerun tests or alter production or test code in this review transaction.
- Treat only native workflow outputs as transaction, ledger, receipt, bundle, and gate-context evidence.

## Intended untracked manifest

`intended-untracked.json` is the complete allowlist for untracked files introduced by this change and its native review workflow.
