package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestNoopInstallerReturnsNotImplemented(t *testing.T) {
	inst := NoopInstaller{}
	step := planning.PlanStep{
		Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"},
	}

	if got := inst.SupportedKind(); got != "" {
		t.Fatalf("SupportedKind = %q, want empty", got)
	}

	result := inst.Install(context.Background(), step)
	if result.Status != StepStatusNotImplemented {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusNotImplemented)
	}
	if result.Ref != step.Ref {
		t.Fatalf("Ref = %#v, want %#v", result.Ref, step.Ref)
	}
	if result.Message == "" {
		t.Fatal("expected a non-empty message describing noop behavior")
	}
}

func TestNoopForKindReturnsNotImplementedForSupportedKind(t *testing.T) {
	tests := []struct {
		kind planning.ResourceKind
	}{
		{planning.ResourceKindTool},
		{planning.ResourceKindRuntime},
		{planning.ResourceKindPackage},
		{planning.ResourceKindDotfile},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			inst := NoopForKind(tt.kind)
			if got := inst.SupportedKind(); got != tt.kind {
				t.Fatalf("SupportedKind = %q, want %q", got, tt.kind)
			}

			step := planning.PlanStep{
				Ref: planning.ResourceRef{Kind: tt.kind, Name: "example"},
			}
			result := inst.Install(context.Background(), step)
			if result.Status != StepStatusNotImplemented {
				t.Fatalf("Status = %q, want %q", result.Status, StepStatusNotImplemented)
			}
			if result.Ref != step.Ref {
				t.Fatalf("Ref = %#v, want %#v", result.Ref, step.Ref)
			}
			if result.Message == "" {
				t.Fatal("expected a non-empty message describing noop behavior")
			}
		})
	}
}

func TestNoopDotfilesProviderReturnsNotImplemented(t *testing.T) {
	provider := NoopDotfilesProvider{}
	ctx := context.Background()
	modules := []string{"shell", "vim"}

	if err := provider.EnsureModules(ctx, modules); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("EnsureModules error = %v, want ErrNotImplemented", err)
	}
	if err := provider.RunDotlink(ctx, modules); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("RunDotlink error = %v, want ErrNotImplemented", err)
	}
}
