package execution

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrEmptyExecutable is returned when a CommandRequest has no executable.
var ErrEmptyExecutable = errors.New("command executable is empty")

// ErrShellFirstNotSupported is returned when an executable looks like a shell
// string. The command model is executable-plus-args only; there is no shell
// string field, no sh -c default, and no pipeline support.
var ErrShellFirstNotSupported = errors.New("shell-first command input is not supported")

// CommandRequest is an explicit executable-plus-args command description.
// It intentionally contains no shell string field, no default shell wrapper,
// and no pipeline support, so callers cannot accidentally opt in to shell
// interpretation.
type CommandRequest struct {
	Executable string
	Args       []string
	Dir        string
	Env        []string
	Timeout    time.Duration
}

// CommandStatus is the outcome of a command execution.
type CommandStatus string

const (
	// CommandStatusSucceeded indicates the process exited with code 0.
	CommandStatusSucceeded CommandStatus = "succeeded"
	// CommandStatusFailed indicates the process exited non-zero or could not
	// be started.
	CommandStatusFailed CommandStatus = "failed"
	// CommandStatusTimedOut indicates the command was cancelled by context or
	// timeout before completing successfully.
	CommandStatusTimedOut CommandStatus = "timed_out"
	// CommandStatusNotRun indicates the command was never executed, for
	// example during validation failure or dry-run mode.
	CommandStatusNotRun CommandStatus = "not_run"
)

// CommandResult captures the outcome of running a command.
type CommandResult struct {
	Request  CommandRequest
	Status   CommandStatus
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Err      error
}

// CommandRunner executes a CommandRequest and returns a CommandResult.
type CommandRunner interface {
	RunCommand(context.Context, CommandRequest) CommandResult
}

// ValidateCommandRequest rejects commands that cannot be represented as
// executable-plus-args. Empty executables and shell-first input are refused
// before any process is started.
func ValidateCommandRequest(req CommandRequest) error {
	if strings.TrimSpace(req.Executable) == "" {
		return ErrEmptyExecutable
	}
	if containsShellMetacharacters(req.Executable) {
		return ErrShellFirstNotSupported
	}
	return nil
}

// containsShellMetacharacters reports whether s contains whitespace or common
// shell metacharacters. This guards against callers passing a shell string as
// the executable name.
func containsShellMetacharacters(s string) bool {
	const shellChars = " \t\n|;&<>()`$\\\"'"
	return strings.ContainsAny(s, shellChars)
}
