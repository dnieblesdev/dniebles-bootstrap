## Exploration: bootstrap-entrypoint

### Current State

`cmd/dbootstrap` already contains the complete narrow orchestration path that a
high-level entrypoint needs:

```text
parse target/safety flags
  -> catalogtoml.LoadFile
  -> environment, executable, config, and dotfiles detectors
  -> planning.BuildPlan
  -> execution.Runner
  -> execution report and exit code
```

`plan` and `apply` both use `buildPlan`, so catalog loading, environment facts,
profile/resource expansion, dependency ordering, presence detection, and
attention-required config reporting are already composed once. `BuildPlan` is
pure and deterministic; the CLI composition root owns host probing.

`apply` is currently the closest bootstrap command. Its default mode and
`--dry-run` use kind-aware noops and never construct an OS command runner.
Only `--yes` enables real execution. `--yes --sudo` is the sole path that
enables sudo-backed APT; `--sudo` alone and `--dry-run --yes` are usage errors.
Confirmed execution is limited to brew-backed tool/package steps, Linux
APT-backed tool/package steps, and selected dotfile steps. Runtime and other
providers remain unsupported/noop.

The existing provider composition is sequential and plan-ordered. It reports
all step outcomes, continues after a failed step, and returns failure only for
confirmed execution failures. Planning/load errors stop before execution.
Missing Homebrew is advisory: the report adds a manual action and does not
install Homebrew. Dotfiles execution reports aggregate module status plus
per-link outcomes and rollback details, while explicitly avoiding acquisition.

The CLI is non-interactive in its own control flow: it reads flags and emits
stdout/stderr, with no prompt or TUI. A subprocess such as `sudo` may still
require a terminal/password, so a non-interactive caller must opt into
`--yes --sudo` only when its environment can satisfy that command contract.

### Affected Areas

- `cmd/dbootstrap/main.go` — command dispatch, shared target parsing, safety-mode parsing, `buildPlan`, provider wiring, and exit semantics.
- `cmd/dbootstrap/render.go` — existing shared plan and execution presenters; no change is indicated because both command names use the same renderer.
- `internal/planning` — existing profile/bundle/resource selection and deterministic dependency ordering; no domain change indicated.
- `internal/catalog/toml` — existing catalog adapter and default `catalog/bootstrap.toml`; no catalog change indicated.
- `internal/environment`, `internal/state`, `internal/config`, `internal/dotfiles` — existing read-only detection seams used by `buildPlan`.
- `internal/execution` — existing Runner, noop contracts, Brew/APT installers, dotfiles provider, manual Homebrew action, and structured reports.
- `cmd/dbootstrap/main_test.go` and `cmd/dbootstrap/render_test.go` — current seams already inject detectors, command availability, installer factories, runners, and dotfiles prerequisites.
- `openspec/specs/apply-command-dry-run/spec.md` and `openspec/specs/execution-contracts/spec.md` — current safety and execution contracts that a bootstrap surface must preserve.

### Composition and Duplication Findings

- The plan pipeline is already shared through `buildPlan`; a new command must
  call it rather than reproduce detector/catalog/planner wiring.
- `runApply` owns the only end-to-end execution composition. A bootstrap entrypoint
  should delegate to that path or to a small shared application function rather
  than create a second Runner configuration.
- Provider eligibility is scanned separately by `planHasBrewBackedInstall`,
  `planHasAptBackedInstall`, and `planHasDotfileSteps`; this is acceptable for a
  thin slice but should not be duplicated for a second command.
- Homebrew presence is probed once while building confirmed installers and can
  be probed again while appending advisory bootstrap guidance. This is safe but
  is a small reporting/composition duplication.
- `aptCommandExists` is currently assigned `execution.BrewCommandExists`.
  Both use `exec.LookPath`, so behavior is equivalent today, but the name is a
  misleading seam and should be corrected only if the implementation slice
  touches this composition.

### Approaches

1. **New `bootstrap` CLI command delegating to the existing apply application path** — Add a user-facing command whose target and safety flags are the same as `apply`, then route it through one shared orchestration function. Keep `apply` as the explicit lower-level command and make `bootstrap` an intentionally thin alias/presenter entrypoint.
   - Pros: gives the workflow a discoverable high-level name; preserves one plan/apply pipeline, one safety contract, one report model, and existing test seams; leaves room for a future default profile without adding provider logic.
   - Cons: adds a command alias and help/docs surface; must define whether bootstrap requires an explicit target or supplies a catalog-defined default.
   - Effort: Low

2. **Make `bootstrap` the primary command and treat `apply` as an alias** — Rename the conceptual workflow and have both names dispatch to the same implementation.
   - Pros: clearer user-facing vocabulary for first-run use; no duplicated behavior if dispatch is shared.
   - Cons: unnecessary compatibility/UX churn; existing `apply` specs and scripts become migration concerns; does not reduce orchestration complexity.
   - Effort: Low/Medium

3. **Keep only `apply` and document it as the bootstrap entrypoint** — Do not add a command; provide a wrapper or documentation convention around `dbootstrap apply`.
   - Pros: smallest code change and no new command contract.
   - Cons: leaves the requested high-level entrypoint implicit; a shell wrapper would risk becoming hidden orchestration; first-run UX remains dependent on knowing target flags.
   - Effort: Low

### Recommendation

Choose **Approach 1**: add a new `bootstrap` CLI command only as a thin
orchestration alias over the existing `apply` path. The implementation should
extract a single internal command/application function only if needed to make
both names share parsing, `buildPlan`, runner construction, reporting, and
exit handling. It must not introduce a second planner, provider registry,
privilege policy, or output model.

For the smallest safe contract, `bootstrap` should accept the same
`--profile`, repeatable `--resource`, `--catalog`, `--dry-run`, `--yes`, and
`--sudo` surface and require an explicit profile/resource target initially.
Do not silently select `dev` or mutate merely because the command is named
bootstrap. A future catalog-declared default can be proposed separately with
an explicit UX contract. Default and dry-run remain non-mutating; only
explicit `--yes` enables the existing provider set, and only `--yes --sudo`
enables sudo APT. No automatic privilege escalation, prompts, acquisition, or
hidden mutation belongs in this change.

Root help must list `bootstrap` with a concise description so the entrypoint is
discoverable from `dbootstrap --help`. Command-specific `bootstrap --help` must
then show its explicit-target and safety-mode guidance without starting the
pipeline.

The command should preserve current semantics:

- catalog/detection/planning failures: report diagnostics and exit `1` before execution;
- usage/flag errors: print command usage and exit `2`;
- default/dry-run noop or advisory outcomes: report and exit `0`;
- confirmed execution failures, including APT non-Linux/timeout and dotfiles
  failures: render the complete report, then exit `1`;
- confirmed partial success: retain all ordered results; do not claim rollback
  for APT and do not retry automatically.

Only syntactic argument failures are rejected before catalog loading and host
probing: no explicit target, malformed resource syntax, unexpected positionals,
or invalid safety-mode combinations. A syntactically valid but unknown profile
or resource is catalog-dependent. It must continue through the existing shared
catalog/detection/config/planning path and yield the same semantic diagnostic,
report, and exit behavior under both command names.

### UX, Non-Interactive, and Recovery Contract

The happy path should be readable in a pipe/CI log: selected target, detected
environment, ordered plan, mode, summary, per-step result, manual actions,
and stable exit status. stdout remains the report; stderr remains diagnostics
and usage/load errors. No interactive confirmation should be added: `--yes`
is the explicit confirmation boundary.

Recovery is rerun-oriented, not transactional. The Runner continues through
the plan after failures, so reports must identify every result. A subsequent
run re-detects state and reruns eligible work; users must resolve failed
prerequisites or dotfiles rollback state before retrying. Existing idempotence
is limited: tool/runtime command presence and dotfile module-directory presence
are detected, but package presence is not modeled, and a dotfile directory is
not proof that links are current. The entrypoint must not imply stronger
convergence than these detectors provide.

### Test Seams

Use the current package-level seams rather than real hosts or commands:

- stub environment, installation, config, and dotfiles detectors;
- inject brew/APT command-presence functions and command runners;
- inject installer factories and dotfiles base/prerequisite resolvers;
- assert command dispatch, report content, mutation boundaries, and exit code.
- assert root-help discoverability and command-specific bootstrap help without
  detector or execution calls.

The focused contract matrix should cover alias parity with `apply`, explicit
target validation, default/dry-run non-mutation, `--yes` and `--yes --sudo`
selection, no automatic sudo, planning short-circuit, mixed-provider order,
missing Homebrew advisory reporting, continued execution after one failure,
confirmed failure exit status, and pipe-friendly report output. Table-driven
tests with `t.TempDir()` and fake command seams match the repository's Go
testing guidance. No real Homebrew, APT, sudo, dotlink, filesystem home, or
network operation should be required.

### Risks

- A default profile hidden behind `bootstrap` could turn a convenience command into implicit scope selection or mutation; require explicit targets until a separate default-profile contract exists.
- Duplicating `runApply` would allow safety, provider eligibility, or exit semantics to drift; centralize dispatch before adding behavior.
- `sudo` can remain unsuitable for CI/non-interactive execution even though flag parsing is non-interactive; document that explicit confirmation does not provide credential handling.
- Continued execution means a profile can be partially applied; reporting and recovery guidance must not imply atomicity.
- Package-presence and dotfile-link state are intentionally incomplete; do not redesign detection or promise full idempotence in this change.
- Existing advisory Homebrew bootstrap guidance must remain non-mutating and must not become a download/install path.

### Ready for Proposal

Yes. The proposal should state that `bootstrap` is a thin alias over the
existing `apply` orchestration, preserve the current safety/provider boundaries,
require explicit targets for this slice, and explicitly exclude TUI work,
providers, catalog changes, package-presence redesign, automatic escalation,
download/bootstrap shell logic, retries, and hidden mutation.
