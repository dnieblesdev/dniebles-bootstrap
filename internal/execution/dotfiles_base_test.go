package execution

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveDotfilesBasePathSourcesAndCanonicalization(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		envSet    bool
		home      string
		canonical string
		wantRaw   string
		wantSrc   DotfilesBaseSource
	}{
		{name: "env override", envValue: "/link/dots", envSet: true, home: "/home/ada", canonical: "/safe/dots", wantRaw: "/link/dots", wantSrc: DotfilesBaseSourceEnv},
		{name: "home default", envSet: false, home: "/home/ada", canonical: "/safe/home-dots", wantRaw: filepath.Join("/home/ada", ".dotfiles"), wantSrc: DotfilesBaseSourceHome},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := fakeDotfilesBaseResolver(tt.envValue, tt.envSet, tt.home)
			resolver.EvalSymlinks = fakeEval(map[string]string{tt.home: tt.home, tt.wantRaw: tt.canonical})
			resolver.Stat = statDirs(tt.canonical, tt.home)

			got, err := resolver.Resolve()
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if got.RawPath != tt.wantRaw || got.CanonicalPath != tt.canonical || got.Source != tt.wantSrc {
				t.Fatalf("Resolve() = %#v, want raw=%q canonical=%q source=%q", got, tt.wantRaw, tt.canonical, tt.wantSrc)
			}
		})
	}
}

func TestResolveDotfilesBasePathRejectsUnsafeWithoutFallback(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		envSet    bool
		home      string
		canonical string
		stat      func(string) (os.FileInfo, error)
		wantEvals []string
	}{
		{name: "empty env", envValue: "", envSet: true, home: "/home/ada", stat: statDirs("/home/ada/.dotfiles", "/home/ada"), wantEvals: nil},
		{name: "eval failure", envValue: "/bad", envSet: true, home: "/home/ada", canonical: "", stat: statDirs("/home/ada/.dotfiles", "/home/ada"), wantEvals: []string{"/home/ada", "/bad"}},
		{name: "relative canonical", envValue: "/dots", envSet: true, home: "/home/ada", canonical: "relative", stat: statDirs("relative", "/home/ada"), wantEvals: []string{"/home/ada", "/dots"}},
		{name: "missing", envValue: "/missing", envSet: true, home: "/home/ada", canonical: "/missing", stat: func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }, wantEvals: []string{"/home/ada", "/missing"}},
		{name: "non-directory", envValue: "/file", envSet: true, home: "/home/ada", canonical: "/file", stat: func(string) (os.FileInfo, error) { return fakeFileInfo{dir: false}, nil }, wantEvals: []string{"/home/ada", "/file"}},
		{name: "root", envValue: "/", envSet: true, home: "/home/ada", canonical: "/", stat: statDirs("/", "/home/ada"), wantEvals: []string{"/home/ada", "/"}},
		{name: "home itself", envValue: "/home/ada", envSet: true, home: "/home/ada", canonical: "/home/ada", stat: statDirs("/home/ada"), wantEvals: []string{"/home/ada", "/home/ada"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var evals []string
			resolver := fakeDotfilesBaseResolver(tt.envValue, tt.envSet, tt.home)
			resolver.EvalSymlinks = func(path string) (string, error) {
				evals = append(evals, path)
				if path == tt.home {
					return tt.home, nil
				}
				if tt.name == "eval failure" {
					return "", errors.New("unresolved")
				}
				return tt.canonical, nil
			}
			resolver.Stat = tt.stat

			if _, err := resolver.Resolve(); err == nil {
				t.Fatal("Resolve() error = nil, want failure")
			}
			if len(evals) != len(tt.wantEvals) {
				t.Fatalf("EvalSymlinks calls = %#v, want %#v", evals, tt.wantEvals)
			}
			for i := range evals {
				if evals[i] != tt.wantEvals[i] {
					t.Fatalf("EvalSymlinks calls = %#v, want %#v", evals, tt.wantEvals)
				}
			}
		})
	}
}

func TestResolveDotfilesBasePathRejectsCanonicalHomeAlias(t *testing.T) {
	resolver := fakeDotfilesBaseResolver("/env/dots", true, "/home-link/ada")
	resolver.EvalSymlinks = fakeEval(map[string]string{
		"/home-link/ada": "/real-home/ada",
		"/env/dots":      "/real-home/ada",
	})
	resolver.Stat = statDirs("/real-home/ada")

	if _, err := resolver.Resolve(); err == nil {
		t.Fatal("Resolve() error = nil, want canonical home rejection")
	}
}

func TestResolveWithDiagnosticReportsHomeSourceWhenHomeLookupFails(t *testing.T) {
	homeErr := errors.New("home unavailable")
	resolver := DotfilesBaseResolver{
		LookupEnv: func(string) (string, bool) { return "", false },
		HomeDir:   func() (string, error) { return "", homeErr },
	}

	_, diagnostic, err := resolver.ResolveWithDiagnostic([]string{"bash"})
	if !errors.Is(err, homeErr) {
		t.Fatalf("ResolveWithDiagnostic() error = %v, want %v", err, homeErr)
	}
	if diagnostic.Source != DotfilesBaseSourceHome {
		t.Fatalf("diagnostic source = %q, want %q", diagnostic.Source, DotfilesBaseSourceHome)
	}
	if diagnostic.AttemptedCandidate != "" || diagnostic.CanonicalPath != "" {
		t.Fatalf("diagnostic paths = %#v, want no candidate or canonical path", diagnostic)
	}
	if diagnostic.Cause != "resolve home directory: home unavailable" {
		t.Fatalf("diagnostic cause = %q", diagnostic.Cause)
	}
}

func TestResolveWithDiagnosticRetainsAttemptedIdentityAndFilesystemCause(t *testing.T) {
	pathErr := &os.PathError{Op: "stat", Path: "/missing", Err: os.ErrNotExist}
	resolver := fakeDotfilesBaseResolver("/missing", true, "/home/ada")
	resolver.EvalSymlinks = fakeEval(map[string]string{
		"/home/ada": "/home/ada",
		"/missing":  "/missing",
	})
	resolver.Stat = func(string) (os.FileInfo, error) { return nil, pathErr }

	base, diagnostic, err := resolver.ResolveWithDiagnostic([]string{"bash", "nvim"})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ResolveWithDiagnostic() error = %v, want wrapped not-exist", err)
	}
	var gotPathErr *os.PathError
	if !errors.As(err, &gotPathErr) || gotPathErr != pathErr {
		t.Fatalf("ResolveWithDiagnostic() error = %v, want original PathError", err)
	}
	if base != (ResolvedDotfilesBase{}) {
		t.Fatalf("base = %#v, want zero unresolved base", base)
	}
	if diagnostic.Source != DotfilesBaseSourceEnv || diagnostic.AttemptedCandidate != "/missing" || diagnostic.CanonicalPath != "" {
		t.Fatalf("diagnostic = %#v, want attempted env candidate only", diagnostic)
	}
	if got, want := diagnostic.Modules, []string{"bash", "nvim"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("diagnostic modules = %#v, want %#v", got, want)
	}
}

func fakeDotfilesBaseResolver(envValue string, envSet bool, home string) DotfilesBaseResolver {
	return DotfilesBaseResolver{
		LookupEnv: func(string) (string, bool) { return envValue, envSet },
		HomeDir:   func() (string, error) { return home, nil },
	}
}

func statDirs(paths ...string) func(string) (os.FileInfo, error) {
	allowed := map[string]bool{}
	for _, path := range paths {
		allowed[path] = true
	}
	return func(path string) (os.FileInfo, error) {
		if !allowed[path] {
			return nil, os.ErrNotExist
		}
		return fakeFileInfo{dir: true}, nil
	}
}

type fakeFileInfo struct{ dir bool }

func (f fakeFileInfo) Name() string { return "fake" }
func (f fakeFileInfo) Size() int64  { return 0 }
func (f fakeFileInfo) Mode() os.FileMode {
	if f.dir {
		return os.ModeDir | 0o700
	}
	return 0o600
}
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return f.dir }
func (f fakeFileInfo) Sys() any           { return nil }
