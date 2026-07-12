package execution

import (
	"context"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

// Runner executes a planning.Plan sequentially, dispatching each step to the
// Installer registered for the step's resource kind.
type Runner struct {
	installers map[planning.ResourceKind]Installer
}

// NewRunner creates a Runner from the provided installers.
func NewRunner(installers ...Installer) *Runner {
	registry := make(map[planning.ResourceKind]Installer, len(installers))
	for _, inst := range installers {
		registry[inst.SupportedKind()] = inst
	}
	return &Runner{installers: registry}
}

// Run executes the plan in order and returns an execution report for all steps.
// Execution does not stop on not_implemented or failed statuses; each step is
// processed and recorded.
func (r *Runner) Run(ctx context.Context, plan planning.Plan) ExecutionReport {
	report := ExecutionReport{Results: make([]StepResult, 0, len(plan.Steps))}
	for _, step := range plan.Steps {
		if isAlreadyInstalledCommandStep(step) {
			report.Results = append(report.Results, StepResult{
				Ref:     step.Ref,
				Status:  StepStatusSkipped,
				Message: "already installed; no mutation attempted",
			})
			continue
		}
		inst, ok := r.installers[step.Ref.Kind]
		if !ok {
			report.Results = append(report.Results, StepResult{
				Ref:     step.Ref,
				Status:  StepStatusNotImplemented,
				Message: "no installer registered for kind",
			})
			continue
		}
		report.Results = append(report.Results, inst.Install(ctx, step))
	}
	return report
}

func isAlreadyInstalledCommandStep(step planning.PlanStep) bool {
	presence := step.Resource.Presence
	return step.Status == planning.PlanStepStatusAlreadyInstalled &&
		(step.Ref.Kind == planning.ResourceKindTool || step.Ref.Kind == planning.ResourceKindRuntime) &&
		presence != nil && presence.Kind == "command_exists" && presence.Name != ""
}
