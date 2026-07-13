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

func TestAptPackageDetectorDetectsOnlyEligibleAptPackages(t *testing.T) {
	jq := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "json-tool"}
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: jq, Resource: planning.Resource{Ref: jq, Install: &planning.InstallMetadata{Provider: "apt", Package: "jq"}, Presence: &planning.PresenceMetadata{Name: "different-command"}}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "git"}}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "brew"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "jq"}}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "empty"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: " "}}},
	}}
	runner := &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "install ok installed"}}}

	got := AptPackageDetector{CommandExists: func(name string) bool { return name == "dpkg-query" }, Runner: runner}.Detect(context.Background(), plan)

	if want := map[planning.ResourceRef]planning.PackagePresence{jq: planning.PackagePresenceInstalled}; !reflect.DeepEqual(got, want) {
		t.Fatalf("presence = %#v, want %#v", got, want)
	}
	if got, want := runner.requests, []execution.CommandRequest{{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "jq"}, Timeout: aptPackagePresenceTimeout}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("requests = %#v, want %#v", got, want)
	}
}

func TestAptPackageDetectorClassifiesStatuses(t *testing.T) {
	ref := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"}
	plan := planning.Plan{Steps: []planning.PlanStep{{Ref: ref, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "jq"}}}}}
	tests := []struct {
		name   string
		exists bool
		runner execution.CommandRunner
		want   planning.PackagePresence
	}{
		{name: "install ok installed", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "install ok installed"}}}, want: planning.PackagePresenceInstalled},
		{name: "hold ok installed", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "hold ok installed"}}}, want: planning.PackagePresenceInstalled},
		{name: "install ok unpacked", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "install ok unpacked"}}}, want: planning.PackagePresenceAbsent},
		{name: "install ok half-configured", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "install ok half-configured"}}}, want: planning.PackagePresenceAbsent},
		{name: "deinstall ok config-files", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "deinstall ok config-files"}}}, want: planning.PackagePresenceAbsent},
		{name: "exact not found signature", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching jq"}}}, want: planning.PackagePresenceAbsent},
		{name: "not found with contradictory stdout", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusFailed, ExitCode: 1, Stdout: "install ok installed", Stderr: "dpkg-query: no packages found matching jq"}}}, want: planning.PackagePresenceUnknown},
		{name: "exit one wrong stderr", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "some other error"}}}, want: planning.PackagePresenceUnknown},
		{name: "empty stdout success", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: ""}}}, want: planning.PackagePresenceUnknown},
		{name: "malformed two fields", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "ok installed"}}}, want: planning.PackagePresenceUnknown},
		{name: "malformed four fields", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "install ok installed extra"}}}, want: planning.PackagePresenceUnknown},
		{name: "malformed invalid desired action", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "download ok installed"}}}, want: planning.PackagePresenceUnknown},
		{name: "malformed invalid error flag", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "install broken installed"}}}, want: planning.PackagePresenceUnknown},
		{name: "malformed invalid package status", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "install ok partial"}}}, want: planning.PackagePresenceUnknown},
		{name: "missing dpkg-query", want: planning.PackagePresenceUnknown},
		{name: "nil runner", exists: true, want: planning.PackagePresenceUnknown},
		{name: "timeout", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusTimedOut, Err: context.DeadlineExceeded}}}, want: planning.PackagePresenceUnknown},
		{name: "runner error", exists: true, runner: &recordingCommandRunner{results: []execution.CommandResult{{Status: execution.CommandStatusFailed, Err: errors.New("failed")}}}, want: planning.PackagePresenceUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AptPackageDetector{CommandExists: func(string) bool { return tt.exists }, Runner: tt.runner, Timeout: time.Second}.Detect(context.Background(), plan)
			if got[ref] != tt.want {
				t.Fatalf("presence = %q, want %q", got[ref], tt.want)
			}
		})
	}
}

func TestAptPackageDetectorProbesEachEligibleStepOnce(t *testing.T) {
	jq := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"}
	curl := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "curl"}
	plan := planning.Plan{Steps: []planning.PlanStep{
		{Ref: jq, Resource: planning.Resource{Ref: jq, Install: &planning.InstallMetadata{Provider: "apt", Package: "jq"}}},
		{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "vim"}, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "brew", Package: "vim"}}},
		{Ref: curl, Resource: planning.Resource{Ref: curl, Install: &planning.InstallMetadata{Provider: "apt", Package: "curl"}}},
	}}
	runner := &recordingCommandRunner{results: []execution.CommandResult{
		{Status: execution.CommandStatusSucceeded, ExitCode: 0, Stdout: "install ok installed"},
		{Status: execution.CommandStatusFailed, ExitCode: 1, Stderr: "dpkg-query: no packages found matching curl"},
	}}

	got := AptPackageDetector{CommandExists: func(name string) bool { return name == "dpkg-query" }, Runner: runner}.Detect(context.Background(), plan)

	want := map[planning.ResourceRef]planning.PackagePresence{
		jq:   planning.PackagePresenceInstalled,
		curl: planning.PackagePresenceAbsent,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("presence = %#v, want %#v", got, want)
	}
	wantRequests := []execution.CommandRequest{
		{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "jq"}, Timeout: aptPackagePresenceTimeout},
		{Executable: "dpkg-query", Args: []string{"--show", "--showformat=${Status}", "curl"}, Timeout: aptPackagePresenceTimeout},
	}
	if !reflect.DeepEqual(runner.requests, wantRequests) {
		t.Fatalf("requests = %#v, want %#v", runner.requests, wantRequests)
	}
}

func TestApplyAptPackagePresenceCopiesPlanAndAddsUnknownAttention(t *testing.T) {
	ref := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "jq"}
	plan := planning.Plan{Steps: []planning.PlanStep{{Ref: ref, Resource: planning.Resource{Install: &planning.InstallMetadata{Provider: "apt", Package: "jq"}}}}}

	decorated := ApplyAptPackagePresence(plan, map[planning.ResourceRef]planning.PackagePresence{ref: planning.PackagePresenceUnknown})

	if plan.Steps[0].PackagePresence != planning.PackagePresenceUnchecked || len(plan.Steps[0].AttentionReasons) != 0 {
		t.Fatalf("original plan mutated: %#v", plan.Steps[0])
	}
	if got := decorated.Steps[0].PackagePresence; got != planning.PackagePresenceUnknown {
		t.Fatalf("presence = %q, want unknown", got)
	}
	if got := decorated.Steps[0].AttentionReasons; !reflect.DeepEqual(got, []string{"APT package presence could not be determined; no mutation attempted"}) {
		t.Fatalf("attention = %#v", got)
	}
}
