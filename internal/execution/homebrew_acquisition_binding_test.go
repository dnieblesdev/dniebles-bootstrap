//go:build linux && (amd64 || arm64)

package execution

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func stagedFile(t *testing.T, contents string) *os.File {
	t.Helper()
	staged, err := os.CreateTemp(t.TempDir(), "staged")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := staged.WriteString(contents); err != nil {
		t.Fatal(err)
	}
	if _, err := staged.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	return staged
}

func TestBindVerifiedStreamsExactBytesAndHandlesPartialWrites(t *testing.T) {
	staged := stagedFile(t, "approved-bytes")
	defer staged.Close()
	want := sha256.Sum256([]byte("approved-bytes"))
	sealed, err := bindVerified(staged, want, binderHooks{maxWrite: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer sealed.Close()
	got, err := io.ReadAll(sealed.File)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("approved-bytes")) || !sealed.Sealed {
		t.Fatalf("memfd bytes = %q, sealed = %t", got, sealed.Sealed)
	}
}

func TestBindVerifiedHashesBytesWrittenAfterStagedMutation(t *testing.T) {
	staged := stagedFile(t, "firstoriginal!")
	defer staged.Close()
	mutated, wantBytes := false, []byte("firstchanged!!")
	sealed, err := bindVerified(staged, sha256.Sum256(wantBytes), binderHooks{readSize: 5, afterRead: func() {
		if mutated {
			return
		}
		mutated = true
		if _, err := staged.WriteAt([]byte("changed!!"), 5); err != nil {
			t.Fatal(err)
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer sealed.Close()
	got, err := io.ReadAll(sealed.File)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, wantBytes) {
		t.Fatalf("memfd bytes = %q, want %q", got, wantBytes)
	}
}

func TestBindVerifiedMutationAndFailuresNeverReachBash(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hooks binderHooks
		want  [32]byte
	}{
		{"digest mismatch", binderHooks{}, sha256.Sum256([]byte("different-approved-bytes"))},
		{"read failure", binderHooks{readErr: errors.New("read failed")}, sha256.Sum256([]byte("mutable-bytes"))},
		{"write failure", binderHooks{writeErr: errors.New("write failed")}, sha256.Sum256([]byte("mutable-bytes"))},
		{"hash failure", binderHooks{hashErr: errors.New("hash failed")}, sha256.Sum256([]byte("mutable-bytes"))},
		{"rewind failure", binderHooks{seekErr: errors.New("rewind failed")}, sha256.Sum256([]byte("mutable-bytes"))},
		{"seal failure", binderHooks{sealErr: errors.New("seal failed")}, sha256.Sum256([]byte("mutable-bytes"))},
		{"seal validation failure", binderHooks{validateErr: errors.New("validation failed")}, sha256.Sum256([]byte("mutable-bytes"))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			staged := stagedFile(t, "mutable-bytes")
			defer staged.Close()
			sealed, err := bindVerified(staged, tc.want, tc.hooks)
			if err == nil || sealed != nil {
				if sealed != nil {
					sealed.Close()
				}
				t.Fatalf("sealed = %#v, err = %v", sealed, err)
			}
		})
	}
}

func TestHomebrewAcquirerNeverDispatchesPackages(t *testing.T) {
	stage, err := newHomebrewStage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	runner := &recordingCommandRunner{}
	brew := stagedFile(t, "brew")
	if err := brew.Chmod(0o700); err != nil {
		t.Fatal(err)
	}
	acquirer := NewHomebrewAcquirer(HomebrewAcquirerDependencies{
		NewStage: func(string) (*homebrewStage, error) { return stage, nil },
		Download: func(_ context.Context, _ httpDoer, staged *homebrewStage) error {
			return os.WriteFile(staged.Path, []byte("reviewed installer"), 0o600)
		},
		Bind: func(*os.File, [32]byte, binderHooks) (*sealedScript, error) {
			return &sealedScript{File: brew, Sealed: true}, nil
		},
		Runner:   runner,
		LookPath: func(string) (string, error) { return brew.Name(), nil },
		Stat: func(path string) (os.FileInfo, error) {
			if path == homebrewDefaultBinary {
				return nil, os.ErrNotExist
			}
			return os.Stat(path)
		},
	})
	result := acquirer.Acquire(context.Background())
	if result.Err != nil || !result.Acquired || result.PackageDispatchAllowed {
		t.Fatalf("result = %#v", result)
	}
	if runner.calls != 2 || runner.request.Executable != brew.Name() || len(runner.request.Args) != 1 || runner.request.Args[0] != "--version" {
		t.Fatalf("runner calls = %d, request = %#v", runner.calls, runner.request)
	}
}

func TestHomebrewAcquirerFailurePaths(t *testing.T) {
	for _, tc := range []struct {
		name          string
		downloadErr   error
		commandStatus []CommandStatus
		wantError     string
		wantBinds     int
		wantCommands  int
	}{
		{name: "download failure stops before binding", downloadErr: errors.New("download failed"), wantError: "download Homebrew installer", wantBinds: 0, wantCommands: 0},
		{name: "installer execution failure stops before lookup", commandStatus: []CommandStatus{CommandStatusFailed}, wantError: "execute Homebrew installer", wantBinds: 1, wantCommands: 1},
		{name: "revalidation failure is terminal", commandStatus: []CommandStatus{CommandStatusSucceeded, CommandStatusFailed}, wantError: "revalidate Homebrew", wantBinds: 1, wantCommands: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stage, err := newHomebrewStage(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			sealed := stagedFile(t, "verified installer")
			defer sealed.Close()
			brew := stagedFile(t, "brew")
			defer brew.Close()
			if err := brew.Chmod(0o700); err != nil {
				t.Fatal(err)
			}
			downloads, binds := 0, 0
			runner := &scriptedCommandRunner{statuses: tc.commandStatus}
			acquirer := NewHomebrewAcquirer(HomebrewAcquirerDependencies{
				NewStage: func(string) (*homebrewStage, error) { return stage, nil },
				Download: func(_ context.Context, _ httpDoer, staged *homebrewStage) error {
					downloads++
					if tc.downloadErr != nil {
						return tc.downloadErr
					}
					return os.WriteFile(staged.Path, []byte("reviewed installer"), 0o600)
				},
				Bind: func(*os.File, [32]byte, binderHooks) (*sealedScript, error) {
					binds++
					return &sealedScript{File: sealed, Sealed: true}, nil
				},
				Runner:   runner,
				LookPath: func(string) (string, error) { return brew.Name(), nil },
				Stat: func(path string) (os.FileInfo, error) {
					if path == homebrewDefaultBinary {
						return nil, os.ErrNotExist
					}
					return os.Stat(path)
				},
			})

			result := acquirer.Acquire(context.Background())
			if result.Err == nil || !strings.Contains(result.Err.Error(), tc.wantError) || result.Acquired {
				t.Fatalf("result = %#v, want %q failure", result, tc.wantError)
			}
			if downloads != 1 || binds != tc.wantBinds || runner.calls != tc.wantCommands {
				t.Fatalf("downloads=%d binds=%d commands=%d", downloads, binds, runner.calls)
			}
		})
	}
}

type scriptedCommandRunner struct {
	calls    int
	statuses []CommandStatus
}

func (r *scriptedCommandRunner) RunCommand(_ context.Context, request CommandRequest) CommandResult {
	status := r.statuses[r.calls]
	r.calls++
	return CommandResult{Request: request, Status: status, Err: errors.New("command failed")}
}
