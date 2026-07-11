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

// BrewOrAptInstaller keeps Runner dispatch keyed by resource kind while routing
// only explicit brew and apt provider metadata to their respective delegates.
func BrewOrAptInstaller(kind planning.ResourceKind, brew, apt Installer) Installer {
	return &brewOrAptInstaller{kind: kind, brew: brew, apt: apt}
}

type brewOrAptInstaller struct {
	kind      planning.ResourceKind
	brew, apt Installer
}

func (i *brewOrAptInstaller) SupportedKind() planning.ResourceKind { return i.kind }

func (i *brewOrAptInstaller) Install(ctx context.Context, step planning.PlanStep) StepResult {
	if step.Resource.Install == nil {
		return StepResult{Ref: step.Ref, Status: StepStatusNotImplemented, Message: "no install metadata for this resource", Err: ErrNotImplemented}
	}
	var delegate Installer
	switch step.Resource.Install.Provider {
	case "brew":
		delegate = i.brew
	case "apt":
		delegate = i.apt
	default:
		return StepResult{Ref: step.Ref, Status: StepStatusNotImplemented, Message: "unsupported install provider for this resource", Err: ErrNotImplemented}
	}
	if delegate == nil {
		return StepResult{Ref: step.Ref, Status: StepStatusNotImplemented, Message: "no installer registered for this provider", Err: ErrNotImplemented}
	}
	return delegate.Install(ctx, step)
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
