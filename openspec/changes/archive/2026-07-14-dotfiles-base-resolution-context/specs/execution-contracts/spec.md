# Delta for execution-contracts

## ADDED Requirements

### Requirement: Base diagnostic context is execution-owned and safe to render

Execution results MUST transport base source, attempted candidate, optional canonical base, selected modules, and a safe cause as a dedicated base diagnostic context. The context MUST distinguish attempted from canonical paths and MUST NOT expose a canonical path for an unresolved or rejected base.

#### Scenario: Resolution failure renders the attempted candidate

- GIVEN base resolution fails for an environment or home-convention candidate
- WHEN the execution result is rendered
- THEN source, attempted candidate, selected modules, and safe cause are shown
- AND the candidate is not labeled as canonical base

#### Scenario: Validated base renders canonical identity

- GIVEN base resolution and filesystem validation succeed
- WHEN the execution result is rendered
- THEN the canonical base is shown as canonical
- AND the diagnostic remains associated with the validated filesystem identity

#### Scenario: Empty environment value is terminal

- GIVEN `DBOOTSTRAP_DOTFILES_DIR` is explicitly empty
- WHEN base resolution is requested
- THEN the result contains a safe base-resolution failure
- AND no home-convention fallback or executable context is shown

### Requirement: Base diagnostic rendering excludes unrelated execution detail

Base diagnostics MUST render deterministically and MUST be limited to base-resolution context, filesystem identity, selected modules, and safe cause. Rendering MUST NOT infer or include dotlink report outcomes, execution stderr, multi-cause aggregation, or semantic deduplication.

#### Scenario: Base failure remains isolated

- GIVEN a result contains a base-resolution failure
- WHEN its diagnostic is rendered
- THEN only the base diagnostic fields are rendered
- AND no dotlink execution or report detail is invented

## REMOVED Requirements

### Requirement: None

(Reason: No requirement is removed in this change.)
(Migration: None)
