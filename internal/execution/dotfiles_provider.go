package execution

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const DefaultDotlinkTimeout = 2 * time.Minute

var (
	ErrEmptyDotfileModules       = errors.New("dotfile module list is empty")
	ErrInvalidDotfileModule      = errors.New("invalid dotfile module name")
	ErrMissingDotlinkRunner      = errors.New("missing dotfiles command runner")
	ErrDotfilesPathEscapes       = errors.New("dotfiles path escapes base")
	ErrDotlinkCommandFailed      = errors.New("dotlink command failed")
	ErrInconsistentDotlinkReport = errors.New("inconsistent dotlink command and report status")
	ErrDotlinkReportUnavailable  = errors.New("dotlink report unavailable")
)

var dotfileModuleNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type LocalDotfilesProvider struct {
	Base         ResolvedDotfilesBase
	Resolver     DotfilesBaseResolver
	Runner       CommandRunner
	Stat         func(string) (os.FileInfo, error)
	EvalSymlinks func(string) (string, error)
	Timeout      time.Duration
}

func NewLocalDotfilesProvider(runner CommandRunner, resolver DotfilesBaseResolver) *LocalDotfilesProvider {
	return &LocalDotfilesProvider{Runner: runner, Resolver: resolver, Timeout: DefaultDotlinkTimeout}
}

func (p *LocalDotfilesProvider) DotfilesBase() (ResolvedDotfilesBase, error) {
	context := p.ResolveDotfilesExecutionContext(nil)
	return context.Base, context.Err
}

// DotfilesBaseDiagnostic returns safe resolution context for a selected module
// set. It never calls the command runner. A canonical path is exposed only when
// the base has passed canonicalization and safety validation.
func (p *LocalDotfilesProvider) DotfilesBaseDiagnostic(modules []string) DotfilesBaseDiagnostic {
	return p.ResolveDotfilesExecutionContext(modules).Diagnostic
}

func (p *LocalDotfilesProvider) ResolveDotfilesExecutionContext(modules []string) DotfilesExecutionContext {
	if p.Base.CanonicalPath != "" {
		diagnostic := DotfilesBaseDiagnostic{Source: p.Base.Source, AttemptedCandidate: p.Base.RawPath, Modules: append([]string(nil), modules...)}
		if diagnostic.AttemptedCandidate == "" {
			diagnostic.AttemptedCandidate = p.Base.CanonicalPath
		}
		base, err := p.resolvedBase()
		if err != nil {
			diagnostic.Cause = err.Error()
			return DotfilesExecutionContext{Diagnostic: diagnostic, Err: err}
		}
		diagnostic.CanonicalPath = base.CanonicalPath
		return DotfilesExecutionContext{Base: base, Diagnostic: diagnostic, validatedBase: base.CanonicalPath}
	}
	base, diagnostic, err := p.Resolver.ResolveWithDiagnostic(modules)
	return DotfilesExecutionContext{Base: base, Diagnostic: diagnostic, Err: err, validatedBase: base.CanonicalPath}
}

func (p *LocalDotfilesProvider) EnsureModules(_ context.Context, modules []string) error {
	base, err := p.resolvedBase()
	if err != nil {
		return err
	}
	return p.validateRepo(base.CanonicalPath, modules)
}

// RunDotlink preserves the legacy provider seam while consuming the report.
// Callers that need validated report data use RunDotlinkReport.
func (p *LocalDotfilesProvider) RunDotlink(ctx context.Context, modules []string) error {
	report, err := p.RunDotlinkReport(ctx, modules)
	if err != nil {
		return err
	}
	if report.Status == DotlinkReportStatusFailed {
		return ErrDotlinkCommandFailed
	}
	return nil
}

// RunDotlinkReport executes dotlink once and returns only a validated report.
// Stderr is deliberately never inspected.
func (p *LocalDotfilesProvider) RunDotlinkReport(ctx context.Context, modules []string) (DotlinkLinkReport, error) {
	return p.RunDotlinkReportWithExecutionContext(ctx, modules, p.ResolveDotfilesExecutionContext(modules))
}

func (p *LocalDotfilesProvider) RunDotlinkReportWithExecutionContext(ctx context.Context, modules []string, baseContext DotfilesExecutionContext) (DotlinkLinkReport, error) {
	if baseContext.Err != nil {
		return DotlinkLinkReport{}, &DotfilesFailure{Phase: DotfilesPhaseResolution, BaseSnapshot: &baseContext.Diagnostic, PrerequisiteErr: baseContext.Err}
	}
	base := baseContext.Base
	if baseContext.validatedBase == "" || baseContext.validatedBase != base.CanonicalPath {
		return DotlinkLinkReport{}, ErrUnresolvedDotfiles
	}
	runnerCandidate := filepath.Join(base.CanonicalPath, "bin", "dotlink")
	if failure := p.validatePrerequisites(base.CanonicalPath, modules, baseContext.Diagnostic, runnerCandidate); failure != nil {
		return DotlinkLinkReport{}, failure
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = DefaultDotlinkTimeout
	}
	args := make([]string, 0, len(modules)+2)
	args = append(args, "link", "--report=json")
	args = append(args, modules...)
	request := CommandRequest{
		Executable: runnerCandidate,
		Args:       args,
		Dir:        base.CanonicalPath,
		Timeout:    timeout,
	}
	if p.Runner == nil {
		return DotlinkLinkReport{}, &DotfilesFailure{Phase: DotfilesPhaseCommandExecution, Executable: request.Executable, Runner: "CommandRunner", Command: request, ReportStatus: "unavailable", BaseSnapshot: &baseContext.Diagnostic, ExecutionErr: errors.Join(ErrDotlinkCommandFailed, ErrMissingDotlinkRunner)}
	}
	result := p.Runner.RunCommand(ctx, request)
	failure := dotfilesCommandFailure(baseContext.Diagnostic, request, result)
	if strings.TrimSpace(result.Stdout) == "" {
		failure.Phase = DotfilesPhaseReportValidation
		failure.ParseErr = errors.Join(ErrInvalidDotlinkReport, ErrDotlinkReportUnavailable)
		return DotlinkLinkReport{}, failure
	}
	report, parseErr := ParseDotlinkLinkReport([]byte(result.Stdout), modules)
	if parseErr != nil {
		failure.Phase = DotfilesPhaseReportValidation
		failure.ParseErr = errors.Join(ErrInvalidDotlinkReport, parseErr)
		return DotlinkLinkReport{}, failure
	}
	report.CommandStatus = result.Status
	if (report.Status == DotlinkReportStatusSuccess && result.Status != CommandStatusSucceeded) ||
		(report.Status == DotlinkReportStatusFailed && result.Status != CommandStatusFailed) {
		failure.Phase = DotfilesPhaseReportValidation
		failure.ParseErr = errors.Join(ErrInvalidDotlinkReport, ErrInconsistentDotlinkReport)
		return DotlinkLinkReport{}, failure
	}
	if failure.ExecutionErr != nil {
		failure.ReportStatus = report.Status
		return report, failure
	}
	return report, nil
}

func dotfilesCommandFailure(base DotfilesBaseDiagnostic, request CommandRequest, result CommandResult) *DotfilesFailure {
	failure := &DotfilesFailure{Phase: DotfilesPhaseCommandExecution, Executable: request.Executable, Runner: "CommandRunner", Command: request, Stderr: sanitizeDotlinkStderr(result.Stderr), ReportStatus: "unavailable", BaseSnapshot: &base}
	if result.Status != CommandStatusSucceeded {
		cause := result.Err
		if cause == nil {
			cause = fmt.Errorf("command status %s", result.Status)
		}
		failure.ExecutionErr = errors.Join(ErrDotlinkCommandFailed, cause)
		if result.ExitCode != 0 {
			code := result.ExitCode
			failure.ExitCode = &code
		}
	}
	return failure
}

func (p *LocalDotfilesProvider) validatePrerequisites(base string, modules []string, diagnostic DotfilesBaseDiagnostic, runnerCandidate string) *DotfilesFailure {
	if err := p.validateContainedExistingPath(base, runnerCandidate, false); err != nil {
		return prerequisiteFailure(diagnostic, DotfilesPrerequisiteRunner, runnerCandidate, fmt.Errorf("validate dotlink: %w", err))
	}
	if len(modules) == 0 {
		return prerequisiteFailure(diagnostic, DotfilesPrerequisiteModule, "", ErrEmptyDotfileModules)
	}
	for _, module := range modules {
		candidate := filepath.Join(base, module)
		if err := validateDotfileModuleName(module); err != nil {
			return prerequisiteFailure(diagnostic, DotfilesPrerequisiteModule, candidate, err)
		}
		if err := p.validateContainedExistingPath(base, candidate, true); err != nil {
			return prerequisiteFailure(diagnostic, DotfilesPrerequisiteModule, candidate, fmt.Errorf("validate module %q: %w", module, err))
		}
	}
	return nil
}

func prerequisiteFailure(base DotfilesBaseDiagnostic, kind DotfilesPrerequisiteTargetKind, candidate string, cause error) *DotfilesFailure {
	return &DotfilesFailure{
		Phase:              DotfilesPhasePrerequisite,
		BaseSnapshot:       &base,
		PrerequisiteTarget: &DotfilesPrerequisiteTarget{Kind: kind, AttemptedCandidate: candidate},
		PrerequisiteErr:    cause,
	}
}

const (
	maxDotlinkStderrBytes  = 4096
	dotlinkStderrTruncated = "...[truncated]"
)

func sanitizeDotlinkStderr(stderr string) string {
	tokens := make([]string, 0, len(stderr))
	total := 0
	for _, r := range stderr {
		token := string(r)
		if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
			token = fmt.Sprintf(`\x%02x`, r)
		}
		tokens = append(tokens, token)
		total += len(token)
	}
	if total <= maxDotlinkStderrBytes {
		return strings.Join(tokens, "")
	}
	var out strings.Builder
	for _, token := range tokens {
		if out.Len()+len(token)+len(dotlinkStderrTruncated) > maxDotlinkStderrBytes {
			break
		}
		out.WriteString(token)
	}
	out.WriteString(dotlinkStderrTruncated)
	return out.String()
}

func (p *LocalDotfilesProvider) resolvedBase() (ResolvedDotfilesBase, error) {
	if p.Base.CanonicalPath != "" {
		canonical, err := p.validateResolvedBase(p.Base.CanonicalPath)
		if err != nil {
			return ResolvedDotfilesBase{}, err
		}
		p.Base.CanonicalPath = canonical
		return p.Base, nil
	}
	return p.Resolver.Resolve()
}

func (p *LocalDotfilesProvider) validateResolvedBase(path string) (string, error) {
	evalSymlinks := p.EvalSymlinks
	if evalSymlinks == nil {
		evalSymlinks = filepath.EvalSymlinks
	}
	stat := p.Stat
	if stat == nil {
		stat = os.Stat
	}
	homeDir := p.Resolver.HomeDir
	if homeDir == nil {
		homeDir = os.UserHomeDir
	}
	home, err := homeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	canonicalHome, err := canonicalizeDotfilesPath(home, evalSymlinks)
	if err != nil {
		return "", fmt.Errorf("resolve home directory %q: %w", home, err)
	}
	canonicalBase, err := canonicalizeDotfilesPath(path, evalSymlinks)
	if err != nil {
		return "", fmt.Errorf("resolve dotfiles base %q: %w", path, err)
	}
	if err := validateDotfilesBase(canonicalBase, canonicalHome, stat); err != nil {
		return "", err
	}
	return canonicalBase, nil
}

func (p *LocalDotfilesProvider) validateRepo(base string, modules []string) error {
	if len(modules) == 0 {
		return ErrEmptyDotfileModules
	}
	if err := p.validateContainedExistingPath(base, filepath.Join(base, "bin", "dotlink"), false); err != nil {
		return fmt.Errorf("validate dotlink: %w", err)
	}
	for _, module := range modules {
		if err := validateDotfileModuleName(module); err != nil {
			return err
		}
		if err := p.validateContainedExistingPath(base, filepath.Join(base, module), true); err != nil {
			return fmt.Errorf("validate module %q: %w", module, err)
		}
	}
	return nil
}

func (p *LocalDotfilesProvider) validateContainedExistingPath(base, candidate string, requireDir bool) error {
	evalSymlinks := p.EvalSymlinks
	if evalSymlinks == nil {
		evalSymlinks = filepath.EvalSymlinks
	}
	stat := p.Stat
	if stat == nil {
		stat = os.Stat
	}

	canonical, err := evalSymlinks(candidate)
	if err != nil || canonical == "" {
		if err == nil {
			err = ErrUnresolvedDotfiles
		}
		return err
	}
	if !pathContained(base, canonical) {
		return fmt.Errorf("%w: %q", ErrDotfilesPathEscapes, canonical)
	}
	info, err := stat(canonical)
	if err != nil {
		return err
	}
	if requireDir && !info.IsDir() {
		return ErrDotfilesBaseNotDir
	}
	return nil
}

func validateDotfileModuleName(module string) error {
	if module == "" || module == "." || module == ".." || strings.HasPrefix(module, "-") || filepath.IsAbs(module) || strings.ContainsAny(module, `/\`) || !dotfileModuleNamePattern.MatchString(module) {
		return fmt.Errorf("%w: %q", ErrInvalidDotfileModule, module)
	}
	return nil
}

func pathContained(base, candidate string) bool {
	cleanBase := filepath.Clean(base)
	cleanCandidate := filepath.Clean(candidate)
	if cleanCandidate == cleanBase {
		return true
	}
	rel, err := filepath.Rel(cleanBase, cleanCandidate)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
