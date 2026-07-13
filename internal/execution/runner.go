package execution

import (
	"context"
	"errors"
	"strings"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

var (
	ErrBrewFormulaPresenceUnknown = errors.New("Homebrew formula presence could not be determined")
	ErrAptPackagePresenceUnknown  = errors.New("APT package presence could not be determined")
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
		if isAlreadyInstalledCommandStep(step) || isInstalledBrewFormulaStep(step) || isInstalledAptPackageStep(step) {
			report.Results = append(report.Results, StepResult{
				Ref:     step.Ref,
				Status:  StepStatusSkipped,
				Message: "already installed; no mutation attempted",
			})
			continue
		}
		if isUnknownBrewFormulaStep(step) {
			report.Results = append(report.Results, StepResult{
				Ref:     step.Ref,
				Status:  StepStatusFailed,
				Message: "Homebrew formula presence could not be determined; no mutation attempted",
				Err:     ErrBrewFormulaPresenceUnknown,
			})
			continue
		}
		if isUnknownAptPackageStep(step) {
			report.Results = append(report.Results, StepResult{
				Ref:     step.Ref,
				Status:  StepStatusFailed,
				Message: "APT package presence could not be determined; no mutation attempted",
				Err:     ErrAptPackagePresenceUnknown,
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

func isInstalledBrewFormulaStep(step planning.PlanStep) bool {
	return isEligibleBrewFormulaStep(step) && step.PackagePresence == planning.PackagePresenceInstalled
}

func isUnknownBrewFormulaStep(step planning.PlanStep) bool {
	return isEligibleBrewFormulaStep(step) && step.PackagePresence == planning.PackagePresenceUnknown
}

func isEligibleBrewFormulaStep(step planning.PlanStep) bool {
	return step.Ref.Kind == planning.ResourceKindPackage &&
		step.Resource.Install != nil &&
		step.Resource.Install.Provider == "brew" &&
		strings.TrimSpace(step.Resource.Install.Package) != ""
}

func isInstalledAptPackageStep(step planning.PlanStep) bool {
	return isEligibleAptPackageStep(step) && step.PackagePresence == planning.PackagePresenceInstalled
}

func isUnknownAptPackageStep(step planning.PlanStep) bool {
	return isEligibleAptPackageStep(step) && step.PackagePresence == planning.PackagePresenceUnknown
}

func isEligibleAptPackageStep(step planning.PlanStep) bool {
	return step.Ref.Kind == planning.ResourceKindPackage &&
		step.Resource.Install != nil &&
		step.Resource.Install.Provider == "apt" &&
		strings.TrimSpace(step.Resource.Install.Package) != ""
}
