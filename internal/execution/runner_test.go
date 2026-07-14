package execution

import (
	"context"
	"reflect"
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

func TestRunnerHonorsEligibleBrewFormulaPresence(t *testing.T) {
	installed := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"}
	absent := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}
	unknown := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "fd"}
	packageInstaller := &fakeInstaller{kind: planning.ResourceKindPackage}
	runner := NewRunner(packageInstaller)
	report := runner.Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{
		{Ref: installed, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "jq"}}, PackagePresence: planning.PackagePresenceInstalled},
		{Ref: absent, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "ripgrep"}}, PackagePresence: planning.PackagePresenceAbsent},
		{Ref: unknown, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "fd"}}, PackagePresence: planning.PackagePresenceUnknown},
	}})

	if got := report.Results[0]; got.Status != StepStatusSkipped || got.Message != "already installed; no mutation attempted" {
		t.Fatalf("installed result = %#v", got)
	}
	if got := report.Results[1]; got.Status != StepStatusInstalled {
		t.Fatalf("absent result = %#v, want existing installer dispatch", got)
	}
	if got := report.Results[2]; got.Status != StepStatusFailed || got.Message != "Homebrew formula presence could not be determined; no mutation attempted" || got.Err == nil {
		t.Fatalf("unknown result = %#v", got)
	}
	if got, want := len(packageInstaller.calls), 1; got != want || packageInstaller.calls[0].Ref != absent {
		t.Fatalf("installer calls = %#v, want absent step only", packageInstaller.calls)
	}
}

func TestRunnerIgnoresPackagePresenceForNonMatchingProvider(t *testing.T) {
	installer := &fakeInstaller{kind: planning.ResourceKindPackage}
	step := planning.PlanStep{
		Ref:             planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"},
		Resource:        planning.Resource{Install: &planning.InstallMetadata{Provider: "other", Package: "jq"}},
		PackagePresence: planning.PackagePresenceInstalled,
	}
	report := NewRunner(installer).Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{step}})
	if got := report.Results[0].Status; got != StepStatusInstalled {
		t.Fatalf("status = %q, want installer dispatch", got)
	}
	if got := len(installer.calls); got != 1 {
		t.Fatalf("installer calls = %d, want 1", got)
	}
}

func TestRunnerHonorsEligibleAptPackagePresence(t *testing.T) {
	installed := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "held-pkg"}
	absent := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "missing-pkg"}
	unknown := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ambiguous-pkg"}
	packageInstaller := &fakeInstaller{kind: planning.ResourceKindPackage}
	runner := NewRunner(packageInstaller)
	report := runner.Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{
		{Ref: installed, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "held-pkg"}}, PackagePresence: planning.PackagePresenceInstalled},
		{Ref: absent, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "missing-pkg"}}, PackagePresence: planning.PackagePresenceAbsent},
		{Ref: unknown, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "ambiguous-pkg"}}, PackagePresence: planning.PackagePresenceUnknown},
	}})

	if got := report.Results[0]; got.Status != StepStatusSkipped || got.Message != "already installed; no mutation attempted" {
		t.Fatalf("installed result = %#v", got)
	}
	if got := report.Results[1]; got.Status != StepStatusInstalled {
		t.Fatalf("absent result = %#v, want existing installer dispatch", got)
	}
	if got := report.Results[2]; got.Status != StepStatusFailed || got.Message != "APT package presence could not be determined; no mutation attempted" || got.Err == nil {
		t.Fatalf("unknown result = %#v", got)
	}
	if got, want := len(packageInstaller.calls), 1; got != want || packageInstaller.calls[0].Ref != absent {
		t.Fatalf("installer calls = %#v, want absent step only", packageInstaller.calls)
	}
}

func TestRunnerAptPartialStatesDispatch(t *testing.T) {
	packageInstaller := &fakeInstaller{
		kind: planning.ResourceKindPackage,
		results: []StepResult{
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "unpacked-pkg"}, Status: StepStatusInstalled},
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "half-configured-pkg"}, Status: StepStatusInstalled},
		},
	}
	runner := NewRunner(packageInstaller)
	report := runner.Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "unpacked-pkg"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "unpacked-pkg"}}, PackagePresence: planning.PackagePresenceAbsent},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "half-configured-pkg"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "half-configured-pkg"}}, PackagePresence: planning.PackagePresenceAbsent},
	}})

	if got, want := len(report.Results), 2; got != want {
		t.Fatalf("result count = %d, want %d", got, want)
	}
	for i, want := range []StepStatus{StepStatusInstalled, StepStatusInstalled} {
		if report.Results[i].Status != want {
			t.Fatalf("result[%d].Status = %q, want %q", i, report.Results[i].Status, want)
		}
	}
	if got, want := len(packageInstaller.calls), 2; got != want {
		t.Fatalf("installer calls = %d, want %d", got, want)
	}
}

func TestRunnerPreservesOrderWithAptPresence(t *testing.T) {
	a := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "a"}
	b := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "b"}
	c := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "c"}
	packageInstaller := &fakeInstaller{
		kind: planning.ResourceKindPackage,
		results: []StepResult{
			{Ref: c, Status: StepStatusInstalled},
		},
	}
	toolInstaller := &fakeInstaller{kind: planning.ResourceKindTool}
	runner := NewRunner(packageInstaller, toolInstaller)
	report := runner.Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{
		{Ref: a, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "a"}}, PackagePresence: planning.PackagePresenceInstalled},
		{Ref: b, Resource: planning.Resource{Ref: b}},
		{Ref: c, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "c"}}, PackagePresence: planning.PackagePresenceAbsent},
	}})

	wantOrder := []planning.ResourceRef{a, b, c}
	if got := len(report.Results); got != len(wantOrder) {
		t.Fatalf("result count = %d, want %d", got, len(wantOrder))
	}
	for i, want := range wantOrder {
		if report.Results[i].Ref != want {
			t.Fatalf("result[%d].Ref = %#v, want %#v", i, report.Results[i].Ref, want)
		}
	}
	if report.Results[0].Status != StepStatusSkipped {
		t.Fatalf("result[0].Status = %q, want skipped", report.Results[0].Status)
	}
	if report.Results[1].Status != StepStatusInstalled {
		t.Fatalf("result[1].Status = %q, want installed", report.Results[1].Status)
	}
	if report.Results[2].Status != StepStatusInstalled {
		t.Fatalf("result[2].Status = %q, want installed", report.Results[2].Status)
	}
	if len(packageInstaller.calls) != 1 || packageInstaller.calls[0].Ref != c {
		t.Fatalf("package installer calls = %#v, want c only", packageInstaller.calls)
	}
	if len(toolInstaller.calls) != 1 || toolInstaller.calls[0].Ref != b {
		t.Fatalf("tool installer calls = %#v, want b only", toolInstaller.calls)
	}
}

func TestRunnerIgnoresPackagePresenceForInvalidAptPackage(t *testing.T) {
	installer := &fakeInstaller{kind: planning.ResourceKindPackage}
	step := planning.PlanStep{
		Ref:             planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"},
		Resource:        planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: ""}},
		PackagePresence: planning.PackagePresenceInstalled,
	}
	report := NewRunner(installer).Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{step}})
	if got := report.Results[0].Status; got != StepStatusInstalled {
		t.Fatalf("status = %q, want installer dispatch", got)
	}
	if got := len(installer.calls); got != 1 {
		t.Fatalf("installer calls = %d, want 1", got)
	}
}

func TestRunnerIgnoresAptPresenceForBrewPackage(t *testing.T) {
	installer := &fakeInstaller{kind: planning.ResourceKindPackage}
	step := planning.PlanStep{
		Ref:             planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"},
		Resource:        planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "jq"}},
		PackagePresence: planning.PackagePresenceInstalled,
	}
	report := NewRunner(installer).Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{step}})
	if got := report.Results[0].Status; got != StepStatusSkipped {
		t.Fatalf("status = %q, want brew skip", got)
	}
	if got := len(installer.calls); got != 0 {
		t.Fatalf("installer calls = %d, want 0", got)
	}
}

func TestRunnerAptUncheckedPresenceDispatches(t *testing.T) {
	installer := &fakeInstaller{kind: planning.ResourceKindPackage}
	step := planning.PlanStep{
		Ref:             planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"},
		Resource:        planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "jq"}},
		PackagePresence: planning.PackagePresenceUnchecked,
	}
	report := NewRunner(installer).Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{step}})
	if got := report.Results[0].Status; got != StepStatusInstalled {
		t.Fatalf("status = %q, want installer dispatch", got)
	}
	if got := len(installer.calls); got != 1 {
		t.Fatalf("installer calls = %d, want 1", got)
	}
}

func TestRunnerAptIneligibleStepIgnoresPresence(t *testing.T) {
	toolInstaller := &fakeInstaller{kind: planning.ResourceKindTool}
	pkgInstaller := &fakeInstaller{kind: planning.ResourceKindPackage}
	report := NewRunner(toolInstaller, pkgInstaller).Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "jq"}, Resource: planning.Resource{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "jq"}}, PackagePresence: planning.PackagePresenceInstalled},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "jq"}}, PackagePresence: planning.PackagePresenceInstalled},
	}})
	if got, want := len(report.Results), 2; got != want {
		t.Fatalf("result count = %d, want %d", got, want)
	}
	if report.Results[0].Status != StepStatusInstalled {
		t.Fatalf("tool status = %q, want installed", report.Results[0].Status)
	}
	if report.Results[1].Status != StepStatusSkipped {
		t.Fatalf("apt status = %q, want skipped", report.Results[1].Status)
	}
	if got := len(toolInstaller.calls); got != 1 {
		t.Fatalf("tool installer calls = %d, want 1", got)
	}
	if got := len(pkgInstaller.calls); got != 0 {
		t.Fatalf("package installer calls = %d, want 0", got)
	}
}

func TestRunnerEmptyPlan(t *testing.T) {
	runner := NewRunner(&fakeInstaller{kind: planning.ResourceKindTool})
	report := runner.Run(context.Background(), planning.Plan{})

	if len(report.Results) != 0 {
		t.Fatalf("len(Results) = %d, want 0", len(report.Results))
	}
}

func TestRunnerPreservesAttentionReasonsForEveryOutcome(t *testing.T) {
	delegated := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "delegated"}
	installer := &fakeInstaller{kind: planning.ResourceKindTool, results: []StepResult{{
		Ref: delegated, Status: StepStatusInstalled, Message: "delegated install completed",
	}}}
	reasons := []string{"A", "B"}
	tests := []struct {
		name        string
		step        planning.PlanStep
		wantStatus  StepStatus
		wantMessage string
	}{
		{
			name:        "delegated installed",
			step:        planning.PlanStep{Ref: delegated, AttentionReasons: reasons},
			wantStatus:  StepStatusInstalled,
			wantMessage: "delegated install completed",
		},
		{
			name: "already installed skipped",
			step: planning.PlanStep{
				Ref:      planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "present"},
				Resource: planning.Resource{Presence: &planning.PresenceMetadata{Kind: "command_exists", Name: "present"}},
				Status:   planning.PlanStepStatusAlreadyInstalled, AttentionReasons: reasons,
			},
			wantStatus: StepStatusSkipped, wantMessage: "already installed; no mutation attempted",
		},
		{
			name: "unknown brew failed",
			step: planning.PlanStep{
				Ref:             planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "unknown"},
				Resource:        planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "unknown"}},
				PackagePresence: planning.PackagePresenceUnknown, AttentionReasons: reasons,
			},
			wantStatus: StepStatusFailed, wantMessage: "Homebrew formula presence could not be determined; no mutation attempted",
		},
		{
			name:       "unsupported",
			step:       planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "unsupported"}, AttentionReasons: reasons},
			wantStatus: StepStatusNotImplemented, wantMessage: "no installer registered for kind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := NewRunner(installer).Run(context.Background(), planning.Plan{Steps: []planning.PlanStep{tt.step}})
			got := report.Results[0]
			if got.Status != tt.wantStatus || got.Message != tt.wantMessage {
				t.Fatalf("result = %#v, want status=%q message=%q", got, tt.wantStatus, tt.wantMessage)
			}
			if !reflect.DeepEqual(got.AttentionReasons, reasons) {
				t.Fatalf("attention reasons = %#v, want %#v", got.AttentionReasons, reasons)
			}
		})
	}
}

func TestRunnerCopiesAttentionReasonsForEachResult(t *testing.T) {
	reasons := []string{"missing config", "runtime decoration", "missing config"}
	steps := []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "first"}, AttentionReasons: reasons},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "second"}, AttentionReasons: reasons},
	}
	report := NewRunner(&fakeInstaller{kind: planning.ResourceKindTool}).Run(context.Background(), planning.Plan{Steps: steps})

	report.Results[0].AttentionReasons[0] = "changed"
	report.Results[0].AttentionReasons = append(report.Results[0].AttentionReasons, "extra")
	if !reflect.DeepEqual(steps[0].AttentionReasons, reasons) || !reflect.DeepEqual(steps[1].AttentionReasons, reasons) {
		t.Fatalf("plan reasons mutated: %#v", steps)
	}
	if !reflect.DeepEqual(report.Results[1].AttentionReasons, reasons) {
		t.Fatalf("sibling reasons = %#v, want %#v", report.Results[1].AttentionReasons, reasons)
	}
}
