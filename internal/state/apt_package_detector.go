package state

import (
	"context"
	"strings"
	"time"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/execution"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

const aptPackagePresenceTimeout = 30 * time.Second

const unknownAptPackagePresenceReason = "APT package presence could not be determined; no mutation attempted"

// Valid dpkg status values for the three-field Status output.
// Any successful stdout that parses to three fields but contains an unknown
// value is treated as ambiguous/unknown so it never dispatches the installer.
var (
	aptDesiredActions = map[string]bool{
		"install":   true,
		"hold":      true,
		"deinstall": true,
		"purge":     true,
		"unknown":   true,
	}
	aptErrorFlags = map[string]bool{
		"ok":        true,
		"reinstreq": true,
	}
	aptPackageStatuses = map[string]bool{
		"not-installed":    true,
		"config-files":     true,
		"half-installed":   true,
		"unpacked":         true,
		"half-configured":  true,
		"triggers-awaited": true,
		"triggers-pending": true,
		"installed":        true,
	}
)

// AptPackageDetector probes only eligible APT packages through injected seams.
type AptPackageDetector struct {
	CommandExists func(string) bool
	Runner        execution.CommandRunner
	Timeout       time.Duration
}

// Detect returns a presence outcome for each eligible APT package in plan order.
func (d AptPackageDetector) Detect(ctx context.Context, plan planning.Plan) map[planning.ResourceRef]planning.PackagePresence {
	presence := make(map[planning.ResourceRef]planning.PackagePresence)
	for _, step := range plan.Steps {
		if !isEligibleAptPackageStep(step) {
			continue
		}
		if d.CommandExists == nil || !d.CommandExists("dpkg-query") || d.Runner == nil {
			presence[step.Ref] = planning.PackagePresenceUnknown
			continue
		}
		timeout := d.Timeout
		if timeout <= 0 {
			timeout = aptPackagePresenceTimeout
		}
		packageName := strings.TrimSpace(step.Resource.Install.Package)
		result := d.Runner.RunCommand(ctx, execution.CommandRequest{
			Executable: "dpkg-query",
			Args:       []string{"--show", "--showformat=${Status}", packageName},
			Timeout:    timeout,
		})
		presence[step.Ref] = classifyAptPackageResult(packageName, result)
	}
	return presence
}

func classifyAptPackageResult(packageName string, result execution.CommandResult) planning.PackagePresence {
	if result.Err == nil && result.Status == execution.CommandStatusSucceeded && result.ExitCode == 0 {
		fields := strings.Fields(result.Stdout)
		if len(fields) != 3 {
			return planning.PackagePresenceUnknown
		}
		if !aptDesiredActions[fields[0]] || !aptErrorFlags[fields[1]] || !aptPackageStatuses[fields[2]] {
			return planning.PackagePresenceUnknown
		}
		if fields[1] == "ok" && fields[2] == "installed" {
			return planning.PackagePresenceInstalled
		}
		return planning.PackagePresenceAbsent
	}
	if result.Status == execution.CommandStatusFailed && result.ExitCode == 1 {
		expected := "dpkg-query: no packages found matching " + packageName
		if strings.TrimSpace(result.Stdout) == "" && strings.Contains(result.Stderr, expected) {
			return planning.PackagePresenceAbsent
		}
	}
	return planning.PackagePresenceUnknown
}

// ApplyAptPackagePresence returns an execution-plan copy decorated with detected presence.
func ApplyAptPackagePresence(plan planning.Plan, presence map[planning.ResourceRef]planning.PackagePresence) planning.Plan {
	copyPlan := planning.Plan{Steps: append([]planning.PlanStep(nil), plan.Steps...)}
	for index := range copyPlan.Steps {
		step := &copyPlan.Steps[index]
		if !isEligibleAptPackageStep(*step) {
			continue
		}
		state, ok := presence[step.Ref]
		if !ok {
			continue
		}
		step.PackagePresence = state
		if state == planning.PackagePresenceUnknown {
			step.AttentionReasons = append(append([]string(nil), step.AttentionReasons...), unknownAptPackagePresenceReason)
		}
	}
	return copyPlan
}

func isEligibleAptPackageStep(step planning.PlanStep) bool {
	return step.Ref.Kind == planning.ResourceKindPackage &&
		step.Resource.Install != nil &&
		step.Resource.Install.Provider == "apt" &&
		strings.TrimSpace(step.Resource.Install.Package) != ""
}
