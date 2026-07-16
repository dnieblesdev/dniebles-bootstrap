//go:build linux

package execution

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
)

var ErrUnsafeHomebrewStage = errors.New("Homebrew staging file is unsafe")

type homebrewStage struct {
	Dir  string
	Path string
}

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

func newHomebrewStage(parent string) (*homebrewStage, error) {
	dir, err := os.MkdirTemp(parent, "dbootstrap-homebrew-")
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	path := filepath.Join(dir, "installer.sh")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	if err := file.Close(); err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	return &homebrewStage{Dir: dir, Path: path}, nil
}

func (s *homebrewStage) Close() error { return os.RemoveAll(s.Dir) }

func reopenOwnedStagedFile(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return nil, ErrUnsafeHomebrewStage
	}
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, fmt.Errorf("open staged installer without following links: %w", err)
	}
	info, err = file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || int(stat.Uid) != os.Getuid() {
		file.Close()
		return nil, ErrUnsafeHomebrewStage
	}
	return file, nil
}

// downloadPinnedInstaller sends exactly one literal HTTPS request. Its caller
// supplies a redirect-disabled client; the effective URL check is a second,
// fail-closed guard against redirect-following transports.
func downloadPinnedInstaller(ctx context.Context, client httpDoer, stage *homebrewStage) error {
	if stage == nil {
		return ErrUnsafeHomebrewStage
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, homebrewInstallerURL, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.Request == nil || response.Request.URL == nil {
		return errors.New("Homebrew installer download failed")
	}
	if err := validatePinnedDownload(req.URL.String(), response.Request.URL.String()); err != nil {
		return err
	}
	file, err := os.OpenFile(stage.Path, os.O_WRONLY|os.O_TRUNC|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return fmt.Errorf("open private stage for download: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || int(stat.Uid) != os.Getuid() {
		file.Close()
		return ErrUnsafeHomebrewStage
	}
	_, copyErr := io.Copy(file, response.Body)
	closeErr := file.Close()
	return preservePrimaryError(copyErr, closeErr)
}

type sealedScript struct {
	File   *os.File
	Sealed bool
}

func (s *sealedScript) Close() error {
	if s == nil || s.File == nil {
		return nil
	}
	return s.File.Close()
}

func bashRequestForSealedScript(script *sealedScript) (CommandRequest, []*os.File, error) {
	if script == nil || script.File == nil || !script.Sealed {
		return CommandRequest{}, nil, errors.New("sealed Homebrew installer is unavailable")
	}
	return CommandRequest{Executable: "/bin/bash", Args: []string{"/proc/self/fd/3"}}, []*os.File{script.File}, nil
}

// executeSealedScript is the sole Bash boundary. Validation happens before
// invoking the runner, and only the sealed memfd is inherited as child FD 3.
func executeSealedScript(ctx context.Context, runner CommandRunner, script *sealedScript) CommandResult {
	req, files, err := bashRequestForSealedScript(script)
	if err != nil {
		return CommandResult{Status: CommandStatusNotRun, ExitCode: -1, Err: err}
	}
	req.ExtraFiles = files
	return runner.RunCommand(ctx, req)
}

func preservePrimaryError(primary, cleanup error) error {
	if cleanup == nil {
		return primary
	}
	if primary == nil {
		return cleanup
	}
	return fmt.Errorf("%w (cleanup: %v)", primary, cleanup)
}
