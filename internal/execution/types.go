package execution

import "github.com/dnieblesdev/dniebles-bootstrap/internal/planning"

// StepStatus is a runtime execution outcome. It is intentionally separate from
// planning.PlanStepStatus to avoid semantic drift between intended and executed work.
type StepStatus string

const (
	StepStatusInstalled      StepStatus = "installed"
	StepStatusFailed         StepStatus = "failed"
	StepStatusSkipped        StepStatus = "skipped"
	StepStatusNotImplemented StepStatus = "not_implemented"
)

// StepResult describes the outcome of executing a single plan step.
type StepResult struct {
	Ref            planning.ResourceRef
	Status         StepStatus
	Message        string
	Err            error
	LinkDetails    []LinkDetail
	Failure        *LinkFailure
	Rollback       LinkRollback
	BaseDiagnostic *DotfilesBaseDiagnostic
}

// LinkOutcome is the execution-owned outcome for one validated dotlink entry.
// It deliberately does not reuse the provider report type.
type LinkOutcome string

const (
	LinkOutcomeChanged    LinkOutcome = "changed"
	LinkOutcomeUnchanged  LinkOutcome = "unchanged"
	LinkOutcomeFailed     LinkOutcome = "failed"
	LinkOutcomeRolledBack LinkOutcome = "rolled_back"
)

// LinkDetail retains the validated, ordered facts for one link operation.
type LinkDetail struct {
	Module  string
	Source  string
	Target  string
	Outcome LinkOutcome
	Cause   *LinkCause
}

type LinkCause struct {
	Code    string
	Message string
}

type LinkFailure struct {
	Module string
	Cause  LinkCause
}

type LinkRollback struct {
	Attempted bool
	Completed bool
	Removed   []string
}

// DotfilesBaseDiagnostic holds safe base-resolution context. CanonicalPath is
// populated only after canonicalization and safety validation succeed.
type DotfilesBaseDiagnostic struct {
	Source             DotfilesBaseSource
	AttemptedCandidate string
	CanonicalPath      string
	Modules            []string
	Cause              string
}

// ManualAction is a provider-owned, non-mutating action that the operator must
// perform manually before the plan can be applied. It contains no executable
// fields and cannot be run by the engine.
type ManualAction struct {
	ID           string
	Title        string
	Reason       string
	Instructions []string
}

// ExecutionReport aggregates the results of executing a plan.
type ExecutionReport struct {
	Results       []StepResult
	ManualActions []ManualAction
}
