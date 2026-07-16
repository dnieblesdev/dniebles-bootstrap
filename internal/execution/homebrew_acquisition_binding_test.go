//go:build linux && (amd64 || arm64)

package execution

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"os"
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
