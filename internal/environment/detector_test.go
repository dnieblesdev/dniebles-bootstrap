package environment

import (
	"errors"
	"testing"
)

func TestDetectorDetect(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		goarch   string
		env      map[string]string
		files    map[string]string
		wantOS   string
		wantArch string
		wantDist string
		wantWSL  bool
	}{
		{
			name:     "maps runtime distro and WSL env signal",
			goos:     "linux",
			goarch:   "amd64",
			env:      map[string]string{"WSL_DISTRO_NAME": "Ubuntu"},
			files:    map[string]string{osReleasePath: "NAME=Ubuntu\nID=ubuntu\n"},
			wantOS:   "linux",
			wantArch: "amd64",
			wantDist: "ubuntu",
			wantWSL:  true,
		},
		{
			name:     "maps WSL kernel signal when env absent",
			goos:     "linux",
			goarch:   "arm64",
			files:    map[string]string{osReleasePath: "ID=debian\n", procKernelOSRelPath: "5.15.90.1-microsoft-standard-WSL2"},
			wantOS:   "linux",
			wantArch: "arm64",
			wantDist: "debian",
			wantWSL:  true,
		},
		{
			name:     "missing optional files falls back conservatively",
			goos:     "linux",
			goarch:   "amd64",
			wantOS:   "linux",
			wantArch: "amd64",
			wantDist: "",
			wantWSL:  false,
		},
		{
			name:     "non linux skips distro and keeps WSL false without evidence",
			goos:     "darwin",
			goarch:   "arm64",
			files:    map[string]string{osReleasePath: "ID=ubuntu\n"},
			wantOS:   "darwin",
			wantArch: "arm64",
			wantDist: "",
			wantWSL:  false,
		},
		{
			name:     "blank runtime values are not invented",
			goos:     "",
			goarch:   "",
			wantOS:   "",
			wantArch: "",
			wantDist: "",
			wantWSL:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := Detector{
				Runtime: func() (string, string) { return tt.goos, tt.goarch },
				Env: func(key string) (string, bool) {
					value, ok := tt.env[key]
					return value, ok
				},
				ReadFile: func(path string) (string, error) {
					content, ok := tt.files[path]
					if !ok {
						return "", errors.New("missing fixture")
					}
					return content, nil
				},
			}

			got := detector.Detect()
			if got.OS != tt.wantOS || got.Arch != tt.wantArch || got.Distro != tt.wantDist || got.WSL != tt.wantWSL {
				t.Fatalf("Detect() = {OS:%q Arch:%q Distro:%q WSL:%t}, want {OS:%q Arch:%q Distro:%q WSL:%t}", got.OS, got.Arch, got.Distro, got.WSL, tt.wantOS, tt.wantArch, tt.wantDist, tt.wantWSL)
			}
		})
	}
}

func TestParseOSReleaseID(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "plain id", content: "ID=ubuntu\n", want: "ubuntu"},
		{name: "double quoted id", content: "NAME=Ubuntu\nID=\"ubuntu\"\n", want: "ubuntu"},
		{name: "single quoted id", content: "ID='debian'\n", want: "debian"},
		{name: "comments and malformed lines ignored", content: "# comment\nNAME\nID=fedora\n", want: "fedora"},
		{name: "missing id", content: "NAME=Ubuntu\n", want: ""},
		{name: "empty id", content: "ID=\n", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseOSReleaseID(tt.content); got != tt.want {
				t.Fatalf("parseOSReleaseID() = %q, want %q", got, tt.want)
			}
		})
	}
}
