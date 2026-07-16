# dniebles-bootstrap

`dniebles-bootstrap` is the planning home for a personal development-environment bootstrapper. It will become a domain-first Go orchestrator that plans and runs profile installs and point installs while keeping dotfiles ownership external.

## Current status

This repository provides a domain-first Go CLI for deterministic planning and explicitly confirmed execution. Planning remains pure; host probing is performed by read-only adapters at the CLI composition root.

- Go application code includes planning, catalog decoding, execution reporting, provider-aware installers, and table-driven unit/integration tests.
- A repository-local TOML catalog fixture exists at `catalog/bootstrap.toml`; it decodes into planning inputs while planner-owned semantics remain in `internal/planning`.
- The accepted direction is captured under `openspec/changes/archive/2026-07-03-design-bootstrap-orchestrator/`, `openspec/changes/first-go-planning-slice/`, and `openspec/changes/catalog-toml-adapter/`.

## CLI usage

Run the CLI from the repository root. The default catalog is `catalog/bootstrap.toml`; use `--catalog <path>` for another local catalog.

## Operational workflow

### Quick path

1. Inspect selected work: `go run ./cmd/dbootstrap plan --profile dev`.
2. Review non-mutating reporting: `go run ./cmd/dbootstrap apply --profile dev` or add `--dry-run`.
3. Confirm eligible execution deliberately: `go run ./cmd/dbootstrap apply --profile dev --yes`.

Select targets with `--profile <name>`, repeatable `--resource <kind:name>`, and `--catalog <path>`. `bootstrap` accepts the same target and safety flags and uses the same execution workflow as `apply`.

### Commands and safety modes

| Command or flag | Behavior |
|---|---|
| `plan` | Inspects the selected work and renders planning statuses; it does not mutate the host. |
| `apply` | Reports execution results by default; only `--yes` confirms eligible execution. |
| `bootstrap` | Uses the same apply execution semantics for an explicit selection; provider/bootstrap needs remain advisory. |
| `--dry-run` | Reports the dry-run mode without mutation. It cannot be combined with `--yes`. |
| `--yes` | Explicitly confirms supported eligible execution. Default and dry-run modes do not mutate the host. |
| `--sudo` | Is meaningful only with confirmed `--yes` where the provider supports it; it does not independently enable mutation. |

### Explicit Homebrew acquisition (Linux/WSL)

If selected work needs Homebrew and `brew` is missing, acquisition is a separate, explicit prerequisite step. On Linux or WSL only, request it with both flags:

```bash
go run ./cmd/dbootstrap apply --profile dev --yes --acquire-homebrew
```

The command downloads one reviewed, commit-pinned installer to private local staging, verifies its pinned SHA-256, and executes only those verified local bytes. It never uses a remote `curl | bash`-style pipeline. After the installer exits, `dbootstrap` revalidates that `brew` is usable.

Acquisition is terminal: when revalidation succeeds, no target packages are installed in that run. Start a new terminal command to install the selected packages:

```bash
go run ./cmd/dbootstrap apply --profile dev --yes
```

`--yes` alone and `--acquire-homebrew` alone remain non-mutating guidance. The acquisition flow is unavailable on unsupported platforms and fails before downloading there. Download, verification, installer, or post-install Brew revalidation failures also stop safely; no target package installation continues. Fix the reported problem, then rerun deliberately.

### Confirmed reruns

A confirmed `apply --yes` or `bootstrap --yes` avoids installer mutation only when planning has marked an eligible `tool` or `runtime` as `already_installed` after reliable configured-command detection. The resource must have non-nil presence metadata with `Presence.Kind == "command_exists"` and a non-empty `Presence.Name`. The result is reported as `unchanged`: `already installed; no mutation attempted`.

This is intentionally narrow. In confirmed mode only, an eligible Brew package may also be unchanged after the exact read-only query `brew list --formula <InstallMetadata.Package>` positively proves its formula is installed. The result is `already installed; no mutation attempted`. A missing Brew executable, timeout, failed or ambiguous query is reported as failed and never authorizes installation. Default and dry-run modes do not perform this probe. APT packages, casks, versions, configuration, and dotfiles keep their normal behavior.

### Results and recovery

Reports keep the original plan order, including mixed results:

| Result | Meaning |
|---|---|
| `changed` | The eligible confirmed action completed. |
| `unchanged` | No action was needed or mutation was not attempted. |
| `not supported yet` | The selected action has no supported execution path in this mode. |
| `failed` | The action failed; confirmed eligible failures produce a non-zero result. |

Execution continues according to existing behavior after a non-terminal step failure. Fix the reported cause, then rerun deliberately. This workflow performs no automatic retry or rollback.

### Advisory bootstrap boundary

When a required provider or bootstrap dependency is missing, bootstrap output is manual/advisory guidance only. This workflow does not clone, fetch, install, retry, or otherwise acquire that dependency automatically.

## Goals and non-goals

| Goal | Decision |
|------|----------|
| Fresh-machine bootstrap | Provide a path to make `dbootstrap` available and then let the Go application own orchestration. |
| Profile installs | Plan and execute named environment profiles made of bundles, tools, runtimes, packages, and dotfiles requests. |
| Point installs | Install or reconcile one requested tool, runtime, package, bundle, or capability without pulling unrelated scope. |
| Domain-first core | Keep planning, dependency ordering, execution, and reporting in one shared core. |
| CLI now, TUI later | Start with a CLI interface and preserve a future Bubble Tea TUI as a thin interface over the same core. |

| Non-goal | Boundary |
|----------|----------|
| Dotfiles internals | `~/.dotfiles` owns modules, configs, assets, symlinks, validations, and `dotlink` semantics. |
| Shell orchestration | A shell wrapper may acquire `dbootstrap`, but it must not resolve catalogs, run installers, or own reporting. |
| Automatic convergence | Confirmed execution is explicit and limited; it does not provide package/version/configuration reconciliation or general idempotency. |

## Install flows

### Profile install

1. Detect environment facts: OS, distro, WSL status, and CPU architecture.
2. Resolve the requested profile from the repository catalog.
3. Expand bundles, tools, runtimes, packages, and dotfiles module requests.
4. Build a dependency-ordered plan.
5. Execute through installer and dotfiles provider adapters.
6. Report structured results: installed, already-installed, skipped, failed, or attention-required.

Missing expected dotfiles configuration should not block tool installation. It must remain visible as an attention-required result.

### Point install

1. Detect environment facts.
2. Resolve only the requested point target.
3. Avoid unrelated catalog resources.
4. Install or skip based on existing state.
5. Request only the dotfiles modules needed for that point target.
6. Report the result and any missing configuration attention item.

## First-run bootstrap entrypoint

The first-run entrypoint is intentionally small. Its job is to make `dbootstrap` available, then hand control to the Go application.

Supported entrypoint paths:

- Download and install a compatible released `dbootstrap` binary when available.
- Install or use Go to compile/run `dbootstrap` from this repository when a binary is unavailable or source execution is preferred.

After `dbootstrap` starts, the Go application owns catalog resolution, dotfiles integration, installer selection, dependency ordering, plan execution, and operational reporting.

## Direct binary installation (Linux/WSL)

`install.sh` installs a released `dbootstrap` binary and its bundled catalog without a package manager. It supports Linux and WSL on `amd64` and `arm64` only; other platforms fail with a clear error.

### Managed paths

| Asset | Default path | Override |
|---|---|---|
| Binary | `${XDG_BIN_HOME:-$HOME/.local/bin}/dbootstrap` | `XDG_BIN_HOME` |
| Catalog | `${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/catalog/bootstrap.toml` | `XDG_DATA_HOME` or `--catalog` |
| Install state | `${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/install-state.toml` | `XDG_DATA_HOME` |
| PATH state | `${XDG_DATA_HOME:-$HOME/.local/share}/dbootstrap/shell-path-state.toml` | `XDG_DATA_HOME` |

### Install a reviewed release

Download the immutable installer and its checksum for one explicit release. Verify the file, inspect it, then run that local file. Do not execute a remote response.

```bash
VERSION=v1.2.3
BASE_URL="https://github.com/dnieblesdev/dniebles-bootstrap/releases/download/${VERSION}"
INSTALLER="dbootstrap_install_${VERSION}.sh"

# Download both immutable assets to local files.
curl -fsSL -o "${INSTALLER}" "${BASE_URL}/${INSTALLER}"
curl -fsSL -o "${INSTALLER}.sha256" "${BASE_URL}/${INSTALLER}.sha256"

# Verify and inspect before local execution.
sha256sum --check --status --strict "${INSTALLER}.sha256"
sed -n '1,240p' "${INSTALLER}"

# Install once and opt in to Bash PATH setup in the same invocation.
bash "${INSTALLER}" --setup-path bash --shell-file "${HOME}/.bashrc"
```

### Opt in to PATH setup

Select one supported shell and exactly one explicit startup file on the initial installation invocation; a matching managed installation rejects a later setup-only invocation unless you use the explicit `--force` lifecycle override. The installer does not auto-detect a shell or startup file.

```bash
# Use this instead of the Bash invocation above for an initial Zsh installation.
bash "${INSTALLER}" --setup-path zsh --shell-file "${HOME}/.zshrc"
```

The selected file takes effect in a fresh interactive shell (`exec bash -i` or `exec zsh -i`). To update the current Bash session instead, run `source "${HOME}/.bashrc"`. The installer edits only the selected marked block and refuses ambiguous, modified, symlinked, or unmarked shell configuration.

### Reinstall, upgrade, or downgrade

A matching managed installation refuses to overwrite itself unless you explicitly confirm the lifecycle change:

```bash
./install.sh --force
```

`--force` also permits upgrades and downgrades. An unmanaged file at a managed path always aborts, even with `--force`.

### Uninstall

`--uninstall` removes only unchanged managed binary, catalog, and state files, plus an unchanged installer-owned PATH block when one exists. It refuses to delete modified files so your changes are preserved.

```bash
./install.sh --uninstall
```

### Catalog behavior after installation

Once installed, `dbootstrap` reads the installed catalog at the managed path by default. It does not depend on the repository checkout or the current working directory. Use `--catalog <path>` when you want to run against a different catalog, such as the one in a local repository clone.

### Scope and privilege boundaries

- No `sudo`, package manager, or elevated privilege is used.
- macOS, Windows, and other architectures are not supported by this path.
- The reviewed flow does not use `curl | bash`, moving/latest installer URLs, or automatic shell/profile detection.
- PATH setup supports only explicit Bash or Zsh startup-file targets; it never edits multiple files.
- The installer does not adopt or remove existing unmarked installations.

## Homebrew stable channel (Linux/WSL)

Use the primary repository's custom-URL tap on Linux/WSL `amd64` or `arm64`:

```bash
brew tap dnieblesdev/dniebles-bootstrap https://github.com/dnieblesdev/dniebles-bootstrap.git
brew install dnieblesdev/dniebles-bootstrap/dbootstrap
```

The formula is pinned to public stable `v0.1.0`, installs the catalog at `$(brew --prefix)/share/dbootstrap/catalog/bootstrap.toml`, and is intentionally unavailable on macOS. Pull requests validate the local formula first; the merged `main` formula is then smoke-tested on a GitHub-hosted Linux runner through the exact custom-URL tap. No standalone tap is required. See [Homebrew stable channel details](docs/homebrew-stable-channel.md) for pins, receipts, lifecycle, and rollback.

## Architecture direction

Future implementation should preserve these layers:

| Layer | Responsibility |
|-------|----------------|
| Domain/core | Profiles, catalog concepts, plans, plan steps, validation semantics, and structured statuses. |
| Application/use cases | Profile planning, point planning, execution orchestration, and result aggregation. |
| Infrastructure | TOML catalog adapter, installers, command runner, first-run acquisition, git sparse checkout, and dotfiles provider. |
| Interfaces | CLI first; future TUI as a thin presenter/controller over the same use cases. |

## Dotfiles boundary

`dniebles-bootstrap` integrates with dotfiles as an external provider. It may request modules, use partial clone and sparse checkout strategies, and invoke `dotlink` as a provider operation.

It must not own or duplicate dotfiles module internals, declarative profile semantics, symlink lifecycle, asset layout, validations, or configuration files.

## Catalog direction

The catalog belongs in this repository. TOML is the first implemented authoring format because it is readable and maps well to Go structs, but the domain model remains format-agnostic and schema-versioned so the format can evolve later.

The TOML adapter lives in `internal/catalog/toml`, and the initial fixture lives in `catalog/bootstrap.toml`. Adapter validation is intentionally shallow: TOML syntax, required fields, duplicate IDs, supported refs, and basic local references stay in the adapter; dependency expansion, environment filtering, missing config attention, and other planner semantics stay in `internal/planning`.

## Project guidance

- Use SDD/OpenSpec artifacts before implementation.
- Keep generated technical artifacts in English.
- Keep `.atl/` local and ignored.
- See `AGENT.md` for repository operating rules.
