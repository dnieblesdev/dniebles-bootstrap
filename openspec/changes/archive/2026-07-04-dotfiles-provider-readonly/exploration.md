# Exploration: dotfiles-provider-readonly

## Current State

The planning domain (`internal/planning/types.go`) already defines `ResourceKindDotfile` and `ResourceRef` supports dotfile references. The builder handles dotfile refs in dependency ordering identically to other resource kinds. However, dotfile resources are **not usable** from the catalog through the CLI because:

| Layer | Status | Detail |
|-------|--------|--------|
| TOML catalog schema | ❌ Missing | `catalogFile` has `[[tools]]`, `[[runtimes]]`, `[[packages]]` but **no** `[[dotfiles]]` section |
| TOML validation | ❌ Rejects | `supportedKind()` only accepts `tool`, `runtime`, `package` — `dotfile` refs fail with "unsupported resource kind" |
| TOML mapping | ❌ Skipped | `mapCatalog()` only maps tools/runtimes/packages into `Resources` |
| Catalog fixture | ❌ None | `catalog/bootstrap.toml` declares zero dotfile resources |
| State detector | ⚠️ Ignores | `isDetectableKind()` skips dotfiles; only tools and runtimes are PATH-detected |
| Config detector | ⚠️ Partial | Reads `~/.dotfiles/config/` for config-key presence, but has no module-awareness |
| CLI wiring | ❌ None | No dotfiles-specific detection before `BuildPlan` |

**What dotfiles data exists on the host today**:

```
~/.dotfiles/
├── .git/                    ← Git repo present
├── bash/                    ← dotfiles module (aliases, functions, config, prompt, completions, .bashrc)
├── zsh/                     ← dotfiles module (aliases, functions, config, completions, .zshrc, tools)
├── git/                     ← dotfiles module (.gitconfig)
├── config/                  ← dotfiles module (.config/starship.toml)
├── profiles/                ← dotfiles manifests (base.sh, interactive.sh)
├── dotlink/                 ← dotlink runtime (dotlink.sh, manifest.sh, tests/)
├── env/                     ← environment sourcing files (non-module)
├── wsl/                     ← WSL-specific files
├── linux/                   ← Linux-specific files
└── bin/                     ← bin/dotlink entrypoint
```

Known modules from `dotlink/manifest.sh`: `DOTLINK_KNOWN_MODULES=(bash git zsh config)`. Dotlink's `status` and `verify` subcommands are read-only drift detectors — they report linked/missing/conflicting state without modifying files.

## Affected Areas

- `internal/catalog/toml/schema.go` — must add `Dotfiles []resourceEntry` to `catalogFile`, `supportedKind()` must accept `dotfile`
- `internal/catalog/toml/catalog.go` — `mapCatalog` must include dotfiles in `Resources`
- `internal/catalog/toml/validate.go` — must collect and validate dotfile refs
- `internal/catalog/toml/catalog_test.go` — must exercise dotfile decoding/validation
- `internal/planning/types.go` — may need a new `DotfilesState` type (or reuse `InstallationState`)
- `internal/planning/builder.go` — may need to accept and consume dotfiles state
- `internal/planning/builder_test.go` — must cover dotfiles state integration
- `catalog/bootstrap.toml` — optional: could add a sample dotfile resource
- **NEW** `internal/dotfiles/detector.go` — read-only dotfiles awareness adapter
- **NEW** `internal/dotfiles/detector_test.go` — host-independent seam tests
- `cmd/dbootstrap/main.go` — CLI composition root must wire new detector
- `cmd/dbootstrap/main_test.go` — CLI tests must inject dotfiles stub
- `cmd/dbootstrap/render.go` — may need small rendering adjustments (render still works with existing field)

## Approaches

### 1. Extend state detector for dotfiles (minimal merge approach)

Add dotfile resource kind detection to the existing `internal/state/detector.go`. When `ref.Kind == ResourceKindDotfile`, check directory existence at `~/.dotfiles/<name>/` instead of PATH lookup. Add a `PathExists` seam to the `Detector` struct alongside the existing `LookPath` seam. Dotfiles presence flows through the same `InstallationState.PresentResources` map.

- **Pros**: No new package, no `BuildPlan` signature change, reuses existing plumbing, smallest diff
- **Cons**: Mixes PATH-based detection (Linux tooling) with filesystem detection (dotfiles structure) in one adapter; state detector loses single-responsibility focus; harder to evolve a proper `DotfilesProvider` interface later
- **Effort**: Low

### 2. Separate `internal/dotfiles/` adapter (clean separation approach)

Create `internal/dotfiles/detector.go` as a standalone read-only adapter following the exact pattern of `internal/environment`, `internal/state`, and `internal/config`. The detector has injectable seams (`BasePath`, `PathExists`, `ReadDir`), produces a `DotfilesState` that reports repo presence and module directory presence. The CLI composition root merges `DotfilesState` output into the existing `InstallationState` before calling `BuildPlan` (or the planner gets a new `DotfilesState` parameter). TOML catalog gains `[[dotfiles]]` support independently.

- **Pros**: Clean separation of concerns; each adapter has one job; follows established `internal/<domain>/detector.go` pattern with package-level `Detect()` and struct with injectable seams; easy to evolve into a full `DotfilesProvider` later; independently testable; mirrors directory structure of `internal/environment`, `internal/state`, `internal/config`
- **Cons**: Creates a new package; may require either a new `DotfilesState` type in planning or a merge helper in CLI composition root; slightly larger change
- **Effort**: Medium

### 3. Catalog-only: `[[dotfiles]]` support without a detector

Only extend the TOML catalog to accept `[[dotfiles]]` entries, route them through `mapCatalog` into `Resources`, and let the planner process them with existing `ConfigState`/`InstallationState` mechanics. No runtime dotfiles discovery — dotfile resources would always show as `planned` (or `attention_required` for missing config).

- **Pros**: Trivially small scope; makes dotfiles declarable in the catalog immediately; no new packages or detector logic
- **Cons**: No runtime awareness — can't distinguish present vs. missing dotfiles modules; `dbootstrap plan` output treats `dotfile:bash` and `dotfile:missing-module` identically; misses the "awareness layer" goal
- **Effort**: Very Low

## Recommendation

**Approach 2 (separate `internal/dotfiles/` adapter)** is the right choice for this slice. It delivers real awareness while keeping the separation clean:

1. **TOML catalog gains `[[dotfiles]]` support** — `catalogFile` gets a `Dotfiles []resourceEntry` field, `supportedKind()` accepts `dotfile`, `collectResourceRefs` and `validateDependencyRefs` are called for dotfiles, `mapCatalog` maps them into `Resources`. In practice this means ~3 files changed in `internal/catalog/toml/` plus test updates.

2. **`internal/dotfiles/detector.go`** — a `Detector` struct with:
   - `BasePath string` (defaults to `$HOME/.dotfiles`)
   - `PathExists func(string) bool` injectable seam
   - `ReadDir func(string) ([]os.DirEntry, error)` injectable seam
   - Returns `DotfilesFacts` containing `RepoPresent bool` and `ModulePresent map[string]bool`

3. **Planner integration** — two clean options:
   - Option A: Add `DotfilesFacts` parameter to `BuildPlan`. Clean but widens the signature.
   - Option B: Merge dotfiles module presence into `InstallationState.PresentResources` at the CLI composition root (like `tool:git: true`). The planner already knows how to handle `already_installed` for present resources. This requires NO planner signature change — the CLI does a one-liner merge after detecting both states.

   **Option B is recommended** because it preserves the existing planner contract, makes dotfiles presence naturally show as `already_installed` in plan output, and keeps the change surface smaller. The merge logic at the CLI is: for each dotfile module that exists, add `ResourceRef{Kind: ResourceKindDotfile, Name: module}` → `true` to the installation state map.

4. **CLI wiring** — follows existing pattern: `detectDotfilesState` is a package-level variable in `main.go` (like `detectConfigState`), initialized to `dotfiles.Detect`, overridable in tests.

5. **Test patterns** — follow existing host-independent seam injection: table-driven detector tests with fake `PathExists`/`ReadDir`, CLI test stubs, and planner tests that cover dotfile resources flowing through `InstallationState`.

## Risks

- **DotfilesState vs InstallationState merge semantics**: Marking a dotfiles module as `already_installed` when its directory exists could be misleading — the module directory existing doesn't mean symlinks are set up. But for the READ-ONLY bootstrapper, it's sufficient: the plan can tell you "this module is available" vs "this module directory is missing." Actual dotlink operations are out of scope.
- **Module name convention**: The dotfiles repo uses directory names like `bash`, `zsh`, `git`, `config` as module identifiers. The catalog `dotfile` resource name should match these directory names. A `dotfile:shell` resource would check for `~/.dotfiles/shell/` — which doesn't exist today. This is correct behavior: the plan would report `attention_required` or `planned` for it.
- **Dotfiles repo gating**: If `~/.dotfiles/` doesn't exist at all, all dotfile modules report absent. This is the right behavior for a fresh-machine bootstrap scenario — the plan should flag all dotfiles as missing until the repo is cloned.
- **Boundary drift risk**: The `internal/dotfiles/` adapter must remain read-only and must not grow dotlink invocation, sparse checkout, or symlink management in this slice. Those belong in a future `dotfiles-provider` or `apply` slice.

## Ready for Proposal

Yes — the next step should define the proposal covering:
- Extending TOML catalog with `[[dotfiles]]`
- Creating `internal/dotfiles/` as a read-only adapter
- Dotfiles state flowing into existing `InstallationState` (no planner signature change)
- CLI composition root wiring with injectable test seams
- A sample dotfile resource in the catalog fixture (e.g., `dotfile:bash` depending on `tool:git`)
