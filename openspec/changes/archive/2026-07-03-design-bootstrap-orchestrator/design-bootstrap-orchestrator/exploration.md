# Exploration: design-bootstrap-orchestrator

### Current State
`dniebles-bootstrap` is an almost-empty repo: a README, `.gitignore`, and local `.atl` state. There is no Go module, no `openspec/` tree yet, and no implementation to refactor. The only existing architecture signal is the hard boundary that dotfiles module ownership remains in `~/.dotfiles`.

#### Domain language and bounded contexts
- **Bootstrap Orchestration**: owns planning, execution, retries, verification, and user-facing workflows.
- **Environment Detection**: owns OS, distro, WSL, and architecture detection.
- **Catalog & Planning**: owns declarative definitions for tools, runtimes, packages, bundles, profiles, and future resources.
- **Dotfiles Integration**: owns provider selection, module lookup, sparse checkout requests, partial clone behavior, and dotlink invocation.
- **Execution Infrastructure**: owns installers, runners, logging, command execution, and transport concerns.

#### Primary domain entities
- **Profile**: declarative install target composed of tools, runtimes, packages, bundles, and dotfiles module requests.
- **Tool**: installable capability with optional or mandatory config.
- **Runtime**: execution platform dependency such as Go, Node, Python, or Java.
- **Package**: OS-level package or formula.
- **Bundle**: reusable group of tools/resources reused across profiles.
- **Catalog**: authoritative declarative registry of installable resources and relationships.
- **Plan**: ordered, validated set of actions to reach a target state.
- **Runner**: executes plan steps and captures results.
- **Installer**: strategy for a specific resource type or platform.
- **DotfilesProvider**: external bridge to `~/.dotfiles` operations.
- **DotfilesModule**: requested module artifact known to dotfiles, not bootstrap.
- **EnvironmentDetector**: collects platform facts for resolution and plan selection.

#### Key relationships
- Profiles reference bundles and resources; bundles reference tools/runtimes/packages; resources may reference dotfiles modules.
- Plans are produced from catalog + environment facts + profile/point target.
- Installers operate on resources; runners coordinate installers; dotfiles provider is invoked only after module validation and sparse checkout selection.
- Bootstrap must not own module internals, symlink lifecycle, asset structure, or validation semantics for dotfiles.

#### Lifecycle flows
- **Profile install**: detect environment → resolve profile → expand bundles/resources → verify existing state → install tools/runtimes/packages → update/clone dotfiles → sparse-checkout needed modules → run dotlink → verify result.
- **Point install**: resolve catalog entry → verify already installed → install if needed → inspect associated dotfiles modules → validate modules exist in dotfiles → sparse-checkout those modules → run dotlink only for requested scope.

#### Catalog format analysis
- **TOML**: strong fit for Go, readable for humans, supports comments, stable for small-to-medium declarative registries, and maps well to typed structs.
- **YAML**: flexible and familiar, but easy to misindent, ambiguous typing is common, and schema drift becomes subtle in large catalogs.
- **JSON**: precise and machine-friendly, but poor for hand-edited catalogs because it lacks comments and is noisy for nested profiles.
- **Recommendation**: TOML for the first catalog version; it is the best balance of human-editability, Go ergonomics, and declarative structure. If future resource metadata becomes highly nested or schema-heavy, keep the domain model format-agnostic so a later migration is possible.

#### Boundary with `~/.dotfiles`
- Bootstrap owns orchestration, not dotfiles internals.
- Dotfiles remains responsible for modules, profiles, configurations, symlinks, assets, validations, and `dotlink`.
- Bootstrap may request specific modules, perform sparse checkout, and invoke `dotlink`, but it must not model or mutate module internals.
- Assets may live in dotfiles later; bootstrap should treat them as provider-managed.

#### Docs strategy now
- **AGENT.md**: add a short repository operating guide covering scope boundary, SDD workflow, dotfiles separation, `.atl/` ignore rule, and no-code-before-design expectations.
- **README.md**: rewrite as an orientation doc with purpose, non-goals, boundary with dotfiles, supported flows, and a “what lives where” summary.
- Keep both docs concise and review-friendly; they should help a future contributor avoid architectural drift, not teach implementation details.

### Affected Areas
- `README.md` — needs to become the first concise statement of bootstrap scope and boundaries.
- `AGENT.md` — should be added as the repo operating guide for future agents.
- `openspec/changes/design-bootstrap-orchestrator/exploration.md` — this exploration artifact.
- `openspec/config.yaml` and `openspec/specs/` — should be created next if SDD continues past exploration.

### Approaches
1. **Domain-first orchestrator core** — model bootstrap as an application core that plans and executes environment work, with CLI and TUI as thin interfaces.
   - Pros: clean separation, avoids duplicating logic, supports future resources beyond tools, keeps dotfiles concerns external.
   - Cons: upfront design work, needs disciplined boundaries.
   - Effort: High

2. **Command-driven installer** — shape the system around direct commands like `install` and `apply`, with lighter domain modeling.
   - Pros: faster to start, simpler mental model.
   - Cons: risks hardcoding workflows, makes profile planning and future resource types harder, encourages logic leakage into interfaces.
   - Effort: Medium

### Recommendation
Use the domain-first orchestrator core. The user’s requirements are explicitly about planning, dependency resolution, profile composition, point installs, and multiple interfaces over one core. That is a domain orchestration problem, not a script wrapper. TOML is the best initial catalog format because it is easy to read, comments well, and maps cleanly to Go structs.

### Risks
- Catalog format choice can lock in painful schema evolution if picked too narrowly.
- Dotfiles boundary drift: bootstrap may accidentally absorb module knowledge instead of treating dotfiles as an external provider.
- TUI/CLI divergence if execution logic is not centralized early.

### Ready for Proposal
Yes — the next step should define the proposal, especially domain language, bounded contexts, catalog format decision, and repo docs strategy.
