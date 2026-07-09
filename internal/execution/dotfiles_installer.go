package execution

import (
	"context"
	"errors"
	"fmt"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

var (
	ErrUnsupportedDotfilesStep = errors.New("unsupported dotfiles step")
	ErrMissingDotfilesProvider = errors.New("missing dotfiles provider")
)

type DotfilesInstaller struct {
	provider DotfilesProvider
}

func NewDotfilesInstaller(provider DotfilesProvider) *DotfilesInstaller {
	return &DotfilesInstaller{provider: provider}
}

func (i *DotfilesInstaller) SupportedKind() planning.ResourceKind {
	return planning.ResourceKindDotfile
}

func (i *DotfilesInstaller) Install(ctx context.Context, step planning.PlanStep) StepResult {
	if step.Ref.Kind != planning.ResourceKindDotfile {
		return failedStep(step, ErrUnsupportedDotfilesStep.Error(), ErrUnsupportedDotfilesStep)
	}
	if i.provider == nil {
		return failedStep(step, ErrMissingDotfilesProvider.Error(), ErrMissingDotfilesProvider)
	}
	module := step.Ref.Name
	if err := i.provider.RunDotlink(ctx, []string{module}); err != nil {
		return failedStep(step, fmt.Sprintf("dotfile module %s failed", module), err)
	}
	return StepResult{
		Ref:     step.Ref,
		Status:  StepStatusInstalled,
		Message: fmt.Sprintf("installed dotfile module %s", module),
	}
}
