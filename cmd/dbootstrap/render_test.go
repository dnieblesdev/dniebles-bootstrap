package main

import (
	"bytes"
	"testing"

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
	renderPlanResult(&stdout, "dev", "catalog/bootstrap.toml", planning.EnvironmentFacts{OS: "linux", Arch: "amd64"}, result)
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
