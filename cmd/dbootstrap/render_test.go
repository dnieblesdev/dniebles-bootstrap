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

func TestRenderExecutionReportIsDistinctFromPlanRendering(t *testing.T) {
	report := execution.ExecutionReport{
		Results: []execution.StepResult{
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}, Status: execution.StepStatusNotImplemented, Message: "noop installer does not perform real installation"},
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}, Status: execution.StepStatusNotImplemented, Message: "noop installer does not perform real installation"},
		},
	}

	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeDryRun, report)

	wantStdout := "Execution Report\n" +
		"Mode: dry-run\n" +
		"\n" +
		"Steps:\n" +
		"1. tool:git [not_implemented] noop installer does not perform real installation\n" +
		"2. package:ripgrep [not_implemented] noop installer does not perform real installation\n" +
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
		"Steps:\n" +
		"1. tool:fd [not_implemented] noop installer does not perform real installation\n" +
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

func TestRenderExecutionReportWarnsInConfirmedMode(t *testing.T) {
	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeConfirmed, execution.ExecutionReport{})

	if got := stdout.String(); !bytes.Contains([]byte(got), []byte("Warning: confirmed mode may run real brew install commands")) {
		t.Fatalf("stdout missing confirmed warning: %q", got)
	}
}

func TestRenderExecutionReportHandlesEmptyReport(t *testing.T) {
	var stdout bytes.Buffer
	renderExecutionReport(&stdout, applyModeDefaultNonMutating, execution.ExecutionReport{})

	wantStdout := "Execution Report\n" +
		"Mode: default-non-mutating\n" +
		"\n" +
		"Steps:\n" +
		"- none\n" +
		"\n" +
		"Manual Actions:\n" +
		"- none\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("stdout = %q, want %q", got, wantStdout)
	}
}
