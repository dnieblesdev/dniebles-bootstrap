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

func renderExecutionReport(w io.Writer, mode applyMode, report execution.ExecutionReport) {
	fmt.Fprintln(w, "Execution Report")
	fmt.Fprintf(w, "Mode: %s\n", mode)
	if mode == applyModeConfirmed {
		fmt.Fprintln(w, "Confirmed mode: brew-backed tool/package steps and selected dotfile resources may have changed this machine; runtime, non-brew, unselected, and unsupported steps remain non-mutating or not supported yet.")
	}
	fmt.Fprintln(w)

	if len(report.Results) == 0 {
		fmt.Fprintln(w, "No actionable steps were selected; nothing to apply.")
		fmt.Fprintln(w)
		renderManualActions(w, report)
		return
	}

	renderExecutionSummary(w, report.Results)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Steps:")
	for index, result := range report.Results {
		fmt.Fprintf(w, "%d. %s [%s]", index+1, renderRef(result.Ref), renderExecutionStepStatus(result.Status))
		if result.Message != "" {
			fmt.Fprintf(w, " %s", result.Message)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w)
	renderManualActions(w, report)
}

func renderExecutionSummary(w io.Writer, results []execution.StepResult) {
	counts := executionSummaryCounts(results)
	fmt.Fprintln(w, "Summary:")
	for _, category := range executionSummaryCategories() {
		fmt.Fprintf(w, "- %s: %d\n", category, counts[category])
	}
}

func executionSummaryCounts(results []execution.StepResult) map[string]int {
	counts := make(map[string]int, len(executionSummaryCategories()))
	for _, category := range executionSummaryCategories() {
		counts[category] = 0
	}
	for _, result := range results {
		counts[executionSummaryCategory(result.Status)]++
	}
	return counts
}

func executionSummaryCategories() []string {
	return []string{"changed", "unchanged", "not supported yet", "failed"}
}

func executionSummaryCategory(status execution.StepStatus) string {
	switch status {
	case execution.StepStatusInstalled:
		return "changed"
	case execution.StepStatusSkipped:
		return "unchanged"
	case execution.StepStatusNotImplemented:
		return "not supported yet"
	case execution.StepStatusFailed:
		return "failed"
	default:
		return "failed"
	}
}

func renderExecutionStepStatus(status execution.StepStatus) string {
	return executionSummaryCategory(status)
}

func renderManualActions(w io.Writer, report execution.ExecutionReport) {
	fmt.Fprintln(w, "Manual Actions:")
	if len(report.ManualActions) == 0 {
		fmt.Fprintln(w, "- none")
		return
	}
	for _, action := range report.ManualActions {
		fmt.Fprintf(w, "- %s: %s\n", action.ID, action.Title)
		fmt.Fprintf(w, "  reason: %s\n", action.Reason)
		for _, instruction := range action.Instructions {
			fmt.Fprintf(w, "  instruction: %s\n", instruction)
		}
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
