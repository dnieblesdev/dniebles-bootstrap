//go:build linux

package execution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	memfdAllowSealing = 0x0002
	fAddSeals         = 1033
	fGetSeals         = 1034
	fSealSeal         = 0x0001
	fSealShrink       = 0x0002
	fSealGrow         = 0x0004
	fSealWrite        = 0x0008
	allMemfdSeals     = fSealSeal | fSealShrink | fSealGrow | fSealWrite
)

var ErrHomebrewDigest = errors.New("Homebrew installer digest mismatch")

type binderHooks struct {
	maxWrite    int
	readSize    int
	afterRead   func()
	readErr     error
	writeErr    error
	hashErr     error
	seekErr     error
	sealErr     error
	validateErr error
}

// bindVerified binds Bash execution to the exact bytes written to a sealable memfd.
func bindVerified(staged *os.File, expected [32]byte, hooks binderHooks) (_ *sealedScript, err error) {
	memfd, err := newSealableMemfd("dbootstrap-homebrew")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			memfd.Close()
		}
	}()
	hash := sha256.New()
	bufferSize := 32 * 1024
	if hooks.readSize > 0 && hooks.readSize < bufferSize {
		bufferSize = hooks.readSize
	}
	buffer := make([]byte, bufferSize)
	for {
		if hooks.readErr != nil {
			return nil, hooks.readErr
		}
		n, readErr := staged.Read(buffer)
		if n > 0 && hooks.afterRead != nil {
			hooks.afterRead()
		}
		for written := 0; written < n; {
			if hooks.writeErr != nil {
				return nil, hooks.writeErr
			}
			limit := n - written
			if hooks.maxWrite > 0 && limit > hooks.maxWrite {
				limit = hooks.maxWrite
			}
			m, writeErr := memfd.Write(buffer[written : written+limit])
			if m > 0 {
				if hooks.hashErr != nil {
					return nil, hooks.hashErr
				}
				if _, hashErr := hash.Write(buffer[written : written+m]); hashErr != nil {
					return nil, hashErr
				}
				written += m
			}
			if writeErr != nil {
				return nil, writeErr
			}
			if m == 0 {
				return nil, io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}
	var actual [32]byte
	copy(actual[:], hash.Sum(nil))
	if actual != expected {
		return nil, ErrHomebrewDigest
	}
	if hooks.seekErr != nil {
		return nil, hooks.seekErr
	}
	if _, err := memfd.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	if hooks.sealErr != nil {
		return nil, hooks.sealErr
	}
	if err := addMemfdSeals(memfd.Fd()); err != nil {
		return nil, err
	}
	if hooks.validateErr != nil {
		return nil, hooks.validateErr
	}
	if err := validateMemfdSeals(memfd.Fd()); err != nil {
		return nil, err
	}
	return &sealedScript{File: memfd, Sealed: true}, nil
}

type HomebrewAcquirerDependencies struct {
	Client   httpDoer
	NewStage func(string) (*homebrewStage, error)
	Download func(context.Context, httpDoer, *homebrewStage) error
	Bind     func(*os.File, [32]byte, binderHooks) (*sealedScript, error)
	Runner   CommandRunner
	LookPath func(string) (string, error)
	Stat     func(string) (os.FileInfo, error)
}
type HomebrewAcquirer struct{ dependencies HomebrewAcquirerDependencies }

func NewHomebrewAcquirer(dependencies HomebrewAcquirerDependencies) *HomebrewAcquirer {
	if dependencies.Client == nil {
		dependencies.Client = &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	if dependencies.NewStage == nil {
		dependencies.NewStage = newHomebrewStage
	}
	if dependencies.Download == nil {
		dependencies.Download = downloadPinnedInstaller
	}
	if dependencies.Bind == nil {
		dependencies.Bind = func(staged *os.File, digest [32]byte, _ binderHooks) (*sealedScript, error) {
			return bindVerified(staged, digest, binderHooks{})
		}
	}
	if dependencies.Runner == nil {
		dependencies.Runner = NewOSCommandRunner()
	}
	if dependencies.LookPath == nil {
		dependencies.LookPath = exec.LookPath
	}
	if dependencies.Stat == nil {
		dependencies.Stat = os.Stat
	}
	return &HomebrewAcquirer{dependencies: dependencies}
}

func acquireHomebrewLinux(ctx context.Context) HomebrewAcquisitionResult {
	return NewHomebrewAcquirer(HomebrewAcquirerDependencies{}).Acquire(ctx)
}

// Acquire stages, verifies, executes, and revalidates Homebrew without ever
// invoking a target package installer.
func (a *HomebrewAcquirer) Acquire(ctx context.Context) (result HomebrewAcquisitionResult) {
	stage, err := a.dependencies.NewStage("")
	if err != nil {
		return HomebrewAcquisitionResult{Err: fmt.Errorf("stage Homebrew installer: %w", err)}
	}
	defer func() { result.Err = preservePrimaryError(result.Err, stage.Close()) }()
	if err := a.dependencies.Download(ctx, a.dependencies.Client, stage); err != nil {
		return HomebrewAcquisitionResult{Err: fmt.Errorf("download Homebrew installer: %w", err)}
	}
	staged, err := reopenOwnedStagedFile(stage.Path)
	if err != nil {
		return HomebrewAcquisitionResult{Err: fmt.Errorf("validate staged Homebrew installer: %w", err)}
	}
	defer staged.Close()
	digestBytes, _ := hex.DecodeString(homebrewInstallerDigest)
	var digest [32]byte
	copy(digest[:], digestBytes)
	script, err := a.dependencies.Bind(staged, digest, binderHooks{})
	if err != nil {
		return HomebrewAcquisitionResult{Err: fmt.Errorf("verify Homebrew installer: %w", err)}
	}
	defer script.Close()
	if command := executeSealedScript(ctx, a.dependencies.Runner, script); command.Status != CommandStatusSucceeded {
		return HomebrewAcquisitionResult{Err: fmt.Errorf("execute Homebrew installer: %w", command.Err)}
	}
	brew, err := a.lookupBrew()
	if err != nil {
		return HomebrewAcquisitionResult{Err: err}
	}
	if command := a.dependencies.Runner.RunCommand(ctx, CommandRequest{Executable: brew, Args: []string{"--version"}}); command.Status != CommandStatusSucceeded {
		return HomebrewAcquisitionResult{Err: fmt.Errorf("revalidate Homebrew: %w", command.Err)}
	}
	return HomebrewAcquisitionResult{Acquired: true}
}

func (a *HomebrewAcquirer) lookupBrew() (string, error) {
	if isExecutableRegularFile(a.dependencies.Stat, homebrewDefaultBinary) {
		return homebrewDefaultBinary, nil
	}
	path, err := a.dependencies.LookPath("brew")
	if err == nil && filepath.IsAbs(path) && isExecutableRegularFile(a.dependencies.Stat, path) {
		return path, nil
	}
	return "", errors.New("Homebrew was not found after installer execution")
}

func isExecutableRegularFile(stat func(string) (os.FileInfo, error), path string) bool {
	info, err := stat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
}

func newSealableMemfd(name string) (*os.File, error) {
	nameBytes, err := syscall.BytePtrFromString(name)
	if err != nil {
		return nil, err
	}
	fd, _, errno := syscall.Syscall(memfdCreateSyscall(), uintptr(unsafe.Pointer(nameBytes)), uintptr(memfdAllowSealing), 0)
	if errno != 0 {
		return nil, errno
	}
	return os.NewFile(fd, name), nil
}

func addMemfdSeals(fd uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_FCNTL, fd, fAddSeals, allMemfdSeals)
	if errno != 0 {
		return errno
	}
	return nil
}

func validateMemfdSeals(fd uintptr) error {
	seals, _, errno := syscall.Syscall(syscall.SYS_FCNTL, fd, fGetSeals, 0)
	if errno != 0 {
		return errno
	}
	if seals&allMemfdSeals != allMemfdSeals {
		return errors.New("memfd seals are incomplete")
	}
	return nil
}
