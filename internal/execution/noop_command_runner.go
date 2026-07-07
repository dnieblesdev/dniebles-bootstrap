package execution

import "context"

// NoopCommandRunner is a deterministic dry-run command runner. It never
// starts a process, never mutates the host, and always returns the same
// not-run result shape for a given request.
type NoopCommandRunner struct{}

// NewNoopCommandRunner creates a deterministic no-op command runner.
func NewNoopCommandRunner() *NoopCommandRunner {
	return &NoopCommandRunner{}
}

// RunCommand returns a deterministic CommandResult without invoking the host.
// The request is preserved so callers can audit what would have run.
func (r *NoopCommandRunner) RunCommand(_ context.Context, req CommandRequest) CommandResult {
	return CommandResult{
		Request:  req,
		Status:   CommandStatusNotRun,
		ExitCode: -1,
	}
}
