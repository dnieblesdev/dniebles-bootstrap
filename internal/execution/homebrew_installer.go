package execution

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

var (
	// ErrUnsupportedInstallProvider is returned when install metadata is absent
	// or does not select the Homebrew provider.
	ErrUnsupportedInstallProvider = errors.New("unsupported install provider")
	// ErrMissingInstallPackage is returned when Homebrew metadata omits a package.
	ErrMissingInstallPackage = errors.New("missing install package")
	// ErrMissingHomebrew is returned when brew is unavailable on PATH.
	ErrMissingHomebrew = errors.New("missing homebrew executable")
	// ErrMissingCommandRunner is returned when the installer has no runner seam.
	ErrMissingCommandRunner = errors.New("missing command runner")
	// ErrMissingCommandExists is returned when the installer has no presence seam.
	ErrMissingCommandExists = errors.New("missing command exists seam")
)

// HomebrewInstaller installs brew-backed resources through explicit command
// requests. CLI composition decides when it is safe to wire this installer.
type HomebrewInstaller struct {
	kind   planning.ResourceKind
	runner CommandRunner
	exists CommandExists
}

// NewHomebrewInstaller creates an installer for one resource kind using the
// supplied command runner and command-presence seam.
func NewHomebrewInstaller(kind planning.ResourceKind, runner CommandRunner, exists CommandExists) *HomebrewInstaller {
	return &HomebrewInstaller{kind: kind, runner: runner, exists: exists}
}

// SupportedKind reports the resource kind handled by this installer.
func (i *HomebrewInstaller) SupportedKind() planning.ResourceKind { return i.kind }

// Install validates structured brew metadata, checks that brew exists, and runs
// exactly `brew install <package>` through the injected CommandRunner.
func (i *HomebrewInstaller) Install(ctx context.Context, step planning.PlanStep) StepResult {
	packageName, err := brewPackage(step.Resource.Install)
	if err != nil {
		return failedStep(step, err.Error(), err)
	}
	if i.exists == nil {
		return failedStep(step, ErrMissingCommandExists.Error(), ErrMissingCommandExists)
	}
	if !i.exists("brew") {
		return failedStep(step, "brew executable is not available on PATH", ErrMissingHomebrew)
	}
	if i.runner == nil {
		return failedStep(step, ErrMissingCommandRunner.Error(), ErrMissingCommandRunner)
	}

	result := i.runner.RunCommand(ctx, CommandRequest{
		Executable: "brew",
		Args:       []string{"install", packageName},
	})
	if result.Status == CommandStatusSucceeded {
		return StepResult{
			Ref:     step.Ref,
			Status:  StepStatusInstalled,
			Message: fmt.Sprintf("installed %s with Homebrew", packageName),
		}
	}

	return StepResult{
		Ref:     step.Ref,
		Status:  StepStatusFailed,
		Message: fmt.Sprintf("brew install %s failed with status %s", packageName, result.Status),
		Err:     commandResultError(result),
	}
}

func brewPackage(metadata *planning.InstallMetadata) (string, error) {
	if metadata == nil || metadata.Provider != "brew" {
		return "", ErrUnsupportedInstallProvider
	}
	packageName := strings.TrimSpace(metadata.Package)
	if packageName == "" {
		return "", ErrMissingInstallPackage
	}
	return packageName, nil
}

func failedStep(step planning.PlanStep, message string, err error) StepResult {
	return StepResult{Ref: step.Ref, Status: StepStatusFailed, Message: message, Err: err}
}

func commandResultError(result CommandResult) error {
	if result.Err != nil {
		return result.Err
	}
	return fmt.Errorf("command status %s", result.Status)
}
