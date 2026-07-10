package execution

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

var (
	ErrUnsupportedDotfilesStep = errors.New("unsupported dotfiles step")
	ErrMissingDotfilesProvider = errors.New("missing dotfiles provider")
)

type DotfilesBaseReporter interface {
	DotfilesBase() (ResolvedDotfilesBase, error)
}

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
	modules := []string{module}
	baseContext := i.baseContext(modules)
	if err := i.provider.RunDotlink(ctx, modules); err != nil {
		return failedStep(step, fmt.Sprintf("dotfile module %s failed: %v%s", module, err, baseContext), err)
	}
	return StepResult{
		Ref:     step.Ref,
		Status:  StepStatusInstalled,
		Message: fmt.Sprintf("installed dotfile module %s%s", module, baseContext),
	}
}

func (i *DotfilesInstaller) baseContext(modules []string) string {
	reporter, ok := i.provider.(DotfilesBaseReporter)
	if !ok {
		return ""
	}
	base, err := reporter.DotfilesBase()
	if err != nil {
		return ""
	}
	return fmt.Sprintf(" (dotfiles base: %s; source: %s; modules: %s)", base.CanonicalPath, base.Source, strings.Join(modules, ", "))
}
