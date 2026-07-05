## Exploration: installer-provider-interfaces

### Current State

The `dniebles-bootstrap` codebase has a fully working pure planning pipeline:

- **`internal/planning`** — domain types (`Plan`, `PlanStep`, `PlanStepResult`, `PlanStepStatus`) and a pure `BuildPlan()` function. All statuses are planning-time: `planned`, `skipped`, `attention_required`, `already_installed`, `error`.
- **`cmd/dbootstrap`** — CLI `plan` command that loads catalog, runs detectors, calls `BuildPlan()`, and renders results. No `apply` command exists.
- **Infrastructure adapters** — all read-only detectors: `internal/environment` (OS/arch/distro/WSL), `internal/state` (PATH-based tool/runtime presence), `internal/dotfiles` (module-dir existence), `internal/config` (config-key file existence), `internal/catalog/toml` (TOML decoding).
- **No execution code anywhere** — zero installer, runner, command-execution, or dotfiles-operation code exists. The README explicitly says "Missing: execute through installer and dotfiles provider adapters."

The planning domain follows clean architecture strictly: it imports nothing outside `internal/planning`, uses caller-supplied state, and never probes the host. All detectors live in separate packages and are injected at the CLI composition root.

### Affected Areas

- **New: `internal/execution/`** — execution boundary contracts (Installer, Runner, DotfilesProvider interfaces, StepResult types). Does not exist yet.
- **`internal/planning/`** — No change. Planning remains pure. Execution contracts live outside, referencing planning types only as input data.
- **`internal/dotfiles/` (detector)** — No change. The existing read-only detector stays; the new `DotfilesProvider` interface is a separate execution contract, not a detector replacement.
- **`cmd/dbootstrap/`** — No change in this slice. The `apply` command wiring comes later.
- **`go.mod`** — No new dependencies needed (interface-only code requires zero third-party imports).
- **Active specs** — Existing specs (environment-detection, installation-state, config-state-awareness, dotfiles-provider, point-install-planning) are unaffected. Execution contracts are new domain territory.

### Approaches

1. **Single `internal/execution` package with contracts only** — Create one package containing `Installer` interface, `Runner` type/interface, `DotfilesProvider` interface, and execution-time `StepResult`/`StepStatus` types. Include noop/stub mocks so contracts compile and are testable without real execution.

   - **Pros**: Single cohesive package; low package-count overhead; matches the `internal/planning` pattern of one focused package; easy to discover and review; noop mocks prove contract viability immediately.
   - **Cons**: Could grow large when implementations arrive; single package mixes runner and provider concerns (though both are execution-layer).
   - **Effort**: Low (contracts + noop stubs + tests ≈ 150-250 lines)

2. **Three separate packages: `internal/runner`, `internal/installer`, `internal/provider`** — Split execution contracts by concern. The runner orchestrates; the installer interface lives in its own package; the dotfiles provider interface lives in its own package.

   - **Pros**: Maximum separation of concerns; each package stays tiny; clear ownership boundaries for future implementations.
   - **Cons**: Premature split for contracts-only slice; cross-package references between runner and installer; more files and directories for what is currently interface definitions only; higher review cognitive load.
   - **Effort**: Medium (≈ 250-350 lines across packages, more boilerplate)

3. **Contracts inline in `internal/planning`** — Add execution interfaces/stubs to the planning package.

   - **Pros**: Zero new packages; everything in one place.
   - **Cons**: Violates clean architecture — planning becomes impure by defining execution contracts; `BuildPlan` purity is compromised by association; contradicts the instruction "execution contracts live outside `internal/planning`"; hard No.
   - **Effort**: Low (but architecturally wrong)

### Recommendation

**Approach 1** — single `internal/execution/` package with contracts and noop stubs. This mirrors the established pattern (`internal/planning` is one focused package with types + builder). When implementation arrives (real installers, command runner), contracts can be split if needed — but premature splitting for an interface-only slice adds more structure than value.

The package would contain:

| File | Purpose |
|------|---------|
| `types.go` | `StepResult`, `StepStatus` (execution-time), `ExecutionReport` |
| `installer.go` | `Installer` interface (`Install(ctx, step) (StepResult, error)`, `SupportedKind() ResourceKind`) |
| `runner.go` | `Runner` struct (maps `ResourceKind → Installer`, iterates plan steps) |
| `provider.go` | `DotfilesProvider` interface (`PrepareModules`, `RunDotlink`) |
| `noop.go` | Noop stubs for Installer, Runner, DotfilesProvider — return `not_implemented` or identity |
| `*_test.go` | Table-driven tests proving contracts compile, noop stubs return expected statuses, runner delegates to correct installer by kind |

Design decisions embedded in this approach:

- **Execution statuses are separate from planning statuses.** Planning statuses (`planned`, `skipped`, `attention_required`, `already_installed`, `error`) describe *what the plan intends*. Execution statuses (`installed`, `failed`, `skipped`, `not_implemented`) describe *what actually happened*. No overlap, no confusion.
- **`Installer` interface is kind-scoped.** Each installer handles one `ResourceKind` (tool, runtime, package, dotfile). The `Runner` dispatches by kind. This matches the Go idiom of small, focused interfaces.
- **`DotfilesProvider` is a separate execution contract**, not an installer. The orchestrator design doc describes it as handling partial clone, sparse checkout, and `dotlink` invocation — cross-cutting operations that span multiple dotfile resources, not per-resource installs.
- **`Runner` accepts a plan and returns execution results.** It does not modify the plan or planning state. It is a pure orchestration boundary: given installers and a plan, produce results.
- **Noop stubs return `not_implemented` status.** This satisfies "avoiding actual command execution or mutation" while proving the contracts wire together. Future slices replace noops with real implementations.

### Risks

- **Status vocabulary confusion**: Planning has `planned`/`skipped`/`attention_required`; execution has `installed`/`failed`/`skipped`/`not_implemented`. Both share `skipped` but with different semantics (planning: environment mismatch; execution: pre-condition not met). Document clearly; use separate Go types (`PlanStepStatus` vs `StepStatus`).
- **Premature Runner design**: The Runner interface/struct must be general enough to support future real implementations (parallel execution, retries, timeouts) without locking in design now. Mitigation: start with synchronous sequential Runner; if concurrency/retry is needed later, add a `RunnerOption` or extract a `Runner` interface.
- **DotfilesProvider interface scope creep**: The orchestrator design mentions git operations (partial clone, sparse checkout) and `dotlink` invocation. The interface should expose these as named operations, not as a generic "execute" method. Risk: over-specifying git internals before the real implementation proves what's needed. Mitigation: start with `EnsureModules(ctx, modules) error` and `RunDotlink(ctx, modules) error` — high-level enough to defer git strategy.
- **No real integration test possible**: With only noop stubs, there is no integration path to prove the full `plan → run` pipeline end-to-end. This is expected and acceptable — the stubs prove contracts compile and delegate correctly; integration comes when real installers exist.

### Ready for Proposal

Yes. The planning pipeline is complete. The architecture has clear seams: planning domain is pure, CLI composition root wires detectors, and execution has no contracts yet. This slice fills that gap cleanly — one package, interface definitions, noop stubs, table-driven tests. No existing code changes. Proceed to `sdd-propose`.
