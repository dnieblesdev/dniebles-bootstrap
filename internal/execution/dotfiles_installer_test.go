package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestDotfilesInstallerMapsDotfileNameOnly(t *testing.T) {
	provider := &fakeDotfilesProvider{}
	installer := NewDotfilesInstaller(provider)
	step := planning.PlanStep{
		Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"},
		Resource: planning.Resource{
			Ref:         planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "other"},
			Description: "ignored",
			Install:     &planning.InstallMetadata{Provider: "shell", Package: "danger"},
		},
	}

	result := installer.Install(context.Background(), step)
	if result.Status != StepStatusInstalled {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusInstalled)
	}
	if len(provider.modules) != 1 || len(provider.modules[0]) != 1 || provider.modules[0][0] != "bash" {
		t.Fatalf("provider modules = %#v, want [[bash]]", provider.modules)
	}
}

func TestDotfilesInstallerRejectsNonDotfileWithoutProviderCall(t *testing.T) {
	provider := &fakeDotfilesProvider{}
	installer := NewDotfilesInstaller(provider)
	step := planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}}

	result := installer.Install(context.Background(), step)
	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
	}
	if len(provider.modules) != 0 {
		t.Fatalf("provider modules = %#v, want none", provider.modules)
	}
}

func TestDotfilesInstallerProviderErrorFailsStep(t *testing.T) {
	providerErr := errors.New("dotlink failed")
	installer := NewDotfilesInstaller(&fakeDotfilesProvider{err: providerErr})
	step := planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "nvim"}}

	result := installer.Install(context.Background(), step)
	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
	}
	if !errors.Is(result.Err, providerErr) {
		t.Fatalf("Err = %v, want provider error", result.Err)
	}
}

func TestDotfilesInstallerRequiresProvider(t *testing.T) {
	installer := NewDotfilesInstaller(nil)
	step := planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}}

	result := installer.Install(context.Background(), step)
	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
	}
}

type fakeDotfilesProvider struct {
	modules [][]string
	err     error
}

func (f *fakeDotfilesProvider) EnsureModules(context.Context, []string) error { return nil }
func (f *fakeDotfilesProvider) RunDotlink(_ context.Context, modules []string) error {
	copied := append([]string(nil), modules...)
	f.modules = append(f.modules, copied)
	return f.err
}
