# Tasks: Dotfiles Base Resolution Context

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 650–780 (implementation and tests) |
| 400-line budget risk | High |
| 800-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR: resolver/context, gate/transport, rendering, and verification |
| Delivery strategy | exception-ok |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: High
800-line budget risk: Low
Size exception: accepted by the user for this target

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | Resolve and validate typed base context, gate the executable, transport diagnostics, render base facts, and verify | Single PR | One independently verifiable target; size exception accepted |

## Phase 1: Resolver and Context RED

- [x] 1.1 **RED** — Extend `internal/execution/dotfiles_base_test.go` with table-driven env/home, empty, missing, unsafe, and non-directory cases; assert attempted-only diagnostics, empty canonical fields, and selected modules.
- [x] 1.2 **RED** — Add resolver tests for wrapped filesystem failures using `errors.Is` and `errors.As`, with injected functions and `t.TempDir()` where filesystem identity is exercised.
- [x] 1.3 **RED** — Add `internal/execution/dotfiles_provider_test.go` cases proving invalid resolution cannot expose an executable or call the fake runner, while valid resolution must use canonical identity.

## Phase 2: Resolver and Context GREEN

- [x] 2.1 **GREEN** — Modify `internal/execution/dotfiles_base.go` to centralize source selection, canonicalization, safety/directory validation, safe causes, and `%w` error preservation; publish canonical paths only after validation.
- [x] 2.2 **GREEN** — Modify `internal/execution/types.go` and `internal/execution/provider.go` to carry `DotfilesBaseDiagnostic`, original error, validated base, and private validation proof without exposing rejected canonical paths.

## Phase 3: Gate, Transport, and Rendering RED

- [x] 3.1 **RED** — Extend `internal/execution/dotfiles_provider_test.go` for empty-env terminal behavior and canonical `<base>/bin/dotlink` derivation; assert zero runner calls on every invalid prerequisite.
- [x] 3.2 **RED** — Add `internal/execution/dotfiles_installer_test.go` coverage that failed dotfile results retain source, attempted candidate, modules, safe cause, and no canonical path.
- [x] 3.3 **RED** — Extend `cmd/dbootstrap/render_test.go` with exact deterministic success/failure base output, including canonical-versus-attempted labels and absence of unrelated fields.

## Phase 4: Gate, Transport, and Rendering GREEN

- [x] 4.1 **GREEN** — Modify `internal/execution/dotfiles_provider.go` to reject missing, errored, empty, or mismatched validation proofs before prerequisite checks, executable construction, or runner use.
- [x] 4.2 **GREEN** — Modify `internal/execution/dotfiles_installer.go` to attach the dedicated base diagnostic to failed step results without changing existing result contracts.
- [x] 4.3 **GREEN** — Modify `cmd/dbootstrap/render.go` to render only deterministic base source, attempted/canonical identity, modules, and safe cause.

## Phase 5: Verification

- [x] 5.1 Run `gofmt` on changed Go files and focused `go test ./internal/execution ./cmd/dbootstrap`.
- [x] 5.2 Run `go test ./...`, `go vet ./...`, and confirm the diff remains at or below 800 changed lines and contains no out-of-scope behavior.
