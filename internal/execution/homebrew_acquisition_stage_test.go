//go:build linux && (amd64 || arm64)

package execution

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHomebrewPinIsImmutable(t *testing.T) {
	if homebrewInstallerURL != "https://raw.githubusercontent.com/Homebrew/install/4b0227cf8416504142d23893368c2e1d211d5191/install.sh" {
		t.Fatalf("installer URL = %q, want reviewed literal commit URL", homebrewInstallerURL)
	}
	if homebrewInstallerDigest != "99287f194a8b3c9e6b0203a11a5fa54518be57209343e6bb954dec4635796d9d" {
		t.Fatalf("installer digest = %q, want reviewed SHA-256", homebrewInstallerDigest)
	}
	if err := validatePinnedDownload(homebrewInstallerURL, homebrewInstallerURL); err != nil {
		t.Fatalf("validatePinnedDownload() = %v, want nil", err)
	}
	for _, effectiveURL := range []string{"https://example.invalid/install.sh", homebrewInstallerURL + "?redirected"} {
		if err := validatePinnedDownload(homebrewInstallerURL, effectiveURL); err == nil {
			t.Fatalf("validatePinnedDownload(%q) = nil, want redirect rejection", effectiveURL)
		}
	}
}

func TestHomebrewStageIsPrivateAndRejectsUnsafeFiles(t *testing.T) {
	stage, err := newHomebrewStage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer stage.Close()
	info, err := os.Stat(stage.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("stage dir mode = %o, want 0700", info.Mode().Perm())
	}
	if err := os.WriteFile(stage.Path, []byte("installer"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := reopenOwnedStagedFile(stage.Path)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
	for _, tc := range []struct {
		name  string
		setup func(*testing.T, string)
	}{
		{"symlink", func(t *testing.T, path string) {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("/etc/passwd", path); err != nil {
				t.Fatal(err)
			}
		}},
		{"world-readable", func(t *testing.T, path string) {
			if err := os.Chmod(path, 0o644); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "stage")
			if err := os.WriteFile(path, []byte("installer"), 0o600); err != nil {
				t.Fatal(err)
			}
			tc.setup(t, path)
			if file, err := reopenOwnedStagedFile(path); err == nil {
				file.Close()
				t.Fatal("reopenOwnedStagedFile() succeeded for unsafe stage")
			}
		})
	}
}

func TestDownloadPinnedInstallerRejectsRedirectAndSubstitutedBytes(t *testing.T) {
	for _, tc := range []struct {
		name, effectiveURL, body string
		wantErr                  bool
	}{
		{"literal URL", homebrewInstallerURL, "reviewed bytes", false}, {"redirect", "https://example.invalid/install.sh", "reviewed bytes", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stage, err := newHomebrewStage(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer stage.Close()
			doer := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != homebrewInstallerURL {
					t.Fatalf("request URL = %q", req.URL)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tc.body)), Request: &http.Request{URL: mustParseURL(t, tc.effectiveURL)}}, nil
			})
			err = downloadPinnedInstaller(context.Background(), doer, stage)
			if (err != nil) != tc.wantErr {
				t.Fatalf("downloadPinnedInstaller() = %v, wantErr %t", err, tc.wantErr)
			}
			if !tc.wantErr {
				got, err := os.ReadFile(stage.Path)
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != tc.body {
					t.Fatalf("staged bytes = %q, want %q", got, tc.body)
				}
			}
		})
	}
}

func TestBashRequestUsesOnlyInheritedFDThree(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "sealed")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	req, files, err := bashRequestForSealedScript(&sealedScript{File: file, Sealed: true})
	if err != nil {
		t.Fatal(err)
	}
	if req.Executable != "/bin/bash" || len(req.Args) != 1 || req.Args[0] != "/proc/self/fd/3" {
		t.Fatalf("request = %#v, want /bin/bash /proc/self/fd/3", req)
	}
	if len(files) != 1 || files[0] != file {
		t.Fatalf("ExtraFiles = %#v, want sealed FD only", files)
	}
	for _, forbidden := range []string{"-c", "http", "|", "staged"} {
		if strings.Contains(strings.Join(append([]string{req.Executable}, req.Args...), " "), forbidden) {
			t.Fatalf("request contains forbidden %q", forbidden)
		}
	}
}

func TestExecuteSealedScriptRejectsUnsealedAndUsesExactFDThreeRequest(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "sealed")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	runner := &recordingCommandRunner{}
	if result := executeSealedScript(context.Background(), runner, &sealedScript{File: file}); result.Err == nil {
		t.Fatal("unsealed script execution succeeded")
	}
	if runner.calls != 0 {
		t.Fatalf("Bash calls = %d, want 0", runner.calls)
	}
	result := executeSealedScript(context.Background(), runner, &sealedScript{File: file, Sealed: true})
	if result.Status != CommandStatusSucceeded || runner.calls != 1 {
		t.Fatalf("result = %#v, calls = %d", result, runner.calls)
	}
	if runner.request.Executable != "/bin/bash" || strings.Join(runner.request.Args, " ") != "/proc/self/fd/3" || len(runner.request.ExtraFiles) != 1 || runner.request.ExtraFiles[0] != file {
		t.Fatalf("request = %#v", runner.request)
	}
}

func TestHomebrewAcquirerPreservesPrimaryErrorDuringCleanup(t *testing.T) {
	primary := errors.New("digest mismatch")
	err := preservePrimaryError(primary, errors.New("cleanup failed"))
	if !errors.Is(err, primary) || !strings.Contains(err.Error(), "cleanup failed") {
		t.Fatalf("error = %v", err)
	}
}

type recordingCommandRunner struct {
	calls   int
	request CommandRequest
}

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

type roundTripFunc func(*http.Request) (*http.Response, error)

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
func (r *recordingCommandRunner) RunCommand(_ context.Context, req CommandRequest) CommandResult {
	r.calls++
	r.request = req
	return CommandResult{Request: req, Status: CommandStatusSucceeded}
}
