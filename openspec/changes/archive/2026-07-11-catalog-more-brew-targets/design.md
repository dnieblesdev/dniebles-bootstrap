# Design: Catalog More Brew Targets

## Technical Approach

Extend the declarative default catalog with `package:jq`, using the existing package schema and the same `brew` install and `command_exists` presence metadata pattern as `package:ripgrep`. Append the resource to `bundle:cli`; the existing `profile:dev` already selects that bundle. Update the focused default-catalog fixture test to make the new decoded resource, bundle membership, and `dev` planning result contractual. No production Go behavior changes are required.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Represent jq | Add one `[[packages]]` record with `provider = "brew"`, `package = "jq"`, and `command_exists`/`jq` | New resource kind; new provider metadata | The TOML adapter already maps this metadata into inert `planning.Resource` fields and the Homebrew installer already supports package resources. |
| Select jq for development | Append `package:jq` to existing `bundle:cli` | New bundle/profile; direct `dev` profile resource | `dev` already selects `cli`; preserving that composition proves selection without changing profile semantics. |
| Protect behavior | Extend `TestLoadFileAndBuildPlanFromFixture` | Execution/provider tests; a separate test harness | The focused fixture test loads the real catalog, asserts decoded metadata, and builds a plan without executing commands. It directly covers the delta contract at the smallest boundary. |

## Data Flow

```text
catalog/bootstrap.toml
  package:jq + bundle:cli membership
              |
              v
catalog/toml LoadFile -> planning.Catalog -> BuildPlan(profile:dev)
              |                              |
              v                              v
       metadata retained              plan includes package:jq
```

`InstallMetadata` and `PresenceMetadata` remain desired-state data during planning; this change neither invokes providers nor performs presence checks.

## File Changes

| File | Action | Description |
|---|---|---|
| `catalog/bootstrap.toml` | Modify | Add the jq package entry and append `package:jq` to `bundle:cli`. |
| `internal/catalog/toml/catalog_test.go` | Modify | Add a jq resource ref and extend the real-catalog expected model, planned step order, status, and metadata assertions. |

## Interfaces / Contracts

No new interfaces or types. The catalog contract adds this existing metadata shape:

```toml
[[packages]]
id = "jq"
install = { provider = "brew", package = "jq" }
presence = { kind = "command_exists", name = "jq" }
```

`bundle:cli` will contain `tool:git`, `package:ripgrep`, and `package:jq`. A `profile:dev` plan will retain existing resources and include jq through that bundle.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit/contract | Real catalog decoding and preserved jq metadata | Extend `TestLoadFileAndBuildPlanFromFixture` expected `planning.Catalog`. |
| Integration | Existing `dev` selection includes jq in deterministic plan order | Extend the same test's `BuildPlan` expected refs and planned status assertions; no command runner is wired. |
| E2E | Not applicable | Do not execute Homebrew or apply; behavior is explicitly out of scope. |

Run `go test ./internal/catalog/toml` first, then `go test ./...`; retain strict TDD during implementation.

## Migration / Rollout

No migration required. The catalog-only addition makes jq part of future `dev` plans. Rollback removes its catalog and bundle entries together with the matching contract assertions.

## Open Questions

None.
