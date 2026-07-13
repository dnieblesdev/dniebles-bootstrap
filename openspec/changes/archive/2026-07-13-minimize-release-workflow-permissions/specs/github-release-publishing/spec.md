# Delta for GitHub Release Publishing

## MODIFIED Requirements

### Requirement: Validate the release tag

The publish workflow MUST require a `v`-prefixed strict SemVer input and use it
as both build version and release tag, rejecting invalid input before either
downstream build or publication job can run. (Previously: invalid input was
required to fail before building or publishing.)

#### Scenario: Valid stable version

- GIVEN a maintainer dispatches with `v1.2.3`
- WHEN validation succeeds
- THEN the called build and release tag both use `v1.2.3`

#### Scenario: Invalid or unprefixed version

- GIVEN a maintainer dispatches with `1.2.3`, `v1`, or an invalid SemVer
- WHEN validation runs
- THEN the workflow fails before the build or publish jobs start

### Requirement: Restrict publication authority

The workflow-level permissions mapping MUST contain only `contents: read`.
Only the publication job MAY have `contents: write` and `actions: read`; the
called build MUST retain read-only access, and manual builds MUST remain
artifact-only. No global `actions: write` permission MAY be granted.
(Previously: only the publish job was required to have `contents: write`.)

#### Scenario: Permissions are inspected

- GIVEN the release-publish workflow is evaluated
- WHEN workflow-level and effective job permissions are inspected
- THEN the global mapping contains only `contents: read`
- AND only publish has `contents: write` and `actions: read`

#### Scenario: Permission removal preserves behavior

- GIVEN a valid release request reaches the publish job
- WHEN the called build and artifact download execute
- THEN release validation, artifact publication, and checksum verification behave unchanged

## ADDED Requirements

### Requirement: Preserve non-permission behavior

Removing the unused global write grant MUST NOT change release triggers, version
validation, artifact identity, checksum verification, or publication outcomes.

#### Scenario: Workflow behavior remains unchanged

- GIVEN the workflow is run with a valid or invalid version
- WHEN the workflow completes or fails validation
- THEN its observable behavior matches the pre-change workflow except for the removed permission
