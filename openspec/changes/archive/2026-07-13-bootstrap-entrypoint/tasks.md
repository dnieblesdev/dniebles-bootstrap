# Tasks: Bootstrap CLI Entrypoint

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | 650–800 (implementation plus parity tests) |
| 800-line budget risk | High; accepted `size:exception` |
| Chained PRs recommended | No |
| Suggested split | One direct commit/push on `main` |
| Delivery strategy | exception-ok |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|---|---|---|---|
| 1 | Shared command dispatch and orchestration | Direct commit | `main`; include focused tests; no renderer changes |
| 2 | Full parity and regression matrix | Same direct commit | Depends on Unit 1; verify once, without redundant reruns |

## Phase 1: Shared CLI Boundary

- [x] 1.1 Modify `cmd/dbootstrap/main.go` to dispatch `bootstrap`, list it in root help, and provide name-aware command usage/help.
- [x] 1.2 Extract the current apply flow into one command-name-aware runner; retain `apply` flags, defaults, output, and exit behavior.
- [x] 1.3 Keep `cmd/dbootstrap/render.go` unchanged; ensure command name is presentation-only and cannot alter plans, providers, reports, or exit mapping.

## Phase 2: Validation and Orchestration

- [x] 2.1 Validate missing targets, malformed resources, positionals, `--dry-run --yes`, and `--sudo` without `--yes` before catalog, detector, runner, or provider work.
- [x] 2.2 Route syntactically valid unknown profiles/resources through the shared catalog, host detection, configuration, planner, diagnostic, report, and failure path.
- [x] 2.3 Preserve default/dry-run non-mutation, `--yes` eligibility, `--yes --sudo` APT behavior, partial execution reporting, and apply compatibility.

## Phase 3: Parity Verification

- [x] 3.1 Extend `cmd/dbootstrap/main_test.go` with root-help, `bootstrap --help`, and no-probe assertions using injected seams.
- [x] 3.2 Add table-driven apply/bootstrap comparisons for default, dry-run, yes, and yes+sudo: plans, reports, exit statuses, and command requests.
- [x] 3.3 Add parity cases for syntactic failures, unknown-target semantic failures, catalog/config/environment prerequisites, and ordered partial failures; run the focused suite once.

## Completion Evidence

- `apply -h` and `apply --help` retain their parser-driven usage failure, stderr
  output, and exit status; only `bootstrap` intercepts its aliases for successful
  command help.
- The parity matrix covers unknown profiles and syntactically valid unknown
  resources, catalog loading, missing configuration, environment mismatch, and
  ordered confirmed partial failure. The partial-failure case explicitly asserts
  `package:first [changed]` before `package:second [failed]` for both `apply`
  and `bootstrap`. Each comparison asserts output, exit status, environment
  probes, and command-runner calls where applicable.
- Final evidence: `gofmt`, `go test ./cmd/dbootstrap`, `go test ./...`,
  `go vet ./...`, and `git diff --check` passed. `render.go` remains unchanged.
