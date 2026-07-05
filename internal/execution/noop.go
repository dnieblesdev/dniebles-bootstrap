package execution

import (
	"context"
	"errors"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

// ErrNotImplemented is returned by noop execution paths to signal that the
// requested operation has no real implementation in the current slice.
var ErrNotImplemented = errors.New("not_implemented")

// NoopInstaller is a safe installer stub that returns not_implemented without
// performing any command execution or host mutation.
type NoopInstaller struct{}

// SupportedKind reports the resource kind this installer claims to support.
// Noop installers match no concrete kind by default.
func (NoopInstaller) SupportedKind() planning.ResourceKind { return "" }

// Install returns a not_implemented result without executing anything.
func (NoopInstaller) Install(_ context.Context, step planning.PlanStep) StepResult {
	return StepResult{
		Ref:     step.Ref,
		Status:  StepStatusNotImplemented,
		Message: "noop installer does not perform real installation",
	}
}

// NoopDotfilesProvider is a safe provider stub that returns not_implemented
// without cloning, applying, installing, or mutating dotfiles.
type NoopDotfilesProvider struct{}

// EnsureModules returns ErrNotImplemented without touching the filesystem.
func (NoopDotfilesProvider) EnsureModules(_ context.Context, _ []string) error {
	return ErrNotImplemented
}

// RunDotlink returns ErrNotImplemented without invoking dotlink.
func (NoopDotfilesProvider) RunDotlink(_ context.Context, _ []string) error {
	return ErrNotImplemented
}
