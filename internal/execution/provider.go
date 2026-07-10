package execution

import "context"

// DotfilesProvider is the high-level execution boundary for dotfiles workflows.
// It remains separate from the read-only dotfiles detector and owns no planning logic.
type DotfilesProvider interface {
	EnsureModules(ctx context.Context, modules []string) error
	RunDotlink(ctx context.Context, modules []string) error
}

// DotlinkReportProvider exposes the validated dotlink report boundary without
// forcing legacy provider consumers to translate report details prematurely.
type DotlinkReportProvider interface {
	RunDotlinkReport(ctx context.Context, modules []string) (DotlinkLinkReport, error)
}

// DotfilesBaseDiagnosticReporter supplies safe base-resolution facts for
// execution results without exposing command output.
type DotfilesBaseDiagnosticReporter interface {
	DotfilesBaseDiagnostic(modules []string) DotfilesBaseDiagnostic
}

// DotfilesExecutionContext keeps one resolved base and its safe diagnostic
// together so rendering and command execution cannot disagree.
type DotfilesExecutionContext struct {
	Base          ResolvedDotfilesBase
	Diagnostic    DotfilesBaseDiagnostic
	Err           error
	validatedBase string
}

type DotfilesExecutionContextProvider interface {
	ResolveDotfilesExecutionContext(modules []string) DotfilesExecutionContext
	RunDotlinkReportWithExecutionContext(context.Context, []string, DotfilesExecutionContext) (DotlinkLinkReport, error)
}
