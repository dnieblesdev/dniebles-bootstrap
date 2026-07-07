package execution

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

const helperEnv = "DBOOTSTRAP_TEST_HELPER"

// TestMain enables this test binary to be reused as a helper process. When the
// helper environment variable is set, runHelper is invoked instead of the test
// suite. This avoids depending on arbitrary host tools.
func TestMain(m *testing.M) {
	if os.Getenv(helperEnv) != "" {
		runHelper()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func runHelper() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "helper: missing subcommand")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "echo":
		fmt.Println(strings.Join(os.Args[2:], " "))
	case "stderr":
		fmt.Fprintln(os.Stderr, strings.Join(os.Args[2:], " "))
	case "exit":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "helper: usage exit <code> <stdout> <stderr>")
			os.Exit(2)
		}
		code, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "helper: bad exit code %q\n", os.Args[2])
			os.Exit(2)
		}
		fmt.Fprint(os.Stdout, os.Args[3])
		fmt.Fprint(os.Stderr, os.Args[4])
		os.Exit(code)
	case "sleep":
		// Block until the parent context cancels this process.
		time.Sleep(24 * time.Hour)
	default:
		fmt.Fprintf(os.Stderr, "helper: unknown subcommand %q\n", os.Args[1])
		os.Exit(2)
	}
}

// helperRequest builds a CommandRequest that re-executes this test binary in
// helper mode for the given subcommand.
func helperRequest(subcommand string, args ...string) CommandRequest {
	return CommandRequest{
		Executable: os.Args[0],
		Args:       append([]string{subcommand}, args...),
		Env:        append(os.Environ(), helperEnv+"="+subcommand),
	}
}

func TestValidateCommandRequestAcceptsArgvOnly(t *testing.T) {
	tests := []struct {
		name    string
		req     CommandRequest
		wantErr error
	}{
		{
			name:    "valid executable",
			req:     CommandRequest{Executable: "git"},
			wantErr: nil,
		},
		{
			name:    "valid executable with absolute path",
			req:     CommandRequest{Executable: "/usr/bin/git"},
			wantErr: nil,
		},
		{
			name:    "empty executable",
			req:     CommandRequest{Executable: ""},
			wantErr: ErrEmptyExecutable,
		},
		{
			name:    "whitespace-only executable",
			req:     CommandRequest{Executable: "   "},
			wantErr: ErrEmptyExecutable,
		},
		{
			name:    "shell string with spaces",
			req:     CommandRequest{Executable: "sh -c"},
			wantErr: ErrShellFirstNotSupported,
		},
		{
			name:    "pipeline metacharacter",
			req:     CommandRequest{Executable: "git|cat"},
			wantErr: ErrShellFirstNotSupported,
		},
		{
			name:    "shell substitution metacharacter",
			req:     CommandRequest{Executable: "$(which git)"},
			wantErr: ErrShellFirstNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommandRequest(tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateCommandRequest() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestOSCommandRunnerSuccess(t *testing.T) {
	runner := NewOSCommandRunner()
	req := helperRequest("echo", "hello", "world")

	result := runner.RunCommand(context.Background(), req)

	if result.Status != CommandStatusSucceeded {
		t.Fatalf("Status = %q, want %q", result.Status, CommandStatusSucceeded)
	}
	if result.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello world\n" {
		t.Fatalf("Stdout = %q, want %q", result.Stdout, "hello world\n")
	}
	if result.Stderr != "" {
		t.Fatalf("Stderr = %q, want empty", result.Stderr)
	}
	if result.Err != nil {
		t.Fatalf("Err = %v, want nil", result.Err)
	}
	if result.Request.Executable != req.Executable {
		t.Fatalf("Request.Executable = %q, want %q", result.Request.Executable, req.Executable)
	}
	if result.Duration < 0 {
		t.Fatalf("Duration = %v, want non-negative", result.Duration)
	}
}

func TestOSCommandRunnerCapturesStderr(t *testing.T) {
	runner := NewOSCommandRunner()
	req := helperRequest("stderr", "warning", "message")

	result := runner.RunCommand(context.Background(), req)

	if result.Status != CommandStatusSucceeded {
		t.Fatalf("Status = %q, want %q", result.Status, CommandStatusSucceeded)
	}
	if result.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "" {
		t.Fatalf("Stdout = %q, want empty", result.Stdout)
	}
	if result.Stderr != "warning message\n" {
		t.Fatalf("Stderr = %q, want %q", result.Stderr, "warning message\n")
	}
}

func TestOSCommandRunnerFailure(t *testing.T) {
	runner := NewOSCommandRunner()
	req := helperRequest("exit", "3", "stdout-data", "stderr-data")

	result := runner.RunCommand(context.Background(), req)

	if result.Status != CommandStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, CommandStatusFailed)
	}
	if result.ExitCode != 3 {
		t.Fatalf("ExitCode = %d, want 3", result.ExitCode)
	}
	if result.Stdout != "stdout-data" {
		t.Fatalf("Stdout = %q, want %q", result.Stdout, "stdout-data")
	}
	if result.Stderr != "stderr-data" {
		t.Fatalf("Stderr = %q, want %q", result.Stderr, "stderr-data")
	}
	if result.Err == nil {
		t.Fatal("expected non-nil error for non-zero exit")
	}
}

func TestOSCommandRunnerMissingExecutable(t *testing.T) {
	runner := NewOSCommandRunner()
	req := CommandRequest{Executable: "this-binary-does-not-exist-abc123"}

	result := runner.RunCommand(context.Background(), req)

	if result.Status != CommandStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, CommandStatusFailed)
	}
	if result.ExitCode != -1 {
		t.Fatalf("ExitCode = %d, want -1", result.ExitCode)
	}
	if result.Err == nil {
		t.Fatal("expected non-nil error for missing executable")
	}
}

func TestOSCommandRunnerValidationFailure(t *testing.T) {
	runner := NewOSCommandRunner()
	req := CommandRequest{Executable: ""}

	result := runner.RunCommand(context.Background(), req)

	if result.Status != CommandStatusNotRun {
		t.Fatalf("Status = %q, want %q", result.Status, CommandStatusNotRun)
	}
	if result.ExitCode != -1 {
		t.Fatalf("ExitCode = %d, want -1", result.ExitCode)
	}
	if !errors.Is(result.Err, ErrEmptyExecutable) {
		t.Fatalf("Err = %v, want ErrEmptyExecutable", result.Err)
	}
}

func TestOSCommandRunnerTimeout(t *testing.T) {
	runner := NewOSCommandRunner()
	req := helperRequest("sleep")
	req.Timeout = 50 * time.Millisecond

	result := runner.RunCommand(context.Background(), req)

	if result.Status != CommandStatusTimedOut {
		t.Fatalf("Status = %q, want %q", result.Status, CommandStatusTimedOut)
	}
	if result.ExitCode != -1 {
		t.Fatalf("ExitCode = %d, want -1", result.ExitCode)
	}
	if !errors.Is(result.Err, context.DeadlineExceeded) {
		t.Fatalf("Err = %v, want context.DeadlineExceeded", result.Err)
	}
}

func TestOSCommandRunnerExternalCancellation(t *testing.T) {
	runner := NewOSCommandRunner()
	req := helperRequest("sleep")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result := runner.RunCommand(ctx, req)

	if result.Status != CommandStatusTimedOut {
		t.Fatalf("Status = %q, want %q", result.Status, CommandStatusTimedOut)
	}
	if result.ExitCode != -1 {
		t.Fatalf("ExitCode = %d, want -1", result.ExitCode)
	}
	if !errors.Is(result.Err, context.Canceled) {
		t.Fatalf("Err = %v, want context.Canceled", result.Err)
	}
}

func TestNoopCommandRunnerDoesNotExecute(t *testing.T) {
	runner := NewNoopCommandRunner()
	req := helperRequest("echo", "should", "not", "run")

	result := runner.RunCommand(context.Background(), req)

	if result.Status != CommandStatusNotRun {
		t.Fatalf("Status = %q, want %q", result.Status, CommandStatusNotRun)
	}
	if result.ExitCode != -1 {
		t.Fatalf("ExitCode = %d, want -1", result.ExitCode)
	}
	if result.Stdout != "" {
		t.Fatalf("Stdout = %q, want empty", result.Stdout)
	}
	if result.Stderr != "" {
		t.Fatalf("Stderr = %q, want empty", result.Stderr)
	}
	if result.Err != nil {
		t.Fatalf("Err = %v, want nil", result.Err)
	}
	if result.Duration != 0 {
		t.Fatalf("Duration = %v, want 0", result.Duration)
	}
	if result.Request.Executable != req.Executable {
		t.Fatalf("Request was not preserved")
	}
}

func TestNoopCommandRunnerIsDeterministic(t *testing.T) {
	runner := NewNoopCommandRunner()
	req := helperRequest("echo", "x")

	first := runner.RunCommand(context.Background(), req)
	second := runner.RunCommand(context.Background(), req)

	if first.Status != second.Status {
		t.Fatalf("status mismatch: %q vs %q", first.Status, second.Status)
	}
	if first.ExitCode != second.ExitCode {
		t.Fatalf("exit code mismatch: %d vs %d", first.ExitCode, second.ExitCode)
	}
	if first.Stdout != second.Stdout || first.Stderr != second.Stderr {
		t.Fatal("output mismatch between noop runs")
	}
}
