# Delta for dotfiles-provider

## ADDED Requirements

### Requirement: Base resolution exposes safe typed context

The provider MUST carry base source, attempted candidate, selected modules, and a safe resolution cause as structured context. A canonical base MUST be populated only after filesystem validation succeeds; failed resolution MUST leave it empty. Filesystem failures MUST preserve identity for `errors.Is` and `errors.As` classification.

#### Scenario: Valid base has canonical identity

- GIVEN a configured base candidate resolves to an existing safe directory
- WHEN the provider validates the base
- THEN the context records its source and attempted candidate
- AND the canonical base contains the resolved filesystem identity

#### Scenario: Failed base retains attempted identity only

- GIVEN an empty, missing, unsafe, or non-directory candidate
- WHEN base resolution fails
- THEN the context records the source, attempted candidate, selected modules, and safe cause
- AND canonical base is empty

#### Scenario: Wrapped filesystem errors remain classifiable

- GIVEN a filesystem operation returns a wrapped missing or invalid-path error
- WHEN the failure is propagated
- THEN callers can classify the underlying cause with `errors.Is` or `errors.As`

### Requirement: Dotlink executable requires a validated canonical base

The provider MUST NOT construct or expose the dotlink executable path until a canonical base has been resolved and validated. Once valid, the executable context MUST be derived from that canonical base and remain associated with the same filesystem identity.

#### Scenario: Invalid base omits executable context

- GIVEN base canonicalization or validation fails
- WHEN provider prerequisites are evaluated
- THEN no dotlink executable path is derived
- AND the command runner is not called

#### Scenario: Valid base derives executable context

- GIVEN a canonical validated base exists
- WHEN provider prerequisites are evaluated
- THEN executable context is derived beneath that canonical base

## REMOVED Requirements

### Requirement: None

(Reason: No requirement is removed in this change.)
(Migration: None)
