package execution

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestAptInstallerBuildsExplicitVectors(t *testing.T) {
	tests := []struct {
		name       string
		sudo       bool
		wantLookup []string
		want       CommandRequest
	}{
		{name: "direct", wantLookup: []string{"apt-get"}, want: CommandRequest{Executable: "apt-get", Args: []string{"install", "-y", "--", "ripgrep"}, Timeout: 10 * time.Minute}},
		{name: "explicit sudo", sudo: true, wantLookup: []string{"apt-get", "sudo"}, want: CommandRequest{Executable: "sudo", Args: []string{"apt-get", "install", "-y", "--", "ripgrep"}, Timeout: 10 * time.Minute}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded}}
			var lookups []string
			installer := NewAptInstaller(planning.ResourceKindPackage, runner, func(name string) bool { lookups = append(lookups, name); return true }, tt.sudo)

			result := installer.Install(context.Background(), aptStep("  ripgrep  ", "apt"))

			if result.Status != StepStatusInstalled {
				t.Fatalf("Status = %q, want %q", result.Status, StepStatusInstalled)
			}
			if !reflect.DeepEqual(lookups, tt.wantLookup) {
				t.Fatalf("lookups = %#v, want %#v", lookups, tt.wantLookup)
			}
			if !reflect.DeepEqual(runner.calls, []CommandRequest{tt.want}) {
				t.Fatalf("requests = %#v, want %#v", runner.calls, []CommandRequest{tt.want})
			}
		})
	}
}

func TestAptInstallerRejectsUnsafeMetadataWithoutProbing(t *testing.T) {
	tests := []struct {
		name     string
		metadata *planning.InstallMetadata
		wantErr  error
	}{
		{name: "missing", wantErr: ErrUnsupportedInstallProvider},
		{name: "other provider", metadata: &planning.InstallMetadata{Provider: "brew", Package: "ripgrep"}, wantErr: ErrUnsupportedInstallProvider},
		{name: "empty", metadata: &planning.InstallMetadata{Provider: "apt", Package: "  "}, wantErr: ErrMissingInstallPackage},
		{name: "option", metadata: &planning.InstallMetadata{Provider: "apt", Package: "--allow-unauthenticated"}, wantErr: ErrUnsafeInstallPackage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{}
			lookups := 0
			installer := NewAptInstaller(planning.ResourceKindPackage, runner, func(string) bool { lookups++; return true }, false)
			result := installer.Install(context.Background(), aptStepWithMetadata(tt.metadata))
			if result.Status != StepStatusFailed || !errors.Is(result.Err, tt.wantErr) {
				t.Fatalf("result = %#v, want failed with %v", result, tt.wantErr)
			}
			if lookups != 0 || len(runner.calls) != 0 {
				t.Fatalf("lookups = %d, calls = %d, want zero", lookups, len(runner.calls))
			}
		})
	}
}

func TestAptInstallerAvailabilityAndCommandFailuresAreStructured(t *testing.T) {
	tests := []struct {
		name       string
		sudo       bool
		exists     func(string) bool
		command    CommandResult
		wantStatus CommandStatus
		wantCalls  int
	}{
		{name: "missing apt", exists: func(string) bool { return false }, wantCalls: 0},
		{name: "missing sudo", sudo: true, exists: func(name string) bool { return name == "apt-get" }, wantCalls: 0},
		{name: "failed command", exists: func(string) bool { return true }, command: CommandResult{Status: CommandStatusFailed, Err: errors.New("exit 1")}, wantStatus: CommandStatusFailed, wantCalls: 1},
		{name: "timed out command", exists: func(string) bool { return true }, command: CommandResult{Status: CommandStatusTimedOut, Err: context.DeadlineExceeded}, wantStatus: CommandStatusTimedOut, wantCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{result: tt.command}
			installer := NewAptInstaller(planning.ResourceKindPackage, runner, tt.exists, tt.sudo)
			result := installer.Install(context.Background(), aptStep("ripgrep", "apt"))
			if result.Status != StepStatusFailed || len(runner.calls) != tt.wantCalls {
				t.Fatalf("result status = %q, calls = %d", result.Status, len(runner.calls))
			}
			if tt.wantStatus != "" && !errors.Is(result.Err, tt.command.Err) {
				t.Fatalf("Err = %v, want %v outcome preserved", result.Err, tt.wantStatus)
			}
		})
	}
}

func aptStep(pkg, provider string) planning.PlanStep {
	return aptStepWithMetadata(&planning.InstallMetadata{Provider: provider, Package: pkg})
}

func aptStepWithMetadata(metadata *planning.InstallMetadata) planning.PlanStep {
	ref := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}
	return planning.PlanStep{Ref: ref, Resource: planning.Resource{Ref: ref, Install: metadata}}
}
