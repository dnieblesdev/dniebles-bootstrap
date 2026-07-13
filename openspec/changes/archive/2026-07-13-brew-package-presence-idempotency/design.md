# Confirmed Homebrew Formula Presence Design

Confirmed `apply` and `bootstrap` will preflight only eligible Brew **package** steps, then pass a typed transient result to the execution runner. A package is skipped only after a positive `brew list --formula <InstallMetadata.Package>` result; every inconclusive result is a failed/attention execution outcome and never reaches an installer.

## Scope and boundaries

- Probe only `ResourceKindPackage` resources with `Install.Provider == "brew"` and a trimmed, non-empty `Install.Package`.
- Use the metadata package string exactly; do not use the resource ID or `Presence.Name`.
- Probe only after the pure plan is built and only in confirmed modes (`--yes`, with or without `--sudo`). `plan`, default `apply`, and `--dry-run` retain their current command-presence detection and make no Brew lookup or command-runner call.
- Use one read-only request per eligible selected step: `CommandRequest{Executable: "brew", Args: []string{"list", "--formula", packageName}, Timeout: brewFormulaPresenceTimeout}`. The request has no shell, environment override, sudo, retry, fallback, cask, APT, or version behavior.
- This is an in-memory execution decoration, not catalog/TOML schema or persisted state. The planner remains OS-probe-free.

## Components and contracts

### `internal/planning`

Add a transient package-presence field to `PlanStep` and a small enum:

```go
type PackagePresence string
const (
    PackagePresenceUnchecked PackagePresence = ""
    PackagePresenceInstalled PackagePresence = "installed"
    PackagePresenceAbsent    PackagePresence = "absent"
    PackagePresenceUnknown   PackagePresence = "unknown"
)
```

`BuildPlan` leaves this field `Unchecked`; existing command-presence semantics and all catalog data remain unchanged. The field is set only on a copied execution plan after confirmed-mode probing. It allows the runner to distinguish a positively installed package from an untrusted unknown result without overloading `PlanStepStatus` or changing existing config-attention behavior.

### `internal/state`

Add `BrewFormulaDetector` beside the existing command `Detector`, with injected seams:

```go
type BrewFormulaDetector struct {
    CommandExists func(string) bool
    Runner        execution.CommandRunner
    Timeout       time.Duration
}
func (d BrewFormulaDetector) Detect(ctx context.Context, plan planning.Plan) map[planning.ResourceRef]planning.PackagePresence
```

It iterates the already ordered plan, records only eligible Brew package refs, and performs at most one lookup and one request per eligible step. Nil seams, unavailable `brew`, invalid metadata, or an unavailable runner yield `Unknown`; they never panic or invoke an installer. A package that is not eligible is omitted/`Unchecked` and gets no Brew lookup.

Use a private `classifyBrewFormulaResult(result execution.CommandResult) planning.PackagePresence`:

| Condition | Classification | Execution consequence |
|---|---|---|
| `Status == succeeded` and `ExitCode == 0` | `installed` | unchanged, installer skipped |
| completed Brew `list` exit explicitly recognized as the Brew formula-not-installed diagnostic (exit 1 plus the exact supported `No such keg` diagnostic form) | `absent` | existing Brew installer remains eligible |
| anything else: missing Brew, nil runner, timed out/not-run, runner error, malformed success result, non-zero unrecognized output/status | `unknown` | failed/attention result, installer skipped |

The absence recognizer must require both the expected failed command status/exit code and the supported Brew diagnostic; exit code 1 alone is never absence. This deliberately prefers `unknown` when Homebrew output is localized or changes rather than authorizing a mutation. Query output is not surfaced verbatim.

Use a package constant `brewFormulaPresenceTimeout` (short, fixed, e.g. `30 * time.Second`) in `internal/state`; the existing `OSCommandRunner` enforces it with `context.WithTimeout`. The detector adds no retry or fallback.

Add `ApplyBrewFormulaPresence(plan planning.Plan, presence map[ResourceRef]PackagePresence) planning.Plan` (or an equivalently named copy helper) in `internal/state`. It returns a copied plan and assigns the transient field only for classified eligible refs. `Unknown` also appends a stable, sanitized attention reason such as `"Homebrew formula presence could not be determined; no mutation attempted"`; it does not alter the original planning result.

### `internal/execution`

Extend `Runner.Run` pre-dispatch rules, preserving step order:

1. Retain the existing tool/runtime command-presence skip.
2. If an eligible Brew package has transient `PackagePresenceInstalled`, append `StepStatusSkipped` with exactly `already installed; no mutation attempted`; do not select or call an installer.
3. If an eligible Brew package has transient `PackagePresenceUnknown`, append `StepStatusFailed` with `Homebrew formula presence could not be determined; no mutation attempted` and a package-presence sentinel error; do not select or call an installer.
4. `Absent` and `Unchecked` use the existing provider-aware dispatch unchanged.

Eligibility is revalidated at this boundary (package kind, provider `brew`, trimmed package metadata) so a manually constructed or malformed plan cannot gain a false skip. `hasFailedExecutionResult` already turns the unknown result into confirmed-mode exit failure, while `Runner` continues later ordered steps.

## CLI composition and data flow

In `cmd/dbootstrap/main.go`, keep `buildPlan` unchanged. In `runApplyLike`, after plan errors are handled and only when `isConfirmedMode(mode)`:

1. Lazily create the existing injected OS command runner only if `planHasEligibleBrewPackagePresence(plan)` is true.
2. Construct `state.BrewFormulaDetector` with `brewCommandExists`, that runner, and the fixed timeout.
3. Detect and decorate a copy of `result.Plan` for execution.
4. Build the normal apply runner and execute the decorated plan.

Both `apply` and `bootstrap` share this path. `--sudo` is passed only to the existing APT installer; it is not passed to detection. Keep `appendApplyBootstrap` advisory behavior unchanged. Add a narrow injectable detector factory/function variable in `main.go` only if needed to make CLI composition tests independent of real PATH/process state; do not add a new CLI flag.

The same lazy `CommandRunner` instance may serve the read-only query and later installers, but test call order must prove each eligible query precedes its corresponding install. No runner is constructed for safe modes solely because a Brew package exists.

## File changes

| File | Change |
|---|---|
| `internal/planning/types.go` | Add transient `PackagePresence` enum/field; no TOML/catalog schema change. |
| `internal/state/brew_formula_detector.go` | New eligibility, injected lookup/runner query, conservative classifier, and execution-plan copy helper. |
| `internal/state/brew_formula_detector_test.go` | Strict-TDD unit coverage for eligibility, exact argv/timeout, all classifications, and no-probe cases. |
| `internal/execution/runner.go` | Gate eligible package installed/unknown states before installer dispatch. |
| `internal/execution/runner_test.go` | Prove installed skip, unknown failed/no dispatch, absent dispatch, revalidation, and continued ordered execution. |
| `cmd/dbootstrap/main.go` | Confirmed-mode-only detector composition and plan decoration. |
| `cmd/dbootstrap/main_test.go` | Composition tests for `apply`/`bootstrap`, safe modes, runner ordering, and zero install calls for installed/unknown. |
| `README.md` | Update **Confirmed reruns**: add the narrowly defined positive Brew formula exception and conservative unknown failure behavior; retain all exclusions. |

No APT, cask, installer, catalog, parser, privilege, or shell files change.

## Test-first plan

1. Add failing state tests using fake lookup and `CommandRunner`: metadata package differs from ref/presence name; exact `brew list --formula jq` request and timeout; installed; supported explicit absent; missing Brew; nil runner; timeout; runner error; unrecognized non-zero; malformed success; non-Brew/blank/cask-like metadata no probe.
2. Implement the detector/classifier until those tests pass.
3. Add failing runner tests for installed skip and unknown failure (zero installer calls), absent dispatch, malformed manually supplied states, and later-step continuation/order; implement the guard.
4. Add failing CLI table tests for confirmed `apply` and `bootstrap`: query-before-install order, installed output unchanged/no mutation, unknown failed/no install/non-zero, and mixed plan continuation. Add default and dry-run assertions that neither lookup nor runner factory is called for package presence.
5. Implement composition and README wording. Run focused packages, then mandatory `go test ./...`.

Use table-driven tests and existing recording/sequence fakes. Do not execute real Brew in tests.

## Rollout and rollback

This is an additive, under-400-author-changed-line slice. Roll out with the focused unit/CLI suite and `go test ./...`; no migration or configuration is required. If a Brew diagnostic proves incompatible, its result safely becomes `unknown` (no installation). Rollback is one revert of the transient detector, runner gate, CLI wiring, tests, README wording, and delta specs; prior confirmed installer dispatch resumes.
