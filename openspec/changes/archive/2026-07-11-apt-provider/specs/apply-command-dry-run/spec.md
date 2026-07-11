# Delta for apply-command-dry-run

## MODIFIED Requirements

### Requirement: Apply mode is explicit and safe by default

The `apply` command MUST treat default and `--dry-run` modes as non-mutating and non-probing, and MUST treat `--yes` as the only confirmed mode that may execute eligible brew-backed or APT-backed tool/package steps and selected dotfile steps. Existing Homebrew execution eligibility remains cross-platform. On Linux, APT MUST use `apt-get install -y -- <package>` for `--yes` and `sudo apt-get install -y -- <package>` only for explicit `--yes --sudo`, with a ten-minute request timeout. `--sudo` MUST be rejected unless `--yes` is set and MUST NOT change non-APT providers.
(Previously: only confirmed brew-backed and selected dotfile execution could mutate.)

#### Scenario: Default and dry-run do not probe APT

- GIVEN the user runs `apply` or `apply --dry-run`
- WHEN the command composes and executes its plan
- THEN it neither probes `apt-get` nor invokes an APT command

#### Scenario: Linux confirmed mode permits APT

- GIVEN Linux `apply --yes` selects an eligible APT target
- WHEN confirmed execution is composed
- THEN APT availability may be checked and direct installation may be attempted

#### Scenario: Linux confirmed sudo mode is explicit

- GIVEN Linux `apply --yes --sudo` selects an eligible APT target
- WHEN confirmed execution is composed
- THEN `sudo` and `apt-get` availability may be checked and only `sudo apt-get install -y -- <package>` may be attempted

#### Scenario: Sudo flag is rejected outside confirmed mode

- GIVEN `--sudo` is combined with default mode or `--dry-run`
- WHEN flags are validated
- THEN apply returns a usage error without probing or executing APT

#### Scenario: Non-Linux confirmed mode rejects APT safely

- GIVEN `apply --yes` selects an APT target on a non-Linux host
- WHEN execution is composed
- THEN the result is `StepStatusFailed`, confirmed apply exits non-zero, and zero apt/sudo probes or commands run

## ADDED Requirements

### Requirement: APT apply failures are reported without orchestration

Confirmed APT failures, including missing `apt-get`, missing `sudo` for `--sudo`, command failure, and ten-minute timeout, MUST be rendered as structured failed results and MUST cause the confirmed command to report non-success. The command MUST NOT silently escalate, retry, fall back, bootstrap, update, change repositories, detect package presence, or claim rollback.

#### Scenario: Missing or failed APT execution is visible

- GIVEN confirmed Linux execution lacks `apt-get`, lacks `sudo` for `--sudo`, or the selected APT vector fails
- WHEN the report is rendered
- THEN the APT step is failed with its non-success outcome and no second command is attempted
