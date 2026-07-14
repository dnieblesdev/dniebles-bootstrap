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
	if provider, ok := i.provider.(DotfilesExecutionContextProvider); ok {
		baseResolution := provider.ResolveDotfilesExecutionContext(modules)
		diagnostic := baseResolution.Diagnostic
		report, err := provider.RunDotlinkReportWithExecutionContext(ctx, modules, baseResolution)
		if err != nil {
			result := translateDotlinkReport(step, report, "", &diagnostic)
			result.Status = StepStatusFailed
			result.Message = fmt.Sprintf("dotfile module %s failed", module)
			result.Err = err
			var failure *DotfilesFailure
			if errors.As(err, &failure) {
				result.DotfilesFailure = failure
			}
			return result
		}
		return translateDotlinkReport(step, report, "", &diagnostic)
	}
	var diagnostic *DotfilesBaseDiagnostic
	if reporter, ok := i.provider.(DotfilesBaseDiagnosticReporter); ok {
		value := reporter.DotfilesBaseDiagnostic(modules)
		diagnostic = &value
	}
	reportProvider, ok := i.provider.(DotlinkReportProvider)
	if !ok {
		if err := i.provider.RunDotlink(ctx, modules); err != nil {
			return failedDotfilesStep(step, module, err, "", diagnostic)
		}
		return StepResult{Ref: step.Ref, Status: StepStatusInstalled, Message: fmt.Sprintf("installed dotfile module %s", module), BaseDiagnostic: diagnostic}
	}
	report, err := reportProvider.RunDotlinkReport(ctx, modules)
	if err != nil {
		return failedDotfilesStep(step, module, err, "", diagnostic)
	}
	return translateDotlinkReport(step, report, "", diagnostic)
}

func failedDotfilesStep(step planning.PlanStep, module string, err error, baseContext string, diagnostic *DotfilesBaseDiagnostic) StepResult {
	result := StepResult{Ref: step.Ref, Status: StepStatusFailed, Message: fmt.Sprintf("dotfile module %s failed", module), Err: err, BaseDiagnostic: diagnostic}
	var failure *DotfilesFailure
	if errors.As(err, &failure) {
		result.DotfilesFailure = failure
	}
	return result
}

func translateDotlinkReport(step planning.PlanStep, report DotlinkLinkReport, baseContext string, diagnostic *DotfilesBaseDiagnostic) StepResult {
	result := StepResult{Ref: step.Ref, Message: fmt.Sprintf("installed dotfile module %s%s", step.Ref.Name, baseContext), BaseDiagnostic: diagnostic}
	result.LinkDetails = make([]LinkDetail, 0, len(report.Entries))
	allUnchanged := len(report.Entries) > 0
	for _, entry := range report.Entries {
		detail := LinkDetail{Module: entry.Module, Source: entry.Source, Target: entry.Target, Outcome: LinkOutcome(entry.Outcome)}
		if entry.Cause != nil {
			detail.Cause = &LinkCause{Code: entry.Cause.Code, Message: entry.Cause.Message}
		}
		result.LinkDetails = append(result.LinkDetails, detail)
		if detail.Outcome != LinkOutcomeUnchanged {
			allUnchanged = false
		}
		if detail.Outcome == LinkOutcomeFailed || detail.Outcome == LinkOutcomeRolledBack {
			result.Status = StepStatusFailed
		}
	}
	if report.Failure != nil {
		result.Failure = &LinkFailure{Module: report.Failure.Module, Cause: LinkCause{Code: report.Failure.Cause.Code, Message: report.Failure.Cause.Message}}
		result.Status = StepStatusFailed
	}
	result.Rollback = LinkRollback{Attempted: report.Rollback.Attempted, Completed: report.Rollback.Completed, Removed: append([]string(nil), report.Rollback.Removed...)}
	if report.Status == DotlinkReportStatusFailed {
		result.Status = StepStatusFailed
	}
	if result.Status == StepStatusFailed {
		result.Message = fmt.Sprintf("dotfile module %s failed%s%s", step.Ref.Name, baseContext, rollbackRecoveryContext(result.Rollback))
		return result
	}
	if allUnchanged {
		result.Status = StepStatusSkipped
		result.Message = fmt.Sprintf("unchanged dotfile module %s%s", step.Ref.Name, baseContext)
		return result
	}
	result.Status = StepStatusInstalled
	return result
}

func rollbackRecoveryContext(rollback LinkRollback) string {
	return fmt.Sprintf(" (rollback attempted=%t completed=%t; recovery: verify rollback state and restore affected targets before retrying dotlink)", rollback.Attempted, rollback.Completed)
}
