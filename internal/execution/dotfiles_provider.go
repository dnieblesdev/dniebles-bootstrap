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
	ErrEmptyDotfileModules  = errors.New("dotfile module list is empty")
	ErrInvalidDotfileModule = errors.New("invalid dotfile module name")
	ErrMissingDotlinkRunner = errors.New("missing dotfiles command runner")
	ErrDotfilesPathEscapes  = errors.New("dotfiles path escapes base")
	ErrDotlinkCommandFailed = errors.New("dotlink command failed")
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

func (p *LocalDotfilesProvider) EnsureModules(_ context.Context, modules []string) error {
	base, err := p.resolvedBase()
	if err != nil {
		return err
	}
	return p.validateRepo(base.CanonicalPath, modules)
}

func (p *LocalDotfilesProvider) RunDotlink(ctx context.Context, modules []string) error {
	base, err := p.resolvedBase()
	if err != nil {
		return err
	}
	if err := p.validateRepo(base.CanonicalPath, modules); err != nil {
		return err
	}
	if p.Runner == nil {
		return ErrMissingDotlinkRunner
	}

	timeout := p.Timeout
	if timeout <= 0 {
		timeout = DefaultDotlinkTimeout
	}
	args := make([]string, 0, len(modules)+1)
	args = append(args, "link")
	args = append(args, modules...)
	request := CommandRequest{
		Executable: filepath.Join(base.CanonicalPath, "bin", "dotlink"),
		Args:       args,
		Dir:        base.CanonicalPath,
		Timeout:    timeout,
	}
	result := p.Runner.RunCommand(ctx, request)
	if result.Status == CommandStatusSucceeded {
		return nil
	}
	return fmt.Errorf("%w: %s: %w", ErrDotlinkCommandFailed, result.Status, commandResultError(result))
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
