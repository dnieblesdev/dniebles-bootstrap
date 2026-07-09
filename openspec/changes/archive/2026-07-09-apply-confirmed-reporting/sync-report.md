# Sync Report: apply-confirmed-reporting

## Status

synced

## Domains Synced

- `apply-command-dry-run`

## Canonical Files Updated

- `openspec/specs/apply-command-dry-run/spec.md`

## Requirement Changes Applied

### ADDED

- `Apply execution summary is always rendered`
- `Empty selected plans render an explicit empty state`

### MODIFIED

- `Apply renders execution mode-specific reporting`
  - added user-facing `not supported yet` wording for internal `not_implemented`
  - added confirmed `--yes` framing that only brew-backed `tool` / `package` steps may have changed the machine
  - clarified unsupported / non-brew work remains non-mutating and not supported yet

### PRESERVED Existing Active Requirements

- `Apply command exists with plan-style target flags`
- `Apply reuses the planning pipeline`
- `Apply remains strictly non-mutating`
- `Apply mode is explicit and safe by default`
- `Conflicting safety flags are rejected`
- `Confirmed mode only wires brew-backed installs`
- `No apply command is introduced` history under `## REMOVED Requirements`

## Active Same-Domain Collisions

- None detected.

## Destructive Sync / Approvals

- None required.
- Corrective merge only; no canonical requirement removals were introduced.

## Validation Performed

- Confirmed the corrected canonical spec contains the restored active requirements.
- Confirmed no `(Previously: ...)` text remains in the active canonical spec.
- Confirmed the new requirement headings are present.
- Confirmed the preserved `## REMOVED Requirements` section and history remain present.
- Quick content checks run with `grep` against `openspec/specs/apply-command-dry-run/spec.md`.

## Structured Status and Action Context

- Change: `apply-confirmed-reporting`
- Artifact store: `both` (OpenSpec + Engram)
- OpenSpec directory present and authoritative.
- Execution mode: auto; not workspace-planning.
- Verification report status: PASS; no blockers reported.
- This sync corrected a prior bad merge that had replaced the active canonical spec with only the delta slice.

## Next Recommended Phase

- `sdd-archive` when the parent/orchestrator is ready.
