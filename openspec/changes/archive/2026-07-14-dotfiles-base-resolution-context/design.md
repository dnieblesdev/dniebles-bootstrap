# Design: Dotfiles Base Resolution Context

## Technical Approach

Keep base resolution in `internal/execution` as a pure, injected-filesystem boundary. The resolver will select the environment or home candidate, retain that attempted value, canonicalize it, and validate the canonical directory before publishing it. A small execution context carries the validated base, render-safe diagnostic fields, and the original error. Provider code may derive `<canonical>/bin/dotlink` only from that validated context. `cmd/dbootstrap` renders that context without interpreting command results. This implements both delta specs while preserving external dotfiles ownership.

## Architecture Decisions

| Decision | Options / tradeoff | Chosen rationale |
|---|---|---|
| Resolution ownership | Provider-local checks duplicate resolver logic; resolver owns selection and validation | Centralize source selection, canonicalization, safety checks, and `Stat` failures in `DotfilesBaseResolver` so every caller gets the same attempted/canonical distinction. |
| Diagnostic shape | Format errors early; carry structured context | Use `DotfilesBaseDiagnostic` for source, attempted candidate, optional canonical path, modules, and safe display cause; retain the actual wrapped error separately for `errors.Is`/`errors.As`. Strings alone would lose filesystem identity. |
| Canonical publication | Set canonical path after symlink resolution; set it after validation | Populate `ResolvedDotfilesBase.CanonicalPath` and diagnostic canonical path only after absolute/safe/directory validation succeeds. Failure returns a zero base and empty canonical field. |
| Execution gate | Construct executable eagerly; derive after validation | `DotfilesExecutionContext.validatedBase` is the internal proof. Reject contexts with an error or mismatched/empty proof before validating repository prerequisites or constructing the executable. |
| Rendering boundary | Embed details in failure messages; render structured data | Keep base rendering in `renderLinkDetails`; label valid values `canonical base` and rejected values `attempted candidate`, then emit only the diagnostic fields. |

## Data Flow

```text
env/home candidate
      -> ResolveWithDiagnostic(modules)
      -> canonicalize + validate directory
      -> DotfilesExecutionContext{Base, Diagnostic, Err}
             | failure: canonical empty; wrapped Err preserved
             v
provider gate -> validated canonical base -> executable path
             v
StepResult.BaseDiagnostic -> renderLinkDetails
```

An explicitly empty `DBOOTSTRAP_DOTFILES_DIR` is terminal: record `env`, an empty attempted candidate, and the sentinel cause; do not read the home fallback, derive an executable, or invoke a runner.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/execution/dotfiles_base.go` | Modify | Produce the typed diagnostic during source selection; wrap filesystem operations with `%w`; publish canonical state only after validation. |
| `internal/execution/types.go` | Modify | Define/adjust the render-safe base diagnostic attached to execution results. |
| `internal/execution/provider.go` | Modify | Keep the resolved base, diagnostic, original error, and private validation proof together. |
| `internal/execution/dotfiles_provider.go` | Modify | Consume one context, block invalid contexts before executable construction or runner use, and derive paths from its canonical base. |
| `internal/execution/dotfiles_installer.go` | Modify | Attach base diagnostics to dotfile step results without changing report semantics. |
| `cmd/dbootstrap/render.go` | Modify | Deterministically render canonical versus attempted base facts and the safe cause only. |
| `internal/execution/dotfiles_base_test.go` | Modify | Cover source selection, empty/missing/unsafe/non-directory bases, and wrapped filesystem identity. |
| `internal/execution/dotfiles_provider_test.go` | Modify | Prove canonical gating, executable derivation, and zero runner calls on invalid bases. |
| `internal/execution/dotfiles_installer_test.go` | Modify | Prove the diagnostic reaches failed dotfile results. |
| `cmd/dbootstrap/render_test.go` | Modify | Assert stable success and failure base output without unrelated execution detail. |

## Interfaces / Contracts

`DotfilesExecutionContext` remains the single handoff:

- `Base` is non-zero only after successful validation.
- `Diagnostic.CanonicalPath` is non-empty only when `Base.CanonicalPath` is valid.
- `Err` preserves the original/sentinel or wrapped filesystem cause; callers use `errors.Is` and `errors.As` on it.
- `validatedBase` is private and must equal `Base.CanonicalPath` before provider code derives an executable.

`DotfilesBaseDiagnostic` is display data, not an error transport. It contains source, attempted candidate, canonical path, selected modules, and a safe cause string.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Resolution states and error identity | Table-driven resolver tests with injected lookup, symlink, and stat functions; use `t.TempDir()` for real filesystem cases and assert `errors.Is`/`errors.As`. |
| Unit | Provider gate | Fake `CommandRunner`; assert invalid contexts leave executable absent and make zero calls, while valid contexts use canonical `bin/dotlink`. |
| Integration | Installer-to-render transport | Focused installer and renderer tests using `StepResult.BaseDiagnostic`; assert exact deterministic lines. |

Run focused `go test ./internal/execution ./cmd/dbootstrap`, then `go test ./...`. No external command integration test is needed for this slice.

## Migration / Rollout

No migration required. This is an in-process diagnostic contract with no persisted state or flag.

## Open Questions

None.
