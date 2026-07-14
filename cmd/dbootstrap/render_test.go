package main

import (
	"bytes"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/execution"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestRenderPlanResultIncludesSkippedAttentionAndDiagnostics(t *testing.T) {
	toolGit := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}
	runtimeGo := planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"}
	packageLinux := planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "linux-only"}

	result := planning.PlanResult{
		Plan: planning.Plan{Steps: []planning.PlanStep{
			{
				Ref:       toolGit,
				Resource:  planning.Resource{Ref: toolGit, Description: "Version control"},
				DependsOn: nil,
			},
			{
				Ref:              runtimeGo,
				Resource:         planning.Resource{Ref: runtimeGo, Description: "Go toolchain"},
				DependsOn:        []planning.ResourceRef{toolGit},
				AttentionReasons: []string{"missing required config \"go.env\""},
			},
		}},
		Results: []planning.PlanStepResult{
			{Ref: packageLinux, Status: planning.PlanStepStatusSkipped, Reasons: []string{"environment facts do not match resource conditions"}},
			{Ref: runtimeGo, Status: planning.PlanStepStatusAttentionRequired, Reasons: []string{"missing required config \"go.env\""}},
			{Ref: toolGit, Status: planning.PlanStepStatusPlanned},
			{Status: planning.PlanStepStatusError, Reasons: []string{"unknown bundle \"missing\""}},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	renderPlanResult(&stdout, "dev", nil, "catalog/bootstrap.toml", planning.EnvironmentFacts{OS: "linux", Arch: "amd64"}, result)
	renderDiagnostics(&stderr, result)

	wantStdout := "Plan profile: dev\n" +
		"Catalog: catalog/bootstrap.toml\n" +
		"Environment: os=linux arch=amd64 distro= wsl=false\n" +
		"\n" +
		"Steps:\n" +
		"1. tool:git [planned] Version control\n" +
		"   depends_on: none\n" +
		"   attention: none\n" +
		"2. runtime:go [attention_required] Go toolchain\n" +
		"   depends_on: tool:git\n" +
		"   attention: missing required config \"go.env\"\n" +
		"\n" +
		"Results:\n" +
		"- package:linux-only: skipped\n" +
		"  reason: environment facts do not match resource conditions\n" +
		"- runtime:go: attention_required\n" +
		"  reason: missing required config \"go.env\"\n" +
		"- tool:git: planned\n" +
		"- diagnostic: error\n" +
		"  reason: unknown bundle \"missing\"\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("stdout = %q, want %q", got, wantStdout)
	}

	wantStderr := "Diagnostics:\n- diagnostic: unknown bundle \"missing\"\n"
	if got := stderr.String(); got != wantStderr {
		t.Fatalf("stderr = %q, want %q", got, wantStderr)
	}
}

func TestRenderPlanResultResourceOnlyHeader(t *testing.T) {
	toolGit := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}

	result := planning.PlanResult{
		Plan: planning.Plan{Steps: []planning.PlanStep{
			{
				Ref:      toolGit,
				Resource: planning.Resource{Ref: toolGit, Description: "Version control"},
			},
		}},
		Results: []planning.PlanStepResult{
			{Ref: toolGit, Status: planning.PlanStepStatusPlanned},
		},
	}

	var stdout bytes.Buffer
	renderPlanResult(&stdout, "", []planning.ResourceRef{toolGit}, "catalog/bootstrap.toml", planning.EnvironmentFacts{OS: "linux", Arch: "amd64"}, result)

	wantStdout := "Plan resources: tool:git\n" +
		"Catalog: catalog/bootstrap.toml\n" +
		"Environment: os=linux arch=amd64 distro= wsl=false\n" +
		"\n" +
		"Steps:\n" +
		"1. tool:git [planned] Version control\n" +
		"   depends_on: none\n" +
		"   attention: none\n" +
		"\n" +
		"Results:\n" +
		"- tool:git: planned\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("stdout = %q, want %q", got, wantStdout)
	}
}

func TestRenderExecutionReportRendersSummaryAndUserFacingStatuses(t *testing.T) {
	report := execution.ExecutionReport{
		Results: []execution.StepResult{
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "fd"}, Status: execution.StepStatusInstalled, Message: "installed fd with Homebrew"},
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}, Status: execution.StepStatusSkipped, Message: "skipped because Homebrew must be installed manually"},
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindRuntime, Name: "go"}, Status: execution.StepStatusNotImplemented, Message: "noop installer does not perform real installation"},
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "broken"}, Status: execution.StepStatusFailed, Message: "brew install broken failed"},
		},
	}

	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeDryRun, report)

	wantStdout := "Execution Report\n" +
		"Mode: dry-run\n" +
		"\n" +
		"Summary:\n" +
		"- changed: 1\n" +
		"- unchanged: 1\n" +
		"- not supported yet: 1\n" +
		"- failed: 1\n" +
		"\n" +
		"Steps:\n" +
		"1. tool:fd [changed] installed fd with Homebrew\n" +
		"2. package:ripgrep [unchanged] skipped because Homebrew must be installed manually\n" +
		"3. runtime:go [not supported yet] noop installer does not perform real installation\n" +
		"4. tool:broken [failed] brew install broken failed\n" +
		"\n" +
		"Manual Actions:\n" +
		"- none\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("stdout = %q, want %q", got, wantStdout)
	}
}

func TestRenderExecutionReportRendersManualActions(t *testing.T) {
	report := execution.ExecutionReport{
		Results: []execution.StepResult{
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "fd"}, Status: execution.StepStatusNotImplemented, Message: "noop installer does not perform real installation"},
		},
		ManualActions: []execution.ManualAction{
			{
				ID:     "homebrew:bootstrap",
				Title:  "Install Homebrew",
				Reason: "Homebrew is required by selected resources but is not installed on this host.",
				Instructions: []string{
					"Review the official Homebrew installation documentation before making host changes:",
					"https://brew.sh/",
					"Install Homebrew manually only after you understand the documented steps, then re-run dbootstrap apply.",
				},
			},
		},
	}

	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeDryRun, report)

	wantStdout := "Execution Report\n" +
		"Mode: dry-run\n" +
		"\n" +
		"Summary:\n" +
		"- changed: 0\n" +
		"- unchanged: 0\n" +
		"- not supported yet: 1\n" +
		"- failed: 0\n" +
		"\n" +
		"Steps:\n" +
		"1. tool:fd [not supported yet] noop installer does not perform real installation\n" +
		"\n" +
		"Manual Actions:\n" +
		"- homebrew:bootstrap: Install Homebrew\n" +
		"  reason: Homebrew is required by selected resources but is not installed on this host.\n" +
		"  instruction: Review the official Homebrew installation documentation before making host changes:\n" +
		"  instruction: https://brew.sh/\n" +
		"  instruction: Install Homebrew manually only after you understand the documented steps, then re-run dbootstrap apply.\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("stdout = %q, want %q", got, wantStdout)
	}
}

func TestRenderExecutionReportFramesConfirmedModeMutability(t *testing.T) {
	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeConfirmed, execution.ExecutionReport{})

	got := stdout.String()
	for _, want := range []string{
		"brew-backed tool/package steps, eligible Linux APT-backed tool/package steps, and selected dotfile resources may have changed this machine",
		"unsupported, non-provider-backed, and unselected steps remain non-mutating or not supported yet",
	} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Fatalf("stdout missing %q: %q", want, got)
		}
	}
}

func TestRenderExecutionReportRendersDotlinkDetailsAndBaseDiagnostic(t *testing.T) {
	report := execution.ExecutionReport{Results: []execution.StepResult{{
		Ref:     planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"},
		Status:  execution.StepStatusFailed,
		Message: "dotfile module bash failed",
		LinkDetails: []execution.LinkDetail{
			{Module: "bash", Source: "bashrc", Target: "/home/ada/.bashrc", Outcome: execution.LinkOutcomeChanged},
			{Module: "bash", Source: "profile", Target: "/home/ada/.profile", Outcome: execution.LinkOutcomeUnchanged},
			{Module: "bash", Source: "profile", Target: "/home/ada/.profile", Outcome: execution.LinkOutcomeRolledBack, Cause: &execution.LinkCause{Code: "conflict", Message: "target exists"}},
		},
		Failure:        &execution.LinkFailure{Module: "bash", Cause: execution.LinkCause{Code: "conflict", Message: "target exists"}},
		Rollback:       execution.LinkRollback{Attempted: true, Completed: true, Removed: []string{"/home/ada/.bashrc"}},
		BaseDiagnostic: &execution.DotfilesBaseDiagnostic{Source: execution.DotfilesBaseSourceEnv, AttemptedCandidate: "/tmp/dotfiles", Modules: []string{"bash"}, Cause: "resolve dotfiles base: unavailable"},
	}}}

	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeConfirmed, report)
	got := stdout.String()
	for _, want := range []string{
		"dotfile:bash [failed] dotfile module bash failed",
		"link: changed source=bashrc target=/home/ada/.bashrc",
		"link: unchanged source=profile target=/home/ada/.profile",
		"link: rolled_back source=profile target=/home/ada/.profile",
		"cause: conflict: target exists",
		"aggregate failure: module=bash cause=conflict: target exists",
		"rollback: attempted=true completed=true",
		"rollback removed: /home/ada/.bashrc",
		"dotfiles base: source=env attempted candidate=/tmp/dotfiles modules=bash cause=resolve dotfiles base: unavailable",
	} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Fatalf("stdout missing %q: %q", want, got)
		}
	}
	if bytes.Contains([]byte(got), []byte("canonical base=/tmp/dotfiles")) {
		t.Fatalf("unresolved base rendered as canonical: %q", got)
	}
}

func TestRenderLinkDetailsRendersExecutionFactsAndDeduplicatesBase(t *testing.T) {
	base := &execution.DotfilesBaseDiagnostic{Source: execution.DotfilesBaseSourceEnv, CanonicalPath: "/repo", Modules: []string{"bash"}}
	result := execution.StepResult{BaseDiagnostic: base, DotfilesFailure: &execution.DotfilesFailure{Executable: "/repo/bin/dotlink", Runner: "CommandRunner", Command: execution.CommandRequest{Args: []string{"link", "bash"}}, Stderr: `bad\x1b[31m`, BaseSnapshot: &execution.DotfilesBaseDiagnostic{Source: execution.DotfilesBaseSourceEnv, CanonicalPath: "/repo", Modules: []string{"bash"}}}}
	var output bytes.Buffer
	renderLinkDetails(&output, result)
	got := output.String()
	if bytes.Count([]byte(got), []byte("canonical base=/repo")) != 1 || !bytes.Contains([]byte(got), []byte("executable: /repo/bin/dotlink")) || bytes.Contains([]byte(got), []byte("\x1b")) {
		t.Fatalf("output = %q, want one base and labeled sanitized execution facts", got)
	}
}

func TestRenderLinkDetailsLabelsDifferentBaseSnapshots(t *testing.T) {
	result := execution.StepResult{
		BaseDiagnostic:  &execution.DotfilesBaseDiagnostic{Source: execution.DotfilesBaseSourceEnv, CanonicalPath: "/report", Modules: []string{"bash"}},
		DotfilesFailure: &execution.DotfilesFailure{BaseSnapshot: &execution.DotfilesBaseDiagnostic{Source: execution.DotfilesBaseSourceHome, AttemptedCandidate: "/failure", Modules: []string{"bash"}}},
	}
	var output bytes.Buffer
	renderLinkDetails(&output, result)
	if got := output.String(); !bytes.Contains([]byte(got), []byte("report base context:")) || !bytes.Contains([]byte(got), []byte("failure base context:")) {
		t.Fatalf("output = %q, want explicit labels for differing snapshots", got)
	}
}

func TestRenderExecutionReportPreservesRemovedRollbackPathsWithoutAttemptFlag(t *testing.T) {
	report := execution.ExecutionReport{Results: []execution.StepResult{{
		Ref:      planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"},
		Status:   execution.StepStatusFailed,
		Message:  "dotfile module bash failed",
		Rollback: execution.LinkRollback{Removed: []string{"/home/ada/.bashrc"}},
	}}}

	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeConfirmed, report)
	got := stdout.String()
	for _, want := range []string{
		"rollback: attempted=false completed=false",
		"rollback removed: /home/ada/.bashrc",
	} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Fatalf("stdout missing %q: %q", want, got)
		}
	}
}

func TestRenderExecutionReportLabelsValidatedBaseCanonical(t *testing.T) {
	report := execution.ExecutionReport{Results: []execution.StepResult{{
		Ref:            planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"},
		Status:         execution.StepStatusInstalled,
		BaseDiagnostic: &execution.DotfilesBaseDiagnostic{Source: execution.DotfilesBaseSourceEnv, CanonicalPath: "/repo/.dotfiles", Modules: []string{"bash"}},
	}}}
	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeConfirmed, report)
	if got := stdout.String(); !bytes.Contains([]byte(got), []byte("dotfiles base: canonical base=/repo/.dotfiles source=env modules=bash")) {
		t.Fatalf("stdout missing canonical base context: %q", got)
	}
}

func TestRenderLinkDetailsRendersOnlyDeterministicBaseDiagnosticFacts(t *testing.T) {
	tests := []struct {
		name   string
		result execution.StepResult
		want   string
	}{
		{
			name: "validated canonical base",
			result: execution.StepResult{BaseDiagnostic: &execution.DotfilesBaseDiagnostic{
				Source:        execution.DotfilesBaseSourceHome,
				CanonicalPath: "/home/ada/.dotfiles",
				Modules:       []string{"bash"},
			}},
			want: "   dotfiles base: canonical base=/home/ada/.dotfiles source=home modules=bash\n",
		},
		{
			name: "rejected attempted candidate",
			result: execution.StepResult{BaseDiagnostic: &execution.DotfilesBaseDiagnostic{
				Source:             execution.DotfilesBaseSourceEnv,
				AttemptedCandidate: "/missing",
				Modules:            []string{"bash", "nvim"},
				Cause:              "stat dotfiles base: file does not exist",
			}},
			want: "   dotfiles base: source=env attempted candidate=/missing modules=bash, nvim cause=stat dotfiles base: file does not exist\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			renderLinkDetails(&output, tt.result)
			if got := output.String(); got != tt.want {
				t.Fatalf("renderLinkDetails() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderExecutionReportEscapesTerminalControlsInDotlinkDetails(t *testing.T) {
	report := execution.ExecutionReport{Results: []execution.StepResult{{
		Ref:     planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"},
		Status:  execution.StepStatusFailed,
		Message: "dotfile module bash failed: aggregate\x1b[2Jmessage\n",
		LinkDetails: []execution.LinkDetail{{
			Source:  "bash\x1b[2Jrc",
			Target:  "/home/ada/\x1b]8;;https://example.test\x1b\\.bashrc",
			Outcome: execution.LinkOutcomeFailed,
			Cause:   &execution.LinkCause{Code: "conflict\r", Message: "target\nexists"},
		}},
		Failure:        &execution.LinkFailure{Module: "bash\x1b[31m", Cause: execution.LinkCause{Code: "failed", Message: "bad\tstate"}},
		Rollback:       execution.LinkRollback{Attempted: true, Removed: []string{"/home/ada/\x1b[?25l.bashrc"}},
		BaseDiagnostic: &execution.DotfilesBaseDiagnostic{Source: execution.DotfilesBaseSourceEnv, AttemptedCandidate: "/tmp/\x1b[2Jdotfiles", Modules: []string{"bash\n"}, Cause: "invalid\rcandidate"},
	}}}

	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeConfirmed, report)
	got := stdout.String()
	if bytes.ContainsRune([]byte(got), '\x1b') || bytes.ContainsRune([]byte(got), '\r') {
		t.Fatalf("rendered terminal control character: %q", got)
	}
	for _, want := range []string{"aggregate\\x1b[2Jmessage\\x0a", "bash\\x1b[2Jrc", "target\\x0aexists", "conflict\\x0d", "bad\\x09state", "/tmp/\\x1b[2Jdotfiles"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Fatalf("stdout missing escaped value %q: %q", want, got)
		}
	}
}

func TestRenderExecutionReportHandlesEmptyReport(t *testing.T) {
	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeDefaultNonMutating, execution.ExecutionReport{})

	wantStdout := "Execution Report\n" +
		"Mode: default-non-mutating\n" +
		"\n" +
		"No actionable steps were selected; nothing to apply.\n" +
		"\n" +
		"Manual Actions:\n" +
		"- none\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("stdout = %q, want %q", got, wantStdout)
	}
}
