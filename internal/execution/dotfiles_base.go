package execution

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const dotfilesEnvVar = "DBOOTSTRAP_DOTFILES_DIR"

var (
	ErrEmptyDotfilesBase  = errors.New("dotfiles base path is empty")
	ErrUnsafeDotfilesBase = errors.New("unsafe dotfiles base path")
	ErrDotfilesBaseNotDir = errors.New("dotfiles base path is not a directory")
	ErrUnresolvedDotfiles = errors.New("dotfiles base path could not be resolved")
)

type DotfilesBaseSource string

const (
	DotfilesBaseSourceEnv  DotfilesBaseSource = "env"
	DotfilesBaseSourceHome DotfilesBaseSource = "home"
)

type ResolvedDotfilesBase struct {
	RawPath       string
	CanonicalPath string
	Source        DotfilesBaseSource
}

type DotfilesBaseResolver struct {
	LookupEnv    func(string) (string, bool)
	HomeDir      func() (string, error)
	EvalSymlinks func(string) (string, error)
	Stat         func(string) (os.FileInfo, error)
}

func ResolveDotfilesBasePath() (ResolvedDotfilesBase, error) {
	return DotfilesBaseResolver{}.Resolve()
}

func (r DotfilesBaseResolver) Resolve() (ResolvedDotfilesBase, error) {
	base, _, err := r.ResolveWithDiagnostic(nil)
	return base, err
}

// ResolveWithDiagnostic resolves the base once while retaining the safe context
// needed to report failures without presenting an unvalidated path as canonical.
func (r DotfilesBaseResolver) ResolveWithDiagnostic(modules []string) (ResolvedDotfilesBase, DotfilesBaseDiagnostic, error) {
	diagnostic := DotfilesBaseDiagnostic{Modules: append([]string(nil), modules...)}
	lookupEnv := r.LookupEnv
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}
	homeDir := r.HomeDir
	if homeDir == nil {
		homeDir = os.UserHomeDir
	}
	evalSymlinks := r.EvalSymlinks
	if evalSymlinks == nil {
		evalSymlinks = filepath.EvalSymlinks
	}
	stat := r.Stat
	if stat == nil {
		stat = os.Stat
	}

	raw, source, envSet := "", DotfilesBaseSourceHome, false
	diagnostic.Source = source
	if value, ok := lookupEnv(dotfilesEnvVar); ok {
		raw, source, envSet = value, DotfilesBaseSourceEnv, true
		diagnostic.AttemptedCandidate = raw
		diagnostic.Source = source
		if raw == "" {
			diagnostic.Cause = ErrEmptyDotfilesBase.Error()
			return ResolvedDotfilesBase{}, diagnostic, ErrEmptyDotfilesBase
		}
	}

	home, err := homeDir()
	if err != nil {
		err = fmt.Errorf("resolve home directory: %w", err)
		diagnostic.Cause = err.Error()
		return ResolvedDotfilesBase{}, diagnostic, err
	}
	if !envSet {
		raw, source = filepath.Join(home, ".dotfiles"), DotfilesBaseSourceHome
		diagnostic.AttemptedCandidate = raw
		diagnostic.Source = source
	}

	canonicalHome, err := canonicalizeDotfilesPath(home, evalSymlinks)
	if err != nil {
		err = fmt.Errorf("resolve home directory %q: %w", home, err)
		diagnostic.Cause = err.Error()
		return ResolvedDotfilesBase{}, diagnostic, err
	}

	canonical, err := canonicalizeDotfilesPath(raw, evalSymlinks)
	if err != nil {
		err = fmt.Errorf("resolve dotfiles base %q: %w", raw, err)
		diagnostic.Cause = err.Error()
		return ResolvedDotfilesBase{}, diagnostic, err
	}
	if err := validateDotfilesBase(canonical, canonicalHome, stat); err != nil {
		diagnostic.Cause = err.Error()
		return ResolvedDotfilesBase{}, diagnostic, err
	}
	base := ResolvedDotfilesBase{RawPath: raw, CanonicalPath: filepath.Clean(canonical), Source: source}
	diagnostic.CanonicalPath = base.CanonicalPath
	return base, diagnostic, nil
}

func canonicalizeDotfilesPath(path string, evalSymlinks func(string) (string, error)) (string, error) {
	canonical, err := evalSymlinks(path)
	if err != nil || canonical == "" {
		if err == nil {
			err = ErrUnresolvedDotfiles
		}
		return "", err
	}
	return filepath.Clean(canonical), nil
}

func validateDotfilesBase(path string, canonicalHome string, stat func(string) (os.FileInfo, error)) error {
	clean := filepath.Clean(path)
	cleanHome := filepath.Clean(canonicalHome)
	if clean == "" || !filepath.IsAbs(clean) || clean == string(filepath.Separator) || clean == cleanHome {
		return fmt.Errorf("%w: %q", ErrUnsafeDotfilesBase, path)
	}
	info, err := stat(clean)
	if err != nil {
		return fmt.Errorf("stat dotfiles base %q: %w", clean, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %q", ErrDotfilesBaseNotDir, clean)
	}
	return nil
}
