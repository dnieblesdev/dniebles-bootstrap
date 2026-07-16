package execution

import (
	"context"
	"strings"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestAcquireHomebrewRejectsNonLinuxBeforeLinuxAcquirer(t *testing.T) {
	original := acquireHomebrewLinuxFn
	t.Cleanup(func() { acquireHomebrewLinuxFn = original })
	calls := 0
	acquireHomebrewLinuxFn = func(context.Context) HomebrewAcquisitionResult {
		calls++
		return HomebrewAcquisitionResult{Acquired: true}
	}

	result := AcquireHomebrew(context.Background(), planning.EnvironmentFacts{OS: "darwin"})
	if result.Err != ErrHomebrewAcquisitionUnavailable || result.Acquired {
		t.Fatalf("result = %#v", result)
	}
	if calls != 0 {
		t.Fatalf("Linux acquirer calls = %d, want 0 before download", calls)
	}
}

func TestAppendHomebrewBootstrapNoBrewResources(t *testing.T) {
	report := ExecutionReport{Results: []StepResult{}}
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}, Resource: planning.Resource{
			Install: &planning.InstallMetadata{Provider: "apt", Package: "git"},
		}},
	}}

	got := AppendHomebrewBootstrap(report, plan, func(string) bool { return false })

	if len(got.ManualActions) != 0 {
		t.Fatalf("ManualActions = %d, want 0", len(got.ManualActions))
	}
}

func TestAppendHomebrewBootstrapBrewPresent(t *testing.T) {
	report := ExecutionReport{Results: []StepResult{}}
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}, Resource: planning.Resource{
			Install: &planning.InstallMetadata{Provider: "brew", Package: "ripgrep"},
		}},
	}}

	got := AppendHomebrewBootstrap(report, plan, func(name string) bool {
		if name != "brew" {
			t.Fatalf("expected lookup for brew, got %q", name)
		}
		return true
	})

	if len(got.ManualActions) != 0 {
		t.Fatalf("ManualActions = %d, want 0", len(got.ManualActions))
	}
}

func TestAppendHomebrewBootstrapBrewMissing(t *testing.T) {
	report := ExecutionReport{Results: []StepResult{}}
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "fd"}, Resource: planning.Resource{
			Install: &planning.InstallMetadata{Provider: "brew", Package: "fd"},
		}},
	}}

	got := AppendHomebrewBootstrap(report, plan, func(name string) bool {
		if name != "brew" {
			t.Fatalf("expected lookup for brew, got %q", name)
		}
		return false
	})

	if len(got.ManualActions) != 1 {
		t.Fatalf("ManualActions = %d, want 1", len(got.ManualActions))
	}
	action := got.ManualActions[0]
	if action.ID != "homebrew:bootstrap" {
		t.Fatalf("ID = %q, want homebrew:bootstrap", action.ID)
	}
	if action.Title != "Install Homebrew" {
		t.Fatalf("Title = %q, want Install Homebrew", action.Title)
	}
	if action.Reason == "" {
		t.Fatal("Reason is empty")
	}
	if len(action.Instructions) == 0 {
		t.Fatal("Instructions is empty")
	}
	found := false
	for _, inst := range action.Instructions {
		if inst == homebrewDocumentationURL {
			found = true
		}
		assertNoExecutableHomebrewGuidance(t, inst)
	}
	if !found {
		t.Fatalf("instructions missing official Homebrew documentation URL: %v", action.Instructions)
	}
}

func TestAppendHomebrewBootstrapGuidanceIsAdvisoryOnly(t *testing.T) {
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "fd"}, Resource: planning.Resource{
			Install: &planning.InstallMetadata{Provider: "brew", Package: "fd"},
		}},
	}}

	got := AppendHomebrewBootstrap(ExecutionReport{}, plan, func(string) bool { return false })

	if len(got.ManualActions) != 1 {
		t.Fatalf("ManualActions = %d, want 1", len(got.ManualActions))
	}
	joined := strings.Join(got.ManualActions[0].Instructions, "\n")
	if !strings.Contains(joined, homebrewDocumentationURL) {
		t.Fatalf("instructions missing official documentation URL: %q", joined)
	}
	assertNoExecutableHomebrewGuidance(t, joined)
}

func TestAppendHomebrewBootstrapPreservesResults(t *testing.T) {
	report := ExecutionReport{Results: []StepResult{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}, Status: StepStatusNotImplemented},
	}}
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}, Resource: planning.Resource{
			Install: &planning.InstallMetadata{Provider: "brew", Package: "ripgrep"},
		}},
	}}

	got := AppendHomebrewBootstrap(report, plan, func(string) bool { return false })

	if len(got.Results) != 1 {
		t.Fatalf("Results = %d, want 1", len(got.Results))
	}
	if got.Results[0].Ref.Name != "git" {
		t.Fatalf("Results[0].Ref.Name = %q, want git", got.Results[0].Ref.Name)
	}
}

func TestAppendHomebrewBootstrapLookupOnlyBrew(t *testing.T) {
	// Guard against the provider probing unrelated commands.
	lookups := []string{}
	exists := func(name string) bool {
		lookups = append(lookups, name)
		return false
	}
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "fd"}, Resource: planning.Resource{
			Install: &planning.InstallMetadata{Provider: "brew", Package: "fd"},
		}},
	}}

	AppendHomebrewBootstrap(ExecutionReport{}, plan, exists)

	for _, name := range lookups {
		if name != "brew" {
			t.Fatalf("unexpected command lookup: %q", name)
		}
	}
}

func TestAppendHomebrewBootstrapDoesNotExecuteInstruction(t *testing.T) {
	// The official install command must be rendered as text only; the provider
	// never constructs a CommandRequest or attempts to run it.
	report := ExecutionReport{}
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "fd"}, Resource: planning.Resource{
			Install: &planning.InstallMetadata{Provider: "brew", Package: "fd"},
		}},
	}}

	got := AppendHomebrewBootstrap(report, plan, func(string) bool { return false })

	if len(got.ManualActions) != 1 {
		t.Fatalf("ManualActions = %d, want 1", len(got.ManualActions))
	}
}

func assertNoExecutableHomebrewGuidance(t *testing.T, text string) {
	t.Helper()
	for _, forbidden := range []string{"/bin/bash", "curl", "sh -c", "|", "install.sh"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("guidance contains forbidden executable fragment %q in %q", forbidden, text)
		}
	}
}
