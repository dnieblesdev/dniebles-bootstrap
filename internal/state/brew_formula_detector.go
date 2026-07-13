package state

import (
	"context"
	"strings"
	"time"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/execution"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

const brewFormulaPresenceTimeout = 30 * time.Second

const unknownBrewFormulaPresenceReason = "Homebrew formula presence could not be determined; no mutation attempted"

// BrewFormulaDetector probes only eligible Brew package formulas through injected seams.
type BrewFormulaDetector struct {
	CommandExists func(string) bool
	Runner        execution.CommandRunner
	Timeout       time.Duration
}

// Detect returns a presence outcome for each eligible Brew package in plan order.
func (d BrewFormulaDetector) Detect(ctx context.Context, plan planning.Plan) map[planning.ResourceRef]planning.PackagePresence {
	presence := make(map[planning.ResourceRef]planning.PackagePresence)
	for _, step := range plan.Steps {
		if !isEligibleBrewFormulaStep(step) {
			continue
		}
		if d.CommandExists == nil || !d.CommandExists("brew") || d.Runner == nil {
			presence[step.Ref] = planning.PackagePresenceUnknown
			continue
		}
		timeout := d.Timeout
		if timeout <= 0 {
			timeout = brewFormulaPresenceTimeout
		}
		packageName := strings.TrimSpace(step.Resource.Install.Package)
		result := d.Runner.RunCommand(ctx, execution.CommandRequest{
			Executable: "brew",
			Args:       []string{"list", "--formula", packageName},
			Timeout:    timeout,
		})
		presence[step.Ref] = classifyBrewFormulaResult(result)
	}
	return presence
}

func classifyBrewFormulaResult(result execution.CommandResult) planning.PackagePresence {
	if result.Err == nil && result.Status == execution.CommandStatusSucceeded && result.ExitCode == 0 {
		return planning.PackagePresenceInstalled
	}
	if result.Status == execution.CommandStatusFailed && result.ExitCode == 1 && strings.Contains(result.Stderr, "No such keg") {
		return planning.PackagePresenceAbsent
	}
	return planning.PackagePresenceUnknown
}

// ApplyBrewFormulaPresence returns an execution-plan copy decorated with detected presence.
func ApplyBrewFormulaPresence(plan planning.Plan, presence map[planning.ResourceRef]planning.PackagePresence) planning.Plan {
	copyPlan := planning.Plan{Steps: append([]planning.PlanStep(nil), plan.Steps...)}
	for index := range copyPlan.Steps {
		step := &copyPlan.Steps[index]
		if !isEligibleBrewFormulaStep(*step) {
			continue
		}
		state, ok := presence[step.Ref]
		if !ok {
			continue
		}
		step.PackagePresence = state
		if state == planning.PackagePresenceUnknown {
			step.AttentionReasons = append(append([]string(nil), step.AttentionReasons...), unknownBrewFormulaPresenceReason)
		}
	}
	return copyPlan
}

func isEligibleBrewFormulaStep(step planning.PlanStep) bool {
	return step.Ref.Kind == planning.ResourceKindPackage &&
		step.Resource.Install != nil &&
		step.Resource.Install.Provider == "brew" &&
		strings.TrimSpace(step.Resource.Install.Package) != ""
}
