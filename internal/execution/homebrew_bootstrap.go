package execution

import (
	"context"
	"errors"
	"os/exec"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

// CommandExists reports whether a command named name is available on the host.
// It is a safe presence seam backed by exec.LookPath and does not spawn a
// process or shell.
type CommandExists func(name string) bool

type HomebrewAcquisitionResult struct {
	Acquired               bool
	PackageDispatchAllowed bool
	Err                    error
}

var ErrHomebrewAcquisitionUnavailable = errors.New("Homebrew acquisition unavailable outside Linux/WSL")

// AcquireHomebrew is deliberately terminal: it only prepares Homebrew and never
// dispatches target package installation.
func AcquireHomebrew(ctx context.Context, facts planning.EnvironmentFacts) HomebrewAcquisitionResult {
	if facts.OS != "linux" {
		return HomebrewAcquisitionResult{Err: ErrHomebrewAcquisitionUnavailable}
	}
	return acquireHomebrewLinux(ctx)
}

// BrewCommandExists reports whether the brew command is present on the host.
func BrewCommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

const homebrewDocumentationURL = "https://brew.sh/"

// AppendHomebrewBootstrap enriches report with a manual Homebrew bootstrap
// action when the plan contains brew-backed resources and brew is not present
// on the host. It only reads plan metadata and calls exists("brew"); it never
// executes the install command or installs target packages.
func AppendHomebrewBootstrap(report ExecutionReport, plan planning.Plan, exists CommandExists) ExecutionReport {
	if !planNeedsHomebrew(plan) {
		return report
	}
	if exists("brew") {
		return report
	}

	report.ManualActions = append(report.ManualActions, ManualAction{
		ID:     "homebrew:bootstrap",
		Title:  "Install Homebrew",
		Reason: "Homebrew is required by selected resources but is not installed on this host.",
		Instructions: []string{
			"Review the official Homebrew installation documentation before making host changes:",
			homebrewDocumentationURL,
			"Install Homebrew manually only after you understand the documented steps, then re-run dbootstrap apply.",
		},
	})
	return report
}

func planNeedsHomebrew(plan planning.Plan) bool {
	for _, step := range plan.Steps {
		if step.Resource.Install != nil && step.Resource.Install.Provider == "brew" {
			return true
		}
	}
	return false
}
