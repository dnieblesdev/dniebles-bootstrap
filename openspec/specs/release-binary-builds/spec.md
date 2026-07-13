# Release Binary Builds Specification

## Purpose

Define manually initiated GitHub Actions builds that produce versioned,
self-contained binary archives for supported release targets without
publishing a release.

## Requirements

### Requirement: Manually build supported binary archives

The workflow MUST run only when manually dispatched and MUST produce static
`dbootstrap` archives for Linux amd64, Linux arm64, and Windows amd64. Linux
archives MUST use `tar.gz`; the Windows archive MUST use `zip`.

#### Scenario: Dispatch creates the target archives

- GIVEN a maintainer dispatches the binary-build workflow
- WHEN the workflow completes successfully
- THEN it produces one archive for each of Linux amd64, Linux arm64, and Windows amd64
- AND each archive uses the required target-specific format

#### Scenario: Unsupported automatic trigger is ignored

- GIVEN a push or pull request occurs
- WHEN GitHub evaluates the binary-build workflow triggers
- THEN no binary-build workflow run is created

### Requirement: Inject and expose the build version

The build MUST accept a dispatch version when supplied, inject that version
into the binary, and expose it through `dbootstrap --version`. A build with no
injected version MUST report the local default `dev`.

#### Scenario: Dispatch version is reported

- GIVEN a maintainer dispatches a build with version `v1.2.3`
- WHEN the resulting binary is invoked with `--version`
- THEN it reports `v1.2.3`

#### Scenario: Local default remains available

- GIVEN the binary is built without version injection
- WHEN it is invoked with `--version`
- THEN it reports `dev`

### Requirement: Package the catalog with every binary

Every archive MUST contain its target binary and `catalog/bootstrap.toml` at
the expected package path. The workflow MUST have permission to read repository
contents required to check out and package these files.

#### Scenario: Archive is self-contained

- GIVEN a target build succeeds
- WHEN its archive is inspected
- THEN it contains the executable and `catalog/bootstrap.toml`

### Requirement: Generate and upload checksummed artifacts

The workflow MUST generate a SHA-256 checksum for every archive and MUST upload
the archives and checksum files as one cataloged workflow artifact bundle.

#### Scenario: Successful run exposes verifiable outputs

- GIVEN all target builds succeed
- WHEN the workflow artifact is downloaded
- THEN it contains all three archives and their SHA-256 checksum files
- AND each checksum identifies the corresponding archive

#### Scenario: Build failure prevents incomplete delivery

- GIVEN any target compilation or packaging step fails
- WHEN the workflow finishes
- THEN the workflow is unsuccessful
- AND it MUST NOT upload a successful complete artifact bundle

### Requirement: Exclude release publishing

The binary-build workflow MUST NOT create releases, tags, package channels, or
other publishing destinations, whether direct or called; outputs remain artifacts.
(Previously: only manually dispatched builds were specified as artifact-only.)

#### Scenario: Manual build does not publish

- GIVEN a binary-build workflow completes successfully
- WHEN its side effects are inspected
- THEN no release, tag, or external publication is created

#### Scenario: Reusable build does not publish

- GIVEN the publish workflow invokes the binary-build workflow
- WHEN the called build completes
- THEN it exposes artifacts without creating a release or tag
