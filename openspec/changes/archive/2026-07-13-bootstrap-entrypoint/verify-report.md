```yaml
schema: gentle-ai.verify-result/v1
evidence_revision: sha256:34022680efd00b1bafa9369138b45e62689b371160629279443f321e3b17383d
verdict: pass
blockers: 0
critical_findings: 0
requirements: 7/7
scenarios: 17/17
test_command: go test ./cmd/dbootstrap -run 'TestRunBootstrap|TestRunApplyLike|TestRunApplyHelpRetainsParserUsageFailure' -count=1; go test ./... -count=1
test_exit_code: 0
test_output_hash: sha256:34022680efd00b1bafa9369138b45e62689b371160629279443f321e3b17383d
build_command: go build ./...; go vet ./...; test -z "$(gofmt -l $(git ls-files '*.go'))"; git diff --check
build_exit_code: 0
build_output_hash: sha256:34022680efd00b1bafa9369138b45e62689b371160629279443f321e3b17383d
```

# Verification Report: Bootstrap CLI Entrypoint

**Change:** `bootstrap-entrypoint`  
**Evidence file:** `/tmp/opencode/bootstrap-entrypoint-verify-evidence.txt`  
**Evidence SHA-256:** `34022680efd00b1bafa9369138b45e62689b371160629279443f321e3b17383d`

## Completeness

| Metric | Result |
|---|---:|
| Tasks | 9/9 complete |
| Requirements | 6/6 compliant |
| Scenarios | 16/16 covered by passing runtime tests |

## Runtime Evidence

| Check | Result |
|---|---|
| Focused bootstrap CLI tests | PASS |
| Full suite: `go test ./... -count=1` | PASS |
| Build: `go build ./...` | PASS |
| Vet: `go vet ./...` | PASS |
| Formatting check | PASS |
| Diff check: `git diff --check` | PASS |

The evidence file records the exact command output. No coverage threshold is configured.

## Specification Compliance

| Requirement / scenario group | Passing runtime coverage |
|---|---|
| Explicit profile/resource selection and valid invocation parity | `TestRunBootstrapMatchesApplyAcrossSafetyModes` |
| Missing target, malformed resource, positionals, and invalid safety combinations fail before probing | `TestRunApplyLikeRejectsSyntacticInputBeforeProbing` |
| Unknown profile and unknown resource follow the semantic path | `TestRunBootstrapMatchesApplyForUnknownProfile`, `TestRunBootstrapMatchesApplyForUnknownResource` |
| Default, dry-run, confirmed, and confirmed-sudo safety parity | `TestRunBootstrapMatchesApplyAcrossSafetyModes` |
| Catalog, config, and environment prerequisite failures are safe and equivalent | `TestRunBootstrapMatchesApplyForPrerequisites` |
| Confirmed partial failure preserves result order and failed exit parity | `TestRunBootstrapMatchesApplyForPartialFailure` |
| Root help discoverability and bootstrap command help are non-probing | `TestRunBootstrapHelp` |
| Existing apply help behavior remains compatible | `TestRunApplyHelpRetainsParserUsageFailure` |

## Design Coherence

`apply` delegates to `runApplyLike("apply", ...)` and `bootstrap` calls `runApplyLike("bootstrap", ...)`. The command name controls dispatch and bootstrap-only standalone help; planning, provider selection, rendering, and exit mapping remain shared. `cmd/dbootstrap/render.go` is unchanged.

## Issues

- **CRITICAL:** None.
- **WARNING:** None.
- **SUGGESTION:** Add a coverage baseline separately if the project adopts a threshold.

## Verdict

**PASS** — all 9 tasks are complete; the exact spec totals are 6 requirements and 16 scenarios; every required scenario group has fresh passing runtime evidence.
