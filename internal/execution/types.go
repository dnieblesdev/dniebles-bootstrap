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
	Ref     planning.ResourceRef
	Status  StepStatus
	Message string
	Err     error
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
