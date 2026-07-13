package version

import (
	"os/exec"
	"testing"
)

func TestNormalizeCmd(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"normal tag", "v1.2.3", "v1.2.3"},
		{"slash branch", "feature/new-thing", "feature-new-thing"},
		{"empty", "", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", "./cmd/normalize", "--version", tt.in)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go run ./cmd/normalize --version %q: %v\n%s", tt.in, err, out)
			}
			got := string(out)
			// CombinedOutput includes a trailing newline from fmt.Println.
			if len(got) > 0 && got[len(got)-1] == '\n' {
				got = got[:len(got)-1]
			}
			if got != tt.want {
				t.Fatalf("normalize output = %q, want %q", got, tt.want)
			}
		})
	}
}
