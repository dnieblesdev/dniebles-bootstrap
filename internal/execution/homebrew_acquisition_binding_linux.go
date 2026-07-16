//go:build linux

package execution

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"os"
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

type HomebrewAcquisitionResult struct{ PackageDispatchAllowed bool }
type HomebrewAcquirerDependencies struct{}
type HomebrewAcquirer struct{ dependencies HomebrewAcquirerDependencies }

func NewHomebrewAcquirer(dependencies HomebrewAcquirerDependencies) *HomebrewAcquirer {
	return &HomebrewAcquirer{dependencies: dependencies}
}

// Acquire is terminal and package-dispatch-free until CLI orchestration is wired.
func (a *HomebrewAcquirer) Acquire(_ context.Context) HomebrewAcquisitionResult {
	return HomebrewAcquisitionResult{PackageDispatchAllowed: false}
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
