package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestBrewOnlyInstallerDelegatesBrewMetadata(t *testing.T) {
	delegate := &fakeInstaller{
		kind: planning.ResourceKindPackage,
		results: []StepResult{{
			Ref:    planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"},
			Status: StepStatusInstalled,
		}},
	}
	installer := BrewOnlyInstaller(planning.ResourceKindPackage, delegate)
	step := brewOnlyStep(planning.ResourceKindPackage, "ripgrep", &planning.InstallMetadata{Provider: "brew", Package: "ripgrep"})

	result := installer.Install(context.Background(), step)

	if result.Status != StepStatusInstalled {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusInstalled)
	}
	if len(delegate.calls) != 1 {
		t.Fatalf("delegate calls = %d, want 1", len(delegate.calls))
	}
}

func TestBrewOnlyInstallerReturnsNotImplementedForNonBrewMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata *planning.InstallMetadata
	}{
		{name: "missing metadata", metadata: nil},
		{name: "unsupported provider", metadata: &planning.InstallMetadata{Provider: "apt", Package: "ripgrep"}},
		{name: "empty provider", metadata: &planning.InstallMetadata{Package: "ripgrep"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delegate := &fakeInstaller{kind: planning.ResourceKindPackage}
			installer := BrewOnlyInstaller(planning.ResourceKindPackage, delegate)

			result := installer.Install(context.Background(), brewOnlyStep(planning.ResourceKindPackage, "ripgrep", tt.metadata))

			if result.Status != StepStatusNotImplemented {
				t.Fatalf("Status = %q, want %q", result.Status, StepStatusNotImplemented)
			}
			if !errors.Is(result.Err, ErrNotImplemented) {
				t.Fatalf("Err = %v, want %v", result.Err, ErrNotImplemented)
			}
			if len(delegate.calls) != 0 {
				t.Fatalf("delegate calls = %d, want 0", len(delegate.calls))
			}
		})
	}
}

func TestBrewOnlyInstallerMissingDelegateReturnsNotImplemented(t *testing.T) {
	installer := BrewOnlyInstaller(planning.ResourceKindTool, nil)
	step := brewOnlyStep(planning.ResourceKindTool, "fd", &planning.InstallMetadata{Provider: "brew", Package: "fd"})

	result := installer.Install(context.Background(), step)

	if result.Status != StepStatusNotImplemented {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusNotImplemented)
	}
	if !errors.Is(result.Err, ErrNotImplemented) {
		t.Fatalf("Err = %v, want %v", result.Err, ErrNotImplemented)
	}
}

func TestBrewOnlyInstallerSupportedKind(t *testing.T) {
	installer := BrewOnlyInstaller(planning.ResourceKindTool, &fakeInstaller{kind: planning.ResourceKindTool})

	if got := installer.SupportedKind(); got != planning.ResourceKindTool {
		t.Fatalf("SupportedKind() = %q, want %q", got, planning.ResourceKindTool)
	}
}

func TestBrewOrAptInstallerRoutesOnlyMatchingProvider(t *testing.T) {
	brew := &fakeInstaller{kind: planning.ResourceKindPackage, results: []StepResult{{Status: StepStatusInstalled}}}
	apt := &fakeInstaller{kind: planning.ResourceKindPackage, results: []StepResult{{Status: StepStatusInstalled}}}
	installer := BrewOrAptInstaller(planning.ResourceKindPackage, brew, apt)

	for _, tt := range []struct {
		provider          string
		wantBrew, wantApt int
	}{{"brew", 1, 0}, {"apt", 1, 1}, {"asdf", 1, 1}} {
		result := installer.Install(context.Background(), brewOnlyStep(planning.ResourceKindPackage, "ripgrep", &planning.InstallMetadata{Provider: tt.provider, Package: "ripgrep"}))
		if tt.provider == "asdf" && result.Status != StepStatusNotImplemented {
			t.Fatalf("unsupported provider status = %q", result.Status)
		}
		if len(brew.calls) != tt.wantBrew || len(apt.calls) != tt.wantApt {
			t.Fatalf("provider %q: brew calls=%d apt calls=%d", tt.provider, len(brew.calls), len(apt.calls))
		}
	}
}

func brewOnlyStep(kind planning.ResourceKind, name string, metadata *planning.InstallMetadata) planning.PlanStep {
	ref := planning.ResourceRef{Kind: kind, Name: name}
	return planning.PlanStep{
		Ref: ref,
		Resource: planning.Resource{
			Ref:     ref,
			Install: metadata,
		},
	}
}
