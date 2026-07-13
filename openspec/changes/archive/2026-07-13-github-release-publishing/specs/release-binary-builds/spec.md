# Delta for Release Binary Builds

## ADDED Requirements

### Requirement: Support reusable verified builds

The binary-build workflow MUST support release-publish invocation with an
explicit version while retaining manual dispatch. A successful called build
MUST expose the direct-build bundle, archives, and SHA-256 files.

#### Scenario: Publish workflow calls the build

- GIVEN the publish workflow supplies a valid release version
- WHEN the reusable build completes successfully
- THEN it exposes three archives and three matching checksums

#### Scenario: Direct manual behavior remains unchanged

- GIVEN a maintainer manually dispatches the build workflow
- WHEN it completes successfully
- THEN it produces artifacts only and does not publish

#### Scenario: Called build fails

- GIVEN compilation or packaging fails during a reusable invocation
- WHEN the build workflow finishes
- THEN no successful complete artifact bundle is exposed

## MODIFIED Requirements

### Requirement: Exclude release publishing

The binary-build workflow MUST NOT create releases, tags, package channels, or
other publishing destinations, whether direct or called; outputs remain artifacts.
(Previously: only manually dispatched builds were specified as artifact-only.)

#### Scenario: Manual build does not publish

- GIVEN a binary-build workflow is manually dispatched
- WHEN it completes successfully
- THEN no release, tag, or external publication is created

#### Scenario: Reusable build does not publish

- GIVEN the publish workflow invokes the binary-build workflow
- WHEN the called build completes
- THEN it exposes artifacts without creating a release or tag
