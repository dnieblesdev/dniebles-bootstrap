package execution

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestLocalDotfilesProviderBuildsExactCommand(t *testing.T) {
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded, Stdout: string(readDotlinkReportFixture(t, "all-changed.json"))}}
	provider := newFakeLocalProvider("/repo", runner)

	report, err := provider.RunDotlinkReport(context.Background(), []string{"bash"})
	if err != nil {
		t.Fatalf("RunDotlinkReport() error = %v", err)
	}
	if report.Status != DotlinkReportStatusSuccess {
		t.Fatalf("report status = %q, want success", report.Status)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("runner calls = %d, want 1", len(runner.calls))
	}
	want := CommandRequest{Executable: "/repo/bin/dotlink", Args: []string{"link", "--report=json", "bash"}, Dir: "/repo", Timeout: DefaultDotlinkTimeout}
	if !reflect.DeepEqual(runner.calls[0], want) {
		t.Fatalf("CommandRequest = %#v, want %#v", runner.calls[0], want)
	}
}

func TestDotfilesFailurePreservesIndependentCausesAndBoundedStderr(t *testing.T) {
	exit := &exec.ExitError{}
	syntax := &json.SyntaxError{}
	failure := &DotfilesFailure{ExecutionErr: errors.Join(ErrDotlinkCommandFailed, exit), ParseErr: errors.Join(ErrInvalidDotlinkReport, syntax), Stderr: sanitizeDotlinkStderr(strings.Repeat("é\x1b", 5000))}
	if !errors.Is(failure, ErrDotlinkCommandFailed) || !errors.Is(failure, ErrInvalidDotlinkReport) {
		t.Fatalf("failure did not preserve sentinel identities: %v", failure)
	}
	var gotExit *exec.ExitError
	var gotSyntax *json.SyntaxError
	if !errors.As(failure, &gotExit) || !errors.As(failure, &gotSyntax) || len(failure.Stderr) > 4096 || !strings.HasSuffix(failure.Stderr, dotlinkStderrTruncated) || strings.Contains(failure.Stderr, "\x1b") {
		t.Fatalf("failure = %#v, want joined typed causes and bounded escaped stderr", failure)
	}
}

func TestLocalDotfilesProviderComposesExecutionAndReportOutcomes(t *testing.T) {
	exit := &exec.ExitError{}
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusFailed, Err: exit, ExitCode: 7, Stdout: "{", Stderr: "bad\x1b[31m"}}
	report, err := newFakeLocalProvider("/repo", runner).RunDotlinkReport(context.Background(), []string{"bash"})
	if report.Status != "" || !errors.Is(err, ErrDotlinkCommandFailed) || !errors.Is(err, ErrInvalidDotlinkReport) {
		t.Fatalf("report=%#v err=%v, want discarded invalid report and both causes", report, err)
	}
	var got *DotfilesFailure
	if !errors.As(err, &got) || got.ExitCode == nil || *got.ExitCode != 7 || got.Executable != "/repo/bin/dotlink" {
		t.Fatalf("failure = %#v, want structured canonical execution context", got)
	}
}

func TestLocalDotfilesProviderFailedCommandMalformedReportPreservesAllCauses(t *testing.T) {
	exit := &exec.ExitError{}
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusFailed, Err: exit, Stdout: `{"schema_version":]}`}}
	provider := newFakeLocalProvider("/repo", runner)
	_, err := provider.RunDotlinkReportWithExecutionContext(context.Background(), []string{"bash"}, provider.ResolveDotfilesExecutionContext([]string{"bash"}))
	var gotExit *exec.ExitError
	var gotSyntax *json.SyntaxError
	if !errors.Is(err, ErrDotlinkCommandFailed) || !errors.Is(err, ErrInvalidDotlinkReport) || !errors.As(err, &gotExit) || !errors.As(err, &gotSyntax) {
		t.Fatalf("provider error = %v, want both sentinels and concrete causes", err)
	}
}

func TestLocalDotfilesProviderMissingRunnerRetainsCanonicalCommandContext(t *testing.T) {
	runner := &fakeCommandRunner{}
	provider := newFakeLocalProvider("/repo", runner)
	provider.Runner = nil
	_, err := provider.RunDotlinkReportWithExecutionContext(context.Background(), []string{"bash"}, provider.ResolveDotfilesExecutionContext([]string{"bash"}))
	var failure *DotfilesFailure
	if !errors.As(err, &failure) || failure.Runner != "CommandRunner" || failure.Executable != "/repo/bin/dotlink" {
		t.Fatalf("failure = %#v, want runner and canonical executable", failure)
	}
	want := CommandRequest{Executable: "/repo/bin/dotlink", Args: []string{"link", "--report=json", "bash"}, Dir: "/repo", Timeout: DefaultDotlinkTimeout}
	if !reflect.DeepEqual(failure.Command, want) || !errors.Is(err, ErrMissingDotlinkRunner) {
		t.Fatalf("failure = %#v, want command request and missing-runner cause", failure)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("runner calls = %d, want 0", len(runner.calls))
	}
}

func TestLocalDotfilesProviderReconcilesCommandAndReport(t *testing.T) {
	exit := &exec.ExitError{}
	tests := []struct {
		name      string
		result    CommandResult
		wantErr   error
		wantExit  *exec.ExitError
		wantSeen  bool
		wantPhase DotfilesPhase
	}{
		{name: "success report on success", result: CommandResult{Status: CommandStatusSucceeded, Stdout: string(readDotlinkReportFixture(t, "all-changed.json"))}, wantSeen: true},
		{name: "failed report on failed command", result: CommandResult{Status: CommandStatusFailed, Err: exit, Stdout: string(readDotlinkReportFixture(t, "failed.json")), Stderr: "human output must be ignored"}, wantErr: ErrDotlinkCommandFailed, wantExit: exit, wantSeen: true, wantPhase: DotfilesPhaseCommandExecution},
		{name: "failed report on timed out command", result: CommandResult{Status: CommandStatusTimedOut, Stdout: string(readDotlinkReportFixture(t, "failed.json"))}, wantErr: ErrInconsistentDotlinkReport, wantPhase: DotfilesPhaseReportValidation},
		{name: "failed report on command not run", result: CommandResult{Status: CommandStatusNotRun, Stdout: string(readDotlinkReportFixture(t, "failed.json"))}, wantErr: ErrInconsistentDotlinkReport, wantPhase: DotfilesPhaseReportValidation},
		{name: "missing report on failed command", result: CommandResult{Status: CommandStatusFailed, Stderr: string(readDotlinkReportFixture(t, "failed.json"))}, wantErr: ErrDotlinkCommandFailed, wantPhase: DotfilesPhaseReportValidation},
		{name: "invalid report on failed command", result: CommandResult{Status: CommandStatusFailed, Stdout: "not JSON"}, wantErr: ErrDotlinkCommandFailed, wantPhase: DotfilesPhaseReportValidation},
		{name: "success report on failed command", result: CommandResult{Status: CommandStatusFailed, Stdout: string(readDotlinkReportFixture(t, "status-exit-mismatch.json"))}, wantErr: ErrInconsistentDotlinkReport, wantPhase: DotfilesPhaseReportValidation},
		{name: "failed report on success command", result: CommandResult{Status: CommandStatusSucceeded, Stdout: string(readDotlinkReportFixture(t, "failed.json"))}, wantErr: ErrInconsistentDotlinkReport, wantPhase: DotfilesPhaseReportValidation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{result: tt.result}
			provider := newFakeLocalProvider("/repo", runner)
			report, err := provider.RunDotlinkReportWithExecutionContext(context.Background(), []string{"bash"}, provider.ResolveDotfilesExecutionContext([]string{"bash"}))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RunDotlinkReport() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantPhase != "" {
				var failure *DotfilesFailure
				if !errors.As(err, &failure) || failure.Phase != tt.wantPhase {
					t.Fatalf("failure = %#v, want phase %q", failure, tt.wantPhase)
				}
			}
			var gotExit *exec.ExitError
			if tt.wantExit != nil && (!errors.Is(err, ErrDotlinkCommandFailed) || !errors.As(err, &gotExit) || gotExit != tt.wantExit) {
				t.Fatalf("RunDotlinkReportWithExecutionContext() error = %v, want ErrDotlinkCommandFailed and exit error %p", err, tt.wantExit)
			}
			if tt.wantSeen && (report.Status == "" || len(report.Entries) == 0) {
				t.Fatalf("report = %#v, error = %v, want retained validated report", report, err)
			}
			if !tt.wantSeen && err != nil && (report.Status != "" || len(report.Entries) != 0) {
				t.Fatalf("report = %#v, want no trusted report details on failure", report)
			}
			if len(runner.calls) != 1 {
				t.Fatalf("runner calls = %d, want 1", len(runner.calls))
			}
		})
	}
}

func TestLocalDotfilesProviderLegacyBoundaryFailsForValidatedFailedReport(t *testing.T) {
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusFailed, Stdout: string(readDotlinkReportFixture(t, "failed.json"))}}
	provider := newFakeLocalProvider("/repo", runner)

	if err := provider.RunDotlink(context.Background(), []string{"bash"}); !errors.Is(err, ErrDotlinkCommandFailed) {
		t.Fatalf("RunDotlink() error = %v, want ErrDotlinkCommandFailed", err)
	}
}

func TestLocalDotfilesProviderBaseDiagnosticKeepsUnresolvedCandidateNonCanonical(t *testing.T) {
	provider := &LocalDotfilesProvider{
		Base:     ResolvedDotfilesBase{RawPath: "/candidate/dotfiles", CanonicalPath: "/candidate/dotfiles", Source: DotfilesBaseSourceEnv},
		Resolver: DotfilesBaseResolver{HomeDir: func() (string, error) { return "/home/ada", nil }},
		Runner:   &fakeCommandRunner{},
		Stat:     func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		EvalSymlinks: fakeEval(map[string]string{
			"/home/ada":           "/home/ada",
			"/candidate/dotfiles": "/candidate/dotfiles",
		}),
	}

	diagnostic := provider.DotfilesBaseDiagnostic([]string{"bash"})
	if diagnostic.AttemptedCandidate != "/candidate/dotfiles" || diagnostic.Source != DotfilesBaseSourceEnv || diagnostic.CanonicalPath != "" || diagnostic.Cause == "" {
		t.Fatalf("diagnostic = %#v, want attempted env candidate without canonical path and with cause", diagnostic)
	}
	if !reflect.DeepEqual(diagnostic.Modules, []string{"bash"}) {
		t.Fatalf("diagnostic modules = %#v, want [bash]", diagnostic.Modules)
	}
}

func TestLocalDotfilesProviderBaseDiagnosticRetainsEnvCandidateWhenHomeLookupFails(t *testing.T) {
	homeErr := errors.New("home unavailable")
	provider := &LocalDotfilesProvider{Resolver: DotfilesBaseResolver{
		LookupEnv: func(string) (string, bool) { return "/work/dotfiles", true },
		HomeDir:   func() (string, error) { return "", homeErr },
	}}

	diagnostic := provider.DotfilesBaseDiagnostic([]string{"bash"})
	if got, want := diagnostic.Source, DotfilesBaseSourceEnv; got != want {
		t.Fatalf("Source = %q, want %q", got, want)
	}
	if got, want := diagnostic.AttemptedCandidate, "/work/dotfiles"; got != want {
		t.Fatalf("AttemptedCandidate = %q, want %q", got, want)
	}
	if got, want := diagnostic.Cause, "resolve home directory: home unavailable"; got != want {
		t.Fatalf("Cause = %q, want home lookup failure", diagnostic.Cause)
	}
}

func TestLocalDotfilesProviderValidationFailuresDoNotRun(t *testing.T) {
	tests := []struct {
		name    string
		modules []string
		setup   func(*LocalDotfilesProvider)
	}{
		{name: "empty modules", modules: nil},
		{name: "missing dotlink", modules: []string{"bash"}, setup: func(p *LocalDotfilesProvider) { p.Stat = statDirs("/repo", "/repo/bash") }},
		{name: "dotlink escapes", modules: []string{"bash"}, setup: func(p *LocalDotfilesProvider) {
			p.EvalSymlinks = fakeEval(map[string]string{"/repo/bin/dotlink": "/tmp/dotlink", "/repo/bash": "/repo/bash"})
		}},
		{name: "missing module", modules: []string{"missing"}},
		{name: "module escapes", modules: []string{"bash"}, setup: func(p *LocalDotfilesProvider) {
			p.EvalSymlinks = fakeEval(map[string]string{"/repo/bin/dotlink": "/repo/bin/dotlink", "/repo/bash": "/tmp/bash"})
		}},
		{name: "missing runner", modules: []string{"bash"}, setup: func(p *LocalDotfilesProvider) { p.Runner = nil }},
		{name: "leading dash", modules: []string{"-bad"}},
		{name: "separator", modules: []string{"bad/name"}},
		{name: "absolute", modules: []string{"/bad"}},
		{name: "dot", modules: []string{"."}},
		{name: "dotdot", modules: []string{".."}},
		{name: "bad character", modules: []string{"bad name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded}}
			provider := newFakeLocalProvider("/repo", runner)
			if tt.setup != nil {
				tt.setup(provider)
			}

			if err := provider.RunDotlink(context.Background(), tt.modules); err == nil {
				t.Fatal("RunDotlink() error = nil, want failure")
			}
			if len(runner.calls) != 0 {
				t.Fatalf("runner calls = %d, want 0", len(runner.calls))
			}
		})
	}
}

func TestLocalDotfilesProviderPrerequisiteFailuresRetainAttemptedCandidates(t *testing.T) {
	tests := []struct {
		name       string
		modules    []string
		setup      func(*LocalDotfilesProvider)
		wantKind   DotfilesPrerequisiteTargetKind
		wantTarget string
		wantErr    error
	}{
		{name: "missing runner", modules: []string{"bash"}, setup: func(p *LocalDotfilesProvider) { p.Stat = statDirs("/repo", "/repo/bash") }, wantKind: DotfilesPrerequisiteRunner, wantTarget: "/repo/bin/dotlink", wantErr: os.ErrNotExist},
		{name: "missing module", modules: []string{"missing"}, wantKind: DotfilesPrerequisiteModule, wantTarget: "/repo/missing", wantErr: os.ErrNotExist},
		{name: "escaping module", modules: []string{"bash"}, setup: func(p *LocalDotfilesProvider) {
			p.EvalSymlinks = fakeEval(map[string]string{"/home/ada": "/home/ada", "/repo": "/repo", "/repo/bin/dotlink": "/repo/bin/dotlink", "/repo/bash": "/tmp/bash"})
		}, wantKind: DotfilesPrerequisiteModule, wantTarget: "/repo/bash", wantErr: ErrDotfilesPathEscapes},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{}
			provider := newFakeLocalProvider("/repo", runner)
			if tt.setup != nil {
				tt.setup(provider)
			}

			_, err := provider.RunDotlinkReport(context.Background(), tt.modules)
			var failure *DotfilesFailure
			if !errors.As(err, &failure) || failure.Phase != DotfilesPhasePrerequisite || failure.PrerequisiteTarget == nil || failure.PrerequisiteTarget.Kind != tt.wantKind || failure.PrerequisiteTarget.AttemptedCandidate != tt.wantTarget || !errors.Is(err, tt.wantErr) {
				t.Fatalf("failure = %#v, error = %v; want prerequisite candidate %q and %v", failure, err, tt.wantTarget, tt.wantErr)
			}
			if len(runner.calls) != 0 {
				t.Fatalf("runner calls = %d, want 0", len(runner.calls))
			}
		})
	}
}

func TestLocalDotfilesProviderCommandFailureAndTimeout(t *testing.T) {
	tests := []struct {
		name   string
		result CommandResult
	}{
		{name: "failed", result: CommandResult{Status: CommandStatusFailed, Err: errors.New("exit 1")}},
		{name: "timed out", result: CommandResult{Status: CommandStatusTimedOut, Err: context.DeadlineExceeded}},
		{name: "not run", result: CommandResult{Status: CommandStatusNotRun}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{result: tt.result}
			provider := newFakeLocalProvider("/repo", runner)

			if err := provider.RunDotlink(context.Background(), []string{"bash"}); err == nil {
				t.Fatal("RunDotlink() error = nil, want failure")
			}
			if len(runner.calls) != 1 {
				t.Fatalf("runner calls = %d, want 1", len(runner.calls))
			}
		})
	}
}

func TestLocalDotfilesProviderEnsureModulesValidatesWithoutRunning(t *testing.T) {
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded}}
	provider := newFakeLocalProvider("/repo", runner)

	if err := provider.EnsureModules(context.Background(), []string{"bash"}); err != nil {
		t.Fatalf("EnsureModules() error = %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("runner calls = %d, want 0", len(runner.calls))
	}
}

func TestLocalDotfilesProviderUsesCanonicalInjectedBaseForCommand(t *testing.T) {
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded, Stdout: string(readDotlinkReportFixture(t, "all-changed.json"))}}
	provider := &LocalDotfilesProvider{
		Base:     ResolvedDotfilesBase{CanonicalPath: "/repo-link"},
		Resolver: DotfilesBaseResolver{HomeDir: func() (string, error) { return "/home/ada", nil }},
		Runner:   runner,
		Stat:     statDirs("/repo", "/repo/bin/dotlink", "/repo/bash"),
		EvalSymlinks: fakeEval(map[string]string{
			"/home/ada":         "/home/ada",
			"/repo-link":        "/repo",
			"/repo/bin/dotlink": "/repo/bin/dotlink",
			"/repo/bash":        "/repo/bash",
		}),
	}

	if err := provider.RunDotlink(context.Background(), []string{"bash"}); err != nil {
		t.Fatalf("RunDotlink() error = %v", err)
	}
	want := CommandRequest{Executable: "/repo/bin/dotlink", Args: []string{"link", "--report=json", "bash"}, Dir: "/repo", Timeout: DefaultDotlinkTimeout}
	if !reflect.DeepEqual(runner.calls, []CommandRequest{want}) {
		t.Fatalf("runner calls = %#v, want %#v", runner.calls, []CommandRequest{want})
	}
}

func TestLocalDotfilesProviderRejectsUnsafeInjectedBase(t *testing.T) {
	tests := []struct {
		name string
		base string
		stat func(string) (os.FileInfo, error)
	}{
		{name: "root", base: "/", stat: statDirs("/")},
		{name: "home", base: "/home/ada", stat: statDirs("/home/ada")},
		{name: "alias to home", base: "/home-link/ada", stat: statDirs("/home/ada")},
		{name: "relative", base: "relative", stat: statDirs("relative")},
		{name: "missing", base: "/missing", stat: func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }},
		{name: "non-directory", base: "/file", stat: func(string) (os.FileInfo, error) { return fakeFileInfo{dir: false}, nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded}}
			provider := &LocalDotfilesProvider{
				Base:     ResolvedDotfilesBase{CanonicalPath: tt.base},
				Resolver: DotfilesBaseResolver{HomeDir: func() (string, error) { return "/home/ada", nil }},
				Runner:   runner,
				Stat:     tt.stat,
				EvalSymlinks: func(path string) (string, error) {
					if path == "/home/ada" {
						return "/home/ada", nil
					}
					if tt.name == "alias to home" && path == "/home-link/ada" {
						return "/home/ada", nil
					}
					if path == tt.base {
						return tt.base, nil
					}
					return "", os.ErrNotExist
				},
			}

			if err := provider.RunDotlink(context.Background(), []string{"bash"}); err == nil {
				t.Fatal("RunDotlink() error = nil, want unsafe base failure")
			}
			if len(runner.calls) != 0 {
				t.Fatalf("runner calls = %d, want 0", len(runner.calls))
			}
		})
	}
}

func TestLocalDotfilesProviderRejectsForgedUnsafeExecutionContext(t *testing.T) {
	runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded}}
	provider := newFakeLocalProvider("/repo", runner)
	provider.EvalSymlinks = fakeEval(map[string]string{
		"/home/ada": "/home/ada",
		"/":         "/",
	})

	_, err := provider.RunDotlinkReportWithExecutionContext(context.Background(), []string{"bash"}, DotfilesExecutionContext{
		Base: ResolvedDotfilesBase{CanonicalPath: "/"},
	})
	if !errors.Is(err, ErrUnresolvedDotfiles) {
		t.Fatalf("RunDotlinkReportWithExecutionContext() error = %v, want rejected unvalidated base error", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("runner calls = %d, want 0", len(runner.calls))
	}
}

func TestLocalDotfilesProviderRejectsExecutionContextsWithoutMatchingValidationProof(t *testing.T) {
	tests := []struct {
		name    string
		context DotfilesExecutionContext
	}{
		{
			name:    "missing proof",
			context: DotfilesExecutionContext{Base: ResolvedDotfilesBase{CanonicalPath: "/repo"}},
		},
		{
			name: "mismatched proof",
			context: DotfilesExecutionContext{
				Base:          ResolvedDotfilesBase{CanonicalPath: "/repo"},
				validatedBase: "/other",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeCommandRunner{result: CommandResult{Status: CommandStatusSucceeded, Stdout: string(readDotlinkReportFixture(t, "all-changed.json"))}}
			provider := newFakeLocalProvider("/repo", runner)

			_, err := provider.RunDotlinkReportWithExecutionContext(context.Background(), []string{"bash"}, tt.context)
			if err == nil {
				t.Fatal("RunDotlinkReportWithExecutionContext() error = nil, want rejected context")
			}
			if len(runner.calls) != 0 {
				t.Fatalf("runner calls = %d, want 0", len(runner.calls))
			}
		})
	}
}

func newFakeLocalProvider(base string, runner CommandRunner) *LocalDotfilesProvider {
	home := "/home/ada"
	return &LocalDotfilesProvider{
		Base:     ResolvedDotfilesBase{CanonicalPath: base},
		Resolver: DotfilesBaseResolver{HomeDir: func() (string, error) { return home, nil }},
		Runner:   runner,
		Stat:     statDirs(base, base+"/bin/dotlink", base+"/bash", base+"/nvim"),
		EvalSymlinks: fakeEval(map[string]string{
			home:                  home,
			base:                  base,
			base + "/bin/dotlink": base + "/bin/dotlink",
			base + "/bash":        base + "/bash",
			base + "/nvim":        base + "/nvim",
		}),
		Timeout: DefaultDotlinkTimeout,
	}
}

func fakeEval(paths map[string]string) func(string) (string, error) {
	return func(path string) (string, error) {
		if got, ok := paths[path]; ok {
			return got, nil
		}
		return "", os.ErrNotExist
	}
}
