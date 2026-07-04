## Exploration: wire-installation-state-cli

### Current State

The `installation-state-detector` slice has been committed and archived. It delivered:

- **`internal/planning/types.go`**: `InstallationState` struct with `PresentResources map[ResourceRef]bool` and `PlanStepStatusAlreadyInstalled = "already_installed"`.
- **`internal/planning/builder.go`**: `BuildPlan` accepts `InstallationState` as its fifth parameter. After environment matching succeeds, status becomes `already_installed` for present resources, including when missing config exists (attention reasons remain attached as metadata).
- **`internal/state/detector.go`**: `Detector` struct with injectable `LookPath PathLookup` seam. Detects `tool` and `runtime` refs via PATH lookup. `package` and `dotfile` refs are excluded via `isDetectableKind()`. Defaults to `exec.LookPath` when seam is nil.
- **`internal/state/detector_test.go`**: Table-driven tests with injected fake `LookPath`. One test exercises the nil-seam default path with a non-existent executable name.

The CLI (`cmd/dbootstrap/main.go`) currently passes empty `planning.InstallationState{}` mechanically:

```go
result := planning.BuildPlan(
    catalog,
    planning.PlanRequest{Profile: *profile},
    facts,
    planning.ConfigState{},
    planning.InstallationState{},  // <-- always empty
)
```

The existing environment-facts test seam uses a package-level variable:

```go
var detectEnvironmentFacts = environment.Detect
```

Tests stub it:

```go
func stubEnvironmentFacts(t *testing.T, facts planning.EnvironmentFacts) {
    t.Helper()
    original := detectEnvironmentFacts
    detectEnvironmentFacts = func() planning.EnvironmentFacts { return facts }
    t.Cleanup(func() { detectEnvironmentFacts = original })
}
```

No equivalent seam exists for installation state.

The renderer (`cmd/dbootstrap/render.go`) already prints `[status]` labels mechanically from `PlanStepResult.Status`. Since `already_installed` is already a defined `PlanStepStatus`, the step line would render it automatically. Attention reasons are preserved by the planner even under `already_installed`, so the existing `attention:` line works without changes.

### Affected Areas

- `cmd/dbootstrap/main.go` — needs to call `state.Detect(catalog)` between catalog load and `BuildPlan`. Needs a package-level `detectInstallationState` variable for test seams. The `state` import must be added.
- `cmd/dbootstrap/main_test.go` — needs a `stubInstallationState` helper and new test cases for already-installed output. Existing test cases remain unchanged (they stub environment facts and will pass empty state implicitly).
- `cmd/dbootstrap/render.go` — no structural changes needed. The renderer mechanically prints status and attention reasons from the planner output. Verified: `already_installed` is already a `PlanStepStatus` constant and the step line renders `[status]` generically.
- `internal/state/detector.go` — no changes. Already handles `tool`/`runtime` detection with PATH lookup seam and correctly excludes `package`/`dotfile`.
- `internal/planning/` — no changes. `BuildPlan` and types already support `InstallationState`.

### Key Questions Answered

#### 1. Smallest useful slice

Exactly three changes:

1. **`main.go`**: Add `var detectInstallationState = state.Detect`. Call it after catalog load, pass result to `BuildPlan` instead of empty `InstallationState{}`.
2. **`main_test.go`**: Add `stubInstallationState` helper. Add test cases showing `already_installed` status for git (present on PATH in test fixture) and mixed already_installed/planned/attention_required output.
3. Update existing test expected output strings — the `"success uses adapter and planner with exact output"` case will change because `tool:git` will show `[already_installed]` instead of `[planned]` when the fixture marks it present.

This is ~30 lines of production code and ~50 lines of test changes.

#### 2. CLI calls internal/state without becoming an installer/runner

Follow the exact pattern already established by `internal/environment`:

```go
import "github.com/dnieblesdev/dniebles-bootstrap/internal/state"

var detectInstallationState = state.Detect
```

In `runPlan`:

```go
installation := detectInstallationState(catalog)
result := planning.BuildPlan(catalog, ..., facts, ..., installation)
```

The CLI never installs, never runs commands (PATH lookup only uses `exec.LookPath`, not `exec.Command`), and never mutates state. It remains a read-only detector fed into a pure planner. This is identical in spirit to `detectEnvironmentFacts` which reads OS/proc/env without side effects.

#### 3. Packages remain undetected/unknown

Yes. `internal/state/detector.go` already enforces this via `isDetectableKind()`:

```go
func isDetectableKind(kind planning.ResourceKind) bool {
    return kind == planning.ResourceKindTool || kind == planning.ResourceKindRuntime
}
```

`ResourceKindPackage` and `ResourceKindDotfile` return false and are skipped. The catalog contains `package:ripgrep` which will remain `[planned]` (not `[already_installed]`) because the detector never marks it present. This is correct — package manager checks require dpkg/rpm/brew integration, explicitly out of scope.

#### 4. Renderer already supports already-installed + attention reasons

The renderer prints `[status]` mechanically via `statusByRef` → `PlanStepResult.Status`. Since `PlanStepStatusAlreadyInstalled` is already defined, the step line already renders it:

```
1. tool:git [already_installed] Version control
   depends_on: none
   attention: none

2. runtime:go [already_installed] Go toolchain
   depends_on: tool:git
   attention: missing required config "go.env"
```

The planner preserves attention reasons (missing config keys) under `already_installed` status. The renderer's existing `step.AttentionReasons` display works without changes. No visual distinction (color/icon) is needed in this slice — that belongs in a future TUI/render slice.

#### 5. Host-independent testing via injected seam

Use the same pattern as `stubEnvironmentFacts`:

```go
var detectInstallationState = state.Detect

func stubInstallationState(t *testing.T, installation planning.InstallationState) {
    t.Helper()
    original := detectInstallationState
    detectInstallationState = func(planning.Catalog) planning.InstallationState {
        return installation
    }
    t.Cleanup(func() { detectInstallationState = original })
}
```

Test scenario: set `InstallationState{PresentResources: map[ResourceRef]bool{toolGit: true}}` and verify the output shows `[already_installed]` for git. Tests never touch the real host PATH. The nil-seam default path (`state.Detect` using `exec.LookPath`) is already tested in `internal/state/detector_test.go`.

#### 6. Out of scope — confirmed

| Concern | Status | Evidence |
|---|---|---|
| **Installers** | Out of scope | No installation code exists anywhere. `internal/planning` is pure data. |
| **Package manager checks** | Out of scope | `isDetectableKind()` excludes `package`. `dpkg`/`rpm`/`brew` are future concerns. |
| **Apply/install command** | Out of scope | `main.go` line 46 returns `unknown command "apply"`. |
| **Dotfiles runtime** | Out of scope | `isDetectableKind()` excludes `dotfile`. No dotfiles runtime exists. |
| **TUI** | Out of scope | CLI uses `fmt.Fprintf` to `io.Writer`. No Bubbletea or TUI code exists. |

### Approaches

Only one viable approach exists — the pattern is already established:

1. **Package-level variable seam** — Mirror `detectEnvironmentFacts` pattern.
   - Pros: Consistent with existing codebase. Tests remain host-independent. Mechanical to understand. Minimal change surface.
   - Cons: One more package-level variable (but the pattern is already accepted).
   - Effort: Low

A discarded alternative (direct call without seam) would make tests host-dependent or require a mock catalog, breaking the established pattern.

### Recommendation

Wire `state.Detect` into `runPlan` using the same package-level variable seam pattern as `detectEnvironmentFacts`. The renderer needs no changes. Add test cases for already_installed output.

This is the naturally smallest next slice: the detector exists, the planner understands state, only the wiring glue is missing. The change touches ~80 lines total (production + tests) and completes the closure of the installation-state vertical from detector → planner → CLI output.

### Risks

- Existing test `"success uses adapter and planner with exact output"` will change output because the catalog fixture will use real `exec.LookPath` by default. If `git` is on the test host's PATH, it will show `[already_installed]`. This must be handled by either (a) always stubbing installation state in that test, or (b) accepting host-dependent output. Recommendation: always stub installation state in the "success" test case (default to empty state), then add a separate test case that exercises already_installed with a populated fixture. This keeps the existing test deterministic.
- The `detectInstallationState` variable and `stubInstallationState` helper must be carefully named so they don't create confusion with environment-facts stubbing. Using `InstallationState` (not just `State`) in the name is already disambiguating.
- The `internal/state` import adds a new dependency to `cmd/dbootstrap`. This is the intended architecture — the CLI is the composition root that wires adapters into the domain.

### Ready for Proposal

Yes. Progress to `sdd-propose` for `wire-installation-state-cli`. The orchestrator should tell the user this is a low-risk wiring slice of ~80 lines that completes the state detection vertical.
