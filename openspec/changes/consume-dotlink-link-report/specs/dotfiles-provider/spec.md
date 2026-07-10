# Delta for dotfiles-provider

## ADDED Requirements

### Requirement: Dotlink JSON reports are the only execution source of truth

For confirmed dotfile execution, the system MUST request `dotlink link --report=json MODULE...` and treat stdout as the only candidate structured outcome source. Stderr MUST NOT be parsed as report input or as a fallback.

When stdout is present, the system MUST parse and validate it regardless of `CommandStatus`. Parsing MUST first reject duplicate JSON object keys at every object level—including top-level report, entry, cause, failure, and rollback objects—before decoding the domain/wire report. `json.Decoder.DisallowUnknownFields` alone is insufficient. The parser MUST also reject malformed JSON, unknown fields, trailing documents, unsupported schema versions, missing/unknown modules, invalid outcomes, incomplete reports, and semantic contradictions. No invalid-report path may parse human-readable stdout or stderr as a compatibility fallback.

A valid JSON report with `status: failed` and non-success command status MUST preserve its validated entries, causes, and rollback details and produce a failed Bootstrap result. When command status is non-success and stdout is absent, malformed, or invalid, the system MUST produce a generic safe failed result. A report with `status: success` and non-success command status, or `status: failed` and success command status, MUST be treated as inconsistent and fail safely.

#### Scenario: Duplicate keys fail before domain decode

- GIVEN stdout contains a duplicate top-level JSON key or a duplicate key in an entry, cause, failure, or rollback object
- WHEN the report is consumed
- THEN the report is rejected as invalid
- AND execution produces a safe failed result
- AND no human-output parsing fallback is used

#### Scenario: Valid failed report survives a non-zero process exit

- GIVEN confirmed Dotlink execution exits non-zero
- AND stdout contains a valid schema v1 report with `status: failed`
- WHEN the provider consumes the command result
- THEN it preserves validated entries, safe causes, and rollback detail
- AND the aggregate Bootstrap result is failed

#### Scenario: Missing or malformed report after command failure fails safely

- GIVEN confirmed Dotlink execution exits non-zero
- AND stdout is absent, malformed, or fails validation
- WHEN the provider consumes the command result
- THEN it returns a generic safe failed result
- AND no human-output parsing fallback is used

#### Scenario: Command and report success states must agree

- GIVEN stdout contains a valid report
- AND the report status and command status disagree
- WHEN the provider consumes the result
- THEN it treats the result as inconsistent and failed safely

### Requirement: Dotlink report outcomes are rendered per entry

The system MUST render each validated per-link detail from the execution result, rather than infer every upstream outcome from a module `StepStatus`.
A successful aggregate with all unchanged entries maps to a skipped module summary; a successful aggregate with one or more changed entries and no failures maps to an installed module summary. Any failed or rolled_back entry, or an aggregate failed report, maps to a failed module summary and a non-zero confirmed apply exit.

Rendering MUST show each entry’s `changed`, `unchanged`, `failed`, or `rolled_back` outcome with source, target, safe cause, and rollback detail when supplied.

#### Scenario: Mixed changed and unchanged entries are reported truthfully

- GIVEN a valid successful JSON report includes one `changed` entry and one `unchanged` entry
- WHEN the report is rendered
- THEN the module summary is installed
- AND the output shows changed and unchanged outcomes for the respective links

#### Scenario: Failed and rolled_back entries preserve diagnostics

- GIVEN a valid failed JSON report includes `failed` or `rolled_back` entries
- WHEN the report is rendered
- THEN the module summary is failed
- AND the output shows each failed or rolled_back entry with available safe cause or rollback detail

## MODIFIED Requirements

### Requirement: Local dotfiles execution requires explicit safe prerequisites

The local dotfiles execution core MUST resolve the base path from `DBOOTSTRAP_DOTFILES_DIR` when set and non-empty, MUST fail safely with no fallback when it is set empty, and MUST resolve `~/.dotfiles` only when the env var is unset.
It MUST canonicalize symlinks with `EvalSymlinks` before validation and MUST fail safely when the canonical base is missing, unresolved, unsafe, not an existing directory, `/`, or the user's home directory itself.

On resolution failure, it MUST report the base-resolution source, attempted source/candidate, selected modules, and safe cause. It MUST NOT label the attempted candidate as `canonical base`. It MAY render `canonical base` only after successful canonicalization and safety validation.

It MUST fail safely when `bin/dotlink` is missing or resolves outside the validated canonical repository, validate module names before command construction, reject missing/out-of-repository selected modules, and MUST NOT attempt remote acquisition.

#### Scenario: Base resolution failure does not claim a false canonical base

- GIVEN base resolution fails for an env or home-convention candidate
- WHEN the failure is rendered
- THEN the attempted source/candidate, selected modules, and safe cause remain visible
- AND the unresolved candidate is not labeled as canonical base

#### Scenario: Safe prerequisites allow provider execution

- GIVEN a selected dotfile module name and a validated canonical local base
- AND `bin/dotlink` and the selected module exist inside that repository
- WHEN local execution runs through the provider
- THEN dotlink may be requested through the injected command runner
