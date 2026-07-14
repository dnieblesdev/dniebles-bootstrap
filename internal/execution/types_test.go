package execution

import (
	"errors"
	"io/fs"
	"os/exec"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestStepStatusVocabulary(t *testing.T) {
	tests := []struct {
		name   string
		status StepStatus
		want   string
	}{
		{"installed", StepStatusInstalled, "installed"},
		{"failed", StepStatusFailed, "failed"},
		{"skipped", StepStatusSkipped, "skipped"},
		{"not_implemented", StepStatusNotImplemented, "not_implemented"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.status); got != tt.want {
				t.Fatalf("status = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDotfilesFailurePreservesPrerequisiteTargetAndIndependentCauses(t *testing.T) {
	exit := &exec.ExitError{}
	failure := &DotfilesFailure{
		Phase: DotfilesPhasePrerequisite,
		PrerequisiteTarget: &DotfilesPrerequisiteTarget{
			Kind:               DotfilesPrerequisiteRunner,
			AttemptedCandidate: "/repo/bin/dotlink",
		},
		PrerequisiteErr: errors.Join(fs.ErrNotExist, &fs.PathError{Op: "stat", Path: "/repo/bin/dotlink", Err: fs.ErrNotExist}),
		ExecutionErr:    errors.Join(ErrDotlinkCommandFailed, exit),
		ParseErr:        ErrInvalidDotlinkReport,
	}

	if failure.Phase != DotfilesPhasePrerequisite || failure.PrerequisiteTarget.Kind != DotfilesPrerequisiteRunner || failure.PrerequisiteTarget.AttemptedCandidate != "/repo/bin/dotlink" {
		t.Fatalf("failure transport = %#v, want prerequisite runner candidate", failure)
	}
	if !errors.Is(failure, fs.ErrNotExist) || !errors.Is(failure, ErrDotlinkCommandFailed) || !errors.Is(failure, ErrInvalidDotlinkReport) {
		t.Fatalf("failure did not preserve independent sentinel causes: %v", failure)
	}
	var gotExit *exec.ExitError
	var gotPath *fs.PathError
	if !errors.As(failure, &gotExit) || !errors.As(failure, &gotPath) {
		t.Fatalf("failure did not preserve typed causes: %v", failure)
	}
}

func TestStepResultShape(t *testing.T) {
	ref := planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}
	wantErr := errors.New("boom")

	result := StepResult{
		Ref:     ref,
		Status:  StepStatusFailed,
		Message: "something went wrong",
		Err:     wantErr,
	}

	if result.Ref != ref {
		t.Fatalf("Ref = %#v, want %#v", result.Ref, ref)
	}
	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
	}
	if result.Message != "something went wrong" {
		t.Fatalf("Message = %q, want %q", result.Message, "something went wrong")
	}
	if result.Err != wantErr {
		t.Fatalf("Err = %v, want %v", result.Err, wantErr)
	}
}

func TestExecutionReportAggregatesResults(t *testing.T) {
	report := ExecutionReport{
		Results: []StepResult{
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}, Status: StepStatusInstalled},
			{Ref: planning.ResourceRef{Kind: planning.ResourceKindPackage, Name: "ripgrep"}, Status: StepStatusSkipped},
		},
	}

	if len(report.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(report.Results))
	}
	if report.Results[0].Status != StepStatusInstalled {
		t.Fatalf("first status = %q, want installed", report.Results[0].Status)
	}
	if report.Results[1].Status != StepStatusSkipped {
		t.Fatalf("second status = %q, want skipped", report.Results[1].Status)
	}
}
