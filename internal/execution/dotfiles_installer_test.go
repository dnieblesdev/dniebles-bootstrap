package execution

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

func TestDotfilesInstallerMapsDotfileNameOnly(t *testing.T) {
	provider := &fakeDotfilesProvider{}
	installer := NewDotfilesInstaller(provider)
	step := planning.PlanStep{
		Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"},
		Resource: planning.Resource{
			Ref:         planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "other"},
			Description: "ignored",
			Install:     &planning.InstallMetadata{Provider: "shell", Package: "danger"},
		},
	}

	result := installer.Install(context.Background(), step)
	if result.Status != StepStatusInstalled {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusInstalled)
	}
	if len(provider.modules) != 1 || len(provider.modules[0]) != 1 || provider.modules[0][0] != "bash" {
		t.Fatalf("provider modules = %#v, want [[bash]]", provider.modules)
	}
}

func TestDotfilesInstallerRejectsNonDotfileWithoutProviderCall(t *testing.T) {
	provider := &fakeDotfilesProvider{}
	installer := NewDotfilesInstaller(provider)
	step := planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindTool, Name: "git"}}

	result := installer.Install(context.Background(), step)
	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
	}
	if len(provider.modules) != 0 {
		t.Fatalf("provider modules = %#v, want none", provider.modules)
	}
}

func TestDotfilesInstallerProviderErrorFailsStep(t *testing.T) {
	providerErr := errors.New("dotlink failed")
	installer := NewDotfilesInstaller(&fakeDotfilesProvider{err: providerErr})
	step := planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "nvim"}}

	result := installer.Install(context.Background(), step)
	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
	}
	if !errors.Is(result.Err, providerErr) {
		t.Fatalf("Err = %v, want provider error", result.Err)
	}
}

func TestDotfilesInstallerRequiresProvider(t *testing.T) {
	installer := NewDotfilesInstaller(nil)
	step := planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}}

	result := installer.Install(context.Background(), step)
	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
	}
}

func TestDotfilesInstallerTranslatesValidatedReports(t *testing.T) {
	tests := []struct {
		name       string
		fixture    string
		wantStatus StepStatus
		wantDetail []LinkOutcome
	}{
		{name: "changed", fixture: "all-changed.json", wantStatus: StepStatusInstalled, wantDetail: []LinkOutcome{LinkOutcomeChanged}},
		{name: "unchanged", fixture: "all-unchanged.json", wantStatus: StepStatusSkipped, wantDetail: []LinkOutcome{LinkOutcomeUnchanged}},
		{name: "mixed", fixture: "mixed.json", wantStatus: StepStatusInstalled, wantDetail: []LinkOutcome{LinkOutcomeChanged, LinkOutcomeUnchanged}},
		{name: "failed", fixture: "failed.json", wantStatus: StepStatusFailed, wantDetail: []LinkOutcome{LinkOutcomeFailed}},
		{name: "rolled back", fixture: "rolled-back.json", wantStatus: StepStatusFailed, wantDetail: []LinkOutcome{LinkOutcomeRolledBack}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := ParseDotlinkLinkReport(readDotlinkReportFixture(t, tt.fixture), []string{"bash"})
			if err != nil {
				t.Fatalf("ParseDotlinkLinkReport() error = %v", err)
			}
			result := NewDotfilesInstaller(&fakeDotfilesProvider{report: report}).Install(context.Background(), planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}})
			if result.Status != tt.wantStatus {
				t.Fatalf("Status = %q, want %q", result.Status, tt.wantStatus)
			}
			if len(result.LinkDetails) != len(tt.wantDetail) {
				t.Fatalf("LinkDetails = %#v, want %d details", result.LinkDetails, len(tt.wantDetail))
			}
			for index, want := range tt.wantDetail {
				if result.LinkDetails[index].Outcome != want {
					t.Fatalf("LinkDetails[%d].Outcome = %q, want %q", index, result.LinkDetails[index].Outcome, want)
				}
			}
		})
	}
}

func TestDotfilesInstallerAggregateFailedStatusOverridesSuccessfulEntries(t *testing.T) {
	report, err := ParseDotlinkLinkReport([]byte(`{"schema_version":1,"modules":["bash"],"status":"failed","entries":[{"module":"bash","source":"bashrc","target":"/home/ada/.bashrc","outcome":"changed"},{"module":"bash","source":"profile","target":"/home/ada/.profile","outcome":"unchanged"}],"failure":{"module":"bash","cause":{"code":"link_failed","message":"target exists"}},"rollback":{"attempted":false,"completed":false,"removed":[]}}`), []string{"bash"})
	if err != nil {
		t.Fatalf("ParseDotlinkLinkReport() error = %v", err)
	}
	report.Failure = nil // Isolate aggregate status from the required failed-report failure context.

	result := NewDotfilesInstaller(&fakeDotfilesProvider{report: report}).Install(context.Background(), planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}})
	if result.Status != StepStatusFailed {
		t.Fatalf("Status = %q, want %q when aggregate report status is failed", result.Status, StepStatusFailed)
	}
	for _, want := range []string{
		"dotfile module bash failed",
		"rollback attempted=false",
		"verify rollback state and restore affected targets before retrying dotlink",
	} {
		if !strings.Contains(result.Message, want) {
			t.Fatalf("Message = %q, want recovery guidance %q", result.Message, want)
		}
	}
}

func TestDotfilesInstallerFailedReportsAlwaysIncludeRecoveryGuidance(t *testing.T) {
	tests := []struct {
		name   string
		report DotlinkLinkReport
	}{
		{
			name: "aggregate failure without changed entries",
			report: DotlinkLinkReport{
				Status:  DotlinkReportStatusFailed,
				Entries: []DotlinkLinkEntry{{Module: "bash", Outcome: DotlinkLinkOutcomeUnchanged}},
			},
		},
		{
			name: "rolled back entry without changed entry",
			report: DotlinkLinkReport{
				Status:   DotlinkReportStatusFailed,
				Entries:  []DotlinkLinkEntry{{Module: "bash", Outcome: DotlinkLinkOutcomeRolledBack}},
				Rollback: DotlinkRollback{Removed: []string{"/home/ada/.bashrc"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewDotfilesInstaller(&fakeDotfilesProvider{report: tt.report}).Install(context.Background(), planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}})
			if result.Status != StepStatusFailed {
				t.Fatalf("Status = %q, want %q", result.Status, StepStatusFailed)
			}
			for _, want := range []string{"recovery:", "verify rollback state", "restore affected targets", "before retrying dotlink"} {
				if !strings.Contains(result.Message, want) {
					t.Fatalf("Message = %q, missing recovery guidance %q", result.Message, want)
				}
			}
		})
	}
}

func TestDotfilesInstallerPreservesOrderedLinkDetailsAndFailureContext(t *testing.T) {
	report, err := ParseDotlinkLinkReport(readDotlinkReportFixture(t, "mixed.json"), []string{"bash"})
	if err != nil {
		t.Fatalf("ParseDotlinkLinkReport() error = %v", err)
	}
	result := NewDotfilesInstaller(&fakeDotfilesProvider{report: report}).Install(context.Background(), planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}})
	if got, want := result.LinkDetails[0].Source, "bashrc"; got != want {
		t.Fatalf("first source = %q, want %q", got, want)
	}
	if got, want := result.LinkDetails[1].Target, "/home/ada/.profile"; got != want {
		t.Fatalf("second target = %q, want %q", got, want)
	}

	failed, err := ParseDotlinkLinkReport(readDotlinkReportFixture(t, "rolled-back.json"), []string{"bash"})
	if err != nil {
		t.Fatalf("ParseDotlinkLinkReport() error = %v", err)
	}
	result = NewDotfilesInstaller(&fakeDotfilesProvider{report: failed}).Install(context.Background(), planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}})
	if result.LinkDetails[0].Cause == nil || result.Failure == nil || !result.Rollback.Attempted {
		t.Fatalf("result = %#v, want detail cause, aggregate failure, and rollback", result)
	}
}

func TestDotfilesInstallerProviderErrorHasNoInferredLinks(t *testing.T) {
	providerErr := errors.New("safe provider failure")
	result := NewDotfilesInstaller(&fakeDotfilesProvider{err: providerErr}).Install(context.Background(), planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}})
	if result.Status != StepStatusFailed || len(result.LinkDetails) != 0 || !errors.Is(result.Err, providerErr) {
		t.Fatalf("result = %#v, want failed result without inferred links", result)
	}
}

func TestDotfilesInstallerPreservesFailedReportAndExecutionError(t *testing.T) {
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusFailed, ExitCode: 2, Err: errors.New("exit 2"), Stdout: string(readDotlinkReportFixture(t, "failed.json"))}}
	result := NewDotfilesInstaller(newFakeLocalProvider("/repo", runner)).Install(context.Background(), planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}})
	if result.Status != StepStatusFailed || len(result.LinkDetails) == 0 || !errors.Is(result.Err, ErrDotlinkCommandFailed) || result.DotfilesFailure == nil {
		t.Fatalf("result = %#v, want failed step with retained report and execution failure", result)
	}
	if strings.Contains(result.Message, "dotfiles base") || result.BaseDiagnostic == nil || result.BaseDiagnostic.CanonicalPath != "/repo" {
		t.Fatalf("result = %#v, want base-free message and unchanged primary base diagnostic", result)
	}
}

func TestDotfilesInstallerResolvesBaseOnceForDiagnosticMessageAndExecution(t *testing.T) {
	homeCalls := 0
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded, Stdout: string(readDotlinkReportFixture(t, "all-changed.json"))}}
	provider := &LocalDotfilesProvider{
		Resolver: DotfilesBaseResolver{
			LookupEnv: func(string) (string, bool) { return "/repo", true },
			HomeDir: func() (string, error) {
				homeCalls++
				return "/home/ada", nil
			},
			Stat: statDirs("/repo"),
			EvalSymlinks: fakeEval(map[string]string{
				"/home/ada": "/home/ada",
				"/repo":     "/repo",
			}),
		},
		Runner: runner,
		Stat:   statDirs("/repo", "/repo/bin/dotlink", "/repo/bash"),
		EvalSymlinks: fakeEval(map[string]string{
			"/home/ada":         "/home/ada",
			"/repo":             "/repo",
			"/repo/bin/dotlink": "/repo/bin/dotlink",
			"/repo/bash":        "/repo/bash",
		}),
	}

	result := NewDotfilesInstaller(provider).Install(context.Background(), planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}})
	if result.Status != StepStatusInstalled {
		t.Fatalf("Status = %q, want %q", result.Status, StepStatusInstalled)
	}
	if homeCalls != 1 {
		t.Fatalf("home directory lookups = %d, want one shared base resolution", homeCalls)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("runner calls = %d, want 1", len(runner.calls))
	}
}

func TestDotfilesInstallerRetainsFailedBaseDiagnostic(t *testing.T) {
	resolver := fakeDotfilesBaseResolver("", true, "/home/ada")
	provider := &LocalDotfilesProvider{Resolver: resolver, Runner: &fakeCommandRunner{}}

	result := NewDotfilesInstaller(provider).Install(context.Background(), planning.PlanStep{Ref: planning.ResourceRef{Kind: planning.ResourceKindDotfile, Name: "bash"}})
	if result.Status != StepStatusFailed || !errors.Is(result.Err, ErrEmptyDotfilesBase) {
		t.Fatalf("result = %#v, want failed empty-base result", result)
	}
	if result.BaseDiagnostic == nil {
		t.Fatal("BaseDiagnostic = nil, want retained resolution diagnostic")
	}
	diagnostic := result.BaseDiagnostic
	if diagnostic.Source != DotfilesBaseSourceEnv || diagnostic.AttemptedCandidate != "" || diagnostic.CanonicalPath != "" || diagnostic.Cause != ErrEmptyDotfilesBase.Error() {
		t.Fatalf("diagnostic = %#v, want safe empty-env failure", diagnostic)
	}
	if len(diagnostic.Modules) != 1 || diagnostic.Modules[0] != "bash" {
		t.Fatalf("diagnostic modules = %#v, want [bash]", diagnostic.Modules)
	}
}

type fakeDotfilesProvider struct {
	modules [][]string
	err     error
	report  DotlinkLinkReport
}

func (f *fakeDotfilesProvider) EnsureModules(context.Context, []string) error { return nil }
func (f *fakeDotfilesProvider) RunDotlink(_ context.Context, modules []string) error {
	copied := append([]string(nil), modules...)
	f.modules = append(f.modules, copied)
	return f.err
}
func (f *fakeDotfilesProvider) RunDotlinkReport(ctx context.Context, modules []string) (DotlinkLinkReport, error) {
	if err := f.RunDotlink(ctx, modules); err != nil {
		return DotlinkLinkReport{}, err
	}
	return f.report, nil
}
