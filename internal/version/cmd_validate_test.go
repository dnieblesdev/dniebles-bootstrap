package version

import (
	"os/exec"
	"testing"
)

func TestValidateCmdRelease(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		wantPrerelease string
		wantErr        bool
	}{
		{"stable", "v1.2.3", "prerelease=false", false},
		{"prerelease", "v1.2.3-rc.1", "prerelease=true", false},
		{"build metadata only", "v1.2.3+build.123", "prerelease=false", false},
		{"unprefixed", "1.2.3", "", true},
		{"partial", "v1.2", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", "./cmd/validate", "--version", tt.version, "-release")
			out, err := cmd.CombinedOutput()
			got := string(out)
			if len(got) > 0 && got[len(got)-1] == '\n' {
				got = got[:len(got)-1]
			}
			if tt.wantErr {
				if err == nil {
					t.Fatalf("go run ./cmd/validate -release --version %q: want error, got %q", tt.version, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("go run ./cmd/validate -release --version %q: %v\n%s", tt.version, err, out)
			}
			if got != tt.wantPrerelease {
				t.Fatalf("go run ./cmd/validate -release --version %q = %q, want %q", tt.version, got, tt.wantPrerelease)
			}
		})
	}
}

func TestValidateCmd(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"valid version", "v1.2.3", false},
		{"invalid version", "v1.2.3; rm -rf /", true},
		{"empty version", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"run", "./cmd/validate"}
			if tt.version != "" {
				args = append(args, "--version", tt.version)
			}
			cmd := exec.Command("go", args...)
			err := cmd.Run()
			if tt.wantErr && err == nil {
				t.Fatalf("go run ./cmd/validate --version %q: want error, got nil", tt.version)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("go run ./cmd/validate --version %q: want nil, got %v", tt.version, err)
			}
		})
	}
}
