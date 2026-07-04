package environment

import (
	"os"
	"runtime"
	"strings"

	"github.com/dnieblesdev/dniebles-bootstrap/internal/planning"
)

const (
	osReleasePath       = "/etc/os-release"
	procVersionPath     = "/proc/version"
	procKernelOSRelPath = "/proc/sys/kernel/osrelease"
)

type RuntimeSource func() (goos, goarch string)
type EnvSource func(key string) (string, bool)
type FileSource func(path string) (string, error)

type Detector struct {
	Runtime  RuntimeSource
	Env      EnvSource
	ReadFile FileSource
}

func Detect() planning.EnvironmentFacts {
	return Detector{}.Detect()
}

func (d Detector) Detect() planning.EnvironmentFacts {
	runtimeSource := d.Runtime
	if runtimeSource == nil {
		runtimeSource = defaultRuntimeSource
	}
	envSource := d.Env
	if envSource == nil {
		envSource = os.LookupEnv
	}
	fileSource := d.ReadFile
	if fileSource == nil {
		fileSource = defaultFileSource
	}

	goos, goarch := runtimeSource()
	facts := planning.EnvironmentFacts{OS: goos, Arch: goarch}
	if goos == "linux" {
		facts.Distro = detectDistro(fileSource)
	}
	facts.WSL = detectWSL(envSource, fileSource)
	return facts
}

func defaultRuntimeSource() (string, string) {
	return runtime.GOOS, runtime.GOARCH
}

func defaultFileSource(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func detectDistro(readFile FileSource) string {
	content, err := readFile(osReleasePath)
	if err != nil {
		return ""
	}
	return parseOSReleaseID(content)
}

func detectWSL(env EnvSource, readFile FileSource) bool {
	for _, key := range []string{"WSL_DISTRO_NAME", "WSL_INTEROP"} {
		if value, ok := env(key); ok && value != "" {
			return true
		}
	}

	for _, path := range []string{procVersionPath, procKernelOSRelPath} {
		content, err := readFile(path)
		if err != nil {
			continue
		}
		text := strings.ToLower(content)
		if strings.Contains(text, "microsoft") || strings.Contains(text, "wsl") {
			return true
		}
	}
	return false
}
