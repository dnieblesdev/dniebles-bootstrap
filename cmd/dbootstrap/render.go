package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	if isConfirmedMode(mode) {
		fmt.Fprintln(w, "Confirmed mode: brew-backed tool/package steps, eligible Linux APT-backed tool/package steps, and selected dotfile resources may have changed this machine; unsupported, non-provider-backed, and unselected steps remain non-mutating or not supported yet.")
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
			fmt.Fprintf(w, " %s", sanitizeTerminalText(result.Message))
		}
		fmt.Fprintln(w)
		renderLinkDetails(w, result)
	}

	fmt.Fprintln(w)
	renderManualActions(w, report)
}

func renderLinkDetails(w io.Writer, result execution.StepResult) {
	for _, detail := range result.LinkDetails {
		fmt.Fprintf(w, "   link: %s source=%s target=%s\n", detail.Outcome, sanitizeTerminalText(detail.Source), sanitizeTerminalText(detail.Target))
		if detail.Cause != nil {
			fmt.Fprintf(w, "   cause: %s: %s\n", sanitizeTerminalText(detail.Cause.Code), sanitizeTerminalText(detail.Cause.Message))
		}
	}
	if result.Failure != nil {
		fmt.Fprintf(w, "   aggregate failure: module=%s cause=%s: %s\n", sanitizeTerminalText(result.Failure.Module), sanitizeTerminalText(result.Failure.Cause.Code), sanitizeTerminalText(result.Failure.Cause.Message))
	}
	if result.Rollback.Attempted || result.Rollback.Completed || len(result.Rollback.Removed) > 0 {
		fmt.Fprintf(w, "   rollback: attempted=%t completed=%t\n", result.Rollback.Attempted, result.Rollback.Completed)
		for _, removed := range result.Rollback.Removed {
			fmt.Fprintf(w, "   rollback removed: %s\n", sanitizeTerminalText(removed))
		}
	}
	if failure := result.DotfilesFailure; failure != nil {
		if result.BaseDiagnostic != nil && failure.BaseSnapshot != nil && !sameBaseDiagnostic(*result.BaseDiagnostic, failure.BaseSnapshot) {
			renderBaseDiagnostic(w, "report base context", result.BaseDiagnostic)
			renderBaseDiagnostic(w, "failure base context", failure.BaseSnapshot)
		} else if result.BaseDiagnostic != nil {
			renderBaseDiagnostic(w, "dotfiles base", result.BaseDiagnostic)
		} else if failure.BaseSnapshot != nil {
			renderBaseDiagnostic(w, "failure base context", failure.BaseSnapshot)
		}
		if failure.Phase != "" {
			fmt.Fprintf(w, "   phase: %s\n", sanitizeBoundedDiagnosticText(string(failure.Phase)))
		}
		if target := failure.PrerequisiteTarget; target != nil {
			fmt.Fprintf(w, "   attempted %s candidate: %s\n", sanitizeBoundedDiagnosticText(string(target.Kind)), sanitizeBoundedDiagnosticText(target.AttemptedCandidate))
		}
		if cause := dotfilesFailureCause(failure); cause != "" {
			fmt.Fprintf(w, "   cause: %s\n", cause)
		}
		if failure.Executable != "" || failure.Runner != "" || len(failure.Command.Args) != 0 {
			fmt.Fprintf(w, "   executable: %s\n   runner: %s\n   command: %s\n", sanitizeTerminalText(failure.Executable), sanitizeTerminalText(failure.Runner), sanitizeTerminalText(strings.Join(failure.Command.Args, " ")))
		}
		if failure.ExitCode != nil {
			fmt.Fprintf(w, "   exit code: %d\n", *failure.ExitCode)
		}
		if failure.Stderr != "" {
			fmt.Fprintf(w, "   stderr: %s\n", sanitizeTerminalText(failure.Stderr))
		}
	} else if result.BaseDiagnostic != nil {
		renderBaseDiagnostic(w, "dotfiles base", result.BaseDiagnostic)
	}
}

func dotfilesFailureCause(failure *execution.DotfilesFailure) string {
	if errors.Is(failure, fs.ErrNotExist) {
		return "path does not exist"
	}
	if errors.Is(failure, execution.ErrDotfilesPathEscapes) {
		return "path escapes dotfiles base"
	}
	if errors.Is(failure, execution.ErrInvalidDotfileModule) {
		return "invalid dotfile module"
	}
	if errors.Is(failure, execution.ErrInvalidDotlinkReport) {
		return "invalid dotlink report"
	}
	if errors.Is(failure, execution.ErrDotlinkCommandFailed) {
		return "dotlink command failed"
	}
	return ""
}

func renderBaseDiagnostic(w io.Writer, label string, diagnostic *execution.DotfilesBaseDiagnostic) {
	if diagnostic == nil {
		return
	}
	modules := sanitizeBoundedDiagnosticText(strings.Join(diagnostic.Modules, ", "))
	if diagnostic.CanonicalPath != "" {
		fmt.Fprintf(w, "   %s: canonical base=%s source=%s modules=%s\n", label, sanitizeBoundedDiagnosticText(diagnostic.CanonicalPath), sanitizeBoundedDiagnosticText(string(diagnostic.Source)), modules)
		return
	}
	fmt.Fprintf(w, "   %s: source=%s attempted candidate=%s modules=%s", label, sanitizeBoundedDiagnosticText(string(diagnostic.Source)), sanitizeBoundedDiagnosticText(diagnostic.AttemptedCandidate), modules)
	if diagnostic.Cause != "" {
		fmt.Fprintf(w, " cause=%s", sanitizeBoundedDiagnosticText(diagnostic.Cause))
	}
	fmt.Fprintln(w)
}

func sameBaseDiagnostic(primary execution.DotfilesBaseDiagnostic, snapshot *execution.DotfilesBaseDiagnostic) bool {
	if snapshot == nil || primary.Source != snapshot.Source || primary.AttemptedCandidate != snapshot.AttemptedCandidate || primary.CanonicalPath != snapshot.CanonicalPath || primary.Cause != snapshot.Cause || len(primary.Modules) != len(snapshot.Modules) {
		return false
	}
	for i := range primary.Modules {
		if primary.Modules[i] != snapshot.Modules[i] {
			return false
		}
	}
	return true
}

func sanitizeTerminalText(value string) string {
	var sanitized strings.Builder
	for _, r := range value {
		if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
			fmt.Fprintf(&sanitized, `\x%02x`, r)
			continue
		}
		sanitized.WriteRune(r)
	}
	return sanitized.String()
}

const (
	maxRenderedDiagnosticBytes  = 4096
	renderedDiagnosticTruncated = "...[truncated]"
)

func sanitizeBoundedDiagnosticText(value string) string {
	sanitized := sanitizeTerminalText(value)
	if len(sanitized) <= maxRenderedDiagnosticBytes {
		return sanitized
	}

	var bounded strings.Builder
	bounded.Grow(maxRenderedDiagnosticBytes)
	for _, r := range sanitized {
		encoded := string(r)
		if bounded.Len()+len(encoded)+len(renderedDiagnosticTruncated) > maxRenderedDiagnosticBytes {
			break
		}
		bounded.WriteString(encoded)
	}
	bounded.WriteString(renderedDiagnosticTruncated)
	return bounded.String()
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
