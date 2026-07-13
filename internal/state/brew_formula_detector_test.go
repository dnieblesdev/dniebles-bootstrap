package state

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/execution"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

type recordingCommandRunner struct {
	results  []execution.CommandResult
	requests []execution.CommandRequest
}

func (r *recordingCommandRunner) RunCommand(_ context.Context, request execution.CommandRequest) execution.CommandResult {
	r.requests = append(r.requests, request)
	if len(r.results) == 0 {
		return execution.CommandResult{Status: execution.CommandStatusNotRun}
	}
	result := r.results[0]
	r.results = r.results[1:]
	return result
}

func TestBrewFormulaDetectorDetectsOnlyEligibleFormulaPackages(t *testing.T) {
	jq := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "json-tool"}
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: jq, Resource: planning.Resource{Ref: jq, Install: &planning.InstallMetadata{Provider: "brew", Package: "jq"}, Presence: &planning.PresenceMetadata{Name: "different-command"}}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "git"}}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "apt"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "jq"}}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "empty"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: " "}}},
	}}
	runner := &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0}}}

	got := BrewFormulaDetector{CommandExists: func(name string) bool { return name == "brew" }, Runner: runner}.Detect(context.Background(), plan)

	if want := map[planning.ResourceRef]planning.PackagePresence{jq: planning.PackagePresenceInstalled}; !reflect.DeepEqual(got, want) {
		t.Fatalf("presence = %#v, want %#v", got, want)
	}
	if got, want := runner.requests, []execution.CommandRequest{{Executable: "brew", Args: []string{"list", "--formula", "jq"}, Timeout: brewFormulaPresenceTimeout}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("requests = %#v, want %#v", got, want)
	}
}

func TestBrewFormulaDetectorClassifiesFailuresAsUnknownAndExplicitAbsence(t *testing.T) {
	ref := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"}
	plan := planning.Plan{Steps: []planning.PlanStep{{Ref: ref, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "jq"}}}}}
	tests := []struct {
		name   string
		exists bool
		runner execution.CommandRunner
		want   planning.PackagePresence
	}{
		{name: "missing brew", want: planning.PackagePresenceUnknown},
		{name: "nil runner", exists: true, want: planning.PackagePresenceUnknown},
		{name: "timeout", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusTimedOut, Err: context.DeadlineExceeded}}}, want: planning.PackagePresenceUnknown},
		{name: "runner error", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusFailed, Err: errors.New("failed")}}}, want: planning.PackagePresenceUnknown},
		{name: "bare exit one", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusFailed, ExitCode: 1}}}, want: planning.PackagePresenceUnknown},
		{name: "explicit formula absent", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "Error: No such keg: /usr/local/Cellar/jq", Err: errors.New("exit 1")}}}, want: planning.PackagePresenceAbsent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BrewFormulaDetector{CommandExists: func(string) bool { return tt.exists }, Runner: tt.runner, Timeout: time.Second}.Detect(context.Background(), plan)
			if got[ref] != tt.want {
				t.Fatalf("presence = %q, want %q", got[ref], tt.want)
			}
		})
	}
}

func TestApplyBrewFormulaPresenceCopiesPlanAndAddsUnknownAttention(t *testing.T) {
	ref := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"}
	plan := planning.Plan{Steps: []planning.PlanStep{{Ref: ref, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "jq"}}}}}

	decorated := ApplyBrewFormulaPresence(plan, map[planning.ResourceRef]planning.PackagePresence{ref: planning.PackagePresenceUnknown})

	if plan.Steps[0].PackagePresence != planning.PackagePresenceUnchecked || len(plan.Steps[0].AttentionReasons) != 0 {
		t.Fatalf("original plan mutated: %#v", plan.Steps[0])
	}
	if got := decorated.Steps[0].PackagePresence; got != planning.PackagePresenceUnknown {
		t.Fatalf("presence = %q, want unknown", got)
	}
	if got := decorated.Steps[0].AttentionReasons; !reflect.DeepEqual(got, []string{"Homebrew formula presence could not be determined; no mutation attempted"}) {
		t.Fatalf("attention = %#v", got)
	}
}
