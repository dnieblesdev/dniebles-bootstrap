# Delta for repository-guidance

## ADDED Requirements

### Requirement: Repository orientation docs

The system MUST provide README.md and AGENT.md as the primary repository orientation documents.

#### Scenario: README explains scope

- GIVEN a new contributor opens the repository
- WHEN they read README.md
- THEN they find purpose, boundaries, and major flows
- AND implementation details are not required to understand scope

#### Scenario: AGENT.md explains operating rules

- GIVEN an agent or contributor opens AGENT.md
- WHEN they follow the guide
- THEN they find the working rules for the repository
- AND the dotfiles boundary and SDD workflow are explicit

#### Scenario: Generated artifact language is explicit

- GIVEN an agent or contributor produces technical artifacts
- WHEN they check AGENT.md
- THEN generated docs, specs, code comments, and user-facing strings default to English
- AND repository artifacts do not inherit conversational language unless explicitly requested

### Requirement: Non-goals remain explicit

The system MUST state that this change does not implement Go application code and does not own dotfiles internals.

#### Scenario: Documentation clarifies exclusions

- GIVEN the change description is reviewed
- WHEN a reader checks non-goals
- THEN Go implementation is excluded from this change
- AND dotfiles internals remain owned externally

#### Scenario: Bootstrap scope stays reviewable

- GIVEN a future implementation proposal
- WHEN it is compared to this spec
- THEN the no-code boundary is visible
- AND future work can proceed without ambiguity

### Requirement: Local agent state remains untracked

The system MUST keep `.atl/` as local ignored agent registry state.

#### Scenario: Local registry is not versioned

- GIVEN `.atl/` exists locally
- WHEN repository guidance is followed
- THEN `.atl/` remains ignored by git
- AND no SDD artifact depends on committing `.atl/` content
