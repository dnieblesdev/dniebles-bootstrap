package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/execution"
	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func renderPlanResult(w io.Writer, profile string, resources []planning.ResourceRef, catalogPath string, facts planning.EnvironmentFacts, result planning.PlanResult) {
	if profile != "" {
		fmt.Fprintf(w, "Plan profile: %s\n", profile)
	} else {
		fmt.Fprintf(w, "Plan resources: %s\n", renderRefs(resources))
	}
	fmt.Fprintf(w, "Catalog: %s\n", catalogPath)
	fmt.Fprintf(w, "Environment: os=%s arch=%s distro=%s wsl=%t\n", facts.OS, facts.Arch, facts.Distro, facts.WSL)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Steps:")
	if len(result.Plan.Steps) == 0 {
		fmt.Fprintln(w, "- none")
	} else {
		statusByRef := resultStatuses(result.Results)
		for index, step := range result.Plan.Steps {
			status := statusByRef[step.Ref]
			if status == "" {
				status = planning.PlanStepStatusPlanned
			}
			fmt.Fprintf(w, "%d. %s [%s] %s\n", index+1, renderRef(step.Ref), status, step.Resource.Description)
			fmt.Fprintf(w, "   depends_on: %s\n", renderRefs(step.DependsOn))
			fmt.Fprintf(w, "   attention: %s\n", renderReasons(step.AttentionReasons))
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Results:")
	if len(result.Results) == 0 {
		fmt.Fprintln(w, "- none")
		return
	}
	for _, stepResult := range result.Results {
		fmt.Fprintf(w, "- %s: %s\n", renderResultRef(stepResult.Ref), stepResult.Status)
		if len(stepResult.Reasons) > 0 {
			for _, reason := range stepResult.Reasons {
				fmt.Fprintf(w, "  reason: %s\n", reason)
			}
		}
	}
}

func renderExecutionReport(w io.Writer, report execution.ExecutionReport) {
	fmt.Fprintln(w, "Execution Report")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Steps:")
	if len(report.Results) == 0 {
		fmt.Fprintln(w, "- none")
		return
	}
	for index, result := range report.Results {
		fmt.Fprintf(w, "%d. %s [%s]", index+1, renderRef(result.Ref), result.Status)
		if result.Message != "" {
			fmt.Fprintf(w, " %s", result.Message)
		}
		fmt.Fprintln(w)
	}
}

func renderDiagnostics(w io.Writer, result planning.PlanResult) {
	wroteHeader := false
	for _, stepResult := range result.Results {
		if stepResult.Status != planning.PlanStepStatusError {
			continue
		}
		if !wroteHeader {
			fmt.Fprintln(w, "Diagnostics:")
			wroteHeader = true
		}
		if len(stepResult.Reasons) == 0 {
			fmt.Fprintf(w, "- %s: error\n", renderResultRef(stepResult.Ref))
			continue
		}
		for _, reason := range stepResult.Reasons {
			fmt.Fprintf(w, "- %s: %s\n", renderResultRef(stepResult.Ref), reason)
		}
	}
}

func resultStatuses(results []planning.PlanStepResult) map[planning.ResourceRef]planning.PlanStepStatus {
	statuses := make(map[planning.ResourceRef]planning.PlanStepStatus, len(results))
	for _, result := range results {
		if result.Ref != (planning.ResourceRef{}) {
			statuses[result.Ref] = result.Status
		}
	}
	return statuses
}

func renderRefs(refs []planning.ResourceRef) string {
	if len(refs) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(refs))
	for _, ref := range refs {
		parts = append(parts, renderRef(ref))
	}
	return strings.Join(parts, ", ")
}

func renderReasons(reasons []string) string {
	if len(reasons) == 0 {
		return "none"
	}
	return strings.Join(reasons, "; ")
}

func renderResultRef(ref planning.ResourceRef) string {
	if ref == (planning.ResourceRef{}) {
		return "diagnostic"
	}
	return renderRef(ref)
}

func renderRef(ref planning.ResourceRef) string {
	return string(ref.Kind) + ":" + ref.Name
}
