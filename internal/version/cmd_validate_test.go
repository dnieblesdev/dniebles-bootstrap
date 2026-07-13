package version

import (
	"os/exec"
	"testing"
)

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
