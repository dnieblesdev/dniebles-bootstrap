package execution

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

const aptCommandTimeout = 10 * time.Minute

var (
	ErrUnsafeInstallPackage = errors.New("unsafe install package")
	ErrMissingApt           = errors.New("missing apt-get executable")
	ErrMissingSudo          = errors.New("missing sudo executable")
)

type AptExecutionReason string

const AptExecutionUnsupportedOS AptExecutionReason = "unsupported_os"

// AptExecutionError describes a failed APT execution path without claiming rollback.
type AptExecutionError struct {
	Reason        AptExecutionReason
	OS            string
	CommandStatus CommandStatus
	ExitCode      int
}

func (e *AptExecutionError) Error() string {
	return fmt.Sprintf("apt execution %s on %s (command status %s)", e.Reason, e.OS, e.CommandStatus)
}

// AptInstaller uses an injected executable-plus-arguments command seam.
type AptInstaller struct {
	kind   planning.ResourceKind
	runner CommandRunner
	exists CommandExists
	sudo   bool
}

func NewAptInstaller(kind planning.ResourceKind, runner CommandRunner, exists CommandExists, sudo bool) *AptInstaller {
	return &AptInstaller{kind: kind, runner: runner, exists: exists, sudo: sudo}
}

func (i *AptInstaller) SupportedKind() planning.ResourceKind { return i.kind }

func (i *AptInstaller) Install(ctx context.Context, step planning.PlanStep) StepResult {
	packageName, err := aptPackage(step.Resource.Install)
	if err != nil {
		return failedStep(step, err.Error(), err)
	}
	if i.exists == nil {
		return failedStep(step, ErrMissingCommandExists.Error(), ErrMissingCommandExists)
	}
	if !i.exists("apt-get") {
		return failedStep(step, "apt-get executable is not available on PATH", ErrMissingApt)
	}
	if i.sudo && !i.exists("sudo") {
		return failedStep(step, "sudo executable is not available on PATH", ErrMissingSudo)
	}
	if i.runner == nil {
		return failedStep(step, ErrMissingCommandRunner.Error(), ErrMissingCommandRunner)
	}

	request := CommandRequest{Executable: "apt-get", Args: []string{"install", "-y", "--", packageName}, Timeout: aptCommandTimeout}
	if i.sudo {
		request = CommandRequest{Executable: "sudo", Args: append([]string{"apt-get"}, request.Args...), Timeout: aptCommandTimeout}
	}
	result := i.runner.RunCommand(ctx, request)
	if result.Status == CommandStatusSucceeded {
		return StepResult{Ref: step.Ref, Status: StepStatusInstalled, Message: fmt.Sprintf("installed %s with APT", packageName)}
	}
	return StepResult{Ref: step.Ref, Status: StepStatusFailed, Message: fmt.Sprintf("apt install %s failed with status %s", packageName, result.Status), Err: commandResultError(result)}
}

func aptPackage(metadata *planning.InstallMetadata) (string, error) {
	if metadata == nil || metadata.Provider != "apt" {
		return "", ErrUnsupportedInstallProvider
	}
	packageName := strings.TrimSpace(metadata.Package)
	if packageName == "" {
		return "", ErrMissingInstallPackage
	}
	if strings.HasPrefix(packageName, "-") {
		return "", ErrUnsafeInstallPackage
	}
	return packageName, nil
}

type nonLinuxAptInstaller struct {
	kind planning.ResourceKind
	os   string
}

func NewNonLinuxAptInstaller(kind planning.ResourceKind, os string) Installer {
	return nonLinuxAptInstaller{kind: kind, os: os}
}

func (i nonLinuxAptInstaller) SupportedKind() planning.ResourceKind { return i.kind }

func (i nonLinuxAptInstaller) Install(_ context.Context, step planning.PlanStep) StepResult {
	err := &AptExecutionError{Reason: AptExecutionUnsupportedOS, OS: i.os, CommandStatus: CommandStatusNotRun, ExitCode: 1}
	return StepResult{Ref: step.Ref, Status: StepStatusFailed, Message: err.Error(), Err: err}
}
