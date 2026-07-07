package execution

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

// OSCommandRunner runs commands as real OS processes using exec.CommandContext.
// It honors cwd/env, captures stdout/stderr, measures duration, and reports
// exit codes. It does not interpret shell strings.
type OSCommandRunner struct{}

// NewOSCommandRunner creates a real OS process runner.
func NewOSCommandRunner() *OSCommandRunner {
	return &OSCommandRunner{}
}

// RunCommand executes req as an OS process. The request is validated first;
// if it is invalid, a CommandResult with StatusNotRun is returned. A
// non-positive Timeout leaves the supplied context unchanged; otherwise a
// child timeout is applied.
func (r *OSCommandRunner) RunCommand(ctx context.Context, req CommandRequest) CommandResult {
	result := CommandResult{
		Request:  req,
		Status:   CommandStatusNotRun,
		ExitCode: -1,
	}

	if err := ValidateCommandRequest(req); err != nil {
		result.Err = err
		return result
	}

	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, req.Executable, req.Args...)
	if req.Dir != "" {
		cmd.Dir = req.Dir
	}
	if req.Env != nil {
		cmd.Env = req.Env
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	result.Duration = time.Since(start)

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if runErr != nil {
		if ctx.Err() != nil {
			result.Status = CommandStatusTimedOut
			result.Err = ctx.Err()
			return result
		}

		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			result.Status = CommandStatusFailed
			result.ExitCode = exitErr.ExitCode()
			result.Err = exitErr
			return result
		}

		result.Status = CommandStatusFailed
		result.Err = runErr
		return result
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	} else {
		result.ExitCode = 0
	}
	result.Status = CommandStatusSucceeded
	return result
}
