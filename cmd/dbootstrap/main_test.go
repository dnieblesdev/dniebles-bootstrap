package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPlanCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:     "success uses adapter and planner with exact output",
			args:     []string{"plan", "--profile", "dev", "--catalog", "../../catalog/bootstrap.toml"},
			wantCode: exitSuccess,
			wantStdout: "Plan profile: dev\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro= wsl=false\n" +
				"\n" +
				"Steps:\n" +
				"1. tool:git [planned] Version control\n" +
				"   depends_on: none\n" +
				"   attention: none\n" +
				"2. package:ripgrep [planned] Fast text search\n" +
				"   depends_on: tool:git\n" +
				"   attention: none\n" +
				"3. runtime:go [attention_required] Go toolchain\n" +
				"   depends_on: tool:git\n" +
				"   attention: missing required config \"go.env\"\n" +
				"\n" +
				"Results:\n" +
				"- package:ripgrep: planned\n" +
				"- runtime:go: attention_required\n" +
				"  reason: missing required config \"go.env\"\n" +
				"- tool:git: planned\n",
			wantStderr: "",
		},
		{
			name:       "missing profile is a stable usage error",
			args:       []string{"plan"},
			wantCode:   exitUsage,
			wantStdout: "",
			wantStderr: "Usage: dbootstrap plan --profile <name> [--catalog <path>]\nerror: --profile is required\n",
		},
		{
			name:     "unknown profile exits with diagnostics",
			args:     []string{"plan", "--profile", "missing", "--catalog", "../../catalog/bootstrap.toml"},
			wantCode: exitFailure,
			wantStdout: "Plan profile: missing\n" +
				"Catalog: ../../catalog/bootstrap.toml\n" +
				"Environment: os=linux arch=amd64 distro= wsl=false\n" +
				"\n" +
				"Steps:\n" +
				"- none\n" +
				"\n" +
				"Results:\n" +
				"- diagnostic: error\n" +
				"  reason: unknown profile \"missing\"\n",
			wantStderr: "Diagnostics:\n- diagnostic: unknown profile \"missing\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			gotCode := run(tt.args, &stdout, &stderr)

			if gotCode != tt.wantCode {
				t.Fatalf("run() exit code = %d, want %d", gotCode, tt.wantCode)
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

func TestRunPlanCatalogLoadErrors(t *testing.T) {
	tests := []struct {
		name       string
		catalog    string
		wantStderr string
	}{
		{
			name:       "missing catalog path",
			catalog:    filepath.Join(t.TempDir(), "missing.toml"),
			wantStderr: "no such file or directory",
		},
		{
			name:       "invalid catalog input",
			catalog:    writeFile(t, t.TempDir(), "invalid.toml", "[[tools]"),
			wantStderr: "toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			gotCode := run([]string{"plan", "--profile", "dev", "--catalog", tt.catalog}, &stdout, &stderr)

			if gotCode != exitFailure {
				t.Fatalf("run() exit code = %d, want %d", gotCode, exitFailure)
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if !strings.Contains(stderr.String(), "error: load catalog ") || !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr = %q, want load error containing %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestRunUsageErrors(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{
			name: "missing command",
			args: nil,
			wantStderr: "Usage: dbootstrap <command> [options]\n" +
				"\n" +
				"Commands:\n" +
				"  plan    Build a deterministic plan for a profile\n" +
				"error: command is required\n",
		},
		{
			name: "unknown command",
			args: []string{"apply"},
			wantStderr: "Usage: dbootstrap <command> [options]\n" +
				"\n" +
				"Commands:\n" +
				"  plan    Build a deterministic plan for a profile\n" +
				"error: unknown command \"apply\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			gotCode := run(tt.args, &stdout, &stderr)

			if gotCode != exitUsage {
				t.Fatalf("run() exit code = %d, want %d", gotCode, exitUsage)
			}
			if stdout.String() != "" {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", got, tt.wantStderr)
			}
		})
	}
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}
