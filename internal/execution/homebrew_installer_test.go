package execution

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

type fakeCommandRunner struct {
	result CommandResult
	calls  []CommandRequest
}

func (f *fakeCommandRunner) RunCommand(_ context.Context, req CommandRequest) CommandResult {
	f.calls = append(f.calls, req)
	f.result.Request = req
	return f.result
}

func TestHomebrewInstallerSupportedKind(t *testing.T) {
	installer := NewHomebrewInstaller(planning.ResourceKindPackage, &fakeCommandRunner{}, func(string) bool { return true })

	if got := installer.SupportedKind(); got != planning.ResourceKindPackage {
		t.Fatalf("SupportedKind() = %q, want %q", got, planning.ResourceKindPackage)
	}
}

func TestHomebrewInstallerSuccessBuildsExactCommand(t *testing.T) {
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded, ExitCode: 0}}
	lookups := []string{}
	installer := NewHomebrewInstaller(planning.ResourceKindPackage, runner, func(name string) bool {
		lookups = append(lookups, name)
		return true
	})
	step := brewStep(planning.ResourceKindPackage, "ripgrep", &planning.InstallMetadata{Provider: "brew", Package: "ripgrep"})

	result := installer.Install(context.Background(), step)

	if result.Status != StepStatusInstalled {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusInstalled)
	}
	if result.Err != nil {
		t.Fatalf("Err = %v, want nil", result.Err)
	}
	if !reflect.DeepEqual(lookups, []string{"brew"}) {
		t.Fatalf("lookups = %#v, want [brew]", lookups)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("runner calls = %d, want 1", len(runner.calls))
	}
	want := CommandRequest{Executable: "brew", Args: []string{"install", "ripgrep"}}
	if !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("CommandRequest = %#v, want %#v", runner.calls[0], want)
	}
}

func TestHomebrewInstallerCommandFailureIsStructured(t *testing.T) {
	commandErr := errors.New("exit status 42")
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusFailed, ExitCode: 42, Err: commandErr}}
	installer := NewHomebrewInstaller(planning.ResourceKindTool, runner, func(string) bool { return true })
	step := brewStep(planning.ResourceKindTool, "fd", &planning.InstallMetadata{Provider: "brew", Package: "fd"})

	result := installer.Install(context.Background(), step)

	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
	}
	if !errors.Is(result.Err, commandErr) {
		t.Fatalf("Err = %v, want command error", result.Err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("runner calls = %d, want 1", len(runner.calls))
	}
}

func TestHomebrewInstallerRejectsInvalidMetadataWithoutRunning(t *testing.T) {
	tests := []struct {
		name     string
		metadata *planning.InstallMetadata
		wantErr  error
	}{
		{name: "nil metadata", metadata: nil, wantErr: ErrUnsupportedInstallProvider},
		{name: "unsupported provider", metadata: &planning.InstallMetadata{Provider: "apt", Package: "ripgrep"}, wantErr: ErrUnsupportedInstallProvider},
		{name: "missing package", metadata: &planning.InstallMetadata{Provider: "brew", Package: ""}, wantErr: ErrMissingInstallPackage},
		{name: "blank package", metadata: &planning.InstallMetadata{Provider: "brew", Package: "   "}, wantErr: ErrMissingInstallPackage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded}}
			lookupCalled := false
			installer := NewHomebrewInstaller(planning.ResourceKindPackage, runner, func(string) bool {
				lookupCalled = true
				return true
			})

			result := installer.Install(context.Background(), brewStep(planning.ResourceKindPackage, "ripgrep", tt.metadata))

			if result.Status != StepStatusFailed {
				t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
			}
			if !errors.Is(result.Err, tt.wantErr) {
				t.Fatalf("Err = %v, want %v", result.Err, tt.wantErr)
			}
			if lookupCalled {
				t.Fatal("presence seam was called for invalid metadata")
			}
			if len(runner.calls) != 0 {
				t.Fatalf("runner calls = %d, want 0", len(runner.calls))
			}
		})
	}
}

func TestHomebrewInstallerMissingBrewDoesNotRunCommand(t *testing.T) {
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded}}
	lookups := []string{}
	installer := NewHomebrewInstaller(planning.ResourceKindTool, runner, func(name string) bool {
		lookups = append(lookups, name)
		return false
	})
	step := brewStep(planning.ResourceKindTool, "fd", &planning.InstallMetadata{Provider: "brew", Package: "fd"})

	result := installer.Install(context.Background(), step)

	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
	}
	if !errors.Is(result.Err, ErrMissingHomebrew) {
		t.Fatalf("Err = %v, want %v", result.Err, ErrMissingHomebrew)
	}
	if !reflect.DeepEqual(lookups, []string{"brew"}) {
		t.Fatalf("lookups = %#v, want [brew]", lookups)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("runner calls = %d, want 0", len(runner.calls))
	}
}

func TestHomebrewInstallerRequiresInjectedSeams(t *testing.T) {
	tests := []struct {
		name    string
		runner  CommandRunner
		exists  CommandExists
		wantErr error
	}{
		{name: "missing presence seam", runner: &fakeCommandRunner{}, exists: nil, wantErr: ErrMissingCommandExists},
		{name: "missing runner", runner: nil, exists: func(string) bool { return true }, wantErr: ErrMissingCommandRunner},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installer := NewHomebrewInstaller(planning.ResourceKindPackage, tt.runner, tt.exists)
			step := brewStep(planning.ResourceKindPackage, "ripgrep", &planning.InstallMetadata{Provider: "brew", Package: "ripgrep"})

			result := installer.Install(context.Background(), step)

			if result.Status != StepStatusFailed {
				t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
			}
			if !errors.Is(result.Err, tt.wantErr) {
				t.Fatalf("Err = %v, want %v", result.Err, tt.wantErr)
			}
		})
	}
}

func brewStep(kind planning.ResourceKind, name string, metadata *planning.InstallMetadata) planning.PlanStep {
	ref := planning.ResourceRef{Kind: kind, Name: name}
	return planning.PlanStep{
		Ref: ref,
		Resource: planning.Resource{
			Ref:     ref,
			Install: metadata,
		},
	}
}
