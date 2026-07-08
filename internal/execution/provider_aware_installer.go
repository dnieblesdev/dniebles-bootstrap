package execution

import (
	"context"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

// BrewOnlyInstaller routes only brew-backed resources to the supplied installer.
// Missing metadata or any other provider returns not_implemented without
// executing commands.
func BrewOnlyInstaller(kind planning.ResourceKind, brew Installer) Installer {
	return &providerAwareInstaller{kind: kind, provider: "brew", delegate: brew}
}

type providerAwareInstaller struct {
	kind     planning.ResourceKind
	provider string
	delegate Installer
}

func (i *providerAwareInstaller) SupportedKind() planning.ResourceKind { return i.kind }

func (i *providerAwareInstaller) Install(ctx context.Context, step planning.PlanStep) StepResult {
	metadata := step.Resource.Install
	if metadata == nil || metadata.Provider != i.provider {
		return StepResult{
			Ref:     step.Ref,
			Status:  StepStatusNotImplemented,
			Message: "no brew install metadata for this resource",
			Err:     ErrNotImplemented,
		}
	}
	if i.delegate == nil {
		return StepResult{
			Ref:     step.Ref,
			Status:  StepStatusNotImplemented,
			Message: "no brew installer registered",
			Err:     ErrNotImplemented,
		}
	}
	return i.delegate.Install(ctx, step)
}
