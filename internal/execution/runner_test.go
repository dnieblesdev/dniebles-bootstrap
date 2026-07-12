package execution

import (
	"context"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

// fakeInstaller records every Install call and returns a configured result.
type fakeInstaller struct {
	kind    planning.ResourceKind
	results []StepResult
	calls   []planning.PlanStep
	callIdx int
}

func (f *fakeInstaller) SupportedKind() planning.ResourceKind { return f.kind }

func (f *fakeInstaller) Install(_ context.Context, step planning.PlanStep) StepResult {
	f.calls = append(f.calls, step)
	idx := f.callIdx
	f.callIdx++
	if idx < len(f.results) {
		return f.results[idx]
	}
	return StepResult{Ref: step.Ref, Status: StepStatusInstalled}
}

func TestRunnerDispatchesSequentiallyByKind(t *testing.T) {
	toolInst := &fakeInstaller{
		kind: planning.ResourceKindTool,
		results: []StepResult{
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}, Status: StepStatusInstalled},
		},
	}
	pkgInst := &fakeInstaller{
		kind: planning.ResourceKindPackage,
		results: []StepResult{
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}, Status: StepStatusSkipped},
		},
	}
	runtimeInst := &fakeInstaller{
		kind: planning.ResourceKindRuntime,
		results: []StepResult{
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"}, Status: StepStatusNotImplemented},
		},
	}

	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"}},
	}}

	runner := NewRunner(toolInst, pkgInst, runtimeInst)
	report := runner.Run(context.Background(), plan)

	if len(report.Results) != 3 {
		t.Fatalf("len(Results) = %d, want 3", len(report.Results))
	}

	wantOrder := []planning.ResourceRef{
		{Kind: planning.ResourceKindTool, Name: "git"},
		{Kind: planning.ResourceKindPackage, Name: "ripgrep"},
		{Kind: planning.ResourceKindRuntime, Name: "go"},
	}
	for i, want := range wantOrder {
		if report.Results[i].Ref != want {
			t.Fatalf("result[%d].Ref = %#v, want %#v", i, report.Results[i].Ref, want)
		}
	}

	if len(toolInst.calls) != 1 || toolInst.calls[0].Ref != wantOrder[0] {
		t.Fatalf("tool installer calls = %#v, want one call for git", toolInst.calls)
	}
	if len(pkgInst.calls) != 1 || pkgInst.calls[0].Ref != wantOrder[1] {
		t.Fatalf("package installer calls = %#v, want one call for ripgrep", pkgInst.calls)
	}
	if len(runtimeInst.calls) != 1 || runtimeInst.calls[0].Ref != wantOrder[2] {
		t.Fatalf("runtime installer calls = %#v, want one call for go", runtimeInst.calls)
	}

	wantStatuses := []StepStatus{StepStatusInstalled, StepStatusSkipped, StepStatusNotImplemented}
	for i, want := range wantStatuses {
		if report.Results[i].Status != want {
			t.Fatalf("result[%d].Status = %q, want %q", i, report.Results[i].Status, want)
		}
	}
}

func TestRunnerSkipsOnlyEligibleAlreadyInstalledSteps(t *testing.T) {
	tool := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "vim"}
	runtime := planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"}
	packageRef := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "vim"}
	installer := &fakeInstaller{kind: planning.ResourceKindTool}
	runtimeInstaller := &fakeInstaller{kind: planning.ResourceKindRuntime}
	packageInstaller := &fakeInstaller{kind: planning.ResourceKindPackage}
	runner := NewRunner(installer, runtimeInstaller, packageInstaller)
	validPresence := &planning.PresenceMetadata{Kind: "command_exists", Name: "vim"}
	report := runner.Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{
		{Ref: tool, Resource: planning.Resource{Ref: tool, Presence: validPresence}, Status: planning.PlanStepStatusAlreadyInstalled},
		{Ref: runtime, Resource: planning.Resource{Ref: runtime, Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "go"}}, Status: planning.PlanStepStatusAlreadyInstalled},
		{Ref: tool, Resource: planning.Resource{Ref: tool}, Status: planning.PlanStepStatusAlreadyInstalled},
		{Ref: packageRef, Resource: planning.Resource{Ref: packageRef, Presence: validPresence}, Status: planning.PlanStepStatusAlreadyInstalled},
	}})
	if got, want := len(report.Results), 4; got != want {
		t.Fatalf("result count = %d, want %d", got, want)
	}
	for _, index := range []int{0, 1} {
		if got := report.Results[index]; got.Status != StepStatusSkipped || got.Message != "already installed; no mutation attempted" {
			t.Fatalf("result[%d] = %#v, want idempotent skip", index, got)
		}
	}
	if got, want := len(installer.calls), 1; got != want {
		t.Fatalf("tool installer calls = %d, want %d for invalid metadata", got, want)
	}
	if got, want := len(runtimeInstaller.calls), 0; got != want {
		t.Fatalf("runtime installer calls = %d, want %d", got, want)
	}
	if got, want := len(packageInstaller.calls), 1; got != want {
		t.Fatalf("package installer calls = %d, want %d", got, want)
	}
}

func TestRunnerContinuesOnMissingInstaller(t *testing.T) {
	toolInst := &fakeInstaller{
		kind: planning.ResourceKindTool,
		results: []StepResult{
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}, Status: StepStatusInstalled},
		},
	}

	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "shell"}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "curl"}},
	}}

	runner := NewRunner(toolInst)
	report := runner.Run(context.Background(), plan)

	if len(report.Results) != 3 {
		t.Fatalf("len(Results) = %d, want 3", len(report.Results))
	}

	if report.Results[0].Status != StepStatusInstalled {
		t.Fatalf("result[0].Status = %q, want installed", report.Results[0].Status)
	}
	if report.Results[1].Status != StepStatusNotImplemented {
		t.Fatalf("result[1].Status = %q, want not_implemented", report.Results[1].Status)
	}
	if report.Results[2].Status != StepStatusInstalled {
		t.Fatalf("result[2].Status = %q, want installed", report.Results[2].Status)
	}

	if len(toolInst.calls) != 2 {
		t.Fatalf("tool installer calls = %d, want 2", len(toolInst.calls))
	}
}

func TestRunnerEmptyPlan(t *testing.T) {
	runner := NewRunner(&fakeInstaller{kind: planning.ResourceKindTool})
	report := runner.Run(context.Background(), planning.Plan{})

	if len(report.Results) != 0 {
		t.Fatalf("len(Results) = %d, want 0", len(report.Results))
	}
}
