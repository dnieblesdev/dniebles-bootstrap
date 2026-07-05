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

// ExecutionReport aggregates the results of executing a plan.
type ExecutionReport struct {
	Results []StepResult
}
