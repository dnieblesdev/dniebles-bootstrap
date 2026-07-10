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
