# Proposal: Homebrew Bootstrap Provider

## Intent

Introduce a safe Homebrew bootstrap provider that detects when `brew` is missing and reports an explicit, reviewable bootstrap action. Because the official install path is remote-script based, this slice MUST NOT execute it; it prepares trustworthy planning/reporting before any future mutation wiring.

## Scope

### In Scope
- Detect missing Homebrew availability needed by Homebrew-backed resources.
- Produce a clear bootstrap action/report with official manual install instructions.
- Preserve `apply` safety: default, `--dry-run`, and current `--yes` remain non-mutating.

### Out of Scope
- Executing `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`.
- Installing target packages with `brew install`.
- Adding raw catalog command fields, shell-first metadata, or bypassing the apply safety contract.

## Capabilities

### New Capabilities
- `homebrew-bootstrap-provider`: Detects missing Homebrew and reports a provider-owned, non-mutating bootstrap action with manual instructions.

### Modified Capabilities
- `apply-command-dry-run`: Apply reporting may include Homebrew bootstrap actions, but all accepted modes remain non-mutating.
- `execution-contracts`: Adds provider-specific Homebrew bootstrap semantics that report planned/manual work without side effects.

## Approach

Add a small provider boundary in `internal/execution` that models Homebrew bootstrap separately from package installation. Use structured catalog provider metadata only to recognize Homebrew-backed needs; do not add shell fields. Wire CLI reporting so `dbootstrap apply` can show missing-brew bootstrap guidance while still using the existing noop/safety contract.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/execution/` | New/Modified | Homebrew bootstrap provider/report types; no remote script execution. |
| `cmd/dbootstrap/main.go` | Modified | Compose provider reporting behind existing apply safety modes. |
| `cmd/dbootstrap/render.go` | Modified | Render explicit Homebrew bootstrap/manual action guidance. |
| `openspec/specs/` | Modified/New | Add provider capability and update apply/execution deltas. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Remote script executed too early | High | This slice forbids execution, even with `--yes`. |
| Homebrew bootstrap confused with package install | Med | Keep bootstrap provider separate from target package installers. |
| Shell-first metadata sneaks into catalog | Med | Reuse structured provider metadata only. |

## Rollback Plan

Remove the Homebrew provider/reporting code and delta specs. `apply` returns to existing noop execution reports, with no catalog or host migration required.

## Dependencies

- Completed `apply-safety-contract` behavior from commit `443f0f1`.
- Official Homebrew install command used only as displayed manual instruction.

## Success Criteria

- [ ] Missing Homebrew is reported as an explicit bootstrap action with manual instructions.
- [ ] `dbootstrap apply`, `--dry-run`, and `--yes` perform no host mutation.
- [ ] No raw command metadata or target package installation is introduced.
