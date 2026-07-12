# Proposal: Bootstrap Entrypoint Primary Delivery Record

## Intent

Establish this change as the authoritative current delivery record for the already-implemented `dbootstrap bootstrap` entrypoint behavior. The implementation is not changed by this proposal.

## Scope

### In Scope

- Record the delivered bootstrap entrypoint behavior as the current OpenSpec change.
- Preserve the implementation's explicit-target, apply-equivalent orchestration and safety expectations for subsequent specification and verification artifacts.
- State the review-record constraint that requires this independent current record.

### Out of Scope

- Source-code, test, catalog, provider, or runtime behavior changes.
- Changes to the historical `openspec/changes/bootstrap-entrypoint/` directory.
- Reopening or converting the legacy review projection.

## Delivery Record

The existing `openspec/changes/bootstrap-entrypoint/` change remains historical evidence. Its legacy review projection cannot consume compact-v2 receipt authority, so it cannot serve as the authoritative current delivery record.

This independent change is the authoritative current record for the already-implemented bootstrap entrypoint behavior. It intentionally does not alter the older change or declare a formal relationship to it.

## Capabilities

### Current Capability Record

- `bootstrap-entrypoint`: `dbootstrap bootstrap` provides an explicit-target front door that follows the established apply pipeline and safety model.

## Expected Follow-up Artifacts

| Artifact | Purpose |
| --- | --- |
| Specification | Capture the delivered command, parity, validation, and safety contract. |
| Design | Record the shared-pipeline implementation shape. |
| Tasks and verification | Reconcile the existing implementation with the authoritative delivery record. |

## Success Criteria

- [ ] The new change records the delivered bootstrap entrypoint behavior without modifying implementation.
- [ ] The older `bootstrap-entrypoint` change remains intact as historical evidence.
- [ ] Subsequent artifacts use this change as the authoritative current delivery record.
