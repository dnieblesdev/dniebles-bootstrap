# Delta for execution-contracts

## MODIFIED Requirements

### Requirement: Execution contracts remain non-mutating for apply

`internal/execution` MUST use noop contracts by default and in `--dry-run`, and MUST allow real execution only for confirmed brew-backed tool/package steps (with their existing cross-platform eligibility), confirmed Linux APT-backed tool/package steps, and confirmed selected dotfile steps. APT MUST be composed only for confirmed `--yes`, using direct `apt-get install -y -- <package>` or the explicit `sudo apt-get install -y -- <package>` vector for `--yes --sudo`, with a ten-minute `CommandRequest.Timeout`; no kind-based Runner redesign or provider-registry redesign is permitted.
(Previously: confirmed real execution was limited to brew-backed steps and selected dotfiles.)

#### Scenario: Default and dry-run remain noops

- GIVEN `apply` runs without `--yes` or with `--dry-run`
- WHEN execution is dispatched
- THEN APT is neither probed nor invoked and results remain non-mutating

#### Scenario: Confirmed Linux APT is provider-gated

- GIVEN `apply --yes` runs on Linux with an APT-backed `tool` or `package`
- WHEN the Runner dispatches by kind
- THEN only the APT provider gate may delegate to one of the two explicit APT vectors with `-y --` and the ten-minute timeout

#### Scenario: Other provider execution remains unchanged

- GIVEN an APT step is non-Linux, unsupported, or not selected for confirmed execution
- WHEN the Runner processes it
- THEN non-Linux selected APT returns `StepStatusFailed` without APT/sudo probes or commands; other unsupported work retains its existing outcome without shell, automatic escalation, retry, fallback, bootstrap, update, repository, presence, or rollback behavior

#### Scenario: Bootstrap data stays advisory

- GIVEN Homebrew bootstrap need data is attached
- WHEN execution contracts report it
- THEN the data remains non-mutating and reviewable

#### Scenario: Core provider is dormant until composed

- GIVEN the dotfiles provider and installer exist in `internal/execution`
- WHEN no caller composes them into the confirmed apply runner
- THEN no dotlink execution is possible through the CLI
- AND existing noop execution behavior is unchanged

### Requirement: CLI composition uses injectable execution seams

The CLI apply composition root MUST allow tests to inject Linux facts, `apt-get` availability, `sudo` availability, and the `CommandRunner` used to prove APT. Production composition MUST not probe or expose APT through default or dry-run modes.
(Previously: injectable seams covered brew and dotfiles execution only.)

#### Scenario: Opt-in fixture proves the command vector

- GIVEN a custom catalog or fixture supplies an explicit APT target and fake seams
- WHEN confirmed Linux execution is tested
- THEN the test can assert the exact direct vector without a real external command

#### Scenario: Production composition remains confirmed-only

- GIVEN production apply dependencies are used
- WHEN the user runs default apply or `apply --dry-run`
- THEN APT is not composed with a mutating runner or probed

#### Scenario: Test proof covers both privilege modes

- GIVEN a custom APT target and fake Linux, executable, and runner seams
- WHEN confirmed tests run with and without `--sudo`
- THEN they assert both vectors and all failure outcomes without a real external command
