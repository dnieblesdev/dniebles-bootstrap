# GitHub Release Publishing Specification

## Purpose

Define GitHub Release publication from verified build artifacts.

## Requirements

### Requirement: Validate the release tag

The publish workflow MUST require a `v`-prefixed strict SemVer input and use it
as both build version and release tag, rejecting invalid input before either
downstream build or publication job can run.

#### Scenario: Valid stable version

- GIVEN a maintainer dispatches with `v1.2.3`
- WHEN validation succeeds
- THEN the called build and release tag both use `v1.2.3`

#### Scenario: Invalid or unprefixed version

- GIVEN a maintainer dispatches with `1.2.3`, `v1`, or an invalid SemVer
- WHEN validation runs
- THEN the workflow fails before building or publishing

### Requirement: Publish the called build outputs

The workflow MUST publish exactly three archives and matching SHA-256 files
from its called build, verifying checksums before publication and never altering files.

#### Scenario: Verified assets are published

- GIVEN the called build succeeds and all checksums match
- WHEN publication completes
- THEN the release contains exactly those three archives and three checksums

#### Scenario: Verification fails

- GIVEN an archive or checksum does not match
- WHEN verification runs
- THEN publication fails and no release is created or uploaded

### Requirement: Restrict publication authority

The workflow-level permissions mapping MUST contain only `contents: read`.
Only the publication job MAY have `contents: write` and `actions: read`; the
called build MUST retain read-only access, and manual builds MUST remain
artifact-only. No global `actions: write` permission MAY be granted.

#### Scenario: Permissions are inspected

- GIVEN either workflow is evaluated
- WHEN its effective job permissions are inspected
- THEN only the publish job has `contents: write`

### Requirement: Preserve non-permission behavior

Removing the unused global write grant MUST NOT change release triggers, version
validation, artifact identity, checksum verification, or publication outcomes.

#### Scenario: Workflow behavior remains unchanged

- GIVEN the workflow is run with a valid or invalid version
- WHEN the workflow completes or fails validation
- THEN its observable behavior matches the pre-change workflow except for the removed permission

### Requirement: Prevent overwrites and capture evidence

The workflow MUST fail when the requested tag or release exists and MUST NOT
overwrite it. It MUST support prereleases and expose dispatch evidence identifying
the input tag and resulting prerelease state.

#### Scenario: Prerelease evidence

- GIVEN a valid prerelease such as `v1.2.3-rc.1`
- WHEN publication completes
- THEN the release is visibly marked prerelease and the run records that evidence

#### Scenario: Existing release is protected

- GIVEN the requested tag or release already exists
- WHEN the publish workflow runs
- THEN it fails without modifying the existing tag or release

### Requirement: Preserve scope boundaries

The workflow MUST NOT publish packages, sign artifacts, generate changelogs, or
trigger automatically; no behavior outside release publication is required.

#### Scenario: No scope creep

- GIVEN a publish workflow completes successfully
- WHEN repository and package destinations are inspected
- THEN only the requested GitHub Release is created
