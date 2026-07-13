# GitHub Actions Build Validation Specification

## Purpose

Define the GitHub Actions gate that validates Go changes targeting the `main` branch without producing or distributing build artifacts.

## Requirements

### Requirement: Validate main branch changes in GitHub Actions

The repository MUST run a GitHub Actions validation workflow for every push to `main` and every pull request targeting `main`. The workflow MUST use the Go version declared by `go.mod` and MUST execute the checks sequentially in this order: `go test ./...`, `go vet ./...`, and `go build ./...`.

#### Scenario: Push to main passes all validation checks

- GIVEN a commit is pushed to the `main` branch
- WHEN GitHub Actions starts the validation workflow
- THEN it runs `go test ./...`, `go vet ./...`, and `go build ./...` in order
- AND the workflow succeeds when all three commands succeed

#### Scenario: Pull request targeting main runs validation

- GIVEN a pull request targets the `main` branch
- WHEN the pull request is opened, synchronized, or reopened
- THEN GitHub Actions runs the same three checks in the same order
- AND the workflow reports failure if any check fails

#### Scenario: Failed check prevents a successful validation result

- GIVEN one validation command exits unsuccessfully
- WHEN the workflow executes the checks
- THEN the workflow MUST report failure
- AND it MUST NOT report the commit or pull request as fully validated

### Requirement: Do not generate or publish artifacts

The validation workflow MUST NOT create, upload, publish, sign, release, or otherwise distribute build artifacts. `go build ./...` is a validation check only; its output MUST NOT be treated as a release or delivery artifact.

#### Scenario: Validation completes without artifact publication

- GIVEN a push or pull request triggers the workflow
- WHEN all validation commands complete
- THEN the workflow performs no artifact upload or publication
- AND it creates no repository-managed artifact output

#### Scenario: Build validation failure remains non-distributable

- GIVEN `go build ./...` fails
- WHEN the workflow ends
- THEN the workflow reports the validation failure
- AND it MUST NOT publish or upload partial build output
